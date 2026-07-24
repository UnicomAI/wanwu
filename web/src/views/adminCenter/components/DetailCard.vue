<template>
  <div class="detail-card">
    <div class="card-title">{{ title }}</div>
    <div class="card-content">
      <slot />
    </div>
  </div>
</template>

<script>
export default {
  props: {
    title: {
      type: String,
      required: true,
    },
  },
};
</script>

<style lang="scss" scoped>
.detail-card {
  background: #fff;
  border-radius: 8px;
  padding: 20px 24px;
  margin-bottom: 20px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  border-left: 4px solid #409eff;
  padding-left: 12px;
  margin-bottom: 20px;
}

.card-content {
  padding-left: 16px;
}

::v-deep .el-form-item__label {
  color: #909399;
  font-size: 14px;
}
</style>
