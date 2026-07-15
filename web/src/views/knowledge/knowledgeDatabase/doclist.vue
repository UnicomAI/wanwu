<template>
  <div class="page-wrapper full-content">
    <div class="page-title">
      <i
        class="el-icon-arrow-left"
        @click="$router.push('/knowledge')"
        style="margin-right: 10px; font-size: 20px; cursor: pointer"
      ></i>
      {{ knowledgeName }}
      <div style="margin-left: 34px; font-weight: normal; color: #6b7280">
        uuid: {{ docQuery.knowledgeId }}
        <copyIcon :text="docQuery.knowledgeId" type="icon" />
      </div>
    </div>
    <div class="block table-wrap list-common wrap-fullheight">
      <el-container class="konw_container">
        <el-main class="noPadding">
          <el-container>
            <el-header class="classifyTitle">
              <div class="searchInfo">
                <search-input
                  class="cover-input-icon"
                  :placeholder="$t('knowledgeManage.docPlaceholder')"
                  ref="searchInput"
                  @handleSearch="handleSearch"
                />
                <el-select
                  v-model="docQuery.metaType"
                  class="no-border-input"
                  size="small"
                  style="width: 110px; margin-right: 15px"
                  @change="handleMetaTypeChange"
                >
                  <el-option label="String" value="string" />
                  <el-option label="Number" value="number" />
                  <el-option label="Time" value="time" />
                </el-select>
                <search-input
                  v-if="docQuery.metaType !== 'time'"
                  class="cover-input-icon"
                  :placeholder="$t('knowledgeManage.metaPlaceholder')"
                  ref="searchInputMeta"
                  @handleSearch="handleSearchByMeta"
                />
                <el-date-picker
                  v-else
                  v-model="metaDateRange"
                  :end-placeholder="
                    $t('knowledgeManage.meta.metaValueEndPlaceholder')
                  "
                  :start-placeholder="
                    $t('knowledgeManage.meta.metaValueStartPlaceholder')
                  "
                  class="no-border-input"
                  size="small"
                  style="width: 330px"
                  type="datetimerange"
                  @change="handleMetaDateChange"
                />
              </div>

              <div class="content_title">
                <el-button
                  size="mini"
                  type="primary"
                  icon="el-icon-refresh"
                  @click="reload"
                ></el-button>
                <template v-if="showGraphReport">
                  <el-dropdown
                    v-for="(group, index) in graphDropdownGroups"
                    :key="group.label"
                    @command="handleCommand"
                    :style="{ margin: index === 0 ? '0 10px' : '' }"
                  >
                    <el-button size="mini" type="primary">
                      {{ group.label }}
                      <i :class="['el-icon--right', group.icon]"></i>
                    </el-button>
                    <el-dropdown-menu slot="dropdown">
                      <el-dropdown-item
                        v-for="item in group.items"
                        :key="item.command"
                        :command="item.command"
                      >
                        {{ item.label }}
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </el-dropdown>
                </template>

                <el-button
                  size="mini"
                  type="primary"
                  @click="showMeta"
                  v-if="hasManagePerm"
                >
                  {{ $t('knowledgeManage.docList.metaDataManagement') }}
                </el-button>
                <el-button
                  size="mini"
                  type="primary"
                  @click="
                    $router.push(
                      `/knowledge/hitTest?knowledgeId=${docQuery.knowledgeId}&graphSwitch=${graphSwitch}&category=${category}`,
                    )
                  "
                >
                  {{ $t('knowledgeManage.hitTest.name') }}
                </el-button>
                <el-button
                  size="mini"
                  type="primary"
                  :underline="false"
                  @click="handleUpload"
                  v-if="hasManagePerm"
                >
                  {{ $t('knowledgeManage.fileUpload') }}
                </el-button>
                <template v-if="hasManagePerm">
                  <el-dropdown
                    v-for="(group, index) in dropdownGroups"
                    :key="group.label"
                    @command="handleCommand"
                    :style="{ margin: index === 0 ? '0 10px' : '' }"
                  >
                    <el-button size="mini" type="primary">
                      {{ group.label }}
                      <i :class="['el-icon--right', group.icon]"></i>
                    </el-button>
                    <el-dropdown-menu slot="dropdown">
                      <el-dropdown-item
                        v-for="item in group.items"
                        :key="item.command"
                        :command="item.command"
                      >
                        {{ item.label }}
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </el-dropdown>
                </template>
              </div>
            </el-header>
            <el-main class="noPadding" v-loading="tableLoading">
              <el-alert
                :title="title_tips"
                type="warning"
                show-icon
                style="margin-bottom: 10px"
                v-if="showTips"
              ></el-alert>
              <el-descriptions
                style="margin-bottom: 10px"
                title=""
                :column="1"
                border
              >
                <el-descriptions-item
                  :label="$t('knowledgeManage.knowledgeName')"
                  labelStyle="width: 120px"
                >
                  {{ knowledgeName }}
                  <i
                    v-if="[POWER_TYPE_SYSTEM_ADMIN].includes(permissionType)"
                    class="el-icon-edit-outline"
                    style="cursor: pointer"
                    @click="showEdit"
                  ></i>
                </el-descriptions-item>
                <el-descriptions-item
                  :label="$t('knowledgeManage.desc')"
                  labelStyle="width: 120px"
                >
                  <span>
                    {{ description || $t('knowledgeManage.zeroData') }}
                  </span>
                  <i
                    v-if="[POWER_TYPE_SYSTEM_ADMIN].includes(permissionType)"
                    class="el-icon-edit-outline"
                    style="cursor: pointer"
                    @click="showEdit"
                  ></i>
                </el-descriptions-item>
                <el-descriptions-item
                  label="Embedding"
                  labelStyle="width: 120px"
                >
                  <div class="keyword-tags">
                    <template v-if="embeddingModel">
                      <img
                        :src="convertModelIcon(embeddingModel?.avatar?.path)"
                        class="model-img"
                      />
                      <span class="model-name">
                        {{ embeddingModel.displayName }}
                      </span>
                      <template
                        v-if="
                          embeddingModel.tags && embeddingModel.tags.length > 0
                        "
                      >
                        <el-tag
                          v-for="(item, index) in embeddingModel.tags"
                          :key="index"
                          size="small"
                          color="#E6F0FF"
                          class="keyword-tag"
                        >
                          {{ item.text }}
                        </el-tag>
                      </template>
                    </template>
                    <span v-else>{{ $t('knowledgeManage.zeroData') }}</span>
                  </div>
                </el-descriptions-item>
                <el-descriptions-item
                  v-if="graphLlmModel"
                  :label="$t('knowledgeManage.graph.graphModel')"
                  labelStyle="width: 120px"
                >
                  <div class="keyword-tags">
                    {{ graphLlmModel.displayName }}
                    <template
                      v-if="graphLlmModel.tags && graphLlmModel.tags.length > 0"
                    >
                      <el-tag
                        v-for="(item, index) in graphLlmModel.tags"
                        :key="index"
                        class="keyword-tag"
                        color="#E6F0FF"
                        size="small"
                      >
                        {{ item.text }}
                      </el-tag>
                    </template>
                  </div>
                </el-descriptions-item>
                <el-descriptions-item
                  :label="$t('knowledgeManage.keyWordConfig')"
                  labelStyle="width: 120px"
                >
                  <div class="keyword-tags">
                    <template v-if="keywords && keywords.length > 0">
                      <el-tag
                        v-for="(item, index) in keywords"
                        :key="index"
                        size="small"
                        color="#E6F0FF"
                        class="keyword-tag"
                      >
                        {{ item.name }} : {{ item.alias }}
                      </el-tag>
                    </template>
                    <span v-else>{{ $t('knowledgeManage.zeroData') }}</span>
                    <i
                      class="el-icon-edit-outline"
                      style="cursor: pointer"
                      @click="$router.push(`/knowledge/keyword`)"
                    ></i>
                  </div>
                </el-descriptions-item>
              </el-descriptions>
              <el-table
                ref="dataTable"
                :data="tableData"
                style="width: 100%"
                :row-key="'docId'"
                :header-cell-style="{ background: '#F9F9F9', color: '#999999' }"
                @selection-change="handleSelectionChange"
              >
                <el-table-column
                  type="selection"
                  reserve-selection
                  v-if="hasManagePerm"
                  key="selection"
                  width="55"
                ></el-table-column>
                <el-table-column
                  prop="docName"
                  :label="$t('knowledgeManage.fileName')"
                  min-width="180"
                >
                  <template slot-scope="scope">
                    <el-popover
                      placement="bottom-start"
                      :content="scope.row.docName"
                      trigger="hover"
                      width="300"
                    >
                      <span slot="reference">
                        {{
                          scope.row.docName.length > 20
                            ? scope.row.docName.slice(0, 20) + '...'
                            : scope.row.docName
                        }}
                      </span>
                    </el-popover>
                  </template>
                </el-table-column>
                <el-table-column
                  prop="docType"
                  :label="$t('knowledgeManage.fileStyle')"
                ></el-table-column>
                <el-table-column
                  prop="segmentMethod"
                  :label="$t('knowledgeManage.docList.segmentMode')"
                >
                  <template slot-scope="scope">
                    <span>
                      {{ getSegmentMethodText(scope.row.segmentMethod) }}
                    </span>
                  </template>
                </el-table-column>
                <el-table-column
                  prop="author"
                  :label="$t('knowledgeManage.author')"
                ></el-table-column>
                <el-table-column
                  prop="uploadTime"
                  :label="$t('knowledgeManage.importTime')"
                  width="200"
                ></el-table-column>
                <el-table-column prop="status" width="150">
                  <template #header>
                    <div style="display: flex; align-items: center">
                      <span>{{ $t('knowledgeManage.currentStatus') }}</span>
                      <FilterPopover
                        style="margin-left: 5px"
                        :options="KNOWLEDGE_STATUS_OPTIONS"
                        @applyFilter="filterCurrentStatus"
                      />
                    </div>
                  </template>
                  <template slot-scope="scope">
                    <span
                      :class="[
                        [
                          KNOWLEDGE_STATUS_CHECK_FAIL,
                          KNOWLEDGE_STATUS_FAIL,
                        ].includes(scope.row.status)
                          ? 'error'
                          : '',
                      ]"
                    >
                      {{ getCurrentStatus(scope.row.status) }}
                    </span>
                    <el-tooltip
                      class="item"
                      effect="light"
                      :content="scope.row.errorMsg ? scope.row.errorMsg : ''"
                      placement="top"
                      v-if="scope.row.status === KNOWLEDGE_STATUS_FAIL"
                      popper-class="custom-tooltip"
                    >
                      <span
                        class="el-icon-warning"
                        style="margin-left: 5px; color: #e6a23c"
                      ></span>
                    </el-tooltip>
                    <i
                      class="el-icon-refresh"
                      style="margin-left: 5px; color: #409eff"
                      @click="handleRetry(scope.row)"
                      v-if="
                        scope.row.status === KNOWLEDGE_STATUS_FAIL &&
                        scope.row.docType !== 'url'
                      "
                    ></i>
                  </template>
                </el-table-column>
                <!-- 任务进度列 -->
                <el-table-column
                  :label="$t('knowledgeManage.docProgress')"
                  prop="docProgress"
                  width="100"
                >
                  <template slot-scope="scope">
                    <el-progress
                      :percentage="scope.row.docProgress || 0"
                      :status="getProgressStatus(scope.row.status)"
                      :stroke-width="6"
                      :width="40"
                      type="circle"
                    />
                  </template>
                </el-table-column>
                <el-table-column
                  v-if="graphSwitch"
                  key="graphStatus"
                  prop="graphStatus"
                  width="150"
                >
                  <template #header>
                    <div style="display: flex; align-items: center">
                      <span>{{ $t('knowledgeManage.graph.graphStatus') }}</span>
                      <FilterPopover
                        style="margin-left: 5px"
                        :options="
                          KNOWLEDGE_GRAPH_STATUS_OPTIONS.filter(
                            opt => opt.value !== KNOWLEDGE_GRAPH_STATUS_INITIAL,
                          )
                        "
                        @applyFilter="filterGraphStatus"
                      />
                    </div>
                  </template>
                  <template slot-scope="scope">
                    <span>
                      {{ getGraphStatus(scope.row.graphStatus) }}
                    </span>
                    <el-tooltip
                      class="item"
                      effect="light"
                      :content="scope.row.graphErrMsg ?? ''"
                      placement="top"
                      v-if="scope.row.graphStatus === STATUS_FAILED"
                      popper-class="custom-tooltip"
                    >
                      <span
                        class="el-icon-warning"
                        style="margin-left: 5px; color: #e6a23c"
                      ></span>
                    </el-tooltip>
                  </template>
                </el-table-column>
                <!-- 图谱进度列 -->
                <el-table-column
                  v-if="graphSwitch"
                  key="graphProgress"
                  :label="$t('knowledgeManage.graphProgress')"
                  prop="graphProgress"
                  width="100"
                >
                  <template slot-scope="scope">
                    <el-progress
                      :percentage="scope.row.graphProgress || 0"
                      :status="getGraphProgressStatus(scope.row.graphStatus)"
                      :stroke-width="6"
                      :width="40"
                      type="circle"
                    />
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
                      :disabled="
                        [
                          KNOWLEDGE_STATUS_CHECKING,
                          KNOWLEDGE_STATUS_ANALYSING,
                        ].includes(Number(scope.row.status))
                      "
                      v-if="hasManagePerm"
                    >
                      {{ $t('common.button.delete') }}
                    </el-button>
                    <el-button
                      size="mini"
                      round
                      @click="handleConfig([scope.row.docId])"
                      :disabled="
                        ![KNOWLEDGE_STATUS_FINISH].includes(
                          Number(scope.row.status),
                        ) || scope.row.docType === 'url'
                      "
                      v-if="hasManagePerm"
                    >
                      {{ $t('knowledgeManage.segmentConfig') }}
                    </el-button>
                    <el-button size="mini" round @click="handleView(scope.row)">
                      {{ $t('knowledgeManage.view') }}
                    </el-button>
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
    <!-- 元数据管理 -->
    <el-dialog
      :title="$t('knowledgeManage.docList.metaDataManagement')"
      :visible.sync="metaVisible"
      width="550px"
      :before-close="handleClose"
    >
      <mataData
        ref="mataData"
        @updateMeta="updateMeta"
        type="create"
        :knowledgeId="docQuery.knowledgeId"
        class="mataData"
      />
      <span slot="footer" class="dialog-footer">
        <!-- <el-button @click="handleClose">
          {{ $t('common.button.cancel') }}
        </el-button> -->
        <el-button type="primary" @click="createMeta">
          {{ $t('common.button.create') }}
        </el-button>
        <el-button type="primary" @click="submitMeta" :disabled="isDisabled">
          {{ $t('common.button.confirm') }}
        </el-button>
      </span>
    </el-dialog>

    <!-- 批量编辑元数据值弹窗 -->
    <batchMetaData
      ref="batchMetaData"
      :selectedDocIds="selectedDocIds"
      @reLoadDocList="reLoadDocList"
    />
    <!-- 批量编辑元数据值操作框 -->
    <BatchMetaButton
      ref="BatchMetaButton"
      :selectedCount="selectedTableData.length"
      type="knowledge"
      @showBatchMeta="showBatchMeta"
      @handleBatchDelete="handleBatchDelete"
      @handleBatchExport="handleBatchExport"
      @handleBatchConfig="handleBatchConfig"
      @handleMetaCancel="handleMetaCancel"
    />
    <!-- 导出记录 -->
    <exportRecord ref="exportRecord" />
    <createKnowledge
      ref="createKnowledge"
      :category="KNOWLEDGE"
      @reloadData="reload"
    />
  </div>
