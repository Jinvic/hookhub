package ech0

func subjectFor(topic string) string {
	switch topic {
	case TopicUserCreated:
		return "用户创建"
	case TopicUserUpdated:
		return "用户更新"
	case TopicUserDeleted:
		return "用户删除"
	case TopicEchoCreated:
		return "动态发布"
	case TopicEchoUpdated:
		return "动态编辑"
	case TopicEchoDeleted:
		return "动态删除"
	case TopicCommentCreated:
		return "评论创建"
	case TopicCommentStatusUpdated:
		return "评论状态变更"
	case TopicCommentDeleted:
		return "评论删除"
	case TopicResourceUploaded:
		return "资源上传"
	case TopicSystemSnapshot:
		return "系统快照"
	case TopicSystemExport:
		return "系统导出"
	case TopicSystemSnapshotSchedUpdate:
		return "快照计划更新"
	case TopicWebhookTest:
		return "Webhook 测试"
	default:
		return topic
	}
}
