package models

const (
	OrderTime  = "time"
	OrderScore = "score"
)

// ParaSignUp 注册参数
type ParaSignUp struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	RePassword string `json:"re_password" binding:"required,eqfield=Password"`
}

// ParamLogin 登录请求参数
type ParaLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 投票数据
type ParaVoteData struct {
	// UserID 从请求中获取当前的用户  不用从结构体
	PostID    string `json:"post_id" binding:"required"`              // 帖子id
	Direction int8   `json:"direction,string" binding:"oneof=1 0 -1"` // 赞成票 1 反对票-1 取消投票0 required会把0值过滤 所以删掉required
}

// ParamPostList 获取帖子列表query string参数
type ParamPostList struct {
	Page  int64  `form:"page"`
	Size  int64  `form:"size"`
	Order string `form:"order"`
}
