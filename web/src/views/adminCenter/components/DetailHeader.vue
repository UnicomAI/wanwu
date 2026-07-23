<template>
  <div class="model-header">
    <div class="model-avatar">
      <img :src="avatar || defaultIcon" alt="" />
    </div>
    <div class="model-meta">
      <div class="model-name-row">
        <span class="model-name">{{ name || '-' }}</span>
        <el-tag
          v-for="(tag, index) in tags"
          :key="index"
          :type="tag.type || ''"
          class="header-tag"
          size="small"
        >
          {{ tag.text }}
        </el-tag>
      </div>
      <div v-if="description" class="model-desc">{{ description }}</div>
    </div>
  </div>
</template>

<script>
const defaultIcon = require('@/assets/imgs/model_default_icon.png');

export default {
  data() {
    return {
      defaultIcon,
    };
  },
  props: {
    avatar: {
      type: String,
      default: '',
    },
    name: {
      type: String,
      default: '',
    },
    tags: {
      type: Array,
      default: () => [],
    },
    description: {
      type: String,
      default: '',
    },
  },
};
</script>

<style lang="scss" scoped>
.model-header {
  display: flex;
  align-items: flex-start;
  padding: 20px 24px;
}

.model-avatar {
  width: 64px;
  height: 64px;
  border-radius: 8px;
  overflow: hidden;
  flex-shrink: 0;
  margin-right: 16px;
  background: #f0f2f5;
  display: flex;
  align-items: center;
  justify-content: center;
  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

.model-meta {
  flex: 1;
}

.model-name-row {
  display: flex;
  align-items: center;
  margin-bottom: 6px;
}

.model-name {
  font-size: 18px;
  font-weight: 600;
  margin-right: 12px;
}

.header-tag {
  margin-right: 8px;
}

.model-desc {
  font-size: 14px;
  color: #909399;
}
</style>
