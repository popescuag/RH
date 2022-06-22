package delta

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"testing"
	"time"

	s "github.com/popescuag/RH/internal/pkg/signature"
	"github.com/stretchr/testify/assert"
)

var (
	chunks       = buildRandomChunks(512, 6)
	smallerChunk = buildRandomChunk(64)
)

func TestDeltas(t *testing.T) {
	testCases := []struct {
		name           string
		signature      s.SignatureData
		newFile        []byte
		expectedResult []byte
	}{
		{
			name:           "identical files",
			signature:      s.BuildSignatureData(chunks[0:3], 512),
			newFile:        buildNewFile1(),
			expectedResult: buildOutput1(),
		},
		{
			name:           "2 chunks identical, last chunk different",
			signature:      s.BuildSignatureData(chunks[0:3], 512),
			newFile:        buildNewFile2(),
			expectedResult: buildOutput2(),
		},

		{
			name:           "3rd chunk deleted, the rest are the same",
			signature:      s.BuildSignatureData(chunks[0:3], 512),
			newFile:        buildNewFile3(),
			expectedResult: buildOutput3(),
		},

		{
			name:           "2nd chunk replaced, the rest are the same",
			signature:      s.BuildSignatureData(chunks[0:3], 512),
			newFile:        buildNewFile4(),
			expectedResult: buildOutput4(),
		},

		{
			name:           "Shifted chunks",
			signature:      s.BuildSignatureData(chunks[0:3], 512),
			newFile:        buildNewFile5(),
			expectedResult: buildOutput5(),
		},
		{
			name:           "No common chunks",
			signature:      s.BuildSignatureData(chunks[0:3], 512),
			newFile:        buildNewFile6(),
			expectedResult: buildOutput6(),
		},
		{
			name:           "Smaller last chunk",
			signature:      s.BuildSignatureData(chunks[0:3], 512),
			newFile:        buildNewFile7(),
			expectedResult: buildOutput7(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var wg sync.WaitGroup

			prf, pwf := io.Pipe()
			// Pass new file data
			go writeBytes(tc.newFile, pwf, t)

			//Read output
			pro, pwo := io.Pipe()
			wg.Add(1)
			go func(r *io.PipeReader, len int, t *testing.T) {
				output := make([]byte, len)
				br, _ := io.ReadFull(pro, output)
				equal := bytes.Compare(output, tc.expectedResult)
				assert.Equal(t, 0, equal, "Expected 0, got %v", equal)
				assert.Equal(t, len, br)
				defer wg.Done()
			}(pro, len(tc.expectedResult), t)

			err := createDelta(tc.signature, prf, int64(len(tc.newFile)), pwo)
			assert.Nil(t, err)
			pwo.Close()

			wg.Wait()

			t.Logf("Test %v completed", tc.name)
		})
	}
}

func writeBytes(arr []byte, w *io.PipeWriter, t *testing.T) {
	bw, err := io.Copy(w, bytes.NewReader(arr))
	if err != nil {
		t.Errorf("%v", err)
	}
	assert.Equal(t, bw, int64(len(arr)))
	if bw < int64(len(arr)) {
		t.Errorf("Less bytes written than expected: %v", bw)
	}
	w.Close()
}

func buildRandomChunks(chunkSize int, chunkCount int) [][]byte {
	randomChunks := make([][]byte, chunkCount)
	for i := 0; i < chunkCount; i++ {
		randomChunks[i] = buildRandomChunk(chunkSize)
	}
	return randomChunks
}

func buildRandomChunk(chunkSize int) []byte {
	chunk := make([]byte, chunkSize)
	rand.Seed(time.Now().UnixNano())
	rand.Read(chunk)
	return chunk
}

func buildNewFile1() []byte {
	return bytes.Join(chunks[0:3], make([]byte, 0))
}

func buildOutput1() []byte {
	//512|P,4,0P,4,1P,4,2
	return []byte(fmt.Sprintf("512%v%v%v4%v0%v%v4%v1%v%v4%v2", dataSeparator, pointerMark, fieldSeparator,
		fieldSeparator, pointerMark, fieldSeparator, fieldSeparator, pointerMark, fieldSeparator, fieldSeparator))
}

