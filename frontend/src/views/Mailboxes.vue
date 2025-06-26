<template>
  <div class="mailboxes">
    <el-card>
      <template #header>
        <div class="page-header">
          <span>邮箱管理</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            添加邮箱
          </el-button>
        </div>
      </template>
      
      <!-- 搜索和筛选区域 -->
      <div class="filter-section">
        <el-row :gutter="20">
          <el-col :span="8">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索邮箱名称或地址"
              clearable
              @input="handleSearch"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
          </el-col>
          <el-col :span="6">
            <el-select v-model="statusFilter" placeholder="状态筛选" clearable @change="handleFilter">
              <el-option label="全部状态" value="" />
              <el-option label="激活" value="active" />
              <el-option label="停用" value="inactive" />
            </el-select>
          </el-col>
          <el-col :span="6">
            <el-select v-model="protocolFilter" placeholder="协议筛选" clearable @change="handleFilter">
              <el-option label="全部协议" value="" />
              <el-option label="IMAP" value="IMAP" />
              <el-option label="POP3" value="POP3" />
            </el-select>
          </el-col>
          <el-col :span="4">
            <el-button @click="refreshData" :loading="loading">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </el-col>
        </el-row>
      </div>

      <!-- 邮箱列表表格 -->
      <el-table 
        :data="filteredMailboxes" 
        style="width: 100%" 
        v-loading="loading"
        :default-sort="{ prop: 'created_at', order: 'descending' }"
      >
        <el-table-column prop="name" label="邮箱名称" min-width="150">
          <template #default="scope">
            <div class="name-cell">
              <span class="name">{{ scope.row.name }}</span>
              <span class="description" v-if="scope.row.description">{{ scope.row.description }}</span>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="email" label="邮箱地址" min-width="200" />
        
        <el-table-column prop="protocol" label="协议" width="80">
          <template #default="scope">
            <el-tag size="small" :type="scope.row.protocol === 'IMAP' ? 'primary' : 'success'">
              {{ scope.row.protocol }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="host" label="服务器" width="120" />
        
        <el-table-column prop="port" label="端口" width="80" />
        
        <el-table-column label="SSL" width="60">
          <template #default="scope">
            <el-icon v-if="scope.row.ssl" class="ssl-icon" color="#67c23a">
              <Check />
            </el-icon>
            <el-icon v-else color="#f56c6c">
              <Close />
            </el-icon>
          </template>
        </el-table-column>
        
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag 
              :type="scope.row.status === 'active' ? 'success' : 'danger'"
              size="small"
            >
              {{ scope.row.status === 'active' ? '激活' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="连接状态" width="120">
          <template #default="scope">
            <div class="connection-status">
              <el-button
                :type="getConnectionStatusType(scope.row.id)"
                size="small"
                :loading="testingConnections.includes(scope.row.id)"
                @click="testConnection(scope.row)"
              >
                {{ getConnectionStatusText(scope.row.id) }}
              </el-button>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="created_at" label="创建时间" width="160">
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
                @click="handleEdit(scope.row)"
                :icon="Edit"
              >
                编辑
              </el-button>
              <el-button 
                :type="scope.row.status === 'active' ? 'warning' : 'success'" 
                size="small" 
                @click="toggleStatus(scope.row)"
                :icon="Switch"
              >
                {{ scope.row.status === 'active' ? '停用' : '启用' }}
              </el-button>
              <el-dropdown @command="(command) => handleDropdownCommand(command, scope.row)">
                <el-button size="small" type="info">
                  更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item 
                      command="test"
                    >
                      测试连接
                    </el-dropdown-item>
                    <el-dropdown-item 
                      command="duplicate"
                    >
                      复制邮箱
                    </el-dropdown-item>
                    <el-dropdown-item 
                      command="delete"
                      class="text-danger"
                      divided
                    >
                      删除
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
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

    <!-- 邮箱添加/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑邮箱' : '添加邮箱'"
      width="600px"
      :close-on-click-modal="false"
    >
      <MailboxForm
        ref="mailboxFormRef"
        v-model="formData"
        :loading="testingForm || submitting"
      />
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="info" @click="testFormConnection" :loading="testingForm">
            测试连接
          </el-button>
          <el-button type="primary" @click="handleSubmit" :loading="submitting">
            {{ isEdit ? '更新' : '创建' }}
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import { 
  Plus, Search, Refresh, Edit, Delete, Switch, 
  Check, Close, ArrowDown 
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { mailboxAPI } from '@/api'
import MailboxForm from '@/components/MailboxForm.vue'

// 响应式数据
const loading = ref(false)
const mailboxes = ref([])
const searchKeyword = ref('')
const statusFilter = ref('')
const protocolFilter = ref('')
const testingConnections = ref([])
const connectionStatuses = ref({})

// 分页数据
const pagination = ref({
  page: 1,
  size: 20,
  total: 0
})

// 对话框和表单
const dialogVisible = ref(false)
const isEdit = ref(false)
const mailboxFormRef = ref()
const testingForm = ref(false)
const submitting = ref(false)

// 表单数据
const formData = ref({
  name: '',
  email: '',
  protocol: 'IMAP',
  host: '',
  port: 993,
  username: '',
  password: '',
  ssl: true,
  description: ''
})

// 计算属性
const filteredMailboxes = computed(() => {
  return mailboxes.value.filter(mailbox => {
    const matchesKeyword = searchKeyword.value === '' || 
      mailbox.name.toLowerCase().includes(searchKeyword.value.toLowerCase()) ||
      mailbox.email.toLowerCase().includes(searchKeyword.value.toLowerCase())
    
    const matchesStatus = statusFilter.value === '' || 
      mailbox.status === statusFilter.value
    
    const matchesProtocol = protocolFilter.value === '' || 
      mailbox.protocol === protocolFilter.value
    
    return matchesKeyword && matchesStatus && matchesProtocol
  })
})

// 生命周期
onMounted(() => {
  loadMailboxes()
})

// 方法
const loadMailboxes = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.value.page,
      size: pagination.value.size
    }
    const response = await mailboxAPI.list(params)
    mailboxes.value = response.data.list || []
    pagination.value.total = response.data.total || 0
  } catch (error) {
    console.error('加载邮箱列表失败:', error)
    ElMessage.error('加载邮箱列表失败')
  } finally {
    loading.value = false
  }
}

const refreshData = () => {
  loadMailboxes()
}

const handleSearch = () => {
  // 搜索是实时的，通过计算属性实现
}

const handleFilter = () => {
  // 筛选是实时的，通过计算属性实现
}

const handleSizeChange = (size) => {
  pagination.value.size = size
  pagination.value.page = 1
  loadMailboxes()
}

const handleCurrentChange = (page) => {
  pagination.value.page = page
  loadMailboxes()
}

const handleAdd = () => {
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
}

const handleEdit = async (row) => {
  isEdit.value = true
  try {
    // 获取包含密码的完整邮箱信息
    const response = await mailboxAPI.getForEdit(row.id)
    formData.value = { ...response.data }
    dialogVisible.value = true
  } catch (error) {
    console.error('获取邮箱详情失败:', error)
    ElMessage.error('获取邮箱详情失败')
    // 如果获取失败，使用原有数据但清空密码
    formData.value = { ...row }
    formData.value.password = ''
    dialogVisible.value = true
  }
}

const resetForm = () => {
  formData.value = {
    name: '',
    email: '',
    protocol: 'IMAP',
    host: '',
    port: 993,
    username: '',
    password: '',
    ssl: true,
    description: ''
  }
  if (mailboxFormRef.value) {
    mailboxFormRef.value.resetFields()
  }
}

const testFormConnection = async () => {
  try {
    const isValid = await mailboxFormRef.value.validate()
    if (!isValid) {
      ElMessage.error('请先填写完整的邮箱配置信息')
      return
    }
    
    testingForm.value = true
    const testData = { 
      ...formData.value,
      username: formData.value.email // 使用邮箱地址作为用户名
    }
    
    const response = await mailboxAPI.test(testData)
    
    // 检查响应数据中的连接状态
    if (response.data && response.data.status === 'connected') {
      ElMessage.success('连接测试成功！')
    } else {
      // 连接失败，显示具体错误信息
      const errorMessage = response.data?.message || '连接测试失败'
      ElMessage.error(`连接测试失败: ${errorMessage}`)
    }
  } catch (error) {
    // HTTP层面的错误
    if (error.response) {
      ElMessage.error(`连接测试失败: ${error.response.data?.message || '未知错误'}`)
    } else {
      ElMessage.error('连接测试失败，请检查配置信息')
    }
  } finally {
    testingForm.value = false
  }
}

const handleSubmit = async () => {
  try {
    const isValid = await mailboxFormRef.value.validate()
    if (!isValid) {
      return
    }
    
    submitting.value = true
    const submitData = { 
      ...formData.value,
      username: formData.value.email // 使用邮箱地址作为用户名
    }
    
    if (isEdit.value) {
      // 编辑时，如果密码为空则不更新密码
      if (!submitData.password) {
        delete submitData.password
      }
      await mailboxAPI.update(submitData.id, submitData)
      ElMessage.success('邮箱配置更新成功')
    } else {
      await mailboxAPI.create(submitData)
      ElMessage.success('邮箱配置创建成功')
    }
    
    dialogVisible.value = false
    loadMailboxes()
  } catch (error) {
    ElMessage.error(isEdit.value ? '更新邮箱配置失败' : '创建邮箱配置失败')
  } finally {
    submitting.value = false
  }
}

const testConnection = async (row) => {
  if (testingConnections.value.includes(row.id)) return
  
  testingConnections.value.push(row.id)
  try {
    const response = await mailboxAPI.test(row.id)
    
    // 检查响应数据中的连接状态
    if (response.data && response.data.status === 'connected') {
      connectionStatuses.value[row.id] = 'success'
      ElMessage.success(`${row.name} 连接测试成功`)
    } else {
      // 连接失败，显示具体错误信息
      const errorMessage = response.data?.message || '连接失败'
      connectionStatuses.value[row.id] = 'error'
      ElMessage.error(`${row.name} 连接测试失败: ${errorMessage}`)
    }
  } catch (error) {
    connectionStatuses.value[row.id] = 'error'
    if (error.response) {
      ElMessage.error(`${row.name} 连接测试失败: ${error.response.data?.message || '连接失败'}`)
    } else {
      ElMessage.error(`${row.name} 连接测试失败，请检查配置信息`)
    }
  } finally {
    testingConnections.value = testingConnections.value.filter(id => id !== row.id)
    // 3秒后清除连接状态
    setTimeout(() => {
      delete connectionStatuses.value[row.id]
    }, 3000)
  }
}

const toggleStatus = async (row) => {
  try {
    const newStatus = row.status === 'active' ? 'inactive' : 'active'
    await mailboxAPI.update(row.id, { status: newStatus })
    
    row.status = newStatus
    ElMessage.success(`邮箱已${newStatus === 'active' ? '启用' : '停用'}`)
  } catch (error) {
    ElMessage.error('状态更新失败')
  }
}

// 下拉菜单命令处理
const handleDropdownCommand = async (command, row) => {
  switch (command) {
    case 'test':
      await testConnection(row)
      break
    case 'duplicate':
      await handleDuplicate(row)
      break
    case 'delete':
      await handleDelete(row)
      break
  }
}

// 复制邮箱
const handleDuplicate = async (row) => {
  try {
    const duplicateData = {
      name: `${row.name} - 副本`,
      email: row.email,
      protocol: row.protocol,
      host: row.host,
      port: row.port,
      username: row.username,
      password: '', // 密码需要重新设置
      ssl: row.ssl,
      status: 'inactive', // 副本默认为停用状态
      description: `${row.description || ''} (复制自 ${row.name})`
    }
    
    // 设置表单数据并打开编辑对话框
    isEdit.value = false
    formData.value = duplicateData
    dialogVisible.value = true
    ElMessage.info('请设置新邮箱的密码')
  } catch (error) {
    console.error('复制邮箱失败:', error)
    ElMessage.error('复制邮箱失败')
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除邮箱 "${row.name}" 吗？此操作不可恢复。`,
      '删除确认',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await mailboxAPI.delete(row.id)
    ElMessage.success('邮箱删除成功')
    loadMailboxes()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除邮箱失败')
    }
  }
}

const getConnectionStatusType = (id) => {
  const status = connectionStatuses.value[id]
  if (status === 'success') return 'success'
  if (status === 'error') return 'danger'
  return 'primary'
}

const getConnectionStatusText = (id) => {
  const status = connectionStatuses.value[id]
  if (status === 'success') return '连接正常'
  if (status === 'error') return '连接失败'
  return '测试连接'
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-section {
  margin-bottom: 20px;
}

.name-cell {
  display: flex;
  flex-direction: column;
}

.name {
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.description {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
}

.ssl-icon {
  font-size: 16px;
}

.connection-status {
  display: flex;
  align-items: center;
}

.pagination-wrapper {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

:deep(.el-table .el-table__cell) {
  padding: 8px 0;
}

:deep(.el-input-number .el-input__inner) {
  text-align: left;
}

.text-danger {
  color: #f56c6c;
}

:deep(.el-dropdown-menu__item.text-danger:hover) {
  background-color: #fef0f0;
  color: #f56c6c;
}
</style> 