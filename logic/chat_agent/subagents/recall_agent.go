package subagents

//func RecallAgent() adk.Agent {
//	a, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
//		Name:        "RecallAgent",
//		Description: "决定当前消息是否调用向量检索工具",
//		Instruction: `
//你是一个游戏社区舆情智能体，你负责决定当前对话是否需要向量化，并在向量数据库中检索相关内容
//
//`,
//		Model: model.NewChatModel(),
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	return a
//}
