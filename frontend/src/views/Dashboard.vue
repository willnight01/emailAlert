<template>
  <div class="dashboard-page">
    <!-- 页面标题 -->
    <div class="page-title">
      <h2>仪表盘</h2>
      <p class="page-subtitle">统一邮件告警平台运行概览</p>
      <div class="title-actions">
        <el-button 
          @click="refreshDashboard" 
          :loading="loading" 
          type="primary" 
          size="small"
          :icon="loading ? '' : 'Refresh'"
        >
          {{ loading ? '刷新中...' : '刷新数据' }}
        </el-button>
      </div>
    </div>

    <!-- 系统状态概览 -->
    <div class="status-overview" v-if="systemHealth">
      <div class="overall-status">
        <div class="status-icon" :class="systemHealth.status">
          <i :class="getStatusIcon(systemHealth.status)"></i>
        </div>
        <div class="status-info">
          <h3>{{ getOverallStatusText(systemHealth.status) }}</h3>
          <p>系统运行时间: {{ systemHealth.uptime || '未知' }}</p>
        </div>
      </div>
      <div class="system-metrics">
        <div class="metric-item">
          <span class="metric-label">版本</span>
          <span class="metric-value">{{ systemHealth.version || 'v1.0.0' }}</span>
        </div>
        <div class="metric-item">
          <span class="metric-label">服务数</span>
          <span class="metric-value">{{ systemHealth.summary?.total_services || 0 }}</span>
        </div>
        <div class="metric-item">
          <span class="metric-label">健康服务</span>
          <span class="metric-value text-success">{{ systemHealth.summary?.healthy_services || 0 }}</span>
        </div>
      </div>
    </div>

    <!-- 核心指标卡片 -->
    <el-row :gutter="24" class="stats-cards">
      <el-col :span="6">
        <div class="stat-card" @click="goToPage('/mailboxes')">
          <div class="stat-icon-wrapper">
            <el-icon class="stat-icon primary"><Message /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ businessStats.total_mailboxes || 0 }}</div>
            <div class="stat-label">邮箱总数</div>
            <div class="stat-desc">活跃: {{ businessStats.active_mailboxes || 0 }}</div>
          </div>
          <div class="stat-trend">
            <i class="el-icon-right"></i>
          </div>
        </div>
      </el-col>
      
      <el-col :span="6">
        <div class="stat-card" @click="goToPage('/rule-groups')">
          <div class="stat-icon-wrapper">
            <el-icon class="stat-icon warning"><Setting /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ businessStats.total_rules || 0 }}</div>
            <div class="stat-label">告警规则</div>
            <div class="stat-desc">活跃: {{ businessStats.active_rules || 0 }}</div>
          </div>
          <div class="stat-trend">
            <i class="el-icon-right"></i>
          </div>
        </div>
      </el-col>
      
      <el-col :span="6">
        <div class="stat-card" @click="goToPage('/channels')">
          <div class="stat-icon-wrapper">
            <el-icon class="stat-icon info"><Bell /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ businessStats.total_channels || 0 }}</div>
            <div class="stat-label">通知渠道</div>
            <div class="stat-desc">活跃: {{ businessStats.active_channels || 0 }}</div>
          </div>
          <div class="stat-trend">
            <i class="el-icon-right"></i>
          </div>
        </div>
      </el-col>
      
      <el-col :span="6">
        <div class="stat-card" @click="goToPage('/alerts')">
          <div class="stat-icon-wrapper">
            <el-icon class="stat-icon success"><SuccessFilled /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ businessStats.total_alerts || 0 }}</div>
            <div class="stat-label">告警记录</div>
            <div class="stat-desc">今日: {{ businessStats.today_alerts || 0 }}</div>
          </div>
          <div class="stat-trend">
            <i class="el-icon-right"></i>
          </div>
        </div>
      </el-col>
    </el-row>
    
    <!-- 主要内容区域 -->
    <el-row :gutter="24" class="main-content">
      <!-- 系统服务状态 -->
      <el-col :span="8">
        <div class="content-card">
          <div class="card-header">
            <h3>系统服务</h3>
            <el-button 
              @click="goToPage('/system')" 
              size="small" 
              type="primary" 
              link
            >
              查看详情
            </el-button>
          </div>
          <div class="service-list" v-loading="loading">
            <div class="service-item" v-for="service in systemServices" :key="service.name">
              <div class="service-info">
                <div class="service-name">
                  <i :class="getServiceIcon(service.name)" class="service-icon"></i>
                  {{ service.name }}
                </div>
                <div class="service-desc">{{ service.message }}</div>
              </div>
              <div class="service-status" :class="service.status">
                <span class="status-dot"></span>
                {{ getStatusText(service.status) }}
              </div>
            </div>
          </div>
        </div>
      </el-col>
      
      <!-- 系统资源监控 -->
      <el-col :span="8">
        <div class="content-card">
          <div class="card-header">
            <h3>系统资源</h3>
            <el-button 
              @click="goToPage('/system')" 
              size="small" 
              type="primary" 
              link
            >
              查看详情
            </el-button>
          </div>
          <div class="resource-metrics" v-loading="loading">
            <div class="metric-grid">
              <div class="metric-item" v-for="metric in systemMetrics" :key="metric.name">
                <div class="metric-header">
                  <span class="metric-name">{{ metric.name }}</span>
                  <span class="metric-value">{{ metric.value }}{{ metric.unit }}</span>
                </div>
                <div class="metric-progress">
                  <div class="progress-circle" :style="{ 
                    background: `conic-gradient(${getProgressColor(metric.percentage)} ${metric.percentage * 3.6}deg, #f0f0f0 0deg)` 
                  }">
                    <div class="progress-center">
                      <div class="progress-percentage">{{ metric.percentage }}%</div>
                    </div>
                  </div>
                </div>
                <div class="metric-status" :class="getMetricStatus(metric)">
                  {{ getMetricStatusText(metric) }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </el-col>
      
      <!-- 快捷操作 -->
      <el-col :span="8">
        <div class="content-card">
          <div class="card-header">
            <h3>快捷操作</h3>
          </div>
          <div class="quick-actions">
            <div class="action-item" @click="goToPage('/mailboxes', { action: 'create' })">
              <el-icon class="action-icon"><Plus /></el-icon>
              <div class="action-content">
                <div class="action-title">添加邮箱</div>
                <div class="action-desc">配置新的邮箱监控</div>
              </div>
            </div>
            <div class="action-item" @click="goToPage('/rule-groups', { action: 'create' })">
              <el-icon class="action-icon"><DocumentAdd /></el-icon>
              <div class="action-content">
                <div class="action-title">创建规则</div>
                <div class="action-desc">添加告警匹配规则</div>
              </div>
            </div>
            <div class="action-item" @click="goToPage('/channels', { action: 'create' })">
              <el-icon class="action-icon"><ChatLineSquare /></el-icon>
              <div class="action-content">
                <div class="action-title">配置渠道</div>
                <div class="action-desc">设置通知推送渠道</div>
              </div>
            </div>
            <div class="action-item" @click="goToPage('/templates', { action: 'create' })">
              <el-icon class="action-icon"><Edit /></el-icon>
              <div class="action-content">
                <div class="action-title">编辑模版</div>
                <div class="action-desc">自定义消息模版</div>
              </div>
            </div>
          </div>
        </div>
      </el-col>
    </el-row>
    
    <!-- 最近告警和统计图表 -->
    <el-row :gutter="24" class="bottom-content">
      <el-col :span="16">
        <div class="content-card">
          <div class="card-header">
            <h3>最近告警</h3>
            <div class="header-actions">
              <el-button 
                @click="refreshRecentAlerts" 
                :loading="alertsLoading"
                size="small" 
                type="primary"
                link
              >
                刷新
              </el-button>
              <el-button 
                @click="goToPage('/alerts')" 
                size="small" 
                type="primary"
                link
              >
                查看全部
              </el-button>
            </div>
          </div>
          <div class="alerts-table" v-loading="alertsLoading">
            <el-table :data="recentAlerts" style="width: 100%" empty-text="暂无告警记录">
              <el-table-column prop="created_at" label="时间" width="160">
                <template #default="scope">
                  {{ formatTime(scope.row.created_at) }}
                </template>
              </el-table-column>
              <el-table-column label="邮箱源" width="140">
                <template #default="scope">
                  {{ scope.row.mailbox?.name || '未知邮箱' }}
                </template>
              </el-table-column>
              <el-table-column prop="subject" label="主题" min-width="200" show-overflow-tooltip />
              <el-table-column label="规则" width="120" show-overflow-tooltip>
                <template #default="scope">
                  {{ scope.row.rule_group?.name || '无规则' }}
                </template>
              </el-table-column>
              <el-table-column prop="status" label="状态" width="100">
                <template #default="scope">
                  <el-tag :type="getAlertStatusType(scope.row.status)" size="small">
                    {{ getAlertStatusText(scope.row.status) }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="80">
                <template #default="scope">
                  <el-button 
                    @click="viewAlert(scope.row)" 
                    size="small" 
                    type="primary"
                    link
                  >
                    详情
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </el-col>
      
      <el-col :span="8">
        <div class="content-card">
          <div class="card-header">
            <h3>性能概览</h3>
          </div>
          <div class="performance-overview" v-loading="loading">
            <div class="perf-item">
              <div class="perf-label">CPU核心</div>
              <div class="perf-value">{{ systemResources.cpu?.cores || 0 }} 核心</div>
              <div class="perf-desc">处理能力充足</div>
            </div>
            <div class="perf-item">
              <div class="perf-label">内存使用</div>
              <div class="perf-value">{{ formatMemorySize(systemResources.memory?.alloc || 0) }}</div>
              <div class="perf-desc">使用率: {{ systemResources.memory?.usage_percent || 0 }}%</div>
            </div>
            <div class="perf-item">
              <div class="perf-label">并发数</div>
              <div class="perf-value">{{ systemResources.goroutines || 0 }} 个</div>
              <div class="perf-desc">系统负载轻</div>
            </div>
            <div class="perf-item">
              <div class="perf-label">成功率</div>
              <div class="perf-value">{{ businessStats.success_rate || 0 }}%</div>
              <div class="perf-desc">运行稳定</div>
            </div>
          </div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Message, 
  Setting, 
  Bell, 
  SuccessFilled, 
  Plus, 
  DocumentAdd, 
  ChatLineSquare, 
  Edit 
} from '@element-plus/icons-vue'
import { monitorAPI, systemAPI, alertsAPI } from '@/api'

const router = useRouter()

// 响应式数据
const loading = ref(false)
const alertsLoading = ref(false)

const systemHealth = ref(null)
const systemServices = ref([])
const systemResources = ref({})
const systemMetrics = ref([])
const businessStats = ref({})
const recentAlerts = ref([])

let refreshTimer = null

// 生命周期
onMounted(() => {
  loadDashboardData()
  // 设置定时刷新（每30秒）
  refreshTimer = setInterval(() => {
    loadDashboardData(false) // 静默刷新
  }, 30000)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})

// 方法
const loadDashboardData = async (showLoading = true) => {
  if (showLoading) loading.value = true
  
  try {
    // 并行加载数据
    const [healthResponse, statsResponse] = await Promise.all([
      systemAPI.health().catch(() => null),
      systemAPI.stats().catch(() => null)
    ])
    
    // 系统健康状态
    if (healthResponse?.data) {
      systemHealth.value = healthResponse.data
      systemServices.value = healthResponse.data.services || []
      systemResources.value = healthResponse.data.resources || {}
      
      // 构建系统资源指标
      const resources = healthResponse.data.resources
      if (resources) {
        systemMetrics.value = [
          {
            name: 'CPU核心数',
            value: resources.cpu?.cores || 0,
            unit: '个',
            percentage: Math.min((resources.cpu?.cores || 0) * 10, 100),
            type: 'cpu'
          },
          {
            name: '内存使用',
            value: formatMemorySize(resources.memory?.alloc || 0),
            unit: '',
            percentage: Math.round(resources.memory?.usage_percent || 0),
            type: 'memory',
            rawValue: resources.memory?.alloc || 0
          },
          {
            name: 'Goroutine数',
            value: resources.goroutines || 0,
            unit: '个',
            percentage: Math.min((resources.goroutines || 0) * 2, 100),
            type: 'goroutine',
            rawValue: resources.goroutines || 0
          },
          {
            name: 'GC次数',
            value: resources.gc?.num_gc || 0,
            unit: '次',
            percentage: Math.min((resources.gc?.num_gc || 0) * 5, 100),
            type: 'gc'
          }
        ]
      }
    }
    
    // 业务统计数据
    if (statsResponse?.data) {
      businessStats.value = statsResponse.data.business || {}
    }
    
    // 加载最近告警
    await loadRecentAlerts(false)
    
  } catch (error) {
    console.error('加载仪表盘数据失败:', error)
    if (showLoading) {
      ElMessage.error('加载数据失败')
    }
  } finally {
    if (showLoading) loading.value = false
  }
}



const loadRecentAlerts = async (showLoading = true) => {
  if (showLoading) alertsLoading.value = true
  
  try {
    const response = await alertsAPI.list({ 
      page: 1, 
      size: 10,
      sort_by: 'created_at',
      sort_order: 'desc'
    })
    if (response?.data?.alerts) {
      recentAlerts.value = response.data.alerts
    } else {
      recentAlerts.value = []
    }
  } catch (error) {
    console.error('加载最近告警失败:', error)
    recentAlerts.value = []
  } finally {
    if (showLoading) alertsLoading.value = false
  }
}

const refreshDashboard = () => {
  loadDashboardData(true)
}

const refreshRecentAlerts = () => {
  loadRecentAlerts(true)
}



const goToPage = (path, query = {}) => {
  router.push({ path, query })
}

const viewAlert = (alert) => {
  router.push(`/alerts?id=${alert.id}`)
}

// 工具方法
const formatTime = (timeStr) => {
  if (!timeStr) return '-'
  return new Date(timeStr).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const formatMemorySize = (bytes) => {
  return systemAPI.formatMemorySize(bytes)
}

const getStatusIcon = (status) => {
  const icons = {
    'healthy': 'el-icon-success-filled',
    'degraded': 'el-icon-warning-filled',
    'unhealthy': 'el-icon-error-filled'
  }
  return icons[status] || 'el-icon-info-filled'
}

const getOverallStatusText = (status) => {
  const texts = {
    'healthy': '系统运行正常',
    'degraded': '系统运行降级',
    'unhealthy': '系统运行异常'
  }
  return texts[status] || '系统状态未知'
}

const getServiceIcon = (serviceName) => {
  const icons = {
    '数据库连接': 'el-icon-coin',
    '邮件监控服务': 'el-icon-message',
    '通知服务': 'el-icon-bell',
    '缓存服务': 'el-icon-files'
  }
  return icons[serviceName] || 'el-icon-service'
}

const getStatusText = (status) => {
  return systemAPI.getStatusText(status)
}

// 系统资源监控相关方法
const getProgressColor = (percentage) => {
  if (percentage < 50) return '#52c41a'
  if (percentage < 80) return '#faad14'
  return '#ff4d4f'
}

const getMetricStatus = (metric) => {
  switch (metric.type) {
    case 'memory':
      if (metric.percentage < 50) return 'excellent'
      if (metric.percentage < 80) return 'good'
      return 'warning'
    case 'goroutine':
      if (metric.rawValue < 50) return 'excellent'
      if (metric.rawValue < 100) return 'good'
      return 'warning'
    case 'cpu':
      return 'info'
    case 'gc':
      return 'info'
    default:
      return 'info'
  }
}

const getMetricStatusText = (metric) => {
  const status = getMetricStatus(metric)
  const statusTexts = {
    'excellent': '优秀',
    'good': '良好',
    'warning': '需关注',
    'info': '正常'
  }
  return statusTexts[status] || '正常'
}

const getAlertStatusType = (status) => {
  const types = {
    'pending': 'warning',
    'sent': 'success',
    'failed': 'danger',
    'cancelled': 'info'
  }
  return types[status] || 'info'
}

const getAlertStatusText = (status) => {
  const texts = {
    'pending': '待处理',
    'sent': '已发送',
    'failed': '失败',
    'cancelled': '已取消'
  }
  return texts[status] || '未知'
}
</script>

<style scoped>
.dashboard-page {
  padding: 20px;
  background: #f5f7fa;
  min-height: calc(100vh - 60px);
}

/* 页面标题 */
.page-title {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 24px;
}

.page-title h2 {
  color: #1f2937;
  font-size: 28px;
  font-weight: 600;
  margin: 0 0 8px 0;
}

.page-subtitle {
  color: #6b7280;
  font-size: 14px;
  margin: 0;
}

.title-actions {
  display: flex;
  gap: 12px;
}

/* 状态概览 */
.status-overview {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 24px;
  color: white;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 4px 20px rgba(102, 126, 234, 0.3);
}

.overall-status {
  display: flex;
  align-items: center;
  gap: 16px;
}

.status-icon {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  background: rgba(255, 255, 255, 0.2);
}

.status-icon.healthy {
  background: rgba(82, 196, 26, 0.2);
}

.status-icon.degraded {
  background: rgba(250, 173, 20, 0.2);
}

.status-icon.unhealthy {
  background: rgba(255, 77, 79, 0.2);
}

.status-info h3 {
  margin: 0 0 4px 0;
  font-size: 20px;
  font-weight: 600;
}

.status-info p {
  margin: 0;
  opacity: 0.9;
  font-size: 14px;
}

.system-metrics {
  display: flex;
  gap: 32px;
}

.metric-item {
  text-align: center;
}

.metric-label {
  display: block;
  font-size: 12px;
  opacity: 0.8;
  margin-bottom: 4px;
}

.metric-value {
  display: block;
  font-size: 18px;
  font-weight: 600;
}

.text-success {
  color: #52c41a;
}

/* 统计卡片 */
.stats-cards {
  margin-bottom: 24px;
}

.stat-card {
  background: white;
  border-radius: 12px;
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  border: 1px solid #e5e7eb;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
  overflow: hidden;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
}

.stat-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, #667eea, #764ba2);
  opacity: 0;
  transition: opacity 0.2s ease;
}

.stat-card:hover::before {
  opacity: 1;
}

.stat-icon-wrapper {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-icon {
  font-size: 24px;
}

.stat-icon.primary {
  color: #6366f1;
  background: rgba(99, 102, 241, 0.1);
}

.stat-icon.warning {
  color: #f59e0b;
  background: rgba(245, 158, 11, 0.1);
}

.stat-icon.info {
  color: #06b6d4;
  background: rgba(6, 182, 212, 0.1);
}

.stat-icon.success {
  color: #10b981;
  background: rgba(16, 185, 129, 0.1);
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: #1f2937;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 14px;
  color: #374151;
  font-weight: 500;
  margin-bottom: 2px;
}

.stat-desc {
  font-size: 12px;
  color: #6b7280;
}

.stat-trend {
  color: #9ca3af;
  font-size: 16px;
}

/* 主要内容区域 */
.main-content {
  margin-bottom: 24px;
}

.content-card {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  border: 1px solid #e5e7eb;
  height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f3f4f6;
}

.card-header h3 {
  margin: 0;
  color: #1f2937;
  font-size: 18px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 8px;
}

/* 服务列表 */
.service-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.service-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background: #f9fafb;
  border-radius: 8px;
  border: 1px solid #f3f4f6;
  transition: all 0.2s ease;
}

.service-item:hover {
  background: #f3f4f6;
  transform: translateY(-1px);
}

.service-info {
  flex: 1;
}

.service-name {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  color: #1f2937;
  margin-bottom: 4px;
}

.service-icon {
  color: #6b7280;
  font-size: 16px;
}

.service-desc {
  font-size: 12px;
  color: #6b7280;
}

.service-status {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 500;
  padding: 4px 8px;
  border-radius: 12px;
}

.service-status.healthy {
  color: #059669;
  background: rgba(5, 150, 105, 0.1);
}

.service-status.degraded {
  color: #d97706;
  background: rgba(217, 119, 6, 0.1);
}

.service-status.unhealthy {
  color: #dc2626;
  background: rgba(220, 38, 38, 0.1);
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}

/* 系统资源监控 */
.resource-metrics {
  min-height: 300px;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.metric-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 16px;
  background: #f9fafb;
  border-radius: 8px;
  border: 1px solid #f3f4f6;
  transition: all 0.2s ease;
}

.metric-item:hover {
  background: #f3f4f6;
  transform: translateY(-1px);
}

.metric-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  margin-bottom: 12px;
  width: 100%;
}

