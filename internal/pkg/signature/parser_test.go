package signature

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSignatureFile(t *testing.T) {
	testCases := []struct {
		name           string
		inputData      []byte
		expectedResult SignatureData
		err            error
	}{
		{
			name:           "Valid signature file",
			inputData:      buildValidSignatureFile(),
			expectedResult: buildValidParseOutput(),
		},
		{
			name:      "Invalid signature file metadata: chunkSize too small",
			inputData: buildInvalidSignatureMetadata1(),
			expectedResult: SignatureData{
				ChunkSize: 64,
			},
			err: errors.New("invalid signature file metadata: chunkSize too small"),
		},
		{
			name:           "Invalid signature file metadata: too many or too few elements",
			inputData:      buildInvalidSignatureMetadata2(),
			expectedResult: SignatureData{},
			err:            errors.New("invalid signature file metadata: too many or too few elements"),
		},
		{
			name:      "Invalid signature file: size too small",
			inputData: buildInvalidSignatureMetadata3(),
			expectedResult: SignatureData{
				ChunkSize:     512,
				ChunkCount:    2,
				LastChunkSize: 512,
			},
			err: errors.New("invalid signature file: size too small"),
		},
		{
			name:      "Invalid signature file: size too large",
			inputData: buildInvalidSignatureMetadata4(),
			expectedResult: SignatureData{
				ChunkSize:     512,
				ChunkCount:    2,
				LastChunkSize: 512,
			},
			err: errors.New("invalid signature file: size too large"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var wg sync.WaitGroup
			// For input
			pri, pwi := io.Pipe()

			wg.Add(1)
			go func(w *io.PipeWriter, t *testing.T) {
				bw, err := io.Copy(w, bytes.NewReader(tc.inputData))
				if err != nil {
					t.Errorf("%v", err)
				}
				assert.Equal(t, bw, int64(len(tc.inputData)))
				if bw < int64(len(tc.inputData)) {
					t.Errorf("Less bytes written than expected: %v", bw)
				}
				w.Close()
				wg.Done()
			}(pwi, t)

			signatureData, err := parseSignatureFile(pri, int64(len(tc.inputData)))

			wg.Wait()

			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.expectedResult, signatureData)
			log.Printf("Test %v complete", tc.name)
		})
	}

}

func buildValidSignatureFile() []byte {
	// 2 chunks, 512 bytes each
	sum512 := sha256.Sum256(make([]byte, 512))
	pattern := METADATA_FORMAT + "%x%x"
	return []byte(fmt.Sprintf(pattern, 512, 2, 512, sum512, sum512))
}

func buildValidParseOutput() SignatureData {
	sum512 := fmt.Sprintf("%x", sha256.Sum256(make([]byte, 512)))
	return SignatureData{
		ChunkSize:     512,
		ChunkCount:    2,
		LastChunkSize: 512,
		Sums:          []string{sum512, sum512},
	}
}

func buildInvalidSignatureMetadata1() []byte {
	// chunk size too small
	sum512 := sha256.Sum256(make([]byte, 512))
	pattern := METADATA_FORMAT + "%x%x"
	return []byte(fmt.Sprintf(pattern, 64, 2, 64, sum512, sum512))
}

func buildInvalidSignatureMetadata2() []byte {
	// too few metadata elements
	sum512 := sha256.Sum256(make([]byte, 512))
	pattern := "%d%d%x%x"
	return []byte(fmt.Sprintf(pattern, 512, 5, sum512, sum512))
}

func buildInvalidSignatureMetadata3() []byte {
	// size too small
	sum512 := sha256.Sum256(make([]byte, 512))
	pattern := METADATA_FORMAT + "%x"
	return []byte(fmt.Sprintf(pattern, 512, 2, 512, sum512))
}

func buildInvalidSignatureMetadata4() []byte {
	// size too large
	sum512 := sha256.Sum256(make([]byte, 512))
	pattern := METADATA_FORMAT + "%x%x%x"
	return []byte(fmt.Sprintf(pattern, 512, 2, 512, sum512, sum512, sum512))
}
