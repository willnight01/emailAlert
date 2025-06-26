<template>
  <el-form
    ref="formRef"
    :model="formData"
    :rules="formRules"
    label-width="100px"
    label-position="left"
  >
    <el-form-item label="邮箱名称" prop="name">
      <el-input 
        v-model="formData.name" 
        placeholder="请输入邮箱名称"
        :disabled="loading"
      />
    </el-form-item>
    
    <el-form-item label="邮箱地址" prop="email">
      <el-input 
        v-model="formData.email" 
        placeholder="请输入邮箱地址"
        :disabled="loading"
        @blur="autoFillUsername"
      />
    </el-form-item>
    
    <el-form-item label="协议类型" prop="protocol">
      <el-select 
        v-model="formData.protocol" 
        placeholder="请选择协议类型" 
        style="width: 100%"
        :disabled="loading"
        @change="onProtocolChange"
      >
        <el-option label="IMAP" value="IMAP" />
        <el-option label="POP3" value="POP3" />
      </el-select>
    </el-form-item>
    
    <el-row :gutter="20">
      <el-col :span="14">
        <el-form-item label="服务器地址" prop="host">
          <el-input 
            v-model="formData.host" 
            placeholder="如: imap.gmail.com"
            :disabled="loading"
            @blur="autoFillPort"
          />
        </el-form-item>
      </el-col>
      <el-col :span="10">
        <el-form-item label="端口" prop="port">
          <el-input-number 
            v-model="formData.port" 
            :min="1" 
            :max="65535" 
            style="width: 100%" 
            placeholder="端口号"
            :disabled="loading"
            :controls-position="right"
            size="default"
          />
        </el-form-item>
      </el-col>
    </el-row>
    
    <el-form-item label="用户名" prop="username">
      <el-input 
        v-model="formData.username" 
        placeholder="通常为邮箱地址"
        :disabled="loading"
      />
    </el-form-item>
    
    <el-form-item label="密码" prop="password">
      <el-input 
        v-model="formData.password" 
        type="text" 
        placeholder="请输入密码或应用专用密码"
        :disabled="loading"
      />
      <div class="password-hint">
        <el-text size="small" type="info">
          <el-icon><InfoFilled /></el-icon>
          对于Gmail等邮箱，请使用应用专用密码而非登录密码
        </el-text>
      </div>
    </el-form-item>
    
    <el-form-item label="SSL加密">
      <el-switch 
        v-model="formData.ssl" 
        active-text="启用"
        inactive-text="禁用"
        :disabled="loading"
        @change="onSSLChange"
      />
    </el-form-item>
    
    <el-form-item label="描述">
      <el-input 
        v-model="formData.description" 
        type="textarea" 
        :rows="3" 
        placeholder="邮箱配置描述（可选）"
        :disabled="loading"
      />
    </el-form-item>

    <!-- 常用邮箱快捷配置 -->
    <el-form-item label="快捷配置">
      <el-button-group>
        <el-button 
          size="small" 
          @click="applyQuickConfig('gmail')"
          :disabled="loading"
        >
          Gmail
        </el-button>
        <el-button 
          size="small" 
          @click="applyQuickConfig('outlook')"
          :disabled="loading"
        >
          Outlook
        </el-button>
        <el-button 
          size="small" 
          @click="applyQuickConfig('qq')"
          :disabled="loading"
        >
          QQ邮箱
        </el-button>
        <el-button 
          size="small" 
          @click="applyQuickConfig('163')"
          :disabled="loading"
        >
          163邮箱
        </el-button>
      </el-button-group>
      <div class="quick-config-hint">
        <el-text size="small" type="info">点击快捷配置按钮自动填充服务器配置</el-text>
      </div>
    </el-form-item>
  </el-form>
</template>

<script setup>
import { ref, watch } from 'vue'
import { InfoFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

// Props
const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({})
  },
  loading: {
    type: Boolean,
    default: false
  }
})

// Emits
const emit = defineEmits(['update:modelValue', 'validate'])

// Refs
const formRef = ref()

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
  description: '',
  ...props.modelValue
})

// 表单验证规则
const formRules = {
  name: [
    { required: true, message: '请输入邮箱名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
  ],
  protocol: [
    { required: true, message: '请选择协议类型', trigger: 'change' }
  ],
  host: [
    { required: true, message: '请输入服务器地址', trigger: 'blur' }
  ],
  port: [
    { required: true, message: '请输入端口号', trigger: 'blur' },
    { type: 'number', min: 1, max: 65535, message: '端口号范围 1-65535', trigger: 'blur' }
  ],
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ]
}

