<template>
  <div class="page-wrapper full-content">
    <div class="page-title">
      <span class="el-icon-arrow-left back" @click="goBack"></span>
      {{
        mode === 'config'
          ? title
          : $t('knowledgeManage.knowledgeDatabase.fileUpload.addFile')
      }}
      <LinkIcon type="knowledge" />
    </div>
    <div class="table-box">
      <div class="fileUpload">
        <!-- 文件上传 -->
        <div v-if="mode !== 'config'" class="upload-section">
          <p class="section-title">
            {{
              $t('knowledgeManage.knowledgeDatabase.fileUpload.fileUpload')
            }}：
          </p>
          <div class="fileBtn" v-if="category === 2">
            <el-radio-group v-model="fileType" @change="fileTypeChange">
              <el-radio-button label="fileMultiModal" v-if="category === 2">
                {{
                  $t(
                    'knowledgeManage.knowledgeDatabase.fileUpload.fileMultiModal',
                  )
                }}
              </el-radio-button>
              <el-radio-button label="file">
                {{ $t('knowledgeManage.knowledgeDatabase.fileUpload.file') }}
              </el-radio-button>
              <!--              <el-radio-button label="fileUrl">-->
              <!--                {{ $t('knowledgeManage.knowledgeDatabase.fileUpload.fileUrl') }}-->
              <!--              </el-radio-button>-->
              <!--              <el-radio-button label="url">-->
              <!--                {{ $t('knowledgeManage.knowledgeDatabase.fileUpload.url') }}-->
              <!--              </el-radio-button>-->
            </el-radio-group>
          </div>
          <div
            element-loading-background="rgba(255, 255, 255, 0.5)"
            v-if="fileType !== 'url'"
          >
            <div class="dialog-body">
              <el-upload
                :class="['upload-box']"
                drag
                action=""
                :show-file-list="false"
                :auto-upload="false"
                :multiple="fileType !== 'fileUrl'"
                :limit="fileType === 'fileUrl' ? 1 : undefined"
                :accept="acceptType"
                :file-list="fileList"
                :on-change="uploadOnChange"
              >
                <div>
                  <div>
                    <img
                      :src="require('@/assets/imgs/uploadImg.png')"
                      class="upload-img"
                    />
                    <p class="click-text">
                      {{
                        $t(
                          'knowledgeManage.knowledgeDatabase.fileUpload.clickText',
                        )
                      }}
                      <span class="clickUpload">
                        {{
                          $t(
                            'knowledgeManage.knowledgeDatabase.fileUpload.clickUpload',
                          )
                        }}
                      </span>
                    </p>
                  </div>
                  <div class="tips">
                    <p v-if="fileType === 'file'">
                      <span class="red">*</span>
                      {{
                        $t(
                          'knowledgeManage.knowledgeDatabase.fileUpload.uploadTips1',
                        )
                      }}
                    </p>
                    <p v-if="fileType === 'file'">
                      <span class="red">*</span>
                      {{
                        $t(
                          'knowledgeManage.knowledgeDatabase.fileUpload.uploadTips2',
                        )
                      }}
                    </p>
                    <template
                      v-if="fileType === 'fileMultiModal'"
                      v-for="uploadLimit in uploadLimitList"
                    >
                      <p v-if="uploadLimit.fileType === 'video'">
                        <span class="red">*</span>
                        {{
                          $t(
                            'knowledgeManage.multiKnowledgeDatabase.uploadTipsVideo',
                            {
                              extList: uploadLimit.extList.join('、'),
                              maxSize: uploadLimit.maxSize,
                            },
                          )
                        }}
                      </p>
                      <p v-if="uploadLimit.fileType === 'audio'">
                        <span class="red">*</span>
                        {{
                          $t(
                            'knowledgeManage.multiKnowledgeDatabase.uploadTipsAudio',
                            { extList: uploadLimit.extList.join('、') },
                          )
                        }}
                      </p>
                      <p v-if="uploadLimit.fileType === 'image'">
                        <span class="red">*</span>
                        {{
                          $t(
                            'knowledgeManage.multiKnowledgeDatabase.uploadTipsImage',
                            {
                              extList: uploadLimit.extList.join('、'),
                              maxSize: uploadLimit.maxSize,
                            },
                          )
                        }}
                      </p>
                    </template>
                    <p v-if="fileType === 'fileUrl'">
                      <span class="red">*</span>
                      {{
                        $t(
                          'knowledgeManage.knowledgeDatabase.fileUpload.uploadTips3',
                        )
                      }}
                      <a
                        class="template_downLoad"
                        href="#"
                        @click.prevent.stop="downloadTemplate"
                      >
                        {{ $t('common.fileUpload.templateClick') }}
                      </a>
                    </p>
                    <p v-if="fileType === 'fileUrl'">
                      <span class="red">*</span>
                      {{
                        $t(
                          'knowledgeManage.knowledgeDatabase.fileUpload.uploadTips4',
                        )
                      }}
                    </p>
                  </div>
                </div>
              </el-upload>
            </div>
          </div>
          <div class="el-upload-url" v-else>
            <div class="upload-url">
              <urlAnalysis
                :categoryId="knowledgeId"
                ref="urlUpload"
                @handleLoading="handleLoading"
                @handleSetData="handleSetData"
              />
            </div>
          </div>
        </div>

        <!-- 上传文件的列表 -->
        <div class="file-list" v-if="fileList.length > 0 && mode !== 'config'">
          <transition name="el-zoom-in-top">
            <ul class="document_lise">
              <li
                v-for="(file, index) in fileList"
                :key="index"
                class="document_lise_item"
              >
                <div style="padding: 8px 0" class="lise_item_box">
                  <span class="size">
                    <img :src="require('@/assets/imgs/fileicon.png')" />
                    {{ file.name }}
                    <span class="file-size">
                      {{ filterSize(file.size) }}
                    </span>
                    <el-progress
                      :percentage="file.percentage"
                      v-if="file.percentage !== 100"
                      :status="file.progressStatus"
                      max="100"
                      class="progress"
                    ></el-progress>
                  </span>
                  <span class="handleBtn">
                    <span>
                      <span v-if="file.percentage === 100">
                        <i
                          class="el-icon-check check success"
                          v-if="file.progressStatus === 'success'"
                        ></i>
                        <i class="el-icon-close close fail" v-else></i>
                      </span>
                      <i
                        class="el-icon-loading"
                        v-else-if="
                          file.percentage !== 100 && index === fileIndex
                        "
                      ></i>
                    </span>
                    <span style="margin-left: 30px">
                      <i
                        class="el-icon-error error"
                        @click="handleRemove(file, index)"
                      ></i>
                    </span>
                  </span>
                </div>
              </li>
            </ul>
          </transition>
        </div>
        <!-- 参数设置 -->
        <p class="section-title" v-if="mode !== 'config'">
          {{
            $t('knowledgeManage.knowledgeDatabase.fileUpload.paramSetting')
          }}：
        </p>
        <div class="parse-mode-list" v-if="mode !== 'config'">
          <div
            v-for="item in parseModeList"
            :key="item.label"
            :class="[
              'parse-mode-item',
              parseMode === item.label ? 'activeAnalyzer' : '',
            ]"
            @click="parseMode = item.label"
          >
            <p class="analyzerItem_text">{{ item.text }}</p>
            <h3 class="analyzerItem_desc">{{ item.desc }}</h3>
          </div>
        </div>
        <div class="params_form">
          <el-form
            :model="ruleForm"
            ref="ruleForm"
            label-width="140px"
            class="demo-ruleForm"
            @submit.native.prevent
            label-position="left"
          >
            <template v-if="parseMode === PARSE_MODE_TEMPLATE">
              <el-form-item
                class="parse-template-item"
                :label="$t('knowledgeManage.parseTemplate.title')"
              >
                <parseTemplateSelect
                  ref="parseTemplateSelect"
                  v-model="ruleForm.parseTemplate"
                  :scope="fileType === 'fileMultiModal' ? 'media' : 'doc'"
                  :requiredDocTypes="missingDocTypes"
                  collapsible
                  @unbound-change="unboundDocTypes = $event"
                >
                  <template #footer>
                    <el-checkbox
                      v-if="isSystemAdmin"
                      v-model="ruleForm.overrideTemplate"
                    >
                      {{
                        $t('knowledgeManage.parseTemplate.overrideKnowledge')
                      }}
                    </el-checkbox>
                    <el-tooltip
                      v-if="isSystemAdmin"
                      :content="
                        $t('knowledgeManage.parseTemplate.overrideKnowledgeTip')
                      "
                      placement="top"
                    >
                      <span class="el-icon-question question"></span>
                    </el-tooltip>
                  </template>
                </parseTemplateSelect>
              </el-form-item>
            </template>
            <parseConfigForm
              v-else
              ref="parseConfig"
              :value="ruleForm"
              :multiModal="fileType === 'fileMultiModal'"
              :mediaTypes="[...fileFormatSet]"
              @asr-change="handleASR"
            />
          </el-form>
        </div>

        <!-- 元数据管理：在参数设置框之外 -->
        <div class="meta-section" v-if="mode !== 'config'">
          <p class="section-title meta-label">
            {{ $t('knowledgeManage.metadataManagement') }}
            <span class="optional-tip">
              {{ $t('knowledgeManage.parseTemplate.optional') }}
            </span>
            <el-tooltip
              :content="$t('knowledgeManage.metadataManagementTips')"
              placement="right"
            >
              <span class="el-icon-question question"></span>
            </el-tooltip>
          </p>
          <mataData
            ref="mataData"
            @updateMeta="updateMeta"
            :knowledgeId="knowledgeId"
            :withCompressed="withCompressed"
          />
        </div>
        <div class="next">
          <el-button
            type="primary"
            size="mini"
            @click="submitInfo"
            :disabled="!confirmFlag"
            :loading="urlLoading"
          >
            {{ $t('common.button.confirm') }}
          </el-button>
          <el-button
            size="mini"
            @click="formReset"
            v-if="parseMode !== PARSE_MODE_TEMPLATE"
          >
            {{ $t('common.button.restore') }}
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
import urlAnalysis from '../component/urlAnalysis.vue';
import uploadChunk from '@/mixins/uploadChunk';
import {
  docImport,
  updateDocConfig,
  getDocConfig,
  getDocList,
  getDocLimit,
  getDocDetail,
} from '@/api/knowledge';
import { delfile } from '@/api/chunkFile';
import { getParseTemplateList } from '@/api/parseTemplate';
import {
  DOC_TYPE_LIST,
  bindListToMap,
  bindMapToList,
  getDocTypeByFileName,
} from '../parseTemplate/config';
import FileIcon from '@/components/FileIcon.vue';
import { selectASRList } from '@/api/modelAccess';
import LinkIcon from '@/components/linkIcon.vue';
import parseConfigForm from '../component/parseConfigForm.vue';
import parseTemplateSelect from '../component/parseTemplateSelect.vue';
import mataData from '../component/metadata.vue';
import { USER_API } from '@/utils/requestConstants';
import { POWER_TYPE_SYSTEM_ADMIN } from '@/views/knowledge/constants';
import {
  PARSE_MODE_CUSTOM,
  PARSE_MODE_LIST,
  PARSE_MODE_TEMPLATE,
} from '../config';
import { deepMerge, filterSize, authDownload } from '@/utils/util';

