<template>
  <el-dialog
    v-model="visible"
    title="模版预览"
    width="80%"
    :close-on-click-modal="false"
  >
    <div class="template-preview">
      <div class="preview-header">
        <div class="template-info">
          <h3>{{ templateData?.name }}</h3>
          <div class="template-meta">
            <el-tag :type="getTypeTagType(templateData?.type)" size="small">
              {{ getTypeLabel(templateData?.type) }}
            </el-tag>
            <el-tag v-if="templateData?.is_default" type="success" size="small">默认模版</el-tag>
            <span class="meta-text">更新时间：{{ formatTime(templateData?.updated_at) }}</span>
          </div>
        </div>
        <div class="preview-actions">
          <el-button @click="refreshPreview" :loading="loading">
            <el-icon><Refresh /></el-icon>
            刷新预览
          </el-button>
        </div>
      </div>
      
      <el-tabs v-model="activeTab" class="preview-tabs">
        <!-- 预览效果 -->
        <el-tab-pane label="预览效果" name="preview">
          <div class="preview-content">
            <div v-if="templateData?.type === 'email' && templateData?.subject" class="subject-section">
              <h4>邮件主题</h4>
              <div class="subject-preview">
                {{ previewData.subject || templateData.subject }}
              </div>
            </div>
            
            <div class="content-section">
              <h4>内容预览</h4>
              <div 
                class="content-preview"
                :class="{
                  'html-preview': templateData?.type === 'email',
                  'markdown-preview': templateData?.type === 'dingtalk' || templateData?.type === 'markdown',
                  'text-preview': templateData?.type === 'wechat'
                }"
                v-html="renderPreviewContent"
              >
              </div>
            </div>
          </div>
        </el-tab-pane>
        
        <!-- 原始模版 -->
        <el-tab-pane label="原始模版" name="template">
          <div class="template-source">
            <div v-if="templateData?.type === 'email' && templateData?.subject" class="subject-source">
              <h4>邮件主题模版</h4>
              <el-input
                :model-value="templateData.subject"
                type="textarea"
                :rows="2"
                readonly
                class="template-textarea"
              />
            </div>
            
            <div class="content-source">
              <h4>内容模版</h4>
              <el-input
                :model-value="templateData?.content || ''"
                type="textarea"
                :rows="15"
                readonly
                class="template-textarea"
              />
            </div>
          </div>
        </el-tab-pane>
        
        <!-- 使用变量 -->
        <el-tab-pane label="使用变量" name="variables">
          <div class="variables-section">
            <h4>模版中使用的变量</h4>
            <div v-if="previewData.used_vars && previewData.used_vars.length > 0" class="variable-list">
              <el-tag 
                v-for="variable in previewData.used_vars" 
                :key="variable"
                size="medium"
                class="variable-tag"
              >
                {{ variable }}
              </el-tag>
            </div>
            <div v-else class="no-variables">
              <el-empty description="该模版暂未使用任何变量" />
            </div>
          </div>
        </el-tab-pane>
        
        <!-- 示例数据 -->
        <el-tab-pane label="示例数据" name="data">
          <div class="sample-data">
            <h4>预览使用的示例数据</h4>
            <el-input
              :model-value="JSON.stringify(sampleData, null, 2)"
              type="textarea"
              :rows="20"
              readonly
              class="json-textarea"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
    
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="visible = false">关闭</el-button>
        <el-button type="primary" @click="handleEdit">
          <el-icon><Edit /></el-icon>
          编辑模版
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Edit } from '@element-plus/icons-vue'
import { templatesAPI } from '@/api'

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

const emit = defineEmits(['update:modelValue', 'edit'])

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const loading = ref(false)
const activeTab = ref('preview')

const previewData = ref({
  subject: '',
  content: '',
  used_vars: []
})

// 示例数据
const sampleData = {
  Email: {
    Subject: '【告警】生产环境数据库连接异常',
    Sender: 'monitor@example.com',
    Content: '生产环境数据库连接出现异常，连接超时。请立即检查数据库服务器状态。错误信息：Connection timeout after 30 seconds',
    HTMLContent: '<p>生产环境数据库连接出现异常，连接超时。</p><p>请立即检查数据库服务器状态。</p><p><strong>错误信息：</strong>Connection timeout after 30 seconds</p>',
    ReceivedAt: '2024-01-15 14:30:25',
    Size: 1024,
    MessageID: '<123456789@example.com>'
  },
  Alert: {
    ID: 1001,
    Subject: '【告警】生产环境数据库连接异常',
    Status: 'active',
    Content: '生产环境数据库连接出现异常，连接超时。请立即检查数据库服务器状态。',
    CreatedAt: '2024-01-15 14:30:25',
    ReceivedAt: '2024-01-15 14:30:25'
  },
  Rule: {
    ID: 1,
    Name: '生产环境数据库告警',
    MatchType: 'keyword',
    MatchValue: '数据库,连接,异常,超时',
    Priority: 9,
    Description: '用于监控生产环境数据库相关告警'
  },
  Mailbox: {
    ID: 1,
    Name: '生产环境监控邮箱',
    Email: 'monitor@example.com',
    Host: 'imap.example.com',
    Port: 993
  },
  System: {
    AppName: '统一邮件告警平台',
    AppVersion: 'v1.0.0',
    ServerName: 'alert-server-01',
    Environment: 'production'
  },
  Time: {
    Now: Math.floor(Date.now() / 1000),
    NowFormat: new Date().toLocaleString('zh-CN'),
    Today: new Date().toLocaleDateString('zh-CN'),
    Year: new Date().getFullYear(),
    Month: String(new Date().getMonth() + 1).padStart(2, '0'),
    Day: String(new Date().getDate()).padStart(2, '0'),
    Hour: String(new Date().getHours()).padStart(2, '0'),
    Minute: String(new Date().getMinutes()).padStart(2, '0'),
    Second: String(new Date().getSeconds()).padStart(2, '0'),
    Weekday: ['星期日', '星期一', '星期二', '星期三', '星期四', '星期五', '星期六'][new Date().getDay()]
  }
}

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
    email: '邮件模版',
    dingtalk: '钉钉模版',
    wechat: '企业微信模版',
    markdown: 'Markdown模版'
  }
  return typeMap[type] || type
}

