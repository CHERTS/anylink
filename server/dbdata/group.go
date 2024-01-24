package dbdata

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/cherts/anylink/base"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	Allow = "allow"
	Deny  = "deny"
	All   = "all"
)

// The maximum number of characters for domain name diversion is 20,000
const DsMaxLen = 20000

type GroupLinkAcl struct {
	// Top-down matching default allow * *
	Action string     `json:"action"` // allow、deny
	Val    string     `json:"val"`
	Port   uint16     `json:"port"`
	IpNet  *net.IPNet `json:"ip_net"`
	Note   string     `json:"note"`
}

type ValData struct {
	Val    string `json:"val"`
	IpMask string `json:"ip_mask"`
	Note   string `json:"note"`
}

type GroupNameId struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// type Group struct {
// 	Id               int                    `json:"id" xorm:"pk autoincr not null"`
// 	Name             string                 `json:"name" xorm:"varchar(60) not null unique"`
// 	Note             string                 `json:"note" xorm:"varchar(255)"`
// 	AllowLan         bool                   `json:"allow_lan" xorm:"Bool"`
// 	ClientDns        []ValData              `json:"client_dns" xorm:"Text"`
// 	RouteInclude     []ValData              `json:"route_include" xorm:"Text"`
// 	RouteExclude     []ValData              `json:"route_exclude" xorm:"Text"`
// 	DsExcludeDomains string                 `json:"ds_exclude_domains" xorm:"Text"`
// 	DsIncludeDomains string                 `json:"ds_include_domains" xorm:"Text"`
// 	LinkAcl          []GroupLinkAcl         `json:"link_acl" xorm:"Text"`
// 	Bandwidth        int                    `json:"bandwidth" xorm:"Int"`                           // bandwidth limit
// 	Auth             map[string]interface{} `json:"auth" xorm:"not null default '{}' varchar(255)"` // verification method
// 	Status           int8                   `json:"status" xorm:"Int"`                              // 1 normal
// 	CreatedAt        time.Time              `json:"created_at" xorm:"DateTime created"`
// 	UpdatedAt        time.Time              `json:"updated_at" xorm:"DateTime updated"`
// }

func GetGroupNames() []string {
	var datas []Group
	err := Find(&datas, 0, 0)
	if err != nil {
		base.Error(err)
		return nil
	}
	var names []string
	for _, v := range datas {
		names = append(names, v.Name)
	}
	return names
}

func GetGroupNamesNormal() []string {
	var datas []Group
	err := FindWhere(&datas, 0, 0, "status=1")
	if err != nil {
		base.Error(err)
		return nil
	}
	var names []string
	for _, v := range datas {
		names = append(names, v.Name)
	}
	return names
}

func GetGroupNamesIds() []GroupNameId {
	var datas []Group
	err := Find(&datas, 0, 0)
	if err != nil {
		base.Error(err)
		return nil
	}
	var names []GroupNameId
	for _, v := range datas {
		names = append(names, GroupNameId{Id: v.Id, Name: v.Name})
	}
	return names
}

