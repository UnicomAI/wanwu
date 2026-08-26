<template>
  <div
    class="file-tree-wrapper"
    @mousedown.capture="handleExternalPointer"
    @contextmenu.capture="handleExternalPointer"
  >
    <div class="file-tree-header">
      <div class="header-icons">
        <svg-icon
          :class="[
            'header-icon svg-icon-btn',
            { active: activeView === 'files' },
          ]"
          icon-class="skillWorkspaceFolder"
          :title="$t('generalAgent.skill.skillWorkBench.common.files')"
          @click.native="$emit('switch-view', 'files')"
        />
        <svg-icon
          :class="[
            'header-icon svg-icon-btn',
            { active: activeView === 'search' },
          ]"
          icon-class="skillWorkspaceSearch"
          :title="$t('generalAgent.skill.skillWorkBench.common.search')"
          @click.native="$emit('switch-view', 'search')"
        />
        <span class="git-icon-wrap">
          <svg-icon
            :class="[
              'header-icon svg-icon-btn',
              { active: activeView === 'git' },
            ]"
            icon-class="gitBranch"
            :title="$t('generalAgent.skill.skillWorkBench.common.git')"
            @click.native="$emit('switch-view', 'git')"
          />
          <span v-if="effectiveGitChangeCount > 0" class="git-change-count">
            {{ effectiveGitChangeCount > 99 ? '99+' : effectiveGitChangeCount }}
          </span>
        </span>
      </div>
      <div class="header-actions">
        <i
          class="el-icon-refresh header-icon"
          :title="$t('generalAgent.skill.skillWorkBench.common.refresh')"
          :class="{ spinning: manualLoading }"
          @click="refreshFiles"
        ></i>
      </div>
    </div>
    <input
      ref="uploadInput"
      class="upload-input"
      type="file"
      multiple
      accept="*/*"
      @change="handleUploadChange"
    />
    <div class="file-tree-content" @scroll="hideContextMenu">
      <el-tree
        v-if="treeData.length > 0"
        :data="treeData"
        :props="treeProps"
        node-key="path"
        :expand-on-click-node="false"
        :default-expanded-keys="defaultExpandedKeys"
        @node-click="handleNodeClick"
        @node-contextmenu="handleNodeContextMenu"
        ref="fileTree"
      >
        <span
          class="custom-tree-node"
          slot-scope="{ node, data }"
          :class="[
            { 'is-placeholder': data.isEmptyPlaceholder },
            statusClass(data),
          ]"
        >
          <template v-if="data.isEditing">
            <el-input
              :ref="'edit-' + data.path"
              v-model="data.editName"
              class="tree-edit-input"
              size="mini"
              :maxlength="maxEntryNameLength"
              :placeholder="
                $t('generalAgent.skill.skillWorkBench.fileTree.name')
              "
              @keydown.enter.native.prevent="commitEditing(data)"
              @keydown.esc.native.prevent="cancelEditing(data)"
              @blur="handleEditBlur(data)"
            />
          </template>
          <template v-else>
            <i
              :class="getFileIcon(data).icon"
              class="file-icon"
              :style="{
                color: data.isEmptyPlaceholder
                  ? '#999'
                  : getFileIcon(data).color,
              }"
            ></i>
            <el-tooltip :content="node.label" placement="top" :open-delay="300">
              <span class="file-name">{{ node.label }}</span>
            </el-tooltip>
            <span
              v-if="statusMarker(data)"
              :class="['status-marker', statusMarker(data).className]"
            >
              {{ statusMarker(data).label }}
            </span>
          </template>
        </span>
      </el-tree>
      <div v-else class="empty-state">
        <i class="el-icon-folder-opened"></i>
        <p>
          {{
            manualLoading
              ? $t('generalAgent.skill.skillWorkBench.fileTree.loading')
              : $t('generalAgent.skill.skillWorkBench.fileTree.empty')
          }}
        </p>
        <el-button
          size="mini"
          type="text"
          @click="handleRootCommand('new-file')"
        >
          {{ $t('generalAgent.skill.skillWorkBench.fileTree.newFile') }}
        </el-button>
      </div>
    </div>

    <!-- 右键菜单-->
    <el-dropdown
      ref="contextDropdown"
      trigger="click"
      size="mini"
      class="context-menu-dropdown"
      placement="bottom-start"
      @command="handleContextMenuCommand"
      :style="{ left: contextMenuX + 'px', top: contextMenuY + 'px' }"
    >
      <span ref="contextTrigger" class="dropdown-trigger-node"></span>
      <el-dropdown-menu slot="dropdown" class="file-tree-dropdown-menu">
        <template v-if="contextMenuTarget && contextMenuTarget.isDir">
          <el-dropdown-item command="new-file">
            {{ $t('generalAgent.skill.skillWorkBench.fileTree.newFile') }}
          </el-dropdown-item>
          <el-dropdown-item command="new-folder">
            {{ $t('generalAgent.skill.skillWorkBench.fileTree.newFolder') }}
          </el-dropdown-item>
          <el-dropdown-item command="upload">
            {{ $t('generalAgent.skill.skillWorkBench.fileTree.upload') }}
          </el-dropdown-item>
        </template>
        <el-dropdown-item command="rename">
          {{ $t('generalAgent.skill.skillWorkBench.fileTree.rename') }}
        </el-dropdown-item>
        <el-dropdown-item command="copy-path">
          {{ $t('generalAgent.skill.skillWorkBench.fileTree.copyPath') }}
        </el-dropdown-item>
        <el-dropdown-item command="download">
          {{ $t('common.button.download') }}
        </el-dropdown-item>
        <el-dropdown-item command="delete" class="text-danger">
          {{ $t('common.button.delete') }}
        </el-dropdown-item>
      </el-dropdown-menu>
    </el-dropdown>
  </div>
