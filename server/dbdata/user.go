package dbdata

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cherts/anylink/pkg/utils"
	"github.com/xlzd/gotp"
)

// type User struct {
// 	Id       int    `json:"id"  xorm:"pk autoincr not null"`
// 	Username string `json:"username" storm:"not null unique"`
// 	Nickname string `json:"nickname"`
// 	Email    string `json:"email"`
// 	// Password  string    `json:"password"`
// 	PinCode    string    `json:"pin_code"`
// 	OtpSecret  string    `json:"otp_secret"`
// 	DisableOtp bool      `json:"disable_otp"` // disable otp
// 	Groups     []string  `json:"groups"`
// 	Status     int8      `json:"status"` // 1 normal
// 	SendEmail  bool      `json:"send_email"`
// 	CreatedAt  time.Time `json:"created_at"`
// 	UpdatedAt  time.Time `json:"updated_at"`
// }

func SetUser(v *User) error {
	var err error
	if v.Username == "" || len(v.Groups) == 0 {
		return errors.New("Username or group error")
	}

	planPass := v.PinCode
	// Automatically generate password
	if len(planPass) < 6 {
		planPass = utils.RandomRunes(8)
	}
	v.PinCode = planPass

	if v.OtpSecret == "" {
		v.OtpSecret = gotp.RandomSecret(32)
	}

	// Determine whether the group is valid
	ng := []string{}
	groups := GetGroupNames()
	for _, g := range v.Groups {
		if utils.InArrStr(groups, g) {
			ng = append(ng, g)
		}
	}
	if len(ng) == 0 {
		return errors.New("Username or group error")
	}
	v.Groups = ng

	v.UpdatedAt = time.Now()
	if v.Id > 0 {
		err = Set(v)
	} else {
		err = Add(v)
	}

	return err
}

// Verify user login information
func CheckUser(name, pwd, group string) error {
	// Get logged in group data
	groupData := &Group{}
	err := One("Name", group, groupData)
	if err != nil || groupData.Status != 1 {
		return fmt.Errorf("%s - %s", name, "User group error")
	}
	// InitializeAuth
	if len(groupData.Auth) == 0 {
		groupData.Auth["type"] = "local"
	}
	authType := groupData.Auth["type"].(string)
	// Local authentication method
	if authType == "local" {
		return checkLocalUser(name, pwd, group)
	}
	// Other authentication methods, support customization
	_, ok := authRegistry[authType]
	if !ok {
		return fmt.Errorf("%s %s", "Unknown authentication method: ", authType)
	}
	auth := makeInstance(authType).(IUserAuth)
	return auth.checkUser(name, pwd, groupData)
}

// Verify local user login information
func checkLocalUser(name, pwd, group string) error {
	// TODO Serious Problem
	// return nil

	pl := len(pwd)
	if name == "" || pl < 6 {
		return fmt.Errorf("%s %s", name, "wrong password")
	}
	v := &User{}
	err := One("Username", name, v)
	if err != nil || v.Status != 1 {
		switch v.Status {
		case 0:
			return fmt.Errorf("%s %s", name, "The user does not exist or the user is deactivated")
		case 2:
			return fmt.Errorf("%s %s", name, "User has expired")
		}
	}
	// Determine user group information
	if !utils.InArrStr(v.Groups, group) {
		return fmt.Errorf("%s %s", name, "User group error")
	}
	// Determine otp information
	pinCode := pwd
	if !v.DisableOtp {
		pinCode = pwd[:pl-6]
		otp := pwd[pl-6:]
		if !checkOtp(name, otp, v.OtpSecret) {
			return fmt.Errorf("%s %s", name, "Dynamic code error")
		}
	}

	// Determine user password
	if pinCode != v.PinCode {
		return fmt.Errorf("%s %s", name, "wrong password")
	}

	return nil
}

// After the user expiration time is reached, update the user status and return a user slice whose status is expired.
func CheckUserlimittime() (limitUser []interface{}) {
	if _, err := xdb.Where("limittime <= ?", time.Now()).And("status = ?", 1).Update(&User{Status: 2}); err != nil {
		return
	}
	user := make(map[int64]User)
	if err := xdb.Where("status != ?", 1).Find(user); err != nil {
		return
	}
	for _, v := range user {
		limitUser = append(limitUser, v.Username)
	}
	return
}

var (
	userOtpMux = sync.Mutex{}
	userOtp    = map[string]time.Time{}
)

func init() {
	go func() {
		expire := time.Second * 60

		for range time.Tick(time.Second * 10) {
			tnow := time.Now()
			userOtpMux.Lock()
			for k, v := range userOtp {
				if tnow.After(v.Add(expire)) {
					delete(userOtp, k)
				}
			}
			userOtpMux.Unlock()
		}
	}()
}

// Determine token information
func checkOtp(name, otp, secret string) bool {
	key := fmt.Sprintf("%s:%s", name, otp)

	userOtpMux.Lock()
	defer userOtpMux.Unlock()

	// Token can only be used once
	if _, ok := userOtp[key]; ok {
		// already exists
		return false
	}
	userOtp[key] = time.Now()

	totp := gotp.NewDefaultTOTP(secret)
	unix := time.Now().Unix()
	verify := totp.Verify(otp, unix)

	return verify
}
