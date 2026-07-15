<template>
  <span class="helpDoc-link-text" @click="handleClick">
    {{ text }}
  </span>
</template>

<script>
import { mapGetters } from 'vuex';

export default {
  props: {
    type: '',
    text: '',
  },
  data() {
    return {
      basePath: this.$basePath,
      linkList: {},
    };
  },
  mounted() {
    const { linkList } = this.commonInfo.data || {};
    this.linkList = linkList || {};
  },
  computed: {
    ...mapGetters('user', ['commonInfo']),
  },
  methods: {
    handleClick() {
      const docUrl = this.linkList[this.type];
      if (docUrl) window.open(docUrl);
    },
  },
};
</script>

<style lang="scss" scoped>
.helpDoc-link-text {
  font-size: 12px;
  margin-left: 8px;
  color: #55575f;
  cursor: pointer;
}
.helpDoc-link-text:hover {
  text-decoration: underline;
}
</style>
