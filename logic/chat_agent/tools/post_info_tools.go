package tools

import (
	"context"
	"gamepulse/dao/pgsql"
	"gamepulse/models"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
)

func ChatPostInfoTool() tool.InvokableTool {
	return utils.NewTool(
		&schema.ToolInfo{
			Name:  "search_community_posts", // 改个更贴切的名字
			Desc:  "根据社区名称查询最近的帖子列表，用于分析社区舆论或热门话题",
			Extra: nil,
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"community_name": {
					Type: schema.String,
					Desc: "社区的名称，例如 '鸣潮'、'原神' 等",
				},
				"sentiment_label": {
					Type: schema.String,
					Desc: "帖子的情绪，只有以下四种 'positive','negative','neutral','all'",
				},
				"last_time": {
					Type: schema.Integer,
					Desc: "查询近几日的帖子，根据用户的语义决定，默认为7天",
				},
			}),
		}, getPostInfo)
}

func getPostInfo(ctx context.Context, post models.ChatPostBackend) ([]*models.ChatPostBackendResult, error) {
	resultList, err := pgsql.ChatGetPostListByCommunity(post)
	if err != nil {
		zap.L().Error("pgsql.ChatGetPostListByCommunity error!", zap.Error(err))
		return nil, err
	}
	return resultList, nil
}