</template>

<script>
import { getSkillWorkspaceFiles } from '@/api/skillResource/skillWorkSpace';
import { getFileIcon } from '@/utils/fileIcons';

export default {
  name: 'FileTree',
  props: {
    customSkillId: {
      type: String,
      required: true,
    },
    activeView: {
      type: String,
      default: 'files',
    },
    polling: {
      type: Boolean,
      default: false,
    },
    refreshInterval: {
      type: Number,
      default: 2000,
    },
    gitStatusFiles: {
      type: Array,
      default: () => [],
    },
    gitChangeCount: {
      type: Number,
      default: 0,
    },
  },
  data() {
    return {
      treeData: [],
      manualLoading: false,
      pollingTimer: null,
      contextMenuX: 0,
      contextMenuY: 0,
      contextMenuTarget: null,
      pendingUploadDirectory: '',
      editSequence: 0,
      maxEntryNameLength: 255,
      defaultExpandedKeys: [], // 保存展开的节点路径
      treeProps: {
        children: 'children',
        label: 'name',
      },
    };
  },
  computed: {
    effectiveGitChangeCount() {
      if (this.gitChangeCount > 0) return this.gitChangeCount;
      return new Set(
        (this.gitStatusFiles || [])
          .map(file => this.normalizePath(file && file.path))
          .filter(Boolean),
      ).size;
    },
  },
  mounted() {
    this.fetchFiles(true);
    window.addEventListener('resize', this.hideContextMenu);
  },
  beforeDestroy() {
    this.stopPolling();
    window.removeEventListener('resize', this.hideContextMenu);
  },
  methods: {
    async fetchFiles(showLoading = false) {
      if (!this.customSkillId) return;
      if (showLoading) this.manualLoading = true;
      try {
        const res = await getSkillWorkspaceFiles(this.customSkillId);
        if (res.code === 0 && res.data) {
          // 预处理：为空目录添加占位子节点，以便显示展开箭头
          const newTreeData = this.processTreeData(res.data.files || []);

          // 保存当前展开状态
          this.saveExpandedState();

          // 更新数据
          this.treeData = newTreeData;

          // 如果是首次加载且没有保存的展开状态，默认展开所有目录
          if (showLoading && this.defaultExpandedKeys.length === 0) {
            this.$nextTick(() => {
              this.expandAllDirectories();
            });
          }
        }
      } catch (error) {
        console.error('Failed to fetch files:', error);
      } finally {
        if (showLoading) this.manualLoading = false;
      }
    },
    // 保存当前展开状态
    saveExpandedState() {
      if (this.$refs.fileTree) {
        const nodesMap = this.$refs.fileTree.store.nodesMap;
        this.defaultExpandedKeys = [];
        for (const path in nodesMap) {
          if (nodesMap[path].expanded && !nodesMap[path].isLeaf) {
            this.defaultExpandedKeys.push(path);
          }
        }
      }
    },
    // 展开所有目录
    expandAllDirectories() {
      const expandedPaths = [];
      const collectPaths = nodes => {
        if (!Array.isArray(nodes)) return;
        nodes.forEach(node => {
          if (node.isDir) {
            expandedPaths.push(node.path);
            if (node.children) {
              collectPaths(node.children);
            }
          }
        });
      };
      collectPaths(this.treeData);
      this.defaultExpandedKeys = expandedPaths;
    },
    // 递归处理树数据，为空目录添加占位节点
    processTreeData(nodes) {
      if (!Array.isArray(nodes)) return nodes;

      return nodes.map(node => {
        if (node.isDir) {
          const children = node.children || [];
          if (children.length === 0) {
            // 空目录：添加一个隐藏的占位节点
            return {
              ...node,
              children: [
                {
                  path: `${node.path}/.empty`,
                  name: this.$t(
                    'generalAgent.skill.skillWorkBench.fileTree.emptyDir',
                  ),
                  isDir: false,
                  isEmptyPlaceholder: true,
                },
              ],
            };
          } else {
            return {
              ...node,
              children: this.processTreeData(children),
            };
          }
        }
        return node;
      });
    },
    refreshFiles() {
      this.fetchFiles(true);
    },
    startPolling() {
      if (!this.customSkillId) {
        return;
      }
      this.stopPolling();
      this.pollingTimer = setInterval(() => {
        this.fetchFiles(false); // 轮询不显示 loading
      }, this.refreshInterval);
    },
    stopPolling() {
      if (this.pollingTimer) {
        clearInterval(this.pollingTimer);
        this.pollingTimer = null;
      }
    },
    handleNodeContextMenu(event, data) {
      event.preventDefault();
      event.stopPropagation();
      if (!data || data.isEmptyPlaceholder) {
        this.hideContextMenu();
        return;
      }

      this.contextMenuTarget = data;
      // 稍微偏移，确保菜单在鼠标右下方，且不被光标完全遮挡
      this.contextMenuX = event.clientX + 2;
      this.contextMenuY = event.clientY + 2;

      this.$nextTick(() => {
        if (this.$refs.contextTrigger) {
          this.$refs.contextTrigger.click();
        }
      });
    },
    handleContextMenuCommand(command) {
      const target = this.contextMenuTarget;
      if (!target) return;
      if (command === 'download') this.$emit('download-file', target);
      else if (command === 'delete') this.$emit('delete-file', target);
      else if (command === 'rename') this.startRename(target);
      else if (command === 'copy-path') this.copyRelativePath(target);
      else if (command === 'new-file') this.startCreate(target.path, false);
      else if (command === 'new-folder') this.startCreate(target.path, true);
      else if (command === 'upload') this.openUpload(target.path);
      this.hideContextMenu();
    },
    hideContextMenu() {
      const dropdown = this.$refs.contextDropdown;
      if (dropdown && dropdown.visible) {
        dropdown.hide();
      }
      this.contextMenuTarget = null;
    },
    handleNodeClick(data, node) {
      this.hideContextMenu();
      // 忽略空目录占位节点点击
      if (data.isEmptyPlaceholder) {
        return;
      }
      if (data.isDir) {
        // 点击目录时切换展开/折叠状态
        if (node.expanded) {
          node.collapse();
        } else {
          node.expand();
        }
      } else {
        this.$emit('file-click', data);
      }
    },
    getFileIcon(data) {
      if (data.isEmptyPlaceholder) {
        return { icon: 'el-icon-document', color: '#999' };
      }
      if (data.isDir) return { icon: 'el-icon-folder', color: '#dcb67a' };
      return getFileIcon(data.name);
    },
    normalizePath(path) {
      return String(path || '')
        .replace(/\\/g, '/')
        .replace(/^\/+|\/+$/g, '');
    },
    statusForPath(path) {
      const normalizedPath = this.normalizePath(path);
      if (!normalizedPath) return null;
      return (this.gitStatusFiles || []).find(
        file => this.normalizePath(file && file.path) === normalizedPath,
      );
    },
    hasDirtyDescendant(path) {
      const prefix = `${this.normalizePath(path)}/`;
      return (this.gitStatusFiles || []).some(file => {
        const candidate = this.normalizePath(file && file.path);
        return candidate && candidate.indexOf(prefix) === 0;
      });
    },
    statusMarker(data) {
      if (!data || data.isEmptyPlaceholder || data.isEditing) return null;
      const status = data.isDir ? null : this.statusForPath(data.path);
      const changeType = String(
        (status && status.changeType) || '',
      ).toLowerCase();
      if (status && (changeType === 'untracked' || changeType === 'added')) {
        return { label: 'U', className: 'status-untracked' };
      }
      if (status && ['modified', 'renamed', 'deleted'].includes(changeType)) {
        return { label: 'M', className: 'status-modified' };
      }
      return null;
    },
    statusClass(data) {
      if (!data || data.isEmptyPlaceholder || data.isEditing) return '';
      if (data.isDir) {
        if (!this.hasDirtyDescendant(data.path)) return '';
        return [
          'is-dirty-directory',
          this.hasUntrackedDescendant(data.path)
            ? 'status-untracked-directory'
            : 'status-modified-directory',
        ];
      }
      const marker = this.statusMarker(data);
      return marker ? marker.className : '';
    },
    hasUntrackedDescendant(path) {
      const normalizedPath = this.normalizePath(path);
      const prefix = normalizedPath ? `${normalizedPath}/` : '';
      return (this.gitStatusFiles || []).some(file => {
        const candidate = this.normalizePath(file && file.path);
        const changeType = String(
          (file && file.changeType) || '',
        ).toLowerCase();
        const isUnderDirectory =
          candidate &&
          (candidate.indexOf(prefix) === 0 || candidate === normalizedPath);
        return (
          isUnderDirectory &&
          (changeType === 'untracked' || changeType === 'added')
        );
      });
    },
    expandDirectory(path) {
      const normalizedPath = this.normalizePath(path);
      if (!normalizedPath || !this.$refs.fileTree) return;
      const node = this.$refs.fileTree.getNode(normalizedPath);
      if (node && !node.expanded) node.expand();
    },
    findChildren(directoryPath) {
      const targetPath = this.normalizePath(directoryPath);
      if (!targetPath) return this.treeData;
      const walk = nodes => {
        if (!Array.isArray(nodes)) return null;
        for (let i = 0; i < nodes.length; i += 1) {
          const item = nodes[i];
          if (this.normalizePath(item.path) === targetPath) {
            if (!Array.isArray(item.children)) this.$set(item, 'children', []);
            return item.children;
          }
          const result = walk(item.children);
          if (result) return result;
        }
        return null;
      };
      return walk(this.treeData);
    },
    findNode(path) {
      const targetPath = this.normalizePath(path);
      const walk = nodes => {
        if (!Array.isArray(nodes)) return null;
        for (let i = 0; i < nodes.length; i += 1) {
          if (this.normalizePath(nodes[i].path) === targetPath) return nodes[i];
          const result = walk(nodes[i].children);
          if (result) return result;
        }
        return null;
      };
      return walk(this.treeData);
    },
    validateEntryName(name) {
      const value = String(name || '').trim();
      if (!value || value === '.' || value === '..') return false;
      // 名称只能是单个文件名。发送到接口前拒绝分隔符、控制字符及保留的
      // .git 条目。
      return (
        value.length <= this.maxEntryNameLength &&
        !/[\\/\u0000-\u001f\u007f]/.test(value) &&
        value.toLowerCase() !== '.git'
      );
    },
    startCreate(directoryPath, isDir) {
      const parentPath = this.normalizePath(directoryPath);
      const children = this.findChildren(parentPath);
      if (!children) {
        this.$message.warning(
          this.$t(
            'generalAgent.skill.skillWorkBench.fileTree.directoryUnavailable',
          ),
        );
        return;
      }
      if (children.some(child => child.isEditing)) return;
      const id = `.__new__${++this.editSequence}`;
      const temporary = {
        path: parentPath ? `${parentPath}/${id}` : id,
        name: '',
        isDir: Boolean(isDir),
        isEditing: true,
        isNew: true,
        editName: '',
        parentPath,
      };
      const placeholderIndex = children.findIndex(
        child => child.isEmptyPlaceholder,
      );
      if (placeholderIndex >= 0) this.$delete(children, placeholderIndex);
      children.push(temporary);
      this.$nextTick(() => {
        this.expandDirectory(parentPath);
        // 展开折叠的 Element-UI 树节点会触发下一次渲染；等待渲染完成后再获取
        // 输入框引用，确保在折叠文件夹中创建条目后可以立即编辑。
        this.$nextTick(() => {
          const editor = this.$refs[`edit-${temporary.path}`];
          if (editor && editor.focus) editor.focus();
        });
      });
    },
    startRename(data) {
      if (!data || data.isEmptyPlaceholder || data.isEditing) return;
      this.$set(data, 'isEditing', true);
      this.$set(data, 'editName', data.name);
      this.$nextTick(() => {
        const editor = this.$refs[`edit-${data.path}`];
        if (editor && editor.focus) {
          editor.focus();
          if (editor.select) editor.select();
        }
      });
    },
    findEditingEntry(nodes = this.treeData) {
      if (!Array.isArray(nodes)) return null;
      for (let i = 0; i < nodes.length; i += 1) {
        const node = nodes[i];
        if (node && node.isEditing) return node;
        const editingChild = this.findEditingEntry(node && node.children);
        if (editingChild) return editingChild;
      }
      return null;
    },
    handleExternalPointer(event) {
      const editing = this.findEditingEntry();
      if (!editing || !editing.isNew || String(editing.editName || '').trim()) {
        return;
      }
      const target = event && event.target;
      if (target && target.closest && target.closest('.tree-edit-input')) {
        return;
      }
      this.cancelEditing(editing);
    },
    handleEditBlur(data) {
      if (!data || !data.isEditing) return;
      if (!String(data.editName || '').trim()) {
        this.cancelEditing(data);
        return;
      }
      this.commitEditing(data);
    },
    commitEditing(data) {
      if (!data || !data.isEditing) return;
      const newName = String(data.editName || '').trim();
      if (!this.validateEntryName(newName)) {
        this.$message.warning(
          this.$t('generalAgent.skill.skillWorkBench.fileTree.invalidName'),
        );
        return;
      }
      if (data.isNew) {
        const parentPath = this.normalizePath(data.parentPath);
        const path = parentPath ? `${parentPath}/${newName}` : newName;
        this.expandDirectory(parentPath);
        this.$set(data, 'name', newName);
        this.$set(data, 'path', path);
        this.$set(data, 'isEditing', false);
        this.$emit(data.isDir ? 'create-folder' : 'create-file', {
          path,
          name: newName,
          parentPath,
        });
      } else {
        if (newName === data.name) {
          this.$set(data, 'isEditing', false);
          return;
        }
        this.$set(data, 'isEditing', false);
        this.$emit('rename-entry', { entry: data, newName });
      }
    },
    cancelEditing(data) {
      if (!data || !data.isEditing) return;
      if (data.isNew) {
        const parentPath = this.normalizePath(data.parentPath);
        const remove = nodes => {
          if (!Array.isArray(nodes)) return false;
          const index = nodes.indexOf(data);
          if (index >= 0) {
            this.$delete(nodes, index);
            return true;
          }
          return nodes.some(child => remove(child.children));
        };
        remove(this.treeData);
        this.restoreEmptyDirectoryPlaceholder(parentPath);
      } else {
        this.$set(data, 'isEditing', false);
      }
    },
    restoreEmptyDirectoryPlaceholder(directoryPath) {
      const normalizedPath = this.normalizePath(directoryPath);
      if (!normalizedPath) return;
      const directory = this.findNode(normalizedPath);
      if (
        !directory ||
        !directory.isDir ||
        !Array.isArray(directory.children) ||
        directory.children.length > 0
      ) {
        return;
      }
      this.$set(directory, 'children', [
        {
          path: `${normalizedPath}/.empty`,
          name: this.$t('generalAgent.skill.skillWorkBench.fileTree.emptyDir'),
          isDir: false,
          isEmptyPlaceholder: true,
        },
      ]);
    },
    handleRootCommand(command) {
      if (command === 'new-file') this.startCreate('', false);
      else if (command === 'new-folder') this.startCreate('', true);
      else if (command === 'upload') this.openUpload('');
    },
    openUpload(directoryPath) {
      this.pendingUploadDirectory = this.normalizePath(directoryPath);
      const input = this.$refs.uploadInput;
      if (!input) return;
      input.value = '';
      input.click();
    },
    handleUploadChange(event) {
      const files = Array.from(
        (event && event.target && event.target.files) || [],
      );
      const path = this.pendingUploadDirectory;
      this.pendingUploadDirectory = '';
      if (event && event.target) event.target.value = '';
      if (files.length > 0) this.$emit('upload-files', { path, files });
    },
    async copyRelativePath(data) {
      const path = this.normalizePath(data && data.path);
      if (!path) return;
      try {
        if (navigator.clipboard && navigator.clipboard.writeText) {
          await navigator.clipboard.writeText(path);
        } else {
          const textarea = document.createElement('textarea');
          textarea.value = path;
          textarea.setAttribute('readonly', '');
          textarea.style.position = 'fixed';
          textarea.style.opacity = '0';
          document.body.appendChild(textarea);
          textarea.select();
          document.execCommand('copy');
          document.body.removeChild(textarea);
        }
        this.$message.success(
          this.$t('generalAgent.skill.skillWorkBench.fileTree.copySuccess'),
        );
      } catch (error) {
        this.$message.error(
          this.$t('generalAgent.skill.skillWorkBench.fileTree.copyFailed'),
        );
      }
    },
  },
  watch: {
    customSkillId(newVal, oldVal) {
      if (newVal !== oldVal) {
        this.treeData = [];
        this.defaultExpandedKeys = []; // 重置展开状态
        if (newVal) {
          this.fetchFiles(true);
          // 如果 polling 已经是 true，启动轮询
          if (this.polling) {
            this.startPolling();
          }
        }
      }
    },
    polling: {
      handler(newVal) {
        if (newVal) {
          this.startPolling();
        } else {
          this.stopPolling();
        }
      },
      immediate: true,
    },
  },
};
</script>

