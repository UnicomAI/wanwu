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
        <span v-if="!setType" class="vertical-form-title">Search Method Configuration</span>
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
              <span>Semantic[{{item.mixTypeRange}}]</span>
              <span>Keyword[{{(1 - (item.mixTypeRange || 0)).toFixed(1)}}]</span>
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
                content="The reranking model reorders initial search results based on semantic matching between candidate documents and user questions, further improving the relevance and accuracy of final returned results."
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
                content="Used to control the number of most relevant document segments returned in the retrieval phase. These document segments will be sent to the generation model to generate the final answer."
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
              <span class="content-name">Max Context Length</span>
              <el-tooltip
                class="item"
                effect="dark"
                content="Maximum number of context dialogue rounds to save."
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
              <span class="content-name">Score Threshold</span>
              <el-tooltip
                class="item"
                effect="dark"
                content="Similarity threshold for search results, results below this value will be filtered."
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
      isSettingFromConfig: false, // Add flag to distinguish whether value is set from config
      formInline: {
        knowledgeMatchParams: {
          keywordPriority: 0.8, //Keyword weight
          matchType: "", //vector（Vector search）、text（Text search）、mix（Hybrid search: vector + text）
          priorityMatch: 1, //Weight matching, only set to 1
          rerankModelId: "", //rerankModelid
          threshold: 0.4, //Filter score threshold
          semanticsPriority: 0.2, //Semantic weight
          topK:5, //topK Get top N rows
          maxHistory:0//Max context
        },
      },
      initialEditForm:null,
      searchTypeData: [
        {
          name: "Vector search",
          value: "vector",
          desc: "Find semantically similar and diversely expressed text segments through vector similarity, suitable for understanding and recalling semantically related information.",
          icon: "el-icon-menu",
          isWeight: false,
          showContent: false,
        },
        {
          name: "Full-text Search",
          value: "text",
          desc: "Based on keyword matching, efficiently query text segments containing specified words, suitable for precise search",
          icon: "el-icon-document",
          isWeight: false,
          showContent: false,
        },
        {
          name: "Hybrid Search",
          value: "mix",
          desc: "Combines vector and keyword search, integrating semantic understanding with keyword matching, balancing relevance and accuracy to improve search results.",
          icon: "el-icon-s-grid",
          isWeight: true,
          Weight: "",
          mixTypeValue: "weight",
          showContent: false,
          mixTypeRange: 0.2,
          mixType: [
            {
              name: "Weight Settings",
              value: "weight",
              desc: "Use weight settings to adjust the influence of each retrieval method. By tuning the weights you control how semantic similarity and keyword matching contribute to the final ranking.",
            },
            {
              name: "RerankModel",
              value: "rerank",
              desc: "The reranking model reorders initial search results based on semantic matching between candidate documents and user questions, further improving the relevance and accuracy of final returned results.",
            },
          ],
        },
      ],
    };
  },
  watch: {
    formInline: {
      handler(newVal) {
        // If value is set from config, do not trigger sendConfigInfo
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
          this.isSettingFromConfig = true; // Set flag
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

          // Use nextTick to ensure DOM update completes before resetting flag
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
    // Preload data to avoid delay when opening dropdown for the first time
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
      // Directly trigger event to avoid debounce delay
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