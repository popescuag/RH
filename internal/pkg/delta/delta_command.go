package delta

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	s "github.com/popescuag/RH/internal/pkg/signature"
)

const (
	pointerMark    = "P"
	newChunkMark   = "N"
	fieldSeparator = ","
	dataSeparator  = "|"
)

// GetDelta = computes deltas based on signature data and the new file
func GetDelta(signatureData []byte, newData []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	len64 := int64(len(newData))
	sd, err := s.ParseFromReader(io.NopCloser(bytes.NewReader(signatureData)))
	createDelta(sd, io.NopCloser(bytes.NewReader(newData)), len64, buf)

	return buf.Bytes(), err
}

func Compute(signatureFile string, newFile string, deltaFile string) error {
	signatureData, err := s.ParseFromFile(signatureFile)
	if err != nil {
		return err
	}

	f, err := os.Open(newFile)
	if err != nil {
		return err
	}
	inputReader := io.NopCloser(bufio.NewReader(f))
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	out, err := os.Create(deltaFile)
	if err != nil {
		return err
	}
	defer out.Close()

	return createDelta(signatureData, inputReader, fi.Size(), out)
}

func createDelta(signatureData s.SignatureData, newFile io.ReadCloser, newFileSize int64, output io.Writer) error {
	//Write metadata first
	_, err := output.Write([]byte(fmt.Sprintf("%d%v", signatureData.Metadata.ChunkSize, dataSeparator)))
	if err != nil {
		return err
	}
	defer newFile.Close()

	var totalBytesRead int64
	for {
		chunk := make([]byte, signatureData.Metadata.ChunkSize)
		br, err := newFile.Read(chunk)
		if err != nil {
			return err
		}
		totalBytesRead += int64(br)

		checksum := s.GetChecksum(chunk[:br])
		index := findChecksum(checksum, signatureData.Checksums)

		// Write a pointer to a chunk from the original file if the signature of this chunk was found
		// or the new chunk otherwise
		if index == -1 {
			err = writeNewChunk(chunk[:br], output)
		} else {
			err = writePointer(uint32(index), output)
		}

		if err != nil {
			return err
		}

		if totalBytesRead == newFileSize {
			break
		}
	}
	return nil
}

func writePointer(index uint32, out io.Writer) error {
	_, err := out.Write([]byte(fmt.Sprintf("%v%v%d%v%d", pointerMark, fieldSeparator,
		4, fieldSeparator, index)))
	return err
}

func writeNewChunk(newChunk []byte, out io.Writer) error {
	_, err := out.Write([]byte(fmt.Sprintf("%v%v%v%v", newChunkMark, fieldSeparator,
		len(newChunk), fieldSeparator)))
	if err != nil {
		return err
	}
	_, err = out.Write(newChunk)
	return err
}

func findChecksum(newChecksum string, checksums []string) int {
	index := -1
	for i, checksum := range checksums {
		if newChecksum == checksum {
			index = i
			break
		}
	}
	return index
}
