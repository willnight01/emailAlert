<template>
  <div class="rule-groups">
    <el-card>
      <template #header>
        <div class="page-header">
          <div>
            <span>告警规则管理</span>
            <el-tag type="info" size="small" style="margin-left: 8px;">增强版多维度规则引擎</el-tag>
          </div>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            创建规则组
          </el-button>
        </div>
      </template>

      <!-- 搜索和筛选区域 -->
      <div class="filter-section">
        <el-row :gutter="20">
          <el-col :span="8">
            <el-input
              v-model="searchParams.name"
              placeholder="搜索规则组名称或描述"
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
              v-model="searchParams.status"
              placeholder="状态筛选"
              @change="handleSearch"
              clearable
            >
              <el-option label="全部状态" value="" />
              <el-option label="激活" value="active" />
              <el-option label="停用" value="inactive" />
            </el-select>
          </el-col>
          <el-col :span="6">
            <el-select
              v-model="searchParams.mailbox_id"
              placeholder="邮箱筛选"
              @change="handleSearch"
              clearable
            >
              <el-option label="全部邮箱" value="" />
              <el-option
                v-for="option in mailboxOptions"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              />
            </el-select>
          </el-col>
          <el-col :span="4">
            <el-button @click="resetSearch" :loading="loading">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </el-col>
        </el-row>
      </div>
      
      <!-- 规则组列表表格 -->
      <el-table 
        :data="ruleGroups" 
        style="width: 100%" 
        v-loading="loading"
        :default-sort="{ prop: 'created_at', order: 'descending' }"
      >
        <el-table-column prop="name" label="规则组名称" min-width="260">
          <template #default="scope">
            <div class="name-cell">
              <span class="name">{{ scope.row.name }}</span>
              <div class="meta-info">
                <span class="mailbox-info">{{ scope.row.mailbox?.name }}</span>
                <el-divider direction="vertical" />
                <span class="condition-count">{{ scope.row.conditions?.length || 0 }}个条件</span>
                <el-divider direction="vertical" />
                <el-tag size="small" :type="scope.row.logic === 'and' ? 'success' : 'warning'">
                  {{ scope.row.logic?.toUpperCase() }}
                </el-tag>
              </div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="description" label="描述" min-width="280" show-overflow-tooltip>
          <template #default="scope">
            <span v-if="scope.row.description" class="description-text">{{ scope.row.description }}</span>
            <span v-else class="no-description">暂无描述</span>
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
        
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="scope">
            <div class="action-buttons">
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
                @click="handleToggleStatus(scope.row)"
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
                      测试规则
                    </el-dropdown-item>
                    <el-dropdown-item 
                      command="duplicate"
                    >
                      复制规则
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
            </div>
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

    <!-- 规则组表单对话框 -->
    <el-dialog
      v-model="formVisible"
      :title="isEdit ? '编辑规则组' : '创建规则组'"
      width="900px"
      :close-on-click-modal="false"
      @close="resetForm"
    >
      <div class="rule-group-form">
        <!-- 基本信息 -->
        <el-card class="form-section">
          <template #header>
            <span>基本信息</span>
          </template>
          <el-form
            ref="basicFormRef"
            :model="formData"
            :rules="basicFormRules"
            label-width="120px"
          >
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item label="规则组名称" prop="name">
                  <el-input v-model="formData.name" placeholder="请输入规则组名称" style="font-size: 14px;" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="邮箱源" prop="mailbox_id">
                  <el-select v-model="formData.mailbox_id" placeholder="请选择邮箱" style="width: 100%; font-size: 14px;">
                    <el-option
                      v-for="option in mailboxOptions"
                      :key="option.value"
                      :label="option.label"
                      :value="option.value"
                    />
                  </el-select>
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="16">
              <el-col :span="8">
                <el-form-item label="条件逻辑" prop="logic">
                  <el-select v-model="formData.logic" placeholder="请选择" style="width: 100%; font-size: 14px;">
                    <el-option label="AND (全部满足)" value="and" />
                    <el-option label="OR (任一满足)" value="or" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="8">
                <el-form-item label="优先级" prop="priority">
                  <el-input-number
                    v-model="formData.priority"
                    :min="1"
                    :max="10"
                    style="width: 100%; font-size: 14px;"
                  />
                </el-form-item>
              </el-col>
              <el-col :span="8">
                <el-form-item label="状态">
                  <el-radio-group v-model="formData.status" style="font-size: 14px;">
                    <el-radio label="active">激活</el-radio>
                    <el-radio label="inactive">停用</el-radio>
                  </el-radio-group>
                </el-form-item>
              </el-col>
            </el-row>
            <el-form-item label="通知渠道">
              <el-select
                v-model="formData.channel_ids"
                multiple
                placeholder="请选择通知渠道"
                style="width: 100%; font-size: 14px;"
                collapse-tags
                collapse-tags-tooltip
              >
                <el-option
                  v-for="option in channelOptions"
                  :key="option.value"
                  :label="option.label"
                  :value="option.value"
                >
                  <div class="channel-option">
                    <span>{{ option.label }}</span>
                    <el-tag :type="getChannelTypeColor(option.type)" size="small">{{ option.type }}</el-tag>
                  </div>
                </el-option>
              </el-select>
              <div class="input-tip">
                提示：可选择多个通知渠道，告警时会同时发送到所有选中的渠道
              </div>
            </el-form-item>
            <el-form-item label="描述">
              <el-input
                v-model="formData.description"
                type="textarea"
                :rows="3"
                placeholder="请输入规则组描述"
                style="font-size: 14px;"
              />
            </el-form-item>
          </el-form>
        </el-card>

        <!-- 匹配条件 -->
        <el-card class="form-section">
          <template #header>
            <div class="condition-header">
              <span>匹配条件配置</span>
              <el-button type="primary" size="small" @click="addCondition">
                <el-icon><Plus /></el-icon>
                添加条件
              </el-button>
            </div>
          </template>
          
          <div v-if="conditions.length === 0" class="empty-conditions">
            <el-empty description="暂无匹配条件，请添加至少一个条件">
              <el-button type="primary" @click="addCondition">添加第一个条件</el-button>
            </el-empty>
          </div>

          <div v-else class="conditions-list">
            <div
              v-for="(condition, index) in conditions"
              :key="index"
              class="condition-item"
            >
              <div class="condition-header-line">
                <span class="condition-number">条件 {{ index + 1 }}</span>
                <el-button type="danger" size="small" text @click="removeCondition(index)">
                  <el-icon><Delete /></el-icon>
                  删除
                </el-button>
              </div>
              
              <el-form
                :model="condition"
                :rules="conditionFormRules"
                label-width="120px"
                size="small"
              >
                <el-row :gutter="16">
                  <el-col :span="6">
                    <el-form-item label="匹配字段" prop="field_type">
                      <el-select v-model="condition.field_type" placeholder="请选择" style="width: 100%; font-size: 14px;">
                        <el-option
                          v-for="option in fieldTypeOptions"
                          :key="option.value"
                          :label="option.label"
                          :value="option.value"
                        />
                      </el-select>
                    </el-form-item>
                  </el-col>
                  <el-col :span="6">
                    <el-form-item label="匹配类型" prop="match_type">
                      <el-select v-model="condition.match_type" placeholder="请选择" style="width: 100%; font-size: 14px;">
                        <el-option
                          v-for="option in matchTypeOptions"
                          :key="option.value"
                          :label="option.label"
                          :value="option.value"
                        />
                      </el-select>
                    </el-form-item>
                  </el-col>
                  <el-col :span="6">
                    <el-form-item label="关键词逻辑" prop="keyword_logic">
                      <el-select v-model="condition.keyword_logic" placeholder="请选择" style="width: 100%; font-size: 14px;">
                        <el-option label="AND (全部)" value="and" />
                        <el-option label="OR (任一)" value="or" />
                      </el-select>
                    </el-form-item>
                  </el-col>
                  <el-col :span="6">
                    <el-form-item label="优先级" prop="priority">
                      <el-input-number
                        v-model="condition.priority"
                        :min="1"
                        :max="10"
                        style="width: 100%; font-size: 14px;"
                      />
                    </el-form-item>
                  </el-col>
                </el-row>
                <el-form-item label="关键词" prop="keywords">
                  <el-input
                    v-model="condition.keywords"
                    type="textarea"
                    :rows="3"
                    placeholder="请输入关键词，多个关键词用逗号分隔，如：错误,异常,失败"
                    style="font-size: 14px;"
                  />
                  <div class="input-tip">
                    提示：多个关键词用逗号分隔，匹配逻辑由上方的"关键词逻辑"控制
                  </div>
                </el-form-item>
                <el-form-item label="状态">
                  <el-radio-group v-model="condition.status" style="font-size: 14px;">
                    <el-radio label="active">启用</el-radio>
                    <el-radio label="inactive">停用</el-radio>
                  </el-radio-group>
                </el-form-item>
                <el-form-item label="描述">
                  <el-input
                    v-model="condition.description"
                    placeholder="请输入条件描述（可选）"
                    style="font-size: 14px;"
                  />
                </el-form-item>
              </el-form>
            </div>
          </div>
        </el-card>
      </div>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="resetForm">取消</el-button>
          <el-button type="primary" @click="handleSave" :loading="saving">
            {{ isEdit ? '更新' : '创建' }}
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 测试规则组对话框 -->
    <el-dialog
      v-model="testVisible"
      title="规则组测试"
      width="60%"
    >
      <el-form :model="testData" label-width="100px">
        <el-form-item label="邮件主题">
          <el-input v-model="testData.subject" placeholder="请输入测试邮件主题" />
        </el-form-item>
        <el-form-item label="发件人">
          <el-input v-model="testData.sender" placeholder="请输入发件人邮箱" />
        </el-form-item>
        <el-form-item label="收件人">
          <el-input v-model="testData.to" placeholder="请输入收件人邮箱" />
        </el-form-item>
        <el-form-item label="抄送人">
          <el-input v-model="testData.cc" placeholder="请输入抄送人邮箱（可选）" />
        </el-form-item>
        <el-form-item label="邮件内容">
          <el-input
            v-model="testData.content"
            type="textarea"
            :rows="4"
            placeholder="请输入测试邮件内容"
          />
        </el-form-item>
        <el-form-item label="附件名称">
          <el-input v-model="testData.attachment_name" placeholder="请输入附件名称（可选）" />
        </el-form-item>
      </el-form>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="testVisible = false">取消</el-button>
          <el-button type="primary" @click="runTest" :loading="testing">执行测试</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, List, Delete, Edit, Switch, Refresh, ArrowDown } from '@element-plus/icons-vue'
