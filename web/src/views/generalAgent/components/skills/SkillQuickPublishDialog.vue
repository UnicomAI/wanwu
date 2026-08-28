<template>
  <el-dialog
    :title="$t('generalAgent.skill.skillWorkBench.quickPublish.title')"
    :visible.sync="dialogVisible"
    width="680px"
    append-to-body
    :close-on-click-modal="!submitting"
    :close-on-press-escape="!submitting"
    :show-close="!submitting"
    :before-close="beforeClose"
    custom-class="quick-publish-dialog-wrapper"
  >
    <div class="quick-publish-dialog">
      <div
        v-if="publishFailed"
        class="quick-publish-notice quick-publish-notice--error"
      >
        <i class="el-icon-warning-outline"></i>
        <div>
          <div class="notice-title">
            {{
              $t(
                'generalAgent.skill.skillWorkBench.quickPublish.publishFailedTitle',
              )
            }}
          </div>
          <div class="notice-description">
            {{
              $t(
                'generalAgent.skill.skillWorkBench.quickPublish.publishFailedDescription',
              )
            }}
          </div>
        </div>
      </div>

      <template v-else>
        <div class="quick-publish-notice">
          <i class="el-icon-info"></i>
          <span>
            {{
              $t(
                'generalAgent.skill.skillWorkBench.quickPublish.scopeDescription',
              )
            }}
          </span>
        </div>

        <div class="change-scope-header">
          <span>
            {{
              $t('generalAgent.skill.skillWorkBench.quickPublish.changeScope', {
                count: localFiles.length,
              })
            }}
          </span>
          <el-button type="text" @click="viewGitDetails">
            {{
              $t('generalAgent.skill.skillWorkBench.quickPublish.viewDetails')
            }}
          </el-button>
        </div>

        <div class="change-scope-list">
          <div
            v-for="group in fileGroups"
            :key="group.key"
            v-show="group.files.length > 0"
            class="change-group"
          >
            <div class="change-group-title">
              {{ group.label }}
            </div>
            <div
              v-for="file in group.files"
              :key="`${group.key}-${file.path}-${file.changeType}`"
              class="change-file"
            >
              <span :class="['change-badge', file.changeType]">
                {{ changeTypeShortLabel(file.changeType) }}
              </span>
              <span class="change-path" :title="file.path">
                {{ file.path }}
              </span>
              <span class="change-type">
                {{ changeTypeLabel(file.changeType) }}
              </span>
            </div>
          </div>
        </div>

        <div class="commit-message-field">
          <div class="field-label">
            {{
              $t('generalAgent.skill.skillWorkBench.quickPublish.commitMessage')
            }}
          </div>
          <el-input
            v-model="commitMessage"
            type="textarea"
            :rows="3"
            :disabled="submitting"
            :placeholder="
              $t(
                'generalAgent.skill.skillWorkBench.quickPublish.commitMessagePlaceholder',
              )
            "
          />
          <div v-if="showCommitMessageError" class="field-error">
            {{
              $t(
                'generalAgent.skill.skillWorkBench.quickPublish.commitMessageRequired',
              )
            }}
          </div>
        </div>

        <div class="publish-summary">
          <div class="summary-item">
            <span class="summary-label">
              {{ $t('generalAgent.skill.skillWorkBench.quickPublish.version') }}
            </span>
            <span>{{ publishData.version }}</span>
          </div>
          <div class="summary-item">
            <span class="summary-label">
              {{ $t('generalAgent.skill.skillWorkBench.quickPublish.scope') }}
            </span>
            <span>{{ publishScopeLabel }}</span>
          </div>
          <div class="summary-item summary-item--full">
            <span class="summary-label">
              {{
                $t(
                  'generalAgent.skill.skillWorkBench.quickPublish.versionDescription',
                )
              }}
            </span>
            <span>{{ publishData.desc }}</span>
          </div>
        </div>
      </template>
    </div>

    <span slot="footer" class="dialog-footer">
      <el-button :disabled="submitting" @click="dialogVisible = false">
        {{ $t('common.button.cancel') }}
      </el-button>
      <el-button
        type="primary"
        :loading="submitting"
        :disabled="disabled || (!publishFailed && !commitMessage.trim())"
        @click="confirmQuickPublish"
      >
        {{ primaryButtonText }}
      </el-button>
    </span>
  </el-dialog>
</template>

