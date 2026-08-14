<template>
  <div class="version-detail-container">
    <div class="version-detail">
      <h3>{{ $t('list.version.detail') }}</h3>
      <div class="detail-form">
        <el-form ref="publishForm" :model="publishForm" :rules="publishRules">
          <el-form-item :label="$t('list.version.no')" prop="version">
            <el-input v-model="publishForm.version" :disabled="true"></el-input>
          </el-form-item>
          <el-form-item :label="$t('list.version.desc')" prop="desc">
            <el-input
              v-model="publishForm.desc"
              :placeholder="$t('list.version.descPlaceholder')"
              :readonly="isPublished"
            ></el-input>
          </el-form-item>
          <el-form-item
            v-if="appType === AGENT"
            :label="$t('agent.form.hideKnowledge')"
          >
            <el-switch
              v-model="publishForm.extra.hideKnowledge"
              :active-value="1"
              :inactive-value="0"
              active-color="var(--color)"
              :disabled="isPublished"
            ></el-switch>
          </el-form-item>
          <el-form-item
            :label="$t('list.version.publishType')"
            prop="publishType"
          >
            <el-radio-group
              v-model="publishForm.publishType"
              :disabled="isPublished"
            >
              <div>
                <el-radio label="private">
                  {{ $t('agent.form.publishType') }}
                </el-radio>
              </div>
              <div>
                <el-radio label="organization">
                  {{ $t('agent.form.publishType1') }}
                </el-radio>
              </div>
              <div>
                <el-radio label="public">
                  {{ $t('agent.form.publishType2') }}
                </el-radio>
              </div>
            </el-radio-group>
          </el-form-item>

          <div v-if="!isPublished" class="saveBtn">
            <el-button size="mini" type="primary" @click="savePublish">
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
      />
    </div>
  </div>
</template>

<script>
import VersionTimeLine from '@/components/versionTimeLine.vue';
import { getAppLatestVersion, updateAppVersion } from '@/api/appspace';
import { AGENT } from '@/utils/commonSet';

export default {
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
      AGENT,
      publishForm: {
        publishType: 'private',
        version: '',
        desc: '',
        extra: {
          hideKnowledge: 0,
        },
      },
      isPublished: false,
      publishRules: {
        version: [
          {
            required: true,
            message: this.$t('list.version.noMsg'),
            trigger: 'blur',
          },
          {
            pattern: /^v\d+\.\d+\.\d+$/,
            message: this.$t('list.version.versionMsg'),
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
    getAppLatestVersion({
      appId: this.appId,
      appType: this.appType,
    }).then(res => {
      const rawExtra = res.data && res.data.extra;
      let responseExtra = {};
      if (typeof rawExtra === 'string') {
        try {
          responseExtra = JSON.parse(rawExtra || '{}');
        } catch (e) {
          responseExtra = {};
        }
      } else {
        responseExtra = rawExtra || {};
      }
      const extra = {
        ...this.publishForm.extra,
        ...responseExtra,
      };
      this.publishForm = {
        ...this.publishForm,
        ...res.data,
        extra: {
          ...extra,
          hideKnowledge: Number(extra.hideKnowledge) === 1 ? 1 : 0,
        },
      };
      this.isPublished = Boolean(res.data && res.data.publishType);
    });
  },
  methods: {
    savePublish() {
      this.$refs.publishForm.validate(valid => {
        if (valid) {
          const data = {
            appId: this.appId,
            appType: this.appType,
            desc: this.publishForm.desc,
            publishType: this.publishForm.publishType,
          };
          if (this.appType === AGENT) {
            data.extra = JSON.stringify({
              ...(this.publishForm.extra || {}),
              hideKnowledge: this.publishForm.extra.hideKnowledge ? 1 : 0,
            });
          }
          updateAppVersion(data).then(res => {
            if (res.code === 0) {
              this.$message.success(this.$t('common.info.save'));
              this.$refs.versionTimeline.getAppVersionList();
            }
          });
        }
      });
    },
  },
};
</script>

<style scoped>
.version-detail-container {
  padding: 20px;
  overflow-y: auto;
  height: calc(100vh - 120px);
}

.version-detail,
.version-history {
  border: 1px solid #e6e6e6;
  border-radius: 8px;
  margin-bottom: 20px;
  padding: 20px;
}

.version-detail h3,
.version-history h3 {
  font-size: 18px;
  margin-bottom: 20px;
  border-bottom: 1px solid #e6e6e6;
  padding-bottom: 10px;
}

.detail-form {
  margin-top: 20px;
}

.version-history {
  margin-top: 20px;
}
</style>
