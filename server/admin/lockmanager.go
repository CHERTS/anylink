package admin

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cherts/anylink/base"
)

type LockInfo struct {
	Description string     `json:"description"` // Lock reason
	Username    string     `json:"username"`    // username
	IP          string     `json:"ip"`          // IP address
	State       *LockState `json:"state"`       // Lock status information
}
type LockState struct {
	Locked       bool      `json:"locked"`      // Locked
	FailureCount int       `json:"attempts"`    // Number of failures
	LockTime     time.Time `json:"lock_time"`   // Lock deadline
	LastAttempt  time.Time `json:"lastAttempt"` // Time of last attempt
}
type IPWhitelists struct {
	IP   net.IP
	CIDR *net.IPNet
}

type LockManager struct {
	mu sync.Mutex
	// LoginStatus   sync.Map                         // Login status
	ipLocks       map[string]*LockState            // Global IP lock status
	userLocks     map[string]*LockState            // Global user lock status
	ipUserLocks   map[string]map[string]*LockState // Single user IP lock status
	ipWhitelists  []IPWhitelists                   // Global IP whitelist, including IP addresses and CIDR ranges
	cleanupTicker *time.Ticker
}

var lockmanager *LockManager
var once sync.Once

func GetLockManager() *LockManager {
	once.Do(func() {
		lockmanager = &LockManager{
			// LoginStatus:  sync.Map{},
			ipLocks:      make(map[string]*LockState),
			userLocks:    make(map[string]*LockState),
			ipUserLocks:  make(map[string]map[string]*LockState),
			ipWhitelists: make([]IPWhitelists, 0),
		}
	})
	return lockmanager
}

const defaultGlobalLockStateExpirationTime = 3600

func InitLockManager() {
	lm := GetLockManager()
	if base.Cfg.AntiBruteForce {
		if base.Cfg.GlobalLockStateExpirationTime <= 0 {
			base.Cfg.GlobalLockStateExpirationTime = defaultGlobalLockStateExpirationTime
		}
		lm.StartCleanupTicker()
		lm.InitIPWhitelist()
	}
}

func GetLocksInfo(w http.ResponseWriter, r *http.Request) {
	lm := GetLockManager()
	locksInfo := lm.GetLocksInfo()

	RespSucess(w, locksInfo)
}

func UnlockUser(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	lockinfo := LockInfo{}
	if err := json.Unmarshal(body, &lockinfo); err != nil {
		RespError(w, RespInternalErr, err)
		return
	}

	if lockinfo.State == nil {
		RespError(w, RespInternalErr, "Locked user not found!")
		return
	}
	lm := GetLockManager()

	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Find lockout status by username and IP
	var state *LockState
	switch {
	case lockinfo.IP == "" && lockinfo.Username != "":
		state = lm.userLocks[lockinfo.Username] // Global User Lockout
	case lockinfo.Username != "" && lockinfo.IP != "":
		if userIPMap, exists := lm.ipUserLocks[lockinfo.Username]; exists {
			state = userIPMap[lockinfo.IP] // Single User IP Lock
		}
	default:
		state = lm.ipLocks[lockinfo.IP] // Global IP Lock
	}

	if state == nil || !state.Locked {
		RespError(w, RespInternalErr, "Lock status Not found or unlocked")
		return
	}

	lm.Unlock(state)
	base.Info("Unlock successfully:", lockinfo.Description, lockinfo.Username, lockinfo.IP)

	RespSucess(w, "Unlocked successfully!")
}

func (lm *LockManager) GetLocksInfo() []LockInfo {
	var locksInfo []LockInfo

	lm.mu.Lock()
	defer lm.mu.Unlock()

	for ip, state := range lm.ipLocks {
		if base.Cfg.MaxGlobalIPBanCount > 0 && state.Locked {
			info := LockInfo{
				Description: "Global IP lock",
				Username:    "",
				IP:          ip,
				State: &LockState{
					Locked:       state.Locked,
					FailureCount: state.FailureCount,
					LockTime:     state.LockTime,
					LastAttempt:  state.LastAttempt,
				},
			}
			locksInfo = append(locksInfo, info)
		}
	}

	for username, state := range lm.userLocks {
		if base.Cfg.MaxGlobalUserBanCount > 0 && state.Locked {
			info := LockInfo{
				Description: "Global user lockout",
				Username:    username,
				IP:          "",
				State: &LockState{
					Locked:       state.Locked,
					FailureCount: state.FailureCount,
					LockTime:     state.LockTime,
					LastAttempt:  state.LastAttempt,
				},
			}
			locksInfo = append(locksInfo, info)
		}
	}

	for username, ipStates := range lm.ipUserLocks {
		for ip, state := range ipStates {
			if base.Cfg.MaxBanCount > 0 && state.Locked {
				info := LockInfo{
					Description: "Single user IP lock",
					Username:    username,
					IP:          ip,
					State: &LockState{
						Locked:       state.Locked,
						FailureCount: state.FailureCount,
						LockTime:     state.LockTime,
						LastAttempt:  state.LastAttempt,
					},
				}
				locksInfo = append(locksInfo, info)
			}
		}
	}
	return locksInfo
}

