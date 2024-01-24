package admin

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/cherts/anylink/dbdata"
	"github.com/cherts/anylink/pkg/utils"
	mapset "github.com/deckarep/golang-set"
	"github.com/spf13/cast"
	"github.com/xuri/excelize/v2"
)

func UserUpload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(8 << 20)
	file, header, err := r.FormFile("file")
	if err != nil || !strings.Contains(header.Filename, ".xlsx") || !strings.Contains(header.Filename, ".xls") {
		RespError(w, RespInternalErr, "File parsing failed: only xlsx or xls files supported")
		return
	}
	defer file.Close()

	// go/path-injection
	// base.Cfg.FilesPath It can be accessed directly from the outside, but files cannot be uploaded here.
	fileName := path.Join(os.TempDir(), utils.RandomRunes(10))
	newFile, err := os.Create(fileName)
	if err != nil {
		RespError(w, RespInternalErr, "Failed to create file:", err)
		return
	}
	defer newFile.Close()

	io.Copy(newFile, file)
	if err = UploadUser(newFile.Name()); err != nil {
		RespError(w, RespInternalErr, err)
		os.Remove(fileName)
		return
	}
	os.Remove(fileName)
	RespSucess(w, "Batch added successfully")
}

func UploadUser(file string) error {
	f, err := excelize.OpenFile(file)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			return
		}
	}()
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return err
	}
	if rows[0][0] != "id" || rows[0][1] != "username" || rows[0][2] != "nickname" || rows[0][3] != "email" || rows[0][4] != "pin_code" || rows[0][5] != "limittime" || rows[0][6] != "otp_secret" || rows[0][7] != "disable_otp" || rows[0][8] != "groups" || rows[0][9] != "status" || rows[0][10] != "send_email" {
		return fmt.Errorf("Batch addition failed, the table format is incorrect")
	}
	var k []interface{}
	for _, v := range dbdata.GetGroupNames() {
		k = append(k, v)
	}
	for index, row := range rows {
		if index == 0 {
			continue
		}
		id, _ := strconv.Atoi(row[0])
		if len(row[4]) < 6 {
			row[4] = utils.RandomRunes(8)
		}
		limittime, _ := time.ParseInLocation("2006-01-02 15:04:05", row[5], time.Local)
		disableOtp, _ := strconv.ParseBool(row[7])
		var group []string
		if row[8] == "" {
			return fmt.Errorf("Data error in line %d, user group is not allowed to be empty", index)
		}
		for _, v := range strings.Split(row[8], ",") {
			if s := mapset.NewSetFromSlice(k); s.Contains(v) {
				group = append(group, v)
			} else {
				return fmt.Errorf("User group [%s] does not exist, please check the data in row %d", v, index)
			}
		}
		status := cast.ToInt8(row[9])
		sendmail, _ := strconv.ParseBool(row[10])
		// createdAt, _ := time.ParseInLocation("2006-01-02 15:04:05", row[11], time.Local)
		// updatedAt, _ := time.ParseInLocation("2006-01-02 15:04:05", row[12], time.Local)
		user := &dbdata.User{
			Id:         id,
			Username:   row[1],
			Nickname:   row[2],
			Email:      row[3],
			PinCode:    row[4],
			LimitTime:  &limittime,
			OtpSecret:  row[6],
			DisableOtp: disableOtp,
			Groups:     group,
			Status:     status,
			SendEmail:  sendmail,
			// CreatedAt:  createdAt,
			// UpdatedAt:  updatedAt,
		}
		if err := dbdata.AddBatch(user); err != nil {
			return fmt.Errorf("Please check whether the data in row %d contains duplicate users.", index)
		}
		if user.SendEmail {
			if err := userAccountMail(user); err != nil {
				return err
			}
		}
	}
	return nil
}
