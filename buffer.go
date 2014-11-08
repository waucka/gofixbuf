// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Modifications made by Alexander Wauck

package gofixbuf

// Simple byte buffer for marshaling data.

import (
	"errors"
	"io"
	"unicode/utf8"
)

// A Buffer is a variable-sized buffer of bytes with Write methods.
// The zero value for Buffer is an empty buffer ready to use.
type Buffer struct {
	buf       []byte            // contents are the bytes buf
	off       int               // write at &buf[off]
	runeBytes [utf8.UTFMax]byte // avoid allocation of slice on each WriteByte or Rune
}

// ErrTooLarge is passed to panic if memory cannot be allocated to store data in a buffer.
var ErrTooLarge = errors.New("gofixbuf.Buffer: too large")

// Bytes returns a slice of the contents of the buffer;
// len(b.Bytes()) == b.Len().  If the caller changes the contents of the
// returned slice, the contents of the buffer will change provided there
// are no intervening method calls on the Buffer.
func (b *Buffer) Bytes() []byte { return b.buf[:] }

// String returns the contents of the buffer
// as a string.  If the Buffer is a nil pointer, it returns "<nil>".
func (b *Buffer) String() string {
	if b == nil {
		// Special case, useful in debugging.
		return "<nil>"
	}
	return string(b.buf)
}

// Len returns the number of bytes in the buffer;
// b.Len() == len(b.Bytes()).
func (b *Buffer) Len() int { return len(b.buf) }

func (b *Buffer) Cap() int {
	return cap(b.buf)
}

// Reset resets the buffer so it represents the full slice.
func (b *Buffer) Reset() {
	b.off = 0
}

// checkLen checks that the buffer has space for n more bytes.
// It returns the index where bytes should be written.
// If the buffer doesn't have room, it will return ErrTooLarge.
func (b *Buffer) checkLen(n int) (int, error) {
	if b.off + n > cap(b.buf) {
		return b.off, ErrTooLarge
	}
	return b.off, nil
}

// Write appends the contents of p to the buffer, growing the buffer as
// needed. The return value n is the length of p; err is always nil. If the
// buffer becomes too large, Write will panic with ErrTooLarge.
func (b *Buffer) Write(p []byte) (n int, err error) {
	m, e := b.checkLen(len(p))
	if e != nil {
		return 0, e
	}
	b.off = m + len(p)
	return copy(b.buf[m:], p), nil
}

// WriteString appends the contents of s to the buffer, growing the buffer as
// needed. The return value n is the length of s; err is always nil. If the
// buffer becomes too large, WriteString will panic with ErrTooLarge.
func (b *Buffer) WriteString(s string) (n int, err error) {
	m, e := b.checkLen(len(s))
	if e != nil {
		return 0, e
	}
	b.off = m + len(s)
	return copy(b.buf[m:], s), nil
}

// ReadFrom reads data from r until EOF and appends it to the buffer, growing
// the buffer as needed. The return value n is the number of bytes read. Any
// error except io.EOF encountered during the read is also returned. If the
// buffer becomes too large, ReadFrom will panic with ErrTooLarge.
func (b *Buffer) ReadFrom(r io.Reader) (n int64, err error) {
	for {
		m, e := r.Read(b.buf[b.off:cap(b.buf)])
		b.off += m
		n += int64(m)
		if e != nil || m == 0 {
			return n, e
		}
	}
}

// WriteByte appends the byte c to the buffer, growing the buffer as needed.
// The returned error is always nil, but is included to match bufio.Writer's
// WriteByte. If the buffer becomes too large, WriteByte will panic with
// ErrTooLarge.
func (b *Buffer) WriteByte(c byte) error {
	m, e := b.checkLen(1)
	if e != nil {
		return e
	}
	b.buf[m] = c
	b.off++
	return nil
}

// WriteRune appends the UTF-8 encoding of Unicode code point r to the
// buffer, returning its length and an error, which is always nil but is
// included to match bufio.Writer's WriteRune. The buffer is grown as needed;
// if it becomes too large, WriteRune will panic with ErrTooLarge.
func (b *Buffer) WriteRune(r rune) (n int, err error) {
	if r < utf8.RuneSelf {
		b.WriteByte(byte(r))
		return 1, nil
	}
	n = utf8.EncodeRune(b.runeBytes[0:], r)
	b.Write(b.runeBytes[0:n])
	return n, nil
}

// NewBuffer creates and initializes a new Buffer using buf as its initial
// contents.  It is intended to size the internal buffer for writing. To do
// that, buf should have the desired capacity but a length of zero.
//
// In most cases, new(Buffer) (or just declaring a Buffer variable) is
// sufficient to initialize a Buffer.
func NewBuffer(buf []byte) *Buffer { return &Buffer{buf: buf} }
