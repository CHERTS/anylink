package cron

import (
	"github.com/cherts/anylink/base"
	"github.com/cherts/anylink/dbdata"
)

// Clear user activity log
func ClearUserActLog() {
	lifeDay, timesUp := isClearTime()
	if !timesUp {
		return
	}
	// When the audit log is saved permanently, exit
	if lifeDay <= 0 {
		return
	}
	affected, err := dbdata.UserActLogIns.ClearUserActLog(getTimeAgo(lifeDay))
	base.Info("Cron ClearUserActLog: ", affected, err)
}
