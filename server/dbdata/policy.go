package dbdata

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

func GetPolicy(Username string) *Policy {
	policyData := &Policy{}
	err := One("Username", Username, policyData)
	if err != nil {
		return policyData
	}
	return policyData
}

func SetPolicy(p *Policy) error {
	var err error
	if p.Username == "" {
		return errors.New("username error (empty)")
	}

	// Contains routing
	routeInclude := []ValData{}
	for _, v := range p.RouteInclude {
		if v.Val != "" {
			if v.Val == All {
				routeInclude = append(routeInclude, v)
				continue
			}

			ipMask, ipNet, err := parseIpNet(v.Val)
			if err != nil {
				return errors.New("RouteInclude error" + err.Error())
			}

			if strings.Split(ipMask, "/")[0] != ipNet.IP.String() {
				errMsg := fmt.Sprintf("RouteInclude error: Wrong network address, suggestion: change %s to %s", v.Val, ipNet)
				return errors.New(errMsg)
			}

			v.IpMask = ipMask
			routeInclude = append(routeInclude, v)
		}
	}
	p.RouteInclude = routeInclude
	// exclude route
	routeExclude := []ValData{}
	for _, v := range p.RouteExclude {
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
	p.RouteExclude = routeExclude

	// DNS judgment
	clientDns := []ValData{}
	for _, v := range p.ClientDns {
		if v.Val != "" {
			ip := net.ParseIP(v.Val)
			if ip.String() != v.Val {
				return errors.New("DNS IP error")
			}
			clientDns = append(clientDns, v)
		}
	}
	if len(routeInclude) == 0 || (len(routeInclude) == 1 && routeInclude[0].Val == "all") {
		if len(clientDns) == 0 {
			return errors.New("Default route, a DNS must be set")
		}
	}
	p.ClientDns = clientDns

	// Domain name split tunneling, cannot be filled in at the same time
	p.DsIncludeDomains = strings.TrimSpace(p.DsIncludeDomains)
	p.DsExcludeDomains = strings.TrimSpace(p.DsExcludeDomains)
	if p.DsIncludeDomains != "" && p.DsExcludeDomains != "" {
		return errors.New("Include/exclude domain names cannot be filled in at the same time")
	}
	// Verify the format containing the domain name
	err = CheckDomainNames(p.DsIncludeDomains)
	if err != nil {
		return errors.New("Incorrect domain name included: " + err.Error())
	}
	// Verify the format of excluded domain names
	err = CheckDomainNames(p.DsExcludeDomains)
	if err != nil {
		return errors.New("Wrong domain name to exclude: " + err.Error())
	}

	p.UpdatedAt = time.Now()
	if p.Id > 0 {
		err = Set(p)
	} else {
		err = Add(p)
	}

	return err
}
