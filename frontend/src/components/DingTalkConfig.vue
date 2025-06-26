<template>
  <div class="dingtalk-config">
    <el-form-item label="推送类型" prop="type">
      <el-radio-group v-model="localConfig.type" @change="handleTypeChange">
        <el-radio label="robot">群机器人</el-radio>
        <el-radio label="work">工作通知</el-radio>
      </el-radio-group>
    </el-form-item>

    <!-- 群机器人配置 -->
    <template v-if="localConfig.type === 'robot'">
      <el-form-item label="Webhook地址" prop="webhook_url">
        <el-input 
          v-model="localConfig.webhook_url" 
          placeholder="请输入群机器人Webhook地址"
          @input="emitChange"
        />
        <div class="form-tip">
          <el-text type="info" size="small">
            在钉钉群中添加自定义机器人，复制Webhook地址
          </el-text>
        </div>
      </el-form-item>
      <el-form-item label="加签密钥" prop="secret">
        <el-input 
          v-model="localConfig.secret" 
          placeholder="请输入加签密钥（可选）"
          type="password"
          show-password
          @input="emitChange"
        />
        <div class="form-tip">
          <el-text type="info" size="small">
            如果设置了加签安全设置，请输入密钥
          </el-text>
        </div>
      </el-form-item>
    </template>

    <!-- 工作通知配置 -->
    <template v-if="localConfig.type === 'work'">
      <el-form-item label="AppKey" prop="app_key">
        <el-input 
          v-model="localConfig.app_key" 
          placeholder="请输入应用AppKey"
          @input="emitChange"
        />
      </el-form-item>
      <el-form-item label="AppSecret" prop="app_secret">
        <el-input 
          v-model="localConfig.app_secret" 
          placeholder="请输入应用AppSecret"
          type="password"
          show-password
          @input="emitChange"
        />
      </el-form-item>
      <el-form-item label="AgentId" prop="agent_id">
        <el-input-number 
          v-model="localConfig.agent_id" 
          placeholder="请输入应用AgentId"
          style="width: 100%"
          @change="emitChange"
        />
      </el-form-item>
      <el-form-item label="接收用户" prop="user_ids">
        <el-input 
          v-model="localConfig.user_ids" 
          placeholder="请输入接收用户ID，多个用逗号分隔"
          @input="emitChange"
        />
        <div class="form-tip">
          <el-text type="info" size="small">
            用户ID，多个用逗号分隔，如：userid1,userid2
          </el-text>
        </div>
      </el-form-item>
      <el-form-item label="接收部门" prop="dept_ids">
        <el-input 
          v-model="localConfig.dept_ids" 
          placeholder="请输入接收部门ID，多个用逗号分隔（可选）"
          @input="emitChange"
        />
      </el-form-item>
      <el-form-item label="发送全员" prop="to_all_user">
        <el-switch 
          v-model="localConfig.to_all_user" 
          @change="emitChange"
        />
        <div class="form-tip">
          <el-text type="info" size="small">
            开启后将发送给企业全部用户
          </el-text>
        </div>
      </el-form-item>
    </template>

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
const localConfig = reactive({
  type: 'robot',
  webhook_url: '',
  secret: '',
  app_key: '',
  app_secret: '',
  agent_id: null,
  user_ids: '',
  dept_ids: '',
  to_all_user: false
})

// 监听props变化
watch(() => props.modelValue, (newVal) => {
  if (newVal && typeof newVal === 'object') {
    Object.assign(localConfig, {
      type: 'robot',
      webhook_url: '',
      secret: '',
      app_key: '',
      app_secret: '',
      agent_id: null,
      user_ids: '',
      dept_ids: '',
      to_all_user: false,
      ...newVal
    })
  }
}, { immediate: true })

const emitChange = () => {
  emit('update:modelValue', { ...localConfig })
}

const handleTypeChange = () => {
  // 切换类型时清空其他字段
  if (localConfig.type === 'robot') {
    localConfig.app_key = ''
    localConfig.app_secret = ''
    localConfig.agent_id = null
    localConfig.user_ids = ''
    localConfig.dept_ids = ''
    localConfig.to_all_user = false
  } else {
    localConfig.webhook_url = ''
    localConfig.secret = ''
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
.dingtalk-config {
  margin-top: 10px;
}

.form-tip {
  margin-top: 5px;
}
</style> 