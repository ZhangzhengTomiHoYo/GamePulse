package controllers

import (
	"errors"
	"fmt"
	"gamepulse/dao/pgsql"
	"gamepulse/logic"
	"gamepulse/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// SignUpHandler 用户注册接口
// @Summary 用户注册接口
// @Description 用户注册接口
// @Tags 用户相关接口
// @Accept application/json
// @Produce application/json
// @Param request body models.ParaSignUp true "注册参数"
// @Security ApiKeyAuth
// @Success 200 {object} _ResponseData "{"code":1000, "msg":"success", "data":null}"
// @Router /signup [post]
func SignUpHandler(c *gin.Context) {
	// 1. 获取参数和参数校验
	// 使用new 而不是 var 返回指针，方便后续函数编写
	//var p models.ParaSignUp
	p := new(models.ParaSignUp)
	// 注意！ 只能判断一些简单的错误，比如字段类型对不对int传了string，是否为json
	// 和业务相关的，要手动判断，比如是否为空、password和re_password是否一致
	// add: 实际上gin内置了validator库，用于判断，而且用法十分简单，打tag就行
	//
	// 有人说，前端用js已经判断过了，但是如果禁用了js，或者用脚本攻击了js，所以不要相信前端，后端是一定要做的
	if err := c.ShouldBindJSON(p); err != nil {
		// 1.1 请求参数有误 直接返回响应
		// func Error(msg string, fields ...zap.Field)
		// 第一个参数是msg,第二个及后续参数都是「结构化日志字段（Field）」，作用是给日志附加「可解析的关键信息」（而非单纯拼接字符串）
		// 结构体用zap.Any()
		zap.L().Error("SignUp with invalid param", zap.Error(err))
		// 判断err是不是validator.ValidationErrors类型
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			// 现在就很清爽
			ResponseError(c, CodeInvalidParam)
			return
		}
		//
		ResponseErrorWithMsg(c, CodeInvalidParam, removeTopStruct(errs.Translate(trans)))
		return
	}

	// 手动对请求参数进行详细的业务规则校验
	//if len(p.Username) == 0 || len(p.Password) == 0 || len(p.RePassword) == 0 || p.RePassword != p.Password {
	//	zap.L().Error("SignUp with invalid param")
	//	c.JSON(http.StatusOK, gin.H{
	//		"msg": "请求参数有误",
	//	})
	//}

	// 2. 业务处理
	if err := logic.SignUp(p); err != nil {
		zap.L().Error("logic.SignUp failed", zap.Error(err))
		if errors.Is(err, pgsql.ErrorUserExist) {
			ResponseError(c, CodeUserExist)
			return
		}
		ResponseError(c, CodeServeBusy)
		return
	}
	// 3.返回响应
	ResponseSuccess(c, nil)
}

// LoginHandler 用户登录接口
// @Summary 用户登录接口
// @Description 用户登录接口
// @Tags 用户相关接口
// @Accept application/json
// @Produce application/json
// @Param object body models.ParaLogin true "登录参数"
// @Security ApiKeyAuth
// @Success 200 {object} _ResponseData
// @Router /login [post]
func LoginHandler(c *gin.Context) {
	// 1.获取请求参数及参数校验
	p := new(models.ParaLogin)
	if err := c.BindJSON(p); err != nil {
		zap.L().Error("Login with invalid param", zap.Error(err))
		// 判断err是不是validator.ValidationErrors类型
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}
		ResponseErrorWithMsg(c, CodeInvalidParam, removeTopStruct(errs.Translate(trans)))
		return
	}
	//2.业务逻辑处理
	user, err := logic.Login(p)
	if err != nil {
		zap.L().Error("logic.Login failed", zap.String("username", p.Username), zap.Error(err))
		if errors.Is(err, pgsql.ErrorUserNotExist) {
			ResponseError(c, CodeUserNoteExist)
			return
		}
		ResponseError(c, CodeInvalidPassword)
		return
	}
	//3.返回响应
	ResponseSuccess(c, gin.H{
		"user_id":   fmt.Sprintf("%d", user.UserID), // 前端最大为1<<53-1 后端int64最大为1<<63-1 此处转换为字符串
		"user_name": user.Username,
		"token":     user.Token,
	})
}
