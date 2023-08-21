//go:build unit
// +build unit

package app

import "testing"

func TestCageStatusValidate(t *testing.T) {
	tests := []struct {
		status CageStatus
		valid  bool
	}{
		{CageStatusActive, true},
		{CageStatusDown, true},
		{CageStatusUnspecified, false},
		{CageStatus("foo"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			err := tt.status.Validate()
			if want, got := tt.valid, err == nil; want != got {
				t.Errorf("Expected %t got %t", want, got)
			}
		})
	}
}

func TestCageStatusIsUnspecified(t *testing.T) {
	tests := []struct {
		status CageStatus
		empty  bool
	}{
		{CageStatusActive, false},
		{CageStatusDown, false},
		{CageStatusUnspecified, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if want, got := tt.empty, tt.status.IsUnspecified(); want != got {
				t.Errorf("Expected %t got %t", want, got)
			}
		})
	}
}
