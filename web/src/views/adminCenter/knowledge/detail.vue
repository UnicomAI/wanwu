<template>
  <DetailLayout :back-title="backTitle" back-path="/adminCenter/knowledge">
    <DetailHeader
      :avatar="avatarUrl"
      :description="base.description"
      :name="base.name"
      :tags="headerTags"
    />

    <DetailCard :title="$t('adminCenter.common.basicInfo')">
      <InfoGrid :items="basicItems" />
    </DetailCard>

    <DetailCard :title="$t('adminCenter.common.config')">
      <QaDoclist
        v-if="base.category === QA"
        :knowledge-id="docQuery.knowledgeId"
        :list-api="getAdminKnowledgeQaPairList"
        :readonly="true"
        style="margin-left: -30px"
      />
      <Doclist
        v-else
        :knowledge-id="docQuery.knowledgeId"
        :list-api="getAdminKnowledgeFileList"
        :readonly="true"
        style="margin-left: -30px"
      />
    </DetailCard>
  </DetailLayout>
</template>

<script>
import {
  getAdminKnowledgeBase,
  getAdminKnowledgeFileList,
  getAdminKnowledgeQaPairList,
} from '@/api/adminCenter';
import { getParseTemplateList } from '@/api/parseTemplate';
import { avatarSrc } from '@/utils/util';
import { KNOWLEDGE, QA, MULTIMODAL } from '@/views/knowledge/constants';
import { DOC_TYPE_LIST } from '@/views/knowledge/parseTemplate/config';
import DetailLayout from '../components/DetailLayout.vue';
import DetailHeader from '../components/DetailHeader.vue';
import DetailCard from '../components/DetailCard.vue';
import InfoGrid from '../components/InfoGrid.vue';
import Doclist from '@/views/knowledge/knowledgeDatabase/doclist.vue';
import QaDoclist from '@/views/knowledge/qaDatabase/docList.vue';
import detailMixin from '../mixins/detailMixin';

export default {
  components: {
    DetailLayout,
    DetailHeader,
    DetailCard,
    InfoGrid,
    Doclist,
    QaDoclist,
  },
  mixins: [detailMixin],
  data() {
    return {
      base: {},
      parseTemplateList: [],
      getAdminKnowledgeFileList,
      getAdminKnowledgeQaPairList,
      QA,
      moduleTitleKey: 'adminCenter.pageModules.resourcePool.knowledge.title',
    };
  },
  computed: {
    docQuery() {
      return { knowledgeId: this.$route.query.knowledgeId };
    },
    headerTags() {
      const map = {
        [KNOWLEDGE]: this.$t('knowledgeManage.textKnowledgeDatabase.title'),
        [QA]: this.$t('knowledgeManage.qaDatabase.title'),
        [MULTIMODAL]: this.$t('knowledgeManage.multiKnowledgeDatabase.title'),
      };
      const text = map[this.base.category];
      return text ? [{ text, type: '' }] : [];
    },
    avatarUrl() {
      return avatarSrc(
        this.base.avatar?.path,
        require('@/assets/imgs/knowledgeIcon.png'),
      );
    },
    embeddingModel() {
      return this.base.embeddingModel || null;
    },
    keywords() {
      return this.base.keywords || [];
    },
    embeddingModelField() {
      const m = this.embeddingModel;
      if (!m) return null;
      return {
        icon: this.convertModelIcon(m?.avatar?.path),
        name: m.displayName,
        tags: m.tags || [],
      };
    },
    keywordTags() {
      return (this.keywords || []).map(k => ({
        text: `${k.name} : ${k.alias}`,
      }));
    },
    parseTemplateTags() {
      const bind = this.base.parseTemplate || [];
      return bind.reduce((acc, { docType, templateId }) => {
        if (!templateId) return acc;
        const template = this.parseTemplateList.find(
          item => item.templateId === templateId,
        );
        if (!template || template.builtIn) return acc;
        const docTypeItem = DOC_TYPE_LIST.find(
          item => item.docType === docType,
        );
        acc.push({
          text: `${docTypeItem ? docTypeItem.name : docType}：${template.name}`,
        });
        return acc;
      }, []);
    },
    basicItems() {
      const b = this.base;
      return [
        {
          label: this.$t(
            'adminCenter.pageModules.resourcePool.knowledge.detail.embeddingModel',
          ),
          model: this.embeddingModelField,
        },
        {
          label: this.$t(
            'adminCenter.pageModules.resourcePool.knowledge.detail.keywords',
          ),
          tags: this.keywordTags,
        },
        {
          label: this.$t(
            'adminCenter.pageModules.resourcePool.knowledge.detail.parseTemplate',
          ),
          tags: this.parseTemplateTags,
        },
        {
          label: this.$t('adminCenter.common.creator'),
          value: b.ownerUserName,
        },
        { label: this.$t('adminCenter.common.org'), value: b.ownerOrgName },
        {
          label: this.$t('adminCenter.common.updateTime'),
          value: b.updatedAt,
        },
      ];
    },
  },
  mounted() {
    this.fetchBase();
  },
  methods: {
    fetchBase() {
      const knowledgeId = this.$route.query.knowledgeId;
      if (!knowledgeId) return;
      getAdminKnowledgeBase({ knowledgeId }).then(res => {
        this.base = res?.data ?? {};
        this.fetchParseTemplates();
      });
    },
    fetchParseTemplates() {
      const bind = this.base.parseTemplate || [];
      if (!bind.length) return;
      getParseTemplateList().then(res => {
        if (res.code === 0) this.parseTemplateList = res.data.list || [];
      });
    },
  },
};
</script>
