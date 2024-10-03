package utils

import (
	"fmt"

	"github.com/0xSumeet/go_api/internal/configs"
)

// Validate page number for request
func ValidatePageNumber(page int) (bool, error) {
	if page < 1 {
		return false, fmt.Errorf("Invalid Page number")
	}
	return true, nil
}

// Validate data limit for pagination
func ValidateDataLimit(limit int) int {
	if limit <= 0 {
		return config.DefaultLimit
	} else if limit > config.MaximumLimit {
		return config.MaximumLimit
	}
	return limit
}

