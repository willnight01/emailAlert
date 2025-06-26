<template>
  <div class="webhook-config">
    <el-form-item label="Webhook地址" prop="url">
      <el-input 
        v-model="localConfig.url" 
        placeholder="请输入Webhook URL地址"
        @input="emitChange"
      />
      <div class="form-tip">
        <el-text type="info" size="small">
          支持HTTP/HTTPS协议，如：https://api.example.com/webhook
        </el-text>
      </div>
    </el-form-item>

    <el-form-item label="请求方法" prop="method">
      <el-select v-model="localConfig.method" placeholder="选择请求方法" @change="emitChange">
        <el-option value="GET" label="GET" />
        <el-option value="POST" label="POST" />
        <el-option value="PUT" label="PUT" />
        <el-option value="PATCH" label="PATCH" />
        <el-option value="DELETE" label="DELETE" />
      </el-select>
    </el-form-item>

    <el-form-item label="数据格式" prop="content_type">
      <el-select v-model="localConfig.content_type" placeholder="选择数据格式" @change="emitChange">
        <el-option value="application/json" label="JSON" />
        <el-option value="application/x-www-form-urlencoded" label="表单格式" />
        <el-option value="text/plain" label="纯文本" />
        <el-option value="text/xml" label="XML" />
        <el-option value="application/xml" label="Application XML" />
      </el-select>
    </el-form-item>

    <el-form-item label="认证方式" prop="auth_type">
      <el-select v-model="localConfig.auth_type" placeholder="选择认证方式" @change="handleAuthTypeChange">
        <el-option value="none" label="无认证" />
        <el-option value="basic" label="Basic认证" />
        <el-option value="bearer" label="Bearer Token" />
        <el-option value="apikey" label="API Key" />
      </el-select>
    </el-form-item>

    <!-- Basic认证 -->
    <template v-if="localConfig.auth_type === 'basic'">
      <el-form-item label="用户名" prop="username">
        <el-input 
          v-model="localConfig.username" 
          placeholder="请输入用户名"
          @input="emitChange"
        />
      </el-form-item>
      <el-form-item label="密码" prop="password">
        <el-input 
          v-model="localConfig.password" 
          placeholder="请输入密码"
          type="password"
          show-password
          @input="emitChange"
        />
      </el-form-item>
    </template>

    <!-- Bearer Token -->
    <template v-if="localConfig.auth_type === 'bearer'">
      <el-form-item label="Token" prop="token">
        <el-input 
          v-model="localConfig.token" 
          placeholder="请输入Bearer Token"
          type="password"
          show-password
          @input="emitChange"
        />
      </el-form-item>
    </template>

    <!-- API Key -->
    <template v-if="localConfig.auth_type === 'apikey'">
      <el-form-item label="API Key名称" prop="api_key_name">
        <el-input 
          v-model="localConfig.api_key_name" 
          placeholder="请输入API Key名称"
          @input="emitChange"
        />
        <div class="form-tip">
          <el-text type="info" size="small">
            Header名称，如：X-API-Key、Authorization
          </el-text>
        </div>
      </el-form-item>
      <el-form-item label="API Key值" prop="api_key_value">
        <el-input 
          v-model="localConfig.api_key_value" 
          placeholder="请输入API Key值"
          type="password"
          show-password
          @input="emitChange"
        />
      </el-form-item>
    </template>

    <el-form-item label="自定义Headers" prop="headers">
      <el-input 
        v-model="customHeaders" 
        type="textarea" 
        placeholder="请输入自定义HTTP头部（JSON格式，可选）"
        :rows="3"
        @input="handleHeadersChange"
      />
      <div class="form-tip">
        <el-text type="info" size="small">
          JSON格式，如：{"Content-Type": "application/json", "User-Agent": "MyApp"}
        </el-text>
      </div>
    </el-form-item>

    <el-form-item label="消息模板" prop="template">
      <el-input 
        v-model="localConfig.template" 
        type="textarea" 
        placeholder="请输入消息模板（可选）"
        :rows="4"
        @input="emitChange"
      />
              <div class="form-tip">
          <el-text type="info" size="small">
            支持变量：{{title}}、{{content}}、{{timestamp}}
          </el-text>
        </div>
    </el-form-item>

    <el-form-item label="超时时间" prop="timeout">
      <el-input-number 
        v-model="localConfig.timeout" 
        placeholder="请求超时时间（秒）"
        :min="1"
        :max="300"
        style="width: 100%"
        @change="emitChange"
      />
      <div class="form-tip">
        <el-text type="info" size="small">
          请求超时时间，单位秒，建议5-30秒
        </el-text>
      </div>
    </el-form-item>

    <el-form-item label="验证SSL" prop="verify_ssl">
      <el-switch 
        v-model="localConfig.verify_ssl" 
        @change="emitChange"
      />
      <div class="form-tip">
        <el-text type="info" size="small">
          关闭后将跳过SSL证书验证（不推荐）
        </el-text>
      </div>
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
import { ref, reactive, watch, onMounted } from 'vue'
import { ElMessage } from 'element-plus'

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({})
  }
})

const emit = defineEmits(['update:modelValue', 'test'])

const testing = ref(false)
const customHeaders = ref('')

const localConfig = reactive({
  url: '',
  method: 'POST',
  content_type: 'application/json',
  auth_type: 'none',
  username: '',
  password: '',
  token: '',
  api_key_name: '',
  api_key_value: '',
  headers: {},
  template: '',
  timeout: 10,
  verify_ssl: true
})

// 监听props变化
watch(() => props.modelValue, (newVal) => {
  if (newVal && typeof newVal === 'object') {
    Object.assign(localConfig, {
      url: '',
      method: 'POST',
      content_type: 'application/json',
      auth_type: 'none',
      username: '',
      password: '',
      token: '',
      api_key_name: '',
      api_key_value: '',
      headers: {},
      template: '',
      timeout: 10,
      verify_ssl: true,
      ...newVal
    })
    
    // 更新headers显示
    try {
      customHeaders.value = JSON.stringify(localConfig.headers || {}, null, 2)
    } catch (error) {
      customHeaders.value = ''
    }
  }
}, { immediate: true })

const emitChange = () => {
  emit('update:modelValue', { ...localConfig })
}

const handleAuthTypeChange = () => {
  // 切换认证类型时清空相关字段
  if (localConfig.auth_type === 'none') {
    localConfig.username = ''
    localConfig.password = ''
    localConfig.token = ''
    localConfig.api_key_name = ''
    localConfig.api_key_value = ''
  } else if (localConfig.auth_type === 'basic') {
    localConfig.token = ''
    localConfig.api_key_name = ''
    localConfig.api_key_value = ''
  } else if (localConfig.auth_type === 'bearer') {
    localConfig.username = ''
    localConfig.password = ''
    localConfig.api_key_name = ''
    localConfig.api_key_value = ''
  } else if (localConfig.auth_type === 'apikey') {
    localConfig.username = ''
    localConfig.password = ''
    localConfig.token = ''
  }
  emitChange()
}

const handleHeadersChange = (value) => {
  customHeaders.value = value
  try {
    localConfig.headers = value ? JSON.parse(value) : {}
  } catch (error) {
    localConfig.headers = {}
  }
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
.webhook-config {
  margin-top: 10px;
}

.form-tip {
  margin-top: 5px;
}
</style> 