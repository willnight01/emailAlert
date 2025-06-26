<template>
  <div class="rule-condition-form">
    <div class="condition-header">
      <span class="condition-title">{{ title }}</span>
      <el-button v-if="showDelete" type="danger" size="small" text @click="handleDelete">
        <el-icon><Delete /></el-icon>
        删除条件
      </el-button>
    </div>

    <el-form
      :model="condition"
      :rules="formRules"
      label-width="120px"
      size="small"
    >
      <el-row :gutter="20">
        <el-col :span="6">
          <el-form-item label="匹配字段" prop="field_type">
            <el-select 
              v-model="condition.field_type" 
              placeholder="请选择字段" 
              style="width: 100%"
              @change="handleFieldChange"
            >
              <el-option
                v-for="option in fieldTypeOptions"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              >
                <div class="option-item">
                  <span>{{ option.label }}</span>
                  <span class="option-desc">{{ option.description }}</span>
                </div>
              </el-option>
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item label="匹配类型" prop="match_type">
            <el-select 
              v-model="condition.match_type" 
              placeholder="请选择类型" 
              style="width: 100%"
              @change="handleMatchTypeChange"
            >
              <el-option
                v-for="option in matchTypeOptions"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              >
                <div class="option-item">
                  <span>{{ option.label }}</span>
                  <span class="option-desc">{{ option.description }}</span>
                </div>
              </el-option>
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="5">
          <el-form-item label="关键词逻辑" prop="keyword_logic">
            <el-select 
              v-model="condition.keyword_logic" 
              placeholder="请选择" 
              style="width: 100%"
            >
              <el-option label="AND (全部)" value="and">
                <div class="option-item">
                  <span>AND (全部)</span>
                  <span class="option-desc">所有关键词都必须匹配</span>
                </div>
              </el-option>
              <el-option label="OR (任一)" value="or">
                <div class="option-item">
                  <span>OR (任一)</span>
                  <span class="option-desc">任意一个关键词匹配即可</span>
                </div>
              </el-option>
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="4">
          <el-form-item label="优先级" prop="priority">
            <el-input-number
              v-model="condition.priority"
              :min="1"
              :max="10"
              style="width: 100%"
              placeholder="1-10"
            />
          </el-form-item>
        </el-col>
        <el-col :span="3">
          <el-form-item label="状态">
            <el-switch
              v-model="condition.status"
              active-value="active"
              inactive-value="inactive"
              active-text="启用"
              inactive-text="停用"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item label="关键词" prop="keywords">
        <div class="keywords-input">
          <el-input
            v-model="condition.keywords"
            type="textarea"
            :rows="3"
            :placeholder="keywordsPlaceholder"
            @input="handleKeywordsChange"
          />
          <div class="keywords-help">
            <div class="help-title">关键词输入说明：</div>
            <ul class="help-list">
              <li>多个关键词用<strong>逗号</strong>分隔，如：错误,异常,失败</li>
              <li>支持<strong>空格</strong>作为分隔符，如：错误 异常 失败</li>
              <li>关键词匹配逻辑由上方的"关键词逻辑"控制</li>
              <li v-if="condition.match_type === 'regex'">正则表达式模式：输入一个完整的正则表达式</li>
            </ul>
          </div>
        </div>
      </el-form-item>

      <el-form-item label="条件描述">
        <el-input
          v-model="condition.description"
          placeholder="请输入条件描述（可选）"
          maxlength="200"
          show-word-limit
        />
      </el-form-item>

      <!-- 实时预览 -->
      <el-form-item label="匹配预览">
        <div class="match-preview">
          <div class="preview-content">
            <el-tag v-if="condition.field_type" type="info" size="small">
              {{ getFieldTypeText(condition.field_type) }}
            </el-tag>
            <el-tag v-if="condition.match_type" type="" size="small">
              {{ getMatchTypeText(condition.match_type) }}
            </el-tag>
            <span class="preview-text">{{ previewText }}</span>
          </div>
        </div>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup>
import { computed, defineProps, defineEmits } from 'vue'
import { Delete } from '@element-plus/icons-vue'

