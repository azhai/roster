package buffer

import (
	"bytes"
	"io"
	"sort"
	// "syscall"
)

type Position = uint32

func NewPosition(n int) Position {
	return Position(uint(n))
}

func FromPosition(pos Position) int64 {
	return int64(uint64(pos))
}

type ReaderIndex interface {
	SetUnitSize(size int)
	Count() int
	ReadIndex(data []byte, idx int) (int, error)
}

func BinSearch(r ReaderIndex, target []byte, checkTop bool) int {
	key := make([]byte, len(target))
	count := r.Count()
	find := func(i int) bool {
		r.ReadIndex(key, i)
		return bytes.Compare(target, key) < 0
	}
	if find(0) {
		return -1 // 超出下限
	} else if checkTop && !find(-1) {
		return -2 // 超出上限
	}
	return sort.Search(count, find) - 1
}

type Catalog struct {
	reader     io.ReaderAt
	begin, end int64
	count      int
	unitSize   int
}

func NewCatalog(reader io.ReaderAt, begin, end Position) *Catalog {
	return &Catalog{
		reader: reader,
		begin:  FromPosition(begin),
		end:    FromPosition(end),
	}
}

func (c *Catalog) SetUnitSize(size int) {
	c.unitSize = size
	if c.unitSize > 0 && c.end > c.begin {
		c.count = int(c.end-c.begin) / c.unitSize
	}
}

func (c *Catalog) Count() int {
	return c.count
}

func (c *Catalog) ReadAt(data []byte, offset int64) (int, error) {
	if offset >= 0 {
		offset += c.begin
	} else {
		offset += c.end
	}
	return c.reader.ReadAt(data, offset)
}

func (c *Catalog) ReadIndex(data []byte, idx int) (int, error) {
	offset := int64(c.unitSize * idx)
	return c.ReadAt(data, offset)
}

type BufCatalog struct {
	data     []byte
	length   int
	count    int
	unitSize int
}

func NewBufCatalog(reader io.ReaderAt, begin, end Position) (*BufCatalog, error) {
	offset := FromPosition(begin)
	size := int(FromPosition(end) - offset)
	c := &BufCatalog{data: make([]byte, size)}
	n, err := reader.ReadAt(c.data, offset)
	if err == nil {
		c.length = n
	}
	return c, err
}

func (c *BufCatalog) SetUnitSize(size int) {
	c.unitSize = size
	if c.unitSize > 0 {
		c.count = c.length / c.unitSize
	}
}

func (c *BufCatalog) Count() int {
	return c.count
}

func (c *BufCatalog) ReadAt(data []byte, offset int64) (int, error) {
	start := int(offset)
	stop := start + len(data)
	return copy(data[:], c.data[start:stop]), nil
}

func (c *BufCatalog) ReadIndex(data []byte, idx int) (int, error) {
	if idx < 0 {
		idx += c.count
	}
	offset := int64(c.unitSize * idx)
	return c.ReadAt(data, offset)
}

/*
type MemCatalog struct {
	BufCatalog
}

func NewMemCatalog(fd int, begin, end Position) (*MemCatalog, error) {
	var c *MemCatalog
	offset := FromPosition(begin)
	size := int(FromPosition(end) - offset)
	data, err := syscall.Mmap(fd, offset, size,
		syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err == nil {
		c = &MemCatalog{
			BufCatalog{data: data, length: len(data)},
		}
	}
	return c, err
}

func (c *MemCatalog) Close() error {
	if c.data != nil {
		return syscall.Munmap(c.data)
	}
	return nil
}
*/
