<template>
  <div class="alerts">
    <el-card>
      <template #header>
        <div class="page-header">
          <span>告警历史</span>
        </div>
      </template>

      <!-- 搜索筛选 -->
      <div class="search-filter">
        <el-row :gutter="20">
          <el-col :span="6">
            <el-input
              v-model="searchQuery.subject"
              placeholder="搜索告警主题"
              @input="handleSearch"
              clearable
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
          </el-col>
          <el-col :span="5">
            <el-input
              v-model="searchQuery.sender"
              placeholder="搜索发件人"
              @input="handleSearch"
              clearable
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
          </el-col>
          <el-col :span="4">
            <el-select
              v-model="searchQuery.status"
              placeholder="选择状态"
              @change="handleSearch"
              clearable
            >
              <el-option label="待处理" value="pending" />
              <el-option label="已发送" value="sent" />
              <el-option label="发送失败" value="failed" />
            </el-select>
          </el-col>
          <el-col :span="7">
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
          <el-col :span="2">
            <el-button @click="handleReset" :icon="Refresh">重置</el-button>
          </el-col>
        </el-row>
      </div>
      
      <!-- 告警列表 -->
      <el-table 
        :data="alerts" 
        v-loading="loading" 
        style="width: 100%"
        @sort-change="handleSortChange"
      >
        <el-table-column prop="subject" label="告警主题" min-width="200" show-overflow-tooltip />
        <el-table-column prop="sender" label="发件人" width="180" show-overflow-tooltip />
        <el-table-column label="邮箱源" width="150">
          <template #default="scope">
            {{ scope.row.mailbox?.name || '未知邮箱' }}
          </template>
        </el-table-column>
        <el-table-column label="匹配规则" width="150">
          <template #default="scope">
            {{ scope.row.rule_group?.name || '无规则' }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" sortable>
          <template #default="scope">
            <el-tag :type="getStatusColor(scope.row.status)">
              {{ getStatusLabel(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="发送渠道" width="120">
          <template #default="scope">
            <el-tag
              v-if="scope.row.sent_channels"
              type="success"
              size="small"
            >
              {{ scope.row.sent_channels.split(',').length }}个渠道
            </el-tag>
            <span v-else class="text-gray-400">未发送</span>
          </template>
        </el-table-column>
        <el-table-column prop="retry_count" label="重试次数" width="100" />
        <el-table-column prop="received_at" label="接收时间" width="180" sortable>
          <template #default="scope">
            {{ formatDate(scope.row.received_at) }}
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
                v-if="scope.row.status === 'failed' || scope.row.status === 'pending'"
                type="warning" 
                size="small" 
                @click="handleRetry(scope.row)"
                :icon="Refresh"
              >
                重试发送
              </el-button>
              <el-dropdown 
                @command="(command) => handleDropdownCommand(command, scope.row)"
              >
                <el-button type="info" size="small">
                  更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item 
                      v-if="scope.row.status === 'pending'"
                      command="sent"
                    >
                      标记为已发送
                    </el-dropdown-item>
                    <el-dropdown-item 
                      v-if="scope.row.status === 'pending'"
                      command="failed"
                    >
                      标记为失败
                    </el-dropdown-item>
                    <el-dropdown-item 
                      v-if="scope.row.status === 'sent'"
                      command="retry"
                    >
                      重新发送
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

    <!-- 告警详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="告警详情"
      width="800px"
      @close="handleDetailDialogClose"
    >
      <div v-if="currentAlert" class="alert-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="告警主题">
            {{ currentAlert.subject }}
          </el-descriptions-item>
          <el-descriptions-item label="发件人">
            {{ currentAlert.sender }}
          </el-descriptions-item>
          <el-descriptions-item label="邮箱源">
            {{ currentAlert.mailbox?.name || '未知邮箱' }}
          </el-descriptions-item>
          <el-descriptions-item label="匹配规则">
            {{ currentAlert.rule_group?.name || '无规则' }}
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusColor(currentAlert.status)">
              {{ getStatusLabel(currentAlert.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="重试次数">
            {{ currentAlert.retry_count || 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="接收时间">
            {{ formatDate(currentAlert.received_at) }}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatDate(currentAlert.created_at) }}
          </el-descriptions-item>
        </el-descriptions>
        
        <el-divider>发送渠道</el-divider>
        <div v-if="currentAlert.sent_channels" class="sent-channels">
          <el-tag
            v-for="channel in currentAlert.sent_channels.split(',')"
            :key="channel"
            type="success"
            style="margin-right: 8px; margin-bottom: 8px;"
          >
            {{ channel }}
          </el-tag>
        </div>
        <div v-else class="text-gray-400">未发送到任何渠道</div>
        
        <el-divider>错误信息</el-divider>
        <div v-if="currentAlert.error_msg" class="error-msg">
          <el-alert
            :title="currentAlert.error_msg"
            type="error"
            :closable="false"
            show-icon
          />
        </div>
        <div v-else class="text-gray-400">无错误信息</div>
        
        <el-divider>邮件内容</el-divider>
        <div class="email-content">
          <el-input
            v-model="currentAlert.content"
            type="textarea"
            :rows="8"
            readonly
            placeholder="邮件内容"
          />
        </div>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="detailDialogVisible = false">关闭</el-button>
          <el-button 
            v-if="currentAlert && (currentAlert.status === 'failed' || currentAlert.status === 'pending')"
            type="warning" 
            @click="handleRetry(currentAlert)"
          >
            重试发送
          </el-button>
          <el-button 
            v-if="currentAlert && currentAlert.status === 'sent'"
            type="primary" 
            @click="handleRetry(currentAlert)"
          >
            重新发送
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
import { alertsAPI } from '@/api'
import { formatDate } from '@/utils'

// 响应式数据
const alerts = ref([])
const loading = ref(false)

// 搜索筛选
const searchQuery = reactive({
  subject: '',
  sender: '',
  status: ''
})

// 排序
const sortOptions = reactive({
  sort_by: 'created_at',
  sort_order: 'desc'
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
const currentAlert = ref(null)

// 方法
const fetchAlerts = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      size: pagination.size,
      ...searchQuery,
      ...sortOptions
    }
    
    // 处理日期范围
    if (dateRange.value && dateRange.value.length === 2) {
      params.start_date = formatDate(dateRange.value[0], 'YYYY-MM-DD')
      params.end_date = formatDate(dateRange.value[1], 'YYYY-MM-DD')
    }
    
    // 过滤空值参数
    Object.keys(params).forEach(key => {
      if (params[key] === '' || params[key] === null || params[key] === undefined) {
        delete params[key]
      }
    })
    
    const response = await alertsAPI.list(params)
    alerts.value = response.data.alerts || []
    pagination.total = response.data.total || 0
  } catch (error) {
    console.error('获取告警列表失败:', error)
    ElMessage.error('获取告警列表失败')
  } finally {
    loading.value = false
  }
}



const handleSearch = () => {
  pagination.page = 1
  fetchAlerts()
}

const handleReset = () => {
  Object.assign(searchQuery, {
    subject: '',
    sender: '',
    status: ''
  })
  dateRange.value = ''
  Object.assign(sortOptions, {
    sort_by: 'created_at',
    sort_order: 'desc'
  })
  pagination.page = 1
  fetchAlerts()
}

const handleSortChange = ({ column, prop, order }) => {
  if (prop && order) {
    sortOptions.sort_by = prop
    sortOptions.sort_order = order === 'ascending' ? 'asc' : 'desc'
    pagination.page = 1
    fetchAlerts()
  }
}

const handleSizeChange = (val) => {
  pagination.size = val
  pagination.page = 1
  fetchAlerts()
}

const handleCurrentChange = (val) => {
  pagination.page = val
  fetchAlerts()
}

const handleView = (alert) => {
  currentAlert.value = alert
  detailDialogVisible.value = true
}

const handleRetry = async (alert) => {
  try {
    await ElMessageBox.confirm(
      `确定要重试发送告警 "${alert.subject}" 吗？`,
      '确认重试',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await alertsAPI.retry(alert.id)
    ElMessage.success('重试发送成功')
    fetchAlerts()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('重试告警失败:', error)
      ElMessage.error('重试发送失败')
    }
  }
}

// 下拉菜单命令处理
const handleDropdownCommand = async (command, alert) => {
  switch (command) {
    case 'sent':
    case 'failed':
      await handleStatusChange(alert, command)
      break
    case 'retry':
      await handleRetry(alert)
      break
    case 'delete':
      await handleDelete(alert)
      break
  }
}

// 删除告警记录
const handleDelete = async (alert) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除告警记录 "${alert.subject}" 吗？此操作不可恢复。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await alertsAPI.delete(alert.id)
    ElMessage.success('删除成功')
    fetchAlerts()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleStatusChange = async (alert, status) => {
  try {
    await alertsAPI.updateStatus(alert.id, { status })
    ElMessage.success('状态更新成功')
    fetchAlerts()
  } catch (error) {
    ElMessage.error('状态更新失败')
  }
}

const handleDetailDialogClose = () => {
  currentAlert.value = null
}

// 工具方法
const getStatusColor = (status) => {
  const colors = {
    pending: 'warning',
    sent: 'success',
    failed: 'danger'
  }
  return colors[status] || 'info'
}

const getStatusLabel = (status) => {
  const labels = {
    pending: '待处理',
    sent: '已发送',
    failed: '发送失败'
  }
  return labels[status] || status
}

// 生命周期
onMounted(() => {
  fetchAlerts()
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

.alert-detail {
  max-height: 600px;
  overflow-y: auto;
}

.sent-channels {
  min-height: 40px;
}

.error-msg {
  margin: 10px 0;
}

.email-content {
  margin-top: 10px;
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