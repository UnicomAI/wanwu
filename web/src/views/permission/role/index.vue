<template>
  <div class="role-page">
    <org-switcher
      v-model="selectedOrgId"
      class="role-page__switcher"
      @change="handleOrgChange"
    />
    <div class="table-wrap list-common wrap-fullheight role-page__main">
      <div class="role-page__content">
        <div class="section-title">{{ $t('role.title') }}</div>
        <div class="section-desc">{{ $t('role.desc') }}</div>
        <div class="toolbar">
          <search-input
            :placeholder="$t('role.form.role')"
            ref="searchInput"
            @handleSearch="getTableData"
          />
          <el-button
            class="add-bt"
            size="mini"
            type="primary"
            @click="preUpdate()"
          >
            <img src="@/assets/imgs/addRole.png" alt="" />
            <span>{{ $t('role.button.create') }}</span>
          </el-button>
        </div>

        <div v-loading="loading" class="role-grid card-wrap">
          <div
            v-for="item in tableData"
            :key="item.roleId"
            class="role-card"
            @click="showRoleDetail(item)"
          >
            <!-- 第一行：左侧头像+名称，右侧开关 -->
            <div class="card-header">
              <div class="header-left">
                <div class="avatar">
                  <img
                    :src="avatarSrc(item.avatar?.path || defaultAvatar)"
                    alt=""
                  />
                </div>
                <el-tooltip
                  :content="item.name"
                  placement="top"
                  :open-delay="300"
                >
                  <span class="role-name">{{ item.name }}</span>
                </el-tooltip>
                <el-tag v-if="item.isGlobal" size="mini" class="role-scope-tag">
                  {{ $t('role.card.scopeGlobal') }}
                </el-tag>
                <el-tag v-if="item.isAdmin" size="mini" class="role-scope-tag">
                  {{ $t('role.card.scopeBuiltin') }}
                </el-tag>
              </div>
              <div @click.stop>
                <el-switch
                  @change="
                    val => {
                      changeStatus(item, val);
                    }
                  "
                  :disabled="
                    !(
                      !item.isAdmin &&
                      (!item.isGlobal || (item.isGlobal && isSystem))
                    )
                  "
                  v-model="item.status"
                />
              </div>
            </div>

            <!-- 第二行：描述 -->
            <div class="card-desc">
              <template v-if="item.remark">
                <el-tooltip :content="item.remark">
                  <span class="role-name">{{ item.remark }}</span>
                </el-tooltip>
              </template>
              <template v-else>--</template>
            </div>

            <!-- 第三行：分配权限（多彩标签） -->
            <div class="card-perms">
              <template v-if="item.permissions && item.permissions.length">
                <el-tag
                  v-for="(perm, idx) in getRolePermissions(
                    item.permissions,
                  ).slice(0, 5)"
                  :key="idx"
                  size="mini"
                  :color="getPermColor(idx).backgroundColor"
                  :style="{ color: getPermColor(idx).color }"
                  class="perm-tag"
                >
                  {{ perm.name }}
                </el-tag>
                <template
                  v-if="getRolePermissions(item.permissions).length > 5"
                >
                  <el-tooltip
                    placement="top"
                    :content="
                      getRolePermissions(item.permissions)
                        .slice(5)
                        .map(p => p.name)
                        .join('、')
                    "
                  >
                    <el-tag size="mini" class="perm-tag perm-tag-more">
                      +{{ getRolePermissions(item.permissions).length - 5 }}
                    </el-tag>
                  </el-tooltip>
                </template>
              </template>
              <span v-else class="text-muted">--</span>
            </div>

            <!-- 第四行：用户数量 + 编辑/删除图标 -->
            <div class="card-footer">
              <span class="user-count">
                <i class="el-icon-user user-icon"></i>
                {{ $t('role.card.userCount', { count: item.userCount || 0 }) }}
              </span>
              <div
                class="card-actions"
                v-if="
                  !item.isAdmin &&
                  (!item.isGlobal || (item.isGlobal && isSystem))
                "
              >
                <i
                  class="el-icon-edit action-icon"
                  @click.stop="preUpdate(item)"
                ></i>
                <i
                  class="el-icon-delete action-icon delete-icon"
                  @click.stop="preDel(item)"
                ></i>
              </div>
            </div>
          </div>
        </div>
        <el-empty
          class="noData"
          v-if="!loading && !(tableData && tableData.length)"
          :description="$t('common.noData')"
        ></el-empty>
      </div>
    </div>

    <el-dialog
      :title="
        isEdit
          ? isSystem
            ? $t('role.button.globalEdit')
            : $t('role.button.edit')
          : isSystem
            ? $t('role.button.globalCreate')
            : $t('role.button.create')
      "
      :visible.sync="dialogVisible"
      width="580px"
      append-to-body
      :close-on-click-modal="false"
      :before-close="handleClose"
    >
      <el-form
        :model="form"
        :rules="rules"
        ref="form"
        style="margin-top: -16px"
      >
        <el-form-item :label="$t('role.table.avatar')" prop="avatar">
          <upload-avatar v-model="form.avatar" />
        </el-form-item>
        <el-form-item :label="$t('role.table.name')" prop="name">
          <el-input
            v-model="form.name"
            :placeholder="$t('common.hint.roleName')"
            clearable
          />
        </el-form-item>
        <el-form-item
          :label="$t('role.dialog.remark')"
          prop="remark"
          class="mark-textArea"
        >
          <el-input
            type="textarea"
            :rows="3"
            v-model="form.remark"
            :placeholder="$t('common.input.placeholder')"
            maxlength="100"
            show-word-limit
            clearable
          />
        </el-form-item>
        <el-form-item :label="$t('role.dialog.perm')" prop="permissions">
          <select-tree
            ref="permTree"
            :data-list="permList"
            :default-value="defaultPermValue"
            :tree-key-map="{ value: 'perm' }"
            :disabled="row.isAdmin"
            @handleChange="changeTree"
          />
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button size="small" @click="handleClose">
          {{ $t('common.button.cancel') }}
        </el-button>
        <el-button
          size="small"
          type="primary"
          :loading="submitLoading"
          @click="handleSubmit"
        >
          {{ $t('common.button.confirm') }}
        </el-button>
      </span>
    </el-dialog>

    <roleDetailDialog
      :isSystem="isSystem"
      :orgId="selectedOrgId"
      ref="roleDetailDialog"
      :preUpdate="preUpdate"
    />
  </div>
