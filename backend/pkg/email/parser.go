package email

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/quotedprintable"
	"net/textproto"
	"regexp"
	"strings"
)

// EmailParser 简化的邮件解析器
type EmailParser struct{}

// NewEmailParser 创建新的邮件解析器
func NewEmailParser() *EmailParser {
	return &EmailParser{}
}

// ParseContent 解析邮件内容，提取纯文本和HTML内容
func (p *EmailParser) ParseContent(rawEmail string) (textContent, htmlContent string, err error) {
	// 分离邮件头和正文
	headerEnd := strings.Index(rawEmail, "\r\n\r\n")
	if headerEnd == -1 {
		headerEnd = strings.Index(rawEmail, "\n\n")
		if headerEnd == -1 {
			return p.fallbackParseContent(rawEmail), "", nil
		}
	}

	headerStr := rawEmail[:headerEnd]
	bodyStr := rawEmail[headerEnd+4:]
	if strings.Index(rawEmail, "\r\n\r\n") == -1 {
		bodyStr = rawEmail[headerEnd+2:]
	}

	// 解析邮件头
	headers, err := p.parseHeaders(headerStr)
	if err != nil {
		log.Printf("解析邮件头失败: %v", err)
		return p.fallbackParseContent(rawEmail), "", nil
	}

	// 获取Content-Type
	contentType := headers.Get("Content-Type")
	if contentType == "" {
		contentType = "text/plain; charset=utf-8"
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		log.Printf("解析Content-Type失败: %v", err)
		mediaType = "text/plain"
		params = make(map[string]string)
	}

	// 根据内容类型解析
	if strings.HasPrefix(mediaType, "multipart/") {
		boundary := params["boundary"]
		if boundary != "" {
			textContent, htmlContent = p.parseMultipart(bodyStr, boundary)
		} else {
			textContent = p.fallbackParseContent(rawEmail)
		}
	} else {
		// 单部分邮件
		encoding := headers.Get("Content-Transfer-Encoding")
		content := p.decodeContent(bodyStr, encoding)

		// 处理字符集
		if charset, ok := params["charset"]; ok && charset != "" {
			content = p.convertCharset(content, charset)
		}

		if strings.HasPrefix(mediaType, "text/html") {
			htmlContent = content
			textContent = p.stripHTMLTags(content)
		} else {
			textContent = content
		}
	}

	// 如果都没有内容，使用后备方法
	if textContent == "" && htmlContent == "" {
		textContent = p.fallbackParseContent(rawEmail)
	}

	return textContent, htmlContent, nil
}

// parseHeaders 解析邮件头
func (p *EmailParser) parseHeaders(headerStr string) (textproto.MIMEHeader, error) {
	reader := strings.NewReader(headerStr + "\r\n\r\n")
	tp := textproto.NewReader(bufio.NewReader(reader))
	return tp.ReadMIMEHeader()
}

