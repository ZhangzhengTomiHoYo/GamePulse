package redis

import (
	"bluebell/models"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

func getIDsFromKey(key string, page, size int64) ([]string, error) {
	// 2. 确定查询的索引起始点
	start := (page - 1) * size
	// Redis命令对于[0,2]是左闭右也闭的查询 zrange bluebell:post:score 0 2 withscores
	// 所以查3个是 索引为 2  要减一
	end := start + size - 1
	// 3. ZREVRANGE 查询 按分数从大到校的顺序查询指定数量的元素
	return client.ZRevRange(key, start, end).Result()
}

func GetPostIdsInOrder(p *models.ParamPostList) ([]string, error) {
	// 从redis获取id
	// 1. 根据用户请求中携带的order参数确定要查询的redis key
	key := getRedisKey(KeyPostTimeZSet)
	if p.Order == models.OrderScore {
		key = getRedisKey(KeyPostScoreZSet)
	}
	return getIDsFromKey(key, p.Page, p.Size)
}

// GetPostVoteData 根据ids查询每篇帖子的投赞成票的数据
func GetPostVoteData(ids []string) (data []int64, err error) {
	//data = make([]int64, 0, len(ids))
	//for _, id := range ids {
	//	key := getRedisKey(KeyPostVotedZSetPF + id)
	//	// 查找key中分数是1的元素的数量 -> 统计每篇帖子的赞成票的数量
	//	v := client.ZCount(key, "1", "1").Val()
	//}

	// 使用事务 一次发送多条命令 减少RTT
	pipeline := client.Pipeline()
	for _, id := range ids {
		key := getRedisKey(KeyPostVotedZSetPF + id)
		pipeline.ZCount(key, "1", "1")
	}
	cmders, err := pipeline.Exec()
	if err != nil {
		return nil, err
	}
	data = make([]int64, 0, len(cmders))
	for _, cmder := range cmders {
		v := cmder.(*redis.IntCmd).Val()
		data = append(data, v)
	}
	return
}

// GetCommuntiyPostIdsInOrder 按社区查询ids
func GetCommunityPostIdsInOrder(p *models.ParamPostList) ([]string, error) {
	orderKey := getRedisKey(KeyPostTimeZSet)
	if p.Order == models.OrderScore {
		orderKey = getRedisKey(KeyPostScoreZSet)
	}

	// 使用zinterstore把分区的帖子与帖子分数的 zset 生成一个新的zset
	// 针对新的zset按之前的逻辑取数据

	// 社区的key
	cKey := getRedisKey(KeyCommunitySetPF + fmt.Sprintf("%d", p.CommunityID))
	// 利用缓存key减少zinterstore执行的次数
	// 60秒一过 就又要进if
	key := orderKey + fmt.Sprintf("%d", p.CommunityID)
	if client.Exists(key).Val() == 0 {
		// 不存在，需要计算
		pipeline := client.Pipeline()
		pipeline.ZInterStore(key, redis.ZStore{
			Aggregate: "MAX",
		}, cKey, orderKey) // zinterscore 计算
		pipeline.Expire(key, 60*time.Second) // 设置超时时间
		_, err := pipeline.Exec()
		if err != nil {
			return nil, err
		}
	}
	// 存在的话就直接根据key查询
	return getIDsFromKey(key, p.Page, p.Size)
}
