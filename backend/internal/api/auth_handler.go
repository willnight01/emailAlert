package api

import (
	"net/http"

	"emailAlert/internal/model"
	"emailAlert/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("请求参数错误", err.Error()))
		return
	}

	response, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse("登录失败", err.Error()))
		return
	}

	// 设置Cookie (HttpOnly, Secure在生产环境中, SameSite策略)
	c.SetCookie(
		"session_id",       // cookie名称
		response.SessionID, // cookie值
		24*60*60,           // 最大生存时间(秒) - 24小时
		"/",                // 路径
		"",                 // 域名（空表示当前域名）
		false,              // secure（生产环境建议为true）
		true,               // httpOnly
	)

	c.JSON(http.StatusOK, SuccessResponse("登录成功", response))
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从Cookie获取session ID
	sessionID, err := c.Cookie("session_id")
	if err == nil && sessionID != "" {
		// 撤销session
		h.authService.RevokeSession(sessionID)
	}

	// 清除Cookie
	c.SetCookie(
		"session_id",
		"",
		-1, // 过期时间设为负数，立即过期
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, SuccessResponse("登出成功", nil))
}

// Profile 获取用户信息
func (h *AuthHandler) Profile(c *gin.Context) {
	// 从Cookie获取session ID
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse("未找到会话信息", ""))
		return
	}

	user, err := h.authService.ValidateSession(sessionID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse("会话无效", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse("获取用户信息成功", gin.H{
		"username": user.Username,
		"role":     user.Role,
	}))
}

// LoginHandler 登录处理器
func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	response, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"data":    response,
	})
}

// GetUsersHandler 获取用户列表（仅管理员）
func (h *AuthHandler) GetUsersHandler(c *gin.Context) {
	// 检查当前用户是否为管理员
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	currentUser := user.(*model.User)
	if currentUser.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	users, err := h.authService.LoadUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 不返回密码信息
	var safeUsers []gin.H
	for _, u := range users {
		safeUsers = append(safeUsers, gin.H{
			"username": u.Username,
			"role":     u.Role,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取用户列表成功",
		"data":    safeUsers,
	})
}

// CreateUserHandler 创建用户（仅管理员）
func (h *AuthHandler) CreateUserHandler(c *gin.Context) {
	// 检查当前用户是否为管理员
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	currentUser := user.(*model.User)
	if currentUser.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	newUser := &model.User{
		Username: req.Username,
		Password: req.Password,
		Role:     req.Role,
	}

	err := h.authService.CreateUser(newUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "创建用户成功"})
}

// UpdateUserHandler 更新用户（仅管理员）
func (h *AuthHandler) UpdateUserHandler(c *gin.Context) {
	// 检查当前用户是否为管理员
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	currentUser := user.(*model.User)
	if currentUser.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名不能为空"})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	updatedUser := &model.User{
		Username: req.Username,
		Password: req.Password,
		Role:     req.Role,
	}

	err := h.authService.UpdateUser(username, updatedUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新用户成功"})
}

// DeleteUserHandler 删除用户（仅管理员）
func (h *AuthHandler) DeleteUserHandler(c *gin.Context) {
	// 检查当前用户是否为管理员
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	currentUser := user.(*model.User)
	if currentUser.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}

	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名不能为空"})
		return
	}

	// 不能删除自己
	if username == currentUser.Username {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能删除自己"})
		return
	}

	err := h.authService.DeleteUser(username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除用户成功"})
}
