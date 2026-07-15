<template>
  <div class="channel-config">
    <!-- 添加渠道账号表单 -->
    <div class="form-section bind-wrapper">
      <div class="section-title">{{ $t('channel.addAccountTitle') }}</div>
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        size="small"
      >
        <div style="width: 60%">
          <!-- 选择渠道类型 -->
          <el-form-item :label="$t('channel.channelType')" prop="channelType">
            <div class="channel-type-list">
              <div
                v-for="item in channelTypeOptions"
                :key="item.value"
                :class="[
                  'channel-type-item',
                  { active: form.channelType === item.value },
                ]"
                @click="form.channelType = item.value"
              >
                <div class="channel-icon">
                  <img :src="item.img" alt="" />
                </div>
                <span>{{ item.label }}</span>
              </div>
            </div>

            <!-- 微信扫码连接区域 -->
            <div class="qr-connect-area">
              <el-button
                type="primary"
                plain
                round
                size="medium"
                @click="handleScanConnect"
              >
                <i class="iconfont icon-scan"></i>
                {{ $t('channel.scanConnect')
                }}{{
                  form.channelType === WECHAT
                    ? $t('channel.wechat')
                    : $t('channel.dingtalk')
                }}
              </el-button>
              <p class="qr-hint">
                {{
                  form.channelType === WECHAT
                    ? $t('channel.scanHint')
                    : $t('channel.dingtalkHint')
                }}
              </p>
            </div>
          </el-form-item>

          <!-- 渠道名称 -->
          <el-form-item :label="$t('channel.channelName')" prop="name">
            <el-input
              v-model="form.name"
              :placeholder="$t('common.hint.text')"
              maxlength="50"
              show-word-limit
            />
            <div class="field-hint">{{ $t('channel.nameHint') }}</div>
          </el-form-item>

          <!-- 应用类型 -->
          <el-form-item :label="$t('channel.appType')" prop="appType">
            <el-select
              v-model="form.appType"
              :placeholder="$t('common.select.placeholder')"
              style="width: 100%"
              @change="handleAppTypeChange"
            >
              <el-option
                v-for="item in appTypeOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </el-form-item>

          <!-- 关联应用 -->
          <el-form-item
            v-if="isBindAppType"
            :label="$t('channel.bindApp')"
            prop="appId"
          >
            <el-select
              v-model="form.appId"
              :placeholder="$t('channel.bindAppPlaceholder')"
              style="width: 100%"
              filterable
            >
              <el-option
                v-for="item in appList"
                :key="item.appId"
                :label="item.name"
                :value="item.appId"
              />
            </el-select>
          </el-form-item>

          <!-- 关联模型 -->
          <el-form-item
            v-if="isModelType"
            :label="$t('channel.bindModel')"
            prop="modelUuid"
          >
            <el-select
              v-model="form.modelUuid"
              :placeholder="$t('common.select.placeholder')"
              style="width: 100%"
              filterable
            >
              <el-option
                v-for="item in modelList"
                :key="item.uuid"
                :label="item.displayName"
                :value="item.uuid"
              />
            </el-select>
          </el-form-item>

          <!-- 关联场景 -->
          <el-form-item
            v-if="form.appType === GENERAL_AGENT"
            :label="$t('channel.bindScene')"
            prop="agentId"
          >
            <el-select
              v-model="form.agentId"
              :placeholder="$t('common.select.placeholder')"
              style="width: 100%"
              filterable
            >
              <el-option
                v-for="item in sceneList"
                :key="item.agentId"
                :label="item.agentName"
                :value="item.agentId"
              />
            </el-select>
          </el-form-item>

          <!-- 关联数字员工 -->
          <el-form-item
            v-if="form.appType === DIGITAL_EMPLOYEE"
            :label="$t('channel.bindDigitalEmployee')"
            prop="agentId"
          >
            <el-select
              v-model="form.agentId"
              :placeholder="$t('common.select.placeholder')"
              style="width: 100%"
              filterable
            >
              <el-option
                v-for="item in employeeList"
                :key="item.id"
                :label="item.name"
                :value="item.id"
              />
            </el-select>
          </el-form-item>

          <!-- 关联API Key -->
          <el-form-item :label="$t('channel.bindApiKey')" prop="apiKeyId">
            <el-select
              v-model="form.apiKeyId"
              :placeholder="$t('common.select.placeholder')"
              filterable
              style="width: 100%"
            >
              <el-option
                v-for="item in apiKeyList"
                :key="item.keyId"
                :label="item.name"
                :value="item.keyId"
              />
            </el-select>
          </el-form-item>

          <!-- 渠道配置展示 -->
          <div v-if="justifyConfig()">
            <el-form-item
              v-for="item in Object.keys(form.config)"
              :label="item"
              :key="item"
              :prop="item"
            >
              <el-input v-model="form.config[item]" placeholder="" disabled />
            </el-form-item>
          </div>
        </div>
        <!-- 按钮 -->
        <el-form-item class="btn-row">
          <el-button type="primary" @click="handleSave" :loading="saveLoading">
            {{ $t('channel.saveConfig') }}
          </el-button>
          <el-button @click="handleReset">
            {{ $t('common.button.cancel') }}
          </el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 已配置渠道列表 -->
    <div class="form-section table-section table-wrap list-common">
      <div class="table-header">
        <div class="section-title">{{ $t('channel.configuredList') }}</div>
        <el-input
          v-model="searchName"
          :placeholder="$t('channel.searchPlaceholder')"
          prefix-icon="el-icon-search"
          clearable
          size="small"
          @keyup.enter.native="searchData"
          @clear="searchData"
          style="width: 220px"
        />
      </div>
      <el-table
        :data="tableData"
        :header-cell-style="{ background: '#F9F9F9', color: '#999999' }"
        v-loading="loading"
        style="width: 100%"
      >
        <!-- 渠道 -->
        <el-table-column
          :label="$t('channel.table.channel')"
          align="left"
          width="140"
        >
          <template slot-scope="scope">
            <div class="channel-cell">
              <div class="channel-icon-mini">
                <img
                  :src="getCurrentChannel(scope.row.channelType).img"
                  alt=""
                />
              </div>
              <span>
                {{ getCurrentChannel(scope.row.channelType).label }}
              </span>
            </div>
          </template>
        </el-table-column>
        <!-- 渠道名称 -->
        <el-table-column
          prop="name"
          :label="$t('channel.table.name')"
          align="left"
          width="150"
        />
        <!-- 应用类型 -->
        <el-table-column
          prop="appType"
          :label="$t('channel.table.appType')"
          align="left"
        >
          <template slot-scope="scope">
            <div>
              {{ getChannelAppTypeLabel(scope.row.appType) }}
            </div>
          </template>
        </el-table-column>
        <!-- 应用名称 -->
        <el-table-column
          prop="appName"
          :label="$t('channel.table.appName')"
          align="left"
        >
          <template slot-scope="scope">
            {{ scope.row.appName || '--' }}
          </template>
        </el-table-column>
        <!-- 创建时间 -->
        <el-table-column
          prop="createdAt"
          :label="$t('channel.table.createdAt')"
          align="left"
        />
        <!-- 开关 -->
        <el-table-column
          :label="$t('channel.table.switch')"
          align="left"
          width="140"
        >
          <template slot-scope="scope">
            <el-switch
              @change="val => handleChangeStatus(scope.row, val)"
              v-model="scope.row.enabled"
            />
          </template>
        </el-table-column>
        <!-- 状态 -->
        <el-table-column
          :label="$t('channel.table.status')"
          align="left"
          width="100"
        >
          <template slot-scope="scope">
            <el-tag
              v-for="(tag, idx) in parseStatusTags(scope.row)"
              :key="idx"
              size="mini"
              :type="tag.type"
              style="margin-right: 4px; margin-bottom: 2px"
            >
              {{ tag.text }}
            </el-tag>
          </template>
        </el-table-column>
        <!-- 操作 -->
        <el-table-column
          :label="$t('common.table.operation')"
          align="left"
          width="120"
        >
          <template slot-scope="scope">
            <el-button size="mini" type="text" @click="handleEdit(scope.row)">
              {{ $t('common.button.edit') }}
            </el-button>
            <!--隐藏，目前无需断开，可直接删除时断开-->
            <!--<el-button
              size="mini"
              type="text"
              @click="handleDisconnect(scope.row)"
            >
              {{ $t('channel.disconnect') }}
            </el-button>-->
            <el-button
              size="mini"
              type="text"
              class="danger-btn"
              @click="handleDelete(scope.row)"
            >
              {{ $t('common.button.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <Pagination
        class="pagination"
        ref="pagination"
        :listApi="listApi"
        @refreshData="refreshData"
      />
    </div>
    <EditDialog ref="editDialog" @success="fetchTableData" />
    <ScanDialog ref="scanDialog" @success="onScanSuccess" />
  </div>
</template>

<script>
import Pagination from '@/components/pagination.vue';
import EditDialog from './components/editDialog.vue';
import ScanDialog from './components/scanDialog.vue';
import {
  deleteChannel,
  fetchChannelList,
  changeChannelStatus,
  disconnectChannel,
  createChannel,
  getApiSelect,
  getAppSelect,
  getModelSelect,
  getSceneSelect,
  getEmployeeSelect,
} from '@/api/channel';
import { AGENT, AppType } from '@/utils/commonSet';
import {
  WECHAT,
  DING_TALK,
  GENERAL_AGENT,
  DIGITAL_EMPLOYEE,
  APP_TYPE_OPTIONS,
} from './constants';

export default {
  name: 'ChannelConfig',
  components: { Pagination, EditDialog, ScanDialog },
  data() {
    return {
      WECHAT,
      GENERAL_AGENT,
      DIGITAL_EMPLOYEE,
      AppType,
      listApi: fetchChannelList,
      loading: false,
      saveLoading: false,
      searchName: '',
      form: {
        channelType: WECHAT,
        name: '',
        appType: AGENT,
        appId: '',
        modelUuid: '',
        agentId: '',
        apiKeyId: '',
        config: {},
      },
      rules: {
        channelType: [
          {
            required: true,
            message: this.$t('common.select.placeholder'),
            trigger: 'change',
          },
        ],
        name: [
          {
            pattern: this.$config.commonTextReg,
            message: this.$t('common.hint.text'),
            trigger: 'blur',
          },
          {
            min: 2,
            max: 50,
            message: this.$t('common.hint.textLimit'),
            trigger: 'blur',
          },
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
        ],
        appType: [
          {
            required: true,
            message: this.$t('common.select.placeholder'),
            trigger: 'change',
          },
        ],
        appId: [
          {
            required: true,
            message: this.$t('common.select.placeholder'),
            trigger: 'change',
          },
        ],
        modelUuid: [
          {
            required: true,
            message: this.$t('common.select.placeholder'),
            trigger: 'change',
          },
        ],
        agentId: [
          {
            required: true,
            message: this.$t('common.select.placeholder'),
            trigger: 'change',
          },
        ],
        apiKeyId: [
          {
            required: true,
            message: this.$t('common.select.placeholder'),
            trigger: 'change',
          },
        ],
      },
      channelTypeOptions: [
        {
          value: WECHAT,
          label: this.$t('channel.wechat'),
          img: require('@/assets/imgs/wechat.png'),
        },
        {
          value: DING_TALK,
          label: this.$t('channel.dingtalk'),
          img: require('@/assets/imgs/dingtalk.png'),
        },
      ],
      appTypeOptions: APP_TYPE_OPTIONS,
      appList: [],
      modelList: [],
      sceneList: [],
      employeeList: [],
      apiKeyList: [],
      tableData: [],
    };
  },
  computed: {
    isBindAppType() {
      return this.form.appType === AGENT;
    },
    isModelType() {
      return (
        this.form.appType === GENERAL_AGENT ||
        this.form.appType === DIGITAL_EMPLOYEE
      );
    },
  },
  mounted() {
    this.fetchTableData();
    this.fetchAppList();
    this.fetchApiKeyList();
  },
  methods: {
    refreshData(data) {
      this.tableData = data;
    },
    async fetchTableData(params) {
      const searchInfo = {
        ...(this.searchName && { name: this.searchName }),
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
      this.fetchTableData({ pageNo: 1 });
    },
    handleAppTypeChange(newVal) {
      this.form.appId = '';
      this.form.modelUuid = '';
      this.form.agentId = '';
      if (newVal === AGENT) {
        this.fetchAppList();
      } else if (newVal === GENERAL_AGENT || newVal === DIGITAL_EMPLOYEE) {
        this.fetchModelList();
      }
      if (newVal === GENERAL_AGENT) {
        this.fetchSceneList();
      }
      if (newVal === DIGITAL_EMPLOYEE) {
        this.fetchEmployeeList();
      }
      this.$nextTick(() => {
        this.$refs.formRef.clearValidate(['appId', 'modelUuid', 'agentId']);
      });
    },
    async fetchAppList() {
      const res = await getAppSelect(this.form.appType);
      this.appList = res.data?.list || [];
    },
    async fetchApiKeyList() {
      const res = await getApiSelect();
      this.apiKeyList = res.data?.list || [];
    },
    async fetchModelList() {
      const res = await getModelSelect();
      this.modelList = res.data?.list || [];
    },
    async fetchSceneList() {
      const res = await getSceneSelect();
      this.sceneList = res.data?.wgaAgentList || [];
    },
    async fetchEmployeeList() {
      const res = await getEmployeeSelect();
      this.employeeList = res.data?.list || res.data || [];
    },
    getChannelAppTypeLabel(appType) {
      const option = this.appTypeOptions.find(item => item.value === appType);
      return option?.label || AppType[appType] || '--';
    },
    justifyConfig() {
      return Object.keys(this.form.config || {}).length > 0;
    },
    formatValue() {
      const value = { ...this.form };
      if (value.appType === AGENT) {
        delete value.modelUuid;
        delete value.agentId;
      } else if ([GENERAL_AGENT, DIGITAL_EMPLOYEE].includes(value.appType)) {
        delete value.appId;
      }
      return value;
    },
    handleSave() {
      if (!this.justifyConfig()) {
        this.$message.warning(this.$t('channel.configEmpty'));
        return;
      }
      this.$refs.formRef.validate(valid => {
        if (!valid) return;
        this.saveLoading = true;
        createChannel(this.formatValue())
          .then(() => {
            this.saveLoading = false;
            this.$message.success(this.$t('common.message.success'));
            this.handleReset();
            this.fetchTableData();
          })
          .catch(() => {
            this.saveLoading = false;
          });
      });
    },
    handleReset() {
      this.$refs.formRef && this.$refs.formRef.resetFields();
      this.form = {
        channelType: WECHAT,
        name: '',
        appType: AGENT,
        appId: '',
        modelUuid: '',
        agentId: '',
        apiKeyId: '',
        config: {},
      };
    },
    handleEdit(row) {
      this.$refs.editDialog.open(row);
    },
    handleScanConnect() {
      this.$refs.scanDialog.open(this.form.channelType);
    },
    onScanSuccess(data) {
      this.form.config = data.credentials || {};
    },
    handleDisconnect(row) {
      this.$confirm(
        this.$t('channel.disconnectConfirm'),
        this.$t('common.confirm.title'),
        {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          type: 'warning',
        },
      ).then(() => {
        disconnectChannel(row.id).then(() => {
          this.$message.success(this.$t('common.message.success'));
          this.fetchTableData();
        });
      });
    },
    handleDelete(row) {
      this.$confirm(
        this.$t('channel.deleteConfirm'),
        this.$t('common.confirm.title'),
        {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          type: 'warning',
        },
      ).then(() => {
        deleteChannel(row.id).then(() => {
          this.$message.success(this.$t('common.message.success'));
          this.fetchTableData();
        });
      });
    },
    handleChangeStatus(row, val) {
      changeChannelStatus(row.id, {
        enabled: val,
      }).then(() => {
        this.$message.success(this.$t('common.message.success'));
        this.fetchTableData();
      });
    },
    getCurrentChannel(channelType) {
      const type = channelType || WECHAT;
      return this.channelTypeOptions.find(item => item.value === type) || {};
    },
    parseStatusTags(row) {
      const statusMap = {
        loggedIn: {
          text: this.$t('channel.status.connected'),
          type: 'success',
        },
        waitingLogin: {
          text: this.$t('channel.status.waitingLogin'),
          type: 'warning',
        },
        error: { text: this.$t('channel.status.error'), type: 'danger' },
        offline: {
          text: this.$t('channel.status.disconnected'),
          type: 'info',
        },
      };
      return statusMap[row.status]
        ? [statusMap[row.status]]
        : [{ text: row.status, type: '' }];
    },
  },
};
</script>

<style lang="scss" scoped>
.channel-config {
  padding-bottom: 24px;
  .form-section {
    padding: 24px 20px;
    border: 1px solid #ebeef5;
    border-radius: 8px;

    .section-title {
      font-size: 16px;
      font-weight: 600;
      color: #303133;
      margin-bottom: 16px;
    }
  }

  .bind-wrapper {
    margin-top: 35px;
    margin-bottom: 20px;
  }

  .table-section {
    padding-bottom: 0;
    .table-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;

      .section-title {
        font-size: 16px;
        font-weight: 600;
        color: #303133;
      }
    }
  }

  .channel-type-list {
    display: flex;
    gap: 32px;
  }

  .channel-type-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    cursor: pointer;
    opacity: 0.6;
    transition: opacity 0.2s;

    &:hover,
    &.active {
      opacity: 1;
    }

    &.active .channel-icon {
      box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
    }

    span {
      font-size: 13px;
      color: #606266;
    }
  }

  .channel-icon {
    width: 44px;
    height: 44px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;

    img {
      width: 100%;
      height: 100%;
      border-radius: 50%;
    }
  }

  .qr-connect-area {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 30px 0;
    border: 1px dashed #dcdfe6;
    border-radius: 8px;
    margin-top: 12px;

    .qr-hint {
      font-size: 12px;
      color: #909399;
      margin-top: 10px;
    }
  }

  .field-hint {
    font-size: 12px;
    color: #909399;
    margin-top: 4px;
  }

  .field-tip {
    font-size: 12px;
    color: #f56c6c;
    margin-top: 4px;
  }

  .btn-row {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-top: 8px;
    justify-content: flex-end;

    .warning-text {
      font-size: 13px;
      color: #f56c6c;
      margin-left: 12px;
    }
  }

  .channel-cell {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .channel-icon-mini {
    width: 26px;
    height: 26px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;

    img {
      width: 100%;
      height: 100%;
      border-radius: 50%;
    }
  }

  .pagination-wrap {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 16px;
    padding-bottom: 16px;

    .total-info {
      font-size: 13px;
      color: #909399;
    }
  }

  ::v-deep .danger-btn {
    color: #f56c6c;
  }

  ::v-deep .el-switch__label * {
    font-size: 12px;
  }

  ::v-deep .operation.el-button--text.el-button {
    padding: 3px 8px 3px 0;
    border-right: 1px solid #eaeaea !important;
  }
}
</style>
