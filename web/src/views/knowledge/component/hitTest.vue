<template>
  <div class="page-wrapper full-content">
    <div class="page-title">
      <i
        class="el-icon-arrow-left"
        @click="goBack"
        style="margin-right: 10px; font-size: 20px; cursor: pointer"
      >
      </i>
      {{$t('knowledgeManage.hitTest.name')}}
      <LinkIcon type="knowledge-hit" />
    </div>
    <div class="block wrap-fullheight">
      <div class="test-left test-box">
        <div class="hitTest_input">
          <h3>{{$t('knowledgeManage.hitTest.title')}}</h3>
          <el-input
            type="textarea"
            :rows="4"
            v-model="question"
            class="test_ipt"
          />
          <div class="test_btn">
            <el-button
              type="primary"
              size="small"
              @click="startTest"
            >{{$t('knowledgeManage.hitTest.hitTestBtn')}}<span class="el-icon-caret-right"></span></el-button>
          </div>
        </div>
         <div class="hitTest_input meta_box">
          <h3>{{$t('knowledgeManage.hitTest.metaDataFilter')}}</h3>
          <metaSet ref="metaSet" class="metaSet" :knowledgeId="knowledgeId" />
        </div>
        <div class="test_form">
          <searchConfig ref="searchConfig" @sendConfigInfo="sendConfigInfo" />
        </div>
        <div class="hitTest_input graph_box" v-if="graphSwitch">
          <graphSwitch ref="graphSwitch" @graphSwitchchange="graphSwitchchange" :label="$t('knowledgeManage.hitTest.graph')"/>
        </div>
      </div>
      <div class="test-right test-box">
        <div class="result_title">
          <h3>{{$t('knowledgeManage.hitTest.hitTestResult')}}</h3>
          <img
            src="@/assets/imgs/nodata_2x.png"
            v-if="searchList.length >0"
          />
        </div>
        <div
          class="result"
          v-loading="resultLoading"
        >
          <div
            v-if="searchList.length >0"
            class="result_box"
          >
            <div
              v-for="(item,index) in searchList"
              :key="'result'+index"
              class="resultItem"
            >
              <div class="resultTitle">
                <span>
                  <span class="tag"  @click="showSectionDetail(index)">{{$t('knowledgeManage.section')}}#{{index+1}}</span>
                  <span v-if="['graph','community_report'].includes(item.contentType)" class="segment-type">
                    {{item.contentType === 'graph' ? '#' + $t('knowledgeManage.hitTest.graph') : '#' + $t('knowledgeManage.hitTest.communityReport')}}
                  </span>
                  <span v-else>
                    <span class="segment-type">{{item.childContentList && item.childContentList.length > 0 ? '#' + $t('knowledgeManage.hitTest.parentSonSegment') : '#' + $t('knowledgeManage.hitTest.commonSegment')}}</span>
                    <span class="segment-length" v-if="item.childContentList && item.childContentList.length > 0" @click="showSectionDetail(index)">#{{item.childContentList.length || 0}}{{$t('knowledgeManage.hitTest.childSegmentCount')}}</span>
                  </span>
                </span>
                <span class="score">{{$t('knowledgeManage.hitTest.hitScore')}}: {{(formatScore(score[index]))}}</span>
              </div>
              <div>
                <div class="resultContent">
                  {{item.snippet.slice(0,100)}}...
                </div>
                <div class="resultChildContent" v-if="item.childContentList.length > 0">
                  <el-collapse 
                    v-model="activeNames" 
                    class="section-collapse"
                    :accordion="false"
                  >
                    <el-collapse-item 
                      :name="`${index}`"
                      class="segment-collapse-item"
                    >
                      <template slot="title">
                        <span class="sub-badge">{{$t('knowledgeManage.hitTest.hitChildSegment', {count: item.childContentList.length})}}</span>
                      </template>
                      <div class="segment-content">
                        <div 
                          v-for="(child, childIndex) in item.childContentList" 
                          :key="childIndex"
                          class="child-item"
                        >
                          <div class="child-header">
                            <span class="child-header-content">
                              <span class="segment-badge">C-{{ childIndex + 1 }}</span>
                              <span class="segment-content">{{child.childSnippet}}</span>
                            </span>
                            <span class="segment-score">
                              <span class="score-value">{{$t('knowledgeManage.hitTest.hitScore')}}: {{ formatScore(item.childScore[childIndex]) }}</span>
                            </span>
                          </div>
                        </div>
                      </div>
                    </el-collapse-item>
                  </el-collapse>
                </div>
                <div class="file_name">{{$t('knowledgeManage.hitTest.fileName')}}：{{item.title}}</div>
              </div>
            </div>
          </div>
          <div
            v-else
            class="nodata"
          >
            <img src="@/assets/imgs/nodata_2x.png" />
            <p class="nodata_tip">{{$t('knowledgeManage.hitTest.noData')}}</p>
          </div>
        </div>
        <!-- 分段DetailsArea域 -->
        <sectionShow ref="sectionShow" />
      </div>
    </div>
  </div>