import { ruleGroupsAPI, channelsAPI } from '@/api'

// 响应式数据
const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const formVisible = ref(false)
const testVisible = ref(false)
const isEdit = ref(false)

const ruleGroups = ref([])
const mailboxOptions = ref([])
const matchTypeOptions = ref([])
const fieldTypeOptions = ref([])
const channelOptions = ref([])

// 分页数据
const pagination = reactive({
  page: 1,
  size: 10,
  total: 0
})

// 搜索参数
const searchParams = reactive({
  name: '',
  mailbox_id: null,
  logic: '',
  status: ''
})

// 表单数据
const formData = reactive({
  name: '',
  mailbox_id: null,
  logic: 'and',
  priority: 1,
  status: 'active',
  description: '',
  channel_ids: [] // 新增：选中的通知渠道ID列表
})

// 条件数据
const conditions = ref([])

// 测试数据
const testData = reactive({
  subject: '',
  sender: '',
  to: '',
  cc: '',
  content: '',
  attachment_name: ''
})

// 表单验证规则
const basicFormRules = {
  name: [
    { required: true, message: '请输入规则组名称', trigger: 'blur' }
  ],
  mailbox_id: [
    { required: true, message: '请选择邮箱', trigger: 'change' }
  ],
  logic: [
    { required: true, message: '请选择条件逻辑', trigger: 'change' }
  ]
}