</template>

<script>
import SearchInput from '@/components/searchInput.vue';
import SelectTree from '../components/selectTree.vue';
import SelectOrgTree from '../components/selectOrgTree.vue';
import RoleDetailDialog from './roleDetailDialog.vue';
import OrgSwitcher from '../components/orgSwitcher.vue';
import {
  fetchPermTree,
  createRole,
  editRole,
  deleteRole,
  fetchRoleList,
  changeRoleStatus,
} from '@/api/permission/role';
import { mapActions } from 'vuex';
import { checkPerm } from '@/router/permission';
import { PERMS } from '@/router/constants';
import { avatarSrc } from '@/utils/util';
import { OrgTagColorList } from '@/utils/commonSet';
import UploadAvatar from '@/components/uploadAvatar.vue';

export default {
  components: {
    SearchInput,
    SelectTree,
    SelectOrgTree,
    RoleDetailDialog,
    OrgSwitcher,
    UploadAvatar,
  },
  data() {
    return {
      isSystem: false, // 是否系统下的判断不依赖外部的 isSystem 判断了，只依赖左侧组织树返回的 isSystem 字段判断
      selectedOrgId: '',
      loading: false,
      isEdit: false,
      permList: [],
      defaultPermValue: [],
      defaultAvatar: '/v1/static/icon/role-default-icon.png',
      form: {
        name: '',
        remark: '',
        permissions: [],
        avatar: {
          path: '',
          key: '',
        },
      },
      rules: {
        name: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
          {
            min: 1,
            max: 64,
            message: this.$t('common.hint.roleNameLimit'),
            trigger: 'blur',
          },
          {
            pattern: /^[a-zA-Z0-9_.\u4e00-\u9fa5]+$/,
            message: this.$t('common.hint.roleName'),
            trigger: 'blur',
          },
        ],
        remark: [
          {
            max: 100,
            message: this.$t('common.hint.remarkLimit'),
            trigger: 'blur',
          },
        ],
      },
      tableData: [],
      dialogVisible: false,
      submitLoading: false,
      row: {},
    };
  },
  methods: {
    ...mapActions('user', ['getPermissionInfo']),
    avatarSrc,
    changeTree(perms) {
      this.form.permissions = perms;
    },
    async getPermTree() {
      const { data } = await fetchPermTree({ orgId: this.selectedOrgId });
      this.permList = data.routes || [];
    },
    async getTableData() {
      const searchInput = this.$refs.searchInput;
      const searchInfo = {
        ...(searchInput.value && { name: searchInput.value }),
        orgId: this.selectedOrgId,
        pageNo: 1,
        pageSize: 100,
      };
      this.loading = true;
      try {
        const res = await fetchRoleList(searchInfo);
        this.tableData = res.data?.list || [];
      } finally {
        this.loading = false;
      }
    },
    showRoleDetail(item) {
      this.$refs.roleDetailDialog.openDialog({
        ...item,
        permList: this.getRolePermissions(item.permissions),
      });
    },
    handleOrgChange(org) {
      this.isSystem = org?.isSystem || false;
      this.getTableData();
      this.getPermTree();
    },
    setFormValue(row) {
      const obj = { ...this.form };
      for (let key in obj) {
        obj[key] =
          row && row[key] ? row[key] : Array.isArray(obj[key]) ? [] : '';
      }
      this.form = obj;
    },
    handleClose() {
      this.$refs.form.resetFields();
      this.dialogVisible = false;
    },
    getPermColor(index) {
      return OrgTagColorList[index % OrgTagColorList.length];
    },
    getRolePermissions(permissions) {
      const perms = permissions || [];
      const permKeys = perms.map(item => item.perm);
      return (
        perms
          .map(item => {
            if (permKeys.some(key => key.includes(`${item.perm}.`))) {
              return null;
            } else {
              return item;
            }
          })
          .filter(item => item) || []
      );
    },
    preUpdate(row) {
      this.row = row || {};
      this.isEdit = Boolean(row);
      const defaultAvatar = { path: this.defaultAvatar, key: '' };
      if (row) {
        // 处理一级权限返回问题
        const permissions = this.getRolePermissions(row.permissions);
        this.setFormValue({
          ...row,
          permissions: permissions.map(item => item.perm),
          avatar: row.avatar || defaultAvatar,
        });
        this.defaultPermValue = permissions || [];
      } else {
        this.setFormValue({ avatar: defaultAvatar });
        this.defaultPermValue = [];
      }

      this.dialogVisible = true;
      this.$nextTick(() => {
        this.$refs.form && this.$refs.form.clearValidate();
      });
    },
    preDel(row) {
      this.$confirm(
        this.$t('role.confirm.delete'),
        this.$t('common.confirm.title'),
        {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          type: 'warning',
        },
      ).then(async () => {
        let res = await deleteRole({
          roleId: row.roleId,
          orgId: this.selectedOrgId,
        });
        if (res.code === 0) {
          this.$message.success(this.$t('common.message.success'));
          await this.getTableData();
        }
      });
    },
    changeStatus(row, val) {
      this.$confirm(
        val
          ? this.$t('role.switch.startHint')
          : this.$t('role.switch.stopHint'),
        this.$t('common.confirm.title'),
        {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          type: 'warning',
        },
      )
        .then(async () => {
          let res = await changeRoleStatus({
            roleId: row.roleId,
            status: val,
            orgId: this.selectedOrgId,
          });
          if (res.code === 0) {
            this.$message.success(this.$t('common.message.success'));
            await this.getTableData();
          }
        })
        .catch(() => {
          this.getTableData();
        });
    },
    handleSubmit() {
      this.$refs.form.validate(async valid => {
        if (!valid) return;

        this.submitLoading = true;
        const params = {
          ...this.form,
          isGlobal: this.isSystem,
          orgId: this.selectedOrgId,
        };
        if (this.isEdit) {
          params.roleId = this.row.roleId;
          params.isGlobal = this.row.isGlobal;
        }

        try {
          const res = this.isEdit
            ? await editRole(params)
            : await createRole(params);
          if (res.code === 0) {
            this.$message.success(this.$t('common.message.success'));
            this.dialogVisible = false;

            // 如果当前用户有这个角色，则更新权限
            const permission = this.$store.state.user.permission || {};
            const roles = permission.roles
              ? permission.roles.map(item => item.id)
              : [];
            if (roles.includes(this.row.roleId)) {
              await this.getPermissionInfo();
              if (checkPerm(PERMS.ADMIN_CENTER)) {
                await this.getTableData();
                return;
              }
              window.location.reload();
              return;
            }
            await this.getTableData();
          }
        } finally {
          this.submitLoading = false;
        }
      });
    },
  },
};
</script>

