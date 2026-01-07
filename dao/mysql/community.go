package mysql

import (
	"bluebell/models"
	"database/sql"

	"go.uber.org/zap"
)

func GetCommunityList() (communityList []*models.Community, err error) {
	// 查数据库 查找到所有的community 并返回
	sqlStr := "select community_id, community_name from community"
	if err := db.Select(&communityList, sqlStr); err != nil {
		if err == sql.ErrNoRows {
			zap.L().Warn("there is no community in db")
			err = nil
		}
	}
	return // 隐式return communityList, err
}
