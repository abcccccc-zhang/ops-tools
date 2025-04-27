import { createRouter, createWebHistory } from 'vue-router';
import EncryptDecryptView from '../views/encrypy/EncryptDecryptView.vue';
import TestPageView from '../views/test/TestPageView.vue';

const routes = [
  {
    path: '/encrypt-decrypt',
    name: 'EncryptDecrypt',
    component: EncryptDecryptView,
  },
  {
    path: '/test',
    name: 'TestPage',
    component: TestPageView,
  },
  {
    path: '/',
    redirect: '/encrypt-decrypt', // 默认重定向到加解密工具
  },
];

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes,
});

export default router;