<style lang="scss" scoped>
.role-page {
  display: flex;
  align-items: flex-start;
  height: 100%;

  &__switcher {
    flex-shrink: 0;
    align-self: stretch;
  }

  &__main {
    flex: 1;
    min-width: 0;
    margin-left: 10px;
    background: #f8fafc;
    border-radius: 10px;
    padding: 14px 0;
  }

  &__content {
    height: calc(100vh - 210px);
    overflow-y: auto;
    padding: 0 14px 10px;
  }

  .section-title {
    font-size: 16px;
    font-weight: bold;
    color: #555;
  }

  .section-desc {
    font-size: 11px;
    color: #999;
    margin-bottom: 12px;
  }

  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 14px;

    .add-bt {
      flex-shrink: 0;
      margin-left: 16px;
      img {
        width: 16px;
        margin-right: 5px;
        display: inline-block;
        vertical-align: middle;
      }
      span {
        display: inline-block;
        vertical-align: middle;
      }
    }
  }

  .role-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 16px;
  }

  .role-card {
    background: #fff;
    border-radius: 8px;
    padding: 20px;
    display: flex;
    flex-direction: column;
    box-shadow: 0 2px 8px 0 rgba(16, 18, 25, 0.102);
    min-width: 0;
    overflow: hidden;
    cursor: pointer;
    &:hover {
      box-shadow: 0 4px 10px 0 rgba(16, 18, 25, 0.12);
    }

    /* 第一行：左侧头像+名称，右侧开关 */
    .card-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 12px;
      min-width: 0;

      .header-left {
        display: flex;
        align-items: center;
        flex: 1;
        min-width: 0;
      }

      .avatar {
        width: 40px;
        height: 40px;
        border-radius: 8px;
        object-fit: cover;
        background: #fff;
        box-shadow: 0 1px 4px 0 rgba(0, 0, 0, 0.15);
        margin-right: 10px;
        & img {
          width: 100%;
          height: 100%;
          border-radius: 8px;
        }
      }

      .role-name {
        font-size: 15px;
        font-weight: 600;
        color: #303133;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        /*flex: 1;*/
        margin-right: 8px;
      }
    }

    .role-scope-tag {
      color: $tag_color !important;
      background-color: $tag_bg !important;
      border: none !important;
      margin-right: 8px;
    }

    /* 第二行：描述 */
    .card-desc {
      font-size: 13px;
      color: #909399;
      line-height: 1.5;
      margin-bottom: 14px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    /* 第三行：权限标签（多彩） */
    .card-perms {
      display: flex;
      flex-wrap: wrap;
      gap: 6px;
      margin-bottom: 18px;
      flex: 1;

      .perm-tag {
        margin: 0;
        border: none !important;

        &.perm-tag-more {
          background: $tag_bg !important;
          color: $tag_color !important;
        }
      }

      .text-muted {
        font-size: 12px;
        color: #c0c4cc;
      }
    }

    /* 第四行：用户数量 + 编辑/删除图标 */
    .card-footer {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding-top: 8px;

      .user-count {
        font-size: 12px;
        color: #909399;
        display: flex;
        align-items: center;

        .user-icon {
          margin-right: 4px;
        }
      }

      .card-actions {
        display: flex;
        align-items: center;
        gap: 14px;

        .action-icon {
          font-size: 16px;
          color: #909399;
          cursor: pointer;
          transition: color 0.2s;

          &:hover {
            color: #409eff;
          }

          &.delete-icon:hover {
            color: #f56c6c;
          }
        }
      }
    }
  }
}

.mark-textArea ::v-deep {
  .el-textarea__inner {
    font-family: inherit;
    font-size: inherit;
  }
}
</style>
