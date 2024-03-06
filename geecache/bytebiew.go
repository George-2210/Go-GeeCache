package geecache

// ByteView 表示一个不可变的只读数据。
type ByteView struct {
	b []byte
}

// Len 返回数据的长度。
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice 将数据作为字节切片返回一个副本。
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// String 将数据作为字符串返回，如果需要会进行复制。
func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
