<template>
  <div class="agent-from-content" :class="{ 'isDisabled': isPublish }">
    <div class="form-header">
      <div class="header-left">
        <span class="el-icon-arrow-left btn" @click="goBack"></span>
        <!-- <span class="header-left-title">Text Q&AEdit</span> -->
        <div class="basicInfo">
          <div class="img">
            <img :src="editForm.avatar.path ? `/user/api`+ editForm.avatar.path : '@/assets/imgs/bg-logo.png'"  />
          </div>
          <div class="basicInfo-desc">
            <span class="basicInfo-title">{{editForm.name || '无Information'}}</span>
            <span class="el-icon-edit-outline editIcon" @click="editAgent"></span>
            <LinkIcon type="rag" />
            <p>{{editForm.desc || '无Information'}}</p>
          </div>
          </div>
      </div>
      <div class="header-right">
        <div class="header-api">
          <el-tag  effect="plain" class="root-url">API根地址</el-tag>
          {{apiURL}}
        </div>
        <el-button @click="openApiDialog" plain class="apikeyBtn" size="small" >
          <img :src="require('@/assets/imgs/apikey.png')" />
          API密钥
        </el-button>
        <el-button size="small" type="primary" @click="handlePublish" style="padding:13px 12px;">Publish<span class="el-icon-arrow-down" style="margin-left:5px;"></span></el-button>
        <div class="popover-operation" v-if="showOperation">
          <div>
            <el-radio :label="'private'" v-model="scope">私密Publish为App：仅自己可见</el-radio>
          </div>
          <div>
            <el-radio :label="'organization'" v-model="scope">公开Publish为App：组织内可见</el-radio>
          </div>
          <div>
            <el-radio :label="'public'" v-model="scope">公开Publish为App：全局可见</el-radio>
          </div>
          <div class="saveBtn">
            <el-button size="mini" type="primary" @click="savePublish">保 存</el-button>
          </div>
        </div>
      </div>
    </div>
    <div class="agent_form">
      <div class="drawer-form">
        <div class="model-box">
          <div class="block prompt-box">
            <p class="block-title common-set">
              <span class="common-set-label">
                <img :src="require('@/assets/imgs/require.png')" class="required-label"/>
                Model选择
                <span class="model-tips">[ 暂Not Supported选择图文问答类Model ]</span>
              </span>
              <span class="el-icon-s-operation operation" @click="showModelSet"></span>
            </p>
            <div class="rl">
              <el-select
                v-model="editForm.modelParams"
                placeholder="可EnterModelNameSearch"
                @visible-change="visibleChange"
                loading-text="ModelLoading..."
                class="cover-input-icon model-select"
                :disabled="isPublish"
                :loading="modelLoading"
                filterable
                value-key="modelId"
              >
                <el-option
                  v-for="(item,index) in modleOptions"
                  :key="item.modelId"
                  :label="item.displayName"
                  :value="item.modelId"
                >
                  <div class="model-option-content">
                    <span class="model-name">{{ item.displayName }}</span>
                    <div class="model-select-tags" v-if="item.tags && item.tags.length > 0">
                      <span
                        v-for="(tag, tagIdx) in item.tags"
                        :key="tagIdx"
                        class="model-select-tag"
                      >{{ tag.text }}</span>
                    </div>
                  </div>
                </el-option>
              </el-select>
            </div>
          </div>
          <div class="block recommend-box">
            <p class="block-title common-set">
              <span class="common-set-label">
                <img :src="require('@/assets/imgs/require.png')" class="required-label"/>
                关联Knowledge Base
              </span>
              <span>
                <span class="common-add" @click="showKnowledgeDiglog">
                  <span class="el-icon-plus"></span>
                  <span class="handleBtn">Add</span>
                </span>
              </span>
            </p>
            <div class="rl knowledge-conent">
              <div class="tool-right tool">
                  <div class="action-list">
                    <div v-for="(n,i) in editForm.knowledgebases" class="action-item" :key="'knowledge'+ i">
                       <div class="name" style="color: #333">
                        <span>{{n.name}}</span>
                       </div>
                        <div class="bt">
                          <el-tooltip class="item" effect="dark" content="元Data 滤" placement="top-start">
                            <span class="el-icon-setting del" @click="showMetaSet(n,i)" style="margin-right:10px;"></span>
                          </el-tooltip>
                          <span class="el-icon-delete del" @click="delKnowledge(i)"></span>
                      </div>
                    </div>
                  </div>
              </div>
            </div>
          </div>
        </div>
        <div class="block safety-box">
            <p class="block-title common-set">
              <span class="common-set-label">
                <img :src="require('@/assets/imgs/require.png')" class="required-label"/>
                检索方式Configuration
              </span>
            </p>
          <div class="rl">
            <searchConfig ref="searchConfig" @sendConfigInfo="sendConfigInfo" :setType="'rag'" :config="editForm.knowledgeConfig"/>
          </div>
        </div>
        <div class="block prompt-box safety-box">
            <p class="block-title tool-title">
            <span class="block-title-text">
              安全护栏Configuration
              <el-tooltip class="item" effect="dark" content="实 when 拦截高风险Content of EnterAnd输出，保障Content安全合规。" placement="top">
                  <span class="el-icon-question question-tips"></span>
              </el-tooltip>
            </span>
            <span class="common-add">
              <span @click="showSafety">
                <span class="el-icon-s-operation" ></span>
                <span class="handleBtn" style="margin-right:10px;">Configuration</span>
              </span>
              <el-switch v-model="editForm.safetyConfig.enable" :disabled="!(editForm.safetyConfig.tables || []).length"></el-switch>
            </span>
          </p>
        </div>
        <div class="block prompt-box safety-box" v-if="showGraphSwitch">
          <graphSwitch ref="graphSwitch" @graphSwitchchange="graphSwitchchange" :label="'知识图谱'" :graphSwitch="editForm.knowledgeConfig.useGraph"/>
        </div>
      </div>
      <div class="drawer-test">
        <Chat :chatType="'test'" :editForm="editForm"/>
      </div>
    </div>
    <!-- EditAgent -->
    <CreateTxtQues ref="createTxtQues" :type="'edit'" :editForm="editForm" @updateInfo="getDetail"/>
    <!-- ModelSetting -->
    <ModelSet @setModelSet="setModelSet" ref="modelSetDialog" :modelConfig="editForm.modelConfig" />
    <!-- Knowledge BaseSetting -->
    <knowledgeSet @setKnowledgeSet="setKnowledgeSet" ref="knowledgeSetDialog" :knowledgeConfig="editForm.knowledgeConfig" />
    <!-- apikey -->
    <ApiKeyDialog ref="apiKeyDialog" :appId="editForm.appId" :appType="'rag'" />
    <setSafety ref="setSafety" @sendSafety="sendSafety" />
    <!-- Knowledge Base选择 -->
    <knowledgeSelect ref="knowledgeSelect" @getKnowledgeData="getKnowledgeData" />
    <!-- 元DataSetting -->
    <el-dialog
      :visible.sync="metaSetVisible"
      width="1050px"
      class="metaSetVisible"
      :before-close="handleMetaClose">
      <template #title>
         <div class="metaHeader">
          <h3>Configuration元Data 滤</h3>
          <span>[ BySetting of 元Data，对Knowledge Base内Information进行更加细化 of 筛选与检索控制。]</span>
         </div>
      </template>
      <metaSet ref="metaSet"  :knowledgeId="currentKnowledgeId" :currentMetaData="currentMetaData"/>
      <span slot="footer" class="dialog-footer">
        <el-button @click="handleMetaClose">取 消</el-button>
        <el-button type="primary" @click="submitMeta">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import {getApiKeyRoot,appPublish} from "@/api/appspace";
