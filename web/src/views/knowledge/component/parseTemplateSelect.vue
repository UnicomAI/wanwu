<template>
  <div class="parse-template-select">
    <div class="template-grid">
      <div
        class="template-item"
        v-for="item in docTypeList"
        :key="item.docType"
      >
        <p class="item-label">
          <FileIcon class="item-icon" :type="item.icon" size="16px" />
          <span class="item-name">{{ item.name }}</span>
          <span
            v-if="!disabled"
            class="item-edit"
            @click="openTemplatePage(item.docType)"
          >
            {{ $t('common.button.edit') }}
          </span>
        </p>
        <el-select
          :value="bind[item.docType] || builtInValue(item.docType)"
          :disabled="disabled"
          @change="handleChange(item.docType, $event)"
        >
          <el-option
            :label="$t('knowledgeManage.parseTemplate.builtIn')"
            :value="builtInValue(item.docType)"
          ></el-option>
          <el-option
            v-for="template in customTemplates(item.docType)"
            :key="template.templateId"
            :label="template.name"
            :value="template.templateId"
          ></el-option>
          <el-option
            class="create-template-option"
            :label="$t('knowledgeManage.parseTemplate.createTemplate')"
            :value="CREATE_OPTION"
          ></el-option>
        </el-select>
      </div>
    </div>
  </div>
</template>
<script>
import { getParseTemplateList } from '@/api/parseTemplate';
import FileIcon from '@/components/FileIcon.vue';
import { DOC_TYPE_LIST } from '../parseTemplate/config';

const CREATE_OPTION = '__create__';

export default {
  name: 'ParseTemplateSelect',
  components: { FileIcon },
  props: {
    // 各文档类型选定的模板，格式 {docType: templateId}
    value: { type: Object, default: () => ({}) },
    disabled: { type: Boolean, default: false },
  },
  data() {
    return {
      CREATE_OPTION,
      docTypeList: DOC_TYPE_LIST,
      bind: { ...this.value },
      grouped: {},
      needRefresh: false,
    };
  },
  watch: {
    value(val) {
      this.bind = { ...val };
    },
  },
  created() {
    this.getList();
  },
  mounted() {
    window.addEventListener('focus', this.refreshAfterEdit);
  },
  beforeDestroy() {
    window.removeEventListener('focus', this.refreshAfterEdit);
  },
  methods: {
    // 内置（默认）模板由后端保底落库，取它自己的 id
    builtInValue(docType) {
      const builtIn = (this.grouped[docType] || []).find(item => item.builtIn);
      return builtIn ? builtIn.templateId : '';
    },
    // 内置模板单独渲染，列表里只留用户自建的
    customTemplates(docType) {
      return (this.grouped[docType] || []).filter(item => !item.builtIn);
    },
    currentTemplate(docType) {
      const templateId = this.bind[docType] || this.builtInValue(docType);
      return (
        (this.grouped[docType] || []).find(
          item => item.templateId === templateId,
        ) || null
      );
    },
    // 另开标签编辑，避免带走当前页已选的文件
    openTemplatePage(docType) {
      const template = this.currentTemplate(docType);
      const { href } = this.$router.resolve({
        path: '/knowledge/parseTemplate',
        query: {
          docType,
          templateId: template ? template.templateId : undefined,
        },
      });
      this.needRefresh = true;
      window.open(href, '_blank');
    },
    refreshAfterEdit() {
      if (!this.needRefresh) return;
      this.needRefresh = false;
      this.getList();
    },
    getList() {
      getParseTemplateList().then(res => {
        if (res.code !== 0) return;
        this.grouped = (res.data.list || []).reduce((acc, template) => {
          (acc[template.docType] = acc[template.docType] || []).push(template);
          return acc;
        }, {});
      });
    },
    handleChange(docType, templateId) {
      if (templateId === CREATE_OPTION) {
        this.$emit('create');
        this.$router.push('/knowledge/parseTemplate');
        return;
      }
      this.$set(this.bind, docType, templateId);
      this.$emit('input', { ...this.bind });
    },
  },
};
</script>
<style lang="scss" scoped>
.parse-template-select {
  .template-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    gap: 8px 12px;
  }

  .template-item {
    .item-label {
      display: flex;
      align-items: center;
      gap: 6px;
      margin: 0 0 6px;
      font-size: 12px;
      color: #666;
      line-height: 1.4;
    }

    .item-icon {
      flex: none;
      font-size: 16px;
    }

    .item-name {
      flex: 1;
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
    }

    .item-edit {
      flex: none;
      cursor: pointer;
      color: #8a9099;
      transition: color 0.2s;

      &:hover {
        color: #5a6cf3;
      }
    }

    .el-select {
      width: 100%;
    }
  }
}
</style>

<style lang="scss">
// 下拉面板挂在 body 上，样式不能写在 scoped 块里
.create-template-option {
  color: #5a6cf3 !important;
}
</style>
