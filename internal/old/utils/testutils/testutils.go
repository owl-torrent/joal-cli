package testutils

import (
	"github.com/go-playground/validator/v10"
	"net/url"
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

// Await for the WaitGroup to unlock until the timeout occurs.
// If the WaitGroup unlock no error is returned, if the timeout occurs first an error is returned
func WaitOrFailAfterTimeout(t *testing.T, wg *sync.WaitGroup, timeout time.Duration) {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return // completed normally
	case <-time.After(timeout):
		t.Fail() // timed out
		t.Errorf("\n%s", "timed out while waiting for wait group")
	}
}

func MustParseUrl(str string) *url.URL {
	parse, err := url.Parse(str)
	if err != nil {
		panic(err)
	}
	return parse
}
