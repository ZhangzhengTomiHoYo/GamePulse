package pgsql

import "errors"

var (
	ErrorUserExist       = errors.New("用户已存在(dao层user.go)")
	ErrorUserNotExist    = errors.New("用户不存在(dao层user.go)")
	ErrorInvalidPassword = errors.New("密码错误(dao层user.go)")
	ErrorInvalidID       = errors.New("无效的ID")
)
