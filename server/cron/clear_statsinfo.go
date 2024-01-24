package cron

import (
	"time"

	"github.com/cherts/anylink/base"
	"github.com/cherts/anylink/dbdata"
)

const siLifeDay = 30

// Clear chart data
func ClearStatsInfo() {
	_, timesUp := isClearTime()
	if !timesUp {
		return
	}
	ts := getTimeAgo(siLifeDay)
	for _, item := range dbdata.StatsInfoIns.Actions {
		affected, err := dbdata.StatsInfoIns.ClearStatsInfo(item, ts)
		base.Info("Cron ClearStatsInfo  "+item+": ", affected, err)
	}
}

// Is it time to "clean up"?
func isClearTime() (int, bool) {
	dataLog, err := dbdata.SettingGetAuditLog()
	if err != nil {
		base.Error("Cron SettingGetLog: ", err)
		return -1, false
	}
	currentTime := time.Now().Format("15:04")
	// When the "cleaning time" has not arrived, return
	if dataLog.ClearTime != currentTime {
		return -1, false
	}
	return dataLog.LifeDay, true
}

// Get the cleanup date based on the storage duration
func getTimeAgo(days int) string {
	var timeS string
	ts := time.Now().AddDate(0, 0, -days)
	tsZero := time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, time.Local)
	timeS = tsZero.Format(dbdata.LayoutTimeFormat)
	return timeS
}
