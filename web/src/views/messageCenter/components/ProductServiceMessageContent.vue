<template>
  <div class="product-service-message-content">
    <template v-for="(part, index) in contentParts">
      <el-button
        v-if="part.action"
        :key="index"
        type="text"
        class="product-service-message-content__action"
        @click="handleAction(part.action)"
      >
        {{ part.text }}
      </el-button>
      <span v-else :key="index">{{ part.text }}</span>
    </template>
  </div>
</template>

<script>
const MESSAGE_TYPE_ROUTE_MAP = Object.freeze({
  agent: '/explore/agent',
  workflow: '/explore/workflow',
  chatflow: '/explore/workflow',
  rag: '/explore/rag',
  skill: '/skillSquare/detail',
  knowledge: '/knowledge/doclist',
  qaKnowledge: '/knowledge/qa/docList',
  model: '/modelAccess',
  about: null,
});
export default {
  name: 'ProductServiceMessageContent',
  props: {
    msgText: {
      type: String,
      default: '',
    },
    actions: {
      type: Array,
      default: () => [],
    },
  },
  computed: {
    contentParts() {
      const parts = [];
      const pattern = /\$\$\{([\s\S]*?)\}\$\$/g;
      let lastIndex = 0;
      let actionIndex = 0;
      let match = pattern.exec(this.msgText);

      while (match) {
        if (match.index > lastIndex) {
          parts.push({
            type: 'text',
            text: this.msgText.slice(lastIndex, match.index),
          });
        }
        parts.push({
          type: 'action',
          text: match[1],
          action: this.actions[actionIndex],
        });
        actionIndex += 1;
        lastIndex = pattern.lastIndex;
        match = pattern.exec(this.msgText);
      }

      if (lastIndex < this.msgText.length) {
        parts.push({ type: 'text', text: this.msgText.slice(lastIndex) });
      }
      return parts;
    },
  },
  methods: {
    getRouteLocation(action) {
      if (!action) return null;

      const { msgType, actionParams = {} } = action;
      const { id, skillType } = actionParams;
      const path = MESSAGE_TYPE_ROUTE_MAP[msgType];
      if (!path || !id) return null;

      let query = { id };
      if (msgType === 'skill') {
        if (!skillType) return null;
        query = { skillId: id, skillType };
      } else if (msgType === 'model') {
        query = { modelId: id, from: 'messageCenter' };
      } else if (msgType === 'knowledge' || msgType === 'qaKnowledge') {
        return { path: `${path}/${id}` };
      }

      return { path, query };
    },
    handleAction(action) {
      if (action && action.msgType === 'about') {
        this.$store.commit('layout/OPEN_ABOUT_DIALOG');
        return;
      }

      const location = this.getRouteLocation(action);
      if (!location) return;

      if (location.external) {
        if (action.actionType === 'blank') {
          window.open(location.url, '_blank', 'noopener');
        } else {
          window.location.assign(location.url);
        }
        return;
      }

      if (action.actionType === 'blank') {
        const { href } = this.$router.resolve(location);
        window.open(href, '_blank', 'noopener');
        return;
      }
      this.$router.push(location);
    },
  },
};
</script>

<style lang="scss" scoped>
.product-service-message-content {
  padding: 8px 48px;
  color: #606266;
  line-height: 22px;
  white-space: pre-wrap;

  &__action {
    padding: 0;
    font-size: inherit;
    line-height: inherit;
    vertical-align: baseline;
  }
}
</style>
