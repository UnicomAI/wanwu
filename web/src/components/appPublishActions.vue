<template>
  <div class="app-publish-actions">
    <!-- 版本历史 -->
    <VersionPopover
      ref="versionPopover"
      v-if="publishType"
      style="pointer-events: auto"
      :appId="appId"
      :appType="appType"
      @reloadData="reloadData"
      @previewVersion="previewVersion"
      @export="handleExport"
    />
    <!-- 发布配置 -->
    <el-button
      v-if="publishType"
      size="small"
      type="primary"
      style="padding: 9px 12px; margin-left: 8px"
      @click="handlePublishSet"
    >
      <span class="el-icon-setting"></span>
      {{ $t('agent.form.publishConfig') }}
    </el-button>
    <!-- 发布按钮弹窗 -->
    <el-popover
      v-model="publishPopoverVisible"
      placement="bottom-end"
      trigger="click"
      style="margin-left: 8px"
    >
      <el-button
        slot="reference"
        size="small"
        type="primary"
        :disabled="disabled"
        style="padding: 9px 12px"
      >
        {{ $t('common.button.publish') }}
        <span class="el-icon-arrow-down" style="margin-left: 5px"></span>
      </el-button>
      <el-form ref="publishForm" :model="publishForm" :rules="publishRules">
        <el-form-item :label="$t('list.version.no')" prop="version">
          <el-input
            v-model="publishForm.version"
            :placeholder="$t('list.version.noPlaceholder')"
          ></el-input>
        </el-form-item>
        <el-form-item :label="$t('list.version.desc')" prop="desc">
          <el-input
            v-model="publishForm.desc"
            :placeholder="$t('list.version.descPlaceholder')"
          ></el-input>
        </el-form-item>
        <el-form-item
          :label="$t('list.version.publishType')"
          prop="publishType"
        >
          <el-radio-group v-model="publishForm.publishType">
            <div>
              <el-radio label="private">
                {{
                  appType === 'agent'
                    ? $t('agent.form.publishType')
                    : $t('app.commonPublishType.private')
                }}
              </el-radio>
            </div>
            <div>
              <el-radio label="organization">
                {{
                  appType === 'agent'
                    ? $t('agent.form.publishType1')
                    : $t('app.commonPublishType.organization')
                }}
              </el-radio>
            </div>
            <div>
              <el-radio label="public">
                {{
                  appType === 'agent'
                    ? $t('agent.form.publishType2')
                    : $t('app.commonPublishType.public')
                }}
              </el-radio>
            </div>
          </el-radio-group>
        </el-form-item>
        <div class="saveBtn" style="text-align: right; margin-top: 10px">
          <el-button
            size="mini"
            type="primary"
            :loading="savingPublish"
            :disabled="disabled"
            @click="savePublish"
          >
            {{ $t('common.button.save') }}
          </el-button>
        </div>
      </el-form>
    </el-popover>
  </div>
</template>

<script>
import VersionPopover from '@/components/versionPopover.vue';
import { appPublish } from '@/api/appspace';
import { downloadCustomSkillVersion } from '@/api/templateSquare';
import { resDownloadFile } from '@/utils/util';

export default {
  name: 'AppPublishActions',
  components: {
    VersionPopover,
  },
  props: {
    appId: {
      type: String,
      required: true,
    },
    appType: {
      type: String,
      required: true,
    },
    appName: {
      type: String,
      default: '',
    },
    publishType: {
      type: String,
      default: '',
    },
    beforePublish: {
      type: Function,
      default: null,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      publishPopoverVisible: false,
      savingPublish: false,
      publishForm: {
        publishType: 'private',
        version: '',
        desc: '',
      },
    };
  },
  computed: {
    publishRules() {
      return {
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
      };
    },
  },
  watch: {
    disabled(value) {
      if (value) {
        this.publishPopoverVisible = false;
      }
    },
  },
  methods: {
    reloadData() {
      this.$emit('reload-data');
    },
    previewVersion(item) {
      this.$emit('preview-version', item);
    },
    handleExport(item) {
      if (this.appType === 'skill') {
        downloadCustomSkillVersion({
          skillId: this.appId,
          version: item.version,
        }).then(res => {
          resDownloadFile(res, `${this.appName || ''}_${item.version}.zip`);
        });
      }
    },
    handlePublishSet() {
      const routeMap = {
        agent: '/agent/publishSet',
        rag: '/rag/publishSet',
        skill: '/generalAgent/skills/publishConfig',
      };

      const targetPath = routeMap[this.appType];
      this.$router.push({
        path: targetPath,
        query: {
          appId: this.appId,
          appType: this.appType,
          name: this.appName,
        },
      });
    },
    validatePublishForm() {
      return new Promise(resolve => {
        this.$refs.publishForm.validate(valid => resolve(valid));
      });
    },
    getPublishData() {
      return {
        appId: this.appId,
        appType: this.appType,
        publishType: this.publishForm.publishType,
        desc: this.publishForm.desc,
        version: this.publishForm.version,
      };
    },
    async executePublish(data) {
      if (this.disabled) return null;

      const res = await appPublish(data);
      if (res.code === 0) {
        if (this.appType === 'skill') {
          this.$router.push('/skillSquare?type=mine');
        } else {
          this.$router.push({ path: '/explore' });
        }
      }
      return res;
    },
    async savePublish() {
      if (this.disabled || this.savingPublish) return;

      const valid = await this.validatePublishForm();
      if (!valid) return;

      const data = this.getPublishData();
      this.savingPublish = true;
      try {
        let shouldPublish = true;
        if (this.beforePublish) {
          try {
            shouldPublish = (await this.beforePublish({ ...data })) !== false;
          } catch (error) {
            console.error('beforePublish error', error);
            shouldPublish = false;
          }
        }

        if (this.disabled || !shouldPublish) {
          this.publishPopoverVisible = false;
          return;
        }

        await this.executePublish(data);
      } finally {
        this.savingPublish = false;
      }
    },
  },
};
</script>

<style lang="scss" scoped>
.app-publish-actions {
  display: flex;
  align-items: center;
}
</style>
