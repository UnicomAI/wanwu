<template>
  <div class="version-detail-container">
    <div class="version-detail">
      <h3>{{ $t('list.version.detail') }}</h3>
      <div class="detail-form">
        <el-form
          ref="publishForm"
          :model="publishForm"
          :rules="publishRules"
          label-position="top"
        >
          <el-form-item :label="$t('list.version.no')" prop="version">
            <el-input v-model="publishForm.version" :disabled="true"></el-input>
          </el-form-item>
          <el-form-item :label="$t('list.version.desc')" prop="desc">
            <el-input
              v-model="publishForm.desc"
              type="textarea"
              :rows="3"
              :placeholder="$t('list.version.descPlaceholder')"
            ></el-input>
          </el-form-item>
          <el-form-item
            :label="$t('list.version.publishType')"
            prop="publishType"
          >
            <el-radio-group v-model="publishForm.publishType">
              <div class="radio-item">
                <el-radio label="private">
                  {{ $t('tempSquare.skills.publishType') }}
                </el-radio>
              </div>
              <div class="radio-item">
                <el-radio label="organization">
                  {{ $t('tempSquare.skills.publishType1') }}
                </el-radio>
              </div>
              <div class="radio-item">
                <el-radio label="public">
                  {{ $t('tempSquare.skills.publishType2') }}
                </el-radio>
              </div>
            </el-radio-group>
          </el-form-item>

          <el-form-item :label="$t('tempSquare.skills.closedSource')">
            <el-switch v-model="publishForm.isClosedSource" />
            <span class="closed-source-tip">{{
              $t('tempSquare.skills.closedSourceTip')
            }}</span>
          </el-form-item>

          <div class="save-action">
            <el-button
              size="medium"
              type="primary"
              @click="savePublish"
              :loading="saving"
            >
              {{ $t('common.button.save') }}
            </el-button>
          </div>
        </el-form>
      </div>
    </div>

    <div class="version-history">
      <h3>{{ $t('list.version.history') }}</h3>
      <VersionTimeLine
        ref="versionTimeline"
        :appId="appId"
        :appType="appType"
        where="webUrl"
        @export="handleExport"
      />
    </div>
  </div>
</template>

<script>
import VersionTimeLine from '@/components/versionTimeLine.vue';
import { getAppLatestVersion, updateAppVersion } from '@/api/appspace';
import { downloadCustomSkillVersion } from '@/api/templateSquare';
import { resDownloadFile } from '@/utils/util';

export default {
  name: 'SkillCreateScope',
  props: {
    appType: {
      type: String,
      required: true,
    },
    appId: {
      type: String,
      required: true,
    },
  },
  components: {
    VersionTimeLine,
  },
  data() {
    return {
      saving: false,
      publishForm: {
        publishType: 'private',
        version: '',
        desc: '',
        isClosedSource: false,
      },
      publishRules: {
        version: [
          {
            required: true,
            message: this.$t('list.version.noMsg'),
            trigger: 'blur',
          },
        ],
        desc: [
          {
            required: true,
            message: this.$t('list.version.descPlaceholder'),
            trigger: 'blur',
          },
        ],
        publishType: [
          {
            required: true,
            message: this.$t('common.select.placeholder'),
            trigger: 'change',
          },
        ],
      },
    };
  },
  created() {
    this.fetchLatestVersion();
  },
  methods: {
    fetchLatestVersion() {
      if (!this.appId) return;
      getAppLatestVersion({
        appId: this.appId,
        appType: this.appType,
      }).then(res => {
        if (res.code === 0 && res.data) {
          this.publishForm = {
            ...this.publishForm,
            ...res.data,
          };
        }
      });
    },
    savePublish() {
      this.$refs.publishForm.validate(valid => {
        if (valid) {
          this.saving = true;
          updateAppVersion({
            appId: this.appId,
            appType: this.appType,
            desc: this.publishForm.desc,
            publishType: this.publishForm.publishType,
            isClosedSource: this.publishForm.isClosedSource,
          })
            .then(res => {
              if (res.code === 0) {
                this.$message.success(this.$t('common.info.save'));
                if (this.$refs.versionTimeline) {
                  this.$refs.versionTimeline.getAppVersionList();
                }
              }
            })
            .finally(() => {
              this.saving = false;
            });
        }
      });
    },
    handleExport(item) {
      downloadCustomSkillVersion({
        skillId: this.appId,
        version: item.version,
      }).then(response => {
        resDownloadFile(
          response,
          `${this.$route.query.name || ''}_${item.version}.zip`,
        );
      });
    },
  },
};
</script>

<style lang="scss" scoped>
.version-detail-container {
  padding: 20px;
  overflow-y: auto;
  height: 100%;
}

.version-detail,
.version-history {
  border: 1px solid #e6e6e6;
  border-radius: 8px;
  margin-bottom: 20px;
  padding: 20px;
  background: #fff;

  h3 {
    font-size: 18px;
    margin-bottom: 20px;
    border-bottom: 1px solid #e6e6e6;
    padding-bottom: 15px;
    font-weight: 600;
  }
}

.detail-form {
  max-width: 600px;

  .radio-item {
    margin-bottom: 12px;
    &:last-child {
      margin-bottom: 0;
    }
  }

  .save-action {
    margin-top: 30px;
    padding-top: 20px;
    border-top: 1px dashed #eee;
  }
}

.version-history {
  margin-top: 20px;
}

.closed-source-tip {
  margin-left: 10px;
  color: #999;
  font-size: 12px;
}
</style>
