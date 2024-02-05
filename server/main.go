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
var (
	appVer    string
	commitId  string
	buildDate string
)

func main() {
	admin.UiData = uiData
	base.APP_VER = appVer
	base.CommitId = commitId
	base.BuildDate = buildDate

	base.Start()
	handler.Start()

	signalWatch()
}

func signalWatch() {
	base.Info("Server pid: ", os.Getpid())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGALRM, syscall.SIGUSR2)
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
