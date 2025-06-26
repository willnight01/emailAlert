<template>
  <div class="login-container">
    <!-- 背景动画元素 -->
    <div class="bg-animation">
      <div class="floating-shape shape-1"></div>
      <div class="floating-shape shape-2"></div>
      <div class="floating-shape shape-3"></div>
      <div class="floating-shape shape-4"></div>
    </div>
    
    <div class="login-card">
      <div class="login-header">
        <div class="logo-container">
          <img src="/logo.png" alt="Logo" class="login-logo" />
          <div class="logo-glow"></div>
        </div>
        <h1 class="app-title">EmailAlert</h1>
        <h2 class="app-subtitle">邮件告警平台</h2>
        <p class="login-desc">请登录您的账户</p>
      </div>
      
      <el-form
        ref="loginFormRef"
        :model="loginForm"
        :rules="loginRules"
        class="login-form"
        @submit.prevent="handleLogin"
      >
        <el-form-item prop="username" class="form-item-animated">
          <el-input
            v-model="loginForm.username"
            placeholder="用户名"
            size="large"
            prefix-icon="User"
            :disabled="loading"
            class="animated-input"
          />
        </el-form-item>
        
        <el-form-item prop="password" class="form-item-animated">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="密码"
            size="large"
            prefix-icon="Lock"
            :disabled="loading"
            show-password
            @keyup.enter="handleLogin"
            class="animated-input"
          />
        </el-form-item>
        
        <el-form-item class="form-item-animated">
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            @click="handleLogin"
            class="login-button"
          >
            <span v-if="!loading">登录</span>
            <span v-else class="loading-text">登录中...</span>
          </el-button>
        </el-form-item>
      </el-form>
      
      <div class="login-footer">
        <div class="copyright-info">
          <p class="version">v1.0.0</p>
          <p class="maintainer">维护者：Eason Zhang</p>
          <p class="github-link">
            <a href="https://github.com/wilnight01/emailAlert" target="_blank" rel="noopener noreferrer">
              <el-icon><Link /></el-icon>
              开源项目地址
            </a>
          </p>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock, Link } from '@element-plus/icons-vue'
import { authAPI } from '@/api'
import { useUserStore } from '@/store'

const router = useRouter()
const userStore = useUserStore()
const loginFormRef = ref()
const loading = ref(false)

const loginForm = reactive({
  username: '',
  password: ''
})

const loginRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ]
}

const handleLogin = async () => {
  if (!loginFormRef.value) return
  
  try {
    await loginFormRef.value.validate()
    loading.value = true
    
    const response = await authAPI.login(loginForm)
    
    if (response.data) {
      // 使用store来处理登录状态
      userStore.login({
        username: response.data.username,
        role: response.data.role
      })
      
      ElMessage.success('登录成功')
      
      // 使用 nextTick 确保状态更新后再跳转
      await new Promise(resolve => setTimeout(resolve, 100))
      
      // 跳转到首页
      router.push('/dashboard')
    }
  } catch (error) {
    console.error('登录失败:', error)
    ElMessage.error(error.response?.data?.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
  position: relative;
  overflow: hidden;
}

/* 背景动画 */
.bg-animation {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 1;
}

.floating-shape {
  position: absolute;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  animation: float 6s ease-in-out infinite;
}

.shape-1 {
  width: 80px;
  height: 80px;
  top: 20%;
  left: 10%;
  animation-delay: 0s;
}

.shape-2 {
  width: 60px;
  height: 60px;
  top: 60%;
  right: 15%;
  animation-delay: 2s;
}

.shape-3 {
  width: 100px;
  height: 100px;
  bottom: 20%;
  left: 20%;
  animation-delay: 4s;
}

.shape-4 {
  width: 40px;
  height: 40px;
  top: 30%;
  right: 30%;
  animation-delay: 1s;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0px) rotate(0deg);
    opacity: 0.7;
  }
  50% {
    transform: translateY(-20px) rotate(180deg);
    opacity: 1;
  }
}

