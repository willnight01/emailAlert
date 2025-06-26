<template>
  <el-dialog
    v-model="visible"
    title="模版变量帮助"
    width="70%"
    :close-on-click-modal="false"
  >
    <div class="variable-help">
      <div class="help-header">
        <el-alert
          title="变量使用说明"
          type="info"
          show-icon
          :closable="false"
        >
          <template #default>
            <p>在模版中使用变量时，请使用双大括号包围变量名，如：<code v-pre>{{.Email.Subject}}</code></p>
            <p>变量名区分大小写，请按照下面的格式准确使用</p>
          </template>
        </el-alert>
      </div>
      
      <el-tabs v-model="activeTab" class="variable-tabs">
        <el-tab-pane label="邮件变量" name="email">
          <VariableCategory
            title="邮件相关变量"
            description="从邮件中提取的基本信息"
            :variables="emailVariables"
            @insert="handleInsert"
          />
        </el-tab-pane>
        
        <el-tab-pane label="告警变量" name="alert">
          <VariableCategory
            title="告警相关变量"
            description="告警处理过程中生成的信息"
            :variables="alertVariables"
            @insert="handleInsert"
          />
        </el-tab-pane>
        
        <el-tab-pane label="规则变量" name="rule">
          <VariableCategory
            title="规则相关变量"
            description="匹配的告警规则信息"
            :variables="ruleVariables"
            @insert="handleInsert"
          />
        </el-tab-pane>
        
        <el-tab-pane label="邮箱变量" name="mailbox">
          <VariableCategory
            title="邮箱相关变量"
            description="邮箱配置信息"
            :variables="mailboxVariables"
            @insert="handleInsert"
          />
        </el-tab-pane>
        
        <el-tab-pane label="系统变量" name="system">
          <VariableCategory
            title="系统相关变量"
            description="系统环境信息"
            :variables="systemVariables"
            @insert="handleInsert"
          />
        </el-tab-pane>
        
        <el-tab-pane label="时间变量" name="time">
          <VariableCategory
            title="时间相关变量"
            description="各种时间格式"
            :variables="timeVariables"
            @insert="handleInsert"
          />
        </el-tab-pane>
        
        <el-tab-pane label="常用组合" name="examples">
          <div class="examples-section">
            <h3>常用变量组合示例</h3>
            <div class="example-items">
              <div 
                v-for="example in examples" 
                :key="example.name"
                class="example-item"
              >
                <div class="example-header">
                  <h4>{{ example.name }}</h4>
                  <el-button 
                    size="small" 
                    type="primary" 
                    @click="handleInsert(example.template)"
                  >
                    插入模版
                  </el-button>
                </div>
                <div class="example-description">{{ example.description }}</div>
                <div class="example-template">
                  <el-input
                    :model-value="example.template"
                    type="textarea"
                    :rows="3"
                    readonly
                  />
                </div>
              </div>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </el-dialog>
</template>

<script setup>
import { ref, computed } from 'vue'
import VariableCategory from './VariableCategory.vue'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'insert'])

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const activeTab = ref('email')

// 邮件变量
const emailVariables = [
  {
    name: '.Email.Subject',
    description: '邮件主题',
    example: '【告警】系统异常通知'
  },
  {
    name: '.Email.Sender',
    description: '发件人邮箱地址',
    example: 'monitor@example.com'
  },
  {
    name: '.Email.Content',
    description: '邮件正文内容',
    example: '数据库连接异常，请立即处理'
  },
  {
    name: '.Email.HTMLContent',
    description: '邮件HTML内容',
    example: '<p>数据库连接异常，请立即处理</p>'
  },
  {
    name: '.Email.ReceivedAt',
    description: '邮件接收时间',
    example: '2024-01-15 14:30:25'
  },
  {
    name: '.Email.Size',
    description: '邮件大小（字节）',
    example: '1024'
  },
  {
    name: '.Email.MessageID',
    description: '邮件唯一标识符',
    example: '<123456@example.com>'
  }
]

// 告警变量
const alertVariables = [
  {
    name: '.Alert.ID',
    description: '告警ID',
    example: '1001'
  },
  {
    name: '.Alert.Subject',
    description: '告警主题（处理后）',
    example: '【告警】系统异常通知'
  },
  {
    name: '.Alert.Status',
    description: '告警状态',
    example: 'active'
  },
  {
    name: '.Alert.Content',
    description: '告警内容（处理后）',
    example: '数据库连接异常，请立即处理'
  },
  {
    name: '.Alert.CreatedAt',
    description: '告警创建时间',
    example: '2024-01-15 14:30:25'
  },
  {
    name: '.Alert.ReceivedAt',
    description: '邮件接收时间',
    example: '2024-01-15 14:30:25'
  }
]

// 规则变量
const ruleVariables = [
  {
    name: '.Rule.ID',
    description: '规则ID',
    example: '1'
  },
  {
    name: '.Rule.Name',
    description: '规则名称',
    example: '生产环境告警规则'
  },
  {
    name: '.Rule.MatchType',
    description: '匹配类型',
    example: 'keyword'
  },
  {
    name: '.Rule.MatchValue',
    description: '匹配值',
    example: '错误,异常,告警'
  },
  {
    name: '.Rule.Priority',
    description: '规则优先级',
    example: '9'
  },
  {
    name: '.Rule.Description',
    description: '规则描述',
    example: '用于匹配生产环境的告警信息'
  }
]

