<template>
    <el-card>
      <h2>加密</h2>

      <el-form :model="encryptForm" ref="encryptForm" status-icon @submit.native.prevent="handleSubmit">
        <el-form-item label="加密算法" prop="algorithm">
          <el-select v-model="selectedAlgorithm" placeholder="选择加密算法">
            <el-option label="PBEWithMD5AndTripleDES" value="PBEWithMD5AndTripleDES"></el-option>
            <el-option label="aes" value="aes"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="Text to Encrypt(一般填密码)" prop="encryptMsg" >
          <el-input
            v-model="encryptMsg"
            name="encryptMsg"
            type="textarea"
            placeholder="text"
            autocomplete="on"
          />
          <!-- style="width: 240px"
            :rows="2"
            type="textarea" -->
        </el-form-item>
        <el-form-item label="Secret Key(一般填加密的key)" prop="encryptKey">
          <el-input
            v-model="encryptKey"
            autocomplete="on"
            name="encryptMsg"
            placeholder="Secret Key"
            type = "text"
          />
          <!-- style="width: 240px"
            :rows="2"
            type="textarea" -->
        </el-form-item>
        <div>
          <el-button type="primary" native-type="submit" @click="validateEncrypt">加密</el-button>
          <el-button type="default" @click="resetEncrypt">重置</el-button>
        </div>
      </el-form>
      <p v-if="encryptedString">加密结果: {{ encryptedString }}</p>
      <el-button v-if="encryptedString" type="primary" @click="copyToClipboard(encryptedString)">
      复制加密结果
    </el-button>
      <p v-if="response">{{ response }}</p>
    </el-card>
  
    <el-card style="margin-top: 20px;">
      <h2>解密</h2>
      <el-form :model="decryptForm" ref="decryptForm" status-icon @submit.native.prevent="handleSubmit">
        <el-form-item label="加密算法" prop="algorithm">
          <el-select v-model="selectedAlgorithm" placeholder="选择加密算法">
            <el-option label="PBEWithMD5AndTripleDES" value="PBEWithMD5AndTripleDES"></el-option>
            <el-option label="aes" value="aes"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="Encrypted Text(一般填加密之后的text)" prop="decryptString">
          <el-input
            v-model="decryptString"
            style="width: 240px"
            :rows="2"
            autocomplete="on"
            name="decryptString"
            type="textarea"
            placeholder="text"
          />
        </el-form-item>
        <el-form-item label="Secret Key(填对应加密的key)" prop="decryptKey">
          <el-input
            v-model="decryptKey"
            style="width: 240px"
            :rows="2"
            autocomplete="on"
            name="decryptKey"
            type="text"
            placeholder="Secret Key"
          />
        </el-form-item>
        <div>
          <el-button type="primary" native-type="submit" @click="validateDecrypt">解密</el-button>
          <el-button type="default" @click="resetDecrypt">重置</el-button>
        </div>
      </el-form>
      <p v-if="decryptedMsg">解密结果: {{ decryptedMsg }}</p>
      <el-button v-if="decryptedMsg" type="primary" @click="copyToClipboard(decryptedMsg)">
      复制解密结果
    </el-button>
    </el-card>
  </template>
  
  <script>
  import { encryptMessage } from '../../api/encrypt';
  import { decryptMessage } from '../../api/decrypt';
  
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
        algorithm: '',
        selectedAlgorithm: 'PBEWithMD5AndTripleDES', // 默认选择算法 
        encryptForm: {},
        decryptForm: {}
      };
    },
    methods: {
      handleSubmit(event) {
    // 阻止默认提交
    event.preventDefault();

    // 触发表单验证
    this.$refs.encryptForm.validate(valid => {
      if (valid) {
        // 验证通过，执行Ajax请求
        this.submitForm();
      } else {
        console.log('表单验证失败');
      }
    });
  },
      async validateEncrypt() {
        this.$refs.encryptForm.validate(async (valid) => {
          if (valid) {
            if (this.hasInvalidCharacters(this.encryptMsg,"Text to Encrypt") || this.hasInvalidCharacters(this.encryptKey,"Secret Key")) {
          return; // 如果有错误，直接返回
        }
            await this.encrypt();
          } else {
            alert('请确保输入的所有字段都是有效的！');
          }
        });
      },
      async validateDecrypt() {
        this.$refs.decryptForm.validate(async (valid) => {
          if (valid) {
            if (this.hasInvalidCharacters(this.decryptString,"Encrypted Text") || this.hasInvalidCharacters(this.decryptKey,"Secret Key")) {
          return;
        }
            await this.decrypt();
          } else {
            alert('请确保输入的所有字段都是有效的！');
          }
        });
      },
      hasInvalidCharacters(text,field) {
        if (this.isEmpty(text)) {
      this.$message.error(`${field} 不能为空！`);
      return true; // 为空时返回 true，表示有无效字符
    }

    // 正则表达式检查空格、制表符和换行符
    const invalidPattern = /[\s\t\n]/; // 这里的\s包含所有空白字符
    if (invalidPattern.test(text,field)) {
      this.$message.error(`${field} 不能包含空格、制表符或换行符。`);
      return true; // 含有无效字符时返回 true
    }

    return false; // 如果没有问题，返回 false
  },
  isEmpty(text) {
    // 检查字符串是否为空
    return !text || text.trim() === '';
  },
      async encrypt() {
        try {
          const response = await encryptMessage(this.encryptMsg, this.encryptKey, this.selectedAlgorithm);
          this.encryptedString = response.encrypted_string;
        } catch (error) {
          console.error("加密失败:", error.message);
          alert("加密失败: " + error.message);
        }
      },
  //     copyToClipboard(text) {
  //       if (window.isSecureContext && navigator.clipboard) {
  //   navigator.clipboard.writeText(text).then(() => {
  //     this.$message.success('复制成功');
  //   }).catch(err => {
  //     this.$message.error('复制失败: ' + err);
  //   });
  //       }else{
  //         const textarea = document.createElement('textarea');
  //               textarea.value = text;
  //               document.body.appendChild(textarea);
  //               textarea.select();
  //               document.execCommand('copy');
  //               document.body.removeChild(textarea);
  //               this.$message.success('复制成功');
  //       }
  // },
  copyToClipboard(text) {
    const showMessage = (message, isSuccess) => {
        if (isSuccess) {
            this.$message.success(message);
        } else {
            this.$message.error(message);
        }
    };

    if (window.isSecureContext && navigator.clipboard) {
        navigator.clipboard.writeText(text)
            .then(() => showMessage('复制成功', true))
            .catch(err => showMessage('复制失败: ' + err, false));
    } else {
        const textarea = document.createElement('textarea');
        textarea.value = text;
        document.body.appendChild(textarea);
        textarea.select();

        try {
            const successful = document.execCommand('copy');
            if (successful) {
                showMessage('复制成功', true);
            } else {
                showMessage('复制失败: 执行命令失败', false);
            }
        } catch (err) {
            showMessage('复制失败: ' + err, false);
        } finally {
            document.body.removeChild(textarea);
        }
    }
},
      async decrypt() {
        try {
          const response = await decryptMessage(this.decryptString, this.decryptKey, this.selectedAlgorithm);
          this.decryptedMsg = response.decrypted_msg;
        } catch (error) {
          console.error("解密失败:", error.message);
          alert("解密失败:(检查信息是否正确)" + error.message);
        }
      },
      resetEncrypt() {
        this.encryptMsg = '';
        this.encryptKey = '';
        this.encryptedString = '';
        this.response = '';
        this.selectedAlgorithm = 'PBEWithMD5AndTripleDES'; // 重置算法选择
        this.$refs.encryptForm.resetFields(); // 重置表单字段
      //   // 清除 localStorage 中的记录
      // localStorage.removeItem("encryptKey");
      // localStorage.removeItem("encryptMsg");
      },
      resetDecrypt() {
        this.decryptString = '';
        this.decryptKey = '';
        this.decryptedMsg = '';
        this.response = '';
        this.selectedAlgorithm = 'PBEWithMD5AndTripleDES'; // 重置算法选择
        this.$refs.decryptForm.resetFields(); // 重置表单字段
      },
    },
  };
  </script>
  
  <style scoped>
  .el-card {
    padding: 20px;
  }
  </style>
  