package logic

import (
	"bluebell/models"
	"bluebell/setting"
	"errors"
	"strings"

	"go.uber.org/zap"
)

func EmbedPostAsync(post *models.Post) error {
	if post == nil {
		return errors.New("post is nil")
	}
	if strings.TrimSpace(post.Title) == "" && strings.TrimSpace(post.Content) == "" {
		return errors.New("post title and content are empty")
	}
	if setting.Conf == nil || setting.Conf.LLMConfig == nil {
		return errors.New("llm config not initialized")
	}
	if strings.TrimSpace(setting.Conf.LLMConfig.APIKey) == "" {
		return errors.New("llm api key is empty")
	}

	postID := post.ID
	communityID := post.CommunityID
	title := post.Title
	content := post.Content

	go func() {
		if err := embedAndSavePost(postID, communityID, title, content); err != nil {
			zap.L().Error("analyzeAndSavePost failed",
				zap.Int64("postID", postID),
				zap.Int64("communityID", communityID),
				zap.Error(err))
		}
	}()

	return nil
}

func embedAndSavePost() {

}