import CreateTxtQues from "@/components/createApp/createRag.vue"
import ModelSet from "./modelSetDialog.vue";
import metaSet from "@/components/metaSet";
import knowledgeSet from "./knowledgeSetDialog.vue"
import ApiKeyDialog from "./ApiKeyDialog";
import setSafety from "@/components/setSafety";
import { getRerankList,selectModelList } from "@/api/modelAccess";
import { getRagInfo,updateRagConfig } from "@/api/rag";
import Chat from "./chat";
import searchConfig from '@/components/searchConfig.vue';
import LinkIcon from "@/components/linkIcon.vue";
import graphSwitch from "@/components/graphSwitch.vue"
import knowledgeSelect from "@/components/knowledgeSelect.vue"
export default {
  components: {
    LinkIcon,
    Chat,
    CreateTxtQues,
    ModelSet,
    knowledgeSet,
    ApiKeyDialog,
    setSafety,
    searchConfig,
    knowledgeSelect,
    metaSet,
    graphSwitch
  },
  data() {
    return {
      knowledgeIndex:-1,
      currentKnowledgeId:'',
      currentMetaData:{},
      metaSetVisible:false,
      rerankOptions:[],
      showOperation:false,
      scope:'public',
      localKnowledgeConfig:{},
      editForm:{
        appId:'',
        avatar:{},
        name:'',
        desc:'',
        modelParams:'',
        modelConfig:{
          temperature:0.14,
          topP:0.85,
          frequencyPenalty:1.1,
          temperatureEnable:true,
          topPEnable:true,
          frequencyPenaltyEnable:true
        },
        rerankParams:'',
        knowledgebases:[],
        knowledgeConfig:{
          keywordPriority: 0.8, //关Key词权重
          matchType: "mix", //vector（向量检索）、text（Text检索）、mix（混合检索：向量+Text）
          priorityMatch: 1, //权重Match，只Has在混合检索模式下，选择权重Setting后，这个才Setting为1
          rerankModelId: "", //rerankModelid
          semanticsPriority: 0.2, //语义权重
          topK: 5, //topK Get最高 of 几行
          threshold: 0.4, // 滤分数阈Value
          maxHistory:0,//
          useGraph:false
        },
        safetyConfig:{
          enable: false,
          tables:[]
        }
      },
      initialEditForm:null,
      apiURL:'',
      modelLoading:false,
      wfDialogVisible: false,
      workFlowInfos: [],
      workflowList: [],
      modelParams:'',
      rerankParams:'',
      platform: this.$platform,
      isPublish: false,
      modleOptions: [],
      selectKnowledge: [],
      loadingPercent: 10,
      nameStatus: "",
      saved: false, //Button
      loading: false, //Button
      t: null,
      logoFileList: [],
      debounceTimer:null, //防抖计 when 器
      isUpdating: false, // 防止重复Update标记
      isSettingFromDetail: false, // 防止DetailsDataTriggerUpdate标记
    };
  },
  watch:{
    editForm: {
    handler(newVal) {
      // IfYes从DetailsSetting of Data，不TriggerUpdate逻辑
      if (this.isSettingFromDetail) {
        return;
      }
      
      if(this.debounceTimer){
        clearTimeout(this.debounceTimer)
      }
      this.debounceTimer = setTimeout(() =>{
          const props = ['modelParams', 'modelConfig', 'knowledgebases', 'knowledgeConfig','safetyConfig'];
          const changed = props.some(prop => {
          return JSON.stringify(newVal[prop]) !== JSON.stringify(
              (this.initialEditForm || {})[prop]
            );
          });
          if (changed && !this.isUpdating) {
            const isMixPriorityMatch = newVal['knowledgeConfig']['matchType'] === 'mix' && newVal['knowledgeConfig']['priorityMatch'];
            if(newVal['modelParams']!== '' || (isMixPriorityMatch && !newVal['knowledgeConfig']['rerankModelId'])){
              this.updateInfo();
            }
          }
      },500)
    },
    deep: true
    },
  },
  computed:{
      showGraphSwitch() {
      return this.editForm.knowledgebases && this.editForm.knowledgebases.some(item => item.graphSwitch === 1)
    }
  },
  mounted() {
    this.initialEditForm = JSON.parse(JSON.stringify(this.editForm));
  },
  created() {
    this.getModelData(); //GetModelList
    this.getRerankData(); //GetrerankModel
    if (this.$route.query.id) {
      this.editForm.appId = this.$route.query.id;
      setTimeout(() => {
        this.getDetail();//GetDetails
        this.apiKeyRootUrl(); //Getapi跟Address
      }, 500);
    }
        //判断YesNoPublish
    if (this.$route.query.publish) {
      this.isPublish = true;
    }
  },
  methods: {
    graphSwitchchange(val){
      this.editForm.knowledgeConfig.useGraph = val;
    },
    submitMeta(){
      const metaData  = this.$refs.metaSet.getMetaData();
      if(this.$refs.metaSet.validateRequiredFields(metaData['metaDataFilterParams']['metaFilterParams'])){
        this.$message.warning('Missing information, please complete')
        return
      }
      this.$set(this.editForm.knowledgebases, this.knowledgeIndex, { ...this.editForm.knowledgebases[this.knowledgeIndex], ...metaData });
      this.metaSetVisible = false;
    },
    delKnowledge(index){
      this.editForm.knowledgebases.splice(index,1)
    },
    handleMetaClose(){
      this.metaSetVisible = false;
    },
    getKnowledgeData(data){
      const originalIds = new Set(this.editForm.knowledgebases.map(item => item.id));
      const newItems = data.filter(item => !originalIds.has(item.id));
      this.editForm.knowledgebases.push(...newItems);
    },
    showMetaSet(e,index){
      this.currentKnowledgeId = e.id;
      this.knowledgeIndex = index;
      this.currentMetaData = {}
      this.$nextTick(() =>{
         this.currentMetaData = e.metaDataFilterParams;
      })
      this.metaSetVisible = true;
    },
    showKnowledgeDiglog(){
      this.$refs.knowledgeSelect.showDialog(this.editForm.knowledgebases)
    },
    sendConfigInfo(data){
      this.editForm.knowledgeConfig = { ...data.knowledgeMatchParams };
    },
    sendSafety(data){
      const tablesData = data.map(({ tableId, tableName }) => ({ tableId, tableName }));
      this.editForm.safetyConfig.tables = tablesData;
    },
    showSafety(){
      this.$refs.setSafety.showDialog(this.editForm.safetyConfig.tables);
    },
    goBack(){
      this.$router.go(-1);
    },
    getDetail(){//GetDetails
      this.isSettingFromDetail = true; // Setting标志位，防止TriggerUpdate逻辑
      getRagInfo({ragId:this.editForm.appId}).then(res =>{
        if(res.code === 0){
            this.editForm.avatar = res.data.avatar;
            this.editForm.name = res.data.name;
            this.editForm.desc = res.data.desc;
            this.editForm.modelParams = res.data.modelConfig.modelId;
            if(res.data.safetyConfig && res.data.safetyConfig !== null){
              this.editForm.safetyConfig = res.data.safetyConfig;
            }
            if(res.data.modelConfig.config !== null){
              this.editForm.modelConfig = res.data.modelConfig.config;
            }
            this.editForm.rerankParams = res.data.rerankConfig.modelId;
            const knowledgeData = res.data.knowledgeBaseConfig.knowledgebases;
            if(knowledgeData && knowledgeData.length > 0){
              this.editForm.knowledgebases = knowledgeData;
            }
            if(res.data.knowledgeBaseConfig.config !== null){
              this.editForm.knowledgeConfig = res.data.knowledgeBaseConfig.config;
              const {matchType,priorityMatch} = res.data.knowledgeBaseConfig.config
              if(matchType === ''){
                this.editForm.knowledgeConfig = {
                  ...this.editForm.knowledgeConfig,
                  matchType:'mix',
                  priorityMatch:1
                }
              }
            }
            
            this.editForm.knowledgeConfig.rerankModelId = res.data.rerankConfig.modelId;
            // UsenextTick确保所HasDataSetting完成后再Reset标志位
            this.$nextTick(() => {
              this.isSettingFromDetail = false;
            });
        } else {
          this.isSettingFromDetail = false;
        }
      }).catch(() => {
        this.isSettingFromDetail = false;
      });
    },
    getRerankData(){
      getRerankList().then(res =>{
        if(res.code === 0){
          this.rerankOptions = res.data.list || []
        }
      })
    },
    handlePublish(){
      this.showOperation = !this.showOperation;
    },
    savePublish(){
      const { matchType, priorityMatch, rerankModelId } = this.editForm.knowledgeConfig;
      const isMixPriorityMatch = matchType === 'mix' && priorityMatch;
      if(this.editForm.modelParams === ''){
        this.$message.warning('Please selectModel！')
        return false
      }
      if(!isMixPriorityMatch && !rerankModelId){
        this.$message.warning('Please选rerank择Model！')
        return false
      }
      if(this.editForm.knowledgebases.length === 0){
        this.$message.warning('Please select关联Knowledge Base！')
        return false
      }
      const data = {appId:this.editForm.appId,appType:'rag',publishType:this.scope}
      appPublish(data).then(res =>{
        if(res.code === 0){
          this.$router.push({path:'/explore'})
        }
      })
    },
    apiKeyRootUrl(){
      const data = {appId:this.editForm.appId,appType:'rag'}
      getApiKeyRoot(data).then(res =>{
        if(res.code === 0){
          this.apiURL = res.data || ''
        }
      })
    },
    openApiDialog(){
      this.$refs.apiKeyDialog.showDialog()
    },
    setModelSet(data){
      this.editForm.modelConfig = data;
    },
    showModelSet(){
      this.$refs.modelSetDialog.showDialog()
    },
    showKnowledgeSet(){
      this.$refs.knowledgeSetDialog.showDialog()
    },
    setKnowledgeSet(data){
      this.editForm.knowledgeConfig = data;
    },
    editAgent(){
      this.$refs.createTxtQues.openDialog()
    },
    visibleChange(val){
      if(val){
        this.getModelData();
      }
    },
    rerankVisible(val){
      if(val){
        this.getRerankData();
      }
    },
    async getModelData() {
      this.modelLoading = true;
      const res = await selectModelList();
      if (res.code === 0) {
        this.modleOptions = (res.data.list || []).filter(item => {
            return item.config && item.config.visionSupport !== 'support'; 
        });

        this.modelLoading = false;
      }
      this.modelLoading = false;
    },
    async updateInfo() {
      if (this.isUpdating) return; // 防止重复Call
      
      this.isUpdating = true;
      try {
        //Knowledge BaseData
        //ModelData
        const modeInfo = this.modleOptions.find(item => item.modelId === this.editForm.modelParams)
        if(this.editForm.knowledgeConfig.matchType === 'mix' && this.editForm.knowledgeConfig.priorityMatch === 1){
          this.editForm.knowledgeConfig.rerankModelId = ''
        }
        const rerankInfo = this.rerankOptions.find(item => item.modelId === this.editForm.knowledgeConfig.rerankModelId)
        let fromParams = {
          ragId:this.editForm.appId,
          knowledgeBaseConfig:{
            knowledgebases:this.editForm.knowledgebases,
            config:this.editForm.knowledgeConfig
          },
          modelConfig:{
            config:this.editForm.modelConfig,
            displayName: modeInfo.displayName,
            model: modeInfo.model,
            modelId: modeInfo.modelId,
            modelType: modeInfo.modelType,
            provider: modeInfo.provider,
          },
          rerankConfig:{
            displayName: rerankInfo ? rerankInfo.displayName : '',
            model: rerankInfo ? rerankInfo.model : '',
            modelId: rerankInfo ? rerankInfo.modelId : '',
            modelType: rerankInfo ? rerankInfo.modelType : '',
            provider: rerankInfo ? rerankInfo.provider : '',
          },
          safetyConfig:this.editForm.safetyConfig,
        }
        const res = await updateRagConfig(fromParams)
        
        // UpdateSuccess后，Update initialEditForm 避免重复Trigger
        if (res.code === 0) {
          this.initialEditForm = JSON.parse(JSON.stringify(this.editForm));
          this.getDetail();//GetDetails
        }
      } catch (error) {
        console.error('UpdateConfigurationFailed:', error);
      } finally {
        this.isUpdating = false;
      }
    }
  }
};
</script>

