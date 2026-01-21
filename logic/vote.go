package logic

import (
	"bluebell/dao/redis"
	"bluebell/models"
	"strconv"

	"go.uber.org/zap"
)

// 基于用户投票的相关算法 http://www.ruanyifeng.com/blog/algorithm/

// 本项目使用简化版的投票分数
// 投一票就加432分 一天是86400秒/200 -> 需要200张赞成票可以给你的帖子续一天
// Reference：《Redis实战》

/* 投票的几种情况
direction=1时，有两种情况：
	1. 之前没有投过票，现在投赞成票 --> 更新分数和投票记录
	2. 之前投反对票，现在改投赞成票 --> 更新分数和投票记录
direction=0时，有两种情况：
	1. 之前投过赞成票，现在要取消投票 --> 更新分数和投票记录
	2. 之前投过反对票，现在要取消投票 --> 更新分数和投票记录
direction=-1时，有两种情况：
	1. 之前没有投过票，现在投反对票 --> 更新分数和投票记录
	2. 之前投过赞成票，现在投反对票 --> 更新分数和投票记录

投票记录直接修改变量就行，但是分数的更改稍微麻烦，因为涉及到计算

投票的限制：
	（好几年前的微博，不会有人看，是冷数据，有人故意翻出来，对后端要求很高）
	每个帖子子发表之日起一个星期内允许用户投票，超过一个星期就不允许再投票了
		1. 到期之后将redis中保存的票数存到mysql中
		2. 到期之后删除那个KeyPostVotedZSetPF
*/

// VoteForPost 为帖子投票的函数
func VoteForPost(userID int64, p *models.ParaVoteData) error {
	zap.L().Debug("VoteForPost",
		zap.Int64("userID", userID),
		zap.String("postID", p.PostID),
		zap.Int8("direction", p.Direction))
	return redis.VoteForPost(strconv.Itoa(int(userID)), p.PostID, float64(p.Direction))
}
