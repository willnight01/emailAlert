package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"

	"emailAlert/internal/model"
)

// AuthService 认证服务接口
type AuthService interface {
	Login(username, password string) (*model.LoginResponse, error)
	ValidateSession(sessionID string) (*model.User, error)
	LoadUsers() ([]model.User, error)
	CreateUser(user *model.User) error
	UpdateUser(username string, user *model.User) error
	DeleteUser(username string) error
	GetUser(username string) (*model.User, error)
	RevokeSession(sessionID string) error
	CreateSession(user *model.User) (string, error)
}

// SessionInfo session信息结构
type SessionInfo struct {
	User      *model.User `json:"user"`
	ExpiresAt time.Time   `json:"expires_at"`
	CreatedAt time.Time   `json:"created_at"`
}

type authService struct {
	configPath      string
	sessions        map[string]*SessionInfo // Session存储
	sessionMutex    sync.RWMutex            // 保护session map的读写锁
	sessionDuration time.Duration           // session有效期，默认24小时
}

// NewAuthService 创建认证服务
func NewAuthService(configPath string) AuthService {
	service := &authService{
		configPath:      configPath,
		sessions:        make(map[string]*SessionInfo),
		sessionDuration: 24 * time.Hour, // 默认24小时过期
	}

	// 启动session清理协程
	go service.startSessionCleanup()

	return service
}

// SimpleUser 简单用户结构（用于配置文件）
type SimpleUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// UsersConfig 用户配置文件结构
type UsersConfig struct {
	Users []SimpleUser `json:"users"`
}

// LoadUsers 从配置文件加载用户
func (s *authService) LoadUsers() ([]model.User, error) {
	var configFile string
	if filepath.IsAbs(s.configPath) {
		configFile = s.configPath
	} else {
		configFile = s.configPath
	}

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, errors.New("无法读取用户配置文件: " + configFile + " - " + err.Error())
	}

	var config UsersConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, errors.New("用户配置文件格式错误")
	}

	// 转换为model.User类型
	var users []model.User
	for _, su := range config.Users {
		users = append(users, model.User{
			Username: su.Username,
			Password: su.Password,
			Role:     su.Role,
		})
	}

	return users, nil
}

// Login 用户登录
func (s *authService) Login(username, password string) (*model.LoginResponse, error) {
	users, err := s.LoadUsers()
	if err != nil {
		return nil, errors.New("加载用户失败: " + err.Error())
	}

	// 查找用户
	var user *model.User
	for _, u := range users {
		if u.Username == username && u.Password == password {
			user = &u
			break
		}
	}

	if user == nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 创建session
	sessionID, err := s.CreateSession(user)
	if err != nil {
		return nil, errors.New("创建会话失败: " + err.Error())
	}

	return &model.LoginResponse{
		SessionID: sessionID,
		Username:  user.Username,
		Role:      user.Role,
	}, nil
}

// CreateSession 创建新的session
func (s *authService) CreateSession(user *model.User) (string, error) {
	sessionID := s.generateSessionID()
	now := time.Now()

	s.sessionMutex.Lock()
	s.sessions[sessionID] = &SessionInfo{
		User:      user,
		ExpiresAt: now.Add(s.sessionDuration),
		CreatedAt: now,
	}
	s.sessionMutex.Unlock()

	return sessionID, nil
}

// ValidateSession 验证session
func (s *authService) ValidateSession(sessionID string) (*model.User, error) {
	if sessionID == "" {
		return nil, errors.New("会话ID为空")
	}

	s.sessionMutex.RLock()
	sessionInfo, exists := s.sessions[sessionID]
	s.sessionMutex.RUnlock()

	if !exists {
		return nil, errors.New("无效的会话")
	}

	// 检查session是否过期
	if time.Now().After(sessionInfo.ExpiresAt) {
		// session已过期，删除它
		s.sessionMutex.Lock()
		delete(s.sessions, sessionID)
		s.sessionMutex.Unlock()
		return nil, errors.New("会话已过期")
	}

	// 自动续期session（每次访问都延长24小时）
	s.sessionMutex.Lock()
	sessionInfo.ExpiresAt = time.Now().Add(s.sessionDuration)
	s.sessionMutex.Unlock()

	return sessionInfo.User, nil
}

// RevokeSession 撤销session（登出）
func (s *authService) RevokeSession(sessionID string) error {
	if sessionID == "" {
		return nil // 没有session就不需要撤销
	}

	s.sessionMutex.Lock()
	defer s.sessionMutex.Unlock()

	if _, exists := s.sessions[sessionID]; !exists {
		return errors.New("会话不存在")
	}

	delete(s.sessions, sessionID)
	return nil
}

// startSessionCleanup 启动session清理协程
func (s *authService) startSessionCleanup() {
	ticker := time.NewTicker(1 * time.Hour) // 每小时清理一次
	defer ticker.Stop()

	for range ticker.C {
		s.cleanupExpiredSessions()
	}
}

// cleanupExpiredSessions 清理过期的session
func (s *authService) cleanupExpiredSessions() {
	s.sessionMutex.Lock()
	defer s.sessionMutex.Unlock()

	now := time.Now()
	for sessionID, sessionInfo := range s.sessions {
		if now.After(sessionInfo.ExpiresAt) {
			delete(s.sessions, sessionID)
		}
	}
}

// generateSessionID 生成随机session ID
func (s *authService) generateSessionID() string {
	bytes := make([]byte, 32) // 256位随机数
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CreateUser 创建用户
func (s *authService) CreateUser(user *model.User) error {
	// 读取现有用户
	users, err := s.LoadUsers()
	if err != nil {
		return err
	}

	// 检查用户名是否已存在
	for _, u := range users {
		if u.Username == user.Username {
			return errors.New("用户名已存在")
		}
	}

	// 添加新用户
	users = append(users, *user)
	return s.saveUsers(users)
}

// UpdateUser 更新用户信息
func (s *authService) UpdateUser(username string, user *model.User) error {
	users, err := s.LoadUsers()
	if err != nil {
		return err
	}

	// 查找并更新用户
	found := false
	for i, u := range users {
		if u.Username == username {
			users[i] = *user
			found = true
			break
		}
	}

	if !found {
		return errors.New("用户不存在")
	}

	return s.saveUsers(users)
}

// DeleteUser 删除用户
func (s *authService) DeleteUser(username string) error {
	users, err := s.LoadUsers()
	if err != nil {
		return err
	}

	// 查找并删除用户
	newUsers := make([]model.User, 0)
	found := false
	for _, u := range users {
		if u.Username != username {
			newUsers = append(newUsers, u)
		} else {
			found = true
		}
	}

	if !found {
		return errors.New("用户不存在")
	}

	return s.saveUsers(newUsers)
}

// GetUser 获取用户信息
func (s *authService) GetUser(username string) (*model.User, error) {
	users, err := s.LoadUsers()
	if err != nil {
		return nil, err
	}

	for _, u := range users {
		if u.Username == username {
			return &u, nil
		}
	}

	return nil, errors.New("用户不存在")
}

// saveUsers 保存用户到配置文件
func (s *authService) saveUsers(users []model.User) error {
	// 转换为配置文件格式
	var simpleUsers []SimpleUser
	for _, u := range users {
		simpleUsers = append(simpleUsers, SimpleUser{
			Username: u.Username,
			Password: u.Password,
			Role:     u.Role,
		})
	}

	config := UsersConfig{Users: simpleUsers}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(s.configPath, data, 0644)
}
