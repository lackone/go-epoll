package go_epoll

import (
	"io"
)

type Buffer struct {
	buf   []byte //数据缓冲
	start int    //数据开始下标
	end   int    //数据结束下标
}

func NewBuffer(bytes []byte) *Buffer {
	return &Buffer{
		buf:   bytes,
		start: 0,
		end:   0,
	}
}

// 数据长度
func (b *Buffer) Len() int {
	return b.end - b.start
}

// 缓冲容量
func (b *Buffer) Cap() int {
	return cap(b.buf)
}

// 获取数据
func (b *Buffer) Bytes() []byte {
	return b.buf[b.start:b.end]
}

// 获取buf
func (b *Buffer) Buf() []byte {
	return b.buf
}

func (b *Buffer) GetStart() int {
	return b.start
}

func (b *Buffer) SetStart(start int) {
	b.start = start
}

func (b *Buffer) GetEnd() int {
	return b.end
}

func (b *Buffer) SetEnd(end int) {
	b.end = end
}

// 尝试获取n个字节数据，如果不够，则返回错误，注意，这里并不会移动start和end下标
func (b *Buffer) TryGet(n int) ([]byte, error) {
	if b.Len() >= n {
		buf := b.buf[b.start : b.start+n]
		return buf, nil
	}
	return nil, DataNotEnough
}

// 读数据
func (b *Buffer) Read(p []byte) (n int, err error) {
	if b.Len() <= 0 {
		b.Reset()
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	n = copy(p, b.buf[b.start:b.end])
	b.start += n
	return
}

// 从start开始，偏移offset位，读取n个字节
func (b *Buffer) ReadAt(offset, n int) ([]byte, error) {
	if b.Len() <= 0 {
		b.Reset()
		return nil, io.EOF
	}
	if b.Len() < (offset + n) {
		return nil, DataNotEnough
	}
	b.start += offset
	buf := b.buf[b.start : b.start+n]
	b.start += n
	return buf, nil
}

// 读取所有数据
func (b *Buffer) ReadAll() ([]byte, error) {
	return b.ReadAt(0, b.Len())
}

// 写数据
func (b *Buffer) Write(p []byte) (n int, err error) {
	if b.tryGrow(len(p)) {
		b.grow()
	}
	n = copy(b.buf[b.end:], p)
	b.end += n
	return
}

// 扩容
func (b *Buffer) grow() {
	var newCap int
	const BK4 = 1 << 12
	// 容量低于4KB时2倍扩容，否则按照当前容量的1/4递增
	if b.Cap() < BK4 {
		newCap = b.Cap() * 2
	} else {
		newCap = b.Cap() + (b.Cap() / 4)
	}
	buf := make([]byte, newCap)
	copy(buf, b.buf)
	b.buf = buf
	return
}

// 判断是否要扩容
func (b *Buffer) tryGrow(n int) bool {
	if (b.end + n) > b.Cap() {
		return true
	}
	return false
}

// 重置
func (b *Buffer) Reset() {
	if b.start == 0 {
		return
	}
	copy(b.buf, b.buf[b.start:b.end])
	b.end -= b.start
	b.start = 0
}