export default {
  name: 'KnowledgeFileUpload',
  components: {
    LinkIcon,
    urlAnalysis,
    parseConfigForm,
    parseTemplateSelect,
    mataData,
  },
  mixins: [uploadChunk],
  data() {
    return {
      splitterValue: '',
      tableData: [],
      uploadLimitList: [],
      maxSizeAudio: 9999,
      confirmFlag: true,
      urlValidate: false,
      fileType:
        Number(this.$route.query.category) === 2 ? 'fileMultiModal' : 'file',
      withCompressed: false,
      knowledgeId: this.$route.query.id,
      knowledgeName: this.$route.query.name,
      mode: this.$route.query.mode,
      title: this.$route.query.title,
      docIdList: Array.isArray(this.$route.query.docIdList)
        ? this.$route.query.docIdList
        : [this.$route.query.docIdList].filter(id => id !== undefined),
      category: Number(this.$route.query.category),
      fileList: [],
      fileUrl: '',
      docInfoList: [],
      ruleForm: {
        docAnalyzer: ['text'],
        docMetaData: [], //元数据管理数据
        docPreprocess: ['replaceSymbols'], //'deleteLinks','replaceSymbols'
        docSegment: {
          segmentType: '0', //0是自动分段，1是自定义分段
          // splitter: ["！", "。", "？", "?", "!", ".", "......"],
          splitter: ['\n\n'],
          maxSplitter: 1024,
          overlap: 0.2,
          segmentMethod: '0', //0是通用分段，1是父子分段
          subMaxSplitter: 200, //父子分段必填
          // subSplitter:["！", "。", "？", "?", "!", ".", "......"]//父子分段必填
          subSplitter: ['\n'],
        },
        docInfoList: [],
        docImportType: 0,
        knowledgeId: this.$route.query.id,
        parserModelId: '',
        asrModelId: '',
        multimodalModelId: '',
        parseTemplate: {},
        overrideTemplate: true,
      },
      PARSE_MODE_TEMPLATE,
      // 由解析模板组件播报：当前没有模板可用的文档类型
      unboundDocTypes: [],
      // 修改已上传文档的解析配置时没有模板入口，只能走手动配置
      parseMode:
        this.$route.query.mode === 'config'
          ? PARSE_MODE_CUSTOM
          : PARSE_MODE_TEMPLATE,
      parseModeList: PARSE_MODE_LIST,
      audioTemplates: [],
      asrOptions: [],
      ruleFormBackup: {},
      urlLoading: false,
    };
  },
  watch: {
    fileList() {
      if (this.fileType === 'fileMultiModal') {
        this.ruleForm.docAnalyzer = ['text'];
      }
      if (this.parseMode === PARSE_MODE_TEMPLATE) {
        this.refreshTemplateAudioLimit();
        return;
      }
      this.confirmFlag = !(
        this.fileFormatSet.has('audio') && !this.ruleForm.asrModelId
      );
      this.verifyASR();
    },
    parseMode() {
      this.refreshTemplateAudioLimit();
    },
    // 模板绑定变了、模板列表或 ASR 列表回来了，选中的 ASR 模型都会跟着变
    templateAsrModel() {
      this.refreshTemplateAudioLimit();
    },
    // 刚缺模板就展开，别等点了确定才让用户找
    missingDocTypes(val, old) {
      if (val.length && !old.length) this.focusMissingTemplates();
    },
  },
  computed: {
    isSystemAdmin() {
      return (
        Number(this.$route.query.permissionType) === POWER_TYPE_SYSTEM_ADMIN
      );
    },
    // 模板模式的音频大小上限来自 audio 模板里配的 ASR 模型
    templateAsrModel() {
      const template = this.audioTemplates.find(
        item => item.templateId === this.ruleForm.parseTemplate.audio,
      );
      if (!template || !template.asrModelId) return null;
      return (
        this.asrOptions.find(item => item.modelId === template.asrModelId) ||
        null
      );
    },
    // 传了文件又没模板可用的类型，标红引导用户补选
    missingDocTypes() {
      return this.unboundDocTypes.filter(d => this.uploadedDocTypes.has(d));
    },
    // 本次上传涉及的解析模板文档类型，和 fileFormatSet 只覆盖音视频图不同
    uploadedDocTypes() {
      return new Set(
        this.fileList
          .map(file => getDocTypeByFileName(file.name))
          .filter(Boolean),
      );
    },
    fileFormatSet() {
      const fileFormatSet = new Set();
      for (const file of this.fileList) {
        const fileType = file.name.split('.').pop().toLowerCase();
        for (const uploadLimit of this.uploadLimitList) {
          const extList = uploadLimit.extList || [];
          if (extList.includes(fileType)) {
            fileFormatSet.add(uploadLimit.fileType);
          }
        }
      }
      return fileFormatSet;
    },
    acceptType() {
      switch (this.fileType) {
        case 'file':
          return '.pdf,.docx,.doc,.txt,.xlsx,.xls,.zip,.tar.gz,.csv,.pptx,.html,.md,.ofd,.wps';
        case 'fileMultiModal':
          return (
            '.' +
            this.uploadLimitList.flatMap(item => item.extList || []).join(',.')
          );
        case 'fileUrl':
          return '.xlsx';
        default:
          return '';
      }
    },
  },
  async created() {
    const query = this.$route.query;
    this.ruleFormBackup = structuredClone(this.ruleForm);
    if (query.mode !== 'config') {
      await this.getKnowledgeParseTemplate();
      this.getAudioTemplates();
    }
    if (query.mode === 'config' && this.docIdList.length === 1) {
      await getDocConfig({
        docId: this.docIdList[0],
        knowledgeId: this.knowledgeId,
      }).then(res => {
        if (res.code === 0) {
          this.ruleForm = deepMerge(this.ruleForm, res.data);
          this.ruleFormBackup = structuredClone(this.ruleForm);
          this.ruleForm.docAnalyzer = [...this.ruleForm.docAnalyzer];
          this.$nextTick(() => this.$refs.parseConfig.refresh());
        }
      });
    }
    if (this.category === 2) {
      if (this.docIdList.length > 0)
        getDocList({
          docName: '',
          graphStatus: [-1],
          knowledgeId: this.knowledgeId,
          docIdList: this.docIdList,
          metaValue: '',
          pageNo: 0,
          pageSize: 10,
          status: [-1],
        }).then(res => {
          if (res.code === 0) {
            this.fileList = res.data.list.map(item => ({
              name: item.docName,
              size: item.fileSize,
            }));
            this.fileType = res.data.list[0].isMultimodal
              ? 'fileMultiModal'
              : 'file';
            if (this.fileType === 'fileMultiModal')
              this.ruleForm.docAnalyzer = ['text'];
          }
        });
      getDocLimit({ knowledgeId: this.knowledgeId })
        .then(res => {
          if (res.code === 0) {
            this.uploadLimitList = res.data.uploadLimitList;
          } else {
            this.$router.back();
            this.$message.error(
              this.$t('knowledgeManage.multiKnowledgeDatabase.fileLimitError'),
            );
          }
        })
        .catch(() => {
          this.$router.back();
          this.$message.error(
            this.$t('knowledgeManage.multiKnowledgeDatabase.fileLimitError'),
          );
        });
    }
  },
  methods: {
    updateMeta(data) {
      this.ruleForm.docMetaData = data;
    },
    validateMetaData() {
      const hasEmptyField = this.ruleForm.docMetaData.some(item => {
        const isMetaKeyEmpty =
          !item.metaKey ||
          (typeof item.metaKey === 'string' && item.metaKey.trim() === '');
        const isMetaRuleRequired = item.metadataType !== 'value';
        const isMetaRuleEmpty =
          isMetaRuleRequired &&
          (!item.metaRule ||
            (typeof item.metaRule === 'string' && item.metaRule.trim() === ''));
        return isMetaKeyEmpty || isMetaRuleEmpty;
      });
      if (hasEmptyField) {
        this.$message.error(this.$t('knowledgeManage.metadataRequired'));
        return false;
      }
      return true;
    },
    goBack() {
      this.$router.go(-1);
    },
    handleASR(value, option) {
      if (!value || !option) {
        if (!this.fileFormatSet.has('audio')) this.confirmFlag = true;
        this.maxSizeAudio = 9999;
        return;
      }
      this.maxSizeAudio = option.config.maxAsrFileSize;
      this.$nextTick(() => {
        this.verifyASR();
      });
    },
    verifyASR() {
      if (
        this.fileList.some(file => file.size / 1024 / 1024 >= this.maxSizeAudio)
      ) {
        this.$message.warning(
          this.$t('knowledgeManage.multiKnowledgeDatabase.audioSizeLimit', {
            maxSize: this.maxSizeAudio,
          }),
        );
        this.confirmFlag = false;
      } else if (
        this.ruleForm.asrModelId ||
        this.parseMode === PARSE_MODE_TEMPLATE
      ) {
        this.confirmFlag = true;
      }
      this.$forceUpdate();
    },
    handleSetData(data) {
      this.docInfoList = [];
      data.map(item => {
        this.docInfoList.push({
          docName: item.fileName,
          docSize: item.fileSize,
          docUrl: item.url,
          docType: 'url',
        });
      });
    },
    async downloadTemplate() {
      try {
        await authDownload(
          `${USER_API}/files/docs/url_import_template.xlsx`,
          'url_import_template.xlsx',
        );
      } catch (error) {
        this.$message.error(this.$t('knowledgeManage.fileDownloadFailed'));
      }
    },
    handleLoading(val, result) {
      this.urlLoading = val;
      if (result === 'success') {
        this.reset();
      }
    },
    reset() {
      if (this.source.length > 0) {
        for (const sourceItem of this.source) {
          sourceItem.cancel();
        }
      }
      let ids = [];
      if (this.fileList.length > 0) {
        this.fileList.map(item => {
          if (item.id) {
            if (item.id.includes(',')) {
              //rag一体机没有此逻辑
              const list = item.id.split(',');
              list.map(item => {
                ids.push(item);
              });
            } else {
              ids.push(item.id);
            }
          }
        });
        if (ids.length > 0) {
          this.deleteData({ id: ids }); //取消时删除文件
        }
      }
      this.$refs['uplodForm'].resetFields();
      this.uplodForm.knowValue = null;
      this.fileList = [];
      this.resultDisabled = true;
      this.source = [];
      this.fileUuid = '';
      this.$emit('handleSetOpen', { isShow: false, knowValue: null });
      this.uploading = false;
    },
    // 删除已上传文件
    handleRemove(item, index) {
      if (item.percentage < 100) {
        this.fileList.splice(index, 1);
        this.cancelAndRestartNextRequests();
        return;
      }
      this.delfile({
        fileList: [this.resList[index]['name']],
        isExpired: true,
      });
      this.fileList = this.fileList.filter(files => files.name !== item.name);
      if (this.fileList.length === 0) {
        this.file = null;
      } else {
        this.fileIndex--;
      }
      if (this.docInfoList.length > 0) {
        this.docInfoList.splice(index, 1);
      }
    },
    delfile(data) {
      delfile(data).then(res => {
        if (res.code === 0) {
          this.$message.success(this.$t('common.info.delete'));
        }
      });
    },
    filterSize,
    fileTypeChange() {
      // 取消所有正在进行的上传请求
      this.cancelAllRequests();

      // 重置上传相关状态
      this.fileIndex = 0;
      this.file = null;
      this.resList = [];

      this.docInfoList = [];
      this.fileList = [];
    },
    // 知识库上已选定的解析模板：有绑定则默认走「使用模板」
    async getKnowledgeParseTemplate() {
      const res = await getDocDetail({
        knowledgeId: this.knowledgeId,
      }).catch(() => null);
      if (!res || res.code !== 0) return;
      this.ruleForm.parseTemplate = bindListToMap(res.data.parseTemplate);
      this.ruleFormBackup = structuredClone(this.ruleForm);
    },
    // 模板模式校验音频要用到 audio 模板和它引用的 ASR 模型，只有多模态知识库会有音频
    getAudioTemplates() {
      if (this.fileType !== 'fileMultiModal') return;
      getParseTemplateList({ docType: 'audio' }).then(res => {
        if (res.code === 0) this.audioTemplates = res.data.list || [];
      });
      selectASRList().then(res => {
        if (res.code === 0) this.asrOptions = res.data.list || [];
      });
    },
    // 模板模式没有 ASR 下拉，上限只能从绑定的模板反查，模板/文件变化时都要重算
    refreshTemplateAudioLimit() {
      if (this.parseMode !== PARSE_MODE_TEMPLATE) return;
      const model = this.templateAsrModel;
      this.maxSizeAudio = model ? model.config.maxAsrFileSize : 9999;
      this.verifyASR();
    },
    // 使用模板：解析参数由后端按文档类型套用模板，前端不再下发分段/解析配置
    submitWithTemplate() {
      if (!this.validateTemplateBind() || !this.validateMetaData()) {
        return;
      }
      this.ruleForm.docMetaData.forEach(item => {
        delete item.metadataType;
      });
      const data = {
        knowledgeId: this.knowledgeId,
        docImportType: this.fileType === 'fileUrl' ? 2 : 0,
        docInfoList: this.docInfoList,
        docMetaData: this.ruleForm.docMetaData,
        useTemplate: true,
        parseTemplate: bindMapToList(this.ruleForm.parseTemplate),
        overrideTemplate: this.isSystemAdmin && this.ruleForm.overrideTemplate,
      };
      docImport(data).then(res => {
        if (res.code === 0) {
          this.$router.push({
            path: `/knowledge/doclist/${this.knowledgeId}`,
            query: { name: this.knowledgeName, done: 'fileUpload' },
          });
        }
      });
    },
    // 内置模板也能被清空，凡是本次传了文件又没模板的类型都得先补上
    validateTemplateBind() {
      const missing = this.missingDocTypes;
      if (!missing.length) return true;
      this.$msgbox({
        title: this.$t(
          'knowledgeManage.parseTemplate.mediaTemplateMissingTitle',
        ),
        message: this.missingTemplateMessage(missing),
        confirmButtonText: this.$t('common.button.confirm'),
        customClass: 'media-template-box',
        type: 'warning',
      })
        // 关掉弹窗就展开模板区，让标红的那几项直接落在视野里
        .finally(() => this.focusMissingTemplates());
      return false;
    },
    // 类型名做成带文件图标的 chip，和下方网格用同一套图标，一眼对得上
    missingTemplateMessage(missing) {
      const h = this.$createElement;
      const chips = missing.map(docType => {
        const item = DOC_TYPE_LIST.find(d => d.docType === docType) || {};
        return h('span', { class: 'type-chip', key: docType }, [
          h(FileIcon, { props: { type: item.icon, size: '16px' } }),
          h('span', { class: 'chip-name' }, item.name),
        ]);
      });
      return h('div', { class: 'missing-body' }, [
        h(
          'p',
          { class: 'missing-lead' },
          this.$t('knowledgeManage.parseTemplate.mediaTemplateMissingLead'),
        ),
        h('div', { class: 'type-chips' }, chips),
        h(
          'p',
          { class: 'missing-hint' },
          this.$t('knowledgeManage.parseTemplate.mediaTemplateMissingHint'),
        ),
      ]);
    },
    // 展开模板区并把标红的那几项滚进视野
    focusMissingTemplates() {
      const select = this.$refs.parseTemplateSelect;
      if (!select) return;
      select.expand();
      this.$nextTick(() => {
        const el = select.$el.querySelector('.template-item.is-error');
        if (el) el.scrollIntoView({ behavior: 'smooth', block: 'center' });
      });
    },
    submitInfo() {
      if (this.mode !== 'config' && !this.validateFiles()) {
        return;
      }
      if (this.parseMode === PARSE_MODE_TEMPLATE) {
        this.submitWithTemplate();
        return;
      }
      // 模板字段只在模板模式提交，手动配置入口剥掉；payload 在副本上组装，不污染表单状态
      const data = { ...this.ruleForm };
      delete data.parseTemplate;
      delete data.overrideTemplate;
      if (this.ruleForm.docSegment.segmentType === '1')
        this.ruleForm.docSegment.segmentMethod = '0';
      const { segmentMethod, segmentType, splitter, subSplitter } =
        this.ruleForm.docSegment;
      this.$refs.ruleForm.validate(valid => {
        if (!valid) {
          return false;
        }
        if (
          (segmentMethod === '1' &&
            (splitter.length === 0 || subSplitter.length === 0)) ||
          (segmentMethod !== '1' &&
            segmentType === '1' &&
            splitter.length === 0)
        ) {
          this.$refs.ruleForm.validate();
          return false;
        }
        this.$refs.ruleForm.clearValidate([
          'docSegment.splitter',
          'docSegment.subSplitter',
        ]);
        if (!this.validateMetaData()) {
          return false;
        }
        this.ruleForm.docMetaData.forEach(item => {
          delete item.metadataType;
        });

        if (this.fileType === 'file' || this.fileType === 'fileMultiModal') {
          data.docImportType = 0;
        } else if (this.fileType === 'fileUrl') {
          data.docImportType = 2;
        } else {
          data.docImportType = 1;
        }

        data.docInfoList = this.docInfoList;
        if (this.ruleForm.asrModelId) data.docAnalyzer.push('asr');
        if (this.ruleForm.multimodalModelId)
          data.docAnalyzer.push('multimodal');
        if (
          this.ruleForm.docSegment.segmentType === '0' &&
          this.ruleForm.docSegment.segmentMethod !== '1'
        ) {
          delete data.docSegment.splitter;
          delete data.docSegment.maxSplitter;
          delete data.docSegment.overlap;
        }

        if (this.mode === 'config') {
          data.docIdList = this.docIdList;
          updateDocConfig(data).then(res => {
            if (res.code === 0) {
              this.$router.push({
                path: `/knowledge/doclist/${this.knowledgeId}`,
                query: { name: this.knowledgeName, done: 'fileUpload' },
              });
            }
          });
        } else
          docImport(data).then(res => {
            if (res.code === 0) {
              this.$router.push({
                path: `/knowledge/doclist/${this.knowledgeId}`,
                query: { name: this.knowledgeName, done: 'fileUpload' },
              });
            }
          });
      });
    },
    formReset() {
      this.ruleForm = structuredClone(this.ruleFormBackup);
      this.confirmFlag = !(
        this.fileFormatSet.has('audio') && !this.ruleForm.asrModelId
      );
      this.$refs.parseConfig?.reset();
      this.$refs.ruleForm.clearValidate();
    },
    uploadOnChange(file, fileList) {
      if (!fileList.length) return;
      // 先进行验证
      const isValid =
        this.verifyEmpty(file) &&
        this.verifyFormat(file) &&
        this.verifyRepeat(file);

      if (!isValid) return;

      this.fileList.push(file);
      setTimeout(() => {
        this.fileList.map((file, index) => {
          if (file.progressStatus && file.progressStatus !== 'success') {
            this.$set(file, 'progressStatus', 'exception');
            this.$set(file, 'showRetry', 'false');
            this.$set(file, 'showResume', 'false');
            this.$set(file, 'showRemerge', 'false');
            if (file.size > this.maxSizeBytes) {
              this.$set(file, 'fileType', 'maxFile');
            } else {
              this.$set(file, 'fileType', 'minFile');
            }
          }
        });
      }, 10);

      // 开始切片上传(如果没有文件正在上传)
      if (this.file === null) {
        this.startUpload();
      } else if (this.file.progressStatus === 'success') {
        // 如果上传当中有新的文件加入
        this.startUpload(this.fileIndex);
      }
    },
    refreshFile(index) {
      //重新上传文件
      this.fileList[index]['showRetry'] = 'false';
      this.fileList[index]['percentage'] = 0;
      this.startUpload(index);
    },
    resumeFile(index) {
      //续传文件
      this.fileList[index]['showResume'] = 'false';
      this.nextChunkIndex = this.uploadedChunks;
      this.processNextChunk();
    },
    remergeFile(index) {
      //重新上传
      this.mergeChunks();
    },
    uploadFile(fileName, oldName) {
      let type = oldName.split('.').pop();
      const docType =
        type === 'gz' ? '.tar.gz' : '.' + oldName.split('.').pop();
      this.docInfoList.push({
        docId: fileName,
        docName: oldName,
        docSize: this.fileList[this.fileIndex].size,
        docType,
      });
      this.fileIndex++;
      if (this.fileIndex < this.fileList.length) {
        this.startUpload(this.fileIndex);
      }
    },
    //  验证文件为空
    verifyEmpty(file) {
      if (file.size <= 0) {
        this.$message.warning(
          file.name + this.$t('knowledgeManage.filterFile'),
        );
        return false;
      }
      return true;
    },
    //  验证文件格式
    verifyFormat(file) {
      let nameType;
      if (this.fileType === 'fileMultiModal') {
        nameType = this.uploadLimitList.flatMap(item => item.extList || []);
      } else if (this.fileType === 'fileUrl') {
        // URL文件上传只允许xlsx格式
        nameType = ['xlsx'];
      } else {
        nameType = [
          'pdf',
          'docx',
          'doc',
          'pptx',
          'zip',
          'tar.gz',
          'xlsx',
          'xls',
          'csv',
          'txt',
          'html',
          'md',
          'ofd',
          'wps',
        ];
      }
      const fileName = file.name;
      const isSupportedFormat = nameType.some(ext =>
        fileName.endsWith(`.${ext}`),
      );
      if (!isSupportedFormat) {
        this.$message.warning(
          file.name + this.$t('knowledgeManage.fileTypeError'),
        );
        return false;
      }

      const fileType = file.name.split('.').pop();
      for (const uploadLimit of this.uploadLimitList) {
        const extList = uploadLimit.extList || [];
        if (extList.includes(fileType)) {
          if (uploadLimit.fileType !== 'audio') {
            if (file.size / 1024 / 1024 >= uploadLimit.maxSize) {
              this.$message.error(
                this.$t(
                  `knowledgeManage.multiKnowledgeDatabase.${uploadLimit.fileType}SizeLimit`,
                  { maxSize: uploadLimit.maxSize },
                ),
              );
              return false;
            }
          }
          return true;
        }
      }
      const limit200 = [
        'pdf',
        'docx',
        'doc',
        'pptx',
        'zip',
        'tar.gz',
        'ofd',
        'wps',
      ];
      const limit20 = ['xlsx', 'xls', 'csv', 'txt', 'html', 'md'];

      if (limit200.includes(fileType) && file.size / 1024 / 1024 >= 200) {
        this.$message.error(this.$t('knowledgeManage.limitSize') + '200MB!');
        return false;
      }

      if (limit20.includes(fileType) && file.size / 1024 / 1024 >= 20) {
        this.$message.error(this.$t('knowledgeManage.limitSize') + '20MB!');
        return false;
      }
      return true;
    },
    //  验证文件重复
    verifyRepeat(file) {
      const isDuplicate = this.fileList.some(item => item.name === file.name);

      if (isDuplicate) {
        this.$message.warning(file.name + this.$t('knowledgeManage.fileExist'));
        return false;
      }
      return true;
    },
    // 提交前的文件校验（原「下一步」的校验，单页后并入提交）
    validateFiles() {
      this.withCompressed = this.fileList.some(file => {
        const fileName = file.name;
        return fileName.endsWith('.zip') || fileName.endsWith('.tar.gz');
      });
      //上传文件类型
      if (
        this.fileType === 'file' ||
        this.fileType === 'fileUrl' ||
        this.fileType === 'fileMultiModal'
      ) {
        if (this.fileIndex < this.fileList.length) {
          this.$message.warning('文件上传中...');
          return false;
        }
        if (this.fileList.length === 0) {
          this.$message.warning('请上传文件!');
          return false;
        }
      }
      //url逐条上传
      if (this.fileType === 'url') {
        if (this.docInfoList.length === 0) {
          this.$message.warning('请上输入url!');
          return false;
        }
      }
      return true;
    },
  },
};
</script>
<style lang="scss" scoped>
.red {
  color: red;
}

