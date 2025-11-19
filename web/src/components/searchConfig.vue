<template>
  <el-form
    :model="formInline"
    ref="formInline"
    :inline="false"
    class="searchConfig"
  >
    <el-form-item
      class="vertical-form-item"
    >
    <template #label>
        <span v-if="!setType" class="vertical-form-title">检索方式Configuration</span>
    </template>
      <div
        v-for="item in searchTypeData"
        :class="['searchType-list',{ 'active': item.showContent }]"
      >
        <div
          class="searchType-title"
          @click="clickSearch(item)"
        >
          <span :class="[item.icon,'img']"></span>
          <div class="title-content">
            <div class="title-box">
              <h3 class="title-name">{{item.name}}</h3>
              <p class="title-desc">{{item.desc}}</p>
            </div>
            <span :class="item.showContent?'el-icon-arrow-up':'el-icon-arrow-down'"></span>
          </div>
        </div>
        <div
          class="searchType-content"
          v-if="item.showContent"
        >
          <div
            v-if="item.isWeight"
            class="weightType-box"
          >
            <div
              v-for="mixItem in item.mixType"
              :class="['weightType',{ 'active': mixItem.value === item.mixTypeValue }]"
              @click.stop="mixTypeClick(item,mixItem)"
            >
              <p class="weightType-name">{{mixItem.name}}</p>
              <p class="weightType-desc">{{mixItem.desc}}</p>
            </div>
          </div>
          <el-row
            v-if="item.isWeight && item.mixTypeValue === 'weight'"
            @click.stop
          >
            <el-col class="mixTypeRange-title">
              <span>语义[{{item.mixTypeRange}}]</span>
              <span>关键词[{{(1 - (item.mixTypeRange || 0)).toFixed(1)}}]</span>
            </el-col>
            <el-col>
              <el-slider
                v-model="item.mixTypeRange"
                show-stops
                :step="0.1"
                :max="1"
                @change="rangeChage($event)"
              >
              </el-slider>
            </el-col>
          </el-row>
          <el-row v-if="showRerank(item)">
            <el-col>
              <span class="content-name">RerankModel</span>
              <el-tooltip
                class="item"
                effect="dark"
                content="重排序Model会根据候选Document与用户问题 of 语义匹配度，对初步检索结果进行重新排序从而进一步提升最终Back结果 of 相关性And准确性。"
                placement="right"
              >
                <span class="el-icon-question tips"></span>
              </el-tooltip>
            </el-col>
            <el-col>
              <el-select
                clearable
                filterable
                style="width:100%;"
                loading-text="ModelLoading..."
                v-model="formInline.knowledgeMatchParams.rerankModelId"
                @visible-change="visibleChange($event)"
                @change="handleRerankChange"
                placeholder="Please select"
                :loading="rerankLoading"
              >
                <el-option
                  v-for="item in rerankOptions"
                  :key="item.modelId"
                  :label="item.displayName"
                  :value="item.modelId"
                >
                </el-option>
              </el-select>
            </el-col>
          </el-row>
          <el-row>
            <el-col>
              <span class="content-name">TopK</span>
              <el-tooltip
                class="item"
                effect="dark"
                content="Used for控制检索阶段Back of 最相关 of Document片段 of Count。这些Document片段将被送入GenerateModel中，Used for Generate最终 of 回答。"
                placement="right"
              >
                <span class="el-icon-question tips"></span>
              </el-tooltip>
            </el-col>
            <el-col>
              <el-slider
                :min="1"
                :max="10"
                :step="1"
                v-model="formInline.knowledgeMatchParams.topK"
                show-input
              >
              </el-slider>
            </el-col>
          </el-row>
          <el-row v-if=showHistory(item)>
            <el-col>
              <span class="content-name">最长上下文</span>
              <el-tooltip
                class="item"
                effect="dark"
                content="Save of 最长 of 上下文对话轮数。"
                placement="right"
              >
                <span class="el-icon-question tips"></span>
              </el-tooltip>
            </el-col>
            <el-col>
              <el-slider
                :min="0"
                :max="100"
                :step="1"
                v-model="formInline.knowledgeMatchParams.maxHistory"
                show-input
              >
              </el-slider>
            </el-col>
          </el-row>
          <el-row>
            <el-col>
              <span class="content-name">Score阈Value</span>
              <el-tooltip
                class="item"
                effect="dark"
                content="检索结果 of 相似度阈Value，低于该Value of 结果将被 滤。"
                placement="right"
              >
                <span class="el-icon-question tips"></span>
              </el-tooltip>
            </el-col>
            <el-col>
              <el-slider
                :min="0"
                :max="1"
                :step="0.1"
                v-model="formInline.knowledgeMatchParams.threshold"
                show-input
              >
              </el-slider>
            </el-col>
          </el-row>
        </div>
      </div>
    </el-form-item>
  </el-form>
