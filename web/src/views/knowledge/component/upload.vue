<template>
  <el-dialog
    :title="$t('knowledgeManage.fileUpload')"
    :visible.sync="openDialog"
    :close-on-click-modal="false"
    width="50%"
    :show-close="false"
  >
    <el-form
      :model="uplodForm"
      ref="uplodForm"
      :rules="uplodRules"
      label-width="110px"
      class="cust_Form"
    >
      <el-form-item
        :label="$t('knowledgeManage.uploadToKnowledge')"
        prop="knowValue"
      >
        <treeselect
          :clearable="true"
          :options="knowledgeData"
          :no-options-text="$t('createApp.noData')"
          :multiple="false"
          :disableBranchNodes="true"
          :normalizer="normalizeOptions"
          v-model="uplodForm.knowValue"
          :placeholder="$t('createApp.knowledgePlaceholder')"
          ref="optionRef"
          :disabled="fileList.length>0"
        ></treeselect>
      </el-form-item>
       <el-form-item
        :label="$t('knowledgeManage.analyticMethod')"
        required
      >
      <el-checkbox-group v-model="uplodForm.plugin">
        <el-checkbox label="text" disabled>{{$t('knowledgeManage.textExtraction')}}</el-checkbox>
        <el-checkbox label="ocr">{{$t('knowledgeManage.OCRAnalysis')}}</el-checkbox>
      </el-checkbox-group>
      </el-form-item>
      <el-form-item
        :label="$t('knowledgeManage.importDoc')"
        required
      >
       <!-- rag一体机上不Show -->
        <el-radio-group
          v-model="upLoadType"
          @change="handleRadioChange"
          v-removeAriaHidden
          v-if="platform !== 'YWD_RAG' && platform !== 'HW_RAG'"
        >
          <el-radio label="1">{{$t('knowledgeManage.formFile')}}</el-radio>
          <!-- 华为一体机隐藏urlUpload -->
          <el-radio label="2">{{$t('knowledgeManage.fromURL')}}</el-radio>
        </el-radio-group>

        <div v-if="upLoadType === '1'">
          <el-upload
            ref="uploadFile"
            class="upload-demo"
            :show-file-list="false"
            :auto-upload="false"
            action=""
            :limit="5"
            accept=".pdf,.docx,.doc,.txt,.xlsx,.xls,.zip,.tar.gz,.csv,.pptx,.html"
            :on-change="handleChange"
            :on-remove="handleRemove"
            :on-exceed="handleExceed"
            multiple
            :file-list="fileList"
            :disabled="!uplodForm.knowValue"
          >
            <el-button size="small" :disabled="!uplodForm.knowValue">{{$t('knowledgeManage.selectFile')}}</el-button>
            <p class="fileNumber">
              {{$t('knowledgeManage.uploadFileNum')}}：{{ fileList.length }} {{$t('knowledgeManage.successNum')}}：{{ fileIndex  }}{{$t('knowledgeManage.number')}}
            </p>
            <div
              slot="tip"
              class="el-upload__tip"
            >
              <div class="uploadTips">
               {{$t('knowledgeManage.uploadTips')}}
              </div>
              <div class="uploadTips">
                {{$t('knowledgeManage.uploadTips1')}}
              </div>
              <div class="uploadTips">
                {{$t('knowledgeManage.uploadTips2')}}
              </div>
              <transition name="el-zoom-in-top">
                <ul class="document_lise">
                  <li
                    v-for="(file, index) in fileList"
                    :key="index"
                  >
                    <div style="padding:8px 0;">
                      <span class="size">{{ file.name }}<span>{{ filterSize(file.size) }}</span></span>
                      <span>
                        <span v-if="file.percentage === 100">
                          <i class="el-icon-check check success" v-if="file.progressStatus === 'success'"></i>
                          <i class="el-icon-close close fail" v-else></i>
                        </span>
                        <i class="el-icon-loading" v-else-if="file.percentage !== 100 && index === fileIndex"></i>
                      </span>
                      <span style="margin-left:30px;">
                        <i class="el-icon-error error" @click="handleRemove(file)"></i>
                      </span>
                    </div>
                    <div style="display:flex;align-items: center;">
                        <el-progress
                          :percentage="file.percentage"
                          v-if="file.percentage !== 100"
                          :status="file.progressStatus"
                          max="100"
                          style="width:360px;"
                        ></el-progress>
                        <el-link type="success" :underline="false" v-if="file.showRetry === 'true'" @click="refreshFile(index)">Retry</el-link>
                        <el-link type="success" :underline="false" style="margin-left:10px;" v-if="file.showResume === 'true' && file.percentage > 0" @click="resumeFile(index)">续传</el-link>
                        <el-link type="success" :underline="false" style="margin-left:10px;" v-if="file.showRemerge === 'true'" @click="remergeFile(index)">续传</el-link>
                    </div>
                  </li>
                </ul>
              </transition>
            </div>
          </el-upload>
        </div>
        <div v-if="upLoadType === '2'">
          <el-tabs type="border-card" v-model="urlActive">
            <el-tab-pane :label="$t('knowledgeManage.addSeparately')" name="first">
              <urlAnalysis
                v-if="urlActive === 'first'"
                :categoryId="uplodForm.knowValue"
                :validate="urlValidate"
                ref="urlUpload"
                @handleLoading="handleLoading"
                @handleValidate="handleValidate"
                @handleSetDisabled="handleSetDisabled"
              ></urlAnalysis>
            </el-tab-pane>
            <el-tab-pane :label="$t('knowledgeManage.batchAdd')" name="second">
              <urlBatch
                v-if="urlActive === 'second'"
                :categoryId="uplodForm.knowValue"
                :validate="urlValidate"
                ref="urlBatchUpload"
                @handleSetBatchStatus="handleSetBatchStatus"
                @handleValidate="handleValidate"
                @handleSetBatchDisabled="handleSetBatchDisabled"
              ></urlBatch>
            </el-tab-pane>
          </el-tabs>
        </div>
      </el-form-item>
    </el-form>
    <span
      slot="footer"
      class="dialog-footer"
    >
      <el-button
        v-if="upLoadType === '1'"
        :disabled="resultDisabled || loadingResult || fileList.length <= 0"
        type="primary"
        size="mini"
        :loading="saveLoading"
        @click="saveUpload"
      >{{$t('knowledgeManage.confirmImport')}}</el-button>
      <el-button
        type="primary"
        v-if="upLoadType === '2' && urlActive === 'first'"
        size="mini"
        :disabled="urlSave"
        @click="handleSave"
        :loading="urlLoading"
        >{{$t('knowledgeManage.confirmImport')}}</el-button
      >
      <!-- urlUploadBatchAddButton -->
      <el-button
        type="primary"
        v-if="upLoadType === '2' && urlActive === 'second'"
        size="mini"
        :disabled="urlBatchDis"
        @click="handlebatchSave"
        :loading="urlLoading"
        >{{$t('knowledgeManage.confirmImport')}}</el-button
      >
      <el-button
        @click="reset"
        size="mini"
      >{{$t('createApp.cancel')}}</el-button>
    </span>
  </el-dialog>