</template>
<script>
import { hitTest } from "@/api/knowledge";
import { md } from "@/mixins/marksown-it";
import { formatScore } from "@/utils/util";
import searchConfig from '@/components/searchConfig.vue';
import LinkIcon from "@/components/linkIcon.vue";
import metaSet from "@/components/metaSet";
import graphSwitch from "@/components/graphSwitch.vue"
import sectionShow from "./sectionShow.vue";
export default {
  components:{LinkIcon, searchConfig, metaSet, sectionShow,graphSwitch}, 
  data() {
    return {
      md: md,
      question: "",
      resultLoading: false,
      knowledgeIdList:{},
      searchList: [],
      score: [],
      formInline:null,
      knowledgeId:this.$route.query.knowledgeId,
      name:this.$route.query.name,
      graphSwitch:this.$route.query.graphSwitch || false,
      activeNames: [],
      useGraph:false
    };
  },
  methods: {
    formatScore,
    goBack() {
      this.$router.go(-1);
    },
    graphSwitchchange(val){
      this.useGraph = val
    },
    sendConfigInfo(data){
      this.formInline = data;
    },
    startTest() {
      const metaData  = this.$refs.metaSet.getMetaData();
      this.knowledgeIdList = {
        ...metaData,
        id:this.knowledgeId,
        name:this.name
      }

      if (this.question === "") {
        this.$message.warning(this.$t('knowledgeManage.hitTest.inputQuestion'));
        return;
      }
      if(this.formInline === null){
        this.$message.warning(this.$t('knowledgeManage.hitTest.selectSearchType'));
        return;
      }
      const { matchType, priorityMatch, rerankModelId } = this.formInline.knowledgeMatchParams;
      if ((matchType !== 'mix' || priorityMatch !== 1) && !rerankModelId) {
        this.$message.warning(this.$t('knowledgeManage.hitTest.selectRerankModel'));
        return;
      }
      if(matchType === 'mix' && priorityMatch === 1){
        this.formInline.knowledgeMatchParams.rerankModelId = '';
      }
      if(this.$refs.metaSet.validateRequiredFields(this.knowledgeIdList['metaDataFilterParams']['metaFilterParams'])){
        this.$message.warning(this.$t('knowledgeManage.hitTest.fillInMissingInfo'))
        return
      }
      const { knowledgeMatchParams } = this.formInline;
      this.$set(knowledgeMatchParams, 'useGraph', this.useGraph);
      const data = {
        ...this.formInline,
        knowledgeList:[this.knowledgeIdList],
        question: this.question,
      };
      this.test(data);
    },
    test(data) {
      this.resultLoading = true;
      this.searchList = [];
      this.score = [];
      hitTest(data).then((res) => {
        if (res.code === 0) {
          this.searchList = res.data !== null ? res.data.searchList : [];
          this.score = res.data !== null ? res.data.score : [];
          // Setting所Has子分段Defaultexpand
          this.activeNames = [];
          this.searchList.forEach((item, index) => {
            if (item.childContentList && item.childContentList.length > 0) {
              this.activeNames.push(`${index}`);
            }
          });

          this.resultLoading = false;
        } else {
          this.searchList = [];
          this.resultLoading = false;
        }
      }).catch(() =>{
        this.resultLoading = false;
      })
    },
    
    // Show分段Details弹框
    showSectionDetail(index) {
      const currentItem = this.searchList[index];
      const currentScore = parseFloat(this.score[index]) || 0;
      const data = {
        searchList:currentItem,
        score:currentScore,
      }
      this.$refs.sectionShow.showDiaglog(data);
    },
  },
};
</script>
<style lang="scss" scoped>
.full-content {
  display: flex;
  flex-direction: column;
  .page-title {
    border-bottom: 1px solid #d9d9d9;
  }
  .block {
    margin: 30px 10px;
    display: flex;
    height: calc(100% - 123px);
    gap: 20px;
    .test-box {
       flex:1;
       height:100%;
       overflow-y:auto;
      .hitTest_input {
        background: #fff;
        border-radius: 6px;
        border: 1px solid #e9ecef;
        padding: 0 20px;
        h3 {
          padding: 30px 0 10px 0;
          font-size: 14px;
          font-weight: bold;
        }
        .test_ipt {
          padding-bottom: 10px;
        }
        .test_btn {
          padding: 10px 0;
          display: flex;
          justify-content: flex-end;
        }
      }
      .test_form {
        margin-top: 20px;
        padding: 20px;
        background: #fff;
        border-radius: 6px;
        border: 1px solid #e9ecef;
      }
    }
    .test-right {
      background: #fff;
      border-radius: 6px;
      border: 1px solid #e9ecef;
      height: 100%;
      padding: 20px;
      box-sizing:border-box;
      display: flex;
      flex-direction: column;
      .result_title {
        display: flex;
        justify-content: space-between;
        h3 {
          padding: 10px 0 10px 0;
          font-size: 14px;
        }
        img {
          width: 150px;
        }
      }
      .resultContent{
        img{
          width:100%;
        }
      }
      .result {
        flex: 1;
        width: 100%;
        display: flex;
        flex-direction: column;
        min-height: 0;
        .result_box {
          width: 100%;
          flex: 1;
          overflow-y: scroll;
          .resultItem {
            background: #f7f8fa;
            border-radius: 6px;
            margin-bottom: 20px;
            padding: 20px;
            color: #666666;
            line-height: 1.8;
            .resultTitle {
              display: flex;
              align-items: center;
              justify-content: space-between;
              padding: 10px 0;
              .tag {
                color: $color;
                display: inline-block;
                background: #d2d7ff;
                padding: 0 10px;
                border-radius: 4px;
                cursor: pointer;
              }
              .segment-type {
                padding: 0 5px;
              }
              .segment-length{
                cursor: pointer;
              }
              .segment-length:hover{
                color: $color;
              }
              .segment-type,
              .segment-length {
                color: #999;
                font-size: 12px;
              }
              .score {
                color: $color;
                font-weight: bold;
              }
            }
            .file_name {
              border-top: 1px dashed #d9d9d9;
              margin: 10px 0;
              padding-top: 10px;
              font-weight: bold;
            }
            .resultChildContent {
              margin-top: 10px;
              
              
              .section-collapse {
                border: none !important;
                background: transparent !important;
                
                /deep/ .el-collapse-item__arrow {
                  display: none !important;
                }
                
                /deep/ .el-collapse-item__header {
                  background: transparent !important;
                  border: none !important;
                  padding: 0 !important;
                }
                
                /deep/ .el-collapse-item__wrap {
                  background: transparent !important;
                  border: none !important;
                }
                
                /deep/ .el-collapse-item__content {
                  background: transparent !important;
                  border: none !important;
                  padding: 0 !important;
                }
                
                .segment-collapse-item {
                  .sub-badge {
                    color: #666666;
                    font-size: 14px;
                    font-weight: 800;
                  }
                  
                  .segment-content {
                    .child-item {
                      padding: 10px 0;
                      background: #f9f9f9;
                      border-radius: 4px;
                      
                      .child-header {
                        display: flex;
                        justify-content: space-between;
                        align-items: center;
                        margin-bottom: 8px;
                        .child-header-content {
                          flex: 1;
                          display: flex;
                          align-items: center;
                          min-width: 0;
                          
                          .segment-content {
                            flex: 1;
                            min-width: 0;
                            overflow: hidden;
                            text-overflow: ellipsis;
                            white-space: nowrap;
                            font-size: 14px;
                            color: #333;
                            line-height: 1.4;
                          }
                        }
                        .segment-badge {
                          background-color: #eaecf9;
                          padding: 6px 12px;
                          border-radius: 4px;
                          color: $color;
                          font-size: 12px;
                          min-width: 40px;
                          text-align: center;
                          font-weight: 500;
                          margin-right: 8px;
                          flex-shrink: 0;
                        }
                        
                        .segment-score {
                          flex-shrink: 0;
                          margin-left: 12px;
                          
                          .score-value {
                            color: $color;
                            font-weight: 500;
                            font-size: 12px;
                            white-space: nowrap;
                          }
                        }
                      }
                    }
                  }
                }
              }
              
            }
          }
        }
        .nodata {
          width: 100%;
          height: 100%;
          display: flex;
          align-items: center;
          justify-content: center;
          flex-direction: column;
          align-self: center; /* 仅该element纵向居in */
          .nodata_tip {
            padding: 10px 0;
            color: #595959;
          }
        }
      }
    }
    .meta_box{
      margin-top:20px;
      padding: 0 20px 20px 20px !important;
      .metaSet{
        width:100%;
      }
    }
    .graph_box{
      margin-top:20px;
    }
  }
}
</style>