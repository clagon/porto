package app

import (
	"fmt"
	"os/exec"
	"runtime"
)

// defaultBrowserOpener は、OSに対応したブラウザ起動オブジェクトを返します。
func defaultBrowserOpener() BrowserOpener {
	return osBrowserOpener{}
}

// osBrowserOpener は、OS ごとの標準ブラウザ起動を実装します。
type osBrowserOpener struct{}

// Open は、実行中のプラットフォームに対応したOSコマンドを非同期的に実行し、Webブラウザで指定URLを開きます。
func (osBrowserOpener) Open(url string) error {
	cmdName, args, err := browserCommandForGOOS(runtime.GOOS, url)
	if err != nil {
		return err
	}
	return exec.Command(cmdName, args...).Start()
}

func browserCommandForGOOS(goos, url string) (string, []string, error) {
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