const conditionFormRules = {
  field_type: [
    { required: true, message: '请选择匹配字段', trigger: 'change' }
  ],
  match_type: [
    { required: true, message: '请选择匹配类型', trigger: 'change' }
  ],
  keywords: [
    { required: true, message: '请输入关键词', trigger: 'blur' }
  ]
}

// 引用
const basicFormRef = ref()

// 计算属性
const getPriorityType = (priority) => {
  if (priority >= 8) return 'danger'
  if (priority >= 5) return 'warning'
  return 'success'
}

const getMatchTypeText = (type) => {
  const map = {
    equals: '完全匹配',
    contains: '包含匹配',
    startsWith: '前缀匹配',
    endsWith: '后缀匹配',
    regex: '正则表达式',
    notContains: '不包含'
  }
  return map[type] || type
}

const getFieldTypeText = (type) => {
  const map = {
    subject: '邮件主题',
    from: '发件人',
    to: '收件人',
    cc: '抄送人',
    body: '邮件正文',
    attachment_name: '附件名称'
  }
  return map[type] || type
}

const getChannelTypeColor = (type) => {
  const map = {
    wechat: 'success',
    dingtalk: 'primary',
    email: 'warning',
    webhook: 'info'
  }
  return map[type] || 'info'
}

// 方法
const loadRuleGroups = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      size: pagination.size,
      ...searchParams
    }
    
    const response = await ruleGroupsAPI.list(params)
    ruleGroups.value = response.data.items || []
    pagination.total = response.data.total || 0
  } catch (error) {
    console.error('获取规则组列表失败:', error)
    ElMessage.error('获取规则组列表失败')
  } finally {
    loading.value = false
  }
}

