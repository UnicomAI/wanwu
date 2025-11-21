<template>
  <div
    class="agent-from-content"
    :class="{ 'isDisabled': isPublish }"
  >
    <div class="form-header">
      <div class="header-left">
        <span
          class="el-icon-arrow-left btn"
          @click="goBack"
        ></span>
        <div class="basicInfo">
          <div class="img">
            <img :src="editForm.avatar.path ? `/user/api`+ editForm.avatar.path : '@/assets/imgs/bg-logo.png'" />
          </div>
          <div class="basicInfo-desc">
            <span class="basicInfo-title">{{(editForm.name || $t('agent.form.noInfo')).length > 12 ? (editForm.name || $t('agent.form.noInfo')).substring(0, 12) + '...' : (editForm.name || $t('agent.form.noInfo'))}}</span>
            <span
              class="el-icon-edit-outline editIcon"
              @click="editAgent"
            ></span>
            <LinkIcon type="agent" />
            <p>{{editForm.desc || $t('agent.form.noInfo')}}</p>
          </div>
        </div>
      </div>
      <div class="header-right">
        <el-button
          size="small"
          type="primary"
          style="padding:13px 12px;"
          @click="handlePublishSet"
        >
          <span class="el-icon-setting"></span>
          {{ $t('agent.form.publishConfig') }}
        </el-button>
        <el-button
          size="small"
          type="primary"
          @click="handlePublish"
          style="padding:13px 12px;"
        >{{ $t('agent.form.publish') }}<span
            class="el-icon-arrow-down"
            style="margin-left:5px;"
          ></span></el-button>
        <div
          class="popover-operation"
          v-if="showOperation"
        >
          <div>
            <el-radio
              :label="'private'"
              v-model="scope"
            >{{ $t('agent.form.publishType') }}</el-radio>
          </div>
          <div>
            <el-radio
              :label="'organization'"
              v-model="scope"
            >{{ $t('agent.form.publishType1') }}</el-radio>
          </div>
          <div>
            <el-radio
              :label="'public'"
              v-model="scope"
            >{{ $t('agent.form.publishType2') }}</el-radio>
          </div>
          <div class="saveBtn">
            <el-button
              size="mini"
              type="primary"
              @click="savePublish"
            >{{ $t('common.button.save') }}</el-button>
          </div>
        </div>
      </div>
    </div>
    <!-- AgentConfiguration -->
    <div class="agent_form">
      <div class="block prompt-box drawer-info">
        <div class="promptTitle">
          <h3>{{ $t('agent.form.systemPrompt') }}</h3>
          <div>
            <el-tooltip class="item" effect="dark" :content="$t('agent.form.submitToPrompt')" placement="top-start">
              <span class="el-icon-folder-add" @click="handleShowPrompt"></span>
            </el-tooltip>
            <el-tooltip class="item" effect="dark" :content="$t('tempSquare.promptOptimize')" placement="top-start">
              <span style="margin-left: 5px" class="el-icon-s-help" @click="showPromptOptimize"></span>
            </el-tooltip>
          </div>
        </div>
        <div class="rl" style="padding: 10px;">
          <el-input
            class="desc-input "
            v-model="editForm.instructions"
            :placeholder="$t('agent.form.promptTips')"
            type="textarea"
            show-word-limit
            :rows="12"
          ></el-input>
        </div>
        <promptTemplate ref="promptTemplate" />
      </div>
      <div class="drawer-form">
        <div class="agnetSet">
          <h3 class="labelTitle">{{ $t('agent.form.agentConfig') }}</h3>
          <div class="block prompt-box">
            <p class="block-title model-title">
              <span class="label">
                <img
                  :src="require('@/assets/imgs/require.png')"
                  class="required-label"
                />
                {{ $t('agent.form.modelSelect') }}
              </span>
              <span
                class="el-icon-s-operation operation"
                @click="showModelSet"
              ></span>
            </p>
            <div class="rl">
              <el-select
                v-model="editForm.modelParams"
                :placeholder="$t('agent.form.modelSearchPlaceholder')"
                @visible-change="visibleChange"
                :loading-text="$t('agent.toolDetail.modelLoadingText')"
                class="cover-input-icon model-select"
                :disabled="isPublish"
                :loading="modelLoading"
                filterable
                value-key="modelId"
                @change="handleModelChange($event)"
              >
                <el-option
                  class="model-option-item"
                  v-for="(item) in modleOptions"
                  :key="item.modelId"
                  :value="item.modelId"
                  :label="item.displayName"
                >
                  <div class="model-option-content">
                    <span class="model-name">{{ item.displayName }}</span>
                    <div
                      class="model-select-tags"
                      v-if="item.tags && item.tags.length > 0"
                    >
                      <span
                        v-for="(tag, tagIdx) in item.tags"
                        :key="tagIdx"
                        class="model-select-tag"
                      >{{ tag.text }}</span>
                    </div>
                  </div>
                </el-option>
              </el-select>
              <div
                class="model-select-tips"
                v-if="editForm.visionsupport === 'support'"
              >{{ $t('agent.form.visionModelTips') }}</div>
            </div>
          </div>
          <div class="block prompt-box">
            <p class="block-title">
              <img
                :src="require('@/assets/imgs/require.png')"
                class="required-label"
              />
              {{ $t('agent.form.prologue') }}
            </p>
            <div class="rl">
              <el-input
                class="desc-input"
                v-model="editForm.prologue"
                maxlength="100"
                :placeholder="$t('agent.form.prologuePlaceholder')"
                type="textarea"
              ></el-input>
              <span class="el-input__count">{{editForm.prologue.length}}/100</span>
            </div>
          </div>
          <div class="block recommend-box">
            <p class="block-title recommend-title">
              <span>{{ $t('agent.form.recommendQuestion') }}</span>
              <span
                @click="addRecommend"
                class="common-add"
              >
                <span class="el-icon-plus"></span>
                <span class="handleBtn">{{ $t('agent.add') }}</span>
              </span>
            </p>
            <div
              class="recommend-item"
              v-for="(n,i) in editForm.recommendQuestion"
              @mouseenter="activeIndex = i"
              @mouseleave="activeIndex = -1"
            >
              <el-input
                class="recommend--input"
                v-model.lazy="n.value"
                maxlength="50"
                :key="`${i}rml`"
              ></el-input>
              <span
                class="el-icon-delete recommend-del"
                @click="clearRecommend(n,i)"
                v-if="activeIndex === i"
              ></span>
            </div>
          </div>
        </div>
        <div class="common-box">
          <div class="block recommend-box">
            <p class="block-title tool-title">
              <span>{{ $t('agent.form.linkKnowledge') }}</span>
              <span>
                <span
                  class="common-add"
                  @click="showKnowledgeDiglog"
                >
                  <span class="el-icon-plus"></span>
                  <span class="handleBtn">{{ $t('agent.add') }}</span>
                </span>
                <span
                  class="common-add"
                  @click="showKnowledgeSet"
                >
                  <span class="el-icon-s-operation"></span>
                  <span class="handleBtn set">{{ $t('agent.form.config') }}</span>
                </span>
              </span>
            </p>
            <div class="rl tool-conent">
              <div class="tool-right tool">
                <div class="action-list">
                  <div
                    v-for="(n,i) in editForm.knowledgebases"
                    class="action-item"
                    :key="'knowledge'+ i"
                  >
                    <div
                      class="name"
                      style="color: #333"
                    >
                      <span>{{n.name || n.knowledgeName}}</span>
                    </div>
                    <div class="bt">
                      <el-tooltip
                        class="item"
                        effect="dark"
                        :content="$t('agent.form.metaDataFilter')"
                        placement="top-start"
                      >
                        <span
                          class="el-icon-setting del"
                          @click="showMetaSet(n,i)"
                          style="margin-right:10px;"
                        ></span>
                      </el-tooltip>
                      <span
                        class="el-icon-delete del"
                        @click="delKnowledge(i)"
                      ></span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="block recommend-box tool-box">
          <p class="block-title tool-title">
            <span>{{ $t('agent.form.tool') }}</span>
            <span
              @click="addTool"
              class="common-add"
            >
              <span class="el-icon-plus"></span>
              <span class="handleBtn">{{ $t('agent.add') }}</span>
            </span>
          </p>
          <div class="rl tool-conent">
            <div
              class="tool-right tool"
              v-show="allTools.length"
            >
              <div class="action-list">
                <div
                  class="action-item"
                  v-for="(n, i) in allTools"
                  :key="`${i}ac`"
                >
                  <div class="name">
                  <div class="toolImg">
                    <img :src="'/user/api/'+n.avatar.path" v-show="n.avatar && n.avatar.path" />
                  </div>
                  <el-tooltip class="item" effect="dark" :content="displayName(n)" placement="top-start">
                    <span>{{ displayName(n).length > 20 ? displayName(n).substring(0, 20) + '...' : displayName(n) }}</span>
                  </el-tooltip>
                  <el-tooltip class="item" effect="dark" :content="n.mcpName || n.toolName" placement="top-start">
                    <span class="el-icon-info desc-info" v-if="n.mcpName || n.toolName"></span>
                  </el-tooltip>
                  </div>
                  <div class="bt">
                    <span class="el-icon-s-operation bt-operation"  @click="handleBuiltin(n)" v-if="n.type === 'action' && n.toolType && n.toolType === 'builtin'"></span>
                    <el-switch
                      v-model="n.enable"
                      class="bt-switch"
                      @change="toolSwitch(n,n.type,n.enable)"
                    ></el-switch>
                    <span
                      @click="toolRemove(n,n.type)"
                      class="el-icon-delete del"
                    ></span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="block prompt-box link-box">
          <p class="block-title tool-title">
            <span>
              {{ $t('agent.form.safetyConfig') }}
              <el-tooltip
                class="item"
                effect="dark"
                :content="$t('agent.form.safetyConfigTips')"
                placement="top"
              >
                <span class="el-icon-question question-tips"></span>
              </el-tooltip>
            </span>
            <span class="common-add">
              <span @click="showSafety">
              <span class="el-icon-s-operation"></span>
              <span
                class="handleBtn"
                style="margin-right:10px;"
              >{{ $t('agent.form.config') }}</span> 
              </span>
              <el-switch
                v-model="editForm.safetyConfig.enable"
                :disabled="!(editForm.safetyConfig.tables || []).length"
              ></el-switch>
            </span>
          </p>
        </div>
        <div class="block prompt-box link-box" v-if="editForm.visionsupport === 'support'">
          <p class="block-title tool-title">
            <span>
              {{ $t('agent.form.vision') }}
              <el-tooltip
                class="item"
                effect="dark"
                :content="$t('agent.form.visionTips')"
                placement="top"
              >
                <span class="el-icon-question question-tips"></span>
              </el-tooltip>
            </span>
            <span class="common-add" @click="showVisualSet">
              <span class="el-icon-s-operation"></span>
              <span
                class="handleBtn"
                style="margin-right:10px;"
              >{{ $t('agent.form.config') }}</span>
            </span>
          </p>
        </div>
        <!-- Knowledge Graph开关 -->
        <div class="block prompt-box link-box" v-if="showGraphSwitch">
            <graphSwitch ref="graphSwitch" @graphSwitchchange="graphSwitchchange" :label="$t('knowledgeManage.create.knowledgeGraph')" :graphSwitch="editForm.knowledgeConfig.useGraph"/>
        </div>
      </div>
      <div class="drawer-test">
        <Chat
          :editForm="editForm"
          :chatType="'test'"
        />
      </div>
    </div>

    <!-- EditAgent -->
    <CreateIntelligent
      ref="createIntelligentDialog"
      :type="'edit'"
      :editForm="editForm"
      @updateInfo="getAppDetail"
    />
    <!-- ModelSetting -->
    <ModelSet
      @setModelSet="setModelSet"
      ref="modelSetDialog"
      :modelform="editForm.modelConfig"
      :limitMaxTokens="limitMaxTokens"
    />
    <!-- selectToolType -->
    <ToolDiaglog
      ref="toolDiaglog"
      @updateDetail="updateDetail"
      :assistantId="editForm.assistantId"
    />
    <!-- Sensitive WordSetting -->
    <setSafety
      ref="setSafety"
      @sendSafety="sendSafety"
    />
    <!-- Knowledge Base召回ParameterConfiguration -->
    <knowledgeSetDialog
      ref="knowledgeSetDialog"
      @setKnowledgeSet="setKnowledgeSet"
    />
    <!-- Knowledge Base Selection -->
    <knowledgeSelect
      ref="knowledgeSelect"
      @getKnowledgeData="getKnowledgeData"
    />
    <!-- 视图Setting -->
    <visualSet
      ref="visualSet"
      @sendVisual="sendVisual"
    />
    <!-- 内置ToolDetails -->
    <ToolDeatail ref="toolDeatail" @updateDetail="updateDetail" />
    <!-- SubmittoTip词 -->
    <createPrompt :isCustom="true" :type="promptType" ref="createPrompt" @reload="updatePrompt"/>
    <!-- Tip词优化 -->
    <PromptOptimize ref="promptOptimize" @promptSubmit="promptSubmit" />
    <!-- Metadata Settings -->
    <el-dialog
      :visible.sync="metaSetVisible"
      width="1050px"
      class="metaSetVisible"
      :before-close="handleMetaClose"
    >
      <template #title>
        <div class="metaHeader">
          <h3>{{ $t('agent.form.configMetaDataFilter') }}</h3>
          <span>{{ $t('agent.form.metaDataFilterDesc') }}</span>
        </div>
      </template>
      <metaSet
        ref="metaSet"
        :knowledgeId="currentKnowledgeId"
        :currentMetaData="currentMetaData"
      />
      <span
        slot="footer"
        class="dialog-footer"
      >
        <el-button @click="handleMetaClose">{{ $t('common.button.cancel') }}</el-button>
        <el-button
          type="primary"
          @click="submitMeta"
        >{{ $t('common.button.confirm') }}</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { appPublish } from "@/api/appspace";
