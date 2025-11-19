<template>
  <div class="page-wrapper full-content">
    <div class="page-title">
      <i
        class="el-icon-arrow-left"
        @click="goBack"
        style="margin-right: 10px; font-size: 20px; cursor: pointer"
      >
      </i>
      安全护栏
      <LinkIcon type="safety" />
    </div>
    <div class="block table-wrap list-common wrap-fullheight">
      <el-container class="konw_container">
        <el-main class="noPadding">
          <el-container>
            <el-header class="classifyTitle">
              <div class="searchInfo">
                <!-- <search-input class="cover-input-icon" :placeholder="$t('knowledgeManage.docPlaceholder')" ref="searchInput" @handleSearch="handleSearch" /> -->
              </div>

              <div class="content_title">
                <el-button size="mini" type="primary"  @click="showCreate">AddSensitive Word</el-button>
                <el-button size="mini" type="primary" @click="showReply">ReplySetting</el-button>
              </div>
            </el-header>
            <el-main
              class="noPadding"
              v-loading="tableLoading"
            >
              <el-table
                :data="tableData"
                style="width: 100%"
                :header-cell-style="{ background: '#F9F9F9', color: '#999999' }"
              >
                <el-table-column
                  prop="word"
                  label="Sensitive Word"
                >
                </el-table-column>
                <el-table-column
                  prop="sensitiveType"
                  label="Type"
                >
                <template slot-scope="scope">
                  <span>{{safetyType[scope.row.sensitiveType]}}</span>
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
                      type="primary"
                    >{{$t('common.button.delete')}}</el-button>
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
    <createWord ref="createWord" @reload="reload"/>
    <setReply ref="setReply" />
  </div>
</template>

<script>
import Pagination from "@/components/pagination.vue";
import createWord from './createWord.vue';
import setReply from './setReply.vue';
import { SafetyType } from "@/utils/commonSet";
import {getSensitiveWord,delSensitiveWord} from "@/api/safety";
import LinkIcon from "@/components/linkIcon.vue";
export default {
  components:{LinkIcon, createWord,setReply,Pagination},
  data() {
    return {
      safetyType:SafetyType,
      loading:false,
      tableLoading:false,
      docQuery: {
        tableId:this.$route.params.id
      },
      fileList: [],
      listApi: getSensitiveWord,
      title_tips:'',
      showTips:false,
      tableData: [],
      currentKnowValue:null,
    };
  },
  mounted(){
    this.getTableData(this.docQuery)
  },
  methods: {
    reload(){
      this.getTableData(this.docQuery)
    },
    showCreate(){
      this.$refs.createWord.showDialog(this.docQuery.tableId)
    },
    showReply(){
      this.$refs.setReply.showDialog(this.docQuery.tableId)
    },
    goBack(){
      this.$router.push({path:'/safety'})
    },
    reload(){
      this.getTableData(this.docQuery)
    },
    handleDel(data){
       this.$confirm(`确定DeleteSensitive Word-${data.word}`,this.$t('knowledgeManage.tip'),
        {
          confirmButtonText:  this.$t('common.button.confirm'),
          cancelButtonText: this.$t('common.button.cancel'),
          type: "warning"
        }
      )
        .then(async () => {
          let jsondata = {tableId:this.docQuery.tableId,wordId:data.wordId}
          this.loading = true;
          let res = await delSensitiveWord(jsondata);
          if (res.code === 0) {
            this.$message.success('DeleteSuccess');
            this.getTableData(this.docQuery)
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
    },
    async download(url,name){
      const res = await downDoc(url)
      const blobUrl = window.URL.createObjectURL(res) // 将blobObject转为一个URL
      const link = document.createElement('a')
      link.href = blobUrl
      link.download = name
      link.click() // StartDownload
      window.URL.revokeObjectURL(link.href) // Download完毕DeleteaTag
    },
    handleUpload() {
      this.$router.push({path:'/knowledge/fileUpload',query:{id:this.docQuery.knowledgeId,name:this.knowledgeName}})
    },
    refreshData(data) {
      this.tableData = data
    }
  }
};
</script>
<style lang="scss" scoped>
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
  border-color: #eee; /* Setting边框颜色 */
  background-color: #fff; /* Setting背景颜色 */
  color: #666; /* Setting文字颜色 */
}
.custom-tooltip.el-tooltip__popper[x-placement^="top"] .popper__arrow::after {
  border-top-color: #fff !important;
}
.custom-tooltip.el-tooltip__popper.is-light[x-placement^="top"] .popper__arrow {
  border-top-color: #ccc !important;
}
</style>