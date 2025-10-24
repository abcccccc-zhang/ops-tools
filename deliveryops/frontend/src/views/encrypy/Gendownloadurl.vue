<template>
    <el-container style="height: 100vh;">
      <Sidebar />
      <el-container>
        <el-header>
          <h1 style="color: white;"></h1>
        </el-header>
        <el-main>
          <Gendownloadurl />
        </el-main>
      </el-container>
    </el-container>
  </template>
  
  <script>
  import Gendownloadurl from '../../components/encrypt/Gendownloadurl.vue';
  import Sidebar from '../../components/sidebar/Sidebar.vue';
  
  export default {
    components: {
      Gendownloadurl,
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
  