</template>
<script>
import Treeselect from "@riophae/vue-treeselect";
import axios from "axios";
import {guid} from '@/utils/util'
import urlAnalysis from "./urlAnalysis.vue";
import urlBatch from "./urlBatch.vue";
import { getDocList,importDoc,saveImportDoc,deleteInvalid } from "@/api/knowledge";
import uploadChunk from "@/mixins/uploadChunk";
// 防抖Function
function debounce(fn, wait) {
    let timeout = null;
    return function() {
        let context = this;
        let args = arguments;
        if (timeout) clearTimeout(timeout);
        let callNow = !timeout;
        timeout = setTimeout(() => {
            timeout = null;
        }, wait);
        if (callNow) fn.apply(context, args);
    };
}
export default {
  components: {
    Treeselect,
    urlAnalysis,
    urlBatch
  },
  mixins: [uploadChunk],
  props: {
    open: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      urlBatchDis:true,
      urlActive:'first',
      platform:this.$platform,
      openDialog: false,
      urlSave: false,
      upDisabled: true,
      urlValidate: false,
      urlLoading: false,
      upLoadType: "1", // 1:从DocumentUpload；2:从urlUpload
      source: [],
      saveLoading:false,
      uploading: false,
      resultDisabled: true,
      uplodForm: {
        knowValue: null,
        plugin:['text','ocr']//'ocr'
      },
      uplodRules: {
        knowValue: [
          { required: true, message: this.$t('knowledgeManage.uploadFileTips'), trigger: "change" },
        ]
      },
      // fileList: [],
      knowledgeData: [],
      successNum: 0,
      fileUuid:''
    };
  },
  created() {
    this.getClassfyDoc();
    this.handleChange = debounce(this.handleChange,1000)
  },
  watch: {
    fileList: {
      handler(val) {
        if (val.length > 0) {
          this.$nextTick(() => {
            this.upDisabled = false;
          });
        } else {
          this.$nextTick(() => {
            this.upDisabled = true;
          });
        }
      },
      deep: true,
    },
    open: {
      handler(val) {
        this.openDialog = val
        val && this.getClassfyDoc();
      },
      immediate:true,
      deep: true,
    },
  },
  methods: {
    showDialog() {
      this.openDialog = true;
    },
    handleSetDisabled(val) {
      this.urlSave = val;
    },
    handleChange(file, fileList) {
      if (!fileList.length) return;
      this.fileList = fileList;
      if(
        this.verifyEmpty(file) !== false &&
        this.verifyFormat(file) !== false &&
        this.verifyRepeat(file) !== false
      ){
        setTimeout(() => {
          this.fileList.map((file, index) => {
            if(file.progressStatus && file.progressStatus !=='success'){
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
        },10)
        //开始切片Upload(IfNoFile正在Upload)
        if(this.file === null){
          this.startUpload();
        }else{//IfUpload当中Has新 of File加入
          if(this.file.progressStatus === 'success'){
            this.startUpload(this.fileIndex)
          }
        }
      }
    },
    refreshFile(index){//重新UploadFile
      this.fileList[index]['showRetry'] = 'false';
      this.fileList[index]['percentage'] = 0;
      this.startUpload(index)
    },
    resumeFile(index){//续传File
      this.fileList[index]['showResume'] = 'false';
      this.nextChunkIndex = this.uploadedChunks;
      this.processNextChunk();
    },
    remergeFile(index){//重新Upload
      this.mergeChunks()
    },
    //  VerifyFileIs Empty
    verifyEmpty(file) {
      const isLt1GB = file.size / 1024 / 1024 / 1024 < 1;
      if (file.size <= 0) {
        setTimeout(() => {
          this.$message.warning(file.name + this.$t('knowledgeManage.filterFile'));
          this.fileList = this.fileList.filter(
            (files) => files.name !== file.name
          );
        }, 50);
        return false
      }
      return true
    },
    //  VerifyFileFormat
    verifyFormat(file) {
      const nameType = ['docx','doc','pdf','xlsx','xls','txt','zip','tar.gz','csv','pptx','html']
      const fileName = file.name
      const isSupportedFormat = nameType.some(ext => fileName.endsWith(`.${ext}`));
      if (!isSupportedFormat) {
        setTimeout(() => {
          this.$message.warning(file.name + this.$t('knowledgeManage.fileTypeError'));
          this.fileList = this.fileList.filter(
            (files) => files.name !== file.name
          );
        }, 50);
        return false
      }else{
        let isLimit = file.size / 1024 / 1024 < 700;
        let num = 700;
        if(!isLimit){
          setTimeout(() => {
            this.$message.error(this.$t('knowledgeManage.limitSize')+`${num}MB!`);
            this.fileList = this.fileList.filter(
              (files) => files.name !== file.name
            );
          }, 50);
          return false
        }
        return  true
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
            this.$message.warning(current.name + this.$t('knowledgeManage.fileExist'));
            res = false
          }
          return accumulator;
        }, []);
        return res;
      }, 50);
    },
    // Delete已UploadFile
    handleRemove(item) {
      this.fileList = this.fileList.filter((files) => files.name !== item.name);
      if(this.fileList.length === 0){
        this.file = null
      }else{
        this.fileIndex--
      }
    },
    handleExceed() {
      this.$message.error(this.$t('knowledgeManage.fileLimitError'));
    },
    saveUpload() {
      //SaveSuccessResult
      let ids = [];
      this.fileList.map((item) => {
        if (item.id && item.id.includes(",")) {
          ids = ids.concat(item.id.split(","));
        } else {
          ids.push(item.id);
        }
      });
      const data = { id: ids };
      this.saveLoading = true;
      saveImportDoc(data).then(res =>{
        if (res.code === 0) {
          this.$message.success(this.$t('knowledgeManage.operateSuccess'));
          this.successNum = 0;
          this.fileList = [];
          this.fileUuid='';
          this.resultDisabled = true;
          this.saveLoading = false;
          this.$emit("handleSetOpen", {isShow:false,knowValue:this.uplodForm.knowValue});
          this.uplodForm.knowValue = null;
        }
      }).catch(error =>{
          this.saveLoading = false;
      })

    },
    submitVisible() {
      //Upload
      this.$refs["uplodForm"].validate((valid) => {
        if (valid) {
          if (!this.fileList.length) {
            this.$message.warning(this.$t('knowledgeManage.selectFile'));
            return;
          }
          if(this.fileUuid === ''){
            this.fileUuid = guid()
          }
          this.fileList.map((file, index) => {
            if (file.status === "ready" && file.back !== "true") {
              const type = file.name.split(".")[1];
              let formData = new FormData();
              formData.append("file", file.raw);
              formData.append("fileType", type);
              formData.append("categoryId", this.uplodForm.knowValue);
              formData.append("batchId", this.fileUuid);
              this.uploadFile(formData, index);
            }
          });
        }
      });
    },
     uploadFile(fileName='') {
      const file= this.fileList[this.fileIndex];
      if(this.fileUuid === ''){ this.fileUuid = guid()}
      const regex = /\.([^.]*)$/;
      const type = file.name.match(regex);
      let formData = new FormData();
      if(this.isChunk){
        formData.append("file", '');
      }else{
        formData.append("file", file.raw);
      }
      formData.append("fileType", type?type[1]:'');
      formData.append("fileName", fileName);//合并完Back of name
      formData.append("fileOriginName", file.name);
      formData.append("categoryId", this.uplodForm.knowValue);
      formData.append("plugin", this.uplodForm.plugin);
      formData.append("batchId", this.fileUuid);
      const cancel = axios.CancelToken.source(); //Create一个Cancel令牌
      this.source.push(cancel);
       importDoc(formData, cancel.token).then(res =>{
          if (res.code === 0) {
            if(Array.isArray(res.data)){
              //Compress包会Back多个id(ArrayFormat)
              this.$set(this.fileList[this.fileIndex],'id',res.data.join(','))
            }else{
              //单个FileUpload会Back一个id(String)
              this.$set(this.fileList[this.fileIndex],'id',res.data)
            }
            this.$set(this.fileList[this.fileIndex], "progressStatus", "success");
            this.$set(this.fileList[this.fileIndex], "percentage", 100);
            this.resultDisabled = false;
            this.$message.success(this.$t('knowledgeManage.uploadSuccess'));
            this.fileIndex++;
            if (this.fileIndex < this.fileList.length) {
              this.startUpload(this.fileIndex);
            }
          }
       }).catch(error =>{

       })

    },
    reset() {
      if (this.source.length > 0) {
        for (let i = 0; i < this.source.length; i++) {
          this.source[i].cancel();
        }
      }
      let ids = []
      if(this.fileList.length > 0){
        this.fileList.map(item => {
          if(item.id){
            if(item.id.includes(',')){//rag一体机No此逻辑
              const list = item.id.split(',')
              list.map(item =>{
                ids.push(item)
              })
            }else{
              ids.push(item.id)
            }
          }
        })
        if(ids.length > 0){
          this.deleteData({id:ids})//Cancel when DeleteFile
        }
      }
      this.$refs["uplodForm"].resetFields();
      this.uplodForm.knowValue = null;
      this.fileList = [];
      this.resultDisabled = true;
      this.source = [];
      this.fileUuid = '';
      this.$emit("handleSetOpen", {isShow:false,knowValue:null});
      this.uploading = false;
      this.$refs.urlUpload &&
        this.$refs.urlUpload.resetForm("dynamicValidateForm");
    },
    async deleteData(data){
       const res = await deleteInvalid(data)
       if(res.code === 0){
          this.$message.success(this.$t('knowledgeManage.clearSuccess'))
       }
    },
    async getClassfyDoc() {
      //GetDocument知识分Class
      const res = await getDocList();
      if (res && res.code === 0) {
        this.knowledgeData = res.data;
      } else {
        this.$message.error(res.message);
      }
    },
    normalizeOptions(node) {
      if (node.children == null || node.children == "null") {
        delete node.children;
      }
      return {
        id: node.id,
        label: node.categoryName,
        children: node.children,
      };
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
    handleRadioChange(val) {
      this.resultDisabled = true;
      this.uploading = false;
      this.urlLoading = false;
      this.fileList = [];
    },
    handleSave() {
      this.$refs["uplodForm"].validate((valid) => {
        this.$refs.urlUpload && this.$refs.urlUpload.handleSave();
      });
    },
    handlebatchSave() {
      this.$refs["uplodForm"].validate((valid) => {
        this.$refs.urlBatchUpload &&
          this.$refs.urlBatchUpload.handleSave(this.uplodForm.plugin);
      });
    },
    handleLoading(val, result) {
      this.urlLoading = val;
      if (result === "success") {
        this.reset();
      }
    },
    handleValidate() {
      this.$refs["uplodForm"].validate((valid) => {
        this.urlValidate = valid; // true/false
      });
    },
     handleSetBatchStatus(val) {
      this.urlLoading = val.value
      if (val.type === "result" && val.value === "success") {
        this.urlBatchDis = false;
        this.reset();
      }
    },
    // VerifyurlBatchUpload ButtonYesNo可点击
    handleSetBatchDisabled(val) {
      this.urlBatchDis = val;
    }
  },
  computed: {
    loadingResult() {
      for (let i = 0; i < this.fileList.length; i++) {
        if (this.fileList[i].loading === true) {
          return true;
        }
      }
      return false;
    },
  }
};
</script>
<style lang="scss">
.success{color: #67c23a;}
.fail{color: #e60001;}
.cust_Form {
  .fileNumber {
    margin-left: 10px;
    display: inline-block;
    padding: 0 20px;
    line-height: 2;
    background: rgb(243, 243, 243);
    border-radius: 8px;
  }
  .fileNumber{
    color: #606266 !important;
  }
  .uploadTips {
    color: #aaabb0;
    font-size: 12px;
    height: 30px;

    &:first-child {
      color: #e60001;
    }
  }
  .document_lise {
    list-style: none;
    li {
      font-size: 12px;
      padding: 7px;
      border-radius: 3px;
      line-height: 1;
      .el-icon-success {
        color: #67c23a;
      }
      .error{
        cursor: pointer;
        font-size: 18px;
      }
      .result_icon {
        float: right;
      }
      .size {
        font-weight: bold;
        margin-right: 10px;
        span{
          margin-left:30px;
          font-weight: unset !important;
        }
      }
    }
    .document_error {
      color: #e60001;
    }
  }
}
</style>