import { store } from "@/store/index";
import { mapGetters,mapActions } from "vuex";
import CreateIntelligent from "@/components/createApp/createIntelligent";
import setSafety from "@/components/setSafety";
import visualSet from "./visualSet";
import metaSet from "@/components/metaSet";
import ModelSet from "./modelSetDialog";
import { selectModelList, getRerankList } from "@/api/modelAccess";
import {
  deleteMcp,
  enableMcp,
  getAgentDetail,
  delWorkFlowInfo,
  delActionInfo,
  putAgentInfo,
  enableWorkFlow,
  enableAction,
  delCustomBuiltIn,
  switchCustomBuiltIn
} from "@/api/agent";
import ToolDiaglog from "./toolDialog";
import ToolDeatail from "./toolDetail";
import knowledgeSetDialog from "./knowledgeSetDialog";
import { readWorkFlow } from "@/api/workflow";
import Chat from "./chat";
import LinkIcon from "@/components/linkIcon.vue";
import graphSwitch from "@/components/graphSwitch.vue"
import promptTemplate from "./prompt/index.vue";
import createPrompt from "@/components/createApp/createPrompt.vue"
import knowledgeSelect from "@/components/knowledgeSelect.vue";
import PromptOptimize from "@/components/promptOptimize.vue";
export default {
  components: {
    LinkIcon,
    Chat,
    CreateIntelligent,
    ModelSet,
    ToolDiaglog,
    setSafety,
    visualSet,
    knowledgeSetDialog,
    knowledgeSelect,
    metaSet,
    ToolDeatail,
    promptTemplate,
    createPrompt,
    graphSwitch,
    PromptOptimize
  },
  provide() {
    return {
      getPrompt: this.getPrompt
    }
  },
  watch: {
    editForm: {
      handler(newVal, oldVal) {
        // If data is set from details, do not trigger update logic
        if (this.isSettingFromDetail) return;

        if (this.debounceTimer) {
          clearTimeout(this.debounceTimer);
        }
        this.debounceTimer = setTimeout(() => {
          const props = [
            "modelParams",
            "modelConfig",
            "prologue",
            "knowledgebases",
            "instructions",
            "safetyConfig",
            "recommendQuestion",
            "visionConfig"
          ];

          const changed = props.some((prop) => {
            return (
              JSON.stringify(newVal[prop]) !==
              JSON.stringify((this.initialEditForm || {})[prop])
            );
          });

          if (changed) {
            if (newVal["modelParams"] !== "" && newVal["prologue"] !== "") {
              this.updateInfo();
            }
          }
        }, 500);
      },
      deep: true,
    },
  },
  computed: {
    ...mapGetters("app", ["cacheData"]),
    ...mapGetters("user", ["commonInfo"]),
  },
  data() {
    return {
      promptType: 'create',
      limitMaxTokens: 4096,
      knowledgeIndex: -1,
      currentKnowledgeId: "",
      currentMetaData: {},
      metaSetVisible: false,
      knowledgeCheckData: [],
      activeIndex: -1,
      showOperation: false,
      appId: "",
      scope: "public",
      rerankOptions: [],
      initialEditForm: null,
      editForm: {
        visionsupport: "",
        assistantId: "",
        avatar: {},
        name: "",
        desc: "",
        rerankParams: "",
        modelParams: "",
        prologue: "", //开场白
        instructions: "", //SystemTip词
        knowledgebases: [],
        visionConfig:{//视觉Configuration
          picNum: 3,
          maxPicNum:6
        },
        knowledgeConfig: {
          keywordPriority: 0.8, //Keyword weight
          matchType: "mix", //vector（Vector search）、text（Text search）、mix（Hybrid search: vector + text）
          priorityMatch: 1, //Weight matching, only set to 1
          rerankModelId: "", //rerankModelid
          semanticsPriority: 0.2, //Semantic weight
          topK: 5, //topK Get top N rows
          threshold: 0.4, //Filter score threshold
          maxHistory: 0, //Max context
          useGraph:false
        },
        recommendQuestion: [{ value: "" }],
        modelConfig: {
          temperature: 0.7,
          topP: 1,
          frequencyPenalty: 0,
          presencePenalty: 0,
          maxTokens: 512,
          maxTokensEnable: true,
          frequencyPenaltyEnable: true,
          temperatureEnable: true,
          topPEnable: true,
          presencePenaltyEnable: true,
        },
        safetyConfig: {
          enable: false,
          tables: [],
        },
      },
      apiURL: "",
      hasPluginPermission: false,
      modelLoading: false,
      wfDialogVisible: false,
      workFlowInfos: [],
      actionInfos: [],
      mcpInfos: [],
      allTools: [], //所Has of Tool
      workflowList: [],
      modelParams: {},
      platform: this.$platform,
      isPublish: false,
      modleOptions: [],
      selectKnowledge: [],
      knowledgeData: [],
      loadingPercent: 10,
      nameStatus: "",
      saved: false, //Button
      loading: false, //Button
      t: null,
      logoFileList: [],
      imageUrl: "",
      defaultLogo: require("@/assets/imgs/bg-logo.png"),
      debounceTimer: null, //Debounce timer
      isSettingFromDetail: false, // Prevent detail data trigger update flag
      nameMap: {
        workflow: {
          displayName: "Workflow",
          propName: "name",
        },
        mcp: {
          displayName: "MCPTool",
          propName: "actionName",
        },
        action: {
          displayName: "CustomTool",
          propName: "actionName",
        },
        // 可以ContinueAddOtherType
        default: {
          displayName: "not知Tool",
          propName: "name", // DefaultPropertyName
        },
      },
    };
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
      this.editForm.assistantId = this.$route.query.id;
      setTimeout(() => {
        this.getAppDetail();
      }, 500);
    }
    //Check if published
    if (this.$route.query.publish) {
      this.isPublish = true;
    }
    //CheckYesNoHasPlugin管理 of Permission
    const accessCert = localStorage.getItem("access_cert");
    const permission = accessCert
      ? JSON.parse(accessCert).user.permission.orgPermission
      : "";
    this.hasPluginPermission = permission.indexOf("plugin") !== -1;
  },
  beforeDestroy() {
    store.dispatch("app/initState");
    this.clearMaxPicNum();
  },
  methods: {
    ...mapActions("app", ["setMaxPicNum","clearMaxPicNum"]),
    graphSwitchchange(val){
      this.editForm.knowledgeConfig.useGraph = val;
    },
    updatePrompt(){
        this.$refs.promptTemplate.getPromptTemplateList()
    },
    handleShowPrompt(){
      this.$refs.createPrompt.openDialog({prompt: this.editForm.instructions});
    },
    showPromptOptimize() {
      if (!this.editForm.instructions) {
        this.$message.warning(this.$t('tempSquare.promptOptimizeHint'))
        return
      }
      this.$refs.promptOptimize.openDialog({prompt: this.editForm.instructions})
    },
    promptSubmit(prompt) {
      this.editForm.instructions = prompt
    },
    getPrompt(prompt){
      this.editForm.instructions = prompt;
    },
    handleBuiltin(n){
      this.$refs.toolDeatail.showDiaglog(n)
    },
    showVisualSet(){
      this.$refs.visualSet.showDialog(this.editForm.visionConfig);
    },
    sendVisual(data){
      this.editForm.visionConfig.picNum = data.picNum;
    },
    handleModelChange(val) {
      this.setModelInfo(val);
    },
    setModelInfo(val) {
      const selectedModel = this.modleOptions.find(
        (item) => item.modelId === val
      );
      if (selectedModel) {
        this.editForm.visionsupport = selectedModel.config.visionSupport;
        const maxTokens = selectedModel.config.maxTokens;
        this.limitMaxTokens = maxTokens && maxTokens > 0 ? maxTokens : 4096;
      }
    },
    submitMeta() {
      const metaData = this.$refs.metaSet.getMetaData();
      if (
        this.$refs.metaSet.validateRequiredFields(
          metaData["metaDataFilterParams"]["metaFilterParams"]
        )
      ) {
        this.$message.warning(this.$t('agent.form.incompleteInfo'));
        return;
      }
      this.$set(this.editForm.knowledgebases, this.knowledgeIndex, {
        ...this.editForm.knowledgebases[this.knowledgeIndex],
        ...metaData,
      });
      this.metaSetVisible = false;
    },
    delKnowledge(index) {
      this.editForm.knowledgebases.splice(index, 1);
    },
    getKnowledgeData(data) {
      const originalIds = new Set(this.editForm.knowledgebases.map(item => item.id));
      const newItems = data.filter(item => !originalIds.has(item.id));
      this.editForm.knowledgebases.push(...newItems);
    },
    handleMetaClose() {
      this.metaSetVisible = false;
    },
    showMetaSet(e, index) {
      this.currentKnowledgeId = e.id;
      this.currentMetaData = {};
      this.$nextTick(() => {
        this.currentMetaData = e.metaDataFilterParams;
      });
      this.knowledgeIndex = index;
      this.metaSetVisible = true;
    },
    showKnowledgeDiglog() {
      this.$refs.knowledgeSelect.showDialog(this.editForm.knowledgebases);
    },
    handlePublishSet() {
      this.$router.push({
        path: `/agent/publishSet`,
        query: {
          appId: this.editForm.assistantId,
          appType: "agent",
          name: this.editForm.name,
        },
      });
    },
    setKnowledgeSet(data) {
      this.editForm.knowledgeConfig = data;
    },
    displayName(item) {
      const config = this.nameMap[item.type] || this.nameMap["default"];
      return item[config.propName];
    },
    updateDetail() {
      this.getAppDetail();
    },
    showKnowledgeSet() {
      if (!this.editForm.knowledgebases.length) return;
      this.$refs.knowledgeSetDialog.showDialog(this.editForm.knowledgeConfig);
    },
    showSafety() {
      this.$refs.setSafety.showDialog(this.editForm.safetyConfig.tables);
    },
    sendSafety(data) {
      const tablesData = data.map(({ tableId, tableName }) => ({
        tableId,
        tableName,
      }));
      this.editForm.safetyConfig.tables = tablesData;
    },
    actionSwitch(id) {
      enableAction({ actionId: id }).then((res) => {
        if (res.code === 0) {
          this.getAppDetail();
        }
      });
    },
    toolSwitch(n, type, enable) {
      if (type === "workflow") {
        this.workflowSwitch(n.workFlowId, enable);
      } else if (type === "mcp") {
        this.mcpSwitch(n, enable);
      } else {
        this.customSwitch(n, enable);
      }
    },
    customSwitch(n, enable) {
      switchCustomBuiltIn({
        assistantId: this.editForm.assistantId,
        actionName:n.actionName,
        toolId:n.toolId,
        toolType:n.toolType,
        enable,
      })
        .then((res) => {
          if (res.code === 0) {
            this.getAppDetail();
          }
        })
        .catch(() => {});
    },
    mcpSwitch(n, enable) {
      enableMcp({ assistantId: this.editForm.assistantId,actionName:n.actionName, enable,mcpId:n.mcpId,mcpType:n.mcpType })
        .then((res) => {
          if (res.code === 0) {
            this.getAppDetail();
          }
        })
        .catch(() => {});
    },
    workflowSwitch(id, enable) {
      enableWorkFlow({
        assistantId: this.editForm.assistantId,
        workFlowId: id,
        enable,
      })
        .then((res) => {
          if (res.code === 0) {
            this.getAppDetail();
          }
        })
        .catch(() => {});
    },
    addTool() {
      const data = {
        mcpInfos: this.mcpInfos,
        workFlowInfos: this.workFlowInfos,
        customInfos: this.actionInfos,
      };
      this.$refs.toolDiaglog.showDialog(data);
    },
    rerankVisible(val) {
      if (val) {
        this.getRerankData();
      }
    },
    getRerankData() {
      getRerankList().then((res) => {
        if (res.code === 0) {
          this.rerankOptions = res.data.list || [];
        }
      });
    },
    goBack() {
      this.$router.go(-1);
    },
    handlePublish() {
      this.showOperation = !this.showOperation;
    },
    savePublish() {
      if (this.editForm.modelParams === "") {
        this.$message.warning(this.$t('agent.form.selectModel'));
        return false;
      }
      if (this.editForm.prologue === "") {
        this.$message.warning(this.$t('agent.form.inputPrologue'));
        return false;
      }
      const data = {
        appId: this.editForm.assistantId,
        appType: "agent",
        publishType: this.scope,
      };
      appPublish(data).then((res) => {
        if (res.code === 0) {
          this.$router.push({ path: "/explore" });
        }
      });
    },
    setModelSet(data) {
      this.editForm.modelConfig = data;
    },
    showModelSet() {
      this.$refs.modelSetDialog.showDialog();
    },
    editAgent() {
      this.$refs.createIntelligentDialog.openDialog();
    },
    async getWorkFlowDetail(n, index) {
      let params = {
        workflowID: n.appId,
      };
      let res = await readWorkFlow(params);
      if (res.code === 0) {
        this.doCreateWorkFlow(n.appId, res.data.base64OpenAPISchema, index);
      }
    },
    preAddWorkflow() {
      this.wfDialogVisible = true;
    },
    toolRemove(n, type) {
      if (type === "workflow") {
        this.doDeleteWorkflow(n.workFlowId);
      } else if (type === "mcp") {
        this.mcpRemove(n);
      } else {
        this.customRemove(n);
      }
    },
    customRemove(n) {
      delCustomBuiltIn({ assistantId: this.editForm.assistantId, toolId: n.toolId, toolType: n.toolType,actionName: n.actionName})
        .then((res) => {
          if (res.code === 0) {
            this.$message.success(this.$t('agent.form.deleteSuccess'));
            this.getAppDetail();
          }
        })
        .catch((err) => {});
    },
    mcpRemove(n) {
      deleteMcp({ assistantId: this.editForm.assistantId,actionName:n.actionName,mcpId:n.mcpId,mcpType:n.mcpType})
        .then((res) => {
          if (res.code === 0) {
            this.$message.success(this.$t('agent.form.deleteSuccess'));
            this.getAppDetail();
          }
        })
        .catch((err) => {});
    },
    visibleChange(val) {
      //下拉框Show of  when 候RequestModelList
      if (val) {
        this.getModelData();
      }
    },
    async getModelData() {
      this.modelLoading = true;
      const res = await selectModelList();
      if (res.code === 0) {
        this.modleOptions = res.data.list || [];
        this.modelLoading = false;
      }
      this.modelLoading = false;
    },
    async updateInfo() {
      //ModelData
      let modeInfo;
      if (
        typeof this.editForm.modelParams === "object" &&
        this.editForm.modelParams
      ) {
        modeInfo = this.editForm.modelParams;
      } else {
        modeInfo = this.modleOptions.find(
          (item) => item.modelId === this.editForm.modelParams
        );
      }
      const rerankInfo = this.rerankOptions.find(
        (item) => item.modelId === this.editForm.knowledgeConfig.rerankModelId
      );
      const recommendQuestion = this.editForm.recommendQuestion.map(
        (item) => item.value
      );
      const params = {
        assistantId: this.editForm.assistantId,
        prologue: this.editForm.prologue,
        recommendQuestion:
          recommendQuestion.length > 0 && recommendQuestion[0] !== ""
            ? recommendQuestion
            : [],
        instructions: this.editForm.instructions,
        knowledgeBaseConfig: {
          config: this.editForm.knowledgeConfig,
          knowledgebases: this.editForm.knowledgebases,
        },
        modelConfig: {
          config: this.editForm.modelConfig,
          displayName: modeInfo.displayName,
          model: modeInfo.model,
          modelId: modeInfo.modelId,
          modelType: modeInfo.modelType,
          provider: modeInfo.provider,
        },
        safetyConfig: this.editForm.safetyConfig,
        visionConfig: {picNum:this.editForm.visionConfig.picNum},
        rerankConfig: rerankInfo
          ? {
              displayName: rerankInfo.displayName,
              model: rerankInfo.model,
              modelId: rerankInfo.modelId,
              modelType: rerankInfo.modelType,
              provider: rerankInfo.provider,
            }
          : {},
      };
      let res = await putAgentInfo(params);
      if (res.code === 0) {
        this.getAppDetail();
      }
    },
    startLoading(val) {
      this.loadingPercent = val;
      if (val === 0) {
        this.loading = true;
      }
      if (val === 100) {
        setTimeout(() => {
          this.loading = false;
        }, 500);
      }
    },
    async getAppDetail() {
      this.startLoading(0);
      this.isSettingFromDetail = true;
      let res = await getAgentDetail({ assistantId: this.editForm.assistantId });
      if (res.code === 0) {
        this.startLoading(100);
        let data = res.data;
        this.editForm.knowledgeConfig =
          res.data.knowledgeBaseConfig.config.matchType === ""
            ? this.editForm.knowledgeConfig
            : res.data.knowledgeBaseConfig.config;
        this.editForm.knowledgeConfig.rerankModelId =
          res.data.rerankConfig.modelId;
        const knowledgeData = res.data.knowledgeBaseConfig.knowledgebases;
        if (knowledgeData && knowledgeData.length > 0) {
          this.editForm.knowledgebases = knowledgeData;
        }
        this.editForm = {
          ...this.editForm,
          avatar: data.avatar || {},
          prologue: data.prologue || "", //开场白
          name: data.name || "",
          desc: data.desc || "",
          instructions: data.instructions || "", //SystemTip词
          rerankParams: data.rerankConfig.modelId || "",
          visionConfig: data.visionConfig,//ImageConfiguration
          modelConfig:
            data.modelConfig.config !== null
              ? data.modelConfig.config
              : this.editForm.modelConfig,
          modelParams: data.modelConfig.modelId || "",
          recommendQuestion:
            data.recommendQuestion && data.recommendQuestion.length > 0
              ? data.recommendQuestion.map((n, index) => {
                  return {
                    value: n,
                  };
                })
              : [],
          safetyConfig:
            data.safetyConfig !== null
              ? data.safetyConfig
              : this.editForm.safetyConfig,
        };

        //SettingModelInformation
        this.setModelInfo(data.modelConfig.modelId);

        //回显CustomPlugin
        this.workFlowInfos = data.workFlowInfos || [];
        this.mcpInfos = data.mcpInfos || [];
        this.actionInfos = data.toolInfos || [];
        this.allTools = [
          ...this.workFlowInfos.map((item) => ({ ...item, type: "workflow" })),
          ...this.mcpInfos.map((item) => ({ ...item, type: "mcp" })),
          ...this.actionInfos.map((item) => ({ ...item, type: "action" })),
        ];

        this.setMaxPicNum(this.editForm.visionConfig.picNum);
        
        this.$nextTick(() => {
          this.isSettingFromDetail = false;
        });
      } else {
        this.isSettingFromDetail = false;
      }
    },
    async doDeleteWorkflow(workFlowId) {
      if (this.editForm.assistantId) {
        let res = await delWorkFlowInfo({
          workFlowId,
          assistantId: this.editForm.assistantId,
        });
        if (res.code === 0) {
          this.$message.success(this.$t("agent.delPluginTips"));
          this.getAppDetail();
        }
      } else {
        this.$message.error(this.$t("agent.otherTips"));
      }
    },
    //推荐问题
    addRecommend() {
      if (this.editForm.recommendQuestion.length > 3) {
        return;
      }
      this.editForm.recommendQuestion.push({ value: "" });
    },
    clearRecommend(n, index) {
      if (this.editForm.recommendQuestion.length === 1) return;
      this.editForm.recommendQuestion.splice(index, 1);
      this.activeIndex = -1;
    },
    async preDelAction(actionId) {
      this.$confirm(
        this.$t("createApp.delActionTips"),
        this.$t("knowledgeManage.tip"),
        {
          confirmButtonText: this.$t("createApp.save"),
          cancelButtonText: this.$t("createApp.cancel"),
          type: "warning",
        }
      )
        .then(async () => {
          let res = await delActionInfo({ actionId });
          if (res.code === 0) {
            this.$message.success(this.$t("createApp.delSuccess"));
            this.getAppDetail();
          }
        })
        .catch(() => {});
    },
  },
};
</script>

