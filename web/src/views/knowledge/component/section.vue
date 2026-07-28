<template>
  <div
    class="section page-wrapper"
    v-loading="loading.itemStatus"
    :class="{ 'disable-clicks': obj.disable === 'true' }"
  >
    <div class="title">
      <i
        class="el-icon-arrow-left"
        @click="goBack"
        style="margin-right: 20px; font-size: 20px; cursor: pointer"
      ></i>
      {{ obj.name }}
    </div>

    <el-descriptions :column="3" :size="''" border class="margin-top" title="">
      <el-descriptions-item :label="$t('knowledgeManage.fileName')">
        {{ res.fileName }}
      </el-descriptions-item>
      <el-descriptions-item :label="$t('knowledgeManage.splitNum')">
        {{ res.segmentTotalNum }}
      </el-descriptions-item>
      <el-descriptions-item :label="$t('knowledgeManage.importTime')">
        {{ res.uploadTime }}
      </el-descriptions-item>
      <el-descriptions-item :label="$t('knowledgeManage.chunkType')">
        {{
          Number(res.segmentType) === 0
            ? $t('knowledgeManage.autoChunk')
            : $t('knowledgeManage.autoConfigChunk')
        }}
      </el-descriptions-item>
      <el-descriptions-item :label="$t('knowledgeManage.setMaxLength')">
        {{ String(res.maxSegmentSize) }}
      </el-descriptions-item>
      <el-descriptions-item :label="$t('knowledgeManage.markSplit')">
        {{ String(res.splitter).replace(/\n/g, '\\n') }}
      </el-descriptions-item>
      <el-descriptions-item :label="$t('knowledgeManage.metaData')">
        <template v-if="metaDataList && metaDataList.length > 0">
          <span
            v-for="(item, index) in metaDataList.slice(0, 3)"
            :key="index"
            class="metaItem"
          >
            {{ item.metaKey }}:
            {{
              item.metaValueType === 'time'
                ? formatTimestamp(item.metaValue)
                : item.metaValue
            }}
          </span>
          <el-tooltip
            v-if="metaDataList.length > 3"
            :content="filterData(metaDataList.slice(3))"
            placement="bottom"
          >
            <span class="metaItem">...</span>
          </el-tooltip>
        </template>
        <span v-else>{{ $t('knowledgeManage.zeroData') }}</span>
        <span
          v-if="
            metaDataList &&
            [
              POWER_TYPE_EDIT,
              POWER_TYPE_ADMIN,
              POWER_TYPE_SYSTEM_ADMIN,
            ].includes(permissionType) &&
            obj.disable !== 'true' &&
            obj.readonly !== 'true'
          "
          class="el-icon-edit-outline editIcon"
          @click="showDatabase(metaDataList || [])"
        ></span>
      </el-descriptions-item>
      <el-descriptions-item :label="$t('knowledgeManage.metaDataRules')">
        <template v-if="metaRuleList && metaRuleList.length > 0">
          <span
            v-for="(item, index) in metaRuleList.slice(0, 3)"
            :key="index"
            class="metaItem"
          >
            {{ item.metaKey }}: {{ item.metaRule }}
            <span v-if="index < metaRuleList.slice(0, 3).length - 1"></span>
          </span>
          <el-tooltip
            v-if="metaRuleList.length > 3"
            :content="filterRule(metaRuleList.slice(3))"
            placement="bottom"
          >
            <span class="metaItem">...</span>
          </el-tooltip>
        </template>
        <span v-else>{{ $t('knowledgeManage.zeroData') }}</span>
      </el-descriptions-item>
      <el-descriptions-item :label="$t('knowledgeManage.batchAddSplit')">
        <span>{{ res.segmentImportStatus || '-' }}</span>
      </el-descriptions-item>
      <el-descriptions-item :label="$t('knowledgeManage.parsingMethod')">
        <div class="keyword-tags">
          <template v-if="res.docAnalyzerText?.length">
            <template v-for="(item, index) in res.docAnalyzerText">
              <span>
                {{ item.text }}
              </span>
              <el-tooltip
                v-if="item.displayName"
                effect="light"
                placement="top"
                popper-class="custom-tooltip"
                style="pointer-events: auto !important"
              >
                <div slot="content" class="tooltip-content">
                  <img
                    :src="convertModelIcon(item?.avatar?.path)"
                    class="model-img"
                  />
                  <span class="model-name">{{ item.displayName }}</span>
                  <template v-if="item.tags?.length">
                    <el-tag
                      v-for="(tag, tagIndex) in item.tags"
                      :key="'tag-' + tagIndex"
                      class="keyword-tag"
                      color="#E6F0FF"
                      size="small"
                    >
                      {{ tag.text }}
                    </el-tag>
                  </template>
                </div>
                <i class="el-icon-question question-icon" />
              </el-tooltip>
              <span v-if="index < res.docAnalyzerText.length - 1">;</span>
            </template>
          </template>
          <span v-else>-</span>
        </div>
      </el-descriptions-item>
    </el-descriptions>

    <div v-if="obj.disable !== 'true'" class="btn">
      <search-input
        style="visibility: hidden"
        ref="searchInput"
        :placeholder="$t('knowledgeManage.segmentPlaceholder')"
        @handleSearch="handleSearch"
      />
      <div>
        <el-button
          v-if="
            [
              POWER_TYPE_EDIT,
              POWER_TYPE_ADMIN,
              POWER_TYPE_SYSTEM_ADMIN,
            ].includes(permissionType)
          "
          :loading="loading.start"
          size="mini"
          type="primary"
          @click="createChunk(false)"
        >
          {{ $t('knowledgeManage.create.createChunk') }}
        </el-button>
        <el-button
          v-if="
            [
              POWER_TYPE_EDIT,
              POWER_TYPE_ADMIN,
              POWER_TYPE_SYSTEM_ADMIN,
            ].includes(permissionType)
          "
          :loading="loading.start"
          size="mini"
          type="primary"
          @click="handleStatus('start')"
        >
          {{ $t('knowledgeManage.allRun') }}
        </el-button>
        <el-button
          v-if="
            [
              POWER_TYPE_EDIT,
              POWER_TYPE_ADMIN,
              POWER_TYPE_SYSTEM_ADMIN,
            ].includes(permissionType)
          "
          :loading="loading.stop"
          size="mini"
          type="primary"
          @click="handleStatus('stop')"
        >
          {{ $t('knowledgeManage.allStop') }}
        </el-button>
      </div>
    </div>

    <div class="container">
      <!-- 左侧：文件预览面板（仅在非禁用状态且 canPreview 为 true 时显示） -->
      <div
        v-if="obj.disable !== 'true' && res?.canPreview === true"
        class="section-preview-panel"
      >
        <file-preview-drawer
          :blob="previewBlob"
          :file-name="previewFileName"
          :loading="previewLoading"
          :panel-style="{ width: '100%', height: '100%' }"
          :showClose="false"
          :visible="true"
        />
      </div>
      <!-- 不可预览时的兜底提示 -->
      <div
        v-else-if="obj.disable !== 'true' && res?.canPreview === false"
        class="section-preview-panel preview-fail-panel"
      >
        <el-empty
          :description="
            res?.previewFailReason ||
            '' ||
            $t('knowledgeManage.previewUnavailable')
          "
        />
      </div>

      <!-- 右侧：分段列表 -->
      <div
        :class="{ 'full-width': obj.disable === 'true' }"
        class="section-content-panel"
      >
        <div class="card">
          <template v-if="res.contentList.length > 0 && obj.disable !== 'true'">
            <el-card
              v-for="(item, index) in res.contentList"
              :key="index"
              class="box-card segment-card"
            >
              <div slot="header" class="clearfix">
                <span>
                  {{ $t('knowledgeManage.split') + ':' + item.contentNum }}
                  <span class="segment-type">
                    #{{
                      item.isParent
                        ? $t('knowledgeManage.config.parentSonSegment')
                        : $t('knowledgeManage.config.commonSegment')
                    }}
                  </span>
                  <span class="segment-length" v-if="!item.isParent">
                    #{{ item.content.length
                    }}{{ $t('knowledgeManage.character') }}
                  </span>
                  <span class="segment-child" v-if="item.childNum">
                    #{{ item.childNum || 0
                    }}{{ $t('knowledgeManage.childSegmentUnit') }}
                  </span>
                </span>
                <div>
                  <el-switch
                    style="padding: 3px 0"
                    v-model="item.available"
                    active-color="var(--color)"
                    v-if="
                      [
                        POWER_TYPE_EDIT,
                        POWER_TYPE_ADMIN,
                        POWER_TYPE_SYSTEM_ADMIN,
                      ].includes(permissionType)
                    "
                    @change="handleStatusChange(item, index)"
                  ></el-switch>
                  <el-dropdown
                    @command="handleCommand"
                    placement="bottom"
                    v-if="
                      [
                        POWER_TYPE_EDIT,
                        POWER_TYPE_ADMIN,
                        POWER_TYPE_SYSTEM_ADMIN,
                      ].includes(permissionType)
                    "
                  >
                    <span class="el-dropdown-link">
                      <i class="el-icon-more more"></i>
                    </span>
                    <el-dropdown-menu slot="dropdown">
                      <el-dropdown-item
                        class="card-delete"
                        :command="{ type: 'delete', item }"
                      >
                        <i class="el-icon-delete card-opera-icon" />
                        {{ $t('common.button.delete') }}
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </el-dropdown>
                </div>
              </div>
              <div
                class="text item"
                v-html="Md2Img(item.content)"
                @click="handleClick(item, index)"
              ></div>
              <div
                class="tagList"
                v-if="
                  [
                    POWER_TYPE_EDIT,
                    POWER_TYPE_ADMIN,
                    POWER_TYPE_SYSTEM_ADMIN,
                  ].includes(permissionType)
                "
              >
                <span class="el-icon-price-tag icon-tag"></span>
                <span
                  :class="['smartDate', 'tagList']"
                  @click.stop="addTag(item.labels, item.contentId)"
                  v-if="item.labels.length === 0"
                >
                  {{ $t('keyword.create') }}
                </span>
                <span
                  class="tagList-item"
                  @click.stop="addTag(item.labels, item.contentId)"
                  v-else
                >
                  {{ formattedTagNames(item.labels) }}
                </span>
              </div>
            </el-card>
          </template>
          <el-empty
            v-else
            :description="$t('knowledgeManage.noData')"
          ></el-empty>
        </div>

        <div
          v-if="obj.disable !== 'true'"
          class="list-common"
          style="text-align: right; flex-shrink: 0; padding: 10px 0"
        >
          <el-pagination
            :current-page="page.pageNo"
            :page-size="page.pageSize"
            :page-sizes="page.pageSizeList"
            :total="page.total"
            background
            layout="total, prev, pager, next, jumper"
            @current-change="handleCurrentChange"
            @size-change="handleSizeChange"
          ></el-pagination>
        </div>
      </div>
    </div>

    <el-dialog
      v-if="dialogVisible"
      :title="$t('knowledgeManage.detailView')"
      :visible.sync="dialogVisible"
      width="60%"
      :show-close="false"
      v-loading="loading.dialog"
      class="section-dialog"
      :close-on-click-modal="false"
    >
      <div slot="title">
        <span style="font-size: 16px">
          {{ $t('knowledgeManage.detailView') }}
        </span>
        <el-switch
          @change="handleDetailStatusChange"
          style="float: right; padding: 3px 0"
          v-model="cardObj[0].available"
          active-color="var(--color)"
          v-if="
            [
              POWER_TYPE_EDIT,
              POWER_TYPE_ADMIN,
              POWER_TYPE_SYSTEM_ADMIN,
            ].includes(permissionType)
          "
        ></el-switch>
      </div>
      <div class="dialog-content">
        <el-table
          :data="cardObj"
          border
          style="width: 100%"
          :header-cell-style="{
            background: '#F9F9F9',
            color: '#999999',
          }"
        >
          <el-table-column
            prop="content"
            align="center"
            :render-header="renderHeader"
          >
            <template slot-scope="scope">
              <uploadImgMd
                v-if="isMultiModal"
                :placeholder="
                  $t('knowledgeManage.create.chunkContentPlaceholder')
                "
                v-model="scope.row.content"
                :permission-type="permissionType"
                :knowledgeId="obj.knowledgeId"
              ></uploadImgMd>
              <el-input
                v-else
                type="textarea"
                v-model="scope.row.content"
                :autosize="{ minRows: 3, maxRows: 5 }"
                class="full-width-textarea"
                :disabled="[POWER_TYPE_READ].includes(permissionType)"
              ></el-input>
              <div
                v-if="
                  cardObj[0]['isParent'] &&
                  [
                    POWER_TYPE_EDIT,
                    POWER_TYPE_ADMIN,
                    POWER_TYPE_SYSTEM_ADMIN,
                  ].includes(permissionType)
                "
                style="
                  display: flex;
                  justify-content: flex-end;
                  padding: 10px 0;
                "
              >
                <el-button
                  type="primary"
                  @click="handleSubmit"
                  :loading="submitLoading"
                >
                  {{ $t('knowledgeManage.create.saveAndReparseChild') }}
                </el-button>
              </div>
              <div
                class="segment-list"
                v-if="scope.row.childContent.length > 0"
              >
                <el-collapse v-model="activeNames" class="section-collapse">
                  <el-collapse-item
                    v-for="(segment, index) in scope.row.childContent"
                    :key="index"
                    :name="index"
                    class="segment-collapse-item"
                  >
                    <template slot="title">
                      <span class="segment-badge">C-{{ index + 1 }}</span>
                      <div class="segment-actions">
                        <span
                          v-if="
                            !editingSegments[
                              `${scope.row.contentId}-${index}`
                            ] &&
                            [
                              POWER_TYPE_EDIT,
                              POWER_TYPE_ADMIN,
                              POWER_TYPE_SYSTEM_ADMIN,
                            ].includes(permissionType)
                          "
                          class="action-btn edit-btn"
                          @click.stop="editSegment(scope.row, index)"
                        >
                          <i class="el-icon-edit-outline"></i>
                          {{ $t('common.button.edit') }}
                        </span>
                        <span
                          v-if="
                            !editingSegments[
                              `${scope.row.contentId}-${index}`
                            ] &&
                            [
                              POWER_TYPE_EDIT,
                              POWER_TYPE_ADMIN,
                              POWER_TYPE_SYSTEM_ADMIN,
                            ].includes(permissionType)
                          "
                          class="action-btn delete-btn"
                          @click.stop="deleteSegment(scope.row, index)"
                        >
                          <i class="el-icon-delete"></i>
                          {{ $t('common.button.delete') }}
                        </span>
                        <span
                          v-if="
                            editingSegments[`${scope.row.contentId}-${index}`]
                          "
                          class="action-btn save-btn"
                          @click.stop="confirmEdit(scope.row, index)"
                        >
                          <i class="el-icon-check"></i>
                          {{ $t('common.button.save') }}
                        </span>
                        <span
                          v-if="
                            editingSegments[`${scope.row.contentId}-${index}`]
                          "
                          class="action-btn cancel-btn"
                          @click.stop="cancelEdit(scope.row, index)"
                        >
                          <i class="el-icon-close"></i>
                          {{ $t('common.button.cancel') }}
                        </span>
                      </div>
                    </template>
                    <div class="segment-content">
                      <div
                        v-if="
                          !editingSegments[`${scope.row.contentId}-${index}`]
                        "
                        class="content-display"
                        v-html="Md2Img(segment.content)"
                      ></div>
                      <div v-else class="content-edit">
                        <uploadImgMd
                          v-if="isMultiModal"
                          :placeholder="
                            $t('knowledgeManage.create.chunkContentPlaceholder')
                          "
                          v-model="segment.content"
                          :permission-type="permissionType"
                          :knowledgeId="obj.knowledgeId"
                          @input="
                            newContent =>
                              (editingContent[
                                `${scope.row.contentId}-${index}`
                              ] = newContent)
                          "
                        ></uploadImgMd>
                        <el-input
                          v-else
                          v-model="
                            editingContent[`${scope.row.contentId}-${index}`]
                          "
                          type="textarea"
                          :rows="3"
                          :placeholder="
                            $t('knowledgeManage.create.chunkContentPlaceholder')
                          "
                          class="edit-input"
                        />
                      </div>
                    </div>
                  </el-collapse-item>
                </el-collapse>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <span slot="footer" class="dialog-footer">
        <el-button
          type="primary"
          @click="handleSubmit"
          :loading="submitLoading"
          v-if="!cardObj[0]['isParent']"
        >
          {{ $t('common.confirm.confirm') }}
        </el-button>
        <el-button
          type="primary"
          @click="createChunk(true)"
          v-if="
            cardObj[0]['isParent'] &&
            [
              POWER_TYPE_EDIT,
              POWER_TYPE_ADMIN,
              POWER_TYPE_SYSTEM_ADMIN,
            ].includes(permissionType)
          "
          :disabled="submitLoading"
        >
          {{ $t('knowledgeManage.create.addChildChunk') }}
        </el-button>
        <el-button
          type="primary"
          @click="handleClose"
          :disabled="submitLoading"
        >
          {{ $t('knowledgeManage.close') }}
        </el-button>
      </span>
    </el-dialog>
    <dataBaseDialog
      v-if="obj.readonly !== 'true'"
      ref="dataBase"
      @updateData="updateData"
      :knowledgeId="obj.knowledgeId"
      :name="obj.knowledgeName"
    />
    <tagDialog
      ref="tagDialog"
      type="section"
      :title="title"
      :currentList="currentList"
      @sendList="sendList"
    />
    <createChunk
      ref="createChunk"
      @updateDataBatch="updateDataBatch"
      @updateData="updateData"
      :parentId="cardObj[0]['contentId']"
      @updateChildData="updateChildData"
    />
  </div>
