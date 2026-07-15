<template>
  <el-dialog
    title=""
    :visible.sync="visible"
    width="760px"
    custom-class="role-detail-dialog"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <div v-if="role" class="role-detail">
      <!-- 头像 + 名称 -->
      <div class="detail-header">
        <div class="header-left">
          <div class="avatar">
            <img :src="avatarSrc(role.avatar?.path || defaultAvatar)" alt="" />
          </div>
          <div class="header-info">
            <span class="role-name">{{ role.name }}</span>
          </div>
        </div>
      </div>

      <!-- 详情字段 -->
      <div class="detail-body">
        <div class="detail-item">
          <div class="detail-label">{{ $t('role.detail.scope') }}</div>
          <div class="detail-value">
            {{
              role.scope === 'global'
                ? $t('role.dialog.scopeGlobal')
                : $t('role.dialog.scopeOrg')
            }}
          </div>
        </div>
        <div class="detail-item">
          <div class="detail-label">{{ $t('role.detail.description') }}</div>
          <div class="detail-value">{{ role.remark || '--' }}</div>
        </div>
        <div class="detail-row">
          <div>
            <div class="detail-label">{{ $t('role.detail.status') }}</div>
            <div class="role-status">
              <el-tag :type="role.status ? 'success' : 'danger'" size="small">
                {{
                  role.status
                    ? $t('role.detail.enabled')
                    : $t('role.detail.disabled')
                }}
              </el-tag>
            </div>
          </div>
          <div>
            <div class="detail-label">{{ $t('role.detail.userCount') }}</div>
            <div class="detail-value">
              {{ $t('role.card.userCount', { count: role.userCount || 0 }) }}
            </div>
          </div>
        </div>

        <div class="detail-perm">
          <div class="detail-perm-title">
            {{ $t('role.detail.permissions') }}
          </div>
          <div class="detail-value">
            <template v-if="role.permList && role.permList.length">
              <el-tag
                v-for="(perm, idx) in role.permList"
                :key="idx"
                size="mini"
                :color="getPermColor(idx).backgroundColor"
                :style="{ color: getPermColor(idx).color, border: 'none' }"
                class="perm-tag"
              >
                {{ perm.name }}
              </el-tag>
            </template>
            <span v-else class="text-muted">--</span>
          </div>
        </div>

        <div class="detail-users list-common">
          <div class="detail-perm-title">
            {{ $t('role.detail.associatedUsers') }}
            <el-input
              v-model="userSearch"
              :placeholder="$t('role.detail.searchUserPlaceholder')"
              prefix-icon="el-icon-search"
              size="small"
              clearable
              class="user-search-input no-border-input"
              @keyup.enter.native="searchData"
              @clear="searchData()"
            />
          </div>
          <el-table
            :data="userList"
            style="width: 100%"
            size="small"
            class="user-table"
            v-loading="loading"
            max-height="280px"
          >
            <el-table-column
              :label="$t('role.detail.userName')"
              prop="userName"
            >
              <template slot-scope="scope">
                <div class="user-cell">
                  <div class="user-cell__avatar">
                    <img
                      v-if="scope.row.avatar && scope.row.avatar.path"
                      :src="avatarSrc(scope.row.avatar.path)"
                      alt=""
                    />
                  </div>
                  <div>
                    {{ scope.row.userName }}
                  </div>
                </div>
              </template>
            </el-table-column>
            <el-table-column
              :label="$t('role.detail.org')"
              prop="orgs"
              width="150"
            >
              <template slot-scope="scope">
                <template v-if="scope.row.orgs?.length > 0">
                  <el-tooltip
                    :content="scope.row.orgs.map(o => o.name).join('、')"
                  >
                    <div class="table-org-name">
                      {{ scope.row.orgs.map(o => o.name).join('、') }}
                    </div>
                  </el-tooltip>
                </template>
                <template v-else>--</template>
              </template>
            </el-table-column>
            <el-table-column :label="$t('role.detail.phone')" prop="phone">
              <template slot-scope="scope">
                {{ scope.row.phone || '--' }}
              </template>
            </el-table-column>
            <el-table-column :label="$t('role.detail.email')" prop="email">
              <template slot-scope="scope">
                {{ scope.row.email || '--' }}
              </template>
            </el-table-column>
            <el-table-column
              :label="$t('common.table.operation')"
              width="80"
              align="center"
            >
              <template #default="{ row }">
                <el-button
                  type="text"
                  size="mini"
                  class="remove-text-btn"
                  @click="removeUser(row)"
                >
                  {{ $t('role.detail.removeUser') }}
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

    <span slot="footer" class="dialog-footer">
      <el-button size="small" @click="handleClose">
        {{ $t('common.button.back') }}
      </el-button>
      <el-button
        v-if="!role.isAdmin && (!role.isGlobal || (role.isGlobal && isSystem))"
        size="small"
        type="primary"
        @click="preUpdate(role)"
      >
        {{ $t('common.button.edit') }}
      </el-button>
    </span>
  </el-dialog>
