package protocol

func Header(dir Direction, group command, params ...byte) []byte {
	h := []byte{byte(dir), byte(group)}
	h = append(h, params...)
	return h
}