const loadOptions = async () => {
  try {
    // 获取邮箱选项
    const mailboxResponse = await ruleGroupsAPI.getMailboxOptions()
    mailboxOptions.value = mailboxResponse.data || []

    // 获取匹配类型选项
    const matchTypeResponse = await ruleGroupsAPI.getMatchTypeOptions()
    matchTypeOptions.value = matchTypeResponse.data || []

    // 获取字段类型选项
    const fieldTypeResponse = await ruleGroupsAPI.getFieldTypeOptions()  
    fieldTypeOptions.value = fieldTypeResponse.data || []

    // 获取通知渠道选项
    const channelResponse = await ruleGroupsAPI.getChannelOptions()
    channelOptions.value = channelResponse.data || []
  } catch (error) {
    console.error('获取选项数据失败:', error)
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadRuleGroups()
}

const resetSearch = () => {
  Object.assign(searchParams, {
    name: '',
    mailbox_id: null,
    logic: '',
    status: ''
  })
  handleSearch()
}

const handleAdd = () => {
  isEdit.value = false
  formVisible.value = true
}

const handleEdit = async (row) => {
  isEdit.value = true
  try {
    // 获取完整的规则组数据（包含条件）
    const response = await ruleGroupsAPI.getWithConditions(row.id)
    const ruleGroupData = response.data
    
    // 填充基本表单数据
    Object.assign(formData, {
      id: ruleGroupData.id,
      name: ruleGroupData.name,
      mailbox_id: ruleGroupData.mailbox_id,
      logic: ruleGroupData.logic,
      priority: ruleGroupData.priority,
      status: ruleGroupData.status,
      description: ruleGroupData.description || '',
      channel_ids: (ruleGroupData.channels || []).map(channel => channel.id)
    })
    
    // 填充条件数据
    conditions.value = (ruleGroupData.conditions || []).map(condition => ({
      field_type: condition.field_type,
      match_type: condition.match_type,
      keywords: condition.keywords,
      keyword_logic: condition.keyword_logic || 'or',
      priority: condition.priority || 1,
      status: condition.status || 'active',
      description: condition.description || ''
    }))
    
    formVisible.value = true
  } catch (error) {
    console.error('获取规则组详情失败:', error)
    ElMessage.error('获取规则组详情失败')
  }
}

// 下拉菜单命令处理
const handleDropdownCommand = async (command, row) => {
  switch (command) {
    case 'test':
      handleTest(row)
      break
    case 'duplicate':
      await handleDuplicate(row)
      break
    case 'delete':
      await handleDelete(row)
      break
  }
}

// 复制规则组
const handleDuplicate = async (row) => {
  try {
    const duplicateData = {
      rule_group: {
        name: `${row.name} - 副本`,
        mailbox_id: row.mailbox_id,
        logic: row.logic,
        priority: row.priority,
        status: 'inactive', // 副本默认为停用状态
        description: `${row.description || ''} (复制自 ${row.name})`,
        channel_ids: row.channels?.map(c => c.id) || []
      },
      conditions: row.conditions || []
    }
    
    await ruleGroupsAPI.createWithConditions(duplicateData)
    ElMessage.success('复制规则组成功')
    loadRuleGroups()
  } catch (error) {
    console.error('复制规则组失败:', error)
    ElMessage.error('复制规则组失败')
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除规则组"${row.name}"吗？这将同时删除其所有匹配条件。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await ruleGroupsAPI.delete(row.id)
    ElMessage.success('删除规则组成功')
    loadRuleGroups()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除规则组失败:', error)
      ElMessage.error('删除规则组失败')
    }
  }
}

