<template>
  <div class="templates-page">
    <el-card>
      <template #header>
        <div class="page-header">
          <div class="header-left">
            <h2>消息模版管理</h2>
            <p class="page-description">管理告警消息的输出模版，支持多种格式和变量替换</p>
          </div>
          <div class="header-right">
            <el-button type="primary" @click="handleAdd">
              <el-icon><Plus /></el-icon>
              添加模版
            </el-button>
          </div>
        </div>
      </template>
      
      <!-- 搜索和筛选区域 -->
      <div class="filter-section">
        <el-row :gutter="16">
          <el-col :span="8">
            <el-input
              v-model="searchForm.name"
              placeholder="搜索模版名称..."
              clearable
              @input="handleSearch"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
          </el-col>
          <el-col :span="6">
            <el-select
              v-model="searchForm.type"
              placeholder="模版类型"
              clearable
              @change="handleSearch"
            >
              <el-option label="全部类型" value="" />
              <el-option label="邮件模版" value="email" />
              <el-option label="钉钉模版" value="dingtalk" />
              <el-option label="企业微信" value="wechat" />
              <el-option label="Markdown" value="markdown" />
            </el-select>
          </el-col>
          <el-col :span="6">
            <el-select
              v-model="searchForm.status"
              placeholder="状态"
              clearable
              @change="handleSearch"
            >
              <el-option label="全部状态" value="" />
              <el-option label="活跃" value="active" />
              <el-option label="停用" value="inactive" />
            </el-select>
          </el-col>
          <el-col :span="4">
            <el-button @click="handleReset">重置</el-button>
          </el-col>
        </el-row>
      </div>
      
      <!-- 模版列表 -->
      <div class="table-section">
        <el-table 
          :data="templateList" 
          v-loading="loading"
          stripe
          style="width: 100%"
        >
          <el-table-column prop="name" label="模版名称" min-width="150">
            <template #default="scope">
              <div class="template-name">
                <span>{{ scope.row.name }}</span>
                <el-tag v-if="scope.row.is_default" type="success" size="small">默认</el-tag>
              </div>
            </template>
          </el-table-column>
          
          <el-table-column prop="type" label="模版类型" width="120">
            <template #default="scope">
              <el-tag :type="getTypeTagType(scope.row.type)" size="small">
                {{ getTypeLabel(scope.row.type) }}
              </el-tag>
            </template>
          </el-table-column>
          
          <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
          
          <el-table-column prop="status" label="状态" width="80">
            <template #default="scope">
              <el-tag :type="scope.row.status === 'active' ? 'success' : 'info'" size="small">
                {{ scope.row.status === 'active' ? '活跃' : '停用' }}
              </el-tag>
            </template>
          </el-table-column>
          
          <el-table-column prop="updated_at" label="更新时间" width="180">
            <template #default="scope">
              {{ formatTime(scope.row.updated_at) }}
            </template>
          </el-table-column>
          
          <el-table-column label="操作" width="250" fixed="right">
            <template #default="scope">
              <el-button-group>
                <el-button 
                  type="primary" 
                  size="small" 
                  @click="handlePreview(scope.row)"
                  :icon="View"
                >
                  预览
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
                        command="setDefault"
                        :disabled="scope.row.is_default"
                      >
                        设为默认
                      </el-dropdown-item>
                      <el-dropdown-item 
                        command="duplicate"
                      >
                        复制模版
                      </el-dropdown-item>
                      <el-dropdown-item 
                        command="toggleStatus"
                        :class="scope.row.status === 'active' ? 'text-warning' : 'text-success'"
                      >
                        {{ scope.row.status === 'active' ? '停用' : '启用' }}
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
        <div class="pagination-section">
          <el-pagination
            v-model:current-page="pagination.page"
            v-model:page-size="pagination.size"
            :page-sizes="[10, 20, 50, 100]"
            :total="pagination.total"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="loadTemplateList"
            @current-change="loadTemplateList"
          />
        </div>
      </div>
    </el-card>
    
    <!-- 模版表单对话框 -->
    <TemplateForm
      v-model="formVisible"
      :template-data="currentTemplate"
      @success="handleFormSuccess"
    />
    
    <!-- 模版预览对话框 -->
    <TemplatePreview
      v-model="previewVisible"
      :template-data="currentTemplate"
      @edit="handleEditFromPreview"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, View, Edit, ArrowDown } from '@element-plus/icons-vue'
import { templatesAPI } from '@/api'
import TemplateForm from '@/components/TemplateForm.vue'
import TemplatePreview from '@/components/TemplatePreview.vue'

// 响应式数据
const loading = ref(false)
const formVisible = ref(false)
const previewVisible = ref(false)
const templateList = ref([])
const currentTemplate = ref(null)

// 搜索表单
const searchForm = reactive({
  name: '',
  type: '',
  status: ''
})

// 分页信息
const pagination = reactive({
  page: 1,
  size: 20,
  total: 0
})