// 邮箱变量
const mailboxVariables = [
  {
    name: '.Mailbox.ID',
    description: '邮箱ID',
    example: '1'
  },
  {
    name: '.Mailbox.Name',
    description: '邮箱名称',
    example: '监控邮箱'
  },
  {
    name: '.Mailbox.Email',
    description: '邮箱地址',
    example: 'monitor@example.com'
  },
  {
    name: '.Mailbox.Host',
    description: 'IMAP服务器地址',
    example: 'imap.example.com'
  },
  {
    name: '.Mailbox.Port',
    description: 'IMAP端口',
    example: '993'
  }
]

// 系统变量
const systemVariables = [
  {
    name: '.System.AppName',
    description: '应用名称',
    example: '统一邮件告警平台'
  },
  {
    name: '.System.AppVersion',
    description: '应用版本',
    example: 'v1.0.0'
  },
  {
    name: '.System.ServerName',
    description: '服务器名称',
    example: 'alert-server-01'
  },
  {
    name: '.System.Environment',
    description: '运行环境',
    example: 'production'
  }
]

// 时间变量
const timeVariables = [
  {
    name: '.Time.Now',
    description: '当前时间戳',
    example: '1705314625'
  },
  {
    name: '.Time.NowFormat',
    description: '当前时间（格式化）',
    example: '2024-01-15 14:30:25'
  },
  {
    name: '.Time.Today',
    description: '今天日期',
    example: '2024-01-15'
  },
  {
    name: '.Time.Year',
    description: '当前年份',
    example: '2024'
  },
  {
    name: '.Time.Month',
    description: '当前月份',
    example: '01'
  },
  {
    name: '.Time.Day',
    description: '当前日期',
    example: '15'
  },
  {
    name: '.Time.Hour',
    description: '当前小时',
    example: '14'
  },
  {
    name: '.Time.Minute',
    description: '当前分钟',
    example: '30'
  },
  {
    name: '.Time.Second',
    description: '当前秒数',
    example: '25'
  },
  {
    name: '.Time.Weekday',
    description: '星期几',
    example: '星期一'
  }
]

// 常用组合示例
const examples = [
  {
    name: '简单邮件通知',
    description: '最基础的邮件告警通知格式',
    template: `主题：{{.Email.Subject}}
发件人：{{.Email.Sender}}
时间：{{.Email.ReceivedAt}}

內容：
{{.Email.Content}}

通知时间：{{.Time.NowFormat}}`
  },
  {
    name: '详细告警信息',
    description: '包含规则和系统信息的详细告警',
    template: `## 🚨 {{.System.AppName}} 告警通知

**告警信息**
- 主题：{{.Email.Subject}}
- 发件人：{{.Email.Sender}}
- 接收时间：{{.Email.ReceivedAt}}

**匹配规则**
- 规则名称：{{.Rule.Name}}
- 匹配类型：{{.Rule.MatchType}}
- 优先级：{{.Rule.Priority}}

**邮件内容**
{{.Email.Content}}

---
*系统：{{.System.AppName}} | 服务器：{{.System.ServerName}} | 通知时间：{{.Time.NowFormat}}*`
  },
  {
    name: 'HTML邮件模版',
    description: '富文本格式的邮件通知模版',
    template: `<h2 style="color: #e74c3c;">🚨 系统告警通知</h2>

<table style="width: 100%; border-collapse: collapse;">
  <tr>
    <td style="padding: 8px; border: 1px solid #ddd; background-color: #f9f9f9;"><strong>邮件主题</strong></td>
    <td style="padding: 8px; border: 1px solid #ddd;">{{.Email.Subject}}</td>
  </tr>
  <tr>
    <td style="padding: 8px; border: 1px solid #ddd; background-color: #f9f9f9;"><strong>发件人</strong></td>
    <td style="padding: 8px; border: 1px solid #ddd;">{{.Email.Sender}}</td>
  </tr>
  <tr>
    <td style="padding: 8px; border: 1px solid #ddd; background-color: #f9f9f9;"><strong>接收时间</strong></td>
    <td style="padding: 8px; border: 1px solid #ddd;">{{.Email.ReceivedAt}}</td>
  </tr>
</table>

<div style="background-color: #f8f9fa; padding: 15px; margin: 15px 0; border-left: 4px solid #007bff;">
  <h3>邮件内容</h3>
  <p>{{.Email.Content}}</p>
</div>

<hr>
<p style="color: #6c757d; font-size: 12px;">
  通知时间：{{.Time.NowFormat}} | 系统：{{.System.AppName}} v{{.System.AppVersion}}
</p>`
  }
]

const handleInsert = (variable) => {
  emit('insert', variable)
}
</script>

<style scoped>
.variable-help {
  max-height: 70vh;
}

.help-header {
  margin-bottom: 20px;
}

.help-header code {
  background-color: #f1f2f6;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  color: #e74c3c;
}

.variable-tabs {
  height: 60vh;
}

:deep(.el-tabs__content) {
  height: calc(60vh - 50px);
  overflow-y: auto;
}

.examples-section {
  padding: 20px;
}

.examples-section h3 {
  margin-bottom: 20px;
  color: #2c3e50;
}

.example-items {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.example-item {
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  padding: 16px;
  background-color: #fafbfc;
}

.example-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.example-header h4 {
  margin: 0;
  color: #2c3e50;
}

.example-description {
  color: #6c757d;
  font-size: 14px;
  margin-bottom: 12px;
}

.example-template {
  margin-top: 8px;
}

:deep(.example-template .el-textarea__inner) {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  background-color: #f8f9fa;
  border: 1px solid #dee2e6;
}
</style> 