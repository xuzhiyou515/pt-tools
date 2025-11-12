<template>
  <div class="subscribe-container">
    <el-card class="subscribe-card">
      <template #header>
        <div class="card-header">
          <span>订阅管理</span>
          <el-button type="primary" @click="loadSubscribes">刷新列表</el-button>
        </div>
      </template>

      <!-- 添加订阅表单 -->
      <el-card class="add-subscribe-card" shadow="never">
        <template #header>
          <span>添加新订阅</span>
        </template>
        <el-form :model="newSubscribe" :rules="rules" ref="subscribeForm" label-width="100px">
          <el-row :gutter="20">
            <el-col :span="8">
              <el-form-item label="豆瓣ID" prop="douban_id">
                <el-input
                  v-model="newSubscribe.douban_id"
                  placeholder="请输入豆瓣ID"
                  clearable
                />
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="分辨率" prop="resolution">
                <el-select v-model="newSubscribe.resolution" placeholder="请选择分辨率">
                  <el-option label="2160P" :value="0" />
                  <el-option label="1080P" :value="1" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item>
                <el-button type="primary" @click="addSubscribe" :loading="adding">
                  添加订阅
                </el-button>
                <el-button @click="resetForm">重置</el-button>
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>
      </el-card>

      <!-- 批量操作栏 -->
      <div class="batch-actions" v-if="selectedSubscribes.length > 0">
        <el-alert
          :title="`已选择 ${selectedSubscribes.length} 个订阅`"
          type="info"
          :closable="false"
        >
          <template #default>
            <el-button
              type="danger"
              size="small"
              @click="batchDelete"
              :loading="batchDeleting"
            >
              批量删除
            </el-button>
            <el-button
              size="small"
              @click="clearSelection"
            >
              取消选择
            </el-button>
          </template>
        </el-alert>
      </div>

      <!-- 订阅列表 -->
      <el-table
        :data="subscribes"
        stripe
        style="width: 100%"
        v-loading="loading"
        @selection-change="handleSelectionChange"
        ref="subscribeTable"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="name" label="电视剧名称" min-width="200" show-overflow-tooltip />
        <el-table-column prop="douban_id" label="豆瓣ID" min-width="120" />
        <el-table-column prop="resolution" label="分辨率" min-width="120">
          <template #default="scope">
            <el-tag :type="scope.row.resolution === 0 ? 'danger' : 'primary'">
              {{ scope.row.resolution === 0 ? '2160P' : '1080P' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="scope">
            <el-button
              type="danger"
              size="small"
              @click="deleteSubscribe(scope.row)"
              :loading="scope.row.deleting"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && subscribes.length === 0" description="暂无订阅数据" />
    </el-card>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'Subscribe',
  data() {
    return {
      subscribes: [],
      loading: false,
      adding: false,
      batchDeleting: false,
      selectedSubscribes: [],
      newSubscribe: {
        douban_id: '',
        resolution: 1 // 默认1080P
      },
      rules: {
        douban_id: [
          { required: true, message: '请输入豆瓣ID', trigger: 'blur' }
        ],
        resolution: [
          { required: true, message: '请选择分辨率', trigger: 'change' }
        ]
      }
    }
  },
  mounted() {
    this.loadSubscribes()
  },
  methods: {
    async loadSubscribes() {
      this.loading = true
      try {
        const response = await axios.get('/getSubscribeList')
        if (response.data.success) {
          this.subscribes = response.data.data.map(item => ({
            ...item,
            deleting: false
          }))
        } else {
          this.$message.error('获取订阅列表失败: ' + response.data.message)
        }
      } catch (error) {
        this.$message.error('获取订阅列表失败: ' + error.message)
      } finally {
        this.loading = false
      }
    },

    async addSubscribe() {
      try {
        await this.$refs.subscribeForm.validate()
        this.adding = true

        const response = await axios.post('/addSubscribe', this.newSubscribe)
        if (response.data.success) {
          this.$message.success('添加订阅成功')
          this.resetForm()
          await this.loadSubscribes()
        } else {
          this.$message.error('添加订阅失败: ' + response.data.message)
        }
      } catch (error) {
        if (error.message) {
          this.$message.error('添加订阅失败: ' + error.message)
        }
      } finally {
        this.adding = false
      }
    },

    async deleteSubscribe(subscribe) {
      try {
        await this.$confirm('确定要删除这个订阅吗？', '确认删除', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })

        subscribe.deleting = true

        const response = await axios.post('/api/delSubscribe', {
          douban_id: subscribe.douban_id,
          resolution: subscribe.resolution
        })

        if (response.data.success) {
          this.$message.success('删除订阅成功')
          await this.loadSubscribes()
        } else {
          this.$message.error('删除订阅失败: ' + response.data.message)
        }
      } catch (error) {
        if (error !== 'cancel') {
          this.$message.error('删除订阅失败: ' + error.message)
        }
      } finally {
        subscribe.deleting = false
      }
    },

    resetForm() {
      this.$refs.subscribeForm?.resetFields()
      this.newSubscribe = {
        douban_id: '',
        resolution: 1
      }
    },

    // 处理选择变化
    handleSelectionChange(selection) {
      this.selectedSubscribes = selection
    },

    // 清空选择
    clearSelection() {
      this.$refs.subscribeTable?.clearSelection()
      this.selectedSubscribes = []
    },

    // 批量删除
    async batchDelete() {
      if (this.selectedSubscribes.length === 0) {
        this.$message.warning('请先选择要删除的订阅')
        return
      }

      try {
        await this.$confirm(
          `确定要删除选中的 ${this.selectedSubscribes.length} 个订阅吗？`,
          '批量删除确认',
          {
            confirmButtonText: '确定删除',
            cancelButtonText: '取消',
            type: 'warning',
            dangerouslyUseHTMLString: false
          }
        )

        this.batchDeleting = true

        // 并发删除所有选中的订阅
        const deletePromises = this.selectedSubscribes.map(async (subscribe) => {
          try {
            const response = await axios.post('/delSubscribe', {
              douban_id: subscribe.douban_id,
              resolution: subscribe.resolution
            })
            return { subscribe, success: true, response }
          } catch (error) {
            return { subscribe, success: false, error }
          }
        })

        const results = await Promise.all(deletePromises)

        // 统计成功和失败的数量
        const successCount = results.filter(r => r.success).length
        const failureCount = results.length - successCount

        if (failureCount === 0) {
          this.$message.success(`成功删除 ${successCount} 个订阅`)
        } else if (successCount === 0) {
          this.$message.error(`删除失败，请检查网络连接`)
        } else {
          this.$message.warning(`成功删除 ${successCount} 个订阅，${failureCount} 个订阅删除失败`)
        }

        // 清空选择并刷新列表
        this.clearSelection()
        await this.loadSubscribes()

      } catch (error) {
        if (error !== 'cancel') {
          this.$message.error('批量删除失败: ' + error.message)
        }
      } finally {
        this.batchDeleting = false
      }
    }
  }
}
</script>

<style scoped>
.subscribe-container {
  max-width: 1200px;
  margin: 0 auto;
}

.subscribe-card {
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

.add-subscribe-card {
  margin-bottom: 20px;
  background-color: #fafafa;
}

.batch-actions {
  margin-bottom: 20px;
}

.batch-actions .el-alert {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.batch-actions .el-alert__content {
  flex: 1;
}

.batch-actions .el-button {
  margin-left: 10px;
}
</style>