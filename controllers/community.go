package controllers

import (
	"bluebell/logic"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func CommunityHandler(c *gin.Context) {
	// 查询到所有的社区 (community_id, community_name) 以列表的形式返回
	data, err := logic.GetCommunityList()
	if err != nil {
		// 服务端 的 错误日志，这些信息是不给前端看到的
		zap.L().Error("logic.GetCommunityList() failed", zap.Error(err))
		// 只给 客户端/前端 返回状态码
		// 不轻易把服务端错误暴露给外面
		ResponseError(c, CodeServeBusy)
		return
	}
	ResponseSuccess(c, data)
}
