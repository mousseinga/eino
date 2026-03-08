package tools

// GetDefaultTools 获取默认工具列表
func GetDefaultTools() []interface{} {
	return []interface{}{
		NewEchoTool(),
		NewTimeTool(),
		NewCalculatorTool(),
		NewTextProcessorTool(),
	}
}
