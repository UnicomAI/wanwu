<template>
  <div v-if="users.length" class="user-list-section">
    <div class="user-list-title">
      {{
        $t('adminCenter.common.visibleUsers', {
          count: users.length,
        })
      }}
    </div>
    <div class="user-list">
      <div v-for="user in users" :key="user.id" class="user-item">
        <el-avatar :size="36" :src="user.avatar" class="user-avatar">
          {{ user.name ? user.name.charAt(0) : 'U' }}
        </el-avatar>
        <div class="user-info">
          <div class="user-name-row">
            <span class="user-name">{{ user.name }}</span>
            <span class="user-org">{{ user.orgPath || '' }}</span>
          </div>
          <div class="user-contact">
            <span class="user-phone">{{ user.phone || '' }}</span>
            <span class="user-email">{{ user.email || '' }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  props: {
    users: {
      type: Array,
      default: () => [],
    },
  },
};
</script>

<style lang="scss" scoped>
.user-list-section {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid #ebeef5;
}

.user-list-title {
  font-size: 14px;
  color: #606266;
  margin-bottom: 12px;
}

.user-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.user-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  background: #f5f7fa;
  border-radius: 4px;
}

.user-avatar {
  margin-right: 12px;
  flex-shrink: 0;
}

.user-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.user-name-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.user-name {
  font-size: 14px;
  color: #303133;
  font-weight: 500;
}

.user-org {
  font-size: 13px;
  color: #606266;
}

.user-contact {
  display: flex;
  align-items: center;
  gap: 16px;
}

.user-phone,
.user-email {
  font-size: 13px;
  color: #909399;
}
</style>
