<template>
  <div class="parse-template-select">
    <!-- 收起态：只列已选模板，未选则提示走内置默认 -->
    <div v-if="!isOpen" key="summary" class="template-summary">
      <el-tag
        v-for="tag in customTags"
        :key="tag.docType"
        size="small"
        color="#E6F0FF"
        class="summary-tag"
        disable-transitions
      >
        {{ tag.docTypeName }}：{{ tag.templateName }}
      </el-tag>
      <span v-if="loaded && !customTags.length" class="summary-empty">
        {{ $t('knowledgeManage.parseTemplate.noneSelected') }}
      </span>
      <i
        v-if="!disabled"
        class="el-icon-edit-outline summary-edit"
        @click="expanded = true"
      ></i>
    </div>
    <template v-else>
      <div key="grid" class="template-grid">
        <div
          class="template-item"
          :class="{ 'is-error': invalidSet.has(item.docType) }"
          v-for="item in visibleDocTypes"
          :key="item.docType"
        >
          <p class="item-label">
            <FileIcon class="item-icon" :type="item.icon" size="16px" />
            <span class="item-name">
              <span v-if="requiredSet.has(item.docType)" class="item-star">
                *
              </span>
              {{ item.name }}
            </span>
            <span
              v-if="!disabled"
              class="item-edit"
              @click="openTemplatePage(item.docType)"
            >
              {{ $t('common.button.edit') }}
            </span>
          </p>
          <el-select
            :value="selectValue(item.docType)"
            :disabled="disabled"
            :placeholder="$t('knowledgeManage.parseTemplate.noTemplate')"
            @change="handleChange(item.docType, $event)"
          >
            <el-option
              v-for="template in optionTemplates(item.docType)"
              :key="template.templateId"
              :label="
                template.builtIn
                  ? $t('knowledgeManage.parseTemplate.builtIn')
                  : template.name
              "
              :value="template.templateId"
            ></el-option>
            <el-option
              class="create-template-option"
              :label="$t('knowledgeManage.parseTemplate.createTemplate')"
              :value="CREATE_OPTION"
            ></el-option>
          </el-select>
          <p v-if="invalidSet.has(item.docType)" class="item-error">
            {{ $t('knowledgeManage.parseTemplate.templateRequired') }}
          </p>
        </div>
      </div>
      <div v-if="collapsible || $slots.footer" class="grid-footer">
        <span class="footer-extra"><slot name="footer"></slot></span>
        <el-button v-if="collapsible" type="text" @click="expanded = false">
          {{ $t('common.button.fold') }}
        </el-button>
      </div>
    </template>
  </div>
</template>
<script>
import { getParseTemplateList } from '@/api/parseTemplate';
import FileIcon from '@/components/FileIcon.vue';
import { DOC_TYPE_LIST, getMediaType } from '../parseTemplate/config';

const CREATE_OPTION = '__create__';

