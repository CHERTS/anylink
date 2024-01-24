package handler

import (
	"github.com/cherts/anylink/base"
	"github.com/cherts/anylink/dbdata"
	"github.com/cherts/anylink/sessdata"
	"github.com/songgao/water/waterutil"
)

func payloadIn(cSess *sessdata.ConnSession, pl *sessdata.Payload) bool {
	if pl.LType == sessdata.LTypeIPData && pl.PType == 0x00 {
		// Make Acl rule judgment
		check := checkLinkAcl(cSess.Group, pl)
		if !check {
			// If the verification fails, it will be discarded directly.
			return false
		}
	}

	closed := false
	select {
	case cSess.PayloadIn <- pl:
	case <-cSess.CloseChan:
		closed = true
	}

	return closed
}

func putPayloadInBefore(cSess *sessdata.ConnSession, pl *sessdata.Payload) {
	// Asynchronous audit log
	if base.Cfg.AuditInterval >= 0 {
		auditPayload.Add(cSess.Username, pl)
		return
	}
	putPayload(pl)
}

func payloadOut(cSess *sessdata.ConnSession, pl *sessdata.Payload) bool {
	dSess := cSess.GetDtlsSession()
	if dSess == nil {
		return payloadOutCstp(cSess, pl)
	} else {
		return payloadOutDtls(cSess, dSess, pl)
	}
}

func payloadOutCstp(cSess *sessdata.ConnSession, pl *sessdata.Payload) bool {
	closed := false

	select {
	case cSess.PayloadOutCstp <- pl:
	case <-cSess.CloseChan:
		closed = true
	}

	return closed
}

func payloadOutDtls(cSess *sessdata.ConnSession, dSess *sessdata.DtlsSession, pl *sessdata.Payload) bool {
	select {
	case cSess.PayloadOutDtls <- pl:
	case <-dSess.CloseChan:
	}

	return false
}

// Acl rule verification
func checkLinkAcl(group *dbdata.Group, pl *sessdata.Payload) bool {
	if pl.LType == sessdata.LTypeIPData && pl.PType == 0x00 && len(group.LinkAcl) > 0 {
	} else {
		return true
	}

	ipDst := waterutil.IPv4Destination(pl.Data)
	ipPort := waterutil.IPv4DestinationPort(pl.Data)
	ipProto := waterutil.IPv4Protocol(pl.Data)
	// fmt.Println("sent:", ip_dst, ip_port)

	// Give priority to dns port
	for _, v := range group.ClientDns {
		if v.Val == ipDst.String() && ipPort == 53 {
			return true
		}
	}

	for _, v := range group.LinkAcl {
		// Loop to determine IP and port
		if v.IpNet.Contains(ipDst) {
			// 放行允许ip的ping
			if v.Port == ipPort || v.Port == 0 || ipProto == waterutil.ICMP {
				if v.Action == dbdata.Allow {
					return true
				} else {
					return false
				}
			}
		}
	}

	return false
}
