package logic

import (
	"bluebell/dao/minio"
	"bluebell/dao/pgsql"
	"bluebell/dao/redis"
	"bluebell/models"
	"bluebell/pkg/snowflake"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// UploadImage 处理图片上传业务逻辑
func UploadImage(fileHeader *multipart.FileHeader) (string, error) {
	// 1. 提取后缀并统一转小写，防止绕过 (比如 .JPG)
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))

	// 2. 严格的业务白名单校验 (已加入 .avif)
	if ext != ".jpg" && ext != ".png" && ext != ".jpeg" && ext != ".gif" && ext != ".webp" && ext != ".avif" {
		return "", errors.New("unsupported file extension: " + ext)
	}

	// 3. 业务命名规则：雪花算法重命名
	imageID := snowflake.GenID()
	objectName := fmt.Sprintf("%d%s", imageID, ext)

	// 4. 打开文件准备底层读取
	fileObj, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("file open failed: %w", err)
	}
	defer fileObj.Close()

	// 5. 移交给真正的底层 DAO 层（MinIO）
	imageURL, err := minio.UploadFile(objectName, fileObj, fileHeader.Size, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		return "", fmt.Errorf("dao minio upload failed: %w", err)
	}

	return imageURL, nil
}

// DeletePost 删除帖子逻辑
func DeletePost(postID, authorID int64) error {
	// 1. 先查询帖子详情，主要是为了拿到 CommunityID 以便后续清理 Redis 社区集合
	post, err := pgsql.GetPostByID(postID)
	if err != nil {
		zap.L().Error("pgsql.GetPostByID failed in DeletePost",
			zap.Int64("postID", postID),
			zap.Error(err))
		return err // 帖子不存在或查询失败
	}

	// 2. 数据库层面执行软删除（底层的 SQL 已经附带了 author_id = $2 的校验防越权）
	err = pgsql.SoftDeletePost(postID, authorID)
	if err != nil {
		zap.L().Error("pgsql.SoftDeletePost failed",
			zap.Int64("postID", postID),
			zap.Int64("authorID", authorID),
			zap.Error(err))
		return err
	}

	// 3. 执行 Redis 缓存连环清理
	err = redis.RemovePostFromCache(postID, post.CommunityID)
	if err != nil {
		// 缓存清理失败记录严重日志。虽然数据库已经删了，但缓存有残留会导致脏数据
		zap.L().Error("redis.RemovePostFromCache failed",
			zap.Int64("postID", postID),
			zap.Int64("communityID", post.CommunityID),
			zap.Error(err))
		return err
	}

	return nil
}

func CreatePost(p *models.Post) (err error) {
	// 1. 生成post id
	p.ID = snowflake.GenID()
	// 2. 保存到数据库
	// 3. 返回
	err = pgsql.CreatePost(p)
	if err != nil {
		return err
	}
	err = redis.CreatePost(p.ID, p.CommunityID)
	if err != nil {
		return err
	}
	return err
}

// GetPostByID 根据帖子id查询帖子详情数据
func GetPostByID(pid int64) (data *models.ApiPostDetail, err error) {
	// data是一个指针，要初始化
	data = new(models.ApiPostDetail)
	// 查询并组合我们接口想用的数据
	post, err := pgsql.GetPostByID(pid)
	if err != nil {
		zap.L().Error("pgsql.GetPostByID(pid) failed",
			zap.Int64("pid", pid),
			zap.Error(err))
		return
	}
	// 根据作者id查询作者信息
	user, err := pgsql.GetUserByID(post.AuthorID)
	if err != nil {
		zap.L().Error("pgsql.GetUserByID(post.AuthorID) failed",
			zap.Int64("author_id", post.AuthorID),
			zap.Error(err))
		return
	}
	// 根据社区id查询社区详细信息
	community, err := pgsql.GetCommunityDetailByID(post.CommunityID)
	if err != nil {
		zap.L().Error("pgsql.GetCommunityDetailByID(post.CommunityID) failed",
			zap.Int64("author_id", post.AuthorID),
			zap.Error(err))
		return
	}
	// 根据 社区id 在 redis 中查询投票数
	// 第一步：int64 -> string
	pidStr := strconv.FormatInt(pid, 10)

	// 第二步：string -> []string
	pidStrSlice := []string{pidStr}
	nums, err := redis.GetPostVoteData(pidStrSlice)

	// 【核心新增逻辑】：反序列化图片字符串
	var imageURLs []string
	if post.ImageURL != "" && post.ImageURL != "[]" && post.ImageURL != "null" {
		if err := json.Unmarshal([]byte(post.ImageURL), &imageURLs); err != nil {
			zap.L().Warn("json.Unmarshal image_url failed", zap.Error(err))
		}
	}

	data = &models.ApiPostDetail{
		AuthorName:      user.Username,
		VoteNum:         nums[0],
		Post:            post,
		CommunityDetail: community,
	}
	return
}

