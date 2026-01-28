package controllers

import "bluebell/models"

// 专门用来放接口文档用到的models
// 因为我们的接口文档返回的数据格式是一致的，但是具体的data类型不一致
// 加_ 是用来区分的
//

type _ResponsePostList struct {
	Code    ResCode                 `json:"code"`    //业务响应状态码
	Message string                  `json:"message"` //提示信息
	Data    []*models.ApiPostDetail `json:"data"`    //帖子列表数据
}

type _ResponseData struct {
	Code    ResCode `json:"code"`    //业务响应状态码
	Message string  `json:"message"` //提示信息
	Data    any     `json:"data"`    //数据
}

type _ResponsePostData struct {
	Code    ResCode               `json:"code"`    //业务响应状态码
	Message string                `json:"message"` //提示信息
	Data    *models.ApiPostDetail `json:"data"`    //单个帖子数据
}