func SetGroup(g *Group) error {
	var err error
	if g.Name == "" {
		return errors.New("Wrong user group name")
	}

	// Judgment data
	routeInclude := []ValData{}
	for _, v := range g.RouteInclude {
		if v.Val != "" {
			if v.Val == All {
				routeInclude = append(routeInclude, v)
				continue
			}

			ipMask, ipNet, err := parseIpNet(v.Val)

			if err != nil {
				return errors.New("RouteInclude mistake" + err.Error())
			}

			// When delivering routes to Mac systems, they must be standard network addresses.
			if strings.Split(ipMask, "/")[0] != ipNet.IP.String() {
				errMsg := fmt.Sprintf("RouteInclude error: Wrong network address, suggestion: change %s to %s", v.Val, ipNet)
				return errors.New(errMsg)
			}

			v.IpMask = ipMask
			routeInclude = append(routeInclude, v)
		}
	}
	g.RouteInclude = routeInclude
	routeExclude := []ValData{}
	for _, v := range g.RouteExclude {
		if v.Val != "" {
			ipMask, ipNet, err := parseIpNet(v.Val)
			if err != nil {
				return errors.New("RouteExclude error" + err.Error())
			}

			if strings.Split(ipMask, "/")[0] != ipNet.IP.String() {
				errMsg := fmt.Sprintf("RouteInclude error: Wrong network address, suggestion: change %s to %s", v.Val, ipNet)
				return errors.New(errMsg)
			}

			v.IpMask = ipMask
			routeExclude = append(routeExclude, v)
		}
	}
	g.RouteExclude = routeExclude
	// Transform data
	linkAcl := []GroupLinkAcl{}
	for _, v := range g.LinkAcl {
		if v.Val != "" {
			_, ipNet, err := parseIpNet(v.Val)
			if err != nil {
				return errors.New("GroupLinkAcl mistake" + err.Error())
			}
			v.IpNet = ipNet
			linkAcl = append(linkAcl, v)
		}
	}
	g.LinkAcl = linkAcl

	// DNS judgment
	clientDns := []ValData{}
	for _, v := range g.ClientDns {
		if v.Val != "" {
			ip := net.ParseIP(v.Val)
			if ip.String() != v.Val {
				return errors.New("DNS IP error")
			}
			clientDns = append(clientDns, v)
		}
	}
	// Whether to default route
	isDefRoute := len(routeInclude) == 0 || (len(routeInclude) == 1 && routeInclude[0].Val == "all")
	if isDefRoute && len(clientDns) == 0 {
		return errors.New("Default route, a DNS must be set")
	}
	g.ClientDns = clientDns
	// Domain name split tunneling, cannot be filled in at the same time
	g.DsIncludeDomains = strings.TrimSpace(g.DsIncludeDomains)
	g.DsExcludeDomains = strings.TrimSpace(g.DsExcludeDomains)
	if g.DsIncludeDomains != "" && g.DsExcludeDomains != "" {
		return errors.New("Include/exclude domain names cannot be filled in at the same time")
	}
	// Verify the format containing the domain name
	err = CheckDomainNames(g.DsIncludeDomains)
	if err != nil {
		return errors.New("Incorrect domain name included:" + err.Error())
	}
	// Verify the format of excluded domain names
	err = CheckDomainNames(g.DsExcludeDomains)
	if err != nil {
		return errors.New("Wrong domain name to exclude:" + err.Error())
	}
	if isDefRoute && g.DsIncludeDomains != "" {
		return errors.New("Default route, setting \"include domain name\" is not allowed, please reconfigure it.")
	}
	// Logic for handling login methods
	defAuth := map[string]interface{}{
		"type": "local",
	}
	if len(g.Auth) == 0 {
		g.Auth = defAuth
	}
	authType := g.Auth["type"].(string)
	if authType == "local" {
		g.Auth = defAuth
	} else {
		if _, ok := authRegistry[authType]; !ok {
			return errors.New("Unknown authentication method: " + authType)
		}
		auth := makeInstance(authType).(IUserAuth)
		err = auth.checkData(g.Auth)
		if err != nil {
			return err
		}
		// Reset Auth and delete redundant keys
		g.Auth = map[string]interface{}{
			"type":   authType,
			authType: g.Auth[authType],
		}
	}

	g.UpdatedAt = time.Now()
	if g.Id > 0 {
		err = Set(g)
	} else {
		err = Add(g)
	}

	return err
}

func GroupAuthLogin(name, pwd string, authData map[string]interface{}) error {
	g := &Group{Auth: authData}
	authType := g.Auth["type"].(string)
	if _, ok := authRegistry[authType]; !ok {
		return errors.New("Unknown authentication method: " + authType)
	}
	auth := makeInstance(authType).(IUserAuth)
	err := auth.checkData(g.Auth)
	if err != nil {
		return err
	}
	err = auth.checkUser(name, pwd, g)
	return err
}

func parseIpNet(s string) (string, *net.IPNet, error) {
	ip, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		return "", nil, err
	}

	mask := net.IP(ipNet.Mask)
	ipMask := fmt.Sprintf("%s/%s", ip, mask)

	return ipMask, ipNet, nil
}

func CheckDomainNames(domains string) error {
	if domains == "" {
		return nil
	}
	strLen := 0
	str_slice := strings.Split(domains, ",")
	for _, val := range str_slice {
		if val == "" {
			return errors.New(val + " Please separate domain names with commas")
		}
		if !ValidateDomainName(val) {
			return errors.New(val + " Wrong domain name")
		}
		strLen += len(val)
	}
	if strLen > DsMaxLen {
		p := message.NewPrinter(language.English)
		return fmt.Errorf("The character length exceeds the limit, the maximum is %s (excluding commas), please delete some domain names", p.Sprintf("%d", DsMaxLen))
	}
	return nil
}

func ValidateDomainName(domain string) bool {
	RegExp := regexp.MustCompile(`^([a-zA-Z0-9][-a-zA-Z0-9]{0,62}\.)+[A-Za-z]{2,18}$`)
	return RegExp.MatchString(domain)
}
