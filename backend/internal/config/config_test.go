package config

import "testing"

func TestWithDefaults(t *testing.T) {
	tests := []struct {
		name string
		in   Config
		want string
	}{
		{
			name: "empty listen addr uses default",
			in:   Config{},
			want: "127.0.0.1:8080",
		},
		{
			name: "explicit listen addr is preserved",
			in:   Config{ListenAddr: "127.0.0.1:9090"},
			want: "127.0.0.1:9090",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.WithDefaults()
			if got.ListenAddr != tt.want {
				t.Fatalf("ListenAddr = %q, want %q", got.ListenAddr, tt.want)
			}
		})
	}
}
