<template>
  <el-dialog
    :visible.sync="dialogVisible"
    :title="$t('channel.editDialog.title')"
    width="640px"
    append-to-body
    :close-on-click-modal="false"
    :before-close="handleClose"
    custom-class="channel-edit-dialog"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-position="top"
      size="small"
      style="margin-top: -20px"
    >
      <!-- 渠道类型 + 重新扫码绑定 -->
      <el-form-item :label="$t('channel.table.channelType')">
        <div class="channel-type-row">
          <el-input :value="channelTypeLabel" disabled></el-input>
          <el-button
            style="margin-left: 10px"
            type="primary"
            size="small"
            @click="handleReScan"
          >
            {{ $t('channel.reScanBind') }}
          </el-button>
        </div>
      </el-form-item>

      <el-form-item :label="$t('channel.channelName')" prop="name">
        <el-input v-model="form.name" maxlength="50" />
      </el-form-item>

      <el-form-item :label="$t('channel.appType')" prop="appType">
        <el-select
          v-model="form.appType"
          style="width: 100%"
          @change="handleAppTypeChange"
          disabled
        >
          <el-option
            v-for="item in appTypeOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>

      <el-form-item
        v-if="isBindAppType"
        :label="$t('channel.bindApp')"
        prop="appId"
      >
        <el-select
          v-model="form.appId"
          :placeholder="$t('channel.bindAppPlaceholder')"
          filterable
          style="width: 100%"
        >
          <el-option
            v-for="item in appList"
            :key="item.appId"
            :label="item.name"
            :value="item.appId"
          />
        </el-select>
      </el-form-item>

      <el-form-item
        v-if="isModelType"
        :label="$t('channel.bindModel')"
        prop="modelUuid"
      >
        <el-select
          v-model="form.modelUuid"
          :placeholder="$t('common.select.placeholder')"
          filterable
          style="width: 100%"
        >
          <el-option
            v-for="item in modelList"
            :key="item.uuid"
            :label="item.displayName"
            :value="item.uuid"
          />
        </el-select>
      </el-form-item>

      <el-form-item
        v-if="form.appType === GENERAL_AGENT"
        :label="$t('channel.bindScene')"
        prop="agentId"
      >
        <el-select
          v-model="form.agentId"
          :placeholder="$t('common.select.placeholder')"
          filterable
          style="width: 100%"
        >
          <el-option
            v-for="item in sceneList"
            :key="item.agentId"
            :label="item.agentName"
            :value="item.agentId"
          />
        </el-select>
      </el-form-item>

      <el-form-item
        v-if="form.appType === DIGITAL_EMPLOYEE"
        :label="$t('channel.bindDigitalEmployee')"
        prop="agentId"
      >
        <el-select
          v-model="form.agentId"
          :placeholder="$t('common.select.placeholder')"
          filterable
          style="width: 100%"
        >
          <el-option
            v-for="item in employeeList"
            :key="item.id"
            :label="item.name"
            :value="item.id"
          />
        </el-select>
      </el-form-item>

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
    </el-form>

    <template #footer>
      <el-button type="primary" @click="handleSave" :loading="saveLoading">
        {{ $t('channel.saveConfig') }}
      </el-button>
      <el-button @click="handleClose">
        {{ $t('common.button.cancel') }}
      </el-button>
    </template>

    <!-- 扫码绑定弹窗 -->
    <scan-dialog ref="scanDialogRef" @success="onScanSuccess" />
  </el-dialog>
</template>

<script>
import { AGENT } from '@/utils/commonSet';
import {
  editChannel,
  getApiSelect,
  getAppSelect,
  getModelSelect,
  getEmployeeSelect,
  getSceneSelect,
} from '@/api/channel';
import {
  WECHAT,
  DING_TALK,
  GENERAL_AGENT,
  DIGITAL_EMPLOYEE,
  APP_TYPE_OPTIONS,
} from '../constants';
import ScanDialog from './scanDialog.vue';

export default {
  name: 'ChannelEditDialog',
  components: {
    ScanDialog,
  },
  data() {
    return {
      GENERAL_AGENT,
      DIGITAL_EMPLOYEE,
      dialogVisible: false,
      saveLoading: false,
      form: {
        name: '',
        appType: AGENT,
        appId: '',
        modelUuid: '',
        agentId: '',
        apiKeyId: '',
        config: {},
      },
      rules: {
        name: [
          {
            pattern: this.$config.commonTextReg,
            message: this.$t('common.hint.text'),
            trigger: 'blur',
          },
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
          {
            min: 2,
            max: 50,
            message: this.$t('common.hint.textLimit'),
            trigger: 'blur',
          },
        ],
        appId: [
          {
            validator: (rule, value, callback) => {
              if (this.form.appType !== AGENT) return callback();
              if (!value) {
                return callback(
                  new Error(this.$t('common.select.placeholder')),
                );
              }
              callback();
            },
            trigger: 'change',
          },
        ],
        appType: [
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
      appTypeOptions: APP_TYPE_OPTIONS,
      appList: [],
      modelList: [],
      sceneList: [],
      employeeList: [],
      apiKeyList: [],
      row: {},
    };
  },
  computed: {
    channelTypeLabel() {
      if (this.row.channelType === WECHAT) return this.$t('channel.wechat');
      if (this.row.channelType === DING_TALK)
        return this.$t('channel.dingtalk');
      return this.row.channelType;
    },
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
  methods: {
    setFormValue(row) {
      const obj = { ...this.form };
      for (let key in obj) {
        obj[key] =
          row && row[key] ? row[key] : Array.isArray(obj[key]) ? [] : '';
      }
      this.form = obj;
    },
    fetchInit() {
      const value = this.form.appType;
      if (value === AGENT) {
        this.fetchAppList();
      } else if (value === GENERAL_AGENT || value === DIGITAL_EMPLOYEE) {
        this.fetchModelList();
      }
      if (value === GENERAL_AGENT) {
        this.fetchSceneList();
      }
      if (value === DIGITAL_EMPLOYEE) {
        this.fetchEmployeeList();
      }
      this.fetchApiKeyList();
    },
    open(row) {
      this.row = row;
      this.setFormValue({ ...row, appType: row.appType || AGENT });
      this.fetchInit();
      this.dialogVisible = true;
      this.$nextTick(() => {
        this.$refs.formRef && this.$refs.formRef.clearValidate();
      });
    },
    handleClose() {
      this.dialogVisible = false;
      this.$refs.formRef && this.$refs.formRef.resetFields();
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
      this.$refs.formRef.validate(valid => {
        if (!valid) return;
        this.saveLoading = true;
        editChannel(this.row.id, this.formatValue())
          .then(() => {
            this.saveLoading = false;
            this.$message.success(this.$t('common.message.success'));
            this.handleClose();
            this.$emit('success');
          })
          .catch(() => {
            this.saveLoading = false;
          });
      });
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
    /** 重新扫码绑定 */
    handleReScan() {
      this.$refs.scanDialogRef.open(this.row.channelType);
    },
    /** 扫码成功回调，更新渠道配置 */
    onScanSuccess(data) {
      this.form.config = data.credentials || {};
    },
  },
};
</script>

<style lang="scss" scoped>
.channel-edit-dialog {
  border-radius: 8px;
  .el-dialog__body {
    padding: 16px 24px 0;
  }
  .el-dialog__footer {
    padding: 12px 24px 20px;
    text-align: right;
  }
}
.channel-type-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}
</style>
