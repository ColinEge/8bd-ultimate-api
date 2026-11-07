package keyboard

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/ColinEge/8bd-ultimate-api/client"
	"github.com/ColinEge/8bd-ultimate-api/transport"
)

type Exchange struct {
	Write    []byte
	Response []byte
	WriteErr error
	ReadErr  error
}

type KeyboardFixture struct {
	Keyboard  client.Client
	Device    client.DeviceInfo
	Connected bool
	Err       error

	Exchanges        []Exchange
	Writes           [][]byte
	Reads            int
	SessionCloses    int
	ProviderCloses   int
	SessionCloseErr  error
	ProviderCloseErr error
	OpenErr          error
	exchangeIndex    int
	awaitingRead     bool
	scriptErr        error
}

func (f *KeyboardFixture) Reset() error {
	if err := f.Close(); err != nil {
		return err
	}
	*f = KeyboardFixture{Connected: true}
	f.Keyboard = client.New(client.WithDeviceProvider(&provider{fixture: f}))
	devices, err := f.Keyboard.Scan(context.Background())
	if err != nil {
		return err
	}

	f.Device = devices[0]
	return nil
}

func (f *KeyboardFixture) Close() error {
	if f.Keyboard == nil {
		return nil
	}
	if err := f.Keyboard.Close(); err != nil {
		return err
	}
	f.Keyboard = nil
	f.Connected = false
	return nil
}

func (f *KeyboardFixture) RequireConnected() error {
	if f.Keyboard == nil {
		return fmt.Errorf("keyboard fixture has not been initialized")
	}

	if !f.Connected || f.Device == (client.DeviceInfo{}) {
		return fmt.Errorf("no keyboard connected")
	}

	return nil
}

func (f *KeyboardFixture) AssertComplete() error {
	if f.scriptErr != nil {
		return f.scriptErr
	}
	if f.awaitingRead || f.exchangeIndex != len(f.Exchanges) {
		return fmt.Errorf("incomplete script: consumed %d of %d exchanges, awaiting read: %t",
			f.exchangeIndex, len(f.Exchanges), f.awaitingRead)
	}
	return nil
}

func (f *KeyboardFixture) write(data []byte) error {
	f.Writes = append(f.Writes, bytes.Clone(data))
	if f.awaitingRead || f.exchangeIndex >= len(f.Exchanges) {
		f.scriptErr = fmt.Errorf("unexpected write %x", data)
		return f.scriptErr
	}
	exchange := f.Exchanges[f.exchangeIndex]
	if !bytes.Equal(data, exchange.Write) {
		f.scriptErr = fmt.Errorf("write mismatch: got %x, want %x", data, exchange.Write)
		return f.scriptErr
	}
	if exchange.WriteErr != nil {
		f.exchangeIndex++
		return exchange.WriteErr
	}
	f.awaitingRead = true
	return nil
}

func (f *KeyboardFixture) read(ctx context.Context) ([]byte, error) {
	f.Reads++
	if !f.awaitingRead {
		f.scriptErr = errors.New("unexpected read without a matching write")
		return nil, f.scriptErr
	}
	exchange := f.Exchanges[f.exchangeIndex]
	f.exchangeIndex++
	f.awaitingRead = false
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return bytes.Clone(exchange.Response), exchange.ReadErr
}

type provider struct {
	fixture *KeyboardFixture
	closed  bool
}

func (p *provider) ListDevices(ctx context.Context) ([]client.DeviceInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if p.closed {
		return nil, client.ErrDeviceClosed
	}
	if p.fixture.Err != nil {
		return nil, p.fixture.Err
	}
	if !p.fixture.Connected {
		return []client.DeviceInfo{}, nil
	}
	return []client.DeviceInfo{{
		ID: "001-003", VendorID: 0x2dc8, Product: client.ProductTypeKeyboard,
		Bus: 1, Address: 3,
	}}, nil
}

func (p *provider) OpenDevice(ctx context.Context, id client.DeviceID) (*client.OpenedDevice, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if p.closed {
		return nil, client.ErrDeviceClosed
	}
	f := p.fixture
	if f.OpenErr != nil {
		return nil, f.OpenErr
	}
	if !f.Connected || id != f.Device.ID {
		return nil, client.ErrDeviceNotFound
	}
	closed := false
	return &client.OpenedDevice{
		Info: f.Device,
		Transport: &transport.Mock{
			WriteFunc: func(ctx context.Context, data []byte) error {
				if closed {
					return client.ErrDeviceClosed
				}
				return f.write(data)
			},
			ReadFunc: func(ctx context.Context) ([]byte, error) {
				if closed {
					return nil, client.ErrDeviceClosed
				}
				return f.read(ctx)
			},
			CloseFunc: func() error {
				f.SessionCloses++
				closed = true
				return f.SessionCloseErr
			},
		},
	}, nil
}

func (p *provider) Close() error {
	if p.closed {
		return nil
	}
	p.fixture.ProviderCloses++
	if err := p.fixture.ProviderCloseErr; err != nil {
		return err
	}
	p.closed = true
	return nil
}
