<template>
  <div class="createDialog">
    <el-dialog
      :visible.sync="dialogVisible"
      width="760px"
      append-to-body
      :close-on-click-modal="false"
      :before-close="handleClose"
    >
      <template slot="title">
        <div class="dialog-title-wrapper">
          <span class="dialog-title">{{ provider.name || '' }}</span>
          <span class="dialog-desc" v-if="provider.key === yuanjing">
            {{ $t('modelAccess.hint.yuanjing') }}
          </span>
        </div>
      </template>
      <el-form
        :model="{ ...createForm }"
        :rules="rules"
        ref="createForm"
        label-width="130px"
        class="createForm form"
      >
        <el-form-item
          :label="$t('modelAccess.table.modelType')"
          prop="modelType"
        >
          <el-select
            v-model="createForm.modelType"
            :placeholder="$t('common.select.placeholder')"
            :disabled="isEdit"
            style="width: 100%"
          >
            <el-option
              v-for="item in modelType"
              :key="item.key"
              :label="item.name"
              :value="item.key"
            ></el-option>
          </el-select>
          <div
            v-if="
              createForm.modelType === embedding && provider.key === yuanjing
            "
            class="embedding-tip"
          >
            {{ $t('modelAccess.table.embeddingTip') }}
          </div>
        </el-form-item>
        <el-form-item :label="$t('modelAccess.table.model')" prop="model">
          <el-select
            v-if="!isEdit"
            v-model="modelObjStr"
            :placeholder="$t('common.select.placeholder')"
            style="width: 100%"
            @change="handleModelChange"
          >
            <el-option
              v-for="item in modelIdList"
              v-if="item.displayName"
              :key="item.model"
              :label="item.displayName"
              :value="JSON.stringify(item)"
            >
              <div class="model-option-content">
                <div class="model-option-content-left">
                  <span class="model-name">
                    {{ item.displayName }}
                  </span>
                </div>

                <div
                  class="model-select-tags"
                  v-if="item.tags && item.tags.length > 0"
                >
                  <span
                    v-for="(tag, tagIdx) in item.tags"
                    :key="tagIdx"
                    class="model-select-tag"
                  >
                    {{ tag.text }}
                  </span>
                </div>
              </div>
            </el-option>
          </el-select>
          <el-input
            v-if="currentModelId === customModelId || isEdit"
            :disabled="isEdit"
            v-model="createForm.model"
            :placeholder="$t('common.input.placeholder')"
            @blur="changeCustomModelId"
          ></el-input>
        </el-form-item>
        <el-form-item
          :label="$t('modelAccess.table.modelDisplayName')"
          prop="displayName"
        >
          <el-input
            v-model="createForm.displayName"
            :placeholder="$t('common.hint.text')"
            :disabled="!allowEdit"
            maxlength="50"
            show-word-limit
          ></el-input>
        </el-form-item>
        <el-form-item :label="$t('modelAccess.table.picPath')" prop="avatar">
          <el-upload
            class="avatar-uploader"
            action=""
            name="files"
            :show-file-list="false"
            :http-request="handleUploadImage"
            :on-error="handleUploadError"
            accept=".png,.jpg,.jpeg"
            :disabled="!allowEdit"
          >
            <img
              class="upload-img"
              :src="
                createForm.avatar && createForm.avatar.path
                  ? avatarSrc(createForm.avatar.path)
                  : defaultLogo
              "
            />
            <!--<span style="margin-left: 12px; color: #606266 !important;" v-if="createForm.avatar && createForm.avatar.path">
              {{createForm.avatar.path}}
            </span>-->
            <span class="upload-hint">{{ $t('modelAccess.hint.upload') }}</span>
          </el-upload>
        </el-form-item>
        <el-form-item
          :label="$t('modelAccess.table.modelDesc')"
          prop="modelDesc"
        >
          <el-input
            type="text"
            v-model="createForm.modelDesc"
            :placeholder="$t('common.input.placeholder')"
            :disabled="!allowEdit"
            show-word-limit
            maxlength="200"
          ></el-input>
        </el-form-item>
        <el-form-item
          v-if="createForm.modelType === llm"
          label="Function Call"
          prop="functionCalling"
        >
          <el-select
            v-model="createForm.functionCalling"
            :placeholder="$t('common.select.placeholder')"
            :disabled="!allowEdit"
            style="width: 100%"
          >
            <el-option
              v-for="item in functionCalling"
              :key="item.key"
              :label="item.name"
              :value="item.key"
            ></el-option>
          </el-select>
        </el-form-item>
        <el-form-item v-if="showVision()" label="Vision" prop="visionSupport">
          <el-select
            v-model="createForm.visionSupport"
            :placeholder="$t('common.select.placeholder')"
            :disabled="!allowEdit"
            style="width: 100%"
          >
            <el-option
              v-for="item in supportList"
              :key="item.key"
              :label="item.name"
              :value="item.key"
            ></el-option>
          </el-select>
        </el-form-item>
        <!-- 多模态模型去掉支持的文件类型选择 -->
        <!--<el-form-item
          v-if="isMultiModal()"
          :label="$t('modelAccess.table.supportFileType')"
          prop="supportFileTypes"
        >
          <el-select
            v-model="createForm.supportFileTypes"
            :placeholder="$t('common.select.placeholder')"
            :disabled="!allowEdit"
            style="width: 100%"
            multiple
          >
            <el-option
              v-for="item in Object.keys(supportFileTypeObj)"
              :key="item"
              :label="supportFileTypeObj[item]"
              :value="item"
            ></el-option>
          </el-select>
        </el-form-item>-->
        <el-form-item
          v-if="showMaxPicLimit()"
          :label="$t('modelAccess.table.maxPicLimit')"
          prop="maxImageSize"
        >
          <el-input-number
            v-model="createForm.maxImageSize"
            :placeholder="$t('common.input.placeholder')"
            :min="0"
            :disabled="!allowEdit"
          ></el-input-number>
          M
        </el-form-item>
        <!-- 多模态模型去掉视频片限制，和最大文本长度 -->
        <!--<el-form-item
          v-if="showMaxVideoLimit()"
          :label="$t('modelAccess.table.maxVideoLimit')"
          prop="maxVideoClipSize"
        >
          <el-input-number
            v-model="createForm.maxVideoClipSize"
            :placeholder="$t('common.input.placeholder')"
            :min="0"
            :disabled="!allowEdit"
          ></el-input-number>
          M
        </el-form-item>
        <el-form-item
          v-if="isMultiModal()"
          :label="$t('modelAccess.table.maxTextSize')"
          prop="maxTextLength"
        >
          <el-input-number
            v-model="createForm.maxTextLength"
            :placeholder="$t('common.input.placeholder')"
            :min="0"
            :disabled="!allowEdit"
          ></el-input-number>
          tokens
        </el-form-item>-->
        <div v-if="showMaxAudioLimit()">
          <el-form-item
            :label="$t('modelAccess.table.maxAudioLimit')"
            prop="maxAsrFileSize"
          >
            <el-input-number
              v-model="createForm.maxAsrFileSize"
              :placeholder="$t('common.input.placeholder')"
              :min="0"
              :disabled="!allowEdit"
            ></el-input-number>
            M
          </el-form-item>
        </div>
        <el-form-item
          v-if="showContextSize()"
          :label="$t('modelAccess.table.contextSize')"
          prop="contextSize"
        >
          <el-input-number
            v-model="createForm.contextSize"
            :placeholder="$t('common.input.placeholder')"
            :min="0"
            :disabled="!allowEdit"
          ></el-input-number>
          tokens
        </el-form-item>
        <el-form-item
          v-if="createForm.modelType === llm"
          label="Max_token"
          prop="maxTokens"
        >
          <el-input-number
            v-model="createForm.maxTokens"
            :placeholder="$t('common.input.placeholder')"
            :min="0"
            :disabled="!allowEdit"
          ></el-input-number>
          tokens
        </el-form-item>
        <el-form-item
          v-if="provider.key !== ollama && !showAppAndAccessKey()"
          :label="$t('modelAccess.table.apiKey')"
          prop="apiKey"
        >
          <el-input
            type="password"
            v-model="createForm.apiKey"
            :placeholder="
              $t('common.hint.apiKey') + (typeObj.apiKey[provider.key] || '--')
            "
            :disabled="!allowEdit"
          ></el-input>
        </el-form-item>
        <div v-if="showAppAndAccessKey()">
          <el-form-item label="APP Key" prop="appKey">
            <el-input
              type="password"
              v-model="createForm.appKey"
              :placeholder="$t('common.hint.appKey')"
              :disabled="!allowEdit"
            ></el-input>
          </el-form-item>
          <el-form-item label="Access Key" prop="accessKey">
            <el-input
              type="password"
              v-model="createForm.accessKey"
              :placeholder="$t('common.hint.accessKey')"
              :disabled="!allowEdit"
            ></el-input>
          </el-form-item>
        </div>
        <el-form-item
          :label="$t('modelAccess.table.inferUrl')"
          prop="endpointUrl"
        >
          <el-input
            v-model="createForm.endpointUrl"
            :title="
              $t('common.hint.inferUrl') +
              (typeObj.inferUrl[`${createForm.modelType}_${provider.key}`] ||
                typeObj.inferUrl[provider.key] ||
                '--')
            "
            :placeholder="
              $t('common.hint.inferUrl') +
              (typeObj.inferUrl[`${createForm.modelType}_${provider.key}`] ||
                typeObj.inferUrl[provider.key] ||
                '--')
            "
            :disabled="!allowEdit"
          ></el-input>
        </el-form-item>
        <!-- 系统管理员可设置的公开范围（个人、全局），普通用户可设置的公开范围（个人、组织内） -->
        <el-form-item
          :label="$t('modelAccess.table.scopeType')"
          prop="scopeType"
        >
          <el-select
            v-model="createForm.scopeType"
            :placeholder="$t('common.select.placeholder')"
            :disabled="!allowEdit || isEdit"
            style="width: 100%"
          >
            <el-option
              v-for="item in scopeTypeList"
              :key="item.key"
              :label="item.name"
              :value="item.key"
              v-show="isSystem ? item.key !== org : item.key !== all"
            ></el-option>
          </el-select>
        </el-form-item>
        <el-form-item
          v-if="createForm.modelType === llm"
          prop="thinkingSupport"
        >
          <template slot="label">
            <span style="font-size: inherit">
              {{ $t('modelAccess.table.think') }}
            </span>
            <el-tooltip
              :content="$t('modelAccess.table.thinkTips')"
              placement="top"
            >
              <i class="el-icon-question thinking-tips-icon" />
            </el-tooltip>
          </template>
          <el-select
            v-model="createForm.thinkingSupport"
            :placeholder="$t('common.select.placeholder')"
            :disabled="!allowEdit || linkStatus === 'test'"
            style="width: 100%"
            @change="handleChangeThinking"
          >
            <el-option
              v-for="item in supportList"
              :key="item.key"
              :label="item.name"
              :value="item.key"
            ></el-option>
          </el-select>
          <div
            class="thinking-link-test"
            v-if="createForm.thinkingSupport === 'support'"
          >
            <div class="left-test">
              <i class="el-icon-link"></i>
              <span>{{ $t('modelAccess.linkText') }}</span>
            </div>
            <div class="right-result">
              <div v-if="linkStatus === 'test'">
                <i class="el-icon-loading"></i>
                {{ $t('modelAccess.testing') }}
              </div>
              <div v-if="linkStatus === 'success'" class="test-success">
                <i class="el-icon-success"></i>
                {{ $t('modelAccess.testSuccess') }}
              </div>
              <div v-if="linkStatus === 'error'" class="test-error">
                <i class="el-icon-error"></i>
                {{ $t('modelAccess.testError') }}
                <i
                  class="el-icon-refresh-right"
                  @click="handleChangeThinking(support)"
                ></i>
              </div>
            </div>
          </div>
        </el-form-item>
        <!--<el-form-item :label="$t('modelAccess.table.publishTime')" prop="publishDate">
          <el-date-picker
            v-model="createForm.publishDate"
            type="date"
            value-format="yyyy-MM-dd"
            :placeholder="$t('common.select.placeholder')"
            :disabled="!allowEdit"
          >
          </el-date-picker>
        </el-form-item>-->
        <el-form-item v-if="isEdit" label="uuid">
          {{ row.uuid || '--' }}
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer" v-if="allowEdit">
        <el-button @click="handleClose">
          {{ $t('common.button.cancel') }}
        </el-button>
        <el-button
          :disabled="['test', 'error'].includes(linkStatus)"
          :loading="loading"
          type="primary"
          @click="handleSubmit"
        >
          {{ $t('common.button.confirm') }}
        </el-button>
      </span>
    </el-dialog>
  </div>
