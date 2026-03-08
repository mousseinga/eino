package tool

import (
	"ai-eino-interview-agent/internal/model"
	"context"
	"log"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// GetMianshiInfoRequest 获取面试对话记录的请求结构体
type GetMianshiInfoRequest struct {
	UserID   uint   `json:"user_id" jsonschema:"description=用户ID"`
	ReportID uint64 `json:"report_id" jsonschema:"description=报告ID"`
}

// GetMianshiInfoResponse 获取面试对话记录的响应结构体
type GetMianshiInfoResponse struct {
	Data []model.InterviewDialogue `json:"data" jsonschema:"description=面试对话记录列表"`
}

// GetMianshiInfo 获取面试对话记录
func GetMianshiInfo(_ context.Context, req *GetMianshiInfoRequest) (*GetMianshiInfoResponse, error) {
	if req == nil {
		return nil, nil
	}
	data, err := model.InterviewDialogueDao.GetInterviewDialoguesByUserIdAndRecordId(req.UserID, req.ReportID)
	if err != nil {
		log.Printf("get interview dialogues failed: %v", err)
		return nil, err
	}
	return &GetMianshiInfoResponse{
		Data: *data,
	}, nil
}

// GetMianshiInfoTool 创建获取面试对话记录的工具
func GetMianshiInfoTool() tool.InvokableTool {
	t, err := utils.InferTool(
		"get_mianshi_info",
		"获取用户的面试对话记录，包括所有问题和回答",
		GetMianshiInfo,
	)
	if err != nil {
		log.Fatalf("infer tool failed: %v", err)
	}
	return t
}