<script>
import {
  getSkillWorkspaceGitStatus,
  postSkillWorkspaceGitAdd,
  postSkillWorkspaceGitCommit,
} from '@/api/skillResource/skillWorkSpace';

export default {
  name: 'SkillQuickPublishDialog',
  props: {
    value: {
      type: Boolean,
      default: false,
    },
    customSkillId: {
      type: String,
      required: true,
    },
    files: {
      type: Array,
      default: () => [],
    },
    publishData: {
      type: Object,
      default: () => ({}),
    },
    publishHandler: {
      type: Function,
      required: true,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      localFiles: [],
      initialStatusSnapshot: '',
      commitMessage: '',
      publishFormSignature: '',
      submitting: false,
      publishFailed: false,
      commitSucceeded: false,
      showCommitMessageError: false,
    };
  },
  computed: {
    dialogVisible: {
      get() {
        return this.value;
      },
      set(value) {
        if (!this.submitting) {
          this.$emit('input', value);
        }
      },
    },
    unstagedFiles() {
      return this.localFiles.filter(file => !file.staged);
    },
    stagedFiles() {
      return this.localFiles.filter(file => file.staged);
    },
    fileGroups() {
      return [
        {
          key: 'unstaged',
          label: this.$t(
            'generalAgent.skill.skillWorkBench.quickPublish.unstagedChanges',
            { count: this.unstagedFiles.length },
          ),
          files: this.unstagedFiles,
        },
        {
          key: 'staged',
          label: this.$t(
            'generalAgent.skill.skillWorkBench.quickPublish.stagedChanges',
            { count: this.stagedFiles.length },
          ),
          files: this.stagedFiles,
        },
      ];
    },
    publishScopeLabel() {
      const key = `app.commonPublishType.${this.publishData.publishType}`;
      return this.$te(key) ? this.$t(key) : this.publishData.publishType;
    },
    primaryButtonText() {
      if (this.submitting) {
        return this.publishFailed
          ? this.$t(
              'generalAgent.skill.skillWorkBench.quickPublish.retryingPublish',
            )
          : this.$t(
              'generalAgent.skill.skillWorkBench.quickPublish.submitting',
            );
      }
      return this.publishFailed
        ? this.$t('generalAgent.skill.skillWorkBench.quickPublish.retryPublish')
        : this.$t(
            'generalAgent.skill.skillWorkBench.quickPublish.confirmSubmit',
          );
    },
  },
  watch: {
    value(visible) {
      if (visible) {
        this.prepareDialog();
      }
    },
  },
  methods: {
    prepareDialog() {
      const signature = JSON.stringify({
        version: this.publishData.version,
        desc: this.publishData.desc,
        publishType: this.publishData.publishType,
      });

      if (signature !== this.publishFormSignature) {
        this.commitMessage = this.publishData.desc || '';
        this.publishFormSignature = signature;
      }

      this.localFiles = this.files.map(file => ({ ...file }));
      this.initialStatusSnapshot = this.createStatusSnapshot(this.localFiles);
      this.publishFailed = false;
      this.commitSucceeded = false;
      this.showCommitMessageError = false;
    },
    createStatusSnapshot(files) {
      return JSON.stringify(
        files
          .map(file => ({
            path: file.path,
            oldPath: file.oldPath || '',
            changeType: file.changeType,
            staged: Boolean(file.staged),
          }))
          .sort((a, b) => {
            const aKey = `${a.path}:${a.changeType}:${a.staged}`;
            const bKey = `${b.path}:${b.changeType}:${b.staged}`;
            return aKey.localeCompare(bKey);
          }),
      );
    },
    changeTypeShortLabel(changeType) {
      const labels = {
        added: 'A',
        untracked: 'A',
        modified: 'M',
        deleted: 'D',
        renamed: 'R',
      };
      return labels[changeType] || String(changeType || '?')[0].toUpperCase();
    },
    changeTypeLabel(changeType) {
      const key = `generalAgent.skill.skillWorkBench.quickPublish.changeTypes.${changeType}`;
      return this.$te(key) ? this.$t(key) : changeType;
    },
    beforeClose(done) {
      if (!this.submitting) done();
    },
    viewGitDetails() {
      if (this.submitting) return;
      this.$emit('view-git-details');
      this.dialogVisible = false;
    },
    async requestPublish() {
      try {
        const res = await this.publishHandler({ ...this.publishData });
        if (!res || res.code !== 0) {
          this.publishFailed = true;
          return false;
        }
        this.$emit('input', false);
        return true;
      } catch (error) {
        console.error('quick publish error', error);
        this.publishFailed = true;
        return false;
      }
    },
    async confirmQuickPublish() {
      if (this.disabled || this.submitting) return;

      if (this.publishFailed || this.commitSucceeded) {
        this.submitting = true;
        try {
          await this.requestPublish();
        } finally {
          this.submitting = false;
        }
        return;
      }

      if (!this.commitMessage.trim()) {
        this.showCommitMessageError = true;
        return;
      }

      this.showCommitMessageError = false;
      this.submitting = true;
      try {
        const statusRes = await getSkillWorkspaceGitStatus(this.customSkillId);
        if (statusRes.code !== 0) return;

        const latestFiles = (statusRes.data && statusRes.data.files) || [];
        const latestSnapshot = this.createStatusSnapshot(latestFiles);
        if (latestSnapshot !== this.initialStatusSnapshot) {
          this.localFiles = latestFiles;
          this.initialStatusSnapshot = latestSnapshot;
          this.$message.warning(
            this.$t(
              'generalAgent.skill.skillWorkBench.quickPublish.scopeChanged',
            ),
          );
          return;
        }

        if (latestFiles.length > 0) {
          const addRes = await postSkillWorkspaceGitAdd(this.customSkillId, {
            paths: [],
          });
          if (addRes.code !== 0) return;

          const commitRes = await postSkillWorkspaceGitCommit(
            this.customSkillId,
            {
              message: this.commitMessage.trim(),
            },
          );
          if (commitRes.code !== 0) return;

          this.commitSucceeded = true;
          this.$emit('git-changed');
        }

        await this.requestPublish();
      } catch (error) {
        console.error('quick submit and publish error', error);
      } finally {
        this.submitting = false;
      }
    },
  },
};
</script>

