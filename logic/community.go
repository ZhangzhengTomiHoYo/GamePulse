package logic

import (
	"gamepulse/dao/pgsql"
	"gamepulse/models"
)

func GetCommunityList() ([]*models.Community, error) {
	// 查数据 查找到所有的community 并返回
	return pgsql.GetCommunityList()
}

func GetCommunityDetail(id int64) (*models.CommunityDetail, error) {
	return pgsql.GetCommunityDetailByID(id)
}
