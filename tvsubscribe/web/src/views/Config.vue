<template>
  <div class="config-container">
    <el-card class="config-card">
      <template #header>
        <div class="card-header">
          <span>配置管理</span>
          <el-button type="primary" @click="loadConfig">刷新配置</el-button>
        </div>
      </template>

      <el-form :model="config" label-width="150px" v-if="config">
        <el-form-item label="服务器地址">
          <el-input v-model="config.endpoint" placeholder="请输入服务器地址" />
        </el-form-item>

        <el-form-item label="Cookie">
          <el-input
            v-model="config.cookie"
            type="textarea"
            :rows="3"
            placeholder="请输入Cookie"
          />
        </el-form-item>

        <el-form-item label="检查间隔(分钟)">
          <el-input-number
            v-model="config.interval_minutes"
            :min="1"
            :max="1440"
            placeholder="检查间隔"
          />
        </el-form-item>

        <el-form-item label="微信服务器">
          <el-input v-model="config.wechat_server" placeholder="请输入微信服务器地址" />
        </el-form-item>

        <el-form-item label="微信Token">
          <el-input v-model="config.wechat_token" placeholder="请输入微信Token" />
        </el-form-item>

        <el-form-item label="监听端口">
          <el-input-number
            v-model="config.port"
            :min="1"
            :max="65535"
            placeholder="监听端口"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="saveConfig" :loading="saving">
            保存配置
          </el-button>
          <el-button @click="loadConfig">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'Config',
  data() {
    return {
      config: null,
      saving: false
    }
  },
  mounted() {
    this.loadConfig()
  },
  methods: {
    async loadConfig() {
      try {
        const response = await axios.get('/getConfig')
        if (response.data.success) {
          this.config = response.data.data
        } else {
          this.$message.error('获取配置失败: ' + response.data.message)
        }
      } catch (error) {
        this.$message.error('获取配置失败: ' + error.message)
      }
    },

    async saveConfig() {
      this.saving = true
      try {
        const response = await axios.post('/setConfig', this.config)
        if (response.data.success) {
          this.$message.success('配置保存成功')
        } else {
          this.$message.error('保存配置失败: ' + response.data.message)
        }
      } catch (error) {
        this.$message.error('保存配置失败: ' + error.message)
      } finally {
        this.saving = false
      }
    }
  }
}
</script>

<style scoped>
.config-container {
  max-width: 800px;
  margin: 0 auto;
}

.config-card {
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header span {
  font-size: 18px;
  font-weight: 500;
}
</style>