const handleToggleStatus = async (row) => {
  try {
    const newStatus = row.status === 'active' ? 'inactive' : 'active'
    
    // 使用专门的状态更新API
    await ruleGroupsAPI.updateStatus(row.id, { status: newStatus })
    
    row.status = newStatus
    ElMessage.success(`规则组已${newStatus === 'active' ? '启用' : '停用'}`)
  } catch (error) {
    console.error('更新规则组状态失败:', error)
    ElMessage.error('状态更新失败')
  }
}

const handleTest = (row) => {
  // 填充测试数据的默认值
  Object.assign(testData, {
    subject: 'TEST: 系统告警测试',
    sender: 'test@example.com',
    to: 'admin@example.com',
    cc: '',
    content: '这是一条测试告警邮件，用于验证规则组匹配效果。',
    attachment_name: ''
  })
  
  testVisible.value = true
}

const runTest = async () => {
  testing.value = true
  try {
    // TODO: 实现规则组测试逻辑
    ElMessage.success('规则组测试功能开发中...')
    testVisible.value = false
  } catch (error) {
    console.error('测试规则组失败:', error)
    ElMessage.error('测试规则组失败')
  } finally {
    testing.value = false
  }
}

const addCondition = () => {
  conditions.value.push({
    field_type: '',
    match_type: '',
    keywords: '',
    keyword_logic: 'or',
    priority: 1,
    status: 'active',
    description: ''
  })
}

const removeCondition = (index) => {
  conditions.value.splice(index, 1)
}

