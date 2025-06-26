<template>
  <div class="admin-container">
    <div class="page-header">
      <h1>用户管理</h1>
      <p>管理系统用户账号</p>
    </div>

    <div class="content">
      <!-- 操作栏 -->
      <div class="toolbar">
        <el-button type="primary" @click="showCreateDialog">
          <el-icon><Plus /></el-icon>
          添加用户
        </el-button>
        <el-button @click="loadUsers">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>

      <!-- 用户列表 -->
      <el-table 
        :data="users" 
        v-loading="loading"
        style="width: 100%"
        empty-text="暂无用户数据"
      >
        <el-table-column prop="username" label="用户名" width="200" />
        <el-table-column prop="role" label="角色" width="120">
          <template #default="{ row }">
            <el-tag :type="row.role === 'admin' ? 'danger' : 'primary'">
              {{ row.role === 'admin' ? '管理员' : '普通用户' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button 
              type="primary" 
              size="small" 
              @click="showEditDialog(row)"
              :disabled="row.username === currentUser"
            >
              编辑
            </el-button>
            <el-button 
              type="danger" 
              size="small" 
              @click="deleteUser(row.username)"
              :disabled="row.username === currentUser"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 创建/编辑用户对话框 -->
    <el-dialog 
      :title="dialogTitle" 
      v-model="dialogVisible" 
      width="500px"
      @close="resetForm"
    >
      <el-form 
        :model="userForm" 
        :rules="userRules" 
        ref="userFormRef"
        label-width="80px"
      >
        <el-form-item label="用户名" prop="username">
          <el-input 
            v-model="userForm.username" 
            placeholder="请输入用户名"
            :disabled="isEdit"
          />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input 
            v-model="userForm.password" 
            type="password" 
            placeholder="请输入密码"
            show-password
          />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="userForm.role" placeholder="请选择角色" style="width: 100%">
            <el-option label="管理员" value="admin" />
            <el-option label="普通用户" value="user" />
          </el-select>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitForm" :loading="submitting">
            {{ isEdit ? '更新' : '创建' }}
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh } from '@element-plus/icons-vue'
import { useUserStore } from '@/store'
import { usersAPI } from '@/api'

const userStore = useUserStore()

// 响应式数据
const users = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const submitting = ref(false)
const isEdit = ref(false)
const userFormRef = ref()

// 当前用户名
const currentUser = computed(() => userStore.username)

// 对话框标题
const dialogTitle = computed(() => isEdit.value ? '编辑用户' : '添加用户')

// 用户表单数据
const userForm = reactive({
  username: '',
  password: '',
  role: 'user'
})

// 表单验证规则
const userRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 4, max: 50, message: '密码长度在 4 到 50 个字符', trigger: 'blur' }
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ]
}

// 加载用户列表
const loadUsers = async () => {
  loading.value = true
  try {
    const response = await usersAPI.list()
    users.value = response.data || []
  } catch (error) {
    ElMessage.error('加载用户列表失败：' + (error.response?.data?.error || error.message))
  } finally {
    loading.value = false
  }
}

// 显示创建用户对话框
const showCreateDialog = () => {
  isEdit.value = false
  dialogVisible.value = true
  resetForm()
}

// 显示编辑用户对话框
const showEditDialog = (user) => {
  isEdit.value = true
  dialogVisible.value = true
  userForm.username = user.username
  userForm.password = '' // 编辑时密码为空，表示不修改
  userForm.role = user.role
}

// 重置表单
const resetForm = () => {
  if (userFormRef.value) {
    userFormRef.value.resetFields()
  }
  userForm.username = ''
  userForm.password = ''
  userForm.role = 'user'
}

// 提交表单
const submitForm = async () => {
  if (!userFormRef.value) return
  
  const valid = await userFormRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (isEdit.value) {
      // 更新用户
      await usersAPI.update(userForm.username, {
        username: userForm.username,
        password: userForm.password,
        role: userForm.role
      })
      ElMessage.success('用户更新成功')
    } else {
      // 创建用户
      await usersAPI.create({
        username: userForm.username,
        password: userForm.password,
        role: userForm.role
      })
      ElMessage.success('用户创建成功')
    }
    
    dialogVisible.value = false
    loadUsers()
  } catch (error) {
    ElMessage.error((isEdit.value ? '更新' : '创建') + '用户失败：' + (error.response?.data?.error || error.message))
  } finally {
    submitting.value = false
  }
}

// 删除用户
const deleteUser = async (username) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户 "${username}" 吗？此操作不可撤销。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await usersAPI.delete(username)
    ElMessage.success('用户删除成功')
    loadUsers()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除用户失败：' + (error.response?.data?.error || error.message))
    }
  }
}

// 组件挂载时加载数据
onMounted(() => {
  loadUsers()
})
</script>

<style scoped>
.admin-container {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h1 {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.page-header p {
  margin: 0;
  color: var(--el-text-color-regular);
  font-size: 14px;
}

.content {
  background: var(--el-bg-color);
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.toolbar {
  margin-bottom: 20px;
  display: flex;
  gap: 12px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

:deep(.el-table) {
  --el-table-border-color: var(--el-border-color-lighter);
}

:deep(.el-table th) {
  background-color: var(--el-fill-color-light);
}
</style> 