.width100 {
  width: 100%;
}

.activeAnalyzer {
  border-color: $color !important;
}

.question {
  cursor: pointer;
  color: #aaadcc;
  margin-left: 5px;
}

.splitterTag {
  margin-right: 10px;
  border: none;
  background: $color_opacity;
  color: $color;
  border-radius: 3px;
}

.optionInput {
  width: 90%;
  margin: 10px;
}

.splitterOption {
  margin-top: 5px;
}

.el-input-number {
  line-height: 28px !important;
}

::v-deep .el-input-number.is-controls-right .el-input-number__decrease,
::v-deep .el-input-number.is-controls-right .el-input-number__increase {
  line-height: 14px !important;
  border: 0;
}

::v-deep {
  .el-upload {
    width: 100%;
  }

  .el-upload-dragger {
    width: 100%;
  }
}

.fileUpload {
  width: 80%;
  padding-top: 30px;
  margin: 0 auto;

  .fileBtn {
    padding: 20px 0 15px 0;
    display: flex;
    justify-content: center;
  }

  .dialog-body {
    padding: 0;
    width: 100%;

    .upload-title {
      text-align: center;
      font-size: 18px;
      margin-bottom: 20px;
    }

    .upload-box {
      height: auto;
      min-height: 280px;
      width: 100% !important;
      display: flex;

      // ElUpload 在 .upload-box 和 dragger 之间还有一层 .el-upload，
      // 它不是 flex，dragger 撑不满就会贴顶
      // 全局给这两层写了 height:100%，但父级是 min-height，百分比高度解析不了，
      // 反而堵死 flex stretch；改回 auto 由 stretch 撑满，内容再上下居中
      ::v-deep .el-upload {
        display: flex;
        width: 100%;
        height: auto;
      }

      ::v-deep .el-upload-dragger {
        display: flex;
        flex-direction: column;
        justify-content: center;
        width: 100%;
        height: auto;
        overflow: visible;
      }

      .upload-img {
        width: 56px;
        height: 56px;
      }

      .click-text {
        margin-top: 10px;

        .clickUpload {
          color: $color;
          font-weight: bold;
        }
      }

      .el-upload-dragger {
        .el-icon-upload {
          margin: 46px 0 10px 0 !important;
          font-size: 32px !important;
          line-height: 36px !important;
          color: $color;
        }

        .el-upload__text {
          margin-top: -10px;
        }
      }

      .size {
        margin-right: 10px;
      }

      .file-size {
        margin-left: 10px;
      }
    }

    .echo-img-box {
      background-color: transparent !important;

      .echo-img {
        img,
        video {
          width: auto;
          height: 80px;
          margin: 10px auto;
          border-radius: 4px;
          background-color: transparent;
        }

        audio {
          width: 300px;
          height: 54px;
          margin: 50px auto;
        }
      }

      .docFile {
        img {
          margin: 0;
          width: 60px;
          height: 100px;
        }
      }
    }

    .tips {
      padding: 20px 20px;

      p {
        color: #9d8d8d !important;

        .template_downLoad {
          color: $color;
          cursor: pointer;
        }
      }
    }
  }

  .el-upload-url {
    width: 100%;
    padding: 0 20px;

    .upload-url {
      background-color: #fff;
      border: 1px solid #d4d6d9;
      border-radius: 6px;
      height: 100%;
      box-sizing: border-box;
      text-align: center;
      cursor: pointer;
      overflow: hidden;
      padding: 20px;
    }

    .upload-url:hover {
      border-color: $color;
    }
  }
}

