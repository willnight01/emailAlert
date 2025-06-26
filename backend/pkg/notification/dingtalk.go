package notification

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// DingTalkNotifier 钉钉通知器
type DingTalkNotifier struct {
	client *http.Client
}

// NewDingTalkNotifier 创建钉钉通知器
func NewDingTalkNotifier() *DingTalkNotifier {
	return &DingTalkNotifier{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DingTalkRobotMessage 钉钉群机器人消息结构
type DingTalkRobotMessage struct {
	MsgType  string                 `json:"msgtype"`
	Text     *DingTalkRobotText     `json:"text,omitempty"`
	Markdown *DingTalkRobotMarkdown `json:"markdown,omitempty"`
	At       *DingTalkRobotAt       `json:"at,omitempty"`
}

type DingTalkRobotText struct {
	Content string `json:"content"`
}

type DingTalkRobotMarkdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type DingTalkRobotAt struct {
	AtMobiles []string `json:"atMobiles,omitempty"`
	AtUserIds []string `json:"atUserIds,omitempty"`
	IsAtAll   bool     `json:"isAtAll,omitempty"`
}

// DingTalkWorkMessage 钉钉工作通知消息结构
type DingTalkWorkMessage struct {
	AgentId    int64                   `json:"agent_id"`
	UseridList string                  `json:"userid_list"`
	DeptIdList string                  `json:"dept_id_list,omitempty"`
	ToAllUser  bool                    `json:"to_all_user,omitempty"`
	Msg        DingTalkWorkMessageBody `json:"msg"`
}

type DingTalkWorkMessageBody struct {
	MsgType string            `json:"msgtype"`
	Text    *DingTalkWorkText `json:"text,omitempty"`
	OA      *DingTalkWorkOA   `json:"oa,omitempty"`
}

type DingTalkWorkText struct {
	Content string `json:"content"`
}

type DingTalkWorkOA struct {
	MessageUrl string             `json:"message_url,omitempty"`
	Head       DingTalkWorkOAHead `json:"head"`
	Body       DingTalkWorkOABody `json:"body"`
}

type DingTalkWorkOAHead struct {
	BgColor string `json:"bgcolor,omitempty"`
	Text    string `json:"text"`
}

type DingTalkWorkOABody struct {
	Title   string               `json:"title"`
	Form    []DingTalkWorkOAForm `json:"form,omitempty"`
	Content string               `json:"content,omitempty"`
}

type DingTalkWorkOAForm struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// DingTalkAccessTokenResponse 钉钉Access Token响应
type DingTalkAccessTokenResponse struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// DingTalkApiResponse 钉钉API响应
type DingTalkApiResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	TaskId  int64  `json:"task_id,omitempty"`
}

// SendRobotMessage 发送钉钉群机器人消息
func (d *DingTalkNotifier) SendRobotMessage(webhookURL, secret, content, title, msgType string) error {
	if webhookURL == "" {
		return fmt.Errorf("webhook URL不能为空")
	}

	// 如果有密钥，需要生成签名
	if secret != "" {
		timestamp := time.Now().UnixMilli()
		sign, err := d.generateSign(timestamp, secret)
		if err != nil {
			return fmt.Errorf("生成签名失败: %v", err)
		}

		// 添加签名参数
		u, err := url.Parse(webhookURL)
		if err != nil {
			return fmt.Errorf("解析webhook URL失败: %v", err)
		}
		q := u.Query()
		q.Set("timestamp", strconv.FormatInt(timestamp, 10))
		q.Set("sign", sign)
		u.RawQuery = q.Encode()
		webhookURL = u.String()
	}

	var message DingTalkRobotMessage

	switch msgType {
	case "text":
		message = DingTalkRobotMessage{
			MsgType: "text",
			Text: &DingTalkRobotText{
				Content: content,
			},
		}
	case "markdown":
		message = DingTalkRobotMessage{
			MsgType: "markdown",
			Markdown: &DingTalkRobotMarkdown{
				Title: title,
				Text:  content,
			},
		}
	default:
		message = DingTalkRobotMessage{
			MsgType: "text",
			Text: &DingTalkRobotText{
				Content: content,
			},
		}
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	resp, err := d.client.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	var result DingTalkApiResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if result.ErrCode != 0 {
		return fmt.Errorf("钉钉群机器人消息发送失败: %s (错误码: %d)", result.ErrMsg, result.ErrCode)
	}

	return nil
}

// SendWorkMessage 发送钉钉工作通知
func (d *DingTalkNotifier) SendWorkMessage(appKey, appSecret string, agentId int64, userIds, title, content string) error {
	// 获取Access Token
	accessToken, err := d.getAccessToken(appKey, appSecret)
	if err != nil {
		return fmt.Errorf("获取Access Token失败: %v", err)
	}

	// 构建OA消息
	message := DingTalkWorkMessage{
		AgentId:    agentId,
		UseridList: userIds,
		Msg: DingTalkWorkMessageBody{
			MsgType: "oa",
			OA: &DingTalkWorkOA{
				Head: DingTalkWorkOAHead{
					BgColor: "FFBBBBBB",
					Text:    title,
				},
				Body: DingTalkWorkOABody{
					Title:   title,
					Content: content,
				},
			},
		},
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	// 发送消息
	sendURL := fmt.Sprintf("https://oapi.dingtalk.com/topapi/message/corpconversation/asyncsend_v2?access_token=%s", accessToken)
	resp, err := d.client.Post(sendURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	var result DingTalkApiResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if result.ErrCode != 0 {
		return fmt.Errorf("钉钉工作通知发送失败: %s (错误码: %d)", result.ErrMsg, result.ErrCode)
	}

	return nil
}

// getAccessToken 获取钉钉Access Token
func (d *DingTalkNotifier) getAccessToken(appKey, appSecret string) (string, error) {
	tokenURL := fmt.Sprintf("https://oapi.dingtalk.com/gettoken?appkey=%s&appsecret=%s", appKey, appSecret)

	resp, err := d.client.Get(tokenURL)
	if err != nil {
		return "", fmt.Errorf("请求Access Token失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	var tokenResp DingTalkAccessTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if tokenResp.ErrCode != 0 {
		return "", fmt.Errorf("获取Access Token失败: %s (错误码: %d)", tokenResp.ErrMsg, tokenResp.ErrCode)
	}

	return tokenResp.AccessToken, nil
}

// generateSign 生成钉钉机器人签名
func (d *DingTalkNotifier) generateSign(timestamp int64, secret string) (string, error) {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return url.QueryEscape(signature), nil
}

// TestRobotConnection 测试钉钉群机器人连接
func (d *DingTalkNotifier) TestRobotConnection(webhookURL, secret string) error {
	testTitle := "钉钉群机器人连接测试"
	testContent := "这是一条来自邮件告警平台的连接测试消息\n" +
		"发送时间: " + time.Now().Format("2006-01-02 15:04:05") + "\n" +
		"如果您收到此消息，说明钉钉群机器人配置正确！"

	return d.SendRobotMessage(webhookURL, secret, testContent, testTitle, "text")
}

// TestWorkConnection 测试钉钉工作通知连接
func (d *DingTalkNotifier) TestWorkConnection(appKey, appSecret string, agentId int64, userIds string) error {
	testTitle := "钉钉工作通知连接测试"
	testContent := "这是一条来自邮件告警平台的连接测试消息。\n" +
		"发送时间: " + time.Now().Format("2006-01-02 15:04:05") + "\n" +
		"如果您收到此消息，说明钉钉工作通知配置正确！"

	return d.SendWorkMessage(appKey, appSecret, agentId, userIds, testTitle, testContent)
}
