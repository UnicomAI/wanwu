<template>
  <div class="org-page">
    <org-switcher
      ref="orgSwitcher"
      v-model="selectedOrgId"
      class="org-page__switcher"
      @change="handleOrgChange"
    />
    <div class="table-wrap list-common org-page__main">
      <div class="org-page__content">
        <div class="table-box">
          <div class="org-section">
            <div class="section-title">{{ $t('org.currentOrgTitle') }}</div>
            <div class="section-desc">{{ $t('org.currentOrgDesc') }}</div>
            <el-table
              :data="currentOrgTableData"
              :header-cell-style="{ background: '#F9F9F9', color: '#999999' }"
              style="width: 100%"
            >
              <el-table-column :label="$t('org.table.name')" align="left">
                <template slot-scope="scope">
                  <div class="org-name-cell">
                    <img
                      class="org-name-cell__icon"
                      :src="avatarSrc(scope.row.avatar?.path || defaultAvatar)"
                      alt=""
                    />
                    <span>{{ scope.row.name }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column
                prop="admins"
                :label="$t('org.table.orgAdmin')"
                align="left"
              >
                <template slot-scope="scope">
                  <template v-if="scope.row.admins?.length > 0">
                    <el-tooltip :content="scope.row.admins.join('、')">
                      <div class="table-org-admin">
                        {{ scope.row.admins.join('、') }}
                      </div>
                    </el-tooltip>
                  </template>
                  <template v-else>--</template>
                </template>
              </el-table-column>
              <el-table-column :label="$t('org.table.members')" align="left">
                <template slot-scope="scope">
                  {{ scope.row.userCount || 0 }}
                </template>
              </el-table-column>
              <el-table-column
                prop="createdAt"
                :label="$t('org.table.createAt')"
                align="left"
              />
              <el-table-column
                align="left"
                :label="$t('common.table.operation')"
                width="120"
              >
                <template slot-scope="scope">
                  <el-button
                    class="operation"
                    type="text"
                    @click="preUpdate(scope.row)"
                  >
                    {{ $t('common.button.edit') }}
                  </el-button>
                  <el-button type="text" @click="preDel(scope.row)">
                    {{ $t('common.button.delete') }}
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
          <div class="org-section sub-org-section">
            <div class="section-title">{{ $t('org.subOrgTitle') }}</div>
            <div class="section-desc">{{ $t('org.subOrgDesc') }}</div>
            <div class="toolbar">
              <search-input
                :placeholder="$t('org.form.org')"
                ref="searchInput"
                @handleSearch="getTableData"
              />
              <el-button
                class="add-bt"
                size="mini"
                type="primary"
                @click="preUpdate()"
              >
                <img src="@/assets/imgs/addOrg.png" alt="" />
                <span>{{ $t('org.button.create') }}</span>
              </el-button>
            </div>
            <el-table
              :data="tableData"
              :header-cell-style="{ background: '#F9F9F9', color: '#999999' }"
              v-loading="loading"
              style="width: 100%"
            >
              <el-table-column :label="$t('org.table.name')" align="left">
                <template slot-scope="scope">
                  <div class="org-name-cell">
                    <img
                      class="org-name-cell__icon"
                      :src="avatarSrc(scope.row.avatar?.path || defaultAvatar)"
                      alt=""
                    />
                    <span>{{ scope.row.name }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column
                prop="admins"
                :label="$t('org.table.orgAdmin')"
                align="left"
              >
                <template slot-scope="scope">
                  <template v-if="scope.row.admins?.length > 0">
                    <el-tooltip :content="scope.row.admins.join('、')">
                      <div class="table-org-admin">
                        {{ scope.row.admins.join('、') }}
                      </div>
                    </el-tooltip>
                  </template>
                  <template v-else>--</template>
                </template>
              </el-table-column>
              <el-table-column :label="$t('org.table.members')" align="left">
                <template slot-scope="scope">
                  {{ scope.row.userCount || 0 }}
                </template>
              </el-table-column>
              <el-table-column
                align="left"
                :label="$t('org.table.status')"
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
                :label="$t('org.table.createAt')"
                align="left"
              />
              <el-table-column
                align="left"
                :label="$t('common.table.operation')"
                width="120"
              >
                <template slot-scope="scope">
                  <el-button
                    class="operation"
                    type="text"
                    @click="preUpdate(scope.row)"
                  >
                    {{ $t('common.button.edit') }}
                  </el-button>
                  <el-button type="text" @click="preDel(scope.row)">
                    {{ $t('common.button.delete') }}
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
    </div>

    <el-dialog
      :title="isEdit ? $t('org.button.edit') : $t('org.button.create')"
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
        <el-form-item :label="$t('org.table.avatar')" prop="avatar">
          <upload-avatar v-model="form.avatar" />
        </el-form-item>
        <el-form-item :label="$t('org.table.name')" prop="name">
          <el-input
            v-model="form.name"
            :placeholder="$t('common.hint.orgName')"
            clearable
          />
        </el-form-item>
        <el-form-item
          :label="$t('org.dialog.remark')"
          prop="remark"
          class="mark-textArea"
        >
          <el-input
            type="textarea"
            :rows="3"
            v-model="form.remark"
            maxlength="100"
            show-word-limit
            :placeholder="$t('common.input.placeholder')"
            clearable
          />
        </el-form-item>
        <!--组织管理员暂时不支持-->
        <!--<el-form-item :label="$t('org.table.orgAdmin')" prop="adminIds">
          <el-select
            v-model="form.adminIds"
            multiple
            filterable
            popper-class="org-admin-select-popper"
            :placeholder="$t('common.select.placeholder')"
            style="width: 100%"
          >
            <el-option
              v-for="item in userList"
              :key="item.userId"
              :label="item.name"
              :value="item.userId"
            >
              <div class="admin-option">
                <img
                  v-if="item.avatar && item.avatar.path"
                  class="admin-option__avatar"
                  :src="avatarSrc(item.avatar.path)"
                  alt=""
                />
                <span class="admin-option__name">{{ item.name }}</span>
              </div>
            </el-option>
          </el-select>
        </el-form-item>-->
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
  </div>
</template>

<script>
import Pagination from '@/components/pagination.vue';
import SearchInput from '@/components/searchInput.vue';
import OrgSwitcher from '@/views/permission/components/orgSwitcher.vue';
import {
  fetchOrgList,
  fetchOrgDetail,
  createOrg,
  editOrg,
  changeOrgStatus,
  deleteOrg,
} from '@/api/permission/org';
import { mapActions } from 'vuex';
import { avatarSrc } from '@/utils/util';
import UploadAvatar from '@/components/uploadAvatar.vue';

export default {
  components: { Pagination, SearchInput, OrgSwitcher, UploadAvatar },
  data() {
    return {
      listApi: fetchOrgList,
      selectedOrgId: '',
      currentOrg: {},
      defaultAvatar: '/v1/static/icon/org-default-icon.png',
      loading: false,
      isEdit: false,
      form: {
        name: '',
        // adminIds: [],
        remark: '',
        avatar: {
          path: '',
          key: '',
        },
      },
      userList: [],
      rules: {
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
    };
  },
  computed: {
    currentOrgTableData() {
      return this.currentOrg && this.currentOrg.orgId ? [this.currentOrg] : [];
    },
  },
  created() {},
  methods: {
    ...mapActions('user', ['getOrgInfo']),
    avatarSrc,
    handleOrgChange(org) {
      this.currentOrg = org || {};
      this.fetchCurrentOrgDetail();
      this.getTableData({ pageNo: 1 });
    },
    updateOrgTree(delId) {
      this.$refs.orgSwitcher.getOrgTree(delId);
    },
    async fetchCurrentOrgDetail() {
      const res = await fetchOrgDetail({ orgId: this.selectedOrgId });
      if (res.code === 0 && res.data) {
        this.currentOrg = res.data;
      }
    },
    async getTableData(params) {
      const searchInput = this.$refs.searchInput;
      const searchInfo = {
        ...(searchInput.value && { name: searchInput.value }),
        ...(this.selectedOrgId && { orgId: this.selectedOrgId }),
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
    setFormValue(row) {
      const obj = { ...this.form };
      for (let key in obj) {
        if (row) {
          obj[key] = row[key];
        } else {
          obj[key] = Array.isArray(obj[key]) ? [] : '';
        }
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
      this.setFormValue({
        ...row,
        avatar: row?.avatar || { path: this.defaultAvatar, key: '' },
      });

      this.dialogVisible = true;
    },
    preDel(row) {
      this.$confirm(
        this.$t('org.confirm.delete'),
        this.$t('common.confirm.title'),
        {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          type: 'warning',
        },
      ).then(async () => {
        let res = await deleteOrg({ orgId: row.orgId });
        if (res.code === 0) {
          this.$message.success(this.$t('common.message.success'));
          // 删除组织后，更新左侧的组织树
          this.updateOrgTree(row.orgId);
          await this.getTableData();
          await this.getOrgInfo();
        }
      });
    },
    changeStatus(row, val) {
      this.$confirm(
        val ? this.$t('org.switch.startHint') : this.$t('org.switch.stopHint'),
        this.$t('common.confirm.title'),
        {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          type: 'warning',
        },
      )
        .then(async () => {
          let res = await changeOrgStatus({ orgId: row.orgId, status: val });
          if (res.code === 0) {
            this.$message.success(this.$t('common.message.success'));
            await this.getTableData();
            await this.getOrgInfo();
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
        if (this.isEdit) {
          params.orgId = this.row.orgId;
        } else {
          params.orgId = this.selectedOrgId;
        }
        try {
          const res = this.isEdit
            ? await editOrg(params)
            : await createOrg(params);
          if (res.code === 0) {
            this.$message.success(this.$t('common.message.success'));
            this.dialogVisible = false;
            // 新增组织后，更新左侧的组织树
            if (!this.isEdit) this.updateOrgTree();
            await this.getTableData();
            await this.getOrgInfo();
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
.org-page {
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

  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }

  .org-section {
    margin-bottom: 20px;

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
  }

  .sub-org-section {
    margin-top: 24px;
  }

  .org-name-cell {
    display: flex;
    align-items: center;
    &__icon {
      width: 32px;
      height: 32px;
      border-radius: 6px;
      margin-right: 10px;
      flex-shrink: 0;
      object-fit: cover;
    }
  }

  .table-header {
    font-size: 16px;
    font-weight: bold;
    color: #555;
  }

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

  .table-org-admin {
    max-width: 160px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
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

.admin-option {
  display: flex;
  align-items: center;
  &__avatar {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    flex-shrink: 0;
    object-fit: cover;
  }

  &__name {
    margin-left: 10px;
    font-size: 13px;
    color: #303133;
  }
}

::v-deep .operation.el-button--text.el-button {
  padding: 3px 10px 3px 0;
  border-right: 1px solid #eaeaea !important;
}
</style>
