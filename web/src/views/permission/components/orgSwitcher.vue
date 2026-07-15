<template>
  <div class="org-switcher">
    <div class="org-switcher__search">
      <el-input
        class="org-switcher__search-input"
        v-model="searchKeyword"
        :placeholder="$t('orgSwitcher.searchPlaceholder')"
        prefix-icon="el-icon-search"
        size="small"
        clearable
      />
    </div>
    <div class="org-switcher__tree-wrap" v-loading="loading">
      <div
        v-if="orgTreeData.length === 0 && !loading"
        class="org-switcher__empty"
      >
        {{ $t('common.noData') }}
      </div>
      <el-tree
        v-else
        ref="orgTree"
        :data="orgTreeData"
        :props="treeProps"
        node-key="orgId"
        default-expand-all
        highlight-current
        :expand-on-click-node="false"
        :current-node-key="activeOrgId"
        :filter-node-method="filterNode"
        @node-click="handleNodeClick"
      >
        <span slot-scope="{ node, data }" class="org-tree-node">
          <span class="org-tree-node__label">
            <span class="org-tree-node__name" :title="data.name">
              {{ data.name }}
            </span>
            <!--<span class="org-tree-node__count">
              {{ data.userCount || 0 }}
            </span>-->
          </span>
          <!--重命名不支持-->
          <!--<span v-if="isAdmin" class="org-tree-node__actions">
            <el-dropdown
              trigger="hover"
              @command="handleOrgCommand(data, $event)"
            >
              <span class="org-tree-node__more" @click.stop>
                <i class="el-icon-more" />
              </span>
              <el-dropdown-menu slot="dropdown">
                <el-dropdown-item command="rename">
                  {{ $t('orgSwitcher.rename') }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </span>-->
        </span>
      </el-tree>
    </div>

    <!-- Rename dialog -->
    <el-dialog
      :title="$t('orgSwitcher.renameDialog.title')"
      :visible.sync="renameVisible"
      width="420px"
      append-to-body
      :close-on-click-modal="false"
      :before-close="handleRenameClose"
    >
      <el-form
        :model="renameForm"
        :rules="renameRules"
        ref="renameForm"
        style="margin-top: -16px"
      >
        <el-form-item :label="$t('org.table.name')" prop="name">
          <el-input
            v-model="renameForm.name"
            :placeholder="$t('common.hint.orgName')"
            clearable
          />
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button size="small" @click="handleRenameClose">
          {{ $t('common.button.cancel') }}
        </el-button>
        <el-button
          size="small"
          type="primary"
          :loading="renameLoading"
          @click="handleRenameSubmit"
        >
          {{ $t('common.button.confirm') }}
        </el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { mapActions } from 'vuex';
import { fetchOrgTree } from '@/api/permission/org';

export default {
  name: 'OrgSwitcher',
  props: {
    value: {
      type: [String, Number],
      default: '',
    },
  },
  data() {
    return {
      orgTreeData: [],
      loading: false,
      searchKeyword: '',
      activeOrgId: this.value,
      treeProps: {
        children: 'children',
        label: 'name',
      },
      renameVisible: false,
      renameLoading: false,
      renameForm: {
        name: '',
      },
      renameOrg: null,
      renameRules: {
        name: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
          {
            min: 1,
            max: 30,
            message: this.$t('common.hint.orgNameLimit'),
            trigger: 'blur',
          },
          {
            pattern: /^[a-zA-Z0-9-_.@\u4e00-\u9fa5]+$/,
            message: this.$t('common.hint.orgName'),
            trigger: 'blur',
          },
        ],
      },
    };
  },
  computed: {
    isAdmin() {
      return this.$store.state.user.permission.isAdmin || false;
    },
  },
  watch: {
    value(val) {
      this.activeOrgId = val;
      this.$refs.orgTree && this.$refs.orgTree.setCurrentKey(val);
    },
    searchKeyword(val) {
      this.$refs.orgTree && this.$refs.orgTree.filter(val);
    },
  },
  mounted() {
    this.getOrgTree();
  },
  methods: {
    ...mapActions('user', ['getOrgInfo']),
    filterNode(value, data) {
      if (!value) return true;
      return data.name && data.name.toLowerCase().includes(value.toLowerCase());
    },
    findFirstPermNode(nodes) {
      for (const node of nodes) {
        if (node.hasPerm) return node;
        if (node.children && node.children.length) {
          const found = this.findFirstPermNode(node.children);
          if (found) return found;
        }
      }
      return null;
    },
    getOrgTree(delId) {
      this.loading = true;
      fetchOrgTree()
        .then(res => {
          this.orgTreeData = res.data || [];
          this.loading = false;

          this.$nextTick(() => {
            // 默认选中第一个有权限的节点（如果当前没有选择 orgId；或者有选择的 orgId，但是删除的组织是当前选中的组织）
            const isInitActive =
              delId && this.activeOrgId && this.activeOrgId === delId;
            if (
              (!this.activeOrgId || isInitActive) &&
              this.orgTreeData.length > 0
            ) {
              const defaultNode = this.findFirstPermNode(this.orgTreeData);
              if (defaultNode) {
                this.activeOrgId = defaultNode.orgId;
                this.$emit('input', defaultNode.orgId);
                this.$emit('change', defaultNode);
                this.$refs.orgTree &&
                  this.$refs.orgTree.setCurrentKey(defaultNode.orgId);
              }
            } else if (this.activeOrgId) {
              this.$refs.orgTree &&
                this.$refs.orgTree.setCurrentKey(this.activeOrgId);
            }
          });
        })
        .catch(() => {
          this.loading = false;
        });
    },
    handleNodeClick(data) {
      if (!data.hasPerm) {
        this.$message.warning(this.$t('orgSwitcher.noPerm'));
        this.$refs.orgTree &&
          this.$refs.orgTree.setCurrentKey(this.activeOrgId);
        return;
      }
      this.activeOrgId = data.orgId;
      this.$emit('input', data.orgId);
      this.$emit('change', data);
    },
    handleOrgCommand(org, command) {
      if (command === 'rename') {
        this.handleRename(org);
      }
    },
    handleRename(org) {
      this.renameOrg = org;
      this.renameForm.name = org.name || '';
      this.renameVisible = true;
      this.$nextTick(() => {
        this.$refs.renameForm && this.$refs.renameForm.clearValidate();
      });
    },
    handleRenameClose() {
      this.renameVisible = false;
      this.renameOrg = null;
      this.renameForm.name = '';
      this.$refs.renameForm && this.$refs.renameForm.resetFields();
    },
    handleRenameSubmit() {
      this.$refs.renameForm.validate(valid => {
        if (!valid) return;
        this.renameLoading = true;
        try {
          // 重命名请求
        } finally {
          this.renameLoading = false;
        }
      });
    },
  },
};
</script>

