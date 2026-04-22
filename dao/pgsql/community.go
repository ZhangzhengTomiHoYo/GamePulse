package pgsql

import (
	"database/sql"
	"gamepulse/models"

	"go.uber.org/zap"
)

// GetCommunityList 社区分类详情
func GetCommunityList() (communityList []*models.Community, err error) {
	// 查数据库 查找到所有的community 并返回
	sqlStr := `select community_id, community_name from community`
	if err = db.Select(&communityList, sqlStr); err != nil {
		if err == sql.ErrNoRows {
			zap.L().Warn("there is no community in db")
			return nil, nil
		}
		return nil, err
	}
	return communityList, nil // 之前是隐式return communityList, err
}

// GetCommunityDetailByID 根据ID查询社区详情
func GetCommunityDetailByID(id int64) (community *models.CommunityDetail, err error) {
	community = new(models.CommunityDetail)
	sqlStr := `select community_id, community_name, introduction, create_time from community where community_id = $1`
	if err = db.Get(community, sqlStr, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrorInvalidID
		}
		return nil, err
	}
	return community, nil
}