func buildNewFile2() []byte {
	newChunks := [][]byte{
		chunks[0], chunks[1], chunks[5],
	}
	return bytes.Join(newChunks, make([]byte, 0))
}

func buildOutput2() []byte {
	//512|P,4,0P,4,1N,512,...
	var outputData []byte
	commonData := []byte(fmt.Sprintf("512%v%v%v4%v0%v%v4%v1%v%v512%v", dataSeparator, pointerMark, fieldSeparator,
		fieldSeparator, pointerMark, fieldSeparator, fieldSeparator, newChunkMark, fieldSeparator, fieldSeparator))
	outputData = append(outputData, commonData...)
	outputData = append(outputData, chunks[5]...)
	return outputData
}

func buildNewFile3() []byte {
	newChunks := [][]byte{
		chunks[0], chunks[1],
	}
	return bytes.Join(newChunks, make([]byte, 0))
}

func buildOutput3() []byte {
	//512|P,4,0P,4,1
	return []byte(fmt.Sprintf("512%v%v%v4%v0%v%v4%v1", dataSeparator, pointerMark, fieldSeparator, fieldSeparator,
		pointerMark, fieldSeparator, fieldSeparator))
}

func buildNewFile4() []byte {
	newChunks := [][]byte{
		chunks[0], chunks[3], chunks[2],
	}
	return bytes.Join(newChunks, make([]byte, 0))
}

func buildOutput4() []byte {
	//512|P,4,0N,512,...P,4,2
	var outputData []byte
	data := []byte(fmt.Sprintf("512%v%v%v4%v0%v%v512%v", dataSeparator, pointerMark, fieldSeparator, fieldSeparator,
		newChunkMark, fieldSeparator, fieldSeparator))
	outputData = append(outputData, data...)
	outputData = append(outputData, chunks[3]...)
	data = []byte(fmt.Sprintf("%v%v4%v2", pointerMark, fieldSeparator, fieldSeparator))
	outputData = append(outputData, data...)
	return outputData
}

func buildNewFile5() []byte {
	newChunks := [][]byte{
		chunks[2], chunks[0], chunks[1],
	}
	return bytes.Join(newChunks, make([]byte, 0))
}

func buildOutput5() []byte {
	//512|P,4,2P,4,0P,4,1
	return []byte(fmt.Sprintf("512%v%v%v4%v2%v%v4%v0%v%v4%v1", dataSeparator, pointerMark, fieldSeparator,
		fieldSeparator, pointerMark, fieldSeparator, fieldSeparator, pointerMark, fieldSeparator, fieldSeparator))
}

func buildNewFile6() []byte {
	newChunks := [][]byte{
		chunks[3], chunks[4], chunks[5],
	}
	return bytes.Join(newChunks, make([]byte, 0))
}

func buildOutput6() []byte {
	//512|N,512,...N,512,...N,512,...
	var outputData []byte
	data := []byte(fmt.Sprintf("512%v%v%v512%v", dataSeparator, newChunkMark, fieldSeparator, fieldSeparator))
	outputData = append(outputData, data...)
	outputData = append(outputData, chunks[3]...)
	data = []byte(fmt.Sprintf("%v%v512%v", newChunkMark, fieldSeparator, fieldSeparator))
	outputData = append(outputData, data...)
	outputData = append(outputData, chunks[4]...)
	data = []byte(fmt.Sprintf("%v%v512%v", newChunkMark, fieldSeparator, fieldSeparator))
	outputData = append(outputData, data...)
	outputData = append(outputData, chunks[5]...)
	return outputData
}

func buildNewFile7() []byte {
	newChunks := [][]byte{
		chunks[0], chunks[1], smallerChunk,
	}
	return bytes.Join(newChunks, make([]byte, 0))
}

func buildOutput7() []byte {
	//512|P,4,0P,4,1N,512,...
	var outputData []byte
	commonData := []byte(fmt.Sprintf("512%v%v%v4%v0%v%v4%v1%v%v64%v", dataSeparator, pointerMark, fieldSeparator,
		fieldSeparator, pointerMark, fieldSeparator, fieldSeparator, newChunkMark, fieldSeparator, fieldSeparator))
	outputData = append(outputData, commonData...)
	outputData = append(outputData, smallerChunk...)
	return outputData
}
