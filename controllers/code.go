package controllers

type ResCode int64

const (
	CodeSuccess ResCode = 1000 + iota
	CodeInvalidParam
	CodeUserExist
	CodeUserNoteExist
	CodeInvalidPassword
	CodeServeBusy

	CodeInvalidToken
	CodeNeedLogin
	CodePostNotExist
	CodeNoPermission
)

var codeMsgMap = map[ResCode]string{
	CodeSuccess:         "success",
	CodeInvalidParam:    "请求参数错误",
	CodeUserExist:       "用户名已存在",
	CodeUserNoteExist:   "用户名不存在",
	CodeInvalidPassword: "用户名或密码错误",
	// 技巧：比如数据库连接错误，直接返回一个服务繁忙，不需要给前端具体信息，具体信息仅在后端记录到日志
	CodeServeBusy: "服务繁忙",

	CodeInvalidToken: "无效的token",
	CodeNeedLogin:    "需要登录",
}

func (c ResCode) Msg() string {
	msg, ok := codeMsgMap[c]
	if !ok {
		msg = codeMsgMap[CodeServeBusy]
	}
	return msg
}
