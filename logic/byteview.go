package logic

type ByteView struct {
	b []byte
}

func (v ByteView) Len() int {
	return len(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}

func (v ByteView) Bytes() []byte {
	return cloneBytes(v.b)
}

func cloneBytes(b []byte) []byte {
	bytes := make([]byte, len(b))
	copy(bytes, b)
	return bytes
}