// 格式化时间
const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

// 渲染预览内容
const renderPreviewContent = computed(() => {
  if (!previewData.value.content) return '加载中...'
  
  const content = previewData.value.content
  
  switch (props.templateData?.type) {
    case 'email':
      return content // HTML 直接渲染
    case 'dingtalk':
    case 'markdown':
      // 简单的 Markdown 渲染
      return content
        .replace(/^### (.*$)/gm, '<h3>$1</h3>')
        .replace(/^## (.*$)/gm, '<h2>$1</h2>')
        .replace(/^# (.*$)/gm, '<h1>$1</h1>')
        .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
        .replace(/\*(.*?)\*/g, '<em>$1</em>')
        .replace(/^- (.*$)/gm, '<li>$1</li>')
        .replace(/^> (.*$)/gm, '<blockquote>$1</blockquote>')
        .replace(/`([^`]+)`/g, '<code>$1</code>')
        .replace(/\n/g, '<br>')
    case 'wechat':
    default:
      return content.replace(/\n/g, '<br>')
  }
})

// 刷新预览
const refreshPreview = async () => {
  if (!props.templateData?.content) return
  
  loading.value = true
  try {
    const response = await templatesAPI.preview({
      content: props.templateData.content,
      subject: props.templateData.subject || ''
    })
    
    previewData.value = response.data
  } catch (error) {
    console.error('预览失败:', error)
    ElMessage.error('预览失败，请检查模版语法')
  } finally {
    loading.value = false
  }
}

// 编辑模版
const handleEdit = () => {
  emit('edit', props.templateData)
  visible.value = false
}

// 监听模版数据变化，自动刷新预览
watch(() => props.templateData, (newData) => {
  if (newData && visible.value) {
    refreshPreview()
  }
}, { immediate: true })

// 监听对话框显示状态
watch(visible, (show) => {
  if (show && props.templateData) {
    activeTab.value = 'preview'
    refreshPreview()
  }
})
</script>

<style scoped>
.template-preview {
  max-height: 70vh;
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e4e7ed;
}

.template-info h3 {
  margin: 0 0 8px 0;
  color: #2c3e50;
  font-weight: 600;
}

.template-meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.meta-text {
  color: #6c757d;
  font-size: 14px;
}

.preview-tabs {
  height: 60vh;
}

:deep(.el-tabs__content) {
  height: calc(60vh - 50px);
  overflow-y: auto;
}

.preview-content {
  padding: 20px;
}

.subject-section,
.content-section {
  margin-bottom: 24px;
}

.subject-section h4,
.content-section h4 {
  margin: 0 0 12px 0;
  color: #2c3e50;
  font-weight: 600;
}

.subject-preview {
  background-color: #f0f9ff;
  border: 1px solid #bfdbfe;
  border-radius: 4px;
  padding: 12px;
  font-weight: 600;
  color: #1e40af;
}

.content-preview {
  background-color: #ffffff;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  padding: 16px;
  min-height: 300px;
  font-size: 14px;
  line-height: 1.6;
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

.template-source {
  padding: 20px;
}

.subject-source,
.content-source {
  margin-bottom: 24px;
}

.subject-source h4,
.content-source h4 {
  margin: 0 0 12px 0;
  color: #2c3e50;
  font-weight: 600;
}

.template-textarea :deep(.el-textarea__inner) {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  line-height: 1.5;
  background-color: #f8f9fa;
}

.variables-section {
  padding: 20px;
}

.variables-section h4 {
  margin: 0 0 16px 0;
  color: #2c3e50;
  font-weight: 600;
}

.variable-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.variable-tag {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
}

.sample-data {
  padding: 20px;
}

.sample-data h4 {
  margin: 0 0 12px 0;
  color: #2c3e50;
  font-weight: 600;
}

.json-textarea :deep(.el-textarea__inner) {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  line-height: 1.4;
  background-color: #f8f9fa;
}

.dialog-footer {
  text-align: right;
}

.no-variables {
  text-align: center;
  padding: 40px 0;
}
</style> 