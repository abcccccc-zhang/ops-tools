<template>
  <el-card>
    <h2>安装包下载链接生成器</h2>

    <el-form :inline="true" :model="form">
      <el-form-item label="版本号" prop="version">
        <el-input
          v-model="form.version"
          name="version"
          autocomplete="on"
          loadVersionHistory
          placeholder="请输入版本号，如 v5.2"
        />
        <div style="font-size: 12px; color: #999; margin-top: 1px;">
          ⚠️ 单机版 arm 架构只有一个版本号，v5.9.22-arm
          生成链接失败请检查版本是否正确
        </div>
      </el-form-item>

      <el-form-item label="类型">
        <el-radio-group v-model="form.type">
          <el-radio label="single">单机版</el-radio>
          <el-radio label="k8s">K8S 版</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item label="架构">
        <el-select v-model="form.arch" placeholder="请选择架构" style="width: 120px;">
          <el-option label="amd64" value="amd64" />
          <el-option label="arm64" value="arm64" />
        </el-select>
      </el-form-item>

      <el-form-item label="过期时间">
        <el-input v-model="form.expireValue" placeholder="数字" style="width: 100px;" />
        <el-select v-model="form.expireUnit" placeholder="单位" style="width: 80px; margin-left: 5px;">
          <el-option label="分钟" value="m" />
          <el-option label="小时" value="h" />
        </el-select>
      </el-form-item>

      <el-form-item>
        <el-button type="primary" @click="fetchPackageList">获取安装包路径</el-button>
      </el-form-item>
    </el-form>

    <!-- 单机版 -->
    <div v-if="form.type === 'single' && packageList.length">
      <h3>单机版安装包:</h3>
      <p class="single-path">{{ packageList[0].path }}</p>
      <el-button type="success" @click="generateSingleLink(packageList[0].path)">生成链接</el-button>
    </div>

    <!-- K8S 版 -->
    <div v-if="form.type === 'k8s' && packageList.length">
      <h3>K8S 安装包列表:</h3>
      <el-table :data="packageList" style="width: 100%">
        <el-table-column label="路径" prop="path" class-name="path-column">
          <template #default="{ row }">
            <span class="path-column">[{{ row.section }}] {{ row.path }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作">
          <template #default="{ row, $index }">
            <el-button size="mini" @click="generateSingleLink(row.path)">生成链接</el-button>
            <el-button size="mini" type="danger" @click="removePackageRow($index)">-</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-button type="primary" style="margin-top: 10px; margin-right: 5px;" @click="generateAllLinks">
        生成全部下载链接
      </el-button>
      <el-button type="success" style="margin-top: 10px;" @click="clearAllLinks">
        清空全部连接
      </el-button>
    </div>

    <!-- 下载链接展示 -->
    <div v-if="generatedLinks.length" style="margin-top: 10px;">
      <h3>✅ 下载链接:</h3>
      <div v-for="(link, index) in generatedLinks" :key="index">
        <p class="single-path">{{ link.path }} : {{ link.url }}</p>
      </div>
      <el-button type="primary" @click="copyAllLinks()">
        复制全部链接
      </el-button>
    </div>
  </el-card>
</template>

<script>
import { getInstallPackageList, generateDownloadURL } from "@/api/install";

export default {
  data() {
    return {
      form: {
        version: "",
        type: "single",
        arch: "amd64",
        expireValue: 24,
        expireUnit: "h",
      },
      packageList: [],
      generatedLinks: [],
    };
  },
  methods: {
    // 清空所有生成的链接
    clearAllLinks() {
      this.generatedLinks = [];
      this.$message.success("已清空全部链接");
    },

    // 删除单条 K8S 包
    removePackageRow(index) {
      this.packageList.splice(index, 1);
    },

    // 获取安装包列表
    async fetchPackageList() {
      if (!this.form.version || !this.form.version.trim()) {
        this.$message.warning("请输入版本号");
        return;
      }
      try {
        const res = await getInstallPackageList(this.form.version, this.form.type, this.form.arch);
        this.packageList = res.data.data || [];
        this.generatedLinks = [];
      } catch (err) {
        this.$message.error("获取安装包失败: " + err.message);
      }
    },

    // 生成单个下载链接
    async generateSingleLink(path) {
      try {
        const expire = `${this.form.expireValue}${this.form.expireUnit}`;
        const res = await generateDownloadURL(path, expire);

        if (this.form.type === "single") {
          this.generatedLinks = [{ path, url: res.data.url }];
        } else {
          const index = this.generatedLinks.findIndex(l => l.path === path);
          if (index >= 0) {
            this.generatedLinks[index].url = res.data.url;
          } else {
            this.generatedLinks.push({ path, url: res.data.url });
          }
        }
      } catch (err) {
        console.error("生成链接失败:", err); // 打印完整错误到控制台
        this.$message.error(err.response.data.error);
      }
    },

    // 生成全部下载链接
    async generateAllLinks() {
      if (!this.packageList.length) return;
      this.generatedLinks = [];
      for (const item of this.packageList) {
        try {
          const path = item.path || item;
          const expire = `${this.form.expireValue}${this.form.expireUnit}`;
          const res = await generateDownloadURL(path, expire);
          this.generatedLinks.push({ path, url: res.data.url });
        } catch (err) {
          this.$message.error(`生成失败: ${err.message}`);
        }
      }
      this.$message.success("全部链接生成完成");
    },

    // 复制全部链接到剪贴板
    copyAllLinks() {
      if (!this.generatedLinks.length) {
        this.$message.warning("没有可复制的链接");
        return;
      }
      const text = this.generatedLinks.map(l => l.url).join("\n");
      this.copyToClipboard(text);
    },

    copyToClipboard(text) {
      const showMessage = (message, isSuccess) => {
        if (isSuccess) this.$message.success(message);
        else this.$message.error(message);
      };
      if (navigator.clipboard && window.isSecureContext) {
        navigator.clipboard.writeText(text).then(() => showMessage("复制成功", true)).catch(err => showMessage("复制失败: " + err.message, false));
      } else {
        const textarea = document.createElement("textarea");
        textarea.value = text;
        textarea.style.position = "fixed";
        textarea.style.left = "-9999px";
        document.body.appendChild(textarea);
        textarea.focus();
        textarea.select();
        try {
          const successful = document.execCommand("copy");
          showMessage(successful ? "复制成功" : "复制失败", successful);
        } catch (err) {
          showMessage("复制失败: " + err.message, false);
        }
        document.body.removeChild(textarea);
      }
    },
  },
  watch: {
    // 切换单机/K8S，清空数据
    'form.type'(newType) {
      this.packageList = [];
      this.generatedLinks = [];
    }
  }
};
</script>

<style scoped>
/deep/ .path-column .cell {
  font-size: 16px;
  font-weight: 700;
}
.single-path {
  font-size: 16px;
  font-weight: 700;
}
</style>
