package sessdata

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cherts/anylink/base"
	"github.com/cherts/anylink/dbdata"
	mapset "github.com/deckarep/golang-set"
	atomic2 "go.uber.org/atomic"
)

var (
	// session_token -> SessUser
	sessions = make(map[string]*Session)
	// dtlsId -> session_token
	dtlsIds = make(map[string]string)
	sessMux sync.RWMutex
)

// connect sess
type ConnSession struct {
	Sess                *Session
	MasterSecret        string // master_secret of dtls protocol
	IpAddr              net.IP // assigned ip address
	LocalIp             net.IP
	MacHw               net.HardwareAddr // Client mac address, taken out from Session
	Username            string
	RemoteAddr          string
	Mtu                 int
	IfName              string
	Client              string // client mobile pc
	UserAgent           string // Client information
	UserLogoutCode      uint8  // User/client actively logs out
	CstpDpd             int
	Group               *dbdata.Group
	Limit               *LimitRater
	BandwidthUp         atomic2.Uint32 // Use upstream bandwidth Byte
	BandwidthDown       atomic2.Uint32 // Use downlink bandwidth Byte
	BandwidthUpPeriod   atomic2.Uint32 // The total amount of the previous period
	BandwidthDownPeriod atomic2.Uint32
	BandwidthUpAll      atomic2.Uint64 // Total amount of uplink bandwidth used
	BandwidthDownAll    atomic2.Uint64 // Total downlink bandwidth used
	closeOnce           sync.Once
	CloseChan           chan struct{}
	PayloadIn           chan *Payload
	PayloadOutCstp      chan *Payload // Cstp data
	PayloadOutDtls      chan *Payload // Dtls data
	// dSess *DtlsSession
	dSess *atomic.Value
	// compress
	CstpPickCmp CmpEncoding
	DtlsPickCmp CmpEncoding
}

type DtlsSession struct {
	isActive  int32
	CloseChan chan struct{}
	closeOnce sync.Once
	IpAddr    net.IP
}

