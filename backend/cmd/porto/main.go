// Package main は Porto アプリケーションの起動用エントリポイントを提供します。
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/clagon/port-mapper/backend/internal/app"
	"github.com/clagon/port-mapper/backend/internal/domain"
	"github.com/clagon/port-mapper/backend/internal/service"
)

type cliCommand string

const (
	commandServe  cliCommand = "serve"
	commandOpen   cliCommand = "open"
	commandClose  cliCommand = "close"
	commandStatus cliCommand = "status"
)

// cliOptions はコマンドライン引数から解析された起動オプションを表します。
type cliOptions struct {
	ListenAddr           string
	ConfigPath           string
	OpenBrowser          bool
	Command              cliCommand
	Protocol             string
	ExternalPort         int
	InternalIP           string
	InternalPort         int
	Description          string
	LeaseDurationSeconds int
}

// parseArgs はコマンドライン引数を解析し起動オプションにマッピングします。
func parseArgs(args []string) (cliOptions, error) {
	fs := flag.NewFlagSet("port-mapper", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	opts := cliOptions{OpenBrowser: true}
	fs.StringVar(&opts.ListenAddr, "listen-addr", "", "listen address for the local API")
	fs.StringVar(&opts.ConfigPath, "config", "", "path to config.json")
	noBrowser := fs.Bool("no-browser", false, "do not open the browser automatically")

	if err := fs.Parse(args); err != nil {
		return cliOptions{}, err
	}
	if *noBrowser {
		opts.OpenBrowser = false
	}
	return parseCommand(opts, fs.Args())
}

func parseCommand(opts cliOptions, args []string) (cliOptions, error) {
	if len(args) == 0 {
		opts.Command = commandServe
		return opts, nil
	}

	switch args[0] {
	case string(commandOpen):
		return parseOpenCommand(opts, args[1:])
	case string(commandClose):
		return parseCloseCommand(opts, args[1:])
	case string(commandStatus):
		opts.Command = commandStatus
		opts.OpenBrowser = false
		if len(args) > 1 {
			return cliOptions{}, fmt.Errorf("status accepts no arguments")
		}
		return opts, nil
	case string(commandServe):
		opts.Command = commandServe
		if len(args) > 1 {
			return cliOptions{}, fmt.Errorf("serve accepts no arguments")
		}
		return opts, nil
	default:
		return cliOptions{}, fmt.Errorf("unknown command %q", args[0])
	}
}

func parseOpenCommand(opts cliOptions, args []string) (cliOptions, error) {
	fs := flag.NewFlagSet("porto open", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	opts.Command = commandOpen
	opts.OpenBrowser = false
	opts.Protocol = "TCP"
	opts.Description = "Porto CLI"
	fs.StringVar(&opts.Protocol, "protocol", opts.Protocol, "protocol to map: TCP or UDP")
	fs.StringVar(&opts.InternalIP, "internal-ip", "", "internal IP address; auto-detected when omitted")
	fs.IntVar(&opts.InternalPort, "internal-port", 0, "internal port; defaults to the external port")
	fs.StringVar(&opts.Description, "description", opts.Description, "port mapping description")
	fs.IntVar(&opts.LeaseDurationSeconds, "lease", 0, "lease duration in seconds; 0 means permanent")
	if err := fs.Parse(args); err != nil {
		return cliOptions{}, err
	}
	rest := fs.Args()
	if len(rest) != 1 {
		return cliOptions{}, fmt.Errorf("open requires exactly one external port")
	}
	port, err := parsePort(rest[0])
	if err != nil {
		return cliOptions{}, err
	}
	if err := validateCLIProtocol(opts.Protocol); err != nil {
		return cliOptions{}, err
	}
	if opts.InternalPort < 0 || opts.InternalPort > 65535 {
		return cliOptions{}, fmt.Errorf("internal port %d out of range: must be 1-65535", opts.InternalPort)
	}
	if opts.LeaseDurationSeconds < 0 || opts.LeaseDurationSeconds > domain.MaxLeaseDurationSeconds {
		return cliOptions{}, fmt.Errorf("lease duration %d out of range: must be 0-%d", opts.LeaseDurationSeconds, domain.MaxLeaseDurationSeconds)
	}
	opts.ExternalPort = port
	if opts.InternalPort == 0 {
		opts.InternalPort = port
	}
	return opts, nil
}

func parseCloseCommand(opts cliOptions, args []string) (cliOptions, error) {
	fs := flag.NewFlagSet("porto close", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	opts.Command = commandClose
	opts.OpenBrowser = false
	opts.Protocol = "TCP"
	fs.StringVar(&opts.Protocol, "protocol", opts.Protocol, "protocol to close: TCP or UDP")
	if err := fs.Parse(args); err != nil {
		return cliOptions{}, err
	}
	rest := fs.Args()
	if len(rest) != 1 {
		return cliOptions{}, fmt.Errorf("close requires exactly one external port")
	}
	port, err := parsePort(rest[0])
	if err != nil {
		return cliOptions{}, err
	}
	if err := validateCLIProtocol(opts.Protocol); err != nil {
		return cliOptions{}, err
	}
	opts.ExternalPort = port
	return opts, nil
}

func parsePort(raw string) (int, error) {
	port, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid port %q: %w", raw, err)
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port %d out of range: must be 1-65535", port)
	}
	return port, nil
}

func validateCLIProtocol(protocol string) error {
	switch strings.ToUpper(strings.TrimSpace(protocol)) {
	case "TCP", "UDP":
		return nil
	default:
		return fmt.Errorf("invalid protocol %q: must be TCP or UDP", protocol)
	}
}

func (opts cliOptions) portMapping() domain.PortMapping {
	return domain.PortMapping{
		Protocol:             strings.ToUpper(strings.TrimSpace(opts.Protocol)),
		ExternalPort:         opts.ExternalPort,
		InternalIP:           strings.TrimSpace(opts.InternalIP),
		InternalPort:         opts.InternalPort,
		Description:          opts.Description,
		LeaseDurationSeconds: opts.LeaseDurationSeconds,
	}
}

// main はアプリケーションのエントリポイントであり、フラグ解析、DI（依存性の注入）、および初期起動処理を実行します。
func main() {
	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	a, err := app.New(app.AppOptions{
		ListenAddr:  opts.ListenAddr,
		ConfigPath:  opts.ConfigPath,
		OpenBrowser: opts.OpenBrowser,
		Logger:      logger,
	})
	if err != nil {
		log.Fatal(err)
	}

	logger.Info("port-mapper ready",
		"config_path", a.ConfigPath(),
		"listen_addr", a.Addr(),
		"browser_open", opts.OpenBrowser,
	)

	if opts.Command != commandServe {
		if err := runCommand(a, opts); err != nil {
			log.Fatal(err)
		}
		return
	}

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

func runCommand(a *app.App, opts cliOptions) error {
	switch opts.Command {
	case commandOpen:
		if _, err := a.Discover(); err != nil {
			return err
		}
		status, err := a.OpenPort(opts.portMapping())
		if err != nil {
			return err
		}
		fmt.Printf("opened %s %d -> %s:%d\n",
			strings.ToUpper(opts.Protocol),
			opts.ExternalPort,
			localIPForOutput(status, opts.InternalIP),
			opts.InternalPort,
		)
		return nil
	case commandClose:
		if _, err := a.Discover(); err != nil {
			return err
		}
		if _, err := a.ClosePort(opts.portMapping()); err != nil {
			return err
		}
		fmt.Printf("closed %s %d\n", strings.ToUpper(opts.Protocol), opts.ExternalPort)
		return nil
	case commandStatus:
		status, err := a.Discover()
		if err != nil {
			return err
		}
		printStatus(status)
		return nil
	default:
		return fmt.Errorf("unsupported command %q", opts.Command)
	}
}

func localIPForOutput(status service.Status, provided string) string {
	if strings.TrimSpace(provided) != "" {
		return strings.TrimSpace(provided)
	}
	if status.LocalIP != "" {
		return status.LocalIP
	}
	return "auto"
}

func printStatus(status service.Status) {
	if !status.Discovered {
		fmt.Println("router: not discovered")
		return
	}
	fmt.Printf("router: %s\n", status.ControlURL)
	if status.ExternalIP != "" {
		fmt.Printf("external ip: %s\n", status.ExternalIP)
	}
	if status.LocalIP != "" {
		fmt.Printf("local ip: %s\n", status.LocalIP)
	}
	fmt.Printf("ports: %d\n", len(status.Ports))
	for _, mapping := range status.Ports {
		fmt.Printf("- %s %d -> %s:%d %s\n",
			strings.ToUpper(mapping.Protocol),
			mapping.ExternalPort,
			mapping.InternalIP,
			mapping.InternalPort,
			mapping.Description,
		)
	}
}
