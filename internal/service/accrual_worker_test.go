package service

import (
	"testing"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/model"
)

func TestMapAccrualStatus(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "registered becomes processing",
			in:   "REGISTERED",
			want: model.OrderStatusProcessing,
		},
		{
			name: "processing stays processing",
			in:   model.OrderStatusProcessing,
			want: model.OrderStatusProcessing,
		},
		{
			name: "invalid stays invalid",
			in:   model.OrderStatusInvalid,
			want: model.OrderStatusInvalid,
		},
		{
			name: "processed stays processed",
			in:   model.OrderStatusProcessed,
			want: model.OrderStatusProcessed,
		},
		{
			name: "unknown status is ignored",
			in:   "UNKNOWN",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapAccrualStatus(tt.in); got != tt.want {
				t.Fatalf("mapAccrualStatus(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
