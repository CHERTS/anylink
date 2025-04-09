package handler

import (
	"fmt"
	"io"
	"net"

	"github.com/cherts/anylink/base"
	"github.com/cherts/anylink/pkg/arpdis"
	"github.com/cherts/anylink/sessdata"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	gosysctl "github.com/lorenzosaino/go-sysctl"
	"github.com/songgao/packets/ethernet"
	"github.com/songgao/water"
	"github.com/songgao/water/waterutil"
)

const bridgeName = "anylink0"

var (
	// Gateway mac address
	gatewayHw net.HardwareAddr
)

type LinkDriver interface {
	io.ReadWriteCloser
	Name() string
}

func _setGateway() {
	dstAddr := arpdis.Lookup(sessdata.IpPool.Ipv4Gateway, false)
	gatewayHw = dstAddr.HardwareAddr
	// Set to static address mapping
	dstAddr.Type = arpdis.TypeStatic
	arpdis.Add(dstAddr)
}

func _checkTapIp(ifName string) {
	iFace, err := net.InterfaceByName(ifName)
	if err != nil {
		base.Fatal("testTap err: ", err)
	}

	var ifIp net.IP

	addrs, err := iFace.Addrs()
	if err != nil {
		base.Fatal("testTap err: ", err)
	}
	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil || ip.To4() == nil {
			continue
		}
		ifIp = ip
	}

	if !sessdata.IpPool.Ipv4IPNet.Contains(ifIp) {
		base.Fatal("tapIp or Ip network err")
	}
}

func checkTap() {
	_setGateway()
	_checkTapIp(bridgeName)
}

// Create tap network card
func LinkTap(cSess *sessdata.ConnSession) error {
	cfg := water.Config{
		DeviceType: water.TAP,
	}

	ifce, err := water.New(cfg)
	if err != nil {
		base.Error(err)
		return err
	}

	cSess.SetIfName(ifce.Name())

	cmdstr1 := fmt.Sprintf("ip link set dev %s up mtu %d multicast on", ifce.Name(), cSess.Mtu)
	cmdstr2 := fmt.Sprintf("ip link set dev %s master %s", ifce.Name(), bridgeName)
	err = execCmd([]string{cmdstr1, cmdstr2})
	if err != nil {
		base.Error(err)
		_ = ifce.Close()
		return err
	}

	// cmdstr3 := fmt.Sprintf("sysctl -w net.ipv6.conf.%s.disable_ipv6=1", ifce.Name())
	// execCmd([]string{cmdstr3})
	err = gosysctl.Set(fmt.Sprintf("net.ipv6.conf.%s.disable_ipv6", ifce.Name()), "1")
	if err != nil {
		base.Warn(err)
	}

	go allTapRead(ifce, cSess)
	go allTapWrite(ifce, cSess)
	return nil
}

// ========================universal code===========================

