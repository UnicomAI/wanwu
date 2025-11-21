<template>
  <div class="page-wrapper full-content">
    <div class="page-title">
      <i
        class="el-icon-arrow-left"
        @click="goBack"
        style="margin-right: 10px; font-size: 20px; cursor: pointer"
      >
      </i>
      {{knowledgeName}}
    </div>
    <div class="block table-wrap list-common wrap-fullheight">
      <el-container class="konw_container">
        <el-main class="noPadding">
          <el-container>
            <el-header class="classifyTitle">
              <div class="searchInfo">
                <el-select
                  @change="changeOption($event)"
                  v-model="docQuery.status"
                  :placeholder="$t('knowledgeManage.please')"
                  style="width:150px;"
                  class="marginRight no-border-select cover-input-icon"
                >
                  <el-option
                    v-for="item in knowLegOptions"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </el-select>
                <search-input class="cover-input-icon" :placeholder="$t('knowledgeManage.docPlaceholder')" ref="searchInput" @handleSearch="handleSearch" />
              </div>

              <div class="content_title">
                <el-button size="mini" type="primary" icon="el-icon-refresh" @click="reload" >{{$t('common.gpuDialog.reload')}}</el-button>
                <el-button size="mini" type="primary" @click="$router.push(`/knowledge/graphMap/${docQuery.knowledgeId}?name=${knowledgeName}`)" v-if="showGraphReport">{{$t('knowledgeManage.hitTest.graph')}}</el-button>
                <el-button size="mini" type="primary" @click="$router.push(`/knowledge/communityReport?knowledgeId=${docQuery.knowledgeId} &name=${knowledgeName}`)" v-if="showGraphReport">
                  <span>{{$t('knowledgeManage.hitTest.communityReport')}}</span>
                  <el-tooltip class="item" effect="dark" :content="$t('knowledgeManage.docList.communityReportTips')" placement="top">
                    <i class="el-icon-question" style="margin-left: 2px;"></i>
                  </el-tooltip>
                </el-button>
                <el-button size="mini" type="primary" @click="showMeta" v-if="[10,20,30].includes(permissionType)">{{$t('knowledgeManage.docList.metaDataManagement')}}</el-button>
                <el-button size="mini" type="primary" @click="$router.push(`/knowledge/hitTest?knowledgeId=${docQuery.knowledgeId}&graphSwitch=${graphSwitch}`)">{{$t('knowledgeManage.hitTest.name')}}</el-button>
                <el-button
                  size="mini"
                  type="primary"
                  :underline="false"
                  @click="handleUpload"
                  v-if="[10,20,30].includes(permissionType)"
                >{{$t('knowledgeManage.fileUpload')}}</el-button>
              </div>
            </el-header>
            <el-main
              class="noPadding"
              v-loading="tableLoading"
            >
              <el-alert
                :title="title_tips"
                type="warning"
                show-icon
                style="margin-bottom:10px;"
                v-if="showTips"
              ></el-alert>
              <el-table
                ref="dataTable"
                :data="tableData"
                style="width: 100%"
                :row-key="'docId'"
                :header-cell-style="{ background: '#F9F9F9', color: '#999999' }"
                @selection-change="handleSelectionChange"
              >
                <el-table-column
                  type="selection"
                  reserve-selection
                  v-if="[10,20,30].includes(permissionType)"
                  width="55">
                </el-table-column>
                <el-table-column
                  prop="docName"
                  :label="$t('knowledgeManage.fileName')"
                  min-width="180"
                >
                  <template slot-scope="scope">
                    <el-popover
                      placement="bottom-start"
                      :content="scope.row.docName"
                      trigger="hover"
                      width="300"
                    >
                      <span slot="reference">{{scope.row.docName.length>20?scope.row.docName.slice(0,20)+'...':scope.row.docName}}</span>
                    </el-popover>
                  </template>
                </el-table-column>
                <el-table-column
                  prop="docType"
                  :label="$t('knowledgeManage.fileStyle')"
                >
                </el-table-column>
                <el-table-column
                  prop="segmentMethod"
                  :label="$t('knowledgeManage.docList.segmentMode')"
                >
                <template slot-scope="scope">
                  <span>{{ getSegmentMethodText(scope.row.segmentMethod) }}</span>
                </template>
                </el-table-column>
                <el-table-column
                  prop="author"
                  :label="$t('knowledgeManage.author')"
                >
                </el-table-column>
                <el-table-column
                  prop="uploadTime"
                  :label="$t('knowledgeManage.importTime')"
                  width="200"
                >
                </el-table-column>
                <el-table-column
                  prop="status"
                  :label="$t('knowledgeManage.currentStatus')"
                  width="150"
                >
                  <template slot-scope="scope">
                    <span :class="[[4,5].includes(scope.row.status)?'error':'']">{{filterStatus(scope.row.status)}}</span>
                    <el-tooltip
                      class="item"
                      effect="light"
                      :content="scope.row.errorMsg?scope.row.errorMsg:''"
                      placement="top"
                      v-if="scope.row.status === 5"
                      popper-class="custom-tooltip"
                    >
                      <span
                        class="el-icon-warning"
                        style="margin-left:5px;color:#E6A23C;"
                      ></span>
                    </el-tooltip>
                  </template>
                </el-table-column>
                <el-table-column
                  v-if="graphSwitch"
                  prop="graphStatus"
                  :label="$t('knowledgeManage.graph.graphStatus')"
                >
                   <template slot-scope="scope">
                      <span>{{knowledgeGraphStatus[scope.row.graphStatus]}}</span>
                      <el-tooltip
                      class="item"
                      effect="light"
                      :content="scope.row.graphErrMsg?scope.row.graphErrMsg:''"
                      placement="top"
                      v-if="scope.row.graphStatus === 3"
                      popper-class="custom-tooltip"
                    >
                      <span
                        class="el-icon-warning"
                        style="margin-left:5px;color:#E6A23C;"
                      ></span>
                    </el-tooltip>
                   </template>
                </el-table-column>
                <el-table-column
                  :label="$t('knowledgeManage.operate')"
                  width="260"
                >
                  <template slot-scope="scope">
                    <el-button
                      size="mini"
                      round
                      @click="handleDel(scope.row)"
                      :disabled="[2,3].includes(Number(scope.row.status))"
                      v-if="[10,20,30].includes(permissionType)"
                      :type="[2,3].includes(Number(scope.row.status))?'info':''"
                    >{{$t('common.button.delete')}}</el-button>
                    <el-button
                      size="mini"
                      round
                      :type="[0,3,5].includes(Number(scope.row.status))?'info':''"
                      :disabled="[0,3,5].includes(Number(scope.row.status))"
                      @click="handleView(scope.row)"
                    >{{$t('knowledgeManage.view')}}</el-button>
                  </template>
                </el-table-column>
              </el-table>
              <!-- 分页 -->
              <Pagination
                class="pagination table-pagination"
                ref="pagination"
                :listApi="listApi"
                :page_size="10"
                @refreshData="refreshData"
              />
            </el-main>
          </el-container>
        </el-main>
      </el-container>
    </div>
    <!-- Metadata Management -->
    <el-dialog
      :title="$t('knowledgeManage.docList.metaDataManagement')"
      :visible.sync="metaVisible"
      width="550px"
      :before-close="handleClose">
      <mataData ref="mataData" @updateMeata="updateMeata" type="create" :knowledgeId="docQuery.knowledgeId" class="mataData"/>
      <span slot="footer" class="dialog-footer">
        <el-button @click="handleClose">{{$t('common.button.cancel')}}</el-button>
        <el-button type="primary" @click="createMeta">{{$t('common.button.create')}}</el-button>
        <el-button type="primary" @click="submitMeta" :disabled="isDisabled">{{$t('common.button.confirm')}}</el-button>
      </span>
    </el-dialog>
    
    <!-- BatchEditMetadataValue弹窗 -->
    <batchMetaData ref="batchMetaData" :selectedDocIds="selectedDocIds" @reLoadDocList="reLoadDocList" />
    <!-- BatchEditMetadataValueOperation框 -->
    <BatchMetatButton ref="BatchMetatButton" :selectedCount="selectedTableData.length" @showBatchMeta="showBatchMeta" @handleMetaCancel="handleMetaCancel"/>
  </div>
