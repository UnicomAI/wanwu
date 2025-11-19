<template>
  <div class="page-wrapper full-content">
    <div class="page-title">
      <span
        class="el-icon-arrow-left back"
        @click="goBack"
      ></span>
      AddFile
      <LinkIcon type="knowledge" />
    </div>
    <div class="table-box">
      <div class="fileUpload">
        <el-steps
          :active="active"
          class="fileStep"
          align-center
        >
          <el-step title="FileUpload"></el-step>
          <el-step title="ParameterSetting"></el-step>
        </el-steps>

        <!-- FileUpload -->
        <div v-if="active === 1">
          <div class="fileBtn">
            <el-radio-group
              v-model="fileType"
              @change="fileTypeChage"
            >
              <el-radio-button label="file">从FileUpload</el-radio-button>
              <el-radio-button label="fileUrl">urlFileUpload</el-radio-button>
              <el-radio-button label="url">url单条Upload</el-radio-button>
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
                :limit="5"
                multiple
                accept=".pdf,.docx,.doc,.txt,.xlsx,.xls,.zip,.tar.gz,.csv,.pptx,.html,.md,.ofd,.wps"
                :file-list="fileList"
                :on-change="uploadOnChange"
              >
                <div>
                  <div>
                    <img
                      :src="require('@/assets/imgs/uploadImg.png')"
                      class="upload-img"
                    />
                    <p class="click-text">将File拖到此处，OR<span class="clickUpload">点击Upload</span></p>
                  </div>
                  <div class="tips">
                    <p v-if="fileType === 'file'"><span class="red">*</span>您可单独OR者BatchUpload以下格式 of Document：pdf/docx/pptx/doc/wps/ofdFileMax为200MB，xlsx/xls/csv/txt/html/md/FileMax为20MB。zip/tar.gz格式内 of Document需符合各自File格式Upload大小限制</p>
                    <p v-if="fileType === 'file'"><span class="red">*</span>非压缩包File，一次可传5个File，如File页数多，DocumentParse when 间较长，平均3秒/页，Please您耐心等待</p>
                    <p v-if="fileType === 'fileUrl'"><span class="red">*</span>BatchUpload支持.xlsx格式，仅可Upload1个。Document最多可Add100条url，File不超 15mb <a
                        class="template_downLoad"
                        href="#"
                        @click.prevent.stop="downloadTemplate"
                      >模版Download</a></p>
                    <p v-if="fileType === 'fileUrl'"><span class="red">*</span>当前Content不自动Update</p>
                  </div>
                </div>
              </el-upload>
            </div>
          </div>
          <div
            class="el-upload-url"
            v-else
          >
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

        <!-- ParameterSetting -->
        <div
          v-else
          class="params_form"
        >
          <el-form
            :model="ruleForm"
            ref="ruleForm"
            label-width="140px"
            class="demo-ruleForm"
            @submit.native.prevent
            label-position="left"
          >
            <el-form-item label="分段Setting">
              <div class="segmentList">
                <div
                  v-for="segmentItem in segmentList"
                  :key="segmentItem.text"
                  :class="['segmentItem',ruleForm.docSegment.segmentMethod === segmentItem.label ? 'activeAnalyzer' : '']"
                  style="width:50%;"
                  @click="segmentSetClick(segmentItem.label)"
                >
                  <div class="itemImg">
                    <img :src="require(`@/assets/imgs/${segmentItem.img}`)" />
                  </div>
                  <div>
                    <p class="analyzerItem_text">{{segmentItem.text}}</p>
                    <h3 class="analyzerItem_desc">{{segmentItem.desc}}</h3>
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
                      :class="['segmentItem',ruleForm.docSegment.segmentType === segmentCommon.label ? 'activeAnalyzer' : '']"
                      @click="segmentClick(segmentCommon.label)"
                    >
                      <div>
                        <p class="analyzerItem_text">{{segmentCommon.text}}</p>
                        <h3 class="analyzerItem_desc">{{segmentCommon.desc}}</h3>
                      </div>
                    </div>
                  </div>
                </el-form-item>
                <el-form-item
                  v-if="ruleForm.docSegment.segmentType == '1'"
                  prop="docSegment.splitter"
                  :rules="ruleForm.docSegment.segmentType === '1' 
                  ? [{ required: true,validator: validateSplitter('splitter'), message: $t('knowledgeManage.markTips'), trigger: 'blur' }] 
                  : []"
                >
                  <template #label>
                    <span>分段标识</span>
                    <el-tooltip
                      :content="$t('knowledgeManage.splitOptionsTips')"
                      placement="right"
                    >
                      <span class="el-icon-question question"></span>
                    </el-tooltip>
                  </template>
                  <el-tag
                    v-for="(tag,index) in checkSplitter['splitter']"
                    :key="'tag'+index"
                    :disable-transitions="false"
                    class="splitterTag"
                  >
                    {{tag.splitterName.replace(/\n/g, '\\n')}}
                  </el-tag>
                  <el-button
                    class="button-new-tag"
                    size="small"
                    @click="showSplitterSet('splitter')"
                  > + 分段标识Setting</el-button>
                </el-form-item>
                <el-form-item
                  v-if="ruleForm.docSegment.segmentType == '1'"
                  prop="docSegment.maxSplitter"
                  :rules="[
                  { required: true, message: $t('knowledgeManage.splitMax'),trigger:'blur'},
                  { type:'number',min:200,max:4000,message:'Please enter有效范围内 of 数Value',trigger: 'blur'}
                  ]"
                >
                  <template #label>
                      <span>{{$t('knowledgeManage.splitMax')}}</span>
                      <el-tooltip
                        :content="$t('knowledgeManage.splitMaxTips')"
                        placement="right"
                      >
                        <span class="el-icon-question question"></span>
                      </el-tooltip>
                  </template>
                  <div :class="[['0','1','3','4'].includes(ruleForm.docSegment.segmentType)?'':'set']">
                    <el-input type="number" v-model.number="ruleForm.docSegment.maxSplitter" :placeholder="$t('knowledgeManage.splitMax')"></el-input>
                  </div>
                </el-form-item>
                <el-form-item
                  v-if="ruleForm.docSegment.segmentType == '1'"
                  :label="$t('knowledgeManage.overLapNum')"
                  prop="docSegment.overlap"
                  :rules="[
                    { required: true, message:$t('knowledgeManage.overLapNumTips'),trigger:'blur'},
                    { type:'number',min:0,max:1,message:'Please enter有效范围内 of 数Value',trigger: 'blur'}
                  ]"
                >
                  <el-input
                    :min="0"
                    :max="0.25"
                    :step="0.01"
                    type="number" 
                    v-model.number="ruleForm.docSegment.overlap" 
                    placeholder="数Value范围0-0.25" 
                    ></el-input>
                </el-form-item>
              </template>
              <template v-if="this.ruleForm.docSegment.segmentMethod === '1'">
                <div
                  v-for="item in fatSonBlock"
                  :key="item.level"
                  class="commonSet"
                >
                  <h3 class="title"><span class="bar"></span>{{item.title}}</h3>
                  <el-form-item
                    :prop="item.splitterProp"
                    :rules="[
                    { required: true,validator: validateSplitter(item.key), message: $t('knowledgeManage.markTips'), trigger: 'blur' }
                    ]"
                  >
                    <template #label>
                      <span>分段标识</span>
                      <el-tooltip
                        :content="$t('knowledgeManage.splitOptionsTips')"
                        placement="right"
                      >
                        <span class="el-icon-question question"></span>
                      </el-tooltip>
                    </template>
                    <el-tag
                      v-for="(tag,index) in checkSplitter[item.key]"
                      :key="'tag'+index"
                      :disable-transitions="false"
                      class="splitterTag"
                    >
                      {{tag.splitterName.replace(/\n/g, '\\n')}}
                    </el-tag>
                    <el-button
                      class="button-new-tag"
                      size="small"
                      @click="showSplitterSet(item.key)"
                    > + 分段标识Setting</el-button>
                  </el-form-item>
                  <el-form-item
                    :prop="item.maxSplitterProp"
                    :rules="[
                    { required: true, message: $t('knowledgeManage.splitMax'),trigger:'blur'},
                    { type:'number',min:200,max:item.maxSplitterNum,message:'Please enter有效范围内 of 数Value',trigger: 'blur'}
                    ]"
                  >
                    <template #label>
                      <span>可分割MaxValue</span>
                      <el-tooltip
                        :content="$t('knowledgeManage.splitMaxTips')"
                        placement="right"
                      >
                        <span class="el-icon-question question"></span>
                      </el-tooltip>
                    </template>
                    <el-input type="number" :min="200" :max="item.maxSplitterNum" v-model.number="ruleForm.docSegment[item.maxSplitter]" :placeholder="$t('knowledgeManage.splitMax')" @change="maxSplitterChange(item)"></el-input>
                  </el-form-item>
                </div>
              </template>
              <el-form-item
                label="文本预Process规Then"
                prop="docPreprocess"
                v-if="ruleForm.docSegment.segmentType == '1'"
              >
                <el-checkbox-group v-model="ruleForm.docPreprocess">
                  <el-checkbox label="replaceSymbols">替换掉连续 of 空格、换行符And制表符</el-checkbox>
                  <el-checkbox label="deleteLinks">Delete所有URLAnd电子邮件地址</el-checkbox>
                </el-checkbox-group>
              </el-form-item>
              <el-form-item
                label="Parse方式"
                prop="docAnalyzer"
              >
                <el-checkbox-group
                  v-model="ruleForm.docAnalyzer"
                  @change="docAnalyzerChange($event)"
                >
                  <div
                    v-for="analyzerItem in docAnalyzerList"
                    :class="['docAnalyzerList',ruleForm.docAnalyzer.includes(analyzerItem.label) ? 'activeAnalyzer' : '']"
                  >
                    <el-checkbox
                      :label="analyzerItem.label"
                      :disabled="analyzerDisabled(analyzerItem.label)"
                    >{{analyzerItem.text}}</el-checkbox>
                    <h3 class="analyzerItem_desc">{{analyzerItem.desc}}</h3>
                  </div>
                </el-checkbox-group>
              </el-form-item>
              <el-form-item
                prop="parserModelId"
                v-if="ruleForm.docAnalyzer.includes('ocr')||ruleForm.docAnalyzer.includes('model')"
                :rules="[
                    { required: true, message:'Please selectModel',trigger:'blur'}
                ]"
              >
              <template #label>
                <span>{{modelTypeTip[ruleForm.docAnalyzer[1]]['label']}}</span>
                <el-tooltip
                  :content="modelTypeTip[ruleForm.docAnalyzer[1]]['desc']"
                  placement="right"
                >
                  <span class="el-icon-question question"></span>
                </el-tooltip>
              </template>
                <el-select
                  v-model="ruleForm.parserModelId"
                  placeholder="Please select"
                  class="width100"
                >
                  <el-option
                    v-for="item in modelOptions"
                    :key="item.modelId"
                    :label="item.displayName"
                    :value="item.modelId"
                  >
                  </el-option>
                </el-select>
              </el-form-item>
              <el-form-item
                label="元Data管理"
                prop="docAnalyzer"
              >
                <mataData
                  ref="mataData"
                  @updateMeata="updateMeata"
                  :knowledgeId="knowledgeId"
                />
              </el-form-item>
          </el-form>
        </div>
        <!-- UploadFile of List -->
        <div
          class="file-list"
          v-if="fileList.length > 0 && active === 1 "
        >
          <transition name="el-zoom-in-top">
            <ul class="document_lise">
              <li
                v-for="(file, index) in fileList"
                :key="index"
                class="document_lise_item"
              >
                <div
                  style="padding:8px 0;"
                  class="lise_item_box"
                >
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
                        <i
                          class="el-icon-close close fail"
                          v-else
                        ></i>
                      </span>
                      <i
                        class="el-icon-loading"
                        v-else-if="file.percentage !== 100 && index === fileIndex"
                      ></i>
                    </span>
                    <span style="margin-left:30px;">
                      <i
                        class="el-icon-error error"
                        @click="handleRemove(file,index)"
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
            v-if="active === 2"
          >上一步</el-button>
          <el-button
            type="primary"
            size="mini"
            @click="nextStep"
            v-if="active === 1"
            :loading="urlLoading"
          >下一步</el-button>
          <el-button
            type="primary"
            size="mini"
            @click="submitInfo"
            v-if="active === 2"
          >确 定</el-button>
          <el-button
            size="mini"
            @click="formReset"
            v-if="active === 2"
          >重 置</el-button>
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
      @relodData="relodData"
      @checkData="checkData"
    />
  </div>
