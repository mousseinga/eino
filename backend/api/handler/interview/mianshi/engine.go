package mianshi

import (
	"ai-eino-interview-agent/internal/mq"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"ai-eino-interview-agent/chatApp/agent_service/interview"
	"ai-eino-interview-agent/internal/model"
	interviewservice "ai-eino-interview-agent/internal/service/interviews"
)

// InterviewDialogueData 对话数据结构
type InterviewDialogueData struct {
	Question string
	Answer   string
}

// InterviewEngine 面试引擎 - 处理核心面试逻辑
type InterviewEngine struct {
	sessionManager *SessionManager
	interviewSvc   interviewservice.InterviewManager
	writer         io.Writer
}

// NewInterviewEngine 创建面试引擎
func NewInterviewEngine(sessionManager *SessionManager, interviewSvc interviewservice.InterviewManager, writer io.Writer) *InterviewEngine {
	return &InterviewEngine{
		sessionManager: sessionManager,
		interviewSvc:   interviewSvc,
		writer:         writer,
	}
}

// RunInterviewLoop 运行面试循环
// 新逻辑：逐个生成问题，每次生成一道，用户回答后再生成下一道
// 保留前5道题的历史作为上下文
func (e *InterviewEngine) RunInterviewLoop(ctx context.Context, session *InterviewSession) {
	const answerTimeout = 30 * time.Minute
	const heartbeatInterval = 15 * time.Second
	const maxQuestions = 20      // 最多生成30道问题
	const historyContextSize = 5 // 保留前5道题作为历史上下文

	// 创建智能体服务
	agentSvc := interview.NewInterviewAgentService(session.UserID)

	// 确定智能体类型
	agentType := e.selectAgentType(session)

	// 用于存储所有问题和回答
	var allDialogues []*InterviewDialogueData

	// 用于存储最近的历史记录（前5道题）
	type HistoryItem struct {
		Question string
		Answer   string
	}
	var recentHistory []HistoryItem

	// 循环生成30道问题
	for questionIndex := 1; questionIndex <= maxQuestions; questionIndex++ {
		select {
		case <-ctx.Done():
			log.Printf("[Interview Engine] Context cancelled, sessionID: %s, questions generated: %d", session.SessionID, questionIndex-1)
			return
		default:
		}

		// 构建提示词
		var prompt string
		if questionIndex == 1 {
			// 第一道问题：只需要简历ID和难度
			prompt = fmt.Sprintf(`请根据简历ID和难度等级生成一道面试问题。

简历ID: %d
难度等级: %s

要求：
1. 生成一道技术面试问题
2. 返回JSON格式

返回格式：
{
  "question_text": "问题内容"
}`, session.ResumeId, session.Difficulty)
		} else {
			// 后续问题：包含最近5道题的历史上下文
			historyText := ""
			for i, h := range recentHistory {
				historyText += fmt.Sprintf("问题%d：%s\n回答%d：%s\n\n", i+1, h.Question, i+1, h.Answer)
			}

			prompt = fmt.Sprintf(`根据简历ID、难度等级和最近的问答历史，生成下一道面试问题。

简历ID: %d
难度等级: %s

最近的问答历史（前%d道题）：
%s

要求：
1. 根据用户的回答情况，生成更有针对性的下一道问题
2. 避免重复之前问过的问题
3. 逐步深化问题难度
4. 返回JSON格式

返回格式：
{
  "question_text": "问题内容"
}`, session.ResumeId, session.Difficulty, len(recentHistory), historyText)
		}

		// 调用智能体生成一道问题
		var questionResult map[string]interface{}
		err := agentSvc.RunInterviewWithCallback(ctx, agentType, session.HasResume, prompt, func(message string) error {
			// 解析响应
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(message), &result); err == nil {
				questionResult = result
			}
			return nil
		})

		if err != nil {
			log.Printf("[Interview Engine] Failed to generate question %d: %v, sessionID: %s", questionIndex, err, session.SessionID)
			SendErrorEvent(e.writer, fmt.Sprintf("Failed to generate question %d: %s", questionIndex, err.Error()))
			break
		}

		if len(questionResult) == 0 {
			log.Printf("[Interview Engine] Agent returned empty result for question %d, sessionID: %s", questionIndex, session.SessionID)
			SendErrorEvent(e.writer, fmt.Sprintf("Agent returned empty result for question %d", questionIndex))
			break
		}

		// 提取问题文本
		questionText, ok := questionResult["question_text"].(string)
		if !ok || questionText == "" {
			log.Printf("[Interview Engine] Failed to extract question text for question %d, sessionID: %s", questionIndex, session.SessionID)
			SendErrorEvent(e.writer, fmt.Sprintf("Failed to extract question %d", questionIndex))
			break
		}

		// 发送问题事件
		err = SendSSEEvent(e.writer, map[string]interface{}{
			"type":  "question",
			"index": questionIndex,
			"total": maxQuestions,
			"data": map[string]interface{}{
				"question_text": questionText,
			},
		})
		if err != nil {
			log.Printf("[Interview Engine] Failed to send question event: %v", err)
			return
		}

		// 发送就绪事件
		SendReadyEventWithSession(e.writer, questionIndex, session.SessionID)
		e.sessionManager.ClearAnswer(session.SessionID)

		// 等待用户回答
		log.Printf("[Interview Engine] Waiting for answer, sessionID: %s, question: %d/%d", session.SessionID, questionIndex, maxQuestions)
		answer, received := WaitForAnswerWithHeartbeat(ctx, e.sessionManager, session.SessionID, answerTimeout, heartbeatInterval, e.writer)
		if !received {
			log.Printf("[Interview Engine] Answer timeout, sessionID: %s, question: %d", session.SessionID, questionIndex)
			SendErrorEvent(e.writer, fmt.Sprintf("Question %d timeout", questionIndex))
			break
		}

		// 保存当前问题和回答
		dialogue := &InterviewDialogueData{
			Question: questionText,
			Answer:   answer,
		}
		allDialogues = append(allDialogues, dialogue)

		// 更新会话中的问题计数
		session.QuestionCount = int32(questionIndex)

		// 更新最近的历史记录（保留最近5道题）
		recentHistory = append(recentHistory, HistoryItem{
			Question: questionText,
			Answer:   answer,
		})
		if len(recentHistory) > historyContextSize {
			recentHistory = recentHistory[len(recentHistory)-historyContextSize:]
		}

		// 发送进度事件
		err = SendSSEEvent(e.writer, map[string]interface{}{
			"type":     "answer_received",
			"index":    questionIndex,
			"total":    maxQuestions,
			"progress": float64(questionIndex) / float64(maxQuestions) * 100,
		})
		if err != nil {
			log.Printf("[Interview Engine] Failed to send answer_received event: %v", err)
		}

		log.Printf("[Interview Engine] Question %d answered, sessionID: %s", questionIndex, session.SessionID)
	}

	// 所有问题都已回答，保存到数据库
	log.Printf("[Interview Engine] All questions answered, saving %d questions to database, sessionID: %s", len(allDialogues), session.SessionID)
	err := e.saveAllDialogues(ctx, session, allDialogues)
	if err != nil {
		log.Printf("[Interview Engine] Failed to save dialogues: %v, sessionID: %s", err, session.SessionID)
		SendErrorEvent(e.writer, "Failed to save interview data: "+err.Error())
		SendCompleteEvent(e.writer)
		return
	}

	// 发送完成事件
	SendCompleteEvent(e.writer)

	// 发布评估报告生成消息
	if err := mq.PublishEvaluationReport(ctx, session.UserID, session.RecordID); err != nil {
		log.Printf("[Interview Loop] Failed to publish evaluation report message: %v, sessionID: %s", err, session.SessionID)
	}

	// 发布主题评估消息
	if err := mq.PublishTopicEvaluation(ctx, session.UserID, session.RecordID); err != nil {
		log.Printf("[Interview Loop] Failed to publish topic evaluation message: %v, sessionID: %s", err, session.SessionID)
	}
}

