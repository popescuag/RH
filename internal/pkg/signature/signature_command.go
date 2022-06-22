package signature

import (
	"bufio"
	"bytes"
	"io"
	"os"
)

func GetSignature(data []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	len64 := int64(len(data))
	chunkSize := computeChunkSize(len64)
	err := createSignatureFile(io.NopCloser(bytes.NewReader(data)), len64, chunkSize, buf)
	return buf.Bytes(), err
}

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

	md := signatureMetadata{}

	chunkCount := uint32(inputFileSize / int64(chunkSize))
	if inputFileSize%int64(chunkSize) > 0 {
		chunkCount++
	}
	md.ChunkCount = chunkCount
	md.ChunkSize = uint32(chunkSize)

	err := md.write(output)

	//err := binary.Write(output, binary.LittleEndian, binarySignature)
	if err != nil {
		return err
	}
	var totalBytesRead int64
	var chunkIndex uint32
	for {
		chunk := make([]byte, chunkSize)
		br, err := input.Read(chunk)
		if err != nil {
			return err
		}
		totalBytesRead += int64(br)
		//err = binary.Write(output, binary.LittleEndian, sha256.Sum256(chunk[:br]))
		err = writeChecksum(chunk[:br], output)
		if err != nil {
			return err
		}
		if totalBytesRead == inputFileSize {
			break
		}
		chunkIndex++
	}
	return nil
}

func BuildSignatureData(chunks [][]byte, chunkSize uint32) SignatureData {
	chunksCount := len(chunks)
	signatureData := SignatureData{
		Metadata: signatureMetadata{
			ChunkSize:  chunkSize,
			ChunkCount: uint32(chunksCount),
		},
	}
	signatureData.Checksums = make([]string, chunksCount)

	for i := 0; i < chunksCount; i++ {
		buf := new(bytes.Buffer)
		writeChecksum(chunks[i], buf)
		signatureData.Checksums[i] = buf.String()
	}
	return signatureData
}

func computeChunkSize(fileSize int64) int {
	chunkSize := 32
	if fileSize > 5<<20 && fileSize <= 50<<20 {
		chunkSize = 4 << 10 //4k
	} else if fileSize > 50<<20 && fileSize <= 100<<20 {
		chunkSize = 64 << 10 //64k
	} else if fileSize > 100<<20 && fileSize <= 300<<20 {
		chunkSize = 256 << 10 //256k
	} else if fileSize > 300<<20 && fileSize <= 1024<<20 {
		chunkSize = 1 << 20 //1M
	} else if fileSize > 1024<<20 {
		chunkSize = 4 << 20 //4M
	}
	return chunkSize
}