type Session struct {
	mux             sync.RWMutex
	Sid             string // session-id returned by auth
	Token           string // The unique token of session information
	DtlsSid         string // session_id of dtls protocol
	MacAddr         string // Client mac address
	UniqueIdGlobal  string // Client unique identifier
	MacHw           net.HardwareAddr
	UniqueMac       bool   // The client obtains the real device mac
	Username        string // username
	Group           string
	AuthStep        string
	AuthPass        string
	RemoteAddr      string
	UserAgent       string
	DeviceType      string
	PlatformVersion string

	LastLogin time.Time
	IsActive  bool

	// Parameters that need to be set to enable link
	CSess *ConnSession
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func checkSession() {
	// Detect expired sessions
	go func() {
		if base.Cfg.SessionTimeout == 0 {
			return
		}
		timeout := time.Duration(base.Cfg.SessionTimeout) * time.Second
		tick := time.NewTicker(time.Second * 60)
		for range tick.C {
			outToken := []string{}
			sessMux.RLock()
			t := time.Now()
			for k, v := range sessions {
				v.mux.RLock()
				if !v.IsActive {
					if t.Sub(v.LastLogin) > timeout {
						outToken = append(outToken, k)
					}
				}
				v.mux.RUnlock()
			}
			sessMux.RUnlock()

			// Delete expired session
			for _, v := range outToken {
				CloseSess(v, dbdata.UserLogoutTimeout)
			}
		}
	}()
}

// Users whose status is expired will be kicked offline
func CloseUserLimittimeSession() {
	s := mapset.NewSetFromSlice(dbdata.CheckUserlimittime())
	limitTimeToken := []string{}
	sessMux.RLock()
	for _, v := range sessions {
		v.mux.RLock()
		if v.IsActive && s.Contains(v.Username) {
			limitTimeToken = append(limitTimeToken, v.Token)
		}
		v.mux.RUnlock()
	}
	sessMux.RUnlock()
	for _, v := range limitTimeToken {
		CloseSess(v, dbdata.UserLogoutExpire)
	}
}

func GenToken() string {
	// Generate 32-bit token
	bToken := make([]byte, 32)
	rand.Read(bToken)
	return fmt.Sprintf("%x", bToken)
}

func NewSession(token string) *Session {
	if token == "" {
		btoken := make([]byte, 32)
		rand.Read(btoken)
		token = fmt.Sprintf("%x", btoken)
	}

	// generate dtlsn session_id
	dtlsid := make([]byte, 32)
	rand.Read(dtlsid)

	sess := &Session{
		Sid:       fmt.Sprintf("%d", time.Now().Unix()),
		Token:     token,
		DtlsSid:   fmt.Sprintf("%x", dtlsid),
		LastLogin: time.Now(),
	}

	sessMux.Lock()
	sessions[token] = sess
	dtlsIds[sess.DtlsSid] = token
	sessMux.Unlock()
	return sess
}

func (s *Session) NewConn() *ConnSession {
	s.mux.RLock()
	active := s.IsActive
	macAddr := s.MacAddr
	macHw := s.MacHw
	username := s.Username
	uniqueMac := s.UniqueMac
	s.mux.RUnlock()
	if active {
		s.CSess.Close()
	}

	limit := LimitClient(username, false)
	if !limit {
		base.Warn("limit is full", username)
		return nil
	}
	ip := AcquireIp(username, macAddr, uniqueMac)
	if ip == nil {
		LimitClient(username, true)
		return nil
	}

	// Query group information
	group := &dbdata.Group{}
	err := dbdata.One("Name", s.Group, group)
	if err != nil {
		base.Error(err)
		return nil
	}

	cSess := &ConnSession{
		Sess:           s,
		MacHw:          macHw,
		Username:       username,
		IpAddr:         ip,
		closeOnce:      sync.Once{},
		CloseChan:      make(chan struct{}),
		PayloadIn:      make(chan *Payload, 64),
		PayloadOutCstp: make(chan *Payload, 64),
		PayloadOutDtls: make(chan *Payload, 64),
		dSess:          &atomic.Value{},
	}

	dSess := &DtlsSession{
		isActive: -1,
	}
	cSess.dSess.Store(dSess)

	cSess.Group = group
	if group.Bandwidth > 0 {
		// Current limiting settings
		cSess.Limit = NewLimitRater(group.Bandwidth, group.Bandwidth)
	}

	go cSess.ratePeriod()

	s.mux.Lock()
	s.MacAddr = macAddr
	s.IsActive = true
	s.CSess = cSess
	s.mux.Unlock()
	return cSess
}

func (cs *ConnSession) Close() {
	cs.closeOnce.Do(func() {
		base.Info("closeOnce:", cs.IpAddr)
		cs.Sess.mux.Lock()
		defer cs.Sess.mux.Unlock()

		close(cs.CloseChan)
		cs.Sess.IsActive = false
		cs.Sess.LastLogin = time.Now()
		cs.Sess.CSess = nil

		dSess := cs.GetDtlsSession()
		if dSess != nil {
			dSess.Close()
		}

		ReleaseIp(cs.IpAddr, cs.Sess.MacAddr)
		LimitClient(cs.Username, true)
		AddUserActLog(cs)
	})
}

// Create dtls link
func (cs *ConnSession) NewDtlsConn() *DtlsSession {
	ds := cs.dSess.Load().(*DtlsSession)
	isActive := atomic.LoadInt32(&ds.isActive)
	if isActive > 0 {
		// Determine that the original connection exists and do not create it.
		return nil
	}

	dSess := &DtlsSession{
		isActive:  1,
		CloseChan: make(chan struct{}),
		closeOnce: sync.Once{},
		IpAddr:    cs.IpAddr,
	}
	cs.dSess.Store(dSess)
	return dSess
}

// Close dtls link
func (ds *DtlsSession) Close() {
	ds.closeOnce.Do(func() {
		base.Info("closeOnce dtls:", ds.IpAddr)

		atomic.StoreInt32(&ds.isActive, -1)
		close(ds.CloseChan)
	})
}

func (cs *ConnSession) GetDtlsSession() *DtlsSession {
	ds := cs.dSess.Load().(*DtlsSession)
	isActive := atomic.LoadInt32(&ds.isActive)
	if isActive > 0 {
		return ds
	}
	return nil
}

const BandwidthPeriodSec = 10 // Traffic rate statistics period (seconds)

func (cs *ConnSession) ratePeriod() {
	tick := time.NewTicker(time.Second * BandwidthPeriodSec)
	defer tick.Stop()

	for range tick.C {
		select {
		case <-cs.CloseChan:
			return
		default:
		}

		// Real-time traffic clearing
		rtUp := cs.BandwidthUp.Swap(0)
		rtDown := cs.BandwidthDown.Swap(0)
		// Set the traffic per second in the previous cycle
		cs.BandwidthUpPeriod.Swap(rtUp / BandwidthPeriodSec)
		cs.BandwidthDownPeriod.Swap(rtDown / BandwidthPeriodSec)
		// Add up all traffic
		cs.BandwidthUpAll.Add(uint64(rtUp))
		cs.BandwidthDownAll.Add(uint64(rtDown))
	}
}

var MaxMtu = 1460

func (cs *ConnSession) SetMtu(mtu string) {
	if base.Cfg.Mtu > 0 {
		MaxMtu = base.Cfg.Mtu
	}
	cs.Mtu = MaxMtu

	mi, err := strconv.Atoi(mtu)
	if err != nil || mi < 100 {
		return
	}

	if mi < MaxMtu {
		cs.Mtu = mi
	}
}

func (cs *ConnSession) SetIfName(name string) {
	cs.Sess.mux.Lock()
	defer cs.Sess.mux.Unlock()
	cs.IfName = name
}

func (cs *ConnSession) RateLimit(byt int, isUp bool) error {
	if isUp {
		cs.BandwidthUp.Add(uint32(byt))
		return nil
	}
	// Only downstream rate limit
	cs.BandwidthDown.Add(uint32(byt))
	if cs.Limit == nil {
		return nil
	}
	return cs.Limit.Wait(byt)
}

func (cs *ConnSession) SetPickCmp(cate, encoding string) (string, bool) {
	var cmpName string
	if !base.Cfg.Compression {
		return cmpName, false
	}
	var cmp CmpEncoding
	switch {
	// case strings.Contains(encoding, "oc-lz4"):
	// 	cmpName = "oc-lz4"
	// 	cmp = Lz4Cmp{}
	case strings.Contains(encoding, "lzs"):
		cmpName = "lzs"
		cmp = LzsgoCmp{}
	default:
		return cmpName, false
	}
	if cate == "cstp" {
		cs.CstpPickCmp = cmp
	} else {
		cs.DtlsPickCmp = cmp
	}
	return cmpName, true
}

func SToken2Sess(stoken string) *Session {
	stoken = strings.TrimSpace(stoken)
	sarr := strings.Split(stoken, "@")
	token := sarr[1]

	return Token2Sess(token)
}

func Token2Sess(token string) *Session {
	sessMux.RLock()
	defer sessMux.RUnlock()
	return sessions[token]
}

func Dtls2Sess(did string) *Session {
	sessMux.RLock()
	defer sessMux.RUnlock()
	token := dtlsIds[did]
	return sessions[token]
}

func Dtls2CSess(did string) *ConnSession {
	sessMux.RLock()
	defer sessMux.RUnlock()
	token := dtlsIds[did]
	sess := sessions[token]
	if sess == nil {
		return nil
	}

	sess.mux.RLock()
	defer sess.mux.RUnlock()
	return sess.CSess
}

func Dtls2MasterSecret(did string) string {
	sessMux.RLock()
	token := dtlsIds[did]
	sess := sessions[token]
	sessMux.RUnlock()

	if sess == nil {
		return ""
	}

	sess.mux.RLock()
	defer sess.mux.RUnlock()
	if sess.CSess == nil {
		return ""
	}
	return sess.CSess.MasterSecret
}

func DelSess(token string) {
	// sessions.Delete(token)
}

func CloseSess(token string, code ...uint8) {
	sessMux.Lock()
	defer sessMux.Unlock()
	sess, ok := sessions[token]
	if !ok {
		return
	}

	delete(sessions, token)
	delete(dtlsIds, sess.DtlsSid)

	if sess.CSess != nil {
		if len(code) > 0 {
			sess.CSess.UserLogoutCode = code[0]
		}
		sess.CSess.Close()
		return
	}
	AddUserActLogBySess(sess)
}

func CloseCSess(token string) {
	sessMux.RLock()
	defer sessMux.RUnlock()
	sess, ok := sessions[token]
	if !ok {
		return
	}

	if sess.CSess != nil {
		sess.CSess.Close()
	}
}

func DelSessByStoken(stoken string) {
	stoken = strings.TrimSpace(stoken)
	sarr := strings.Split(stoken, "@")
	token := sarr[1]
	CloseSess(token, dbdata.UserLogoutBanner)
}

func AddUserActLog(cs *ConnSession) {
	ua := dbdata.UserActLog{
		Username:        cs.Sess.Username,
		GroupName:       cs.Sess.Group,
		IpAddr:          cs.IpAddr.String(),
		RemoteAddr:      cs.RemoteAddr,
		DeviceType:      cs.Sess.DeviceType,
		PlatformVersion: cs.Sess.PlatformVersion,
		Status:          dbdata.UserLogout,
	}
	ua.Info = dbdata.UserActLogIns.GetInfoOpsById(cs.UserLogoutCode)
	dbdata.UserActLogIns.Add(ua, cs.UserAgent)
}

func AddUserActLogBySess(sess *Session) {
	ua := dbdata.UserActLog{
		Username:        sess.Username,
		GroupName:       sess.Group,
		IpAddr:          "",
		RemoteAddr:      sess.RemoteAddr,
		DeviceType:      sess.DeviceType,
		PlatformVersion: sess.PlatformVersion,
		Status:          dbdata.UserLogout,
	}
	ua.Info = dbdata.UserActLogIns.GetInfoOpsById(dbdata.UserLogoutBanner)
	dbdata.UserActLogIns.Add(ua, sess.UserAgent)
}
