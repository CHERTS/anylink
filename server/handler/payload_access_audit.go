package handler

import (
	"crypto/md5"
	"encoding/binary"
	"runtime/debug"
	"time"

	"github.com/cherts/anylink/base"
	"github.com/cherts/anylink/dbdata"
	"github.com/cherts/anylink/pkg/utils"
	"github.com/cherts/anylink/sessdata"
	"github.com/ivpusic/grpool"
	"github.com/songgao/water/waterutil"
)

const (
	acc_proto_udp = iota + 1
	acc_proto_tcp
	acc_proto_https
	acc_proto_http
)

var (
	auditPayload *AuditPayload
	logBatch     *LogBatch
)

// Analyze audit logs
type AuditPayload struct {
	Pool       *grpool.Pool
	IpAuditMap utils.IMaps
}

// Save audit log
type LogBatch struct {
	Logs    []dbdata.AccessAudit
	LogChan chan dbdata.AccessAudit
}

// Asynchronous writing to pool
func (p *AuditPayload) Add(userName string, pl *sessdata.Payload) {
	select {
	case p.Pool.JobQueue <- func() {
		logAudit(userName, pl)
	}:
	default:
		putPayload(pl)
		base.Error("AccessAudit: AuditPayload channel is full")
	}
}

// Data placement
func (l *LogBatch) Write() {
	if len(l.Logs) == 0 {
		return
	}
	_ = dbdata.AddBatch(l.Logs)
	l.Reset()
}

// Clear data
func (l *LogBatch) Reset() {
	l.Logs = []dbdata.AccessAudit{}
}

// Enable batch writing of data
func logAuditBatch() {
	if base.Cfg.AuditInterval < 0 {
		return
	}
	auditPayload = &AuditPayload{
		Pool:       grpool.NewPool(10, 10240),
		IpAuditMap: utils.NewMap("cmap", 0),
	}
	logBatch = &LogBatch{
		LogChan: make(chan dbdata.AccessAudit, 10240),
	}
	var (
		limit       = 100 // Batch writing to the data table exceeds the upper limit
		outTime     = time.NewTimer(time.Second)
		accessAudit = dbdata.AccessAudit{}
	)

	for {
		// Reset timeout time
		outTime.Reset(time.Second * 1)
		select {
		case accessAudit = <-logBatch.LogChan:
			logBatch.Logs = append(logBatch.Logs, accessAudit)
			if len(logBatch.Logs) >= limit {
				if !outTime.Stop() {
					<-outTime.C
				}
				logBatch.Write()
			}
		case <-outTime.C:
			logBatch.Write()
		}
	}
}

// Parse IP packet data
func logAudit(userName string, pl *sessdata.Payload) {
	defer func() {
		if err := recover(); err != nil {
			base.Error("logAudit is panic: ", err, "\n", string(debug.Stack()), "\n", pl.Data)
		}
		putPayload(pl)
	}()

	if !(pl.LType == sessdata.LTypeIPData && pl.PType == 0x00) {
		return
	}

	ipProto := waterutil.IPv4Protocol(pl.Data)
	// access agreement
	var accessProto uint8
	// Only count tcp and udp access
	switch ipProto {
	case waterutil.TCP:
		accessProto = acc_proto_tcp
	case waterutil.UDP:
		accessProto = acc_proto_udp
	default:
		return
	}
	// When the IP message only contains header information, print LOG and exit.
	ipPl := waterutil.IPv4Payload(pl.Data)
	if len(ipPl) < 4 {
		base.Error("ipPl len < 4", ipPl, pl.Data)
		return
	}
	ipPort := (uint16(ipPl[2]) << 8) | uint16(ipPl[3])
	ipSrc := waterutil.IPv4Source(pl.Data)
	ipDst := waterutil.IPv4Destination(pl.Data)
	b := getByte51()
	key := *b
	copy(key[:16], ipSrc)
	copy(key[16:32], ipDst)
	binary.BigEndian.PutUint16(key[32:34], ipPort)
	key[34] = byte(accessProto)
	copy(key[35:51], []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})

	info := ""
	nu := utils.NowSec().Unix()
	if ipProto == waterutil.TCP {
		tcpPlData := waterutil.IPv4Payload(pl.Data)
		// 24 (ACK PSH)
		if len(tcpPlData) < 14 || tcpPlData[13] != 24 {
			return
		}
		accessProto, info = onTCP(tcpPlData)
		// HTTPS or HTTP
		if accessProto != acc_proto_tcp {
			// Store the key containing only IP data in advance to avoid recording both the domain name and IP data.
			ipKey := make([]byte, 51)
			copy(ipKey, key)
			ipS := utils.BytesToString(ipKey)
			auditPayload.IpAuditMap.Set(ipS, nu)

			key[34] = byte(accessProto)
			// Store the key containing the domain name
			if info != "" {
				md5Sum := md5.Sum([]byte(info))
				copy(key[35:51], md5Sum[:])
			}
		}
	}
	s := utils.BytesToString(key)

	// The judgment already exists and has not expired
	v, ok := auditPayload.IpAuditMap.Get(s)
	if ok && nu-v.(int64) < int64(base.Cfg.AuditInterval) {
		// Recycle byte objects
		putByte51(b)
		return
	}

	auditPayload.IpAuditMap.Set(s, nu)

	audit := dbdata.AccessAudit{
		Username:    userName,
		Protocol:    uint8(ipProto),
		Src:         ipSrc.String(),
		Dst:         ipDst.String(),
		DstPort:     ipPort,
		CreatedAt:   utils.NowSec(),
		AccessProto: accessProto,
		Info:        info,
	}
	select {
	case logBatch.LogChan <- audit:
	default:
		base.Error("AccessAudit: LogChan channel is full")
		return
	}
}
