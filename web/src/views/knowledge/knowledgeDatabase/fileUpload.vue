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
        <el-steps :active="active" class="fileStep" align-center>
          <el-step
            :title="
              $t('knowledgeManage.knowledgeDatabase.fileUpload.fileUpload')
            "
          ></el-step>
          <el-step
            :title="
              $t('knowledgeManage.knowledgeDatabase.fileUpload.paramSetting')
            "
          ></el-step>
        </el-steps>

        <!-- 文件上传 -->
        <div v-if="active === 1">
          <div class="fileBtn">
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
              <el-radio-button label="fileUrl">
                {{ $t('knowledgeManage.knowledgeDatabase.fileUpload.fileUrl') }}
              </el-radio-button>
              <el-radio-button label="url">
                {{ $t('knowledgeManage.knowledgeDatabase.fileUpload.url') }}
              </el-radio-button>
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

        <!-- 参数设置 -->
        <div v-else class="params_form">
          <el-form
            :model="ruleForm"
            ref="ruleForm"
            label-width="140px"
            class="demo-ruleForm"
            @submit.native.prevent
            label-position="left"
          >
            <el-form-item
              :label="
                $t(
                  'knowledgeManage.knowledgeDatabase.fileUpload.segmentSetting',
                )
              "
            >
              <div class="segmentList">
                <div
                  v-for="segmentItem in segmentList"
                  :key="segmentItem.text"
                  :class="[
                    'segmentItem',
                    ruleForm.docSegment.segmentMethod === segmentItem.label
                      ? 'activeAnalyzer'
                      : '',
                  ]"
                  style="width: 50%"
                  @click="segmentSetClick(segmentItem.label)"
                >
                  <div class="itemImg">
                    <img :src="require(`@/assets/imgs/${segmentItem.img}`)" />
                  </div>
                  <div>
                    <p class="analyzerItem_text">{{ segmentItem.text }}</p>
                    <h3 class="analyzerItem_desc">{{ segmentItem.desc }}</h3>
                  </div>
                </div>
              </div>
            </el-form-item>
            <template v-if="this.ruleForm.docSegment.segmentMethod === '0'">
              <el-form-item :label="$t('knowledgeManage.chunkTypeSet')">
                <div class="segmentList">
                  <div
                    v-for="segmentCommon in segmentCommonList"
                    :key="segmentCommon.text"
                    :class="[
                      'segmentItem',
                      ruleForm.docSegment.segmentType === segmentCommon.label
                        ? 'activeAnalyzer'
                        : '',
                    ]"
                    @click="segmentClick(segmentCommon.label)"
                  >
                    <div>
                      <p class="analyzerItem_text">{{ segmentCommon.text }}</p>
                      <h3 class="analyzerItem_desc">
                        {{ segmentCommon.desc }}
                      </h3>
                    </div>
                  </div>
                </div>
              </el-form-item>
              <el-form-item
                v-if="ruleForm.docSegment.segmentType === '1'"
                prop="docSegment.splitter"
                :rules="
                  ruleForm.docSegment.segmentType === '1'
                    ? [
                        {
                          required: true,
                          validator: validateSplitter('splitter'),
                          message: $t('knowledgeManage.markTips'),
                          trigger: 'blur',
                        },
                      ]
                    : []
                "
              >
                <template #label>
                  <span>
                    {{
                      $t(
                        'knowledgeManage.knowledgeDatabase.fileUpload.segmentTips',
                      )
                    }}
                  </span>
                  <el-tooltip
                    :content="$t('knowledgeManage.splitOptionsTips')"
                    placement="right"
                  >
                    <span class="el-icon-question question"></span>
                  </el-tooltip>
                </template>
                <el-tag
                  v-for="(tag, index) in checkSplitter['splitter']"
                  :key="'tag' + index"
                  :disable-transitions="false"
                  class="splitterTag"
                >
                  {{ tag.splitterName.replace(/\n/g, '\\n') }}
                </el-tag>
                <el-button
                  class="button-new-tag"
                  size="small"
                  @click="showSplitterSet('splitter')"
                >
                  {{
                    $t(
                      'knowledgeManage.knowledgeDatabase.fileUpload.segmentTipsSetting',
                    )
                  }}
                </el-button>
              </el-form-item>
              <el-form-item
                v-if="ruleForm.docSegment.segmentType === '1'"
                prop="docSegment.maxSplitter"
                :rules="[
                  {
                    required: true,
                    message: $t('knowledgeManage.splitMax'),
                    trigger: 'blur',
                  },
                  {
                    type: 'number',
                    min: 200,
                    max: 4000,
                    message: $t('knowledgeManage.splitMaxMsg'),
                    trigger: 'blur',
                  },
                ]"
              >
                <template #label>
                  <span>{{ $t('knowledgeManage.splitMax') }}</span>
                  <el-tooltip
                    :content="$t('knowledgeManage.splitMaxTips')"
                    placement="right"
                  >
                    <span class="el-icon-question question"></span>
                  </el-tooltip>
                </template>
                <div
                  :class="[
                    ['0', '1', '3', '4'].includes(
                      ruleForm.docSegment.segmentType,
                    )
                      ? ''
                      : 'set',
                  ]"
                >
                  <el-input
                    type="number"
                    v-model.number="ruleForm.docSegment.maxSplitter"
                    :placeholder="$t('knowledgeManage.splitMax')"
                  ></el-input>
                </div>
              </el-form-item>
              <el-form-item
                v-if="ruleForm.docSegment.segmentType === '1'"
                :label="$t('knowledgeManage.overLapNum')"
                prop="docSegment.overlap"
                :rules="[
                  {
                    required: true,
                    message: $t('knowledgeManage.overLapNumTips'),
                    trigger: 'blur',
                  },
                  {
                    type: 'number',
                    min: 0,
                    max: 1,
                    message: $t('knowledgeManage.overLapNumMsg'),
                    trigger: 'blur',
                  },
                ]"
              >
                <el-input
                  :min="0"
                  :max="0.25"
                  :step="0.01"
                  type="number"
                  v-model.number="ruleForm.docSegment.overlap"
                  :placeholder="$t('knowledgeManage.overLapNumPlaceholder')"
                ></el-input>
              </el-form-item>
            </template>
            <template v-if="this.ruleForm.docSegment.segmentMethod === '1'">
              <div
                v-for="item in fatSonBlock"
                :key="item.level"
                class="commonSet"
              >
                <h3 class="title">
                  <span class="bar"></span>
                  {{ item.title }}
                </h3>
                <el-form-item
                  :prop="item.splitterProp"
                  :rules="[
                    {
                      required: true,
                      validator: validateSplitter(item.key),
                      message: $t('knowledgeManage.markTips'),
                      trigger: 'blur',
                    },
                  ]"
                >
                  <template #label>
                    <span>
                      {{
                        $t(
                          'knowledgeManage.knowledgeDatabase.fileUpload.segmentTips',
                        )
                      }}
                    </span>
                    <el-tooltip
                      :content="$t('knowledgeManage.splitOptionsTips')"
                      placement="right"
                    >
                      <span class="el-icon-question question"></span>
                    </el-tooltip>
                  </template>
                  <el-tag
                    v-for="(tag, index) in checkSplitter[item.key]"
                    :key="'tag' + index"
                    :disable-transitions="false"
                    class="splitterTag"
                  >
                    {{ tag.splitterName.replace(/\n/g, '\\n') }}
                  </el-tag>
                  <el-button
                    class="button-new-tag"
                    size="small"
                    @click="showSplitterSet(item.key)"
                  >
                    {{
                      $t(
                        'knowledgeManage.knowledgeDatabase.fileUpload.segmentTipsSetting',
                      )
                    }}
                  </el-button>
                </el-form-item>
                <el-form-item
                  :prop="item.maxSplitterProp"
                  :rules="[
                    {
                      required: true,
                      message: $t('knowledgeManage.splitMax'),
                      trigger: 'blur',
                    },
                    {
                      type: 'number',
                      min: 200,
                      max: item.maxSplitterNum,
                      message: $t('knowledgeManage.splitMaxMsg'),
                      trigger: 'blur',
                    },
                  ]"
                >
                  <template #label>
                    <span>{{ $t('knowledgeManage.splitMax') }}</span>
                    <el-tooltip
                      :content="$t('knowledgeManage.splitMaxTips')"
                      placement="right"
                    >
                      <span class="el-icon-question question"></span>
                    </el-tooltip>
                  </template>
                  <el-input
                    type="number"
                    :min="200"
                    :max="item.maxSplitterNum"
                    v-model.number="ruleForm.docSegment[item.maxSplitter]"
                    :placeholder="$t('knowledgeManage.splitMax')"
                    @change="maxSplitterChange(item)"
                  ></el-input>
                </el-form-item>
              </div>
            </template>
            <el-form-item
              :label="$t('knowledgeManage.textPreprocessing')"
              prop="docPreprocess"
              v-if="
                ruleForm.docSegment.segmentType === '1' ||
                ruleForm.docSegment.segmentMethod === '1'
              "
            >
              <el-checkbox-group v-model="ruleForm.docPreprocess">
                <el-checkbox label="replaceSymbols">
                  {{ $t('knowledgeManage.replaceSymbols') }}
                </el-checkbox>
                <el-checkbox label="deleteLinks">
                  {{ $t('knowledgeManage.deleteLinks') }}
                </el-checkbox>
              </el-checkbox-group>
            </el-form-item>
            <el-form-item
              :label="$t('knowledgeManage.parsingMethod')"
              v-if="fileType === 'fileMultiModal'"
            >
              <div
                class="segmentList"
                v-if="fileFormatSet.has('video') || fileFormatSet.has('audio')"
              >
                <span style="display: inline-block; width: 100px">
                  <span class="red" v-if="fileFormatSet.has('audio')">*</span>
                  ASR
                </span>
                <modelSelect
                  v-model="ruleForm.asrModelId"
                  :options="asrOptions"
                  clearable
                  @change="handleASR"
                />
              </div>
              <div
                class="segmentList"
                v-if="fileFormatSet.has('video') || fileFormatSet.has('image')"
              >
                <span style="display: inline-block; width: 100px">
                  {{ $t('knowledgeManage.config.visionModal') }}
                </span>
                <modelSelect
                  v-model="ruleForm.multimodalModelId"
                  :options="visionOptions"
                  clearable
                />
              </div>
            </el-form-item>
            <el-form-item
              :label="$t('knowledgeManage.parsingMethod')"
              prop="docAnalyzer"
              v-else
            >
              <el-checkbox-group
                v-model="ruleForm.docAnalyzer"
                @change="docAnalyzerChange($event)"
              >
                <div
                  v-for="analyzerItem in docAnalyzerList"
                  :class="[
                    'docAnalyzerList',
                    ruleForm.docAnalyzer.includes(analyzerItem.label)
                      ? 'activeAnalyzer'
                      : '',
                  ]"
                >
                  <el-checkbox
                    :label="analyzerItem.label"
                    :disabled="analyzerDisabled(analyzerItem.label)"
                  >
                    {{ analyzerItem.text }}
                  </el-checkbox>
                  <h3 class="analyzerItem_desc">{{ analyzerItem.desc }}</h3>
                </div>
              </el-checkbox-group>
            </el-form-item>
            <el-form-item
              prop="parserModelId"
              v-if="
                ruleForm.docAnalyzer.includes('ocr') ||
                ruleForm.docAnalyzer.includes('model')
              "
              :rules="[
                {
                  required: true,
                  message: $t('knowledgeManage.parsingMethodMsg'),
                  trigger: 'blur',
                },
              ]"
            >
              <template #label>
                <span>
                  {{ modelTypeTip[ruleForm.docAnalyzer[1]]['label'] }}
                </span>
                <el-tooltip
                  :content="modelTypeTip[ruleForm.docAnalyzer[1]]['desc']"
                  placement="right"
                >
                  <span class="el-icon-question question"></span>
                </el-tooltip>
              </template>
              <el-select
                v-model="ruleForm.parserModelId"
                :placeholder="$t('common.select.placeholder')"
                class="width100"
              >
                <el-option
                  v-for="item in modelOptions"
                  :key="item.modelId"
                  :label="item.displayName"
                  :value="item.modelId"
                ></el-option>
              </el-select>
            </el-form-item>
            <el-form-item prop="docAnalyzer" v-if="mode !== 'config'">
              <template #label>
                <span>
                  {{ $t('knowledgeManage.metadataManagement') }}
                </span>
                <el-tooltip
                  :content="$t('knowledgeManage.metadataManagementTips')"
                  placement="right"
                >
                  <span class="el-icon-question question"></span>
                </el-tooltip>
              </template>
              <mataData
                ref="mataData"
                @updateMeta="updateMeta"
                :knowledgeId="knowledgeId"
                :withCompressed="withCompressed"
              />
            </el-form-item>
          </el-form>
        </div>
        <!-- 上传文件的列表 -->
        <div class="file-list" v-if="fileList.length > 0 && active === 1">
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
        <div class="next">
          <el-button
            type="primary"
            size="mini"
            @click="preStep"
            v-if="active === 2 && mode !== 'config'"
          >
            {{ $t('knowledgeManage.prevStep') }}
          </el-button>
          <el-button
            type="primary"
            size="mini"
            @click="nextStep"
            v-if="active === 1"
            :loading="urlLoading"
          >
            {{ $t('knowledgeManage.nextStep') }}
          </el-button>
          <el-button
            type="primary"
            size="mini"
            @click="submitInfo"
            v-if="active === 2"
            :disabled="!confirmFlag"
          >
            {{ $t('common.button.confirm') }}
          </el-button>
          <el-button size="mini" @click="formReset" v-if="active === 2">
            {{ $t('common.button.restore') }}
          </el-button>
        </div>
      </div>
    </div>
    <splitterDialog
      ref="splitterDialog"
      :title="titleText"
      :placeholderText="placeholderText"
      :dataList="splitOptions"
      @editItem="editItem"
      @createItem="createItem"
      @delItem="delSplitterItem"
      @reloadData="reloadData"
      @checkData="checkData"
    />
  </div>
