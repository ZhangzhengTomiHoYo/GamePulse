package logic

import (
	"errors"
	"gamepulse/dao/pgsql"
	"gamepulse/models"
	"gamepulse/pkg/jwt"
	"gamepulse/pkg/snowflake"
	"strings"

	"go.uber.org/zap"
)

// 存放业务逻辑的代码

var invitationCodes = map[string]struct{}{
	"20260426": {},
	"HR2026":   {},
}

var ErrorInvalidInvitationCode = errors.New("invalid invitation code")

func SignUp(p *models.ParaSignUp) (err error) {
	if _, ok := invitationCodes[strings.TrimSpace(p.InvitationCode)]; !ok {
		return ErrorInvalidInvitationCode
	}

	// 1.判断用户存不存在

	if err = pgsql.CheckUserExist(p.Username); err != nil {
		return err
	}

	// 2.生成UID
	userID := snowflake.GenID()
	// 构造一个User实例
	user := &models.User{
		UserID:   userID,
		Username: p.Username,
		Password: p.Password,
	}
	// 3.保存进数据库
	return pgsql.InsertUser(user)
}

func Login(p *models.ParaLogin) (user *models.User, err error) {
	user = &models.User{
		Username: p.Username,
		Password: p.Password,
	}
	// 传递的是指针，就能拿到user.UserID
	if err := pgsql.Login(user); err != nil {
		return nil, err
	}

	token, err := jwt.GenToken(user.UserID, user.Username)
	if err != nil {
		zap.L().Error("jwt.GenToken err", zap.Error(err))
		return nil, err
	}
	user.Token = token

	return user, err
}
