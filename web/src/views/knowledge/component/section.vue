<template>
  <div class="section" v-loading="loading.itemStatus">
    <div class="title">
      <i
        class="el-icon-arrow-left"
        @click="$router.go(-1)"
        style="margin-right: 20px; font-size: 20px; cursor: pointer"
      ></i
      >{{ obj.name }}
    </div>
    <div class="container">
      <el-descriptions
        class="margin-top"
        title=""
        :column="3"
        :size="''"
        border
      >
        <el-descriptions-item :label="$t('knowledgeManage.fileName')">{{
          res.fileName
        }}</el-descriptions-item>
        <el-descriptions-item :label="$t('knowledgeManage.splitNum')">
          {{ res.segmentTotalNum }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('knowledgeManage.importTime')">{{
          res.uploadTime
        }}</el-descriptions-item>
        <el-descriptions-item :label="$t('knowledgeManage.chunkType')">{{
          Number(res.segmentType) === 0 ? $t('knowledgeManage.autoChunk') : $t('knowledgeManage.autoConfigChunk')
        }}</el-descriptions-item>
        <el-descriptions-item :label="$t('knowledgeManage.setMaxLength')">{{
          String(res.maxSegmentSize)
        }}</el-descriptions-item>
        <el-descriptions-item :label="$t('knowledgeManage.markSplit')">{{
          String(res.splitter).replace(/\n/g, '\\n')
        }}</el-descriptions-item>
        <el-descriptions-item label="Metadata">
          <template v-if="metaDataList && metaDataList.length > 0">
            <span
                v-for="(item, index) in metaDataList.slice(0, 3)"
                :key="index"
                class="metaItem"
            >
              {{ item.metaKey }}: {{ item.metaValueType === 'time' ? formatTimestamp(item.metaValue) : item.metaValue }}
            </span>
            <el-tooltip v-if="metaDataList.length > 3" :content="filterData(metaDataList.slice(3))" placement="bottom">
              <span class="metaItem">...</span>
            </el-tooltip>
          </template>
          <span v-else>No data</span>
          <span class="el-icon-edit-outline editIcon" @click="showDatabase(metaDataList || [])" v-if="metaDataList && [10,20,30].includes(permissionType)"></span>
        </el-descriptions-item>
        <el-descriptions-item label="Metadata Rules">
          <template v-if="metaRuleList && metaRuleList.length > 0">
            <span v-for="(item, index) in metaRuleList.slice(0, 3)" :key="index" class="metaItem">
              {{ item.metaKey }}: {{ item.metaRule }}<span v-if="index < metaRuleList.slice(0, 3).length - 1"> </span>
            </span>
            <el-tooltip v-if="metaRuleList.length > 3" :content="filterRule(metaRuleList.slice(3))" placement="bottom">
              <span class="metaItem">...</span>
            </el-tooltip>
          </template>
          <span v-else>No data</span>
        </el-descriptions-item>
        <el-descriptions-item label="Batch Add Segment Status">
          <span>{{res.segmentImportStatus}}</span>
        </el-descriptions-item>
      </el-descriptions>

      <div class="btn">
         <el-button
          type="primary"
          @click="createChunk(false)"
          size="mini"
          :loading="loading.start"
          v-if="[10,20,30].includes(permissionType)"
          >Add Segment</el-button
        >
        <el-button
          type="primary"
          @click="handleStatus('start')"
          size="mini"
          :loading="loading.start"
          v-if="[10,20,30].includes(permissionType)"
          >{{$t('knowledgeManage.allRun')}}</el-button
        >
        <el-button
          type="primary"
          @click="handleStatus('stop')"
          size="mini"
          :loading="loading.stop"
          v-if="[10,20,30].includes(permissionType)"
          >{{$t('knowledgeManage.allStop')}}</el-button
        >
      </div>

      <div class="card">
        <el-row :gutter="20" v-if="res.contentList.length > 0">
          <el-col
            :span="6"
            v-for="(item, index) in res.contentList"
            :key="index"
            class="card-box"
          >
            <el-card class="box-card">
              <div slot="header" class="clearfix">
                <span>
                  {{ $t('knowledgeManage.split')+":" + item.contentNum }}
                  <span class="segment-type">#{{ item.isParent?"Parent-child Segment":"General Segment" }}</span>
                  <span class="segment-length" v-if="!item.isParent">#{{ item.content.length }}{{$t('knowledgeManage.character')}}</span>
                  <span class="segment-child" v-if="item.childNum">#{{ item.childNum || 0 }} child segments</span>
                </span>
                <div>
                  <el-switch
                    style="padding: 3px 0;"
                    v-model="item.available"
                    active-color="var(--color)"
                    v-if="[10,20,30].includes(permissionType)"
                    @change="handleStatusChange(item, index)"
                  >
                  </el-switch>
                  <el-dropdown @command="handleCommand" placement="bottom" v-if="[10,20,30].includes(permissionType)">
                    <span class="el-dropdown-link">
                      <i class="el-icon-more more"></i>
                    </span>
                    <el-dropdown-menu slot="dropdown">
                      <el-dropdown-item class="card-delete" :command="{type: 'delete', item}">
                        <i class="el-icon-delete card-opera-icon" />
                        {{$t('common.button.delete')}}
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </el-dropdown>
                </div>
              </div>
              <div class="text item" @click="handleClick(item, index)">
                {{ item.content }}
              </div>
              <div class="tagList" v-if="[10,20,30].includes(permissionType)">
                <span :class="['smartDate','tagList']" @click.stop="addTag(item.labels,item.contentId)" v-if="item.labels.length === 0">
                  <span class="el-icon-price-tag icon-tag"></span>
                  CreateKeyword
                </span>
                <span class="tagList-item" @click.stop="addTag(item.labels,item.contentId)" v-else>{{formattedTagNames(item.labels) }}</span>
              </div>
            </el-card>
          </el-col>
        </el-row>
        <el-empty v-else :description="$t('knowledgeManage.noData')"></el-empty>
      </div>

      <div class="list-common" style="text-align: right">
        <el-pagination
          background
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
          :current-page="page.pageNo"
          :page-sizes="page.pageSizeList"
          :page-size="page.pageSize"
          layout="total, prev, pager, next, jumper"
          :total="page.total"
        >
        </el-pagination>
      </div>
    </div>

    <el-dialog
      v-if="dialogVisible"
      :title="$t('knowledgeManage.detailView')"
      :visible.sync="dialogVisible"
      width="60%"
      :show-close="false"
      v-loading="loading.dialog"
      class="section-dialog"
    >
      <div slot="title">
        <span style="font-size: 16px">{{$t('knowledgeManage.detailView')}}</span>
        <el-switch
          @change="handleDetailStatusChange"
          style="float: right; padding: 3px 0"
          v-model="cardObj[0].available"
          active-color="var(--color)"
          v-if="[10,20,30].includes(permissionType)"
        >
        </el-switch>
      </div>
      <div class="dialog-content">
        <el-table
          :data="cardObj"
          border
          style="width: 100%"
          :header-cell-style="{
            background: '#F9F9F9',
            color: '#999999',
          }"
        >
          <el-table-column
            prop="content"
            align="center"
            :render-header="renderHeader"
          >
          <template slot-scope="scope">
              <el-input 
                type="textarea"
                v-model="scope.row.content"
                :autosize="{ minRows: 3, maxRows: 5}"
                class="full-width-textarea"
                :disabled="[0].includes(permissionType)"
                >
              </el-input>
              <div  v-if="cardObj[0]['isParent'] && [10,20,30].includes(permissionType)" style="display: flex;justify-content: flex-end;padding: 10px 0;">
                <el-button type="primary" @click="handleSubmit"  :loading="submitLoading">Save and reparse child segments</el-button>
              </div>
              <div class="segment-list" v-if="scope.row.childContent.length > 0">
                <el-collapse 
                  v-model="activeNames" 
                  class="section-collapse"
                >
                  <el-collapse-item 
                    v-for="(segment, index) in scope.row.childContent" 
                    :key="index"
                    :name="index"
                    class="segment-collapse-item"
                  >
                    <template slot="title">
                      <span class="segment-badge">C-{{ index + 1 }}</span>
                      <div class="segment-actions">
                        <span v-if="!editingSegments[`${scope.row.contentId}-${index}`] && [10,20,30].includes(permissionType)" class="action-btn edit-btn" @click.stop="editSegment(scope.row, index)">
                          <i class="el-icon-edit-outline"></i>Edit
                        </span>
                        <span v-if="!editingSegments[`${scope.row.contentId}-${index}`] && [10,20,30].includes(permissionType)" class="action-btn delete-btn" @click.stop="deleteSegment(scope.row, index)">
                          <i class="el-icon-delete"></i>Delete
                        </span>
                        <span v-if="editingSegments[`${scope.row.contentId}-${index}`]" class="action-btn save-btn" @click.stop="confirmEdit(scope.row, index)">
                          <i class="el-icon-check"></i>Save
                        </span>
                        <span v-if="editingSegments[`${scope.row.contentId}-${index}`]" class="action-btn cancel-btn" @click.stop="cancelEdit(scope.row, index)">
                          <i class="el-icon-close"></i>Cancel
                        </span>
                      </div>
                    </template>
                    <div class="segment-content">
                      <div v-if="!editingSegments[`${scope.row.contentId}-${index}`]" class="content-display">
                        {{ segment.content }}
                      </div>
                      <div v-else class="content-edit">
                        <el-input
                          v-model="editingContent[`${scope.row.contentId}-${index}`]"
                          type="textarea"
                          :rows="3"
                          placeholder="Please enterContent"
                          class="edit-input"
                        />
                      </div>
                    </div>
                  </el-collapse-item>
                </el-collapse>
              </div>
          </template>
          </el-table-column>
        </el-table>
      </div>

      <span slot="footer" class="dialog-footer">
        <el-button type="primary" @click="handleSubmit" :loading="submitLoading" v-if="!cardObj[0]['isParent']">Confirm</el-button>
        <el-button type="primary" @click="createChunk(true)" v-if="cardObj[0]['isParent'] && [10,20,30].includes(permissionType)"" :disabled="submitLoading">Add Child Segment</el-button>
        <el-button type="primary" @click="handleClose" :disabled="submitLoading">{{$t('knowledgeManage.close')}}</el-button>
      </span>
    </el-dialog>
    <dataBaseDialog ref="dataBase" @updateData="updateData" :knowledgeId="obj.knowledgeId" :name="obj.knowledgeName"/>
    <tagDialog ref="tagDialog" type="section" :title="title" :currentList="currentList" @sendList="sendList" />
    <createChunk ref="createChunk"  @updateDataBatch="updateDataBatch" @updateData="updateData" :parentId="cardObj[0]['contentId']" @updateChildData="updateChildData"/>
  </div>
