// import './assets/main.css' // 引入 main.css 样式文件
// createApp(App).mount('#app')
import { createApp } from 'vue'
import App from './App.vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import router from './router'; // 确保路径正确
// 引入 App.vue 组件
// import App from './App.vue' 
// 创建应用，并将 App 根组件挂载到 <div id="#app"></div> 中
// 使用
const app = createApp(App)
app.use(ElementPlus)
app.use(router); 
app.mount('#app')