// parseMultipart 解析多部分邮件
func (p *EmailParser) parseMultipart(body, boundary string) (textContent, htmlContent string) {
	// 分割各部分
	parts := strings.Split(body, "--"+boundary)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || part == "--" {
			continue
		}

		// 查找部分头的结束位置
		headerEnd := strings.Index(part, "\r\n\r\n")
		if headerEnd == -1 {
			headerEnd = strings.Index(part, "\n\n")
			if headerEnd == -1 {
				continue
			}
		}

		partHeaderStr := part[:headerEnd]
		partBodyStr := part[headerEnd+4:]
		if strings.Index(part, "\r\n\r\n") == -1 {
			partBodyStr = part[headerEnd+2:]
		}

		// 解析部分头
		partHeaders, err := p.parseHeaders(partHeaderStr)
		if err != nil {
			log.Printf("解析部分头失败: %v", err)
			continue
		}

		// 获取Content-Type
		partContentType := partHeaders.Get("Content-Type")
		if partContentType == "" {
			continue
		}

		partMediaType, partParams, err := mime.ParseMediaType(partContentType)
		if err != nil {
			log.Printf("解析部分Content-Type失败: %v", err)
			continue
		}

		// 解码内容
		encoding := partHeaders.Get("Content-Transfer-Encoding")
		content := p.decodeContent(partBodyStr, encoding)

		// 处理字符集
		if charset, ok := partParams["charset"]; ok && charset != "" {
			content = p.convertCharset(content, charset)
		}

		// 根据媒体类型分配内容
		switch partMediaType {
		case "text/plain":
			if textContent == "" {
				textContent = content
			}
		case "text/html":
			if htmlContent == "" {
				htmlContent = content
				// 如果没有纯文本版本，从HTML中提取
				if textContent == "" {
					textContent = p.stripHTMLTags(content)
				}
			}
		case "multipart/alternative":
			// 递归处理嵌套的多部分邮件
			if nestedBoundary, ok := partParams["boundary"]; ok {
				nestedText, nestedHTML := p.parseMultipart(partBodyStr, nestedBoundary)
				if textContent == "" {
					textContent = nestedText
				}
				if htmlContent == "" {
					htmlContent = nestedHTML
				}
			}
		}
	}

	return textContent, htmlContent
}

// decodeContent 根据编码方式解码内容
func (p *EmailParser) decodeContent(content, encoding string) string {
	encoding = strings.ToLower(strings.TrimSpace(encoding))

	switch encoding {
	case "base64":
		// Base64解码
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(content))
		if err != nil {
			log.Printf("Base64解码失败: %v", err)
			return content // 返回原始内容
		}
		return string(decoded)

	case "quoted-printable":
		// Quoted-Printable解码
		reader := quotedprintable.NewReader(strings.NewReader(content))
		decoded, err := io.ReadAll(reader)
		if err != nil {
			log.Printf("Quoted-Printable解码失败: %v", err)
			return content
		}
		return string(decoded)

	default:
		// 无编码或不支持的编码，直接返回
		return content
	}
}

// convertCharset 转换字符集（简化实现）
func (p *EmailParser) convertCharset(content, charset string) string {
	// 这里可以根据需要实现字符集转换
	// 对于常见的UTF-8、GBK等，Go的标准库已经能够处理
	return content
}

// stripHTMLTags 从HTML中提取纯文本
func (p *EmailParser) stripHTMLTags(html string) string {
	// 简单的HTML标签去除
	content := html

	// 处理常见的HTML实体
	content = strings.ReplaceAll(content, "&nbsp;", " ")
	content = strings.ReplaceAll(content, "&lt;", "<")
	content = strings.ReplaceAll(content, "&gt;", ">")
	content = strings.ReplaceAll(content, "&amp;", "&")
	content = strings.ReplaceAll(content, "&quot;", "\"")

	// 处理换行标签
	content = strings.ReplaceAll(content, "<br>", "\n")
	content = strings.ReplaceAll(content, "<br/>", "\n")
	content = strings.ReplaceAll(content, "<br />", "\n")
	content = strings.ReplaceAll(content, "<BR>", "\n")
	content = strings.ReplaceAll(content, "<BR/>", "\n")
	content = strings.ReplaceAll(content, "<BR />", "\n")

	// 处理段落和div标签
	content = strings.ReplaceAll(content, "<p>", "\n")
	content = strings.ReplaceAll(content, "</p>", "\n")
	content = strings.ReplaceAll(content, "<P>", "\n")
	content = strings.ReplaceAll(content, "</P>", "\n")
	content = strings.ReplaceAll(content, "<div>", "\n")
	content = strings.ReplaceAll(content, "</div>", "\n")
	content = strings.ReplaceAll(content, "<DIV>", "\n")
	content = strings.ReplaceAll(content, "</DIV>", "\n")

	// 去除所有HTML标签
	re := regexp.MustCompile(`<[^>]*>`)
	content = re.ReplaceAllString(content, "")

	// 清理多余的空行和空白字符
	lines := strings.Split(content, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}

	result := strings.Join(cleanLines, "\n")

	// 限制连续换行数量
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}

	return strings.TrimSpace(result)
}

