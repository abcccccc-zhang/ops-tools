<!-- <template>
  <el-container style="height: 100vh;">
    <el-aside width="200px">
      <el-menu default-active="1">
        <el-menu-item index="1" @click="navigateToEncryptDecrypt">加解密工具</el-menu-item>
      </el-menu>
    </el-aside>
    
    <el-container>
      <el-header>
        <h1 style="color: white;">加密和解密工具</h1>
      </el-header>
      <el-main>
        <el-card>
          <h2>加密</h2>
          <el-form :model="encryptForm" ref="encryptForm" status-icon>
            <el-form-item label="消息" prop="encryptMsg">
              <el-input v-model="encryptMsg" placeholder="输入要加密的消息" />
            </el-form-item>
            <el-form-item label="加密密钥" prop="encryptKey">
              <el-input v-model="encryptKey" placeholder="输入加密密钥" />
            </el-form-item>
            <div>
              <el-button type="primary" @click="validateEncrypt">加密</el-button>
              <el-button type="default" @click="resetEncrypt">重置</el-button>
            </div>
          </el-form>
          <p v-if="encryptedString">加密结果: {{ encryptedString }}</p>
          <p v-if="response">{{ response }}</p>
        </el-card>

        <el-card style="margin-top: 20px;">
          <h2>解密</h2>
          <el-form :model="decryptForm" ref="decryptForm" status-icon>
            <el-form-item label="已加密的消息" prop="decryptString">
              <el-input v-model="decryptString" placeholder="输入已加密的消息" />
            </el-form-item>
            <el-form-item label="解密密钥" prop="decryptKey">
              <el-input v-model="decryptKey" placeholder="输入解密密钥" />
            </el-form-item>
            <div>
              <el-button type="primary" @click="validateDecrypt">解密</el-button>
              <el-button type="default" @click="resetDecrypt">重置</el-button>
            </div>
          </el-form>
          <p v-if="decryptedMsg">解密结果: {{ decryptedMsg }}</p>
        </el-card>
      </el-main>
    </el-container>
  </el-container>
</template>

<script>
import axios from 'axios';

export default {
  data() {
    return {
      response: '',
      encryptMsg: '',
      encryptKey: '',
      encryptedString: '',
      decryptString: '',
      decryptKey: '',
      decryptedMsg: '',
      encryptForm: {},
      decryptForm: {},
    };
  },
  methods: {
    async validateEncrypt() {
      this.$refs.encryptForm.validate(async (valid) => {
        if (valid) {
          await this.encrypt();
        } else {
          alert('请确保输入的所有字段都是有效的！');
        }
      });
    },
    async validateDecrypt() {
      this.$refs.decryptForm.validate(async (valid) => {
        if (valid) {
          await this.decrypt();
        } else {
          alert('请确保输入的所有字段都是有效的！');
        }
      });
    },
    async encrypt() {
      try {
        const response = await axios.post('http://localhost:8080/api/encrypt', {
          msg: this.encryptMsg,
          encryption_key: this.encryptKey,
        });
        this.encryptedString = response.data.encrypted_string;
      } catch (error) {
        if (error.response) {
          console.error("加密失败:", error.response.data.error);
          alert("加密失败: " + error.response.data.error);
        } else if (error.request) {
          console.error("未收到响应:", error.request);
          this.response = "未收到响应";
        } else {
          console.error("请求发生错误:", error.message);
          alert("请求发生错误: " + error.message);
        }
      }
    },
    async decrypt() {
      try {
        const response = await axios.post('http://192.168.7.134:8080/api/decrypt', {
          encrypted_string: this.decryptString,
          encryption_key: this.decryptKey,
        });
        this.decryptedMsg = response.data.decrypted_msg;
      } catch (error) {
        console.error("解密失败:", error.response.data.error);
        alert("解密失败: " + error.response.data.error);
      }
    },
    resetEncrypt() {
      this.encryptMsg = '';
      this.encryptKey = '';
      this.encryptedString = '';
      this.response = '';
      this.$refs.encryptForm.resetFields(); // 重置表单字段
    },
    resetDecrypt() {
      this.decryptString = '';
      this.decryptKey = '';
      this.decryptedMsg = '';
      this.$refs.decryptForm.resetFields(); // 重置表单字段
    },
    navigateToEncryptDecrypt() {
      // 这里可以处理页面切换逻辑
      console.log("导航到加解密工具");
    },
  },
};
</script>

<style scoped>
.el-header {
  background-color: #409eff;
  padding: 20px;
  text-align: center;
}
.el-card {
  padding: 20px;
}
</style> -->
<template>
  <router-view />
</template>

<script>
export default {
  name: 'App',
};
</script>

<style>
html, body {
  margin: 0;
  padding: 0;
  font-family: Arial, sans-serif;
  background-color: #f5f5f5;
}
</style>
