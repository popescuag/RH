package signature

type SignatureData struct {
	ChunkSize     int
	ChunkCount    int
	LastChunkSize int
	Sums          []string
}

func Parse(signatureFile string) (SignatureData, error) {
	return SignatureData{}, nil
}
