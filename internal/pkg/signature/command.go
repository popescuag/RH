package signature

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

const METADATA_FORMAT = "%d,%d,%d\n"

func Compute(inputFileName string, outputFile string) error {
	f, err := os.Open(inputFileName)
	if err != nil {
		return err
	}
	defer f.Close()

	input := io.NopCloser(bufio.NewReader(f))
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	inputFileSize := fi.Size()

	chunkSize := computeChunkSize(inputFileSize)
	out, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer out.Close()

	return createSignatureFile(input, inputFileSize, chunkSize, out)
}

func createSignatureFile(input io.ReadCloser, inputFileSize int64, chunkSize int, output io.Writer) error {
	defer input.Close()

	chunkCount := inputFileSize / int64(chunkSize)
	if inputFileSize%int64(chunkSize) > 0 {
		chunkCount++
	}

	lastChunkSize := chunkSize
	if inputFileSize%int64(chunkSize) != 0 {
		lastChunkSize = int(inputFileSize % int64(chunkSize))
	}

	_, err := output.Write([]byte(fmt.Sprintf(METADATA_FORMAT, chunkSize, chunkCount, lastChunkSize)))
	if err != nil {
		return err
	}

	var totalBytesRead int64
	for {
		chunk := make([]byte, chunkSize)
		br, err := input.Read(chunk)
		if err != nil {
			return err
		}
		totalBytesRead += int64(br)
		_, err = output.Write([]byte(fmt.Sprintf("%x", sha256.Sum256(chunk[:br]))))
		if err != nil {
			return err
		}
		if totalBytesRead == inputFileSize {
			break
		}
	}

	return nil
}

func computeChunkSize(fileSize int64) int {
	chunkSize := 512
	if fileSize > 5<<20 && fileSize <= 50<<20 {
		chunkSize = 4 << 10
	} else if fileSize > 50<<20 && fileSize <= 100<<20 {
		chunkSize = 64 << 10
	} else if fileSize > 100<<20 && fileSize <= 300<<20 {
		chunkSize = 256 << 10
	} else if fileSize > 300<<20 && fileSize <= 1024<<20 {
		chunkSize = 1 << 20
	} else if fileSize > 1024<<20 {
		chunkSize = 4 << 20
	}
	return chunkSize
}