.next {
  padding: 20px;
  display: flex;
  justify-content: flex-end;
}

.meta-section {
  margin-top: 24px;
}

.optional-tip {
  color: #999;
  font-weight: 400;
}

.meta-label .question {
  font-weight: 400;
}

.parse-mode-list {
  display: flex;
  gap: 16px;

  .parse-mode-item {
    flex: 1 1 0;
    padding: 14px 20px;
    border: 1px solid #ddd;
    border-radius: 6px;
    background: #fff;
    cursor: pointer;

    .analyzerItem_text {
      margin: 0;
      font-size: 14px;
      font-weight: 600;
      line-height: 1.6;
    }

    .analyzerItem_desc {
      margin: 4px 0 0;
      font-size: 12px;
      font-weight: unset;
      line-height: 1.5;
      color: #b4b3b3;
    }
  }
}

// 方案 A：这个框的标题拉开一档字重，和框内摘要形成主次
.parse-template-item ::v-deep > .el-form-item__label {
  font-weight: 500;
  color: #303133;
}

.section-title {
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.4;
  color: #333;
}

.upload-section {
  margin-bottom: 24px;
}

.params_form {
  margin-top: 16px;
  background: #fff;
  border: 1px solid #d4d6d9;
  border-radius: 6px;

  .el-form {
    padding: 20px 24px;

    // 收起态只有一行字，别让框空得太夸张
    > .el-form-item:last-child {
      margin-bottom: 0;
    }

    .commonSet {
      background: #f6f7fe;
      padding: 15px;
      border-radius: 6px;

      .title {
        font-size: 14px;
        border-bottom: 1px solid #dee1fe;
        padding-bottom: 10px;
        display: flex;
        align-items: center;

        .bar {
          display: inline-block;
          width: 4px;
          height: 14px;
          background: $color;
          margin-right: 5px;
        }
      }
    }

    .commonSet:nth-child(1) {
      margin-bottom: 15px;
    }

    .el-form-item {
      display: flex;
      flex-direction: column;

      ::v-deep {
        .el-form-item__content {
          margin: 0 !important;
        }
      }
    }

    .el-checkbox-group,
    .radioGroup {
      display: flex;
      justify-content: flex-start;
      gap: 15px;

      .docAnalyzerList {
        flex: 1;
        border: 1px solid #ddd;
        padding: 0 10px 10px 10px;
        border-radius: 6px;
        cursor: pointer;

        .analyzerItem_desc {
          display: block;
          color: #b4b3b3;
          font-size: 12px;
          font-weight: unset;
          line-height: 1;
        }
      }
    }

    .segmentList {
      display: flex;
      gap: 15px;

      .segmentItem {
        display: flex;
        align-items: center;
        cursor: pointer;
        border: 1px solid #ddd;
        padding: 10px;
        border-radius: 6px;
        gap: 15px;
        width: 50%;

        .itemImg {
          width: 45px;
          height: 45px;
          border: 1px solid #eeeded;
          border-radius: 8px;
          display: flex;
          justify-content: center;
          align-items: center;
          box-shadow: 0px 2px 4px -2px rgba(16, 24, 40, 0.06);

          img {
            width: 25px;
            height: fit-content;
          }
        }

        .analyzerItem_text {
          font-size: 14px;
          font-weight: 600;
          line-height: 1.8;
        }

        .analyzerItem_desc {
          line-height: 1.2;
          color: #b4b3b3;
          font-weight: unset;
        }
      }
    }
  }
}