</template>
<script>
import urlAnalysis from '../component/urlAnalysis.vue';
import uploadChunk from '@/mixins/uploadChunk';
import {
  docImport,
  ocrSelectList,
  delSplitter,
  getSplitter,
  createSplitter,
  editSplitter,
  parserSelect,
  updateDocConfig,
  getDocConfig,
  getDocList,
  getDocLimit,
} from '@/api/knowledge';
import { delfile } from '@/api/chunkFile';
import LinkIcon from '@/components/linkIcon.vue';
import splitterDialog from '../component/splitterDialog.vue';
import mataData from '../component/metadata.vue';
import { USER_API } from '@/utils/requestConstants';
import {
  SEGMENT_COMMON_LIST,
  SEGMENT_LIST,
  DOC_ANALYZER_LIST,
  FAT_SON_BLOCK,
  MODEL_TYPE_TIP,
} from '../config';
import { deepMerge, filterSize, authDownload } from '@/utils/util';
import modelSelect from '@/components/modelSelect.vue';
import { selectASRList, selectModelList } from '@/api/modelAccess';

export default {
  name: 'KnowledgeFileUpload',
  components: { modelSelect, LinkIcon, urlAnalysis, splitterDialog, mataData },
  mixins: [uploadChunk],
  data() {
    const validateSplitter = type => {
      return (rule, value, callback) => {
        if (this.checkSplitter[type].length === 0) {
          callback(new Error(this.$t('knowledgeManage.splitterRequired')));
        } else {
          callback();
        }
      };
    };
    return {
      validateSplitter: validateSplitter,
      placeholderText: this.$t('knowledgeManage.placeholderText'),
      titleText: this.$t('knowledgeManage.titleText'),
      splitterValue: '',
      tableData: [],
      modelOptions: [],
      asrOptions: [],
      visionOptions: [],
      uploadLimitList: [],
      maxSizeAudio: 9999,
      confirmFlag: true,
      urlValidate: false,
      active: this.$route.query.mode === 'config' ? 2 : 1,
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
      segmentType: '',
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
      },
      ruleFormBackup: {},
      checkSplitter: {
        splitter: [],
        subSplitter: [],
      },
      splitOptions: [],
      urlLoading: false,
      segmentCommonList: SEGMENT_COMMON_LIST,
      segmentList: SEGMENT_LIST,
      docAnalyzerList: DOC_ANALYZER_LIST,
      fatSonBlock: FAT_SON_BLOCK,
      modelTypeTip: MODEL_TYPE_TIP,
    };
  },
  computed: {
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
    if (query.mode === 'config' && this.docIdList.length === 1) {
      await getDocConfig({
        docId: this.docIdList[0],
        knowledgeId: this.knowledgeId,
      }).then(res => {
        if (res.code === 0) {
          this.ruleForm = deepMerge(this.ruleForm, res.data);
          this.ruleFormBackup = structuredClone(this.ruleForm);
          this.ruleForm.docAnalyzer = [...this.ruleForm.docAnalyzer];
          this.getModelOptions();
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
      selectASRList().then(res => {
        if (res.code === 0) {
          this.asrOptions = res.data.list || [];
        }
      });
      selectModelList().then(res => {
        if (res.code === 0) {
          this.visionOptions = (res.data.list || []).filter(
            item => item.config.visionSupport === 'support',
          );
        }
      });
    }
    await this.getSplitterList('');
    await this.custom();
  },
  methods: {
    getModelOptions() {
      this.getOcrList();
    },
    maxSplitterChange(item) {
      if (item.level === 'parent') {
        const parentMaxValue = this.ruleForm.docSegment.maxSplitter;
        const sonBlock = this.fatSonBlock.find(block => block.level === 'son');
        if (sonBlock) {
          sonBlock.maxSplitterNum = parentMaxValue;
          if (this.ruleForm.docSegment.subMaxSplitter > parentMaxValue) {
            this.ruleForm.docSegment.subMaxSplitter = parentMaxValue;
            this.$message.warning(
              this.$t('knowledgeManage.childSegmentMaxAdjusted', {
                parentMaxValue,
              }),
            );
          }
        }
      } else if (item.level === 'son') {
        const sonMaxValue = this.ruleForm.docSegment.subMaxSplitter;
        const parentMaxValue = this.ruleForm.docSegment.maxSplitter;
        if (sonMaxValue > parentMaxValue) {
          this.ruleForm.docSegment.subMaxSplitter = parentMaxValue;
          this.$message.warning(
            this.$t('knowledgeManage.childSegmentMaxAdjustedTips', {
              parentMaxValue,
            }),
          );
        }
      }
    },
    getParserList() {
      parserSelect()
        .then(res => {
          if (res.code === 0) {
            this.modelOptions = res.data.list || [];
          }
        })
        .catch(() => {});
    },
    docAnalyzerChange(val) {
      this.ruleForm.parserModelId = '';
      this.modelOptions = [];
      if (val.length === 3) {
        this.ruleForm.docAnalyzer = [val[0], val[2]];
      }
      this.getModelOptions();
    },
    segmentClick(label) {
      this.ruleForm.docSegment.segmentType = label;
    },
    segmentSetClick(label) {
      this.ruleForm.docSegment.segmentMethod = label;
    },
    analyzerDisabled(label) {
      if (label === 'text') return true;
    },
    custom() {
      this.$nextTick(() => {
        const { splitter, subSplitter } = this.ruleForm.docSegment;
        const filterByType = values =>
          this.splitOptions.filter(item => values.includes(item.splitterValue));
        this.checkSplitter = {
          splitter: filterByType(splitter),
          subSplitter: filterByType(subSplitter),
        };
      });
    },
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
    checkData(data) {
      this.checkSplitter[this.segmentType] = data;
      this.ruleForm.docSegment[this.segmentType] = data.map(
        item => item.splitterValue,
      );
    },
    reloadData(name) {
      this.getSplitterList(name);
    },
    async getSplitterList(splitterName) {
      const res = await getSplitter({ splitterName });
      if (res.code === 0) {
        this.splitOptions = (res.data.knowledgeSplitterList || []).map(
          item => ({
            ...item,
            showDel: false,
            showIpt: false,
          }),
        );
      }
    },
    editItem(item) {
      editSplitter({
        splitterId: item.splitterId,
        splitterName: item.splitterName,
        splitterValue: item.splitterName,
      }).then(res => {
        if (res.code === 0) {
          item.showIpt = false;
          this.getSplitterList('');
        }
      });
    },
    createItem(item) {
      createSplitter({
        splitterName: item.splitterName,
        splitterValue: item.splitterName,
      }).then(res => {
        if (res.code === 0) {
          item.showIpt = false;
          this.getSplitterList('');
        }
      });
    },
    async delSplitterItem(item) {
      this.$confirm(
        this.$t(
          'knowledgeManage.knowledgeDatabase.fileUpload.deleteSplitterConfirm',
          { splitterName: item.splitterName },
        ),
        this.$t(
          'knowledgeManage.knowledgeDatabase.fileUpload.deleteSplitterTitle',
        ),
        {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          type: 'warning',
        },
      )
        .then(async () => {
          const res = await delSplitter({ splitterId: item.splitterId });
          if (res.code === 0) {
            this.getSplitterList('');
          }
        })
        .catch(error => {
          this.getSplitterList('');
        });
    },
    showSplitterSet(type) {
      this.segmentType = type;
      this.$refs.splitterDialog.showDialog(
        this.checkSplitter[this.segmentType],
      );
    },
    goBack() {
      this.$router.go(-1);
    },
    getOcrList() {
      ocrSelectList().then(res => {
        if (res.code === 0) {
          this.modelOptions = res.data.list || [];
        }
      });
    },
    handleASR(value) {
      if (!value) {
        if (!this.fileFormatSet.has('audio')) this.confirmFlag = true;
        this.maxSizeAudio = 9999;
        return;
      }
      this.maxSizeAudio = this.asrOptions.find(
        option => option.modelId === value,
      ).config.maxAsrFileSize;
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
      } else if (this.ruleForm.asrModelId) {
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
    submitInfo() {
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
          this.ruleForm.docImportType = 0;
        } else if (this.fileType === 'fileUrl') {
          this.ruleForm.docImportType = 2;
        } else {
          this.ruleForm.docImportType = 1;
        }

        this.ruleForm.docInfoList = this.docInfoList;
        if (this.ruleForm.asrModelId) this.ruleForm.docAnalyzer.push('asr');
        if (this.ruleForm.multimodalModelId)
          this.ruleForm.docAnalyzer.push('multimodal');
        let data = null;
        if (
          this.ruleForm.docSegment.segmentType === '0' &&
          this.ruleForm.docSegment.segmentMethod !== '1'
        ) {
          data = this.ruleForm;
          delete data.docSegment.splitter;
          delete data.docSegment.maxSplitter;
          delete data.docSegment.overlap;
        } else {
          data = this.ruleForm;
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
      this.checkSplitter = {
        splitter: [],
        subSplitter: [],
      };
      this.splitOptions = this.splitOptions.map(item => ({
        ...item,
        checked: false,
      }));
      this.confirmFlag = !(
        this.fileFormatSet.has('audio') && !this.ruleForm.asrModelId
      );
      this.getModelOptions();
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
    nextStep() {
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
      this.active = 2;
      if (this.fileType === 'fileMultiModal')
        this.ruleForm.docAnalyzer = ['text'];
      this.confirmFlag = !(
        this.fileFormatSet.has('audio') && !this.ruleForm.asrModelId
      );
      this.verifyASR();
    },
    preStep() {
      this.active = 1;
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

  .fileStep {
    width: 40%;
    margin: 0 auto;
  }

  .fileBtn {
    padding: 20px 0 15px 0;
    display: flex;
    justify-content: center;
  }

  .dialog-body {
    padding: 0 20px;
    width: 100%;

    .upload-title {
      text-align: center;
      font-size: 18px;
      margin-bottom: 20px;
    }

    .upload-box {
      height: auto;
      min-height: 190px;
      width: 100% !important;

      .upload-img {
        width: 56px;
        height: 56px;
        margin-top: 30px;
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

.params_form {
  margin-top: 10px;
  background: #fff;
  border: 1px solid #d4d6d9;
  border-radius: 6px;
  max-height: 65vh;
  overflow-y: auto;

  .el-form {
    padding: 30px;

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
  padding: 20px;

  .document_lise_item {
    cursor: pointer;
    padding: 5px 10px;
    list-style: none;
    background: #fff;
    border-radius: 4px;
    box-shadow: 1px 2px 2px #ddd;
    display: flex;
    align-items: center;
    margin-bottom: 10px;

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
