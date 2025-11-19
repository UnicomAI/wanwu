<template>
  <el-dialog
    :visible.sync="dialogVisible"
    width="600px"
    :before-close="handleClose"
  >
    <!-- Title and Description -->
    <template slot="title">
      <div class="custom-title">
        <div class="header-section">
          <h2 class="dialog-title">{{actionDetail.action && actionDetail.action.name}}</h2>
          <!-- <p class="dialog-subtitle">{{actionDetail.action && actionDetail.action.description}}</p> -->
        </div>
      </div>
    </template>

    <!-- API Key Section -->
    <div class="api-key-section">
      <div class="api-key-label">API Key</div>
      <div class="api-key-input-group">
        <el-input
          v-model="apiKey"
          :placeholder="$t('agent.toolDetail.inputApikey')"
          class="api-key-input"
          showPassword
          @input="inputApiKey"
        />
        <div class="api-key-buttons">
          <el-button style="width: 100px" size="mini" type="primary"  @click="changeApiKey" :disabled="isDisabled">
            {{actionDetail.apiKey ? $t('tool.builtIn.update') : $t('common.confirm.confirm')}}
          </el-button>
        </div>
      </div>
    </div>
    <!-- rerank Section -->
    <div class="api-key-section rerank-section" v-if="currentItem && currentItem.toolId === 'bochawebsearch'">
    <div class="api-key-label">Rerank</div>
      <el-select
          v-model="rerankId"
          :placeholder="$t('agent.toolDetail.selectRerank')"
          @visible-change="rerankVisible"
          @change="rerankChange"
          :loading-text="$t('agent.toolDetail.modelLoadingText')"
          class="cover-input-icon"
          style="width:100%;"
          filterable
          clearable
      >
          <el-option
          v-for="(item,index) in rerankOptions"
          :key="item.modelId"
          :label="item.displayName"
          :value="item.modelId"
          >
          </el-option>
      </el-select>
    </div>

    <div class="parameters-section">
      <el-table :data="parametersData" border class="parameters-table">
        <el-table-column prop="key" :label="$t('agent.toolDetail.parameters')" width="120" />
        <el-table-column prop="type" :label="$t('agent.toolDetail.type')" width="100" />
        <el-table-column prop="description" :label="$t('agent.toolDetail.description')" />
        <el-table-column
          prop="required"
          :label="$t('agent.toolDetail.required')"
          width="100"
          align="center"
          :formatter="(row, column, cellValue) => (cellValue ? $t('agent.toolDetail.yes'): $t('agent.toolDetail.no'))"
        />
      </el-table>
    </div>
  </el-dialog>
</template>

<script>
import {toolActionDetail,updateRerank} from "@/api/agent";
import { changeApiKey } from "@/api/mcp";
import { getRerankList} from "@/api/modelAccess";
export default {
  data() {
    return {
      dialogVisible: false,
      actionDetail:{
        action: {
          name: '',
          description: ''
        },
        apiKey: ''
      },
      apiKey: '',
      parametersData: [],
      currentItem:null,
      rerankOptions:[],
      rerankId:'',
      isDisabled:false
    }
  },
  methods: {
    inputApiKey(){
      this.isDisabled = false;
    },
     rerankVisible(val){
      if(val){
          this.getRerankData();
      }
    },
    rerankChange(val){
      if(val){
        const data = {
          assistantId:this.$route.query.id,
          toolId:this.currentItem.toolId,
          toolConfig:{
            rerankId:this.rerankId
          }
        }
        updateRerank(data).then(res =>{
          if(res.code === 0){
            this.$emit('updateDetail')
          }
        }).catch(() =>{})
      }
    },
    getRerankData(){
      getRerankList().then(res =>{
          if(res.code === 0){
          this.rerankOptions = res.data.list || []
          }
      })
    },
    changeApiKey(){
      changeApiKey({
        apiKey: this.apiKey,
        toolSquareId: this.currentItem.toolId
      }).then((res) => {
        if (res.code === 0) {
          this.$message.success(this.$t('common.message.success'))
          this.getDeatil(this.currentItem)
        }
      })
    },
    handleClose() {
      this.dialogVisible = false;
    },
    showDiaglog(n){
      this.dialogVisible = true;
      this.currentItem = n;
      if(n.toolId === 'bochawebsearch'){
        this.getRerankData();
        this.rerankId = n.toolConfig.rerankId || '';
      }
      this.getDeatil(n);
    },
    getDeatil(n){
      toolActionDetail({actionName:n.actionName,toolId:n.toolId,toolType:n.toolType}).then(res =>{
        if(res.code === 0){
          if(res.data.apiKey !== ''){
            this.isDisabled = true
          }
          this.actionDetail = res.data || {}
          const base = { action: { name: '', description: '' }, apiKey: '' }
          const payload = res.data || {}
          this.actionDetail = Object.assign({}, base, payload, {
            action: Object.assign({}, base.action, (payload.action || {})),
            apiKey: typeof payload.apiKey === 'string' ? payload.apiKey : ''
          })
          this.apiKey = res.data.apiKey
          const {properties,required} = (this.actionDetail.action && this.actionDetail.action.inputSchema) ? this.actionDetail.action.inputSchema : { properties: {}, required: [] };
          this.parametersData = this.toParametersData(properties,required);
        }
      }).catch(() =>{})
    },
    toParametersData(schemaObject, requiredKeys){
      return Object.entries(schemaObject).map(([key, def]) => ({
        key,
        type: def.type,
        description: def.description || '',
        required: requiredKeys.includes(key)
      }));
    }
  }
}
</script>

<style lang="scss" scoped>
/deep/{
  .el-dialog__body {
    padding:20px!important;
  }
  .el-dialog__header{
    padding:10px 20px !important; 
    height:45px !important;
    border-bottom:1px solid #dbdbdb;
  }
}

.header-section {
  .dialog-title {
    font-size: 20px;
    font-weight: bold;
    color: #333;
    margin: 0 0 8px 0;
  }
  
  .dialog-subtitle {
    font-size: 14px;
    color: #666;
    margin: 0;
  }
}

.api-key-section {
  margin-bottom:15px;
  display:flex;
  .api-key-label {
    font-size: 14px;
    font-weight: 500;
    color: #333;
    min-width:56px;
    white-space: nowrap;
  }
  
  .api-key-input-group {
    display: flex;
    align-items: center;
    flex:1;
    gap: 12px;
    .api-key-input {
      flex: 1;
    }
    
    .api-key-buttons {
      display: flex;
      flex-direction: row;
      gap: 8px;
      .confirm-btn,
      .update-btn {
        width: 60px;
        height: 32px;
      }
    }
  }
}

.parameters-section {
  .parameters-table {
    width: 100%;
  }
}

.dialog-footer {
  text-align: center;
  padding: 20px 0 0 0;
  
  .final-confirm-btn {
    width: 120px;
    height: 40px;
    font-size: 16px;
  }
}
</style>