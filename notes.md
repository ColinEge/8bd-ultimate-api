# Protocol notes
All messages should be padded to 33 bytes, for this document, padding is omitted for the sake of brevity

## Profile Write
Profile write command is:
`0x52 0x70 <len> <pkt>`
- len is the number of bytes in this payload
- pkt is the number of packets remaining for this command

### Update
Send command with your desired name UTF16BE encoded

Single packet:
- `0x52 0x70 0x0e 0x00 0x00 0x50 0x00 0x72 0x00 0x30 0x00 0x66 0x00 0x69 0x00 0x6C 0x00 0x65` (Pr0file)

Multi packet:
- `0x52 0x70 0x1d 0x01 0x00 0x74 0x00 0x68 0x00 0x69 0x00 0x73 0x00 0x57 0x00 0x61 0x00 0x73 0x00 0x41 0x00 0x4c 0x00 0x6f 0x00 0x6e 0x00 0x67 0x00 0x4e 0x00 0x61 0x00`
- `0x52 0x70 0x03 0x00 0x6d 0x00 0x65` (thisWasALongName)

Notice how the firmware can handle a character split between two packets (like the `m` in this case), it will just read 29 (0x1d) bytes from packet 1 and 3 (0x03) bytes from packet 2 then concatenate.

### Delete
Send command with no payload (len & pkt as 0x00)

`0x52 0x70 0x00 0x00`

### Create
Send a delete command then an update command

- `0x52 0x70 0x00 0x00`
- `0x52 0x70 0x0e 0x00 0x50 0x72 0x30 0x66 0x69 0x6C 0x65` (Pr0file)


## Read Remapped Keys
Remapped keys command is:
`0x52 0x81`

The response we get is `0x54 0x81` then a list of <keycode, hidPage>.

With our buffer being 33 bytes, the command marker being 2 bytes,
we only have 31 bytes to play with (15 keys) and one byte spare.

We can accept multiple packets of keys because the last byte is used as
a keep reading marker (`0x01` for read again and `0x00` for done).

## Key and action list