</template>
<script>
import { getSectionList,setSectionStatus,sectionLabels,delSegment,editSegment,getSegmentChild,delSegmentChild,updateSegmentChild } from "@/api/knowledge";
import dataBaseDialog from './dataBaseDialog';
import tagDialog from './tagDialog.vue';
import createChunk from './chunk/createChunk.vue'
import {mapGetters} from 'vuex';
export default {
  components:{dataBaseDialog,tagDialog,createChunk},
  data() {
    return {
      submitLoading:false,
      oldContent:'',
      title:'CreateKeyword',
      dialogVisible: false,
      editingSegments: {},
      editingContent: {},
      obj: {},
      cardObj: [
        {
          available: false,
          content: "",
          childContent:[],
          contentId: "",
          len: 20,
        },
      ],
      value: true,
      activeStatus: false,
      activeNames: [], 
      page: {
        pageNo: 1,
        pageSize: 8,
        pageSizeList: [10, 15, 20, 50],
        total: 0,
      },
      loading: {
        start: false,
        stop: false,
        itemStatus: false,
        dialog: false,
      },
      res: {
        contentList: [],
      },
      metaDataList: [],
      metaRuleList: [],
      currentList:[],
      contentId:'',
      timer:null,
      refreshCount:0,
    };
  },
  computed: {
    ...mapGetters('app', ['permissionType'])
  },
  created() {
    this.obj = this.$route.query;
    this.getList();
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
    createChunk(isChildChunk){
      this.$refs.createChunk.showDiglog(this.obj.id,isChildChunk)
    },
    updateChildData(){
      setTimeout(() => {
        this.handleParse();
      }, 1000);
    },
    formatScore(score) {
      if (typeof score !== 'number') {
        return '0.00000';
      }
      return score.toFixed(5);
    },
    editSegment(row, index) {
      const key = `${row.contentId}-${index}`;
      this.$set(this.editingSegments, key, true);
      this.$set(this.editingContent, key, row.childContent[index].content);
      
      this.$nextTick(() => {
        if (!this.activeNames.includes(index)) {
          this.activeNames.push(index);
        }
      });
    },
    cancelEdit(row, index) {
      const key = `${row.contentId}-${index}`;
      this.$set(this.editingSegments, key, false);
      this.$delete(this.editingContent, key);
    },
    confirmEdit(row, index) {
      const key = `${row.contentId}-${index}`;
      const newContent = this.editingContent[key];
      
      if (!newContent || newContent.trim() === '') {
        this.$message.warning('ContentCannot be empty');
        return;
      }
      updateSegmentChild({
        childChunk:{
          content: newContent.trim(),
          chunkNo:row['childContent'][index].childNum
        },
        docId: this.obj.id,
        parentChunkNo: row.contentNum,
        parentId: row.contentId
      }).then(res => {
        if (res.code === 0) {
          this.$message.success('UpdateSuccess');
          this.handleParse();
          this.$set(this.editingSegments, key, false);
          this.$delete(this.editingContent, key);
        } else {
          this.$message.error('UpdateFailed');
        }
      }).catch(() => {
        this.$message.error('UpdateFailed');
      });
    },
    handleParse(){
      getSegmentChild({contentId:this.cardObj[0]['contentId'],docId:this.obj.id}).then(res =>{
        if(res.code === 0){
          this.cardObj[0].childContent = res.data.contentList || [];
          this.activeNames = this.cardObj[0].childContent.map((_, index) => index);
        }
      }).catch(() =>{})
    },
    deleteSegment(row, index) {
      this.$confirm('ConfirmtoDelete这 child segments吗？', 'Tip', {
        confirmButtonText: 'Confirm',
        cancelButtonText: 'Cancel',
        type: 'warning'
      }).then(() => {
        delSegmentChild({docId:this.obj.id,parentId:row['childContent'][index].parentId,parentChunkNo:row.contentNum,ChildChunkNoList:[row['childContent'][index].childNum]}).then(res =>{
          if(res.code === 0){
            this.$message.success('DeleteSuccess');
            this.handleParse();
          }
        }).catch(() => {
          this.$message.error('DeleteFailed');
        });
      });
    },
    updateDataBatch(){
      this.startTimer();
    },
    startTimer(){
      this.clearTimer();
      if (this.refreshCount >= 2) {
        return;
      }
      const delay = this.refreshCount === 0 ? 1000 : 3000;
      this.timer = setTimeout(() =>{
        this.getList()
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
    handleSubmit(){
      const hasChanges = this.oldContent !== this.cardObj[0]['content'];
      
      if(!hasChanges){
        this.$message.warning('No changes')
        return false;
      }
      
      this.submitLoading = true;
      editSegment({content:this.cardObj[0]['content'],contentId:this.cardObj[0]['contentId'],docId:this.obj.id}).then(res =>{
        if(res.code === 0){
          this.$message.success('OperationSuccess');
            this.dialogVisible = false;
            this.submitLoading = false;
            this.getList();
        }
      }).catch(() =>{
        this.submitLoading = false;
      })
    },
    handleCommand(value){
      const {type, item} = value || {}
       switch (type) {
          case 'delete':
            this.delSection(item)
            break
        }
    },
    delSection(item){
      delSegment({contentId:item.contentId,docId:this.obj.id}).then(res =>{
        if(res.code === 0){
          this.$message.success('DeleteSuccess');
          this.getList();
        }
      }).catch(() =>{})
    },
    sendList(data){
      const labels = data.map(item => item.tagName)
      sectionLabels({contentId:this.contentId,docId:this.obj.id,labels}).then(res =>{
        if(res.code === 0){
          this.getList();
          this.$refs.tagDialog.handleClose();
        }
      }).catch(err =>{})
    },
    addTag(data,id){
      if(data.length > 0){
          this.currentList = data.map(item =>({
          tagName:item,
          checked: false,
          showDel: false,
          showIpt: false
        }))
      }else{
        this.currentList = []
      }
      this.contentId = id
      this.$refs.tagDialog.showDiaglog();
    },
    formattedTagNames(data){
      let tags = ''
      if(!Array.isArray(data) || data.length === 0){
        return '';
      }
      if(data.length > 3){
        tags = data.slice(0, 3).join(', ') + (data.length > 3 ? '...' : '');
      }else{
        tags = data.join(', ');
      }
      return tags;
    },
    updateData(){
      this.getList();
    },
    showDatabase(data){
      this.$refs.dataBase.showDiglog(data,this.obj.id)
    },
    filterData(data){
      return data.map(item => {
        let value = item.metaValue;
        if (item.metaValueType === 'time') {
          value = this.formatTimestamp(value);
        }
        return `${item.metaKey}:${value}`;
      }).join(", ");
    },
    formatTimestamp(timestamp) {
      if (timestamp === '') return '';
      const date = new Date(Number(timestamp));
      const year = date.getFullYear();
      const month = String(date.getMonth() + 1).padStart(2, '0');
      const day = String(date.getDate()).padStart(2, '0');
      const hours = String(date.getHours()).padStart(2, '0');
      const minutes = String(date.getMinutes()).padStart(2, '0');
      const seconds = String(date.getSeconds()).padStart(2, '0');
      return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
    },
    filterRule(rule){
      return rule.map(item => `${item.metaKey}:${item.metaRule}`).join(", ")
    },
    getList() {
      this.loading.itemStatus = true;
      getSectionList({
        docId: this.obj.id,
        pageNo: this.page.pageNo,
        pageSize:this.page.pageSize
      })
        .then((res) => {
          this.loading.itemStatus = false;
          this.res = res.data;
          this.page.total = this.res.segmentTotalNum;
          this.metaRuleList = res.data.metaDataList.filter(item => item.metaRule);
          this.metaDataList = res.data.metaDataList;
        })
        .catch(() => {
          this.loading.itemStatus = false;
        });
    },
    handleClick(item, index) {
      this.dialogVisible = true;
      this.oldContent = item.content;
      const obj = JSON.parse(JSON.stringify(item));
      this.$nextTick(() => {
        this.$set(obj,'childContent',[]);
        this.cardObj = [obj];
        if(this.cardObj[0].isParent){
          this.handleParse();
        }
        this.activeStatus = obj.available;
        this.activeNames = [];
      });
    },
    handleCurrentChange(val) {
      this.page.pageNo = val;
      this.getList();
    },
    handleSizeChange(val) {
      this.page.pageSize = val;
      this.getList();
    },
    handleDetailStatusChange(val) {
      this.loading.dialog = true;
      setSectionStatus({
        docId: this.obj.id,
        contentStatus: String(val),
        contentId: this.cardObj[0].contentId,
        all:false,
      })
        .then((res) => {
          this.loading.dialog = false;
          if (res.code === 0) {
            this.$message.success(this.$t('knowledgeManage.operateSuccess'));
          } else {
            this.cardObj[0].available = !this.cardObj[0].available;
          }
        })
        .catch(() => {
          this.loading.dialog = false;
          this.cardObj[0].contentStatus = !this.cardObj[0].contentStatus;
        });
    },
    handleStatusChange(item, index) {
      this.loading.itemStatus = true;
      setSectionStatus({
        docId: this.obj.id,
        contentStatus: String(item.available),
        contentId: item.contentId,
        all: false,
      })
        .then((res) => {
          this.loading.itemStatus = false;
          if (res.code === 0) {
            this.$message.success(this.$t('knowledgeManage.operateSuccess'));
            this.getList();
          } else {
            this.res.contentList[index].available =
              !this.res.contentList[index].available;
          }
        })
        .catch(() => {
          this.res[index].contentStatus = !this.res[index].contentStatus;
          this.loading.itemStatus = false;
        });
    },
    handleStatus(type) {
      this.loading.itemStatus = true;
      setSectionStatus({
        docId: this.obj.id,
        contentStatus: type==='start' ? "true" :"false",
        contentId: "",
        all:true,
      })
        .then((res) => {
          this.loading.itemStatus = false;
          if (res.code === 0) {
            this.$message.success(this.$t('knowledgeManage.operateSuccess'));
            this.getList();
          }
        })
        .catch(() => {
          this.loading.itemStatus = false;
        });
    },
    renderHeader(h, { column, $index }) {
      const columnHtml =
        this.$t('knowledgeManage.section') +
        this.cardObj[0].contentNum +
        this.$t('knowledgeManage.length')+ " :" +
        this.cardObj[0].content.length +
        this.$t('knowledgeManage.character');
      return h("span", {
        domProps: {
          innerHTML: columnHtml,
        },
      });
    },
    handleClose() {
      this.dialogVisible = false;
      if (this.cardObj[0].available === this.activeStatus) return;
      this.getList();
    },
  },
};
</script>
<style lang="scss">
.dialog-content {
  max-height:55vh!important;
  overflow-y: auto;
}
.segment-list {
  margin-top: 10px;
  
  .section-collapse {
    background-color: #f7f8fa;
    border-radius: 6px;
    border: 1px solid $color;
    overflow: hidden;
    
    /deep/ .el-collapse {
      border: none;
      border-radius: 6px;
    }
    
    /deep/ .el-collapse-item__header {
      background-color: #f7f8fa;
      border-bottom: 1px solid #e4e7ed;
      padding: 12px 20px;
      font-weight: normal;
      border-left: none;
      border-right: none;
      border-top: none;
      display: flex !important;
      align-items: center !important;
      justify-content: space-between !important;
      width: 100%;
      position: relative;
      
      &:hover {
        background-color: #f0f2f5;
      }
    }
    
    /deep/ .el-collapse-item__content {
      padding: 15px 20px;
      background-color: #fff;
      border-bottom: 1px solid #e4e7ed;
      border-left: none;
      border-right: none;
      border-top: none;
    }

    /deep/ .el-collapse-item__header .el-collapse-item__arrow,
    .el-collapse-item__arrow,
    [class*="el-collapse-item__arrow"] {
      display: none !important;
    }

    /deep/ .el-collapse-item:last-child .el-collapse-item__content {
      border-bottom: none;
    }
    
    
    /deep/ .el-collapse-item__header::after {
      display: none !important;
    }
     
    .segment-badge {
      color: $color;
      font-size: 12px;
      min-width: 40px;
      text-align: center;
      font-weight: 500;
      margin-right: 120px;
    }
    .segment-actions {
        display: flex;
        gap: 8px;
        align-items: center;
        flex: 1;
        justify-content: flex-end;
        margin-right: 10px;
        .action-btn {
          display: inline-flex;
          align-items: center;
          gap: 4px;
          padding: 4px 8px;
          border-radius: 4px;
          cursor: pointer;
          font-size: 12px;
          transition: all 0.3s ease;
          
          i {
            font-size: 14px;
          }
          
          &.edit-btn {
            color: $btn_bg;
            
            &:hover {
              color: #2a3cc7;
            }
          }
          
          &.delete-btn {
            color: $btn_bg;
            
            &:hover {
              color: #2a3cc7;
            }
          }
          
          &.save-btn {
            color: $btn_bg;
            
            &:hover {
              color: #2a3cc7;
            }
          }
          
          &.cancel-btn {
            color: #909399;
            
            &:hover {
              color: #606266;
            }
          }
        }
      }
    
    .segment-score {
      display: flex;
      align-items: center;
      position: absolute;
      right: 20px;
      top: 50%;
      transform: translateY(-50%);
      
      .score-label {
        font-size: 12px;
        color: $color;
        font-weight: bold;
        margin-right: 5px;
      }
      
      .score-value {
        font-size: 14px;
        color: $color;
        font-weight: bold;
        font-family: 'Courier New', monospace;
      }
    }
    
    .segment-content {
      padding: 10px;
      text-align: left;
      
      .content-display {
        word-wrap: break-word;
        line-height: 1.5;
      }
      
      .content-edit {
        .edit-input {
          /deep/ .el-textarea__inner {
            border: 1px solid $color;
            border-radius: 4px;
            resize: vertical;
          }
        }
      }
    }
    
    /deep/ .el-collapse-item__content {
      font-size: 14px;
      color: #333;
      line-height: 1.5;
      text-align: left;
      word-wrap: break-word;
      word-break: break-all;
      overflow-wrap: break-word;
      
      .segment-action {
        color: #999;
        font-size: 12px;
        margin-left: 8px;
      }
      
      .auto-save {
        color: #666;
        font-size: 12px;
        margin-left: 8px;
        font-style: italic;
      }
      
    }
  }
}

  .smartDate{
      padding-top:3px;
      color:#888888;
  }
  .tagList{
    cursor: pointer;
    .icon-tag{
      transform: rotate(-40deg);
      margin-right:3px;
    }
    .tagList-item{
      color:#888;
    }
  }
  .tagList > .tagList-item:hover{
      color:$color;
  }
.showMore{
  margin-left:5px;
  background:$color_opacity;
  padding:2px;
  border-radius:4px;
}
.metaItem{
  margin-left:5px;
  background:$color_opacity;
  padding:2px;
  border-radius:4px;
}
.editIcon{
  cursor: pointer;
  color:$color;
  font-size:16px;
  display: inline-block;
  margin-left:5px;
}
.section {
  width: 100%;
  height: 100%;
  padding: 20px 20px 30px 20px;
  margin: auto;
  overflow: auto;

  .el-divider--horizontal {
    margin: 30px 0;
  }
  .title {
    font-size: 18px;
    font-weight: bold;
    color: #333;
    padding: 10px 0;
  }

  .container {
    min-width: 980px;
    padding: 15px;
    height: calc(100% - 45px);
    /*background: #fff;
    box-shadow: 0 1px 6px rgba(0, 0, 0, 0.3);*/
    border-radius: 5px;
    overflow: auto;

    .el-descriptions :not(.is-bordered) .el-descriptions-item__cell {
      &:nth-child(even) {
        width: 25%;
      }
      padding: 10px;
    }
    .btn {
      padding: 10px 0;
      text-align: right;
    }

    .card {
      flex-wrap: wrap;
      .el-row {
        margin: 0 !important;
      }
      .text {
        font-size: 14px;
      }

      .item {
        height: 120px;
        margin-bottom: 18px;
        display: -webkit-box;
        -webkit-line-clamp: 6;
        -webkit-box-orient: vertical;
        overflow: hidden;
        text-overflow: ellipsis;
      }

      .clearfix{
        display:flex;
        justify-content:space-between;
        align-items:center;
      }
      .card-box {
        margin-bottom: 10px;

        .box-card {
          &:hover {
            cursor: pointer;
            transform: scale(1.03);
          }
          .more{
            margin-left:5px;
            cursor: pointer;
            transform: rotate(90deg);
            font-size: 16px;
            color: #8c8c8f;
          }
        }
        
        .segment-type {
          margin: 0 5px;
          color: #999;
          font-size: 12px;
        }
        
        .segment-length {
          color: #999;
          font-size: 12px;
        }
        
        .segment-child {
          color: #999;
          font-size: 12px;
          padding-left: 5px;
        }
      }

      .el-card__header {
        padding: 8px 20px;
      }
    }
  }
}
</style>