// 属性定义
const props = defineProps({
  condition: {
    type: Object,
    required: true
  },
  title: {
    type: String,
    default: '匹配条件'
  },
  showDelete: {
    type: Boolean,
    default: true
  },
  fieldTypeOptions: {
    type: Array,
    default: () => []
  },
  matchTypeOptions: {
    type: Array,
    default: () => []
  }
})

// 事件定义
const emit = defineEmits(['delete', 'change'])

// 表单验证规则
const formRules = {
  field_type: [
    { required: true, message: '请选择匹配字段', trigger: 'change' }
  ],
  match_type: [
    { required: true, message: '请选择匹配类型', trigger: 'change' }
  ],
  keywords: [
    { required: true, message: '请输入关键词', trigger: 'blur' },
    { min: 1, message: '关键词不能为空', trigger: 'blur' }
  ],
  keyword_logic: [
    { required: true, message: '请选择关键词逻辑', trigger: 'change' }
  ]
}

// 计算属性
const keywordsPlaceholder = computed(() => {
  if (props.condition.match_type === 'regex') {
    return '请输入正则表达式，如：^ERROR.*database.*$'
  }
  return '请输入关键词，多个关键词用逗号分隔，如：错误,异常,失败'
})

const previewText = computed(() => {
  const { field_type, match_type, keywords, keyword_logic } = props.condition
  
  if (!field_type || !match_type || !keywords) {
    return '请完善条件配置'
  }

  const fieldText = getFieldTypeText(field_type)
  const matchText = getMatchTypeText(match_type)
  
  if (match_type === 'regex') {
    return `当${fieldText}${matchText}："${keywords}"`
  }

  const keywordList = keywords.split(/[,，\s]+/).filter(k => k.trim())
  if (keywordList.length === 0) {
    return '请输入关键词'
  }

  const logicText = keyword_logic === 'and' ? '同时包含' : '包含任一'
  const keywordText = keywordList.slice(0, 3).join('、') + 
    (keywordList.length > 3 ? `等${keywordList.length}个关键词` : '')

  return `当${fieldText}${logicText}："${keywordText}"`
})

// 方法
const getFieldTypeText = (type) => {
  const map = {
    subject: '邮件主题',
    from: '发件人',
    to: '收件人',
    cc: '抄送人',
    body: '邮件正文',
    attachment_name: '附件名称'
  }
  return map[type] || type
}

const getMatchTypeText = (type) => {
  const map = {
    equals: '完全匹配',
    contains: '包含匹配',
    startsWith: '前缀匹配',
    endsWith: '后缀匹配',
    regex: '正则匹配',
    notContains: '不包含'
  }
  return map[type] || type
}

const handleDelete = () => {
  emit('delete')
}

const handleFieldChange = () => {
  emit('change', 'field_type', props.condition.field_type)
}

const handleMatchTypeChange = () => {
  emit('change', 'match_type', props.condition.match_type)
}

const handleKeywordsChange = () => {
  emit('change', 'keywords', props.condition.keywords)
}
</script>

<style scoped>
.rule-condition-form {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 16px;
  background-color: #fafbfc;
  transition: all 0.3s ease;
}

.rule-condition-form:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.1);
}

.condition-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e4e7ed;
}

.condition-title {
  font-weight: 600;
  color: #409eff;
  font-size: 16px;
}

.option-item {
  display: flex;
  flex-direction: column;
}

.option-desc {
  font-size: 12px;
  color: #999;
  margin-top: 2px;
}

.keywords-input {
  width: 100%;
}

.keywords-help {
  margin-top: 8px;
  padding: 12px;
  background-color: #f0f9ff;
  border-radius: 6px;
  border-left: 4px solid #409eff;
}

.help-title {
  font-weight: 600;
  color: #409eff;
  margin-bottom: 8px;
  font-size: 14px;
}

.help-list {
  margin: 0;
  padding-left: 20px;
}

.help-list li {
  margin-bottom: 4px;
  color: #666;
  font-size: 13px;
  line-height: 1.4;
}

.help-list strong {
  color: #409eff;
}

.match-preview {
  padding: 12px;
  background-color: #f5f7fa;
  border-radius: 6px;
  border: 1px solid #e4e7ed;
}

.preview-content {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.preview-text {
  color: #606266;
  font-size: 14px;
  font-weight: 500;
}
</style> 