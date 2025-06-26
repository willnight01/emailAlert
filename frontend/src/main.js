import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import { createPinia } from 'pinia'

// Element Plus
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

// 自定义样式
import './assets/styles/index.css'

// Store
import { useUserStore } from './store'

const app = createApp(App)

// 注册所有图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(createPinia())

// 初始化用户状态
const userStore = useUserStore()
userStore.init()

app.use(router)
app.use(ElementPlus)

app.mount('#app') 