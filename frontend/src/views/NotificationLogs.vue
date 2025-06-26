<template>
  <div class="notification-logs">
    <el-card>
      <template #header>
        <div class="page-header">
          <span>通知发送历史</span>
        </div>
      </template>

      <!-- 搜索筛选 -->
      <div class="search-filter">
        <el-row :gutter="20">
          <el-col :span="5">
            <el-select
              v-model="searchQuery.channel_id"
              placeholder="选择渠道"
              @change="handleSearch"
              clearable
            >
              <el-option
                v-for="channel in channelOptions"
                :key="channel.id"
                :label="channel.name"
                :value="channel.id"
              />
            </el-select>
          </el-col>
          <el-col :span="5">
            <el-select
              v-model="searchQuery.status"
              placeholder="选择状态"
              @change="handleSearch"
              clearable
            >
              <el-option label="待发送" value="pending" />
              <el-option label="发送成功" value="success" />
              <el-option label="发送失败" value="failed" />
            </el-select>
          </el-col>
          <el-col :span="6">
            <el-date-picker
              v-model="dateRange"
              type="daterange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              @change="handleSearch"
              clearable
            />
          </el-col>
          <el-col :span="6">
            <el-input
              v-model="searchQuery.content"
              placeholder="搜索发送内容"
              @input="handleSearch"
              clearable
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
          </el-col>
          <el-col :span="2">
            <el-button @click="handleReset" :icon="Refresh">重置</el-button>
          </el-col>
        </el-row>
      </div>
      
      <!-- 通知记录列表 -->
      <el-table 
        :data="logs" 
        v-loading="loading" 
        style="width: 100%"
        @sort-change="handleSortChange"
      >
        <el-table-column label="渠道信息" width="180">
          <template #default="scope">
            <div>
              <div style="font-weight: 500;">{{ scope.row.channel?.name || '未知渠道' }}</div>
              <el-tag :type="getChannelTypeColor(scope.row.channel?.type)" size="small">
                {{ getChannelTypeLabel(scope.row.channel?.type) }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="告警信息" min-width="200">
          <template #default="scope">
            <div v-if="scope.row.alert">
              <div class="alert-subject">{{ scope.row.alert.subject }}</div>
              <div class="alert-sender">发件人: {{ scope.row.alert.sender }}</div>
            </div>
            <span v-else class="text-gray-400">无关联告警</span>
          </template>
        </el-table-column>
        <el-table-column prop="content" label="发送内容" min-width="300" show-overflow-tooltip>
          <template #default="scope">
            <div class="content-preview">
              {{ getContentPreview(scope.row.content) }}
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="发送状态" width="120" sortable>
          <template #default="scope">
            <el-tag :type="getStatusColor(scope.row.status)">
              {{ getStatusLabel(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="retry_count" label="重试次数" width="100" align="center" />
        <el-table-column prop="sent_at" label="发送时间" width="180" sortable>
          <template #default="scope">
            <div v-if="scope.row.sent_at">
              {{ formatDate(scope.row.sent_at) }}
            </div>
            <span v-else class="text-gray-400">未发送</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" sortable>
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="scope">
            <el-button-group>
              <el-button 
                type="primary" 
                size="small" 
                @click="handleView(scope.row)"
                :icon="View"
              >
                查看详情
              </el-button>
              <el-button 
                v-if="scope.row.status === 'failed'"
                type="warning" 
                size="small" 
                @click="handleRetry(scope.row)"
                :icon="Refresh"
              >
                重试
              </el-button>
              <el-dropdown @command="(command) => handleDropdownCommand(command, scope.row)">
                <el-button size="small" type="info">
                  更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item 
                      v-if="scope.row.status !== 'failed'"
                      command="retry"
                    >
                      重试发送
                    </el-dropdown-item>
                    <el-dropdown-item 
                      command="download"
                    >
                      下载内容
                    </el-dropdown-item>
                    <el-dropdown-item 
                      command="delete"
                      class="text-danger"
                      divided
                    >
                      删除记录
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
      
      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.size"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 通知详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="通知发送详情"
      width="800px"
      @close="handleDetailDialogClose"
    >
      <div v-if="currentLog" class="log-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="渠道名称">
            {{ currentLog.channel?.name || '未知渠道' }}
          </el-descriptions-item>
          <el-descriptions-item label="渠道类型">
            <el-tag :type="getChannelTypeColor(currentLog.channel?.type)">
              {{ getChannelTypeLabel(currentLog.channel?.type) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="发送状态">
            <el-tag :type="getStatusColor(currentLog.status)">
              {{ getStatusLabel(currentLog.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="重试次数">
            {{ currentLog.retry_count || 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatDate(currentLog.created_at) }}
          </el-descriptions-item>
          <el-descriptions-item label="发送时间">
            {{ currentLog.sent_at ? formatDate(currentLog.sent_at) : '未发送' }}
          </el-descriptions-item>
        </el-descriptions>
        
        <el-divider>关联告警</el-divider>
        <div v-if="currentLog.alert" class="alert-info">
          <el-descriptions :column="1" border>
            <el-descriptions-item label="告警主题">
              {{ currentLog.alert.subject }}
            </el-descriptions-item>
            <el-descriptions-item label="发件人">
              {{ currentLog.alert.sender }}
            </el-descriptions-item>
            <el-descriptions-item label="接收时间">
              {{ formatDate(currentLog.alert.received_at) }}
            </el-descriptions-item>
          </el-descriptions>
        </div>
        <div v-else class="text-gray-400">无关联告警信息</div>
        
        <el-divider>发送内容</el-divider>
        <div class="send-content">
          <el-input
            v-model="currentLog.content"
            type="textarea"
            :rows="8"
            readonly
            placeholder="发送内容"
          />
        </div>
        
        <el-divider>错误信息</el-divider>
        <div v-if="currentLog.error_msg" class="error-info">
          <el-alert
            :title="currentLog.error_msg"
            type="error"
            :closable="false"
            show-icon
          />
        </div>
        <div v-else class="text-gray-400">无错误信息</div>
        
        <el-divider>响应数据</el-divider>
        <div v-if="currentLog.response_data" class="response-data">
          <el-input
            v-model="currentLog.response_data"
            type="textarea"
            :rows="4"
            readonly
            placeholder="响应数据"
          />
        </div>
        <div v-else class="text-gray-400">无响应数据</div>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="detailDialogVisible = false">关闭</el-button>
          <el-button 
            v-if="currentLog && currentLog.status === 'failed'"
            type="warning" 
            @click="handleRetry(currentLog)"
          >
            重试发送
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive } from 'vue'
import { Search, Refresh, ArrowDown, View } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { channelsAPI, notificationLogsAPI } from '@/api'
import { formatDate } from '@/utils'

// 响应式数据
const logs = ref([])
const loading = ref(false)
const channelOptions = ref([])

// 搜索筛选
const searchQuery = reactive({
  channel_id: '',
  status: '',
  content: ''
})

const dateRange = ref('')

// 分页
const pagination = reactive({
  page: 1,
  size: 20,
  total: 0
})

// 详情对话框
const detailDialogVisible = ref(false)
const currentLog = ref(null)

// 方法
const fetchLogs = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      size: pagination.size,
      ...searchQuery
    }
    
    // 处理日期范围
    if (dateRange.value && dateRange.value.length === 2) {
      params.start_date = formatDate(dateRange.value[0], 'YYYY-MM-DD')
      params.end_date = formatDate(dateRange.value[1], 'YYYY-MM-DD')
    }
    
    // 调用通知日志API
    const response = await notificationLogsAPI.list(params)
    
    logs.value = response.data.logs || []
    pagination.total = response.data.total || 0
  } catch (error) {
    console.error('获取通知历史失败:', error)
    ElMessage.error('获取通知历史失败')
  } finally {
    loading.value = false
  }
}

const fetchChannelOptions = async () => {
  try {
    const response = await channelsAPI.list({ size: 100 })
    channelOptions.value = response.data.channels || []
  } catch (error) {
    console.error('获取渠道选项失败:', error)
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchLogs()
}

const handleReset = () => {
  Object.assign(searchQuery, {
    channel_id: '',
    status: '',
    content: ''
  })
  dateRange.value = ''
  pagination.page = 1
  fetchLogs()
}

const handleSortChange = ({ column, prop, order }) => {
  // 实现排序逻辑
  console.log('排序变化:', { column, prop, order })
}

const handleSizeChange = (val) => {
  pagination.size = val
  pagination.page = 1
  fetchLogs()
}

const handleCurrentChange = (val) => {
  pagination.page = val
  fetchLogs()
}

const handleView = (log) => {
  currentLog.value = log
  detailDialogVisible.value = true
}

// 下拉菜单命令处理
const handleDropdownCommand = async (command, log) => {
  switch (command) {
    case 'retry':
      await handleRetry(log)
      break
    case 'download':
      await handleDownload(log)
      break
    case 'delete':
      await handleDelete(log)
      break
  }
}

// 下载内容
const handleDownload = async (log) => {
  try {
    const content = log.content || '无内容'
    const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `notification_${log.id}_${new Date().getTime()}.txt`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    ElMessage.success('下载成功')
  } catch (error) {
    console.error('下载失败:', error)
    ElMessage.error('下载失败')
  }
}

// 删除记录
const handleDelete = async (log) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除这条通知记录吗？此操作不可恢复。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    // 调用删除API
    await notificationLogsAPI.delete(log.id)
    ElMessage.success('删除成功')
    fetchLogs()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

const handleRetry = async (log) => {
  try {
    await ElMessageBox.confirm(
      `确定要重试发送通知吗？`,
      '确认重试',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    // 调用重试API
    await notificationLogsAPI.retry(log.id)
    ElMessage.success('通知已加入重试队列')
    fetchLogs()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('重试发送失败:', error)
      ElMessage.error('重试发送失败')
    }
  }
}

const handleDetailDialogClose = () => {
  currentLog.value = null
}

// 工具方法
const getStatusColor = (status) => {
  const colors = {
    pending: 'warning',
    success: 'success',
    failed: 'danger'
  }
  return colors[status] || 'info'
}

const getStatusLabel = (status) => {
  const labels = {
    pending: '待发送',
    success: '发送成功',
    failed: '发送失败'
  }
  return labels[status] || status
}

const getChannelTypeLabel = (type) => {
  const labels = {
    wechat: '企业微信',
    dingtalk: '钉钉',
    email: '邮件',
    webhook: 'Webhook'
  }
  return labels[type] || type
}

const getChannelTypeColor = (type) => {
  const colors = {
    wechat: 'success',
    dingtalk: 'primary',
    email: 'warning',
    webhook: 'info'
  }
  return colors[type] || 'info'
}

const getContentPreview = (content) => {
  if (!content) return '无内容'
  return content.length > 100 ? content.substring(0, 100) + '...' : content
}

// 生命周期
onMounted(() => {
  fetchChannelOptions()
  fetchLogs()
})
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-filter {
  margin-bottom: 20px;
}

.pagination-container {
  margin-top: 20px;
  text-align: right;
}

.alert-subject {
  font-weight: 500;
  margin-bottom: 4px;
}

.alert-sender {
  font-size: 12px;
  color: var(--el-text-color-regular);
}

.content-preview {
  max-height: 60px;
  overflow: hidden;
  line-height: 1.4;
}

.log-detail {
  max-height: 600px;
  overflow-y: auto;
}

.alert-info,
.send-content,
.error-info,
.response-data {
  margin: 10px 0;
}

.text-gray-400 {
  color: #9ca3af;
}

.dialog-footer {
  text-align: right;
}

.text-danger {
  color: #f56c6c;
}

:deep(.el-dropdown-menu__item.text-danger:hover) {
  background-color: #fef0f0;
  color: #f56c6c;
}
</style> 