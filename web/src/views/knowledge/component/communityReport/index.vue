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
        <el-descriptions-item :label="$t('knowledgeManage.communityReport.name')">
          {{ $t('knowledgeManage.communityReport.communityReport') }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('knowledgeManage.communityReport.segmentTotalNum')">
          {{ res.total }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('knowledgeManage.communityReport.uploadTime')">
          {{ formatDate(res.createdAt) }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('knowledgeManage.communityReport.segmentType')">{{
         communityReportStatus[res.status]
        }}</el-descriptions-item>
      </el-descriptions>

      <div class="btn">
        <el-button
          type="primary"
          icon="el-icon-refresh"
          @click="refreshData"
          size="mini"
          :loading="loading.itemStatus"
        >{{ $t('common.gpuDialog.reload') }}</el-button>
        <el-button
            type="primary"
            @click="generateReport"
            size="mini"
            :loading="loading.stop"
            :disabled="!res.canGenerate || permissionType === 0"
            >{{ res.generateLabel === '' ? $t('knowledgeManage.communityReport.generate') : res.generateLabel }}</el-button>
        <el-button
          type="primary"
          @click="createReport"
          size="mini"
          :loading="loading.stop"
          :disabled="!res.canAddReport || permissionType === 0"
          >{{ $t('knowledgeManage.communityReport.addCommunityReport') }}</el-button>
      </div>

      <div class="card">
        <el-row :gutter="20" v-if="res && res.list && res.list.length > 0">
          <el-col
            :span="6"
            v-for="(item, index) in res.list"
            :key="index"
            class="card-box"
          >
            <el-card class="box-card">
              <div slot="header" class="clearfix">
                <el-tooltip :content="item.title" placement="top" :disabled="item.title.length <= 10">
                  <span>{{ item.title.length > 10 ? item.title.substring(0, 10) + '...' : item.title }}</span>
                </el-tooltip>
                <div>
                  <el-dropdown @command="handleCommand" placement="bottom" v-if="permissionType !== 0">
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
          :current-page.sync="page.pageNo"
          :page-sizes="page.pageSizeList"
          :page-size="page.pageSize"
          layout="total, prev, pager, next, jumper"
          :total="page.total"
        >
        </el-pagination>
      </div>
    </div>
    <createReport ref="createReport" @refreshData="refreshData"></createReport>
  </div>
</template>
<script>
import { getCommunityReportList,delCommunityReport,generateCommunityReport} from "@/api/knowledge";
import { COMMUNITY_REPORT_STATUS } from "@/views/knowledge/config";
import {mapGetters} from 'vuex';
import commonMixin from "@/mixins/common";
import createReport from "./create.vue";
export default {
  components:{createReport},
  mixins: [commonMixin],
  data() {
    return {
      obj: {},
      page: {
        pageNo: 1,
        pageSize: 8,
        pageSizeList: [10, 15, 20, 50],
        total: 0,
      },
      loading: {
        stop: false,
        itemStatus: false,
      },
      res: {
        contentList: [],
      },
      communityReportStatus: COMMUNITY_REPORT_STATUS,
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
  methods: {
    formatDate(value) {
      if (value === null || value === undefined || value === '') {
        return '-'
      }
      let dateValue = value
      if (typeof value === 'number' || (typeof value === 'string' && /^\d+$/.test(value))) {
        const timestamp = typeof value === 'string' ? parseInt(value) : value
        if (timestamp.toString().length === 10) {
          dateValue = timestamp * 1000
        } else {
          dateValue = timestamp
        }
      }
      const dateFormatFilter =
        (this.$options.filters && this.$options.filters.dateFormat) || null
      return dateFormatFilter ? dateFormatFilter(dateValue) : dateValue
    },
    refreshData() {
      setTimeout(() => {
        this.getList();
      },500);
    },
    createReport(){
      this.$refs.createReport.showDialog(this.obj.knowledgeId,'add');
    },
    generateReport(){
      generateCommunityReport({knowledgeId:this.obj.knowledgeId}).then(res =>{
        if(res.code === 0){
          this.$message.success(this.$t('knowledgeManage.communityReport.generateSuccess'));
          this.getList();
        }
      })
    },
    handleCommand(value){
      const {type, item} = value || {}
       switch (type) {
          case 'delete':
            this.delReport(item)
            break
        }
    },
    delReport(item){
      delCommunityReport({contentId:item.contentId,knowledgeId:this.obj.knowledgeId}).then(res =>{
        if(res.code === 0){
          this.$message.success(this.$t('knowledgeManage.communityReport.deleteSuccess'));
          this.getList();
        }
      })
    },
    getList() {
      this.loading.itemStatus = true;
      getCommunityReportList({
        knowledgeId: this.obj.knowledgeId,
        pageNo: this.page.pageNo,
        pageSize:this.page.pageSize
      })
        .then((res) => {
          this.loading.itemStatus = false;
          this.res = res.data;
          this.page.total = this.res.total;
          if((!this.res.list || this.res.list.length === 0) && this.page.pageNo > 1){
            this.page.pageNo = 1;
            this.getList();
          }
        })
        .catch(() => {
          this.loading.itemStatus = false;
        });
    },
    handleClick(item, index) {
      if(this.permissionType === 0) return;
      // 点击CardEvent，可根据需RequestAdd功能
      this.$refs.createReport.showDialog(this.obj.knowledgeId, 'edit', item)
    },
    handleCurrentChange(val) {
      this.page.pageNo = val;
      this.getList();
    },
    handleSizeChange(val) {
      this.page.pageSize = val;
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