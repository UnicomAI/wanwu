<template>
  <div class="user-page">
    <org-switcher
      v-model="selectedOrgId"
      class="user-page__switcher"
      @change="handleOrgChange"
    />
    <div class="table-wrap list-common wrap-fullheight user-page__main">
      <div class="user-page__content">
        <div class="section-title">{{ $t('user.title') }}</div>
        <div class="toolbar">
          <div class="toolbar-left">
            <search-input
              :placeholder="$t('user.form.user')"
              ref="searchInput"
              @handleSearch="getTableData"
            />
            <el-select
              v-model="filterRoleIds"
              :placeholder="$t('user.form.roleFilter')"
              multiple
              collapse-tags
              clearable
              class="role-filter no-border-select"
              @change="searchData"
            >
              <el-option
                v-for="item in roleList"
                :key="item.id"
                :label="item.name"
                :value="item.id"
              />
            </el-select>
          </div>
          <div class="toolbar-actions">
            <el-dropdown v-if="!isSystem" @command="handleCommand">
              <el-button class="add-bt" size="mini" type="primary">
                <img src="@/assets/imgs/addUser.png" alt="" />
                {{ $t('user.button.create') }}
                <i class="el-icon-arrow-down"></i>
              </el-button>
              <el-dropdown-menu slot="dropdown">
                <el-dropdown-item command="onceAdd">
                  {{ $t('user.button.onceAdd') }}
                </el-dropdown-item>
                <el-dropdown-item command="batchAdd">
                  {{ $t('user.button.batchAdd') }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
            <el-button
              v-if="!isSystem"
              class="add-bt invite-bt"
              size="mini"
              @click="handleInviteUser"
            >
              <img src="@/assets/imgs/inviteUser.png" alt="" />
              <span>{{ $t('user.button.invite') }}</span>
            </el-button>
          </div>
        </div>
        <div class="table-box">
          <el-table
            :data="tableData"
            :header-cell-style="{ background: '#F9F9F9', color: '#999999' }"
            v-loading="loading"
            style="width: 100%"
          >
            <el-table-column :label="$t('user.table.username')" align="left">
              <template slot-scope="scope">
                <div class="user-cell">
                  <div class="user-cell__avatar">
                    <img
                      v-if="scope.row.avatar && scope.row.avatar.path"
                      :src="avatarSrc(scope.row.avatar.path)"
                      alt=""
                    />
                  </div>
                  <div
                    class="user-cell__info"
                    @click="showUserDetail(scope.row)"
                  >
                    <span class="user-cell__username">
                      {{ scope.row.username }}
                    </span>
                    <span class="user-cell__email">
                      {{ scope.row.email || '--' }}
                    </span>
                  </div>
                </div>
              </template>
            </el-table-column>
            <el-table-column
              v-if="!isSystem"
              :label="$t('user.detail.role')"
              align="left"
            >
              <template slot-scope="scope">
                <div
                  v-if="scope.row.orgs && scope.row.orgs.length"
                  v-for="(orgItem, orgIndex) in scope.row.orgs"
                  :key="orgItem.org.id + orgIndex"
                >
                  {{
                    Array.isArray(orgItem.roles)
                      ? orgItem.roles.map(item => item.name).join(',') || '--'
                      : '--'
                  }}
                </div>
              </template>
            </el-table-column>

            <el-table-column
              prop="phone"
              :label="$t('user.dialog.phone')"
              align="left"
            >
              <template slot-scope="scope">
                {{ scope.row.phone || '--' }}
              </template>
            </el-table-column>
            <el-table-column
              align="left"
              :label="$t('user.table.status')"
              width="100"
            >
              <template slot-scope="scope">
                <div style="height: 26px">
                  <el-switch
                    @change="
                      val => {
                        changeStatus(scope.row, val);
                      }
                    "
                    style="display: block; height: 22px; line-height: 22px"
                    v-model="scope.row.status"
                  />
                </div>
              </template>
            </el-table-column>
            <el-table-column
              prop="createdAt"
              :label="$t('user.table.createAt')"
              align="left"
            />
            <el-table-column
              align="left"
              :label="$t('common.table.operation')"
              width="180"
            >
              <template slot-scope="scope">
                <el-button
                  class="operation"
                  type="text"
                  @click="preUpdate(scope.row)"
                >
                  {{ $t('common.button.edit') }}
                </el-button>
                <el-button
                  class="operation"
                  type="text"
                  @click="preDel(scope.row)"
                >
                  {{
                    isSystem
                      ? $t('common.button.delete')
                      : $t('common.button.remove')
                  }}
                </el-button>
                <el-button type="text" @click="resetPsw(scope.row)">
                  {{ $t('user.table.resetPassword') }}
                </el-button>
              </template>
            </el-table-column>
          </el-table>
          <Pagination
            class="pagination"
            ref="pagination"
            :listApi="listApi"
            @refreshData="refreshData"
          />
        </div>
      </div>
    </div>

    <el-dialog
      :title="isEdit ? $t('user.button.edit') : $t('user.button.create')"
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
        <el-form-item :label="$t('user.table.username')" prop="username">
          <el-input
            v-model="form.username"
            :disabled="isEdit"
            :placeholder="$t('common.hint.userName')"
            clearable
          />
        </el-form-item>
        <el-form-item
          :label="$t('user.dialog.password')"
          :prop="!isEdit ? 'password' : ''"
        >
          <el-input
            v-model="form.password"
            type="password"
            :disabled="isEdit"
            :placeholder="isEdit ? '******' : $t('user.dialog.pwdPlaceholder')"
            clearable
          />
        </el-form-item>
        <el-form-item :label="$t('user.dialog.phone')" prop="phone">
          <el-input
            v-model="form.phone"
            :placeholder="$t('common.input.placeholder')"
            clearable
          />
        </el-form-item>
        <el-form-item :label="$t('user.dialog.email')" prop="email">
          <el-input
            :disabled="isEdit"
            v-model="form.email"
            :placeholder="
              isEdit ? $t('common.noBindEmail') : $t('common.input.placeholder')
            "
            clearable
          />
        </el-form-item>
        <el-form-item
          v-if="!isSystem"
          :label="$t('user.table.role')"
          prop="roleIds"
        >
          <el-select
            v-model="form.roleIds"
            :placeholder="$t('common.select.placeholder')"
            :disabled="row.username === 'admin'"
            style="width: 540px"
            clearable
          >
            <el-option
              v-for="item in roleList"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
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

    <!--邀请用户-->
    <el-dialog
      v-if="!isSystem"
      :title="$t('user.button.invite')"
      :visible.sync="inviteVisible"
      width="580px"
      append-to-body
      :close-on-click-modal="false"
      :before-close="handleInviteClose"
    >
      <el-form
        :model="inviteForm"
        :rules="inviteRules"
        ref="inviteForm"
        style="margin-top: -16px"
      >
        <el-form-item :label="$t('user.inviteDialog.user')" prop="userId">
          <el-select
            filterable
            v-model="inviteForm.userId"
            :placeholder="$t('user.inviteDialog.searchPlaceholder')"
            style="width: 540px"
            :filter-method="searchInviteUserList"
            clearable
          >
            <el-option
              v-for="item in inviteUserList"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button size="small" @click="handleInviteClose">
          {{ $t('common.button.cancel') }}
        </el-button>
        <el-button
          size="small"
          type="primary"
          :loading="submitLoading"
          @click="inviteUser"
        >
          {{ $t('common.button.confirm') }}
        </el-button>
      </span>
    </el-dialog>
    <resetPwd :orgId="selectedOrgId" ref="resetPwd" />
    <BatchAddDialog
      :orgId="selectedOrgId"
      ref="batchAdd"
      @reloadData="searchData"
    />
    <userDetailDialog
      ref="userDetailDialog"
      @resetPassword="handleDetailResetPassword"
      @toggleStatus="handleDetailToggleStatus"
    />
  </div>
</template>

<script>
import Pagination from '@/components/pagination.vue';
import resetPwd from '../components/resetPwd';
import SearchInput from '@/components/searchInput.vue';
import OrgSwitcher from '../components/orgSwitcher.vue';
import { avatarSrc } from '@/utils/util';
import { rsaEncrypt } from '@/utils/crypto';
import { debounce } from 'throttle-debounce';
import {
  fetchUserList,
  fetchInviteUser,
  fetchRoleList,
  inviteUser,
  createUser,
  editUser,
  deleteUser,
  changeUserStatus,
} from '@/api/permission/user';
import { mapActions } from 'vuex';
import { checkPerm } from '@/router/permission';
import { PERMS } from '@/router/constants';
import BatchAddDialog from './batchAddDialog.vue';
import UserDetailDialog from './userDetailDialog.vue';

export default {
  components: {
    Pagination,
    resetPwd,
    SearchInput,
    OrgSwitcher,
    BatchAddDialog,
    UserDetailDialog,
  },
  data() {
    const checkPassword = (rule, value, callback) => {
      let reg =
        /^(?=.*[a-zA-Z])(?=.*\d)(?=.*[~!@#$%^&*()_+`\-={}:";'<>?,./]).{8,20}$/;
      if (!reg.test(value)) {
        callback(new Error(this.$t('user.dialog.passwordError')));
      } else {
        return callback();
      }
    };
    const checkPhone = (rule, value, callback) => {
      let reg = /^1[3-9][0-9]{9}$/;
      if (value && !reg.test(value)) {
        callback(new Error(this.$t('user.dialog.phoneError')));
      } else {
        return callback();
      }
    };
    return {
      isSystem: false, // 是否系统下的判断不依赖外部的 isSystem 判断了，只依赖左侧组织树返回的 isSystem 字段判断
      selectedOrgId: '',
      listApi: fetchUserList,
      loading: false,
      isEdit: false,
      inviteLoading: false,
      inviteUserList: [],
      roleList: [],
      filterRoleIds: [],
      form: {
        username: '',
        password: '',
        phone: '',
        email: '',
        roleIds: '',
      },
      inviteForm: {
        userId: '',
      },
      rules: {
        username: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
          {
            min: 2,
            max: 20,
            message: this.$t('common.hint.userNameLimit'),
            trigger: 'blur',
          },
          {
            pattern: /^(?!_)[a-zA-Z0-9_.\u4e00-\u9fa5]+$/,
            message: this.$t('common.hint.userName'),
            trigger: 'blur',
          }, // 结尾：(?!.*?_$)
        ],
        password: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
          { validator: checkPassword, trigger: 'blur' },
        ],
        phone: [{ validator: checkPhone, trigger: 'blur' }],
        email: [
          // { required: true, message: this.$t('common.input.placeholder'), trigger: 'blur' },
          {
            pattern: /^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(.[a-zA-Z0-9_-]+)+$/,
            message: this.$t('common.hint.emailError'),
            trigger: 'blur',
          },
        ],
      },
      userPhoneRules: {
        phone: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
          { validator: checkPhone, trigger: 'blur' },
        ],
      },
      inviteRules: {
        userId: [
          {
            required: true,
            message: this.$t('common.select.placeholder'),
            trigger: 'change',
          },
        ],
      },
      tableData: [],
      inviteVisible: false,
      dialogVisible: false,
      submitLoading: false,
      row: {},
    };
  },
  created() {
    this.getInviteUserList = debounce(500, async name => {
      const { data } = await fetchInviteUser({
        name,
        orgId: this.selectedOrgId,
      });
      this.inviteUserList = data.select || [];
    });
  },
  methods: {
    ...mapActions('user', ['getPermissionInfo']),
    avatarSrc,
    handleCommand(command) {
      switch (command) {
        case 'onceAdd':
          this.preUpdate();
          break;
        case 'batchAdd':
          this.showBatchAddDialog();
          break;
      }
    },
    showBatchAddDialog() {
      this.$refs.batchAdd.openDialog();
    },
    async getRoleList() {
      const { data } = await fetchRoleList({ orgId: this.selectedOrgId });
      this.roleList = data.select || [];
    },
    searchData() {
      this.getTableData({ pageNo: 1 });
    },
    handleOrgChange(org) {
      this.isSystem = org?.isSystem || false;
      this.getTableData({ pageNo: 1 });
      this.getRoleList();
    },
    showUserDetail(row) {
      this.$refs.userDetailDialog.openDialog(row);
    },
    handleDetailResetPassword(user) {
      this.$refs.resetPwd.openDialog(user);
    },
    handleDetailToggleStatus(user) {
      const val = !user.status;
      this.changeStatus(user, val, () => {
        user.status = val;
      });
    },
    async getTableData(params) {
      const searchInput = this.$refs.searchInput;
      const searchInfo = {
        ...(searchInput.value && { name: searchInput.value }),
        ...(this.selectedOrgId && { orgId: this.selectedOrgId }),
        ...(this.filterRoleIds?.length && {
          roleIds: this.filterRoleIds.join(','),
        }),
        ...params,
      };
      this.loading = true;
      try {
        this.tableData = await this.$refs.pagination.getTableData(searchInfo);
      } finally {
        this.loading = false;
      }
    },
    // 获取从分页组件传递的 data
    refreshData(data) {
      this.tableData = data;
    },
    searchInviteUserList(val) {
      if (val) this.getInviteUserList(val);
    },
    handleInviteUser() {
      this.inviteVisible = true;
      this.getInviteUserList();
    },
    inviteUser() {
      this.$refs.inviteForm.validate(async valid => {
        if (!valid) return;
        this.submitLoading = true;
        try {
          const res = await inviteUser({
            ...this.inviteForm,
            orgId: this.selectedOrgId,
          });
          if (res.code === 0) {
            this.$message.success(this.$t('user.inviteDialog.success'));
            this.handleInviteClose();
            await this.getTableData();
          }
        } finally {
          this.submitLoading = false;
        }
      });
    },
    handleInviteClose() {
      this.inviteVisible = false;
      for (let key in this.inviteForm) {
        this.inviteForm[key] = '';
      }
      this.$refs.inviteForm.resetFields();
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
    preUpdate(row) {
      this.row = row || {};
      this.isEdit = Boolean(row);
      if (row) {
        const curOrg = row.orgs ? row.orgs[0] || {} : {};
        this.setFormValue({
          ...row,
          roleIds: curOrg.roles && curOrg.roles[0] ? curOrg.roles[0].id : '',
        });
      } else {
        this.setFormValue();
      }

      const commonInfo = this.$store.state.user.commonInfo.data || {};
      this.rules = commonInfo.userPhoneRequired
        ? { ...this.rules, ...this.userPhoneRules }
        : this.rules;
      this.dialogVisible = true;
      this.$nextTick(() => {
        this.$refs.form && this.$refs.form.clearValidate();
      });
    },
    preDel(row) {
      this.$confirm(
        this.isSystem
          ? this.$t('user.confirm.delete')
          : this.$t('user.confirm.remove'),
        this.$t('common.confirm.title'),
        {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          type: 'warning',
        },
      ).then(async () => {
        let res = await deleteUser({
          orgId: this.selectedOrgId,
          userId: row.userId,
        });
        if (res.code === 0) {
          this.$message.success(this.$t('common.message.success'));
          await this.getTableData();
        }
      });
    },
    resetPsw(row) {
      this.$refs.resetPwd.openDialog(row);
    },
    changeStatus(row, val, callback) {
      this.$confirm(
        val
          ? this.$t('user.switch.startHint')
          : this.$t('user.switch.stopHint'),
        this.$t('common.confirm.title'),
        {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          type: 'warning',
        },
      )
        .then(async () => {
          let res = await changeUserStatus({
            orgId: this.selectedOrgId,
            userId: row.userId,
            status: val,
          });
          if (res.code === 0) {
            this.$message.success(this.$t('common.message.success'));
            callback && callback();
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
        const params = { ...this.form };
        const { cipher, keyId } = await rsaEncrypt(this.form.password);
        params.cipher = cipher;
        params.keyId = keyId;
        delete params.password;
        params.roleIds = params.roleIds ? [params.roleIds] : [];
        params.orgId = this.selectedOrgId;
        if (this.isEdit) params.userId = this.row.userId;

        try {
          const res = this.isEdit
            ? await editUser(params)
            : await createUser(params);
          if (res.code === 0) {
            this.$message.success(this.$t('common.message.success'));
            this.dialogVisible = false;

            // 如果修改的是当前用户，则更新权限
            const useInfo = this.$store.state.user.userInfo || {};
            if (useInfo.uid === this.row.userId) {
              await this.getPermissionInfo();
              if (checkPerm(PERMS.ADMIN_CENTER)) {
                await this.getTableData();
                return;
              }
              location.reload();
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
.user-page {
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
    padding: 0 14px;
  }

  .section-title {
    font-size: 16px;
    font-weight: bold;
    color: #555;
    margin-bottom: 12px;
  }

  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }

  .toolbar-left {
    display: flex;
    align-items: center;
    flex: 1;
    min-width: 0;
  }

  .role-filter {
    width: 240px;
  }

  .toolbar-actions {
    display: flex;
    align-items: center;
    gap: 15px;
    flex-shrink: 0;
  }

  .table-header {
    font-size: 16px;
    font-weight: bold;
    color: #555;
  }

  .add-bt {
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

  .invite-bt {
    color: $color;
    border-color: $color;
    background: rgba(255, 255, 255, 0) !important;
    margin: 0;
  }

  ::v-deep .el-switch__label * {
    font-size: 13px;
  }
}

.mark-textArea ::v-deep {
  .el-textarea__inner {
    font-family: inherit;
    font-size: inherit;
  }
}

::v-deep .operation.el-button--text.el-button {
  padding: 3px 10px 3px 0;
  border-right: 1px solid #eaeaea !important;
}

.user-cell {
  display: flex;
  align-items: center;
  min-width: 0;
  padding: 4px 0;

  &__avatar {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    flex-shrink: 0;
    overflow: hidden;
    background: #f0f2f5;
    margin-right: 10px;
    img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
  }

  &__info {
    display: flex;
    flex-direction: column;
    min-width: 0;
    line-height: 1.4;
    cursor: pointer;
  }

  &__username {
    color: #333;
    cursor: pointer;
    &:hover {
      text-decoration: underline;
    }
  }

  &__email {
    font-size: 12px;
    color: #909399;
    margin-top: 2px;
    max-width: 200px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}
</style>