</template>
<script>
import urlAnalysis from "./urlAnalysis.vue";
import uploadChunk from "@/mixins/uploadChunk";
import {
  docImport,
  ocrSelectList,
  delSplitter,
  getSplitter,
  createSplitter,
  editSplitter,
  parserSelect,
} from "@/api/knowledge";
import { delfile } from "@/api/chunkFile";
import LinkIcon from "@/components/linkIcon.vue";
import splitterDialog from "./splitterDialog.vue";
import mataData from "./metadata.vue";
import {USER_API} from "@/utils/requestConstants"
import {
  SEGMENT_COMMON_LIST,
  SEGMENT_LIST,
  DOC_ANALYZER_LIST,
  FAT_SON_BLOCK,
  MODEL_TYPE_TIP
} from "../config";
export default {
  components: { LinkIcon, urlAnalysis, splitterDialog, mataData },
  mixins: [uploadChunk],
  data() {
    const validateSplitter = (type) => {
      return (rule, value, callback) =>{
        if (this.checkSplitter[type].length === 0) {
          callback(new Error(this.$t("knowledgeManage.splitterRequired")));
        } else {
          callback();
        }
      }
    };
    return {
      validateSplitter: validateSplitter,
      placeholderText: "Search分隔符",
      titleText: "Create分隔符",
      splitterValue: "",
      tableData: [],
      modelOptions: [],
      urlValidate: false,
      active: 1,
      fileType: "file",
      knowledgeId: this.$route.query.id,
      knowledgeName: this.$route.query.name,
      fileList: [],
      fileUrl: "",
      docInfoList: [],
      segmentType:'',
      ruleForm: {
        docAnalyzer: ["text"],
        docMetaData: [], //元Data管理Data
        docPreprocess: ["replaceSymbols"], //'deleteLinks','replaceSymbols'
        docSegment: {
          segmentType: "0",
          // splitter: ["！", "。", "？", "?", "!", ".", "......"],
          splitter:["\n\n"],
          maxSplitter: 1024,
          overlap: 0.2,
          segmentMethod:"0",//0Yes通用分段，1Yes父子分段
          subMaxSplitter:200,//父子分段Required
          // subSplitter:["！", "。", "？", "?", "!", ".", "......"]//父子分段Required
          subSplitter:["\n"]
        },
        docInfoList: [],
        docImportType: 0,
        knowledgeId: this.$route.query.id,
        parserModelId: "",
      },
      checkSplitter:{
        splitter:[],
        subSplitter:[]
      },
      splitOptions: [],
      urlLoading: false,
      segmentCommonList: SEGMENT_COMMON_LIST,
      segmentList: SEGMENT_LIST,
      docAnalyzerList: DOC_ANALYZER_LIST,
      fatSonBlock:FAT_SON_BLOCK,
      modelTypeTip:MODEL_TYPE_TIP
    };
  },
  async created() {
    await this.getSplitterList("");
    await this.custom();
  },
  methods: {
    maxSplitterChange(item){
      if(item.level === 'parent'){
        const parentMaxValue = this.ruleForm.docSegment.maxSplitter;
        const sonBlock = this.fatSonBlock.find(block => block.level === 'son');
        if (sonBlock) {
          sonBlock.maxSplitterNum = parentMaxValue;
          if (this.ruleForm.docSegment.subMaxSplitter > parentMaxValue) {
            this.ruleForm.docSegment.subMaxSplitter = parentMaxValue;
            this.$message.warning(`子分段MaxValue已调整为 ${parentMaxValue}，不能超 父分段 of MaxValue`);
          }
        }
      }else if(item.level === 'son'){
        const sonMaxValue = this.ruleForm.docSegment.subMaxSplitter;
        const parentMaxValue = this.ruleForm.docSegment.maxSplitter;
        if (sonMaxValue > parentMaxValue) {
          this.ruleForm.docSegment.subMaxSplitter = parentMaxValue;
          this.$message.warning(`子分段MaxValue不能超 父分段 of MaxValue ${parentMaxValue}`);
        }
      }
    },
    getParserList() {
      parserSelect()
        .then((res) => {
          if (res.code === 0) {
            this.modelOptions = res.data.list || [];
          }
        })
        .catch(() => {});
    },
    docAnalyzerChange(val) {
      this.ruleForm.parserModelId = "";
      this.modelOptions = [];
      if(val.length === 3){
        this.ruleForm.docAnalyzer = [val[0],val[2]]
      }
      if (this.ruleForm.docAnalyzer.includes("ocr")) {
        this.getOcrList();
      } else if (val.includes("model")) {
        this.getParserList();
      }
    },
    segmentClick(label) {
      this.ruleForm.docSegment.segmentType = label;
    },
    segmentSetClick(label) {
      this.ruleForm.docSegment.segmentMethod = label;
    },
    analyzerDisabled(label) {
      if (label === "text") return true;
    },
    custom() {
      this.$nextTick(() => {
        const { splitter, subSplitter } = this.ruleForm.docSegment;
        const filterByType = (values) => 
          this.splitOptions.filter(item => 
            values.includes(item.splitterValue) && item.type === "preset"
          );
        this.checkSplitter = {
          splitter: filterByType(splitter),
          subSplitter: filterByType(subSplitter)
        };
      });
    },
    updateMeata(data) {
      this.ruleForm.docMetaData = data;
    },
    validateMetaData() {
      const hasEmptyField = this.ruleForm.docMetaData.some((item) => {
        const isMetaKeyEmpty =
          !item.metaKey ||
          (typeof item.metaKey === "string" && item.metaKey.trim() === "");
        const isMetaRuleRequired = item.metadataType !== "value";
        const isMetaRuleEmpty =
          isMetaRuleRequired &&
          (!item.metaRule ||
            (typeof item.metaRule === "string" && item.metaRule.trim() === ""));
        return isMetaKeyEmpty || isMetaRuleEmpty;
      });
      if (hasEmptyField) {
        this.$message.error("元Data管理存在未填写 of Required字段");
        return false;
      }
      return true;
    },
    checkData(data) {
      this.checkSplitter[this.segmentType] = data;
      this.ruleForm.docSegment[this.segmentType] = data.map(
        (item) => item.splitterValue
      );
    },
    relodData(name) {
      this.getSplitterList(name);
    },
    async getSplitterList(splitterName) {
      const res = await getSplitter({ splitterName });
      if (res.code === 0) {
        this.splitOptions = (res.data.knowledgeSplitterList || []).map(
          (item) => ({
            ...item,
            showDel: false,
            showIpt: false,
          })
        );
      }
    },
    editItem(item) {
      editSplitter({
        splitterId: item.splitterId,
        splitterName: item.splitterName,
        splitterValue: item.splitterName,
      }).then((res) => {
        if (res.code === 0) {
          item.showIpt = false;
          this.getSplitterList("");
        }
      });
    },
    createItem(item) {
      createSplitter({
        splitterName: item.splitterName,
        splitterValue: item.splitterName,
      }).then((res) => {
        if (res.code === 0) {
          item.showIpt = false;
          this.getSplitterList("");
        }
      });
    },
    async delSplitterItem(item) {
      this.$confirm(
        `Delete分隔符${item.splitterName}`,
        "Confirm要Delete当前分隔符？",
        {
          confirmButtonText: "确定",
          cancelButtonText: "Cancel",
          type: "warning",
        }
      )
        .then(async () => {
          const res = await delSplitter({ splitterId: item.splitterId });
          if (res.code === 0) {
            this.getSplitterList("");
          }
        })
        .catch((error) => {
          this.getSplitterList("");
        });
    },
    showSplitterSet(type) {
      this.segmentType = type;
      this.$refs.splitterDialog.showDiaglog(this.checkSplitter[this.segmentType]);
    },
    goBack() {
      this.$router.go(-1);
    },
    getOcrList() {
      ocrSelectList().then((res) => {
        if (res.code === 0) {
          this.modelOptions = res.data.list || [];
        }
      });
    },
    handleSetData(data) {
      this.docInfoList = [];
      data.map((item) => {
        this.docInfoList.push({
          docName: item.fileName,
          docSize: item.fileSize,
          docUrl: item.url,
          docType: "url",
        });
      });
    },
    async downloadTemplate() {
      const url = `${USER_API}/static/docs/url_import_template.xlsx`;
      const fileName = "url_import_template.xlsx";
      try {
        const response = await fetch(url);
        if (!response.ok) throw new Error("FileDoes not existOR服务器Error");

        const blob = await response.blob();
        const blobUrl = URL.createObjectURL(blob);

        const a = document.createElement("a");
        a.href = blobUrl;
        a.download = fileName;
        a.click();

        URL.revokeObjectURL(blobUrl); // 释放Memory
      } catch (error) {
        alert("FileDownloadFailed，PleaseLaterRetry！");
      }
    },
    handleLoading(val, result) {
      this.urlLoading = val;
      if (result === "success") {
        this.reset();
      }
    },
    reset() {
      if (this.source.length > 0) {
        for (let i = 0; i < this.source.length; i++) {
          this.source[i].cancel();
        }
      }
      let ids = [];
      if (this.fileList.length > 0) {
        this.fileList.map((item) => {
          if (item.id) {
            if (item.id.includes(",")) {
              //rag一体机No此逻辑
              const list = item.id.split(",");
              list.map((item) => {
                ids.push(item);
              });
            } else {
              ids.push(item.id);
            }
          }
        });
        if (ids.length > 0) {
          this.deleteData({ id: ids }); //Cancel when DeleteFile
        }
      }
      this.$refs["uplodForm"].resetFields();
      this.uplodForm.knowValue = null;
      this.fileList = [];
      this.resultDisabled = true;
      this.source = [];
      this.fileUuid = "";
      this.$emit("handleSetOpen", { isShow: false, knowValue: null });
      this.uploading = false;
    },
    // Delete已UploadFile
    handleRemove(item, index) {
      if(item.percentage < 100){
        this.fileList.splice(index,1);
        this.cancelAllRequests();//Cancel所HasRequest
        return;
      }
      this.delfile({
        fileList: [this.resList[index]["name"]],
        isExpired: true,
      });
      this.fileList = this.fileList.filter((files) => files.name !== item.name);
      if (this.fileList.length === 0) {
        this.file = null;
      } else {
        this.fileIndex--;
      }
      if (this.docInfoList.length > 0) {
        this.docInfoList.splice(index,1);
      }
    },
    delfile(data) {
      delfile(data).then((res) => {
        if (res.code === 0) {
          this.$message.success("DeleteSuccess");
        }
      });
    },
    filterSize(size) {
      if (!size) return "";
      var num = 1024.0; //byte
      if (size < num) return size + "B";
      if (size < Math.pow(num, 2)) return (size / num).toFixed(2) + "KB"; //kb
      if (size < Math.pow(num, 3))
        return (size / Math.pow(num, 2)).toFixed(2) + "MB"; //M
      if (size < Math.pow(num, 4))
        return (size / Math.pow(num, 3)).toFixed(2) + "G"; //G
      return (size / Math.pow(num, 4)).toFixed(2) + "T"; //T
    },
    fileTypeChage() {
      // Cancel所Has正在进行 of UploadRequest
      this.cancelAllRequests();
      
      // ResetUpload相关Status
      this.fileIndex = 0;
      this.file = null;
      this.resList = [];

      this.docInfoList = [];
      this.fileList = [];
    },
    submitInfo() {
      const { segmentMethod,segmentType, splitter,subSplitter } = this.ruleForm.docSegment;
      if (
        (segmentMethod === '1' && (splitter.length === 0 || subSplitter.length === 0)) ||
        (segmentMethod !== '1' && segmentType === '1' && splitter.length === 0)
      ) {
        this.$refs.ruleForm.validate();
        return false;
      }
      this.$refs.ruleForm.clearValidate(["docSegment.splitter","docSegment.subSplitter"]);

      if (!this.validateMetaData()) {
        return;
      }

      this.ruleForm.docMetaData.forEach((item) => {
        delete item.metadataType;
      });

      if (this.fileType === "file") {
        this.ruleForm.docImportType = 0;
      } else if (this.fileType === "fileUrl") {
        this.ruleForm.docImportType = 2;
      } else {
        this.ruleForm.docImportType = 1;
      }
      this.ruleForm.docInfoList = this.docInfoList;
      let data = null;
      if (this.ruleForm.docSegment.segmentType == "0" && this.ruleForm.docSegment.segmentMethod !== "1") {
        data = this.ruleForm;
        delete data.docSegment.splitter;
        delete data.docSegment.maxSplitter;
        delete data.docSegment.overlap;
      } else {
        data = this.ruleForm;
      }
      docImport(data).then((res) => {
        if (res.code === 0) {
          this.$router.push({
            path: `/knowledge/doclist/${this.knowledgeId}`,
            query: { name: this.knowledgeName, done: "fileUpload" },
          });
        }
      });
    },
    formReset() {
      this.ruleForm = {
        docAnalyzer: ["text"],
        docMetaData: [], //元Data管理Data
        docPreprocess: ["replaceSymbols"], //'deleteLinks','replaceSymbols'
        docSegment: {
          segmentType: this.ruleForm.docSegment.segmentType,
          splitter: [], //"！","。","？","?","!",".","......"
          maxSplitter: 200,
          overlap: 0.2,
        },
        docInfoList: [],
        docImportType: 0,
        knowledgeId: this.$route.query.id,
        ocrModelId: "",
      };
      this.checkSplitter = {
        splitter:[],
        subSplitter:[]
      };
      this.splitOptions = this.splitOptions.map((item) => ({
        ...item,
        checked: false,
      }));
      this.$refs.ruleForm.clearValidate();
    },
    uploadOnChange(file, fileList) {
      if (!fileList.length) return;
      this.fileList = fileList;
      if (
        this.verifyEmpty(file) !== false &&
        this.verifyFormat(file) !== false &&
        this.verifyRepeat(file) !== false
      ) {
        setTimeout(() => {
          this.fileList.map((file, index) => {
            if (file.progressStatus && file.progressStatus !== "success") {
              this.$set(file, "progressStatus", "exception");
              this.$set(file, "showRetry", "false");
              this.$set(file, "showResume", "false");
              this.$set(file, "showRemerge", "false");
              if (file.size > this.maxSizeBytes) {
                this.$set(file, "fileType", "maxFile");
              } else {
                this.$set(file, "fileType", "minFile");
              }
            }
          });
        }, 10);
        //开始切片Upload(IfNoFile正在Upload)
        if (this.file === null) {
          this.startUpload();
        } else {
          //IfUpload当中Has新 of File加入
          if (this.file.progressStatus === "success") {
            this.startUpload(this.fileIndex);
          }
        }
      }
    },
    refreshFile(index) {
      //重新UploadFile
      this.fileList[index]["showRetry"] = "false";
      this.fileList[index]["percentage"] = 0;
      this.startUpload(index);
    },
    resumeFile(index) {
      //续传File
      this.fileList[index]["showResume"] = "false";
      this.nextChunkIndex = this.uploadedChunks;
      this.processNextChunk();
    },
    remergeFile(index) {
      //重新Upload
      this.mergeChunks();
    },
    uploadFile(fileName, oldName) {
      let type = oldName.split(".").pop();
      const docType =
        type === "gz" ? ".tar.gz" : "." + oldName.split(".").pop();
      this.docInfoList.push({
        docId: fileName,
        docName: oldName,
        docType,
      });
      this.fileIndex++;
      if (this.fileIndex < this.fileList.length) {
        this.startUpload(this.fileIndex);
      }
    },
    //  VerifyFileIs Empty
    verifyEmpty(file) {
      const isLt1GB = file.size / 1024 / 1024 / 1024 < 1;
      if (file.size <= 0) {
        setTimeout(() => {
          this.$message.warning(
            file.name + this.$t("knowledgeManage.filterFile")
          );
          this.fileList = this.fileList.filter(
            (files) => files.name !== file.name
          );
        }, 50);
        return false;
      }
      return true;
    },
    //  VerifyFileFormat
    verifyFormat(file) {
      const nameType = [
        "pdf",
        "docx",
        "doc",
        "pptx",
        "zip",
        "tar.gz",
        "xlsx",
        "xls",
        "csv",
        "txt",
        "html",
        "md",
        "ofd",
        "wps",
      ];
      const fileName = file.name;
      const isSupportedFormat = nameType.some((ext) =>
        fileName.endsWith(`.${ext}`)
      );
      if (!isSupportedFormat) {
        setTimeout(() => {
          this.$message.warning(
            file.name + this.$t("knowledgeManage.fileTypeError")
          );
          this.fileList = this.fileList.filter(
            (files) => files.name !== file.name
          );
        }, 50);
        return false;
      } else {
        const fileType = file.name.split(".").pop();
        const limit200 = [
          "pdf",
          "docx",
          "doc",
          "pptx",
          "zip",
          "tar.gz",
          "ofd",
          "wps",
        ];
        const limit20 = ["xlsx", "xls", "csv", "txt", "html", "md"];
        let isLimit200 = file.size / 1024 / 1024 < 200;
        let isLimit20 = file.size / 1024 / 1024 < 20;
        let num = 0;
        if (limit200.includes(fileType)) {
          num = 200;
          if (!isLimit200) {
            setTimeout(() => {
              this.$message.error(
                this.$t("knowledgeManage.limitSize") + `${num}MB!`
              );
              this.fileList = this.fileList.filter(
                (files) => files.name !== file.name
              );
            }, 50);
            return false;
          }
          return true;
        } else if (limit20.includes(fileType)) {
          num = 20;
          if (!isLimit20) {
            setTimeout(() => {
              this.$message.error(
                this.$t("knowledgeManage.limitSize") + `${num}MB!`
              );
              this.fileList = this.fileList.filter(
                (files) => files.name !== file.name
              );
            }, 50);
            return false;
          }
          return true;
        }
        return true;
      }
    },
    //  VerifyFileFormat
    verifyRepeat(file) {
      let res = true;
      setTimeout(() => {
        this.fileList = this.fileList.reduce((accumulator, current) => {
          const length = accumulator.filter(
            (obj) => obj.name === current.name
          ).length;
          if (length === 0) {
            accumulator.push(current);
          } else {
            this.$message.warning(
              current.name + this.$t("knowledgeManage.fileExist")
            );
            res = false;
          }
          return accumulator;
        }, []);
        return res;
      }, 50);
    },
    nextStep() {
      //UploadFileType
      if (this.fileType === "file" || this.fileType === "fileUrl") {
        if (this.fileIndex < this.fileList.length) {
          this.$message.warning("FileUpload中...");
          return false;
        }
        if (this.fileList.length === 0) {
          this.$message.warning("PleaseUploadFile!");
          return false;
        }
      }
      //url逐条Upload
      if (this.fileType === "url") {
        if (this.docInfoList.length === 0) {
          this.$message.warning("Please上Enterurl!");
          return false;
        }
      }
      this.active = 2;
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
.width100{
  width:100%;
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
/deep/.el-input-number.is-controls-right .el-input-number__decrease,
/deep/.el-input-number.is-controls-right .el-input-number__increase {
  line-height: 14px !important;
  border: 0;
}
/deep/ {
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
    padding:30px;
    .commonSet {
      background: #f6f7fe;
      padding: 15px;
      border-radius: 6px;
      .title{
        font-size:14px;
        border-bottom: 1px solid #dee1fe ;
        padding-bottom:10px;
        display: flex;
        align-items: center;
        .bar{
          display:inline-block;
          width:4px;
          height:14px;
          background: $color;
          margin-right:5px;
        }
      }
    }
    .commonSet:nth-child(1){
      margin-bottom:15px;
    }
    .el-form-item {
      display: flex;
      flex-direction: column;
      /deep/ {
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
        width:50%;
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