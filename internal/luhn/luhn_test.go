package luhn

import "testing"

func TestValid(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{
			name:   "valid order number",
			number: "9278923470",
			want:   true,
		},
		{
			name:   "valid short number",
			number: "0",
			want:   true,
		},
		{
			name:   "invalid order number",
			number: "1234567890",
			want:   false,
		},
		{
			name:   "empty number",
			number: "",
			want:   false,
		},
		{
			name:   "contains letters",
			number: "123abc",
			want:   false,
		},
		{
			name:   "contains space",
			number: "123 456",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Valid(tt.number)
			if got != tt.want {
				t.Fatalf("Valid(%q) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}