export default {
  name: 'ParseTemplateSelect',
  components: { FileIcon },
  props: {
    // 各文档类型选定的模板，格式 {docType: templateId}
    value: { type: Object, default: () => ({}) },
    disabled: { type: Boolean, default: false },
    // 展示范围：all 全部 / doc 文本表格 / media 视频音频图片
    scope: {
      type: String,
      default: 'all',
      validator: v => ['all', 'doc', 'media'].includes(v),
    },
    // 新建多模态知识库时，替用户选上已有的第一个媒体模板
    autoBindMedia: { type: Boolean, default: false },
    // 上传文件场景下十三项太占版面，默认折叠
    collapsible: { type: Boolean, default: false },
    // 校验未通过时由父级传入，标红引导用户补选
    requiredDocTypes: { type: Array, default: () => [] },
  },
  data() {
    return {
      CREATE_OPTION,
      expanded: false,
      docTypeList: DOC_TYPE_LIST,
      bind: { ...this.value },
      grouped: {},
      loaded: false,
      needRefresh: false,
    };
  },
  computed: {
    visibleDocTypes() {
      if (this.scope === 'all') return this.docTypeList;
      const wantMedia = this.scope === 'media';
      return this.docTypeList.filter(
        item => this.isMedia(item.docType) === wantMedia,
      );
    },
    isOpen() {
      return !this.collapsible || this.expanded;
    },
    requiredSet() {
      return new Set(this.requiredDocTypes);
    },
    invalidSet() {
      return new Set(this.requiredDocTypes.filter(d => !this.selectValue(d)));
    },
    // 实际没有模板可用的类型，父级据此判断本次上传缺哪几项
    // 列表没回来之前一律算「还不知道」，否则会误判成全都没模板
    unboundDocTypes() {
      if (!this.loaded) return [];
      return this.visibleDocTypes
        .filter(item => !this.selectValue(item.docType))
        .map(item => item.docType);
    },
    // 收起时只展示偏离内置（默认）的那几项
    customTags() {
      return this.visibleDocTypes.reduce((acc, item) => {
        const templateId = this.bind[item.docType];
        if (!templateId) return acc;
        const template = (this.grouped[item.docType] || []).find(
          t => t.templateId === templateId,
        );
        // 列表未返回时取不到模板，先不展示，避免闪出原始 id
        if (!template || template.builtIn) return acc;
        acc.push({
          docType: item.docType,
          docTypeName: item.name,
          templateName: template.name,
        });
        return acc;
      }, []);
    },
  },
  watch: {
    value(val) {
      this.bind = { ...val };
    },
    // 媒体类型露出来了才谈得上自动绑定
    scope() {
      this.bindDefaultMediaTemplates();
    },
    unboundDocTypes: {
      immediate: true,
      handler(val) {
        this.$emit('unbound-change', val);
      },
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
    // 供父级在校验失败后展开，引导用户补选模板
    expand() {
      this.expanded = true;
    },
    isMedia(docType) {
      return getMediaType(docType) !== 'doc';
    },
    // 只有从没配过的类型才回落内置；手动清空过的留空
    isConfigured(docType) {
      return Object.hasOwn(this.bind, docType);
    },
    // 值必须在选项里存在，否则 el-select 会把原始 templateId 画出来
    selectValue(docType) {
      const templateId = this.isConfigured(docType)
        ? this.bind[docType]
        : this.builtInValue(docType);
      if (!templateId) return '';
      return this.optionTemplates(docType).some(
        t => t.templateId === templateId,
      )
        ? templateId
        : '';
    },
    // 内置（默认）模板由后端保底落库，取它自己的 id；媒体类型不认内置
    builtInValue(docType) {
      if (this.isMedia(docType)) return '';
      const builtIn = (this.grouped[docType] || []).find(item => item.builtIn);
      return builtIn ? builtIn.templateId : '';
    },
    // 顺序跟解析模板页一致：自建的按新到旧在前，内置（默认）沉底
    optionTemplates(docType) {
      const list = this.grouped[docType] || [];
      return this.isMedia(docType) ? list.filter(t => !t.builtIn) : list;
    },
    currentTemplate(docType) {
      const templateId = this.selectValue(docType);
      return (
        (this.grouped[docType] || []).find(
          item => item.templateId === templateId,
        ) || null
      );
    },
    // 另开标签编辑/新建，避免带走当前页已填的表单
    openTemplatePage(docType, create) {
      const template = this.currentTemplate(docType);
      const { href } = this.$router.resolve({
        path: '/knowledge/parseTemplate',
        query: {
          docType,
          templateId: create || !template ? undefined : template.templateId,
          create: create ? '1' : undefined,
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
          acc[template.docType] = acc[template.docType] || [];
          acc[template.docType].push(template);
          return acc;
        }, {});
        this.loaded = true;
        this.bindDefaultMediaTemplates();
      });
    },
    // 媒体类型没有内置模板，未绑定就用不了；取列表第一个（接口按创建时间倒序）
    bindDefaultMediaTemplates() {
      if (!this.autoBindMedia) return;
      let changed = false;
      this.visibleDocTypes.forEach(item => {
        if (!this.isMedia(item.docType) || this.isConfigured(item.docType)) {
          return;
        }
        const first = (this.grouped[item.docType] || [])[0];
        if (!first) return;
        this.$set(this.bind, item.docType, first.templateId);
        changed = true;
      });
      if (changed) this.$emit('input', { ...this.bind });
    },
    handleChange(docType, templateId) {
      if (templateId === CREATE_OPTION) {
        this.openTemplatePage(docType, true);
        return;
      }
      this.$set(this.bind, docType, templateId || '');
      this.$emit('input', { ...this.bind });
    },
  },
};
</script>
<style lang="scss" scoped>
.parse-template-select {
  .template-summary {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
    min-height: 32px;
  }

  .summary-tag {
    border-color: transparent;
  }

  .summary-empty {
    font-size: 12px;
    color: #606266;
  }

  .summary-edit {
    cursor: pointer;
    font-size: 16px;
    color: #5a6cf3;
  }

  .grid-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    margin-top: 8px;
  }

  .footer-extra {
    display: inline-flex;
    align-items: center;
  }

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

    .item-star {
      color: #f56c6c;
    }

    .item-error {
      margin: 4px 0 0;
      font-size: 12px;
      line-height: 1.2;
      color: #f56c6c;
    }

    // 未绑定的媒体类型标红引导，效果对齐 el-form-item 的 is-error
    &.is-error ::v-deep .el-input__inner {
      border-color: #f56c6c;

      &:focus {
        border-color: #f56c6c;
      }
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
