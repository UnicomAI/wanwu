<template>
  <div class="skill-workspace-explorer">
    <FileTree
      ref="fileTree"
      v-show="sidebarView === 'files'"
      :customSkillId="customSkillId"
      :activeView="sidebarView"
      :gitStatusFiles="gitStatusFiles"
      :gitChangeCount="gitChangeCount"
      @file-click="$emit('open-file', $event)"
      @download-file="downloadFile"
      @delete-file="handleDeleteFile"
      @create-file="handleCreateFile"
      @create-folder="handleCreateFolder"
      @rename-entry="handleRenameEntry"
      @upload-files="handleUploadFiles"
      @switch-view="sidebarView = $event"
    />

    <div v-show="sidebarView === 'search'" class="search-panel">
      <div class="panel-header">
        <div class="header-icons">
          <svg-icon
            :class="[
              'header-icon svg-icon-btn',
              { active: sidebarView === 'files' },
            ]"
            icon-class="skillWorkspaceFolder"
            :title="$t('generalAgent.skill.skillWorkBench.common.files')"
            @click.native="sidebarView = 'files'"
          />
          <svg-icon
            :class="[
              'header-icon svg-icon-btn',
              { active: sidebarView === 'search' },
            ]"
            icon-class="skillWorkspaceSearch"
            :title="$t('generalAgent.skill.skillWorkBench.common.search')"
            @click.native="sidebarView = 'search'"
          />
          <span class="git-icon-wrap">
            <svg-icon
              :class="[
                'header-icon svg-icon-btn',
                { active: sidebarView === 'git' },
              ]"
              icon-class="gitBranch"
              :title="$t('generalAgent.skill.skillWorkBench.common.git')"
              @click.native="sidebarView = 'git'"
            />
            <span v-if="gitChangeCount > 0" class="git-change-count">
              {{ gitChangeCount > 99 ? '99+' : gitChangeCount }}
            </span>
          </span>
        </div>
      </div>
      <div class="search-input-wrap">
        <el-input
          v-model="searchKeyword"
          :placeholder="
            $t('generalAgent.skill.skillWorkBench.search.placeholder')
          "
          size="mini"
          prefix-icon="el-icon-search"
          clearable
          @clear="clearSearch"
        />
        <div class="search-options">
          <el-tooltip
            :content="
              $t('generalAgent.skill.skillWorkBench.search.caseSensitive')
            "
            placement="bottom"
          >
            <span
              :class="['option-btn', { active: caseSensitive }]"
              @click="caseSensitive = !caseSensitive"
            >
              Aa
            </span>
          </el-tooltip>
          <el-tooltip
            :content="$t('generalAgent.skill.skillWorkBench.search.wholeWord')"
            placement="bottom"
          >
            <span
              :class="['option-btn', { active: wholeWord }]"
              @click="wholeWord = !wholeWord"
            >
              |ab|
            </span>
          </el-tooltip>
          <el-tooltip
            :content="$t('generalAgent.skill.skillWorkBench.search.regex')"
            placement="bottom"
          >
            <span
              :class="['option-btn', { active: useRegex }]"
              @click="useRegex = !useRegex"
            >
              .*
            </span>
          </el-tooltip>
          <el-tooltip
            :content="
              $t('generalAgent.skill.skillWorkBench.search.includeExclude')
            "
            placement="bottom"
          >
            <i
              :class="[
                'option-icon el-icon-setting',
                { active: showAdvancedSearch },
              ]"
              @click="showAdvancedSearch = !showAdvancedSearch"
            ></i>
          </el-tooltip>
        </div>
        <div v-if="showAdvancedSearch" class="advanced-options">
          <el-input
            v-model="includePattern"
            :placeholder="
              $t('generalAgent.skill.skillWorkBench.search.includePlaceholder')
            "
            size="mini"
            clearable
          />
          <el-input
            v-model="excludePattern"
            :placeholder="
              $t('generalAgent.skill.skillWorkBench.search.excludePlaceholder')
            "
            size="mini"
            clearable
          />
        </div>
        <div v-if="searching" class="search-status">
          <i class="el-icon-loading"></i>
          {{ $t('generalAgent.skill.skillWorkBench.search.searching') }}
        </div>
      </div>

      <div class="search-results" v-if="searchResults.length > 0">
        <div class="result-toolbar">
          <span class="result-count">
            {{
              $t('generalAgent.skill.skillWorkBench.search.resultCount', {
                count: searchResults.length,
              })
            }}
          </span>
          <div class="toolbar-actions">
            <el-tooltip
              :content="
                $t('generalAgent.skill.skillWorkBench.search.clearResults')
              "
              placement="bottom"
            >
              <i class="el-icon-delete" @click="clearSearch"></i>
            </el-tooltip>
            <el-tooltip
              :content="
                viewMode === 'list'
                  ? $t('generalAgent.skill.skillWorkBench.search.treeView')
                  : $t('generalAgent.skill.skillWorkBench.search.listView')
              "
              placement="bottom"
            >
              <i
                :class="
                  viewMode === 'list' ? 'el-icon-s-grid' : 'el-icon-s-unfold'
                "
                @click="toggleViewMode"
              ></i>
            </el-tooltip>
          </div>
        </div>

        <div v-if="viewMode === 'list'" class="result-list">
          <div
            v-for="(result, index) in searchResults"
            :key="index"
            class="result-item"
            @click="openSearchResult(result)"
          >
            <div class="result-file">
              <i
                :class="getFileIcon(result.path.split('/').pop()).icon"
                :style="{
                  color: getFileIcon(result.path.split('/').pop()).color,
                }"
              ></i>
              <span class="file-path">{{ result.path }}</span>
              <span class="result-line">:{{ result.line }}</span>
            </div>
            <div class="result-content">{{ result.content.trim() }}</div>
          </div>
        </div>

        <div v-else class="result-tree">
          <SearchResultTreeNode
            v-for="node in searchResultTree"
            :key="node.path"
            :node="node"
            :level="0"
            @toggle-node="toggleTreeNode"
            @result-click="handleTreeResultClick"
            @match-click="handleTreeMatchClick"
          />
        </div>
      </div>
      <div class="search-empty" v-else-if="searchDone && searchKeyword">
        <p>{{ $t('generalAgent.skill.skillWorkBench.search.noResults') }}</p>
      </div>
    </div>

    <div v-show="sidebarView === 'git'" class="git-panel">
      <div class="panel-header">
        <div class="header-icons">
          <svg-icon
            :class="[
              'header-icon svg-icon-btn',
              { active: sidebarView === 'files' },
            ]"
            icon-class="skillWorkspaceFolder"
            :title="$t('generalAgent.skill.skillWorkBench.common.files')"
            @click.native="sidebarView = 'files'"
          />
          <svg-icon
            :class="[
              'header-icon svg-icon-btn',
              { active: sidebarView === 'search' },
            ]"
            icon-class="skillWorkspaceSearch"
            :title="$t('generalAgent.skill.skillWorkBench.common.search')"
            @click.native="sidebarView = 'search'"
          />
          <svg-icon
            :class="[
              'header-icon svg-icon-btn',
              { active: sidebarView === 'git' },
            ]"
            icon-class="gitBranch"
            :title="$t('generalAgent.skill.skillWorkBench.common.git')"
          />
        </div>
        <i
          class="el-icon-refresh header-icon"
          :title="$t('generalAgent.skill.skillWorkBench.common.refresh')"
          :class="{ spinning: gitStatusLoading || gitLoading }"
          @click="refreshGit"
        ></i>
      </div>

      <div class="git-staging-area">
        <div class="git-section-title">
          <span>{{ $t('generalAgent.skill.skillWorkBench.git.changes') }}</span>
          <div class="section-actions">
            <i
              class="el-icon-check section-action-icon"
              :title="$t('generalAgent.skill.skillWorkBench.git.stageAll')"
              @click="gitStageAll"
            ></i>
          </div>
        </div>

        <div v-if="unstagedFiles.length > 0" class="git-file-group">
          <div class="git-group-header" @click="toggleGroup('unstaged')">
            <i
              :class="
                groupExpanded.unstaged
                  ? 'el-icon-arrow-down'
                  : 'el-icon-arrow-right'
              "
            ></i>
            <span>
              {{
                $t('generalAgent.skill.skillWorkBench.git.unstagedChanges', {
                  count: unstagedFiles.length,
                })
              }}
            </span>
          </div>
          <div v-show="groupExpanded.unstaged" class="git-group-content">
            <div
              v-for="file in unstagedFiles"
              :key="'u-' + file.path"
              class="git-file-item"
              :class="{
                active: activeGitDiffId === workingDiffId(file, false),
              }"
              @click="selectWorkingFile(file, false)"
            >
              <span :class="['change-type-badge', file.changeType]">
                {{ file.changeType[0].toUpperCase() }}
              </span>
              <span class="file-path" :title="file.path">
                {{ file.path }}
              </span>
              <span
                :class="[
                  'file-action svg-file-action',
                  { disabled: discardingPaths[file.path] },
                ]"
                :title="$t('generalAgent.skill.skillWorkBench.git.discard')"
                @click.stop="gitDiscardFile(file)"
              >
                <svg-icon class-name="abandon-icon" icon-class="u-turn-left" />
              </span>
              <i
                class="el-icon-plus file-action"
                :title="$t('generalAgent.skill.skillWorkBench.git.stage')"
                @click.stop="gitStageFile(file.path)"
              ></i>
            </div>
          </div>
        </div>

        <div v-if="stagedFiles.length > 0" class="git-file-group">
          <div class="git-group-header" @click="toggleGroup('staged')">
            <i
              :class="
                groupExpanded.staged
                  ? 'el-icon-arrow-down'
                  : 'el-icon-arrow-right'
              "
            ></i>
            <span>
              {{
                $t('generalAgent.skill.skillWorkBench.git.stagedChanges', {
                  count: stagedFiles.length,
                })
              }}
            </span>
          </div>
          <div v-show="groupExpanded.staged" class="git-group-content">
            <div
              v-for="file in stagedFiles"
              :key="'s-' + file.path"
              class="git-file-item"
              :class="{ active: activeGitDiffId === workingDiffId(file, true) }"
              @click="selectWorkingFile(file, true)"
            >
              <span :class="['change-type-badge', file.changeType]">
                {{ file.changeType[0].toUpperCase() }}
              </span>
              <span class="file-path" :title="file.path">
                {{ file.path }}
              </span>
              <i
                class="el-icon-minus file-action"
                :title="$t('generalAgent.skill.skillWorkBench.git.unstage')"
                @click.stop="gitUnstageFile(file.path)"
              ></i>
            </div>
          </div>
        </div>

        <div
          v-if="gitStatusFiles.length === 0 && !gitStatusLoading"
          class="git-empty"
        >
          {{ $t('generalAgent.skill.skillWorkBench.git.noChanges') }}
        </div>

        <div class="git-commit-input" v-if="stagedFiles.length > 0">
          <el-input
            v-model="gitCommitMessage"
            size="mini"
            :placeholder="
              $t('generalAgent.skill.skillWorkBench.git.commitPlaceholder')
            "
            @keydown.enter.ctrl.native="gitCommit"
          />
          <el-button
            size="mini"
            type="primary"
            :disabled="!gitCommitMessage.trim()"
            @click="gitCommit"
          >
            {{ $t('generalAgent.skill.skillWorkBench.git.commit') }}
          </el-button>
        </div>
      </div>

      <div class="git-commit-list">
        <div class="git-section-title">
          {{ $t('generalAgent.skill.skillWorkBench.git.history') }}
        </div>
        <div
          v-for="commit in gitCommits"
          :key="commit.hash"
          class="git-commit-item"
          :class="{ active: activeGitDiffId === commitDiffId(commit) }"
          @click="selectGitCommit(commit)"
        >
          <el-button
            class="commit-restore-btn"
            type="text"
            :title="$t('generalAgent.skill.skillWorkBench.git.restore')"
            :aria-label="$t('generalAgent.skill.skillWorkBench.git.restore')"
            :loading="restoringCommitHash === commit.hash"
            :disabled="!!restoringCommitHash"
            @click.stop="restoreGitCommit(commit)"
          >
            <svg-icon class-name="restore-icon" icon-class="history" />
          </el-button>
          <div class="commit-message">{{ commit.message }}</div>
          <div class="commit-meta">
            <span class="commit-hash">{{ commit.hash.substring(0, 7) }}</span>
            <span class="commit-tags" v-if="commit.tags && commit.tags.length">
              <span class="commit-tag" :title="commit.tags.join(', ')">
                {{ latestTag(commit.tags) }}
              </span>
              <el-tooltip
                v-if="commit.tags.length > 1"
                :content="commit.tags.join(', ')"
                placement="top"
              >
                <span class="commit-tag commit-tag-more">
                  +{{ commit.tags.length - 1 }}
                </span>
              </el-tooltip>
            </span>
            <span class="commit-time">{{ formatGitTime(commit.time) }}</span>
          </div>
        </div>
        <div v-if="gitCommits.length === 0 && !gitLoading" class="git-empty">
          {{ $t('generalAgent.skill.skillWorkBench.git.noHistory') }}
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import FileTree from './FileTree.vue';
import SearchResultTreeNode from './SearchResultTreeNode.vue';
import {
  searchSkillWorkspace,
  getSkillWorkspaceGitLog,
  getSkillWorkspaceGitDiff,
  getSkillWorkspaceGitDiffWorking,
  getSkillWorkspaceGitDiffStaged,
  getSkillWorkspaceGitStatus,
  createSkillWorkspaceFile,
  createSkillWorkspaceDirectory,
  renameSkillWorkspaceFile,
  uploadSkillWorkspaceFiles,
  postSkillWorkspaceGitAdd,
  postSkillWorkspaceGitReset,
  postSkillWorkspaceGitCommit,
  postSkillWorkspaceGitDiscard,
  downloadSkillWorkspace,
  deleteSkillWorkspaceFile,
} from '@/api/skillResource/skillWorkSpace';
import { postSkillWorkspaceGitRestore } from '@/api/generalAgent';
import { getFileIcon } from '@/utils/fileIcons';
import { resDownloadFile } from '@/utils/util';

