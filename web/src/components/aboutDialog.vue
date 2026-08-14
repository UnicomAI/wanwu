<template>
  <el-dialog
    :visible="visible"
    width="720px"
    append-to-body
    :close-on-click-modal="false"
    :before-close="handleClose"
    custom-class="about-dialog"
    :title="$t('menu.about')"
  >
    <div class="about-wrap">
      <div class="about-info-row">
        <span class="about-label">{{ $t('about.currentVersion') }}</span>
        <span class="about-version-value">
          <span class="version-dot"></span>
          {{ about.version || '1.0' }}
        </span>
      </div>
      <div class="about-section-title">{{ $t('about.log') }}</div>
      <div class="about-changelog">
        <MdRender :content="content || $t('common.noData')" />
      </div>
    </div>
  </el-dialog>
</template>

<script>
import { mapGetters } from 'vuex';
import { avatarSrc } from '@/utils/util';
import MdRender from '@/components/mdRender.vue';
import { getAboutDetail } from '@/api/user';

export default {
  components: {
    MdRender,
  },
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      about: {},
      content: '',
    };
  },
  watch: {
    visible(value) {
      if (value) this.fetchAboutDetail();
    },
    commonInfo: {
      handler(val) {
        const { about } = val.data || {};
        this.about = about || {};
      },
      deep: true,
    },
  },
  computed: {
    ...mapGetters('user', ['commonInfo']),
  },
  methods: {
    avatarSrc,
    fetchAboutDetail() {
      getAboutDetail().then(res => {
        this.content = res.data?.releaseNotes || '';
      });
    },
    handleClose(done) {
      this.$emit('update:visible', false);
      done();
    },
  },
};
</script>

<style lang="scss" scoped>
.about-wrap {
  padding: 0 0 12px;
  margin-top: -20px;
}
.about-info-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}
.about-label {
  font-size: 14px;
  color: #333;
  font-weight: 500;
}
.about-version-value {
  font-size: 16px;
  font-weight: 600;
  color: #121212;
  display: flex;
  align-items: center;
  gap: 6px;
}
.version-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #f59e0b;
  display: inline-block;
  flex-shrink: 0;
}
.about-section-title {
  font-size: 14px;
  color: #333;
  font-weight: 500;
  margin-bottom: 10px;
}
.about-changelog {
  background: #fafafa;
  border-radius: 8px;
  padding: 18px 22px;
  max-height: 400px;
  overflow-y: auto;
}
.about-dialog ::v-deep {
  border-radius: 12px !important;
  .el-dialog__header {
    padding: 20px 24px 10px;
    border-bottom: none;
  }
  .el-dialog__body {
    padding: 8px 24px 20px;
  }
  .el-dialog__headerbtn {
    display: none;
  }
}
</style>
