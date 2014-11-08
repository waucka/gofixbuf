# gofixbuf

## What is it?

`gofixbuf` is an implementation of a byte buffer derived from `bytes.Buffer`.
However, it is not a drop-in replacement.  This version cannot grow.  Also,
all means of reading the buffer other than `Bytes()` and `String()` have been removed.
Thus, it is, in a sense, "fixed"

Here's an example of the kind of situation `gofixbuf` was invented for:

```go
func (self *SomeChunkedDataThingy) Read(p []byte) (int, error) {
	outBuf := gofixbuf.NewBuffer(p)
	var totalBytesWritten int64 = 0
	for {
		if self.readingChunkData == nil {
			chunkData, err := self.getCurrentChunkData()
			if err != nil {
				return int(totalBytesWritten), err
			}
			self.readingChunkData = bytes.NewBuffer(chunkData)
		}
		bytesWritten, err := outBuf.ReadFrom(self.readingChunkData)
		totalBytesWritten += bytesWritten
		if err != nil && err != io.EOF {
			return int(totalBytesWritten), err
		}
		if err == io.EOF {
			self.readingChunkData = nil
			self.readingChunkNum++
		}
		if bytesWritten == 0 {
			return int(totalBytesWritten), nil
		}
	}
}
```

## Types and Values

```go
type Buffer struct {
	buf       []byte
	off       int
	runeBytes [utf8.UTFMax]byte
}

var ErrTooLarge = errors.New("gofixbuf.Buffer: too large")
```

## Functions

```go
func NewBuffer(buf []byte) *Buffer
func (b *Buffer) Bytes() []byte
func (b *Buffer) String() string
func (b *Buffer) Len() int
func (b *Buffer) Cap() int
func (b *Buffer) Reset()
func (b *Buffer) Write(p []byte) (n int, err error)
func (b *Buffer) WriteString(s string) (n int, err error)
func (b *Buffer) WriteByte(c byte) error
func (b *Buffer) WriteRune(r rune) (n int, err error)
func (b *Buffer) ReadFrom(r io.Reader) (n int64, err error)
```
