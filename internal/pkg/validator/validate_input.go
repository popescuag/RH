package validator

import (
	"errors"
	"fmt"
	"os"

	"github.com/popescuag/RH/internal/pkg/signature"
)

const (
	SIGNATURE_CMD = "signature"
	DELTA_CMD     = "delta"
	min_file_size = 64
)

func ValidateInputParams(params []string) error {
	if len(params) < 3 {
		return fmt.Errorf("3 or more parameters expected (%d provided)", len(params))
	}

	err := validateOperation(params[0])
	if err != nil {
		return err
	}

	if params[0] == SIGNATURE_CMD {
		_, err = validateSignatureParams(params[1:])
	} else {
		err = validateDeltaParams(params[1:])
	}

	return err
}

func validateOperation(operation string) error {
	if operation != SIGNATURE_CMD && operation != DELTA_CMD {
		return errors.New("first paramter should be signature or delta")
	}
	return nil
}

func validateSignatureParams(params []string) (int64, error) {
	if len(params) != 2 {
		return 0, fmt.Errorf("signature function requires exactly 2 parameters (%d provided)", len(params))
	}

	s, err := os.Stat(params[0])
	if err != nil {
		return 0, fmt.Errorf("file %v not found", params[0])
	}

	if s.Size() < min_file_size {
		return 0, fmt.Errorf("input file %v too small", params[0])
	}

	return s.Size(), nil
}

func validateDeltaParams(params []string) error {
	if len(params) != 3 {
		return fmt.Errorf("delta function requires exactly 3 parameters (%d provided)", len(params))
	}

	_, err := os.Stat(params[0])
	if err != nil {
		return fmt.Errorf("signature file %v not found", params[0])
	}

	_, err = os.Stat(params[1])
	if err != nil {
		return fmt.Errorf("file %v cannot be found", params[1])
	}

	_, err = signature.ParseFromFile(params[0])
	if err != nil {
		return fmt.Errorf("file %v is not a valid signature file", params[0])
	}

	f, err := os.Create(params[2])
	if err != nil {
		return fmt.Errorf("cannot create %v file", params[2])
	}
	defer f.Close()

	return nil
}
