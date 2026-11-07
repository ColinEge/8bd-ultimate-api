package client

import "errors"

// USB / Device errors
var (
	ErrInvalidKey        = errors.New("invalid physical key")
	ErrInvalidAction     = errors.New("invalid action")
	ErrUnsupported       = errors.New("unsupported operation or action")
	ErrMalformedResponse = errors.New("malformed device response")
	ErrDeviceRejected    = errors.New("device rejected operation")
	ErrNilContext        = errors.New("gousb context is nil")

	ErrNotDetected          = errors.New("no 8BitDo Device found")
	ErrDeviceNotFound       = errors.New("device not found")
	ErrEndpointsNotFound    = errors.New("required USB endpoints not found")
	ErrInterfaceClaimFailed = errors.New("failed to claim USB interface")
	ErrDeviceClosed         = errors.New("operation on closed device")

	ErrInvalidUsageCode = errors.New("invalid usage code")
	ErrUSBWriteFailed   = errors.New("USB write failed")
	ErrUSBReadFailed    = errors.New("USB read failed")
	ErrUnexpectedPrefix = errors.New("unexpected prefix")

	ErrProfileQueryFailed    = errors.New("failed to query profile information")
	ErrMappingQueryFailed    = errors.New("failed to query key mappings")
	ErrProfileMutationFailed = errors.New("profile mutation failed")
	ErrInvalidProfileName    = errors.New("invalid profile name")

	Err = errors.New("")
)
