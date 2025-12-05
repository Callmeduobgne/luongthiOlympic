// Copyright 2024 IBN Network (ICTU Blockchain Network)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	batchIDPattern   = regexp.MustCompile(`^[A-Za-z0-9_-]{3,50}$`)
	packageIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{3,50}$`)
	datePattern      = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

// ValidateBatchID validates a batch ID
func ValidateBatchID(batchID string) error {
	if strings.TrimSpace(batchID) == "" {
		return fmt.Errorf("batch ID cannot be empty")
	}
	if !batchIDPattern.MatchString(batchID) {
		return fmt.Errorf("batch ID must be 3-50 characters and contain only alphanumeric, underscore, or hyphen")
	}
	return nil
}

// ValidatePackageID validates a package ID
func ValidatePackageID(packageID string) error {
	if strings.TrimSpace(packageID) == "" {
		return fmt.Errorf("package ID cannot be empty")
	}
	if !packageIDPattern.MatchString(packageID) {
		return fmt.Errorf("package ID must be 3-50 characters and contain only alphanumeric, underscore, or hyphen")
	}
	return nil
}

// ValidateDate validates a date string in YYYY-MM-DD format
func ValidateDate(dateStr string) error {
	if strings.TrimSpace(dateStr) == "" {
		return fmt.Errorf("date cannot be empty")
	}
	if !datePattern.MatchString(dateStr) {
		return fmt.Errorf("date must be in YYYY-MM-DD format")
	}
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date: %v", err)
	}
	return nil
}

// ValidateString validates a string with max length
func ValidateString(value, fieldName string, maxLength int) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}
	if len(value) > maxLength {
		return fmt.Errorf("%s must be less than %d characters", fieldName, maxLength)
	}
	return nil
}

// ValidateWeight validates package weight
func ValidateWeight(weight float64) error {
	if weight <= 0 {
		return fmt.Errorf("weight must be greater than 0")
	}
	if weight > 1000000 { // 1 ton max
		return fmt.Errorf("weight must be less than 1,000,000 grams")
	}
	return nil
}

// ValidatePagination validates limit and offset for pagination
func ValidatePagination(limit, offset int) error {
	if limit < 1 || limit > 1000 {
		return fmt.Errorf("limit must be between 1 and 1000")
	}
	if offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}
	return nil
}

// ValidateDateRange validates that startDate is before endDate
func ValidateDateRange(startDate, endDate string) error {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return fmt.Errorf("invalid start date: %v", err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return fmt.Errorf("invalid end date: %v", err)
	}
	if start.After(end) {
		return fmt.Errorf("start date cannot be after end date")
	}
	return nil
}
