<template>
  <el-dialog
    :visible.sync="dialogVisible"
    title="Hit Segment Details"
    width="70%"
    @close="handleClose"
    class="section-dialog"
  >
    <div class="section-show-container">
      <!-- Parent segment area -->
      <div class="parent-segment" v-if="parentSegment">
        <div class="segment-header">
          <span class="parent-badge" v-if="['graph','community_report'].includes(parentSegment.contentType)">{{parentSegment.contentType === 'graph' ? 'Knowledge Graph' : 'Community Report'}}</span>
          <span class="parent-badge" v-else>{{segmentList.length > 0 ? 'Parent Segment' :'General Segment'}}</span>
          <div class="parent-score">
            <span class="score-label">Hit Score:</span>
            <span class="score-value">{{ formatScore(parentSegment.score) }}</span>
          </div>
        </div>
        <div class="parent-content" v-html="parentSegment.content ? md.render(parentSegment.content) : 'No Content'">
        </div>
      </div>

      <!-- Child segment area -->
        <div class="sub-segments" v-if="segmentList.length > 0">
        <div class="segment-header">
          <span class="sub-badge">Hit {{ segmentList.length }} child segments</span>
        </div>
      </div>
      <div class="collapse-wrapper" v-if="segmentList.length > 0">
        <el-collapse 
          v-model="activeNames" 
          class="section-collapse"
          :accordion="false"
        >
        <el-collapse-item 
          v-for="(segment, index) in segmentList" 
          :key="index"
          :name="index"
          class="segment-collapse-item"
        >
          <template slot="title">
            <span class="segment-badge">C-{{ index + 1 }}</span>
            <span class="segment-score">
              <span class="score-label">Hit Score:</span>
              <span class="score-value">{{ formatScore(childscore[index]) }}</span>
            </span>
          </template>
          <div class="segment-content" v-html="segment.content ? md.render(segment.content) : 'No Content'"></div> 
        </el-collapse-item>
        </el-collapse>
      </div>
    </div>
  </el-dialog>
</template>

<script>
import { formatScore } from '@/utils/util'
import { md } from "@/mixins/marksown-it";
export default {
  name: 'SectionShow',
  data() {
    return {
      md: md,
      dialogVisible: false,
      activeNames: [],
      parentSegment: {},
      segmentList: [],
      childscore:[]
    }
  },
  methods: {
    formatScore,
    // Show dialog
    showDiaglog(data) {
      if (data) {
        // Update parent segment data
        if (data.searchList) {
          this.parentSegment = {
            score: parseFloat(data.score) || 0,
            content: data.searchList.snippet||'No Content',
            contentType: data.searchList.contentType
          };
        }
        
        // Update child segment data
        if (data.searchList && Array.isArray(data.searchList.childContentList)) {
          this.childscore = data.searchList.childScore;
          this.segmentList = data.searchList.childContentList.map(segment => ({
            content: segment.childSnippet || '',
            autoSave: Boolean(segment.autoSave),
            score: parseFloat(segment.score) || 0
          }));
          // Set all collapse items to expanded state
          this.activeNames = this.segmentList.map((_, index) => index);
        }
        
        
      }
      this.dialogVisible = true;
    },
    // Close弹框
    handleClose() {
      this.dialogVisible = false;
    },
  }
}
</script>

<style lang="scss" scoped>
.section-dialog {
  /deep/ .el-dialog__body {
    padding: 0 20px 20px 20px;
    max-height:70vh;
    overflow-y: auto;
  }
}

.section-show-container {
  .parent-segment {
    padding: 20px 20px 0 20px;
    background: #fff;
    border-radius: 8px;
    .segment-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 15px;
      
      .parent-badge {
        background-color: #d2d7ff;
        color: $color;
        padding: 6px 12px;
        border-radius: 6px;
        font-size: 12px;
        font-weight: 500;
      }
      
      .parent-score {
        display: flex;
        align-items: center;
        
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
    }
    
    .parent-content {
      text-align: left;
      background-color: #f7f8fa;
      padding: 15px;
      border-radius: 6px;
      border: 1px solid $color;
      line-height: 1.6;
      
      .parent-item {
        margin-bottom: 10px;
        font-size: 14px;
        color: #333;
        line-height: 1.5;
        text-align: left;
        
        .segment-action {
          color: #999;
          font-size: 12px;
          margin-left: 8px;
        }
      }
    }
  }
  
  .sub-segments {
    padding: 20px 20px 0 20px;
    
    .segment-header {
      margin-bottom: 15px;
      
      .sub-badge {
        background-color: #d2d7ff;
        color: $color;
        padding: 6px 12px;
        border-radius: 6px;
        font-size: 12px;
        font-weight: 500;
      }
    }
  }
  
  .collapse-wrapper {
    border: 1px solid $color;
    background: #f7f8fa;
    border-radius: 6px;
    margin: 0 20px 20px 20px;
    overflow: hidden;
  }
  
  .section-collapse {
    padding: 0;
    background: transparent;
    
    /deep/ .el-collapse {
      border: none !important;
      border-top: none !important;
      border-bottom: none !important;
      border-left: none !important;
      border-right: none !important;
      background: transparent !important;
      border-radius: 0 !important;
    }
    
    /deep/ .el-collapse-item {
      border: none !important;
      margin-bottom: 0 !important;
      background: transparent !important;
    }
    
    
    /deep/ .el-collapse-item__header {
      background: transparent;
      border: none;
      padding: 0 10px;
      display: flex;
      justify-content: space-between;
      align-items: center;
      
      &:hover {
        background: #f0f2f5;
      }
    }
    
    /deep/ .el-collapse-item__content {
      padding: 0;
      background: transparent;
    }
    
    /deep/ .el-collapse-item__arrow {
      display: none !important;
    }
    
    .segment-collapse-item {
      .segment-badge {
        color: $color;
        font-size: 14px;
        font-weight: 600;
        margin-right: 15px;
      }
      
      .segment-score {
        display: flex;
        align-items: center;
        
        .score-label {
          font-size: 12px;
          color: $color;
          margin-right: 5px;
        }
        
        .score-value {
          font-size: 12px;
          color: $color;
          font-weight: 600;
          font-family: 'Courier New', monospace;
        }
      }
      
      .segment-content {
        font-size: 14px;
        color: #333;
        line-height: 1.6;
        text-align: left;
        padding:10px;
      }
    }
  }
}

// Response式设计
@media (max-width: 768px) {
  .section-show-container {
    .parent-segment,
    .sub-segments {
      padding: 15px;
    }
    
    .section-collapse {
      /deep/ .el-collapse-item__header {
        flex-direction: column;
        align-items: flex-start;
        
        .segment-badge {
          margin-bottom: 8px;
          margin-right: 0;
        }
        
        .segment-score {
          position: static;
          transform: none;
          margin-top: 8px;
        }
      }
    }
  }
}
</style>