// saveAllDialogues 保存所有30道问题到数据库
// 新逻辑：直接保存所有问题，不再区分主问题和追问
func (e *InterviewEngine) saveAllDialogues(ctx context.Context, session *InterviewSession, questions []*InterviewDialogueData) error {
	log.Printf("[Interview Engine] Saving %d questions to database, sessionID: %s", len(questions), session.SessionID)

	// 逐个保存每道问题
	for i, q := range questions {
		dialogue := &model.InterviewDialogue{
			UserID:    session.UserID,
			ReportID:  session.RecordID,
			Question:  q.Question,
			Answer:    q.Answer,
			CreatedAt: time.Now(),
		}

		if err := model.InterviewDialogueDao.Create(dialogue); err != nil {
			return fmt.Errorf("failed to save question %d: %w", i+1, err)
		}

		if (i+1)%10 == 0 {
			log.Printf("[Interview Engine] Saved %d/%d questions, sessionID: %s", i+1, len(questions), session.SessionID)
		}
	}

	log.Printf("[Interview Engine] Successfully saved all %d questions to database, sessionID: %s", len(questions), session.SessionID)
	return nil
}

// selectAgentType 根据面试类型和领域选择智能体类型
func (e *InterviewEngine) selectAgentType(session *InterviewSession) interview.InterviewAgentType {
	// 综合面试
	if session.Type == "综合面试" {
		switch session.Domain {
		case "校招简历面试":
			return interview.ComprehensiveSchool
		case "社招简历面试":
			return interview.ComprehensiveSocial
		default:
			// 社招简历面试为默认选项
			return interview.ComprehensiveSocial
		}
	}

	// 专项面试
	switch session.Domain {
	case "Java":
		return interview.SpecializedJava
	case "MQ":
		return interview.SpecializedMQ
	case "MySQL":
		return interview.SpecializedMySQL
	case "Redis":
		return interview.SpecializedRedis
	case "Go":
		fallthrough
	default:
		return interview.SpecializedGo
	}
}
