<template>
  <div>
    <div class="table-wrap list-common wrap-fullheight">
      <div class="table-box">
        <div class="table-box-header">
          <search-input
            style="margin-right: 2px; margin-bottom: 20px"
            :placeholder="$t('oauth.name')"
            ref="searchInput"
            @handleSearch="getTableData"
          />
          <el-button
            size="mini"
            type="primary"
            @click="preInsert"
            icon="el-icon-plus"
            class="add-btn"
          >
            {{ $t('common.button.create') }}
          </el-button>
        </div>
        <el-table
          :data="tableData"
          :header-cell-style="{ background: '#F9F9F9', color: '#999999' }"
          v-loading="loading"
          style="width: 100%"
        >
          <el-table-column prop="name" :label="$t('oauth.name')" align="left" />
          <el-table-column prop="desc" :label="$t('oauth.desc')" align="left" />
          <el-table-column prop="clientId" label="Client Id" align="left" />
          <el-table-column
            prop="clientSecret"
            label="Client Secret"
            align="left"
          />
          <el-table-column
            prop="redirectUri"
            label="Redirect Uri"
            align="left"
          />
          <el-table-column align="left" :label="$t('common.table.operation')">
            <template slot-scope="scope">
              <el-switch
                @change="val => changeStatus(scope.row, val)"
                v-model="scope.row.status"
                :active-text="$t('common.switch.start')"
                :inactive-text="$t('common.switch.stop')"
              />
            </template>
          </el-table-column>
          <el-table-column
            align="left"
            :label="$t('common.table.operation')"
            width="180"
          >
            <template slot-scope="scope">
              <el-button size="mini" type="text" @click="preUpdate(scope.row)">
                {{ $t('common.button.edit') }}
              </el-button>
              <el-button size="mini" type="text" @click="preDel(scope.row)">
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

    <!-- 创建/编辑弹窗 -->
    <el-dialog
      :title="isEdit ? $t('oauth.edit') : $t('oauth.create')"
      :visible.sync="dialogVisible"
      width="600px"
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
        <el-form-item :label="$t('oauth.name')" prop="name">
          <el-input
            v-model="form.name"
            :placeholder="$t('common.input.placeholder')"
            clearable
          />
        </el-form-item>
        <el-form-item :label="$t('oauth.desc')" prop="desc">
          <el-input
            type="textarea"
            :rows="3"
            v-model="form.desc"
            :placeholder="$t('common.input.placeholder')"
            clearable
          />
        </el-form-item>
        <el-form-item label="Redirect Uri" prop="redirectUri">
          <el-input
            v-model="form.redirectUri"
            :placeholder="$t('common.input.placeholder')"
            clearable
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
  </div>
</template>

<script>
import Pagination from '@/components/pagination.vue';
import SearchInput from '@/components/searchInput.vue';
import {
  fetchOAuthList,
  createOAuth,
  editOAuth,
  deleteOAuth,
  changeOAuthStatus,
} from '@/api/permission/oauth';
import { deleteUser } from '@/api/permission/user';

export default {
  components: { Pagination, SearchInput },
  data() {
    return {
      listApi: fetchOAuthList,
      loading: false,
      dialogVisible: false,
      submitLoading: false,
      isEdit: false,
      form: {
        name: '',
        desc: '',
        clientId: '',
        redirectUri: '',
        status: true,
      },
      rules: {
        name: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
        ],
        desc: [
          {
            max: 200,
            message: this.$t('common.hint.remarkLimit'),
            trigger: 'blur',
          },
        ],
        redirectUri: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
        ],
      },
      tableData: [],
      row: {},
    };
  },
  mounted() {
    this.getTableData();
  },
  methods: {
    async getTableData() {
      const searchInfo = {
        ...(this.$refs.searchInput.value && {
          name: this.$refs.searchInput.value,
        }),
      };
      this.loading = true;
      try {
        this.tableData = await this.$refs.pagination.getTableData(searchInfo);
      } finally {
        this.loading = false;
      }
    },
    refreshData(data) {
      this.tableData = data;
    },
    preInsert() {
      this.isEdit = false;
      this.form = {
        name: '',
        desc: '',
        clientId: '',
        redirectUri: '',
        status: true,
      };
      this.dialogVisible = true;
    },
    preUpdate(row) {
      this.isEdit = true;
      this.row = row;
      this.form = { ...row };
      this.dialogVisible = true;
    },
    preDel(row) {
      this.$confirm(
        this.$t('oauth.deleteHint'),
        this.$t('common.confirm.title'),
        {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          type: 'warning',
        },
      ).then(() => {
        return deleteOAuth({ clientId: row.clientId }).then(res => {
          if (res.code === 0) {
            this.$message.success(this.$t('common.message.success'));
            this.getTableData();
          }
        });
      });
    },
    handleClose() {
      this.$refs.form.resetFields();
      this.dialogVisible = false;
    },
    changeStatus(row, val) {
      changeOAuthStatus({ clientId: row.clientId, status: val }).then(res => {
        if (res.code === 0) {
          this.$message.success(this.$t('common.message.success'));
          this.getTableData();
        }
      });
    },
    handleSubmit() {
      this.$refs.form.validate(async valid => {
        if (!valid) return;
        this.submitLoading = true;
        try {
          const params = { ...this.form };
          if (this.isEdit) {
            await editOAuth(params);
          } else {
            await createOAuth(params);
          }
          this.$message.success(this.$t('common.message.success'));
          this.dialogVisible = false;
          await this.getTableData();
        } finally {
          this.submitLoading = false;
        }
      });
    },
  },
};
</script>

<style lang="scss" scoped>
.table-box {
  max-height: calc(100vh - 160px);
  overflow-y: auto;
  padding: 5px 10px 10px 20px;

  .table-box-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .add-btn {
    margin-left: 20px;
    margin-bottom: 20px;
  }

  ::v-deep .el-switch__label * {
    font-size: 13px;
  }
}

::v-deep .operation.el-button--text.el-button {
  padding: 3px 10px 3px 0;
  border-right: 1px solid #eaeaea !important;
}
</style>
