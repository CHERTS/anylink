// AnyLink is an enterprise-level remote office VPN software that can support multiple people using it online at the same time.

//go:build !windows
// +build !windows

package main

import (
	"embed"
	"os"
	"os/signal"
	"syscall"

	"github.com/cherts/anylink/admin"
	"github.com/cherts/anylink/base"
	"github.com/cherts/anylink/handler"
)

//go:embed ui
var uiData embed.FS

// Program version
var CommitId string

func main() {
	base.CommitId = CommitId
	admin.UiData = uiData

	base.Start()
	handler.Start()
	signalWatch()
}

func signalWatch() {
	base.Info("Server pid:", os.Getpid())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGALRM)
	for {
		sig := <-sigs
		base.Info("Get signal:", sig)
		switch sig {
		case syscall.SIGUSR2:
			// reload
			base.Info("Reload")
		default:
			// stop
			base.Info("Stop")
			handler.Stop()
			return
		}
	}
}
