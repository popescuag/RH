package delta

import "io"

const (
	POINTER   = "pointer"
	NEW_CHUNK = "new"
)

func Compute(signatureFile string, newlFile string, deltaFile string) error {
	return nil
}

func createDeltaFile(signatureFile io.ReadCloser, newFile io.ReadCloser, output io.Writer) error {
	return nil
}
