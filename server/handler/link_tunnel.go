package handler

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"text/template"

	"github.com/cherts/anylink/base"
	"github.com/cherts/anylink/dbdata"
	"github.com/cherts/anylink/sessdata"
)

var (
	hn string
)

func init() {
	// Get host name
	hn, _ = os.Hostname()
}

func HttpSetHeader(w http.ResponseWriter, key string, value string) {
	w.Header()[key] = []string{value}
}

func HttpAddHeader(w http.ResponseWriter, key string, value string) {
	w.Header()[key] = append(w.Header()[key], value)
}

func LinkTunnel(w http.ResponseWriter, r *http.Request) {
	// TODO Debug information output
	if base.GetLogLevel() == base.LogLevelTrace {
		hd, _ := httputil.DumpRequest(r, true)
		base.Trace("LinkTunnel: ", string(hd))
	}

	// Determine the value of session-token
	cookie, err := r.Cookie("webvpn")
	if err != nil || cookie.Value == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	sess := sessdata.SToken2Sess(cookie.Value)
	if sess == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Open link
	cSess := sess.NewConn()
	if cSess == nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Client information
	cstpMtu := r.Header.Get("X-CSTP-MTU")
	cstpBaseMtu := r.Header.Get("X-CSTP-Base-MTU")
	masterSecret := r.Header.Get("X-DTLS-Master-Secret")
	localIp := r.Header.Get("X-Cstp-Local-Address-Ip4")
	// export ip
	exportIp4 := r.Header.Get("X-Cstp-Remote-Address-Ip4")
	mobile := r.Header.Get("X-Cstp-License")

	cSess.SetMtu(cstpMtu)
	cSess.MasterSecret = masterSecret
	cSess.RemoteAddr = r.RemoteAddr
	cSess.UserAgent = strings.ToLower(r.UserAgent())
	cSess.LocalIp = net.ParseIP(localIp)
	cstpKeepalive := base.Cfg.CstpKeepalive
	cstpDpd := base.Cfg.CstpDpd
	cSess.Client = "pc"
	if mobile == "mobile" {
		// Mobile client
		cstpKeepalive = base.Cfg.MobileKeepalive
		cstpDpd = base.Cfg.MobileDpd
		cSess.Client = "mobile"
	}
	cSess.CstpDpd = cstpDpd

	dtlsPort := "4433"
	if strings.Contains(base.Cfg.ServerDTLSAddr, ":") {
		ss := strings.Split(base.Cfg.ServerDTLSAddr, ":")
		dtlsPort = ss[1]
	}

	base.Info(sess.Username, cSess.IpAddr, cSess.MacHw, cSess.Client, mobile)

	// Detect cipher suites
	dtlsCiphersuite := checkDtls12Ciphersuite(r.Header.Get("X-Dtls12-Ciphersuite"))
	base.Trace("dtlsCiphersuite", dtlsCiphersuite)

	// Return client data
	HttpSetHeader(w, "Server", fmt.Sprintf("%s %s", base.APP_NAME, base.APP_VER))
	HttpSetHeader(w, "X-CSTP-Version", "1")
	HttpSetHeader(w, "X-CSTP-Server-Name", fmt.Sprintf("%s %s", base.APP_NAME, base.APP_VER))
	HttpSetHeader(w, "X-CSTP-Protocol", "Copyright (c) 2004 Cisco Systems, Inc.")
	HttpSetHeader(w, "X-CSTP-Address", cSess.IpAddr.String())             // assigned ip address
	HttpSetHeader(w, "X-CSTP-Netmask", sessdata.IpPool.Ipv4Mask.String()) // subnet mask
	HttpSetHeader(w, "X-CSTP-Hostname", hn)                               // machine name
	HttpSetHeader(w, "X-CSTP-Base-MTU", cstpBaseMtu)
	// Default search domain for client dns
	if base.Cfg.DefaultDomain != "" {
		HttpSetHeader(w, "X-CSTP-Default-Domain", base.Cfg.DefaultDomain)
	}

	// Compression
	if cmpName, ok := cSess.SetPickCmp("cstp", r.Header.Get("X-Cstp-Accept-Encoding")); ok {
		HttpSetHeader(w, "X-CSTP-Content-Encoding", cmpName)
	}
	if cmpName, ok := cSess.SetPickCmp("dtls", r.Header.Get("X-Dtls-Accept-Encoding")); ok {
		HttpSetHeader(w, "X-DTLS-Content-Encoding", cmpName)
	}

	// Set user policy
	SetUserPolicy(cSess.Username, cSess.Group)

	// Allow the local LAN to access the VPN network and must be placed first in the route
	if cSess.Group.AllowLan {
		HttpSetHeader(w, "X-CSTP-Split-Exclude", "0.0.0.0/255.255.255.255")
	}
	// Dns address
	for _, v := range cSess.Group.ClientDns {
		HttpAddHeader(w, "X-CSTP-DNS", v.Val)
	}
	// Allowed routes
	for _, v := range cSess.Group.RouteInclude {
		if strings.ToLower(v.Val) == dbdata.All {
			continue
		}
		HttpAddHeader(w, "X-CSTP-Split-Include", v.IpMask)
	}
	// Route not allowed
	for _, v := range cSess.Group.RouteExclude {
		HttpAddHeader(w, "X-CSTP-Split-Exclude", v.IpMask)
	}
	// Exclude export ip routing (export ip is not encrypted for transmission)
	if base.Cfg.ExcludeExportIp && exportIp4 != "" {
		HttpAddHeader(w, "X-CSTP-Split-Exclude", exportIp4+"/255.255.255.255")
	}

	HttpSetHeader(w, "X-CSTP-Lease-Duration", "1209600") // IP address lease period
	HttpSetHeader(w, "X-CSTP-Session-Timeout", "none")
	HttpSetHeader(w, "X-CSTP-Session-Timeout-Alert-Interval", "60")
	HttpSetHeader(w, "X-CSTP-Session-Timeout-Remaining", "none")
	HttpSetHeader(w, "X-CSTP-Idle-Timeout", "18000")
	HttpSetHeader(w, "X-CSTP-Disconnected-Timeout", "18000")
	HttpSetHeader(w, "X-CSTP-Keep", "true")
	HttpSetHeader(w, "X-CSTP-Tunnel-All-DNS", "false")

	HttpSetHeader(w, "X-CSTP-Rekey-Time", "43200") // 172800
	HttpSetHeader(w, "X-CSTP-Rekey-Method", "new-tunnel")
	HttpSetHeader(w, "X-DTLS-Rekey-Time", "43200")
	HttpSetHeader(w, "X-DTLS-Rekey-Method", "new-tunnel")

	HttpSetHeader(w, "X-CSTP-DPD", fmt.Sprintf("%d", cstpDpd))
	HttpSetHeader(w, "X-CSTP-Keepalive", fmt.Sprintf("%d", cstpKeepalive))
	// HttpSetHeader(w, "X-CSTP-Banner", banner.Banner)
	HttpSetHeader(w, "X-CSTP-MSIE-Proxy-Lockdown", "true")
	HttpSetHeader(w, "X-CSTP-Smartcard-Removal-Disconnect", "true")

	HttpSetHeader(w, "X-CSTP-MTU", fmt.Sprintf("%d", cSess.Mtu)) // 1399
	HttpSetHeader(w, "X-DTLS-MTU", fmt.Sprintf("%d", cSess.Mtu))

	HttpSetHeader(w, "X-DTLS-Session-ID", sess.DtlsSid)
	HttpSetHeader(w, "X-DTLS-Port", dtlsPort)
	HttpSetHeader(w, "X-DTLS-DPD", fmt.Sprintf("%d", cstpDpd))
	HttpSetHeader(w, "X-DTLS-Keepalive", fmt.Sprintf("%d", cstpKeepalive))
	HttpSetHeader(w, "X-DTLS12-CipherSuite", dtlsCiphersuite)

	HttpSetHeader(w, "X-CSTP-License", "accept")
	HttpSetHeader(w, "X-CSTP-Routing-Filtering-Ignore", "false")
	HttpSetHeader(w, "X-CSTP-Quarantine", "false")
	HttpSetHeader(w, "X-CSTP-Disable-Always-On-VPN", "false")
	HttpSetHeader(w, "X-CSTP-Client-Bypass-Protocol", "false")
	HttpSetHeader(w, "X-CSTP-TCP-Keepalive", "false")
	// Set up domain name split tunneling (not supported on mobile devices)
	if mobile != "mobile" {
		SetPostAuthXml(cSess.Group, w)
	}

	w.WriteHeader(http.StatusOK)

	hClone := w.Header().Clone()
	buf := &bytes.Buffer{}
	_ = hClone.Write(buf)
	base.Debug("LinkTunnel Response Header:", buf.String())

	hj := w.(http.Hijacker)
	conn, bufRW, err := hj.Hijack()
	if err != nil {
		base.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Start data processing
	switch base.Cfg.LinkMode {
	case base.LinkModeTUN:
		err = LinkTun(cSess)
	case base.LinkModeTAP:
		err = LinkTap(cSess)
	case base.LinkModeMacvtap:
		err = LinkMacvtap(cSess)
	}
	if err != nil {
		conn.Close()
		base.Error(err)
		return
	}
	dbdata.UserActLogIns.Add(dbdata.UserActLog{
		Username:        sess.Username,
		GroupName:       sess.Group,
		IpAddr:          cSess.IpAddr.String(),
		RemoteAddr:      cSess.RemoteAddr,
		DeviceType:      sess.DeviceType,
		PlatformVersion: sess.PlatformVersion,
		Status:          dbdata.UserConnected,
	}, cSess.UserAgent)

	go LinkCstp(conn, bufRW, cSess)
}

// Set up domain name split tunneling
func SetPostAuthXml(g *dbdata.Group, w http.ResponseWriter) error {
	if g.DsExcludeDomains == "" && g.DsIncludeDomains == "" {
		return nil
	}
	tmpl, err := template.New("post_auth_xml").Parse(ds_domains_xml)
	if err != nil {
		return err
	}
	var result bytes.Buffer
	err = tmpl.Execute(&result, g)
	if err != nil {
		return err
	}
	xmlAuth := ""
	for _, v := range strings.Split(result.String(), "\n") {
		xmlAuth += strings.TrimSpace(v)
	}
	HttpSetHeader(w, "X-CSTP-Post-Auth-XML", xmlAuth)
	return nil
}

// Set user policies to override Group attribute values
func SetUserPolicy(username string, g *dbdata.Group) {
	userPolicy := dbdata.GetPolicy(username)
	if userPolicy.Id != 0 && userPolicy.Status == 1 {
		base.Debug(username + " use UserPolicy")
		g.AllowLan = userPolicy.AllowLan
		g.ClientDns = userPolicy.ClientDns
		g.RouteInclude = userPolicy.RouteInclude
		g.RouteExclude = userPolicy.RouteExclude
		g.DsExcludeDomains = userPolicy.DsExcludeDomains
		g.DsIncludeDomains = userPolicy.DsIncludeDomains
	}
}