const handleSave = async () => {
  if (!basicFormRef.value) return

  try {
    await basicFormRef.value.validate()
  } catch (error) {
    ElMessage.error('请检查基本信息填写')
    return
  }

  if (conditions.value.length === 0) {
    ElMessage.error('请至少添加一个匹配条件')
    return
  }

  // 验证所有条件
  for (let i = 0; i < conditions.value.length; i++) {
    const condition = conditions.value[i]
    if (!condition.field_type || !condition.match_type || !condition.keywords) {
      ElMessage.error(`请完善条件 ${i + 1} 的配置`)
      return
    }
  }

  saving.value = true
  try {
    const requestData = {
      rule_group: { ...formData },
      conditions: conditions.value,
      channel_ids: formData.channel_ids || []
    }

    if (isEdit.value) {
      await ruleGroupsAPI.updateWithConditions(formData.id, requestData)
      ElMessage.success('更新规则组成功')
    } else {
      await ruleGroupsAPI.createWithConditions(requestData)
      ElMessage.success('创建规则组成功')
    }

    formVisible.value = false
    loadRuleGroups()
  } catch (error) {
    console.error('保存规则组失败:', error)
    ElMessage.error('保存规则组失败')
  } finally {
    saving.value = false
  }
}

const resetForm = () => {
  formVisible.value = false
  
  // 重置基本表单数据
  Object.assign(formData, {
    id: null,
    name: '',
    mailbox_id: null,
    logic: 'and',
    priority: 1,
    status: 'active',
    description: '',
    channel_ids: []
  })
  
  // 重置条件数据
  conditions.value = []
  
  // 清除表单验证
  if (basicFormRef.value) {
    basicFormRef.value.clearValidate()
  }
}

const handleSizeChange = (newSize) => {
  pagination.size = newSize
  pagination.page = 1
  loadRuleGroups()
}

const handleCurrentChange = (newPage) => {
  pagination.page = newPage
  loadRuleGroups()
}

// 初始化
onMounted(() => {
  loadOptions()
  loadRuleGroups()
})
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
  margin-bottom: 4px;
  font-size: 14px;
  line-height: 1.4;
}

.meta-info {
  display: flex;
  align-items: center;
  font-size: 13px;
  color: var(--el-text-color-secondary);
  flex-wrap: wrap;
  gap: 2px;
}

.mailbox-info {
  color: var(--el-text-color-regular);
  font-size: 13px;
}

.condition-count {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.description-text {
  color: var(--el-text-color-regular);
  line-height: 1.5;
  font-size: 14px;
  word-break: break-word;
}

.no-description {
  color: var(--el-text-color-placeholder);
  font-style: italic;
  font-size: 13px;
}

.pagination-wrapper {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.rule-group-form {
  max-height: 70vh;
  overflow-y: auto;
}

.form-section {
  margin-bottom: 20px;
}

.condition-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.empty-conditions {
  text-align: center;
  padding: 40px 0;
}

.conditions-list {
  space-y: 16px;
}

.condition-item {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 16px;
  background-color: #fafafa;
  transition: all 0.3s ease;
}

.condition-item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.1);
}

.condition-header-line {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.condition-number {
  font-weight: 600;
  color: #409eff;
  font-size: 15px;
}

.condition-keywords {
  word-break: break-all;
  max-width: 200px;
  display: inline-block;
}

.input-tip {
  font-size: 13px;
  color: #909399;
  margin-top: 6px;
  line-height: 1.4;
}

.channel-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  padding: 2px 0;
}

.channel-option span {
  flex: 1;
  font-size: 14px;
}

.text-gray {
  color: #909399;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

:deep(.el-table .el-table__cell) {
  padding: 12px 8px;
  font-size: 14px;
}

:deep(.el-table th.el-table__cell) {
  background-color: #f8f9fa;
  font-weight: 500;
  font-size: 14px;
}

:deep(.el-input-number .el-input__inner) {
  text-align: left;
}

.action-buttons {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.action-buttons .el-button {
  font-size: 13px;
}

:deep(.el-form-item__label) {
  font-size: 14px;
  font-weight: 500;
}

:deep(.el-input__inner) {
  font-size: 14px;
}

:deep(.el-textarea__inner) {
  font-size: 14px;
  line-height: 1.5;
}

.text-danger {
  color: #f56c6c;
}

:deep(.el-dropdown-menu__item.text-danger:hover) {
  background-color: #fef0f0;
  color: #f56c6c;
}
</style> 