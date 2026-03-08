package tool

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// PDFToTextRequest 大模型调用工具的入参结构体（明确参数要求）
type PDFToTextRequest struct {
	FilePath string `json:"file_path" jsonschema:"required,description=本地PDF文件的绝对路径（例如：D:\\test\\document.pdf 或 /home/user/document.pdf）"`
	ToPages  bool   `json:"to_pages" jsonschema:"default=false,description=是否按页面分割文本（true=分页输出，false=合并所有页为一个文本，默认false）"`
}

// PDFToTextResult 工具返回的结构化结果（大模型可直接解析）
type PDFToTextResult struct {
	Success    bool                   `json:"success" jsonschema:"description=解析是否成功"`
	Content    string                 `json:"content,omitempty" jsonschema:"description=合并后的纯文本（ToPages=false时返回）"`
	Pages      []PDFPageText          `json:"pages,omitempty" jsonschema:"description=分页文本（ToPages=true时返回）"`
	TotalPages int                    `json:"total_pages" jsonschema:"description=PDF总页数"`
	ErrorMsg   string                 `json:"error_msg,omitempty" jsonschema:"description=错误信息（失败时返回）"`
	Meta       map[string]interface{} `json:"meta,omitempty" jsonschema:"description=元数据（方便追溯）"`
}

// PDFPageText 单页文本结构（分页模式下使用）
type PDFPageText struct {
	PageNum int    `json:"page_num" jsonschema:"description=页码（从1开始）"`
	Content string `json:"content" jsonschema:"description=单页纯文本"`
}

// ConvertPDFToText 核心逻辑：PDF转纯文本（工具执行入口）
func ConvertPDFToText(ctx context.Context, req *PDFToTextRequest) (*PDFToTextResult, error) {
	result := PDFToTextResult{
		Meta: map[string]interface{}{
			"file_path":  req.FilePath,
			"to_pages":   req.ToPages,
			"parse_time": time.Now().Format("2006-01-02 15:04:05"),
			"method":     "pdftotext_cli", // 标记使用 CLI 方式
		},
	}
	// 1. 参数校验
	if req.FilePath == "" {
		result.Success = false
		result.ErrorMsg = "参数错误：必须传入 file_path"
		return &result, errors.New(result.ErrorMsg)
	}

	// 2. 检查文件是否存在
	if _, err := os.Stat(req.FilePath); err != nil {
		result.Success = false
		result.ErrorMsg = fmt.Sprintf("文件不存在或无法访问: %v", err)
		return &result, errors.New(result.ErrorMsg)
	}

	log.Printf("[ConvertPDFToText] 开始调用 pdftotext 解析文件: %s", req.FilePath)
	startTime := time.Now()

	// 3. 使用 pdftotext CLI 工具解析
	// 命令：pdftotext -layout <filepath> -
	// "-" 表示输出到 stdout
	cmd := exec.CommandContext(ctx, "pdftotext", "-layout", req.FilePath, "-")

	// 设置缓冲区捕获输出
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	// 执行命令
	err := cmd.Run()

	duration := time.Since(startTime)
	if err != nil {
		log.Printf("[ConvertPDFToText] pdftotext 执行失败 (耗时 %v): %v, Stderr: %s", duration, err, errBuf.String())

		// 检查是否是 context 超时
		if ctx.Err() == context.DeadlineExceeded {
			result.ErrorMsg = "PDF解析超时"
		} else {
			result.ErrorMsg = fmt.Sprintf("PDF工具执行失败: %v, 详情: %s", err, errBuf.String())
		}

		result.Success = false
		return &result, errors.New(result.ErrorMsg)
	}

	content := outBuf.String()
	log.Printf("[ConvertPDFToText] pdftotext 执行成功 (耗时 %v), 获取到 %d 字符", duration, len(content))

	// 4. 构造成功结果
	result.Success = true
	//由于 pdftotext CLI 一次输出所有文本，我们暂时简化处理：
	// 即使 ToPages=true，也只在 TotalPages 中返回 1，并在 Content 中返回全部内容。
	// 大模型通常只需要全文即可。
	if req.ToPages {
		result.Pages = []PDFPageText{{PageNum: 1, Content: content}}
		result.TotalPages = 1
	} else {
		result.Content = content
		result.TotalPages = 1 // 无法精确获取页数，设为1
	}

	return &result, nil
}

// CreatePDFToTextTool 创建工具实例（供Eino框架注册，大模型识别）
func CreatePDFToTextTool() tool.InvokableTool {
	// 工具元信息：大模型识别的关键（名称、描述、参数定义）
	pdfTool, err := utils.InferTool("pdf_to_text", "将本地PDF文件转换为纯文本，仅支持文本型PDF（可复制文字），不支持扫描件、加密PDF。需传入本地PDF的绝对路径，可选择按页面分割或合并所有页。", ConvertPDFToText)
	if err != nil {
		log.Fatalf("infer tool failed: %v", err)
	}
	fmt.Println("✅ PDF转纯文本工具初始化完成（大模型可调用）")
	return pdfTool
}
