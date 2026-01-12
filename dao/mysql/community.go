package mysql

import (
	"bluebell/models"
	"database/sql"

	"go.uber.org/zap"
)

// GetCommunityList 社区分类详情
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

// GetCommunityDetailByID 根据ID查询社区详情
func GetCommunityDetailByID(id int64) (community *models.CommunityDetail, err error) {
	community = new(models.CommunityDetail)
	sqlstr := `select 
		community_id, community_name, introduction, create_time
		from community
		where community_id = ?
	`
	if err = db.Get(community, sqlstr, id); err != nil {
		if err == sql.ErrNoRows {
			err = ErrorInvalidID
		}
	}
	return community, err
}
