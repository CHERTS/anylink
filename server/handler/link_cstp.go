package handler

import (
	"bufio"
	"encoding/binary"
	"net"
	"time"

	"github.com/cherts/anylink/base"
	"github.com/cherts/anylink/dbdata"
	"github.com/cherts/anylink/pkg/utils"
	"github.com/cherts/anylink/sessdata"
)

func LinkCstp(conn net.Conn, bufRW *bufio.ReadWriter, cSess *sessdata.ConnSession) {
	base.Debug("LinkCstp connect ip:", cSess.IpAddr, "user:", cSess.Username, "rip:", conn.RemoteAddr())
	defer func() {
		base.Debug("LinkCstp return", cSess.Username, cSess.IpAddr)
		_ = conn.Close()
		cSess.Close()
	}()

	var (
		err     error
		n       int
		dataLen uint16
		dead    = time.Duration(cSess.CstpDpd+5) * time.Second
	)

	go cstpWrite(conn, bufRW, cSess)

	for {

		// Set timeout limit
		err = conn.SetReadDeadline(utils.NowSec().Add(dead))
		if err != nil {
			base.Error("SetDeadline: ", cSess.Username, err)
			return
		}
		// hdata := make([]byte, BufferSize)
		pl := getPayload()
		n, err = bufRW.Read(pl.Data)
		if err != nil {
			base.Error("read hdata: ", cSess.Username, err)
			return
		}

		// Current limiting settings
		err = cSess.RateLimit(n, true)
		if err != nil {
			base.Error(err)
		}

		switch pl.Data[6] {
		case 0x07: // KEEPALIVE
			// do nothing
			// base.Debug("recv keepalive", cSess.IpAddr)
		case 0x05: // DISCONNECT
			cSess.UserLogoutCode = dbdata.UserLogoutClient
			base.Debug("DISCONNECT", cSess.Username, cSess.IpAddr)
			return
		case 0x03: // DPD-REQ
			// base.Debug("recv DPD-REQ", cSess.IpAddr)
			pl.PType = 0x04
			if payloadOutCstp(cSess, pl) {
				return
			}
		case 0x04:
		// log.Println("recv DPD-RESP")
		case 0x08: // decompress
			if cSess.CstpPickCmp == nil {
				continue
			}
			dst := getByteFull()
			nn, err := cSess.CstpPickCmp.Uncompress(pl.Data[8:], *dst)
			if err != nil {
				putByte(dst)
				base.Error("cstp decompress error", err, nn)
				continue
			}
			binary.BigEndian.PutUint16(pl.Data[4:6], uint16(nn))
			pl.Data = append(pl.Data[:8], (*dst)[:nn]...)
			putByte(dst)
			fallthrough
		case 0x00: // DATA
			// Get data length
			dataLen = binary.BigEndian.Uint16(pl.Data[4:6]) // 4,5
			// Fix cstp data length overflow error
			if 8+dataLen > BufferSize {
				base.Error("recv error dataLen", cSess.Username, dataLen)
				continue
			}
			// Remove header
			copy(pl.Data, pl.Data[8:8+dataLen])
			// Update slice length
			pl.Data = pl.Data[:dataLen]
			// pl.Data = append(pl.Data[:0], pl.Data[8:8+dataLen]...)
			if payloadIn(cSess, pl) {
				return
			}
		}
	}
}

func cstpWrite(conn net.Conn, bufRW *bufio.ReadWriter, cSess *sessdata.ConnSession) {
	defer func() {
		base.Debug("cstpWrite return", cSess.Username, cSess.IpAddr)
		_ = conn.Close()
		cSess.Close()
	}()

	var (
		err error
		n   int
		pl  *sessdata.Payload
	)

	for {
		select {
		case pl = <-cSess.PayloadOutCstp:
		case <-cSess.CloseChan:
			return
		}

		if pl.LType != sessdata.LTypeIPData {
			continue
		}

		if pl.PType == 0x00 {
			isCompress := false
			if cSess.CstpPickCmp != nil && len(pl.Data) > base.Cfg.NoCompressLimit {
				dst := getByteFull()
				size, err := cSess.CstpPickCmp.Compress(pl.Data, (*dst)[8:])
				if err == nil && size < len(pl.Data) {
					copy((*dst)[:8], plHeader)
					binary.BigEndian.PutUint16((*dst)[4:6], uint16(size))
					(*dst)[6] = 0x08
					pl.Data = append(pl.Data[:0], (*dst)[:size+8]...)
					isCompress = true
				}
				putByte(dst)
			}
			if !isCompress {
				// Get data length
				l := len(pl.Data)
				// Expand capacity first +8
				pl.Data = pl.Data[:l+8]
				// Data move back
				copy(pl.Data[8:], pl.Data)
				// Add header information
				copy(pl.Data[:8], plHeader)
				// Update header length
				binary.BigEndian.PutUint16(pl.Data[4:6], uint16(l))
			}
		} else {
			pl.Data = append(pl.Data[:0], plHeader...)
			// Set header type
			pl.Data[6] = pl.PType
		}

		n, err = conn.Write(pl.Data)
		if err != nil {
			base.Error("write err", cSess.Username, err)
			return
		}

		putPayload(pl)

		// Current limiting settings
		err = cSess.RateLimit(n, false)
		if err != nil {
			base.Error(err)
		}
	}
}
