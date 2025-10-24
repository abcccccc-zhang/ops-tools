<template>
    <el-container style="height: 100vh;">
      <Sidebar />
      <el-container>
        <el-header>
          <h1 style="color: white;">测试页</h1>
        </el-header>
        <el-main>
          <TestPage />
        </el-main>
      </el-container>
    </el-container>
  </template>
  
  <script>
  import TestPage from '../../components/test/TestPage.vue';
  import Sidebar from '../../components/sidebar/Sidebar.vue';
  
  export default {
    components: {
      TestPage,
      Sidebar,
    },
  };
  </script>
  
  <style scoped>
  .el-header {
    background-color: #409eff;
    padding: 20px;
    text-align: center;
  }
  </style>
  