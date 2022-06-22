package signature

import (
	"bytes"
	"io"
	"log"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSignatures(t *testing.T) {
	testCases := []struct {
		name           string
		inputData      []byte
		expectedResult []byte
		chunkSize      int
	}{
		{
			name:           "Last chunk smaller than the rest",
			inputData:      buildInput1(),
			expectedResult: buildOutput1(),
			chunkSize:      512,
		},
		{
			name:           "All chunks equal",
			inputData:      buildInput2(),
			expectedResult: buildOutput2(),
			chunkSize:      512,
		},
		{
			name:           "Large file 640MB",
			inputData:      buildInput3(),
			expectedResult: buildOutput3(),
			chunkSize:      1 << 20,
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

			//For output
			pro, pwo := io.Pipe()
			wg.Add(1)
			go func(r *io.PipeReader, len int, t *testing.T) {
				output := make([]byte, len)
				br, _ := io.ReadFull(pro, output)
				// t.Logf("%v\n%v", tc.name, string(output))
				// t.Logf("\n%v", string(tc.expectedResult))
				equal := bytes.Compare(output, tc.expectedResult)
				assert.Equal(t, 0, equal)
				assert.Equal(t, len, br)
				//t.Logf("%v %v", len, br)
				wg.Done()
			}(pro, len(tc.expectedResult), t)

			err := createSignatureFile(pri, int64(len(tc.inputData)), tc.chunkSize, pwo)
			pwo.Close()

			assert.Nil(t, err)
			wg.Wait()
			log.Printf("Test %v complete", tc.name)
		})
	}
}

func buildInput1() []byte {
	return make([]byte, 1<<10+10)
}

func buildOutput1() []byte {
	buf := new(bytes.Buffer)
	// 3 chunks, first 2 512, last 10 bytes
	chunk512 := make([]byte, 512)
	chunk10 := make([]byte, 10)

	md := signatureMetadata{
		ChunkSize:  512,
		ChunkCount: 3,
	}
	md.write(buf)
	writeChecksum(chunk512, buf)
	writeChecksum(chunk512, buf)
	writeChecksum(chunk10, buf)

	return buf.Bytes()
}

func buildInput2() []byte {
	return make([]byte, 1<<10)
}

func buildOutput2() []byte {
	buf := new(bytes.Buffer)
	// 2 chunks, 512 bytes each
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

func buildInput3() []byte {
	return make([]byte, 640<<20)
}

func buildOutput3() []byte {
	buf := new(bytes.Buffer)
	chunkSize := 1 << 20
	chunkCount := 640
	chunk1M := make([]byte, chunkSize)

	md := signatureMetadata{
		ChunkSize:  uint32(chunkSize),
		ChunkCount: uint32(chunkCount),
	}
	md.write(buf)
	for i := 0; i < chunkCount; i++ {
		writeChecksum(chunk1M, buf)
	}

	return buf.Bytes()
}

func TestComputeChunkSize(t *testing.T) {
	testCases := []struct {
		name           string
		fileSize       int64
		expectedResult int
	}{
		{
			name:           "XS file",
			fileSize:       2 << 10,
			expectedResult: 32,
		},
		{
			name:           "S file",
			fileSize:       7 << 20,
			expectedResult: 4 << 10,
		},
		{
			name:           "M file",
			fileSize:       60 << 20,
			expectedResult: 64 << 10,
		},
		{
			name:           "L file",
			fileSize:       150 << 20,
			expectedResult: 256 << 10,
		},
		{
			name:           "XL file",
			fileSize:       600 << 20,
			expectedResult: 1 << 20,
		},
		{
			name:           "XXL file",
			fileSize:       1025 << 20,
			expectedResult: 4 << 20,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedResult, computeChunkSize(tc.fileSize))
		})
	}
}
