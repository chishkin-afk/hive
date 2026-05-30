package config

import "fmt"

type InvalidCertPathError struct {
	original error
	path     string
}

func (icpe *InvalidCertPathError) Error() string {
	return fmt.Sprintf("path to certificate is invalid: %s", icpe.path)
}

func (icpe *InvalidCertPathError) Unwrap() error {
	return icpe.original
}

type InvalidKeyPathError struct {
	original error
	path     string
}

func (icpe *InvalidKeyPathError) Error() string {
	return fmt.Sprintf("path to key is invalid: %s", icpe.path)
}

func (icpe *InvalidKeyPathError) Unwrap() error {
	return icpe.original
}
