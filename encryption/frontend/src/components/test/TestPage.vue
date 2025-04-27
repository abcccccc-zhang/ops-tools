<template>
    <el-card>
      <h2>测试功能</h2>
      <el-input v-model="testInput" placeholder="输入测试内容" />
      <el-button type="primary" @click="runTest">运行测试</el-button>
      <p v-if="testResult">测试结果: {{ testResult }}</p>
    </el-card>
  </template>
  
  <script>
  export default {
    data() {
      return {
        testInput: '',
        testResult: '',
      };
    },
    methods: {
      runTest() {
        // 这里可以执行测试逻辑
        this.testResult = `测试成功: ${this.testInput}`;
      },
    },
  };
  </script>
  
  <style scoped>
  .el-card {
    padding: 20px;
  }
  </style>
  