<style lang="scss" scoped>
.model-option-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  
  .model-name {
    flex-shrink: 0;
    font-weight: 500;
  }
  
  .model-select-tags {
    display: flex;
    flex-wrap: nowrap;
    gap: 4px;
    flex-shrink: 0;
    margin-top: 4px;
    .model-select-tag {
      background-color: #f0f2ff;
      color: $color;
      border-radius: 4px;
      padding: 2px 8px;
      font-size: 10px;
      line-height: 1.2;
    }
  }
}
.isDisabled .header-right,.isDisabled .drawer-form > div{
  user-select: none;
  pointer-events: none !important;      
}
/deep/{
  .apikeyBtn{
    padding: 12px 10px;
    border:1px solid $btn_bg;
    color: $btn_bg;
    display:flex;
    align-items: center;
    img{
      height:14px;
    }
  }
  .metaSetVisible{
    .el-dialog__header{
      border-bottom:1px solid #dbdbdb;
    }
    .el-dialog__body{
      max-height:400px;
      overflow-y: auto;
    }
  }
}

.metaHeader{
  display:flex;
  justify-content: flex-start;
  h3{
    font-size:18px;
  }
  span{
    margin-left:10px;
    color:#666;
    display:inline-block;
    padding-top:5px;
  }
}

.question {
  cursor: pointer;
  color: #999;
  margin-left: 8px;
}
::selection {
  color: #1a2029;
  background: #c8deff;
}
.question {
  cursor: pointer;
  color: #ccc;
  margin-left: 6px;
}
.form-header{
  width:100%;
  height:60px;
  display:flex;
  justify-content:space-between;
  align-items:center;
  padding:0 20px;
  position:relative;
  border-bottom:1px solid #dbdbdb;
  .popover-operation{
    position:absolute;
    bottom:-122px;
    right:20px;
    background:#fff;
    box-shadow: 0px 1px 7px rgba(0, 0, 0, 0.3);
    padding:10px 20px;
    border-radius:6px;
    z-index:999;
    .saveBtn{
      display:flex;
      justify-content:center;
      padding:10px 0;
    }
  }
  .header-left{
    display: flex;
    align-items: center;
    .btn{
      margin-right:10px;
      font-size:18px;
      cursor: pointer;
    }
    .header-left-title{
      font-size:18px;
      color: $color_title;
      font-weight: bold;
    }
    .basicInfo{
      display: flex;
      align-items:center;
      border-radius:8px;
      padding:10px 0;
      .img{
        padding:10px;
        img{
          border:1px solid #eee;
          border-radius:6px;
          width:32px;
          height:32px;
          object-fit: cover;
        }
      }
      .basicInfo-desc{
        flex:1;
        .editIcon{
          cursor: pointer;
          margin-left: 5px;
          font-size: 16px;
          color: #6b7280;
        }
      }
      .basicInfo-title{
        display:inline-block;
        font-weight:800;
        font-size:14px;
      }
      p {
        color: #6b7280;
        font-size: 12px;
        margin: 0;
        line-height: 1.4;
      }
    }
  }
  .header-right{
    display: flex;
    align-items:center;
    .header-api{
    padding: 6px 10px;
    box-shadow: 1px 2px 2px #ddd;
    background-color: #fff;
    margin:0 10px;
    border-radius:6px;
    .root-url{
      background-color:#ECEEFE;
      color:$color;
      border:none;
    }
  }
  }
  
}
.agent-from-content{
  height:100%;
  width:100%;
  overflow: hidden;
}
.agent_form{
  padding:0 10px;
  display: flex;
  justify-content: space-between;
  gap: 20px;
  height:calc(100% - 60px);
  .drawer-form {
    width:50% ;
    position: relative;
    height:100%;
    padding:0 10px;
    border-radius: 6px;
    overflow-y: auto;
    display:flex;
    flex-direction: column;
    margin:10px 0;
    .editIcon{
      font-size: 16px;
      margin-left: 5px;
      cursor: pointer;
    }
  /deep/.el-input__inner,
  /deep/.el-textarea__inner {
    background-color: transparent !important;
    border: 1px solid #d3d7dd !important;
    font-family: "Microsoft YaHei", Arial, sans-serif;
    padding: 15px;
  }
  .flex {
    width: 100%;
    display: flex;
    justify-content: space-between;
  }
  .model-box{
    background:#F7F8FA;
    box-shadow: 0 1px 4px 0 rgba(0, 0, 0, 0.15);
    border-radius:8px;
    padding:20px 15px;
    margin-bottom:10px;
    .block{
      margin-bottom:10px;
    }
    .model-tips{
      color: #999;
      margin-left: 10px;
      font-weight:normal!important;
    }
  }
  .safety-box{
    background:#F7F8FA;
    box-shadow: 0 1px 4px 0 rgba(0, 0, 0, 0.15);
    border-radius:8px;
    padding:10px 15px;
    margin-top:14px;
    .block-title{
      line-height: 30px;
      font-size: 15px;
      font-weight: bold;
      display: flex;
      align-items: center;
      .block-title-text{
        font-size:15px;
      }
      .handleBtn{
        cursor: pointer;
      }
    }
    .tool-title{
      justify-content: space-between;
    }
  }
  .common-set{
    display:flex;
    justify-content: space-between;
    .common-set-label{
      display:flex;
      align-items:center;
      font-size: 15px;
      font-weight:bold;
    }
  }
  /*通用*/
  .block {
    margin-bottom: 24px;
    .block-title {
      line-height: 30px;
      font-size: 15px;
      font-weight: bold;
      display: flex;
      align-items:center;
      .title_tips {
        color: #999;
        margin-left: 20px;
        font-weight: normal;
      }
    }
    .tool-conent{
      display:flex;
      justify-content:space-between;
      gap:10px;
      .tool{
        width:50%;
        height:300px;
        max-height:300px;
      }
    }
    .knowledge-conent{
      display: flex;
      justify-content: space-between;
      gap: 10px;
      .tool {
        width: 100%;
        max-height: 300px;
        .action-list {
          width: 100%;
          display: grid;
          grid-template-columns: repeat(2, minmax(0, 1fr));
          gap: 10px;
        }
      }
    }
    .model-select{
      width:100%;
    }
    .operation{
      text-align:center;
      cursor:pointer;
      font-size: 16px;
      padding-right:10px;
    }
    .common-add{
      cursor: pointer;
    }
    .operation:hover{
      color:$color;
    }
    .tips {
      display: flex;
      align-items: center;
      margin-bottom: 5px;
      .block-title-tips {
        color: #ccc;
        margin-right: 10px;
      }
    }
    .paramsSet {
      padding: 10px;
    }
    .required-label{
      width: 18px;
      height:18px;
      margin-right: 4px;
    }
    .block-tip {
      color: #919eac;
    }
  }
  .el-input__count {
    color: #909399;
    background: #fafafa;
    position: absolute;
    font-size: 12px;
    bottom: 5px;
    right: 10px;
  }
  /*新建App*/
  .name-box {
    height: 90px;
    line-height: 90px;
    font-size: 22px;
    display: flex;
    .name-input {
      width: 100%;
    }
    .input-echo {
      font-size: 22px;
      .name-edit {
        margin-left: 20px;
        cursor: pointer;
        font-size: 16px;
      }
    }
  }
  .logo-box {
    margin-top: 20px;
    .right-input-box {
      flex: 1;
      width: 0;
      margin-left: 20px;
    }
    .instructions-input {
      margin-top: 10px;
    }
  }
  .logo-upload {
    width: 120px;
    height: 120px;
    margin-top: 3px;
    /deep/ {
      .el-upload {
        width: 100%;
        height: 100%;
      }
      .echo-img {
        img {
          object-fit: cover;
          height: 100%;
        }
        .echo-img-tip {
          position: absolute;
          width: 100%;
          bottom: 0;
          background: #33333396;
          color: #fff !important;
          font-size: 12px;
          line-height: 26px;
          z-index: 10;
        }
      }
    }
  }
  /deep/.desc-input {
    .el-textarea__inner {
      height: 90px !important;
    }
  }
  .systemPrompt-tip {
    background-color: #f1f1f1;
    border-radius: 6px;
    line-height: 24px;
    color: #919eac;
    margin-top: 10px;
    padding: 8px 20px;
  }
  /*推荐问题*/
  .recommend-box {
    .recommend-item {
      margin-bottom: 12px;
      .recommend--input {
        width: calc(100% - 60px);
      }
      .close--icon {
        display: inline-block;
        width: 60px;
        line-height: 40px;
        text-align: center;
        cursor: pointer;
        color: #333;
        &:hover {
          font-weight: bold;
        }
      }
    }
  }

  /*知识增强*/
  .knowledge-config-com {
    margin-top: 10px;
  }

  /*action*/
  .api-box {
    padding-bottom: 60px;
  }

  /*Plugin*/
  .plugin-box {
    .el-checkbox-group {
      margin-top: 10px;
    }
    .plugin-checkbox /deep/.el-checkbox__inner.is-checked.el-checkbox__inner {
      background-color: #409eff;
      border-color: #409eff;
    }
  }

  /*footer*/
  .footer {
    position: absolute;
    height: 80px;
    padding: 20px 0;
    bottom: 0;
    left: 0;
    right: 0;
    text-align: center;
    border-top: 1px solid #d3d7dd;
  }
}
.drawer-test{
  width:50%;
  background:#F7F8FA;
  border-radius:6px;
  border-radius:8px;
  margin:10px 0;
  box-shadow: 0 1px 4px 0 rgba(0, 0, 0, 0.15);
}
}


.loading-progress {
  width: 100%;
  top: -4px;
  z-index: 1;
  position: fixed;
  left: -2px;
  right: -2px;
}
.action-list {
  margin: 10px 0 15px 0;
  width: 100%;
  .action-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    border: 1px solid #ddd;
    border-radius: 6px;
    margin-bottom: 5px;
    width: 100%;
    .name {
      width: 60%;
      box-sizing: border-box;
      padding: 10px 20px;
      cursor: pointer;
      color: #2c7eea;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    .bt {
      text-align: center;
      width: 40%;
      display: flex;
      justify-content: flex-end;
      padding-right: 10px;
      box-sizing: border-box;
      cursor: pointer;
      .del {
        color: $btn_bg;
        font-size: 16px;
      }
    }
  }
}
</style>
<style lang="scss">
.vue-treeselect .vue-treeselect__menu-container {
  z-index: 9999 !important;
}
.custom-tooltip.is-light {
  border-color: #ccc; /* Setting边框颜色 */
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

