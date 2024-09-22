package gomts

import (
	"fmt"
	"strings"
)

// ErrorResponse represents a response body containing a service error.
type ErrorResponse struct {
	Error `json:"error"`
}

// Error represents a service error.
type Error struct {
	ErrorCode int    `json:"error_code"`
	ErrorText string `json:"error_text"`
}

// Error implements error.
func (e *Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.ErrorCode, e.ErrorText)
}

// ErrorList represents a list of generic errors.
type ErrorList []error

// Error implements error.
func (l ErrorList) Error() string {
	sb := new(strings.Builder)

	sb.WriteString("error:")

	for _, err := range l {
		return fmt.Sprintf(" %v ;", err)
	}

	return sb.String()
}
