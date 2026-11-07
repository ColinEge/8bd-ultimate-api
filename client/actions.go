package client

// Named actions use the built-in bindables.yaml encodings. Functions return
// independent values, not mutable global arrays. Actions are not physical IDs:
// for example, F13 is a binding target, not a physical key named KeyF13.
func ActionA() Action { return Action{7, 0, 0x04} }
func ActionB() Action { return Action{7, 0, 0x05} }
func ActionC() Action { return Action{7, 0, 0x06} }
func ActionD() Action { return Action{7, 0, 0x07} }
func ActionE() Action { return Action{7, 0, 0x08} }
func ActionF() Action { return Action{7, 0, 0x09} }
func ActionG() Action { return Action{7, 0, 0x0a} }
func ActionH() Action { return Action{7, 0, 0x0b} }
func ActionI() Action { return Action{7, 0, 0x0c} }
func ActionJ() Action { return Action{7, 0, 0x0d} }
func ActionK() Action { return Action{7, 0, 0x0e} }
func ActionL() Action { return Action{7, 0, 0x0f} }
func ActionM() Action { return Action{7, 0, 0x10} }
func ActionN() Action { return Action{7, 0, 0x11} }
func ActionO() Action { return Action{7, 0, 0x12} }
func ActionP() Action { return Action{7, 0, 0x13} }
func ActionQ() Action { return Action{7, 0, 0x14} }
func ActionR() Action { return Action{7, 0, 0x15} }
func ActionS() Action { return Action{7, 0, 0x16} }
func ActionT() Action { return Action{7, 0, 0x17} }
func ActionU() Action { return Action{7, 0, 0x18} }
func ActionV() Action { return Action{7, 0, 0x19} }
func ActionW() Action { return Action{7, 0, 0x1a} }
func ActionX() Action { return Action{7, 0, 0x1b} }
func ActionY() Action { return Action{7, 0, 0x1c} }
func ActionZ() Action { return Action{7, 0, 0x1d} }

func Action1() Action { return Action{7, 0, 0x1e} }
func Action2() Action { return Action{7, 0, 0x1f} }
func Action3() Action { return Action{7, 0, 0x20} }
func Action4() Action { return Action{7, 0, 0x21} }
func Action5() Action { return Action{7, 0, 0x22} }
func Action6() Action { return Action{7, 0, 0x23} }
func Action7() Action { return Action{7, 0, 0x24} }
func Action8() Action { return Action{7, 0, 0x25} }
func Action9() Action { return Action{7, 0, 0x26} }
func Action0() Action { return Action{7, 0, 0x27} }

func ActionF1() Action  { return Action{7, 0, 0x3a} }
func ActionF2() Action  { return Action{7, 0, 0x3b} }
func ActionF3() Action  { return Action{7, 0, 0x3c} }
func ActionF4() Action  { return Action{7, 0, 0x3d} }
func ActionF5() Action  { return Action{7, 0, 0x3e} }
func ActionF6() Action  { return Action{7, 0, 0x3f} }
func ActionF7() Action  { return Action{7, 0, 0x40} }
func ActionF8() Action  { return Action{7, 0, 0x41} }
func ActionF9() Action  { return Action{7, 0, 0x42} }
func ActionF10() Action { return Action{7, 0, 0x43} }
func ActionF11() Action { return Action{7, 0, 0x44} }
func ActionF12() Action { return Action{7, 0, 0x45} }
func ActionF13() Action { return Action{7, 0, 0x68} }
func ActionF14() Action { return Action{7, 0, 0x69} }
func ActionF15() Action { return Action{7, 0, 0x6a} }
func ActionF16() Action { return Action{7, 0, 0x6b} }
func ActionF17() Action { return Action{7, 0, 0x6c} }
func ActionF18() Action { return Action{7, 0, 0x6d} }
func ActionF19() Action { return Action{7, 0, 0x6e} }
func ActionF20() Action { return Action{7, 0, 0x6f} }
func ActionF21() Action { return Action{7, 0, 0x70} }
func ActionF22() Action { return Action{7, 0, 0x71} }
func ActionF23() Action { return Action{7, 0, 0x72} }
func ActionF24() Action { return Action{7, 0, 0x73} }

