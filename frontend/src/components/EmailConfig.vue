<template>
  <div class="email-config">
    <el-form-item label="SMTP服务器" prop="host">
      <el-input 
        v-model="localConfig.host" 
        placeholder="请输入SMTP服务器地址"
        @input="emitChange"
      />
      <div class="form-tip">
        <el-text type="info" size="small">
          如：smtp.gmail.com, smtp.qq.com, smtp.126.com
        </el-text>
      </div>
    </el-form-item>

    <el-form-item label="端口" prop="port">
      <el-input-number 
        v-model="localConfig.port" 
        placeholder="请输入SMTP端口"
        :min="1"
        :max="65535"
        style="width: 100%"
        @change="emitChange"
      />
      <div class="form-tip">
        <el-text type="info" size="small">
          SSL: 465, TLS: 587, 无加密: 25
        </el-text>
      </div>
    </el-form-item>

    <el-form-item label="用户名" prop="username">
      <el-input 
        v-model="localConfig.username" 
        placeholder="请输入SMTP用户名"
        @input="emitChange"
      />
    </el-form-item>

    <el-form-item label="密码" prop="password">
      <el-input 
        v-model="localConfig.password" 
        placeholder="请输入SMTP密码"
        type="password"
        show-password
        @input="emitChange"
      />
      <div class="form-tip">
        <el-text type="info" size="small">
          Gmail等需要使用应用专用密码
        </el-text>
      </div>
    </el-form-item>

    <el-form-item label="SSL加密" prop="ssl">
      <el-switch 
        v-model="localConfig.ssl" 
        @change="emitChange"
      />
      <div class="form-tip">
        <el-text type="info" size="small">
          开启后使用SSL/TLS加密连接
        </el-text>
      </div>
    </el-form-item>

    <el-form-item label="发件人" prop="from">
      <el-input 
        v-model="localConfig.from" 
        placeholder="请输入发件人邮箱地址"
        @input="emitChange"
      />
    </el-form-item>

    <el-form-item label="发件人姓名" prop="from_name">
      <el-input 
        v-model="localConfig.from_name" 
        placeholder="请输入发件人姓名（可选）"
        @input="emitChange"
      />
    </el-form-item>

    <el-form-item label="收件人" prop="to">
      <el-input 
        v-model="toAddresses" 
        placeholder="请输入收件人邮箱地址，多个用逗号分隔"
        @input="handleToChange"
      />
      <div class="form-tip">
        <el-text type="info" size="small">
          多个收件人用逗号分隔，如：user1@example.com,user2@example.com
        </el-text>
      </div>
    </el-form-item>

    <el-form-item label="抄送" prop="cc">
      <el-input 
        v-model="ccAddresses" 
        placeholder="请输入抄送邮箱地址，多个用逗号分隔（可选）"
        @input="handleCcChange"
      />
    </el-form-item>

    <el-form-item label="邮件格式" prop="format">
      <el-radio-group v-model="localConfig.format" @change="emitChange">
        <el-radio label="text">纯文本</el-radio>
        <el-radio label="html">HTML</el-radio>
        <el-radio label="mixed">混合格式</el-radio>
      </el-radio-group>
      <div class="form-tip">
        <el-text type="info" size="small">
          混合格式推荐使用，兼容性最好
        </el-text>
      </div>
    </el-form-item>

    <el-form-item label="邮件主题" prop="subject">
      <el-input 
        v-model="localConfig.subject" 
        placeholder="请输入邮件主题模板（可选）"
        @input="emitChange"
      />
              <div class="form-tip">
          <el-text type="info" size="small">
            支持变量：{{title}}、{{date}}、{{time}}
          </el-text>
        </div>
    </el-form-item>

    <el-form-item label="邮件优先级" prop="priority">
      <el-select v-model="localConfig.priority" placeholder="选择邮件优先级" @change="emitChange">
        <el-option :value="1" label="高优先级" />
        <el-option :value="3" label="普通优先级" />
        <el-option :value="5" label="低优先级" />
      </el-select>
    </el-form-item>

    <el-form-item>
      <el-button 
        type="primary" 
        @click="handleTest"
        :loading="testing"
        icon="Connection"
      >
        测试配置
      </el-button>
    </el-form-item>
  </div>
</template>

<script setup>
import { ref, reactive, watch, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({})
  }
})

const emit = defineEmits(['update:modelValue', 'test'])

const testing = ref(false)
const localConfig = reactive({
  host: '',
  port: 587,
  username: '',
  password: '',
  ssl: true,
  from: '',
  from_name: '',
  to: [],
  cc: [],
  format: 'mixed',
  subject: '',
  priority: 3
})

// 用于显示的地址字符串
const toAddresses = ref('')
const ccAddresses = ref('')

// 监听props变化
watch(() => props.modelValue, (newVal) => {
  if (newVal && typeof newVal === 'object') {
    Object.assign(localConfig, {
      host: '',
      port: 587,
      username: '',
      password: '',
      ssl: true,
      from: '',
      from_name: '',
      to: [],
      cc: [],
      format: 'mixed',
      subject: '',
      priority: 3,
      ...newVal
    })
    
    // 更新显示字符串
    toAddresses.value = Array.isArray(localConfig.to) ? localConfig.to.join(',') : (localConfig.to || '')
    ccAddresses.value = Array.isArray(localConfig.cc) ? localConfig.cc.join(',') : (localConfig.cc || '')
  }
}, { immediate: true })

const emitChange = () => {
  emit('update:modelValue', { ...localConfig })
}

const handleToChange = (value) => {
  toAddresses.value = value
  localConfig.to = value ? value.split(',').map(email => email.trim()).filter(email => email) : []
  emitChange()
}

const handleCcChange = (value) => {
  ccAddresses.value = value
  localConfig.cc = value ? value.split(',').map(email => email.trim()).filter(email => email) : []
  emitChange()
}

const handleTest = async () => {
  testing.value = true
  try {
    await emit('test', { ...localConfig })
  } finally {
    testing.value = false
  }
}

onMounted(() => {
  // 初始化时触发一次
  emitChange()
})
</script>

<style scoped>
.email-config {
  margin-top: 10px;
}

.form-tip {
  margin-top: 5px;
}
</style> 