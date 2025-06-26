<template>
  <div class="variable-category">
    <div class="category-header">
      <h3>{{ title }}</h3>
      <p class="category-description">{{ description }}</p>
    </div>
    
    <div class="variable-list">
      <div 
        v-for="variable in variables" 
        :key="variable.name"
        class="variable-item"
      >
        <div class="variable-info">
          <div class="variable-name">
            <code>{{ getVariableTemplate(variable.name) }}</code>
            <el-button 
              size="small" 
              type="primary" 
              link
              @click="handleInsert(variable.name)"
            >
              插入
            </el-button>
          </div>
          <div class="variable-description">{{ variable.description }}</div>
          <div class="variable-example">
            <span class="example-label">示例值：</span>
            <span class="example-value">{{ variable.example }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  title: {
    type: String,
    required: true
  },
  description: {
    type: String,
    required: true
  },
  variables: {
    type: Array,
    required: true
  }
})

const emit = defineEmits(['insert'])

const handleInsert = (variableName) => {
  emit('insert', variableName)
}

const getVariableTemplate = (variableName) => {
  return `{{${variableName}}}`
}
</script>

<style scoped>
.variable-category {
  padding: 20px;
}

.category-header {
  margin-bottom: 24px;
  padding-bottom: 12px;
  border-bottom: 2px solid #e4e7ed;
}

.category-header h3 {
  margin: 0 0 8px 0;
  color: #2c3e50;
  font-weight: 600;
}

.category-description {
  margin: 0;
  color: #6c757d;
  font-size: 14px;
}

.variable-list {
  display: grid;
  gap: 16px;
}

.variable-item {
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  padding: 16px;
  background-color: #fafbfc;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.variable-item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.1);
}

.variable-name {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.variable-name code {
  background-color: #f1f2f6;
  padding: 4px 8px;
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  color: #e74c3c;
  font-weight: 600;
}

.variable-description {
  color: #495057;
  font-size: 14px;
  margin-bottom: 8px;
  line-height: 1.5;
}

.variable-example {
  font-size: 13px;
  color: #6c757d;
}

.example-label {
  color: #868e96;
  font-weight: 500;
}

.example-value {
  color: #28a745;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  background-color: #f8f9fa;
  padding: 2px 4px;
  border-radius: 3px;
  margin-left: 4px;
}
</style> 