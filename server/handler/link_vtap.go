package handler

import (
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"github.com/cherts/anylink/base"
	"github.com/cherts/anylink/pkg/utils"
	"github.com/cherts/anylink/sessdata"
)

// link vtap
const vTapPrefix = "lvtap"

type Vtap struct {
	*os.File
	ifName string
}

func (v *Vtap) Close() error {
	v.File.Close()
	cmdstr := fmt.Sprintf("ip link del %s", v.ifName)
	return execCmd([]string{cmdstr})
}

func checkMacvtap() {
	// Load macvtap
	base.CheckModOrLoad("macvtap")

	_setGateway()
	_checkTapIp(base.Cfg.Ipv4Master)

	ifName := "anylinkMacvtap"

	// Enable promiscuous mode for the primary network card
	cmdstr1 := fmt.Sprintf("ip link set dev %s promisc on", base.Cfg.Ipv4Master)
	// Test macvtap functionality
	cmdstr2 := fmt.Sprintf("ip link add link %s name %s type macvtap mode bridge", base.Cfg.Ipv4Master, ifName)
	cmdstr3 := fmt.Sprintf("ip link del %s", ifName)
	err := execCmd([]string{cmdstr1, cmdstr2, cmdstr3})
	if err != nil {
		base.Fatal(err)
	}
}

// Create Macvtap network card
func LinkMacvtap(cSess *sessdata.ConnSession) error {
	capL := sessdata.IpPool.IpLongMax - sessdata.IpPool.IpLongMin
	ipN := utils.Ip2long(cSess.IpAddr) % capL
	ifName := fmt.Sprintf("%s%d", vTapPrefix, ipN)

	cSess.SetIfName(ifName)

	cmdstr1 := fmt.Sprintf("ip link add link %s name %s type macvtap mode bridge", base.Cfg.Ipv4Master, ifName)
	alias := utils.ParseName(cSess.Group.Name + "." + cSess.Username)
	cmdstr2 := fmt.Sprintf("ip link set dev %s up mtu %d address %s alias %s", ifName, cSess.Mtu, cSess.MacHw, alias)

	err := execCmd([]string{cmdstr1, cmdstr2})
	if err != nil {
		base.Error(err)
		return err
	}
	cmdstr3 := fmt.Sprintf("sysctl -w net.ipv6.conf.%s.disable_ipv6=1", ifName)
	execCmd([]string{cmdstr3})

	return createVtap(cSess, ifName)
}

// func checkIpvtap() {

// }

// Create Ipvtap network card
func LinkIpvtap(cSess *sessdata.ConnSession) error {
	return nil
}

type ifReq struct {
	Name  [0x10]byte
	Flags uint16
	pad   [0x28 - 0x10 - 2]byte
}

func createVtap(cSess *sessdata.ConnSession, ifName string) error {
	// Initialize ifName
	inf, err := net.InterfaceByName(ifName)
	if err != nil {
		base.Error(err)
		return err
	}

	tName := fmt.Sprintf("/dev/tap%d", inf.Index)

	var fdInt int

	fdInt, err = syscall.Open(tName, syscall.O_RDWR|syscall.O_NONBLOCK, 0)
	if err != nil {
		return err
	}

	var flags uint16 = syscall.IFF_TAP | syscall.IFF_NO_PI
	var req ifReq
	req.Flags = flags

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fdInt),
		uintptr(syscall.TUNSETIFF),
		uintptr(unsafe.Pointer(&req)),
	)
	if errno != 0 {
		return os.NewSyscallError("ioctl", errno)
	}

	file := os.NewFile(uintptr(fdInt), tName)
	ifce := &Vtap{file, ifName}

	go allTapRead(ifce, cSess)
	go allTapWrite(ifce, cSess)
	return nil
}

// Destroy unclosed vtap
func destroyVtap() {
	its, err := net.Interfaces()
	if err != nil {
		base.Error(err)
		return
	}
	for _, v := range its {
		if strings.HasPrefix(v.Name, vTapPrefix) {
			// Delete the original network card
			cmdstr := fmt.Sprintf("ip link del %s", v.Name)
			execCmd([]string{cmdstr})
		}
	}
}
