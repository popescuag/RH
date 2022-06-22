package signature

import (
	"bytes"
	"errors"
	"io"
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
			name:           "Invalid signature file metadata: chunkSize too small",
			inputData:      buildInvalidSignatureMetadata1(),
			expectedResult: SignatureData{},
			err:            errors.New("invalid signature file metadata: chunkSize too small"),
		},
		{
			name:           "Invalid signature file: size too small",
			inputData:      buildInvalidSignatureMetadata2(),
			expectedResult: SignatureData{},
			err:            errors.New("invalid signature file: size too small"),
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

			signatureData, err := ParseFromReader(pri)

			wg.Wait()

			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.expectedResult, signatureData)
			t.Logf("Test %v complete", tc.name)
		})
	}

}

func buildValidSignatureFile() []byte {
	// 2 chunks, 512 bytes each
	buf := new(bytes.Buffer)
	chunk512 := make([]byte, 512)

	md := signatureMetadata{
		ChunkSize:  512,
		ChunkCount: 2,
	}
	md.write(buf)
	writeChecksum(chunk512, buf)
	writeChecksum(chunk512, buf)

	return buf.Bytes()
}

func buildValidParseOutput() SignatureData {
	buf := new(bytes.Buffer)
	writeChecksum(make([]byte, 512), buf)
	sum512 := buf.Bytes()
	return SignatureData{
		Metadata: signatureMetadata{
			ChunkSize:  512,
			ChunkCount: 2},
		Checksums: []string{string(sum512), string(sum512)},
	}
}

func buildInvalidSignatureMetadata1() []byte {
	// chunk size too small
	buf := new(bytes.Buffer)

	md := signatureMetadata{
		ChunkSize:  31,
		ChunkCount: 2,
	}
	md.write(buf)
	return buf.Bytes()
}

func buildInvalidSignatureMetadata2() []byte {
	// size too small
	buf := new(bytes.Buffer)
	chunk512 := make([]byte, 512)

	md := signatureMetadata{
		ChunkSize:  512,
		ChunkCount: 2,
	}
	md.write(buf)
	writeChecksum(chunk512, buf)
	return buf.Bytes()
}
