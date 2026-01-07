package logic

import (
	"bluebell/dao/mysql"
	"bluebell/models"
)

func GetCommunityList() ([]*models.Community, error) {
	// 查数据 查找到所有的community 并返回
	return mysql.GetCommunityList()
}