// Initialize IP whitelist
func (lm *LockManager) InitIPWhitelist() {
	ipWhitelist := strings.Split(base.Cfg.IPWhitelist, ",")
	for _, ipWhitelist := range ipWhitelist {
		ipWhitelist = strings.TrimSpace(ipWhitelist)
		if ipWhitelist == "" {
			continue
		}

		_, ipNet, err := net.ParseCIDR(ipWhitelist)
		if err == nil {
			lm.ipWhitelists = append(lm.ipWhitelists, IPWhitelists{CIDR: ipNet})
			continue
		}

		ip := net.ParseIP(ipWhitelist)
		if ip != nil {
			lm.ipWhitelists = append(lm.ipWhitelists, IPWhitelists{IP: ip})
			continue
		}
	}
}

// Check if the IP is in the whitelist
func (lm *LockManager) IsWhitelisted(ip string) bool {
	clientIP := net.ParseIP(ip)
	if clientIP == nil {
		return false
	}
	for _, ipWhitelist := range lm.ipWhitelists {
		if ipWhitelist.CIDR != nil && ipWhitelist.CIDR.Contains(clientIP) {
			return true
		}
		if ipWhitelist.IP != nil && ipWhitelist.IP.Equal(clientIP) {
			return true
		}
	}
	return false
}

func (lm *LockManager) StartCleanupTicker() {
	lm.cleanupTicker = time.NewTicker(1 * time.Minute)
	go func() {
		for range lm.cleanupTicker.C {
			lm.CleanupExpiredLocks()
		}
	}()
}

// Periodically clean up expired locks
func (lm *LockManager) CleanupExpiredLocks() {
	now := time.Now()
	lm.mu.Lock()
	defer lm.mu.Unlock()

	for ip, state := range lm.ipLocks {
		if !lm.CheckLockState(state, now, base.Cfg.GlobalIPBanResetTime) ||
			now.Sub(state.LastAttempt) > time.Duration(base.Cfg.GlobalLockStateExpirationTime)*time.Second {
			delete(lm.ipLocks, ip)
		}
	}

	for user, state := range lm.userLocks {
		if !lm.CheckLockState(state, now, base.Cfg.GlobalUserBanResetTime) ||
			now.Sub(state.LastAttempt) > time.Duration(base.Cfg.GlobalLockStateExpirationTime)*time.Second {
			delete(lm.userLocks, user)
		}
	}

	for user, ipMap := range lm.ipUserLocks {
		for ip, state := range ipMap {
			if !lm.CheckLockState(state, now, base.Cfg.BanResetTime) ||
				now.Sub(state.LastAttempt) > time.Duration(base.Cfg.GlobalLockStateExpirationTime)*time.Second {
				delete(ipMap, ip)
				if len(ipMap) == 0 {
					delete(lm.ipUserLocks, user)
				}
			}
		}
	}
}

// Check Global IP Lock
func (lm *LockManager) CheckGlobalIPLock(ip string, now time.Time) bool {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	state, exists := lm.ipLocks[ip]
	if !exists {
		return false
	}

	return lm.CheckLockState(state, now, base.Cfg.GlobalIPBanResetTime)
}

// Checking Global User Lockout
func (lm *LockManager) CheckGlobalUserLock(username string, now time.Time) bool {
	// I don't know why Cisco AnyConnect sends an empty user request every time it connects...
	if username == "" {
		return false
	}
	lm.mu.Lock()
	defer lm.mu.Unlock()

	state, exists := lm.userLocks[username]
	if !exists {
		return false
	}
	return lm.CheckLockState(state, now, base.Cfg.GlobalUserBanResetTime)
}

// Check IP lockout for a single user
func (lm *LockManager) CheckUserIPLock(username, ip string, now time.Time) bool {
	// I don't know why Cisco AnyConnect sends an empty user request every time it connects...
	if username == "" {
		return false
	}
	lm.mu.Lock()
	defer lm.mu.Unlock()

	userIPMap, userExists := lm.ipUserLocks[username]
	if !userExists {
		return false
	}

	state, ipExists := userIPMap[ip]
	if !ipExists {
		return false
	}

	return lm.CheckLockState(state, now, base.Cfg.BanResetTime)
}

