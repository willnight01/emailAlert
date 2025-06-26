<template>
  <el-dialog
    v-model="visible"
    :title="isEdit ? '编辑模版' : '添加模版'"
    width="90%"
    :close-on-click-modal="false"
    @closed="handleClosed"
  >
    <div class="template-form">
      <el-row :gutter="20">
        <!-- 左侧表单 -->
        <el-col :span="12">
          <el-form 
            ref="formRef" 
            :model="form" 
            :rules="rules" 
            label-width="80px"
          >
            <el-form-item label="模版名称" prop="name">
              <el-input 
                v-model="form.name" 
                placeholder="请输入模版名称"
                clearable
              />
            </el-form-item>
            
            <el-form-item label="模版类型" prop="type">
              <el-select 
                v-model="form.type" 
                placeholder="请选择模版类型"
                style="width: 100%"
                @change="handleTypeChange"
              >
                <el-option label="邮件模版" value="email" />
                <el-option label="钉钉模版" value="dingtalk" />
                <el-option label="企业微信" value="wechat" />
                <el-option label="Markdown" value="markdown" />
              </el-select>
            </el-form-item>
            
            <el-form-item 
              v-if="form.type === 'email'" 
              label="邮件主题" 
              prop="subject"
            >
              <el-input 
                v-model="form.subject" 
                placeholder="支持变量，如：【告警】{{.System.AppName}} - {{.Email.Subject}}"
                clearable
              />
            </el-form-item>
            
            <el-form-item label="模版内容" prop="content">
              <div class="code-editor">
                <div class="editor-toolbar">
                  <el-button-group>
                    <el-button 
                      size="small" 
                      @click="insertVariable"
                      :icon="Plus"
                    >
                      插入变量
                    </el-button>
                    <el-button 
                      size="small" 
                      @click="showVariableHelp = true"
                      :icon="QuestionFilled"
                    >
                      变量帮助
                    </el-button>
                    <el-button 
                      size="small" 
                      @click="formatContent"
                      :icon="Document"
                    >
                      格式化
                    </el-button>
                  </el-button-group>
                </div>
                <el-input
                  v-model="form.content"
                  type="textarea"
                  :rows="12"
                  placeholder="请输入模版内容，支持变量替换，如：{{.Email.Subject}}"
                  style="font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;"
                />
              </div>
            </el-form-item>
            
            <el-form-item label="是否默认">
              <el-switch 
                v-model="form.is_default"
                active-text="设为默认模版"
                inactive-text="普通模版"
              />
            </el-form-item>
            
            <el-form-item label="描述">
              <el-input 
                v-model="form.description" 
                type="textarea"
                :rows="3"
                placeholder="请输入模版描述"
              />
            </el-form-item>
          </el-form>
        </el-col>
        
        <!-- 右侧预览 -->
        <el-col :span="12">
          <div class="preview-section">
            <div class="preview-header">
              <span>实时预览</span>
              <el-button 
                size="small" 
                type="primary" 
                @click="refreshPreview"
                :loading="previewLoading"
              >
                刷新预览
              </el-button>
            </div>
            
            <el-tabs v-model="activePreviewTab" class="preview-tabs">
              <el-tab-pane 
                v-if="form.type === 'email' && form.subject" 
                label="邮件主题" 
                name="subject"
              >
                <div class="preview-content subject-preview">
                  {{ previewData.subject || '主题预览...' }}
                </div>
              </el-tab-pane>
              
              <el-tab-pane label="内容预览" name="content">
                <div 
                  class="preview-content"
                  :class="{
                    'html-preview': form.type === 'email',
                    'markdown-preview': form.type === 'dingtalk' || form.type === 'markdown',
                    'text-preview': form.type === 'wechat'
                  }"
                  v-html="renderPreviewContent"
                >
                </div>
              </el-tab-pane>
              
              <el-tab-pane label="使用变量" name="variables">
                <div class="used-variables">
                  <el-tag 
                    v-for="variable in previewData.used_vars" 
                    :key="variable"
                    size="small"
                    class="variable-tag"
                  >
                    {{ variable }}
                  </el-tag>
                  <div v-if="!previewData.used_vars || previewData.used_vars.length === 0" class="no-variables">
                    暂未使用任何变量
                  </div>
                </div>
              </el-tab-pane>
            </el-tabs>
          </div>
        </el-col>
      </el-row>
    </div>
    
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="visible = false">取消</el-button>
        <el-button 
          type="primary" 
          @click="handleSubmit"
          :loading="submitLoading"
        >
          {{ isEdit ? '更新' : '创建' }}
        </el-button>
      </div>
    </template>
    
    <!-- 变量帮助对话框 -->
    <VariableHelp v-model="showVariableHelp" @insert="handleInsertVariable" />
    
    <!-- 变量选择器 -->
    <VariableSelector 
      v-model="showVariableSelector" 
      @select="handleSelectVariable"
    />
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, QuestionFilled, Document } from '@element-plus/icons-vue'
import { templatesAPI } from '@/api'
import VariableHelp from './VariableHelp.vue'
import VariableSelector from './VariableSelector.vue'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  templateData: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['update:modelValue', 'success'])

