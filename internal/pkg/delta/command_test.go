package delta

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"testing"

	"github.com/popescuag/RH/internal/pkg/signature"
	"github.com/stretchr/testify/assert"
)

func TestDeltas(t *testing.T) {
	testCases := []struct {
		name           string
		signature      []byte
		newFile        []byte
		expectedResult []byte
	}{
		{
			name:           "identical files",
			signature:      buildSignature1(),
			newFile:        buildNewFile1(),
			expectedResult: buildOutput1(),
		},
		/*
			{
				name:           "2 chunks identical, last chunk different",
				signature:      buildSignature1(),
				newFile:        buildNewFile1(),
				expectedResult: buildOutput1(),
			},
			{
				name:           "3rd chunk deleted, the rest are the same",
				signature:      buildSignature1(),
				newFile:        buildNewFile1(),
				expectedResult: buildOutput1(),
			},
			{
				name:           "2nd chunk replaced, the rest are the same",
				signature:      buildSignature1(),
				newFile:        buildNewFile1(),
				expectedResult: buildOutput1(),
			},
			{
				name:           "Shifted chunks",
				signature:      buildSignature1(),
				newFile:        buildNewFile1(),
				expectedResult: buildOutput1(),
			},
			{
				name:           "No common chunks",
				signature:      buildSignature1(),
				newFile:        buildNewFile1(),
				expectedResult: buildOutput1(),
			},
			{
				name:           "1 common chunk",
				signature:      buildSignature1(),
				newFile:        buildNewFile1(),
				expectedResult: buildOutput1(),
			},
		*/
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
				assert.Equal(t, equal, 0)
				assert.Equal(t, br, len)
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

func buildSignature1() []byte {
	chunkSize := 1024
	metadata := []byte(fmt.Sprintf(signature.METADATA_FORMAT, chunkSize, 3, chunkSize))
	data := []byte(fmt.Sprintf("%x", sha256.Sum256(make([]byte, chunkSize))))
	var signature []byte
	signature = append(signature, metadata...)
	for i := 0; i < 3; i++ {
		signature = append(signature, data...)
	}
	return signature
}

func buildNewFile1() []byte {
	return make([]byte, 3<<10)
}

func buildOutput1() []byte {
	return make([]byte, 1)
}
