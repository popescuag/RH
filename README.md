# RH

## Usage
### Signature
`go run cmd/main.go signature /path/to/input/file /path/to/signature/file`

From other modules use `api.GetSignature`

### Delta
`go run cmd/main.go delta /path/to/signature/file /path/to/new/file /path/to/delta/file`

From other modules use `api.GetDelta`