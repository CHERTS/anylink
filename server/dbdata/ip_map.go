package dbdata

import (
	"errors"
	"net"
	"time"
)

type IpMap struct {
	Id        int       `json:"id" xorm:"pk autoincr not null"`
	IpAddr    string    `json:"ip_addr" xorm:"varchar(32) not null unique"`
	MacAddr   string    `json:"mac_addr" xorm:"varchar(32) not null unique"`
	UniqueMac bool      `json:"unique_mac" xorm:"Bool index"`
	Username  string    `json:"username" xorm:"varchar(60)"`
	Keep      bool      `json:"keep" xorm:"Bool"` // Keep ip-mac binding
	KeepTime  time.Time `json:"keep_time" xorm:"DateTime"`
	Note      string    `json:"note" xorm:"varchar(255)"` // Description
	LastLogin time.Time `json:"last_login" xorm:"DateTime"`
	UpdatedAt time.Time `json:"updated_at" xorm:"DateTime updated"`
}

func SetIpMap(v *IpMap) error {
	var err error

	if len(v.IpAddr) < 4 || len(v.MacAddr) < 6 {
		return errors.New("IP or MAC error")
	}

	macHw, err := net.ParseMAC(v.MacAddr)
	if err != nil {
		return errors.New("MAC error")
	}
	// Unify the format of macAddr
	v.MacAddr = macHw.String()

	v.UpdatedAt = time.Now()
	if v.Id > 0 {
		err = Set(v)
	} else {
		err = Add(v)
	}
	return err
}
