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
            <el-col :span="12">
              <el-form-item label="豆瓣ID/名称" prop="douban_id">
                <el-autocomplete
                  v-model="newSubscribe.douban_id"
                  :fetch-suggestions="querySearchAsync"
                  placeholder="请输入豆瓣ID或电视剧名称"
                  clearable
                  @select="handleSelect"
                  @input="handleInput"
                  :trigger-on-focus="false"
                  style="width: 100%"
                >
                  <template #default="{ item }">
                    <div class="search-result-item">
                      <div class="search-result-content">
                        <div class="poster-container">
                          <img
                            v-if="item.img"
                            :src="getDoubanImageUrl(item.img)"
                            :alt="item.title"
                            class="poster-image"
                            @error="handleImageError"
                            referrerpolicy="no-referrer"
                          />
                          <div v-else class="poster-placeholder">
                            <i class="el-icon-picture-outline"></i>
                          </div>
                        </div>
                        <div class="search-info">
                          <div class="title">{{ item.title }}</div>
                          <div class="meta-info">
                            <span v-if="item.year" class="meta-item year-item">
                              <i class="el-icon-date"></i>
                              {{ item.year }}
                            </span>
                            <span v-if="item.episode" class="meta-item episode-item">
                              <i class="el-icon-video-camera"></i>
                              {{ item.episode }}集
                            </span>
                          </div>
                        </div>
                      </div>
                    </div>
                  </template>
                </el-autocomplete>
              </el-form-item>
            </el-col>
            <el-col :span="6">
              <el-form-item label="分辨率" prop="resolution">
                <el-select v-model="newSubscribe.resolution" placeholder="请选择分辨率" style="width: 100%">
                  <el-option label="2160P" :value="0" />
                  <el-option label="1080P" :value="1" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="6">
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
              type="success"
              size="small"
              @click="batchTrigger"
              :loading="batchTriggering"
            >
              批量触发
            </el-button>
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
        <el-table-column label="电视剧名称" min-width="200" show-overflow-tooltip>
          <template #default="scope">
            <a
              v-if="scope.row.name"
              :href="getDoubanUrl(scope.row.douban_id)"
              target="_blank"
              class="douban-link"
              :title="`在豆瓣中查看《${scope.row.name}》`"
            >
              {{ scope.row.name }}
            </a>
            <span v-else class="no-name">未获取到名称</span>
          </template>
        </el-table-column>
        <el-table-column label="豆瓣ID" min-width="120">
          <template #default="scope">
            <a
              :href="getDoubanUrl(scope.row.douban_id)"
              target="_blank"
              class="douban-link douban-id-link"
              :title="`在豆瓣中查看ID: ${scope.row.douban_id}`"
            >
              {{ scope.row.douban_id }}
            </a>
          </template>
        </el-table-column>
        <el-table-column prop="resolution" label="分辨率" min-width="120">
          <template #default="scope">
            <el-tag :type="scope.row.resolution === 0 ? 'danger' : 'primary'">
              {{ scope.row.resolution === 0 ? '2160P' : '1080P' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <el-button
              type="success"
              size="small"
              @click="triggerSubscribe(scope.row)"
              :loading="scope.row.triggering"
            >
              触发
            </el-button>
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
      batchTriggering: false,
      selectedSubscribes: [],
      newSubscribe: {
        douban_id: '',
        resolution: 1 // 默认1080P
      },
      searchResults: [],
      searching: false,
      rules: {
        douban_id: [
          { required: true, message: '请输入豆瓣ID或电视剧名称', trigger: 'blur' }
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
    // 豆瓣搜索功能
    async querySearchAsync(queryString, callback) {
      if (!queryString || queryString.trim().length === 0) {
        callback([])
        return
      }

      // 如果输入的是纯数字，不进行搜索
      if (/^\d+$/.test(queryString)) {
        callback([])
        return
      }

      this.searching = true
      try {
        const response = await axios.get('/searchDouBan', {
          params: { name: queryString }
        })

        if (response.data.success) {
          const results = response.data.data.map(item => ({
            value: `${item.title} (${item.douban_id})`,
            title: item.title,
            douban_id: item.douban_id,
            img: item.img || '',
            year: item.year || '',
            episode: item.episode || ''
          }))
          callback(results)
        } else {
          console.error('搜索失败:', response.data.message)
          callback([])
        }
      } catch (error) {
        console.error('搜索请求失败:', error)
        callback([])
      } finally {
        this.searching = false
      }
    },

    // 处理选择搜索结果
    handleSelect(item) {
      // 只设置豆瓣ID，不设置显示值
      this.newSubscribe.douban_id = item.douban_id
    },

    // 处理输入变化
    handleInput(value) {
      // 如果用户手动输入了数字ID，保持原值
      // 如果用户输入了文字，会触发搜索
    },

    // 处理图片加载错误
    handleImageError(event) {
      event.target.style.display = 'none'
      const placeholder = event.target.nextElementSibling
      if (placeholder && placeholder.classList.contains('poster-placeholder')) {
        placeholder.style.display = 'flex'
      }
    },

    // 获取豆瓣图片URL，使用代理避免403错误
    getDoubanImageUrl(originalUrl) {
      if (!originalUrl) return ''
      // 使用自己的图片代理接口，避免豆瓣防盗链
      const encodedUrl = encodeURIComponent(originalUrl)
      return `/proxy/image?url=${encodedUrl}`
    },

    // 生成豆瓣页面URL
    getDoubanUrl(doubanId) {
      if (!doubanId) return '#'
      return `https://movie.douban.com/subject/${doubanId}/`
    },
    async loadSubscribes() {
      this.loading = true
      try {
        const response = await axios.get('/getSubscribeList')
        if (response.data.success) {
          this.subscribes = response.data.data.map(item => ({
            ...item,
            deleting: false,
            triggering: false
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

        const response = await axios.post('/delSubscribe', {
          ids: [subscribe.id]
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

        // 收集所有要删除的ID
        const idsToDelete = this.selectedSubscribes.map(subscribe => subscribe.id).filter(id => id)

        if (idsToDelete.length === 0) {
          this.$message.warning('请先选择要删除的订阅')
          this.batchDeleting = false
          return
        }

        // 使用新的ID数组格式批量删除
        try {
          const response = await axios.post('/delSubscribe', {
            ids: idsToDelete
          })

          if (response.data.success) {
            this.$message.success(response.data.message || '批量删除成功')
          } else {
            this.$message.error('批量删除失败: ' + response.data.message)
          }
        } catch (error) {
          this.$message.error('批量删除失败: ' + error.message)
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
    },

    // 触发单个订阅
    async triggerSubscribe(subscribe) {
      try {
        await this.$confirm('确定要触发这个订阅吗？', '确认触发', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })

        subscribe.triggering = true

        const response = await axios.post('/triggerNow', {
          ids: [subscribe.id]
        })

        if (response.data.success) {
          this.$message.success(response.data.message || '订阅触发成功')
        } else {
          this.$message.error('订阅触发失败: ' + response.data.message)
        }
      } catch (error) {
        if (error !== 'cancel') {
          this.$message.error('订阅触发失败: ' + error.message)
        }
      } finally {
        subscribe.triggering = false
      }
    },

    // 批量触发订阅
    async batchTrigger() {
      if (this.selectedSubscribes.length === 0) {
        this.$message.warning('请先选择要触发的订阅')
        return
      }

      try {
        await this.$confirm(
          `确定要触发选中的 ${this.selectedSubscribes.length} 个订阅吗？`,
          '批量触发确认',
          {
            confirmButtonText: '确定触发',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )

        this.batchTriggering = true

        // 收集所有要触发的ID
        const idsToTrigger = this.selectedSubscribes.map(subscribe => subscribe.id).filter(id => id)

        if (idsToTrigger.length === 0) {
          this.$message.warning('没有有效的订阅ID')
          this.batchTriggering = false
          return
        }

        const response = await axios.post('/triggerNow', {
          ids: idsToTrigger
        })

        if (response.data.success) {
          this.$message.success(response.data.message || '批量触发成功')
        } else {
          this.$message.error('批量触发失败: ' + response.data.message)
        }
      } catch (error) {
        if (error !== 'cancel') {
          this.$message.error('批量触发失败: ' + error.message)
        }
      } finally {
        this.batchTriggering = false
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

/* 搜索结果样式 */
.search-result-item {
  padding: 8px 0;
}

.search-result-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.poster-container {
  flex-shrink: 0;
  width: 40px;
  height: 56px;
  position: relative;
  border-radius: 4px;
  overflow: hidden;
  background-color: #f5f7fa;
}

.poster-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 4px;
}

.poster-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #c0c4cc;
  font-size: 16px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.search-info {
  flex: 1;
  min-width: 0;
}

.search-result-item .title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.4;
  margin-bottom: 4px;
}

.meta-info {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.meta-item {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  font-size: 10px;
  padding: 1px 5px;
  border-radius: 8px;
  color: #606266;
  background-color: #f4f4f5;
  white-space: nowrap;
  line-height: 1.2;
}

.meta-item i {
  font-size: 9px;
}

.year-item {
  color: #409EFF;
  background-color: #ecf5ff;
}

.episode-item {
  color: #67C23A;
  background-color: #f0f9ff;
}

/* 豆瓣链接样式 */
.douban-link {
  color: #409EFF;
  text-decoration: none;
  transition: color 0.3s ease;
  cursor: pointer;
}

.douban-link:hover {
  color: #66b1ff;
  text-decoration: underline;
}

.douban-id-link {
  font-family: 'Courier New', monospace;
  font-size: 13px;
  color: #606266;
}

.douban-id-link:hover {
  color: #409EFF;
}

.no-name {
  color: #C0C4CC;
  font-style: italic;
}
</style>