<style lang="scss" scoped>
.isDisabled .header-right,
.isDisabled .drawer-form > div {
  user-select: none;
  pointer-events: none !important;
}
/deep/ {
  .apikeyBtn {
    border: 1px solid $btn_bg;
    padding: 12px 10px;
    color: $btn_bg;
    display: flex;
    align-items: center;
    img {
      height: 14px;
    }
  }
  .metaSetVisible {
    .el-dialog__header {
      border-bottom: 1px solid #dbdbdb;
    }
    .el-dialog__body {
      max-height: 400px;
      overflow-y: auto;
    }
  }
}
.metaHeader {
  display: flex;
  justify-content: flex-start;
  h3 {
    font-size: 18px;
  }
  span {
    margin-left: 10px;
    color: #666;
    display: inline-block;
    padding-top: 5px;
  }
}
//通用AddButton
.common-add {
  color: #595959;
  cursor: pointer;
  margin-left: 10px;
  .handleBtn,
  .el-icon-plus {
    font-size: 13px !important;
    padding: 0 2px;
  }
  .set {
    margin-left: 1px;
  }
  .el-icon-plus {
    font-weight: bold;
  }
}
.model-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  .label {
    display: flex;
    align-items: center;
    font-size: 15px;
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
.basicInfo {
  display: flex;
  align-items: center;
  border-radius: 12px;
  padding: 16px;
  .img {
    margin-right: 10px;
    img {
      border-radius: 6px;
      width: 32px;
      height: 32px;
      object-fit: cover;
      box-shadow: 0 2px 4px 0 rgba(0, 0, 0, 0.1);
    }
  }
  .basicInfo-desc {
    flex: 1;
  }
  .basicInfo-title {
    display: inline-block;
    font-weight: 600;
    font-size: 14px;
    color: #1f2937;
  }
  .editIcon {
    font-size: 16px;
    margin-left: 5px;
    cursor: pointer;
    color: #6b7280;
  }
  p {
    color: #6b7280;
    font-size: 12px;
    margin: 0;
    line-height: 1.2;
  }
}
.form-header {
  width: 100%;
  height: 60px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  position: relative;
  border-bottom: 1px solid #dbdbdb;
  .popover-operation {
    position: absolute;
    bottom: -122px;
    right: 20px;
    background: #fff;
    box-shadow: 0px 1px 7px rgba(0, 0, 0, 0.3);
    padding: 10px 20px;
    border-radius: 6px;
    z-index: 999;
    .saveBtn {
      display: flex;
      justify-content: center;
      padding: 10px 0;
    }
  }
  .header-left {
    display: flex;
    align-items: center;
    .btn {
      margin-right: 10px;
      font-size: 18px;
      cursor: pointer;
    }
    .header-left-title {
      font-size: 18px;
      color: $color_title;
      font-weight: bold;
    }
  }
  .header-right {
    display: flex;
    align-items: center;
  }
}
.agent-from-content {
  height:100%;
  width: 100%;
  overflow: hidden!important;
}
.agent_form {
  padding: 0 10px;
  display: flex;
  justify-content: space-between;
  gap: 10px;
  height: calc(100% - 60px);
  .drawer-info {
    position:relative;
    width: 30%;
    margin: 10px 0;
    border-radius: 6px;
    background: #f7f8fa;
    box-shadow: 0 1px 4px 0 rgba(0, 0, 0, 0.15);
    // padding: 10px;
  }
  .labelTitle {
    font-size: 18px;
    font-weight: 800;
    padding: 10px 20px;
  }
  .promptTitle{
    display:flex;
    justify-content:space-between;
    padding:10px 10px 0 10px;
    h3{
      font-size: 18px;
      font-weight: 800;
    }
    span{
      font-size: 16px;
      color:$color;
      cursor:pointer;
      display:inline-block;
      padding:8px;
      border-radius:50%;
      background:#E0E7FF;
    }
  }
  .actionConfig {
    overflow-y: auto;
    width: 60%;
    padding: 0 40px;
  }
  .drawer-form {
    width: 30%;
    margin: 10px 0;
    position: relative;
    height: 100%;
    padding: 0 10px;
    border-radius: 6px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
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
    .link-box {
      background: #f7f8fa;
      box-shadow: 0 1px 4px 0 rgba(0, 0, 0, 0.15);
      border-radius: 8px;
      padding: 10px 20px;
    }
    .common-box {
      background: #f7f8fa;
      box-shadow: 0 1px 4px 0 rgba(0, 0, 0, 0.15);
      border-radius: 8px;
      padding: 5px 20px;
      margin-bottom: 15px;
      .block {
        margin-bottom: 10px;
      }
    }
    .tool-box {
      background: #f7f8fa;
      box-shadow: 0 1px 4px 0 rgba(0, 0, 0, 0.15);
      border-radius: 8px;
      padding: 10px 20px;
    }

    .agnetSet {
      background: #f7f8fa;
      box-shadow: 0 1px 4px 0 rgba(0, 0, 0, 0.15);
      border-radius: 8px;
      margin-bottom: 15px;
      .block {
        padding: 5px 20px;
        margin-bottom: 0px !important;
      }
    }
    /* General */
    .block {
      margin-bottom: 15px;
      .tool-title {
        display: flex;
        justify-content: space-between;
        span {
          font-size: 15px;
        }
      }
      .block-title {
        line-height: 30px;
        font-size: 15px;
        font-weight: bold;
        display: flex;
        align-items: center;
        .title_tips {
          color: #999;
          margin-left: 20px;
          font-weight: normal;
        }
        .question-tips {
          margin-left: 5px;
        }
      }
      .block-link {
        border: 1px solid #ddd;
        padding: 6px 10px;
        border-radius: 6px;
        display: flex;
        justify-content: space-between;
        align-items: center;
        .link-text {
          color: $color;
          display: flex;
          align-items: center;
        }
        .link-operation {
          cursor: pointer;
          margin-right: 5px;
          font-size: 16px;
          line-height:20px;
        }
      }
      .tool-conent {
        display: flex;
        justify-content: space-between;
        gap: 10px;
        .tool {
          width: 100%;
          max-height: 300px;
          overflow-y: auto;
          .action-list {
            width: 100%;
            // display: grid;
            // grid-template-columns: repeat(2, minmax(0, 1fr));
            // gap: 10px;
          }
        }
      }
      .model-select {
        width: 100%;
      }
      .model-select-tips {
        margin-top: 10px;
        color: #dc6803;
      }
      .operation {
        text-align: center;
        cursor: pointer;
        font-size: 16px;
        padding-right: 10px;
      }
      .operation:hover {
        color: $color;
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
      .required-label {
        width: 18px;
        height: 18px;
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
    /* Create App */
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
    /* Recommended Questions */
    .recommend-box {
      .recommend-title {
        display: flex;
        justify-content: space-between;
        span {
          font-size: 15px;
        }
      }
      .recommend-item {
        margin-bottom: 12px;
        display: flex;
        justify-content: space-between;
        position: relative;
        .recommend--input {
          width: calc(100% - 30px);
          margin-right: 30px;
        }
        .recommend-del {
          position: absolute;
          right: 10px;
          top: 10px;
          color: #595959;
          cursor: pointer;
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

    /* Knowledge Enhancement */
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
  .drawer-test {
    width: calc((100% - 320px - 20px) / 2);
    background: #f7f8fa;
    border-radius: 8px;
    margin: 10px 0;
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
      width: 80%;
      box-sizing: border-box;
      padding: 10px;
      cursor: pointer;
      display:flex;
      align-items:center;
      color: #333;
      .desc-info{
        color:#ccc;
        margin-left:4px;
      }
      .toolImg{
        width:30px;
        height:30px;
        border-radius:50%;
        background:#eee;
        margin-right:5px;
        img{
          width:100%;
          height:100%;
          border-radius:50%;
          object-fit: cover;
        }
      }

    }
    .bt {
      text-align: center;
      width: 30%;
      display: flex;
      justify-content: flex-end;
      align-items:center;
      padding-right: 10px;
      box-sizing: border-box;
      cursor: pointer;
      .del {
        color: $btn_bg;
        font-size: 16px;
        line-height:20px;
      }
      .bt-switch {
        margin: 0 6px 0 6px;
      }
      .bt-operation{
        font-size:16px;
        line-height:20px;
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
  border-color: #ccc; /* Set border color */
  background-color: #fff; /* Set background color */
  color: #666; /* Set text color */
}
.custom-tooltip.el-tooltip__popper[x-placement^="top"] .popper__arrow::after {
  border-top-color: #fff !important;
}
.custom-tooltip.el-tooltip__popper.is-light[x-placement^="top"] .popper__arrow {
  border-top-color: #ccc !important;
}
.drawer-test .echo .session-item{
  width:30vw!important;
}
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
</style>

