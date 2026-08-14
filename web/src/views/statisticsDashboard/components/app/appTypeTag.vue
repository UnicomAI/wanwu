<template>
  <span :class="['type-tag', tagClass]">
    {{ tagName }}
  </span>
</template>

<script>
import { TagColorObj, TotalTypeObj } from '@/utils/commonSet';

export default {
  props: {
    appType: {
      type: String,
      default: '',
    },
    emptyText: {
      type: String,
      default: '--',
    },
  },
  computed: {
    tagName() {
      return TotalTypeObj[this.appType] || this.appType || this.emptyText;
    },
    tagClass() {
      return TagColorObj[this.appType] || 'tag-gray';
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/statisticsTag.scss';
</style>
