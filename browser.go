package fileshare

import (
	"log"
	"os/exec"
	"runtime"
)

func openBrowser(url string, autoOpen bool) {
	if !autoOpen {
		return
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	default:
		log.Println("[WARN] Unsupported platform for browser auto-open")
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("[ERROR] Failed to open browser: %v", err)
	}
}
