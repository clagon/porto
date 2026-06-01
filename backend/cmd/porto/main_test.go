package main

import "testing"

func TestParseArgsDefaults(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "defaults",
			args: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := parseArgs(tt.args)
			if err != nil {
				t.Fatalf("parseArgs() error = %v", err)
			}
			if opts.ListenAddr != "" {
				t.Fatalf("ListenAddr = %q, want empty", opts.ListenAddr)
			}
			if opts.ConfigPath != "" {
				t.Fatalf("ConfigPath = %q, want empty", opts.ConfigPath)
			}
			if !opts.OpenBrowser {
				t.Fatal("OpenBrowser = false, want true")
			}
			if opts.Command != commandServe {
				t.Fatalf("Command = %q, want %q", opts.Command, commandServe)
			}
		})
	}
}

func TestParseArgsOverrides(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "overrides",
			args: []string{"--listen-addr", "127.0.0.1:9090", "--config", "/tmp/port-mapper.json", "--no-browser"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := parseArgs(tt.args)
			if err != nil {
				t.Fatalf("parseArgs() error = %v", err)
			}
			if opts.ListenAddr != "127.0.0.1:9090" {
				t.Fatalf("ListenAddr = %q, want %q", opts.ListenAddr, "127.0.0.1:9090")
			}
			if opts.ConfigPath != "/tmp/port-mapper.json" {
				t.Fatalf("ConfigPath = %q, want %q", opts.ConfigPath, "/tmp/port-mapper.json")
			}
			if opts.OpenBrowser {
				t.Fatal("OpenBrowser = true, want false")
			}
			if opts.Command != commandServe {
				t.Fatalf("Command = %q, want %q", opts.Command, commandServe)
			}
		})
	}
}

func TestParseArgsOpenCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want cliOptions
	}{
		{
			name: "minimal open",
			args: []string{"open", "25565"},
			want: cliOptions{
				OpenBrowser:          false,
				Command:              commandOpen,
				Protocol:             "TCP",
				ExternalPort:         25565,
				InternalPort:         25565,
				Description:          "Porto CLI",
				LeaseDurationSeconds: 0,
			},
		},
		{
			name: "udp open with options",
			args: []string{"--config", "/tmp/porto.json", "open", "--protocol", "udp", "--internal-ip", "192.168.1.25", "--internal-port", "19132", "--description", "Bedrock", "--lease", "3600", "19132"},
			want: cliOptions{
				ConfigPath:           "/tmp/porto.json",
				OpenBrowser:          false,
				Command:              commandOpen,
				Protocol:             "udp",
				ExternalPort:         19132,
				InternalIP:           "192.168.1.25",
				InternalPort:         19132,
				Description:          "Bedrock",
				LeaseDurationSeconds: 3600,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseArgs(tt.args)
			if err != nil {
				t.Fatalf("parseArgs() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("parseArgs() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseArgsCloseCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want cliOptions
	}{
		{
			name: "minimal close",
			args: []string{"close", "25565"},
			want: cliOptions{
				OpenBrowser:  false,
				Command:      commandClose,
				Protocol:     "TCP",
				ExternalPort: 25565,
			},
		},
		{
			name: "udp close",
			args: []string{"close", "--protocol", "udp", "19132"},
			want: cliOptions{
				OpenBrowser:  false,
				Command:      commandClose,
				Protocol:     "udp",
				ExternalPort: 19132,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseArgs(tt.args)
			if err != nil {
				t.Fatalf("parseArgs() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("parseArgs() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseArgsCommandErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "unknown command", args: []string{"wat"}},
		{name: "open missing port", args: []string{"open"}},
		{name: "open invalid port", args: []string{"open", "0"}},
		{name: "open invalid protocol", args: []string{"open", "--protocol", "icmp", "25565"}},
		{name: "open invalid internal port", args: []string{"open", "--internal-port", "99999", "25565"}},
		{name: "open invalid lease", args: []string{"open", "--lease", "-1", "25565"}},
		{name: "close invalid protocol", args: []string{"close", "--protocol", "icmp", "25565"}},
		{name: "close extra args", args: []string{"close", "25565", "extra"}},
		{name: "status extra args", args: []string{"status", "extra"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := parseArgs(tt.args); err == nil {
				t.Fatal("parseArgs() error = nil, want error")
			}
		})
	}
}

func TestParseArgsRejectsUnknownFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "unknown flag",
			args: []string{"--does-not-exist"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := parseArgs(tt.args); err == nil {
				t.Fatal("parseArgs() error = nil, want error")
			}
		})
	}
}