const visible = ref(false)
const formRef = ref()
const submitLoading = ref(false)
const previewLoading = ref(false)
const showVariableHelp = ref(false)
const showVariableSelector = ref(false)
const activePreviewTab = ref('content')

const isEdit = computed(() => !!props.templateData?.id)

const form = reactive({
  name: '',
  type: 'email',
  subject: '',
  content: '',
  is_default: false,
  description: ''
})

const rules = {
  name: [
    { required: true, message: '请输入模版名称', trigger: 'blur' },
    { min: 2, max: 50, message: '模版名称长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择模版类型', trigger: 'change' }
  ],
  content: [
    { required: true, message: '请输入模版内容', trigger: 'blur' },
    { min: 10, message: '模版内容至少需要 10 个字符', trigger: 'blur' }
  ]
}

const previewData = ref({
  subject: '',
  content: '',
  used_vars: []
})

// 监听modelValue变化
watch(() => props.modelValue, (val) => {
  visible.value = val
  if (val) {
    initForm()
    nextTick(() => {
      refreshPreview()
    })
  }
})

watch(visible, (val) => {
  emit('update:modelValue', val)
})

// 初始化表单
const initForm = () => {
  if (props.templateData) {
    Object.assign(form, {
      name: props.templateData.name || '',
      type: props.templateData.type || 'email',
      subject: props.templateData.subject || '',
      content: props.templateData.content || '',
      is_default: props.templateData.is_default || false,
      description: props.templateData.description || ''
    })
  } else {
    Object.assign(form, {
      name: '',
      type: 'email',
      subject: '',
      content: '',
      is_default: false,
      description: ''
    })
  }
}

// 类型变化处理
const handleTypeChange = (type) => {
  // 根据类型设置默认主题
  if (type === 'email' && !form.subject) {
    form.subject = '【告警】{{.System.AppName}} - {{.Email.Subject}}'
  }
  
  // 根据类型设置默认内容模版
  if (!form.content) {
    switch (type) {
      case 'email':
        form.content = getDefaultEmailTemplate()
        break
      case 'dingtalk':
        form.content = getDefaultDingTalkTemplate()
        break
      case 'wechat':
        form.content = getDefaultWeChatTemplate()
        break
      case 'markdown':
        form.content = getDefaultMarkdownTemplate()
        break
    }
  }
  
  refreshPreview()
}

// 默认模版内容
const getDefaultEmailTemplate = () => `<h2>🚨 系统告警通知</h2>
<p><strong>邮件主题：</strong>{{.Email.Subject}}</p>
<p><strong>发件人：</strong>{{.Email.Sender}}</p>
<p><strong>接收时间：</strong>{{.Email.ReceivedAt}}</p>
<div style="background-color: #f8f9fa; padding: 10px; margin: 10px 0;">
  <h3>邮件内容</h3>
  <div>{{.Email.Content}}</div>
</div>
<p><small>通知时间：{{.Time.NowFormat}} | 系统：{{.System.AppName}}</small></p>`

const getDefaultDingTalkTemplate = () => `## 🚨 系统告警通知

**告警详情**
- **邮件主题：** {{.Email.Subject}}
- **发件人：** {{.Email.Sender}}
- **接收时间：** {{.Email.ReceivedAt}}

**邮件内容**
> {{.Email.Content}}

**系统信息**
- **通知时间：** {{.Time.NowFormat}}
- **系统：** {{.System.AppName}}`

const getDefaultWeChatTemplate = () => `【系统告警通知】

邮件主题：{{.Email.Subject}}
发件人：{{.Email.Sender}}
接收时间：{{.Email.ReceivedAt}}

邮件内容：
{{.Email.Content}}

通知时间：{{.Time.NowFormat}}
系统：{{.System.AppName}}`

