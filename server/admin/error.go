package admin

// return code
const (
	RespSuccess       = 0
	RespInternalErr   = 1
	RespTokenErr      = 2
	RespUserOrPassErr = 3
	RespParamErr      = 4
)

var RespMap = map[int]string{
	RespTokenErr:      "Client TOKEN error",
	RespUserOrPassErr: "Wrong user name or password",
}
