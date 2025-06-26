package api

// APIResponse 统一API响应结构
type APIResponse struct {
	Code    int         `json:"code"`            // 响应状态码
	Message string      `json:"message"`         // 响应消息
	Data    interface{} `json:"data,omitempty"`  // 响应数据
	Error   string      `json:"error,omitempty"` // 错误信息
}

// SuccessResponse 成功响应
func SuccessResponse(message string, data interface{}) APIResponse {
	return APIResponse{
		Code:    200,
		Message: message,
		Data:    data,
	}
}

// ErrorResponse 错误响应
func ErrorResponse(message string, error string) APIResponse {
	return APIResponse{
		Code:    400,
		Message: message,
		Error:   error,
	}
}
