package sessdata

import (
	"bytes"
	"net"
	"sort"
	"time"

	"github.com/cherts/anylink/pkg/utils"
)

type Online struct {
	Token            string    `json:"token"`
	Username         string    `json:"username"`
	Group            string    `json:"group"`
	MacAddr          string    `json:"mac_addr"`
	UniqueMac        bool      `json:"unique_mac"`
	Ip               net.IP    `json:"ip"`
	RemoteAddr       string    `json:"remote_addr"`
	TunName          string    `json:"tun_name"`
	Mtu              int       `json:"mtu"`
	Client           string    `json:"client"`
	BandwidthUp      string    `json:"bandwidth_up"`
	BandwidthDown    string    `json:"bandwidth_down"`
	BandwidthUpAll   string    `json:"bandwidth_up_all"`
	BandwidthDownAll string    `json:"bandwidth_down_all"`
	LastLogin        time.Time `json:"last_login"`
}

type Onlines []Online

func (o Onlines) Len() int {
	return len(o)
}

func (o Onlines) Less(i, j int) bool {
	return bytes.Compare(o[i].Ip, o[j].Ip) < 0
}

func (o Onlines) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

func OnlineSess() []Online {
	var datas Onlines
	sessMux.Lock()
	for _, v := range sessions {
		v.mux.Lock()
		if v.IsActive {
			val := Online{
				Token:            v.Token,
				Ip:               v.CSess.IpAddr,
				Username:         v.Username,
				Group:            v.Group,
				MacAddr:          v.MacAddr,
				UniqueMac:        v.UniqueMac,
				RemoteAddr:       v.CSess.RemoteAddr,
				TunName:          v.CSess.IfName,
				Mtu:              v.CSess.Mtu,
				Client:           v.CSess.Client,
				BandwidthUp:      utils.HumanByte(v.CSess.BandwidthUpPeriod.Load()) + "/s",
				BandwidthDown:    utils.HumanByte(v.CSess.BandwidthDownPeriod.Load()) + "/s",
				BandwidthUpAll:   utils.HumanByte(v.CSess.BandwidthUpAll.Load()),
				BandwidthDownAll: utils.HumanByte(v.CSess.BandwidthDownAll.Load()),
				LastLogin:        v.LastLogin,
			}
			datas = append(datas, val)
		}
		v.mux.Unlock()
	}
	sessMux.Unlock()
	sort.Sort(&datas)
	return datas
}
