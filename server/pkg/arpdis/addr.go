package arpdis

import (
	"net"
	"sync"
	"time"

	"github.com/cherts/anylink/pkg/utils"
)

const (
	StaleTimeNormal      = time.Minute * 5
	StaleTimeUnreachable = time.Minute * 10

	TypeNormal      = 0
	TypeUnreachable = 1
	TypeStatic      = 2
)

var (
	table   = make(map[string]*Addr, 128)
	tableMu sync.RWMutex
)

type Addr struct {
	IP           net.IP
	HardwareAddr net.HardwareAddr
	disTime      time.Time
	Type         int8
}

func Lookup(ip net.IP, onlyTable bool) *Addr {
	addr := tableLookup(ip.To4())
	if addr != nil || onlyTable {
		return addr
	}

	addr = doLookup(ip.To4())
	Add(addr)
	return addr
}

// Add adds a IP-MAC map to a runtime ARP table.
func tableLookup(ip net.IP) *Addr {
	tableMu.RLock()
	addr := table[ip.To4().String()]
	tableMu.RUnlock()
	if addr == nil {
		return nil
	}

	// Determine aging expiration time
	tSub := utils.NowSec().Sub(addr.disTime)
	switch addr.Type {
	case TypeStatic:
	case TypeNormal:
		if tSub > StaleTimeNormal {
			return nil
		}
	case TypeUnreachable:
		if tSub > StaleTimeUnreachable {
			return nil
		}
	}

	return addr
}

// Add adds a IP-MAC map to a runtime ARP table.
func Add(addr *Addr) {
	if addr == nil {
		return
	}
	if addr.disTime.IsZero() {
		addr.disTime = utils.NowSec()
	}
	ip := addr.IP.To4().String()
	tableMu.Lock()
	defer tableMu.Unlock()
	if a, ok := table[ip]; ok {
		// Static addresses can only be set once
		if a.Type == TypeStatic {
			return
		}
	}
	table[ip] = addr
}

// Delete removes an IP from the runtime ARP table.
func Delete(ip net.IP) {
	tableMu.Lock()
	defer tableMu.Unlock()
	delete(table, ip.To4().String())
}

// List returns the current runtime ARP table.
func List() map[string]*Addr {
	tableMu.RLock()
	defer tableMu.RUnlock()
	return table
}
