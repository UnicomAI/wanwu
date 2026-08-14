<template>
  <div class="page-wrapper full-content">
    <div v-if="!readonly" class="page-title">
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
                  :placeholder="$t('knowledgeManage.questionPlaceholder')"
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

              <div v-if="!readonly" class="content_title">
                <el-button
                  size="mini"
                  type="primary"
                  icon="el-icon-refresh"
                  @click="reload"
                ></el-button>
                <el-button
                  size="mini"
                  type="primary"
                  @click="
                    $router.push(
                      `/knowledge/hitTest?knowledgeId=${docQuery.knowledgeId}&type=qa`,
                    )
                  "
                >
                  {{ $t('knowledgeManage.hitTest.name') }}
                </el-button>
                <el-button
                  size="mini"
                  type="primary"
                  @click="showMeta"
                  v-if="hasManagePerm"
                >
                  {{ $t('knowledgeManage.docList.metaDataManagement') }}
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
                v-if="showTips"
                :title="title_tips"
                show-icon
                style="margin-bottom: 10px"
                type="warning"
              ></el-alert>
              <el-descriptions
                v-if="!readonly"
                :column="1"
                border
                style="margin-bottom: 10px"
                title=""
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
                      {{ embeddingModel.displayName }}
                      <template
                        v-if="
                          embeddingModel.tags && embeddingModel.tags.length > 0
                        "
                      >
                        <el-tag
                          v-for="(item, index) in embeddingModel.tags"
                          :key="index"
                          class="keyword-tag"
                          color="#E6F0FF"
                          size="small"
                        >
                          {{ item.text }}
                        </el-tag>
                      </template>
                    </template>
                    <span v-else>{{ $t('knowledgeManage.zeroData') }}</span>
                  </div>
                </el-descriptions-item>
              </el-descriptions>
              <el-table
                ref="dataTable"
                :data="tableData"
                style="width: 100%"
                :row-key="'qaPairId'"
                :header-cell-style="{ background: '#F9F9F9', color: '#999999' }"
                @selection-change="handleSelectionChange"
              >
                <el-table-column
                  type="selection"
                  reserve-selection
                  key="selection"
                  v-if="hasManagePerm"
                  width="55"
                ></el-table-column>
                <el-table-column
                  prop="question"
                  :label="$t('knowledgeManage.qaDatabase.question')"
                  min-width="100"
                >
                  <template slot-scope="scope">
                    <el-popover
                      placement="bottom-start"
                      :content="scope.row.question"
                      trigger="hover"
                    >
                      <span slot="reference">
                        {{
                          scope.row.question.length > 20
                            ? scope.row.question.slice(0, 20) + '...'
                            : scope.row.question
                        }}
                      </span>
                    </el-popover>
                  </template>
                </el-table-column>
                <el-table-column
                  prop="answer"
                  :label="$t('knowledgeManage.qaDatabase.answer')"
                  width="180"
                >
                  <template slot-scope="scope">
                    <el-popover
                      placement="bottom-start"
                      :content="scope.row.answer"
                      trigger="hover"
                      width="300"
                    >
                      <span slot="reference">
                        {{
                          scope.row.answer.length > 20
                            ? scope.row.answer.slice(0, 20) + '...'
                            : scope.row.answer
                        }}
                      </span>
                    </el-popover>
                  </template>
                </el-table-column>
                <el-table-column
                  prop="metaDataList"
                  :label="$t('knowledgeManage.qaDatabase.metaData')"
                  v-if="hasManagePerm"
                  key="metaDataList"
                >
                  <template slot-scope="scope">
                    <span>
                      {{ getMetaDataText(scope.row.metaDataList) }}
                    </span>
                    <span
                      class="el-icon-edit-outline edit-icon"
                      @click="handleEditMetaData(scope.row)"
                    ></span>
                  </template>
                </el-table-column>
                <el-table-column
                  prop="author"
                  :label="$t('knowledgeManage.author')"
                ></el-table-column>
                <el-table-column
                  prop="switch"
                  :label="$t('user.table.status')"
                  v-if="hasManagePerm"
                  key="switch"
                >
                  <template slot-scope="scope">
                    <el-switch
                      v-model="scope.row.switch"
                      :active-value="true"
                      :inactive-value="false"
                      @change="handleSwitch(scope.row)"
                    ></el-switch>
                  </template>
                </el-table-column>
                <el-table-column prop="status" width="150">
                  <template #header>
                    <div style="display: flex; align-items: center">
                      <span>{{ $t('knowledgeManage.importStatus') }}</span>
                      <FilterPopover
                        style="margin-left: 5px"
                        :options="knowLegOptions"
                        @applyFilter="filterCurrentStatus"
                      />
                    </div>
                  </template>
                  <template slot-scope="scope">
                    <span>
                      {{ qaImportStatus?.[Number(scope.row.status)] ?? '--' }}
                    </span>
                  </template>
                </el-table-column>
                <el-table-column
                  prop="uploadTime"
                  :label="$t('knowledgeManage.importTime')"
                  width="150"
                ></el-table-column>
                <el-table-column
                  :label="$t('knowledgeManage.operate')"
                  width="200"
                  v-if="hasManagePerm"
                  key="operate"
                >
                  <template slot-scope="scope">
                    <el-button
                      size="mini"
                      round
                      :disabled="
                        [
                          STATUS_PENDING,
                          STATUS_PROCESSING,
                          STATUS_FAILED,
                        ].includes(Number(scope.row.status))
                      "
                      :type="
                        [
                          STATUS_PENDING,
                          STATUS_PROCESSING,
                          STATUS_FAILED,
                        ].includes(Number(scope.row.status))
                          ? 'info'
                          : ''
                      "
                      @click="handleEdit(scope.row)"
                    >
                      {{ $t('common.button.edit') }}
                    </el-button>
                    <el-button
                      size="mini"
                      round
                      @click="handleDel(scope.row)"
                      :disabled="
                        [
                          STATUS_PENDING,
                          STATUS_PROCESSING,
                          STATUS_FAILED,
                        ].includes(Number(scope.row.status))
                      "
                      :type="
                        [
                          STATUS_PENDING,
                          STATUS_PROCESSING,
                          STATUS_FAILED,
                        ].includes(Number(scope.row.status))
                          ? 'info'
                          : ''
                      "
                    >
                      {{ $t('common.button.delete') }}
                    </el-button>
                  </template>
                </el-table-column>
              </el-table>
              <!-- 分页 -->
              <Pagination
                class="pagination table-pagination"
                ref="pagination"
                :listApi="resolvedListApi"
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
        <el-button @click="handleClose">
          {{ $t('common.button.cancel') }}
        </el-button>
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
      :type="batchMetaType"
    />
    <!-- 批量编辑元数据值操作框 -->
    <BatchMetaButton
      ref="BatchMetaButton"
      :selectedCount="selectedTableData.length"
      :type="'qa'"
      :batchMetaType="batchMetaType"
      @showBatchMeta="showBatchMeta"
      @handleBatchDelete="handleBatchDelete"
      @handleMetaCancel="handleMetaCancel"
    />
    <!-- 新建/编辑问答对 -->
    <createQa
      ref="createQa"
      @updateData="updateData"
      :knowledgeId="docQuery.knowledgeId"
    />
    <!-- 文件上传 -->
    <fileUpload
      ref="fileUpload"
      @updateData="updateData"
      :knowledgeId="docQuery.knowledgeId"
    />
    <!-- 导出记录 -->
    <exportRecord ref="exportRecord" />
    <createKnowledge
      ref="createKnowledge"
      :category="QA"
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
import createQa from './createQa.vue';
import fileUpload from './fileUpload.vue';
import exportRecord from './exportRecord.vue';
import { getDocDetail, updateDocMeta } from '@/api/knowledge';
import {
  getQaPairList,
  delQaPair,
  switchQaPair,
  qaDocExport,
  qaTips,
} from '@/api/qaDatabase';
import {
  COMMUNITY_IMPORT_STATUS,
  DROPDOWN_GROUPS,
  QA_STATUS_OPTIONS,
} from '../config';
import {
  INITIAL,
  POWER_TYPE_READ,
  POWER_TYPE_ADMIN,
  POWER_TYPE_EDIT,
  POWER_TYPE_SYSTEM_ADMIN,
  ALL,
  STATUS_FAILED,
  STATUS_FINISHED,
  STATUS_PENDING,
  STATUS_PROCESSING,
  QA,
} from '@/views/knowledge/constants';
import CopyIcon from '@/components/copyIcon.vue';
import createKnowledge from '@/views/knowledge/component/create.vue';