func ActionEnter() Action       { return Action{7, 0, 0x28} }
func ActionEsc() Action         { return Action{7, 0, 0x29} }
func ActionBackspace() Action   { return Action{7, 0, 0x2a} }
func ActionTab() Action         { return Action{7, 0, 0x2b} }
func ActionSpace() Action       { return Action{7, 0, 0x2c} }
func ActionCapsLock() Action    { return Action{7, 0, 0x39} }
func ActionPrintScreen() Action { return Action{7, 0, 0x46} }
func ActionScrollLock() Action  { return Action{7, 0, 0x47} }
func ActionPause() Action       { return Action{7, 0, 0x48} }
func ActionInsert() Action      { return Action{7, 0, 0x49} }
func ActionHome() Action        { return Action{7, 0, 0x4a} }
func ActionPageUp() Action      { return Action{7, 0, 0x4b} }
func ActionDelete() Action      { return Action{7, 0, 0x4c} }
func ActionEnd() Action         { return Action{7, 0, 0x4d} }
func ActionPageDown() Action    { return Action{7, 0, 0x4e} }
func ActionRight() Action       { return Action{7, 0, 0x4f} }
func ActionLeft() Action        { return Action{7, 0, 0x50} }
func ActionDown() Action        { return Action{7, 0, 0x51} }
func ActionUp() Action          { return Action{7, 0, 0x52} }

func ActionNumLock() Action        { return Action{7, 0, 0x53} }
func ActionNumpadDivide() Action   { return Action{7, 0, 0x54} }
func ActionNumpadMultiply() Action { return Action{7, 0, 0x55} }
func ActionNumpadMinus() Action    { return Action{7, 0, 0x56} }
func ActionNumpadPlus() Action     { return Action{7, 0, 0x57} }
func ActionNumpadEnter() Action    { return Action{7, 0, 0x58} }
func ActionNumpad1() Action        { return Action{7, 0, 0x59} }
func ActionNumpad2() Action        { return Action{7, 0, 0x5a} }
func ActionNumpad3() Action        { return Action{7, 0, 0x5b} }
func ActionNumpad4() Action        { return Action{7, 0, 0x5c} }
func ActionNumpad5() Action        { return Action{7, 0, 0x5d} }
func ActionNumpad6() Action        { return Action{7, 0, 0x5e} }
func ActionNumpad7() Action        { return Action{7, 0, 0x5f} }
func ActionNumpad8() Action        { return Action{7, 0, 0x60} }
func ActionNumpad9() Action        { return Action{7, 0, 0x61} }
func ActionNumpad0() Action        { return Action{7, 0, 0x62} }
func ActionNumpadDot() Action      { return Action{7, 0, 0x63} }

func ActionLeftCtrl() Action   { return Action{7, 0xe0, 0} }
func ActionLeftShift() Action  { return Action{7, 0xe1, 0} }
func ActionLeftAlt() Action    { return Action{7, 0xe2, 0} }
func ActionLeftMeta() Action   { return Action{7, 0xe3, 0} }
func ActionRightCtrl() Action  { return Action{7, 0xe4, 0} }
func ActionRightShift() Action { return Action{7, 0xe5, 0} }
func ActionRightAlt() Action   { return Action{7, 0xe6, 0} }
func ActionRightMeta() Action  { return Action{7, 0xe7, 0} }
func ActionFn() Action         { return Action{7, 0, 3} }
func ActionMenu() Action       { return Action{7, 0xe1, 0x43} }

