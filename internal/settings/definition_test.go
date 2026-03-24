package settings_test

import (
	"testing"
	"time"

	"github.com/capcom6/lucky-pick-tg-bot/internal/settings"
)

func TestParseDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		want    time.Duration
		wantErr bool
	}{
		{
			name:  "hh mm ss format",
			value: "06:30:15",
			want:  6*time.Hour + 30*time.Minute + 15*time.Second,
		},
		{
			name:  "legacy integer hours format",
			value: "6",
			want:  6 * time.Hour,
		},
		{
			name:  "legacy integer hours format with spaces",
			value: " 12 ",
			want:  12 * time.Hour,
		},
		{
			name:    "invalid text",
			value:   "hello",
			wantErr: true,
		},
		{
			name:    "invalid legacy negative hours",
			value:   "-1",
			wantErr: true,
		},
		{
			name:  "zero legacy hours",
			value: "0",
			want:  0,
		},
		{
			name:  "zero HH:MM:SS",
			value: "00:00:00",
			want:  0,
		},
		{
			name:    "invalid minutes over 59",
			value:   "06:60:00",
			wantErr: true,
		},
		{
			name:    "empty string",
			value:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := settings.ParseDuration(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseDuration(%q) expected error, got nil", tt.value)
				}

				return
			}

			if err != nil {
				t.Fatalf("ParseDuration(%q) unexpected error: %v", tt.value, err)
			}

			if got.Duration != tt.want {
				t.Fatalf("ParseDuration(%q) = %v, want %v", tt.value, got.Duration, tt.want)
			}
		})
	}
}
