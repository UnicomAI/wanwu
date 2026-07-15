<template>
  <CommonLayout
    :isButton="false"
    :showAside="false"
    class="right-page-content-body"
  >
    <template #main-content>
      <div class="app-content">
        <Chat
          :editForm="editForm"
          :chatType="'chat'"
          :maxPicNum="currentMaxPicNum"
          :maxImageSize="currentMaxImageSize"
        />
      </div>
    </template>
  </CommonLayout>
</template>
<script>
import CommonLayout from '@/components/exploreContainer.vue';
import Chat from './components/chat.vue';
import { getRagPublishedInfo } from '@/api/rag';
import { selectModelList } from '@/api/modelAccess';
export default {
  name: 'ExploreRag',
  components: { CommonLayout, Chat },
  data() {
    return {
      editForm: {
        appId: '',
        avatar: {},
        name: '',
        desc: '',
        modelParams: '',
        visionsupport: '',
        modelConfig: {},
        visionConfig: {
          picNum: 0,
        },
        knowledgeBaseConfig: { config: {}, knowledgebases: [] },
        qaKnowledgeBaseConfig: { config: {}, knowledgebases: [] },
        recommendQuestion: [],
      },
      modelOptions: [],
    };
  },
  computed: {
    currentModelId() {
      return (
        this.editForm.modelConfig?.modelId || this.editForm.modelParams || ''
      );
    },
    currentModelInfo() {
      return (
        this.modelOptions.find(item => item.modelId === this.currentModelId) ||
        this.editForm.modelConfig ||
        {}
      );
    },
    currentModelFullConfig() {
      return (
        this.currentModelInfo.fullConfig || this.currentModelInfo.config || {}
      );
    },
    currentMaxPicNum() {
      const visionSupport = this.currentModelFullConfig.visionSupport;
      if (visionSupport === 'support') return 3;
      if (visionSupport) return -1;
      return 1;
    },
    currentMaxImageSize() {
      const size = Number(this.currentModelFullConfig.maxImageSize);
      return size > 0 ? size : null;
    },
  },
  created() {
    if (this.$route.query.id) {
      this.editForm.appId = this.$route.query.id;
      this.getDetail();
    }
  },

  methods: {
    async getDetail() {
      const res = await getRagPublishedInfo({ ragId: this.editForm.appId });
      if (res.code === 0) {
        this.editForm.avatar = res.data.avatar;
        this.editForm.name = res.data.name;
        this.editForm.desc = res.data.desc;
        this.editForm.visionConfig = res.data.visionConfig || { picNum: 0 };
        this.$set(this.editForm, 'modelConfig', res.data.modelConfig || {});
        this.editForm.modelParams = res.data.modelConfig?.modelId || '';
        if (res.data.knowledgeBaseConfig) {
          this.editForm.knowledgeBaseConfig = res.data.knowledgeBaseConfig;
        }
        if (res.data.qaKnowledgeBaseConfig) {
          this.editForm.qaKnowledgeBaseConfig = res.data.qaKnowledgeBaseConfig;
        }
        this.editForm.recommendQuestion = res.data.recommendQuestion?.map(
          item => ({
            value: item,
          }),
        );
        await this.getModelData();
      }
    },
    async getModelData() {
      try {
        const res = await selectModelList();
        if (res.code === 0) {
          this.modelOptions = res.data.list || [];
          this.applyModelFullConfig();
        }
      } catch (error) {
        console.warn('[rag chat] get model list failed', error);
      }
    },
    applyModelFullConfig() {
      const modelId = this.currentModelId;
      if (!modelId) return;
      const selectedModel = this.modelOptions.find(
        item => item.modelId === modelId,
      );
      if (!selectedModel) return;
      this.editForm.visionsupport = selectedModel.config?.visionSupport || '';
      this.$set(this.editForm, 'modelConfig', {
        ...(this.editForm.modelConfig || {}),
        fullConfig: selectedModel.config || {},
      });
    },
    goBack() {
      this.$router.go(-1);
    },
  },
};
</script>
<style lang="scss" scoped>
::v-deep {
  .apikeyBtn {
    padding: 11px 10px;
    border: 1px solid $btn_bg;
    color: $btn_bg;
    display: flex;
    align-items: center;
    img {
      height: 14px;
    }
  }
}
.app-content {
  width: 100%;
  height: 100%;
  position: relative;
  .app-header-api {
    width: 100%;
    padding: 10px;
    position: absolute;
    z-index: 999;
    top: 0;
    left: 0;
    border-bottom: 1px solid #eaeaea;
    display: flex;
    justify-content: space-between;
    align-content: center;
    .app_name {
      font-size: 18px;
      font-weight: bold;
      color: $color_title;
      display: flex;
      align-items: center;
      .goBack {
        font-weight: bold;
        font-size: 16px;
        cursor: pointer;
        margin-right: 15px;
        color: #333;
      }
    }
    .header-api-box {
      display: flex;
      .header-api-url {
        padding: 6px 10px;
        background: #fff;
        margin: 0 10px;
        border-radius: 6px;
        .root-url {
          background-color: #eceefe;
          color: $color;
          border: none;
        }
      }
    }
  }
}
</style>