```yaml
keys:
  - id: esc
    code: 0x29
    label: "ESC"
    default_usage: esc
  - id: f1
    code: 0x3a
    label: "F1"
    default_usage: f1
  - id: f2
    code: 0x3b
    label: "F2"
    default_usage: f2
  - id: f3
    code: 0x3c
    label: "F3"
    default_usage: f3
  - id: f4
    code: 0x3d
    label: "F4"
    default_usage: f4
  - id: f5
    code: 0x3e
    label: "F5"
    default_usage: f5
  - id: f6
    code: 0x3f
    label: "F6"
    default_usage: f6
  - id: f7
    code: 0x40
    label: "F7"
    default_usage: f7
  - id: f8
    code: 0x41
    label: "F8"
    default_usage: f8
  - id: f9
    code: 0x42
    label: "F9"
    default_usage: f9
  - id: f10
    code: 0x43
    label: "F10"
    default_usage: f10
  - id: f11
    code: 0x44
    label: "F11"
    default_usage: f11
  - id: f12
    code: 0x45
    label: "F12"
    default_usage: f12
  - id: prtsc
    code: 0x46
    label: "Prtsc"
    default_usage: prtsc
  - id: scrlk
    code: 0x47
    label: "Scrlk"
    default_usage: scrlk
  - id: pause
    code: 0x48
    label: "Pause"
    default_usage: pause
  - id: grave
    code: 0x35
    label: "`"
    default_usage: grave
  - id: 1
    code: 0x1e
    label: "1"
    default_usage: 1
  - id: 2
    code: 0x1f
    label: "2"
    default_usage: 2
  - id: 3
    code: 0x20
    label: "3"
    default_usage: 3
  - id: 4
    code: 0x21
    label: "4"
    default_usage: 4
  - id: 5
    code: 0x22
    label: "5"
    default_usage: 5
  - id: 6
    code: 0x23
    label: "6"
    default_usage: 6
  - id: 7
    code: 0x24
    label: "7"
    default_usage: 7
  - id: 8
    code: 0x25
    label: "8"
    default_usage: 8
  - id: 9
    code: 0x26
    label: "9"
    default_usage: 9
  - id: 0
    code: 0x27
    label: "0"
    default_usage: 0
  - id: minus
    code: 0x2d
    label: "-"
    default_usage: minus
  - id: equal
    code: 0x2e
    label: "="
    default_usage: equal
  - id: backspace
    code: 0x2a
    label: "Backspace"
    default_usage: backspace
  - id: insert
    code: 0x49
    label: "Insert"
    default_usage: insert
  - id: home
    code: 0x4a
    label: "Home"
    default_usage: home
  - id: pageup
    code: 0x4b
    label: "PgUp"
    default_usage: pageup
  - id: tab
    code: 0x2b
    label: "Tab"
    default_usage: tab
  - id: q
    code: 0x14
    label: "Q"
    default_usage: q
  - id: w
    code: 0x1a
    label: "W"
    default_usage: w
  - id: e
    code: 0x08
    label: "E"
    default_usage: e
  - id: r
    code: 0x15
    label: "R"
    default_usage: r
  - id: t
    code: 0x17
    label: "T"
    default_usage: t
  - id: y
    code: 0x1c
    label: "Y"
    default_usage: y
  - id: u
    code: 0x18
    label: "U"
    default_usage: u
  - id: i
    code: 0x0c
    label: "I"
    default_usage: i
  - id: o
    code: 0x12
    label: "O"
    default_usage: o
  - id: p
    code: 0x13
    label: "P"
    default_usage: p
  - id: leftbrace
    code: 0x2f
    label: "["
    default_usage: leftbrace
  - id: rightbrace
    code: 0x30
    label: "]"
    default_usage: rightbrace
  - id: backslash
    code: 0x31
    label: "\\"
    default_usage: backslash
  - id: delete
    code: 0x4c
    label: "Delete"
    default_usage: delete
  - id: end
    code: 0x4d
    label: "End"
    default_usage: end
  - id: pagedown
    code: 0x4e
    label: "PgDn"
    default_usage: pagedown
  - id: capslock
    code: 0x39
    label: "Caps"
    default_usage: capslock
  - id: a
    code: 0x04
    label: "A"
    default_usage: a
  - id: s
    code: 0x16
    label: "S"
    default_usage: s
  - id: d
    code: 0x07
    label: "D"
    default_usage: d
  - id: f
    code: 0x09
    label: "F"
    default_usage: f
  - id: g
    code: 0x0a
    label: "G"
    default_usage: g
  - id: h
    code: 0x0b
    label: "H"
    default_usage: h
  - id: j
    code: 0x0d
    label: "J"
    default_usage: j
  - id: k
    code: 0x0e
    label: "K"
    default_usage: k
  - id: l
    code: 0x0f
    label: "L"
    default_usage: l
  - id: semicolon
    code: 0x33
    label: ";"
    default_usage: semicolon
  - id: apostrophe
    code: 0x34
    label: "'"
    default_usage: apostrophe
  - id: enter
    code: 0x28
    label: "Enter"
    default_usage: enter
  - id: leftshift
    code: 0x65
    label: "Shift"
    default_usage: leftshift
  - id: z
    code: 0x1d
    label: "Z"
    default_usage: z
  - id: x
    code: 0x1b
    label: "X"
    default_usage: x
  - id: c
    code: 0x06
    label: "C"
    default_usage: c
  - id: v
    code: 0x19
    label: "V"
    default_usage: v
  - id: b
    code: 0x05
    label: "B"
    default_usage: b
  - id: n
    code: 0x11
    label: "N"
    default_usage: n
  - id: m
    code: 0x10
    label: "M"
    default_usage: m
  - id: comma
    code: 0x36
    label: ","
    default_usage: comma
  - id: dot
    code: 0x37
    label: "."
    default_usage: dot
  - id: slash
    code: 0x38
    label: "/"
    default_usage: slash
  - id: rightshift
    code: 0x69
    label: "Shift"
    default_usage: rightshift
  - id: up
    code: 0x52
    label: "↑"
    default_usage: up
  - id: leftctrl
    code: 0x64
    label: "Ctrl"
    default_usage: leftctrl
  - id: windows
    code: 0x67
    label: "Win"
    default_usage: windows
  - id: leftmeta
    code: 0x67
    label: "Win"
    default_usage: leftmeta
  - id: leftalt
    code: 0x66
    label: "Alt"
    default_usage: leftalt
  - id: space
    code: 0x2c
    label: "Space"
    default_usage: space
  - id: rightalt
    code: 0x6a
    label: "Alt"
    default_usage: rightalt
  - id: superb
    code: 0x6c
    label: "B"
    default_usage: superb
  - id: rightmeta
    code: 0x6c
    label: "B"
    default_usage: superb
  - id: supera
    code: 0x6d
    label: "A"
    default_usage: supera
  - id: compose
    code: 0x6d
    label: "Compose"
    default_usage: supera
  - id: menu
    code: 0x6d
    label: "Menu"
    default_usage: supera
  - id: rightctrl
    code: 0x68
    label: "Ctrl"
    default_usage: rightctrl
  - id: left
    code: 0x50
    label: "←"
    default_usage: left
  - id: down
    code: 0x51
    label: "↓"
    default_usage: down
  - id: right
    code: 0x4f
    label: "→"
    default_usage: right
  - id: numlock
    code: 0x53
    label: "Num"
    default_usage: numlock
  - id: kpslash
    code: 0x54
    label: "/"
    default_usage: kpslash
  - id: kpasterisk
    code: 0x55
    label: "*"
    default_usage: kpasterisk
  - id: kpminus
    code: 0x56
    label: "-"
    default_usage: kpminus
  - id: kpplus
    code: 0x57
    label: "+"
    default_usage: kpplus
  - id: kpenter
    code: 0x58
    label: "Enter"
    default_usage: kpenter
  - id: kp1
    code: 0x59
    label: "1"
    default_usage: kp1
  - id: kp2
    code: 0x5a
    label: "2"
    default_usage: kp2
  - id: kp3
    code: 0x5b
    label: "3"
    default_usage: kp3
  - id: kp4
    code: 0x5c
    label: "4"
    default_usage: kp4
  - id: kp5
    code: 0x5d
    label: "5"
    default_usage: kp5
  - id: kp6
    code: 0x5e
    label: "6"
    default_usage: kp6
  - id: kp7
    code: 0x5f
    label: "7"
    default_usage: kp7
  - id: kp8
    code: 0x60
    label: "8"
    default_usage: kp8
  - id: kp9
    code: 0x61
    label: "9"
    default_usage: kp9
  - id: kp0
    code: 0x62
    label: "0"
    default_usage: kp0
  - id: kpdot
    code: 0x63
    label: "."
    default_usage: kpdot
  - id: superb              #  Based on 108-key
    code: 0xe8
    label: "B"
    default_usage: superb
  - id: supera              #  Based on 108-key
    code: 0xe9
    label: "A"
    default_usage: compose

