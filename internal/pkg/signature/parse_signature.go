package signature

import (
	"bufio"
	"errors"
	"io"
	"os"
)

type SignatureData struct {
	Metadata  signatureMetadata
	Checksums []string
}

func ParseFromFile(signatureFile string) (SignatureData, error) {
	f, err := os.Open(signatureFile)
	signatureData := SignatureData{}
	if err != nil {
		return signatureData, err
	}
	defer f.Close()

	input := io.NopCloser(bufio.NewReader(f))
	signatureData, err = ParseFromReader(input)
	return signatureData, err
}

func ParseFromReader(input io.ReadCloser) (SignatureData, error) {
	md := signatureMetadata{}
	signatureData := SignatureData{}
	defer input.Close()

	err := md.read(input)

	if err != nil && err != io.EOF {
		return SignatureData{}, err
	}
	if md.ChunkSize < 32 {
		return signatureData, errors.New("invalid signature file metadata: chunkSize too small")
	}
	signatureData.Metadata.ChunkCount = md.ChunkCount
	signatureData.Metadata.ChunkSize = md.ChunkSize

	signatureData.Checksums = make([]string, signatureData.Metadata.ChunkCount)
	chunksRead := 0
	for i := 0; i < int(md.ChunkCount); i++ {
		sum := make([]byte, 32)
		err = readChecksum(input, sum)
		if err == io.EOF {
			break
		}
		if err != nil && err != io.EOF {
			return SignatureData{}, err
		}
		signatureData.Checksums[i] = string(sum)
		chunksRead++
	}
	if chunksRead < int(md.ChunkCount) {
		return SignatureData{}, errors.New("invalid signature file: size too small")
	}
	if chunksRead > int(md.ChunkCount) {
		return SignatureData{}, errors.New("invalid signature file: size too large")
	}
	return signatureData, nil
}
