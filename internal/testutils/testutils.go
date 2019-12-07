package testutils

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"sync"
	"testing"
	"time"
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

func WaitOrFailAfterTimeout(wg *sync.WaitGroup, timeout time.Duration) error {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return nil // completed normally
	case <-time.After(timeout):
		return errors.New("WaitGroup.Wait() timeout") // timed out
	}
}
