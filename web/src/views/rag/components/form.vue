<template>
  <div
    class="agent-from-content page-wrapper"
    :class="{ 'disable-clicks': disableClick }"
  >
    <div class="form-header">
      <div class="header-left">
        <span class="el-icon-arrow-left btn" @click="goBack"></span>
        <div class="basicInfo">
          <div class="img">
            <img
              :src="
                editForm.avatar.path
                  ? avatarSrc(editForm.avatar.path)
                  : require('@/assets/imgs/bg-logo.png')
              "
            />
          </div>
          <div class="basicInfo-desc">
            <span class="basicInfo-title">{{ editForm.name || '' }}</span>
            <span
              class="el-icon-edit-outline editIcon"
              @click="editAgent"
            ></span>
            <LinkIcon type="rag" />
            <el-tooltip
              effect="dark"
              :content="editForm.desc || ''"
              placement="top"
            >
              <p class="desc-text">
                {{ editForm.desc || '' }}
              </p>
            </el-tooltip>
            <p>
              uuid: {{ this.editForm.appId }}
              <copyIcon :text="this.editForm.appId" type="icon" />
            </p>
          </div>
        </div>
      </div>
      <div class="header-right">
        <div class="header-api" v-if="publishType">
          <el-tag effect="plain" class="root-url">
            {{ $t('rag.form.apiRootUrl') }}
          </el-tag>
          {{ apiURL }}
        </div>
        <el-button
          v-if="publishType"
          @click="$router.push('/openApiKey')"
          plain
          class="apikeyBtn"
          size="small"
        >
          <img :src="require('@/assets/imgs/apikey.png')" />
          {{ $t('rag.form.apiKey') }}
        </el-button>
        <VersionPopover
          ref="versionPopover"
          v-if="publishType"
          :appId="editForm.appId"
          :appType="RAG"
          @reloadData="reloadData"
          @previewVersion="previewVersion"
        />
        <el-button
          v-if="publishType"
          size="small"
          type="primary"
          style="padding: 13px 12px"
          @click="handlePublishSet"
        >
          <span class="el-icon-setting"></span>
          {{ $t('agent.form.publishConfig') }}
        </el-button>
        <el-popover
          placement="bottom-end"
          trigger="click"
          style="margin-left: 13px"
        >
          <el-button
            slot="reference"
            size="small"
            type="primary"
            style="padding: 13px 12px"
          >
            {{ $t('common.button.publish') }}
            <span class="el-icon-arrow-down" style="margin-left: 5px"></span>
          </el-button>
          <el-form ref="publishForm" :model="publishForm" :rules="publishRules">
            <el-form-item :label="$t('list.version.no')" prop="version">
              <el-input
                v-model="publishForm.version"
                :placeholder="$t('list.version.noPlaceholder')"
              ></el-input>
            </el-form-item>
            <el-form-item :label="$t('list.version.desc')" prop="desc">
              <el-input
                v-model="publishForm.desc"
                :placeholder="$t('list.version.descPlaceholder')"
              ></el-input>
            </el-form-item>
            <el-form-item
              :label="$t('list.version.publishType')"
              prop="publishType"
            >
              <el-radio-group v-model="publishForm.publishType">
                <div>
                  <el-radio label="private">
                    {{ $t('agent.form.publishType') }}
                  </el-radio>
                </div>
                <div>
                  <el-radio label="organization">
                    {{ $t('agent.form.publishType1') }}
                  </el-radio>
                </div>
                <div>
                  <el-radio label="public">
                    {{ $t('agent.form.publishType2') }}
                  </el-radio>
                </div>
              </el-radio-group>
            </el-form-item>
            <div class="saveBtn">
              <el-button size="mini" type="primary" @click="savePublish">
                {{ $t('common.button.save') }}
              </el-button>
            </div>
          </el-form>
        </el-popover>
      </div>
    </div>
    <div class="agent_form">
      <div class="drawer-form">
        <div class="block">
          <div class="prompt-box">
            <p class="block-title common-set">
              <span class="common-set-label">
                <img
                  :src="require('@/assets/imgs/require.png')"
                  class="required-label"
                />
                {{ $t('agent.form.modelSelect') }}
              </span>
              <span class="common-add" @click="showModelSet">
                <el-tooltip
                  class="item"
                  effect="dark"
                  :content="$t('agent.form.modelSelectConfigTips')"
                  placement="top-start"
                >
                  <span class="el-icon-s-operation operation">
                    <span class="handleBtn">{{ $t('agent.form.config') }}</span>
                  </span>
                </el-tooltip>
              </span>
            </p>
            <div class="rl">
              <modelSelect
                v-model="editForm.modelParams"
                :options="modelOptions"
                :placeholder="
                  $t('knowledgeManage.create.modelSearchPlaceholder')
                "
                @visible-change="visibleChange"
                @change="handleModelChange($event)"
                :loading-text="$t('knowledgeManage.create.modelLoading')"
                class="cover-input-icon model-select"
                :loading="modelLoading"
                filterable
                warning
              />
            </div>
          </div>
          <!-- 推荐问题 -->
          <div class="recommend-box">
            <p class="block-title recommend-title">
              <span>{{ $t('agent.form.recommendQuestion') }}</span>
              <span class="common-add" @click="addRecommend">
                <span class="el-icon-plus"></span>
                <span class="handleBtn">{{ $t('agent.add') }}</span>
              </span>
            </p>
            <div
              v-for="(n, i) in editForm.recommendQuestion"
              :key="`${i}rml`"
              class="recommend-item"
            >
              <el-input
                :key="`${i}rml`"
                v-model.lazy="n.value"
                class="recommend--input"
                maxlength="50"
              >
                <span
                  slot="suffix"
                  class="el-icon-delete recommend-del"
                  @click="clearRecommend(n, i)"
                ></span>
              </el-input>
            </div>
          </div>
        </div>
        <!-- 问答库配置 -->
        <div class="block">
          <knowledgeDataField
            :knowledgeConfig="editForm.qaKnowledgeBaseConfig"
            :category="QA"
            @getSelectKnowledge="getSelectKnowledge"
            @knowledgeDelete="knowledgeDelete"
            @knowledgeRecallSet="knowledgeRecallSet"
            @updateMetaData="updateMetaData"
            :labelText="$t('app.linkQaDatabase')"
            :type="'qaKnowledgeBaseConfig'"
          />
        </div>
        <!-- 知识库库配置 -->
        <div class="block">
          <knowledgeDataField
            :knowledgeConfig="editForm.knowledgeBaseConfig"
            :category="KNOWLEDGE"
            :knowledgeCategory="getCategory"
            @getSelectKnowledge="getSelectKnowledge"
            @knowledgeDelete="knowledgeDelete"
            @knowledgeRecallSet="knowledgeRecallSet"
            @updateMetaData="updateMetaData"
            :labelText="$t('agent.form.linkKnowledge')"
            :type="'knowledgeBaseConfig'"
          />
        </div>
        <div class="block">
          <p class="block-title common-set">
            <span class="common-set-label">
              {{ $t('agent.form.safetyConfig') }}
              <el-tooltip
                class="item"
                effect="dark"
                :content="$t('agent.form.safetyConfigTips1')"
                placement="top"
              >
                <span class="el-icon-question question-tips"></span>
              </el-tooltip>
            </span>
            <span class="common-add">
              <el-tooltip
                class="item"
                effect="dark"
                :content="$t('agent.form.safetyConfigTips')"
                placement="top-start"
              >
                <span class="el-icon-s-operation operation" @click="showSafety">
                  <span class="handleBtn">{{ $t('agent.form.config') }}</span>
                </span>
              </el-tooltip>
              <el-switch
                v-model="editForm.safetyConfig.enable"
                :disabled="!(editForm.safetyConfig.tables || []).length"
              ></el-switch>
            </span>
          </p>
        </div>
        <div
          class="block"
          v-if="
            editForm.visionsupport === 'support' &&
            getCategory !== KNOWLEDGE &&
            visionsupportRerank
          "
        >
          <p class="block-title common-set">
            <span class="common-set-label">
              {{ $t('agent.form.vision') }}
              <el-tooltip
                class="item"
                effect="dark"
                :content="$t('agent.form.visionTips1')"
                placement="top"
              >
                <span class="el-icon-question question-tips"></span>
              </el-tooltip>
            </span>
            <!--            <span class="common-add" @click="showVisualSet">-->
            <!--              <el-tooltip-->
            <!--                class="item"-->
            <!--                effect="dark"-->
            <!--                :content="$t('agent.form.visionTips')"-->
            <!--                placement="top-start"-->
            <!--              >-->
            <!--                <span class="el-icon-s-operation operation">-->
            <!--                  <span class="handleBtn">{{ $t('agent.form.config') }}</span>-->
            <!--                </span>-->
            <!--              </el-tooltip>-->
            <!--            </span>-->
            <el-switch
              :value="editForm.visionConfig.picNum === 1"
              @input="
                val => {
                  editForm.visionConfig.picNum = val ? 1 : 0;
                }
              "
            ></el-switch>
          </p>
        </div>
      </div>
      <div class="drawer-test block">
        <Chat
          ref="ragDraftChat"
          :editForm="editForm"
          :chatType="'test'"
          :disableClick="disableClick"
          :maxImageSize="maxImageSize"
          :maxPicNum="editForm.visionConfig.picNum"
        />
      </div>
    </div>
    <!-- 编辑智能体 -->
    <CreateTxtQues
      ref="createTxtQues"
      :type="'edit'"
      :editForm="editForm"
      @updateInfo="getDetail"
    />
    <!-- 模型设置 -->
    <ModelSet
      @setModelSet="setModelSet"
      ref="modelSetDialog"
      :modelConfig="editForm.modelConfig"
    />
    <!-- 视图设置 -->
    <visualSet ref="visualSet" @sendVisual="sendVisual" />

    <setSafety ref="setSafety" @sendSafety="sendSafety" />
  </div>
