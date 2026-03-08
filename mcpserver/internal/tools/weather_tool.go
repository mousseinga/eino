package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// WeatherTool 天气查询工具
type WeatherTool struct {
	apiKey string
	client *http.Client
}

// NewWeatherTool 创建新的天气工具实例
func NewWeatherTool(apiKey string) *WeatherTool {
	return &WeatherTool{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Name 返回工具名称
func (t *WeatherTool) Name() string {
	return "get_weather"
}

// Description 返回工具描述
func (t *WeatherTool) Description() string {
	return "Get current weather information for a city using OpenWeatherMap API. Returns temperature, humidity, description, and more."
}

// InputSchema 返回工具的输入参数 schema
func (t *WeatherTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"city": map[string]interface{}{
				"type":        "string",
				"description": "The name of the city to get weather for (e.g., 'Beijing', 'New York', 'London')",
			},
			"units": map[string]interface{}{
				"type":        "string",
				"description": "Temperature units: 'metric' (Celsius), 'imperial' (Fahrenheit), or 'kelvin' (default)",
				"enum":        []string{"metric", "imperial", "kelvin"},
				"default":     "metric",
			},
		},
		"required": []string{"city"},
	}
}

// Execute 执行天气查询
func (t *WeatherTool) Execute(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	// 验证 API Key
	if t.apiKey == "" {
		return mcp.NewToolResultError("API Key is not configured. Please configure the API key in the server."), nil
	}

	// 验证 API Key 格式（OpenWeatherMap API Key 通常是 32 个字符）
	if len(t.apiKey) < 20 {
		return mcp.NewToolResultError("API Key appears to be invalid (too short). Please check your OpenWeatherMap API key."), nil
	}

	// 获取城市名称
	city, ok := arguments["city"].(string)
	if !ok || city == "" {
		return mcp.NewToolResultError("Invalid argument: city must be a non-empty string"), nil
	}

	// 获取单位（默认为 metric）
	units := "metric"
	if u, ok := arguments["units"].(string); ok && u != "" {
		units = u
	}

	// URL 编码城市名称（处理空格和特殊字符）
	encodedCity := url.QueryEscape(city)

	// 构建 API URL
	apiURL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&units=%s&appid=%s",
		encodedCity, units, t.apiKey)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create request: %v", err)), nil
	}

	// 发送请求
	resp, err := t.client.Do(req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch weather data: %v", err)), nil
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read response: %v", err)), nil
	}

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		// 尝试解析 JSON 错误响应
		var errorResp map[string]interface{}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			if msg, ok := errorResp["message"].(string); ok {
				return mcp.NewToolResultError(fmt.Sprintf("Weather API error: %s", msg)), nil
			}
		}

		// 如果返回的是 HTML（通常是 400 错误），提取关键信息
		bodyStr := string(body)
		if len(bodyStr) > 500 {
			bodyStr = bodyStr[:500] + "..."
		}

		// 检查是否是 HTML 响应
		if resp.StatusCode == 400 && (len(body) == 0 || bodyStr[0] == '<') {
			return mcp.NewToolResultError(fmt.Sprintf("Weather API returned status 400 (Bad Request). This usually means the API key is invalid or the request format is incorrect. Please check your API key and try again.")), nil
		}

		return mcp.NewToolResultError(fmt.Sprintf("Weather API returned status %d: %s", resp.StatusCode, bodyStr)), nil
	}

	// 解析响应
	var weatherData map[string]interface{}
	if err := json.Unmarshal(body, &weatherData); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse weather data: %v", err)), nil
	}

	// 格式化响应
	result := formatWeatherResponse(weatherData, units)
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}

// formatWeatherResponse 格式化天气响应数据
func formatWeatherResponse(data map[string]interface{}, units string) map[string]interface{} {
	result := make(map[string]interface{})

	// 提取主要信息
	if main, ok := data["main"].(map[string]interface{}); ok {
		if temp, ok := main["temp"].(float64); ok {
			result["temperature"] = temp
		}
		if feelsLike, ok := main["feels_like"].(float64); ok {
			result["feels_like"] = feelsLike
		}
		if humidity, ok := main["humidity"].(float64); ok {
			result["humidity"] = humidity
		}
		if pressure, ok := main["pressure"].(float64); ok {
			result["pressure"] = pressure
		}
	}

	// 提取天气描述
	if weather, ok := data["weather"].([]interface{}); ok && len(weather) > 0 {
		if w, ok := weather[0].(map[string]interface{}); ok {
			if desc, ok := w["description"].(string); ok {
				result["description"] = desc
			}
			if main, ok := w["main"].(string); ok {
				result["condition"] = main
			}
		}
	}

	// 提取城市信息
	if name, ok := data["name"].(string); ok {
		result["city"] = name
	}
	if sys, ok := data["sys"].(map[string]interface{}); ok {
		if country, ok := sys["country"].(string); ok {
			result["country"] = country
		}
	}

	// 提取风速
	if wind, ok := data["wind"].(map[string]interface{}); ok {
		if speed, ok := wind["speed"].(float64); ok {
			result["wind_speed"] = speed
		}
	}

	// 添加单位信息
	result["units"] = units
	if units == "metric" {
		result["temperature_unit"] = "°C"
		result["wind_speed_unit"] = "m/s"
	} else if units == "imperial" {
		result["temperature_unit"] = "°F"
		result["wind_speed_unit"] = "mph"
	} else {
		result["temperature_unit"] = "K"
		result["wind_speed_unit"] = "m/s"
	}

	return result
}
