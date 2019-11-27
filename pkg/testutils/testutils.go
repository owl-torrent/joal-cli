package testutils

import (
	"github.com/go-playground/validator/v10"
	"testing"
)

type ErrorDescription struct {
	ErrorFieldPath string
	ErrorTag       string
}

func AssertValidateError(t *testing.T, validationErrors validator.ValidationErrors, expectedError ErrorDescription) {
	fieldFound := false
	tagFound := false
	for _, e := range validationErrors {
		if e.Namespace() == expectedError.ErrorFieldPath {
			fieldFound = true
			if e.Tag() == expectedError.ErrorTag {
				tagFound = true
			}
		}
	}
	if !fieldFound || !tagFound {
		t.Errorf("Wanted error was not found, expected '%v' to contains an error for path=%s and tag=%s", validationErrors, expectedError.ErrorFieldPath, expectedError.ErrorTag)
	}
}