const MS_PER_MINUTE = 60000;
const MS_PER_HOUR = 3600000;
const MS_PER_DAY = 86400000;

export default {
  name: 'SkillWorkspaceExplorer',
  components: {
    FileTree,
    SearchResultTreeNode,
  },
  props: {
    customSkillId: {
      type: String,
      required: true,
    },
    activeGitDiffId: {
      type: String,
      default: '',
    },
    checkRenameAllowed: {
      type: Function,
      default: () => true,
    },
  },
  data() {
    return {
      sidebarView: 'files',
      searchKeyword: '',
      searchResults: [],
      searching: false,
      searchDone: false,
      caseSensitive: false,
      wholeWord: false,
      useRegex: false,
      includePattern: '',
      excludePattern: '',
      showAdvancedSearch: false,
      viewMode: 'list',
      searchDebounceTimer: null,
      searchTreeExpandedPaths: {},
      searchTreeRenderKey: 0,
      gitCommits: [],
      gitLoading: false,
      gitStatusFiles: [],
      gitStatusLoading: false,
      gitCommitMessage: '',
      groupExpanded: { unstaged: true, staged: true },
      downloadingPaths: {},
      discardingPaths: {},
      restoringCommitHash: '',
    };
  },
  computed: {
    unstagedFiles() {
      return this.gitStatusFiles.filter(f => !f.staged);
    },
    stagedFiles() {
      return this.gitStatusFiles.filter(f => f.staged);
    },
    gitChangeCount() {
      const paths = new Set();
      (this.gitStatusFiles || []).forEach(file => {
        const path = String((file && file.path) || '')
          .replace(/\\/g, '/')
          .replace(/^\/+|\/+$/g, '');
        if (path) paths.add(path);
      });
      return paths.size;
    },
    searchResultTree() {
      if (this.searchResults.length === 0) return [];

      const expandedPaths = this.searchTreeExpandedPaths;
      const expandedVersion = this.searchTreeRenderKey;
      // Vue 2 无法可靠地响应临时计算节点上新增的属性，因此用此计数器将
      // 展开/折叠变化显式声明为依赖。
      const isExpanded = path =>
        expandedVersion >= 0 && Boolean(expandedPaths[path]);
      const root = { name: '', children: [], isDir: true };
      const nodeMap = new Map();

      this.searchResults.forEach(result => {
        const parts = result.path.split('/');
        let current = root;

        parts.forEach((part, idx) => {
          const path = parts.slice(0, idx + 1).join('/');
          const isFile = idx === parts.length - 1;

          if (!nodeMap.has(path)) {
            const node = {
              name: part,
              path,
              isDir: !isFile,
              children: isFile ? undefined : [],
              matches: isFile ? [] : undefined,
              expanded: !isFile && isExpanded(path),
            };
            nodeMap.set(path, node);
            current.children.push(node);
          }

          if (!isFile) {
            current = nodeMap.get(path);
          } else {
            nodeMap.get(path).matches.push(result);
          }
        });
      });

      const sortChildren = node => {
        if (!node.children) return;
        node.children.sort((a, b) => {
          if (a.isDir !== b.isDir) return a.isDir ? -1 : 1;
          return a.name.localeCompare(b.name);
        });
        node.children.forEach(sortChildren);
      };
      sortChildren(root);

      return root.children;
    },
  },
  watch: {
    searchKeyword(newVal) {
      if (this.searchDebounceTimer) {
        clearTimeout(this.searchDebounceTimer);
      }
      this.searchDebounceTimer = setTimeout(() => {
        if (newVal.trim()) {
          this.doSearch();
        } else {
          this.clearSearch();
        }
      }, 300);
    },
    caseSensitive() {
      if (this.searchKeyword.trim()) this.doSearch();
    },
    wholeWord() {
      if (this.searchKeyword.trim()) this.doSearch();
    },
    useRegex() {
      if (this.searchKeyword.trim()) this.doSearch();
    },
    includePattern() {
      if (this.searchKeyword.trim()) this.doSearch();
    },
    excludePattern() {
      if (this.searchKeyword.trim()) this.doSearch();
    },
    customSkillId: {
      handler(val) {
        this.gitCommits = [];
        this.gitStatusFiles = [];
        this.clearSearch();
        if (val) {
          this.refreshGit();
        }
      },
      immediate: true,
    },
    sidebarView(val) {
      if (val === 'git') {
        this.refreshGit();
      }
    },
  },
  beforeDestroy() {
    if (this.searchDebounceTimer) {
      clearTimeout(this.searchDebounceTimer);
    }
  },
  methods: {
    showGitView() {
      this.sidebarView = 'git';
      return this.refreshGit();
    },
    refreshFiles() {
      if (this.$refs.fileTree) {
        return this.$refs.fileTree.refreshFiles();
      }
      return Promise.resolve();
    },
    refreshGit() {
      return Promise.all([this.fetchGitLog(), this.fetchGitStatus()]);
    },
    async doSearch() {
      if (!this.searchKeyword.trim()) return;
      try {
        this.searching = true;
        this.searchDone = false;
        const res = await searchSkillWorkspace(this.customSkillId, {
          keyword: this.searchKeyword,
          caseSensitive: this.caseSensitive,
          wholeWord: this.wholeWord,
          useRegex: this.useRegex,
          includePattern: this.includePattern,
          excludePattern: this.excludePattern,
        });
        if (res.code === 0 && res.data) {
          this.searchResults = res.data.results || [];
        }
        this.searchDone = true;
      } catch (e) {
        this.$message.error(
          this.$t('generalAgent.skill.skillWorkBench.search.failed'),
        );
        this.searchDone = true;
      } finally {
        this.searching = false;
      }
    },
    clearSearch() {
      this.searchResults = [];
      this.searchDone = false;
      this.searchTreeExpandedPaths = {};
      this.searchTreeRenderKey += 1;
    },
    toggleViewMode() {
      this.viewMode = this.viewMode === 'list' ? 'tree' : 'list';
    },
    handleTreeResultClick(node) {
      if (!node.isDir && node.matches && node.matches.length > 0) {
        this.openSearchResult(node.matches[0]);
      }
    },
    handleTreeMatchClick(match) {
      this.openSearchResult(match);
    },
    toggleTreeNode(node) {
      if (!node || !node.isDir || !node.path) return;
      if (this.searchTreeExpandedPaths[node.path]) {
        this.$delete(this.searchTreeExpandedPaths, node.path);
      } else {
        this.$set(this.searchTreeExpandedPaths, node.path, true);
      }
      this.searchTreeRenderKey += 1;
    },
    openSearchResult(result) {
      this.sidebarView = 'files';
      this.$emit('open-search-result', {
        result,
        keyword: this.searchKeyword,
      });
    },
    async fetchGitLog() {
      if (!this.customSkillId) return;
      this.gitLoading = true;
      try {
        const res = await getSkillWorkspaceGitLog(this.customSkillId, {
          count: 50,
        });
        this.gitCommits = (res.data && res.data.commits) || [];
      } catch (e) {
        console.error('fetchGitLog error', e);
      } finally {
        this.gitLoading = false;
      }
    },
    async fetchGitStatus() {
      if (!this.customSkillId) return;
      this.gitStatusLoading = true;
      try {
        const res = await getSkillWorkspaceGitStatus(this.customSkillId);
        this.gitStatusFiles = (res.data && res.data.files) || [];
      } catch (e) {
        console.error('fetchGitStatus error', e);
      } finally {
        this.gitStatusLoading = false;
      }
    },
    async gitStageFile(path) {
      if (!this.customSkillId) return;
      try {
        await postSkillWorkspaceGitAdd(this.customSkillId, { paths: [path] });
        await this.fetchGitStatus();
      } catch (e) {
        this.$message.error(
          this.$t('generalAgent.skill.skillWorkBench.git.stageFailed'),
        );
        console.error('gitStageFile error', e);
      }
    },
    async gitStageAll() {
      if (!this.customSkillId) return;
      try {
        await postSkillWorkspaceGitAdd(this.customSkillId, { paths: [] });
        await this.fetchGitStatus();
      } catch (e) {
        this.$message.error(
          this.$t('generalAgent.skill.skillWorkBench.git.stageAllFailed'),
        );
        console.error('gitStageAll error', e);
      }
    },
    async gitUnstageFile(path) {
      if (!this.customSkillId) return;
      try {
        await postSkillWorkspaceGitReset(this.customSkillId, {
          paths: [path],
        });
        await this.fetchGitStatus();
      } catch (e) {
        this.$message.error(
          this.$t('generalAgent.skill.skillWorkBench.git.unstageFailed'),
        );
        console.error('gitUnstageFile error', e);
      }
    },
    async gitDiscardFile(file) {
      if (!this.customSkillId || !file || !file.path) return;
      if (this.discardingPaths[file.path]) return;

      this.$set(this.discardingPaths, file.path, true);
      try {
        const res = await postSkillWorkspaceGitDiscard(this.customSkillId, {
          paths: [file.path],
        });
        if (res.code !== 0) {
          this.$message.error(
            res.msg ||
              this.$t('generalAgent.skill.skillWorkBench.git.discardFailed'),
          );
          return;
        }

        this.$emit('discard-file', {
          path: file.path,
          closeIfMissing: file.changeType === 'untracked',
        });
        this.refreshFiles();
        await this.fetchGitStatus();
      } catch (e) {
        this.$message.error(
          this.$t('generalAgent.skill.skillWorkBench.git.discardFailed'),
        );
        console.error('gitDiscardFile error', e);
      } finally {
        this.$delete(this.discardingPaths, file.path);
      }
    },
    async gitCommit() {
      if (!this.customSkillId || !this.gitCommitMessage.trim()) return;
      try {
        await postSkillWorkspaceGitCommit(this.customSkillId, {
          message: this.gitCommitMessage.trim(),
        });
        this.$message.success(
          this.$t('generalAgent.skill.skillWorkBench.git.commitSuccess'),
        );
        this.gitCommitMessage = '';
        await this.refreshGit();
      } catch (e) {
        this.$message.error(
          this.$t('generalAgent.skill.skillWorkBench.git.commitFailed'),
        );
        console.error('gitCommit error', e);
      }
    },
    async restoreGitCommit(commit) {
      if (
        !this.customSkillId ||
        !commit ||
        !commit.hash ||
        this.restoringCommitHash
      ) {
        return;
      }

      try {
        await this.$confirm(
          this.$t('generalAgent.skill.skillWorkBench.git.restoreConfirm'),
          this.$t('common.confirm.title'),
          {
            confirmButtonText: this.$t('common.button.confirm'),
            cancelButtonText: this.$t('common.button.cancel'),
            type: 'warning',
          },
        );
      } catch (e) {
        return;
      }

      this.restoringCommitHash = commit.hash;
      try {
        const res = await postSkillWorkspaceGitRestore(
          this.customSkillId,
          commit.hash,
        );
        if (res.code !== 0) {
          this.$message.error(
            res.msg ||
              this.$t('generalAgent.skill.skillWorkBench.git.restoreFailed'),
          );
          return;
        }

        this.$message.success(this.$t('common.message.success'));
        this.$emit('workspace-restored');
      } catch (e) {
        this.$message.error(
          this.$t('generalAgent.skill.skillWorkBench.git.restoreFailed'),
        );
        console.error('restoreGitCommit error', e);
      } finally {
        this.restoringCommitHash = '';
      }
    },
    async selectWorkingFile(file, staged) {
      const diffId = this.workingDiffId(file, staged);
      const title = file.path.split('/').pop();
      this.$emit('open-git-diff', {
        id: diffId,
        type: 'working',
        title,
        file: { ...file, staged },
        loading: true,
        original: '',
        modified: '',
        patch: '',
      });

      try {
        const request = staged
          ? getSkillWorkspaceGitDiffStaged
          : getSkillWorkspaceGitDiffWorking;
        const res = await request(this.customSkillId, {
          filePath: file.path,
        });
        const data = res && res.data ? res.data : {};

        this.$emit('open-git-diff', {
          id: diffId,
          type: 'working',
          title,
          file: { ...file, staged },
          loading: false,
          original: data.oldContent || '',
          modified: data.newContent || '',
          patch: data.diff || '',
          changedFiles: data.changedFiles || null,
        });
      } catch (e) {
        this.$emit('open-git-diff', {
          id: diffId,
          type: 'working',
          title,
          file: { ...file, staged },
          loading: false,
          original: '',
          modified: '',
          patch: '',
        });
        console.error('selectWorkingFile error', e);
      }
    },
    async selectGitCommit(commit) {
      const diffId = this.commitDiffId(commit);
      this.$emit('open-git-diff', {
        id: diffId,
        type: 'commit',
        title: commit.message || commit.hash.substring(0, 7),
        commit,
        changedFiles: [],
        loading: true,
      });

      try {
        const res = await getSkillWorkspaceGitDiff(this.customSkillId, {
          fromCommit: commit.hash + '~1',
          toCommit: commit.hash,
        });
        this.$emit('open-git-diff', {
          id: diffId,
          type: 'commit',
          title: commit.message || commit.hash.substring(0, 7),
          commit,
          changedFiles: (res.data && res.data.changedFiles) || [],
          loading: false,
        });
      } catch (e) {
        console.error('selectGitCommit error', e);
        this.$emit('open-git-diff', {
          id: diffId,
          type: 'commit',
          title: commit.message || commit.hash.substring(0, 7),
          commit,
          changedFiles: [],
          loading: false,
        });
      }
    },
    async downloadFile(file) {
      if (!this.customSkillId || !file || !file.path) return;
      if (this.downloadingPaths[file.path]) return;

      this.$set(this.downloadingPaths, file.path, true);
      try {
        const blob = await downloadSkillWorkspace(
          this.customSkillId,
          file.path,
        );
        const fileName = this.resolveDownloadFileName(file, this.customSkillId);
        resDownloadFile(blob, fileName);
        this.$message.success(
          this.$t('generalAgent.workspace.downloadSuccess'),
        );
      } catch (error) {
        console.error('download skill workspace file failed:', error);
        this.$message.error(this.$t('generalAgent.workspace.downloadFailed'));
      } finally {
        this.$delete(this.downloadingPaths, file.path);
      }
    },
    async handleDeleteFile(file) {
      if (!this.customSkillId || !file || !file.path) return;
      this.$confirm(
        this.$t('generalAgent.skill.skillWorkBench.fileTree.confirmDelete', {
          name: file.name || file.path.split('/').pop(),
        }),
        this.$t('common.confirm.title'),
        {
          confirmButtonText: this.$t('common.button.confirm'),
          cancelButtonText: this.$t('common.button.cancel'),
          type: 'warning',
        },
      )
        .then(async () => {
          try {
            const res = await deleteSkillWorkspaceFile(
              this.customSkillId,
              file.path,
            );
            if (res.code === 0) {
              this.$message.success(this.$t('common.info.delete'));
              this.refreshFiles();
              this.refreshGit();
              this.$emit('close-tabs-by-path', file.path);
            } else {
              this.$message.error(res.msg || this.$t('common.info.deleteErr'));
            }
          } catch (error) {
            console.error('Delete file error:', error);
            this.$message.error(this.$t('common.info.deleteErr'));
          }
        })
        .catch(() => {});
    },
    async handleCreateFile(payload) {
      if (!payload || !payload.path || !this.customSkillId) return;
      try {
        const res = await createSkillWorkspaceFile(
          this.customSkillId,
          payload.path,
          '',
        );
        if (!res || res.code !== 0) {
          this.$message.error(
            (res && res.msg) ||
              this.$t(
                'generalAgent.skill.skillWorkBench.fileTree.createFailed',
              ),
          );
          await this.refreshFiles();
          return;
        }
        this.$message.success(
          this.$t('generalAgent.skill.skillWorkBench.fileTree.createSuccess'),
        );
        await this.refreshFiles();
        await this.fetchGitStatus();
      } catch (error) {
        await this.refreshFiles();
        this.$message.error(
          this.$t('generalAgent.skill.skillWorkBench.fileTree.createFailed'),
        );
      }
    },
    async handleCreateFolder(payload) {
      if (!payload || !payload.path || !this.customSkillId) return;
      try {
        const res = await createSkillWorkspaceDirectory(
          this.customSkillId,
          payload.path,
        );
        if (!res || res.code !== 0) {
          this.$message.error(
            (res && res.msg) ||
              this.$t(
                'generalAgent.skill.skillWorkBench.fileTree.createFailed',
              ),
          );
          await this.refreshFiles();
          return;
        }
        this.$message.success(
          this.$t('generalAgent.skill.skillWorkBench.fileTree.createSuccess'),
        );
        await this.refreshFiles();
        await this.fetchGitStatus();
      } catch (error) {
        await this.refreshFiles();
        this.$message.error(
          this.$t('generalAgent.skill.skillWorkBench.fileTree.createFailed'),
        );
      }
    },
    async handleRenameEntry(payload) {
      if (
        !payload ||
        !payload.entry ||
        !payload.entry.path ||
        !payload.newName
      ) {
        return;
      }
      if (!this.checkRenameAllowed(payload.entry.path)) {
        this.$message.warning(
          this.$t(
            'generalAgent.skill.skillWorkBench.fileTree.unsavedRenameWarning',
          ),
        );
        return;
      }
      try {
        const res = await renameSkillWorkspaceFile(
          this.customSkillId,
          payload.entry.path,
          payload.newName,
        );
        if (!res || res.code !== 0) {
          this.$message.error(
            (res && res.msg) ||
              this.$t(
                'generalAgent.skill.skillWorkBench.fileTree.renameFailed',
              ),
          );
          this.refreshFiles();
          return;
        }
        this.$message.success(
          this.$t('generalAgent.skill.skillWorkBench.fileTree.renameSuccess'),
        );
        this.refreshFiles();
        await this.fetchGitStatus();
        this.$emit('close-tabs-by-path', payload.entry.path);
      } catch (error) {
        this.$message.error(
          this.$t('generalAgent.skill.skillWorkBench.fileTree.renameFailed'),
        );
      }
    },
    async handleUploadFiles(payload) {
      const files = (payload && payload.files) || [];
      if (!this.customSkillId || files.length === 0) return;
      if (files.some(file => file.webkitRelativePath)) {
        this.$message.warning(
          this.$t(
            'generalAgent.skill.skillWorkBench.fileTree.directoryUploadNotSupported',
          ),
        );
        return;
      }
      try {
        const res = await uploadSkillWorkspaceFiles(
          this.customSkillId,
          (payload && payload.path) || '',
          files,
        );
        if (!res || res.code !== 0) {
          this.$message.error(
            (res && res.msg) ||
              this.$t(
                'generalAgent.skill.skillWorkBench.fileTree.uploadFailed',
              ),
          );
          this.refreshFiles();
          return;
        }
        this.$message.success(
          this.$t('generalAgent.skill.skillWorkBench.fileTree.uploadSuccess'),
        );
        this.refreshFiles();
        await this.fetchGitStatus();
      } catch (error) {
        this.$message.error(
          this.$t('generalAgent.skill.skillWorkBench.fileTree.uploadFailed'),
        );
      }
    },
    resolveDownloadFileName(file, customSkillId) {
      if (!file.isDir) return file.name || file.path.split('/').pop();
      const dirName = file.name || file.path.split('/').pop();
      return `workspace_${customSkillId}_${dirName}.zip`;
    },
    toggleGroup(group) {
      this.$set(this.groupExpanded, group, !this.groupExpanded[group]);
    },
    workingDiffId(file, staged) {
      return `git-working:${staged ? 'staged' : 'unstaged'}:${file.path}`;
    },
    commitDiffId(commit) {
      return `git-commit:${commit.hash}`;
    },
    latestTag(tags) {
      if (!tags || !tags.length) return '';
      const parseVer = t => {
        const m = t.match(/v?(\d+(?:\.\d+)*)/i);
        if (!m) return null;
        return m[1].split('.').map(n => parseInt(n, 10) || 0);
      };
      const cmp = (a, b) => {
        const la = a.length;
        const lb = b.length;
        const max = Math.max(la, lb);
        for (let i = 0; i < max; i++) {
          const da = a[i] || 0;
          const db = b[i] || 0;
          if (da !== db) return db - da;
        }
        return 0;
      };
      const parsed = tags
        .map(t => ({ t, v: parseVer(t) }))
        .sort((a, b) => {
          if (a.v && b.v) return cmp(a.v, b.v);
          if (a.v) return -1;
          if (b.v) return 1;
          return b.t.localeCompare(a.t);
        });
      return parsed[0].t;
    },
    formatGitTime(timestamp) {
      if (!timestamp) return '';
      const date = new Date(timestamp * 1000);
      const now = new Date();
      const diff = now - date;
      if (diff < MS_PER_MINUTE)
        return this.$t('generalAgent.skill.skillWorkBench.git.justNow');
      if (diff < MS_PER_HOUR)
        return this.$t('generalAgent.skill.skillWorkBench.git.minutesAgo', {
          count: Math.floor(diff / MS_PER_MINUTE),
        });
      if (diff < MS_PER_DAY)
        return this.$t('generalAgent.skill.skillWorkBench.git.hoursAgo', {
          count: Math.floor(diff / MS_PER_HOUR),
        });
      return date.toLocaleDateString();
    },
    getFileIcon,
  },
};
</script>

