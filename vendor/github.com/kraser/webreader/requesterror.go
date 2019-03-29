// requesterror
package webreader

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type RequestError struct {
	errTime       time.Time
	errCode       int
	message       string
	url           string
	originalError error
}

func (e RequestError) Error() string {
	return fmt.Sprintf("%v: %v", e.errTime, e.message)
}

func NewRequestError(pResponse *http.Response, url string) error {
	return RequestError{
		errTime:       time.Now(),
		errCode:       pResponse.StatusCode,
		message:       pResponse.Status,
		url:           url,
		originalError: errors.New(strings.Join([]string{pResponse.Status, url}, " AT ")),
	}
}

func q() {
	q := errors.New("ggg")
	fmt.Println(q)
}
