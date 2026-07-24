<template>
  <div class="empty-state">
    <img alt="" class="empty-img" src="@/assets/imgs/empty.png" />
    <p class="empty-tip">{{ tip }}</p>
  </div>
</template>

<script>
export default {
  props: {
    tip: {
      type: String,
      default: '',
    },
  },
};
</script>

<style lang="scss" scoped>
.empty-state {
  text-align: center;
  padding: 24px 0;
}
.empty-img {
  display: block;
  margin: 0 auto;
}
.empty-tip {
  margin-top: 8px;
  color: #909399;
  font-size: 14px;
}
</style>
