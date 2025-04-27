<template>
    <el-aside width="200px">
      <el-menu default-active="1">
        <el-menu-item index="1" @click="navigateToEncryptDecrypt">加解密工具</el-menu-item>
        <el-menu-item index="2" @click="navigateToTest">测试页</el-menu-item>
      </el-menu>
    </el-aside>
  </template>
  
  <script>
  export default {
    methods: {
      navigateToEncryptDecrypt() {
        this.$router.push('/encrypt-decrypt'); // 使用 Vue Router 导航到加解密工具页
      },
      navigateToTest() {
        this.$router.push('/test'); // 使用 Vue Router 导航到测试页
      },
    },
  };
  </script>
  