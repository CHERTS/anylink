package handler

import (
	"fmt"

	"github.com/cherts/anylink/base"
	"github.com/cherts/anylink/pkg/utils"
	"github.com/cherts/anylink/sessdata"
	"github.com/coreos/go-iptables/iptables"
	gosysctl "github.com/lorenzosaino/go-sysctl"
	"github.com/songgao/water"
)

func checkTun() {
	//Test ip command
	base.CheckModOrLoad("tun")

	// Test tun
	cfg := water.Config{
		DeviceType: water.TUN,
	}

	ifce, err := water.New(cfg)
	if err != nil {
		base.Fatal("open tun err: ", err)
	}
	defer ifce.Close()

	cmdstr1 := fmt.Sprintf("ip link set dev %s up mtu %s multicast off", ifce.Name(), "1399")
	err = execCmd([]string{cmdstr1})
	if err != nil {
		base.Fatal("testTun err: ", err)
	}
	// Enable server forwarding
	// err = execCmd([]string{"sysctl -w net.ipv4.ip_forward=1"})
	// if err != nil {
	// 	base.Fatal(err)
	// }
	if base.Cfg.IptablesNat {
		// Add NAT forwarding rules
		ipt, err := iptables.New()
		if err != nil {
			base.Fatal(err)
			return
		}

		// Fix rockyos nat not taking effect
		base.CheckModOrLoad("iptable_filter")
		base.CheckModOrLoad("iptable_nat")
		// base.CheckModOrLoad("xt_comment")

		// add notes
		natRule := []string{"-s", base.Cfg.Ipv4CIDR, "-o", base.Cfg.Ipv4Master, "-m", "comment",
			"--comment", "AnyLink", "-j", "MASQUERADE"}
		if base.InContainer {
			natRule = []string{"-s", base.Cfg.Ipv4CIDR, "-o", base.Cfg.Ipv4Master, "-j", "MASQUERADE"}
		}
		err = ipt.InsertUnique("nat", "POSTROUTING", 1, natRule...)
		if err != nil {
			base.Error(err)
		}
		// add notes
		forwardRule := []string{"-m", "comment", "--comment", "AnyLink", "-j", "ACCEPT"}
		if base.InContainer {
			forwardRule = []string{"-j", "ACCEPT"}
		}
		err = ipt.InsertUnique("filter", "FORWARD", 1, forwardRule...)
		if err != nil {
			base.Error(err)
		}

		base.Info(ipt.List("nat", "POSTROUTING"))
		base.Info(ipt.List("filter", "FORWARD"))
	}
}

// Create tun network card
func LinkTun(cSess *sessdata.ConnSession) error {
	cfg := water.Config{
		DeviceType: water.TUN,
	}

	ifce, err := water.New(cfg)
	if err != nil {
		base.Error(err)
		return err
	}
	// log.Printf("Interface Name: %s\n", ifce.Name())
	cSess.SetIfName(ifce.Name())

	// View alias information through ip link show

	alias := utils.ParseName(cSess.Group.Name + "." + cSess.Username)
	cmdstr1 := fmt.Sprintf("ip link set dev %s up mtu %d multicast off alias %s", ifce.Name(), cSess.Mtu, alias)
	cmdstr2 := fmt.Sprintf("ip addr add dev %s local %s peer %s/32",
		ifce.Name(), base.Cfg.Ipv4Gateway, cSess.IpAddr)
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

	go tunRead(ifce, cSess)
	go tunWrite(ifce, cSess)
	return nil
}

func tunWrite(ifce *water.Interface, cSess *sessdata.ConnSession) {
	defer func() {
		base.Debug("LinkTun return", cSess.IpAddr)
		cSess.Close()
		_ = ifce.Close()
	}()

	var (
		err error
		pl  *sessdata.Payload
	)

	for {
		select {
		case pl = <-cSess.PayloadIn:
		case <-cSess.CloseChan:
			return
		}

		_, err = ifce.Write(pl.Data)
		if err != nil {
			base.Error("tun Write err", err)
			return
		}

		putPayloadInBefore(cSess, pl)
	}
}

func tunRead(ifce *water.Interface, cSess *sessdata.ConnSession) {
	defer func() {
		base.Debug("tunRead return", cSess.IpAddr)
		_ = ifce.Close()
	}()
	var (
		err error
		n   int
	)

	for {
		// data := make([]byte, BufferSize)
		pl := getPayload()
		n, err = ifce.Read(pl.Data)
		if err != nil {
			base.Error("tun Read err", n, err)
			return
		}

		// Update data length
		pl.Data = (pl.Data)[:n]

		// data = data[:n]
		// ip_src := waterutil.IPv4Source(data)
		// ip_dst := waterutil.IPv4Destination(data)
		// ip_port := waterutil.IPv4DestinationPort(data)
		// fmt.Println("sent:", ip_src, ip_dst, ip_port)
		// packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)
		// fmt.Println("read:", packet)

		if payloadOut(cSess, pl) {
			return
		}
	}
}
