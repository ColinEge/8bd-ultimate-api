package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/ColinEge/8bd-kbd-api/transport"
	"github.com/google/gousb"
)

type OpenedDevice struct {
	Transport transport.Transport
	Info      DeviceInfo
}

type DeviceProvider interface {
	// ListDevices returns available devices (metadata only)
	ListDevices(ctx context.Context) ([]DeviceInfo, error)
	// OpenDevice returns a Transport connected to the given device ID
	OpenDevice(ctx context.Context, id DeviceID) (*OpenedDevice, error)
	Close() error
}

type gousbProvider struct {
	ctx              *gousb.Context
	transportFactory TransportFactory
}

func (p *gousbProvider) Close() error {
	return p.ctx.Close()
}

// ListDevices enumerates devices without keeping handles open
func (p *gousbProvider) ListDevices(ctx context.Context) ([]DeviceInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	devices, err := p.ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		return desc != nil &&
			desc.Vendor == gousb.ID(Vendor8BitDo) &&
			(desc.Product == gousb.ID(ProductTypeKeyboard) ||
				desc.Product == gousb.ID(ProductType108Keyboard))
	})
	if err != nil {
		return nil, errors.Join(err, closeDevices(devices))
	}

	if len(devices) == 0 {
		return []DeviceInfo{}, nil
	}

	var infos []DeviceInfo
	var closeErr error
	for _, d := range devices {
		infos = append(infos, DeviceInfo{
			VendorID: uint16(d.Desc.Vendor),
			Product:  Product(d.Desc.Product),
			ID:       deviceIDFromDevice(d),
			Bus:      d.Desc.Bus,
			Address:  d.Desc.Address,
		})
		closeErr = errors.Join(closeErr, d.Close())
	}
	if err := errors.Join(ctx.Err(), closeErr); err != nil {
		return nil, err
	}
	return infos, nil
}

// OpenDevice claims the USB device and returns a transport.Transport
func (p *gousbProvider) OpenDevice(ctx context.Context, id DeviceID) (*OpenedDevice, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Open all devices
	devices, err := p.ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		return desc != nil &&
			desc.Vendor == gousb.ID(Vendor8BitDo) &&
			(desc.Product == gousb.ID(ProductTypeKeyboard) ||
				desc.Product == gousb.ID(ProductType108Keyboard))
	})
	if err != nil {
		return nil, errors.Join(err, closeDevices(devices))
	}

	// Find the device matching the ID
	var dev *gousb.Device
	var closeErr error
	for _, d := range devices {
		if deviceIDFromDevice(d) == id {
			dev = d
		} else {
			closeErr = errors.Join(closeErr, d.Close())
		}
	}
	if dev == nil {
		return nil, errors.Join(ErrDeviceNotFound, closeErr)
	}
	if err := errors.Join(ctx.Err(), closeErr); err != nil {
		return nil, errors.Join(err, dev.Close())
	}

	dev.SetAutoDetach(true)

	cfgNum, err := dev.ActiveConfigNum()
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to get active config: %w", err), dev.Close())
	}

	cfg, err := dev.Config(cfgNum)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to open config %d: %w", cfgNum, err), dev.Close())
	}

	// Claim interfaces
	var intfs []*gousb.Interface
	for _, desc := range cfg.Desc.Interfaces {
		intf, err := cfg.Interface(desc.Number, 0)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("%w: interface %d: %w", ErrInterfaceClaimFailed, desc.Number, err), cleanupInterfaces(intfs, cfg, dev))
		}
		intfs = append(intfs, intf)
	}

	// Use the keyboard interface (e.g., the 3rd)
	if len(intfs) <= 2 {
		return nil, errors.Join(fmt.Errorf("expected at least 3 interfaces, got %d", len(intfs)), cleanupInterfaces(intfs, cfg, dev))
	}
	kbdIntf := intfs[2]

	// Find endpoints
	var epIn *gousb.InEndpoint
	var epOut *gousb.OutEndpoint
	for _, desc := range kbdIntf.Setting.Endpoints {
		if desc.Direction == gousb.EndpointDirectionIn && epIn == nil {
			epIn, err = kbdIntf.InEndpoint(desc.Number)
			if err != nil {
				return nil, errors.Join(fmt.Errorf("%w: endpoint %d: %w", ErrEndpointsNotFound, desc.Number, err), cleanupInterfaces(intfs, cfg, dev))
			}
		} else if desc.Direction == gousb.EndpointDirectionOut && epOut == nil {
			epOut, err = kbdIntf.OutEndpoint(desc.Number)
			if err != nil {
				return nil, errors.Join(fmt.Errorf("%w: endpoint %d: %w", ErrEndpointsNotFound, desc.Number, err), cleanupInterfaces(intfs, cfg, dev))
			}
		}
	}

	if epIn == nil || epOut == nil {
		return nil, errors.Join(ErrEndpointsNotFound, cleanupInterfaces(intfs, cfg, dev))
	}
	if err := ctx.Err(); err != nil {
		return nil, errors.Join(err, cleanupInterfaces(intfs, cfg, dev))
	}

	// Return a USB transport
	tr := p.transportFactory(epIn, epOut, dev, cfg, intfs)
	if tr == nil {
		return nil, errors.Join(fmt.Errorf("transport factory returned nil"), cleanupInterfaces(intfs, cfg, dev))
	}
	return &OpenedDevice{
		Transport: tr,
		Info: DeviceInfo{
			ID: id, VendorID: uint16(dev.Desc.Vendor), Product: Product(dev.Desc.Product),
			Bus: dev.Desc.Bus, Address: dev.Desc.Address,
		},
	}, nil
}

func cleanupInterfaces(intfs []*gousb.Interface, cfg *gousb.Config, dev *gousb.Device) error {
	var err error
	for _, i := range intfs {
		i.Close()
	}
	if cfg != nil {
		err = errors.Join(err, cfg.Close())
	}
	if dev != nil {
		err = errors.Join(err, dev.Close())
	}
	return err
}

func closeDevices(devices []*gousb.Device) error {
	var err error
	for _, device := range devices {
		err = errors.Join(err, device.Close())
	}
	return err
}
