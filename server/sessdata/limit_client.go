package sessdata

import (
	"sync"

	"github.com/cherts/anylink/base"
)

const limitAllKey = "__ALL__"

var (
	limitClient = map[string]int{limitAllKey: 0}
	limitMux    = sync.Mutex{}
)

func LimitClient(user string, close bool) bool {
	limitMux.Lock()
	defer limitMux.Unlock()

	_all := limitClient[limitAllKey]
	c, ok := limitClient[user]
	if !ok { // User does not exist
		limitClient[user] = 0
	}

	if close {
		limitClient[user] = c - 1
		limitClient[limitAllKey] = _all - 1
		return true
	}

	// global judgment
	if _all >= base.Cfg.MaxClient {
		return false
	}

	// Same user limit exceeded
	if c >= base.Cfg.MaxUserClient {
		return false
	}

	limitClient[user] = c + 1
	limitClient[limitAllKey] = _all + 1
	return true
}
