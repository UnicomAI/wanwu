<template>
  <div class="skill-workbench">
    <div class="workbench-tabs">
      <div
        v-for="tab in allTabs"
        :key="tab.id"
        :class="['tab-item', { active: activeTab === tab.id }]"
        @click="activateTab(tab.id)"
      >
        <i
          v-if="tab.icon"
          :class="[tab.icon, 'tab-icon']"
          :style="{ color: tab.iconColor || '' }"
        ></i>
        <span class="tab-name" :title="tab.title">{{ tab.title }}</span>
        <span v-if="tab.modified" class="tab-modified-dot"></span>
        <i
          v-if="tab.closable"
          class="el-icon-close tab-close"
          @click.stop="closeTab(tab.id)"
        ></i>
      </div>
    </div>

    <div class="workbench-content">
      <div v-show="activeTab === 'preview'" class="tab-content">
        <PreviewChat
          :skillPreviewParams="skillPreviewParams"
          @view-workspace="$emit('view-workspace', $event)"
          @download-all="$emit('download-all')"
        />
      </div>

      <div v-show="activeTab === 'config'" class="tab-content">
        <SkillConfig :skillPreviewParams="skillPreviewParams" />
      </div>

      <div
        v-for="file in openedFiles"
        :key="file.path"
        v-show="activeTab === fileTabId(file.path)"
        class="tab-content"
      >
        <SkillFileEditor
          :ref="fileEditorRef(file.path)"
          :customSkillId="customSkillId"
          :file="file"
          :active="activeTab === fileTabId(file.path)"
          :highlightRequest="fileHighlightRequests[file.path]"
          @modified-change="handleFileModifiedChange"
          @meta-change="handleFileMetaChange"
          @saved="handleFileSaved"
        />
      </div>

      <div
        v-for="diff in gitDiffTabs"
        :key="diff.id"
        v-show="activeTab === diff.id"
        class="tab-content"
      >
        <SkillGitDiffTab :customSkillId="customSkillId" :diff="diff" />
      </div>
    </div>
  </div>
</template>

<script>
import PreviewChat from './preview.vue';
import SkillConfig from './config.vue';
import SkillFileEditor from './SkillFileEditor.vue';
import SkillGitDiffTab from './SkillGitDiffTab.vue';
import { getFileIcon } from '@/utils/fileIcons';
import { MAX_FILE_SIZE_BYTES } from './workspaceConstants';