usages:
    # Start of Keyboard/Keypad page
  - id: none
    code: 0x070000
    desc: "(no key pressed)"
  - id: a
    code: 0x070004
    desc: "Keyboard a and A"
  - id: b
    code: 0x070005
    desc: "Keyboard b and B"
  - id: c
    code: 0x070006
    desc: "Keyboard c and C"
  - id: d
    code: 0x070007
    desc: "Keyboard d and D"
  - id: e
    code: 0x070008
    desc: "Keyboard e and E"
  - id: f
    code: 0x070009
    desc: "Keyboard f and F"
  - id: g
    code: 0x07000a
    desc: "Keyboard g and G"
  - id: h
    code: 0x07000b
    desc: "Keyboard h and H"
  - id: i
    code: 0x07000c
    desc: "Keyboard i and I"
  - id: j
    code: 0x07000d
    desc: "Keyboard j and J"
  - id: k
    code: 0x07000e
    desc: "Keyboard k and K"
  - id: l
    code: 0x07000f
    desc: "Keyboard l and L"
  - id: m
    code: 0x070010
    desc: "Keyboard m and M"
  - id: n
    code: 0x070011
    desc: "Keyboard n and N"
  - id: o
    code: 0x070012
    desc: "Keyboard o and O"
  - id: p
    code: 0x070013
    desc: "Keyboard p and P"
  - id: q
    code: 0x070014
    desc: "Keyboard q and Q"
  - id: r
    code: 0x070015
    desc: "Keyboard r and R"
  - id: s
    code: 0x070016
    desc: "Keyboard s and S"
  - id: t
    code: 0x070017
    desc: "Keyboard t and T"
  - id: u
    code: 0x070018
    desc: "Keyboard u and U"
  - id: v
    code: 0x070019
    desc: "Keyboard v and V"
  - id: w
    code: 0x07001a
    desc: "Keyboard w and W"
  - id: x
    code: 0x07001b
    desc: "Keyboard x and X"
  - id: y
    code: 0x07001c
    desc: "Keyboard y and Y"
  - id: z
    code: 0x07001d
    desc: "Keyboard z and Z"
  - id: 1
    code: 0x07001e
    desc: "Keyboard 1 and !"
  - id: 2
    code: 0x07001f
    desc: "Keyboard 2 and @"
  - id: 3
    code: 0x070020
    desc: "Keyboard 3 and #"
  - id: 4
    code: 0x070021
    desc: "Keyboard 4 and $"
  - id: 5
    code: 0x070022
    desc: "Keyboard 5 and %"
  - id: 6
    code: 0x070023
    desc: "Keyboard 6 and ^"
  - id: 7
    code: 0x070024
    desc: "Keyboard 7 and &"
  - id: 8
    code: 0x070025
    desc: "Keyboard 8 and *"
  - id: 9
    code: 0x070026
    desc: "Keyboard 9 and ("
  - id: 0
    code: 0x070027
    desc: "Keyboard 0 and )"
  - id: enter
    code: 0x070028
    desc: "Keyboard Return (ENTER)"
  - id: esc
    code: 0x070029
    desc: "Keyboard ESCAPE"
  - id: backspace
    code: 0x07002a
    desc: "Keyboard DELETE (Backspace)"
  - id: tab
    code: 0x07002b
    desc: "Keyboard Tab"
  - id: space
    code: 0x07002c
    desc: "Keyboard Spacebar"
  - id: minus
    code: 0x07002d
    desc: "Keyboard - and _"
  - id: equal
    code: 0x07002e
    desc: "Keyboard = and +"
  - id: leftbrace
    code: 0x07002f
    desc: "Keyboard [ and {"
  - id: rightbrace
    code: 0x070030
    desc: "Keyboard ] and }"
  - id: backslash
    code: 0x070031
    desc: "Keyboard \\ and |"
  - id: hashtilde
    code: 0x070032
    desc: "Keyboard Non-US # and ~"
  - id: semicolon
    code: 0x070033
    desc: "Keyboard ; and :"
  - id: apostrophe
    code: 0x070034
    desc: 'Keyboard '' and "'
  - id: grave
    code: 0x070035
    desc: "Keyboard ` and ~"
  - id: comma
    code: 0x070036
    desc: "Keyboard , and <"
  - id: dot
    code: 0x070037
    desc: "Keyboard . and >"
  - id: slash
    code: 0x070038
    desc: "Keyboard / and ?"
  - id: capslock
    code: 0x070039
    desc: "Keyboard Caps Lock"
  - id: f1
    code: 0x07003a
    desc: "Keyboard F1"
  - id: f2
    code: 0x07003b
    desc: "Keyboard F2"
  - id: f3
    code: 0x07003c
    desc: "Keyboard F3"
  - id: f4
    code: 0x07003d
    desc: "Keyboard F4"
  - id: f5
    code: 0x07003e
    desc: "Keyboard F5"
  - id: f6
    code: 0x07003f
    desc: "Keyboard F6"
  - id: f7
    code: 0x070040
    desc: "Keyboard F7"
  - id: f8
    code: 0x070041
    desc: "Keyboard F8"
  - id: f9
    code: 0x070042
    desc: "Keyboard F9"
  - id: f10
    code: 0x070043
    desc: "Keyboard F10"
  - id: f11
    code: 0x070044
    desc: "Keyboard F11"
  - id: f12
    code: 0x070045
    desc: "Keyboard F12"
  - id: f13
    code: 0x070068
    desc: "Keyboard F13"
  - id: f14
    code: 0x070069
    desc: "Keyboard F14"
  - id: f15
    code: 0x07006a
    desc: "Keyboard F15"
  - id: f16
    code: 0x07006b
    desc: "Keyboard F16"
  - id: f17
    code: 0x07006c
    desc: "Keyboard F17"
  - id: f18
    code: 0x07006d
    desc: "Keyboard F18"
  - id: f19
    code: 0x07006e
    desc: "Keyboard F19"
  - id: f20
    code: 0x07006f
    desc: "Keyboard F20"
  - id: f21
    code: 0x070070
    desc: "Keyboard F21"
  - id: F22
    code: 0x070071
    desc: "Keyboard F22"
  - id: F23
    code: 0x070072
    desc: "Keyboard F23"
  - id: F24
    code: 0x070073
    desc: "Keyboard F24"
  - id: sysrq
    code: 0x070046
    desc: "Keyboard PrintScreen"
  - id: prtsc
    code: 0x070046
    desc: "Keyboard PrintScreen"
  - id: scrolllock
    code: 0x070047
    desc: "Keyboard Scroll Lock"
  - id: scrlk
    code: 0x070047
    desc: "Keyboard Scroll Lock"
  - id: pause
    code: 0x070048
    desc: "Keyboard Pause"
  - id: insert
    code: 0x070049
    desc: "Keyboard Insert"
  - id: home
    code: 0x07004a
    desc: "Keyboard Home"
  - id: pageup
    code: 0x07004b
    desc: "Keyboard Page Up"
  - id: delete
    code: 0x07004c
    desc: "Keyboard Delete Forward"
  - id: end
    code: 0x07004d
    desc: "Keyboard End"
  - id: pagedown
    code: 0x07004e
    desc: "Keyboard Page Down"
  - id: right
    code: 0x07004f
    desc: "Keyboard Right Arrow"
  - id: left
    code: 0x070050
    desc: "Keyboard Left Arrow"
  - id: down
    code: 0x070051
    desc: "Keyboard Down Arrow"
  - id: up
    code: 0x070052
    desc: "Keyboard Up Arrow"
  - id: numlock
    code: 0x070053
    desc: "Keyboard Num Lock and Clear"
  - id: kpslash
    code: 0x070054
    desc: "Keypad /"
  - id: kpasterisk
    code: 0x070055
    desc: "Keypad *"
  - id: kpminus
    code: 0x070056
    desc: "Keypad -"
  - id: kpplus
    code: 0x070057
    desc: "Keypad +"
  - id: kpenter
    code: 0x070058
    desc: "Keypad ENTER"
  - id: kp1
    code: 0x070059
    desc: "Keypad 1 and End"
  - id: kp2
    code: 0x07005a
    desc: "Keypad 2 and Down Arrow"
  - id: kp3
    code: 0x07005b
    desc: "Keypad 3 and PageDn"
  - id: kp4
    code: 0x07005c
    desc: "Keypad 4 and Left Arrow"
  - id: kp5
    code: 0x07005d
    desc: "Keypad 5"
  - id: kp6
    code: 0x07005e
    desc: "Keypad 6 and Right Arrow"
  - id: kp7
    code: 0x07005f
    desc: "Keypad 7 and Home"
  - id: kp8
    code: 0x070060
    desc: "Keypad 8 and Up Arrow"
  - id: kp9
    code: 0x070061
    desc: "Keypad 9 and Page Up"
  - id: kp0
    code: 0x070062
    desc: "Keypad 0 and Insert"
  - id: kpdot
    code: 0x070063
    desc: "Keypad . and Delete"
  - id: 102nd
    code: 0x070064
    desc: "Keyboard Non-US \\ and |"
  - id: leftctrl
    code: 0x07e000
    desc: "Keyboard Left Control"
  - id: leftshift
    code: 0x07e100
    desc: "Keyboard Left Shift"
  - id: leftalt
    code: 0x07e200
    desc: "Keyboard Left Alt"
  - id: leftmeta
    code: 0x07e300
    desc: "Keyboard Left GUI"
  - id: windows
    code: 0x07e300
    desc: "Keyboard Left GUI"
  - id: rightctrl
    code: 0x07e400
    desc: "Keyboard Right Control"
  - id: rightshift
    code: 0x07e500
    desc: "Keyboard Right Shift"
  - id: rightalt
    code: 0x07e600
    desc: "Keyboard Right Alt"
  - id: rightmeta
    code: 0x07e700
    desc: "Keyboard Right GUI"
  - id: superb
    code: 0x07e800
    desc: "Reserved"
  - id: supera
    code: 0x07e900
    desc: "Reserved"
    # Start of consumer page
  - id: playpause
    code: 0x0ccd00
    desc: "Play/Pause"
  - id: previoussong
    code: 0x0cb500
    desc: "Scan Next Track"
  - id: nextsong
    code: 0x0cb600
    desc: "Scan Previous Track"
  - id: mute
    code: 0x0ce200
    desc: "Mute"
  - id: volumeup
    code: 0x0ce900
    desc: "Volume Increment"
  - id: volumedown
    code: 0x0cea00
    desc: "Volume Decrement"
  - id: calc
    code: 0x0c9201
    desc: "AL Calculator"
  - id: btn_left
    code: 0x010100
    desc: "left click"
    # Not on the HID usages spec
  - id: btn_right
    code: 0x010200
    desc: "right click"
  - id: btn_middle
    code: 0x010400
    desc: "scroll click"
  - id: btn_extra
    code: 0x010800
    desc: "mouse button 4"
  - id: btn_side
    code: 0x011000
    desc: "mouse button 5"

```