</template>

<script>
import { avatarSrc } from '@/utils/util';
import { OrgTagColorList } from '@/utils/commonSet';
import { removeRoleUser, fetchRoleUsers } from '@/api/permission/role';
import Pagination from '@/components/pagination.vue';

export default {
  name: 'RoleDetailDialog',
  components: { Pagination },
  props: {
    isSystem: {
      type: Boolean,
      default: false,
    },
    preUpdate: {
      type: Function,
      default: () => {},
    },
    orgId: {
      type: String,
      default: '',
    },
  },
  data() {
    return {
      defaultAvatar: '/v1/static/icon/role-default-icon.png',
      listApi: fetchRoleUsers,
      visible: false,
      loading: false,
      role: {},
      userSearch: '',
      userList: [],
    };
  },
  methods: {
    avatarSrc,
    async getTableData(params) {
      if (!this.$refs.pagination) return;
      const searchInfo = {
        roleId: this.role.roleId,
        orgId: this.orgId,
        name: this.userSearch,
        ...params,
      };
      this.loading = true;
      try {
        this.tableData = await this.$refs.pagination.getTableData(searchInfo);
      } finally {
        this.loading = false;
      }
    },
    searchData() {
      this.getTableData({ pageNo: 1 });
    },
    openDialog(row) {
      this.visible = true;
      this.role = row || {};
      this.$nextTick(() => {
        this.searchData();
      });
    },
    getPermColor(index) {
      return OrgTagColorList[index % OrgTagColorList.length];
    },
    refreshData(data) {
      this.userList = data;
    },
    handleClose() {
      this.visible = false;
      this.userSearch = '';
    },
    removeUser(row) {
      removeRoleUser({
        orgId: this.orgId,
        roleId: this.role.roleId,
        userId: row.userId,
      }).then(res => {
        if (res.code === 0) {
          this.$message.success(this.$t('common.message.success'));
          this.getTableData();
        }
      });
    },
  },
};
</script>

<style lang="scss" scoped>
.role-detail {
  margin-top: -35px;
  margin-bottom: -20px;

  .detail-header {
    display: flex;
    align-items: center;
    margin-bottom: 15px;

    .header-left {
      display: flex;
      align-items: center;
    }

    .avatar {
      width: 48px;
      height: 48px;
      border-radius: 8px;
      overflow: hidden;
      flex-shrink: 0;
      background: #f0f2f5;

      img {
        width: 100%;
        height: 100%;
        object-fit: cover;
      }
    }

    .header-info {
      display: flex;
      flex-direction: column;
      margin-left: 14px;

      .role-name {
        font-size: 16px;
        font-weight: 600;
        color: #303133;
        margin-bottom: 6px;
      }

      .status-tag {
        display: inline-block;
        font-size: 12px;
        padding: 1px 8px;
        border-radius: 3px;

        &.enabled {
          color: #67c23a;
          background: #f0f9eb;
        }

        &.disabled {
          color: #f56c6c;
          background: #fef0f0;
        }
      }
    }
  }

  .detail-body {
    .detail-row {
      display: flex;
      padding: 8px 0;
      line-height: 1.5;

      & > div {
        width: 50%;
      }
    }

    .detail-label {
      flex-shrink: 0;
      width: 90px;
      color: #909399;
      font-size: 13px;
      padding-right: 12px;
      margin-bottom: 5px;
    }

    .role-status ::v-deep {
      .el-tag {
        border-radius: 12px;
      }
    }

    .detail-value {
      flex: 1;
      color: #303133;
      font-size: 13px;
      word-break: break-all;

      .el-icon-user {
        margin-right: 4px;
        color: #909399;
      }
    }

    .detail-item {
      padding: 8px 0;
    }

    .detail-users {
      margin-top: 14px;
    }

    .detail-perm {
      margin-top: 10px;
      padding-top: 16px;
      padding-bottom: 16px;
      border-bottom: 0.5px solid #f1f5f9;
      border-top: 0.5px solid #f1f5f9;
    }

    .detail-perm-title {
      color: $color_title;
      font-size: 14px;
      font-weight: 500;
      margin-bottom: 10px;
    }
  }

  .perm-tag {
    margin-right: 6px;
    margin-bottom: 4px;
  }

  .user-search-input {
    margin-left: 15px;
    width: 300px;
  }

  .user-table {
    margin-bottom: 10px;

    .remove-text-btn {
      color: #1d4dd7;
      padding: 0;

      &:hover {
        color: #409eff;
      }
    }

    .user-cell {
      display: flex;
      align-items: center;
    }

    .user-cell__avatar {
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
  }

  .user-pagination {
    display: flex;
    justify-content: flex-end;
  }

  .text-muted {
    color: #c0c4cc;
  }

  .table-org-name {
    max-width: 140px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}
</style>
