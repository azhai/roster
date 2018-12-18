package dataset

import (
	"bytes"
	"io"
	"time"

	"github.com/azhai/roster/buffer"
	"github.com/azhai/roster/utils"
)

const (
	FILE_EXT      = ".dat"
	FIX_BYTES     = 4
	SUB_BYTES     = 2
	ITEM_SIZE_MAX = 256
)

type Offset = uint16
type DropSize = int

const (
	DROP_U32 DropSize = iota
	DROP_U24
	DROP_U16
	DROP_U8
)

type Header struct {
	IdxBegin buffer.Position // 第一个顶层索引开始位置
	IdxEnd   buffer.Position // 最后一个顶层索引结束位置
	KeyCount uint32
	KeySize  uint8 // 后面5bit，1 ~ 31
	ItemSize uint8 // 0（变长）~ 256
	Version  []byte
}

func NewHeader(keySize, itemSize int) *Header {
	return &Header{
		KeySize:  uint8(uint(keySize)),
		ItemSize: uint8(uint(itemSize)),
	}
}

func (h *Header) ToTuple() []interface{} {
	return []interface{}{
		&h.IdxBegin, &h.IdxEnd, &h.KeyCount,
		&h.KeySize, &h.ItemSize, &h.Version,
	}
}

func (h *Header) GetHeaderSize() int {
	fixLen := (FIX_BYTES * 2) + 6 // 前5项字节数为固定值
	return fixLen + len(h.Version)
}

func (h *Header) GetIndexSize() (int, int, int) {
	var cntLen = 0
	keyLen := int(int8(h.KeySize & 0x1f))
	haveSub, drops := h.GetFlags()
	if haveSub {
		cntLen = 1
	}
	return keyLen, cntLen, FIX_BYTES - drops
}

// 第1个bit为haveSub，第2、3个bits为drops
func (h *Header) GetFlags() (bool, DropSize) {
	flags := h.KeySize >> 5
	haveSub := flags >= 0x04
	drops := DropSize(int8(flags & 0x03))
	return haveSub, drops
}

func (h *Header) SetFlags(haveSub bool, drops DropSize) {
	flags := uint8(uint(drops))
	if haveSub {
		flags = flags | 0x04
	}
	// 先清空前3bits，再设置新flags
	h.KeySize = (h.KeySize & 0x1f) | (flags << 5)
}

func (h *Header) SetVersion(version string) error {
	if version == "" {
		version = time.Now().Format("060102")
	}
	data, err := utils.Hex2Bin(version)
	if err == nil {
		h.Version = append(data, 0x00)
	}
	return err
}

type DataSet struct {
	reader  io.ReaderAt
	catalog buffer.ReaderIndex
	Header  *Header
}

func NewDataSet(reader io.ReaderAt, header *Header) *DataSet {
	if header == nil {
		header = NewHeader(31, 0)
	}
	d := &DataSet{reader: reader, Header: header}
	if d.Header.IdxBegin == buffer.Position(0) {
		d.LoadHeader()
	}
	d.CreateCatalog()
	return d
}

func (d *DataSet) Close() error {
	var err error
	if d.catalog != nil {
		if v, ok := d.catalog.(io.Closer); ok {
			err = v.Close()
		}
	}
	if d.reader != nil {
		if v, ok := d.reader.(io.Closer); ok {
			err = v.Close()
		}
	}
	return err
}

func (d *DataSet) ReadAt(data []byte, offset int64) (int, error) {
	return d.reader.ReadAt(data, offset)
}

func (d *DataSet) LoadHeader() error {
	size := d.Header.GetHeaderSize()
	data := make([]byte, size)
	if _, err := d.ReadAt(data, 0); err != nil {
		return err
	}
	buf := buffer.NewBinBuffer(0)
	if _, err := buf.Write(data); err != nil {
		return err
	}
	err := buf.LoadTuple(d.Header)
	buf.Close()
	return err
}

func (d *DataSet) CreateCatalog() error {
	var err error
	begin, end := d.Header.IdxBegin, d.Header.IdxEnd
	d.catalog, err = buffer.NewBufCatalog(d.reader, begin, end)
	if err != nil {
		d.catalog = buffer.NewCatalog(d.reader, begin, end)
	}
	return err
}

func (d *DataSet) SearchIndex(target []byte) ([]byte, []byte) {
	keyLen, cntLen, posLen := d.Header.GetIndexSize()
	unitSize := keyLen + cntLen + posLen
	d.catalog.SetUnitSize(unitSize)
	i := buffer.BinSearch(d.catalog, target, true)
	if i < 0 {
		return nil, nil
	}
	index := make([]byte, unitSize)
	d.catalog.ReadIndex(index, i)
	return index[:keyLen], index[unitSize-posLen:]
}

func (d *DataSet) GetRecord(addr []byte) ([]byte, error) {
	data := make([]byte, ITEM_SIZE_MAX)
	offset := int64(buffer.BomUint64(addr))
	_, err := d.ReadAt(data, offset)
	if err == nil {
		n := bytes.IndexByte(data, 0x00)
		if n >= 0 {
			data = data[:n]
		}
	}
	return data, err
}