</template>
<script>
import {
  getSectionList,
  setSectionStatus,
  sectionLabels,
  delSegment,
  editSegment,
  getSegmentChild,
  delSegmentChild,
  updateSegmentChild,
} from '@/api/knowledge';
import { getAdminKnowledgeFileDetail } from '@/api/adminCenter';
import dataBaseDialog from './dataBaseDialog';
import tagDialog from './tagDialog.vue';
import createChunk from './chunk/createChunk.vue';
import FilePreviewDrawer from '@/views/generalAgent/components/FilePreviewDrawer.vue';
import { Md2Img, avatarSrc, getModelDefaultIcon } from '@/utils/util';
import {
  INITIAL,
  POWER_TYPE_READ,
  POWER_TYPE_EDIT,
  POWER_TYPE_ADMIN,
  POWER_TYPE_SYSTEM_ADMIN,
  MULTIMODAL,
} from '@/views/knowledge/constants';
import SearchInput from '@/components/searchInput.vue';
import uploadImgMd from '@/components/uploadImgMd.vue';

export default {
  name: 'KnowledgeSection',
  components: {
    SearchInput,
    dataBaseDialog,
    tagDialog,
    createChunk,
    uploadImgMd,
    FilePreviewDrawer,
  },
  data() {
    return {
      submitLoading: false,
      oldContent: '',
      dialogVisible: false,
      editingSegments: {},
      editingContent: {},
      obj: {},
      cardObj: [
        {
          available: false,
          content: '',
          childContent: [],
          contentId: '',
          len: 20,
        },
      ],
      value: true,
      activeStatus: false,
      activeNames: [],
      loading: {
        start: false,
        stop: false,
        itemStatus: false,
        dialog: false,
      },
      res: {
        contentList: [],
      },
      metaDataList: [],
      metaRuleList: [],
      currentList: [],
      contentId: '',
      timer: null,
      refreshCount: 0,
      // 分页相关
      page: {
        pageNo: 1,
        pageSize: 10,
        pageSizeList: [10, 15, 20, 50],
        total: 0,
      },
      keyword: '',
      // 文件预览相关
      previewLoading: false,
      previewFileName: '',
      previewBlob: null,
      INITIAL,
      POWER_TYPE_READ,
      POWER_TYPE_EDIT,
      POWER_TYPE_ADMIN,
      POWER_TYPE_SYSTEM_ADMIN,
      permissionType: null,
    };
  },
  computed: {
    isMultiModal() {
      return Number(this.obj.category) === MULTIMODAL;
    },
    title() {
      return this.$t('knowledgeManage.createKeyword');
    },
  },
  created() {
    this.obj = this.$route.query;
    this.permissionType = Number(this.$route.query.permissionType);
    this.getList();
  },
  beforeDestroy() {
    this.clearTimer();
  },
  methods: {
    Md2Img,
    goBack() {
      this.obj.readonly === 'true'
        ? this.$router.replace(
            `/adminCenter/knowledge/detail?knowledgeId=${this.obj.knowledgeId}`,
          )
        : this.$router.replace(`/knowledge/doclist/${this.obj.knowledgeId}`);
    },
    convertModelIcon(iconPath) {
      return iconPath ? avatarSrc(iconPath) : getModelDefaultIcon();
    },
    handleSearch(val) {
      // keyword 变化时重置页码
      if (this.keyword !== val) {
        this.keyword = val;
        this.page.pageNo = 1;
        this.getList();
      }
    },
    createChunk(isChildChunk) {
      this.$refs.createChunk.showDialog(
        this.obj.id,
        this.obj.knowledgeId,
        isChildChunk,
        this.obj.category,
      );
    },
    updateChildData() {
      setTimeout(() => {
        this.handleParse();
      }, 1000);
    },
    formatScore(score) {
      if (typeof score !== 'number') {
        return '0.00000';
      }
      return score.toFixed(5);
    },
    editSegment(row, index) {
      const key = `${row.contentId}-${index}`;
      this.$set(this.editingSegments, key, true);
      this.$set(this.editingContent, key, row.childContent[index].content);

      this.$nextTick(() => {
        if (!this.activeNames.includes(index)) {
          this.activeNames.push(index);
        }
      });
    },
    cancelEdit(row, index) {
      const key = `${row.contentId}-${index}`;
      this.$set(this.editingSegments, key, false);
      this.$delete(this.editingContent, key);
    },
    confirmEdit(row, index) {
      const key = `${row.contentId}-${index}`;
      const newContent = this.editingContent[key];

      if (!newContent || newContent.trim() === '') {
        this.$message.warning(this.$t('knowledgeManage.contentEmpty'));
        return;
      }
      updateSegmentChild({
        childChunk: {
          content: newContent.trim(),
          chunkNo: row['childContent'][index].childNum,
        },
        docId: this.obj.id,
        parentChunkNo: row.contentNum,
        parentId: row.contentId,
      })
        .then(res => {
          if (res.code === 0) {
            this.$message.success(this.$t('knowledgeManage.updateSuccess'));
            this.handleParse();
            this.$set(this.editingSegments, key, false);
            this.$delete(this.editingContent, key);
          } else {
            this.$message.error(this.$t('knowledgeManage.updateFail'));
          }
        })
        .catch(() => {
          this.$message.error(this.$t('knowledgeManage.updateFail'));
        });
    },
    handleParse() {
      getSegmentChild({
        contentId: this.cardObj[0]['contentId'],
        docId: this.obj.id,
      })
        .then(res => {
          if (res.code === 0) {
            this.cardObj[0].childContent = res.data.contentList || [];
            this.activeNames = this.cardObj[0].childContent.map(
              (_, index) => index,
            );
          }
        })
        .catch(() => {});
    },
    deleteSegment(row, index) {
      this.$confirm(
        this.$t('knowledgeManage.create.deleteChildChunkTips'),
        this.$t('common.confirm.title'),
        {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          type: 'warning',
        },
      ).then(() => {
        delSegmentChild({
          docId: this.obj.id,
          parentId: row['childContent'][index].parentId,
          parentChunkNo: row.contentNum,
          childChunkNoList: [row['childContent'][index].childNum],
        })
          .then(res => {
            if (res.code === 0) {
              this.$message.success(
                this.$t('knowledgeManage.create.deleteSuccess'),
              );
              this.handleParse();
            }
          })
          .catch(() => {
            this.$message.error(this.$t('knowledgeManage.deleteFail'));
          });
      });
    },
    updateDataBatch() {
      this.startTimer();
    },
    startTimer() {
      this.clearTimer();
      if (this.refreshCount >= 2) {
        return;
      }
      const delay = this.refreshCount === 0 ? 1000 : 3000;
      this.timer = setTimeout(() => {
        this.getList();
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
    handleSubmit() {
      const hasChanges = this.oldContent !== this.cardObj[0]['content'];

      if (!hasChanges) {
        this.$message.warning(this.$t('knowledgeManage.noChange'));
        return false;
      }

      this.submitLoading = true;
      editSegment({
        content: this.cardObj[0]['content'],
        contentId: this.cardObj[0]['contentId'],
        docId: this.obj.id,
      })
        .then(res => {
          if (res.code === 0) {
            this.$message.success(this.$t('knowledgeManage.operateSuccess'));
            this.dialogVisible = false;
            this.getList();
          }
        })
        .finally(() => {
          this.submitLoading = false;
        });
    },
    handleCommand(value) {
      const { type, item } = value || {};
      switch (type) {
        case 'delete':
          this.delSection(item);
          break;
      }
    },
    delSection(item) {
      delSegment({ contentId: item.contentId, docId: this.obj.id })
        .then(res => {
          if (res.code === 0) {
            this.$message.success(
              this.$t('knowledgeManage.create.deleteSuccess'),
            );
            this.getList();
          }
        })
        .catch(() => {});
    },
    sendList(data) {
      const labels = data.map(item => item.tagName);
      sectionLabels({ contentId: this.contentId, docId: this.obj.id, labels })
        .then(res => {
          if (res.code === 0) {
            this.getList();
            this.$refs.tagDialog.handleClose();
          }
        })
        .catch(err => {});
    },
    addTag(data, id) {
      if (data.length > 0) {
        this.currentList = data.map(item => ({
          tagName: item,
          checked: false,
          showDel: false,
          showIpt: false,
        }));
      } else {
        this.currentList = [];
      }
      this.contentId = id;
      this.$refs.tagDialog.showDialog();
    },
    formattedTagNames(data) {
      let tags = '';
      if (!Array.isArray(data) || data.length === 0) {
        return '';
      }
      if (data.length > 3) {
        tags = data.slice(0, 3).join(', ') + (data.length > 3 ? '...' : '');
      } else {
        tags = data.join(', ');
      }
      return tags;
    },
    updateData() {
      this.getList();
    },
    showDatabase(data) {
      this.$refs.dataBase.showDialog(data, this.obj.id);
    },
    filterData(data) {
      return data
        .map(item => {
          let value = item.metaValue;
          if (item.metaValueType === 'time') {
            value = this.formatTimestamp(value);
          }
          return `${item.metaKey}:${value}`;
        })
        .join(', ');
    },
    formatTimestamp(timestamp) {
      if (timestamp === '') return '';
      const date = new Date(Number(timestamp));
      const year = date.getFullYear();
      const month = String(date.getMonth() + 1).padStart(2, '0');
      const day = String(date.getDate()).padStart(2, '0');
      const hours = String(date.getHours()).padStart(2, '0');
      const minutes = String(date.getMinutes()).padStart(2, '0');
      const seconds = String(date.getSeconds()).padStart(2, '0');
      return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
    },
    filterRule(rule) {
      return rule.map(item => `${item.metaKey}:${item.metaRule}`).join(', ');
    },
    getList() {
      this.loading.itemStatus = true;
      // 预览文件只在 created 首次加载时下载一次,后续 getList 不再重复下载
      this.previewLoading = !this.previewFileName;

      const params = {
        keyword: this.keyword,
        docId: this.obj.id,
        pageNo: this.page.pageNo,
        pageSize: this.page.pageSize,
      };
      const request =
        this.obj.readonly === 'true'
          ? getAdminKnowledgeFileDetail({
              ...params,
              knowledgeId: this.obj.knowledgeId,
            })
          : getSectionList(params);
      request
        .then(res => {
          this.res = res.data;
          this.page.total = res.data.segmentTotalNum || 0;
          this.metaRuleList = res.data.metaDataList.filter(
            item => item.metaRule,
          );
          this.metaDataList = res.data.metaDataList;

          if (
            this.previewLoading &&
            res.data?.canPreview === true &&
            res.data?.downloadUrl
          ) {
            const fileName = this.obj.name;
            const hasExtension = fileName.includes('.');
            this.previewFileName = hasExtension ? fileName : `${fileName}.url`;
            this.downloadPreviewFile(res.data.downloadUrl);
          }
        })
        .catch(() => {
          this.$message.error(this.$t('knowledgeManage.getListFail'));
        })
        .finally(() => {
          this.loading.itemStatus = false;
        });
    },
    // 异步下载预览文件,不阻塞列表加载
    async downloadPreviewFile(url) {
      try {
        const response = await fetch(url);
        if (response.ok) {
          this.previewBlob = await response.blob();
        }
      } catch (e) {
        console.error('文件预览下载失败:', e);
      } finally {
        this.previewLoading = false;
      }
    },
    handleCurrentChange(val) {
      this.page.pageNo = val;
      this.getList();
    },
    handleSizeChange(val) {
      this.page.pageSize = val;
      this.page.pageNo = 1;
      this.getList();
    },
    handleClick(item, index) {
      this.dialogVisible = true;
      this.oldContent = item.content;
      const obj = structuredClone(item);
      this.$nextTick(() => {
        this.$set(obj, 'childContent', []);
        this.cardObj = [obj];
        if (this.cardObj[0].isParent) {
          this.handleParse();
        }
        this.activeStatus = obj.available;
        this.activeNames = [];
      });
    },
    handleDetailStatusChange(val) {
      this.loading.dialog = true;
      setSectionStatus({
        docId: this.obj.id,
        contentStatus: String(val),
        contentId: this.cardObj[0].contentId,
        all: false,
      })
        .then(res => {
          this.loading.dialog = false;
          if (res.code === 0) {
            this.$message.success(this.$t('knowledgeManage.operateSuccess'));
          } else {
            this.cardObj[0].available = !this.cardObj[0].available;
          }
        })
        .catch(() => {
          this.loading.dialog = false;
          this.cardObj[0].contentStatus = !this.cardObj[0].contentStatus;
        });
    },
    handleStatusChange(item, index) {
      this.loading.itemStatus = true;
      setSectionStatus({
        docId: this.obj.id,
        contentStatus: String(item.available),
        contentId: item.contentId,
        all: false,
      })
        .then(res => {
          this.loading.itemStatus = false;
          if (res.code === 0) {
            this.$message.success(this.$t('knowledgeManage.operateSuccess'));
            this.getList();
          } else {
            this.res.contentList[index].available =
              !this.res.contentList[index].available;
          }
        })
        .catch(() => {
          this.res[index].contentStatus = !this.res[index].contentStatus;
          this.loading.itemStatus = false;
        });
    },
    handleStatus(type) {
      this.loading.itemStatus = true;
      setSectionStatus({
        docId: this.obj.id,
        contentStatus: type === 'start' ? 'true' : 'false',
        contentId: '',
        all: true,
      })
        .then(res => {
          if (res.code === 0) {
            this.$message.success(this.$t('knowledgeManage.operateSuccess'));
            this.getList();
          }
        })
        .finally(() => {
          this.loading.itemStatus = false;
        });
    },
    renderHeader(h, { column, $index }) {
      const columnHtml =
        this.$t('knowledgeManage.section') +
        this.cardObj[0].contentNum +
        this.$t('knowledgeManage.length') +
        ' :' +
        this.cardObj[0].content.length +
        this.$t('knowledgeManage.character');
      return h('span', {
        domProps: {
          innerHTML: columnHtml,
        },
      });
    },
    handleClose() {
      this.dialogVisible = false;
      if (this.cardObj[0].available === this.activeStatus) return;
      this.getList();
    },
  },
};
</script>
<style lang="scss">
@import '@/style/customTooltip.scss';
.disable-clicks * {
  pointer-events: none;
}

.disable-clicks .title .el-icon-arrow-left {
  pointer-events: auto;
}

.dialog-content {
  max-height: 55vh !important;
  overflow-y: auto;
}

.segment-list {
  margin-top: 10px;

  .section-collapse {
    background-color: #f7f8fa;
    border-radius: 6px;
    border: 1px solid $color;
    overflow: hidden;

    ::v-deep .el-collapse {
      border: none;
      border-radius: 6px;
    }

    ::v-deep .el-collapse-item__header {
      background-color: #f7f8fa;
      border-bottom: 1px solid #e4e7ed;
      padding: 12px 20px;
      font-weight: normal;
      border-left: none;
      border-right: none;
      border-top: none;
      display: flex !important;
      align-items: center !important;
      justify-content: space-between !important;
      width: 100%;
      position: relative;

      &:hover {
        background-color: #f0f2f5;
      }
    }

    ::v-deep .el-collapse-item__content {
      padding: 15px 20px;
      background-color: #fff;
      border-bottom: 1px solid #e4e7ed;
      border-left: none;
      border-right: none;
      border-top: none;
      font-size: 14px;
      color: #333;
      line-height: 1.5;
      text-align: left;
      word-wrap: break-word;
      word-break: break-all;
      overflow-wrap: break-word;

      .segment-action {
        color: #999;
        font-size: 12px;
        margin-left: 8px;
      }

      .auto-save {
        color: #666;
        font-size: 12px;
        margin-left: 8px;
        font-style: italic;
      }
    }

    ::v-deep .el-collapse-item__header .el-collapse-item__arrow,
    .el-collapse-item__arrow,
    [class*='el-collapse-item__arrow'] {
      display: none !important;
    }

    ::v-deep .el-collapse-item:last-child .el-collapse-item__content {
      border-bottom: none;
    }

    ::v-deep .el-collapse-item__header::after {
      display: none !important;
    }

    .segment-badge {
      color: $color;
      font-size: 12px;
      min-width: 40px;
      text-align: center;
      font-weight: 500;
      margin-right: 120px;
    }

    .segment-actions {
      display: flex;
      gap: 8px;
      align-items: center;
      flex: 1;
      justify-content: flex-end;
      margin-right: 10px;

      .action-btn {
        display: inline-flex;
        align-items: center;
        gap: 4px;
        padding: 4px 8px;
        border-radius: 4px;
        cursor: pointer;
        font-size: 12px;
        transition: all 0.3s ease;

        i {
          font-size: 14px;
        }

        &.edit-btn {
          color: $btn_bg;

          &:hover {
            color: #2a3cc7;
          }
        }

        &.delete-btn {
          color: $btn_bg;

          &:hover {
            color: #2a3cc7;
          }
        }

        &.save-btn {
          color: $btn_bg;

          &:hover {
            color: #2a3cc7;
          }
        }

        &.cancel-btn {
          color: #909399;

          &:hover {
            color: #606266;
          }
        }
      }
    }

    .segment-score {
      display: flex;
      align-items: center;
      position: absolute;
      right: 20px;
      top: 50%;
      transform: translateY(-50%);

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

    .segment-content {
      padding: 10px;
      text-align: left;

      .content-display {
        word-wrap: break-word;
        line-height: 1.5;

        img {
          width: auto;
          max-height: 115px;
        }
      }
    }
  }
}

.smartDate {
  padding-top: 3px;
  color: #888888;
}

.tagList {
  cursor: pointer;

  .icon-tag {
    transform: rotate(-40deg);
    margin-right: 3px;
  }

  .tagList-item {
    color: #888;
  }
}

.tagList > .tagList-item:hover {
  color: $color;
}

.showMore {
  margin-left: 5px;
  background: $color_opacity;
  padding: 2px;
  border-radius: 4px;
}

.metaItem {
  margin-left: 5px;
  background: $color_opacity;
  padding: 2px;
  border-radius: 4px;
}

.editIcon {
  cursor: pointer;
  color: $color;
  font-size: 16px;
  display: inline-block;
  margin-left: 5px;
}

.section {
  width: 100%;
  height: calc(100vh - 64px);
  min-height: unset;
  padding: 20px 20px 0 20px;
  margin: auto;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;

  .el-divider--horizontal {
    margin: 30px 0;
  }

  .title {
    font-size: 18px;
    font-weight: bold;
    color: #333;
    padding: 10px 0;
  }

  .el-descriptions :not(.is-bordered) .el-descriptions-item__cell {
    &:nth-child(even) {
      width: 25%;
    }

    padding: 10px;
  }

  .btn {
    display: flex;
    justify-content: space-between;
    padding: 10px 0;
  }

  .container {
    display: flex;
    min-width: 980px;
    flex: 1;
    min-height: 0;
    border-radius: 5px;
    overflow: hidden;

    .section-preview-panel {
      width: 50%;
      min-width: 400px;
      max-width: 60%;
      flex-shrink: 0;
      overflow: hidden;

      // 覆盖 FilePreviewDrawer 内部样式以适应左侧面板
      .preview-panel {
        border-left: none;
        border-right: 1px solid #e4e7ed;
      }

      .resize-handle {
        left: auto;
        right: -3px;
        border-radius: 0 12px 12px 0;
      }

      &.preview-fail-panel {
        display: flex;
        align-items: center;
        justify-content: center;
        border-right: 1px solid #e4e7ed;
        background: #fafafa;
      }
    }

    .section-content-panel {
      flex: 1;
      min-width: 0;
      min-height: 0;
      display: flex;
      flex-direction: column;
      padding: 0 10px;
      overflow: hidden;

      // 当预览面板隐藏时，分段列表占据全部宽度
      &.full-width {
        padding: 0;
      }

      .card {
        flex: 1;
        overflow-y: auto;
        overflow-x: hidden;
        display: flex;
        flex-direction: column;
        gap: 12px;
        padding: 0 2px;
        flex-wrap: nowrap;

        .text {
          font-size: 14px;
        }

        .item {
          min-height: 40px;
          margin-bottom: 10px;
          display: -webkit-box;
          -webkit-line-clamp: 4;
          -webkit-box-orient: vertical;
          overflow: hidden;
          text-overflow: ellipsis;

          img {
            width: auto;
            max-height: 75px;
          }
        }

        .clearfix {
          display: flex;
          justify-content: space-between;
          align-items: center;
        }

        .segment-card {
          flex-shrink: 0;
          margin: 0 10px;

          &:hover {
            cursor: pointer;
            transform: scale(1.03);
          }

          .more {
            margin-left: 5px;
            cursor: pointer;
            transform: rotate(90deg);
            font-size: 16px;
            color: #8c8c8f;
          }

          ::v-deep .el-card__body {
            overflow: hidden;
          }
        }

        .segment-type {
          margin: 0 5px;
          color: #999;
          font-size: 12px;
        }

        .segment-length {
          color: #999;
          font-size: 12px;
        }

        .segment-child {
          color: #999;
          font-size: 12px;
          padding-left: 5px;
        }
      }
    }
  }
}

.keyword-tags {
  align-items: center;
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.tooltip-content {
  display: flex;
  align-items: center;
  gap: 8px;

  .model-img {
    width: 18px;
    height: 18px;
    border-radius: 4px;
  }

  .model-name {
    font-weight: 500;
  }
}

.tooltip-content .keyword-tag {
  margin: 2px 4px 2px 0;
  color: #1a56db;
}
</style>

<style lang="scss" scoped>
.tagList .tagList-item {
  padding: 2px 4px;
  background: rgb(225, 225, 225);
  border-radius: 10px;
  &:hover {
    background: $tag_bg;
  }
}
</style>
<style lang="scss">
.custom-tooltip {
  max-width: none !important;
}
</style>
