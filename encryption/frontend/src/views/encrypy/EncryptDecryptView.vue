<template>
    <el-container style="height: 100vh;">
      <Sidebar />
      <el-container>
        <el-header>
          <h1 style="color: white;">加密和解密工具</h1>
        </el-header>
        <el-main>
          <EncryptDecrypt />
        </el-main>
      </el-container>
    </el-container>
  </template>
  
  <script>
  import EncryptDecrypt from '../../components/encrypt/EncryptDecrypt.vue';
  import Sidebar from '../../components/sidebar/Sidebar.vue';
  
  export default {
    components: {
      EncryptDecrypt,
      Sidebar,
    },
  };
  </script>
  
  <style scoped>
  .el-header {
    background-color: #409eff;
    padding: 50px;
    text-align: center;
  }
  </style>
  