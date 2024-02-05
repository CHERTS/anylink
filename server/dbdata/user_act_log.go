package dbdata

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/cherts/anylink/base"
	"github.com/ivpusic/grpool"
	"github.com/spf13/cast"
	"xorm.io/xorm"
)

const (
	UserAuthFail      = 0 // Authentication failed
	UserAuthSuccess   = 1 // Authentication successful
	UserConnected     = 2 // Connection successful
	UserLogout        = 3 // User logout
	UserLogoutLose    = 0 // User disconnected
	UserLogoutBanner  = 1 // Cancel user banner pop-up window
	UserLogoutClient  = 2 // User actively logs out
	UserLogoutTimeout = 3 // User logs out after timeout
	UserLogoutAdmin   = 4 // The account was kicked offline by the administrator
	UserLogoutExpire  = 5 // The account expired and was kicked offline.
	UserIdleTimeout   = 6 // User idle link timeout
)

type UserActLogProcess struct {
	Pool      *grpool.Pool
	StatusOps []string
	OsOps     []string
	ClientOps []string
	InfoOps   []string
}

var (
	UserActLogIns = &UserActLogProcess{
		Pool: grpool.NewPool(1, 100),
		StatusOps: []string{ // Operation type
			UserAuthFail:    "Authentication failed",
			UserAuthSuccess: "Authentication successful",
			UserConnected:   "Connection succeeded",
			UserLogout:      "User logout",
		},
		OsOps: []string{ // Operating system
			0: "Unknown",
			1: "Windows",
			2: "macOS",
			3: "Linux",
			4: "Android",
			5: "iOS",
		},
		ClientOps: []string{ // Client
			0: "Unknown",
			1: "AnyConnect",
			2: "OpenConnect",
			3: "AnyLink",
		},
		InfoOps: []string{ // Information
			UserLogoutLose:    "User disconnected",
			UserLogoutBanner:  "User cancels pop-up window/logout initiated by client",
			UserLogoutClient:  "User/client actively disconnects",
			UserLogoutTimeout: "Session expired and was kicked offline",
			UserLogoutAdmin:   "The account was kicked offline by the administrator",
			UserLogoutExpire:  "The account expired and was kicked offline.",
			UserIdleTimeout:   "User idle link timeout",
		},
	}
)

// Asynchronously writing user operation logs
func (ua *UserActLogProcess) Add(u UserActLog, userAgent string) {
	// os, client, ver
	os_idx, client_idx, ver := ua.ParseUserAgent(userAgent)
	u.Os = os_idx
	u.Client = client_idx
	u.Version = ver
	u.RemoteAddr = strings.Split(u.RemoteAddr, ":")[0]
	// remove extra characters
	infoSlice := strings.Split(u.Info, " ")
	infoLen := len(infoSlice)
	if infoLen > 1 {
		if u.Username == infoSlice[0] {
			u.Info = strings.Join(infoSlice[1:], " ")
		}
		// delete - char
		if infoLen > 2 && infoSlice[1] == "-" {
			u.Info = u.Info[2:]
		}
	}
	// limit the max length of char
	u.Version = substr(u.Version, 0, 15)
	u.DeviceType = substr(u.DeviceType, 0, 128)
	u.PlatformVersion = substr(u.PlatformVersion, 0, 128)
	u.Info = substr(u.Info, 0, 255)

	UserActLogIns.Pool.JobQueue <- func() {
		err := Add(u)
		if err != nil {
			base.Error("Add UserActLog error: ", err)
		}
	}
}

// Escape operation type to facilitate vue display
func (ua *UserActLogProcess) GetStatusOpsWithTag() interface{} {
	type StatusTag struct {
		Key   int    `json:"key"`
		Value string `json:"value"`
		Tag   string `json:"tag"`
	}
	var res []StatusTag
	for k, v := range ua.StatusOps {
		tag := "info"
		switch k {
		case UserAuthFail:
			tag = "danger"
		case UserAuthSuccess:
			tag = "success"
		case UserConnected:
			tag = ""
		}
		res = append(res, StatusTag{k, v, tag})
	}
	return res
}

func (ua *UserActLogProcess) GetInfoOpsById(id uint8) string {
	return ua.InfoOps[id]
}

// Analysis user agent
func (ua *UserActLogProcess) ParseUserAgent(userAgent string) (os_idx, client_idx uint8, ver string) {
	// Unknown
	if len(userAgent) == 0 {
		return 0, 0, ""
	}
	// OS
	os_idx = 0
	if strings.Contains(userAgent, "windows") {
		os_idx = 1
	} else if strings.Contains(userAgent, "mac os") || strings.Contains(userAgent, "darwin_i386") {
		os_idx = 2
	} else if strings.Contains(userAgent, "darwin_arm") || strings.Contains(userAgent, "apple") {
		os_idx = 5
	} else if strings.Contains(userAgent, "android") {
		os_idx = 4
	} else if strings.Contains(userAgent, "linux") {
		os_idx = 3
	}
	// Client
	client_idx = 0
	if strings.Contains(userAgent, "anyconnect") {
		client_idx = 1
	} else if strings.Contains(userAgent, "openconnect") {
		client_idx = 2
	} else if strings.Contains(userAgent, "anylink") {
		client_idx = 3
	}
	// Version
	uaSlice := strings.Split(userAgent, " ")
	ver = uaSlice[len(uaSlice)-1]
	if ver[0] == 'v' {
		ver = ver[1:]
	}
	if !regexp.MustCompile(`^(\d+\.?)+$`).MatchString(ver) {
		ver = ""
	}
	return
}

// Clear user operation log
func (ua *UserActLogProcess) ClearUserActLog(ts string) (int64, error) {
	affected, err := xdb.Where("created_at < '" + ts + "'").Delete(&UserActLog{})
	return affected, err
}

// Filter user operation logs in the background
func (ua *UserActLogProcess) GetSession(values url.Values) *xorm.Session {
	session := xdb.Where("1=1")
	if values.Get("username") != "" {
		session.And("username = ?", values.Get("username"))
	}
	if values.Get("sdate") != "" {
		session.And("created_at >= ?", values.Get("sdate")+" 00:00:00'")
	}
	if values.Get("edate") != "" {
		session.And("created_at <= ?", values.Get("edate")+" 23:59:59'")
	}
	if values.Get("status") != "" {
		session.And("status = ?", cast.ToUint8(values.Get("status"))-1)
	}
	if values.Get("os") != "" {
		session.And("os = ?", cast.ToUint8(values.Get("os"))-1)
	}
	if values.Get("sort") == "1" {
		session.OrderBy("id desc")
	} else {
		session.OrderBy("id asc")
	}
	return session
}

// Intercept string
func substr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}