<style lang="scss" scoped>
.org-switcher {
  width: 260px;
  min-width: 260px;
  background: #fff;
  border-right: 1px solid #eaeaea;
  display: flex;
  flex-direction: column;
  max-height: calc(100vh - 170px);
  overflow-y: auto;

  &__search {
    padding: 2px 20px 12px 0;
    flex-shrink: 0;
  }

  &__tree-wrap {
    flex: 1;
    overflow-y: auto;
    padding: 4px 12px 4px 0;
  }

  &__empty {
    padding: 40px 16px;
    text-align: center;
    color: #999;
    font-size: 13px;
  }

  &__search-input ::v-deep .el-input__inner {
    border-radius: 12px !important;
  }
}

.org-tree-node {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex: 1;
  font-size: 0;
  padding-right: 8px;
  min-width: 0;

  &__label {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex: 1;
    min-width: 0;
  }

  &__name {
    font-size: 13px;
    color: #333;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    flex: 1;
    min-width: 0;
  }

  &__count {
    font-size: 12px;
    color: #999;
    margin-left: 8px;
    flex-shrink: 0;
  }

  &__actions {
    display: flex;
    align-items: center;
    flex-shrink: 0;
    margin-left: 4px;
  }

  &__more {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    border-radius: 4px;
    cursor: pointer;
    color: #979797;
    font-size: 16px;

    i {
      transform: rotate(90deg);
    }
  }

  &__menu-icon {
    margin-right: 8px;
    font-size: 14px;
  }
}

::v-deep .el-tree {
  .el-tree-node__content {
    background: rgba(255, 255, 255, 0) !important;
    height: 36px;
    border-radius: 12px;
    margin: 2px 0;

    &:hover {
      background: #f5f7fa !important;
      transition: background 0.15s linear;
    }
  }

  .el-tree-node.is-current > .el-tree-node__content {
    background: $color_opacity !important;

    .org-tree-node__name {
      color: $color;
      font-weight: 500;
    }

    .org-tree-node__count,
    .el-icon-more {
      color: $color;
    }
  }
}
</style>