// fallbackParseContent 后备邮件解析方法
func (p *EmailParser) fallbackParseContent(rawEmail string) string {
	lines := strings.Split(rawEmail, "\n")

	// 找到邮件头结束的位置
	var bodyStart int
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			bodyStart = i + 1
			break
		}
	}

	if bodyStart >= len(lines) {
		return ""
	}

	// 提取邮件正文
	bodyLines := lines[bodyStart:]
	bodyContent := strings.Join(bodyLines, "\n")

	// 检查是否是多部分邮件
	if strings.Contains(strings.ToLower(rawEmail), "content-type: multipart/") {
		return p.parseMultipartFallback(bodyContent)
	}

	// 检查是否是base64编码
	if strings.Contains(strings.ToLower(rawEmail), "content-transfer-encoding: base64") {
		if decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(bodyContent)); err == nil {
			return string(decoded)
		}
	}

	// 检查是否是quoted-printable编码
	if strings.Contains(strings.ToLower(rawEmail), "content-transfer-encoding: quoted-printable") {
		reader := quotedprintable.NewReader(strings.NewReader(bodyContent))
		if decoded, err := io.ReadAll(reader); err == nil {
			return string(decoded)
		}
	}

	return bodyContent
}

// parseMultipartFallback 后备多部分邮件解析
func (p *EmailParser) parseMultipartFallback(bodyContent string) string {
	// 查找边界
	parts := strings.Split(bodyContent, "--")

	for _, part := range parts {
		partLower := strings.ToLower(part)

		// 查找text/plain部分
		if strings.Contains(partLower, "content-type: text/plain") {
			return p.extractPartContent(part)
		}
	}

	// 如果没有找到text/plain，查找text/html
	for _, part := range parts {
		partLower := strings.ToLower(part)

		if strings.Contains(partLower, "content-type: text/html") {
			htmlContent := p.extractPartContent(part)
			return p.stripHTMLTags(htmlContent)
		}
	}

	return bodyContent
}

// extractPartContent 从邮件部分中提取内容
func (p *EmailParser) extractPartContent(part string) string {
	lines := strings.Split(part, "\n")

	var contentStart int
	var encoding string

	// 解析部分头部
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			contentStart = i + 1
			break
		}

		lineLower := strings.ToLower(strings.TrimSpace(line))
		if strings.HasPrefix(lineLower, "content-transfer-encoding:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				encoding = strings.TrimSpace(parts[1])
			}
		}
	}

	if contentStart >= len(lines) {
		return ""
	}

	// 提取内容
	contentLines := lines[contentStart:]
	content := strings.Join(contentLines, "\n")
	content = strings.TrimSpace(content)

	// 根据编码解码
	return p.decodeContent(content, encoding)
}

// ExtractEmailInfo 提取邮件基本信息（用于调试）
func (p *EmailParser) ExtractEmailInfo(rawEmail string) map[string]string {
	info := make(map[string]string)

	// 分离邮件头
	headerEnd := strings.Index(rawEmail, "\r\n\r\n")
	if headerEnd == -1 {
		headerEnd = strings.Index(rawEmail, "\n\n")
		if headerEnd == -1 {
			info["error"] = "无法找到邮件头结束位置"
			return info
		}
	}

	headerStr := rawEmail[:headerEnd]

	// 解析邮件头
	headers, err := p.parseHeaders(headerStr)
	if err != nil {
		info["error"] = fmt.Sprintf("解析邮件头失败: %v", err)
		return info
	}

	// 提取基本信息
	info["subject"] = headers.Get("Subject")
	info["from"] = headers.Get("From")
	info["to"] = headers.Get("To")
	info["date"] = headers.Get("Date")
	info["content_type"] = headers.Get("Content-Type")
	info["content_transfer_encoding"] = headers.Get("Content-Transfer-Encoding")

	return info
}
