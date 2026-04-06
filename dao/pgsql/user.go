package pgsql

import (
	"bluebell/models"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
)

// 把每一步数据库操作封装成函数
// 等待logic层根据业务逻辑调用

const secret = "zhangzheng"

// CheckUserExist 检查用户名的用户是否存在
func CheckUserExist(username string) (err error) {
	sqlStr := `select count(user_id) from "user" where username = $1`
	var count int
	if err = db.Get(&count, sqlStr, username); err != nil {
		return err
	}
	if count > 0 {
		return ErrorUserExist
	}
	return nil
}

// InsertUser 向数据库中插入一条新的用户记录
func InsertUser(user *models.User) (err error) {
	user.Password = encryptPassword(user.Password)
	sqlStr := `insert into "user"(user_id, username, password) values($1, $2, $3) returning user_id`
	var insertedUserID int64
	if err = db.QueryRowx(sqlStr, user.UserID, user.Username, user.Password).Scan(&insertedUserID); err != nil {
		return err
	}
	user.UserID = insertedUserID
	return nil
}

func encryptPassword(oPassword string) string {
	h := md5.New()
	h.Write([]byte(secret))
	return hex.EncodeToString(h.Sum([]byte(oPassword)))
}

// Login loads user by username and validates password.
func Login(user *models.User) (err error) {
	oPassword := user.Password
	sqlStr := `select user_id, username, password from "user" where username = $1`
	err = db.Get(user, sqlStr, user.Username)
	// 一般是用户名或密码错误 如果直接告诉用户不存在 就会疯狂的尝试登录网站
	if err == sql.ErrNoRows {
		return ErrorUserNotExist
	}
	if err != nil {
		// 查询数据库失败
		return err
	}
	// 判断密码是否正确
	if encryptPassword(oPassword) != user.Password { // 加密的密码
		return ErrorInvalidPassword
	}
	return nil
}

// GetUserByID 根据id获取用户信息
func GetUserByID(uid int64) (user *models.User, err error) {
	user = new(models.User)
	sqlStr := `select user_id, username from "user" where user_id = $1`
	err = db.Get(user, sqlStr, uid)
	return user, err
}
