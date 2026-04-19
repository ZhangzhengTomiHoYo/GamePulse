package pgsql

import (
	"bluebell/models"
	"time"
)

// GetPostAnalysisByPostID 根据帖子业务 ID 查询分析结果。
func GetPostAnalysisByPostID(postID int64) (analysis *models.PostAnalysis, err error) {
	analysis = new(models.PostAnalysis)
	sqlStr := `select post_id, sentiment_label, sentiment_score, risk_level, topics, keywords, summary, analyzed_at
		from post_analysis
		where post_id = $1`
	err = db.Get(analysis, sqlStr, postID)
	return analysis, err
}

// UpsertPostAnalysis 写入或更新帖子分析结果。
func UpsertPostAnalysis(analysis *models.PostAnalysis) error {
	analyzedAt := analysis.AnalyzedAt
	if analyzedAt.IsZero() {
		analyzedAt = time.Now()
	}

	sqlStr := `insert into post_analysis
		(post_id, sentiment_label, sentiment_score, risk_level, topics, keywords, summary, analyzed_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
		on conflict (post_id) do update set
			sentiment_label = excluded.sentiment_label,
			sentiment_score = excluded.sentiment_score,
			risk_level = excluded.risk_level,
			topics = excluded.topics,
			keywords = excluded.keywords,
			summary = excluded.summary,
			analyzed_at = excluded.analyzed_at`

	_, err := db.Exec(
		sqlStr,
		analysis.PostID,
		analysis.SentimentLabel,
		analysis.SentimentScore,
		analysis.RiskLevel,
		analysis.Topics,
		analysis.Keywords,
		analysis.Summary,
		analyzedAt,
	)
	if err != nil {
		return err
	}

	analysis.AnalyzedAt = analyzedAt
	return nil
}
