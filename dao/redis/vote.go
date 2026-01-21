package redis

import (
	"errors"
	"math"
	"time"

	"github.com/go-redis/redis"
	"go.uber.org/zap"
)

// 本项目使用简化版的投票分数
// 投一票就加432分 一天是86400秒/200 -> 需要200张赞成票可以给你的帖子续一天
// Reference：《Redis实战》

/*
	投票的几种情况

direction=1时，有两种情况：
 1. 之前没有投过票，现在投赞成票 --> 更新分数和投票记录 0-1=-1 abs 1 +432
 2. 之前投反对票，现在改投赞成票 --> 更新分数和投票记录 -1-1=-2 abs 2 +432*2

direction=0时，有两种情况：
 2. 之前投过反对票，现在要取消投票 --> 更新分数和投票记录 -1-0 = -1 abs 1 +432
 1. 之前投过赞成票，现在要取消投票 --> 更新分数和投票记录 1-0 = 1 abs 1 -432

direction=-1时，有两种情况：
 1. 之前没有投过票，现在投反对票 --> 更新分数和投票记录 0-(-1) = 1 abs 1 -432
 2. 之前投过赞成票，现在投反对票 --> 更新分数和投票记录 1-(-1)=2 abd 2 -432*2

投票记录直接修改变量就行，但是分数的更改稍微麻烦，因为涉及到计算

可以观察到，
上三行是加，是，现在的值大于原来的值
下三行是减，是，现在的值小于原来的值

投票的限制：

	（好几年前的微博，不会有人看，是冷数据，有人故意翻出来，对后端要求很高）
	每个帖子子发表之日起一个星期内允许用户投票，超过一个星期就不允许再投票了
		1. 到期之后将redis中保存的票数存到mysql中
		2. 到期之后删除那个KeyPostVotedZSetPF
*/
const (
	oneWeekInSeconds = 7 * 24 * 3600
	scorePerVote     = 432 // 每一篇占的分数值
)

var (
	ErrVoteTimeExpire = errors.New("投票时间已过")
	ErrVoteRepeat     = errors.New("禁止重复投票")
)

func CreatePost(postID int64) error {
	// 时间和分数，要同时成功
	pipeline := client.TxPipeline() // 放到一个事务里面去
	// 帖子时间
	pipeline.ZAdd(getRedisKey(KeyPostTimeZSet), redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: postID,
	})
	// 帖子分数
	pipeline.ZAdd(getRedisKey(KeyPostTimeZSet), redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: postID,
	})
	_, err := pipeline.Exec() // 事务 要么全执行 要么全不执行

	// 关键3：添加日志，排查写入失败原因
	if err != nil {
		zap.L().Error("Redis ZAdd写入帖子失败", zap.Int64("postID", postID), zap.Error(err))
		return err
	}
	zap.L().Info("Redis ZAdd写入帖子成功", zap.Int64("postID", postID))
	return nil
}

func VoteForPost(userID, postID string, value float64) (err error) {
	// 1. 判断投票限制
	// 取帖子的发布时间
	postTime := client.ZScore(getRedisKey(KeyPostTimeZSet), postID).Val()
	if float64(time.Now().Unix())-postTime > oneWeekInSeconds {
		return ErrVoteTimeExpire
	}

	// 2和3需要放到一个pipeline事务中操作

	// 2. 更新帖子的分数
	// 先查之前的投票记录
	oValue := client.ZScore(getRedisKey(KeyPostVotedZSetPF+postID), userID).Val()
	var op float64

	// 更新：如果这一次投票的值和之前保存的值一致，就提示不允许重复投票
	if value == oValue {
		return ErrVoteRepeat
	}

	if value > oValue {
		op = 1
	} else {
		op = -1
	}
	diff := math.Abs(oValue - value)

	pipeline := client.TxPipeline()
	pipeline.ZIncrBy(getRedisKey(KeyPostScoreZSet), op*diff*scorePerVote, postID)

	// 3. 记录用户为改帖子投票的数据
	if value == 0 {
		// 投票数为0 就是把投票删掉了 所以要用redis的ZREM命令要删掉
		pipeline.ZRem(getRedisKey(KeyPostVotedZSetPF+postID), userID).Result()
	} else {
		pipeline.ZAdd(getRedisKey(KeyPostVotedZSetPF+postID), redis.Z{
			Score:  value,  // 赞成还是反对票
			Member: userID, /// 哪个用户
		})
	}
	_, err = pipeline.Exec()

	return err
}
