package signature

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"io"
)

type signatureMetadata struct {
	ChunkSize  uint32
	ChunkCount uint32
}

func (md *signatureMetadata) write(output io.Writer) error {
	return binary.Write(output, binary.LittleEndian, md)
}

func (md *signatureMetadata) read(input io.Reader) error {
	return binary.Read(input, binary.LittleEndian, md)
}

func writeChecksum(chunk []byte, output io.Writer) error {
	return binary.Write(output, binary.LittleEndian, sha256.Sum256(chunk))
}

func readChecksum(input io.Reader, sum []byte) error {
	return binary.Read(input, binary.LittleEndian, sum)
}

func GetChecksum(chunk []byte) string {
	buf := new(bytes.Buffer)
	writeChecksum(chunk, buf)
	return buf.String()
}
