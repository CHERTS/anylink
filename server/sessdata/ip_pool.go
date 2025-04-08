package sessdata

import (
	"net"
	"sync"
	"time"

	"github.com/cherts/anylink/base"
	"github.com/cherts/anylink/dbdata"
	"github.com/cherts/anylink/pkg/utils"
)

var (
	IpPool   = &ipPoolConfig{}
	ipActive = map[string]bool{}
	// ipKeep and ipLease  ipAddr => macAddr
	// ipKeep    = map[string]string{}
	ipPoolMux sync.Mutex
	// Record loop points
	loopCurIp uint32
)

type ipPoolConfig struct {
	// Calculate dynamic IP
	Ipv4Gateway net.IP
	Ipv4Mask    net.IP
	Ipv4IPNet   *net.IPNet
	IpLongMin   uint32
	IpLongMax   uint32
}

func initIpPool() {

	// Address processing
	_, ipNet, err := net.ParseCIDR(base.Cfg.Ipv4CIDR)
	if err != nil {
		panic(err)
	}
	IpPool.Ipv4IPNet = ipNet
	IpPool.Ipv4Mask = net.IP(ipNet.Mask)

	ipv4Gateway := net.ParseIP(base.Cfg.Ipv4Gateway)
	ipStart := net.ParseIP(base.Cfg.Ipv4Start)
	ipEnd := net.ParseIP(base.Cfg.Ipv4End)
	if !ipNet.Contains(ipv4Gateway) || !ipNet.Contains(ipStart) || !ipNet.Contains(ipEnd) {
		panic("IP segment setting error")
	}
	// IP address pool
	IpPool.Ipv4Gateway = ipv4Gateway
	IpPool.IpLongMin = utils.Ip2long(ipStart)
	IpPool.IpLongMax = utils.Ip2long(ipEnd)

	loopCurIp = IpPool.IpLongMin

	// Network address zero value
	// zero := binary.BigEndian.Uint32(ip.Mask(mask))
	// broadcast address
	// one, _ := ipNet.Mask.Size()
	// max := min | uint32(math.Pow(2, float64(32-one))-1)

	// Get IpLease data
	// go cronIpLease()
}

// func cronIpLease() {
// 	getIpLease()
// 	tick := time.NewTicker(time.Minute * 30)
// 	for range tick.C {
// 		getIpLease()
// 	}
// }
//
// func getIpLease() {
// 	xdb := dbdata.GetXdb()
// 	keepIpMaps := []dbdata.IpMap{}
// 	// sNow := time.Now().Add(-1 * time.Duration(base.Cfg.IpLease) * time.Second)
// 	err := xdb.Cols("ip_addr", "mac_addr").Where("keep=?", true).Find(&keepIpMaps)
// 	if err != nil {
// 		base.Error(err)
// 	}
// 	log.Println(keepIpMaps)
// 	ipPoolMux.Lock()
// 	ipKeep = map[string]string{}
// 	for _, v := range keepIpMaps {
// 		ipKeep[v.IpAddr] = v.MacAddr
// 	}
// 	ipPoolMux.Unlock()
// }

func ipInPool(ip net.IP) bool {
	if utils.Ip2long(ip) >= IpPool.IpLongMin && utils.Ip2long(ip) <= IpPool.IpLongMax {
		return true
	}
	return false
}

// AcquireIp Get dynamic ip
func AcquireIp(username, macAddr string, uniqueMac bool) (newIp net.IP) {
	base.Trace("AcquireIp start:", username, macAddr, uniqueMac)
	ipPoolMux.Lock()
	defer func() {
		ipPoolMux.Unlock()
		base.Trace("AcquireIp end:", username, macAddr, uniqueMac, newIp)
		base.Info("AcquireIp ip:", username, macAddr, uniqueMac, newIp)
	}()

	var (
		err  error
		tNow = time.Now()
	)

	// Obtaining the client macAddr
	if uniqueMac {
		// Determine whether it has been allocated
		mi := &dbdata.IpMap{}
		err = dbdata.One("mac_addr", macAddr, mi)
		if err != nil {
			// no data has been found
			if dbdata.CheckErrNotFound(err) {
				return loopIp(username, macAddr, uniqueMac)
			}
			// Query error report
			base.Error(err)
			return nil
		}

		// IP record exists
		base.Trace("uniqueMac:", username, mi)
		ipStr := mi.IpAddr
		ip := net.ParseIP(ipStr)
		// Skip active connections
		_, ok := ipActive[ipStr]
		// Check whether the original IP is in the new IP pool
		// IpPool.Ipv4IPNet.Contains(ip) &&
		// IP complies with the specification
		// Check whether the original IP is in the new IP pool
		if !ok && ipInPool(ip) {
			mi.Username = username
			mi.LastLogin = tNow
			mi.UniqueMac = uniqueMac
			// Write back db data
			_ = dbdata.Set(mi)
			ipActive[ipStr] = true
			return ip
		}

		// IP reservation
		if mi.Keep {
			base.Error(username, macAddr, ipStr, "Reserved ip does not match CIDR")
			return nil
		}

		// Delete current macAddr
		mi = &dbdata.IpMap{MacAddr: macAddr}
		_ = dbdata.Del(mi)
		return loopIp(username, macAddr, uniqueMac)
	}

	// If you don't have a Mac
	ipMaps := []dbdata.IpMap{}
	err = dbdata.FindWhere(&ipMaps, 30, 1, "username=?", username)
	if err != nil {
		// No data found
		if dbdata.CheckErrNotFound(err) {
			return loopIp(username, macAddr, uniqueMac)
		}
		// Query error
		base.Error(err)
		return nil
	}

	// Traverse mac records
	for _, mi := range ipMaps {
		ipStr := mi.IpAddr
		ip := net.ParseIP(ipStr)

		// Skip active connections
		if _, ok := ipActive[ipStr]; ok {
			continue
		}
		// Skip reserved ip
		if mi.Keep {
			continue
		}
		if mi.UniqueMac {
			continue
		}

		// No need to verify the lease period if you don't have a mac
		// mi.LastLogin.Before(leaseTime) &&
		if ipInPool(ip) {
			mi.Username = username
			mi.LastLogin = tNow
			mi.MacAddr = macAddr
			mi.UniqueMac = uniqueMac
			//Write back db data
			_ = dbdata.Set(mi)
			ipActive[ipStr] = true
			return ip
		}
	}

	return loopIp(username, macAddr, uniqueMac)
}