</template>
<script>
import { getRerankList } from "@/api/modelAccess";
export default {
  props:['setType','config'],
  data() {
    return {
      debounceTimer:null,
      rerankOptions: [],
      rerankLoading: false,
      isSettingFromConfig: false, // Add标志位，Used forArea分YesNoYes从configSetting of Value
      formInline: {
        knowledgeMatchParams: {
          keywordPriority: 0.8, //关Key词权重
          matchType: "", //vector（向量检索）、text（Text检索）、mix（混合检索：向量+Text）
          priorityMatch: 1, //权重Match，只Has在混合检索模式下，选择权重Setting后，这个才Setting为1
          rerankModelId: "", //rerankModelid
          threshold: 0.4, // 滤分数阈Value
          semanticsPriority: 0.2, //语义权重
          topK:5, //topK Get最高 of 几行
          maxHistory:0//最长Context
        },
      },
      initialEditForm:null,
      searchTypeData: [
        {
          name: "向量检索",
          value: "vector",
          desc: "By向量相似度找到语义相近、表达多样 of 文本片段，适Used for理解And召回语义相关Information。",
          icon: "el-icon-menu",
          isWeight: false,
          showContent: false,
        },
        {
          name: "全文检索",
          value: "text",
          desc: "基于关键词匹配，能够高效QueryContains指定词汇 of 文本片段，适Used for精确查找",
          icon: "el-icon-document",
          isWeight: false,
          showContent: false,
        },
        {
          name: "混合检索",
          value: "mix",
          desc: "结合向量And关键词检索，融合语义理解与关键词匹配，兼顾相关性And准确性，提升检索效果。",
          icon: "el-icon-s-grid",
          isWeight: true,
          Weight: "",
          mixTypeValue: "weight",
          showContent: false,
          mixTypeRange: 0.2,
          mixType: [
            {
              name: "权重Setting",
              value: "weight",
              desc: "权重Setting功能Used for调整不同检索方式 of 影响力。BySetting权重，可以控制语义相似度And关键词匹配在最终排序中 of 占比。",
            },
            {
              name: "RerankModel",
              value: "rerank",
              desc: "重排序Model会根据候选Document与用户问题 of 语义匹配度，对初步检索结果进行重新排序从而进一步提升最终Back结果 of 相关性And准确性。",
            },
          ],
        },
      ],
    };
  },
  watch: {
    formInline: {
      handler(newVal) {
        // IfYes从configSetting of Value，不TriggersendConfigInfo
        if (this.isSettingFromConfig) {
          return;
        }
        
        if (this.debounceTimer) {
          clearTimeout(this.debounceTimer);
        }
        this.debounceTimer = setTimeout(() => {
          const props = ['knowledgeMatchParams'];
          const changed = props.some(prop => {
            return JSON.stringify(newVal[prop]) !== JSON.stringify(
                (this.initialEditForm || {})[prop]
              );
            });
          if (changed) {
            if(!this.setType){
              delete this.formInline.knowledgeMatchParams.maxHistory;
            }
            this.$emit('sendConfigInfo', this.formInline);
          }
        }, 200);
      },
      deep: true,
      immediate: false
    },
    config:{
      handler(newVal) {
        if(newVal && Object.keys(newVal).length > 0){
          this.isSettingFromConfig = true; // Setting标志位
          const formData = JSON.parse(JSON.stringify(newVal))
          this.formInline.knowledgeMatchParams = formData;
          const { matchType,priorityMatch } = this.formInline.knowledgeMatchParams;
          if(matchType !== ''){
              this.searchTypeData = this.searchTypeData.map((item) => ({
              ...item,
              showContent: item.value === matchType ? true : false,
            }));
            if(matchType === 'mix'){
              this.searchTypeData[2]['mixTypeValue'] = priorityMatch === 1 ? 'weight' : 'rerank';
            }
          }

          // UsenextTick确保DOMUpdate完成后再Reset标志位
          this.$nextTick(() => {
            this.isSettingFromConfig = false;
          });
        }
      },
      deep: true,
      immediate: true
    }
  },
  mounted() {
    this.$nextTick(() => {
      this.initialEditForm = JSON.parse(JSON.stringify(this.formInline));
    });
  },
  created() {
    // 预LoadData，避免首次打开下拉框 when  of Delay
    this.getRerankData();
  },
  methods: {
    rangeChage(val){
      this.formInline.knowledgeMatchParams.keywordPriority = Number((1 - (val || 0)).toFixed(1));
      this.formInline.knowledgeMatchParams.semanticsPriority = val;
    },
    mixTypeClick(item, n) {
      item.mixTypeValue = n.value;
      const { knowledgeMatchParams } = this.formInline;
      knowledgeMatchParams.priorityMatch = n.value === "weight" ? 1 : 0;
      // if(n.value === 'weight'){
      //   knowledgeMatchParams.rerankModelId = '';
      // }
    },
    showRerank(n) {
      return (
        n.value === "vector" ||
        n.value === "text" ||
        (n.value === "mix" && n.mixTypeValue === "rerank")
      );
    },
    showHistory(n){
      return (
       (this.setType === 'rag'||this.setType === 'agent') &&
        (n.value === "vector" ||
         n.value === "text" ||
         (n.value === "mix") //&& n.mixTypeValue === "rerank"
        )
      )
    },
    clickSearch(n) {
      this.formInline.knowledgeMatchParams.matchType = n.value;
      this.searchTypeData = this.searchTypeData.map((item) => ({
        ...item,
        showContent: item.value === n.value ? !item.showContent : false,
      }));
      this.formInline.knowledgeMatchParams.priorityMatch = n.value !== 'mix' ? 0 : 1;
      this.clear();
    },
    clear() {
      this.formInline.knowledgeMatchParams.rerankModelId = "";
      this.formInline.knowledgeMatchParams.keywordPriority = 0.8;
      this.formInline.knowledgeMatchParams.semanticsPriority = 0.2;
      this.formInline.knowledgeMatchParams.threshold = 0.4;
      this.formInline.knowledgeMatchParams.topK = 5;
    },
    getRerankData() {
      this.rerankLoading = true;
      getRerankList().then((res) => {
        if (res.code === 0) {
          this.rerankOptions = res.data.list || [];
        }
      }).finally(() => {
        this.rerankLoading = false;
      });
    },
    visibleChange(val) {
      if (val && this.rerankOptions.length === 0) {
        this.getRerankData();
      }
    },
    handleRerankChange(value) {
      // DirectlyTriggerEvent，避免防抖Delay
      if(!this.setType){
        const formData = JSON.parse(JSON.stringify(this.formInline));
        delete formData.knowledgeMatchParams.maxHistory;
        this.$emit('sendConfigInfo', formData);
      } else {
        this.$emit('sendConfigInfo', this.formInline);
      }
    },
  },
};
</script>
<style lang="scss" scoped>
/deep/ {
  .vertical-form-item {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    .vertical-form-title{
      color:#000;
      font-size:14px;
    }
  }
  .vertical-form-item .el-form-item__label {
    line-height: unset;
    font-size: 14px;
    font-weight: bold;
  }
  .el-form-item__content {
    width: 100%;
  }
  .el-input-number--small{
    line-height: 28px!important;
  }
}
.active {
  border: 1px solid $color !important;
}
.searchConfig {
  .searchType-list:hover {
    border: 1px solid $color;
  }
  .searchType-list {
    border: 1px solid #c0c4cc;
    border-radius: 4px;
    margin: 20px 0;
    padding: 0 10px;
    cursor: pointer;
    .searchType-title {
      display: flex;
      align-items: center;
      .img {
        font-size: 30px;
        text-align: center;
        line-height: 50px;
        color: $color;
        background-color: #fff;
        width: 50px;
        height: 50px;
        border-radius: 8px;
        border: 1px solid #e9e9eb;
        box-shadow: 4px 2px 4px #f1f1f1;
      }
      .title-content {
        flex: 1;
        display: flex;
        margin-left: 10px;
        justify-content: space-between;
        align-items: center;
        .title-name {
          font-size: 16px;
          font-weight: bold;
          line-height: 1;
          padding-top: 10px;
        }
        .title-desc {
          color: #888;
        }
      }
    }
    .searchType-content {
      padding: 20px;
      .tips {
        color: #888;
        margin-left: 5px;
      }
      .content-name {
        font-weight: bold;
      }
      .weightType-box {
        display: flex;
        gap: 20px;
        .weightType {
          border: 1px solid #c0c4cc;
          border-radius: 4px;
          .weightType-name {
            text-align: center;
            font-weight: bold;
            line-height: 2;
            font-size: 16px;
            padding-top: 5px;
          }
          .weightType-desc {
            text-align: center;
            line-height: 1.5;
            padding: 10px;
            color: #888;
          }
        }
      }
      .mixTypeRange-title {
        display: flex;
        align-items: center;
        justify-content: space-between;
        font-weight: bold;
        margin-top: 20px;
        line-height: 1;
      }
    }
  }
}
</style>