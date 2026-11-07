package transport

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/ColinEge/8bd-kbd-api/protocol"
	"github.com/google/gousb"
)

type USB struct {
	epIn   *gousb.InEndpoint
	epOut  *gousb.OutEndpoint
	device *gousb.Device
	config *gousb.Config
	ifaces []*gousb.Interface
}

func NewUSB(epIn *gousb.InEndpoint, epOut *gousb.OutEndpoint, device *gousb.Device, config *gousb.Config, ifaces []*gousb.Interface) *USB {
	return &USB{epIn: epIn, epOut: epOut, device: device, config: config, ifaces: ifaces}
}

func (u *USB) Write(ctx context.Context, data []byte) error {
	num, err := u.epOut.WriteContext(ctx, data)
	if err != nil {
		return fmt.Errorf("usb write %d bytes failed: %w", num, err)
	}
	if num != len(data) {
		return fmt.Errorf("usb write %d of %d bytes: %w", num, len(data), io.ErrShortWrite)
	}
	return nil
}

func (u *USB) Read(ctx context.Context) ([]byte, error) {
	buf := make([]byte, protocol.ReportSize)
	num, err := u.epIn.ReadContext(ctx, buf)
	if err != nil {
		return nil, fmt.Errorf("usb read %d bytes failed: %w", num, err)
	}
	return buf[:num], nil
}

func (u *USB) Close() error {
	var errs []error
	for _, i := range u.ifaces {
		i.Close()
	}
	if e := u.config.Close(); e != nil {
		errs = append(errs, e)
	}
	if e := u.device.Close(); e != nil {
		errs = append(errs, e)
	}
	if len(errs) > 0 {
		return fmt.Errorf("usb close errors: %w", errors.Join(errs...))
	}
	return nil
}
