package browseropener

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Opener opens a URL in the user's default browser.
type Opener struct{}

// New returns a cross-platform browser opener.
func New() Opener { return Opener{} }

// Open starts the platform browser command asynchronously.
func (Opener) Open(url string) error {
	cmdName, args, err := commandForGOOS(runtime.GOOS, url)
	if err != nil {
		return err
	}
	return exec.Command(cmdName, args...).Start()
}

func commandForGOOS(goos, url string) (string, []string, error) {
	switch goos {
	case "windows":
		return "cmd", []string{"/c", "start", "", url}, nil
	case "darwin":
		return "open", []string{url}, nil
	case "linux":
		return "xdg-open", []string{url}, nil
	default:
		return "", nil, fmt.Errorf("unsupported platform %q", goos)
	}
}