<style lang="scss" scoped>
.file-tree-wrapper {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #f3f3f3;
  color: #333;

  .file-tree-header {
    padding: 6px 8px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid #e0e0e0;
    background: #f8f8f8;

    .header-icons {
      display: flex;
      gap: 4px;
    }

    .header-actions {
      display: flex;
      align-items: center;
      gap: 2px;
    }

    .git-icon-wrap {
      position: relative;
      display: inline-flex;
    }

    .git-change-count {
      position: absolute;
      // gitBranch.svg 的 viewBox 有透明留白，按实际图形边界收近徽标。
      right: 1px;
      bottom: -1px;
      min-width: 11px;
      height: 11px;
      padding: 0 2px;
      border-radius: 8px;
      background: #f56c6c;
      color: #fff;
      font-size: 8px;
      line-height: 11px;
      text-align: center;
      pointer-events: none;
      z-index: 2;
    }

    .header-icon {
      width: 24px;
      height: 24px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 15px;
      color: #666;
      cursor: pointer;
      border-radius: 4px;

      &:hover {
        color: #444;
        background: rgba(0, 0, 0, 0.05);
      }
      &.active {
        color: #5983ff;
        background: rgba(89, 131, 255, 0.1);
      }
      &.spinning {
        animation: spin 0.6s linear;
      }

      &.svg-icon-btn {
        font-size: 15px;
        ::v-deep svg {
          width: 15px;
          height: 15px;
        }
      }
    }
  }

  .file-tree-content {
    flex: 1;
    overflow-y: auto;
    padding: 4px 0;

    ::v-deep .el-tree {
      background: transparent;
      color: #333;

      .el-tree-node__content {
        height: 22px;
        line-height: 22px;

        &:hover {
          background-color: #e8e8e8;
        }
      }

      .el-tree-node.is-current > .el-tree-node__content {
        background-color: rgba(89, 131, 255, 0.12);
      }

      // 文件节点（非目录）隐藏展开箭头
      .el-tree-node.is-leaf .el-tree-node__expand-icon {
        display: none;
      }

      .el-tree-node__expand-icon {
        color: #666;
        font-size: 12px;
        &.is-leaf {
          display: none;
        }
      }

      .custom-tree-node {
        display: flex;
        align-items: center;
        font-size: 13px;
        min-width: 0;
        width: 100%;
        box-sizing: border-box;
        padding-right: 14px;

        .file-icon {
          margin-right: 4px;
          font-size: 14px;
          flex-shrink: 0;
        }

        .file-name {
          display: block;
          min-width: 0;
          flex: 1;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }

        .tree-edit-input {
          flex: 1;
          min-width: 0;

          ::v-deep .el-input__inner {
            height: 20px;
            line-height: 20px;
            padding: 0 4px;
            font-size: 12px;
          }
        }

        .status-marker {
          flex-shrink: 0;
          min-width: 10px;
          margin-left: 6px;
          font-size: 11px;
          font-weight: 700;
          text-align: center;
        }

        &.status-modified {
          .file-name {
            color: #b08800;
          }
          .status-marker {
            color: #b08800;
          }
        }

        &.status-untracked {
          .file-name {
            color: #22863a;
          }
          .status-marker {
            color: #22863a;
          }
        }

        &.is-dirty-directory {
          .file-name {
            color: #b08800;
          }
          &::after {
            content: '';
            width: 10px;
            height: 10px;
            margin-left: 6px;
            background: radial-gradient(
              circle,
              #b08800 0 3px,
              transparent 3.5px
            );
            flex-shrink: 0;
          }
        }

        &.status-untracked-directory {
          .file-name {
            color: #22863a;
          }
          &::after {
            background: radial-gradient(
              circle,
              #22863a 0 3px,
              transparent 3.5px
            );
          }
        }

        // 空目录占位节点样式
        &.is-placeholder {
          opacity: 0.6;
          font-style: italic;
          color: #999;
          cursor: default;

          .file-icon {
            display: none;
          }
        }
      }
    }

    .empty-state {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      height: 200px;
      color: #999;

      i {
        font-size: 48px;
        margin-bottom: 12px;
      }
      p {
        font-size: 13px;
        margin: 0;
      }
    }
  }

  .context-menu-dropdown {
    position: fixed;
    visibility: hidden;
    pointer-events: none;

    .dropdown-trigger-node {
      display: block;
      width: 1px;
      height: 1px;
    }
  }

  .upload-input {
    display: none;
  }
}

::v-deep .file-tree-dropdown-menu.el-dropdown-menu {
  padding: 4px 0;
  border-radius: 4px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);

  .el-dropdown-menu__item {
    font-size: 13px;
    line-height: 32px;
    padding: 0 16px;
    display: flex;
    align-items: center;
    gap: 8px;

    i {
      margin: 0;
      font-size: 14px;
    }
    &:hover {
      background-color: #f5f7fa;
      color: #5983ff;
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

.text-danger {
  color: #f56c6c !important;
}
</style>
