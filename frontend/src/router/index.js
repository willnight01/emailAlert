import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/store'
import Layout from '@/views/Layout.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { 
      title: '登录',
      requiresAuth: false
    }
  },
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/',
    component: Layout,
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '仪表盘', requiresAuth: true }
      },
      {
        path: 'mailboxes',
        name: 'Mailboxes',
        component: () => import('@/views/Mailboxes.vue'),
        meta: { title: '邮箱管理', requiresAuth: true }
      },
      {
        path: 'monitor',
        name: 'Monitor',
        component: () => import('@/views/Monitor.vue'),
        meta: { title: '邮件监控', requiresAuth: true }
      },
      {
        path: 'rule-groups',
        name: 'RuleGroups',
        component: () => import('@/views/AlertRules.vue'),
        meta: { title: '告警规则', requiresAuth: true }
      },
      {
        path: 'channels',
        name: 'Channels',
        component: () => import('@/views/Channels.vue'),
        meta: { title: '通知渠道', requiresAuth: true }
      },
      {
        path: 'templates',
        name: 'Templates',
        component: () => import('@/views/Templates.vue'),
        meta: { title: '消息模版', requiresAuth: true }
      },
      {
        path: 'alerts',
        name: 'Alerts',
        component: () => import('@/views/Alerts.vue'),
        meta: { title: '告警历史', requiresAuth: true }
      },
      {
        path: 'notification-logs',
        name: 'NotificationLogs',
        component: () => import('@/views/NotificationLogs.vue'),
        meta: { title: '通知历史', requiresAuth: true }
      },
      {
        path: 'system',
        name: 'System',
        component: () => import('@/views/System.vue'),
        meta: { title: '系统监控', requiresAuth: true }
      },
      {
        path: 'admin',
        name: 'Admin',
        component: () => import('@/views/Admin.vue'),
        meta: { title: '用户管理', requiresAuth: true, requiresAdmin: true }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  userStore.init() // 初始化用户状态
  
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth !== false)
  const requiresAdmin = to.matched.some(record => record.meta.requiresAdmin === true)
  
  if (requiresAuth && !userStore.isLoggedIn) {
    // 需要认证但未登录，跳转到登录页
    next('/login')
  } else if (to.path === '/login' && userStore.isLoggedIn) {
    // 已登录用户访问登录页，跳转到首页
    next('/dashboard')
  } else if (requiresAdmin && userStore.role !== 'admin') {
    // 需要管理员权限但当前用户不是管理员
    next('/dashboard')
  } else {
    next()
  }
})

export default router 