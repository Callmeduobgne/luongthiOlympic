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

package unit

import (
	"testing"

	"github.com/ibn-network/api-gateway/internal/models"
)

func TestTeaBatchStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status models.TeaBatchStatus
		want   bool
	}{
		{"Valid CREATED", models.StatusCreated, true},
		{"Valid VERIFIED", models.StatusVerified, true},
		{"Valid EXPIRED", models.StatusExpired, true},
		{"Invalid status", models.TeaBatchStatus("INVALID"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("TeaBatchStatus.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTeaBatchStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status models.TeaBatchStatus
		want   string
	}{
		{"CREATED status", models.StatusCreated, "CREATED"},
		{"VERIFIED status", models.StatusVerified, "VERIFIED"},
		{"EXPIRED status", models.StatusExpired, "EXPIRED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("TeaBatchStatus.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

