<template>
  <DetailLayout :back-title="backTitle" back-path="/adminCenter/agent">
    <DetailHeader
      :avatar="avatarUrl"
      :description="base.desc"
      :name="base.name"
      :tags="headerTags"
    />

    <DetailCard :title="$t('adminCenter.common.basicInfo')">
      <InfoGrid :items="basicItems" />
      <VisibleUsers :users="visibleUsers" />
    </DetailCard>

    <!-- 系统提示词 -->
    <DetailCard
      :title="
        $t('adminCenter.pageModules.appDevelopment.agent.detail.systemPrompt')
      "
    >
      <div v-if="detail.instructions" class="prompt-text input-box">
        {{ detail.instructions }}
      </div>
      <EmptyState
        v-else
        :tip="
          $t(
            'adminCenter.pageModules.appDevelopment.agent.detail.noSystemPrompt',
          )
        "
      />
    </DetailCard>

    <!-- 模型配置 -->
    <DetailCard
      :title="
        $t('adminCenter.pageModules.appDevelopment.agent.detail.modelConfig')
      "
    >
      <InfoGrid v-if="modelConfig.modelId" :items="modelItems" />
      <EmptyState
        v-else
        :tip="$t('adminCenter.pageModules.appDevelopment.agent.detail.noModel')"
      />
    </DetailCard>

    <!-- 对话配置 -->
    <DetailCard
      :title="
        $t('adminCenter.pageModules.appDevelopment.agent.detail.dialogConfig')
      "
    >
      <div class="dialog-grid">
        <div class="dialog-row">
          <span class="dialog-label">
            {{
              $t(
                'adminCenter.pageModules.appDevelopment.agent.detail.maxHistoryLength',
              )
            }}
          </span>
          <span class="dialog-value">
            {{ numText(detail.memoryConfig?.maxHistoryLength) }}
          </span>
        </div>
        <div class="dialog-row">
          <span class="dialog-label">
            {{
              $t('adminCenter.pageModules.appDevelopment.agent.detail.prologue')
            }}
          </span>
          <div v-if="detail.prologue" class="dialog-value input-box">
            {{ detail.prologue }}
          </div>
          <div v-else class="dialog-value empty-text">
            {{ $t('common.noData') }}
          </div>
        </div>
        <div class="dialog-row is-stack">
          <span class="dialog-label">
            {{
              $t(
                'adminCenter.pageModules.appDevelopment.agent.detail.recommendQuestion',
              )
            }}
          </span>
          <div class="dialog-value">
            <template v-if="recommendQuestions.length">
              <div
                v-for="(q, i) in recommendQuestions"
                :key="`rq-${i}`"
                class="recommend-item"
              >
                <span class="recommend-index">Q{{ i + 1 }}</span>
                <span class="recommend-text">{{ q }}</span>
              </div>
            </template>
            <div v-else class="empty-text">
              {{ $t('common.noData') }}
            </div>
          </div>
        </div>
      </div>
    </DetailCard>

    <!-- 关联知识库 -->
    <DetailCard
      v-if="!isMultipleAgent"
      :title="
        $t('adminCenter.pageModules.appDevelopment.agent.detail.linkKnowledge')
      "
    >
      <template v-if="hasKnowledge">
        <div class="recall-config">
          <InfoGrid :items="recallItems" />
        </div>
        <div class="card-list">
          <div
            v-for="(kb, i) in knowledgeList"
            :key="`kb-${i}`"
            class="resource-card is-clickable"
            @click="goKnowledge(kb)"
          >
            <img
              :src="avatarSrc(kb.avatar?.path, knowledgeDefaultIcon)"
              class="resource-avatar"
            />
            <div class="resource-info">
              <div class="resource-name">{{ kb.name }}</div>
              <div v-if="kb.description" class="resource-desc">
                {{ kb.description }}
              </div>
              <div class="resource-meta">
                <span class="meta-tag">
                  {{
                    kb.share
                      ? $t('knowledgeManage.public')
                      : $t('knowledgeManage.private')
                  }}
                </span>
                <span v-if="kb.orgName" class="meta-tag">{{ kb.orgName }}</span>
                <span v-if="kb.external === 1" class="meta-tag">
                  {{ $t('knowledgeManage.ribbon.external') }}
                </span>
              </div>
            </div>
            <el-tooltip
              v-if="kb.external !== 1"
              :content="$t('agent.form.metaDataFilter')"
              class="item"
              effect="dark"
              placement="top-start"
            >
              <span
                class="el-icon-setting resource-meta-btn"
                @click.stop="openMetaDataFilter(kb)"
              ></span>
            </el-tooltip>
          </div>
        </div>
        <div v-if="canExpandKnowledge" class="view-more">
          <span
            class="view-more-btn"
            @click="knowledgeExpanded = !knowledgeExpanded"
          >
            {{
              knowledgeExpanded
                ? $t(
                    'adminCenter.pageModules.appDevelopment.agent.detail.collapse',
                  )
                : $t(
                    'adminCenter.pageModules.appDevelopment.agent.detail.viewMore',
                  )
            }}
            <i
              :class="
                knowledgeExpanded ? 'el-icon-arrow-up' : 'el-icon-arrow-down'
              "
            ></i>
          </span>
        </div>
      </template>
      <EmptyState
        v-else
        :tip="
          $t('adminCenter.pageModules.appDevelopment.agent.detail.noKnowledge')
        "
      />
    </DetailCard>

    <!-- 关联工具 -->
    <DetailCard
      v-if="!isMultipleAgent"
      :title="
        $t('adminCenter.pageModules.appDevelopment.agent.detail.linkTool')
      "
    >
      <template v-if="hasTools">
        <div class="card-list">
          <div
            v-for="(tool, i) in toolList"
            :key="`tool-${i}`"
            class="resource-card is-clickable"
            @click="goTool(tool)"
          >
            <img
              :src="avatarSrc(tool.avatar?.path, toolDefaultIcon)"
              class="resource-avatar"
            />
            <div class="resource-info">
              <div class="resource-name">{{ toolDisplayName(tool) }}</div>
              <div v-if="toolDescription(tool)" class="resource-desc">
                {{ toolDescription(tool) }}
              </div>
              <div class="resource-meta">
                <span class="meta-tag">{{ toolTypeName(tool) }}</span>
                <span class="meta-tag">{{ boolText(tool.enable) }}</span>
              </div>
            </div>
          </div>
        </div>
        <div v-if="canExpandTools" class="view-more">
          <span class="view-more-btn" @click="toolExpanded = !toolExpanded">
            {{
              toolExpanded
                ? $t(
                    'adminCenter.pageModules.appDevelopment.agent.detail.collapse',
                  )
                : $t(
                    'adminCenter.pageModules.appDevelopment.agent.detail.viewMore',
                  )
            }}
            <i
              :class="toolExpanded ? 'el-icon-arrow-up' : 'el-icon-arrow-down'"
            ></i>
          </span>
        </div>
      </template>
      <EmptyState
        v-else
        :tip="$t('adminCenter.pageModules.appDevelopment.agent.detail.noTool')"
      />
    </DetailCard>

    <!-- 关联智能体 -->
    <DetailCard
      v-if="isMultipleAgent"
      :title="
        $t('adminCenter.pageModules.appDevelopment.agent.detail.linkAgent')
      "
    >
      <template v-if="hasMultiAgents">
        <div class="card-list">
          <div
            v-for="(agent, i) in multiAgentList"
            :key="`ma-${i}`"
            class="resource-card is-clickable"
            @click="goSingleAgent(agent)"
          >
            <img
              :src="avatarSrc(agent.avatar?.path, agentDefaultIcon)"
              class="resource-avatar"
            />
            <div class="resource-info">
              <div class="resource-name">{{ agent.name }}</div>
              <div v-if="agent.desc" class="resource-desc">
                {{ agent.desc }}
              </div>
              <div class="resource-meta">
                <span class="meta-tag">
                  {{ boolText(agent.enable) }}
                </span>
              </div>
            </div>
          </div>
        </div>
        <div v-if="canExpandMultiAgents" class="view-more">
          <span
            class="view-more-btn"
            @click="multiAgentExpanded = !multiAgentExpanded"
          >
            {{
              multiAgentExpanded
                ? $t(
                    'adminCenter.pageModules.appDevelopment.agent.detail.collapse',
                  )
                : $t(
                    'adminCenter.pageModules.appDevelopment.agent.detail.viewMore',
                  )
            }}
            <i
              :class="
                multiAgentExpanded ? 'el-icon-arrow-up' : 'el-icon-arrow-down'
              "
            ></i>
          </span>
        </div>
      </template>
      <EmptyState
        v-else
        :tip="$t('adminCenter.pageModules.appDevelopment.agent.detail.noAgent')"
      />
    </DetailCard>

    <!-- 安全护栏 -->
    <DetailCard
      :title="
        $t(
          'adminCenter.pageModules.appDevelopment.agent.detail.safetyGuardrail',
        )
      "
    >
      <InfoGrid :items="safetyItems" />
    </DetailCard>

    <!-- 追问配置 -->
    <DetailCard
      v-if="!isMultipleAgent"
      :title="
        $t(
          'adminCenter.pageModules.appDevelopment.agent.detail.recommendConfig',
        )
      "
    >
      <InfoGrid :items="recommendItems" />
      <div v-if="recommendConfig.promptEnable" class="recommend-prompt">
        <div class="sub-label">
          {{
            $t(
              'adminCenter.pageModules.appDevelopment.agent.detail.recommendPrompt',
            )
          }}
        </div>
        <div class="prompt-text input-box">
          {{ recommendConfig.prompt || '-' }}
        </div>
      </div>
    </DetailCard>

    <metaDataFilterField
      ref="metaDataFilterField"
      :category="currentKnowledgeCategory"
      :knowledgeId="currentKnowledgeId"
      :metaData="currentMetaData"
      :readonly="true"
    />
  </DetailLayout>