<style lang="scss" scoped>
.quick-publish-dialog {
  color: #303133;
}

::v-deep .quick-publish-dialog-wrapper {
  .el-dialog__header {
    padding-bottom: 10px;
  }
  .el-dialog__body {
    padding: 12px 16px;
  }
}

.quick-publish-notice {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 8px 12px;
  margin-bottom: 6px;
  color: #7a5d12;
  font-size: 14px;
  line-height: 20px;
  background: #fdf6ec;
  border: 1px solid #faecd8;
  border-radius: 4px;

  > i {
    margin-top: 2px;
    color: #e6a23c;
  }

  &--error {
    color: #c45656;
    background: #fef0f0;
    border-color: #fde2e2;

    > i {
      color: #f56c6c;
    }
  }
}

.notice-title {
  font-weight: 600;
}

.notice-description {
  margin-top: 2px;
  color: #606266;
}

.publish-summary {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px 24px;
  padding: 12px 16px;
  margin-top: 16px;
  background: #f7f8fa;
  border-radius: 4px;
}

.summary-item {
  display: flex;
  gap: 12px;
  min-width: 0;
  line-height: 20px;

  &--full {
    grid-column: 1 / -1;
  }
}

.summary-label {
  flex-shrink: 0;
  color: #909399;
}

.change-scope-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 6px;
  font-weight: 600;
}

.change-scope-list {
  max-height: 240px;
  overflow-y: auto;
  border: 1px solid #ebeef5;
  border-radius: 4px;
}

.change-group + .change-group {
  border-top: 1px solid #ebeef5;
}

.change-group-title {
  padding: 7px 12px;
  color: #606266;
  font-size: 12px;
  font-weight: 600;
  background: #f7f8fa;
}

.change-file {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 34px;
  padding: 0 12px;

  & + & {
    border-top: 1px solid #f2f3f5;
  }
}

.change-badge {
  width: 18px;
  flex-shrink: 0;
  color: #409eff;
  font-size: 12px;
  font-weight: 600;
  text-align: center;

  &.added,
  &.untracked {
    color: #67c23a;
  }

  &.deleted {
    color: #f56c6c;
  }

  &.renamed {
    color: #e6a23c;
  }
}

.change-path {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.change-type {
  flex-shrink: 0;
  color: #909399;
  font-size: 12px;
}

.commit-message-field {
  margin-top: 16px;
}

.field-label {
  margin-bottom: 8px;
  font-weight: 600;
}

.field-error {
  margin-top: 4px;
  color: #f56c6c;
  font-size: 12px;
}
</style>
