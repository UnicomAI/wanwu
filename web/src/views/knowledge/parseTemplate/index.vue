<template>
  <div class="page-wrapper full-content">
    <div class="page-title">
      <span class="crumb-link" @click="goBack">
        {{ $t('menu.knowledge') }}
      </span>
      <i class="el-icon-arrow-right crumb-arrow"></i>
      {{ $t('knowledgeManage.parseTemplate.title') }}
      <div class="page-desc">
        {{ $t('knowledgeManage.parseTemplate.desc') }}
      </div>
    </div>
    <div class="block table-wrap list-common wrap-fullheight template-body">
      <div class="doc-type-list">
        <p class="doc-type-title">
          {{ $t('knowledgeManage.parseTemplate.docType') }}
        </p>
        <div
          v-for="item in docTypeList"
          :key="item.docType"
          :class="['doc-type-item', docType === item.docType ? 'active' : '']"
          @click="handleDocTypeChange(item.docType)"
        >
          <FileIcon :type="item.icon" size="24px" />
          <div class="doc-type-info">
            <p class="doc-type-name">{{ item.name }}</p>
            <div class="doc-type-ext">
              <span v-for="ext in item.extList" :key="ext">{{ ext }}</span>
            </div>
          </div>
        </div>
      </div>
      <div class="template-list" v-loading="loading">
        <el-button size="mini" type="primary" @click="addTemplate">
          +{{ $t('knowledgeManage.parseTemplate.newTemplate') }}
        </el-button>
        <p class="empty-tip" v-if="!templateList.length">
          {{ $t('knowledgeManage.parseTemplate.emptyTip') }}
        </p>
        <templateCard
          v-for="(template, index) in templateList"
          :key="template.templateId || `new-${index}`"
          :value="template"
          :mediaType="mediaType"
          :defaultExpanded="isCardExpanded(template, index)"
          @save="handleSave"
          @delete="handleDelete(template, index)"
        />
      </div>
    </div>
  </div>
</template>
<script>
import FileIcon from '@/components/FileIcon.vue';
import templateCard from './component/templateCard.vue';
import { DEFAULT_TEMPLATE_CONFIG, DOC_TYPE_LIST, getMediaType } from './config';
import {
  getParseTemplateList,
  createParseTemplate,
  updateParseTemplate,
  deleteParseTemplate,
} from '@/api/parseTemplate';

export default {
  name: 'ParseTemplate',
  components: { FileIcon, templateCard },
  data() {
    return {
      docTypeList: DOC_TYPE_LIST,
      // 从上传/新建页跳来时按 query 定位到指定文档类型与模板
      docType: DOC_TYPE_LIST.some(
        item => item.docType === this.$route.query.docType,
      )
        ? this.$route.query.docType
        : DOC_TYPE_LIST[0].docType,
      focusTemplateId: this.$route.query.templateId || '',
      // 从下拉框「+创建模板」跳来时，列表加载完直接展开一张空白卡片
      pendingCreate: this.$route.query.create === '1',
      templateList: [],
      loading: false,
    };
  },
  computed: {
    mediaType() {
      return getMediaType(this.docType);
    },
  },
  created() {
    this.getList();
  },
  methods: {
    goBack() {
      this.$router.push('/knowledge');
    },
    getList() {
      this.loading = true;
      getParseTemplateList({ docType: this.docType })
        .then(res => {
          if (res.code !== 0) return;
          const list = res.data.list || [];
          // 视频/音频/图片没有内置（默认）模板，只能用户自建
          this.templateList =
            this.mediaType === 'doc' ? list : list.filter(t => !t.builtIn);
          if (this.pendingCreate) {
            this.pendingCreate = false;
            this.addTemplate();
          }
        })
        .finally(() => {
          this.loading = false;
        });
    },
    isCardExpanded(template, index) {
      if (!template.templateId) return true;
      if (this.focusTemplateId) {
        return template.templateId === this.focusTemplateId;
      }
      return index === 0;
    },
    handleDocTypeChange(docType) {
      if (this.docType === docType) return;
      this.focusTemplateId = '';
      this.docType = docType;
      this.getList();
    },
    // 新建模板以内置（默认）模板打底，媒体类型没有内置就用默认配置
    addTemplate() {
      const builtIn = this.templateList.find(item => item.builtIn);
      const template = structuredClone(builtIn || DEFAULT_TEMPLATE_CONFIG);
      template.templateId = '';
      template.name = '';
      template.builtIn = false;
      this.templateList.unshift(template);
    },
    // 只发后端契约里的字段，别把列表返回的 builtIn/createdAt 带回去
    templatePayload(template) {
      return {
        name: template.name,
        docSegment: template.docSegment,
        docAnalyzer: template.docAnalyzer,
        docPreprocess: template.docPreprocess,
        parserModelId: template.parserModelId,
        asrModelId: template.asrModelId,
        multimodalModelId: template.multimodalModelId,
      };
    },
    handleSave(template) {
      const unsaved = !template.templateId;
      const request = unsaved ? createParseTemplate : updateParseTemplate;
      const data = unsaved
        ? { ...this.templatePayload(template), docType: this.docType }
        : {
            ...this.templatePayload(template),
            templateId: template.templateId,
          };
      request(data).then(res => {
        if (res.code === 0) {
          this.$message.success(this.$t('common.message.success'));
          this.getList();
        }
      });
    },
    handleDelete(template, index) {
      if (!template.templateId) {
        this.templateList.splice(index, 1);
        return;
      }
      this.$confirm(
        this.$t('knowledgeManage.parseTemplate.deleteConfirm'),
        this.$t('common.button.tip'),
        { type: 'warning' },
      ).then(() => {
        deleteParseTemplate({ templateId: template.templateId }).then(res => {
          if (res.code === 0) {
            this.$message.success(this.$t('common.info.delete'));
            this.getList();
          }
        });
      });
    },
  },
};
</script>
<style lang="scss" scoped>
.crumb-link {
  color: #666;
  cursor: pointer;
}
.crumb-arrow {
  margin: 0 8px;
  color: #999;
}
.page-desc {
  margin-top: 8px;
  font-size: 12px;
  font-weight: 400;
  color: #999;
}
.template-body {
  display: flex;
  gap: 24px;
  padding: 20px;
}
.doc-type-list {
  width: 160px;
  flex-shrink: 0;
  border-right: 1px solid #e6e8f0;
  padding-right: 16px;
  overflow-y: auto;
  .doc-type-title {
    margin: 0 0 12px;
    color: #999;
  }
  .doc-type-item {
    display: flex;
    align-items: center;
    padding: 12px;
    margin-bottom: 12px;
    border: 1px solid #e6e8f0;
    border-radius: 8px;
    cursor: pointer;
    &.active {
      border-color: #5a6cf3;
    }
    // 扩展名换两行时不要把图标横向压扁
    .icon-file {
      flex-shrink: 0;
    }
    .doc-type-info {
      margin-left: 8px;
      min-width: 0;
    }
    .doc-type-name {
      margin: 0;
      color: #333;
    }
    .doc-type-ext {
      display: flex;
      flex-wrap: wrap;
      gap: 4px;
      margin-top: 4px;
      span {
        padding: 0 6px;
        font-size: 12px;
        color: #999;
        background: #f5f6fa;
        border-radius: 4px;
      }
    }
  }
}
.template-list {
  flex: 1;
  min-width: 0;
  overflow-y: auto;
  .empty-tip {
    margin-top: 40px;
    text-align: center;
    color: #999;
  }
  .template-card:first-of-type {
    margin-top: 16px;
  }
}
</style>
