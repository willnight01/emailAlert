<template>
  <div class="wechat-config">
    <el-form-item label="推送类型" prop="type">
      <el-radio-group v-model="localConfig.type" @change="handleTypeChange">
        <el-radio label="robot">群机器人</el-radio>
        <el-radio label="app">应用消息</el-radio>
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
            在企业微信群中添加机器人，复制Webhook地址
          </el-text>
        </div>
      </el-form-item>
      <el-form-item label="Key" prop="key">
        <el-input 
          v-model="localConfig.key" 
          placeholder="请输入机器人Key（可选）"
          @input="emitChange"
        />
        <div class="form-tip">
          <el-text type="info" size="small">
            从Webhook地址中提取的key参数
          </el-text>
        </div>
      </el-form-item>
    </template>

    <!-- 应用消息配置 -->
    <template v-if="localConfig.type === 'app'">
      <el-form-item label="企业ID" prop="corp_id">
        <el-input 
          v-model="localConfig.corp_id" 
          placeholder="请输入企业ID"
          @input="emitChange"
        />
      </el-form-item>
      <el-form-item label="应用ID" prop="agent_id">
        <el-input-number 
          v-model="localConfig.agent_id" 
          placeholder="请输入应用ID"
          style="width: 100%"
          @change="emitChange"
        />
      </el-form-item>
      <el-form-item label="应用Secret" prop="secret">
        <el-input 
          v-model="localConfig.secret" 
          placeholder="请输入应用Secret"
          type="password"
          show-password
          @input="emitChange"
        />
      </el-form-item>
      <el-form-item label="接收人" prop="to_user">
        <el-input 
          v-model="localConfig.to_user" 
          placeholder="请输入接收人用户ID，多个用|分隔"
          @input="emitChange"
        />
        <div class="form-tip">
          <el-text type="info" size="small">
            用户ID，多个用|分隔，如：userid1|userid2
          </el-text>
        </div>
      </el-form-item>
      <el-form-item label="接收部门" prop="to_party">
        <el-input 
          v-model="localConfig.to_party" 
          placeholder="请输入接收部门ID，多个用|分隔（可选）"
          @input="emitChange"
        />
      </el-form-item>
      <el-form-item label="接收标签" prop="to_tag">
        <el-input 
          v-model="localConfig.to_tag" 
          placeholder="请输入接收标签ID，多个用|分隔（可选）"
          @input="emitChange"
        />
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
  key: '',
  corp_id: '',
  agent_id: null,
  secret: '',
  to_user: '',
  to_party: '',
  to_tag: ''
})

// 监听props变化
watch(() => props.modelValue, (newVal) => {
  if (newVal && typeof newVal === 'object') {
    Object.assign(localConfig, {
      type: 'robot',
      webhook_url: '',
      key: '',
      corp_id: '',
      agent_id: null,
      secret: '',
      to_user: '',
      to_party: '',
      to_tag: '',
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
    localConfig.corp_id = ''
    localConfig.agent_id = null
    localConfig.secret = ''
    localConfig.to_user = ''
    localConfig.to_party = ''
    localConfig.to_tag = ''
  } else {
    localConfig.webhook_url = ''
    localConfig.key = ''
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
.wechat-config {
  margin-top: 10px;
}

.form-tip {
  margin-top: 5px;
}
</style> 