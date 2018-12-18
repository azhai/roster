// 格式说明
//  ---------------------------
// |  Header 头部
//  ---------------------------
// |  Record 记录区
//  ---------------------------
// |  Subset 子索引，可选
//  ---------------------------
// |  Index  顶层索引
//  ---------------------------

package dataset

import (
	"io"

	"github.com/azhai/roster/buffer"
)

type KeyPair struct {
	Key []byte
	Idx int
}

type Builder struct {
	Header  *Header
	Record  *buffer.BinBuffer
	Subset  *buffer.BinBuffer
	Index   *buffer.BinBuffer
	PosList []buffer.Position
	IdxList []buffer.Position
}

func NewBuilder(header *Header) *Builder {
	if header == nil {
		header = NewHeader(31, 0)
		header.SetVersion("") // 当前日期
	}
	return &Builder{
		Index:  buffer.NewBinBuffer(0),
		Record: buffer.NewBinBuffer(0),
		Header: header,
	}
}

func (b *Builder) Build(w io.Writer, rs []string, ks []*KeyPair) error {
	if err := b.BuildRecord(b.Record, rs); err != nil {
		return err
	}
	if err := b.BuildIndex(b.Index, ks); err != nil {
		return err
	}
	b.Header.IdxEnd = b.Header.IdxBegin + buffer.NewPosition(b.Index.Len())
	buf := buffer.NewBinBuffer(0)
	if err := buf.AddTuple(b.Header); err != nil {
		return err
	}
	if _, err := buf.WriteTo(w); err != nil {
		return err
	}
	if _, err := b.Record.WriteTo(w); err != nil {
		return err
	}
	if b.Subset != nil {
		if _, err := b.Subset.WriteTo(w); err != nil {
			return err
		}
	}
	_, err := b.Index.WriteTo(w)
	return err
}

func (b *Builder) BuildRecord(buf *buffer.BinBuffer, records []string) error {
	var pos buffer.Position
	base := b.Header.GetHeaderSize()
	for _, rec := range records {
		pos = buffer.NewPosition(base + b.Record.Len())
		if err := buf.AddString(rec); err != nil {
			return err
		}
		b.PosList = append(b.PosList, pos)
	}
	b.Header.IdxBegin = buffer.NewPosition(base + b.Record.Len())
	return nil
}

func (b *Builder) BuildIndex(buf *buffer.BinBuffer, keypairs []*KeyPair) error {
	var (
		count = len(b.PosList)
		addr  buffer.Position
	)
	_, drops := b.Header.GetFlags()
	for _, pair := range keypairs {
		err := buf.AddData(pair.Key)
		if err != nil || pair.Idx < 0 || pair.Idx >= count {
			addr = buffer.NewPosition(0)
		} else {
			addr = b.PosList[pair.Idx]
		}
		buf.AddUint32(addr, drops)
	}
	if size := len(keypairs); size > 0 {
		b.Header.KeyCount += uint32(size)
	}
	return nil
}
