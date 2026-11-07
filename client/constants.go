package client

import "fmt"

// Product is a specific 8BitDo device type
type Product uint16

const (
	ProductTypeKeyboard    Product = 0x5200
	ProductType108Keyboard Product = 0x5209
)

func (p Product) String() string {
	switch p {
	case ProductTypeKeyboard:
		return "8BitDo Retro Mechanical Keyboard"
	case ProductType108Keyboard:
		return "8BitDo Retro 108 Mechanical Keyboard"
	default:
		return fmt.Sprintf("Unknown product 0x%x", uint16(p))
	}
}

type Vendor uint16

const (
	Vendor8BitDo Vendor = 0x2dc8
)

func (v Vendor) String() string {
	switch v {
	case Vendor8BitDo:
		return "8BitDo"
	default:
		return fmt.Sprintf("Unknown vendor 0x%x", uint16(v))
	}
}
