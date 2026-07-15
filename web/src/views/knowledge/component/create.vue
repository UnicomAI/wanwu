<template>
  <div>
    <el-dialog
      top="10vh"
      :title="getTitle()"
      :close-on-click-modal="false"
      :visible.sync="dialogVisible"
      width="70%"
      :before-close="handleClose"
      class="knowledge-create-dialog"
    >
      <el-form
        :model="ruleForm"
        ref="ruleForm"
        label-width="140px"
        class="demo-ruleForm"
        :rules="rules"
        @submit.native.prevent
      >
        <!-- tabs -->
        <div
          class="tabs tabs-x-bottom"
          v-if="!isEdit && category === KNOWLEDGE"
        >
          <div
            :class="['tab', { active: tabActive === INTERNAL }]"
            @click="tabClick(INTERNAL)"
          >
            {{ $t('knowledgeManage.internal') }}
          </div>
          <div
            :class="['tab', { active: tabActive === EXTERNAL }]"
            @click="tabClick(EXTERNAL)"
          >
            {{ $t('knowledgeManage.external') }}
          </div>
        </div>
        <div
          class="card"
          v-if="!isEdit && category === KNOWLEDGE && tabActive === INTERNAL"
        >
          <div
            :class="['card-item', { 'is-active': localCategory === KNOWLEDGE }]"
            @click="localCategory = KNOWLEDGE"
          >
            <img
              class="card-img"
              src="@/assets/imgs/textKnowledge.svg"
              alt=""
            />
            <div>
              <div class="card-name">
                {{ $t('knowledgeManage.textKnowledgeDatabase.title') }}
              </div>
              <div class="card-detail">
                {{ $t('knowledgeManage.textKnowledgeDatabase.desc') }}
              </div>
            </div>
          </div>
          <div
            :class="[
              'card-item',
              { 'is-active': localCategory === MULTIMODAL },
            ]"
            @click="localCategory = MULTIMODAL"
          >
            <img
              class="card-img"
              src="@/assets/imgs/multiKnowledge.svg"
              alt=""
            />
            <div>
              <div class="card-name">
                {{ $t('knowledgeManage.multiKnowledgeDatabase.title') }}
              </div>
              <div class="card-detail">
                {{ $t('knowledgeManage.multiKnowledgeDatabase.desc') }}
              </div>
            </div>
          </div>
        </div>
        <el-form-item
          :label="categoryLabelMap['avatar'][category]"
          prop="avatar"
        >
          <upload-avatar
            v-model="ruleForm.avatar"
            :default-avatar="require('@/assets/imgs/knowledgeIcon.png')"
          />
        </el-form-item>
        <el-form-item :label="categoryLabelMap['name'][category]" prop="name">
          <el-input
            v-model="ruleForm.name"
            :placeholder="$t('knowledgeManage.categoryNameRules')"
            maxlength="50"
            show-word-limit
          ></el-input>
        </el-form-item>
        <el-form-item
          :label="categoryLabelMap['desc'][category]"
          prop="description"
        >
          <el-input
            v-model="ruleForm.description"
            :placeholder="$t('common.input.inputDesc')"
            maxlength="200"
            show-word-limit
          ></el-input>
        </el-form-item>
        <template v-if="tabActive === INTERNAL">
          <el-form-item label="Embedding" prop="embeddingModelInfo.modelId">
            <modelSelect
              v-model="ruleForm.embeddingModelInfo.modelId"
              :options="
                localCategory === MULTIMODAL
                  ? multiEmbeddingOptions
                  : EmbeddingOptions
              "
              :disabled="isEdit"
              warning
            />
          </el-form-item>
          <el-form-item
            prop="knowledgeGraph.switch"
            v-if="category === KNOWLEDGE && localCategory === KNOWLEDGE"
          >
            <template #label>
              <span>{{ $t('knowledgeManage.create.knowledgeGraph') }}:</span>
              <el-tooltip
                class="item"
                effect="dark"
                placement="top-start"
                popper-class="knowledge-graph-tooltip"
              >
                <span class="el-icon-question question-icon"></span>
                <template #content>
                  <p
                    v-for="(item, i) in knowledgeGraphTips"
                    :key="i"
                    class="tooltip-item"
                  >
                    <span class="tooltip-title">{{ item.title }}</span>
                    <span class="tooltip-content">{{ item.content }}</span>
                  </p>
                </template>
              </el-tooltip>
            </template>
            <el-switch
              v-model="ruleForm.knowledgeGraph.switch"
              :disabled="isEdit"
            ></el-switch>
          </el-form-item>
          <el-form-item
            :label="$t('knowledgeManage.create.modelSelect') + ':'"
            prop="knowledgeGraph.llmModelId"
            v-if="ruleForm.knowledgeGraph.switch"
          >
            <modelSelect
              v-model="ruleForm.knowledgeGraph.llmModelId"
              :options="knowledgeGraphModelOptions"
              :placeholder="$t('knowledgeManage.create.modelSearchPlaceholder')"
              @visible-change="visibleChange"
              :loading-text="$t('knowledgeManage.create.modelLoading')"
              :loading="modelLoading"
              filterable
              warning
              :disabled="isEdit"
            />
          </el-form-item>
          <el-form-item
            :label="$t('knowledgeManage.create.uploadSchema') + ':'"
            v-if="ruleForm.knowledgeGraph.switch && !isEdit"
          >
            <el-upload
              action=""
              :auto-upload="false"
              :show-file-list="false"
              :on-change="uploadOnChange"
              :file-list="fileList"
              :limit="1"
              drag
              :disabled="isEdit"
              accept=".xlsx,.xls"
              class="upload-box"
            >
              <div>
                <div>
                  <img
                    :src="require('@/assets/imgs/uploadImg.png')"
                    class="upload-img"
                  />
                  <p class="click-text">
                    {{ $t('common.fileUpload.uploadText') }}
                    <span class="clickUpload">
                      {{ $t('common.fileUpload.uploadClick') }}
                    </span>
                  </p>
                </div>
                <div class="tips">
                  <p>
                    <span class="red">*</span>
                    {{ $t('knowledgeManage.create.schemaTip1') }}
                    <a
                      class="template_downLoad"
                      href="#"
                      @click.prevent.stop="downloadTemplate"
                    >
                      {{ $t('knowledgeManage.create.templateDownload') }}
                    </a>
                  </p>
                  <p>
                    <span class="red">*</span>
                    {{ $t('knowledgeManage.create.schemaTip2') }}
                  </p>
                </div>
              </div>
            </el-upload>
            <!-- 上传文件的列表 -->
            <div class="file-list" v-if="fileList.length > 0">
              <transition name="el-zoom-in-top">
                <ul class="document_lise">
                  <li
                    v-for="(file, index) in fileList"
                    :key="index"
                    class="document_lise_item"
                  >
                    <div style="padding: 8px 0" class="lise_item_box">
                      <span class="size">
                        <img :src="require('@/assets/imgs/fileicon.png')" />
                        {{ file.name }}
                        <span class="file-size">
                          {{ filterSize(file.size) }}
                        </span>
                        <el-progress
                          :percentage="file.percentage"
                          v-if="file.percentage !== 100"
                          :status="file.progressStatus"
                          max="100"
                          class="progress"
                        ></el-progress>
                      </span>
                      <span class="handleBtn">
                        <span>
                          <span v-if="file.percentage === 100">
                            <i
                              class="el-icon-check check success"
                              v-if="file.progressStatus === 'success'"
                            ></i>
                            <i class="el-icon-close close fail" v-else></i>
                          </span>
                          <i
                            class="el-icon-loading"
                            v-else-if="
                              file.percentage !== 100 && index === fileIndex
                            "
                          ></i>
                        </span>
                        <span style="margin-left: 30px">
                          <i
                            class="el-icon-error error"
                            @click="handleRemove(file, index)"
                          ></i>
                        </span>
                      </span>
                    </div>
                  </li>
                </ul>
              </transition>
            </div>
          </el-form-item>
        </template>
        <template v-if="tabActive === EXTERNAL">
          <el-form-item
            :label="$t('knowledgeManage.externalSource')"
            prop="externalSource"
          >
            <el-select
              v-model="ruleForm.externalSource"
              :placeholder="$t('common.select.placeholder')"
            >
              <el-option
                v-for="item in externalSourceOptions"
                :key="item.sourceId"
                :label="item.displayName"
                :value="item.sourceId"
              ></el-option>
            </el-select>
          </el-form-item>
          <el-form-item
            :label="$t('knowledgeManage.external') + 'API'"
            prop="externalApiId"
          >
            <el-select
              v-model="ruleForm.externalApiId"
              :placeholder="$t('common.select.placeholder')"
              :loading="externalApiLoading"
              @change="externalApiChange($event)"
            >
              <el-option
                key="createNew"
                :label="
                  $t('common.create') + $t('knowledgeManage.externalKnowledge')
                "
                value="createNew"
              ></el-option>
              <el-option
                v-for="item in externalApiOptions"
                :key="item.externalApiId"
                :label="item.name"
                :value="item.externalApiId"
              ></el-option>
            </el-select>
          </el-form-item>
          <el-form-item
            :label="$t('knowledgeManage.externalKnowledge')"
            prop="externalKnowledgeId"
          >
            <el-select
              v-model="ruleForm.externalKnowledgeId"
              :placeholder="$t('common.select.placeholder')"
              :disabled="!ruleForm.externalApiId"
              :loading="externalKnowledgeLoading"
            >
              <el-option
                v-for="item in externalKnowledgeOptions"
                :key="item.externalKnowledgeId"
                :label="item.externalKnowledgeName"
                :value="item.externalKnowledgeId"
              ></el-option>
            </el-select>
          </el-form-item>
        </template>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button @click="handleClose()">
          {{ $t('common.confirm.cancel') }}
        </el-button>
        <el-button type="primary" @click="submitForm('ruleForm')">
          {{ $t('common.confirm.confirm') }}
        </el-button>
      </span>
    </el-dialog>
  </div>
