package controllers

import (
	"bluebell/logic"
	"bluebell/models"
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func UploadImageHandler(c *gin.Context) {

}

// DeletePostHandler 删除帖子接口
// @Summary 删除帖子接口
// @Description 作者删除自己的帖子（软删除并清理Redis）
// @Tags 帖子相关接口
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path int64 true "帖子id"
// @Security ApiKeyAuth
// @Success 200 {object} _ResponseData
// @Router /post/{id} [delete]
func DeletePostHandler(c *gin.Context) {
	// 1. 获取路径参数
	pidStr := c.Param("id")
	pid, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil {
		zap.L().Error("delete post with invalid param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}

	// 2. 从上下文(Context)获取当前发请求的用户ID (鉴权中间件写入的)
	userID, err := getCurrentUserID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}

	// 3. 调用逻辑层执行删帖
	if err := logic.DeletePost(pid, userID); err != nil {
		zap.L().Error("logic.DeletePost failed", zap.Int64("postID", pid), zap.Int64("userID", userID), zap.Error(err))
		// 注：如果你在 code.go 里加了 CodeNoPermission，这里可以换成对应的错误码
		ResponseError(c, CodeNoPermission)
		return
	}

	// 4. 返回响应
	ResponseSuccess(c, nil)
}

// CreatePostHandler 创建帖子接口
// @Summary 创建帖子接口
// @Description 创建带图片的帖子接口
// @Tags 帖子相关接口
// @Accept application/json
// @Produce application/json
// @Param object body models.ParamCreatePost true "帖子参数"
// @Security ApiKeyAuth
// @Success 200 {object} _ResponseData
// @Router /post [post]
func CreatePostHandler(c *gin.Context) {
	// 1. 获取参数及参数的校验 (注意：这里改成了绑定 ParamCreatePost)
	p := new(models.ParamCreatePost)
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Debug("c.ShouldBindJSON(p) error", zap.Any("err", err))
		zap.L().Error("create post with invalid param")
		ResponseError(c, CodeInvalidParam)
		return
	}

	// 从c取到当前发请求的用户id
	userID, err := getCurrentUserID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}

	// 2. 数据组装：将接收到的请求参数拼装成数据库模型
	post := &models.Post{
		AuthorID:    userID,
		CommunityID: p.CommunityID,
		Title:       p.Title,
		Content:     p.Content,
	}

	// 极其关键的一步：将前端传来的图片数组序列化为 JSON 字符串
	if len(p.ImageURLs) > 0 {
		b, err := json.Marshal(p.ImageURLs)
		if err != nil {
			zap.L().Error("json.Marshal(p.ImageURLs) failed", zap.Error(err))
			ResponseError(c, CodeServeBusy)
			return
		}
		post.ImageURL = string(b) // 变成了类似 '["http://...1.jpg", "http://...2.jpg"]'
	} else {
		post.ImageURL = "[]" // 如果没发图片，存个空数组的 JSON 字符串
	}

	// 3. 创建帖子 (直接把你拼装好的 post 扔给底层)
	if err := logic.CreatePost(post); err != nil {
		zap.L().Error("logic.CreatePost(post) failed", zap.Error(err))
		ResponseError(c, CodeServeBusy)
		return
	}

	// 4. 返回响应
	ResponseSuccess(c, nil)
}

