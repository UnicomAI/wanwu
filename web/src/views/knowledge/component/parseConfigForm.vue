<template>
  <div :class="['parse-config', readonly ? 'is-readonly' : '']">
    <el-form-item
      :label="$t('knowledgeManage.knowledgeDatabase.fileUpload.segmentSetting')"
    >
      <div class="segmentList">
        <div
          v-for="segmentItem in segmentList"
          :key="segmentItem.text"
          :class="[
            'segmentItem',
            form.docSegment.segmentMethod === segmentItem.label
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
    <template v-if="form.docSegment.segmentMethod === '0'">
      <el-form-item :label="$t('knowledgeManage.chunkTypeSet')">
        <div class="segmentList">
          <div
            v-for="segmentCommon in segmentCommonList"
            :key="segmentCommon.text"
            :class="[
              'segmentItem',
              form.docSegment.segmentType === segmentCommon.label
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
        v-if="form.docSegment.segmentType === '1'"
        prop="docSegment.splitter"
        :rules="
          form.docSegment.segmentType === '1'
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
            {{ $t('knowledgeManage.knowledgeDatabase.fileUpload.segmentTips') }}
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
        v-if="form.docSegment.segmentType === '1'"
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
            ['0', '1', '3', '4'].includes(form.docSegment.segmentType)
              ? ''
              : 'set',
          ]"
        >
          <el-input
            type="number"
            v-model.number="form.docSegment.maxSplitter"
            :placeholder="$t('knowledgeManage.splitMax')"
          ></el-input>
        </div>
      </el-form-item>
      <el-form-item
        v-if="form.docSegment.segmentType === '1'"
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
          v-model.number="form.docSegment.overlap"
          :placeholder="$t('knowledgeManage.overLapNumPlaceholder')"
        ></el-input>
      </el-form-item>
    </template>
    <template v-if="form.docSegment.segmentMethod === '1'">
      <div v-for="item in fatSonBlock" :key="item.level" class="commonSet">
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
                $t('knowledgeManage.knowledgeDatabase.fileUpload.segmentTips')
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
            v-model.number="form.docSegment[item.maxSplitter]"
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
        form.docSegment.segmentType === '1' ||
        form.docSegment.segmentMethod === '1'
      "
    >
      <el-checkbox-group v-model="form.docPreprocess">
        <el-checkbox label="replaceSymbols">
          {{ $t('knowledgeManage.replaceSymbols') }}
        </el-checkbox>
        <el-checkbox label="deleteLinks">
          {{ $t('knowledgeManage.deleteLinks') }}
        </el-checkbox>
      </el-checkbox-group>
    </el-form-item>
    <!-- 多模态但还没传文件时，模型选择整片是空的，整项一起隐藏 -->
    <el-form-item
      :label="$t('knowledgeManage.parsingMethod')"
      v-if="multiModal && hasAnyMedia"
    >
      <el-form-item
        v-if="hasMedia('video') || hasMedia('audio')"
        class="model-row"
        prop="asrModelId"
        :rules="
          hasMedia('audio')
            ? [
                {
                  required: true,
                  message: $t('knowledgeManage.parseTemplate.asrRequired'),
                  trigger: 'change',
                },
              ]
            : []
        "
      >
        <div class="segmentList">
          <span style="display: inline-block; width: 100px">
            <span class="red" v-if="hasMedia('audio')">*</span>
            ASR
          </span>
          <modelSelect
            v-model="form.asrModelId"
            :options="asrOptions"
            clearable
            @change="onAsrChange"
          />
        </div>
      </el-form-item>
      <div class="segmentList" v-if="hasMedia('video') || hasMedia('image')">
        <span style="display: inline-block; width: 100px">
          {{ $t('knowledgeManage.config.visionModal') }}
        </span>
        <modelSelect
          v-model="form.multimodalModelId"
          :options="visionOptions"
          clearable
        />
      </div>
    </el-form-item>
    <el-form-item
      :label="$t('knowledgeManage.parsingMethod')"
      prop="docAnalyzer"
      v-else-if="!multiModal"
    >
      <el-checkbox-group
        v-model="form.docAnalyzer"
        @change="docAnalyzerChange($event)"
      >
        <div
          v-for="analyzerItem in docAnalyzerList"
          :class="[
            'docAnalyzerList',
            form.docAnalyzer.includes(analyzerItem.label)
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
        form.docAnalyzer.includes('ocr') || form.docAnalyzer.includes('model')
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
          {{ modelTypeTip[form.docAnalyzer[1]]['label'] }}
        </span>
        <el-tooltip
          :content="modelTypeTip[form.docAnalyzer[1]]['desc']"
          placement="right"
        >
          <span class="el-icon-question question"></span>
        </el-tooltip>
      </template>
      <el-select
        v-model="form.parserModelId"
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
import splitterDialog from './splitterDialog.vue';
import modelSelect from '@/components/modelSelect.vue';
import {
  ocrSelectList,
  getSplitter,
  createSplitter,
  editSplitter,
  delSplitter,
} from '@/api/knowledge';
import { selectASRList, selectModelList } from '@/api/modelAccess';
import {
  SEGMENT_COMMON_LIST,
  SEGMENT_LIST,
  DOC_ANALYZER_LIST,
  FAT_SON_BLOCK,
  MODEL_TYPE_TIP,
} from '../config';

