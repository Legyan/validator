package validator

import (
	"fmt"
)

func ValidateString(v string, maxLen int) error {
	if len(v) > maxLen || len(v) < 1 {
		return fmt.Errorf("%s length must be between 1 and %d", v, maxLen)
	}
	return nil
}
