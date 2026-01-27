package controllers

import (
	"bluebell/logic"
	"bluebell/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func CreatePostHandler(c *gin.Context) {
	// 1. 获取参数及参数的校验
	p := new(models.Post)
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
	p.AuthorID = userID
	// 2. 创建帖子
	if err := logic.CreatePost(p); err != nil {
		zap.L().Error("logic.CreatePost(p) failed", zap.Error(err))
		ResponseError(c, CodeServeBusy)
		return
	}

	// 3. 返回响应
	ResponseSuccess(c, nil)
}

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

// GetListDetailHandler 获取帖子列表的接口
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
	// GET请求参数: /api/v1/post2?page=1&size-10&order=time  (query string 参数)
	// 获取分页参数 gin框架运用反射取出来
	// 注意！请求中有json的是shouldbindjson 此处用的是shouldbindquery
	p := &models.ParamPostList{
		Page:  1, // 应该写到配置文件中，此处为方便才
		Size:  10,
		Order: models.OrderTime,
	}
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

//
//// 根据社区去查询帖子列表
//func GetCommunityPostListHandler(c *gin.Context) {
//	// GET请求参数: /api/v1/post2?page=1&size-10&order=time  (query string 参数)
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
