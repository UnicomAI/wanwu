<template>
  <header class="page-title">
    <i class="el-icon-arrow-left" @click="goBack" />
    <img class="page-title-img" src="@/assets/imgs/org.png" alt="" />
    <span class="page-title-name">{{ title }}</span>
  </header>
</template>

<script>
export default {
  props: {
    title: {
      type: String,
      required: true,
    },
  },
  methods: {
    goBack() {
      this.$router.replace({ path: '/adminCenter' });
    },
  },
};
</script>

<style lang="scss" scoped>
.page-title {
  .el-icon-arrow-left {
    margin-right: 10px;
    font-size: 16px;
    cursor: pointer;
    color: $color_title;
  }
}
</style>