</template>

<script>
import Pagination from '@/components/pagination.vue';
import SearchInput from '@/components/searchInput.vue';
import mataData from '../component/metadata.vue';
import batchMetaData from '../component/meta/batchMetaData.vue';
import BatchMetaButton from '../component/meta/batchMetaButton.vue';
import FilterPopover from '@/components/filterPopover.vue';
import {
  getDocList,
  delDocItem,
  uploadFileTips,
  updateDocMeta,
  exportDoc,
  docReImport,
} from '@/api/knowledge';
import {
  DROPDOWN_GROUPS,
  KNOWLEDGE_GRAPH_STATUS_OPTIONS,
  KNOWLEDGE_STATUS_OPTIONS,
} from '../config';
import {
  INITIAL,
  STATUS_FAILED,
  STATUS_FINISHED,
  POWER_TYPE_EDIT,
  POWER_TYPE_ADMIN,
  POWER_TYPE_SYSTEM_ADMIN,
  KNOWLEDGE_STATUS_UPLOADED,
  ALL,
  KNOWLEDGE_STATUS_PENDING_PROCESSING,
  KNOWLEDGE_STATUS_FINISH,
  KNOWLEDGE_STATUS_CHECKING,
  KNOWLEDGE_STATUS_ANALYSING,
  KNOWLEDGE_STATUS_CHECK_FAIL,
  KNOWLEDGE_STATUS_FAIL,
  KNOWLEDGE_GRAPH_STATUS_INITIAL,
  KNOWLEDGE,
} from '@/views/knowledge/constants';
import exportRecord from '@/views/knowledge/qaDatabase/exportRecord.vue';
import CopyIcon from '@/components/copyIcon.vue';
import createKnowledge from '@/views/knowledge/component/create.vue';
import { selectModelList } from '@/api/modelAccess';
import { avatarSrc, getModelDefaultIcon } from '@/utils/util';

