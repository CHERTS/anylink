package handler

import (
	"encoding/xml"
	"log"
	"os/exec"
)

const BufferSize = 2048

type ClientRequest struct {
	XMLName              xml.Name       `xml:"config-auth"`
	Client               string         `xml:"client,attr"`                 // Usually it’s a VPN
	Type                 string         `xml:"type,attr"`                   // Request type init logout auth-reply
	AggregateAuthVersion string         `xml:"aggregate-auth-version,attr"` // Usually 2
	Version              string         `xml:"version"`                     // Client version number
	GroupAccess          string         `xml:"group-access"`                // Requested address
	GroupSelect          string         `xml:"group-select"`                // Selected group name
	SessionId            string         `xml:"session-id"`
	SessionToken         string         `xml:"session-token"`
	Auth                 auth           `xml:"auth"`
	DeviceId             deviceId       `xml:"device-id"`
	MacAddressList       macAddressList `xml:"mac-address-list"`
}

type auth struct {
	Username          string `xml:"username"`
	Password          string `xml:"password"`
	SecondaryPassword string `xml:"secondary_password"`
}

type deviceId struct {
	ComputerName    string `xml:"computer-name,attr"`
	DeviceType      string `xml:"device-type,attr"`
	PlatformVersion string `xml:"platform-version,attr"`
	UniqueId        string `xml:"unique-id,attr"`
	UniqueIdGlobal  string `xml:"unique-id-global,attr"`
}

type macAddressList struct {
	MacAddress string `xml:"mac-address"`
}

func execCmd(cmdStrs []string) error {
	for _, cmdStr := range cmdStrs {
		cmd := exec.Command("sh", "-c", cmdStr)
		b, err := cmd.CombinedOutput()
		if err != nil {
			log.Println(string(b))
			return err
		}
	}
	return nil
}