.page-title {
  .back {
    font-size: 18px;
    margin-right: 10px;
    cursor: pointer;
  }
}

.file-list {
  // 上间距由 .upload-section 的 24px 给；容器自带 4px 底部内边距，这里补到 24px
  margin-bottom: 20px;

  $file-item-height: 40px;
  $file-item-gap: 8px;

  // 一屏最多展示 10 个文档，超出滚动
  .document_lise {
    max-height: ($file-item-height + $file-item-gap) * 10;
    overflow-y: auto;
    margin: 0;
    // overflow 容器永远裁切，右侧和底部各留一点给阴影
    padding: 0 4px 4px 0;
  }

  .document_lise_item {
    cursor: pointer;
    height: $file-item-height;
    box-sizing: border-box;
    padding: 5px 10px;
    list-style: none;
    background: #fff;
    border-radius: 6px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    display: flex;
    align-items: center;
    margin-bottom: $file-item-gap;

    &:last-child {
      margin-bottom: 0;
    }

    .lise_item_box {
      width: 100%;
      display: flex;
      align-items: center;
      justify-content: space-between;

      .size {
        display: flex;
        align-items: center;

        .progress {
          width: 400px;
          margin-left: 30px;
        }

        img {
          width: 18px;
          height: 18px;
          margin-bottom: -3px;
        }

        .file-size {
          margin-left: 10px;
        }
      }
    }
  }

  .document_lise_item:hover {
    background: #eceefe;
  }
}