</template>

<script>
import { appPublish, getApiKeyRoot } from '@/api/appspace';
import CreateTxtQues from '@/components/createApp/createRag.vue';
import ModelSet from './modelSetDialog.vue';
import metaSet from '@/components/metaSet';
import setSafety from '@/components/setSafety';
import VersionPopover from '@/components/versionPopover';
import {
  getModelDetail,
  getRerankList,
  selectModelList,
} from '@/api/modelAccess';
import { getRagInfo, getRagPublishedInfo, updateRagConfig } from '@/api/rag';
import Chat from './chat';
import chiChat from '@/components/app/chiChat.vue';
import LinkIcon from '@/components/linkIcon.vue';
import knowledgeSelect from '@/components/knowledgeSelect.vue';
import knowledgeDataField from '@/components/app/knowledgeDataField.vue';
import { RAG } from '@/utils/commonSet';
import {
  EXTERNAL,
  KNOWLEDGE,
  QA,
  MULTIMODAL,
  MIX_MULTIMODAL,
} from '@/views/knowledge/constants';
import CopyIcon from '@/components/copyIcon.vue';
import { avatarSrc } from '@/utils/util';
import visualSet from '@/views/agent/components/visualSet.vue';
import modelSelect from '@/components/modelSelect.vue';
export default {
  name: 'RagTest',
  components: {
    modelSelect,
    visualSet,
    CopyIcon,
    LinkIcon,
    Chat,
    CreateTxtQues,
    ModelSet,
    setSafety,
    VersionPopover,
    knowledgeSelect,
    metaSet,
    chiChat,
    knowledgeDataField,
  },
  data() {
    return {
      RAG,
      QA,
      KNOWLEDGE,
      disableClick: false,
      version: '',
      rerankOptions: [],
      localKnowledgeConfig: {},
      publishType: '', // 为空表示未发布，private表示私密，organization表示组织内可见，public表示公开
      publishForm: {
        publishType: 'private',
        version: '',
        desc: '',
      },
      publishRules: {
        version: [
          {
            required: true,
            message: this.$t('list.version.noMsg'),
            trigger: 'blur',
          },
          {
            pattern: /^v\d+\.\d+\.\d+$/,
            message: this.$t('list.version.versionMsg'),
            trigger: 'blur',
          },
        ],
        desc: [
          {
            required: true,
            message: this.$t('list.version.descPlaceholder'),
            trigger: 'blur',
          },
        ],
        publishType: [
          {
            required: true,
            message: this.$t('common.select.placeholder'),
            trigger: 'change',
          },
        ],
      },
      visionsupportRerank: '',
      editForm: {
        appId: '',
        avatar: {},
        name: '',
        desc: '',
        modelParams: '',
        visionsupport: '',
        visionConfig: {
          //视觉配置
          picNum: 0, //为0时不启用图搜功能
        },
        modelConfig: {
          temperature: 0.14,
          topP: 0.85,
          frequencyPenalty: 1.1,
          temperatureEnable: true,
          topPEnable: true,
          frequencyPenaltyEnable: true,
        },
        knowledgeBaseConfig: {
          config: {
            keywordPriority: 0.8, //关键词权重
            matchType: 'mix', //vector（向量检索）、text（文本检索）、mix（混合检索：向量+文本）
            priorityMatch: 1, //权重匹配，只有在混合检索模式下，选择权重设置后，这个才设置为1
            rerankModelId: '', //rerank模型id
            semanticsPriority: 0.2, //语义权重
            topK: 5, //topK 获取最高的几行
            threshold: 0.4, //过滤分数阈值
            maxHistory: 0, //
            useGraph: false,
            chiChat: false,
          },
          knowledgebases: [],
        },
        knowledgeConfig: {},
        qaKnowledgeBaseConfig: {
          knowledgebases: [],
          config: {
            keywordPriority: 0.8, //关键词权重
            matchType: 'mix', //vector（向量检索）、text（文本检索）、mix（混合检索：向量+文本）
            priorityMatch: 1, //权重匹配，只有在混合检索模式下，选择权重设置后，这个才设置为1
            rerankModelId: '', //rerank模型id
            semanticsPriority: 0.2, //语义权重
            topK: 5, //topK 获取最高的几行
            threshold: 0.4, //过滤分数阈值
            maxHistory: 0, //
            useGraph: false,
            chiChat: false,
          },
        },
        safetyConfig: {
          enable: false,
          tables: [],
        },
        recommendQuestion: [
          {
            value: '',
          },
        ],
      },
      initialEditForm: null,
      apiURL: '',
      modelLoading: false,
      modelOptions: [],
      debounceTimer: null, //防抖计时器
      isUpdating: false, // 防止重复更新标记
      isSettingFromDetail: false, // 防止详情数据触发更新标记
    };
  },
  watch: {
    editForm: {
      handler(newVal) {
        // 如果是从详情设置的数据，不触发更新逻辑
        if (this.isSettingFromDetail) {
          return;
        }

        if (this.debounceTimer) {
          clearTimeout(this.debounceTimer);
        }
        this.debounceTimer = setTimeout(() => {
          const props = [
            'modelParams',
            'modelConfig',
            'knowledgeBaseConfig',
            'safetyConfig',
            'qaKnowledgeBaseConfig',
            'visionConfig',
            'recommendQuestion',
          ];
          const changed = props.some(prop => {
            return (
              JSON.stringify(newVal[prop]) !==
              JSON.stringify((this.initialEditForm || {})[prop])
            );
          });
          if (changed && !this.isUpdating) {
            this.updateInfo();
          }
        }, 500);
      },
      deep: true,
    },
    'editForm.knowledgeBaseConfig.config.rerankModelId': {
      handler(newVal) {
        if (newVal)
          getModelDetail({
            modelId: this.editForm.knowledgeBaseConfig.config.rerankModelId,
          }).then(res => {
            if (res.code === 0) {
              this.visionsupportRerank =
                res.data.modelType === 'multimodal-rerank';
            }
          });
        else this.visionsupportRerank = false;
      },
      deep: true,
      immediate: true,
    },
  },
  computed: {
    getCategory() {
      const knowledgebases = this.editForm.knowledgeBaseConfig.knowledgebases;
      if (!knowledgebases || knowledgebases.length === 0) {
        return KNOWLEDGE;
      }

      const categories = knowledgebases.map(kb => kb.category);
      const hasKnowledge = categories.includes(KNOWLEDGE);
      const hasMultiModal = categories.includes(MULTIMODAL);

      if (hasKnowledge && hasMultiModal) {
        return MIX_MULTIMODAL;
      } else if (hasMultiModal) {
        return MULTIMODAL;
      } else {
        return KNOWLEDGE;
      }
    },
    maxImageSize() {
      const selected = this.modelOptions.find(
        item => item.modelId === this.editForm.modelParams,
      );
      const size = Number(selected?.config?.maxImageSize);
      return size > 0 ? size : null;
    },
  },
  mounted() {
    this.initialEditForm = structuredClone(this.editForm);
  },
  created() {
    this.getModelData(); //获取模型列表
    this.getRerankData(); //获取rerank模型
    if (this.$route.query.id) {
      this.editForm.appId = this.$route.query.id;
      setTimeout(() => {
        this.getDetail(); //获取详情
        this.apiKeyRootUrl(); //获取api根地址
      }, 500);
    }
  },

  methods: {
    avatarSrc,
    reloadData() {
      this.disableClick = false;
      this.getDetail();
    },
    previewVersion(item) {
      this.disableClick = !item.isCurrent;
      this.version = item.version || '';
      this.getDetail();
    },
    //获取知识库或问答库选中数据
    getSelectKnowledge(data, type) {
      this.editForm[type]['knowledgebases'] = data;
    },
    //删除知识库或问答库
    knowledgeDelete(index, type) {
      this.editForm[type]['knowledgebases'].splice(index, 1);
    },
    //设置知识库或问答库召回参数
    knowledgeRecallSet(data, type) {
      if (data) {
        this.editForm[type]['config'] = data;
      }
    },
    chiSwitchChange(value) {
      this.$set(this.editForm.knowledgeBaseConfig.config, 'chiChat', value);
    },
    //更新知识库元数据
    updateMetaData(data, index, type) {
      this.$set(this.editForm[type]['knowledgebases'], index, {
        ...this.editForm[type]['knowledgebases'][index],
        ...data,
      });
    },
    sendSafety(data) {
      const tablesData = data.map(({ tableId, tableName }) => ({
        tableId,
        tableName,
      }));
      this.editForm.safetyConfig.tables = tablesData;
    },
    showSafety() {
      this.$refs.setSafety.showDialog(this.editForm.safetyConfig.tables);
    },
    goBack() {
      this.$router.go(-1);
    },
    getDetail() {
      //获取详情
      this.isSettingFromDetail = true; // 设置标志位，防止触发更新逻辑
      let res;
      if (this.version) {
        res = getRagPublishedInfo({
          ragId: this.editForm.appId,
          version: this.version,
        });
      } else
        res = getRagInfo({
          ragId: this.editForm.appId,
        });
      res
        .then(res => {
          if (res.code === 0) {
            this.publishType = res.data.appPublishConfig.publishType;
            this.editForm.avatar = res.data.avatar;
            this.editForm.name = res.data.name;
            this.editForm.desc = res.data.desc;
            this.editForm.visionConfig = res.data.visionConfig;
            // 临时隐藏visionConfig的配置，全都设置为1
            // this.editForm.visionConfig.picNum = 1;

            this.setModelInfo(res.data.modelConfig.modelId);

            if (res.data.qaKnowledgeBaseConfig) {
              this.editForm.qaKnowledgeBaseConfig.knowledgebases =
                res.data.qaKnowledgeBaseConfig.knowledgebases;
              if (res.data.qaKnowledgeBaseConfig.config !== null) {
                this.editForm.qaKnowledgeBaseConfig.config =
                  res.data.qaKnowledgeBaseConfig.config;
              }
            }

            if (res.data.knowledgeBaseConfig) {
              this.editForm.knowledgeBaseConfig.knowledgebases =
                res.data.knowledgeBaseConfig.knowledgebases;
              if (res.data.knowledgeBaseConfig.config !== null) {
                this.editForm.knowledgeBaseConfig.config =
                  res.data.knowledgeBaseConfig.config;
              }
            }

            if (res.data.safetyConfig) {
              this.editForm.safetyConfig = res.data.safetyConfig;
            }

            // 处理推荐问题
            this.editForm.recommendQuestion =
              res.data.recommendQuestion &&
              res.data.recommendQuestion.length > 0
                ? res.data.recommendQuestion.map((n, index) => {
                    return {
                      value: n,
                    };
                  })
                : [
                    {
                      value: '',
                    },
                  ];

            if (res.data.modelConfig.config !== null) {
              this.editForm.modelConfig = res.data.modelConfig.config;
            }

            this.editForm.knowledgeBaseConfig.config.rerankModelId =
              res.data.rerankConfig.modelId;
            this.editForm.qaKnowledgeBaseConfig.config.rerankModelId =
              res.data.qaRerankConfig.modelId;

            this.$nextTick(() => {
              // 更新基准数据，避免 watch 误判
              this.initialEditForm = structuredClone(this.editForm);
              this.isSettingFromDetail = false;
              this.$refs.ragDraftChat?.loadDraftConversationHistory();
            });
          } else {
            this.isSettingFromDetail = false;
          }
        })
        .catch(() => {
          this.isSettingFromDetail = false;
        });
    },
    showVisualSet() {
      this.$refs.visualSet.showDialog(this.editForm.visionConfig);
    },
    sendVisual(data) {
      this.editForm.visionConfig.picNum = data.picNum;
    },
    handleModelChange(val) {
      if (!val) return;
      const selectedModel = this.modelOptions.find(
        item => item.modelId === val,
      );
      if (selectedModel?.config?.thinkingSupport === 'support') {
        this.$set(this.editForm.modelConfig, 'thinkingEnable', true);
      } else {
        this.$delete(this.editForm.modelConfig, 'thinkingEnable');
      }
      this.setModelInfo(val);
    },
    setModelInfo(val) {
      if (!val) return;
      const selectedModel = this.modelOptions.find(
        item => item.modelId === val,
      );
      if (selectedModel) {
        this.editForm.modelParams = val;
        this.editForm.visionsupport = selectedModel.config.visionSupport;
      } else {
        this.editForm.modelParams = '';
        if (val) this.$message.warning(this.$t('agent.form.modelNotSupport'));
      }
    },
    getRerankData() {
      getRerankList().then(res => {
        if (res.code === 0) {
          this.rerankOptions = res.data.list || [];
        }
      });
    },
    handlePublishSet() {
      this.$router.push({
        path: `/rag/publishSet`,
        query: {
          appId: this.editForm.appId,
          appType: RAG,
          name: this.editForm.name,
        },
      });
    },
    savePublish() {
      const { matchType, priorityMatch, rerankModelId } =
        this.editForm.qaKnowledgeBaseConfig.config;
      const isMixPriorityMatch = matchType === 'mix' && priorityMatch;

      if (this.editForm.modelParams === '') {
        this.$message.warning(this.$t('agent.form.selectModel'));
        return false;
      }
      if (
        this.editForm.knowledgeBaseConfig.knowledgebases.length === 0 &&
        this.editForm.qaKnowledgeBaseConfig.knowledgebases.length === 0
      ) {
        this.$message.warning(this.$t('app.selectKnowledge'));
        return false;
      }
      if (
        !this.editForm.knowledgeBaseConfig.knowledgebases.length &&
        this.editForm.qaKnowledgeBaseConfig.knowledgebases.length > 0
      ) {
        if (!isMixPriorityMatch && !rerankModelId) {
          this.$message.warning(this.$t('app.selectRerank'));
          return false;
        }
      }

      this.$refs.publishForm.validate(valid => {
        if (valid) {
          const data = {
            appId: this.editForm.appId,
            appType: RAG,
            publishType: this.publishForm.publishType,
            desc: this.publishForm.desc,
            version: this.publishForm.version,
          };
          appPublish(data).then(res => {
            if (res.code === 0) {
              this.$router.push({ path: '/explore', query: { type: 'rag' } });
            }
          });
        }
      });
    },
    apiKeyRootUrl() {
      const data = { appId: this.editForm.appId, appType: 'rag' };
      getApiKeyRoot(data).then(res => {
        if (res.code === 0) {
          this.apiURL = res.data || '';
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
      this.$refs.createTxtQues.openDialog();
    },
    //推荐问题
    addRecommend() {
      if (this.editForm.recommendQuestion.length > 3) {
        return;
      }
      this.editForm.recommendQuestion.push({
        value: '',
      });
    },
    clearRecommend(n, index) {
      if (this.editForm.recommendQuestion.length === 1) return;
      this.editForm.recommendQuestion.splice(index, 1);
    },
    visibleChange(val) {
      if (val) {
        this.getModelData();
      }
    },
    async getModelData() {
      this.modelLoading = true;
      const res = await selectModelList();
      if (res.code === 0) {
        this.modelOptions = res.data.list || [];

        this.modelLoading = false;
      }
      this.modelLoading = false;
    },
    async updateInfo() {
      if (this.isUpdating) return; // 防止重复调用

      this.isUpdating = true;
      try {
        //模型数据
        const modeInfo = this.modelOptions.find(
          item => item.modelId === this.editForm.modelParams,
        );
        if (
          this.editForm.knowledgeBaseConfig.config.matchType === 'mix' &&
          this.editForm.knowledgeBaseConfig.config.priorityMatch === 1
        ) {
          this.editForm.knowledgeBaseConfig.config.rerankModelId = '';
        }
        const rerankInfo = this.editForm.knowledgeBaseConfig.knowledgebases
          .length
          ? this.rerankOptions.find(
              item =>
                item.modelId ===
                this.editForm.knowledgeBaseConfig.config.rerankModelId,
            )
          : {};
        const qaRerankInfo = this.editForm.qaKnowledgeBaseConfig.knowledgebases
          .length
          ? this.rerankOptions.find(
              item =>
                item.modelId ===
                this.editForm.qaKnowledgeBaseConfig.config.rerankModelId,
            )
          : {};

        const isAllExternalKnowledgeSelected =
          !this.editForm.knowledgeBaseConfig.knowledgebases.some(
            kb => kb.external !== EXTERNAL,
          );
        const _knowledgeBaseConfig = {
          knowledgebases: this.editForm.knowledgeBaseConfig.knowledgebases,
          config: isAllExternalKnowledgeSelected
            ? {
                matchType: 'mix',
                priorityMatch: 1,
                threshold: this.editForm.knowledgeBaseConfig.config.threshold,
                topK: this.editForm.knowledgeBaseConfig.config.topK,
              }
            : this.editForm.knowledgeBaseConfig.config,
        };

        // 处理推荐问题
        const recommendQuestion = this.editForm.recommendQuestion.map(
          item => item.value,
        );

        let fromParams = {
          ragId: this.editForm.appId,
          knowledgeBaseConfig: _knowledgeBaseConfig,
          qaKnowledgeBaseConfig: this.editForm.qaKnowledgeBaseConfig,
          recommendQuestion:
            recommendQuestion.length > 0 && recommendQuestion[0] !== ''
              ? recommendQuestion
              : [],
          modelConfig: {
            config: this.editForm.modelConfig,
            displayName: modeInfo ? modeInfo.displayName : '',
            model: modeInfo ? modeInfo.model : '',
            modelId: modeInfo ? modeInfo.modelId : '',
            modelType: modeInfo ? modeInfo.modelType : '',
            provider: modeInfo ? modeInfo.provider : '',
          },
          rerankConfig: {
            displayName: rerankInfo ? rerankInfo.displayName : '',
            model: rerankInfo ? rerankInfo.model : '',
            modelId: rerankInfo ? rerankInfo.modelId : '',
            modelType: rerankInfo ? rerankInfo.modelType : '',
            provider: rerankInfo ? rerankInfo.provider : '',
          },
          qaRerankConfig: {
            displayName: qaRerankInfo ? qaRerankInfo.displayName : '',
            model: qaRerankInfo ? qaRerankInfo.model : '',
            modelId: qaRerankInfo ? qaRerankInfo.modelId : '',
            modelType: qaRerankInfo ? qaRerankInfo.modelType : '',
            provider: qaRerankInfo ? qaRerankInfo.provider : '',
          },
          safetyConfig: this.editForm.safetyConfig,
          visionConfig: {
            picNum:
              this.editForm.visionsupport === 'support' &&
              this.getCategory !== KNOWLEDGE &&
              this.visionsupportRerank
                ? this.editForm.visionConfig.picNum
                : 0,
          },
        };
        const res = await updateRagConfig(fromParams);
        if (res.code === 0) {
          this.getDetail();
        }
      } catch (error) {
        console.error(error);
      } finally {
        this.isUpdating = false;
      }
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/draft.scss';
.desc-text {
  max-width: 500px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;
}

/*推荐问题*/
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
      ::v-deep {
        .el-input__inner {
          padding-right: 30px;
        }
      }
    }

    .recommend-del {
      width: 25px;
      line-height: 32px;
      color: #595959;
      cursor: pointer;
    }
  }
}
</style>
