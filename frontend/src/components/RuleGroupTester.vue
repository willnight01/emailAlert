<template>
  <div class="rule-group-tester">
    <el-card>
      <template #header>
        <div class="tester-header">
          <span>规则组匹配测试</span>
          <el-button type="primary" @click="runTest" :loading="testing">
            <el-icon><PlayIcon /></el-icon>
            执行测试
          </el-button>
        </div>
      </template>

      <div class="tester-content">
        <!-- 测试邮件输入 -->
        <el-card class="test-email-section">
          <template #header>
            <span>测试邮件数据</span>
          </template>
          
          <el-form :model="testEmail" label-width="120px" size="default">
            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item label="邮件主题" required>
                  <el-input
                    v-model="testEmail.subject"
                    placeholder="请输入测试邮件主题"
                    maxlength="200"
                    show-word-limit
                  />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="发件人邮箱" required>
                  <el-input
                    v-model="testEmail.from"
                    placeholder="请输入发件人邮箱"
                    maxlength="100"
                  />
                </el-form-item>
              </el-col>
            </el-row>
            
            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item label="收件人邮箱">
                  <el-input
                    v-model="testEmail.to"
                    placeholder="请输入收件人邮箱"
                    maxlength="100"
                  />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="抄送人邮箱">
                  <el-input
                    v-model="testEmail.cc"
                    placeholder="请输入抄送人邮箱（可选）"
                    maxlength="100"
                  />
                </el-form-item>
              </el-col>
            </el-row>

            <el-form-item label="邮件正文" required>
              <el-input
                v-model="testEmail.body"
                type="textarea"
                :rows="6"
                placeholder="请输入测试邮件正文内容"
                maxlength="2000"
                show-word-limit
              />
            </el-form-item>

            <el-form-item label="附件名称">
              <el-input
                v-model="testEmail.attachment_name"
                placeholder="请输入附件名称（可选）"
                maxlength="100"
              />
            </el-form-item>
          </el-form>

          <!-- 快速填充示例 -->
          <div class="quick-fill">
            <span class="quick-fill-label">快速填充示例：</span>
            <el-button-group>
              <el-button size="small" @click="fillErrorExample">错误告警</el-button>
              <el-button size="small" @click="fillWarningExample">警告通知</el-button>
              <el-button size="small" @click="fillInfoExample">信息通知</el-button>
              <el-button size="small" @click="clearTestData">清空</el-button>
            </el-button-group>
          </div>
        </el-card>

        <!-- 测试结果 -->
        <el-card v-if="testResult" class="test-result-section">
          <template #header>
            <div class="result-header">
              <span>测试结果</span>
              <el-tag :type="testResult.matched ? 'success' : 'info'" size="large">
                {{ testResult.matched ? '✓ 匹配成功' : '✗ 匹配失败' }}
              </el-tag>
            </div>
          </template>

          <!-- 整体匹配结果 -->
          <div class="overall-result">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="规则组名称">
                {{ ruleGroup?.name || '未命名规则组' }}
              </el-descriptions-item>
              <el-descriptions-item label="条件逻辑">
                <el-tag :type="ruleGroup?.logic === 'and' ? 'success' : 'warning'">
                  {{ ruleGroup?.logic === 'and' ? 'AND (全部满足)' : 'OR (任一满足)' }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="匹配结果">
                <el-tag :type="testResult.matched ? 'success' : 'info'">
                  {{ testResult.matched ? '匹配成功' : '匹配失败' }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="匹配说明">
                {{ testResult.reason || '无说明' }}
              </el-descriptions-item>
            </el-descriptions>
          </div>

          <!-- 详细条件匹配结果 -->
          <el-divider content-position="left">条件匹配详情</el-divider>
          <div class="condition-results">
            <div
              v-for="(conditionResult, index) in testResult.condition_results"
              :key="index"
              class="condition-result-item"
            >
              <el-card>
                <template #header>
                  <div class="condition-result-header">
                    <span>条件 {{ index + 1 }}</span>
                    <el-tag :type="conditionResult.matched ? 'success' : 'danger'" size="small">
                      {{ conditionResult.matched ? '✓ 匹配' : '✗ 不匹配' }}
                    </el-tag>
                  </div>
                </template>

                <div class="condition-result-content">
                  <el-row :gutter="20">
                    <el-col :span="12">
                      <div class="condition-info">
                        <div class="info-item">
                          <span class="info-label">匹配字段：</span>
                          <el-tag type="info" size="small">
                            {{ getFieldTypeText(conditionResult.condition?.field_type) }}
                          </el-tag>
                        </div>
                        <div class="info-item">
                          <span class="info-label">匹配类型：</span>
                          <el-tag type="" size="small">
                            {{ getMatchTypeText(conditionResult.condition?.match_type) }}
                          </el-tag>
                        </div>
                        <div class="info-item">
                          <span class="info-label">关键词：</span>
                          <span class="keywords-text">{{ conditionResult.condition?.keywords }}</span>
                        </div>
                        <div class="info-item">
                          <span class="info-label">关键词逻辑：</span>
                          <el-tag 
                            :type="conditionResult.condition?.keyword_logic === 'and' ? 'success' : 'warning'" 
                            size="small"
                          >
                            {{ conditionResult.condition?.keyword_logic?.toUpperCase() }}
                          </el-tag>
                        </div>
                      </div>
                    </el-col>
                    <el-col :span="12">
                      <div class="match-info">
                        <div class="info-item">
                          <span class="info-label">字段内容：</span>
                          <div class="field-content">{{ conditionResult.field_content || '无内容' }}</div>
                        </div>
                        <div class="info-item">
                          <span class="info-label">匹配关键词：</span>
                          <div class="matched-keywords">
                            <el-tag
                              v-for="keyword in conditionResult.matched_keywords"
                              :key="keyword"
                              type="success"
                              size="small"
                              style="margin-right: 4px; margin-bottom: 4px;"
                            >
                              {{ keyword }}
                            </el-tag>
                            <span v-if="conditionResult.matched_keywords?.length === 0" class="no-match">
                              无匹配关键词
                            </span>
                          </div>
                        </div>
                        <div class="info-item">
                          <span class="info-label">匹配说明：</span>
                          <div class="match-reason">{{ conditionResult.reason || '无说明' }}</div>
                        </div>
                      </div>
                    </el-col>
                  </el-row>
                </div>
              </el-card>
            </div>
          </div>
        </el-card>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, defineProps } from 'vue'
import { ElMessage } from 'element-plus'
import { CaretRight as PlayIcon } from '@element-plus/icons-vue'
import { ruleGroupsAPI } from '@/api'

// 属性定义
const props = defineProps({
  ruleGroup: {
    type: Object,
    default: () => ({})
  },
  conditions: {
    type: Array,
    default: () => []
  }
})

// 响应式数据
const testing = ref(false)
const testResult = ref(null)

// 测试邮件数据
const testEmail = reactive({
  subject: '',
  from: '',
  to: '',
  cc: '',
  body: '',
  attachment_name: ''
})

// 方法
const runTest = async () => {
  // 验证必填字段
  if (!testEmail.subject || !testEmail.from || !testEmail.body) {
    ElMessage.warning('请填写邮件主题、发件人和正文')
    return
  }

  testing.value = true
  try {
    const requestData = {
      rule_group_data: {
        rule_group: props.ruleGroup,
        conditions: props.conditions
      },
      test_email: testEmail
    }

    const response = await ruleGroupsAPI.test(requestData)
    testResult.value = response.data
    
    ElMessage.success('测试完成')
  } catch (error) {
    console.error('测试规则组失败:', error)
    ElMessage.error('测试失败，请检查规则组配置')
  } finally {
    testing.value = false
  }
}

const fillErrorExample = () => {
  Object.assign(testEmail, {
    subject: 'CRITICAL ERROR: Database Connection Failed',
    from: 'system@example.com',
    to: 'admin@example.com',
    cc: 'ops@example.com',
    body: `系统出现严重错误！

错误信息：数据库连接失败
错误代码：DB_CONNECTION_ERROR_001
发生时间：2024-12-14 15:30:45
影响范围：整个应用系统

请立即检查数据库服务器状态并进行修复。

详细错误日志：
- Connection timeout after 30 seconds
- Unable to connect to database server 192.168.1.100:3306
- Max connection pool size reached

紧急联系人：张三 13800138000`,
    attachment_name: 'error_log_20241214.txt'
  })
}

const fillWarningExample = () => {
  Object.assign(testEmail, {
    subject: 'WARNING: High Memory Usage Alert',
    from: 'monitor@example.com',
    to: 'devops@example.com',
    cc: '',
    body: `系统监控告警通知

告警类型：内存使用率过高
当前使用率：85%
告警阈值：80%
服务器：web-server-01

建议立即检查并清理内存使用情况，避免系统性能下降。

监控数据：
- 可用内存：2.4GB / 16GB
- 内存使用趋势：持续上升
- 最高使用率：87%（10分钟前）

请及时处理！`,
    attachment_name: ''
  })
}

const fillInfoExample = () => {
  Object.assign(testEmail, {
    subject: 'INFO: Daily Backup Completed Successfully',
    from: 'backup@example.com',
    to: 'admin@example.com',
    cc: 'backup-team@example.com',
    body: `每日备份任务执行完成

备份时间：2024-12-14 02:00:00 - 02:35:42
备份状态：成功
备份大小：15.6 GB
备份文件：backup_20241214_020000.tar.gz

备份详情：
- 数据库备份：成功 (8.2GB)
- 文件系统备份：成功 (7.4GB)
- 配置文件备份：成功 (0.02GB)
- 备份验证：通过

备份文件已安全存储到远程服务器。`,
    attachment_name: 'backup_report_20241214.pdf'
  })
}

const clearTestData = () => {
  Object.assign(testEmail, {
    subject: '',
    from: '',
    to: '',
    cc: '',
    body: '',
    attachment_name: ''
  })
  testResult.value = null
}

// 工具函数
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
</script>

<style scoped>
.rule-group-tester {
  padding: 20px;
}

.tester-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.tester-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.test-email-section,
.test-result-section {
  width: 100%;
}

.quick-fill {
  margin-top: 16px;
  padding: 12px;
  background-color: #f5f7fa;
  border-radius: 6px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.quick-fill-label {
  font-weight: 500;
  color: #606266;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.overall-result {
  margin-bottom: 20px;
}

.condition-results {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.condition-result-item {
  width: 100%;
}

.condition-result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.condition-result-content {
  margin-top: 16px;
}

.condition-info,
.match-info {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.info-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.info-label {
  font-weight: 500;
  color: #606266;
  min-width: 80px;
  flex-shrink: 0;
}

.keywords-text {
  color: #409eff;
  font-family: monospace;
  background-color: #f0f9ff;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 13px;
}

.field-content {
  background-color: #f5f7fa;
  padding: 8px;
  border-radius: 4px;
  font-size: 13px;
  line-height: 1.4;
  max-height: 100px;
  overflow-y: auto;
  word-break: break-all;
}

.matched-keywords {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  align-items: center;
}

.no-match {
  color: #909399;
  font-style: italic;
  font-size: 13px;
}

.match-reason {
  color: #606266;
  font-size: 13px;
  line-height: 1.4;
  background-color: #f0f9ff;
  padding: 6px 8px;
  border-radius: 4px;
  border-left: 3px solid #409eff;
}
</style> 