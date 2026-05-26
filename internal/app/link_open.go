package app

import (
	"fmt"
	"os/exec"
	"runtime"
)

type LinkOpener interface {
	Open(target string) error
}

type systemLinkOpener struct{}

func NewSystemLinkOpener() LinkOpener {
	return systemLinkOpener{}
}

func (systemLinkOpener) Open(target string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", target)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", target)
	default:
		cmd = exec.Command("xdg-open", target)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open %q: %w", target, err)
	}
	return nil
}
