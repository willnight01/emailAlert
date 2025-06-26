package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// WeChatNotifier 企业微信通知器
type WeChatNotifier struct {
	client *http.Client
}

// NewWeChatNotifier 创建企业微信通知器
func NewWeChatNotifier() *WeChatNotifier {
	return &WeChatNotifier{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// WeChatRobotMessage 企业微信机器人消息结构
type WeChatRobotMessage struct {
	MsgType  string               `json:"msgtype"`
	Text     *WeChatRobotText     `json:"text,omitempty"`
	Markdown *WeChatRobotMarkdown `json:"markdown,omitempty"`
	At       *WeChatRobotAt       `json:"at,omitempty"`
}

type WeChatRobotText struct {
	Content             string   `json:"content"`
	MentionedList       []string `json:"mentioned_list,omitempty"`
	MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
}

type WeChatRobotMarkdown struct {
	Content string `json:"content"`
}

type WeChatRobotAt struct {
	UserIds []string `json:"userIds,omitempty"`
	IsAtAll bool     `json:"isAtAll,omitempty"`
}

// WeChatAppMessage 企业微信应用消息结构
type WeChatAppMessage struct {
	ToUser   string             `json:"touser,omitempty"`
	ToParty  string             `json:"toparty,omitempty"`
	ToTag    string             `json:"totag,omitempty"`
	MsgType  string             `json:"msgtype"`
	AgentID  int                `json:"agentid"`
	Text     *WeChatAppText     `json:"text,omitempty"`
	TextCard *WeChatAppTextCard `json:"textcard,omitempty"`
	Safe     int                `json:"safe,omitempty"`
}

type WeChatAppText struct {
	Content string `json:"content"`
}

type WeChatAppTextCard struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url,omitempty"`
	BtnTxt      string `json:"btntxt,omitempty"`
}

// WeChatAccessTokenResponse 企业微信Access Token响应
type WeChatAccessTokenResponse struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// WeChatApiResponse 企业微信API响应
type WeChatApiResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// SendRobotMessage 发送企业微信机器人消息
func (w *WeChatNotifier) SendRobotMessage(webhookURL, key string, content string, msgType string) error {
	if webhookURL == "" {
		return fmt.Errorf("webhook URL不能为空")
	}

	var message WeChatRobotMessage

	switch msgType {
	case "text":
		message = WeChatRobotMessage{
			MsgType: "text",
			Text: &WeChatRobotText{
				Content: content,
			},
		}
	case "markdown":
		message = WeChatRobotMessage{
			MsgType: "markdown",
			Markdown: &WeChatRobotMarkdown{
				Content: content,
			},
		}
	default:
		message = WeChatRobotMessage{
			MsgType: "text",
			Text: &WeChatRobotText{
				Content: content,
			},
		}
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	resp, err := w.client.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	var result WeChatApiResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if result.ErrCode != 0 {
		return fmt.Errorf("企业微信机器人消息发送失败: %s (错误码: %d)", result.ErrMsg, result.ErrCode)
	}

	return nil
}

// SendAppMessage 发送企业微信应用消息
func (w *WeChatNotifier) SendAppMessage(corpID, secret string, agentID int, toUser, toParty, toTag string, title, content string) error {
	// 获取Access Token
	accessToken, err := w.getAccessToken(corpID, secret)
	if err != nil {
		return fmt.Errorf("获取Access Token失败: %v", err)
	}

	// 构建消息
	message := WeChatAppMessage{
		ToUser:  toUser,
		ToParty: toParty,
		ToTag:   toTag,
		MsgType: "textcard",
		AgentID: agentID,
		TextCard: &WeChatAppTextCard{
			Title:       title,
			Description: content,
		},
		Safe: 0,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	// 发送消息
	sendURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", accessToken)
	resp, err := w.client.Post(sendURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	var result WeChatApiResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if result.ErrCode != 0 {
		return fmt.Errorf("企业微信应用消息发送失败: %s (错误码: %d)", result.ErrMsg, result.ErrCode)
	}

	return nil
}

// getAccessToken 获取企业微信Access Token
func (w *WeChatNotifier) getAccessToken(corpID, secret string) (string, error) {
	tokenURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", corpID, secret)

	resp, err := w.client.Get(tokenURL)
	if err != nil {
		return "", fmt.Errorf("请求Access Token失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	var tokenResp WeChatAccessTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if tokenResp.ErrCode != 0 {
		return "", fmt.Errorf("获取Access Token失败: %s (错误码: %d)", tokenResp.ErrMsg, tokenResp.ErrCode)
	}

	return tokenResp.AccessToken, nil
}

// TestRobotConnection 测试企业微信机器人连接
func (w *WeChatNotifier) TestRobotConnection(webhookURL, key string) error {
	testMessage := "企业微信机器人连接测试 - " + time.Now().Format("2006-01-02 15:04:05")
	return w.SendRobotMessage(webhookURL, key, testMessage, "text")
}

// TestAppConnection 测试企业微信应用连接
func (w *WeChatNotifier) TestAppConnection(corpID, secret string, agentID int, toUser string) error {
	testTitle := "企业微信应用连接测试"
	testContent := "这是一条来自邮件告警平台的连接测试消息\n" +
		"发送时间: " + time.Now().Format("2006-01-02 15:04:05") + "\n" +
		"如果您收到此消息，说明企业微信应用配置正确！"

	return w.SendAppMessage(corpID, secret, agentID, toUser, "", "", testTitle, testContent)
}
