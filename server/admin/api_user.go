package admin

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/cherts/anylink/base"
	"github.com/cherts/anylink/dbdata"
	"github.com/cherts/anylink/pkg/utils"
	"github.com/cherts/anylink/sessdata"
	"github.com/skip2/go-qrcode"
	mail "github.com/xhit/go-simple-mail/v2"
)

func UserList(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	prefix := r.FormValue("prefix")
	prefix = strings.TrimSpace(prefix)
	pageS := r.FormValue("page")
	page, _ := strconv.Atoi(pageS)
	if page < 1 {
		page = 1
	}

	var (
		pageSize = dbdata.PageSize
		count    int
		datas    []dbdata.User
		err      error
	)

	// Query prefix matching
	if len(prefix) > 0 {
		fuzzy := "%" + prefix + "%"
		where := "username LIKE ? OR nickname LIKE ? OR email LIKE ?"

		count = dbdata.FindWhereCount(&dbdata.User{}, where, fuzzy, fuzzy, fuzzy)
		err = dbdata.FindWhere(&datas, pageSize, page, where, fuzzy, fuzzy, fuzzy)
	} else {
		count = dbdata.CountAll(&dbdata.User{})
		err = dbdata.Find(&datas, pageSize, page)
	}

	if err != nil && !dbdata.CheckErrNotFound(err) {
		RespError(w, RespInternalErr, err)
		return
	}

	data := map[string]interface{}{
		"count":     count,
		"page_size": pageSize,
		"datas":     datas,
	}

	RespSucess(w, data)
}

func UserDetail(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	idS := r.FormValue("id")
	id, _ := strconv.Atoi(idS)
	if id < 1 {
		RespError(w, RespParamErr, "username error")
		return
	}

	var user dbdata.User
	err := dbdata.One("Id", id, &user)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}

	RespSucess(w, user)
}

func UserSet(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	defer r.Body.Close()
	data := &dbdata.User{}
	err = json.Unmarshal(body, data)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}

	if len(data.PinCode) < 6 {
		data.PinCode = utils.RandomRunes(8)
		base.Info("User: ", data.Username, "Random password is: ", data.PinCode)
	}
	plainpwd := data.PinCode
	err = dbdata.SetUser(data)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	data.PinCode = plainpwd

	// send email
	if data.SendEmail {
		err = userAccountMail(data)
		if err != nil {
			RespError(w, RespInternalErr, err)
			return
		}
	}
	//Perform expired user detection after modifying user information
	sessdata.CloseUserLimittimeSession()
	RespSucess(w, nil)
}

func UserDel(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	idS := r.FormValue("id")
	id, _ := strconv.Atoi(idS)

	if id < 1 {
		RespError(w, RespParamErr, "Wrong user id")
		return
	}

	user := dbdata.User{Id: id}
	err := dbdata.Del(&user)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	RespSucess(w, nil)
}

func UserOtpQr(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	b64S := r.FormValue("b64")
	idS := r.FormValue("id")
	id, _ := strconv.Atoi(idS)

	var b64 bool
	if b64S == "1" {
		b64 = true
	}
	data, err := userOtpQr(id, b64)
	if err != nil {
		base.Error(err)
	}
	io.WriteString(w, data)
}

func userOtpQr(uid int, b64 bool) (string, error) {
	var user dbdata.User
	err := dbdata.One("Id", uid, &user)
	if err != nil {
		return "", err
	}

	issuer := url.QueryEscape(base.Cfg.Issuer)
	qrstr := fmt.Sprintf("otpauth://totp/%s:%s?issuer=%s&secret=%s", issuer, user.Email, issuer, user.OtpSecret)
	qr, _ := qrcode.New(qrstr, qrcode.High)

	if b64 {
		data, err := qr.PNG(300)
		if err != nil {
			return "", err
		}
		s := base64.StdEncoding.EncodeToString(data)
		return s, nil
	}

	buf := bytes.NewBuffer(nil)
	err = qr.Write(300, buf)
	return buf.String(), err
}

// online user
func UserOnline(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	search_cate := r.FormValue("search_cate")
	search_text := r.FormValue("search_text")
	show_sleeper := r.FormValue("show_sleeper")
	showSleeper, _ := strconv.ParseBool(show_sleeper)
	// one_offline := r.FormValue("one_offline")

	// datas := sessdata.OnlineSess()
	datas := sessdata.GetOnlineSess(search_cate, search_text, showSleeper)

	data := map[string]interface{}{
		"count":     len(datas),
		"page_size": dbdata.PageSize,
		"datas":     datas,
	}

	RespSucess(w, data)
}

func UserOffline(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	token := r.FormValue("token")
	sessdata.CloseSess(token, dbdata.UserLogoutAdmin)
	RespSucess(w, nil)
}

func UserReline(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	token := r.FormValue("token")
	sessdata.CloseCSess(token)
	RespSucess(w, nil)
}

type userAccountMailData struct {
	Issuer       string
	LinkAddr     string
	Group        string
	Username     string
	Nickname     string
	PinCode      string
	LimitTime    string
	OtpImg       string
	OtpImgBase64 string
	DisableOtp   bool
}

func userAccountMail(user *dbdata.User) error {
	// Platform notification
	htmlBody := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
    <title>Hello AnyLink!</title>
</head>
<body>
%s
</body>
</html>
`
	dataOther := &dbdata.SettingOther{}
	err := dbdata.SettingGet(dataOther)
	if err != nil {
		base.Error(err)
		return err
	}
	htmlBody = fmt.Sprintf(htmlBody, dataOther.AccountMail)
	// fmt.Println(htmlBody)

	// The token is valid for 3 days
	expiresAt := time.Now().Unix() + 3600*24*3
	jwtData := map[string]interface{}{"id": user.Id}
	tokenString, err := SetJwtData(jwtData, expiresAt)
	if err != nil {
		return err
	}

	setting := &dbdata.SettingOther{}
	err = dbdata.SettingGet(setting)
	if err != nil {
		base.Error(err)
		return err
	}

	otpData, _ := userOtpQr(user.Id, true)

	data := userAccountMailData{
		Issuer:       base.Cfg.Issuer,
		LinkAddr:     setting.LinkAddr,
		Group:        strings.Join(user.Groups, ","),
		Username:     user.Username,
		Nickname:     user.Nickname,
		PinCode:      user.PinCode,
		OtpImg:       fmt.Sprintf("https://%s/otp_qr?id=%d&jwt=%s", setting.LinkAddr, user.Id, tokenString),
		OtpImgBase64: "data:image/png;base64," + otpData,
		DisableOtp:   user.DisableOtp,
	}

	if user.LimitTime == nil {
		data.LimitTime = "No restrictions"
	} else {
		data.LimitTime = user.LimitTime.Local().Format("2006-01-02")
	}

	w := bytes.NewBufferString("")
	t, _ := template.New("auth_complete").Parse(htmlBody)
	err = t.Execute(w, data)
	if err != nil {
		return err
	}
	// fmt.Println(w.String())

	var attach *mail.File
	if user.DisableOtp {
		attach = nil
	} else {
		imgData, _ := userOtpQr(user.Id, false)
		attach = &mail.File{
			MimeType: "image/png",
			Name:     "userOtpQr.png",
			Data:     []byte(imgData),
			Inline:   true,
		}
	}

	return SendMail(base.Cfg.Issuer, user.Email, w.String(), attach)
}
