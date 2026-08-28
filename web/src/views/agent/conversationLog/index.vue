<template>
  <div class="page-wrapper">
    <div class="page-title">
      <i class="back-btn el-icon-arrow-left" @click="$router.go(-1)" />
      <span class="page-title-name">
        {{
          $route.query.agentName
            ? $route.query.agentName + '-' + $t('agent.log.title')
            : $t('agent.log.title')
        }}
      </span>
    </div>

    <div class="page-wrapper-content">
      <ConversationLogPanel
        :app-id="assistantId"
        app-type="agent"
        :avatar-path="agentAvatarPath"
        :request-service="draftConversationLogService"
      />
    </div>
  </div>
</template>

<script>
import { getAgentInfo } from '@/api/agent';
import ConversationLogPanel from './ConversationLogPanel.vue';
import { draftConversationLogService } from './services';

export default {
  name: 'AgentConversationLog',
  components: {
    ConversationLogPanel,
  },
  data() {
    return {
      agentAvatarPath: '',
      draftConversationLogService,
    };
  },
  computed: {
    assistantId() {
      return this.$route.query.assistantId || '';
    },
  },
  created() {
    this.loadAgentAvatar();
  },
  methods: {
    async loadAgentAvatar() {
      if (!this.assistantId) return;

      try {
        const res = await getAgentInfo({ assistantId: this.assistantId });
        if (res && res.code === 0) {
          this.agentAvatarPath =
            (res.data && res.data.avatar && res.data.avatar.path) || '';
        }
      } catch (error) {
        this.agentAvatarPath = '';
      }
    },
  },
};
</script>

<style scoped lang="scss">
.page-title {
  display: flex;
  align-items: center;
  gap: 8px;

  .back-btn {
    cursor: pointer;
    font-size: 15px;
  }
}
</style>