export default {
  name: 'ParseConfigForm',
  components: { splitterDialog, modelSelect },
  props: {
    // 解析配置对象，就地修改；外层 el-form 的 model 必须是同一个对象
    value: { type: Object, required: true },
    // 解析方式区：true 走 ASR/图文问答模型，false 走文字提取/OCR
    multiModal: { type: Boolean, default: false },
    // 多模态下参与解析的文件类型：video / audio / image
    mediaTypes: { type: Array, default: () => [] },
    // 内置（默认）模板只展示、不可改
    readonly: { type: Boolean, default: false },
  },
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
      validateSplitter,
      form: this.value,
      splitOptions: [],
      checkSplitter: { splitter: [], subSplitter: [] },
      segmentType: '',
      modelOptions: [],
      asrOptions: [],
      visionOptions: [],
      segmentList: SEGMENT_LIST,
      segmentCommonList: SEGMENT_COMMON_LIST,
      docAnalyzerList: DOC_ANALYZER_LIST,
      fatSonBlock: FAT_SON_BLOCK,
      modelTypeTip: MODEL_TYPE_TIP,
      placeholderText: this.$t('knowledgeManage.placeholderText'),
      titleText: this.$t('knowledgeManage.titleText'),
    };
  },
  computed: {
    hasAnyMedia() {
      return ['video', 'audio', 'image'].some(t => this.hasMedia(t));
    },
  },
  watch: {
    value(val) {
      this.form = val;
      this.refresh();
    },
    multiModal() {
      this.getModelOptions();
    },
  },
  async created() {
    this.getModelOptions();
    await this.getSplitterList('');
    this.refresh();
  },
  methods: {
    // 外部回填配置后调用，重刷已选分隔符标签
    refresh() {
      this.$nextTick(() => {
        const { splitter, subSplitter } = this.form.docSegment;
        const filterByType = values =>
          this.splitOptions.filter(item => values.includes(item.splitterValue));
        this.checkSplitter = {
          splitter: filterByType(splitter),
          subSplitter: filterByType(subSplitter),
        };
      });
    },
    // 外部重置表单后调用
    reset() {
      this.checkSplitter = { splitter: [], subSplitter: [] };
      this.splitOptions = this.splitOptions.map(item => ({
        ...item,
        checked: false,
      }));
      this.getModelOptions();
    },
    hasMedia(type) {
      return this.mediaTypes.includes(type);
    },
    onAsrChange(value) {
      // 父组件要用模型上的 maxAsrFileSize 校验音频大小，把选中项一并带出去
      this.$emit(
        'asr-change',
        value,
        this.asrOptions.find(item => item.modelId === value),
      );
    },
    getModelOptions() {
      if (this.multiModal) {
        selectASRList().then(res => {
          if (res.code === 0) this.asrOptions = res.data.list || [];
        });
        selectModelList().then(res => {
          if (res.code === 0) {
            this.visionOptions = (res.data.list || []).filter(
              item => item.config.visionSupport === 'support',
            );
          }
        });
        return;
      }
      this.getOcrList();
    },
    getOcrList() {
      ocrSelectList().then(res => {
        if (res.code === 0) this.modelOptions = res.data.list || [];
      });
    },
    segmentClick(label) {
      this.form.docSegment.segmentType = label;
    },
    segmentSetClick(label) {
      this.form.docSegment.segmentMethod = label;
    },
    analyzerDisabled(label) {
      if (label === 'text') return true;
    },
    docAnalyzerChange(val) {
      this.form.parserModelId = '';
      this.modelOptions = [];
      if (val.length === 3) {
        this.form.docAnalyzer = [val[0], val[2]];
      }
      this.getModelOptions();
    },
    maxSplitterChange(item) {
      if (item.level === 'parent') {
        const parentMaxValue = this.form.docSegment.maxSplitter;
        const sonBlock = this.fatSonBlock.find(block => block.level === 'son');
        if (sonBlock) {
          sonBlock.maxSplitterNum = parentMaxValue;
          if (this.form.docSegment.subMaxSplitter > parentMaxValue) {
            this.form.docSegment.subMaxSplitter = parentMaxValue;
            this.$message.warning(
              this.$t('knowledgeManage.childSegmentMaxAdjusted', {
                parentMaxValue,
              }),
            );
          }
        }
      } else if (item.level === 'son') {
        const sonMaxValue = this.form.docSegment.subMaxSplitter;
        const parentMaxValue = this.form.docSegment.maxSplitter;
        if (sonMaxValue > parentMaxValue) {
          this.form.docSegment.subMaxSplitter = parentMaxValue;
          this.$message.warning(
            this.$t('knowledgeManage.childSegmentMaxAdjustedTips', {
              parentMaxValue,
            }),
          );
        }
      }
    },
    showSplitterSet(type) {
      this.segmentType = type;
      this.$refs.splitterDialog.showDialog(this.checkSplitter[type]);
    },
    checkData(data) {
      this.checkSplitter[this.segmentType] = data;
      this.form.docSegment[this.segmentType] = data.map(
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
        .catch(() => {
          this.getSplitterList('');
        });
    },
  },
};
</script>
<style lang="scss" scoped>
.parse-config.is-readonly {
  pointer-events: none;
  opacity: 0.7;
}

.parse-config {
  // 嵌套的 el-form-item 只为挂校验规则，不要它的标签宽度和外边距
  .model-row {
    margin-bottom: 0;

    ::v-deep > .el-form-item__content {
      margin-left: 0 !important;
    }
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

  .width100 {
    width: 100%;
  }

  .red {
    color: #f56c6c;
  }
}
</style>
