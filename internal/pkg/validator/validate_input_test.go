package validator

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateOperation(t *testing.T) {
	assert.Nil(t, validateOperation(DELTA_CMD))
	assert.Nil(t, validateOperation(SIGNATURE_CMD))
	assert.NotNil(t, validateOperation("dummyOp"))
}

func TestValidateSignatureParams(t *testing.T) {
	validFile := "../../../LICENSE"
	invalidFile := "../../../go.mod"

	invalidFile = validFile

	stat, err := os.Stat(validFile)
	if err != nil {
		t.Error(err)
	}
	size := stat.Size()
	resultSize, resultErr := validateSignatureParams([]string{validFile, "test"})
	assert.Nil(t, resultErr)
	assert.Equal(t, size, resultSize)

	resultSize, resultErr = validateSignatureParams([]string{invalidFile, "test"})
	assert.NotNil(t, resultErr)
	assert.Equal(t, resultSize, int64(0))

	resultSize, resultErr = validateSignatureParams([]string{"xyxyxyxy", "test"})
	assert.Equal(t, resultSize, int64(0))
	assert.NotNil(t, resultErr)

	resultSize, resultErr = validateSignatureParams([]string{"too", "many", "params"})
	assert.Equal(t, resultSize, int64(0))
	assert.NotNil(t, resultErr)

	resultSize, resultErr = validateSignatureParams([]string{})
	assert.Equal(t, resultSize, int64(0))
	assert.NotNil(t, resultErr)
}

func TestValidateDeltaParams(t *testing.T) {

}
