// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Modifications made by Alexander Wauck

package gofixbuf_test

import (
	. "github.com/waucka/gofixbuf"
	"io"
	"bytes"
	"testing"
	"unicode/utf8"
)

const N = 10000      // make this bigger for a larger (and slower) test
var data string      // test data for write tests
var testBytes []byte // test data; same as data but as a slice.

func init() {
	testBytes = make([]byte, N)
	for i := 0; i < N; i++ {
		testBytes[i] = 'a' + byte(i%26)
	}
	data = string(testBytes)
}

// Verify that contents of buf match the string s.
func check(t *testing.T, testname string, buf *Buffer, s string) {
	bytes := buf.Bytes()
	str := buf.String()
	if buf.Len() != len(bytes) {
		t.Errorf("%s: buf.Len() == %d, len(buf.Bytes()) == %d", testname, buf.Len(), len(bytes))
	}

	if buf.Len() != len(str) {
		t.Errorf("%s: buf.Len() == %d, len(buf.String()) == %d", testname, buf.Len(), len(str))
	}

	if buf.Len() != len(s) {
		t.Errorf("%s: buf.Len() == %d, len(s) == %d", testname, buf.Len(), len(s))
	}

	if string(bytes) != s {
		t.Errorf("%s: string(buf.Bytes()) == %q, s == %q", testname, string(bytes), s)
	}
}

func TestNewBuffer(t *testing.T) {
	buf := NewBuffer(testBytes)
	check(t, "NewBuffer", buf, data)
}

func TestBasicOperations(t *testing.T) {
	bslice := bytes.Repeat([]byte{0x41}, 64)
	a64 := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	buf := NewBuffer(bslice)

	check(t, "TestBasicOperations (1)", buf, a64)

	buf.Reset()
	check(t, "TestBasicOperations (2)", buf, a64)

	n, err := buf.Write([]byte(data[0:1]))
	if n != 1 {
		t.Errorf("wrote 1 byte, but n == %d", n)
	}
	if err != nil {
		t.Errorf("err should always be nil, but err == %s", err)
	}
	check(t, "TestBasicOperations (4)", buf, "a" + a64[1:])

	buf.WriteByte(data[1])
	check(t, "TestBasicOperations (5)", buf, "ab" + a64[2:])

	n, err = buf.Write([]byte(data[2:26]))
	if n != 24 {
		t.Errorf("wrote 25 bytes, but n == %d", n)
	}
	check(t, "TestBasicOperations (6)", buf, string(data[0:26] + a64[26:]))

	buf.WriteByte(data[1])
}

func TestWriteLimit(t *testing.T) {
	buf := NewBuffer(make([]byte, 10))

	written, err := buf.Write([]byte("AAAAA"))
	if err != nil {
		t.Errorf("Error while writing 5 bytes: %s", err.Error())
	}
	if written != 5 {
		t.Errorf("Should have written 5 bytes; wrote %d", written)
	}
	written, err = buf.Write([]byte("AAAAA"))
	if err != nil {
		t.Errorf("Error while writing 5 bytes: %s", err.Error())
	}
	if written != 5 {
		t.Errorf("Should have written 5 bytes; wrote %d", written)
	}
	written, err = buf.Write([]byte("A"))
	if err == nil {
		t.Errorf("Wrote 11th byte to 10-byte buffer!")
	}
	if written != 0 {
		t.Errorf("Should have written 0 bytes; wrote %d", written)
	}
	check(t, "TestWriteLimit (1)", buf, "AAAAAAAAAA")
}

func TestWriteByteLimit(t *testing.T) {
	buf := NewBuffer(make([]byte, 10))

	for i := 0; i < 10; i++ {
		err := buf.WriteByte(0x41)
		if err != nil {
			t.Errorf("Error while writing byte: %s", err.Error())
		}
	}
	err := buf.WriteByte(0x41)
	if err == nil {
		t.Errorf("Wrote 11th byte to 10-byte buffer!")
	}
	check(t, "TestWriteByteLimit (1)", buf, "AAAAAAAAAA")
}

func TestWriteStringLimit(t *testing.T) {
	buf := NewBuffer(make([]byte, 10))

	written, err := buf.WriteString("AAAAA")
	if err != nil {
		t.Errorf("Error while writing 5-byte string: %s", err.Error())
	}
	if written != 5 {
		t.Errorf("Should have written 5 bytes; wrote %d", written)
	}
	written, err = buf.WriteString("AAAAA")
	if err != nil {
		t.Errorf("Error while writing 5-byte string: %s", err.Error())
	}
	if written != 5 {
		t.Errorf("Should have written 5 bytes; wrote %d", written)
	}
	written, err = buf.WriteString("A")
	if err == nil {
		t.Errorf("Wrote 11th byte to 10-byte buffer!")
	}
	if written != 0 {
		t.Errorf("Should have written 0 bytes; wrote %d", written)
	}
	check(t, "TestWriteByteLimit (1)", buf, "AAAAAAAAAA")
}

