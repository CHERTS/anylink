package admin

import (
	"errors"
	"net/http"

	"github.com/cherts/anylink/dbdata"
)

func StatsInfoList(w http.ResponseWriter, r *http.Request) {
	var ok bool
	_ = r.ParseForm()
	action := r.FormValue("action")
	scope := r.FormValue("scope")
	ok = dbdata.StatsInfoIns.ValidAction(action)
	if !ok {
		RespError(w, RespParamErr, errors.New("Chart category does not exist"))
		return
	}
	ok = dbdata.StatsInfoIns.ValidScope(scope)
	if !ok {
		RespError(w, RespParamErr, errors.New("Date range does not exist"))
		return
	}
	datas, err := dbdata.StatsInfoIns.GetData(action, scope)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	data := make(map[string]interface{})
	data["datas"] = datas
	RespSucess(w, data)
}
