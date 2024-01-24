package handler

import (
	"crypto/sha1"
	"encoding/hex"
	"os"

	"github.com/cherts/anylink/admin"
	"github.com/cherts/anylink/base"
	"github.com/cherts/anylink/cron"
	"github.com/cherts/anylink/dbdata"
	"github.com/cherts/anylink/sessdata"
)

func Start() {
	dbdata.Start()
	sessdata.Start()
	cron.Start()

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
