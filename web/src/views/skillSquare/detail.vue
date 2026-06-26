<template>
  <SkillDetail
    :detail="detail"
    :recommendList="recommendList"
    :isPublic="isPublic"
    :bgColor="bgColor"
    :backText="backText"
    :visibleVariableConfig="false"
    :visibleHistory="visibleHistory"
    :historyList="historyList"
    :visibleDownload="skillType !== 'mine' && !detail.isClosedSource"
    @init="initData"
    @back="handleBack"
    @download="handleDownload"
    @click-recommend="handleClickRecommend"
  >
    <template #info-header>
      <div class="apiKeyConfig-tips">
        <i class="el-icon-info"></i>
        <span>
          {{
            skillType === 'builtin'
              ? $t('skillSpace.detail.apiKeyEmptyTips_builtin')
              : $t('skillSpace.detail.apiKeyEmptyTips')
          }}
        </span>
      </div>
    </template>
    <template #extra-buttons>
      <el-button
        v-if="['shared'].includes(skillType)"
        type="primary"
        size="mini"
        plain
        :disabled="detail.isShared"
        @click="handleSendToResource(detail)"
      >
        {{
          detail.isShared
            ? $t('skillSpace.isShared')
            : $t('skillSpace.toResource')
        }}
      </el-button>
    </template>
  </SkillDetail>
</template>

<script>
import SkillDetail from '@/components/skills/skillDetail.vue';
import {
  getBuiltinSquareSkillList,
  sendSquareSkillToResource,
  getSquareSkillDetail,
  downloadSquareSkill,
  getSharedSquareSkillList,
  addSharedSkillToResource,
  getSharedSquareSkillDetail,
  downloadSharedSquareSkill,
  getCreatedSquareSkillList,
  getCreatedSquareSkillDetail,
  getCreatedSkillVersionList,
  getSharedSkillVersionList,
} from '@/api/skillSquare';
import { downloadBuiltinSkill } from '@/api/templateSquare';
import { resDownloadFile } from '@/utils/util';

export default {
  name: 'SkillSquareDetail',
  components: {
    SkillDetail,
  },
  data() {
    return {
      skillId: '',
      skillType: '',
      detail: {},
      recommendList: [],
      historyList: [],
      isPublic: false,
      bgColor:
        'linear-gradient(1deg, rgb(247, 252, 255) 50%, rgb(233, 246, 254) 98%)',
      backText: this.$t('skillSpace.detail.backText'),
    };
  },
  computed: {
    visibleHistory() {
      return ['shared', 'mine'].includes(this.skillType);
    },
  },
  created() {
    this.isPublic = this.$route.path.includes('/public/');
  },
  watch: {
    $route: {
      handler() {
        this.initData();
      },
      deep: true,
    },
  },
  methods: {
    initData() {
      const { skillId, skillType } = this.$route.query || {};
      this.skillId = skillId;
      this.skillType = skillType;

      this.getDetailData();
      this.getRecommendList();
      this.getHistoryList();

      // 滚动到顶部
      const main = document.querySelector('.el-main > .page-container');
      if (main) main.scrollTop = 0;
    },
    async getDetailData() {
      let requestApi = getSquareSkillDetail;
      let params = { skillId: this.skillId };
      if (this.skillType === 'shared') {
        requestApi = getSharedSquareSkillDetail;
      } else if (this.skillType === 'mine') {
        requestApi = getCreatedSquareSkillDetail;
        params = { customSkillId: this.skillId };
      }
      const res = await requestApi(params);
      this.detail = res.data || {};
    },
    async getRecommendList() {
      let requestApi = getBuiltinSquareSkillList;
      if (this.skillType === 'builtin') {
        requestApi = getBuiltinSquareSkillList;
      } else if (this.skillType === 'shared') {
        requestApi = getSharedSquareSkillList;
      } else if (this.skillType === 'mine') {
        requestApi = getCreatedSquareSkillList;
      }
      const res = await requestApi();
      this.recommendList =
        res.data.list.filter(item => item.skillId !== this.skillId) || [];
    },
    async getHistoryList() {
      this.historyList = [];
      if (!this.visibleHistory) return;

      const isMine = this.skillType === 'mine';
      const requestApi = isMine
        ? getCreatedSkillVersionList
        : getSharedSkillVersionList;
      const params = isMine
        ? { customSkillId: this.skillId }
        : { skillId: this.skillId };
      const res = await requestApi(params);
      const list = res.data?.list || [];
      this.historyList = list.map(item => ({
        ...item,
        updateTime: item.updateTime || item.updatedAt || item.createdAt,
      }));
    },
    async handleDownload(item) {
      let downloadApi = downloadSquareSkill;
      if (this.skillType === 'builtin') {
        downloadApi = downloadBuiltinSkill;
      } else if (this.skillType === 'shared') {
        downloadApi = downloadSharedSquareSkill;
      }
      const res = await downloadApi({ skillId: item.skillId });
      resDownloadFile(res, `${item.name}.zip`);
    },
    handleSendToResource(info) {
      const requestApi =
        this.skillType === 'shared'
          ? addSharedSkillToResource
          : sendSquareSkillToResource;

      requestApi({ skillId: info.skillId }).then(res => {
        if (res.code === 0) {
          this.$message.success(this.$t('common.info.send'));
          this.$set(this.detail, 'isShared', true);
        }
      });
    },
    handleBack() {
      const path = '/skillSquare';
      this.$router.push({ path });
    },
    handleClickRecommend(val) {
      const skillId = val.skillId;
      const query = { skillId: skillId };
      if (this.skillType === 'builtin') {
        query.skillType = 'builtin';
      } else if (this.skillType === 'shared') {
        query.skillType = 'shared';
      } else if (this.skillType === 'mine') {
        query.skillType = 'mine';
      }
      this.$router.push({
        path: this.$route.path,
        query,
      });
    },
  },
};
</script>

<style lang="scss" scoped>
.apiKeyConfig-tips {
  display: flex;
  align-items: flex-start;
  background: #f0f7ff;
  border: 1px solid #ddecff;
  border-left: 4px solid #409eff;
  color: #5e6d82;
  font-size: 13px;
  padding: 12px 16px;
  border-radius: 4px;
  margin: 15px 0;
  line-height: 1.6;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.02);

  i {
    font-size: 16px;
    color: #409eff;
    margin-right: 10px;
    margin-top: 2px;
    flex-shrink: 0;
  }

  span {
    flex: 1;
  }
}
</style>
