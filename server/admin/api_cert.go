package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cherts/anylink/base"
	"github.com/cherts/anylink/dbdata"
)

func CustomCert(w http.ResponseWriter, r *http.Request) {
	cert, _, err := r.FormFile("cert")
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	key, _, err := r.FormFile("key")
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	certFile, err := os.OpenFile(base.Cfg.CertFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	defer certFile.Close()
	if _, err := io.Copy(certFile, cert); err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	keyFile, err := os.OpenFile(base.Cfg.CertKey, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	defer keyFile.Close()
	if _, err := io.Copy(keyFile, key); err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	if tlscert, _, err := dbdata.ParseCert(); err != nil {
		RespError(w, RespInternalErr, fmt.Sprintf("The certificate is invalid, please re-upload:%v", err))
		return
	} else {
		dbdata.LoadCertificate(tlscert)
	}
	RespSucess(w, "Upload successful")
}
func GetCertSetting(w http.ResponseWriter, r *http.Request) {
	sess := dbdata.GetXdb().NewSession()
	defer sess.Close()
	data := &dbdata.SettingLetsEncrypt{}
	if err := dbdata.SettingGet(data); err != nil {
		dbdata.SettingSessAdd(sess, data)
		RespError(w, RespInternalErr, err)
	}
	userData := &dbdata.LegoUserData{}
	if err := dbdata.SettingGet(userData); err != nil {
		dbdata.SettingSessAdd(sess, userData)
	}
	RespSucess(w, data)
}
func CreatCert(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	defer r.Body.Close()
	config := &dbdata.SettingLetsEncrypt{}
	if err := json.Unmarshal(body, config); err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	if err := dbdata.SettingSet(config); err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	client := dbdata.LeGoClient{}
	if err := client.NewClient(config); err != nil {
		base.Error(err)
		RespError(w, RespInternalErr, fmt.Sprintf("Failed to obtain certificate: %v", err))
		return
	}
	if err := client.GetCert(config.Domain); err != nil {
		base.Error(err)
		RespError(w, RespInternalErr, fmt.Sprintf("Failed to obtain certificate: %v", err))
		return
	}
	RespSucess(w, "Certificate generation successful")
}