export default {
  name: 'SkillWorkbench',
  components: {
    PreviewChat,
    SkillConfig,
    SkillFileEditor,
    SkillGitDiffTab,
  },
  props: {
    skillPreviewParams: {
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    return {
      activeTab: 'preview',
      openedFiles: [],
      gitDiffTabs: [],
      fileModified: {},
      fileHighlightRequests: {},
    };
  },
  computed: {
    customSkillId() {
      return this.skillPreviewParams.customSkillId || '';
    },
    fixedTabs() {
      return [
        {
          id: 'preview',
          title: this.$t('generalAgent.skill.panel.preview'),
          icon: 'el-icon-cpu',
          closable: false,
        },
        {
          id: 'config',
          title: this.$t('generalAgent.skill.panel.variableConfig'),
          icon: 'el-icon-setting',
          closable: false,
        },
      ];
    },
    fileTabs() {
      return this.openedFiles.map(file => {
        const icon = getFileIcon(file.name || file.path.split('/').pop());
        return {
          id: this.fileTabId(file.path),
          title: file.name || file.path.split('/').pop(),
          icon: icon.icon,
          iconColor: icon.color,
          closable: true,
          modified: !!this.fileModified[file.path],
        };
      });
    },
    diffTabs() {
      return this.gitDiffTabs.map(diff => ({
        id: diff.id,
        title:
          diff.type === 'commit'
            ? this.$t('generalAgent.skill.skillWorkBench.git.commitTabTitle', {
                title: diff.title,
              })
            : `Diff ${diff.title}`,
        icon: 'el-icon-document-copy',
        closable: true,
        modified: false,
      }));
    },
    allTabs() {
      return [...this.fixedTabs, ...this.fileTabs, ...this.diffTabs];
    },
    activeGitDiffId() {
      return this.gitDiffTabs.some(diff => diff.id === this.activeTab)
        ? this.activeTab
        : '';
    },
  },
  watch: {
    activeGitDiffId: {
      handler(val) {
        this.$emit('active-git-diff-change', val);
      },
      immediate: true,
    },
    'skillPreviewParams.customSkillId'(val, oldVal) {
      if (val !== oldVal) {
        this.resetWorkspaceTabs();
      }
    },
  },
  methods: {
    activateTab(id) {
      this.activeTab = id;
    },
    fileTabId(path) {
      return `file:${path}`;
    },
    fileEditorRef(path) {
      return `fileEditor:${encodeURIComponent(path)}`;
    },
    normalizeWorkspacePath(path) {
      return String(path || '')
        .replace(/\\/g, '/')
        .replace(/\/+$/, '');
    },
    isSameOrChildPath(path, targetPath) {
      const normalizedPath = this.normalizeWorkspacePath(path);
      const normalizedTarget = this.normalizeWorkspacePath(targetPath);
      if (!normalizedPath || !normalizedTarget) return false;
      return (
        normalizedPath === normalizedTarget ||
        normalizedPath.startsWith(`${normalizedTarget}/`)
      );
    },
    getFileEditor(path) {
      const editorRef = this.$refs[this.fileEditorRef(path)];
      return Array.isArray(editorRef) ? editorRef[0] : editorRef;
    },
    hasUnsavedFiles() {
      return this.openedFiles.some(file => !!this.fileModified[file.path]);
    },
    hasUnsavedFilesByPath(path) {
      return this.openedFiles.some(
        file =>
          this.isSameOrChildPath(file.path, path) &&
          !!this.fileModified[file.path],
      );
    },
    async discardUnsavedFiles() {
      return this.refreshOpenedFiles({ force: true });
    },
    async refreshOpenedFileByPath(payload) {
      const path =
        typeof payload === 'string' ? payload : payload && payload.path;
      if (!path) return null;

      const file = this.openedFiles.find(item => item.path === path);
      if (!file) return null;

      const editor = this.getFileEditor(path);
      if (!editor || !editor.loadFile) {
        return { failed: true, path };
      }

      const result = await editor.loadFile({ force: true, silent: true });
      if (result && result.failed && payload && payload.closeIfMissing) {
        this.doCloseFileTab(path);
      } else if (result && result.failed) {
        this.$message.warning(
          this.$t('generalAgent.skill.skillWorkBench.editor.refreshFailed', {
            count: 1,
          }),
        );
      }
      return result;
    },
    async refreshOpenedFiles(options = {}) {
      const { force = false } = options;
      const results = await Promise.all(
        this.openedFiles.map(async file => {
          const editor = this.getFileEditor(file.path);
          if (!editor || !editor.loadFile) {
            return { failed: true, path: file.path };
          }
          return editor.loadFile({ force, silent: true });
        }),
      );

      const failed = results.filter(item => item && item.failed);
      if (failed.length > 0) {
        this.$message.warning(
          this.$t('generalAgent.skill.skillWorkBench.editor.refreshFailed', {
            count: failed.length,
          }),
        );
      }
      return results;
    },
    openFile(file) {
      if (!file || file.isDir) return;
      if (file.size > MAX_FILE_SIZE_BYTES) {
        this.$message.warning(
          this.$t('generalAgent.skill.skillWorkBench.editor.fileTooLarge'),
        );
        return;
      }

      const existing = this.openedFiles.find(item => item.path === file.path);
      if (!existing) {
        this.openedFiles.push({
          path: file.path,
          name: file.name || file.path.split('/').pop(),
          size: file.size || 0,
        });
      }
      this.activeTab = this.fileTabId(file.path);
    },
    openSearchResult({ result, keyword }) {
      if (!result || !result.path) return;
      this.openFile({
        path: result.path,
        name: result.path.split('/').pop(),
        size: result.size || 0,
      });
      this.$set(this.fileHighlightRequests, result.path, {
        line: result.line,
        keyword,
        seq: Date.now(),
      });
    },
    openGitDiff(diff) {
      if (!diff || !diff.id) return;
      const index = this.gitDiffTabs.findIndex(item => item.id === diff.id);
      if (index >= 0) {
        this.$set(this.gitDiffTabs, index, {
          ...this.gitDiffTabs[index],
          ...diff,
        });
      } else {
        this.gitDiffTabs.push(diff);
      }
      this.activeTab = diff.id;
    },
    closeTab(id) {
      if (id.startsWith('file:')) {
        const path = id.slice('file:'.length);
        if (this.fileModified[path]) {
          this.$confirm(
            this.$t(
              'generalAgent.skill.skillWorkBench.editor.unsavedCloseConfirm',
            ),
            this.$t('common.confirm.title'),
            {
              confirmButtonText: this.$t('common.confirm.confirm'),
              cancelButtonText: this.$t('common.confirm.cancel'),
              type: 'warning',
            },
          )
            .then(() => this.doCloseFileTab(path))
            .catch(() => {});
          return;
        }
        this.doCloseFileTab(path);
        return;
      }

      const index = this.gitDiffTabs.findIndex(diff => diff.id === id);
      if (index >= 0) {
        const nextTab = this.activeTab === id ? this.resolveNextTab(id) : '';
        this.gitDiffTabs.splice(index, 1);
        if (nextTab) this.activeTab = nextTab;
      }
    },
    doCloseFileTab(path) {
      const index = this.openedFiles.findIndex(file => file.path === path);
      if (index < 0) return;
      const id = this.fileTabId(path);
      const nextTab = this.activeTab === id ? this.resolveNextTab(id) : '';

      this.openedFiles.splice(index, 1);
      this.$delete(this.fileModified, path);
      this.$delete(this.fileHighlightRequests, path);

      if (nextTab) this.activeTab = nextTab;
    },
    closeTabsByPath(path) {
      const closePaths = this.openedFiles
        .filter(file => this.isSameOrChildPath(file.path, path))
        .map(file => file.path);

      closePaths.forEach(closePath => {
        this.doCloseFileTab(closePath);
      });
    },
    resolveNextTab(closedTabId) {
      const index = this.allTabs.findIndex(tab => tab.id === closedTabId);
      if (index < 0) return 'preview';

      const nextTab = this.allTabs[index + 1] || this.allTabs[index - 1];
      return nextTab ? nextTab.id : 'preview';
    },
    handleFileModifiedChange({ path, modified }) {
      this.$set(this.fileModified, path, modified);
    },
    handleFileMetaChange({ path, size }) {
      const file = this.openedFiles.find(item => item.path === path);
      if (file) {
        this.$set(file, 'size', size);
      }
    },
    handleFileSaved(path) {
      this.$set(this.fileModified, path, false);
      this.$emit('file-saved', path);
    },
    resetWorkspaceTabs() {
      this.activeTab = 'preview';
      this.openedFiles = [];
      this.gitDiffTabs = [];
      this.fileModified = {};
      this.fileHighlightRequests = {};
    },
  },
};
</script>

<style lang="scss" scoped>
.skill-workbench {
  flex: 1;
  min-width: 0;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #fff;

  .workbench-tabs {
    display: flex;
    align-items: center;
    background: #ececec;
    border-bottom: 1px solid #e0e0e0;
    overflow-x: auto;
    flex-shrink: 0;

    .tab-item {
      display: flex;
      align-items: center;
      padding: 6px 12px;
      cursor: pointer;
      border-right: 1px solid #e0e0e0;
      color: #666;
      font-size: 13px;
      white-space: nowrap;
      flex-shrink: 0;
      max-width: 220px;

      &:hover {
        background: #e0e0e0;
      }
      &.active {
        background: #fff;
        color: #333;
        box-shadow: inset 0 -2px 0 #5983ff;
      }

      .tab-icon {
        font-size: 14px;
        margin-right: 4px;
        flex-shrink: 0;
      }

      .tab-name {
        overflow: hidden;
        text-overflow: ellipsis;
      }

      .tab-modified-dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: #67c23a;
        margin-left: 6px;
        flex-shrink: 0;
      }

      .tab-close {
        font-size: 12px;
        opacity: 0.5;
        margin-left: 6px;
        flex-shrink: 0;
        &:hover {
          opacity: 1;
        }
      }
    }
  }

  .workbench-content {
    flex: 1;
    min-height: 0;
    overflow: hidden;
    position: relative;

    .tab-content {
      height: 100%;
      min-height: 0;
      overflow: hidden;
    }
  }
}
</style>
