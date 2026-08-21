<template>
  <aside class="skill-workspace-container">
    <div class="workspace-header">
      <div class="workspace-title">
        <i class="el-icon-folder-opened"></i>
        <span>
          {{ $t('generalAgent.skill.panel.skillWorkspace') }}
        </span>
      </div>
      <div class="header-actions">
        <div class="security-review-action">
          <el-button
            type="primary"
            size="small"
            :disabled="securityReviewDisabled"
            @click="handleSecurityReview"
          >
            {{ $t('generalAgent.skill.skillWorkBench.securityReview.btn') }}
          </el-button>
          <el-tooltip
            :content="
              $t('generalAgent.skill.skillWorkBench.securityReview.tip')
            "
            placement="top"
          >
            <i class="el-icon-info security-review-tip"></i>
          </el-tooltip>
        </div>

        <AppPublishActions
          ref="publishActions"
          :appId="skillPreviewParams.customSkillId"
          :appType="SKILL"
          :appName="assistantInfo.name"
          :publishType="publishType"
          :beforePublish="beforeSkillPublish"
          :disabled="publishDisabled"
          @reload-data="reloadData"
          @preview-version="previewVersion"
        />
      </div>
    </div>

    <SkillQuickPublishDialog
      v-model="quickPublishVisible"
      :customSkillId="skillPreviewParams.customSkillId"
      :files="quickPublishFiles"
      :publishData="quickPublishData"
      :publishHandler="executeQuickPublish"
      :disabled="publishDisabled"
      @view-git-details="viewGitDetails"
      @git-changed="handleQuickPublishGitChanged"
    />

    <div class="workspace-body" :class="{ 'disable-clicks': disableClick }">
      <SkillWorkspaceExplorer
        ref="explorer"
        :customSkillId="skillPreviewParams.customSkillId"
        :activeGitDiffId="activeGitDiffId"
        @open-file="openFile"
        @open-search-result="openSearchResult"
        @open-git-diff="openGitDiff"
        @close-tabs-by-path="closeTabsByPath"
        @discard-file="handleDiscardFile"
        @workspace-restored="handleWorkspaceRestored"
      />
      <SkillWorkbench
        ref="workbench"
        :skillPreviewParams="skillPreviewParams"
        @active-git-diff-change="activeGitDiffId = $event"
        @file-saved="handleFileSaved"
        @view-workspace="$emit('view-workspace', $event)"
        @download-all="$emit('download-all')"
      />
    </div>
  </aside>
</template>

<script>
import SkillWorkspaceExplorer from './SkillWorkspaceExplorer.vue';
import SkillWorkbench from './SkillWorkbench.vue';
import SkillQuickPublishDialog from './SkillQuickPublishDialog.vue';
import AppPublishActions from '@/components/appPublishActions.vue';
import { getCustomSkillInfo } from '@/api/templateSquare';
import { getSkillWorkspaceGitStatus } from '@/api/skillResource/skillWorkSpace';
import { SKILL } from '@/utils/commonSet';
import skillManager from '../../mixins/skillManager';

