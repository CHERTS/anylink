package sessdata

func Start() {
	initIpPool()
	checkSession()
	saveStatsInfo()
	CloseUserLimittimeSession()
}
