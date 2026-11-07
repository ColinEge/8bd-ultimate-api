package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ColinEge/8bd-kbd-api/transport"
	"github.com/google/gousb"
)

// DeviceID uniquely identifies a connected device by bus-address string.
type DeviceID string

// deviceIDFromDevice constructs a DeviceID from a gousb.Device
func deviceIDFromDevice(d *gousb.Device) DeviceID {
	return DeviceID(fmt.Sprintf("%03d-%03d", d.Desc.Bus, d.Desc.Address))
}

// // deviceIDFromDevice constructs a DeviceID from raw Bus/Address ints
func deviceIDFromBus(bus, addr int) DeviceID {
	return DeviceID(fmt.Sprintf("%03d-%03d", bus, addr))
}

// DeviceInfo is details
type DeviceInfo struct {
	ID       DeviceID // e.g. "001-003"
	VendorID uint16
	Product  Product
	Bus      int
	Address  int
}

type Client interface {
	// Scan finds all connected 8BitDo Devices
	Scan(ctx context.Context) ([]DeviceInfo, error)
	// Connect starts a session with the device info
	Connect(ctx context.Context, id DeviceID) (Keyboard, error)
	// WithKeyboard opens a short-lived connection to the keyboard identified by id,
	// executes the function f, and ensures the session is closed afterward.
	// Any panics inside the function are recovered and returned as an error.
	WithKeyboard(ctx context.Context, id DeviceID, f func(Keyboard) error) error
	Close() error
}

type TransportFactory func(
	epRead *gousb.InEndpoint,
	epWrite *gousb.OutEndpoint,
	dev *gousb.Device,
	cfg *gousb.Config,
	intfs []*gousb.Interface,
) transport.Transport

type ClientOption func(*client)

// Client wraps a Transport and provides higher-level methods
type client struct {
	provider         DeviceProvider
	timeout          time.Duration
	transportFactory TransportFactory
	gate             chan struct{}
	closed           bool
	keyboards        map[*keyboard]struct{}
}

// New creates a new client that works out of the box
// any default settings can be changed with ClientOption
// with some default settings:
//   - provider is gousbProvider
//   - transport is transport.NewUSB()
//   - timeout is 1 minute
func New(opts ...ClientOption) Client {
	c := &client{
		transportFactory: func(epRead *gousb.InEndpoint, epWrite *gousb.OutEndpoint, dev *gousb.Device, cfg *gousb.Config, intfs []*gousb.Interface) transport.Transport {
			return transport.NewUSB(epRead, epWrite, dev, cfg, intfs)
		},
		timeout:   1 * time.Minute,
		gate:      make(chan struct{}, 1),
		keyboards: make(map[*keyboard]struct{}),
	}

	for _, o := range opts {
		o(c)
	}

	if c.provider == nil {
		c.provider = &gousbProvider{ctx: gousb.NewContext(), transportFactory: c.transportFactory}
	}
	c.gate <- struct{}{}
	return c
}

// WithDeviceProvider replaces USB discovery and connection handling.
// The client owns the provider and closes it when Close is called.
func WithDeviceProvider(provider DeviceProvider) ClientOption {
	return func(c *client) {
		c.provider = provider
	}
}

func WithTransportFactory(tf TransportFactory) ClientOption {
	return func(c *client) {
		c.transportFactory = tf
	}
}

func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *client) {
		c.timeout = timeout
	}
}

func (c *client) Close() error {
	<-c.gate
	defer func() { c.gate <- struct{}{} }()
	if c.closed {
		return nil
	}
	var err error
	for keyboard := range c.keyboards {
		err = errors.Join(err, keyboard.Close())
		delete(c.keyboards, keyboard)
	}
	if closeErr := c.provider.Close(); closeErr != nil {
		return errors.Join(err, closeErr)
	}
	c.closed = true
	return err
}

func (c *client) acquire(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.gate:
	}
	if c.closed {
		c.gate <- struct{}{}
		return ErrDeviceClosed
	}
	if err := ctx.Err(); err != nil {
		c.gate <- struct{}{}
		return err
	}
	return nil
}

func (c *client) Scan(ctx context.Context) ([]DeviceInfo, error) {
	if err := c.acquire(ctx); err != nil {
		return nil, err
	}
	defer func() { c.gate <- struct{}{} }()
	infos, err := c.provider.ListDevices(ctx)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return infos, nil
}

func (c *client) Connect(ctx context.Context, id DeviceID) (Keyboard, error) {
	if err := c.acquire(ctx); err != nil {
		return nil, err
	}
	defer func() { c.gate <- struct{}{} }()
	if id == "" {
		return nil, fmt.Errorf("%w: empty device ID", ErrDeviceNotFound)
	}
	if c.timeout <= 0 {
		return nil, fmt.Errorf("operation timeout must be positive")
	}
	dev, err := c.provider.OpenDevice(ctx, id)
	if err != nil {
		return nil, err
	}

	if dev == nil || dev.Transport == nil {
		return nil, fmt.Errorf("device provider returned no transport")
	}
	if dev.Info.ID != id || dev.Info.VendorID != uint16(Vendor8BitDo) ||
		(dev.Info.Product != ProductTypeKeyboard && dev.Info.Product != ProductType108Keyboard) {
		return nil, errors.Join(fmt.Errorf("%w: provider returned invalid device metadata", ErrUnsupported), dev.Transport.Close())
	}
	if err := ctx.Err(); err != nil {
		return nil, errors.Join(err, dev.Transport.Close())
	}
	kbd := &keyboard{
		info:      KeyboardInfo{Device: dev.Info},
		transport: dev.Transport,
		timeout:   c.timeout,
		gate:      make(chan struct{}, 1),
	}
	kbd.gate <- struct{}{}
	c.keyboards[kbd] = struct{}{}
	return kbd, nil
}

func (c *client) WithKeyboard(ctx context.Context, id DeviceID, f func(Keyboard) error) (err error) {
	if f == nil {
		return fmt.Errorf("keyboard callback must not be nil")
	}
	sess, err := c.Connect(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to connect to device %s: %w", id, err)
	}

	defer func() {
		if r := recover(); r != nil {
			if panicErr, ok := r.(error); ok {
				err = fmt.Errorf("panic in session lambda: %w", panicErr)
			} else {
				err = fmt.Errorf("panic in session lambda: %v", r)
			}
		}
		if closeErr := sess.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to close session: %w", closeErr))
		}
		<-c.gate
		delete(c.keyboards, sess.(*keyboard))
		c.gate <- struct{}{}
	}()

	return f(sess)
}
