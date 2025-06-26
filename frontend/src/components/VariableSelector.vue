<template>
  <el-dialog
    v-model="visible"
    title="选择变量"
    width="50%"
    :close-on-click-modal="false"
  >
    <div class="variable-selector">
      <div class="search-bar">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索变量..."
          clearable
          @input="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>
      
      <div class="variable-tree">
        <el-tree
          ref="treeRef"
          :data="treeData"
          node-key="id"
          :props="treeProps"
          :filter-node-method="filterNode"
          :expand-on-click-node="false"
          default-expand-all
        >
          <template #default="{ data }">
            <div class="tree-node">
              <span class="node-label">{{ data.label }}</span>
              <el-button
                v-if="data.variable"
                size="small"
                type="primary"
                @click="handleSelect(data.variable)"
              >
                选择
              </el-button>
            </div>
          </template>
        </el-tree>
      </div>
    </div>
  </el-dialog>
</template>

<script setup>
import { ref, computed, nextTick } from 'vue'
import { Search } from '@element-plus/icons-vue'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'select'])

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const searchKeyword = ref('')
const treeRef = ref()

const treeProps = {
  children: 'children',
  label: 'label'
}

// 构建树形数据
const treeData = [
  {
    id: 'email',
    label: '邮件变量',
    children: [
      { id: 'email-subject', label: '邮件主题', variable: '.Email.Subject' },
      { id: 'email-sender', label: '发件人', variable: '.Email.Sender' },
      { id: 'email-content', label: '邮件内容', variable: '.Email.Content' },
      { id: 'email-html-content', label: 'HTML内容', variable: '.Email.HTMLContent' },
      { id: 'email-received-at', label: '接收时间', variable: '.Email.ReceivedAt' },
      { id: 'email-size', label: '邮件大小', variable: '.Email.Size' },
      { id: 'email-message-id', label: '邮件ID', variable: '.Email.MessageID' }
    ]
  },
  {
    id: 'alert',
    label: '告警变量',
    children: [
      { id: 'alert-id', label: '告警ID', variable: '.Alert.ID' },
      { id: 'alert-subject', label: '告警主题', variable: '.Alert.Subject' },
      { id: 'alert-status', label: '告警状态', variable: '.Alert.Status' },
      { id: 'alert-content', label: '告警内容', variable: '.Alert.Content' },
      { id: 'alert-created-at', label: '创建时间', variable: '.Alert.CreatedAt' },
      { id: 'alert-received-at', label: '接收时间', variable: '.Alert.ReceivedAt' }
    ]
  },
  {
    id: 'rule',
    label: '规则变量',
    children: [
      { id: 'rule-id', label: '规则ID', variable: '.Rule.ID' },
      { id: 'rule-name', label: '规则名称', variable: '.Rule.Name' },
      { id: 'rule-match-type', label: '匹配类型', variable: '.Rule.MatchType' },
      { id: 'rule-match-value', label: '匹配值', variable: '.Rule.MatchValue' },
      { id: 'rule-priority', label: '优先级', variable: '.Rule.Priority' },
      { id: 'rule-description', label: '规则描述', variable: '.Rule.Description' }
    ]
  },
  {
    id: 'mailbox',
    label: '邮箱变量',
    children: [
      { id: 'mailbox-id', label: '邮箱ID', variable: '.Mailbox.ID' },
      { id: 'mailbox-name', label: '邮箱名称', variable: '.Mailbox.Name' },
      { id: 'mailbox-email', label: '邮箱地址', variable: '.Mailbox.Email' },
      { id: 'mailbox-host', label: 'IMAP主机', variable: '.Mailbox.Host' },
      { id: 'mailbox-port', label: 'IMAP端口', variable: '.Mailbox.Port' }
    ]
  },
  {
    id: 'system',
    label: '系统变量',
    children: [
      { id: 'system-app-name', label: '应用名称', variable: '.System.AppName' },
      { id: 'system-app-version', label: '应用版本', variable: '.System.AppVersion' },
      { id: 'system-server-name', label: '服务器名称', variable: '.System.ServerName' },
      { id: 'system-environment', label: '运行环境', variable: '.System.Environment' }
    ]
  },
  {
    id: 'time',
    label: '时间变量',
    children: [
      { id: 'time-now', label: '当前时间戳', variable: '.Time.Now' },
      { id: 'time-now-format', label: '格式化时间', variable: '.Time.NowFormat' },
      { id: 'time-today', label: '今天日期', variable: '.Time.Today' },
      { id: 'time-year', label: '年份', variable: '.Time.Year' },
      { id: 'time-month', label: '月份', variable: '.Time.Month' },
      { id: 'time-day', label: '日期', variable: '.Time.Day' },
      { id: 'time-hour', label: '小时', variable: '.Time.Hour' },
      { id: 'time-minute', label: '分钟', variable: '.Time.Minute' },
      { id: 'time-second', label: '秒', variable: '.Time.Second' },
      { id: 'time-weekday', label: '星期', variable: '.Time.Weekday' }
    ]
  }
]

// 过滤节点
const filterNode = (value, data) => {
  if (!value) return true
  return data.label.includes(value) || (data.variable && data.variable.includes(value))
}

// 处理搜索
const handleSearch = (value) => {
  treeRef.value?.filter(value)
}

// 处理选择
const handleSelect = (variable) => {
  emit('select', variable)
}
</script>

<style scoped>
.variable-selector {
  max-height: 60vh;
}

.search-bar {
  margin-bottom: 16px;
}

.variable-tree {
  max-height: 50vh;
  overflow-y: auto;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  padding: 8px;
}

.tree-node {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  padding-right: 8px;
}

.node-label {
  flex: 1;
  font-size: 14px;
}

:deep(.el-tree-node__content) {
  height: 36px;
}

:deep(.el-tree-node__content:hover) {
  background-color: #f5f7fa;
}
</style> 