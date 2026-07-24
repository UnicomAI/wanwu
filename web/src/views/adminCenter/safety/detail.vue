<template>
  <DetailLayout :back-title="backTitle" back-path="/adminCenter/safety">
    <DetailHeader
      :avatar="avatarUrl"
      :description="base.remark"
      :name="base.tableName"
      :tags="headerTags"
    />

    <DetailCard :title="$t('adminCenter.common.basicInfo')">
      <InfoGrid :items="basicItems" />
    </DetailCard>

    <DetailCard :title="$t('adminCenter.common.config')">
      <el-form class="config-form" label-position="left" label-width="120px">
        <el-form-item
          :label="
            $t('adminCenter.pageModules.resourcePool.safety.detail.reply')
          "
        >
          <el-input
            :autosize="{ minRows: 1, maxRows: 5 }"
            :value="replyText"
            disabled
            resize="none"
            style="padding: 0 30px"
            type="textarea"
          />
        </el-form-item>
        <el-form-item
          :label="
            $t('adminCenter.pageModules.resourcePool.safety.detail.wordList')
          "
        >
          <WordList
            :list-api="getAdminSensitiveWordDetail"
            :readonly="true"
            :table-id="tableId"
            @reply-change="onReply"
          />
        </el-form-item>
      </el-form>
    </DetailCard>
  </DetailLayout>
</template>

<script>
import {
  getAdminSensitiveWordBase,
  getAdminSensitiveWordDetail,
} from '@/api/adminCenter';
import { avatarSrc } from '@/utils/util';
import WordList from '@/views/safety/component/wordList.vue';
import DetailLayout from '../components/DetailLayout.vue';
import DetailHeader from '../components/DetailHeader.vue';
import DetailCard from '../components/DetailCard.vue';
import InfoGrid from '../components/InfoGrid.vue';
import detailMixin from '../mixins/detailMixin';

export default {
  components: {
    DetailLayout,
    DetailHeader,
    DetailCard,
    InfoGrid,
    WordList,
  },
  mixins: [detailMixin],
  data() {
    return {
      base: {},
      reply: '',
      moduleTitleKey: 'adminCenter.pageModules.resourcePool.safety.title',
    };
  },
  computed: {
    tableId() {
      return this.$route.query.tableId;
    },
    avatarUrl() {
      return avatarSrc(
        this.base.avatar?.path,
        require('@/assets/imgs/safety.png'),
      );
    },
    headerTags() {
      const map = {
        personal: this.$t('safety.personal'),
        global: this.$t('safety.global'),
      };
      const text = map[this.base.type];
      return text ? [{ text, type: '' }] : [];
    },
    replyText() {
      return this.reply || this.$t('safety.setReply.defaultReply');
    },
  },
  mounted() {
    this.fetchBase();
  },
  methods: {
    getAdminSensitiveWordDetail,
    fetchBase() {
      const tableId = this.tableId;
      if (!tableId) return;
      getAdminSensitiveWordBase({ tableId }).then(res => {
        this.base = res?.data ?? {};
      });
    },
    onReply(val) {
      this.reply = val;
    },
  },
};
</script>

<style lang="scss" scoped>
.config-form {
  margin-bottom: 20px;
}

::v-deep .el-textarea.is-disabled .el-textarea__inner {
  color: #606266;
}
</style>
