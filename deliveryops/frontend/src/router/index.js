import { createRouter, createWebHistory } from 'vue-router';
import Gendownloadurl from '../views/encrypy/Gendownloadurl.vue';
import TestPageView from '../views/test/TestPageView.vue';

const routes = [
  {
    path: '/gendownloadurl',
    name: 'Gendownloadurl',
    component: Gendownloadurl,
  },
  {
    path: '/test',
    name: 'TestPage',
    component: TestPageView,
  },
  {
    path: '/',
    redirect: '/gendownloadurl', // 默认重定向到加解密工具
  },
];

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes,
});

export default router;