export default {
  name: 'SkillTabs',
  mixins: [skillManager],
  components: {
    SkillWorkspaceExplorer,
    SkillWorkbench,
    SkillQuickPublishDialog,
    AppPublishActions,
  },
  props: {
    skillPreviewParams: {
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    return {
      SKILL,
      publishType: '',
      disableClick: false,
      version: '',
      assistantInfo: {},
      activeGitDiffId: '',
      quickPublishVisible: false,
      quickPublishFiles: [],
      quickPublishData: {},
    };
  },
  computed: {
    securityReviewDisabled() {
      return this.mainIsStreaming || this.previewIsStreaming;
    },
    publishDisabled() {
      return this.mainIsStreaming;
    },
  },
  watch: {
    'skillPreviewParams.customSkillId': {
      handler(val) {
        if (val) {
          this.getAppDetail();
        }
      },
      immediate: true,
    },
  },
  methods: {
    openFile(file) {
      if (this.$refs.workbench) {
        this.$refs.workbench.openFile(file);
      }
    },
    openSearchResult(payload) {
      if (this.$refs.workbench) {
        this.$refs.workbench.openSearchResult(payload);
      }
    },
    openGitDiff(payload) {
      if (this.$refs.workbench) {
        this.$refs.workbench.openGitDiff(payload);
      }
    },
    refreshFiles() {
      if (this.$refs.explorer) {
        this.$refs.explorer.refreshFiles();
      }
    },
    closeTabsByPath(path) {
      if (this.$refs.workbench) {
        this.$refs.workbench.closeTabsByPath(path);
      }
    },
    async handleDiscardFile(payload) {
      if (this.$refs.workbench) {
        await this.$refs.workbench.refreshOpenedFileByPath(payload);
      }
    },
    async handleWorkspaceRestored() {
      await this.refreshWorkspace();
    },
    async refreshWorkspace() {
      this.refreshFiles();
      if (this.$refs.workbench) {
        await this.$refs.workbench.refreshOpenedFiles({ force: true });
      }
      if (this.$refs.explorer) {
        this.refreshGit();
      }
    },
    refreshGit() {
      if (this.$refs.explorer) {
        this.$refs.explorer.refreshGit();
      }
    },
    hasUnsavedFiles() {
      return this.$refs.workbench
        ? this.$refs.workbench.hasUnsavedFiles()
        : false;
    },
    async discardUnsavedFiles() {
      if (!this.$refs.workbench) return [];
      return this.$refs.workbench.discardUnsavedFiles();
    },
    reloadData() {
      this.disableClick = false;
      this.getAppDetail();
      this.$emit('refresh-workspace');
      this.refreshWorkspace();
    },
    previewVersion(item) {
      this.disableClick = !item.isCurrent;
      this.version = item.version || '';
      this.getAppDetail();
    },
    handleFileSaved() {
      this.refreshGit();
    },
    async beforeSkillPublish(data) {
      if (this.publishDisabled) return false;

      if (this.hasUnsavedFiles()) {
        this.$message.warning(
          this.$t(
            'generalAgent.skill.skillWorkBench.quickPublish.unsavedFiles',
          ),
        );
        return false;
      }

      try {
        const res = await getSkillWorkspaceGitStatus(
          this.skillPreviewParams.customSkillId,
        );
        if (res.code !== 0) return false;

        const files = (res.data && res.data.files) || [];
        if (files.length === 0) return true;

        this.quickPublishData = { ...data };
        this.quickPublishFiles = files;
        this.quickPublishVisible = true;
        return false;
      } catch (error) {
        console.error('skill publish precheck error', error);
        return false;
      }
    },
    executeQuickPublish(data) {
      if (this.publishDisabled) {
        return Promise.resolve(null);
      }
      if (!this.$refs.publishActions) {
        return Promise.resolve(null);
      }
      return this.$refs.publishActions.executePublish(data);
    },
    viewGitDetails() {
      if (this.$refs.explorer) {
        this.$refs.explorer.showGitView();
      }
    },
    handleQuickPublishGitChanged() {
      this.refreshGit();
    },
    handleSecurityReview() {
      if (this.securityReviewDisabled) return;
      this.$emit('security-review');
    },
    async getAppDetail() {
      const params = {
        skillId: this.skillPreviewParams.customSkillId,
      };
      const res = await getCustomSkillInfo(params);

      if (res.code === 0 && res.data) {
        this.assistantInfo = res.data;
        this.publishType = res.data.publishType;
      }
    },
  },
};
</script>

<style lang="scss" scoped>
@import '../../styles/variables';

.skill-workspace-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  min-width: 0;
  background: #fff;
  border-left: 1px solid #f0f0f0;
  position: relative;
  z-index: 10;
}

.workspace-header {
  height: $header-height;
  position: relative;
  padding: 0 16px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  flex-shrink: 0;

  .workspace-title {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    color: #333;
    font-size: 14px;
    font-weight: 600;

    i {
      color: #10a37f;
      font-size: 16px;
      flex-shrink: 0;
    }
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;

    .security-review-action {
      display: flex;
      align-items: center;
      gap: 6px;
    }

    .security-review-tip {
      color: #909399;
      font-size: 15px;
      line-height: 1;

      &:hover {
        color: #606266;
      }
    }
  }
}

.workspace-body {
  flex: 1;
  min-height: 0;
  min-width: 0;
  overflow: hidden;
  position: relative;
  display: flex;
  background: #fff;

  &.disable-clicks {
    pointer-events: none;
    opacity: 0.7;
    filter: grayscale(0.5);
  }
}
</style>
