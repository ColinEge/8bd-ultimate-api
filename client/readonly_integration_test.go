package client_test

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ColinEge/8bd-kbd-api/client"
	"github.com/ColinEge/8bd-kbd-api/keydefs"
	"github.com/ColinEge/8bd-kbd-api/transport"
	"github.com/google/gousb"
)

func readOnlyHardwareClient(t *testing.T) client.Client {
	t.Helper()
	if os.Getenv("TEST_8BITDO_READONLY") != "1" {
		t.Skip("set TEST_8BITDO_READONLY=1 to enable read-only hardware tests")
	}
	c := client.New(client.WithTimeout(10*time.Second), client.WithTransportFactory(
		func(in *gousb.InEndpoint, out *gousb.OutEndpoint, dev *gousb.Device, cfg *gousb.Config, intfs []*gousb.Interface) transport.Transport {
			usb := transport.NewUSB(in, out, dev, cfg, intfs)
			return &transport.Mock{
				WriteFunc: usb.Write,
				ReadFunc: func(ctx context.Context) ([]byte, error) {
					packet, err := usb.Read(ctx)
					t.Logf("action response (%d bytes): % x; error=%v", len(packet), packet, err)
					return packet, err
				},
				CloseFunc: usb.Close,
			}
		},
	))
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Errorf("close client: %v", err)
		}
	})
	return c
}

func TestIntegration_ReadOnlyScan(t *testing.T) {
	c := readOnlyHardwareClient(t)
	devices, err := c.Scan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) == 0 {
		t.Fatal("no keyboards found")
	}
	for _, d := range devices {
		if d.ID == "" || d.VendorID != 0x2dc8 ||
			(d.Product != client.ProductTypeKeyboard && d.Product != client.ProductType108Keyboard) {
			t.Fatalf("unexpected device metadata: %+v", d)
		}
		t.Logf("device=%s model=%s", d.ID, d.Product)
	}
}

func TestIntegration_ReadOnlyAction(t *testing.T) {
	c := readOnlyHardwareClient(t)
	id := client.DeviceID(os.Getenv("TEST_8BITDO_DEVICE"))
	keyText := os.Getenv("TEST_8BITDO_KEY")
	if id == "" || keyText == "" {
		t.Fatal("TEST_8BITDO_DEVICE and TEST_8BITDO_KEY are required for action read-back")
	}
	base := 10
	if strings.HasPrefix(keyText, "0x") || strings.HasPrefix(keyText, "0X") {
		base = 0
	}
	key, err := strconv.ParseUint(keyText, base, 8)
	if err != nil {
		t.Fatalf("invalid TEST_8BITDO_KEY: %v", err)
	}
	expectedText, hasExpected := os.LookupEnv("TEST_8BITDO_EXPECT_ACTION")
	if hasExpected && expectedText == "unbound" {
		expectedText = "00-00-00"
	}
	var expected keydefs.ActionCode
	if hasExpected {
		if n := len(strings.Split(expectedText, "-")); n < 3 || n > 6 {
			t.Fatal("TEST_8BITDO_EXPECT_ACTION must contain 3 to 6 hex bytes")
		}
		expected, err = keydefs.ParseActionCode(expectedText)
		if err != nil {
			t.Fatalf("invalid TEST_8BITDO_EXPECT_ACTION: %v", err)
		}
	}
	err = c.WithKeyboard(context.Background(), id, func(s client.Keyboard) error {
		action, mapped, err := readAction(context.Background(), s, keydefs.KeyID(key))
		if err != nil {
			return err
		}
		t.Logf("device=%s key=0x%02x action=%s mapped=%t", id, key, action, mapped)
		if hasExpected && action != expected {
			t.Errorf("got action %s, want %s", action, expected)
		}
		if hasExpected && mapped != (expected != (keydefs.ActionCode{})) {
			t.Errorf("mapped=%t does not match expected action %s", mapped, expected)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
