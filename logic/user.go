package logic

import (
	"bluebell/dao/pgsql"
	"bluebell/models"
	"bluebell/pkg/jwt"
	"bluebell/pkg/snowflake"

	"go.uber.org/zap"
)

// 存放业务逻辑的代码

func SignUp(p *models.ParaSignUp) (err error) {
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
