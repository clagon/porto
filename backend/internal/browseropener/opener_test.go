package browseropener

import "testing"

func TestCommandForGOOS(t *testing.T) {
	tests := []struct {
		name    string
		goos    string
		url     string
		wantCmd string
		wantArg []string
	}{
		{
			name:    "linux uses xdg-open",
			goos:    "linux",
			url:     "http://127.0.0.1:8080/",
			wantCmd: "xdg-open",
			wantArg: []string{"http://127.0.0.1:8080/"},
		},
		{
			name:    "darwin uses open",
			goos:    "darwin",
			url:     "http://127.0.0.1:8080/",
			wantCmd: "open",
			wantArg: []string{"http://127.0.0.1:8080/"},
		},
		{
			name:    "windows uses cmd start",
			goos:    "windows",
			url:     "http://127.0.0.1:8080/",
			wantCmd: "cmd",
			wantArg: []string{"/c", "start", "", "http://127.0.0.1:8080/"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, gotArgs, err := commandForGOOS(tt.goos, tt.url)
			if err != nil {
				t.Fatalf("commandForGOOS() error = %v", err)
			}
			if gotCmd != tt.wantCmd {
				t.Fatalf("commandForGOOS() cmd = %q, want %q", gotCmd, tt.wantCmd)
			}
			if len(gotArgs) != len(tt.wantArg) {
				t.Fatalf("commandForGOOS() args len = %d, want %d", len(gotArgs), len(tt.wantArg))
			}
			for i := range gotArgs {
				if gotArgs[i] != tt.wantArg[i] {
					t.Fatalf("commandForGOOS() args[%d] = %q, want %q", i, gotArgs[i], tt.wantArg[i])
				}
			}
		})
	}
}

func TestCommandForGOOSRejectsUnknownPlatform(t *testing.T) {
	if _, _, err := commandForGOOS("plan9", "http://127.0.0.1:8080/"); err == nil {
		t.Fatal("commandForGOOS() error = nil, want error")
	}
}
