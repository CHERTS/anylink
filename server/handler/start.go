package handler

import (
	"crypto/sha1"
	"encoding/hex"
	"log"
	"os"

	"github.com/cherts/anylink/admin"
	"github.com/cherts/anylink/base"
	"github.com/cherts/anylink/cron"
	"github.com/cherts/anylink/dbdata"
	"github.com/cherts/anylink/sessdata"
	gosysctl "github.com/lorenzosaino/go-sysctl"
)

func Start() {
	dbdata.Start()
	sessdata.Start()
	cron.Start()

	admin.InitLockManager() // Initialize the anti-explosion timer and IP whitelist

	// Enable server forwarding
	err := gosysctl.Set("net.ipv4.ip_forward", "1")
	if err != nil {
		base.Warn(err)
	}

	val, err := gosysctl.Get("net.ipv4.ip_forward")
	if val != "1" {
		log.Fatal("Please exec 'sysctl -w net.ipv4.ip_forward=1' ")
	}
	// os.Exit(0)
	// execCmd([]string{"sysctl -w net.ipv4.ip_forward=1"})

	switch base.Cfg.LinkMode {
	case base.LinkModeTUN:
		checkTun()
	case base.LinkModeTAP:
		checkTap()
	case base.LinkModeMacvtap:
		checkMacvtap()
	default:
		base.Fatal("LinkMode is err")
	}

	// Calculate the hash of profile.xml
	b, err := os.ReadFile(base.Cfg.Profile)
	if err != nil {
		panic(err)
	}
	ha := sha1.Sum(b)
	profileHash = hex.EncodeToString(ha[:])

	go admin.StartAdmin()
	go startTls()
	go startDtls()

	go logAuditBatch()
}

func Stop() {
	_ = dbdata.Stop()
	destroyVtap()
}
