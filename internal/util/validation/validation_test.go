package validation

import (
	"testing"

	"github.com/TiPSYDiPSY/home-task/internal/model/api"
)

func TestDecimal2Validator(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		amount  string
		wantErr bool
	}{
		{"valid - no decimal", "100", false},
		{"valid - one decimal", "100.5", false},
		{"valid - two decimals", "100.50", false},
		{"invalid - three decimals", "100.501", true},
		{"invalid - four decimals", "0.0401", true},
		{"valid - zero with decimals", "0.00", false},
		{"valid - negative", "-10.50", false},
		{"invalid - negative with three decimals", "-10.505", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := api.TransactionRequest{
				State:         "win",
				Amount:        tt.amount,
				TransactionID: "test-123",
			}

			err := validator.ValidateStruct(req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
