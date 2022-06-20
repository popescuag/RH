package delta

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/popescuag/RH/internal/pkg/signature"
	"github.com/stretchr/testify/assert"
)

var (
	chunks = buildRandomChunks(512, 10)
)

func TestOutput(t *testing.T) {
	readOutput()
	t.Error()
}

func TestDeltas(t *testing.T) {

	testCases := []struct {
		name           string
		signature      []byte
		newFile        []byte
		expectedResult []byte
	}{
		{
			name:           "identical files",
			signature:      buildSignature(chunks[1:3], 512),
			newFile:        buildNewFile1(),
			expectedResult: buildOutput1(),
		},
		{
			name:           "2 chunks identical, last chunk different",
			signature:      buildSignature(chunks[1:3], 512),
			newFile:        buildNewFile2(),
			expectedResult: buildOutput2(),
		},
		{
			name:           "3rd chunk deleted, the rest are the same",
			signature:      buildSignature(chunks[1:3], 512),
			newFile:        buildNewFile3(),
			expectedResult: buildOutput3(),
		},
		{
			name:           "2nd chunk replaced, the rest are the same",
			signature:      buildSignature(chunks[1:3], 512),
			newFile:        buildNewFile4(),
			expectedResult: buildOutput4(),
		},
		{
			name:           "Shifted chunks",
			signature:      buildSignature(chunks[1:3], 512),
			newFile:        buildNewFile5(),
			expectedResult: buildOutput5(),
		},
		{
			name:           "No common chunks",
			signature:      buildSignature(chunks[1:3], 512),
			newFile:        buildNewFile6(),
			expectedResult: buildOutput6(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			prs, pws := io.Pipe()
			prf, pwf := io.Pipe()
			// Pass signature+new file data
			go writeBytes(tc.signature, pws, t)
			go writeBytes(tc.newFile, pwf, t)

			//Read output
			pro, pwo := io.Pipe()
			go func(r *io.PipeReader, len int, t *testing.T) {
				output := make([]byte, len)
				br, _ := io.ReadFull(pro, output)
				equal := bytes.Compare(output, tc.expectedResult)
				assert.Equal(t, 0, equal)
				assert.Equal(t, len, br)
			}(pro, len(tc.expectedResult), t)

			err := createDeltaFile(prs, prf, pwo)
			pwo.Close()

			assert.Nil(t, err)
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
		chunk := make([]byte, chunkSize)
		rand.Read(chunk)
		randomChunks[i] = chunk
		rand.Seed(time.Now().UnixNano())
	}

	return randomChunks
}

func buildSignature(chunks [][]byte, chunkSize int) []byte {
	chunksCount := len(chunks)
	metadata := []byte(fmt.Sprintf(signature.METADATA_FORMAT, chunkSize, chunksCount, chunkSize))
	var signature []byte
	signature = append(signature, metadata...)
	for i := 0; i < chunksCount; i++ {
		data := []byte(fmt.Sprintf("%x", sha256.Sum256(chunks[i])))
		signature = append(signature, data...)
	}
	return signature
}

func buildNewFile1() []byte {
	return bytes.Join(chunks[1:3], make([]byte, 0))
}

func buildOutput1() []byte {
	return []byte("512\nP1\nP2\nP3")
}

func buildNewFile2() []byte {
	newChunks := [][]byte{
		chunks[0], chunks[1], chunks[5],
	}
	return bytes.Join(newChunks, make([]byte, 0))
}

func buildOutput2() []byte {
	var outputData []byte
	commonData := []byte("512\nP1\nP2\nN")
	outputData = append(outputData, commonData...)
	outputData = append(outputData, base64Encode(chunks[5])...)
	return outputData
}

func buildNewFile3() []byte {
	newChunks := [][]byte{
		chunks[0], chunks[1],
	}
	return bytes.Join(newChunks, make([]byte, 0))
}

func buildOutput3() []byte {
	return []byte("512\nP1\nP2")
}

func buildNewFile4() []byte {
	newChunks := [][]byte{
		chunks[0], chunks[3], chunks[2],
	}
	return bytes.Join(newChunks, make([]byte, 0))
}

func buildOutput4() []byte {
	var outputData []byte
	data := []byte("512\nP1\nN")
	outputData = append(outputData, data...)
	outputData = append(outputData, base64Encode(chunks[3])...)
	data = []byte("\nP2")
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
	return []byte("512\nP3\nP1\nP2")
}

func buildNewFile6() []byte {
	newChunks := [][]byte{
		chunks[7], chunks[8], chunks[9],
	}
	return bytes.Join(newChunks, make([]byte, 0))
}

func buildOutput6() []byte {
	var outputData []byte
	data := []byte("512\nN")
	outputData = append(outputData, data...)
	outputData = append(outputData, base64Encode(chunks[7])...)
	data = []byte("\nN")
	outputData = append(outputData, data...)
	outputData = append(outputData, base64Encode(chunks[8])...)
	data = []byte("\nN")
	outputData = append(outputData, data...)
	outputData = append(outputData, base64Encode(chunks[9])...)
	return outputData
}

func TestEncodeDecode(t *testing.T) {
	encoded := base64Encode(chunks[0])
	decoded, _ := base64Decode(encoded)
	assert.Equal(t, chunks[0], decoded)
}

func base64Encode(data []byte) []byte {
	return []byte(base64.StdEncoding.EncodeToString(data))
}

func base64Decode(data []byte) ([]byte, error) {
	var result []byte
	result, err := base64.StdEncoding.DecodeString(string(data))
	return result, err
}

func readOutput() {

	outputReader := bytes.NewReader(buildOutput2())
	scanner := bufio.NewScanner(outputReader)
	scanner.Split(bufio.ScanLines)
	firstLine := true

	for scanner.Scan() {
		if firstLine {
			metadata := scanner.Text()
			log.Printf("%v", metadata)
			firstLine = false
			continue
		}
		line := scanner.Text()
		if line[0] == 'P' {
			log.Print(line[1:])
		} else {
			d, err := base64Decode([]byte(line[1:]))
			if err != nil {
				panic(fmt.Sprintf("Error decoding test data: %v", err))
			}
			e := bytes.Equal(chunks[5], d)
			log.Printf("%v", e)
		}
	}
}
