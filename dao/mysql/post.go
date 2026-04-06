package mysql

import (
	"bluebell/models"
	"strconv"

	"github.com/lib/pq"
)

func CreatePost(p *models.Post) error {
	sqlStr := `insert into post(post_id, title, content, author_id, community_id) values($1, $2, $3, $4, $5) returning post_id`
	var insertedPostID int64
	if err := db.QueryRowx(sqlStr, p.ID, p.Title, p.Content, p.AuthorID, p.CommunityID).Scan(&insertedPostID); err != nil {
		return err
	}
	p.ID = insertedPostID
	return nil
}

// GetPostByID 根据id查询单个帖子数据
func GetPostByID(pid int64) (post *models.Post, err error) {
	post = new(models.Post)
	sqlStr := `select post_id, title, content, author_id, community_id, create_time from post where post_id = $1`
	err = db.Get(post, sqlStr, pid)
	return post, err
}

// GetPostList 查询帖子列表函数
func GetPostList(page, size int64) (posts []*models.Post, err error) {
	sqlStr := `select post_id, title, content, author_id, community_id, create_time from post order by create_time desc limit $1 offset $2`
	posts = make([]*models.Post, 0, 2)
	err = db.Select(&posts, sqlStr, size, (page-1)*size)
	return posts, err
}

// 根据给定的id列表查询帖子数据
func GetPostListByIDs(ids []string) (postList []*models.Post, err error) {
	idValues := make([]int64, 0, len(ids))
	for _, id := range ids {
		v, convErr := strconv.ParseInt(id, 10, 64)
		if convErr != nil {
			return nil, convErr
		}
		idValues = append(idValues, v)
	}

	sqlStr := `select post_id, title, content, author_id, community_id, create_time
		from post
		where post_id = any($1::bigint[])
		order by array_position($1::bigint[], post_id)`
	err = db.Select(&postList, sqlStr, pq.Array(idValues))
	return postList, err
}
