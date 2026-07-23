<template>
  <div class="page-wrapper">
    <div class="page-wrapper-content">
      <div class="page-title">
        <i
          class="el-icon-arrow-left"
          style="margin-right: 10px; font-size: 20px; cursor: pointer"
          @click="goBack"
        ></i>
        {{ backTitle }}
      </div>
      <slot />
    </div>
  </div>
</template>

<script>
export default {
  props: {
    backTitle: {
      type: String,
      required: true,
    },
    backPath: {
      type: String,
      required: true,
    },
  },
  methods: {
    goBack() {
      this.$router.replace(this.backPath);
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/views/adminCenter/styles/common.scss';
</style>