var (
	// Record loop points
	loopCurIp uint32
	loopFarIp *dbdata.IpMap
)

func loopIp(username, macAddr string, uniqueMac bool) net.IP {
	var (
		i  uint32
		ip net.IP
	)

	// Reassignment
	loopFarIp = &dbdata.IpMap{LastLogin: time.Now()}

	i, ip = loopLong(loopCurIp, IpPool.IpLongMax, username, macAddr, uniqueMac)
	if ip != nil {
		loopCurIp = i
		return ip
	}

	i, ip = loopLong(IpPool.IpLongMin, loopCurIp, username, macAddr, uniqueMac)
	if ip != nil {
		loopCurIp = i
		return ip
	}

	// After IP allocation, start from the beginning
	loopCurIp = IpPool.IpLongMin

	if loopFarIp.Id > 0 {
		// Use the earliest logged in IP
		ipStr := loopFarIp.IpAddr
		ip = net.ParseIP(ipStr)
		mi := &dbdata.IpMap{IpAddr: ipStr, MacAddr: macAddr, UniqueMac: uniqueMac, Username: username, LastLogin: time.Now()}
		// Write back db data
		_ = dbdata.Set(mi)
		ipActive[ipStr] = true

		return ip
	}

	// All online, no data available
	base.Warn("no ip available, please see ip_map table row", username, macAddr)
	return nil
}

func loopLong(start, end uint32, username, macAddr string, uniqueMac bool) (uint32, net.IP) {
	var (
		err       error
		tNow      = time.Now()
		leaseTime = time.Now().Add(-1 * time.Duration(base.Cfg.IpLease) * time.Second)
	)

	// Global traversal of expired and unreserved IPs
	for i := start; i <= end; i++ {
		ip := utils.Long2ip(i)
		ipStr := ip.String()

		// Skip active connections
		if _, ok := ipActive[ipStr]; ok {
			continue
		}

		mi := &dbdata.IpMap{}
		err = dbdata.One("ip_addr", ipStr, mi)
		if err != nil {
			// no data has been found
			if dbdata.CheckErrNotFound(err) {
				// This ip is not in use
				mi = &dbdata.IpMap{IpAddr: ipStr, MacAddr: macAddr, UniqueMac: uniqueMac, Username: username, LastLogin: tNow}
				_ = dbdata.Add(mi)
				ipActive[ipStr] = true
				return i, ip
			}
			// Query error report
			base.Error(err)
			return 0, nil
		}

		// Query the used IP
		// Skip reserved ip
		if mi.Keep {
			continue
		}
		// Determine the lease term
		if mi.LastLogin.Before(leaseTime) {
			// There is a record indicating that the lease period has expired and can be used directly.
			mi.Username = username
			mi.LastLogin = tNow
			mi.MacAddr = macAddr
			mi.UniqueMac = uniqueMac
			// Write back db data
			_ = dbdata.Set(mi)
			ipActive[ipStr] = true
			return i, ip
		}
		// Other situations determine the earliest landing
		if mi.LastLogin.Before(loopFarIp.LastLogin) {
			loopFarIp = mi
		}
	}

	return 0, nil
}

// Recycle ip
func ReleaseIp(ip net.IP, macAddr string) {
	ipPoolMux.Lock()
	defer ipPoolMux.Unlock()

	delete(ipActive, ip.String())

	mi := &dbdata.IpMap{}
	err := dbdata.One("ip_addr", ip.String(), mi)
	if err == nil {
		mi.LastLogin = time.Now()
		_ = dbdata.Set(mi)
	}
}
