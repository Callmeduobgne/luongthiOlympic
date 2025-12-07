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

