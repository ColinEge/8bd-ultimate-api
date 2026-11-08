package client

// Key constants are firmware physical IDs from keys.yaml, not HID usages.
// They do not guarantee that a particular keyboard model has the key.
const (
	KeyA KeyID = 0x04 + iota
	KeyB
	KeyC
	KeyD
	KeyE
	KeyF
	KeyG
	KeyH
	KeyI
	KeyJ
	KeyK
	KeyL
	KeyM
	KeyN
	KeyO
	KeyP
	KeyQ
	KeyR
	KeyS
	KeyT
	KeyU
	KeyV
	KeyW
	KeyX
	KeyY
	KeyZ
	Key1
	Key2
	Key3
	Key4
	Key5
	Key6
	Key7
	Key8
	Key9
	Key0
	KeyEnter
	KeyEsc
	KeyBackspace
	KeyTab
	KeySpace
	KeyMinus
	KeyEqual
	KeyLeftBracket
	KeyRightBracket
	KeyBackslash
)

const (
	KeySemicolon KeyID = 0x33 + iota
	KeyApostrophe
	KeyGrave
	KeyComma
	KeyDot
	KeySlash
	KeyCapsLock
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyPrintScreen
	KeyScrollLock
	KeyPause
	KeyInsert
	KeyHome
	KeyPageUp
	KeyDelete
	KeyEnd
	KeyPageDown
	KeyRight
	KeyLeft
	KeyDown
	KeyUp
	KeyNumLock
	KeyNumpadDivide
	KeyNumpadMultiply
	KeyNumpadMinus
	KeyNumpadPlus
	KeyNumpadEnter
	KeyNumpad1
	KeyNumpad2
	KeyNumpad3
	KeyNumpad4
	KeyNumpad5
	KeyNumpad6
	KeyNumpad7
	KeyNumpad8
	KeyNumpad9
	KeyNumpad0
	KeyNumpadDot
	KeyLeftCtrl
	KeyLeftShift
	KeyLeftAlt
	KeyLeftMeta
	KeyRightCtrl
	KeyRightShift
	KeyRightAlt
)

const (
	KeySuperB KeyID = 0xe8
	KeySuperA KeyID = 0xe9
)
