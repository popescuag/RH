package api

import (
	"github.com/popescuag/RH/internal/pkg/delta"
	"github.com/popescuag/RH/internal/pkg/signature"
)

func Signature(data []byte) ([]byte, error) {
	return signature.GetSignature(data)
}

func Delta(signatureData []byte, newData []byte) ([]byte, error) {
	return delta.GetDelta(signatureData, newData)
}
