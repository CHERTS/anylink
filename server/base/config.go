package base

const (
	cfgStr = iota
	cfgInt
	cfgBool

	defaultJwt = "abcdef.0123456789.abcdef"
	defaultPwd = "$2a$10$UQ7C.EoPifDeJh6d8.31TeSPQU7hM/NOM2nixmBucJpAuXDQNqNke"
)

type config struct {
	Typ     int
	Name    string
	Short   string
	Usage   string
	ValStr  string
	ValInt  int
	ValBool bool
}

var configs = []config{
	{Typ: cfgStr, Name: "conf", Usage: "Config file", ValStr: "./conf/server.toml", Short: "c"},
	{Typ: cfgStr, Name: "profile", Usage: "profile.xml file", ValStr: "./conf/profile.xml"},
	{Typ: cfgStr, Name: "profile_name", Usage: "Profile name (Used to distinguish configurations of different servers)", ValStr: "anylink"},
	{Typ: cfgStr, Name: "server_addr", Usage: "Server listening address", ValStr: ":443"},
	{Typ: cfgBool, Name: "server_dtls", Usage: "Enable DTLS", ValBool: false},
	{Typ: cfgStr, Name: "server_dtls_addr", Usage: "DTLS listening address", ValStr: ":443"},
	{Typ: cfgStr, Name: "advertise_dtls_addr", Usage: "DTLS external mapping port (same as server_dtls_addr if empty)", ValStr: ""},
	{Typ: cfgStr, Name: "admin_addr", Usage: "Admin listening address", ValStr: ":8800"},
	{Typ: cfgBool, Name: "proxy_protocol", Usage: "TCP proxy protocol", ValBool: false},
	{Typ: cfgStr, Name: "db_type", Usage: "Database type [sqlite3 mysql postgres]", ValStr: "sqlite3"},
	{Typ: cfgStr, Name: "db_source", Usage: "Database source", ValStr: "./conf/anylink.db"},
	{Typ: cfgStr, Name: "cert_file", Usage: "Certificate file", ValStr: "./conf/vpn_cert.pem"},
	{Typ: cfgStr, Name: "cert_key", Usage: "Certificate key", ValStr: "./conf/vpn_cert.key"},
	{Typ: cfgStr, Name: "files_path", Usage: "External download file path", ValStr: "./conf/files"},
	{Typ: cfgStr, Name: "log_path", Usage: "Log file path, default standard output", ValStr: ""},
	{Typ: cfgStr, Name: "log_level", Usage: "Log level [debug info warn error]", ValStr: "debug"},
	{Typ: cfgBool, Name: "http_server_log", Usage: "Turn on the log of go standard library http.Server", ValBool: false},
	{Typ: cfgBool, Name: "pprof", Usage: "Turn on pprof", ValBool: true},
	{Typ: cfgStr, Name: "issuer", Usage: "System name", ValStr: "XXX company VPN"},
	{Typ: cfgStr, Name: "admin_user", Usage: "Admin username", ValStr: "admin"},
	{Typ: cfgStr, Name: "admin_pass", Usage: "Admin password", ValStr: defaultPwd},
	{Typ: cfgStr, Name: "admin_otp", Usage: "Admin user OTP, generate commands ./anylink tool -o", ValStr: ""},
	{Typ: cfgStr, Name: "jwt_secret", Usage: "JWT key", ValStr: defaultJwt},
	{Typ: cfgStr, Name: "link_mode", Usage: "Virtual network type [tun tap macvtap ipvtap]", ValStr: "tun"},
	{Typ: cfgStr, Name: "ipv4_master", Usage: "IPv4 interface", ValStr: "eth0"},
	{Typ: cfgStr, Name: "ipv4_cidr", Usage: "IPv4 address", ValStr: "192.168.90.0/24"},
	{Typ: cfgStr, Name: "ipv4_gateway", Usage: "IPv4 gateway", ValStr: "192.168.90.1"},
	{Typ: cfgStr, Name: "ipv4_start", Usage: "IPv4 start address", ValStr: "192.168.90.100"},
	{Typ: cfgStr, Name: "ipv4_end", Usage: "IPv4 end address", ValStr: "192.168.90.200"},
	{Typ: cfgStr, Name: "default_group", Usage: "Default user group", ValStr: "one"},
	{Typ: cfgStr, Name: "default_domain", Usage: "Default search domain for client dns", ValStr: ""},

	{Typ: cfgInt, Name: "ip_lease", Usage: "IP lease period (seconds)", ValInt: 86400},
	{Typ: cfgInt, Name: "max_client", Usage: "Max user connections", ValInt: 200},
	{Typ: cfgInt, Name: "max_user_client", Usage: "Maximum single user connections", ValInt: 3},
	{Typ: cfgInt, Name: "cstp_keepalive", Usage: "Keepalive time (seconds)", ValInt: 3},
	{Typ: cfgInt, Name: "cstp_dpd", Usage: "Dead link detection time (seconds)", ValInt: 20},
	{Typ: cfgInt, Name: "mobile_keepalive", Usage: "Mobile terminal keepalive connection detection time (seconds)", ValInt: 4},
	{Typ: cfgInt, Name: "mobile_dpd", Usage: "Mobile terminal dead link detection time (seconds)", ValInt: 60},
	{Typ: cfgInt, Name: "mtu", Usage: "Maximum transmission unit MTU", ValInt: 1460},
	{Typ: cfgInt, Name: "idle_timeout", Usage: "Idle link timeout (seconds) - disconnect the link after timeout, 0 turns off this function", ValInt: 0},
	{Typ: cfgInt, Name: "session_timeout", Usage: "Session expiration time (seconds) - used for disconnection and reconnection, 0 will never expire", ValInt: 3600},
	// {Typ: cfgInt, Name: "auth_timeout", Usage: "auth_timeout", ValInt: 0},
	{Typ: cfgInt, Name: "audit_interval", Usage: "Audit deduplication interval (seconds), -1 turns off", ValInt: -1},

	{Typ: cfgBool, Name: "show_sql", Usage: "Display sql statements for debugging", ValBool: false},
	{Typ: cfgBool, Name: "iptables_nat", Usage: "Whether to automatically add NAT", ValBool: true},
	{Typ: cfgBool, Name: "compression", Usage: "Enable compression", ValBool: false},
	{Typ: cfgInt, Name: "no_compress_limit", Usage: "Below and equal to how many bytes are not compressed", ValInt: 256},

	{Typ: cfgBool, Name: "display_error", Usage: "The client displays detailed error information (be careful when opening the online environment)", ValBool: false},
	{Typ: cfgBool, Name: "exclude_export_ip", Usage: "Exclude export ip routing (export ip is not encrypted for transmission)", ValBool: true},
	{Typ: cfgBool, Name: "auth_alone_otp", Usage: "Log in to the separate verification OTP window", ValBool: false},
	{Typ: cfgBool, Name: "encryption_password", Usage: "Whether the user password is encrypted and saved", ValBool: false},

	{Typ: cfgBool, Name: "anti_brute_force", Usage: "Whether to enable the explosion-proof function", ValBool: true},
	{Typ: cfgStr, Name: "ip_whitelist", Usage: "Global IP whitelist, multiple IP whitelists separated by commas, supports single IP and CIDR range", ValStr: "192.168.90.1,172.16.0.0/24"},

	{Typ: cfgInt, Name: "max_ban_score", Usage: "The maximum number of attempts per unit time. 0 means turning off this function.", ValInt: 5},
	{Typ: cfgInt, Name: "ban_reset_time", Usage: "Set the unit time (seconds), if exceeded, reset the count", ValInt: 10},
	{Typ: cfgInt, Name: "lock_time", Usage: "Lockout duration after exceeding the maximum number of attempts (seconds)", ValInt: 300},

	{Typ: cfgInt, Name: "max_global_user_ban_count", Usage: "The maximum number of attempts per unit time for global users. 0 means turning off this function.能", ValInt: 20},
	{Typ: cfgInt, Name: "global_user_ban_reset_time", Usage: "Global user setting unit time (seconds)", ValInt: 600},
	{Typ: cfgInt, Name: "global_user_lock_time", Usage: "Global user lockout time (seconds)", ValInt: 300},

	{Typ: cfgInt, Name: "max_global_ip_ban_count", Usage: "The maximum number of attempts per unit time for the global IP address. 0 means this function is disabled.", ValInt: 40},
	{Typ: cfgInt, Name: "global_ip_ban_reset_time", Usage: "Global IP setting unit time (seconds)", ValInt: 1200},
	{Typ: cfgInt, Name: "global_ip_lock_time", Usage: "Global IP lock time (seconds)", ValInt: 300},

	{Typ: cfgInt, Name: "global_lock_state_expiration_time", Usage: "The global lock state preservation life cycle (seconds), if it exceeds, the record will be deleted", ValInt: 3600},
}

var envs = map[string]string{}
