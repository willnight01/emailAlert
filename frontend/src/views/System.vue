<template>
  <div class="system-page">
    <!-- 页面标题 -->
    <div class="page-title">
      <h2>系统监控</h2>
      <p class="page-subtitle">实时监控系统运行状态和资源使用情况</p>
    </div>

    <!-- 系统状态概览 -->
    <div class="status-overview" v-if="systemHealth">
      <div class="overall-status">
        <div class="status-icon" :class="systemHealth.status">
          <i :class="getStatusIcon(systemHealth.status)"></i>
        </div>
        <div class="status-info">
          <h3>{{ getStatusText(systemHealth.status) }}</h3>
          <p>系统运行时间: {{ systemHealth.uptime }}</p>
        </div>
      </div>
      <div class="system-metrics">
        <div class="metric-item">
          <span class="metric-label">版本</span>
          <span class="metric-value">{{ systemHealth.version }}</span>
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

    <!-- 主要内容区域 -->
    <el-row :gutter="24" class="main-content">
      <!-- 服务状态 -->
      <el-col :span="12">
        <div class="status-card">
          <div class="card-header">
            <h3>服务状态</h3>
            <el-button 
              @click="refreshSystemStatus" 
              :loading="loading" 
              size="small" 
              type="primary" 
              :icon="loading ? '' : 'Refresh'"
              circle
            />
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
      
      <!-- 系统资源 -->
      <el-col :span="12">
        <div class="resource-card">
          <div class="card-header">
            <h3>系统资源</h3>
            <el-tooltip content="点击查看性能说明" placement="top">
              <el-button 
                @click="showPerformanceInfo = !showPerformanceInfo" 
                size="small" 
                type="info" 
                :icon="showPerformanceInfo ? 'Hide' : 'InfoFilled'"
                circle
              />
            </el-tooltip>
          </div>
          
          <!-- 性能说明面板 -->
          <div v-if="showPerformanceInfo" class="performance-info">
            <div class="info-header">
              <i class="el-icon-info-filled"></i>
              性能指标说明
            </div>
            <div class="info-content">
              <div class="info-item">
                <strong>CPU核心数:</strong> 系统处理能力上限，10核心可同时处理多个任务
              </div>
              <div class="info-item">
                <strong>内存使用:</strong> 应用实际占用内存，&lt;50%健康，50-80%需关注，&gt;80%异常
              </div>
              <div class="info-item">
                <strong>Goroutine:</strong> 并发任务数，&lt;100正常，100-1000中等负载，&gt;1000高负载
              </div>
              <div class="info-item">
                <strong>GC次数:</strong> 垃圾回收频率，反映内存分配模式和程序运行时间
              </div>
            </div>
          </div>

          <div class="resource-charts" v-loading="loading">
            <!-- CPU核心数特殊展示 -->
            <div class="cpu-chart-item" v-if="systemMetrics.find(m => m.type === 'cpu')">
                              <div class="cpu-cores-display">
                  <div class="cpu-icon">
                    <i class="el-icon-monitor"></i>
                  </div>
                <div class="cpu-info">
                  <div class="cpu-title">{{ systemMetrics.find(m => m.type === 'cpu').name }}</div>
                  <div class="cpu-value">
                    <span class="cores-number">{{ systemMetrics.find(m => m.type === 'cpu').value }}</span>
                    <span class="cores-unit">{{ systemMetrics.find(m => m.type === 'cpu').unit }}</span>
                  </div>
                  <div class="cpu-description">{{ getMetricAnalysis(systemMetrics.find(m => m.type === 'cpu')) }}</div>
                </div>
                <div class="cpu-status" :class="getMetricStatus(systemMetrics.find(m => m.type === 'cpu'))">
                  <i class="el-icon-success-filled"></i>
                  <span>{{ getMetricStatusText(systemMetrics.find(m => m.type === 'cpu')) }}</span>
                </div>
              </div>
            </div>
            
            <!-- 其他指标保持原有展示方式 -->
            <div class="chart-item" v-for="metric in systemMetrics.filter(m => m.type !== 'cpu')" :key="metric.name">
              <div class="chart-header">
                <div class="chart-title">
                  <span class="chart-name">{{ metric.name }}</span>
                  <span class="chart-value">{{ metric.value }}{{ metric.unit }}</span>
                </div>
                <div class="chart-status" :class="getMetricStatus(metric)">
                  {{ getMetricStatusText(metric) }}
                </div>
              </div>
              
              <!-- 圆形进度图表 -->
              <div class="circular-progress">
                <div class="progress-circle" :style="{ 
                  background: `conic-gradient(${getProgressColor(metric.percentage)} ${metric.percentage * 3.6}deg, #f0f0f0 0deg)` 
                }">
                  <div class="progress-center">
                    <div class="progress-percentage">{{ metric.percentage }}%</div>
                    <div class="progress-label">{{ getMetricLabel(metric.name) }}</div>
                  </div>
                </div>
              </div>
              
              <!-- 指标分析 -->
              <div class="metric-analysis">
                <div class="analysis-text">{{ getMetricAnalysis(metric) }}</div>
                <div class="analysis-reference">{{ getMetricReference(metric.name) }}</div>
              </div>
            </div>
          </div>
        </div>
      </el-col>
    </el-row>
    
    <!-- 业务统计 -->
    <div class="business-stats">
      <div class="card-header">
        <h3>业务统计</h3>
        <el-button 
          @click="refreshBusinessStats" 
          :loading="statsLoading" 
          size="small" 
          type="primary" 
          :icon="statsLoading ? '' : 'Refresh'"
          circle
        />
      </div>
      <div class="stats-grid" v-loading="statsLoading">
        <div class="stat-item" v-for="stat in businessStats" :key="stat.name">
          <div class="stat-icon">
            <i :class="getStatIcon(stat.name)"></i>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stat.value }}</div>
            <div class="stat-name">{{ stat.name }}</div>
            <div class="stat-desc" v-if="stat.desc">{{ stat.desc }}</div>
          </div>
          <div class="stat-trend">
            <div class="trend-indicator" :class="getBusinessTrend(stat)">
              <i :class="getTrendIcon(stat)"></i>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 数据清理功能 -->
    <div class="cleanup-section">
      <div class="card-header">
        <h3>历史数据清理</h3>
        <el-button 
          @click="showCleanupDialog = true" 
          type="danger" 
          size="small"
          :icon="'Delete'"
        >
          清理数据
        </el-button>
      </div>
      <div class="cleanup-info">
        <p class="cleanup-desc">定期清理历史数据可以释放存储空间，提高系统性能。请根据需要选择清理范围。</p>
        <div class="cleanup-stats">
          <div class="stat-item">
            <span class="stat-label">告警记录总数:</span>
            <span class="stat-value">{{ businessStats.find(s => s.name === '告警记录')?.total || 0 }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-label">通知记录总数:</span>
            <span class="stat-value">{{ businessStats.find(s => s.name === '通知记录')?.total || 0 }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 数据清理对话框 -->
    <el-dialog 
      v-model="showCleanupDialog" 
      title="历史数据清理" 
      width="500px"
      :close-on-click-modal="false"
    >
      <div class="cleanup-dialog">
        <div class="cleanup-warning">
          <el-alert
            title="警告：数据清理操作不可逆"
            type="warning"
            description="清理后的数据将无法恢复，请谨慎操作。建议在清理前备份重要数据。"
            show-icon
            :closable="false"
          />
        </div>

        <el-form :model="cleanupForm" label-width="100px" class="cleanup-form">
          <el-form-item label="数据类型">
            <el-radio-group v-model="cleanupForm.dataType">
              <el-radio value="alerts">仅告警历史</el-radio>
              <el-radio value="notifications">仅通知历史</el-radio>
              <el-radio value="both">告警和通知历史</el-radio>
            </el-radio-group>
          </el-form-item>

          <el-form-item label="时间范围">
            <el-radio-group v-model="cleanupForm.timeRange">
              <el-radio value="1month">1个月前</el-radio>
              <el-radio value="3months">3个月前</el-radio>
              <el-radio value="6months">6个月前</el-radio>
              <el-radio value="1year">1年前</el-radio>
              <el-radio value="2years">2年前</el-radio>
              <el-radio value="all">全部数据</el-radio>
            </el-radio-group>
          </el-form-item>

          <el-form-item label="清理预览">
            <div class="cleanup-preview">
              <p>将清理: <strong>{{ getCleanupDescription() }}</strong></p>
              <p class="cleanup-time">清理范围: {{ getTimeRangeDescription() }}</p>
            </div>
          </el-form-item>
        </el-form>

        <div class="cleanup-confirm">
          <el-checkbox v-model="cleanupForm.confirmed">
            我确认已了解数据清理的风险，并同意执行此操作
          </el-checkbox>
        </div>
      </div>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showCleanupDialog = false">取消</el-button>
          <el-button 
            type="danger" 
            @click="executeCleanup"
            :disabled="!cleanupForm.confirmed"
            :loading="cleanupLoading"
          >
            {{ cleanupLoading ? '清理中...' : '确认清理' }}
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 清理结果对话框 -->
    <el-dialog 
      v-model="showResultDialog" 
      title="清理结果" 
      width="400px"
      :close-on-click-modal="false"
    >
      <div class="cleanup-result">
        <div class="result-icon" :class="cleanupResult.success ? 'success' : 'error'">
          <i :class="cleanupResult.success ? 'el-icon-success-filled' : 'el-icon-error-filled'"></i>
        </div>
        <div class="result-content">
          <h4>{{ cleanupResult.success ? '清理完成' : '清理失败' }}</h4>
          <p v-if="cleanupResult.success">
            成功删除 <strong>{{ cleanupResult.deleted_rows }}</strong> 条记录
          </p>
          <p v-if="cleanupResult.success">
            耗时: {{ formatDuration(cleanupResult.duration) }}
          </p>
          <p v-if="!cleanupResult.success" class="error-message">
            {{ cleanupResult.error }}
          </p>
        </div>
      </div>

      <template #footer>
        <span class="dialog-footer">
          <el-button type="primary" @click="closeResultDialog">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { systemAPI } from '@/api'

const loading = ref(false)
const statsLoading = ref(false)
const showPerformanceInfo = ref(false)
const systemHealth = ref(null)
const systemServices = ref([])
const systemMetrics = ref([])
const businessStats = ref([])


// 数据清理相关状态
const showCleanupDialog = ref(false)
const showResultDialog = ref(false)
const cleanupLoading = ref(false)
const cleanupForm = ref({
  dataType: 'both',
  timeRange: '3months',
  confirmed: false
})
const cleanupResult = ref({
  success: false,
  deleted_rows: 0,
  duration: 0,
  error: ''
})

// 获取系统健康状态
const getSystemHealth = async () => {
  try {
    loading.value = true
    const response = await systemAPI.health()
    if (response.code === 200) {
      systemHealth.value = response.data
      systemServices.value = response.data.services || []
      
      // 构建系统资源指标
      const resources = response.data.resources
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
            value: systemAPI.formatMemorySize(resources.memory?.alloc || 0),
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
  } catch (error) {
    console.error('获取系统健康状态失败:', error)
    ElMessage.error('获取系统状态失败')
  } finally {
    loading.value = false
  }
}

// 获取业务统计
const getBusinessStats = async () => {
  try {
    statsLoading.value = true
    const response = await systemAPI.stats()
    if (response.code === 200) {
      const business = response.data.business
      if (business) {
        businessStats.value = [
          {
            name: '邮箱总数',
            value: business.total_mailboxes || 0,
            desc: `活跃: ${business.active_mailboxes || 0}`,
            active: business.active_mailboxes || 0,
            total: business.total_mailboxes || 0
          },
          {
            name: '告警规则',
            value: business.total_rules || 0,
            desc: `活跃: ${business.active_rules || 0}`,
            active: business.active_rules || 0,
            total: business.total_rules || 0
          },
          {
            name: '通知渠道',
            value: business.total_channels || 0,
            desc: `活跃: ${business.active_channels || 0}`,
            active: business.active_channels || 0,
            total: business.total_channels || 0
          },
          {
            name: '告警记录',
            value: business.total_alerts || 0,
            desc: `今日: ${business.today_alerts || 0}`,
            today: business.today_alerts || 0,
            total: business.total_alerts || 0
          },
          {
            name: '通知记录',
            value: business.total_notifications || 0,
            desc: `今日: ${business.today_notifications || 0}`,
            today: business.today_notifications || 0,
            total: business.total_notifications || 0
          }
        ]
      }
    }
  } catch (error) {
    console.error('获取业务统计失败:', error)
    ElMessage.error('获取业务统计失败')
  } finally {
    statsLoading.value = false
  }
}

// 刷新系统状态
const refreshSystemStatus = () => {
  getSystemHealth()
}

// 刷新业务统计
const refreshBusinessStats = () => {
  getBusinessStats()
}

// 获取进度条颜色
const getProgressColor = (percentage) => {
  if (percentage < 50) return '#52c41a'
  if (percentage < 80) return '#faad14'
  return '#ff4d4f'
}

// 获取指标状态
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

// 获取指标状态文本
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

// 获取指标标签
const getMetricLabel = (name) => {
  const labels = {
    'CPU核心数': '处理能力',
    '内存使用': '使用率',
    'Goroutine数': '并发数',
    'GC次数': '回收次数'
  }
  return labels[name] || '指标'
}

// 获取指标分析
const getMetricAnalysis = (metric) => {
  switch (metric.type) {
    case 'cpu':
      return `${metric.value}核心提供充足的并行处理能力`
    case 'memory':
      if (metric.percentage < 50) return '内存使用健康，运行状态良好'
      if (metric.percentage < 80) return '内存使用适中，需要持续观察'
      return '内存使用较高，建议检查内存泄漏'
    case 'goroutine':
      if (metric.rawValue < 50) return '并发负载轻，系统运行流畅'
      if (metric.rawValue < 100) return '并发负载适中，系统处理正常'
      return '并发负载较高，建议优化并发控制'
    case 'gc':
      return `垃圾回收${metric.value}次，内存管理正常`
    default:
      return '运行正常'
  }
}

// 获取指标参考
const getMetricReference = (name) => {
  const references = {
    'CPU核心数': '参考: 核心数反映系统最大并行处理能力',
    '内存使用': '参考: <50%优秀 | 50-80%良好 | >80%需关注',
    'Goroutine数': '参考: <50优秀 | 50-100良好 | >100需关注',
    'GC次数': '参考: 反映程序运行时间和内存分配模式'
  }
  return references[name] || ''
}

// 获取业务趋势
const getBusinessTrend = (stat) => {
  if (stat.name === '告警记录') {
    return stat.today > 0 ? 'up' : 'stable'
  }
  if (stat.active > 0) return 'up'
  return 'stable'
}

// 获取趋势图标
const getTrendIcon = (stat) => {
  const trend = getBusinessTrend(stat)
  const icons = {
    'up': 'el-icon-trend-charts',
    'down': 'el-icon-bottom',
    'stable': 'el-icon-minus'
  }
  return icons[trend] || 'el-icon-minus'
}

// 获取状态颜色
const getStatusColor = (status) => {
  return systemAPI.getStatusColor(status)
}

// 获取状态文本
const getStatusText = (status) => {
  return systemAPI.getStatusText(status)
}

// 获取状态图标
const getStatusIcon = (status) => {
  const icons = {
    'healthy': 'el-icon-success-filled',
    'degraded': 'el-icon-warning-filled',
    'unhealthy': 'el-icon-error-filled'
  }
  return icons[status] || 'el-icon-info-filled'
}

// 获取服务图标
const getServiceIcon = (serviceName) => {
  const icons = {
    '数据库连接': 'el-icon-coin',
    '邮件监控服务': 'el-icon-message',
    '通知服务': 'el-icon-bell',
    '缓存服务': 'el-icon-files'
  }
  return icons[serviceName] || 'el-icon-service'
}

// 获取统计图标
const getStatIcon = (statName) => {
  const icons = {
    '邮箱总数': 'el-icon-message-box',
    '告警规则': 'el-icon-warning',
    '通知渠道': 'el-icon-chat-line-square',
    '告警记录': 'el-icon-document',
    '通知记录': 'el-icon-bell'
  }
  return icons[statName] || 'el-icon-data-line'
}

// 数据清理相关方法
const getCleanupDescription = () => {
  const dataTypeTexts = {
    'alerts': '告警历史数据',
    'notifications': '通知历史数据',
    'both': '告警和通知历史数据'
  }
  return dataTypeTexts[cleanupForm.value.dataType] || '未知数据'
}

const getTimeRangeDescription = () => {
  const timeRangeTexts = {
    '1month': '1个月前的数据',
    '3months': '3个月前的数据',
    '6months': '6个月前的数据',
    '1year': '1年前的数据',
    '2years': '2年前的数据',
    'all': '所有历史数据'
  }
  return timeRangeTexts[cleanupForm.value.timeRange] || '未知范围'
}

const executeCleanup = async () => {
  cleanupLoading.value = true
  
  try {
    const response = await systemAPI.cleanup(cleanupForm.value.dataType, cleanupForm.value.timeRange)
    
    if (response.code === 200) {
      cleanupResult.value = {
        success: true,
        deleted_rows: response.data.deleted_rows,
        duration: response.data.duration,
        error: ''
      }
      
      ElMessage.success(`清理完成，删除了 ${response.data.deleted_rows} 条记录`)
      
      // 刷新业务统计数据
      await getBusinessStats()
    } else {
      throw new Error(response.message || '清理失败')
    }
  } catch (error) {
    cleanupResult.value = {
      success: false,
      deleted_rows: 0,
      duration: 0,
      error: error.message || '清理操作失败'
    }
    
    ElMessage.error('清理失败: ' + error.message)
  } finally {
    cleanupLoading.value = false
    showCleanupDialog.value = false
    showResultDialog.value = true
  }
}

const closeResultDialog = () => {
  showResultDialog.value = false
  // 重置表单
  cleanupForm.value = {
    dataType: 'both',
    timeRange: '3months',
    confirmed: false
  }
}

const formatDuration = (duration) => {
  if (!duration) return '0ms'
  
  // duration 是纳秒，转换为毫秒
  const ms = Math.round(duration / 1000000)
  if (ms < 1000) {
    return `${ms}ms`
  } else {
    return `${(ms / 1000).toFixed(2)}s`
  }
}



// 组件挂载时获取数据
onMounted(() => {
  getSystemHealth()
  getBusinessStats()
})
</script>

<style scoped>
.system-page {
  padding: 20px;
  background: #f5f7fa;
  min-height: calc(100vh - 60px);
}

/* 页面标题 */
.page-title {
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

/* 主要内容区域 */
.main-content {
  margin-bottom: 24px;
}

/* 卡片样式 */
.status-card,
.resource-card {
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

/* 性能说明面板 */
.performance-info {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 20px;
}

.info-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #4a5568;
  margin-bottom: 12px;
  font-size: 14px;
}

.info-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.info-item {
  font-size: 12px;
  color: #718096;
  line-height: 1.4;
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

/* 资源图表 */
.resource-charts {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 24px;
}

.chart-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px;
  background: #f9fafb;
  border-radius: 12px;
  border: 1px solid #f3f4f6;
  transition: all 0.2s ease;
}

.chart-item:hover {
  background: #f3f4f6;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

/* CPU核心数项目特殊样式 */
.cpu-chart-item {
  grid-column: span 2; /* 让CPU展示占据整行 */
}

.chart-header {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.chart-title {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.chart-name {
  font-size: 14px;
  font-weight: 500;
  color: #374151;
}

.chart-value {
  font-size: 18px;
  font-weight: 600;
  color: #6366f1;
}

.chart-status {
  font-size: 12px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 10px;
}

.chart-status.excellent {
  color: #059669;
  background: rgba(5, 150, 105, 0.1);
}

.chart-status.good {
  color: #0891b2;
  background: rgba(8, 145, 178, 0.1);
}

.chart-status.warning {
  color: #d97706;
  background: rgba(217, 119, 6, 0.1);
}

.chart-status.info {
  color: #6b7280;
  background: rgba(107, 114, 128, 0.1);
}

/* CPU核心数特殊展示 */
.cpu-cores-display {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  color: white;
  box-shadow: 0 4px 20px rgba(102, 126, 234, 0.3);
  transition: all 0.3s ease;
}

.cpu-cores-display:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 24px rgba(102, 126, 234, 0.4);
}

.cpu-icon {
  width: 48px;
  height: 48px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  backdrop-filter: blur(10px);
}

.cpu-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.cpu-title {
  font-size: 14px;
  font-weight: 500;
  opacity: 0.9;
}

.cpu-value {
  display: flex;
  align-items: baseline;
  gap: 4px;
}

.cores-number {
  font-size: 32px;
  font-weight: 700;
  line-height: 1;
}

.cores-unit {
  font-size: 16px;
  font-weight: 500;
  opacity: 0.8;
}

.cpu-description {
  font-size: 12px;
  opacity: 0.8;
  line-height: 1.3;
}

.cpu-status {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.15);
  border-radius: 8px;
  backdrop-filter: blur(10px);
}

.cpu-status i {
  font-size: 16px;
}

.cpu-status span {
  font-size: 11px;
  font-weight: 500;
}

/* 圆形进度图表 */
.circular-progress {
  margin: 16px 0;
}

.progress-circle {
  width: 100px;
  height: 100px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.progress-center {
  width: 70px;
  height: 70px;
  background: white;
  border-radius: 50%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.progress-percentage {
  font-size: 16px;
  font-weight: 700;
  color: #1f2937;
}

.progress-label {
  font-size: 10px;
  color: #6b7280;
  margin-top: 2px;
}

/* 指标分析 */
.metric-analysis {
  text-align: center;
  width: 100%;
}

.analysis-text {
  font-size: 12px;
  color: #374151;
  margin-bottom: 4px;
  line-height: 1.4;
}

.analysis-reference {
  font-size: 10px;
  color: #9ca3af;
  line-height: 1.3;
}

/* 业务统计 */
.business-stats {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  border: 1px solid #e5e7eb;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: #f9fafb;
  border-radius: 8px;
  border: 1px solid #f3f4f6;
  transition: all 0.2s ease;
  position: relative;
}

.stat-item:hover {
  background: #f3f4f6;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
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

.stat-name {
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
  position: absolute;
  top: 12px;
  right: 12px;
}

.trend-indicator {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
}

.trend-indicator.up {
  background: rgba(34, 197, 94, 0.1);
  color: #22c55e;
}

.trend-indicator.stable {
  background: rgba(107, 114, 128, 0.1);
  color: #6b7280;
}

/* 数据清理功能样式 */
.cleanup-section {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  border: 1px solid #e5e7eb;
  margin-top: 24px;
}

.cleanup-info {
  margin-top: 16px;
}

.cleanup-desc {
  color: #6b7280;
  font-size: 14px;
  margin-bottom: 16px;
}

.cleanup-stats {
  display: flex;
  gap: 32px;
}

.cleanup-stats .stat-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.cleanup-stats .stat-label {
  color: #374151;
  font-size: 14px;
}

.cleanup-stats .stat-value {
  color: #1f2937;
  font-weight: 600;
  font-size: 16px;
}

/* 清理对话框样式 */
.cleanup-dialog {
  padding: 8px 0;
}

.cleanup-warning {
  margin-bottom: 24px;
}

.cleanup-form {
  margin: 24px 0;
}

.cleanup-form .el-radio-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.cleanup-form .el-radio {
  margin-right: 0;
  margin-bottom: 8px;
}

.cleanup-preview {
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  padding: 12px;
}

.cleanup-preview p {
  margin: 0 0 4px 0;
  font-size: 14px;
}

.cleanup-time {
  color: #6b7280;
  font-size: 12px !important;
}

.cleanup-confirm {
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid #e5e7eb;
}

/* 清理结果样式 */
.cleanup-result {
  text-align: center;
  padding: 20px 0;
}

.result-icon {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  margin: 0 auto 16px;
}

.result-icon.success {
  background: rgba(34, 197, 94, 0.1);
  color: #22c55e;
}

.result-icon.error {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.result-content h4 {
  color: #1f2937;
  font-size: 18px;
  font-weight: 600;
  margin: 0 0 12px 0;
}

.result-content p {
  color: #6b7280;
  font-size: 14px;
  margin: 0 0 8px 0;
}

.error-message {
  color: #ef4444 !important;
  font-weight: 500;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .system-page {
    padding: 16px;
  }
  
  .status-overview {
    flex-direction: column;
    gap: 16px;
    text-align: center;
  }
  
  .system-metrics {
    gap: 16px;
  }
  
  .main-content .el-col {
    margin-bottom: 16px;
  }
  
  .resource-charts {
    grid-template-columns: 1fr;
  }
  
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .cleanup-stats {
    flex-direction: column;
    gap: 12px;
  }
}
</style> 