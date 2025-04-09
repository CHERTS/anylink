package dbdata

import (
	"testing"

	"github.com/cherts/anylink/base"
	"github.com/stretchr/testify/assert"
)

func TestCheckUser(t *testing.T) {
	base.Test()
	ast := assert.New(t)

	preIpData()
	defer closeIpdata()

	group := "group1"

	// add a group
	dns := []ValData{{Val: "1.1.1.1"}}
	route := []ValData{{Val: "192.168.1.0/24"}}
	g := Group{Name: group, Status: 1, ClientDns: dns, RouteInclude: route}
	err := SetGroup(&g)
	ast.Nil(err)
	// Judgment IpMask
	ast.Equal(g.RouteInclude[0].IpMask, "192.168.1.0/255.255.255.0")

	// add a user
	pincode := "a123456"
	u := User{Username: "aaa", PinCode: pincode, Groups: []string{group}, Status: 1}
	err = SetUser(&u)
	ast.Nil(err)

	// Verify PinCode + OtpSecret
	// totp := gotp.NewDefaultTOTP(u.OtpSecret)
	// secret := totp.Now()
	// err = CheckUser("aaa", u.PinCode+secret, group)
	// ast.Nil(err)

	// Verify password individually
	u.DisableOtp = true
	_ = SetUser(&u)
	ext := map[string]any{
		"mac_addr": "",
	}
	err = CheckUser("aaa", pincode, group, ext)
	ast.Nil(err)

	// Add a radius group
	group2 := "group2"
	authData := map[string]interface{}{
		"type": "radius",
		"radius": map[string]string{
			"addr":   "192.168.1.12:1044",
			"secret": "43214132",
		},
	}
	g2 := Group{Name: group2, Status: 1, ClientDns: dns, RouteInclude: route, Auth: authData}
	err = SetGroup(&g2)
	ast.Nil(err)
	err = CheckUser("aaa", "bbbbbbb", group2, ext)
	if ast.NotNil(err) {
		ast.Contains(err.Error(), "aaa Radius server connection abnormality")
	}
	// Add user policy
	dns2 := []ValData{{Val: "8.8.8.8"}}
	route2 := []ValData{{Val: "192.168.2.0/24"}}
	p1 := Policy{Username: "aaa", Status: 1, ClientDns: dns2, RouteInclude: route2}
	err = SetPolicy(&p1)
	ast.Nil(err)
	err = CheckUser("aaa", pincode, group, ext)
	ast.Nil(err)
	// Add an ldap group
	group3 := "group3"
	authData = map[string]interface{}{
		"type": "ldap",
		"ldap": map[string]interface{}{
			"addr":         "192.168.8.12:389",
			"tls":          true,
			"bind_name":    "userfind@abc.com",
			"bind_pwd":     "afdbfdsafds",
			"base_dn":      "dc=abc,dc=com",
			"object_class": "person",
			"search_attr":  "sAMAccountName",
			"member_of":    "cn=vpn,cn=user,dc=abc,dc=com",
		},
	}
	g3 := Group{Name: group3, Status: 1, ClientDns: dns, RouteInclude: route, Auth: authData}
	err = SetGroup(&g3)
	ast.Nil(err)
	err = CheckUser("aaa", "bbbbbbb", group3, ext)
	if ast.NotNil(err) {
		ast.Equal("aaa LDAP server connection exception, please check the server and port", err.Error())
	}
}