</template>
<script>
import {
  addModel,
  editModel,
  fetchModelIdList,
  testLinkThinking,
} from '@/api/modelAccess';
import { uploadAvatar } from '@/api/user';
import { avatarSrc, getModelDefaultIcon } from '@/utils/util';
import {
  PROVIDER_TYPE,
  PROVIDER_OBJ,
  FUNC_CALLING,
  DEFAULT_CALLING,
  SUPPORT,
  NO_SUPPORT,
  SUPPORT_LIST,
  TYPE_OBJ,
  LLM,
  RERANK,
  EMBEDDING,
  MULTIMODAL_EMBEDDING,
  MULTIMODAL_RERANK,
  ASR,
  OPENAI_API,
  OLLAMA,
  YUAN_JING,
  HUOSHAN,
  SHOW_VISION_LIST,
  CUSTOM_MODEL_ID,
  DEFAULT_MODEL_ITEM,
  SUPPORT_FILE_TYPE_OBJ,
  IMAGE,
  VIDEO,
  SCOPE_TYPE_LIST,
  DEFAULT_CONTENT_SIZE,
  DEFAULT_MAX_TOKEN,
  PRIVATE,
  ORG,
  ALL,
} from '../constants';
import LinkIcon from '@/components/linkIcon.vue';

export default {
  components: { LinkIcon },
  data() {
    const minLimit = (rule, value, callback) => {
      if (value !== null && value < 1) {
        callback(new Error(this.$t('modelAccess.hint.minLimit')));
      } else {
        callback();
      }
    };
    return {
      isSystem: this.$store.state.user.permission.isSystem || false,
      allowEdit: true,
      defaultLogo: getModelDefaultIcon(),
      dialogVisible: false,
      modelType: [],
      scopeTypeList: SCOPE_TYPE_LIST,
      org: ORG,
      all: ALL,
      functionCalling: FUNC_CALLING,
      support: SUPPORT,
      supportList: SUPPORT_LIST,
      supportFileTypeObj: SUPPORT_FILE_TYPE_OBJ,
      typeObj: TYPE_OBJ,
      llm: LLM,
      embedding: EMBEDDING,
      ollama: OLLAMA,
      yuanjing: YUAN_JING,
      showVisionList: SHOW_VISION_LIST,
      showContextSizeList: [
        LLM,
        EMBEDDING,
        RERANK,
        MULTIMODAL_RERANK,
        MULTIMODAL_EMBEDDING,
      ],
      customModelId: CUSTOM_MODEL_ID,
      modelIdList: [DEFAULT_MODEL_ITEM],
      currentModelId: '',
      modelObjStr: '',
      linkStatus: '',
      createForm: {
        model: '',
        displayName: '',
        endpointUrl: '',
        scopeType: PRIVATE,
        apiKey: '',
        appKey: '',
        accessKey: '',
        modelType: LLM,
        modelDesc: '',
        contextSize: DEFAULT_CONTENT_SIZE,
        maxTokens: DEFAULT_MAX_TOKEN,
        maxAsrFileSize: 10,
        maxImageSize: 3,
        /*maxVideoClipSize: 10,
        maxTextLength: 512,
        supportFileTypes: [IMAGE, VIDEO],*/
        avatar: {
          key: '',
          path: '',
        },
        // publishDate: '',
        functionCalling: DEFAULT_CALLING,
        visionSupport: NO_SUPPORT,
        thinkingSupport: NO_SUPPORT,
      },
      rules: {
        model: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
          // { pattern: this.$config.commonTextReg, message: this.$t('common.hint.text'), trigger: "blur"}
          // { min: 2, max: 50, message: this.$t('common.hint.textLimit'), trigger: 'blur'},
        ],
        appKey: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
        ],
        accessKey: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
        ],
        maxAsrFileSize: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
          {
            validator: minLimit,
            trigger: 'blur',
          },
        ],
        contextSize: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
        ],
        maxTokens: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
        ],
        maxImageSize: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
          {
            validator: minLimit,
            trigger: 'blur',
          },
        ],
        displayName: [
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
        modelType: [
          {
            required: true,
            message: this.$t('common.select.placeholder'),
            trigger: 'change',
          },
        ],
        endpointUrl: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
          {
            pattern:
              /^(http|ftp|https):\/\/[\w\-_]+(\.[\w\-_]+)+([\w\-\.,@?^=%&:/~\+#]*[\w\-\@?^=%&/~\+#])?$/,
            message: this.$t('modelAccess.hint.urlError'),
            trigger: 'blur',
          },
        ],
        scopeType: [
          {
            required: true,
            message: this.$t('common.select.placeholder'),
            trigger: 'change',
          },
        ],
        /*supportFileTypes: [
          {
            required: true,
            message: this.$t('common.select.placeholder'),
            trigger: 'change',
          },
        ],*/
      },
      row: {},
      provider: {},
      isEdit: false,
      loading: false,
    };
  },
  watch: {
    'createForm.modelType': {
      handler() {
        this.$refs.createForm && this.$refs.createForm.clearValidate();
        if (!this.isEdit) {
          this.setDefaultInferUrl();
          this.clearSelectModelId();
          this.getModelIdList();
        }
      },
      immediate: false,
    },
    'createForm.thinkingSupport': {
      handler(val) {
        if (val === NO_SUPPORT) {
          this.linkStatus = '';
        }
      },
      immediate: false,
    },
    'provider.key': {
      handler(newVal) {
        if (!this.isEdit && newVal) {
          this.setDefaultInferUrl();
          this.getModelIdList();
        }
      },
      immediate: false,
    },
  },
  methods: {
    avatarSrc,
    isMultiModal() {
      return [MULTIMODAL_RERANK, MULTIMODAL_EMBEDDING].includes(
        this.createForm.modelType,
      );
    },
    showAppAndAccessKey() {
      return this.provider.key === HUOSHAN && this.createForm.modelType === ASR;
    },
    showVision() {
      return (
        this.createForm.modelType === LLM &&
        this.showVisionList.includes(this.provider.key)
      );
    },
    showContextSize() {
      return this.showContextSizeList.includes(this.createForm.modelType);
    },
    showFileTypeLimit(type) {
      return (
        this.isMultiModal() && this.createForm.supportFileTypes.includes(type)
      );
    },
    showMaxPicLimit() {
      const { modelType, visionSupport } = this.createForm || {};
      return (
        (modelType === LLM && visionSupport === SUPPORT) || this.isMultiModal() // this.showFileTypeLimit(IMAGE)
      );
    },
    showMaxVideoLimit() {
      return this.showFileTypeLimit(VIDEO);
    },
    showMaxAudioLimit() {
      return this.createForm.modelType === ASR;
    },
    handleChangeThinking(value) {
      if (value === SUPPORT) {
        this.linkStatus = 'test';
        const thinkingData = this.formatSubmitData();
        testLinkThinking(thinkingData)
          .then(res => {
            if (res.code === 0) {
              this.linkStatus = 'success';
            } else {
              this.linkStatus = 'error';
            }
          })
          .catch(() => {
            this.linkStatus = 'error';
          });
      } else {
        this.linkStatus = '';
      }
    },
    getModelIdList() {
      fetchModelIdList({
        provider: this.provider.key,
        modelType: this.createForm.modelType,
      }).then(res => {
        const list = res.data ? res.data.list || [] : [];
        this.modelIdList = [...list, DEFAULT_MODEL_ITEM];
      });
    },
    handleModelChange(modelObjStr) {
      const modelObj = modelObjStr ? JSON.parse(modelObjStr) : {};
      const modelId = modelObj.model || '';
      // currentModelId 用于判断是否展示出自定义的模型 ID
      this.currentModelId = modelId;
      // 根据选择的 model 配置显示
      this.createForm = {
        ...this.createForm,
        model: modelId !== CUSTOM_MODEL_ID ? modelId : '',
        functionCalling: modelObj.functionCalling || DEFAULT_CALLING,
        visionSupport: modelObj.visionSupport || NO_SUPPORT,
        thinkingSupport: modelObj.thinkingSupport || NO_SUPPORT,
      };
      this.$refs.createForm.clearValidate('model');
      // 切换列表中模型 ID 为支持深度思考时，连接状态默认成功
      if (this.createForm.thinkingSupport === SUPPORT) {
        this.linkStatus = 'success';
      }
    },
    changeCustomModelId(value) {
      this.createForm.thinkingSupport = NO_SUPPORT;
    },
    uploadAvatar(file, key) {
      const formData = new FormData();
      const config = { headers: { 'Content-Type': 'multipart/form-data' } };
      formData.append(key, file);
      return uploadAvatar(formData, config);
    },
    handleUploadImage(data) {
      if (data.file) {
        this.uploadAvatar(data.file, 'avatar').then(res => {
          const { key, path } = res.data || {};
          this.createForm.avatar = { key, path };
        });
      }
    },
    handleUploadError() {
      this.$message.error(this.$t('common.message.uploadError'));
    },
    formatValue(data) {
      for (let key in this.createForm) {
        this.createForm[key] = data ? data[key] || '' : '';
      }
    },
    setDefaultInferUrl() {
      const defaultUrl =
        this.typeObj.inferUrl[
          `${this.createForm.modelType}_${this.provider.key}`
        ] || this.typeObj.inferUrl[this.provider.key];
      if (defaultUrl) {
        // OpenAI-API-compatible 供应商不自动设置推理 URL
        this.createForm.endpointUrl =
          this.provider.key === OPENAI_API ? '' : defaultUrl;
      }
    },
    openDialog(title, row, providerType) {
      // 创建或者允许编辑时，可操作
      this.allowEdit = !row || row.allowEdit;
      this.provider = { key: title, name: PROVIDER_OBJ[title] };
      const matchProviderType = providerType || PROVIDER_TYPE;
      const currentProvider =
        matchProviderType.find(item => item.key === title) || {};
      this.modelType = currentProvider.children || [];
      this.createForm.modelType = this.modelType[0]
        ? this.modelType[0].key || LLM
        : LLM;

      // 自动填入推理URL默认值
      if (!row) {
        this.setDefaultInferUrl();
      }

      this.dialogVisible = true;

      this.isEdit = Boolean(row);
      if (this.isEdit) {
        this.row = row || {};
        this.formatValue(row);
        if (row.thinkingSupport === SUPPORT) {
          this.linkStatus = 'success';
        }
      }
    },
    clearSelectModelId() {
      this.createForm.model = '';
      this.modelObjStr = '';
      this.currentModelId = '';
    },
    handleClose() {
      this.dialogVisible = false;
      this.linkStatus = '';
      this.clearSelectModelId();
      this.formatValue({
        modelType: LLM,
        functionCalling: DEFAULT_CALLING,
        visionSupport: NO_SUPPORT,
        thinkingSupport: NO_SUPPORT,
        scopeType: PRIVATE,
        contextSize: DEFAULT_CONTENT_SIZE,
        maxTokens: DEFAULT_MAX_TOKEN,
        maxAsrFileSize: 10,
        maxImageSize: 3,
        /*maxVideoClipSize: 10,
        maxTextLength: 512,
        supportFileTypes: [IMAGE, VIDEO],*/
        avatar: { key: '', path: '' },
      });
      if (this.$refs.createForm) {
        this.$refs.createForm.resetFields();
        this.$refs.createForm.clearValidate();
      }
    },
    formatSubmitData() {
      const {
        apiKey,
        appKey,
        accessKey,
        endpointUrl,
        functionCalling,
        modelType,
        visionSupport,
        thinkingSupport,
        contextSize,
        maxTokens,
        /*maxTextLength,
        maxVideoClipSize,
        supportFileTypes,*/
        maxImageSize,
        maxAsrFileSize,
      } = this.createForm;
      const form = {
        ...this.createForm,
        provider: this.provider.key || '',
        config: {
          endpointUrl,
          ...(this.provider.key !== OLLAMA &&
            !this.showAppAndAccessKey() && { apiKey }),
          ...(modelType === LLM && {
            functionCalling,
            maxTokens,
            thinkingSupport,
          }),
          ...(this.showVision() && { visionSupport }),
          ...(this.showContextSize() && { contextSize }),
          ...(this.showMaxAudioLimit() && { maxAsrFileSize }),
          ...(this.showMaxPicLimit() && { maxImageSize }),
          ...(this.showAppAndAccessKey() && { appKey, accessKey }),
          /*...(this.showMaxVideoLimit() && { maxVideoClipSize }),
          ...(this.isMultiModal() && { supportFileTypes, maxTextLength }),*/
        },
      };
      const deleteKeys = [
        'apiKey',
        'appKey',
        'accessKey',
        'endpointUrl',
        'functionCalling',
        'visionSupport',
        'thinkingSupport',
        'contextSize',
        'maxTokens',
        /*'maxTextLength',
        'maxVideoClipSize',
        'supportFileTypes',*/
        'maxImageSize',
        'maxAsrFileSize',
      ];
      deleteKeys.forEach(key => {
        delete form[key];
      });
      return form;
    },
    handleSubmit() {
      this.$refs.createForm.validate(async valid => {
        if (valid) {
          const form = this.formatSubmitData();

          try {
            this.loading = true;
            const res = this.isEdit
              ? await editModel({ ...form, modelId: this.row.modelId })
              : await addModel(form);
            if (res.code === 0) {
              this.$message.success(this.$t('common.message.success'));
              this.handleClose();
              this.$emit('reloadData', !this.isEdit);
            }
          } finally {
            this.loading = false;
          }
        }
      });
    },
  },
};
</script>
<style lang="scss" scoped>
@import '@/style/modelSelect.scss';
.createForm {
  padding: 0 45px 0 20px;
  .avatar-uploader {
    .upload-img {
      object-fit: cover;
      width: 80px;
      height: 80px;
      border-radius: 8px;
      border: 1px solid #dcdfe6;
      display: inline-block;
      vertical-align: middle;
    }
    .upload-hint {
      display: inline-block;
      margin-left: 12px;
      color: #909399 !important;
    }
  }
  .embedding-tip {
    color: #f56c6c;
    line-height: 16px;
  }
  .thinking-tips-icon {
    margin-left: 4px;
    cursor: pointer;
    font-size: 14px;
  }
  .thinking-link-test {
    display: flex;
    align-items: center;
    margin-top: 8px;
    .left-test {
      border-radius: 4px;
      border: 1px solid #dcdfe6;
      padding: 2px 10px;
      height: 30px;
      line-height: 24px;
      margin-right: 10px;
      .el-icon-link {
        transform: rotate(45deg);
        font-size: 16px;
        color: #1cc78d;
        margin-right: 3px;
      }
    }
    .right-result {
      .test-success {
        color: #67c23a;
      }
      .test-error {
        color: #f56c6c;
        .el-icon-refresh-right {
          cursor: pointer;
          color: #606266;
          margin-left: 5px;
        }
      }
    }
  }
}
.dialog-title-wrapper {
  display: flex;
  align-items: center;
  .dialog-title {
    color: $color_title;
    font-size: 18px;
    font-weight: bold;
  }
  .dialog-desc {
    color: #888;
    margin-left: 20px;
  }
}
</style>