.table-opera-icon {
  font-size: 18px;
}
</style>

<style lang="scss">
// MessageBox 挂在 body 上，样式不能写在 scoped 块里
.media-template-box {
  width: 420px;
  border-radius: 10px;
  box-shadow: 0 8px 32px rgba(31, 35, 41, 0.12);

  .el-message-box__header {
    padding: 20px 24px 0;
  }

  .el-message-box__title {
    padding-left: 30px;
    font-size: 16px;
    font-weight: 600;
    line-height: 24px;
    color: #1f2329;
  }

  .el-message-box__headerbtn {
    top: 18px;
    right: 18px;
    font-size: 15px;
  }

  // 图标提到标题行，-31px 是量出来的：内容区顶到标题中线的距离
  .el-message-box__status {
    top: -31px;
    left: 0;
    font-size: 20px !important;
    transform: none;
  }

  .el-message-box__content {
    padding: 8px 24px 0;
  }

  .el-message-box__message {
    padding-left: 30px;
    // 全局带 status 时给了 padding-right:12px，正文右边会比按钮短一截
    padding-right: 0;
    color: #5c6270;
  }

  .missing-lead {
    margin: 0;
    font-size: 14px;
    line-height: 22px;
  }

  .type-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin: 12px 0;
  }

  .type-chip {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 5px 10px;
    border: 1px solid #ebedf0;
    border-radius: 6px;
    background: #f7f8fa;
    font-size: 13px;
    font-weight: 600;
    line-height: 18px;
    color: #1f2329;
  }

  // #8a9099 在 13px 上只有 3.2:1，不够 WCAG AA 的 4.5:1
  .missing-hint {
    margin: 0;
    font-size: 13px;
    line-height: 20px;
    color: #6b7280;
  }

  // 全局给这个选择器加了 !important，只能同级压回去，否则按钮右边缘和正文差 9px
  .el-message-box__btns {
    padding: 18px 24px 16px !important;
  }

  .el-button--primary {
    min-width: 76px;
    padding: 9px 18px;
    font-size: 14px;
    letter-spacing: 0;
  }
}
</style>