.login-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.15);
  padding: 40px;
  width: 100%;
  max-width: 420px;
  position: relative;
  z-index: 2;
  animation: cardSlideIn 0.8s ease-out;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

@keyframes cardSlideIn {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.login-header {
  text-align: center;
  margin-bottom: 35px;
}

.logo-container {
  position: relative;
  display: inline-block;
  margin-bottom: 20px;
}

.login-logo {
  width: 68px;
  height: 68px;
  animation: logoFloat 3s ease-in-out infinite;
  position: relative;
  z-index: 2;
}

.logo-glow {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 80px;
  height: 80px;
  background: radial-gradient(circle, rgba(102, 126, 234, 0.3) 0%, transparent 70%);
  border-radius: 50%;
  animation: glow 2s ease-in-out infinite alternate;
}

@keyframes logoFloat {
  0%, 100% {
    transform: translateY(0px);
  }
  50% {
    transform: translateY(-8px);
  }
}

@keyframes glow {
  from {
    opacity: 0.5;
    transform: translate(-50%, -50%) scale(1);
  }
  to {
    opacity: 0.8;
    transform: translate(-50%, -50%) scale(1.1);
  }
}

.app-title {
  color: #333;
  margin: 0 0 8px 0;
  font-size: 28px;
  font-weight: 700;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  animation: titleSlideIn 0.8s ease-out 0.2s both;
}

.app-subtitle {
  color: #555;
  margin: 0 0 8px 0;
  font-size: 18px;
  font-weight: 500;
  animation: titleSlideIn 0.8s ease-out 0.4s both;
}

.login-desc {
  color: #666;
  margin: 0;
  font-size: 14px;
  animation: titleSlideIn 0.8s ease-out 0.6s both;
}

@keyframes titleSlideIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.login-form {
  margin-bottom: 20px;
}

.form-item-animated {
  margin-bottom: 24px;
  animation: formSlideIn 0.6s ease-out both;
}

.form-item-animated:nth-child(1) {
  animation-delay: 0.8s;
}

.form-item-animated:nth-child(2) {
  animation-delay: 1s;
}

.form-item-animated:nth-child(3) {
  animation-delay: 1.2s;
}

@keyframes formSlideIn {
  from {
    opacity: 0;
    transform: translateX(-30px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.animated-input :deep(.el-input__wrapper) {
  transition: all 0.3s ease;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.animated-input :deep(.el-input__wrapper):hover {
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.15);
  transform: translateY(-1px);
}

.animated-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.25);
  transform: translateY(-1px);
}

.login-button {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: 500;
  border-radius: 8px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.login-button::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent);
  transition: left 0.5s;
}

.login-button:hover::before {
  left: 100%;
}

.login-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(102, 126, 234, 0.4);
}

.login-button:active {
  transform: translateY(0);
}

.loading-text {
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.7;
  }
}

.login-footer {
  margin-top: 35px;
  padding-top: 25px;
  border-top: 1px solid rgba(240, 240, 240, 0.8);
  animation: footerSlideIn 0.8s ease-out 1.4s both;
}

@keyframes footerSlideIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.copyright-info {
  text-align: center;
}

.copyright-info p {
  margin: 10px 0;
  font-size: 12px;
  color: #999;
  transition: color 0.3s ease;
}

.version {
  font-weight: 600;
  color: #666 !important;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-size: 13px;
}

.maintainer {
  color: #888;
}

.github-link {
  margin-top: 12px;
}

.github-link a {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #409eff;
  text-decoration: none;
  transition: all 0.3s ease;
  padding: 4px 8px;
  border-radius: 6px;
  background: rgba(64, 158, 255, 0.1);
}

.github-link a:hover {
  color: #337ecc;
  background: rgba(64, 158, 255, 0.15);
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.2);
}

.github-link .el-icon {
  font-size: 14px;
  transition: transform 0.3s ease;
}

.github-link a:hover .el-icon {
  transform: rotate(10deg);
}

</style> 