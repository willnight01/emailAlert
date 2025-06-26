import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/store'

// 处理会话失效
const handleSessionExpired = () => {
  const userStore = useUserStore()
  userStore.logout()
  ElMessage.error('会话已过期，请重新登录')
  // 跳转到登录页（这里可以根据实际路由配置调整）
  if (window.location.pathname !== '/login') {
    window.location.href = '/login'
  }
}

// 创建axios实例
const api = axios.create({
  baseURL: '/api/v1',
  timeout: 60000, // 增加到60秒，用于邮件测试等长时间操作
  withCredentials: true, // 启用Cookie发送
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器 (Session认证下不需要添加Authorization头)
api.interceptors.request.use(
  (config) => {
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    const { data } = response
    
    // 如果是文件下载类型，直接返回response
    if (response.config.responseType === 'blob') {
      return response
    }
    
    // 检查业务状态码 - 对于没有code字段的响应，直接返回
    if (data.code !== undefined && data.code !== 200 && data.code !== 201) {
      ElMessage.error(data.message || '请求失败')
      return Promise.reject(new Error(data.message || '请求失败'))
    }
    
    return data
  },
  (error) => {
    let message = '网络错误'
    
    if (error.response) {
      const { status, data } = error.response
      
      switch (status) {
        case 400:
          message = data?.message || '请求参数错误'
          // 对400错误也显示错误消息，但不自动弹出，由组件控制
          console.error('请求错误:', message)
          break
        case 401:
          message = '会话已过期，请重新登录'
          console.log('🚨 收到401错误，会话已过期')
          // 会话失效，直接登出
          handleSessionExpired()
          break
        case 403:
          message = '权限不足'
          ElMessage.error(message)
          break
        case 404:
          message = '请求的资源不存在'
          ElMessage.error(message)
          break
        case 500:
          message = '服务器内部错误'
          ElMessage.error(message)
          break
        default:
          message = data?.message || `请求失败 (${status})`
          ElMessage.error(message)
      }
    } else if (error.code === 'ECONNABORTED') {
      message = '请求超时'
      ElMessage.error(message)
    } else {
      ElMessage.error(message)
    }
    
    return Promise.reject(error)
  }
)

// API接口定义
export const mailboxAPI = {
  // 获取邮箱列表
  list: (params) => api.get('/mailboxes', { params }),
  // 创建邮箱
  create: (data) => api.post('/mailboxes', data),
  // 获取邮箱详情
  get: (id) => api.get(`/mailboxes/${id}`),
  // 获取邮箱详情（包含密码，用于编辑）
  getForEdit: (id) => api.get(`/mailboxes/${id}/edit`),
  // 更新邮箱
  update: (id, data) => api.put(`/mailboxes/${id}`, data),
  // 删除邮箱
  delete: (id) => api.delete(`/mailboxes/${id}`),
  // 测试邮箱连接 - 支持两种方式：测试现有邮箱和测试配置数据
  test: (idOrData) => {
    if (typeof idOrData === 'object') {
      // 测试配置数据，使用更长的超时时间（120秒）
      return api.post('/mailboxes/config-test', idOrData, { timeout: 120000 })
    } else {
      // 测试现有邮箱
      return api.post(`/mailboxes/${idOrData}/test`, {}, { timeout: 120000 })
    }
  }
}

// 旧的告警规则API已移除，统一使用新的规则组API

// 新增：规则组管理API
export const ruleGroupsAPI = {
  // 获取规则组列表
  list: (params) => api.get('/rule-groups', { params }),
  // 创建规则组
  create: (data) => api.post('/rule-groups', data),
  // 创建规则组及其条件
  createWithConditions: (data) => api.post('/rule-groups/with-conditions', data),
  // 获取规则组详情
  get: (id) => api.get(`/rule-groups/${id}`),
  // 获取规则组及其条件
  getWithConditions: (id) => api.get(`/rule-groups/${id}/with-conditions`),
  // 更新规则组
  update: (id, data) => api.put(`/rule-groups/${id}`, data),
  // 更新规则组状态
  updateStatus: (id, data) => api.put(`/rule-groups/${id}/status`, data),
  // 更新规则组及其条件
  updateWithConditions: (id, data) => api.put(`/rule-groups/${id}/with-conditions`, data),
  // 删除规则组
  delete: (id) => api.delete(`/rule-groups/${id}`),
  // 测试规则组
  test: (data) => api.post('/rule-groups/test', data),
  // 获取邮箱选项
  getMailboxOptions: () => api.get('/rule-groups/mailbox-options'),
  // 获取匹配类型选项
  getMatchTypeOptions: () => api.get('/rule-groups/match-type-options'),
  // 获取字段类型选项
  getFieldTypeOptions: () => api.get('/rule-groups/field-type-options'),
  // 获取通知渠道选项
  getChannelOptions: () => api.get('/rule-groups/channel-options')
}

export const channelsAPI = {
  // 获取渠道列表
  list: (params) => api.get('/channels', { params }),
  // 创建渠道
  create: (data) => api.post('/channels', data),
  // 获取渠道详情
  get: (id) => api.get(`/channels/${id}`),
  // 更新渠道
  update: (id, data) => api.put(`/channels/${id}`, data),
  // 删除渠道
  delete: (id) => api.delete(`/channels/${id}`),
  // 测试渠道连接 - 支持两种方式：测试现有渠道和测试配置数据
  test: (idOrData) => {
    if (typeof idOrData === 'object') {
      // 测试配置数据，使用更长的超时时间（120秒）
      return api.post('/channels/config-test', idOrData, { timeout: 120000 })
    } else {
      // 测试现有渠道
      return api.post(`/channels/${idOrData}/test`, {}, { timeout: 120000 })
    }
  },
  // 更新渠道状态
  updateStatus: (id, data) => api.put(`/channels/${id}/status`, data),
  // 发送通知消息
  send: (id, data) => api.post(`/channels/${id}/send`, data),
  // 获取支持的渠道类型
  getTypes: () => api.get('/channels/types')
}

export const templatesAPI = {
  // 获取模版列表
  list: (params) => api.get('/templates', { params }),
  // 创建模版
  create: (data) => api.post('/templates', data),
  // 获取模版详情
  get: (id) => api.get(`/templates/${id}`),
  // 更新模版
  update: (id, data) => api.put(`/templates/${id}`, data),
  // 删除模版
  delete: (id) => api.delete(`/templates/${id}`),
  // 设置默认模版
  setDefault: (id) => api.put(`/templates/${id}/default`),
  // 渲染模版
  render: (id, data) => api.post(`/templates/${id}/render`, data),
  // 预览模版
  preview: (data) => api.post('/templates/preview', data),
  // 获取可用变量列表
  getVariables: () => api.get('/templates/variables'),
  // 获取指定类型模版
  getByType: (type) => api.get(`/templates/type/${type}`),
  // 获取默认模版
  getDefault: (type) => api.get(`/templates/default/${type}`)
}

export const alertsAPI = {
  // 获取告警列表
  list: (params) => api.get('/alerts', { params }),
  // 获取告警详情
  get: (id) => api.get(`/alerts/${id}`),
  // 更新告警状态
  updateStatus: (id, data) => api.put(`/alerts/${id}/status`, data),
  // 删除告警记录
  delete: (id) => api.delete(`/alerts/${id}`),
  // 重试告警发送
  retry: (id) => api.post(`/alerts/${id}/retry`),
  // 获取告警统计信息
  getStats: (params) => api.get('/alerts/stats', { params }),
  // 获取告警趋势数据
  getTrends: (params) => api.get('/alerts/trends', { params }),
  // 批量更新告警状态
  batchUpdate: (data) => api.post('/alerts/batch-update', data)
}

export const monitorAPI = {
  // 启动邮件监控
  start: () => api.post('/monitor/start'),
  // 停止邮件监控
  stop: () => api.post('/monitor/stop'),
  // 获取监控状态
  status: () => api.get('/monitor/status'),
  // 刷新邮箱配置
  refresh: () => api.post('/monitor/refresh'),
  // 获取邮件统计信息
  stats: () => api.get('/monitor/stats'),
  // 创建实时日志连接 (Server-Sent Events)
  createLogStream: (onLog, onError) => {
    // 使用Cookie认证，EventSource会自动携带Cookie
    // 如果需要在URL中传递session_id作为备选方案，可以从Cookie中获取
    let url = '/api/v1/monitor/logs'
    
    // 备选方案：从Cookie中获取session_id并作为查询参数传递
    // 这在某些情况下可能有用，但通常Cookie会自动发送
    const sessionId = document.cookie
      .split(';')
      .find(row => row.trim().startsWith('session_id='))
      ?.split('=')[1]
    
    if (sessionId) {
      url += `?session_id=${encodeURIComponent(sessionId)}`
    }
    
    const eventSource = new EventSource(url)
    
    eventSource.onmessage = (event) => {
      if (event.type === 'log') {
        try {
          const logData = JSON.parse(event.data)
          onLog && onLog(logData)
        } catch (error) {
          console.error('解析日志数据失败:', error)
        }
      }
    }
    
    eventSource.addEventListener('log', (event) => {
      try {
        const logData = JSON.parse(event.data)
        onLog && onLog(logData)
      } catch (error) {
        console.error('解析日志数据失败:', error)
      }
    })
    
    eventSource.addEventListener('heartbeat', (event) => {
      // 心跳，保持连接活跃
      console.debug('收到心跳:', event.data)
    })
    
    eventSource.onerror = (error) => {
      console.error('日志流连接错误:', error)
      onError && onError(error)
    }
    
    return eventSource
  }
}

export const notificationLogsAPI = {
  // 获取通知日志列表
  list: (params) => api.get('/notification-logs', { params }),
  // 获取通知日志详情
  get: (id) => api.get(`/notification-logs/${id}`),
  // 删除通知日志
  delete: (id) => api.delete(`/notification-logs/${id}`),
  // 重试发送通知
  retry: (id) => api.post(`/notification-logs/${id}/retry`),
  // 获取通知日志统计
  getStats: (params) => api.get('/notification-logs/stats', { params })
}

export const systemAPI = {
  // 获取系统状态（兼容现有接口）
  status: () => api.get('/system/status'),
  // 获取系统统计信息
  stats: () => api.get('/system/stats'),
  // 获取系统健康状态
  health: () => api.get('/system/health'),
  // 清理历史数据
  cleanup: (dataType, timeRange) => api.post('/system/cleanup', { data_type: dataType, time_range: timeRange }),
  
  // 格式化内存大小
  formatMemorySize: (bytes) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  },
  
  // 格式化运行时间
  formatUptime: (uptimeStr) => {
    return uptimeStr || '未知'
  },
  
  // 获取状态颜色
  getStatusColor: (status) => {
    const colors = {
      'healthy': '#67c23a',
      'degraded': '#e6a23c', 
      'unhealthy': '#f56c6c',
      'running': '#67c23a',
      'stopped': '#909399',
      'error': '#f56c6c'
    }
    return colors[status] || '#909399'
  },
  
  // 获取状态文本
  getStatusText: (status) => {
    const texts = {
      'healthy': '健康',
      'degraded': '降级',
      'unhealthy': '不健康',
      'running': '运行中',
      'stopped': '已停止',
      'error': '错误'
    }
    return texts[status] || '未知'
  }
}

// 认证API
export const authAPI = {
  // 用户登录
  login: (data) => api.post('/auth/login', data),
  // 用户登出
  logout: () => api.post('/auth/logout'),
  // 获取用户信息
  profile: () => api.get('/auth/profile')
}

// 用户管理API
export const usersAPI = {
  // 获取用户列表
  list: () => api.get('/users'),
  // 创建用户
  create: (data) => api.post('/users', data),
  // 更新用户
  update: (username, data) => api.put(`/users/${username}`, data),
  // 删除用户
  delete: (username) => api.delete(`/users/${username}`)
}

export default api 