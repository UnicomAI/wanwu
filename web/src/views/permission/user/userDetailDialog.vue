<template>
  <el-dialog
    :title="$t('user.detail.title')"
    :visible.sync="visible"
    width="600px"
    append-to-body
    :close-on-click-modal="false"
    :before-close="handleClose"
  >
    <div v-if="user" class="user-detail">
      <!-- 第一部分：头像、名称、邮箱、账号状态，居中显示 -->
      <div class="user-detail__header">
        <div v-if="user.avatar?.path" class="user-detail__avatar">
          <img :src="avatarSrc(user.avatar.path)" alt="" />
        </div>
        <div class="user-detail__name">{{ user.username || '--' }}</div>
        <div class="user-detail__email">{{ user.email || '--' }}</div>
        <div class="user-detail__status">
          <el-tag :type="user.status ? 'success' : 'danger'" size="small">
            {{ user.status ? $t('user.detail.start') : $t('user.detail.stop') }}
          </el-tag>
        </div>
      </div>

      <!-- 第二部分：所属组织、手机号、创建时间，左右结构 -->
      <div class="user-detail__info">
        <div class="info-item">
          <span class="info-label">{{ $t('user.detail.phone') }}</span>
          <span class="info-value">{{ user.phone || '--' }}</span>
        </div>

        <div class="info-item">
          <span class="info-label">{{ $t('user.detail.createAt') }}</span>
          <span class="info-value">{{ user.createdAt || '--' }}</span>
        </div>
      </div>

      <!-- 第三部分：关联用户（组织 - 角色表） -->
      <div class="user-detail__roles">
        <el-table
          :data="relatedRows"
          :show-header="true"
          class="relation-table"
          size="small"
          max-height="228"
        >
          <el-table-column
            :label="$t('user.detail.org')"
            prop="orgName"
            align="center"
            min-width="60%"
          />
          <el-table-column
            :label="$t('user.detail.role')"
            prop="roleNames"
            align="center"
          >
            <template slot-scope="scope">
              <template>
                {{ scope.row.roleNames || '--' }}
              </template>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 第四部分：按钮，左侧重置密码，右侧禁用/启用账号 -->
      <div class="user-detail__actions">
        <el-button @click="handleResetPassword">
          {{ $t('user.detail.resetPassword') }}
        </el-button>
        <el-button
          :type="user.status ? 'danger' : 'primary'"
          @click="handleToggleStatus"
          plain
        >
          {{
            user.status
              ? $t('user.detail.disableAccount')
              : $t('user.detail.enableAccount')
          }}
        </el-button>
      </div>
    </div>
  </el-dialog>
</template>

<script>
import { avatarSrc } from '@/utils/util';

export default {
  name: 'UserDetailDialog',
  data() {
    return {
      visible: false,
      user: {},
    };
  },
  computed: {
    relatedRows() {
      if (!this.user.orgs || !this.user.orgs.length) return [];
      return this.user.orgs.map(item => ({
        orgName: item.org ? item.org.name : '--',
        roleNames:
          Array.isArray(item.roles) && item.roles.length
            ? item.roles.map(r => r.name).join(',')
            : '',
      }));
    },
  },
  methods: {
    avatarSrc,
    openDialog(row) {
      this.visible = true;
      this.user = row || {};
    },
    handleClose() {
      this.visible = false;
    },
    handleResetPassword() {
      this.$emit('resetPassword', this.user);
    },
    handleToggleStatus() {
      this.$emit('toggleStatus', this.user);
    },
  },
};
</script>

<style lang="scss" scoped>
.user-detail {
  margin-top: -20px;

  // 第一部分：居中头像信息
  &__header {
    text-align: center;
    margin-bottom: 30px;
  }
  &__avatar {
    width: 64px;
    height: 64px;
    line-height: 64px;
    border-radius: 50%;
    margin: 0 auto 12px;
    object-fit: cover;
    background: #fff;
    box-shadow: 0 1px 4px 0 rgba(0, 0, 0, 0.15);
    img {
      width: 100%;
      height: 100%;
      border-radius: 50%;
    }
  }
  &__name {
    font-size: 18px;
    font-weight: 600;
    color: #303133;
    margin-bottom: 6px;
  }
  &__email {
    font-size: 14px;
    color: #909399;
    margin-bottom: 10px;
  }
  &__status ::v-deep {
    .el-tag {
      border-radius: 12px;
    }
  }

  // 第二部分：左右结构信息
  &__info {
    margin-bottom: 24px;

    .info-item {
      display: flex;
      align-items: center;
      padding: 10px 0;
      min-width: 0;
      &:last-child {
        border-bottom: 0.5px solid #f1f5f9;
        padding-bottom: 20px;
      }
      .info-label {
        flex-shrink: 0;
        width: 80px;
        color: #909399;
        font-size: 14px;
      }
      .info-value {
        flex: 1;
        color: #303133;
        font-size: 14px;
        word-break: break-all;
        margin-left: 16px;
        text-align: right;

        &--org {
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
          word-break: normal;
        }
      }
    }
  }

  // 第三部分：关联用户（组织 - 角色表）
  &__roles {
    margin-bottom: 24px;
    .section-title {
      font-size: 14px;
      color: $color_title;
      margin-bottom: 10px;
      font-weight: 500;
    }
    .relation-table {
      width: 100%;
      ::v-deep th.el-table__cell {
        background: #f9f9f9;
        color: #999999;
        font-weight: normal;
      }
    }
    .role-tag {
      border-radius: 4px;
      color: $tag_color !important;
      background-color: $tag_bg !important;
      margin-right: 6px;
      &:last-child {
        margin-right: 0;
      }
    }
    .no-roles {
      color: #c0c4cc;
      font-size: 14px;
    }
  }

  // 第四部分：操作按钮
  &__actions {
    display: flex;
    justify-content: space-between;
    padding-top: 16px;
    .el-button {
      width: 50%;
    }
  }
}
</style>
