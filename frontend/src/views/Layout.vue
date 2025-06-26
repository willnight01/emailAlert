<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside :width="isCollapse ? '64px' : '200px'" class="sidebar">
      <div class="logo">
        <img v-if="isCollapse" src="/logo.png" alt="Logo" class="logo-icon" />
        <template v-else>
          <img src="/logo.png" alt="Logo" class="logo-icon" />
          <span class="logo-text">邮件告警平台</span>
        </template>
      </div>
      
      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapse"
        :unique-opened="true"
        router
        class="sidebar-menu"
      >
        <el-menu-item index="/dashboard">
          <el-icon><Monitor /></el-icon>
          <template #title>仪表盘</template>
        </el-menu-item>
        
        <el-menu-item index="/mailboxes">
          <el-icon><Message /></el-icon>
          <template #title>邮箱管理</template>
        </el-menu-item>
        
        <el-menu-item index="/monitor">
          <el-icon><Monitor /></el-icon>
          <template #title>邮件监控</template>
        </el-menu-item>
        
        <el-menu-item index="/rule-groups">
          <el-icon><Setting /></el-icon>
          <template #title>告警规则</template>
        </el-menu-item>
        
        <el-menu-item index="/channels">
          <el-icon><ChatLineSquare /></el-icon>
          <template #title>通知渠道</template>
        </el-menu-item>
        
        <el-menu-item index="/templates">
          <el-icon><Document /></el-icon>
          <template #title>消息模版</template>
        </el-menu-item>
        
        <el-menu-item index="/alerts">
          <el-icon><Bell /></el-icon>
          <template #title>告警历史</template>
        </el-menu-item>
        
        <el-menu-item index="/notification-logs">
          <el-icon><ChatDotRound /></el-icon>
          <template #title>通知历史</template>
        </el-menu-item>
        
        <el-menu-item index="/system">
          <el-icon><Operation /></el-icon>
          <template #title>系统监控</template>
        </el-menu-item>
        
        <el-menu-item v-if="userStore.role === 'admin'" index="/admin">
          <el-icon><UserFilled /></el-icon>
          <template #title>用户管理</template>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <!-- 主内容区域 -->
    <el-container>
      <!-- 顶部导航 -->
      <el-header class="header">
        <div class="header-left">
          <el-button
            type="text"
            @click="toggleCollapse"
            class="collapse-btn"
          >
            <el-icon>
              <Expand v-if="isCollapse" />
              <Fold v-else />
            </el-icon>
          </el-button>
          
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item>{{ currentPageTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        
        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-avatar :size="32" icon="user" />
              <span class="username">{{ userStore.username || '管理员' }}</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 内容区域 -->
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/store'
import {
  Bell,
  Monitor,
  Message,
  Setting,
  ChatLineSquare,
  Document,
  Operation,
  Expand,
  Fold,
  ChatDotRound,
  UserFilled
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const isCollapse = ref(false)

// 当前激活的菜单
const activeMenu = computed(() => route.path)

// 当前页面标题
const currentPageTitle = computed(() => {
  return route.meta.title || '未知页面'
})

// 切换侧边栏折叠状态
const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value
}

// 处理下拉菜单命令
const handleCommand = async (command) => {
  switch (command) {
    case 'logout':
      try {
        await ElMessageBox.confirm(
          '确定要退出登录吗？',
          '提示',
          {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )
        userStore.logout()
        ElMessage.success('已退出登录')
        router.push('/login')
      } catch {
        // 用户取消
      }
      break
    default:
      break
  }
}

onMounted(() => {
  // 组件挂载时的初始化逻辑
})
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.sidebar {
  background-color: var(--el-menu-bg-color);
  border-right: 1px solid var(--el-border-color);
  transition: width 0.3s;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: bold;
  color: var(--el-color-primary);
  border-bottom: 1px solid var(--el-border-color);
  gap: 8px;
}

.logo-icon {
  width: 32px;
  height: 32px;
  object-fit: contain;
}

.logo-text {
  white-space: nowrap;
}

.sidebar-menu {
  border-right: none;
}

.header {
  background-color: var(--el-bg-color);
  border-bottom: 1px solid var(--el-border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.collapse-btn {
  font-size: 18px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.username {
  font-size: 14px;
  color: var(--el-text-color-primary);
}

.main-content {
  background-color: var(--el-bg-color-page);
  padding: 20px;
}
</style> 