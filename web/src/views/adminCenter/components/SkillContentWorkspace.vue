<template>
  <div v-loading="loading" class="skill-content-workspace">
    <div class="workspace-body">
      <div class="file-tree">
        <div class="file-tree-header">
          <i class="el-icon-folder"></i>
          <span>
            {{ $t('generalAgent.skill.skillWorkBench.common.files') }}
          </span>
          <i
            :class="['header-refresh', { spinning: refreshing }]"
            :title="$t('generalAgent.skill.skillWorkBench.common.refresh')"
            class="el-icon-refresh"
            @click="refresh"
          ></i>
        </div>
        <div class="file-tree-content">
          <el-tree
            v-if="treeData.length > 0"
            ref="fileTree"
            :data="treeData"
            :default-expanded-keys="defaultExpandedKeys"
            :expand-on-click-node="false"
            :props="treeProps"
            node-key="path"
            @node-click="handleNodeClick"
          >
            <span
              slot-scope="{ node, data }"
              :class="{ 'is-active': activePath === data.path && !data.isDir }"
              class="custom-tree-node"
            >
              <i
                :class="getFileIcon(data).icon"
                :style="{ color: getFileIcon(data).color }"
                class="file-icon"
              ></i>
              <el-tooltip
                :content="node.label"
                :open-delay="300"
                placement="top"
              >
                <span class="file-name">{{ node.label }}</span>
              </el-tooltip>
            </span>
          </el-tree>
          <div v-else-if="!loading" class="empty-state">
            <i class="el-icon-folder-opened"></i>
            <p>{{ $t('generalAgent.skill.skillWorkBench.fileTree.empty') }}</p>
          </div>
        </div>
      </div>

      <div class="file-viewer">
        <div v-if="fileLoading" class="viewer-placeholder">
          <i class="el-icon-loading"></i>
          <span>
            {{ $t('generalAgent.skill.skillWorkBench.editor.loadingFile') }}
          </span>
        </div>
        <template v-else-if="currentFile">
          <div class="viewer-breadcrumb">
            <i
              :class="getFileIcon(currentFile).icon"
              :style="{ color: getFileIcon(currentFile).color }"
              class="file-icon"
            ></i>
            <span class="file-info">{{ currentFile.path }}</span>
          </div>
          <div class="viewer-pane">
            <MonacoEditor
              :language="currentLanguage"
              :readOnly="true"
              :value="fileContent"
              theme="vs"
            />
          </div>
        </template>
        <div v-else class="viewer-placeholder">
          <i class="el-icon-document"></i>
          <p>
            {{
              $t(
                'adminCenter.pageModules.resourcePool.skill.detail.selectFileTip',
              )
            }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import {
  getSkillContentFiles,
  getSkillContentFile,
} from '@/api/skillResource/skillWorkSpace';
import { getFileIcon as getFileIconByName } from '@/utils/fileIcons';
import { getLanguageByPath } from '@/views/generalAgent/components/skills/workspaceConstants';
import MonacoEditor from '@/components/MonacoEditor/index.vue';

export default {
  name: 'SkillContentWorkspace',
  components: { MonacoEditor },
  props: {
    customSkillId: {
      type: String,
      default: '',
    },
  },
  data() {
    return {
      loading: false,
      refreshing: false,
      treeData: [],
      defaultExpandedKeys: [],
      treeProps: { children: 'children', label: 'name' },
      activePath: '',
      currentFile: null,
      fileContent: '',
      fileLoading: false,
    };
  },
  computed: {
    currentLanguage() {
      return this.currentFile
        ? getLanguageByPath(this.currentFile.path)
        : 'plaintext';
    },
  },
  watch: {
    customSkillId: {
      handler(val, oldVal) {
        if (val !== oldVal) {
          this.resetState();
          if (val) this.fetchFiles();
        }
      },
      immediate: true,
    },
  },
  methods: {
    getFileIcon(data) {
      if (!data) return { icon: 'el-icon-document', color: '#6d8086' };
      if (data.isDir) return { icon: 'el-icon-folder', color: '#dcb67a' };
      return getFileIconByName(data.name);
    },
    resetState() {
      this.treeData = [];
      this.defaultExpandedKeys = [];
      this.activePath = '';
      this.currentFile = null;
      this.fileContent = '';
    },
    async fetchFiles() {
      if (!this.customSkillId) return;
      this.loading = true;
      try {
        const res = await getSkillContentFiles(this.customSkillId);
        if (res.code === 0 && res.data) {
          this.treeData = res.data.files || [];
          this.defaultExpandedKeys = this.collectDirPaths(this.treeData);
        }
      } finally {
        this.loading = false;
      }
    },
    collectDirPaths(nodes) {
      const paths = [];
      const walk = list => {
        if (!Array.isArray(list)) return;
        list.forEach(n => {
          if (n.isDir) {
            paths.push(n.path);
            if (n.children) walk(n.children);
          }
        });
      };
      walk(nodes);
      return paths;
    },
    refresh() {
      this.refreshing = true;
      this.fetchFiles().finally(() => {
        this.refreshing = false;
      });
    },
    handleNodeClick(data) {
      if (data.isDir) return;
      this.activePath = data.path;
      this.loadFile(data);
    },
    async loadFile(file) {
      if (!this.customSkillId || !file || !file.path) return;
      this.currentFile = file;
      this.fileLoading = true;
      this.fileContent = '';
      try {
        const res = await getSkillContentFile(this.customSkillId, file.path);
        if (res.code === 0 && res.data) {
          this.fileContent = res.data.content || '';
        }
      } finally {
        this.fileLoading = false;
      }
    },
  },
};
</script>

<style lang="scss" scoped>
.skill-content-workspace {
  height: 480px;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  overflow: hidden;
  background: #fff;

  .workspace-body {
    display: flex;
    height: 100%;
  }

  .file-tree {
    width: 240px;
    border-right: 1px solid #e4e7ed;
    display: flex;
    flex-direction: column;
    background: #fafafa;
    flex-shrink: 0;
  }

  .file-tree-header {
    display: flex;
    align-items: center;
    padding: 8px 12px;
    border-bottom: 1px solid #e4e7ed;
    background: #f5f7fa;
    font-size: 13px;
    color: #606266;
    gap: 6px;

    > i.el-icon-folder {
      color: #dcb67a;
    }

    span {
      flex: 1;
    }

    .header-refresh {
      cursor: pointer;
      color: #909399;
      padding: 2px;
      border-radius: 3px;
      &:hover {
        color: #409eff;
        background: #ecf5ff;
      }
      &.spinning {
        animation: spin 0.6s linear;
      }
    }
  }

  .file-tree-content {
    flex: 1;
    overflow-y: auto;
    padding: 4px 0;

    ::v-deep .el-tree {
      background: transparent;
      color: #303133;

      .el-tree-node__content {
        height: 26px;
        &:hover {
          background-color: #ecf5ff;
        }
      }

      .el-tree-node.is-leaf .el-tree-node__expand-icon {
        display: none;
      }

      .custom-tree-node {
        display: flex;
        align-items: center;
        width: 100%;
        min-width: 0;
        font-size: 13px;
        padding-right: 8px;

        .file-icon {
          margin-right: 5px;
          font-size: 14px;
          flex-shrink: 0;
        }

        .file-name {
          flex: 1;
          min-width: 0;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }

        &.is-active {
          color: #409eff;
          font-weight: 600;
        }
      }
    }

    .empty-state {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      height: 200px;
      color: #909399;
      i {
        font-size: 40px;
        margin-bottom: 8px;
      }
      p {
        margin: 0;
        font-size: 13px;
      }
    }
  }

  .file-viewer {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .viewer-breadcrumb {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 12px;
    border-bottom: 1px solid #e4e7ed;
    background: #fafafa;
    font-size: 12px;
    color: #606266;
    flex-shrink: 0;

    .file-icon {
      font-size: 14px;
    }

    .file-info {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  .viewer-pane {
    flex: 1;
    min-height: 0;
    overflow: hidden;
  }

  .viewer-placeholder {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 10px;
    color: #909399;
    font-size: 13px;

    i {
      font-size: 40px;
      color: #c0c4cc;
    }
  }
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
