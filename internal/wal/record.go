package wal

import (
	"encoding/binary"
	"fmt"
	"io"
)

type LSN uint64

type Body interface {
	kind() byte
	writeBody(w io.Writer) (int64, error)
	readBody(r io.Reader) (int64, error)
	isBody()
}

type Entry struct {
	LSN  LSN
	Body Body
}

type Begin struct{ XID uint64 }
type Commit struct{ XID uint64 }
type Abort struct{ XID uint64 }
type Update struct {
	XID           uint64
	PageID        uint64
	Offset        uint64
	Before, After []byte
}

func (*Begin) isBody()  {}
func (*Commit) isBody() {}
func (*Abort) isBody()  {}
func (*Update) isBody() {}

const (
	kindBegin  byte = 1
	kindCommit byte = 2
	kindAbort  byte = 3
	kindUpdate byte = 4
)

func (*Begin) kind() byte  { return kindBegin }
func (*Commit) kind() byte { return kindCommit }
func (*Abort) kind() byte  { return kindAbort }
func (*Update) kind() byte { return kindUpdate }

// Small bodies: [xid:8]
func writeXID(w io.Writer, xid uint64) (int64, error) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], xid)
	n, err := w.Write(buf[:])
	return int64(n), err
}

func readXID(r io.Reader, xid *uint64) (int64, error) {
	var buf [8]byte
	n, err := io.ReadFull(r, buf[:])
	if err != nil {
		return int64(n), err
	}
	*xid = binary.LittleEndian.Uint64(buf[:])
	return int64(n), nil
}

func (b *Begin) writeBody(w io.Writer) (int64, error)  { return writeXID(w, b.XID) }
func (b *Commit) writeBody(w io.Writer) (int64, error) { return writeXID(w, b.XID) }
func (b *Abort) writeBody(w io.Writer) (int64, error)  { return writeXID(w, b.XID) }

func (b *Begin) readBody(r io.Reader) (int64, error)  { return readXID(r, &b.XID) }
func (b *Commit) readBody(r io.Reader) (int64, error) { return readXID(r, &b.XID) }
func (b *Abort) readBody(r io.Reader) (int64, error)  { return readXID(r, &b.XID) }

// Update body:
// [0:8]    XID         u64
// [8:16]   PageID      u64
// [16:24]  Offset      u64
// [24:32]  beforeSize  u64
// [32:..]  before
// [N:N+8]  afterSize   u64
// [N+8:..] after
func (u *Update) writeBody(w io.Writer) (int64, error) {
	var hdr [32]byte
	binary.LittleEndian.PutUint64(hdr[0:8], u.XID)
	binary.LittleEndian.PutUint64(hdr[8:16], u.PageID)
	binary.LittleEndian.PutUint64(hdr[16:24], u.Offset)
	binary.LittleEndian.PutUint64(hdr[24:32], uint64(len(u.Before)))

	var total int64
	n, err := w.Write(hdr[:])
	total += int64(n)
	if err != nil {
		return total, err
	}

	n, err = w.Write(u.Before)
	total += int64(n)
	if err != nil {
		return total, err
	}

	var sz [8]byte
	binary.LittleEndian.PutUint64(sz[:], uint64(len(u.After)))
	n, err = w.Write(sz[:])
	total += int64(n)
	if err != nil {
		return total, err
	}

	n, err = w.Write(u.After)
	total += int64(n)
	return total, err
}

func (u *Update) readBody(r io.Reader) (int64, error) {
	var hdr [32]byte
	n, err := io.ReadFull(r, hdr[:])
	total := int64(n)
	if err != nil {
		return total, err
	}
	u.XID = binary.LittleEndian.Uint64(hdr[0:8])
	u.PageID = binary.LittleEndian.Uint64(hdr[8:16])
	u.Offset = binary.LittleEndian.Uint64(hdr[16:24])
	beforeSize := binary.LittleEndian.Uint64(hdr[24:32])

	u.Before = make([]byte, beforeSize)
	n, err = io.ReadFull(r, u.Before)
	total += int64(n)
	if err != nil {
		return total, err
	}

	var sz [8]byte
	n, err = io.ReadFull(r, sz[:])
	total += int64(n)
	if err != nil {
		return total, err
	}
	afterSize := binary.LittleEndian.Uint64(sz[:])

	u.After = make([]byte, afterSize)
	n, err = io.ReadFull(r, u.After)
	total += int64(n)
	return total, err
}

// On-disk entry layout: [kind:1][lsn:8][body...]
func writeEntry(w io.Writer, lsn LSN, b Body) (int64, error) {
	var hdr [9]byte
	hdr[0] = b.kind()
	binary.LittleEndian.PutUint64(hdr[1:9], uint64(lsn))

	var total int64
	n, err := w.Write(hdr[:])
	total += int64(n)
	if err != nil {
		return total, err
	}
	n2, err := b.writeBody(w)
	total += n2
	return total, err
}

func readEntry(r io.Reader) (Entry, error) {
	var hdr [9]byte
	if n, err := io.ReadFull(r, hdr[:]); err != nil {
		if err == io.EOF && n == 0 {
			return Entry{}, nil
		}

		return Entry{}, err
	}
	kind := hdr[0]
	lsn := LSN(binary.LittleEndian.Uint64(hdr[1:9]))

	var body Body
	switch kind {
	case kindBegin:
		body = &Begin{}
	case kindCommit:
		body = &Commit{}
	case kindAbort:
		body = &Abort{}
	case kindUpdate:
		body = &Update{}
	default:
		return Entry{}, fmt.Errorf("wal: unknown record kind %d", kind)
	}

	if _, err := body.readBody(r); err != nil {
		return Entry{}, err
	}
	return Entry{LSN: lsn, Body: body}, nil
}
