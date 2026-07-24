<template>
  <div
    class="page-wrapper wrap-fullheight modelExprience-wrapper right-page-content-body"
  >
    <ConversationPane
      ref="modelExprienceRef"
      mode="modelExprience"
      :modelOptions="modelOptions"
      :modelExperienceId.sync="modelExperienceId"
      @openModelComparison="openModelComparison"
      @refreshHistoryList="getChatHistoryList"
      @pinCurrentChat="handlePinCurrentChat"
    >
      <template #nav>
        <div class="history-wrapper">
          <el-button size="mini" type="primary" @click="createConversation">
            <span class="el-icon-plus"></span>
            {{ $t('modelExprience.createSession') }}
          </el-button>
          <div class="history-list" v-loading="loading" ref="historyList">
            <el-empty
              v-if="!chatList.length"
              :description="$t('common.noData')"
            ></el-empty>
            <template v-else>
              <div
                v-for="(item, index) in chatList"
                :key="index"
                class="history-item"
                :class="[modelExperienceId === item.id && 'active']"
                @click="handleConvesationFeedback(item)"
              >
                <div class="name">{{ item.title }}</div>
                <div class="time">
                  {{ item.createdAt ? formatTimestamp(item.createdAt) : '-' }}
                </div>
                <i
                  class="icon el-icon-close"
                  @click.stop="handleDeleteHistory(item)"
                ></i>
              </div>
            </template>
          </div>
        </div>
      </template>
    </ConversationPane>
    <transition name="fade-slide" mode="out-in">
      <ConversationPane
        v-if="comparisonIds.length"
        ref="modelComparisonRef"
        mode="modelComparison"
        modelExperienceId="0"
        :modelOptions="modelOptions"
        :comparisonIds="comparisonIds"
        @closeModelComparison="closeModelComparison"
      />
    </transition>
  </div>
</template>
<script>
import ConversationPane from './ConversationPane.vue';
import { fetchChatList, deleteChat } from '@/api/modelExprience';
import { selectModelList } from '@/api/modelAccess';
import { formatTimestamp } from '@/utils/util';
export default {
  name: 'ModelExprience',
  components: {
    ConversationPane,
  },
  data() {
    return {
      loading: false,
      chatList: [],
      modelOptions: [],
      comparisonIds: [],
      modelExperienceId: '',
    };
  },
  mounted() {
    this.getChatHistoryList();
    this.getModelList();
  },
  methods: {
    formatTimestamp,
    openModelComparison(ids) {
      this.comparisonIds = ids;
    },
    closeModelComparison() {
      this.comparisonIds = [];
    },
    async getChatHistoryList() {
      this.loading = true;
      let res = await fetchChatList().finally(() => {
        this.loading = false;
      });
      if (res.code === 0) {
        if (res.data && res.data.list && res.data.list.length > 0) {
          this.chatList = res.data.list;
        } else {
          this.chatList = [];
        }
      }
    },
    // 发送消息后直接把当前会话置顶
    handlePinCurrentChat() {
      const activeId = this.modelExperienceId;
      if (!activeId || this.chatList.length <= 1) {
        return;
      }
      const idx = this.chatList.findIndex(item => item.id === activeId);
      if (idx <= 0) {
        return;
      }
      const [current] = this.chatList.splice(idx, 1);
      this.chatList.unshift(current);
      // 滚动条置顶
      this.$nextTick(() => {
        if (this.$refs.historyList) {
          this.$refs.historyList.scrollTop = 0;
        }
      });
    },
    handleConvesationFeedback(chatItem) {
      this.modelExperienceId = chatItem.id;
      this.$refs.modelExprienceRef.initConversation(chatItem);
    },
    handleDeleteHistory(chatItem) {
      this.$confirm(
        this.$t('modelExprience.warning.deleteHistory'),
        this.$t('common.confirm.title'),
        {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          type: 'warning',
        },
      ).then(() => {
        this.loading = true;
        deleteChat({ modelExperienceId: chatItem.id })
          .then(() => {
            this.getChatHistoryList();
          })
          .finally(() => {
            this.loading = false;
          });
      });
    },
    createConversation() {
      this.modelExperienceId = '';
      this.$refs.modelExprienceRef.openSelectModelDialog();
    },
    getModelList() {
      selectModelList().then(res => {
        if (res.code === 0 && res.data && Array.isArray(res.data.list)) {
          this.modelOptions = res.data.list || [];
        }
      });
    },
  },
};
</script>
<style lang="scss" scoped>
.modelExprience-wrapper {
  display: flex;
  flex-direction: row;
  padding: 0;
  position: relative;
  .history-wrapper {
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    width: 250px;
    background-color: #f7f7fc;
    box-shadow: 0 1px 4px 0 rgba(0, 0, 0, 0.15);
    border-radius: 8px;
    padding: 20px 0;
    margin-right: 24px;
    .el-button {
      margin: 0 16px;
    }
    .history-list {
      flex: 1;
      padding: 16px;
      overflow: auto;
      .history-item {
        position: relative;
        border: 1px solid transparent;
        box-shadow: 0 1px 4px 0 rgba(0, 0, 0, 0.15);
        padding: 10px;
        margin-bottom: 15px;
        border-radius: 4px;
        cursor: pointer;
        &.active {
          border-color: $border_color;
          background: $color_opacity;
        }
        &:hover {
          background: $color_opacity;
          .icon {
            opacity: 1;
          }
        }
        .name {
          color: $color_title;
          font-size: 14px;
          overflow: hidden;
          text-overflow: ellipsis;
          display: -webkit-box;
          -webkit-box-orient: vertical;
          line-clamp: 1;
          -webkit-line-clamp: 1;
          cursor: pointer;
        }
        .time {
          margin-top: 10px;
          color: #878aab;
        }
        .icon {
          opacity: 0;
          position: absolute;
          right: 4px;
          top: 4px;
          font-size: 16px;
          cursor: pointer;
        }
      }
    }
  }
  /* 过渡动画样式 */
  .fade-slide-enter-active,
  .fade-slide-leave-active {
    transition: all 0.3s ease;
  }
  .fade-slide-enter-from,
  .fade-slide-leave-to {
    opacity: 0;
    transform: translateY(10px);
  }
}
</style>
