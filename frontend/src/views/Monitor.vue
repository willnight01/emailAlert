<template>
  <div class="monitor">
    <!-- 监控控制面板 -->
    <el-card class="control-panel">
      <template #header>
        <div class="panel-header">
          <div class="panel-title-section">
            <span>邮件监控控制面板</span>
            <el-text size="small" type="info" class="auto-refresh-tip">
              <el-icon><InfoFilled /></el-icon>
              页面会自动刷新最新配置
            </el-text>
          </div>
          <div class="control-actions">
            <el-button 
              type="success" 
              @click="startMonitor" 
              :loading="starting"
              :disabled="monitorStatus === 'running'"
            >
              <el-icon><VideoPlay /></el-icon>
              启动监控
            </el-button>
            <el-button 
              type="danger" 
              @click="stopMonitor" 
              :loading="stopping"
              :disabled="monitorStatus !== 'running'"
            >
              <el-icon><VideoPause /></el-icon>
              停止监控
            </el-button>
            <el-button 
              type="primary" 
              @click="refreshConfig" 
              :loading="refreshing"
              title="重新加载邮箱配置并更新监控服务"
            >
              <el-icon><Refresh /></el-icon>
              刷新配置
            </el-button>
          </div>
        </div>
      </template>
      
      <!-- 监控状态展示 -->
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="status-item">
            <div class="status-icon">
              <el-icon 
                :class="['status-indicator', getStatusClass()]"
                :color="getStatusColor()"
              >
                <component :is="getStatusIcon()" />
              </el-icon>
            </div>
            <div class="status-content">
              <div class="status-title">监控状态</div>
              <div class="status-value">{{ getStatusText() }}</div>
            </div>
          </div>
        </el-col>
        
        <el-col :span="6">
          <div class="status-item">
            <div class="status-icon">
              <el-icon color="#409eff">
                <Message />
              </el-icon>
            </div>
            <div class="status-content">
              <div class="status-title">监控邮箱</div>
              <div class="status-value">{{ monitorStats.activeMailboxes || 0 }}</div>
            </div>
          </div>
        </el-col>
        
        <el-col :span="6">
          <div class="status-item">
            <div class="status-icon">
              <el-icon color="#67c23a">
                <Download />
              </el-icon>
            </div>
            <div class="status-content">
              <div class="status-title">今日邮件</div>
              <div class="status-value">{{ monitorStats.todayEmails || 0 }}</div>
            </div>
          </div>
        </el-col>
        
        <el-col :span="6">
          <div class="status-item">
            <div class="status-icon">
              <el-icon color="#f56c6c">
                <Warning />
              </el-icon>
            </div>
            <div class="status-content">
              <div class="status-title">错误次数</div>
              <div class="status-value">{{ monitorStats.errorCount || 0 }}</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <!-- 邮箱监控详情 -->
    <el-row :gutter="20" class="mt-20">
      <el-col :span="16">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>邮箱监控详情</span>
              <el-button type="text" @click="refreshMailboxStatus">
                <el-icon><Refresh /></el-icon>
                刷新
              </el-button>
            </div>
          </template>
          
          <el-table 
            :data="mailboxStatuses" 
            style="width: 100%" 
            v-loading="loadingStatuses"
          >
            <el-table-column prop="name" label="邮箱名称" width="150" />
            <el-table-column prop="email" label="邮箱地址" width="200" />
            <el-table-column label="监控状态" width="120">
              <template #default="scope">
                <el-tag 
                  :type="scope.row.status === 'monitoring' ? 'success' : 'danger'"
                  size="small"
                >
                  {{ scope.row.status === 'monitoring' ? '监控中' : '已停止' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="lastCheck" label="最后检查" width="160">
              <template #default="scope">
                {{ formatTime(scope.row.lastCheck) }}
              </template>
            </el-table-column>
            <el-table-column prop="emailCount" label="邮件数量" width="100" />
            <el-table-column label="连接状态" width="120">
              <template #default="scope">
                <el-tag 
                  :type="scope.row.connected ? 'success' : 'danger'"
                  size="small"
                >
                  {{ scope.row.connected ? '已连接' : '连接失败' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="errorMessage" label="错误信息" min-width="200">
              <template #default="scope">
                <span v-if="scope.row.errorMessage" class="error-text">
                  {{ scope.row.errorMessage }}
                </span>
                <span v-else class="success-text">正常</span>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      
      <el-col :span="8">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>实时日志</span>
              <el-button type="text" @click="clearLogs">
                <el-icon><Delete /></el-icon>
                清空
              </el-button>
            </div>
          </template>
          
          <div class="log-container" ref="logContainer">
            <div 
              v-for="(log, index) in logs" 
              :key="index" 
              :class="['log-item', log.level]"
            >
              <span class="log-time">{{ formatTime(log.timestamp) }}</span>
              <span class="log-message">{{ log.message }}</span>
            </div>
            <div v-if="logs.length === 0" class="no-logs">
              暂无日志信息
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 邮件统计图表 -->
    <el-row class="mt-20">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>邮件处理统计</span>
              <el-radio-group v-model="statsTimeRange" size="small">
                <el-radio-button label="1h">最近1小时</el-radio-button>
                <el-radio-button label="24h">最近24小时</el-radio-button>
                <el-radio-button label="7d">最近7天</el-radio-button>
              </el-radio-group>
            </div>
          </template>
          
          <div class="chart-container">
            <div class="chart-placeholder">
              <el-icon><TrendCharts /></el-icon>
              <span>邮件处理统计图表（待集成图表库）</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, onActivated, nextTick } from 'vue'
import { 
  VideoPlay, VideoPause, Refresh, Message, Download, Warning,
  Delete, TrendCharts, CircleCheck, CircleClose, Loading, InfoFilled
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { monitorAPI, mailboxAPI } from '@/api'

// 响应式数据
const monitorStatus = ref('stopped') // 'running', 'stopped', 'error'
const starting = ref(false)
const stopping = ref(false)
const refreshing = ref(false)
const loadingStatuses = ref(false)

// 监控统计数据
const monitorStats = ref({
  activeMailboxes: 0,
  todayEmails: 0,
  errorCount: 0
})

// 邮箱状态列表
const mailboxStatuses = ref([])

// 日志相关
const logs = ref([])
const logContainer = ref()

// 统计时间范围
const statsTimeRange = ref('24h')

// 定时器和日志流
let statusTimer = null
let logEventSource = null

// 生命周期
onMounted(async () => {
  // 首先加载监控状态
  await loadMonitorStatus()
  
  // 自动刷新配置以确保显示最新的邮箱配置
  addLog('info', '页面加载时自动刷新配置...')
  try {
    if (monitorStatus.value === 'running') {
      // 如果监控正在运行，通过监控服务刷新配置
      await monitorAPI.refresh()
      addLog('success', '监控配置已自动刷新')
    } else {
      // 如果监控未运行，只需要重新加载邮箱状态
      addLog('info', '监控未运行，直接刷新邮箱状态')
    }
  } catch (error) {
    console.warn('自动刷新配置失败:', error)
    addLog('warning', '自动刷新配置失败，请手动刷新')
  }
  
  // 加载邮箱状态
  await loadMailboxStatuses()
  
  // 开始定时轮询
  startPolling()
})

onUnmounted(() => {
  stopPolling()
})

// 页面激活时（从其他页面切换回来）自动刷新状态
onActivated(async () => {
  addLog('info', '页面激活，检查配置更新...')
  
  // 重新加载监控状态
  await loadMonitorStatus()
  
  // 如果监控正在运行，刷新配置以获取最新邮箱列表
  if (monitorStatus.value === 'running') {
    try {
      await monitorAPI.refresh()
      addLog('success', '配置已更新到最新状态')
    } catch (error) {
      console.warn('激活时刷新配置失败:', error)
    }
  }
  
  // 重新加载邮箱状态
  await loadMailboxStatuses()
})

// 方法
const loadMonitorStatus = async () => {
  try {
    const response = await monitorAPI.status()
    // 修复状态获取：后端返回的是is_running字段，不是status字段
    const isRunning = response.data.is_running
    monitorStatus.value = isRunning ? 'running' : 'stopped'
    
    // 加载统计数据
    if (monitorStatus.value === 'running') {
      loadMonitorStats()
    }
  } catch (error) {
    console.error('加载监控状态失败:', error)
    monitorStatus.value = 'error'
  }
}

const loadMonitorStats = async () => {
  try {
    const response = await monitorAPI.stats()
    monitorStats.value = response.data || {}
  } catch (error) {
    console.error('加载监控统计失败:', error)
  }
}

const loadMailboxStatuses = async () => {
  loadingStatuses.value = true
  try {
    const response = await mailboxAPI.list({ status: 'active' })
    // 修复数据字段名：后端返回的是list字段，不是items字段
    const mailboxes = response.data.list || []
    
    // 增强邮箱状态信息
    mailboxStatuses.value = mailboxes.map(mailbox => ({
      ...mailbox,
      status: monitorStatus.value === 'running' ? 'monitoring' : 'stopped',
      lastCheck: monitorStatus.value === 'running' ? new Date() : null,
      emailCount: 0, // 初始为0，后续可以从统计API获取
      connected: mailbox.status === 'active', // 基于邮箱激活状态
      errorMessage: mailbox.status !== 'active' ? '邮箱未激活' : null
    }))
    
    console.log('加载邮箱监控状态:', mailboxStatuses.value)
  } catch (error) {
    console.error('加载邮箱状态失败:', error)
    ElMessage.error('加载邮箱状态失败')
  } finally {
    loadingStatuses.value = false
  }
}

const startMonitor = async () => {
  starting.value = true
  try {
    await monitorAPI.start()
    monitorStatus.value = 'running'
    ElMessage.success('邮件监控已启动')
    addLog('info', '邮件监控服务已启动')
    loadMonitorStats()
  } catch (error) {
    // 提取错误信息，优先使用响应数据中的消息
    let errorMessage = '未知错误'
    if (error.response?.data?.message) {
      errorMessage = error.response.data.message
    } else if (error.response?.data?.error) {
      errorMessage = error.response.data.error
    } else if (error.message) {
      errorMessage = error.message
    }
    
    console.error('启动监控错误详情:', error)
    ElMessage.error('启动监控失败: ' + errorMessage)
    addLog('error', '启动监控失败: ' + errorMessage)
  } finally {
    starting.value = false
  }
}

const stopMonitor = async () => {
  stopping.value = true
  try {
    await monitorAPI.stop()
    monitorStatus.value = 'stopped'
    ElMessage.success('邮件监控已停止')
    addLog('warning', '邮件监控服务已停止')
  } catch (error) {
    // 提取错误信息，优先使用响应数据中的消息
    let errorMessage = '未知错误'
    if (error.response?.data?.message) {
      errorMessage = error.response.data.message
    } else if (error.response?.data?.error) {
      errorMessage = error.response.data.error
    } else if (error.message) {
      errorMessage = error.message
    }
    
    console.error('停止监控错误详情:', error)
    ElMessage.error('停止监控失败: ' + errorMessage)
    addLog('error', '停止监控失败: ' + errorMessage)
  } finally {
    stopping.value = false
  }
}

const refreshConfig = async () => {
  refreshing.value = true
  try {
    // 如果监控正在运行，刷新监控服务的配置
    if (monitorStatus.value === 'running') {
      await monitorAPI.refresh()
      addLog('success', '监控服务配置已刷新')
    } else {
      addLog('info', '监控未运行，跳过服务配置刷新')
    }
    
    // 无论监控是否运行，都重新加载页面显示的邮箱状态
    await loadMailboxStatuses()
    
    ElMessage.success('配置已刷新')
    addLog('success', '页面配置刷新完成')
  } catch (error) {
    // 提取错误信息，优先使用响应数据中的消息
    let errorMessage = '未知错误'
    if (error.response?.data?.message) {
      errorMessage = error.response.data.message
    } else if (error.response?.data?.error) {
      errorMessage = error.response.data.error
    } else if (error.message) {
      errorMessage = error.message
    }
    
    console.error('刷新配置错误详情:', error)
    ElMessage.error('刷新配置失败: ' + errorMessage)
    addLog('error', '刷新配置失败: ' + errorMessage)
  } finally {
    refreshing.value = false
  }
}

const refreshMailboxStatus = () => {
  loadMailboxStatuses()
}

const addLog = (level, message, timestamp = null) => {
  logs.value.unshift({
    timestamp: timestamp ? new Date(timestamp) : new Date(),
    level,
    message
  })
  
  // 限制日志数量
  if (logs.value.length > 100) {
    logs.value = logs.value.slice(0, 100)
  }
  
  // 滚动到最新日志
  nextTick(() => {
    if (logContainer.value) {
      logContainer.value.scrollTop = 0
    }
  })
}

const clearLogs = () => {
  logs.value = []
}

const startPolling = () => {
  // 定期检查监控状态
  statusTimer = setInterval(() => {
    loadMonitorStatus()
    if (monitorStatus.value === 'running') {
      loadMonitorStats()
    }
  }, 30000) // 30秒检查一次
  
  // 创建实时日志连接
  startLogStream()
}

const startLogStream = () => {
  if (logEventSource) {
    logEventSource.close()
  }
  
  try {
    logEventSource = monitorAPI.createLogStream(
      // 接收日志回调
      (logData) => {
        addLog(logData.level, logData.message, logData.timestamp)
      },
      // 错误处理回调
      (error) => {
        console.error('日志流连接错误:', error)
        addLog('error', '日志连接中断，尝试重新连接...')
        
        // 3秒后重试连接
        setTimeout(() => {
          if (monitorStatus.value === 'running') {
            startLogStream()
          }
        }, 3000)
      }
    )
  } catch (error) {
    console.error('创建日志流失败:', error)
    addLog('error', '无法建立日志连接')
  }
}

const stopPolling = () => {
  if (statusTimer) {
    clearInterval(statusTimer)
    statusTimer = null
  }
  if (logEventSource) {
    logEventSource.close()
    logEventSource = null
  }
}

const getStatusClass = () => {
  switch (monitorStatus.value) {
    case 'running': return 'running'
    case 'stopped': return 'stopped'
    case 'error': return 'error'
    default: return 'unknown'
  }
}

const getStatusColor = () => {
  switch (monitorStatus.value) {
    case 'running': return '#67c23a'
    case 'stopped': return '#909399'
    case 'error': return '#f56c6c'
    default: return '#909399'
  }
}

const getStatusIcon = () => {
  switch (monitorStatus.value) {
    case 'running': return CircleCheck
    case 'stopped': return CircleClose
    case 'error': return Warning
    default: return Loading
  }
}

const getStatusText = () => {
  switch (monitorStatus.value) {
    case 'running': return '运行中'
    case 'stopped': return '已停止'
    case 'error': return '错误'
    default: return '未知'
  }
}

const formatTime = (timestamp) => {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}
</script>

<style scoped>
.monitor {
  padding: 0;
}

.control-panel {
  margin-bottom: 20px;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.panel-title-section {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.auto-refresh-tip {
  display: flex;
  align-items: center;
  gap: 4px;
  opacity: 0.8;
}

.control-actions {
  display: flex;
  gap: 10px;
}

.status-item {
  display: flex;
  align-items: center;
  padding: 20px;
  background: var(--el-fill-color-lighter);
  border-radius: 8px;
  height: 100px;
}

.status-icon {
  margin-right: 15px;
}

.status-icon .el-icon {
  font-size: 32px;
}

.status-indicator {
  animation: pulse 2s infinite;
}

.status-indicator.running {
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% { opacity: 1; }
  50% { opacity: 0.5; }
  100% { opacity: 1; }
}

.status-content {
  flex: 1;
}

.status-title {
  font-size: 14px;
  color: var(--el-text-color-regular);
  margin-bottom: 8px;
}

.status-value {
  font-size: 24px;
  font-weight: bold;
  color: var(--el-text-color-primary);
}

.mt-20 {
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.error-text {
  color: var(--el-color-danger);
  font-size: 12px;
}

.success-text {
  color: var(--el-color-success);
  font-size: 12px;
}

.log-container {
  height: 400px;
  overflow-y: auto;
  background: var(--el-fill-color-lighter);
  border-radius: 4px;
  padding: 10px;
}

.log-item {
  display: flex;
  margin-bottom: 8px;
  font-size: 12px;
  line-height: 1.4;
}

.log-time {
  color: var(--el-text-color-regular);
  margin-right: 10px;
  white-space: nowrap;
}

.log-message {
  flex: 1;
}

.log-item.info .log-message {
  color: var(--el-color-primary);
}

.log-item.success .log-message {
  color: var(--el-color-success);
}

.log-item.warning .log-message {
  color: var(--el-color-warning);
}

.log-item.error .log-message {
  color: var(--el-color-danger);
}

.no-logs {
  text-align: center;
  color: var(--el-text-color-placeholder);
  padding: 50px 0;
}

.chart-container {
  height: 300px;
}

.chart-placeholder {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background-color: var(--el-fill-color-light);
  border-radius: 4px;
  color: var(--el-text-color-placeholder);
  gap: 10px;
}

.chart-placeholder .el-icon {
  font-size: 48px;
}

:deep(.el-table .el-table__cell) {
  padding: 8px 0;
}
</style> 