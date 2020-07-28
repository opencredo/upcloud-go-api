package upcloud

import (
	"encoding/json"
	"fmt"
)

// Error represents an error
type Error struct {
	ErrorCode    string `xml:"error_code" json:"error_code"`
	ErrorMessage string `xml:"error_message" json:"error_message"`
}

func (e *Error) UnmarshalJSON(b []byte) error {
	var v map[string]map[string]string
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	e.ErrorCode = v["error"]["error_code"]
	e.ErrorMessage = v["error"]["error_message"]

	return nil
}

// Error implements the Error interface
func (e *Error) Error() string {
	return fmt.Sprintf("%s (%s)", e.ErrorMessage, e.ErrorCode)
}
