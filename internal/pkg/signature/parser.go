package signature

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
)

type SignatureData struct {
	ChunkSize     int
	ChunkCount    int
	LastChunkSize int
	Sums          []string
}

func Parse(signatureFile string) (SignatureData, error) {
	f, err := os.Open(signatureFile)
	signatureData := SignatureData{}
	if err != nil {
		return signatureData, err
	}
	fi, err := f.Stat()
	if err != nil {
		return signatureData, err
	}
	defer f.Close()

	input := io.NopCloser(bufio.NewReader(f))
	signatureData, err = parseSignatureFile(input, fi.Size())
	return signatureData, err
}

func parseSignatureFile(input io.ReadCloser, fileSize int64) (SignatureData, error) {
	signatureData := SignatureData{}

	//log.Printf("Parse started: %v bytes", fileSize)
	scanner := bufio.NewScanner(input)
	scanner.Split(bufio.ScanLines)
	firstLine := true

	for scanner.Scan() {
		if firstLine {
			// Metadata
			metadata := scanner.Text()
			metadataArr := strings.Split(string(metadata), ",")
			//log.Printf("Metadata read: %v,%v,%v", metadataArr[0], metadataArr[1], metadataArr[2])
			if len(metadataArr) != 3 {
				return signatureData, errors.New("invalid signature file metadata: too many or too few elements")
			}
			v, err := strconv.Atoi(metadataArr[0])
			if err != nil {
				return signatureData, errors.New("invalid signature file metadata: value not an integer")
			}
			signatureData.ChunkSize = v
			v, err = strconv.Atoi(metadataArr[1])
			if err != nil {
				return signatureData, errors.New("invalid signature file metadata: value not an integer")
			}
			if signatureData.ChunkSize < 512 {
				return signatureData, errors.New("invalid signature file metadata: chunkSize too small")
			}

			signatureData.ChunkCount = v
			v, err = strconv.Atoi(metadataArr[2])
			if err != nil {
				return signatureData, errors.New("invalid signature file metadata: value not an integer")
			}
			signatureData.LastChunkSize = v

			requiredSize := int64(len([]byte(metadata)) + 1 + signatureData.ChunkCount*64)
			// Check the size
			if fileSize > requiredSize {
				return signatureData, errors.New("invalid signature file: size too large")
			}
			if fileSize < requiredSize {
				return signatureData, errors.New("invalid signature file: size too small")
			}

			firstLine = false
			continue
		}
		// Sums
		sums := scanner.Text()
		signatureData.Sums = make([]string, signatureData.ChunkCount)
		idx := 0
		for i := 0; i < signatureData.ChunkCount; i++ {
			signatureData.Sums[i] = sums[idx : idx+64]
			idx += 64
			//log.Printf("Sum[%d]: %v", i, signatureData.Sums[i])
		}
	}
	input.Close()

	return signatureData, nil
}