export default {
  name: 'KnowledgeDoclist',
  components: {
    CopyIcon,
    exportRecord,
    Pagination,
    SearchInput,
    mataData,
    batchMetaData,
    BatchMetaButton,
    FilterPopover,
    createKnowledge,
  },
  data() {
    return {
      avatar: '',
      knowledgeName: '',
      description: '',
      category: 0,
      embeddingModel: {},
      keywords: [],
      llmModelId: '',
      graphLlmModel: null, // 知识图谱解析模型信息
      loading: false,
      tableLoading: false,
      docQuery: {
        docIdList: [],
        docName: '',
        metaType: 'string',
        metaValue: '',
        metaStartTime: '',
        metaEndTime: '',
        knowledgeId: '',
        status: [ALL],
        graphStatus: [ALL],
      },
      metaDateRange: null,
      fileList: [],
      listApi: getDocList,
      title_tips: '',
      showTips: false,
      tableData: [],
      KNOWLEDGE_STATUS_OPTIONS,
      knowledgeData: [],
      currentKnowValue: null,
      timer: null,
      refreshCount: 0,
      tagList: [],
      metaVisible: false,
      metaData: [],
      isDisabled: false,
      selectedTableData: [],
      selectedDocIds: [],
      graphSwitch: false,
      showGraphReport: false,
      KNOWLEDGE_GRAPH_STATUS_OPTIONS,
      dropdownGroups: DROPDOWN_GROUPS.slice(0, 1),
      graphDropdownGroups: DROPDOWN_GROUPS.slice(2),
      KNOWLEDGE,
      INITIAL,
      STATUS_FINISHED,
      STATUS_FAILED,
      POWER_TYPE_EDIT,
      POWER_TYPE_ADMIN,
      POWER_TYPE_SYSTEM_ADMIN,
      KNOWLEDGE_STATUS_PENDING_PROCESSING,
      KNOWLEDGE_STATUS_FINISH,
      KNOWLEDGE_STATUS_CHECKING,
      KNOWLEDGE_STATUS_ANALYSING,
      KNOWLEDGE_STATUS_CHECK_FAIL,
      KNOWLEDGE_STATUS_FAIL,
      KNOWLEDGE_GRAPH_STATUS_INITIAL,
      permissionType: null,
    };
  },
  watch: {
    $route: {
      handler(val) {
        if (val.query.done) {
          this.refreshCount = 0;
          this.startTimer();
        }
      },
      immediate: true,
    },
    metaData: {
      handler(val) {
        this.isDisabled = !!(
          val.some(item => !item.metaKey || !item.metaValueType) || !val.length
        );
      },
    },
  },
  computed: {
    hasManagePerm() {
      return [
        POWER_TYPE_EDIT,
        POWER_TYPE_ADMIN,
        POWER_TYPE_SYSTEM_ADMIN,
      ].includes(this.permissionType);
    },
  },
  activated() {
    this.docQuery.knowledgeId = this.$route.params.id;
    this.getTableData(this.docQuery);
  },
  deactivated() {
    this.clearTimer();
    this.handleMetaCancel();
  },
  methods: {
    convertModelIcon(iconPath) {
      return iconPath ? avatarSrc(iconPath) : getModelDefaultIcon();
    },
    showEdit() {
      this.$refs.createKnowledge.showDialog({
        category: this.category,
        knowledgeId: this.docQuery.knowledgeId,
        avatar: this.avatar,
        name: this.knowledgeName,
        description: this.description,
        embeddingModelInfo: this.embeddingModel,
        llmModelId: this.llmModelId,
        graphSwitch: this.graphSwitch ? 1 : 0,
      });
    },
    handleCommand(command) {
      const actions = {
        exportData: this.exportData,
        exportRecord: this.exportRecord,
        goKnowledgeGraph: this.goKnowledgeGraph,
        goCommunityReport: this.goCommunityReport,
      };
      (actions[command] || this.exportData)();
    },
    goKnowledgeGraph() {
      this.$router.push(
        `/knowledge/graphMap/${this.docQuery.knowledgeId}?name=${this.knowledgeName}`,
      );
    },
    goCommunityReport() {
      this.$router.push({
        path: '/knowledge/communityReport',
        query: {
          knowledgeId: this.docQuery.knowledgeId,
          name: this.knowledgeName,
          permissionType: this.permissionType,
        },
      });
    },
    exportData(docIdList) {
      if (!this.docQuery.knowledgeId) {
        this.$message.warning(this.$t('common.noData'));
        return;
      }
      if (this.loading) return;
      const params = {
        knowledgeId: this.docQuery.knowledgeId,
        docIdList: docIdList,
      };
      this.loading = true;
      exportDoc(params)
        .then(res => {
          if (res.code === 0) {
            this.$message.success(this.$t('knowledgeManage.exportAllTips'));
            const data = res.data || {};
            const url = data.fileUrl || data.downloadUrl;
            if (url) {
              window.open(url, '_blank');
            } else if (data.recordCreated) {
              this.exportRecord();
            }
          }
        })
        .finally(() => {
          this.loading = false;
        });
    },
    exportRecord() {
      this.$refs.exportRecord.showDialog(this.docQuery.knowledgeId);
    },
    handleMetaCancel() {
      this.selectedTableData = [];
      this.selectedDocIds = [];
      // 取消所有表格数据的选中状态
      this.$nextTick(() => {
        this.$refs.dataTable?.clearSelection();
      });
    },
    reLoadDocList() {
      this.getTableData(this.docQuery);
      this.handleMetaCancel();
    },
    showBatchMeta() {
      if (!this.selectedTableData || this.selectedTableData.length === 0) {
        this.$message.warning(
          this.$t('knowledgeManage.docList.pleaseSelectDocFirst'),
        );
        return;
      }
      this.$refs.batchMetaData.showDialog();
    },
    handleSelectionChange(val) {
      if (val.length > 100) {
        this.$message.warning(
          this.$t('knowledgeManage.docList.maxSelect100Files'),
        );
        return;
      }
      this.selectedTableData = val;
      this.selectedDocIds = val.map(item => item.docId);
    },
    getSegmentMethodText(value) {
      switch (value) {
        case '0':
          return this.$t('knowledgeManage.config.commonSegment');
        case '1':
          return this.$t('knowledgeManage.config.parentSonSegment');
        default:
          return this.$t('knowledgeManage.docList.unknown');
      }
    },
    createMeta() {
      this.$refs.mataData.createMetaData();
      this.scrollToBottom();
    },
    scrollToBottom() {
      this.$nextTick(() => {
        const container = this.$refs.mataData;
        if (container) {
          container.scrollTop = container.scrollHeight;
        }
      });
    },
    submitMeta() {
      this.isDisabled = true;
      const metaList = this.metaData
        .filter(item => item.option !== '')
        .map(({ metaId, metaKey, metaValueType, option }) => ({
          metaKey,
          ...(option === 'add' ? { metaValueType } : {}),
          option,
          ...(option === 'update' || option === 'delete' ? { metaId } : {}),
        }));
      const data = {
        docId: '',
        knowledgeId: this.docQuery.knowledgeId,
        metaDataList: metaList,
      };
      updateDocMeta(data)
        .then(res => {
          if (res.code === 0) {
            this.$message.success(this.$t('common.message.success'));
            this.$refs.mataData.getList();
            this.metaVisible = false;
            this.isDisabled = false;
          }
        })
        .catch(() => {
          this.isDisabled = false;
        });
    },
    showMeta() {
      this.metaVisible = true;
    },
    updateMeta(data) {
      this.metaData = data;
    },
    handleClose() {
      this.metaVisible = false;
    },
    startTimer() {
      this.clearTimer();
      if (this.refreshCount >= 2) {
        return;
      }
      const delay = this.refreshCount === 0 ? 1000 : 3000;
      this.timer = setTimeout(async () => {
        await this.getTableData(this.docQuery);
        this.refreshCount++;
        this.startTimer();
      }, delay);
    },
    clearTimer() {
      if (this.timer) {
        clearTimeout(this.timer);
        this.timer = null;
      }
    },
    reload() {
      this.getTableData(this.docQuery);
    },
    handleSearch(val) {
      this.docQuery.docName = val;
      this.getTableData(this.docQuery);
    },
    handleSearchByMeta(val) {
      this.docQuery.metaValue = val;
      this.getTableData(this.docQuery);
    },
    handleMetaTypeChange() {
      this.docQuery.metaValue = '';
      this.$refs.searchInputMeta.clearValue();
      this.docQuery.metaStartTime = '';
      this.docQuery.metaEndTime = '';
      this.metaDateRange = null;
      this.getTableData(this.docQuery);
    },
    handleMetaDateChange(val) {
      if (val && val.length === 2) {
        this.docQuery.metaStartTime = new Date(val[0]).getTime().toString();
        this.docQuery.metaEndTime = new Date(val[1]).getTime().toString();
      } else {
        this.docQuery.metaStartTime = '';
        this.docQuery.metaEndTime = '';
      }
      this.getTableData(this.docQuery);
    },
    handleRetry(data) {
      docReImport({
        docIdList: [data.docId],
        knowledgeId: this.docQuery.knowledgeId,
      }).then(res => {
        if (res.code === 0) {
          this.$message.success(this.$t('common.info.retry'));
          this.reLoadDocList();
        }
      });
    },
    async handleDelete(docIdList) {
      this.loading = true;
      try {
        let res = await delDocItem({
          docIdList,
          knowledgeId: this.docQuery.knowledgeId,
        });
        if (res.code === 0) {
          this.$message.success(this.$t('common.info.delete'));
        }
      } finally {
        this.reLoadDocList();
        this.loading = false;
      }
    },
    handleDel(data) {
      this.$confirm(
        this.$t('knowledgeManage.deleteTips'),
        this.$t('knowledgeManage.tip'),
        {
          confirmButtonText: this.$t('common.button.confirm'),
          cancelButtonText: this.$t('common.button.cancel'),
          type: 'warning',
        },
      ).then(() => {
        this.handleDelete([data.docId]);
      });
    },
    handleConfig(docIdList) {
      this.$router.push({
        path: '/knowledge/fileUpload',
        query: {
          id: this.docQuery.knowledgeId,
          name: this.knowledgeName,
          mode: 'config',
          title: this.$t('knowledgeManage.segmentConfig'),
          docIdList: docIdList,
          category: this.category,
        },
      });
    },
    handleBatchDelete() {
      this.$confirm(
        this.$t('knowledgeManage.deleteBatchTips'),
        this.$t('knowledgeManage.tip'),
        {
          confirmButtonText: this.$t('common.button.confirm'),
          cancelButtonText: this.$t('common.button.cancel'),
          type: 'warning',
        },
      ).then(() => {
        this.handleDelete(this.selectedDocIds);
      });
    },
    handleBatchExport() {
      // 统计URL类型文档数量
      const urlDocs = this.selectedTableData.filter(
        doc => doc.docType === 'url',
      );
      const nonUrlDocs = this.selectedTableData.filter(
        doc => doc.docType !== 'url',
      );

      const totalCount = this.selectedTableData.length;
      const urlCount = urlDocs.length;
      const exportCount = nonUrlDocs.length;

      this.$alert(
        this.$t('knowledgeManage.batchExportTips', {
          total: totalCount,
          urlCount: urlCount,
          exportCount: exportCount,
        }),
        this.$t('knowledgeManage.tip'),
        {
          confirmButtonText: this.$t('common.button.confirm'),
          type: 'warning',
        },
      ).then(() => {
        // 用户点击确定后，执行导出（仅当有非URL文档时）
        if (exportCount > 0) {
          this.exportData(nonUrlDocs.map(doc => doc.docId));
        }
      });
    },
    handleBatchConfig() {
      const urlDocs = this.selectedTableData.filter(
        doc => doc.docType === 'url',
      );
      const nonUrlDocs = this.selectedTableData.filter(
        doc => doc.docType !== 'url',
      );

      const unprocessedDocs = nonUrlDocs.filter(
        doc => doc.status !== KNOWLEDGE_STATUS_FINISH,
      );
      const allowedDocs = nonUrlDocs.filter(
        doc => doc.status === KNOWLEDGE_STATUS_FINISH,
      );

      if (
        nonUrlDocs.some(doc => doc.isMultimodal === true) &&
        nonUrlDocs.some(doc => doc.isMultimodal === false)
      ) {
        this.$alert(
          this.$t('knowledgeManage.multimodalMixTips'),
          this.$t('knowledgeManage.tip'),
          {
            confirmButtonText: this.$t('common.button.confirm'),
            type: 'warning',
          },
        );
      } else if (urlDocs.length > 0 || unprocessedDocs.length > 0) {
        let message = this.$t('knowledgeManage.batchConfigTips', {
          total: this.selectedTableData.length,
          unprocessedNum: urlDocs.length + unprocessedDocs.length,
        });
        if (allowedDocs.length > 0) {
          message += this.$t('knowledgeManage.continueTips');
          this.$confirm(message, this.$t('knowledgeManage.tip'), {
            confirmButtonText: this.$t('common.button.confirm'),
            cancelButtonText: this.$t('common.button.cancel'),
            type: 'warning',
          })
            .then(() => {
              this.handleConfig(allowedDocs.map(doc => doc.docId));
            })
            .catch(() => {});
        } else {
          this.$alert(message, this.$t('knowledgeManage.tip'), {
            confirmButtonText: this.$t('common.button.confirm'),
            type: 'warning',
          });
        }
      } else {
        this.handleConfig(nonUrlDocs.map(doc => doc.docId));
      }
    },
    async getTableData(data) {
      this.tableLoading = true;
      this.tableData = await this.$refs['pagination'].getTableData(data);
      this.tableLoading = false;
      this.getTips();
    },
    getTips() {
      uploadFileTips({ knowledgeId: this.docQuery.knowledgeId }).then(res => {
        if (res.code === 0) {
          if (res.data.uploadstatus === 1) {
            this.showTips = true;
            this.title_tips = this.$t('knowledgeManage.refreshTips');
          } else if (res.data.uploadstatus === 2) {
            this.showTips = false;
            this.title_tips = '';
          } else {
            this.showTips = true;
            this.title_tips = res.data.msg;
          }
        }
      });
    },
    filterCurrentStatus(data) {
      this.docQuery.status = data;
      this.getTableData({ ...this.docQuery, pageNo: 1 });
    },
    filterGraphStatus(data) {
      this.docQuery.graphStatus = data;
      this.getTableData({ ...this.docQuery, pageNo: 1 });
    },
    getCurrentStatus(status) {
      if (status === KNOWLEDGE_STATUS_UPLOADED) {
        return this.$t('knowledgeManage.beUploaded');
      }
      const statusOption = KNOWLEDGE_STATUS_OPTIONS.find(
        option => option.value === status,
      );
      return statusOption
        ? statusOption.label
        : this.$t('knowledgeManage.noStatus');
    },
    getGraphStatus(status) {
      const statusOption = KNOWLEDGE_GRAPH_STATUS_OPTIONS.find(
        option => option.value === status,
      );
      return statusOption
        ? statusOption.label
        : this.$t('knowledgeManage.noStatus');
    },
    // 根据状态获取进度条的状态
    getProgressStatus(status) {
      if (status === KNOWLEDGE_STATUS_FINISH) {
        return 'success';
      }
      if (status === KNOWLEDGE_STATUS_FAIL) {
        return 'exception';
      }
      // 处理中状态不设置status，显示默认蓝色
      return undefined;
    },
    // 根据图谱状态获取进度条的状态
    getGraphProgressStatus(graphStatus) {
      if (graphStatus === STATUS_FINISHED) {
        return 'success';
      }
      if (graphStatus === STATUS_FAILED) {
        return 'exception';
      }
      // 处理中状态不设置status，显示默认蓝色
      return undefined;
    },
    handleView(row) {
      this.$router.push({
        path: '/knowledge/section',
        query: {
          id: row.docId,
          type: row.docType,
          name: row.docName,
          knowledgeId: row.knowledgeId,
          knowledgeName: this.knowledgeName,
          category: this.category,
          disable: [
            KNOWLEDGE_STATUS_PENDING_PROCESSING,
            KNOWLEDGE_STATUS_ANALYSING,
            KNOWLEDGE_STATUS_FAIL,
          ].includes(Number(row.status)),
          permissionType: this.permissionType,
        },
      });
    },
    handleUpload() {
      this.$router.push({
        path: '/knowledge/fileUpload',
        query: {
          id: this.docQuery.knowledgeId,
          name: this.knowledgeName,
          category: this.category,
        },
      });
    },
    refreshData(data, tableInfo) {
      this.tableData = data;
      if (tableInfo && tableInfo.docKnowledgeInfo) {
        this.graphSwitch = tableInfo.docKnowledgeInfo.graphSwitch === 1;
        this.showGraphReport = tableInfo.docKnowledgeInfo.showGraphReport;
        this.avatar = tableInfo.docKnowledgeInfo.avatar;
        this.knowledgeName = tableInfo.docKnowledgeInfo.knowledgeName;
        this.description = tableInfo.docKnowledgeInfo.description;
        this.category = tableInfo.docKnowledgeInfo.category;
        this.embeddingModel = tableInfo.docKnowledgeInfo.embeddingModel;
        this.keywords = tableInfo.docKnowledgeInfo.keywords;
        this.llmModelId = tableInfo.docKnowledgeInfo.llmModelId;
        this.permissionType = tableInfo.docKnowledgeInfo.permissionType;
        // 如果开启了知识图谱且有大模型ID，则查询模型详情
        if (this.graphSwitch && this.llmModelId) {
          selectModelList().then(res => {
            if (res.code === 0 && res.data.list) {
              const model = res.data.list.find(
                item => item.modelId === this.llmModelId,
              );
              this.graphLlmModel = model || null;
            }
          });
        } else {
          this.graphLlmModel = null;
        }
      } else {
        this.graphSwitch = false;
        this.showGraphReport = false;
        this.graphLlmModel = null;
      }
    },
  },
};
</script>
<style lang="scss" scoped>
.read-only {
  cursor: not-allowed;
}

.keyword-tags {
  align-items: center;
  display: flex;
  flex-wrap: wrap;
  gap: 4px;

  .model-img {
    width: 18px;
    height: 18px;
    border-radius: 4px;
    vertical-align: middle;
  }

  .model-name {
    font-weight: 500;
  }
}

.keyword-tag {
  margin-right: 4px;
  margin-bottom: 2px;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: $tag_color;
}

.mataData {
  max-height: 400px;
  overflow-y: auto;
}

.edit-icon {
  color: $color;
  cursor: pointer;
  font-size: 16px;
  margin-left: 5px;
}

.doc_tag {
  margin: 0 2px;
}

::v-deep {
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

  .el-dialog__body {
    padding: 10px 20px;
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
        color: #67c23a;
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
  border-color: #eee; /* 设置边框颜色 */
  background-color: #fff; /* 设置背景颜色 */
  color: #666; /* 设置文字颜色 */
}

.custom-tooltip.el-tooltip__popper[x-placement^='top'] .popper__arrow::after {
  border-top-color: #fff !important;
}

.custom-tooltip.el-tooltip__popper.is-light[x-placement^='top'] .popper__arrow {
  border-top-color: #ccc !important;
}
</style>