// 常用邮箱配置
const emailConfigs = {
  gmail: {
    imap: { host: 'imap.gmail.com', port: 993, ssl: true },
    pop3: { host: 'pop.gmail.com', port: 995, ssl: true }
  },
  outlook: {
    imap: { host: 'outlook.office365.com', port: 993, ssl: true },
    pop3: { host: 'outlook.office365.com', port: 995, ssl: true }
  },
  qq: {
    imap: { host: 'imap.qq.com', port: 993, ssl: true },
    pop3: { host: 'pop.qq.com', port: 995, ssl: true }
  },
  163: {
    imap: { host: 'imap.163.com', port: 993, ssl: true },
    pop3: { host: 'pop.163.com', port: 995, ssl: true }
  }
}

// 监听props变化
watch(() => props.modelValue, (newValue) => {
  Object.assign(formData.value, newValue)
}, { deep: true, immediate: true })

// 监听formData变化，触发更新
watch(formData, (newValue) => {
  emit('update:modelValue', { ...newValue })
}, { deep: true })

// 方法
const validate = async () => {
  try {
    await formRef.value.validate()
    return true
  } catch (error) {
    return false
  }
}

const resetFields = () => {
  formRef.value?.resetFields()
}

const clearValidate = () => {
  formRef.value?.clearValidate()
}

// 自动填充用户名
const autoFillUsername = () => {
  if (formData.value.email && !formData.value.username) {
    formData.value.username = formData.value.email
  }
}

// 协议变化处理
const onProtocolChange = () => {
  if (formData.value.host) {
    autoFillPort()
  }
}

// SSL变化处理
const onSSLChange = () => {
  autoFillPort()
}

// 根据服务器地址自动填充端口
const autoFillPort = () => {
  const host = formData.value.host.toLowerCase()
  const protocol = formData.value.protocol.toLowerCase()
  const ssl = formData.value.ssl
  
  // 常见端口配置
  const portMap = {
    imap: ssl ? 993 : 143,
    pop3: ssl ? 995 : 110
  }
  
  // 如果是已知的服务器，使用预设配置
  if (host.includes('gmail.com')) {
    const config = emailConfigs.gmail[protocol]
    if (config) {
      formData.value.port = config.port
      formData.value.ssl = config.ssl
    }
  } else if (host.includes('outlook.office365.com') || host.includes('hotmail.com')) {
    const config = emailConfigs.outlook[protocol]
    if (config) {
      formData.value.port = config.port
      formData.value.ssl = config.ssl
    }
  } else if (host.includes('qq.com')) {
    const config = emailConfigs.qq[protocol]
    if (config) {
      formData.value.port = config.port
      formData.value.ssl = config.ssl
    }
  } else if (host.includes('163.com')) {
    const config = emailConfigs['163'][protocol]
    if (config) {
      formData.value.port = config.port
      formData.value.ssl = config.ssl
    }
  } else {
    // 使用默认端口
    formData.value.port = portMap[protocol] || 993
  }
}

// 应用快捷配置
const applyQuickConfig = (provider) => {
  const config = emailConfigs[provider]
  if (!config) return
  
  const protocolConfig = config[formData.value.protocol.toLowerCase()]
  if (!protocolConfig) return
  
  formData.value.host = protocolConfig.host
  formData.value.port = protocolConfig.port
  formData.value.ssl = protocolConfig.ssl
  
  ElMessage.success(`已应用${provider.toUpperCase()}配置`)
}

// 暴露方法
defineExpose({
  validate,
  resetFields,
  clearValidate,
  formData
})
</script>

<style scoped>
.password-hint {
  margin-top: 5px;
}

.quick-config-hint {
  margin-top: 5px;
}

:deep(.el-input-number .el-input__inner) {
  text-align: left;
  font-size: 14px;
  font-weight: 500;
}

:deep(.el-input-number) {
  min-width: 120px;
}

:deep(.el-input-number .el-input__wrapper) {
  padding: 1px 15px 1px 11px;
}

:deep(.el-form-item__content) {
  line-height: 1.5;
}

/* 端口输入框专用样式优化 */
:deep(.el-row .el-col:last-child .el-input-number) {
  width: 100%;
}

:deep(.el-row .el-col:last-child .el-input-number .el-input__inner) {
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 0.5px;
  text-align: center;
}
</style> 