func ActionGrave() Action        { return Action{7, 0, 0x35} }
func ActionTilde() Action        { return Action{7, 0xe1, 0x35} }
func ActionExclamation() Action  { return Action{7, 0xe1, 0x1e} }
func ActionAt() Action           { return Action{7, 0xe1, 0x1f} }
func ActionHash() Action         { return Action{7, 0xe1, 0x20} }
func ActionDollar() Action       { return Action{7, 0xe1, 0x21} }
func ActionPercent() Action      { return Action{7, 0xe1, 0x22} }
func ActionCaret() Action        { return Action{7, 0xe1, 0x23} }
func ActionAmpersand() Action    { return Action{7, 0xe1, 0x24} }
func ActionAsterisk() Action     { return Action{7, 0xe1, 0x25} }
func ActionLeftParen() Action    { return Action{7, 0xe1, 0x26} }
func ActionRightParen() Action   { return Action{7, 0xe1, 0x27} }
func ActionMinus() Action        { return Action{7, 0, 0x2d} }
func ActionEqual() Action        { return Action{7, 0, 0x2e} }
func ActionUnderscore() Action   { return Action{7, 0xe1, 0x2d} }
func ActionPlus() Action         { return Action{7, 0xe1, 0x2e} }
func ActionLeftBracket() Action  { return Action{7, 0, 0x2f} }
func ActionRightBracket() Action { return Action{7, 0, 0x30} }
func ActionBackslash() Action    { return Action{7, 0, 0x31} }
func ActionLeftBrace() Action    { return Action{7, 0xe1, 0x2f} }
func ActionRightBrace() Action   { return Action{7, 0xe1, 0x30} }
func ActionPipe() Action         { return Action{7, 0xe1, 0x31} }
func ActionSemicolon() Action    { return Action{7, 0, 0x33} }
func ActionApostrophe() Action   { return Action{7, 0, 0x34} }
func ActionColon() Action        { return Action{7, 0xe1, 0x33} }
func ActionDoubleQuote() Action  { return Action{7, 0xe1, 0x34} }
func ActionComma() Action        { return Action{7, 0, 0x36} }
func ActionDot() Action          { return Action{7, 0, 0x37} }
func ActionSlash() Action        { return Action{7, 0, 0x38} }
func ActionLessThan() Action     { return Action{7, 0xe1, 0x36} }
func ActionGreaterThan() Action  { return Action{7, 0xe1, 0x37} }
func ActionQuestion() Action     { return Action{7, 0xe1, 0x38} }

func ActionPrimaryClick() Action { return Action{1, 1, 0} }
func ActionRightClick() Action   { return Action{1, 2, 0} }
func ActionScrollClick() Action  { return Action{1, 4, 0} }
func ActionMouseForward() Action { return Action{1, 0x10, 0} }
func ActionMouseBack() Action    { return Action{1, 8, 0} }
func ActionDoubleClick() Action  { return Action{1, 0x80, 0} }
func ActionScrollUp() Action     { return Action{1, 0, 0, 0, 2} }
func ActionScrollDown() Action   { return Action{1, 0, 0, 0, 0xfe} }
func ActionScrollLeft() Action   { return Action{1, 0, 0, 0, 0, 0xfe} }
func ActionScrollRight() Action  { return Action{1, 0, 0, 0, 0, 2} }

func ActionVolumeUp() Action              { return Action{0x0c, 0xe9, 0} }
func ActionVolumeDown() Action            { return Action{0x0c, 0xea, 0} }
func ActionMute() Action                  { return Action{0x0c, 0xe2, 0} }
func ActionPlayPause() Action             { return Action{0x0c, 0xcd, 0} }
func ActionPreviousTrack() Action         { return Action{0x0c, 0xb6, 0} }
func ActionNextTrack() Action             { return Action{0x0c, 0xb5, 0} }
func ActionMail() Action                  { return Action{0x0c, 0x8a, 1} }
func ActionFileExplorer() Action          { return Action{7, 0xe3, 8} }
func ActionCalculator() Action            { return Action{0x0c, 0x92, 1} }
func ActionBrowser() Action               { return Action{0x0c, 0x23, 2} }
func ActionDisplayBrightnessDown() Action { return Action{0x0c, 0x70, 0} }
func ActionDisplayBrightnessUp() Action   { return Action{0x0c, 0x6f, 0} }
func ActionDisabled() Action              { return Action{7, 0, 0} }