</template>

<script>
import {
  getAdminAssistantBase,
  getAdminAssistantDetail,
} from '@/api/adminCenter';
import { avatarSrc } from '@/utils/util';
import DetailLayout from '../components/DetailLayout.vue';
import DetailHeader from '../components/DetailHeader.vue';
import DetailCard from '../components/DetailCard.vue';
import InfoGrid from '../components/InfoGrid.vue';
import VisibleUsers from '../components/VisibleUsers.vue';
import EmptyState from '../components/EmptyState.vue';
import metaDataFilterField from '@/components/app/metaDataFilterField.vue';
import { MULTIPLE_AGENT, SINGLE_AGENT } from '@/views/agent/constants';
import detailMixin from '../mixins/detailMixin';

const VIEW_MORE_THRESHOLD = 6;

export default {
  components: {
    DetailLayout,
    DetailHeader,
    DetailCard,
    InfoGrid,
    VisibleUsers,
    EmptyState,
    metaDataFilterField,
  },
  mixins: [detailMixin],
  data() {
    return {
      base: {},
      detail: {},
      knowledgeDefaultIcon: require('@/assets/imgs/knowledgeIcon.png'),
      toolDefaultIcon: require('@/assets/imgs/toolImg.png'),
      agentDefaultIcon: require('@/assets/imgs/agent.svg'),
      toolExpanded: false,
      knowledgeExpanded: false,
      multiAgentExpanded: false,
      currentKnowledgeId: '',
      currentKnowledgeCategory: 0,
      currentMetaData: {},
      moduleTitleKey: 'adminCenter.pageModules.appDevelopment.agent.title',
    };
  },
  computed: {
    assistantId() {
      return this.$route.query.assistantId;
    },
    avatarUrl() {
      return avatarSrc(
        this.base.avatar?.path,
        require('@/assets/imgs/bg-logo.png'),
      );
    },
    categoryText() {
      const map = {
        [SINGLE_AGENT]: this.$t('agentDialog.singleAgent'),
        [MULTIPLE_AGENT]: this.$t('agentDialog.multipleAgent'),
      };
      return map[this.base.category] ?? '-';
    },
    isMultipleAgent() {
      return this.base.category === MULTIPLE_AGENT;
    },
    multiAgentList() {
      const list = this.detail.multiAgentInfos || [];
      return this.multiAgentExpanded
        ? list
        : list.slice(0, VIEW_MORE_THRESHOLD);
    },
    hasMultiAgents() {
      return (this.detail.multiAgentInfos || []).length > 0;
    },
    canExpandMultiAgents() {
      return (this.detail.multiAgentInfos || []).length > VIEW_MORE_THRESHOLD;
    },
    headerTags() {
      const tags = [];
      if (this.base.publishStatus)
        tags.push({ text: this.statusText, type: '' });
      if (this.base.category)
        tags.push({ text: this.categoryText, type: 'info' });
      return tags;
    },
    basicItems() {
      return [
        ...this.baseBasicItems(),
        {
          label: this.$t('agent.form.hideKnowledge'),
          value: this.boolText(this.base.hideKnowledge),
        },
      ];
    },
    modelConfig() {
      return this.detail.modelConfig ?? {};
    },
    modelLlm() {
      const cfg = this.modelConfig.config;
      if (cfg && typeof cfg === 'object') return cfg;
      return this.modelConfig;
    },
    modelItems() {
      const c = this.modelLlm;
      return [
        {
          label: this.$t(
            'adminCenter.pageModules.appDevelopment.agent.detail.model',
          ),
          model: {
            icon: this.convertModelIcon(this.modelConfig.avatar?.path),
            name: this.modelConfig.displayName,
            tags: this.modelConfigTags,
          },
        },
        {
          label: this.$t(
            'adminCenter.pageModules.appDevelopment.agent.detail.temperature',
          ),
          value: this.numText(c.temperature),
        },
        {
          label: this.$t(
            'adminCenter.pageModules.appDevelopment.agent.detail.topP',
          ),
          value: this.numText(c.topP),
        },
        {
          label: this.$t(
            'adminCenter.pageModules.appDevelopment.agent.detail.maxTokens',
          ),
          value: this.numText(c.maxTokens),
        },
        {
          label: this.$t(
            'adminCenter.pageModules.appDevelopment.agent.detail.frequencyPenalty',
          ),
          value: this.numText(c.frequencyPenalty),
        },
        {
          label: this.$t(
            'adminCenter.pageModules.appDevelopment.agent.detail.presencePenalty',
          ),
          value: this.numText(c.presencePenalty),
        },
        {
          label: this.$t(
            'adminCenter.pageModules.appDevelopment.agent.detail.thinking',
          ),
          value: this.boolText(c.thinkingEnable),
        },
      ];
    },
    modelConfigTags() {
      const tags = [];
      if (this.modelConfig.modelType)
        tags.push({ text: this.modelConfig.modelType });
      if (this.modelConfig.provider)
        tags.push({ text: this.modelConfig.provider });
      return tags;
    },
    recommendQuestions() {
      return (this.detail.recommendQuestion || []).filter(Boolean);
    },
    knowledgeConfig() {
      return this.detail.knowledgeBaseConfig ?? {};
    },
    recallConfig() {
      return this.knowledgeConfig.config ?? {};
    },
    knowledgeList() {
      const list = this.knowledgeConfig.knowledgebases || [];
      return this.knowledgeExpanded ? list : list.slice(0, VIEW_MORE_THRESHOLD);
    },
    hasKnowledge() {
      return (this.knowledgeConfig.knowledgebases || []).length > 0;
    },
    canExpandKnowledge() {
      return (
        (this.knowledgeConfig.knowledgebases || []).length > VIEW_MORE_THRESHOLD
      );
    },
    matchTypeText() {
      const map = {
        vector: this.$t(
          'adminCenter.pageModules.appDevelopment.agent.detail.matchTypeOptions.vector',
        ),
        text: this.$t(
          'adminCenter.pageModules.appDevelopment.agent.detail.matchTypeOptions.text',
        ),
        mix: this.$t(
          'adminCenter.pageModules.appDevelopment.agent.detail.matchTypeOptions.mix',
        ),
      };
      return (
        map[this.recallConfig.matchType] ?? this.recallConfig.matchType ?? '-'
      );
    },
    recallItems() {
      const c = this.recallConfig;
      const items = [
        {
          label: this.$t(
            'adminCenter.pageModules.appDevelopment.agent.detail.matchType',
          ),
          value: this.matchTypeText,
        },
        {
          label: this.$t(
            'adminCenter.pageModules.appDevelopment.agent.detail.topK',
          ),
          value: this.numText(c.topK),
        },
        {
          label: this.$t(
            'adminCenter.pageModules.appDevelopment.agent.detail.scoreThreshold',
          ),
          value: this.numText(c.threshold),
        },
      ];
      if (this.recallConfig.matchType === 'mix') {
        items.push(
          {
            label: this.$t(
              'adminCenter.pageModules.appDevelopment.agent.detail.mixMode.weight',
            ),
            value: this.numText(c.priorityMatch),
          },
          {
            label: this.$t(
              'adminCenter.pageModules.appDevelopment.agent.detail.semanticsPriority',
            ),
            value: this.numText(c.semanticsPriority),
          },
          {
            label: this.$t(
              'adminCenter.pageModules.appDevelopment.agent.detail.keywordPriority',
            ),
            value: this.numText(c.keywordPriority),
          },
        );
      }
      if (this.rerankConfig.displayName) {
        items.push({
          label: this.$t(
            'adminCenter.pageModules.appDevelopment.agent.detail.rerankModel',
          ),
          model: {
            icon: this.convertModelIcon(this.rerankConfig.avatar?.path),
            name: this.rerankConfig.displayName,
            tags: this.rerankModelTags,
          },
        });
      }
      items.push({
        label: this.$t(
          'adminCenter.pageModules.appDevelopment.agent.detail.useGraph',
        ),
        value: this.boolText(c.useGraph),
      });
      return items;
    },
    rerankConfig() {
      return this.detail.rerankConfig ?? {};
    },
    rerankModelTags() {
      const tags = [];
      if (this.rerankConfig.modelType)
        tags.push({ text: this.rerankConfig.modelType });
      if (this.rerankConfig.provider)
        tags.push({ text: this.rerankConfig.provider });
      return tags;
    },
    toolList() {
      const list = [];
      const push = (arr, kind) =>
        (arr || []).forEach(item => list.push({ ...item, kind }));
      push(this.detail.workFlowInfos, 'workflow');
      push(this.detail.mcpInfos, 'mcp');
      push(this.detail.toolInfos, 'tool');
      push(this.detail.skillInfos, 'skill');
      return this.toolExpanded ? list : list.slice(0, VIEW_MORE_THRESHOLD);
    },
    toolTotal() {
      const d = this.detail;
      return (
        (d.workFlowInfos?.length ?? 0) +
        (d.mcpInfos?.length ?? 0) +
        (d.toolInfos?.length ?? 0) +
        (d.skillInfos?.length ?? 0)
      );
    },
    hasTools() {
      return this.toolTotal > 0;
    },
    canExpandTools() {
      return this.toolTotal > VIEW_MORE_THRESHOLD;
    },
    safetyConfig() {
      return this.detail.safetyConfig ?? {};
    },
    safetyItems() {
      const items = [
        {
          label: this.$t('adminCenter.common.status'),
          value: this.boolText(this.safetyConfig.enable),
        },
      ];
      if (this.safetyConfig.enable) {
        items.push({
          label: this.$t(
            'adminCenter.pageModules.appDevelopment.agent.detail.sensitiveTables',
          ),
          tags: (this.safetyConfig.tables || []).map(t => ({
            text: t.tableName,
            url: t.tableId
              ? `/adminCenter/safety/detail?tableId=${t.tableId}`
              : '',
          })),
        });
      }
      return items;
    },
    recommendConfig() {
      return this.detail.recommendConfig ?? {};
    },
    recommendModelConfig() {
      return this.recommendConfig.modelConfig ?? {};
    },
    recommendItems() {
      const rc = this.recommendConfig;
      const rmc = this.recommendModelConfig;
      const items = [
        {
          label: this.$t('adminCenter.common.status'),
          value: this.boolText(rc.recommendEnable),
        },
      ];
      if (!rc.recommendEnable) return items;
      items.push(
        {
          label: this.$t(
            'adminCenter.pageModules.appDevelopment.agent.detail.recommendModel',
          ),
          model: {
            icon: this.convertModelIcon(rmc.avatar?.path),
            name: rmc.displayName,
            tags: rmc.modelType ? [{ text: rmc.modelType }] : [],
          },
        },
        {
          label: this.$t(
            'adminCenter.pageModules.appDevelopment.agent.detail.maxHistory',
          ),
          value: this.numText(rc.maxHistory),
        },
        {
          label: this.$t(
            'adminCenter.pageModules.appDevelopment.agent.detail.promptEnable',
          ),
          value: this.boolText(rc.promptEnable),
        },
      );
      return items;
    },
  },
  mounted() {
    this.fetchData();
  },
  methods: {
    toolDisplayName(tool) {
      switch (tool.kind) {
        case 'workflow':
          return tool.name;
        case 'mcp':
          return tool.mcpName;
        case 'skill':
          return tool.skillName;
        case 'tool':
          return tool.toolName;
      }
    },
    toolDescription(tool) {
      return tool.description || tool.workFlowDesc || tool.desc || '';
    },
    toolTypeName(tool) {
      const map = {
        workflow: this.$t('appSpace.workflow'),
        mcp: 'MCP',
        tool: this.$t('agent.toolDialog.tool'),
        skill: this.$t('tempSquare.skills.name'),
      };
      return map[tool.kind] ?? '-';
    },
    goKnowledge(kb) {
      if (!kb?.id) return;
      const routeUrl = this.$router.resolve({
        path: '/adminCenter/knowledge/detail',
        query: { knowledgeId: kb.id },
      });
      window.open(routeUrl.href, '_blank');
    },
    openMetaDataFilter(kb) {
      this.currentKnowledgeId = kb.id;
      this.currentKnowledgeCategory = kb.category ?? 0;
      this.currentMetaData = {};
      this.$nextTick(() => {
        this.currentMetaData = kb.metaDataFilterParams;
        this.$refs.metaDataFilterField.showDialog();
      });
    },
    goTool(tool) {
      const routes = {
        workflow: {
          path: '/workflow',
          query: { id: tool.workFlowId, readonly: true },
        },
        mcp: {
          path: '/adminCenter/mcp/detail',
          query: { mcpId: tool.mcpId, type: tool.mcpType },
        },
        tool: {
          path: '/adminCenter/tool/detail',
          query: { toolId: tool.toolId, type: tool.toolType },
        },
        skill: {
          path: '/adminCenter/skill/detail',
          query: { skillId: tool.skillId, type: tool.skillType },
        },
      };
      const route = routes[tool.kind];
      if (!route || !Object.values(route.query).every(v => v != null)) return;
      const routeUrl = this.$router.resolve(route);
      window.open(routeUrl.href, '_blank');
    },
    goSingleAgent(agent) {
      if (!agent?.agentId) return;
      const routeUrl = this.$router.resolve({
        path: '/adminCenter/agent/detail',
        query: { assistantId: agent.agentId },
      });
      window.open(routeUrl.href, '_blank');
    },
    fetchData() {
      if (!this.assistantId) return;
      const params = { assistantId: this.assistantId };
      getAdminAssistantBase(params).then(res => {
        this.base = res?.data ?? {};
      });
      getAdminAssistantDetail(params).then(res => {
        this.detail = res?.data ?? {};
      });
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/views/adminCenter/styles/common.scss';

.sub-label {
  color: #909399;
  font-size: 14px;
  margin: 20px 0 12px;
}

.recommend-prompt {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #ebeef5;
}

.resource-card {
  .resource-meta-btn {
    flex-shrink: 0;
    color: #909399;
    font-size: 16px;
    cursor: pointer;

    &:hover {
      color: #409eff;
    }
  }
}
</style>
