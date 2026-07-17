<template>
  <span :class="['type-tag', tagClass]">
    {{ tagName }}
  </span>
</template>

<script>
import { MODEL_TYPE_OBJ } from '@/views/modelAccess/constants';
import { MODEL_TAG_COLOR } from '@/views/statisticsDashboard/constants';

export default {
  props: {
    modelType: {
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
      return MODEL_TYPE_OBJ[this.modelType] || this.modelType || this.emptyText;
    },
    tagClass() {
      return MODEL_TAG_COLOR[this.modelType] || 'tag-gray';
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/statisticsTag.scss';
</style>
