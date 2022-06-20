package validator

import (
	"errors"
	"fmt"
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
	validFile := "testdata/validFileForSignature"     // larger than 1kb
	invalidFile := "testdata/invalidFileForSignature" // smaller than 1kb

	stat, err := os.Stat(validFile)
	if err != nil {
		t.Error(err)
	}
	size := stat.Size()

	testCases := []struct {
		name               string
		input              []string
		expectedSizeOutput int64
		expectedError      error
	}{
		{
			name:               "Valid test",
			input:              []string{validFile, "test"},
			expectedSizeOutput: size,
			expectedError:      nil,
		},
		{
			name:               "Too small input file",
			input:              []string{invalidFile, "test"},
			expectedSizeOutput: 0,
			expectedError:      fmt.Errorf("input file %v too small", invalidFile),
		},
		{
			name:               "Input file not found",
			input:              []string{"invalidFile", "test"},
			expectedSizeOutput: 0,
			expectedError:      fmt.Errorf("file %v not found", "invalidFile"),
		},
		{
			name:               "Too many params",
			input:              []string{"validFile", "test", "extraParam"},
			expectedSizeOutput: 0,
			expectedError:      errors.New("signature function requires exactly 2 parameters (3 provided)"),
		},
		{
			name:               "Too few params",
			input:              []string{},
			expectedSizeOutput: 0,
			expectedError:      errors.New("signature function requires exactly 2 parameters (0 provided)"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			size, err := validateSignatureParams(tc.input)
			assert.Equal(t, tc.expectedSizeOutput, size)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestValidateDeltaParams(t *testing.T) {
	testCases := []struct {
		name          string
		input         []string
		expectedError error
	}{
		{
			name:          "Valid test",
			input:         []string{"testdata/validSignatureFile", "testdata/validNewFile", "delta"},
			expectedError: nil,
		},
		{
			name:          "Invalid signature file",
			input:         []string{"testdata/invalidSignatureFile", "testdata/validNewFile", "delta"},
			expectedError: errors.New("file testdata/invalidSignatureFile is not a valid signature file"),
		},
		{
			name:          "Invalid new file",
			input:         []string{"testdata/validSignatureFile", "xyxyxy", "delta"},
			expectedError: errors.New("file xyxyxy cannot be found"),
		},
		{
			name:          "Invalid delta file",
			input:         []string{"testdata/validSignatureFile", "testdata/validNewFile", "testdata123/delta"},
			expectedError: errors.New("cannot create testdata123/delta file"),
		},
		{
			name:          "Too many params",
			input:         []string{"testdata/validSignatureFile", "testdata/validNewFile", "delta", "extraParam"},
			expectedError: errors.New("delta function requires exactly 3 parameters (4 provided)"),
		},
		{
			name:          "Too few params",
			input:         []string{},
			expectedError: errors.New("delta function requires exactly 3 parameters (0 provided)"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run(tc.name, func(t *testing.T) {
				err := validateDeltaParams(tc.input)
				assert.Equal(t, tc.expectedError, err)
			})
		})
	}
}