const getDefaultMarkdownTemplate = () => `# 🚨 系统告警通知

| 项目 | 内容 |
|------|------|
| 邮件主题 | {{.Email.Subject}} |
| 发件人 | {{.Email.Sender}} |
| 接收时间 | {{.Email.ReceivedAt}} |

## 邮件内容

\`\`\`
{{.Email.Content}}
\`\`\`

---
*通知时间：{{.Time.NowFormat}} | 系统：{{.System.AppName}}*`

// 刷新预览
const refreshPreview = async () => {
  if (!form.content) return
  
  previewLoading.value = true
  try {
    const response = await templatesAPI.preview({
      content: form.content,
      subject: form.subject
    })
    
    previewData.value = response.data
  } catch (error) {
    console.error('预览失败:', error)
    ElMessage.error('预览失败，请检查模版语法')
  } finally {
    previewLoading.value = false
  }
}

// 渲染预览内容
const renderPreviewContent = computed(() => {
  if (!previewData.value.content) return '内容预览...'
  
  const content = previewData.value.content
  
  switch (form.type) {
    case 'email':
      return content // HTML 直接渲染
    case 'dingtalk':
    case 'markdown':
      // 简单的 Markdown 渲染
      return content
        .replace(/^## (.*$)/gm, '<h2>$1</h2>')
        .replace(/^# (.*$)/gm, '<h1>$1</h1>')
        .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
        .replace(/\*(.*?)\*/g, '<em>$1</em>')
        .replace(/^- (.*$)/gm, '<li>$1</li>')
        .replace(/\n/g, '<br>')
    case 'wechat':
    default:
      return content.replace(/\n/g, '<br>')
  }
})

// 插入变量
const insertVariable = () => {
  showVariableSelector.value = true
}

// 处理变量插入
const handleInsertVariable = (variable) => {
  form.content += `{{${variable}}}`
  refreshPreview()
}

const handleSelectVariable = (variable) => {
  handleInsertVariable(variable)
  showVariableSelector.value = false
}

// 格式化内容
const formatContent = () => {
  if (form.type === 'email') {
    // 简单的HTML格式化
    form.content = form.content
      .replace(/></g, '>\n<')
      .replace(/^\s+/gm, '')
  }
  refreshPreview()
}

// 提交表单
const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    
    submitLoading.value = true
    
    const submitData = {
      name: form.name,
      type: form.type,
      subject: form.subject,
      content: form.content,
      is_default: form.is_default,
      description: form.description
    }
    
    if (isEdit.value) {
      await templatesAPI.update(props.templateData.id, submitData)
      ElMessage.success('模版更新成功')
    } else {
      await templatesAPI.create(submitData)
      ElMessage.success('模版创建成功')
    }
    
    emit('success')
    visible.value = false
  } catch (error) {
    console.error('提交失败:', error)
  } finally {
    submitLoading.value = false
  }
}

// 对话框关闭处理
const handleClosed = () => {
  formRef.value?.resetFields()
  previewData.value = {
    subject: '',
    content: '',
    used_vars: []
  }
}

// 监听内容变化，自动预览
watch([() => form.content, () => form.subject], () => {
  if (form.content) {
    const timer = setTimeout(() => {
      refreshPreview()
    }, 500) // 防抖
    return () => clearTimeout(timer)
  }
}, { deep: true })
</script>

<style scoped>
.template-form {
  max-height: 70vh;
  overflow-y: auto;
}

.code-editor {
  width: 100%;
}

.editor-toolbar {
  margin-bottom: 8px;
  padding: 8px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.preview-section {
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  overflow: hidden;
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background-color: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
  font-weight: 600;
}

.preview-tabs {
  height: 400px;
}

.preview-content {
  padding: 16px;
  min-height: 300px;
  max-height: 350px;
  overflow-y: auto;
  font-size: 14px;
  line-height: 1.6;
}

.subject-preview {
  background-color: #f0f9ff;
  border: 1px solid #bfdbfe;
  border-radius: 4px;
  font-weight: 600;
  color: #1e40af;
}

.html-preview {
  background-color: #ffffff;
}

.markdown-preview {
  background-color: #fafbfc;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif;
}

.text-preview {
  background-color: #f8f9fa;
  white-space: pre-wrap;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}

.used-variables {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 16px;
}

.variable-tag {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}

.no-variables {
  color: #909399;
  font-style: italic;
}

.dialog-footer {
  text-align: right;
}

:deep(.el-textarea__inner) {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace !important;
  font-size: 13px;
  line-height: 1.5;
}

:deep(.el-tabs__content) {
  height: 350px;
  overflow: hidden;
}

:deep(.el-tab-pane) {
  height: 100%;
  overflow-y: auto;
}
</style> 