// 获取模版类型标签类型
const getTypeTagType = (type) => {
  const typeMap = {
    email: 'primary',
    dingtalk: 'success',
    wechat: 'warning',
    markdown: 'info'
  }
  return typeMap[type] || 'info'
}

// 获取模版类型标签
const getTypeLabel = (type) => {
  const typeMap = {
    email: '邮件',
    dingtalk: '钉钉',
    wechat: '微信',
    markdown: 'MD'
  }
  return typeMap[type] || type
}

// 格式化时间
const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

// 加载模版列表
const loadTemplateList = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      size: pagination.size,
      ...searchForm
    }
    
    // 过滤空值
    Object.keys(params).forEach(key => {
      if (params[key] === '' || params[key] === null || params[key] === undefined) {
        delete params[key]
      }
    })
    
    const response = await templatesAPI.list(params)
    templateList.value = response.data.list || []
    pagination.total = response.data.total || 0
  } catch (error) {
    console.error('加载模版列表失败:', error)
    ElMessage.error('加载模版列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索处理
const handleSearch = () => {
  pagination.page = 1
  loadTemplateList()
}

// 重置搜索
const handleReset = () => {
  Object.assign(searchForm, {
    name: '',
    type: '',
    status: ''
  })
  handleSearch()
}

// 添加模版
const handleAdd = () => {
  currentTemplate.value = null
  formVisible.value = true
}

// 编辑模版
const handleEdit = (template) => {
  currentTemplate.value = { ...template }
  formVisible.value = true
}

// 预览模版
const handlePreview = (template) => {
  currentTemplate.value = template
  previewVisible.value = true
}

// 下拉菜单命令处理
const handleDropdownCommand = async (command, template) => {
  switch (command) {
    case 'setDefault':
      await handleSetDefault(template)
      break
    case 'duplicate':
      await handleDuplicate(template)
      break
    case 'toggleStatus':
      await handleToggleStatus(template)
      break
    case 'delete':
      await handleDelete(template)
      break
  }
}

// 设为默认模版
const handleSetDefault = async (template) => {
  try {
    await templatesAPI.setDefault(template.id)
    ElMessage.success('设置默认模版成功')
    loadTemplateList()
  } catch (error) {
    console.error('设置默认模版失败:', error)
    ElMessage.error('设置默认模版失败')
  }
}

// 复制模版
const handleDuplicate = async (template) => {
  try {
    const duplicateData = {
      name: `${template.name} - 副本`,
      type: template.type,
      subject: template.subject,
      content: template.content,
      description: `${template.description || ''} (复制自 ${template.name})`,
      is_default: false,
      status: 'active'
    }
    
    await templatesAPI.create(duplicateData)
    ElMessage.success('复制模版成功')
    loadTemplateList()
  } catch (error) {
    console.error('复制模版失败:', error)
    ElMessage.error('复制模版失败')
  }
}

// 切换状态
const handleToggleStatus = async (template) => {
  try {
    const newStatus = template.status === 'active' ? 'inactive' : 'active'
    await templatesAPI.update(template.id, { status: newStatus })
    ElMessage.success(`模版已${newStatus === 'active' ? '启用' : '停用'}`)
    loadTemplateList()
  } catch (error) {
    console.error('切换状态失败:', error)
    ElMessage.error('切换状态失败')
  }
}

// 删除模版
const handleDelete = async (template) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除模版"${template.name}"吗？此操作不可恢复。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await templatesAPI.delete(template.id)
    ElMessage.success('删除模版成功')
    loadTemplateList()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除模版失败:', error)
      ElMessage.error('删除模版失败')
    }
  }
}

// 表单成功处理
const handleFormSuccess = () => {
  loadTemplateList()
}

// 从预览中编辑
const handleEditFromPreview = (template) => {
  currentTemplate.value = { ...template }
  previewVisible.value = false
  formVisible.value = true
}

// 组件挂载时加载数据
onMounted(() => {
  loadTemplateList()
})
</script>

<style scoped>
.templates-page {
  margin: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.header-left h2 {
  margin: 0 0 8px 0;
  color: #2c3e50;
  font-weight: 600;
}

.page-description {
  margin: 0;
  color: #6c757d;
  font-size: 14px;
}

.filter-section {
  margin-bottom: 20px;
  padding: 20px;
  background-color: #f8f9fa;
  border-radius: 6px;
}

.table-section {
  margin-top: 20px;
}

.template-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pagination-section {
  margin-top: 20px;
  text-align: right;
}

.text-success {
  color: #67c23a;
}

.text-warning {
  color: #e6a23c;
}

.text-danger {
  color: #f56c6c;
}

:deep(.el-dropdown-menu__item.text-success:hover) {
  background-color: #f0f9ff;
  color: #67c23a;
}

:deep(.el-dropdown-menu__item.text-warning:hover) {
  background-color: #fdf6ec;
  color: #e6a23c;
}

:deep(.el-dropdown-menu__item.text-danger:hover) {
  background-color: #fef0f0;
  color: #f56c6c;
}
</style> 