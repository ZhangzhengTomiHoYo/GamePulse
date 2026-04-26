package pgsql

import (
	"gamepulse/models"
	"strings"
)

var communityIDs = map[string]int64{
	"原神":     1,
	"崩坏星穹铁道": 2,
	"绝区零":    3,
	"鸣潮":     4,
	"明日方舟":   5,
	"碧蓝航线":   6,
	"崩坏3":    7,
}

// ChatGetPostListByCommunity 智能体获取对应帖子的信息
func ChatGetPostListByCommunity(backend models.ChatPostBackend) (resultList []*models.ChatPostBackendResult, err error) {
	communityName := strings.TrimSpace(backend.CommunityName)
	community := communityIDs[communityName]
	sentiment := strings.ToLower(strings.TrimSpace(backend.SentimentLabel))
	switch sentiment {
	case "positive", "negative", "neutral", "all":
	default:
		sentiment = "all"
	}

	lastTime := backend.LastTime
	if lastTime <= 0 {
		lastTime = 7
	}
	if lastTime > 30 {
		lastTime = 30
	}

	sqlStr := `
		select
			p.title,
			p.content,
			pa.sentiment_label,
			pa.sentiment_score,
			pa.risk_level,
			pa.topics,
			pa.keywords,
			pa.summary
		from post p
		join post_analysis pa on pa.post_id = p.post_id
		where p.status = 1
		  and ($1 = 0 or p.community_id = $1)
		  and ($2 = 'all' or pa.sentiment_label = $2)
		  and p.create_time >= now() - ($3::int * interval '1 day')
		order by pa.risk_level desc, p.create_time desc
		limit 20`

	resultList = make([]*models.ChatPostBackendResult, 0, 20)
	err = db.Select(&resultList, sqlStr, community, sentiment, lastTime)
	return resultList, err
}
