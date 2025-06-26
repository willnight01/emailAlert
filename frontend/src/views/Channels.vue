<template>
  <div class="channels">
    <el-card>
      <template #header>
        <div class="page-header">
          <span>通知渠道管理</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            添加渠道
          </el-button>
        </div>
      </template>

      <!-- 搜索筛选 -->
      <div class="search-filter">
        <el-row :gutter="20">
          <el-col :span="8">
            <el-input
              v-model="searchQuery.name"
              placeholder="搜索渠道名称"
              @input="handleSearch"
              clearable
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
          </el-col>
          <el-col :span="6">
            <el-select
              v-model="searchQuery.type"
              placeholder="选择渠道类型"
              @change="handleSearch"
              clearable
            >
              <el-option
                v-for="type in channelTypes"
                :key="type"
                :label="getTypeLabel(type)"
                :value="type"
              />
            </el-select>
          </el-col>
          <el-col :span="6">
            <el-select
              v-model="searchQuery.status"
              placeholder="选择状态"
              @change="handleSearch"
              clearable
            >
              <el-option label="激活" value="active" />
              <el-option label="停用" value="inactive" />
            </el-select>
          </el-col>
          <el-col :span="4">
            <el-button @click="handleReset" :icon="Refresh">重置</el-button>
          </el-col>
        </el-row>
      </div>
      
      <!-- 渠道列表 -->
      <el-table 
        :data="channels" 
        v-loading="loading" 
        style="width: 100%"
        @sort-change="handleSortChange"
      >
        <el-table-column prop="name" label="渠道名称" sortable />
        <el-table-column prop="type" label="类型" width="120" sortable>
          <template #default="scope">
            <el-tag :type="getTypeColor(scope.row.type)">
              {{ getTypeLabel(scope.row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100" sortable>
          <template #default="scope">
            <el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'">
              {{ scope.row.status === 'active' ? '激活' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="test_result" label="测试状态" width="120">
          <template #default="scope">
            <el-tag 
              v-if="scope.row.test_result" 
              :type="scope.row.test_result.includes('成功') ? 'success' : 'danger'"
              size="small"
            >
              {{ scope.row.test_result.includes('成功') ? '正常' : '异常' }}
            </el-tag>
            <span v-else class="text-gray-400">未测试</span>
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
                @click="handleTest(scope.row)"
                :loading="testingChannels.includes(scope.row.id)"
                :icon="testingChannels.includes(scope.row.id) ? '' : 'Connection'"
              >
                测试
              </el-button>
              <el-button 
                type="success" 
                size="small" 
                @click="handleEdit(scope.row)"
                :icon="Edit"
              >
                编辑
              </el-button>
              <el-dropdown @command="(command) => handleDropdownCommand(command, scope.row)">
                <el-button size="small" type="info">
                  更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item 
                      command="send"
                    >
                      发送测试
                    </el-dropdown-item>
                    <el-dropdown-item 
                      command="duplicate"
                    >
                      复制渠道
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

    <!-- 添加/编辑渠道对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      @close="handleDialogClose"
    >
      <el-form
        ref="channelFormRef"
        :model="channelForm"
        :rules="channelFormRules"
        label-width="100px"
      >
        <el-form-item label="渠道名称" prop="name">
          <el-input v-model="channelForm.name" placeholder="请输入渠道名称" />
        </el-form-item>
        <el-form-item label="渠道类型" prop="type">
          <el-select v-model="channelForm.type" placeholder="请选择渠道类型" @change="handleTypeChange">
            <el-option
              v-for="type in channelTypes"
              :key="type"
              :label="getTypeLabel(type)"
              :value="type"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="channelForm.status">
            <el-radio label="active">激活</el-radio>
            <el-radio label="inactive">停用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input 
            v-model="channelForm.description" 
            type="textarea" 
            placeholder="请输入渠道描述"
            :rows="2"
          />
        </el-form-item>
        
        <!-- 模版选择 -->
        <el-form-item label="消息模版" prop="template_id">
          <el-select
            v-model="channelForm.template_id"
            placeholder="选择消息模版（不选择则使用默认模版）"
            clearable
            filterable
            style="width: 100%"
          >
            <el-option
              v-for="template in filteredTemplates"
              :key="template.id"
              :label="`${template.name} (${getTypeLabel(template.type)})`"
              :value="template.id"
            >
              <div style="display: flex; justify-content: space-between; align-items: center;">
                <span>{{ template.name }}</span>
                <div style="display: flex; align-items: center; gap: 8px;">
                  <el-tag 
                    v-if="template.is_default" 
                    size="small" 
                    type="success"
                  >
                    默认
                  </el-tag>
                  <span style="color: #999; font-size: 12px">{{ getTypeLabel(template.type) }}</span>
                </div>
              </div>
            </el-option>
          </el-select>
          <div style="font-size: 12px; color: #999; margin-top: 4px;">
            <span v-if="channelForm.type">
              推荐选择"{{ getTypeLabel(channelForm.type) }}"类型的模版，不选择时将使用默认模版
            </span>
            <span v-else>
              请先选择渠道类型，然后选择对应的消息模版
            </span>
          </div>
        </el-form-item>
        
        <!-- 动态配置表单 -->
        <template v-if="channelForm.type">
          <el-divider>{{ getTypeLabel(channelForm.type) }}配置</el-divider>
          <component 
            :is="getConfigComponent(channelForm.type)" 
            v-model="channelConfig"
            :type="channelForm.type"
            @test="handleConfigTest"
          />
        </template>
      </el-form>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button 
            type="primary" 
            @click="handleSubmit"
            :loading="submitting"
          >
            {{ isEditing ? '更新' : '创建' }}
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 发送测试消息对话框 -->
    <el-dialog
      v-model="sendDialogVisible"
      title="发送测试消息"
      width="500px"
    >
      <el-form
        ref="sendFormRef"
        :model="sendForm"
        :rules="sendFormRules"
        label-width="80px"
      >
        <el-form-item label="消息标题" prop="title">
          <el-input v-model="sendForm.title" placeholder="请输入消息标题" />
        </el-form-item>
        <el-form-item label="消息内容" prop="content">
          <el-input 
            v-model="sendForm.content" 
            type="textarea" 
            placeholder="请输入消息内容"
            :rows="4"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="sendDialogVisible = false">取消</el-button>
          <el-button 
            type="primary" 
            @click="handleSendMessage"
            :loading="sending"
          >
            发送
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive, computed } from 'vue'
import { Plus, Search, Refresh, ArrowDown, Edit, Connection } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { channelsAPI, templatesAPI } from '@/api'
import { formatDate } from '@/utils'

// 导入配置组件
import WeChatConfig from '@/components/WeChatConfig.vue'
import DingTalkConfig from '@/components/DingTalkConfig.vue'
import EmailConfig from '@/components/EmailConfig.vue'
import WebhookConfig from '@/components/WebhookConfig.vue'

// 响应式数据
const channels = ref([])
const loading = ref(false)
const channelTypes = ref([])
const availableTemplates = ref([])
const testingChannels = ref([])

// 搜索筛选
const searchQuery = reactive({
  name: '',
  type: '',
  status: ''
})

// 分页
const pagination = reactive({
  page: 1,
  size: 10,
  total: 0
})

// 对话框
const dialogVisible = ref(false)
const dialogTitle = computed(() => isEditing.value ? '编辑渠道' : '添加渠道')
const isEditing = ref(false)
const submitting = ref(false)
const currentChannel = ref(null)

// 计算属性：根据渠道类型筛选模版
const filteredTemplates = computed(() => {
  if (!channelForm.type) {
    return availableTemplates.value
  }
  
  // 优先显示匹配的类型，然后是通用类型
  const matchedTemplates = availableTemplates.value.filter(t => t.type === channelForm.type)
  const otherTemplates = availableTemplates.value.filter(t => t.type !== channelForm.type)
  
  return [...matchedTemplates, ...otherTemplates]
})

// 表单
const channelFormRef = ref()
const channelForm = reactive({
  name: '',
  type: '',
  status: 'active',
  description: '',
  template_id: ''
})

const channelConfig = ref({})

const channelFormRules = {
  name: [
    { required: true, message: '请输入渠道名称', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择渠道类型', trigger: 'change' }
  ]
}

// 发送消息对话框
const sendDialogVisible = ref(false)
const sending = ref(false)
const currentSendChannel = ref(null)

const sendFormRef = ref()
const sendForm = reactive({
  title: '测试消息标题',
  content: '这是一条来自邮件告警平台的测试消息'
})

const sendFormRules = {
  title: [
    { required: true, message: '请输入消息标题', trigger: 'blur' }
  ],
  content: [
    { required: true, message: '请输入消息内容', trigger: 'blur' }
  ]
}

// 方法
const fetchChannels = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      size: pagination.size,
      ...searchQuery
    }
    
    const response = await channelsAPI.list(params)
    channels.value = response.data.channels || []
    pagination.total = response.data.total || 0
  } catch (error) {
    ElMessage.error('获取渠道列表失败')
  } finally {
    loading.value = false
  }
}

const fetchChannelTypes = async () => {
  try {
    const response = await channelsAPI.getTypes()
    channelTypes.value = response.data || []
  } catch (error) {
    console.error('获取渠道类型失败:', error)
    // 设置默认类型
    channelTypes.value = ['wechat', 'dingtalk', 'email', 'webhook']
  }
}

const fetchTemplates = async () => {
  try {
    const response = await templatesAPI.list()
    availableTemplates.value = response.data.list || []
    console.log('获取到的模版列表:', availableTemplates.value)
  } catch (error) {
    console.error('获取模版列表失败:', error)
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchChannels()
}

const handleReset = () => {
  Object.assign(searchQuery, {
    name: '',
    type: '',
    status: ''
  })
  pagination.page = 1
  fetchChannels()
}

const handleSortChange = ({ column, prop, order }) => {
  // 这里可以实现排序逻辑
  console.log('排序变化:', { column, prop, order })
}

const handleSizeChange = (val) => {
  pagination.size = val
  pagination.page = 1
  fetchChannels()
}

const handleCurrentChange = (val) => {
  pagination.page = val
  fetchChannels()
}

const handleAdd = () => {
  isEditing.value = false
  currentChannel.value = null
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (channel) => {
  isEditing.value = true
  currentChannel.value = channel
  
  // 填充表单
  Object.assign(channelForm, {
    name: channel.name,
    type: channel.type,
    status: channel.status,
    description: channel.description,
    template_id: channel.template_id || ''  // 处理null值
  })
  
  // 解析配置
  try {
    channelConfig.value = JSON.parse(channel.config || '{}')
  } catch (error) {
    console.error('解析渠道配置失败:', error)
    channelConfig.value = {}
  }
  
  dialogVisible.value = true
}

// 下拉菜单命令处理
const handleDropdownCommand = async (command, channel) => {
  switch (command) {
    case 'send':
      handleSend(channel)
      break
    case 'duplicate':
      await handleDuplicate(channel)
      break
    case 'delete':
      await handleDelete(channel)
      break
  }
}

// 复制渠道
const handleDuplicate = async (channel) => {
  try {
    const duplicateData = {
      name: `${channel.name} - 副本`,
      type: channel.type,
      status: 'inactive', // 副本默认为停用状态
      description: `${channel.description || ''} (复制自 ${channel.name})`,
      template_id: channel.template_id,
      config: channel.config
    }
    
    await channelsAPI.create(duplicateData)
    ElMessage.success('复制渠道成功')
    fetchChannels()
  } catch (error) {
    console.error('复制渠道失败:', error)
    ElMessage.error('复制渠道失败')
  }
}

const handleDelete = async (channel) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除渠道 "${channel.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await channelsAPI.delete(channel.id)
    ElMessage.success('删除成功')
    fetchChannels()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleTest = async (channel) => {
  testingChannels.value.push(channel.id)
  try {
    await channelsAPI.test(channel.id)
    ElMessage.success('测试成功')
    fetchChannels() // 刷新列表以更新测试状态
  } catch (error) {
    ElMessage.error('测试失败')
  } finally {
    testingChannels.value = testingChannels.value.filter(id => id !== channel.id)
  }
}

const handleSend = (channel) => {
  currentSendChannel.value = channel
  sendDialogVisible.value = true
}

const handleSendMessage = async () => {
  if (!sendFormRef.value) return
  
  try {
    await sendFormRef.value.validate()
    
    sending.value = true
    await channelsAPI.send(currentSendChannel.value.id, sendForm)
    ElMessage.success('发送成功')
    sendDialogVisible.value = false
  } catch (error) {
    if (error !== false) {
      ElMessage.error('发送失败')
    }
  } finally {
    sending.value = false
  }
}

const handleSubmit = async () => {
  if (!channelFormRef.value) return
  
  try {
    await channelFormRef.value.validate()
    
    submitting.value = true
    
    const data = {
      ...channelForm,
      config: JSON.stringify(channelConfig.value),
      template_id: channelForm.template_id || null  // 确保空值转为null
    }
    
    if (isEditing.value) {
      await channelsAPI.update(currentChannel.value.id, data)
      ElMessage.success('更新成功')
    } else {
      await channelsAPI.create(data)
      ElMessage.success('创建成功')
    }
    
    dialogVisible.value = false
    fetchChannels()
  } catch (error) {
    if (error !== false) {
      ElMessage.error(isEditing.value ? '更新失败' : '创建失败')
    }
  } finally {
    submitting.value = false
  }
}

const handleTypeChange = (type) => {
  // 切换类型时重置配置
  channelConfig.value = {}
}

const handleConfigTest = async (config) => {
  // 验证配置是否完整
  if (!channelForm.type) {
    ElMessage.error('请先选择渠道类型')
    return
  }
  
  if (!config || Object.keys(config).length === 0) {
    ElMessage.error('请填写渠道配置信息')
    return
  }

  try {
    console.log('测试配置:', {
      type: channelForm.type,
      config: JSON.stringify(config)
    })
    
    ElMessage.info('正在测试配置，请稍候...')
    
    await channelsAPI.test({
      type: channelForm.type,
      config: JSON.stringify(config)
    })
    ElMessage.success('配置测试成功')
  } catch (error) {
    console.error('配置测试失败:', error)
    
    // 显示详细错误信息
    let errorMessage = '配置测试失败'
    if (error.response?.data?.error) {
      errorMessage += ': ' + error.response.data.error
    } else if (error.message) {
      if (error.code === 'ECONNABORTED') {
        errorMessage = '配置测试超时，请检查网络连接或联系管理员'
      } else {
        errorMessage += ': ' + error.message
      }
    }
    
    ElMessage.error(errorMessage)
  }
}

const handleDialogClose = () => {
  resetForm()
}

const resetForm = () => {
  Object.assign(channelForm, {
    name: '',
    type: '',
    status: 'active',
    description: '',
    template_id: ''
  })
  channelConfig.value = {}
  
  if (channelFormRef.value) {
    channelFormRef.value.clearValidate()
  }
}

// 工具方法
const getTypeLabel = (type) => {
  const labels = {
    wechat: '企业微信',
    dingtalk: '钉钉',
    email: '邮件',
    webhook: 'Webhook'
  }
  return labels[type] || type
}

const getTypeColor = (type) => {
  const colors = {
    wechat: 'success',
    dingtalk: 'primary',
    email: 'warning',
    webhook: 'info'
  }
  return colors[type] || 'info'
}

const getConfigComponent = (type) => {
  const components = {
    wechat: WeChatConfig,
    dingtalk: DingTalkConfig,
    email: EmailConfig,
    webhook: WebhookConfig
  }
  return components[type] || 'div'
}

// 生命周期
onMounted(() => {
  fetchChannelTypes()
  fetchTemplates()
  fetchChannels()
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

.dialog-footer {
  text-align: right;
}

.text-gray-400 {
  color: #9ca3af;
}

.text-danger {
  color: #f56c6c;
}

:deep(.el-dropdown-menu__item.text-danger:hover) {
  background-color: #fef0f0;
  color: #f56c6c;
}
</style> 