func TestCap(t *testing.T) {
	var dbuf Buffer
	if dbuf.Cap() != 0 {
		t.Errorf("Expected capacity of 0; actual capacity is %d", dbuf.Cap())
	}

	buf := NewBuffer(make([]byte, 5))
	if buf.Cap() != 5 {
		t.Errorf("Expected capacity of 5; actual capacity is %d", buf.Cap())
	}

	buf = NewBuffer(make([]byte, 0, 5))
	if buf.Cap() != 5 {
		t.Errorf("Expected capacity of 5; actual capacity is %d", buf.Cap())
	}
}

func TestNil(t *testing.T) {
	var b *Buffer
	if b.String() != "<nil>" {
		t.Errorf("expected <nil>; got %q", b.String())
	}
}

func TestReadFrom(t *testing.T) {
	srcBuf := bytes.NewBuffer(testBytes)
	dstBuf := NewBuffer(make([]byte, 64))
	written, err := dstBuf.ReadFrom(srcBuf)
	if err != nil {
		t.Errorf("Error while reading from buffer: %s", err.Error())
	}
	if written != 64 {
		t.Errorf("Should have written 64 bytes; wrote %d", written)
	}
	check(t, "TestReadFrom (1)", dstBuf, string(testBytes[0:64]))

	written, err = dstBuf.ReadFrom(srcBuf)
	if err != nil {
		t.Errorf("Error while reading from buffer: %s", err.Error())
	}
	if written != 0 {
		t.Errorf("Should have written 0 bytes; wrote %d", written)
	}
	check(t, "TestReadFrom (2)", dstBuf, string(testBytes[0:64]))
}

func TestReadFromEOF(t *testing.T) {
	srcBuf := bytes.NewBuffer(testBytes[0:20])
	dstBuf := NewBuffer(bytes.Repeat([]byte{0x41}, 64))
	written, err := dstBuf.ReadFrom(srcBuf)
	if err != io.EOF {
		if err != nil {
			t.Errorf("Error while reading from buffer: %s", err.Error())
		} else {
			t.Error("Expected io.EOF while reading from buffer")
		}
	}
	if written != 20 {
		t.Errorf("Should have written 20 bytes; wrote %d", written)
	}
	expected := string(testBytes[0:20]) + string(bytes.Repeat([]byte{0x41}, 44))
	check(t, "TestReadFromEOF (1)", dstBuf, expected)
}

func TestWriteTo(t *testing.T) {
	srcBuf := bytes.NewBuffer(testBytes[0:64])
	dstBuf := NewBuffer(make([]byte, 64))
	written, err := srcBuf.WriteTo(dstBuf)
	if err != nil {
		t.Errorf("Error while reading from buffer: %s", err.Error())
	}
	if written != 64 {
		t.Errorf("Should have written 64 bytes; wrote %d", written)
	}
	check(t, "TestWriteTo (1)", dstBuf, string(testBytes[0:64]))
}

func TestLargeWriteTo(t *testing.T) {
	srcBuf := bytes.NewBuffer(testBytes)
	dstBuf := NewBuffer(make([]byte, 64))
	written, err := srcBuf.WriteTo(dstBuf)
	if err != ErrTooLarge {
		if err != nil {
			t.Errorf("Error while reading from buffer: %s", err.Error())
		} else {
			t.Error("Expected ErrTooLarge while reading from buffer")
		}
	}
	if written != 0 {
		t.Errorf("Should have written 0 bytes; wrote %d", written)
	}

	b := make([]byte, 64)
	// Check the resulting bytes
	if !bytes.Equal(dstBuf.Bytes(), b) {
		t.Fatalf("buffer content changed: %q not %q", dstBuf.Bytes(), b)
	}
}

func TestRuneWrites(t *testing.T) {
	const NRune = 1000
	// Built a test slice while we write the data
	b := make([]byte, utf8.UTFMax*NRune)
	buf := NewBuffer(make([]byte, 1872))
	n := 0
	for r := rune(0); r < NRune; r++ {
		size := utf8.EncodeRune(b[n:], r)
		nbytes, err := buf.WriteRune(r)
		if err != nil {
			t.Fatalf("WriteRune(%U) error: %s", r, err)
		}
		if nbytes != size {
			t.Fatalf("WriteRune(%U) expected %d, got %d", r, size, nbytes)
		}
		n += size
	}
	b = b[0:n]

	// Check the resulting bytes
	if !bytes.Equal(buf.Bytes(), b) {
		t.Fatalf("incorrect result from WriteRune: %q not %q", buf.Bytes(), b)
	}
}
