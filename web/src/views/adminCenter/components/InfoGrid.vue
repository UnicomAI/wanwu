<template>
  <div class="info-grid">
    <div
      v-for="(item, index) in normalizedItems"
      :key="index"
      class="info-item"
    >
      <span class="info-label">{{ item.label }}</span>
      <div v-if="item._render" class="info-value keyword-tags">
        <img
          v-if="item._render.icon"
          :src="item._render.icon"
          class="model-img"
        />
        <span v-if="item._render.name" class="model-name">
          {{ item._render.name }}
        </span>
        <el-tag
          v-for="(tag, i) in item._render.tags"
          :key="i"
          class="keyword-tag"
          :class="{ 'is-link': tag.url }"
          color="#E6F0FF"
          size="small"
          @click="goTag(tag)"
        >
          {{ tag.text }}
        </el-tag>
        <span
          v-if="
            !item._render.icon &&
            !item._render.name &&
            !item._render.tags.length
          "
        >
          -
        </span>
      </div>
      <span v-else class="info-value">{{ item.value || '-' }}</span>
    </div>
  </div>
</template>

<script>
export default {
  props: {
    items: {
      type: Array,
      default: () => [],
    },
  },
  computed: {
    normalizedItems() {
      return this.items.map(item => {
        if (item.model) {
          return {
            ...item,
            _render: {
              icon: item.model.icon,
              name: item.model.name,
              tags: item.model.tags || [],
            },
          };
        }
        if (item.tags) {
          return { ...item, _render: { tags: item.tags || [] } };
        }
        return item;
      });
    },
  },
  methods: {
    goTag(tag) {
      if (!tag.url) return;
      const route = this.$router.resolve(tag.url);
      window.open(route.href, '_blank');
    },
  },
};
</script>

<style lang="scss" scoped>
.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px 40px;
}

.info-item {
  display: flex;
  align-items: center;
}

.info-label {
  color: #909399;
  font-size: 14px;
  width: 120px;
  flex-shrink: 0;
}

.info-value {
  color: #303133;
  font-size: 14px;
  flex: 1;
  word-break: break-all;
}

.keyword-tags {
  align-items: center;
  display: flex;
  flex-wrap: wrap;
  gap: 4px;

  .model-img {
    width: 18px;
    height: 18px;
    border-radius: 4px;
    vertical-align: middle;
  }

  .model-name {
    font-weight: 500;
  }
}

.keyword-tag {
  margin-right: 4px;
  margin-bottom: 2px;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: $tag_color;

  &.is-link {
    cursor: pointer;
  }
}
</style>