.metric-name {
  font-size: 12px;
  color: #6b7280;
  font-weight: 500;
}

.metric-value {
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
}

.metric-progress {
  margin-bottom: 12px;
}

.progress-circle {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.progress-center {
  width: 46px;
  height: 46px;
  border-radius: 50%;
  background: white;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.progress-percentage {
  font-size: 11px;
  font-weight: 600;
  color: #1f2937;
  line-height: 1;
}

.metric-status {
  font-size: 11px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 10px;
}

.metric-status.excellent {
  color: #059669;
  background: rgba(5, 150, 105, 0.1);
}

.metric-status.good {
  color: #0891b2;
  background: rgba(8, 145, 178, 0.1);
}

.metric-status.warning {
  color: #d97706;
  background: rgba(217, 119, 6, 0.1);
}

.metric-status.info {
  color: #6366f1;
  background: rgba(99, 102, 241, 0.1);
}

/* 快捷操作 */
.quick-actions {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.action-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #f9fafb;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid #f3f4f6;
}

.action-item:hover {
  background: #f3f4f6;
  transform: translateY(-1px);
}

.action-icon {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-size: 16px;
}

.action-content {
  flex: 1;
}

.action-title {
  font-weight: 500;
  color: #374151;
  margin-bottom: 2px;
}

.action-desc {
  font-size: 12px;
  color: #6b7280;
}

/* 底部内容 */
.bottom-content {
  margin-bottom: 24px;
}

.alerts-table {
  min-height: 200px;
}

/* 性能概览 */
.performance-overview {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.perf-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px;
  background: #f9fafb;
  border-radius: 8px;
  border: 1px solid #f3f4f6;
}

.perf-label {
  font-size: 12px;
  color: #6b7280;
  font-weight: 500;
}

.perf-value {
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
}

.perf-desc {
  font-size: 11px;
  color: #9ca3af;
}

/* 动画效果 */
@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

/* 响应式设计 */
@media (max-width: 768px) {
  .dashboard-page {
    padding: 16px;
  }
  
  .page-title {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }
  
  .status-overview {
    flex-direction: column;
    gap: 16px;
    text-align: center;
  }
  
  .system-metrics {
    gap: 16px;
  }
  
  .stats-cards .el-col {
    margin-bottom: 16px;
  }
  
  .main-content .el-col,
  .bottom-content .el-col {
    margin-bottom: 16px;
  }
}
</style> 