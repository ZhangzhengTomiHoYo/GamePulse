package pgsql

import (
	"bluebell/models"
	"errors"
	"strconv"

	"github.com/lib/pq"
)

// SoftDeletePost 逻辑删除帖子（自带防越权校验）
func SoftDeletePost(postID, authorID int64) error {
	// 只有当 post_id 和 author_id 都匹配，且 status 原本为 1 时，才允许删除
	sqlStr := `UPDATE post SET status = 0, update_time = CURRENT_TIMESTAMP WHERE post_id = $1 AND author_id = $2 AND status = 1`

	res, err := db.Exec(sqlStr, postID, authorID)
	if err != nil {
		return err
	}

	// 校验是否真的修改了数据，如果为0说明不是自己的帖子或帖子已删
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("post not exist or no permission")
	}
	return nil
}

func CreatePost(p *models.Post) error {
	// 增加 image_url 的插入
	sqlStr := `insert into post(post_id, title, content, author_id, community_id, image_url) values($1, $2, $3, $4, $5, $6) returning post_id`
	var insertedPostID int64
	// 记得把 p.ImageURL 传进去
	if err := db.QueryRowx(sqlStr, p.ID, p.Title, p.Content, p.AuthorID, p.CommunityID, p.ImageURL).Scan(&insertedPostID); err != nil {
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
