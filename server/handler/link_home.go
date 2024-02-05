package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cherts/anylink/admin"
	"github.com/cherts/anylink/dbdata"
)

func LinkHome(w http.ResponseWriter, r *http.Request) {
	// fmt.Println(r.RemoteAddr)
	// hu, _ := httputil.DumpRequest(r, true)
	// fmt.Println("DumpHome: ", string(hu))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Del("X-Aggregate-Auth")

	connection := strings.ToLower(r.Header.Get("Connection"))
	userAgent := strings.ToLower(r.UserAgent())
	if connection == "close" && (strings.Contains(userAgent, "anyconnect") || strings.Contains(userAgent, "openconnect")) {
		w.Header().Set("Connection", "close")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	index := &dbdata.SettingOther{}
	if err := dbdata.SettingGet(index); err != nil {
		return
	}
	w.WriteHeader(http.StatusOK)
	if index.Homeindex == "" {
		index.Homeindex = "AnyLink is an enterprise-level remote office SSL VPN software that can support multiple people online at the same time."
	}
	fmt.Fprintln(w, index.Homeindex)
}

func LinkOtpQr(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")

	_ = r.ParseForm()
	idS := r.FormValue("id")
	jwtToken := r.FormValue("jwt")
	data, err := admin.GetJwtData(jwtToken)
	if err != nil || idS != fmt.Sprint(data["id"]) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	admin.UserOtpQr(w, r)
}
