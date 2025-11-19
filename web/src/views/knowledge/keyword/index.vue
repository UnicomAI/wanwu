<template>
  <div class="page-wrapper full-content">
    <div class="page-title">
      <i
        class="el-icon-arrow-left"
        @click="goBack"
        style="margin-right: 10px; font-size: 20px; cursor: pointer"
      >
      </i>
      {{$t('knowledgeManage.keyWordManage')}}
      <LinkIcon type="knowledge-keywords" />
      <div class="keyWordTip">{{$t('knowledgeManage.keyWordTip')}}</div>
    </div>
    <div class="block table-wrap list-common wrap-fullheight">
      <el-container class="konw_container">
        <el-main class="noPadding">
          <el-container>
            <el-header class="classifyTitle">
              <div class="searchInfo">
                <search-input class="cover-input-icon" :placeholder="$t('knowledgeManage.keyWordPlaceholder')" ref="searchInput" @handleSearch="handleSearch" style="width:300px;"/>
              </div>
              <div class="content_title">
                <el-button size="mini" type="primary" icon="el-icon-plus" @click="create">{{$t('knowledgeManage.newKeyWord')}}</el-button>
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
                :data="tableData"
                style="width: 100%"
                :header-cell-style="{ background: '#F9F9F9', color: '#999999' }"
              >
                <el-table-column
                  prop="name"
                  :label="$t('keyword.quesKeyword')"
                >
                </el-table-column>
                <el-table-column
                  prop="alias"
                  :label="$t('keyword.docWord')"
                >
                </el-table-column>
                <el-table-column
                  prop="knowledgeBaseNames"
                  :label="$t('keyword.linkKnowledge')"
                >
                <template slot-scope="scope">
                  <span>{{scope.row.knowledgeBaseNames.join(',')}}</span>
                </template>
                </el-table-column>
                <el-table-column
                  prop="updatedAt"
                  :label="$t('keyword.undateTime')"
                >
                </el-table-column>
                <el-table-column
                  :label="$t('knowledgeManage.operate')"
                  width="260"
                >
                  <template slot-scope="scope">
                    <el-button
                      size="mini"
                      round
                      @click="editItem(scope.row)"
                    >{{$t('common.button.edit')}}</el-button>
                    <el-button
                      size="mini"
                      round
                      @click="handleDel(scope.row)"
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
    <createKeyWords ref="createKeyWords"/>
  </div>
</template>

<script>
import Pagination from "@/components/pagination.vue";
import SearchInput from "@/components/searchInput.vue";
import {delDocItem} from "@/api/knowledge";
import {getKeyWord,delKeyWord} from "@/api/keyword";
import createKeyWords from './create.vue';
import LinkIcon from "@/components/linkIcon.vue";
export default {
  components: {LinkIcon, Pagination,SearchInput,createKeyWords},
  data() {
    return {
      tableLoading:false,
      docQuery: {
        name:''
      },
      listApi: getKeyWord,
      title_tips:'',
      showTips:false,
      tableData: [],
      keyWordeData: []
    };
  },
  mounted(){
    this.getTableData(this.docQuery)
  },
  methods: {
    refreshData(data){
      this.tableData = data
    },
    updateData(){
      this.getTableData(this.docQuery)
    },
    create(){
      this.$refs.createKeyWords.showDialog()
    },
    editItem(item){
      this.$refs.createKeyWords.showDialog(item)
    },
    goBack(){
      this.$router.push({path:'/knowledge'})
    },
    handleSearch(val){
      this.docQuery.name = val;
      this.getTableData(this.docQuery)
    },
    handleDel(data){
       this.$confirm('ConfirmtoDeletecurrentData吗？',this.$t('knowledgeManage.tip'),
        {
          confirmButtonText:  this.$t('common.button.confirm'),
          cancelButtonText: this.$t('common.button.cancel'),
          type: "warning"
        }
      )
        .then(async () => {
          let jsondata = {id:data.id}
          this.tableLoading = true;
          let res = await delKeyWord(jsondata);
          if (res.code === 0) {
            this.$message.success(this.$t('common.info.delInfo'));
            this.getTableData(this.docQuery)//Get知识分ClassData
          }
          this.tableLoading = false;
        })
        .catch((error) => {
          this.getTableData(this.docQuery)
        });
    },
    async getTableData(data){
       this.tableLoading = true;
       this.tableData = await this.$refs["pagination"].getTableData(data);
       this.tableLoading = false;
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
  .keyWordTip{
    padding-top:15px;
    color:#888888;
    font-weight: normal;
  }
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