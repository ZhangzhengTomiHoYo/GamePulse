package pgsql

import "bluebell/models"

// GetPostEmbeddingsByPostID 查询某篇帖子的全部向量分片。
func GetPostEmbeddingsByPostID(postID int64) (embeddings []*models.PostEmbedding, err error) {
	sqlStr := `select id, post_id, chunk_index, chunk_text, community_id, post_create_time,
		model_name, model_version, content_hash, embedding, status, error_msg, create_time, update_time
		from post_embeddings
		where post_id = $1
		order by chunk_index asc, id asc`

	err = db.Select(&embeddings, sqlStr, postID)
	return embeddings, err
}

// CreatePostEmbeddings 批量写入向量分片；如果同一分片重复写入，则按唯一键覆盖。
func CreatePostEmbeddings(embeddings []*models.PostEmbedding) (err error) {
	if len(embeddings) == 0 {
		return nil
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	sqlStr := `insert into post_embeddings
		(post_id, chunk_index, chunk_text, community_id, post_create_time,
		 model_name, model_version, content_hash, embedding, status, error_msg)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		on conflict (post_id, chunk_index, model_name, model_version) do update set
			chunk_text = excluded.chunk_text,
			community_id = excluded.community_id,
			post_create_time = excluded.post_create_time,
			content_hash = excluded.content_hash,
			embedding = excluded.embedding,
			status = excluded.status,
			error_msg = excluded.error_msg,
			update_time = CURRENT_TIMESTAMP`

	for _, embedding := range embeddings {
		if embedding.Status == "" {
			embedding.Status = "pending"
		}

		if _, err = tx.Exec(
			sqlStr,
			embedding.PostID,
			embedding.ChunkIndex,
			embedding.ChunkText,
			embedding.CommunityID,
			embedding.PostCreateTime,
			embedding.ModelName,
			embedding.ModelVersion,
			embedding.ContentHash,
			embedding.Embedding,
			embedding.Status,
			embedding.ErrorMsg,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}