// GetPostDetailHandler 获取单个帖子详情接口
// @Summary 获取单个帖子详情接口
// @Description 获取单个帖子详情接口
// @Tags 帖子相关接口
// @Accept application/json
// @Produce application/json
// @Param id path int64 true "帖子id"
// @Security ApiKeyAuth
// @Success 200 {object} _ResponsePostData
// @Router /post/{id} [get]
func GetPostDetailHandler(c *gin.Context) {
	// 1. 获取参数（从URL中获取帖子的id）
	pidStr := c.Param("id")
	pid, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil {
		zap.L().Error("get pos detail with invalid param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 2. 根据id取出帖子数据（查数据库）
	data, err := logic.GetPostByID(pid)
	if err != nil {
		zap.L().Error("logic.GetPostDetail(pid) failed", zap.Error(err))
		ResponseError(c, CodeServeBusy)
		return
	}
	// 3. 返回响应
	ResponseSuccess(c, data)
}

// GetPageInfo 获取分页参数
func GetPageInfo(c *gin.Context) (int64, int64) {
	// 获取分页参数
	pageStr := c.Query("page")
	sizeStr := c.Query("size")
	var (
		page int64
		size int64
		err  error
	)
	page, err = strconv.ParseInt(pageStr, 10, 64)
	if err != nil {
		page = 1
	}
	size, err = strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		size = 10
	}
	return page, size
}

// GetListDetailHandler 获取一页帖子详情的接口(已弃用)
// @Summary 获取一页帖子详情的接口(已弃用)
// @Description 获取一页帖子详情的接口(已弃用)
// @Tags 帖子相关接口
// @Accept application/json
// @Produce application/json
// @Param page query int true "页码"
// @Param size query int true "每页大小"
// @Param order query string true "排序方式"
// @Security ApiKeyAuth
// @Success 200 {object} _ResponsePostList
// @Router /posts [get]
func GetListDetailHandler(c *gin.Context) {
	// 获取分页参数
	page, size := GetPageInfo(c)
	// 获取数据
	data, err := logic.GetPostList(page, size)
	if err != nil {
		zap.L().Error("logic.GetPostList() failed", zap.Error(err))
		ResponseError(c, CodeServeBusy)
		return
	}
	ResponseSuccess(c, data)
	// 返回响应
}

// 老的注释，供参考
// GetListDetailHandler2 升级版获取帖子列表的接口
// 根据前端传来的参数动态的获取帖子列表
// 按创建时间排序 或者 按照分数排序
// 1. 获取参数
// 2. 去redis查询id列表
// 3. 根据id去数据库查询帖子详细信息
// 4.

// GetListDetailHandler2 升级版帖子列表接口
// @Summary 升级版帖子列表接口
// @Description 可按社区按时间或分数排序查询帖子列表接口
// @Tags 帖子相关接口
// @Accept application/json
// @Produce application/json
// @Param Authorization header string false "Bearer 用户令牌"
// @Param object query models.ParamPostList false "查询参数"
// @Security ApiKeyAuth
// @Success 200 {object} _ResponsePostList
// @Router /posts2 [get]
func GetListDetailHandler2(c *gin.Context) {
	// GET请求参数: /api/v1/post2?page=1&size=10&order=time  (query string 参数)
	// 获取分页参数 gin框架运用反射取出来
	// 注意！请求中有json的是shouldbindjson 此处用的是shouldbindquery
	p := &models.ParamPostList{
		Page:  1, // 应该写到配置文件中，此处为方便才
		Size:  10,
		Order: models.OrderTime,
	}
	// 注意！！！ 此处绑定query
	if err := c.ShouldBindQuery(p); err != nil {
		zap.L().Error("GetListDetailHandler(p) with invalid param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}

	// 获取数据
	data, err := logic.GetPostListNew(p) // 更新：合二为一
	if err != nil {
		zap.L().Error("logic.GetPostListNew() failed", zap.Error(err))
		ResponseError(c, CodeServeBusy)
		return
	}
	ResponseSuccess(c, data)
	// 返回响应
}

// 老的函数，已弃用，供参考
//// 根据社区去查询帖子列表
//func GetCommunityPostListHandler(c *gin.Context) {
//	// GET请求参数: /api/v1/post2?page=1&size=10&order=time  (query string 参数)
//	// 获取分页参数 gin框架运用反射取出来
//	// 注意！请求中有json的是shouldbindjson 此处用的是shouldbindquery
//	p := &models.ParamCommunityPostList{
//		ParamPostList: &models.ParamPostList{
//			Page:  1,
//			Size:  10,
//			Order: models.OrderTime,
//		},
//	}
//	if err := c.ShouldBindQuery(p); err != nil {
//		zap.L().Error("GetListDetailHandler(p) with invalid param", zap.Error(err))
//		ResponseError(c, CodeInvalidParam)
//		return
//	}
//
//	// 获取数据
//	data, err := logic.GetCommunityPostList(p)
//	if err != nil {
//		zap.L().Error("logic.GetPostList() failed", zap.Error(err))
//		ResponseError(c, CodeServeBusy)
//		return
//	}
//	ResponseSuccess(c, data)
//	// 返回响应
//}
