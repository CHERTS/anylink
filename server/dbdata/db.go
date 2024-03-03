package dbdata

import (
	"time"

	"github.com/cherts/anylink/base"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"xorm.io/xorm"
)

var (
	xdb *xorm.Engine
)

func GetXdb() *xorm.Engine {
	return xdb
}

func initDb() {
	var err error
	xdb, err = xorm.NewEngine(base.Cfg.DbType, base.Cfg.DbSource)
	// Initialize xorm time zone
	xdb.DatabaseTZ = time.Local
	xdb.TZLocation = time.Local
	if err != nil {
		base.Fatal(err)
	}

	if base.Cfg.ShowSQL {
		xdb.ShowSQL(true)
	}

	// Initialize database
	err = xdb.Sync2(&User{}, &Setting{}, &Group{}, &IpMap{}, &AccessAudit{}, &Policy{}, &StatsNetwork{}, &StatsCpu{}, &StatsMem{}, &StatsOnline{}, &UserActLog{})
	if err != nil {
		base.Fatal(err)
	}

	// fmt.Println("s1=============", err)
}

func initData() {
	var (
		err error
	)

	// Determine whether to use it for the first time
	install := &SettingInstall{}
	err = SettingGet(install)

	if err == nil && install.Installed {
		// Already installed
		return
	}

	// An error occurred
	if err != ErrNotFound {
		base.Fatal(err)
	}

	err = addInitData()
	if err != nil {
		base.Fatal(err)
	}

}

func addInitData() error {
	var (
		err error
	)

	sess := xdb.NewSession()
	defer sess.Close()

	err = sess.Begin()
	if err != nil {
		return err
	}

	// SettingSmtp
	smtp := &SettingSmtp{
		Host:       "127.0.0.1",
		Port:       25,
		From:       "vpn@xxx.com",
		Encryption: "None",
	}
	err = SettingSessAdd(sess, smtp)
	if err != nil {
		return err
	}

	// SettingAuditLog
	auditLog := SettingGetAuditLogDefault()
	err = SettingSessAdd(sess, auditLog)
	if err != nil {
		return err
	}

	// SettingDnsProvider
	provider := &SettingLetsEncrypt{
		Domain:      "vpn.xxx.com",
		Legomail:    "legomail",
		Name:        "aliyun",
		Renew:       false,
		DNSProvider: DNSProvider{},
	}
	err = SettingSessAdd(sess, provider)
	if err != nil {
		return err
	}
	// LegoUser
	legouser := &LegoUserData{}
	err = SettingSessAdd(sess, legouser)
	if err != nil {
		return err
	}
	// SettingOther
	other := &SettingOther{
		LinkAddr:    "vpn.xxx.com",
		Banner:      "You have connected to the company network, please use it in accordance with company regulations.\nPlease do not perform non-work downloading and video activities!",
		Homeindex:   "AnyLink is an enterprise-level remote office sslvpn software that can support multiple people using it online at the same time.",
		AccountMail: accountMail,
	}
	err = SettingSessAdd(sess, other)
	if err != nil {
		return err
	}

	// Install
	install := &SettingInstall{Installed: true}
	err = SettingSessAdd(sess, install)
	if err != nil {
		return err
	}

	err = sess.Commit()
	if err != nil {
		return err
	}

	g1 := Group{
		Name:         "ops",
		AllowLan:     true,
		ClientDns:    []ValData{{Val: "114.114.114.114"}},
		RouteInclude: []ValData{{Val: All}},
		Status:       1,
	}
	err = SetGroup(&g1)
	if err != nil {
		return err
	}

	return nil
}

func CheckErrNotFound(err error) bool {
	return err == ErrNotFound
}

const accountMail = `<p>Hello, {{.Issuer}}:</p>
<p>&nbsp;&nbsp;Your account has been created and activated.</p>
<p>
    Login address: <b>{{.LinkAddr}}</b> <br/>
    User group: <b>{{.Group}}</b> <br/>
    Username: <b>{{.Username}}</b> <br/>
    PIN code: <b>{{.PinCode}}</b> <br/>
    <!-- 
    User dynamic code (expires after 3 days):<br/>
    <img src="{{.OtpImg}}"/>
    -->
    User OTP code (please save it):<br/>
    <img src="{{.OtpImgBase64}}"/>
</p>
<div>
    Instructions for use:
    <ul>
        <li>Please use OTP software to scan the dynamic code QR code</li>
        <li>Then use anyconnect client to log in</li>
        <li>The login password is [PIN code + dynamic code] (there is no + sign in the middle)</li>
    </ul>
</div>
<p>
    Software download address: https://{{.LinkAddr}}/files/info.txt
</p>`