// Update global IP lock status
func (lm *LockManager) UpdateGlobalIPLock(ip string, now time.Time, success bool) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	state, exists := lm.ipLocks[ip]
	if !exists {
		state = &LockState{}
		lm.ipLocks[ip] = state
	}

	lm.UpdateLockState(state, now, success, base.Cfg.MaxGlobalIPBanCount, base.Cfg.GlobalIPLockTime)
}

// Update global user lock status
func (lm *LockManager) UpdateGlobalUserLock(username string, now time.Time, success bool) {
	// I don't know why Cisco AnyConnect sends an empty user request every time it connects...
	if username == "" {
		return
	}
	lm.mu.Lock()
	defer lm.mu.Unlock()

	state, exists := lm.userLocks[username]
	if !exists {
		state = &LockState{}
		lm.userLocks[username] = state
	}

	lm.UpdateLockState(state, now, success, base.Cfg.MaxGlobalUserBanCount, base.Cfg.GlobalUserLockTime)
}

// Update IP lock status for a single user
func (lm *LockManager) UpdateUserIPLock(username, ip string, now time.Time, success bool) {
	// I don't know why Cisco AnyConnect sends an empty user request every time it connects...
	if username == "" {
		return
	}
	lm.mu.Lock()
	defer lm.mu.Unlock()

	userIPMap, userExists := lm.ipUserLocks[username]
	if !userExists {
		userIPMap = make(map[string]*LockState)
		lm.ipUserLocks[username] = userIPMap
	}

	state, ipExists := userIPMap[ip]
	if !ipExists {
		state = &LockState{}
		userIPMap[ip] = state
	}

	lm.UpdateLockState(state, now, success, base.Cfg.MaxBanCount, base.Cfg.LockTime)
}

// Update lock status
func (lm *LockManager) UpdateLockState(state *LockState, now time.Time, success bool, maxBanCount, lockTime int) {
	if success {
		lm.Unlock(state) // Unlock after successful login
	} else {
		state.FailureCount++
		if state.FailureCount >= maxBanCount {
			state.LockTime = now.Add(time.Duration(lockTime) * time.Second)
			state.Locked = true // Lock when threshold is exceeded
		}
	}
	state.LastAttempt = now
}

// Check lock status
func (lm *LockManager) CheckLockState(state *LockState, now time.Time, resetTime int) bool {
	if state == nil || state.LastAttempt.IsZero() {
		return false
	}

	// If the lock time is exceeded, reset the lock status
	if !state.LockTime.IsZero() && now.After(state.LockTime) {
		lm.Unlock(state) // Unlock after lock-up period
		return false
	}
	// If the window time is exceeded, reset the failure count
	if now.Sub(state.LastAttempt) > time.Duration(resetTime)*time.Second {
		state.FailureCount = 0
		return false
	}
	return state.Locked
}

// Unlock
func (lm *LockManager) Unlock(state *LockState) {
	state.FailureCount = 0
	state.LockTime = time.Time{}
	state.Locked = false
}

// Check lock status
func (lm *LockManager) CheckLocked(username, ipaddr string) bool {
	if !base.Cfg.AntiBruteForce {
		return true
	}

	ip, _, err := net.SplitHostPort(ipaddr) // Extract the pure IP address, removing the port number
	if err != nil {
		base.Error("Checking lock status failed, error in extracting IP address: ", ipaddr)
		return true
	}
	now := time.Now()

	// Check if the IP is in the whitelist
	if lm.IsWhitelisted(ip) {
		return true
	}

	// Check Global IP Lock
	if base.Cfg.MaxGlobalIPBanCount > 0 && lm.CheckGlobalIPLock(ip, now) {
		base.Warn("IP ", ip, "is globally locked. Try again later.")
		return false
	}

	// Checking Global User Lockout
	if base.Cfg.MaxGlobalUserBanCount > 0 && lm.CheckGlobalUserLock(username, now) {
		base.Warn("User ", username, "is globally locked. Try again later.")
		return false
	}

	// Check IP lockout for a single user
	if base.Cfg.MaxBanCount > 0 && lm.CheckUserIPLock(username, ip, now) {
		base.Warn("IP ", ip, "is locked for user", username, "Try again later.")
		return false
	}

	return true
}

// Update user login status
func (lm *LockManager) UpdateLoginStatus(username, ipaddr string, loginStatus bool) {
	ip, _, err := net.SplitHostPort(ipaddr) // Extract the pure IP address, removing the port number
	if err != nil {
		base.Error("Failed to update login status, error in extracting IP address: ", ipaddr)
		return
	}
	now := time.Now()

	// Update user login status
	lm.UpdateGlobalIPLock(ip, now, loginStatus)
	lm.UpdateGlobalUserLock(username, now, loginStatus)
	lm.UpdateUserIPLock(username, ip, now, loginStatus)
}
