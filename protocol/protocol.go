package protocol

// Packet first byte shows direction of the packet
// 54 comes from the keyboard, 52 goes to it.
type Direction byte

const (
	DirIn  Direction = 0x54 // in from keyboard
	DirOut Direction = 0x52 // out to keyboard
)

func (d Direction) String() string {
	switch d {
	case DirIn:
		return "In"
	case DirOut:
		return "Out"
	default:
		return "UnknownDirection"
	}
}

// Packet second byte shows the command group
type command byte

const (
	CommandTransaction command = 0x7b

	CommandWriteProfile    command = 0x70
	CommandReadProfile     command = 0x80
	CommandRecordMacro     command = 0x76
	CommandReadRemapedKeys command = 0x81
	CommandReadKeyAction   command = 0x83
	CommandWriteVolume     command = 0x79

	CommandBindAction command = 0xfa // fa 03 0c 00 aa 09 71
	CommandBindMacro  command = 0x74
)

func (c command) String() string {
	switch c {
	case CommandWriteProfile:
		return "Write Profile"
	case CommandReadProfile:
		return "Read Profile"
	case CommandReadRemapedKeys:
		return "Get Remaped Keys"
	case CommandReadKeyAction:
		return "Get Key Action"
	default:
		return "Unknown Group"
	}
}

const (
	ReportSize = 33
)
