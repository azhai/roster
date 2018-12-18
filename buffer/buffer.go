package buffer

import (
	"bytes"
	"encoding/binary"
	"io"
	"sync"
)

func BomUint64(data []byte) uint64 {
	size := len(data)
	if size > 8 {
		data = data[:8]
	} else if size < 8 {
		zero := bytes.Repeat([]byte{0x00}, 8-size)
		data = append(zero, data...)
	}
	return binary.BigEndian.Uint64(data)
}

type ITuple interface {
	ToTuple() []interface{}
}

type WalkFunc func(item interface{}) error

func IterTuple(t ITuple, f WalkFunc) error {
	for _, item := range t.ToTuple() {
		if err := f(item); err != nil {
			return err
		}
	}
	return nil
}

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(BinBuffer)
	},
}

type BinBuffer struct {
	bytes.Buffer
}

func NewBinBuffer(size int) *BinBuffer {
	b := bufPool.Get().(*BinBuffer)
	b.Reset()
	if size > 0 {
		b.Grow(size)
	}
	return b
}

func (b *BinBuffer) Close() error {
	bufPool.Put(b)
	return nil
}

func (b *BinBuffer) WriteTo(w io.Writer) (int, error) {
	defer b.Close()
	return w.Write(b.Bytes())
}

func (b *BinBuffer) WriteAt(w io.WriteSeeker, offset int64) (int, error) {
	w.Seek(offset, io.SeekStart)
	return b.WriteTo(w)
}

// data需要传引用
func (b *BinBuffer) LoadData(data interface{}) error {
	return binary.Read(b, binary.BigEndian, data)
}

func (b *BinBuffer) LoadTuple(t ITuple) error {
	return IterTuple(t, b.LoadData)
}

func (b *BinBuffer) AddData(data interface{}) error {
	return binary.Write(b, binary.BigEndian, data)
}

func (b *BinBuffer) AddTuple(t ITuple) error {
	return IterTuple(t, b.AddData)
}

func (b *BinBuffer) AddString(str string) error {
	data := append([]byte(str), 0x00)
	return b.AddData(data)
}

func (b *BinBuffer) AddUint32(num uint32, drops int) error {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, num)
	return b.AddData(data[drops:])
}