// GetPostList 获取帖子列表
func GetPostList(page, size int64) (data []*models.ApiPostDetail, err error) {
	posts, err := pgsql.GetPostList(page, size)
	if err != nil {
		return nil, err
	}
	data = make([]*models.ApiPostDetail, 0, len(posts))

	for _, post := range posts {
		// 根据作者id查询作者信息
		user, err := pgsql.GetUserByID(post.AuthorID)
		if err != nil {
			zap.L().Error("pgsql.GetUserByID(post.AuthorID) failed",
				zap.Int64("author_id", post.AuthorID),
				zap.Error(err))
			continue
		}
		// 根据社区id查询社区详细信息
		community, err := pgsql.GetCommunityDetailByID(post.CommunityID)
		if err != nil {
			zap.L().Error("pgsql.GetCommunityDetailByID(post.CommunityID) failed",
				zap.Int64("author_id", post.AuthorID),
				zap.Error(err))
			continue
		}

		// 【核心新增逻辑】：反序列化图片字符串
		var imageURLs []string
		if post.ImageURL != "" && post.ImageURL != "[]" && post.ImageURL != "null" {
			if err := json.Unmarshal([]byte(post.ImageURL), &imageURLs); err != nil {
				zap.L().Warn("json.Unmarshal image_url failed", zap.Error(err))
			}
		}

		postDetail := &models.ApiPostDetail{
			AuthorName:      user.Username,
			Post:            post,
			CommunityDetail: community,
		}
		data = append(data, postDetail)

	}
	return
}

// GetPostList2 获取帖子列表
func GetPostList2(p *models.ParamPostList) (data []*models.ApiPostDetail, err error) {
	// 2. 去redis查询寻id列表
	ids, err := redis.GetPostIdsInOrder(p)
	if err != nil {
		return
	}
	if len(ids) == 0 {
		zap.L().Warn("redis.GetPostIDsInOrder(p) return 0 data")
		return
	}
	// 3. 根据id取MySQL数据库查询帖子详细信息
	// 返回的数据还要按照我给定的id的顺序返回
	posts, err := pgsql.GetPostListByIDs(ids)
	if err != nil {
		return
	}

	// 提前查询好每篇帖子的投票数
	voteData, err := redis.GetPostVoteData(ids)
	if err != nil {
		return
	}

	// 下面的和上面一样
	for idx, post := range posts {
		// 根据作者id查询作者信息
		user, err := pgsql.GetUserByID(post.AuthorID)
		if err != nil {
			zap.L().Error("pgsql.GetUserByID(post.AuthorID) failed",
				zap.Int64("author_id", post.AuthorID),
				zap.Error(err))
			continue
		}
		// 根据社区id查询社区详细信息
		community, err := pgsql.GetCommunityDetailByID(post.CommunityID)
		if err != nil {
			zap.L().Error("pgsql.GetCommunityDetailByID(post.CommunityID) failed",
				zap.Int64("author_id", post.AuthorID),
				zap.Error(err))
			continue
		}

		// 【核心新增逻辑】：反序列化图片字符串
		var imageURLs []string
		if post.ImageURL != "" && post.ImageURL != "[]" && post.ImageURL != "null" {
			if err := json.Unmarshal([]byte(post.ImageURL), &imageURLs); err != nil {
				zap.L().Warn("json.Unmarshal image_url failed", zap.Error(err))
			}
		}

		postDetail := &models.ApiPostDetail{
			AuthorName:      user.Username,
			VoteNum:         voteData[idx],
			Post:            post,
			CommunityDetail: community,
		}
		data = append(data, postDetail)

	}
	return
}

func GetCommunityPostList(p *models.ParamPostList) (data []*models.ApiPostDetail, err error) {
	// 2. 去redis查询寻id列表
	ids, err := redis.GetCommunityPostIdsInOrder(p)
	if err != nil {
		return
	}
	if len(ids) == 0 {
		zap.L().Warn("redis.GetCommunityPostIdsInOrder(p) return 0 data")
		return
	}
	// 3. 根据id取MySQL数据库查询帖子详细信息
	// 返回的数据还要按照我给定的id的顺序返回
	posts, err := pgsql.GetPostListByIDs(ids)
	if err != nil {
		return
	}

	// 提前查询好每篇帖子的投票数
	voteData, err := redis.GetPostVoteData(ids)
	if err != nil {
		return
	}

	// 下面的和上面一样
	for idx, post := range posts {
		// 根据作者id查询作者信息
		user, err := pgsql.GetUserByID(post.AuthorID)
		if err != nil {
			zap.L().Error("pgsql.GetUserByID(post.AuthorID) failed",
				zap.Int64("author_id", post.AuthorID),
				zap.Error(err))
			continue
		}
		// 根据社区id查询社区详细信息
		community, err := pgsql.GetCommunityDetailByID(post.CommunityID)
		if err != nil {
			zap.L().Error("pgsql.GetCommunityDetailByID(post.CommunityID) failed",
				zap.Int64("author_id", post.AuthorID),
				zap.Error(err))
			continue
		}

		// 【核心新增逻辑】：反序列化图片字符串
		var imageURLs []string
		if post.ImageURL != "" && post.ImageURL != "[]" && post.ImageURL != "null" {
			if err := json.Unmarshal([]byte(post.ImageURL), &imageURLs); err != nil {
				zap.L().Warn("json.Unmarshal image_url failed", zap.Error(err))
			}
		}

		postDetail := &models.ApiPostDetail{
			AuthorName:      user.Username,
			VoteNum:         voteData[idx],
			Post:            post,
			CommunityDetail: community,
		}
		data = append(data, postDetail)

	}
	return
}

// GetPostListNew 将两个查询逻辑合二为一的函数
func GetPostListNew(p *models.ParamPostList) (data []*models.ApiPostDetail, err error) {
	// 根据请求参数的不同，执行不同的逻辑
	if p.CommunityID == 0 {
		// 查所有
		data, err = GetPostList2(p)
	} else {
		// 根据社区id查询
		data, err = GetCommunityPostList(p)
	}
	if err != nil {
		zap.L().Error("GetPostListNew failed", zap.Error(err))
	}
	return
}