</template>
<script>
import { mapActions, mapGetters } from 'vuex';
import {
  createKnowledgeItem,
  editKnowledgeItem,
  getExternalAPIList,
  getExternalList,
  createExternal,
  editExternal,
} from '@/api/knowledge';
import { getMultiEmbeddingList, selectModelList } from '@/api/modelAccess';
import { KNOWLEDGE_GRAPH_TIPS } from '../config';
import uploadChunk from '@/mixins/uploadChunk';
import { delfile } from '@/api/chunkFile';
import modelSelect from '@/components/modelSelect.vue';
import {
  INTERNAL,
  EXTERNAL,
  KNOWLEDGE,
  QA,
  MULTIMODAL,
  DB,
} from '@/views/knowledge/constants';
import uploadAvatar from '@/components/uploadAvatar.vue';
import { filterSize, authDownload } from '@/utils/util';

export default {
  props: {
    category: {
      type: Number,
      default: KNOWLEDGE,
    },
  },
  components: {
    uploadAvatar,
    modelSelect,
  },
  mixins: [uploadChunk],
  data() {
    return {
      INTERNAL,
      EXTERNAL,
      KNOWLEDGE,
      QA,
      MULTIMODAL,
      DB,
      dialogVisible: false,
      tabActive: INTERNAL,
      localCategory: this.category,
      ruleForm: {
        name: '',
        description: '',
        embeddingModelInfo: {
          modelId: '',
        },
        knowledgeGraph: {
          llmModelId: '',
          schemaUrl: '',
          switch: false,
        },
        externalSource: 'dify',
        externalApiId: '',
        externalKnowledgeId: '',
      },
      EmbeddingOptions: [],
      multiEmbeddingOptions: [],
      knowledgeGraphModelOptions: [],
      externalSourceOptions: [{ sourceId: 'dify', displayName: 'Dify' }],
      externalApiOptions: [],
      externalKnowledgeOptions: [],
      modelLoading: false,
      externalApiLoading: false,
      externalKnowledgeLoading: false,
      knowledgeGraphTips: KNOWLEDGE_GRAPH_TIPS,
      maxSizeBytes: 0, // 设置为0，所有文件都走切片上传
      rules: {
        name: [
          {
            required: true,
            message: this.$t('knowledgeManage.knowledgeNameRules'),
            trigger: 'blur',
          },
          {
            pattern: this.$config.commonTextReg,
            message: this.$t('common.hint.text'),
            trigger: 'blur',
          },
          {
            min: 2,
            max: 50,
            message: this.$t('common.hint.textLimit'),
            trigger: 'blur',
          },
        ],
        description: [
          {
            required: true,
            message: this.$t('knowledgeManage.inputDesc'),
            trigger: 'blur',
          },
          {
            max: 200,
            message: this.$t('common.hint.descLimit'),
            trigger: 'blur',
          },
        ],
        'embeddingModelInfo.modelId': [
          {
            required: true,
            message: this.$t('common.select.placeholder'),
            trigger: 'blur',
          },
        ],
        'knowledgeGraph.llmModelId': [
          {
            required: true,
            message: this.$t('knowledgeManage.create.selectModel'),
            trigger: 'change',
          },
        ],
        externalSource: [
          {
            required: true,
            message: this.$t('common.select.placeholder'),
            trigger: 'blur',
          },
        ],
        externalApiId: [
          {
            required: true,
            message: this.$t('common.select.placeholder'),
            trigger: 'blur',
          },
        ],
        externalKnowledgeId: [
          {
            required: true,
            message: this.$t('common.select.placeholder'),
            trigger: 'blur',
          },
        ],
      },
      isEdit: false,
      knowledgeId: '',
      categoryLabelMap: {
        avatar: {
          [KNOWLEDGE]: this.$t('knowledgeManage.knowledgeAvatar'),
          [QA]: this.$t('knowledgeManage.qaDatabase.avatar'),
          [DB]: this.$t('knowledgeManage.dbDatabase.avatar'),
        },
        name: {
          [KNOWLEDGE]: this.$t('knowledgeManage.knowledgeName'),
          [QA]: this.$t('knowledgeManage.qaDatabase.name'),
          [DB]: this.$t('knowledgeManage.dbDatabase.name'),
        },
        desc: {
          [KNOWLEDGE]: this.$t('knowledgeManage.desc'),
          [QA]: this.$t('knowledgeManage.qaDatabase.desc'),
          [DB]: this.$t('knowledgeManage.dbDatabase.desc'),
        },
      },
    };
  },
  watch: {
    embeddingList: {
      handler(val) {
        if (val) {
          this.EmbeddingOptions = val;
        }
      },
    },
    'ruleForm.externalApiId': {
      handler(val) {
        this.ruleForm.externalKnowledgeId = '';
        if (val === 'createNew') {
          this.ruleForm.externalApiId = '';
          this.externalKnowledgeOptions = [];
        } else if (val) {
          this.getExternalKnowledgeList();
        } else {
          this.externalKnowledgeOptions = [];
        }
      },
    },
  },
  computed: {
    ...mapGetters('app', ['embeddingList']),
  },
  created() {
    this.getEmbeddingList();
    getMultiEmbeddingList().then(res => {
      if (res.code === 0) {
        this.multiEmbeddingOptions = res.data.list || [];
      }
    });
    this.getModelData(); //获取模型列表
  },
  methods: {
    ...mapActions('app', ['getEmbeddingList']),
    tabClick(type) {
      this.tabActive = type;
    },
    visibleChange(val) {
      //下拉框显示的时候请求模型列表
      if (val) {
        this.getModelData();
      }
    },
    getTitle() {
      if (this.category === KNOWLEDGE) {
        if (this.isEdit) {
          return this.$t('knowledgeManage.editInfo');
        } else {
          return this.$t('knowledgeManage.createKnowledge');
        }
      } else if (this.category === QA) {
        if (this.isEdit) {
          return this.$t('knowledgeManage.qaDatabase.editInfo');
        } else {
          return this.$t('knowledgeManage.qaDatabase.createKnowledge');
        }
      } else if (this.category === DB) {
        if (this.isEdit) {
          return this.$t('knowledgeManage.dbDatabase.editInfo');
        } else {
          return this.$t('knowledgeManage.dbDatabase.createKnowledge');
        }
      }
    },
    async downloadTemplate() {
      try {
        await authDownload(
          '/user/api/v1/files/docs/graph_schema.xlsx',
          'graph_schema.xlsx',
        );
      } catch (error) {
        this.$message.error(this.$t('knowledgeManage.create.downloadFailed'));
      }
    },
    async getModelData() {
      this.modelLoading = true;
      const res = await selectModelList();
      if (res.code === 0) {
        this.knowledgeGraphModelOptions = res.data.list || [];
        this.modelLoading = false;
      }
      this.modelLoading = false;
    },
    async getExternalAPIList() {
      this.externalApiLoading = true;
      const res = await getExternalAPIList();
      if (res.code === 0) {
        this.externalApiOptions = res.data.externalApiList;
        this.externalApiLoading = false;
      }
      this.externalApiLoading = false;
    },
    async getExternalKnowledgeList() {
      this.externalKnowledgeLoading = true;
      const res = await getExternalList({
        externalApiId: this.ruleForm.externalApiId,
      });
      if (res.code === 0) {
        this.externalKnowledgeOptions = res.data.externalKnowledgeList;
        this.externalKnowledgeLoading = false;
      }
      this.externalKnowledgeLoading = false;
    },
    externalApiChange(val) {
      if (val === 'createNew') {
        this.$emit('createExternalApi');
        this.ruleForm.externalApiId = '';
      }
    },
    handleClose() {
      this.dialogVisible = false;
      this.clearform();
    },
    clearform() {
      ((this.isEdit = false), (this.knowledgeId = ''));
      this.$refs.ruleForm.resetFields();
      this.ruleForm = {
        name: '',
        description: '',
        embeddingModelInfo: {
          modelId: '',
        },
        knowledgeGraph: {
          llmModelId: '',
          schemaUrl: '',
          switch: false,
        },
        externalSource: 'dify',
        externalApiId: '',
        externalKnowledgeId: '',
      };
      this.fileList = [];
      this.cancelAllRequests();
      this.file = null;
      this.fileIndex = 0;
      this.fileUuid = '';
    },
    uploadOnChange(file, fileList) {
      if (!fileList.length) return;
      this.fileList = fileList;
      if (
        this.verifyEmpty(file) !== false &&
        this.verifyFormat(file) !== false &&
        this.verifyRepeat(file) !== false
      ) {
        setTimeout(() => {
          this.fileList.map((file, index) => {
            if (file.progressStatus && file.progressStatus !== 'success') {
              this.$set(file, 'progressStatus', 'exception');
              this.$set(file, 'showRetry', 'false');
              this.$set(file, 'showResume', 'false');
              this.$set(file, 'showRemerge', 'false');
              if (file.size > this.maxSizeBytes) {
                this.$set(file, 'fileType', 'maxFile');
              } else {
                this.$set(file, 'fileType', 'minFile');
              }
            }
          });
        }, 10);
        //开始切片上传(如果没有文件正在上传)
        if (this.file === null) {
          this.startUpload();
        } else if (this.file.progressStatus === 'success') {
          //如果上传当中有新的文件加入
          this.startUpload(this.fileIndex);
        }
      }
    },
    //  验证文件为空
    verifyEmpty(file) {
      if (file.size <= 0) {
        setTimeout(() => {
          this.$message.warning(
            file.name + this.$t('knowledgeManage.filterFile'),
          );
          this.fileList = this.fileList.filter(
            files => files.name !== file.name,
          );
        }, 50);
        return false;
      }
      return true;
    },
    //  验证文件格式
    verifyFormat(file) {
      const nameType = ['xlsx', 'xls'];
      const fileName = file.name;
      const isSupportedFormat = nameType.some(ext =>
        fileName.endsWith(`.${ext}`),
      );
      if (!isSupportedFormat) {
        setTimeout(() => {
          this.$message.warning(
            file.name + this.$t('knowledgeManage.fileTypeError'),
          );
          this.fileList = this.fileList.filter(
            files => files.name !== file.name,
          );
        }, 50);
        return false;
      } else {
        const fileType = file.name.split('.').pop();
        const limit20 = ['xlsx', 'xls'];
        let isLimit20 = file.size / 1024 / 1024 < 20;
        let num = 0;
        if (limit20.includes(fileType)) {
          num = 20;
          if (!isLimit20) {
            setTimeout(() => {
              this.$message.error(
                this.$t('knowledgeManage.limitSize') + `${num}MB!`,
              );
              this.fileList = this.fileList.filter(
                files => files.name !== file.name,
              );
            }, 50);
            return false;
          }
          return true;
        }
        return true;
      }
    },
    //  验证文件重复
    verifyRepeat(file) {
      let res = true;
      setTimeout(() => {
        this.fileList = this.fileList.reduce((accumulator, current) => {
          const length = accumulator.filter(
            obj => obj.name === current.name,
          ).length;
          if (length === 0) {
            accumulator.push(current);
          } else {
            this.$message.warning(
              current.name + this.$t('knowledgeManage.fileExist'),
            );
            res = false;
          }
          return accumulator;
        }, []);
        return res;
      }, 50);
    },
    filterSize,
    handleRemove(item, index) {
      if (item.percentage < 100) {
        this.fileList.splice(index, 1);
        this.cancelAndRestartNextRequests();
        return;
      }
      // 如果文件已上传成功，需要删除服务器上的文件
      if (this.resList && this.resList[index] && this.resList[index]['name']) {
        this.delfile({
          fileList: [this.resList[index]['name']],
          isExpired: true,
        });
        this.resList.splice(index, 1);
      }
      this.fileList = this.fileList.filter(files => files.name !== item.name);
      if (this.fileList.length === 0) {
        this.file = null;
        this.ruleForm.knowledgeGraph.schemaUrl = '';
      } else {
        this.fileIndex--;
      }
    },
    delfile(data) {
      delfile(data).then(res => {
        if (res.code === 0) {
          this.$message.success(
            this.$t('knowledgeManage.create.deleteSuccess'),
          );
        }
      });
    },
    uploadFile(fileName, oldName, filePath) {
      this.ruleForm.knowledgeGraph.schemaUrl = filePath || fileName;
      this.fileIndex++;
      if (this.fileIndex < this.fileList.length) {
        this.startUpload(this.fileIndex);
      }
    },
    submitForm(formName) {
      this.$refs[formName].validate(valid => {
        if (valid) {
          if (this.isEdit) {
            this.editKnowledge();
          } else {
            this.createKnowledge();
          }
          this.$parent?.clearIptValue();
        } else {
          return false;
        }
      });
    },
    createKnowledge() {
      const data = {
        ...this.ruleForm,
        category: this.localCategory,
      };
      const request =
        this.tabActive === EXTERNAL
          ? createExternal(data)
          : createKnowledgeItem(data);
      request.then(res => {
        if (res.code === 0) {
          this.$message.success(
            this.$t('knowledgeManage.create.createSuccess'),
          );
          this.$emit('reloadData');
          this.dialogVisible = false;
        }
      });
    },
    editKnowledge() {
      const data = {
        ...this.ruleForm,
        knowledgeId: this.knowledgeId,
      };
      const request =
        this.tabActive === EXTERNAL
          ? editExternal(data)
          : editKnowledgeItem(data);
      request.then(res => {
        if (res.code === 0) {
          this.$message.success(this.$t('knowledgeManage.create.editSuccess'));
          this.$emit('reloadData');
          this.clearform();
          this.dialogVisible = false;
        }
      });
    },
    showDialog(row) {
      console.log(row);
      this.dialogVisible = true;
      this.isEdit = Boolean(row);
      this.tabActive = INTERNAL;
      this.localCategory = this.category;
      if (row) {
        this.localCategory = row.category;
        this.knowledgeId = row.knowledgeId;
        this.ruleForm = {
          avatar: row.avatar,
          name: row.name,
          description: row.description,
          embeddingModelInfo: {
            modelId: row.embeddingModelInfo.modelId,
          },
          knowledgeGraph: {
            llmModelId: row.llmModelId,
            switch: row.graphSwitch === 1,
            schemaUrl: '',
          },
          externalSource: 'dify',
          externalApiId: '',
          externalKnowledgeId: '',
        };
        if (row.external === EXTERNAL) {
          this.ruleForm.externalSource =
            row.externalKnowledgeInfo.externalSource;
          this.ruleForm.externalApiId = row.externalKnowledgeInfo.externalApiId;
          this.ruleForm.externalKnowledgeId =
            row.externalKnowledgeInfo.externalKnowledgeId;
          this.tabActive = EXTERNAL;
          this.externalApiOptions = [
            {
              externalApiId: this.ruleForm.externalApiId,
              name: row.externalKnowledgeInfo.externalApiName,
            },
          ];
          this.externalKnowledgeOptions = [
            {
              externalKnowledgeId: this.ruleForm.externalKnowledgeId,
              externalKnowledgeName:
                row.externalKnowledgeInfo.externalKnowledgeName,
            },
          ];
          this.getExternalKnowledgeList();
          this.$nextTick(() => {
            this.ruleForm.externalKnowledgeId =
              row.externalKnowledgeInfo.externalKnowledgeId;
          });
        }
      } else {
        this.ruleForm = {
          name: '',
          description: '',
          embeddingModelInfo: {
            modelId: '',
          },
          knowledgeGraph: {
            llmModelId: '',
            schemaUrl: '',
            switch: false,
          },
          externalSource: 'dify',
          externalApiId: '',
          externalKnowledgeId: '',
        };
      }
      this.getExternalAPIList();
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/tabs';

.card {
  margin-left: 20px;
  display: flex;

  .card-item {
    flex: 1;
    margin-bottom: 20px;
    margin-right: 20px;
    border-radius: 8px;
    border: 1px solid #d9d9d9;
    padding: 15px 20px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: flex-start;
    .card-img {
      width: 50px;
      height: 50px;
      object-fit: contain;
      padding: 10px 6px;
      background: #ffffff;
      box-shadow: 0 1px 4px 0 rgba(0, 0, 0, 0.15);
      border-radius: 8px;
      border: 0 solid #d9d9d9;
      margin-right: 16px;
    }
    .card-name {
      font-size: 17px;
      font-weight: bold;
      color: $color_title;
      margin-bottom: 5px;
    }
    .card-detail {
      font-size: 12px;
      margin-bottom: 5px;
    }
  }
  .card-item:hover,
  .card-item.is-active {
    box-shadow: 0 1px 4px 0 rgba(0, 0, 0, 0.15);
    border: 1px solid $color;
    .card-name {
      color: $color;
    }
  }
}

.knowledge-create-dialog {
  ::v-deep .el-dialog__body {
    max-height: 60vh;
    overflow-y: auto;
    padding: 20px;
  }

  ::v-deep .el-form-item {
    .el-select {
      width: 100%;
    }
  }
}

.question-icon {
  cursor: pointer;
  color: #909399;
}

.upload-box {
  height: auto;
  min-height: 190px;
  width: 100% !important;

  .upload-img {
    width: 56px;
    height: 56px;
    margin-top: 20px;
  }

  .click-text {
    .clickUpload {
      color: $color;
      font-weight: bold;
    }
  }

  .tips {
    padding: 0 20px;

    p {
      line-height: 1.6;
      color: #666666 !important;

      .red {
        color: #f56c6c;
      }

      .template_downLoad {
        margin-left: 5px;
        color: $color;
        cursor: pointer;
      }
    }
  }
}

.file-list {
  padding: 20px 0;

  .document_lise {
    list-style: none;
    padding: 0;
    margin: 0;
  }

  .document_lise_item {
    cursor: pointer;
    padding: 5px 10px;
    list-style: none;
    background: #fff;
    border-radius: 4px;
    box-shadow: 1px 2px 2px #ddd;
    display: flex;
    align-items: center;
    margin-bottom: 10px;

    .lise_item_box {
      width: 100%;
      display: flex;
      align-items: center;
      justify-content: space-between;

      .size {
        display: flex;
        align-items: center;

        .progress {
          width: 400px;
          margin-left: 30px;
        }

        img {
          width: 18px;
          height: 18px;
          margin-bottom: -3px;
        }

        .file-size {
          margin-left: 10px;
        }
      }

      .handleBtn {
        display: flex;
        align-items: center;

        .check.success {
          color: #67c23a;
        }

        .close.fail {
          color: #f56c6c;
        }

        .error {
          color: #f56c6c;
          cursor: pointer;
          font-size: 18px;
        }
      }
    }
  }

  .document_lise_item:hover {
    background: #eceefe;
  }
}
</style>

<style lang="scss">
.knowledge-graph-tooltip {
  max-width: 400px !important;

  .tooltip-item {
    margin: 0;
    padding: 4px 0;

    .tooltip-title {
      font-weight: bold;
      margin-right: 8px;
    }

    .tooltip-content {
      display: inline-block;
    }
  }
}
</style>