<style lang="scss" scoped>
.skill-workspace-explorer {
  width: 240px;
  height: 100%;
  border-right: 1px solid #e0e0e0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background: #f3f3f3;
  color: #333;
  flex-shrink: 0;

  .search-panel,
  .git-panel {
    height: 100%;
    display: flex;
    flex-direction: column;
    background: #f3f3f3;
    color: #333;
  }

  .panel-header {
    padding: 6px 8px;
    border-bottom: 1px solid #e0e0e0;
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: #f8f8f8;

    .header-icons {
      display: flex;
      gap: 4px;
    }

    .git-icon-wrap {
      position: relative;
      display: inline-flex;
    }

    .git-change-count {
      position: absolute;
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

  .search-input-wrap {
    padding: 8px;

    .search-options {
      display: flex;
      align-items: center;
      gap: 4px;
      margin-top: 6px;

      .option-btn,
      .option-icon {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        width: 24px;
        height: 22px;
        color: #666;
        cursor: pointer;
        border-radius: 3px;
        user-select: none;

        &:hover {
          background: #e8e8e8;
        }
        &.active {
          color: #fff;
          background: #5983ff;
        }
      }

      .option-btn {
        font-size: 11px;
      }

      .option-icon {
        font-size: 14px;
        &.active {
          color: #5983ff;
          background: rgba(89, 131, 255, 0.12);
        }
      }
    }

    .advanced-options {
      margin-top: 6px;
      display: flex;
      flex-direction: column;
      gap: 6px;
    }

    .search-status {
      margin-top: 6px;
      font-size: 12px;
      color: #888;
      display: flex;
      align-items: center;
      gap: 4px;

      i {
        font-size: 14px;
      }
    }
  }

  .search-results {
    flex: 1;
    overflow-y: auto;

    .result-toolbar {
      padding: 6px 12px;
      display: flex;
      justify-content: space-between;
      align-items: center;
      border-bottom: 1px solid #e8e8e8;
      background: #fafafa;

      .result-count {
        font-size: 11px;
        color: #888;
      }

      .toolbar-actions {
        display: flex;
        gap: 8px;

        i {
          font-size: 14px;
          color: #666;
          cursor: pointer;

          &:hover {
            color: #333;
          }
        }
      }
    }

    .result-item {
      padding: 4px 12px;
      cursor: pointer;

      &:hover {
        background: #e8e8e8;
      }

      .result-file {
        display: flex;
        align-items: center;
        font-size: 12px;
        color: #333;
        i {
          margin-right: 4px;
        }
        .file-path {
          max-width: 170px;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
        .result-line {
          color: #888;
          margin-left: 2px;
          flex-shrink: 0;
        }
      }

      .result-content {
        font-size: 12px;
        color: #888;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        padding-left: 16px;
      }
    }
  }

  .search-empty,
  .git-empty {
    padding: 16px 12px;
    font-size: 13px;
    color: #888;
    text-align: center;
  }

  .git-staging-area {
    max-height: 50%;
    overflow-y: auto;
    border-bottom: 1px solid #e0e0e0;
  }

  .git-section-title {
    padding: 6px 12px;
    font-size: 11px;
    font-weight: 600;
    color: #888;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    background: #fafafa;
    border-bottom: 1px solid #f0f0f0;
    display: flex;
    justify-content: space-between;
    align-items: center;

    .section-actions {
      display: flex;
      gap: 4px;

      .section-action-icon {
        font-size: 14px;
        color: #666;
        cursor: pointer;
        padding: 2px;
        border-radius: 3px;

        &:hover {
          color: #333;
          background: #e8e8e8;
        }
      }
    }
  }

  .git-group-header {
    display: flex;
    align-items: center;
    padding: 5px 12px;
    font-size: 11px;
    font-weight: 600;
    color: #555;
    cursor: pointer;
    background: #f5f5f5;
    border-bottom: 1px solid #f0f0f0;

    &:hover {
      background: #eee;
    }

    i {
      margin-right: 4px;
      font-size: 10px;
    }
  }

  .git-file-item {
    display: flex;
    align-items: center;
    padding: 4px 12px;
    cursor: pointer;
    font-size: 12px;

    &:hover {
      background: #f5f5f5;
    }
    &.active {
      background: rgba(89, 131, 255, 0.08);
    }

    .file-path {
      flex: 1;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .file-action {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      width: 20px;
      height: 20px;
      font-size: 12px;
      line-height: 1;
      color: #999;
      cursor: pointer;
      border-radius: 3px;
      flex-shrink: 0;

      &:hover {
        color: #333;
        background: #e0e0e0;
      }

      &.svg-file-action {
        font-size: 12px;
      }

      &.svg-file-action ::v-deep .abandon-icon {
        transform: rotate(45deg);
      }

      &.disabled {
        opacity: 0.5;
        cursor: wait;
        pointer-events: none;
      }
    }
  }

  .change-type-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 16px;
    height: 16px;
    font-size: 10px;
    font-weight: bold;
    border-radius: 3px;
    margin-right: 8px;
    padding: 0 2px;
    flex-shrink: 0;

    &.added {
      background: #e6ffed;
      color: #22863a;
    }
    &.modified {
      background: #fff8c5;
      color: #b08800;
    }
    &.deleted {
      background: #ffebe9;
      color: #cb2431;
    }
    &.renamed {
      background: #ddf4ff;
      color: #0969da;
    }
    &.untracked {
      background: #f0f0f0;
      color: #666;
    }
  }

  .git-commit-input {
    display: flex;
    gap: 6px;
    padding: 8px;
    border-top: 1px solid #f0f0f0;
    background: #fafafa;

    .el-input {
      flex: 1;
    }
  }

  .git-commit-list {
    flex: 1;
    overflow-y: auto;
    padding: 0;
  }

  .git-commit-item {
    position: relative;
    padding: 8px 12px;
    padding-right: 36px;
    cursor: pointer;
    border-bottom: 1px solid #f0f0f0;

    &:hover {
      background: #f5f5f5;
    }
    &.active {
      background: rgba(89, 131, 255, 0.08);
      box-shadow: inset 0 0 0 1px rgba(89, 131, 255, 0.16);
    }

    ::v-deep(.commit-restore-btn) {
      position: absolute;
      top: 6px;
      right: 8px;
      width: 22px;
      height: 22px;
      padding: 0;
      color: #8a8f99 !important;
      border-radius: 3px;
      span {
        color: #8a8f99 !important;
      }

      &:hover {
        color: #333 !important;
        background: #e0e0e0 !important;
      }

      .restore-icon {
        width: 14px;
        height: 14px;
      }
    }

    .commit-message {
      font-size: 13px;
      color: #333;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .commit-meta {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 4px;
      margin-top: 4px;
      font-size: 11px;
      color: #999;

      .commit-hash {
        font-family: monospace;
        flex-shrink: 0;
      }

      .commit-tags {
        display: inline-flex;
        flex-wrap: wrap;
        gap: 3px;
        overflow: hidden;
        min-width: 0;
      }

      .commit-tag {
        font-size: 10px;
        line-height: 1.4;
        padding: 0 5px;
        border-radius: 3px;
        background: rgba(89, 131, 255, 0.1);
        color: #4a6fdb;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        max-width: 80px;

        &.commit-tag-more {
          max-width: none;
          padding: 0 4px;
          background: rgba(0, 0, 0, 0.05);
          color: #999;
        }
      }

      .commit-time {
        flex-shrink: 0;
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
}
</style>