</template>

<script>
import Pagination from "@/components/pagination.vue";
import SearchInput from "@/components/searchInput.vue";
import mataData from './metadata.vue'
import batchMetaData from './meta/batchMetaData.vue'
import BatchMetatButton from './meta/batchMetatButton.vue'
import {getDocList,delDocItem,uploadFileTips,updateDocMeta} from "@/api/knowledge";
import {mapGetters} from 'vuex';
import { KNOWLEDGE_GRAPH_STATUS } from '../config';
export default {
  components: { Pagination,SearchInput,mataData,batchMetaData,BatchMetatButton},
  data() {
    return {
      knowledgeName:'',
      loading:false,
      tableLoading:false,
      docQuery: {
        docName:'',
        knowledgeId:this.$route.params.id,
        status: -1
      },
      fileList: [],
      listApi: getDocList,
      title_tips:'',
      showTips:false,
      tableData: [],
      knowLegOptions:this.getKnowOptions(),
      knowledgeData: [],
      currentKnowValue:null,
      timer:null,
      refreshCount:0,
      tagList:[],
      metaVisible:false,
      metaData:[],
      isDisabled:false,
      selectedTableData:[],
      selectedDocIds:[],
      graphSwitch:false,
      showGraphReport:false,
      knowledgeGraphStatus: KNOWLEDGE_GRAPH_STATUS
    };
  },
  watch:{
    '$route':{
      handler(val){
        if(val.query.done){
          this.startTimer()
        }
      },
      immediate:true
    },
    metaData:{
      handler(val){
        if(val.some(item => !item.metaKey || !item.metaValueType) || !val.length){
          this.isDisabled = true
        }else{
          this.isDisabled = false
        }
      }
    }
  },
  computed: {
    ...mapGetters('app', ['permissionType'])
  },
  mounted(){
    this.getTableData(this.docQuery)
    if (this.permissionType === -1 || this.permissionType === null || this.permissionType === undefined) {
        const savedData = localStorage.getItem('permission_data')
        if (savedData) {
            try {
                const parsed = JSON.parse(savedData)
                const savedPermissionType = parsed && parsed.app && parsed.app.permissionType
                if (savedPermissionType !== undefined && savedPermissionType !== -1) {
                    this.$store.dispatch('app/setPermissionType', savedPermissionType)
                }
            } catch(e) {
            }
        }
    }
  },
  beforeDestroy(){
    this.clearTimer()
  },
  methods: {
    handleMetaCancel(){
      this.selectedTableData = []
      this.selectedDocIds = []
      // Cancel所HasTableData of 选inStatus
      this.$nextTick(() => {
        const table = this.$refs.dataTable
        if (table) {
          table.clearSelection()
        }
      })
    },
    reLoadDocList(){
      this.getTableData(this.docQuery)
      this.selectedTableData = []
      this.selectedDocIds = []
      
      // Cancel所HasTableData of 选inStatus
      this.$nextTick(() => {
        const table = this.$refs.dataTable
        if (table) {
          table.clearSelection()
        }
      })
      
    },
    showBatchMeta(){
      if(!this.selectedTableData || this.selectedTableData.length === 0){
        this.$message.warning(this.$t('knowledgeManage.docList.pleaseSelectDocFirst'));
        return;
      }
      this.$refs.batchMetaData.showDialog();
    },
    handleSelectionChange(val){ 
      if (val.length > 100) {
        this.$message.warning(this.$t('knowledgeManage.docList.maxSelect100Files'));
        return;
      }
      this.selectedTableData = val
      this.selectedDocIds = val.map(item => item.docId)
    },
    getSegmentMethodText(value){
      switch (value) {
        case '0':
          return this.$t('knowledgeManage.config.commonSegment');
        case '1':
          return this.$t('knowledgeManage.config.parentSonSegment');
        default:
          return this.$t('knowledgeManage.docList.unknown');
      }
    },
    createMeta(){
      this.$refs.mataData.createMetaData();
      this.scrollToBottom();
    },
    scrollToBottom() {
      this.$nextTick(() => {
        const container = this.$refs.mataData;
        if (container) {
          container.scrollTop = container.scrollHeight;
        }
      });
     },
    submitMeta(){
      this.isDisabled = true;
      const metaList = this.metaData.filter(item => item.option !== '').map(({metaId,metaKey,metaValueType,option}) =>({
          metaKey,
          ...(option === 'add' ? {metaValueType } : {}),
          option,
          ...(option === 'update'||option === 'delete' ? {metaId } : {})
        }))
      const data = {
        docId:'',
        knowledgeId:this.docQuery.knowledgeId,
        metaDataList:metaList
      }
      updateDocMeta(data).then(res =>{
        if(res.code === 0){
          this.$message.success(this.$t('common.message.success'));
          this.$refs.mataData.getList();
          this.metaVisible = false;
          this.isDisabled = false;
        }
      }).catch(() =>{
          this.isDisabled = false;
      })
    },
    showMeta(){
      this.metaVisible = true;
    },
    updateMeata(data){
      this.metaData = data
    },
    handleClose(){
      this.metaVisible = false;
    },
    startTimer(){
      this.clearTimer();
      if (this.refreshCount >= 2) {
        return;
      }
      const delay = this.refreshCount === 0 ? 1000 : 3000;
      this.timer = setTimeout(() =>{
        this.getTableData(this.docQuery)
        this.refreshCount++;
        this.startTimer()
      },delay)
    },
    clearTimer() {
      if (this.timer) {
        clearInterval(this.timer);
        this.timer = null;
      }
    },
    goBack(){
      this.$router.push({path:'/knowledge'})
    },
    reload(){
      this.getTableData(this.docQuery)
    },
    handleSearch(val){
      this.docQuery.docName = val;
      this.getTableData(this.docQuery)
    },
    getKnowOptions(){
      const commonOptions = [
        { label:this.$t('knowledgeManage.all'), value: -1 },
        { label: this.$t('knowledgeManage.finish'), value: 1 },
        { label: this.$t('knowledgeManage.fail'), value: 5 },
        { label: this.$t('knowledgeManage.analysising'), value: 3 },
        { label: this.$t('knowledgeManage.checkFail'), value: 4 },
        { label: this.$t('knowledgeManage.pendingProcessing'), value: 0 },
        { label: this.$t('knowledgeManage.checking'), value: 2 }
      ];
      return commonOptions;
    },
    handleRemove(item){
      this.fileList = this.fileList.filter((files) => files.name !== item.name);
    },
    handelText(data){
      if(data.length > 0){
        return data.join(',')
      }
    },
    handleHit(){//jumpto命in预测Page
      if(this.$route.path.includes('rag')){
        this.$router.push({path:'/rag/knowledge/hitTest'})
      }else{
        this.$router.push({path:'/knowledge/hitTest'})
      }
    },
    normalizeOptions(node){
      if(node.children == null || node.children == 'null'){
        delete node.children;
      }
      return {
        id: node.id,
        label: node.categoryName,
        children: node.children,
      };
    },
    submitDocname(formName){
      this.$refs[formName].validate((valid) =>{
        if(valid){
          this.modifyDoc({id:this.currentDocdata.id,docName:this.ruleForm.docName})
        }
      })
    },
    async modifyDoc(data){
      this.loading = true;
      const res = await modifyDoc(data)
      if(res.code === 0){
        this.$message.success(this.$t('knowledgeManage.operateSuccess'))
        this.docListVisible = false
        this.getTableData(this.docQuery)
      }
      this.loading = false;
    },
    handleEdit(data){
      this.ruleForm.docName = data.docName
      this.docListVisible = true
      this.currentDocdata = data
    },
    handleDel(data){
       this.$confirm(this.$t('knowledgeManage.deleteTips'),this.$t('knowledgeManage.tip'),
        {
          confirmButtonText:  this.$t('common.button.confirm'),
          cancelButtonText: this.$t('common.button.cancel'),
          type: "warning"
        }
      )
        .then(async () => {
          let jsondata = {docIdList:[data.docId],knowledgeId:this.docQuery.knowledgeId,}
          this.loading = true;
          let res = await delDocItem(jsondata);
          if (res.code === 0) {
            this.$message.success(this.$t('common.info.delInfo'));
            this.getTableData(this.docQuery)//Get知识分ClassData
          }
          this.loading = false;
        })
        .catch((error) => {
          this.getTableData(this.docQuery)
        });
    },
    async getTableData(data){
       this.tableLoading = true;
       this.tableData = await this.$refs["pagination"].getTableData(data);
       this.tableLoading = false;
       this.getTips();
    },
    getTips(){
        uploadFileTips({knowledgeId:this.docQuery.knowledgeId}).then(res =>{
          if(res.code === 0){
              if(res.data.uploadstatus === 1){
                 this.showTips = true
                 this.title_tips = this.$t('knowledgeManage.refreshTips')
              }else if(res.data.uploadstatus === 2){
                  this.showTips = false
                  this.title_tips = ''
              }else{
                this.showTips = true
                this.title_tips = res.data.msg
              }
          }
        })
    },
    changeOption(data) {
      //ByDocument StatusFind
      this.docQuery.status = data
      this.getTableData({...this.docQuery, pageNo: 1})
    },
    filterStatus(status) {
      switch (status) {
        case 0:
          return this.$t('knowledgeManage.pendingProcessing');
          break;
        case 1:
          return this.$t('knowledgeManage.finish');
          break;
        case 2:
          return this.$t('knowledgeManage.checking');
          break;
        case 3:
          return this.$t('knowledgeManage.analysising');
          break;
        case 4:
          return this.$t('knowledgeManage.checkFail');
          break;
        case 5:
          return this.$t('knowledgeManage.fail');
          break;
        case -2:
          return this.$t('knowledgeManage.beUploaded');
          break;
        default:
          return this.$t('knowledgeManage.noStatus');
      }
    },

    handleView(row){
      this.$router.push({
        path: "/knowledge/section",
        query: {
          id: row.docId,
          type: row.docType,
          name:row.docName,
          knowledgeId:row.knowledgeId,
          knowledgeName:this.knowledgeName
        },
      });
    },
    async download(url,name){
      const res = await downDoc(url)
      const blobUrl = window.URL.createObjectURL(res) // 将blobObject转is一URL
      const link = document.createElement('a')
      link.href = blobUrl
      link.download = name
      link.click() // StartDownload
      window.URL.revokeObjectURL(link.href) // Download完毕DeleteaTag
    },
    handleUpload() {
      this.$router.push({path:'/knowledge/fileUpload',query:{id:this.docQuery.knowledgeId,name:this.knowledgeName}})
    },
    refreshData(data,tableInfo) {
      this.tableData = data
      if(tableInfo && tableInfo.docKnowledgeInfo){
        this.graphSwitch = tableInfo.docKnowledgeInfo.graphSwitch === 1 ? true : false
        this.showGraphReport = tableInfo.docKnowledgeInfo.showGraphReport
        this.knowledgeName = tableInfo.docKnowledgeInfo.knowledgeName
      }else{
        this.graphSwitch = false
        this.showGraphReport = false
      }
    }
  }
};
</script>
<style lang="scss" scoped>
.mataData{
  max-height:400px;
  overflow-y: auto;
}
.edit-icon{
  color: $color;
  cursor: pointer;
  font-size: 16px;
  margin-left: 5px;
}
.doc_tag{
  margin:0 2px;
}
/deep/ {
  .el-button.is-disabled,
  .el-button--info.is-disabled {
    color: #c0c4cc !important;
    background-color: #fff !important;
    border-color: #ebeef5 !important;
  }
  .el-tree--highlight-current
    .el-tree-node.is-current
    > .el-tree-node__content {
    background: #ffefef;
  }
  .el-tabs__item.is-active {
    color: #e60001 !important;
  }
  .el-tabs__active-bar {
    background-color: #e60001 !important;
  }
  .el-tabs__content {
    width: 100%;
    height: calc(100% - 40px);
  }
  .el-tab-pane {
    width: 100%;
    height: 100%;
  }
  .el-tree .el-tree-node__content {
    height: 40px;
  }
  .custom-tree-node {
    padding: 0 10px;
  }
  .el-tree .el-tree-node__content:hover {
    background: #ffefef;
  }
  .el-tree-node__expand-icon {
    display: none;
  }
  .el-button.is-round {
    border-color: #dcdfe6;
    color: #606266;
  }
  .el-upload-list {
    max-height: 200px;
    overflow-y: auto;
  }
  .el-dialog__body{
    padding:10px 20px;
  }
}
.fileNumber {
  margin-left: 10px;
  display: inline-block;
  padding: 0 20px;
  line-height: 2;
  background: rgb(243, 243, 243);
  border-radius: 8px;
}
.defalutColor {
  color: #e7e7e7 !important;
}
.border {
  border: 1px solid #e4e7ed;
}
.noPadding {
  padding: 0 10px;
}
.activeColor {
  color: #e60001;
}
.error {
  color: #e60001;
}
.marginRight {
  margin-right: 10px;
}
.full-content {
  //padding: 20px 20px 30px 20px;
  margin: auto;
  overflow: auto;
  //background: #fafafa;
  .title {
    font-size: 18px;
    font-weight: bold;
    color: #333;
    padding: 10px 0;
  }
  .tips {
    font-size: 14px;
    color: #aaabb0;
    margin-bottom: 10px;
  }
  .block {
    width: 100%;
    height: calc(100% - 58px);
    .el-tabs {
      width: 100%;
      height: 100%;
      .konw_container {
        width: 100%;
        height: 100%;
        .tree {
          height: 100%;
          background: none;
          .custom-tree-node {
            width: 100%;
            display: flex;
            justify-content: space-between;
            .icon {
              font-size: 16px;
              transform: rotate(90deg);
              color: #aaabb0;
            }
            .nodeLabel {
              color: #e60001;
              display: flex;
              align-items: center;
              .tag {
                display: block;
                width: 5px;
                height: 5px;
                border-radius: 50%;
                background: #e60001;
                margin-right: 5px;
              }
            }
          }
        }
      }
    }
    .classifyTitle {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 0 10px;
      h2 {
        font-size: 16px;
      }
      .content_title {
        display: flex;
        align-items: center;
        justify-content: flex-end;
      }
    }
  }
  .uploadTips {
    color: #aaabb0;
    font-size: 12px;
    height: 30px;
  }
  .document_lise {
    list-style: none;
    li {
      display: flex;
      justify-content: space-between;
      font-size: 12px;
      padding: 7px;
      border-radius: 3px;
      line-height: 1;
      .el-icon-success {
        display: block;
      }
      .el-icon-error {
        display: none;
      }
      &:hover {
        cursor: pointer;
        background: #eee;

        .el-icon-success {
          display: none;
        }
        .el-icon-error {
          display: block;
        }
      }
      &.document_loading {
        &:hover {
          cursor: pointer;
          background: #eee;

          .el-icon-success {
            display: none;
          }
          .el-icon-error {
            display: none;
          }
        }
      }
      .el-icon-success {
        color: #67c23a;
      }

      .result_icon {
        float: right;
      }
      .size {
        font-weight: bold;
      }
    }
    .document_error {
      color: red;
    }
  }
}
</style>
<style lang="scss">
.custom-tooltip.is-light {
  border-color: #eee; /* Set border color */
  background-color: #fff; /* Set background color */
  color: #666; /* Set text color */
}
.custom-tooltip.el-tooltip__popper[x-placement^="top"] .popper__arrow::after {
  border-top-color: #fff !important;
}
.custom-tooltip.el-tooltip__popper.is-light[x-placement^="top"] .popper__arrow {
  border-top-color: #ccc !important;
}
</style>