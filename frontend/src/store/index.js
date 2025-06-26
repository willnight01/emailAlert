import { defineStore } from 'pinia'

// 应用主状态
export const useAppStore = defineStore('app', {
  state: () => ({
    sidebarCollapsed: false
  }),
  
  actions: {
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed
    }
  }
})

// 用户状态管理
export const useUserStore = defineStore('user', {
  state: () => ({
    user: JSON.parse(localStorage.getItem('user') || 'null'),
    isLoggedIn: false
  }),

  getters: {
    username: (state) => state.user?.username || '',
    role: (state) => state.user?.role || '',
    hasSession: (state) => !!state.user
  },

  actions: {
    // 初始化用户状态
    init() {
      this.isLoggedIn = !!this.user
    },

    // 登录
    login(loginData) {
      const { username, role } = loginData
      this.user = { username, role }
      this.isLoggedIn = true
      
      localStorage.setItem('user', JSON.stringify({ username, role }))
    },

    // 设置用户信息
    setUser(user) {
      this.user = user
      this.isLoggedIn = true
      
      localStorage.setItem('user', JSON.stringify(user))
    },

    // 登出
    logout() {
      this.user = null
      this.isLoggedIn = false
      
      localStorage.removeItem('user')
    }
  }
})

// 邮箱状态管理
export const useMailboxStore = defineStore('mailbox', {
  state: () => ({
    mailboxes: [],
    currentMailbox: null,
    loading: false
  }),
  
  getters: {
    activeMailboxes: (state) => state.mailboxes.filter(m => m.status === 'active'),
    inactiveMailboxes: (state) => state.mailboxes.filter(m => m.status === 'inactive')
  },
  
  actions: {
    setMailboxes(mailboxes) {
      this.mailboxes = mailboxes
    },
    
    addMailbox(mailbox) {
      this.mailboxes.push(mailbox)
    },
    
    updateMailbox(id, updates) {
      const index = this.mailboxes.findIndex(m => m.id === id)
      if (index !== -1) {
        this.mailboxes[index] = { ...this.mailboxes[index], ...updates }
      }
    },
    
    removeMailbox(id) {
      this.mailboxes = this.mailboxes.filter(m => m.id !== id)
    },
    
    setCurrentMailbox(mailbox) {
      this.currentMailbox = mailbox
    },
    
    setLoading(loading) {
      this.loading = loading
    }
  }
})

export default {
  useUserStore,
  useMailboxStore
} 