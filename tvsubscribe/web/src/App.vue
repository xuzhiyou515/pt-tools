<template>
  <div id="app">
    <el-container class="layout-container">
      <el-header class="header">
        <h1>TVSubscribe 管理界面</h1>
      </el-header>
      <el-container>
        <el-aside width="200px" class="sidebar">
          <el-menu
            :default-active="$route.path"
            class="el-menu-vertical"
            router
          >
            <el-menu-item index="/config">
              <el-icon><Setting /></el-icon>
              <span>配置管理</span>
            </el-menu-item>
            <el-menu-item index="/subscribe">
              <el-icon><List /></el-icon>
              <span>订阅管理</span>
            </el-menu-item>
          </el-menu>
        </el-aside>
        <el-main class="main-content">
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script>
import { Setting, List } from '@element-plus/icons-vue'

export default {
  name: 'App',
  components: {
    Setting,
    List
  }
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

#app {
  height: 100vh;
}

.layout-container {
  height: 100%;
}

.header {
  background-color: #409EFF;
  color: white;
  display: flex;
  align-items: center;
  padding: 0 20px;
}

.header h1 {
  font-size: 24px;
  font-weight: 500;
}

.sidebar {
  background-color: #f5f5f5;
  border-right: 1px solid #e6e6e6;
}

.main-content {
  padding: 20px;
  background-color: #fafafa;
}
</style>