func allTapWrite(ifce LinkDriver, cSess *sessdata.ConnSession) {
	defer func() {
		base.Debug("LinkTap return", cSess.IpAddr)
		cSess.Close()
		ifce.Close()
	}()

	var (
		err   error
		dstHw net.HardwareAddr
		pl    *sessdata.Payload
		frame = make(ethernet.Frame, BufferSize)
		ipDst = net.IPv4(1, 2, 3, 4)
	)

	for {
		frame.Resize(BufferSize)

		select {
		case pl = <-cSess.PayloadIn:
		case <-cSess.CloseChan:
			return
		}

		// var frame ethernet.Frame
		switch pl.LType {
		default:
			// log.Println(payload)
		case sessdata.LTypeEthernet:
			copy(frame, pl.Data)
			frame = frame[:len(pl.Data)]

			// packet := gopacket.NewPacket(frame, layers.LayerTypeEthernet, gopacket.Default)
			// fmt.Println("wirteArp:", packet)
		case sessdata.LTypeIPData: // Need to be converted to Ethernet data
			ipSrc := waterutil.IPv4Source(pl.Data)
			if !ipSrc.Equal(cSess.IpAddr) {
				// If the IP address is not assigned to the client, it will be discarded directly.
				continue
			}

			if waterutil.IsIPv6(pl.Data) {
				// Filter out IPv6 data
				continue
			}

			// packet := gopacket.NewPacket(pl.Data, layers.LayerTypeIPv4, gopacket.Default)
			// fmt.Println("get:", packet)

			// Set ipv4 address manually
			ipDst[12] = pl.Data[16]
			ipDst[13] = pl.Data[17]
			ipDst[14] = pl.Data[18]
			ipDst[15] = pl.Data[19]

			dstHw = gatewayHw
			if sessdata.IpPool.Ipv4IPNet.Contains(ipDst) {
				dstAddr := arpdis.Lookup(ipDst, true)
				// fmt.Println("dstAddr", dstAddr)
				if dstAddr != nil {
					dstHw = dstAddr.HardwareAddr
				}
			}

			// fmt.Println("Gateway", ipSrc, ipDst, dstHw)
			frame.Prepare(dstHw, cSess.MacHw, ethernet.NotTagged, ethernet.IPv4, len(pl.Data))
			copy(frame[12+2:], pl.Data)
		}

		// packet := gopacket.NewPacket(frame, layers.LayerTypeEthernet, gopacket.Default)
		// fmt.Println("write:", packet)
		_, err = ifce.Write(frame)
		if err != nil {
			base.Error("tap Write err", err)
			return
		}

		putPayloadInBefore(cSess, pl)
	}
}

func allTapRead(ifce LinkDriver, cSess *sessdata.ConnSession) {
	defer func() {
		base.Debug("tapRead return", cSess.IpAddr)
		ifce.Close()
	}()

	var (
		err   error
		n     int
		data  []byte
		frame = make(ethernet.Frame, BufferSize)
	)

	for {
		frame.Resize(BufferSize)

		n, err = ifce.Read(frame)
		if err != nil {
			base.Error("tap Read err", n, err)
			return
		}
		frame = frame[:n]

		switch frame.Ethertype() {
		default:
			continue
		case ethernet.IPv6:
			continue
		case ethernet.IPv4:
			// Send IP data
			data = frame.Payload()

			ip_dst := waterutil.IPv4Destination(data)
			if !ip_dst.Equal(cSess.IpAddr) {
				// Filter non-native addresses
				// log.Println(ip_dst, sess.Ip)
				continue
			}

			// packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)
			// fmt.Println("put:", packet)

			pl := getPayload()
			// Copy data to pl
			copy(pl.Data, data)
			// Update slice length
			pl.Data = pl.Data[:len(data)]
			if payloadOut(cSess, pl) {
				return
			}

		case ethernet.ARP:
			// Currently only the ARP protocol is implemented
			packet := gopacket.NewPacket(frame, layers.LayerTypeEthernet, gopacket.NoCopy)
			layer := packet.Layer(layers.LayerTypeARP)
			arpReq := layer.(*layers.ARP)

			if !cSess.IpAddr.Equal(arpReq.DstProtAddress) {
				// Filter non-native addresses
				continue
			}

			// fmt.Println("arp", time.Now(), net.IP(arpReq.SourceProtAddress), cSess.IpAddr)
			// fmt.Println(packet)

			// Return ARP data
			src := &arpdis.Addr{IP: cSess.IpAddr, HardwareAddr: cSess.MacHw}
			dst := &arpdis.Addr{IP: arpReq.SourceProtAddress, HardwareAddr: arpReq.SourceHwAddress}
			data, err = arpdis.NewARPReply(src, dst)
			if err != nil {
				base.Error(err)
				return
			}

			// Add arp address from accepted arp information
			addr := &arpdis.Addr{
				IP:           append([]byte{}, dst.IP...),
				HardwareAddr: append([]byte{}, dst.HardwareAddr...),
			}
			arpdis.Add(addr)

			pl := getPayload()
			// Set to layer 2 data type
			pl.LType = sessdata.LTypeEthernet
			// Copy data to pl
			copy(pl.Data, data)
			// Update slice length
			pl.Data = pl.Data[:len(data)]

			if payloadIn(cSess, pl) {
				return
			}

		}
	}
}