export default {
  name: 'KnowledgeQaDoclist',
  components: {
    CopyIcon,
    Pagination,
    SearchInput,
    mataData,
    batchMetaData,
    BatchMetaButton,
    FilterPopover,
    createQa,
    fileUpload,
    exportRecord,
    createKnowledge,
  },
  beforeRouteLeave(to, from, next) {
    // 仅当返回知识库列表(/knowledge)时标记下次进入需重置;去文件详情等子页返回时保留
    if (to.path === '/knowledge') {
      this.shouldResetOnNextEntry = true;
    }
    next();
  },
  props: {
    knowledgeId: {
      type: String,
      default: '',
    },
    readonly: {
      type: Boolean,
      default: false,
    },
    listApi: {
      type: Function,
      default: null,
    },
  },
  data() {
    return {
      // 从知识库返回列表(/knowledge)后再进入时,需重置搜索/页码;从文件详情等子页返回则保留
      shouldResetOnNextEntry: false,
      title_tips: '',
      showTips: false,
      batchMetaType: 'single',
      knowledgeName: '',
      description: '',
      avatar: '',
      embeddingModel: {},
      loading: false,
      tableLoading: false,
      docQuery: {
        name: '',
        metaType: 'string',
        metaValue: '',
        metaStartTime: '',
        metaEndTime: '',
        knowledgeId: '',
        status: [ALL],
      },
      metaDateRange: null,
      fileList: [],
      tableData: [],
      knowLegOptions: QA_STATUS_OPTIONS,
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
      qaImportStatus: COMMUNITY_IMPORT_STATUS,
      dropdownGroups: DROPDOWN_GROUPS.slice(0, 2),
      QA,
      POWER_TYPE_SYSTEM_ADMIN,
      STATUS_FAILED,
      STATUS_FINISHED,
      STATUS_PENDING,
      STATUS_PROCESSING,
      permissionType: null,
    };
  },
  watch: {
    metaData: {
      handler(val) {
        if (
          val.some(item => !item.metaKey || !item.metaValueType) ||
          !val.length
        ) {
          this.isDisabled = true;
        } else {
          this.isDisabled = false;
        }
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
    resolvedListApi() {
      return this.listApi || getQaPairList;
    },
    effectiveKnowledgeId() {
      return this.knowledgeId || this.$route.params.id;
    },
  },
  mounted() {
    if (!this.readonly) return;
    this.docQuery.knowledgeId = this.effectiveKnowledgeId;
    this.getTableData(this.docQuery);
  },
  activated() {
    if (this.readonly) return;
    if (this.shouldResetOnNextEntry) {
      this.shouldResetOnNextEntry = false;
      this.resetDocQueryState();
    }
    this.docQuery.knowledgeId = this.$route.params.id;
    this.getTableData(this.docQuery);
    getDocDetail({ knowledgeId: this.docQuery.knowledgeId }).then(res => {
      if (res.code === 0) {
        this.knowledgeName = res.data.knowledgeName;
        this.description = res.data.description;
        this.avatar = res.data.avatar;
        this.embeddingModel = res.data.embeddingModel;
        this.permissionType = res.data.permissionType;
      }
    });
  },
  deactivated() {
    if (this.readonly) return;
    this.clearTimer();
  },
  methods: {
    resetDocQueryState() {
      this.docQuery.name = '';
      this.docQuery.metaType = 'string';
      this.docQuery.metaValue = '';
      this.docQuery.metaStartTime = '';
      this.docQuery.metaEndTime = '';
      this.docQuery.status = [ALL];
      this.metaDateRange = null;
      this.$refs.searchInput?.clearValue();
      this.$refs.searchInputMeta?.clearValue();
      if (this.$refs.pagination) {
        this.$refs.pagination.pageNo = 1;
      }
    },
    showEdit() {
      this.$refs.createKnowledge.showDialog({
        knowledgeId: this.docQuery.knowledgeId,
        avatar: this.avatar,
        name: this.knowledgeName,
        description: this.description,
        embeddingModelInfo: this.embeddingModel,
      });
    },
    handleCommand(command) {
      const actions = {
        exportData: this.exportData,
        exportRecord: this.exportRecord,
        createQaPair: this.handleCreateQaPair,
        fileUpload: this.handleUpload,
      };
      (actions[command] || this.handleUpload)();
    },
    exportRecord() {
      this.$refs.exportRecord.showDialog(this.docQuery.knowledgeId);
    },
    updateData(type = '') {
      if (type === '') {
        this.getTableData(this.docQuery);
      } else {
        this.startTimer();
      }
    },
    exportData() {
      if (!this.docQuery.knowledgeId) {
        this.$message.warning(this.$t('common.noData'));
        return;
      }
      if (this.loading) return;
      const params = {
        knowledgeId: this.docQuery.knowledgeId,
      };
      this.loading = true;
      qaDocExport(params)
        .then(res => {
          if (res.code === 0) {
            this.$message.success(this.$t('common.message.success'));
            const data = res.data || {};
            const url = data.fileUrl || data.downloadUrl;
            if (url) {
              window.open(url, '_blank');
            } else if (data.recordCreated) {
              this.exportRecord();
            }
          }
        })
        .catch(() => {})
        .finally(() => {
          this.loading = false;
        });
    },
    handleCreateQaPair() {
      this.$refs.createQa.showDialog();
    },
    handleEditMetaData(row) {
      this.$refs.batchMetaData.showDialog(row);
      this.batchMetaType = 'single';
      this.selectedTableData = [row];
      this.selectedDocIds = [row.qaPairId];
    },
    handleSwitch(row) {
      switchQaPair({ qaPairId: row.qaPairId, switch: row.switch }).then(res => {
        if (res.code === 0) {
          this.$message.success(this.$t('common.message.success'));
          this.getTableData(this.docQuery);
        }
      });
    },
    handleMetaCancel() {
      this.selectedTableData = [];
      this.selectedDocIds = [];
      // 取消所有表格数据的选中状态
      this.$nextTick(() => {
        const table = this.$refs.dataTable;
        if (table) {
          table.clearSelection();
        }
      });
    },
    reLoadDocList() {
      this.getTableData(this.docQuery);
      this.selectedTableData = [];
      this.selectedDocIds = [];

      // 取消所有表格数据的选中状态
      this.$nextTick(() => {
        const table = this.$refs.dataTable;
        if (table) {
          table.clearSelection();
        }
      });
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
      this.batchMetaType = 'multiple';
      this.selectedDocIds = val.map(item => item.qaPairId);
    },
    getMetaDataText(list) {
      if (!list || !Array.isArray(list) || list.length === 0) {
        return '';
      }
      return list.map(item => item.metaKey).join(', ');
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
      this.timer = setTimeout(() => {
        this.getTableData(this.docQuery);
        this.refreshCount++;
        this.startTimer();
      }, delay);
    },
    clearTimer() {
      if (this.timer) {
        clearInterval(this.timer);
        this.timer = null;
      }
    },
    reload() {
      this.getTableData(this.docQuery);
    },
    handleSearch(val) {
      this.docQuery.name = val;
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
    async handleDelete(QAPairIdList) {
      this.loading = true;
      try {
        let res = await delQaPair({
          QAPairIdList,
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
      )
        .then(() => {
          this.handleDelete([data.qaPairId]);
        })
        .catch(() => {});
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
      )
        .then(() => {
          this.handleDelete(this.selectedDocIds);
        })
        .catch(() => {});
    },
    async getTableData(data) {
      this.tableLoading = true;
      this.tableData = await this.$refs['pagination'].getTableData(data);
      this.tableLoading = false;
      this.getTips();
    },
    getTips() {
      if (this.readonly) return;
      qaTips({ knowledgeId: this.docQuery.knowledgeId }).then(res => {
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
    handleEdit(row) {
      this.$refs.createQa.showDialog(row);
    },
    handleUpload() {
      this.$refs.fileUpload.showDialog();
    },
    refreshData(data) {
      this.tableData = data;
    },
    filterCurrentStatus(data) {
      this.docQuery.status = data;
      this.getTableData({ ...this.docQuery, pageNo: 1 });
    },
  },
};
</script>
<style lang="scss" scoped>
.keyword-tags {
  align-items: center;
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
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
