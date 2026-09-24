<template>
  <div style="padding-bottom: 5px">
    <!--tab配置-->
    <el-card shadow="never" class="docPage-card">
      <div slot="header">
        <span class="card-title">{{ $t('infoSetting.tabSet') }}</span>
      </div>
      <el-form
        label-width="150px"
        :model="tabForm"
        :rules="tabRules"
        ref="tabForm"
      >
        <el-form-item
          :label="$t('infoSetting.form.labelTitle')"
          prop="tabTitle"
        >
          <el-input v-model="tabForm.tabTitle" style="width: 300px" />
          <div style="font-size: 11px; color: #aaa; margin-top: -9px">
            {{ $t('infoSetting.hint.labelTitle') }}
          </div>
        </el-form-item>
        <el-form-item :label="$t('infoSetting.form.labelIcon')" prop="tabLogo">
          <el-upload
            class="avatar-uploader"
            action=""
            name="files"
            :show-file-list="false"
            :http-request="handleUploadLabelIcon"
            :on-error="handleUploadError"
            accept=".png,.jpg,.jpeg"
          >
            <img
              v-if="tabForm.tabLogo.path"
              :src="
                getLogoPath(tabForm.tabLogo.path, tabForm.tabLogo.previewUrl)
              "
              class="avatar"
            />
            <i
              v-if="!tabForm.tabLogo.path"
              class="el-icon-plus avatar-uploader-icon"
            ></i>
            <span style="margin-left: 12px; color: #aaa !important">
              {{ $t('infoSetting.hint.imgUpload') }}，{{
                $t('infoSetting.hint.logoSizeLimit')
              }}
            </span>
            <div
              style="
                text-align: left;
                font-size: 11px;
                color: #aaa;
                margin-top: -9px;
              "
            >
              {{ $t('infoSetting.hint.labelIcon') }}
            </div>
          </el-upload>
        </el-form-item>
      </el-form>
      <div class="card-footer">
        <el-button
          :loading="tabLoading"
          type="primary"
          @click="handleSubmitTab()"
        >
          {{ $t('infoSetting.form.save') }}
        </el-button>
      </div>
    </el-card>
    <!--登录页配置-->
    <el-card shadow="never" class="docPage-card">
      <div slot="header">
        <span class="card-title">{{ $t('infoSetting.loginBgSet') }}</span>
      </div>
      <el-form
        label-width="150px"
        :model="loginForm"
        :rules="loginRules"
        ref="loginForm"
      >
        <el-form-item :label="$t('infoSetting.form.loginBg')" prop="loginBg">
          <el-upload
            class="avatar-uploader"
            action=""
            name="files"
            :show-file-list="false"
            :http-request="handleUploadLoginBg"
            :on-error="handleUploadError"
            accept=".png,.jpg,.jpeg"
          >
            <img
              v-if="loginForm.loginBg.path"
              :src="
                getLogoPath(
                  loginForm.loginBg.path,
                  loginForm.loginBg.previewUrl,
                )
              "
              class="avatar"
            />
            <i
              v-if="!loginForm.loginBg.path"
              class="el-icon-plus avatar-uploader-icon"
            ></i>
            <span style="margin-left: 12px; color: #aaa !important">
              {{ $t('infoSetting.hint.imgUpload') }}，{{
                $t('infoSetting.hint.bgSizeLimit')
              }}
            </span>
            <div
              style="
                text-align: left;
                font-size: 11px;
                color: #aaa;
                margin-top: -9px;
              "
            >
              {{ $t('infoSetting.hint.loginBg') }}
            </div>
          </el-upload>
        </el-form-item>
        <el-form-item
          :label="$t('infoSetting.form.loginLogo')"
          prop="loginLogo"
        >
          <el-upload
            class="avatar-uploader"
            action=""
            name="files"
            :show-file-list="false"
            :http-request="handleUploadLoginLogo"
            :on-error="handleUploadError"
            accept=".png,.jpg,.jpeg"
          >
            <img
              v-if="loginForm.loginLogo.path"
              :src="
                getLogoPath(
                  loginForm.loginLogo.path,
                  loginForm.loginLogo.previewUrl,
                )
              "
              class="avatar"
            />
            <i
              v-if="!loginForm.loginLogo.path"
              class="el-icon-plus avatar-uploader-icon"
            ></i>
            <span style="margin-left: 12px; color: #aaa !important">
              {{ $t('infoSetting.hint.imgUpload') }}，{{
                $t('infoSetting.hint.logoSizeLimit')
              }}
            </span>
            <div
              style="
                text-align: left;
                font-size: 11px;
                color: #aaa;
                margin-top: -9px;
              "
            >
              {{ $t('infoSetting.hint.loginLogo') }}
            </div>
          </el-upload>
        </el-form-item>
        <el-form-item
          :label="$t('infoSetting.form.loginButtonColor')"
          prop="loginButtonColor"
        >
          <el-color-picker
            v-model="loginForm.loginButtonColor"
            show-alpha
          ></el-color-picker>
          <div
            style="
              text-align: left;
              font-size: 11px;
              color: #aaa;
              margin-top: -9px;
            "
          >
            {{ $t('infoSetting.hint.loginButtonColor') }}
          </div>
        </el-form-item>
      </el-form>
      <div class="card-footer">
        <el-button
          :loading="loginLoading"
          type="primary"
          @click="handelSubmitLogin()"
        >
          {{ $t('infoSetting.form.save') }}
        </el-button>
      </div>
    </el-card>
    <!--平台配置-->
    <el-card shadow="never" class="docPage-card">
      <div slot="header">
        <span class="card-title">{{ $t('infoSetting.platformSet') }}</span>
      </div>
      <el-form ref="form" :model="form" :rules="rules" label-width="150px">
        <el-form-item
          :label="$t('infoSetting.form.platformTitle')"
          prop="homeName"
        >
          <el-input v-model="form.homeName" style="width: 300px" />
          <div style="font-size: 11px; color: #aaa; margin-top: -9px">
            {{ $t('infoSetting.hint.platformTitle') }}
          </div>
        </el-form-item>
        <el-form-item :label="$t('infoSetting.form.logo')" prop="homeLogo">
          <el-upload
            class="avatar-uploader"
            action=""
            name="files"
            :show-file-list="false"
            :http-request="handleUploadLogo"
            :on-error="handleUploadError"
            accept=".png,.jpg,.jpeg"
          >
            <img
              v-if="form.homeLogo.path"
              :src="getLogoPath(form.homeLogo.path, form.homeLogo.previewUrl)"
              class="avatar"
            />
            <i
              v-if="!form.homeLogo.path"
              class="el-icon-plus avatar-uploader-icon"
            ></i>
            <span style="margin-left: 12px; color: #aaa !important">
              {{ $t('infoSetting.hint.imgUpload') }}，{{
                $t('infoSetting.hint.logoSizeLimit')
              }}
            </span>
            <div
              style="
                text-align: left;
                font-size: 11px;
                color: #aaa;
                margin-top: -9px;
              "
            >
              {{ $t('infoSetting.hint.logo') }}
            </div>
          </el-upload>
        </el-form-item>
        <el-form-item
          :label="$t('infoSetting.form.bgColor')"
          prop="homeBgColor"
        >
          <el-radio-group v-model="radio">
            <el-radio label="0">{{ $t('infoSetting.hint.oneColor') }}</el-radio>
            <el-radio label="1">
              {{ $t('infoSetting.hint.linearColor') }}
            </el-radio>
          </el-radio-group>
          <div v-if="radio === '0'">
            <el-color-picker
              v-model="form.homeBgColor"
              show-alpha
            ></el-color-picker>
          </div>
          <div v-if="radio === '1'">
            <el-input v-model="form.homeBgColor" style="width: 300px" />
            <span
              :style="`display: inline-block; vertical-align: middle; width: 32px; height: 32px; border-radius: 2px; margin-left: 14px; background: ${form.homeBgColor}`"
            ></span>
          </div>
          <div
            style="
              text-align: left;
              font-size: 11px;
              color: #aaa;
              margin-top: -9px;
            "
          >
            {{ $t('infoSetting.hint.bgColor') }}
          </div>
        </el-form-item>
      </el-form>
      <div class="card-footer">
        <el-button
          :loading="homeLoading"
          type="primary"
          @click="handleSubmit()"
        >
          {{ $t('infoSetting.form.save') }}
        </el-button>
      </div>
    </el-card>
    <!--通用智能体配置-->
    <el-card class="docPage-card" shadow="never">
      <div slot="header">
        <span class="card-title">{{ $t('infoSetting.generalAgentSet') }}</span>
      </div>
      <el-form
        ref="generalAgentForm"
        :model="generalAgentForm"
        :rules="generalAgentRules"
        label-width="150px"
      >
        <el-form-item
          :label="$t('infoSetting.form.generalAgentMenuName')"
          prop="generalAgentMenuName"
        >
          <el-input
            v-model="generalAgentForm.generalAgentMenuName"
            style="width: 300px"
          />
          <div style="font-size: 11px; color: #aaa; margin-top: -9px">
            {{ $t('infoSetting.hint.generalAgentMenuName') }}
          </div>
        </el-form-item>
        <el-form-item
          :label="$t('infoSetting.form.generalAgentIcon')"
          prop="generalAgentIcon"
        >
          <el-upload
            :http-request="handleUploadGeneralAgentIcon"
            :on-error="handleUploadError"
            :show-file-list="false"
            accept=".png,.jpg,.jpeg"
            action=""
            class="avatar-uploader"
            name="files"
          >
            <img
              v-if="generalAgentForm.generalAgentIcon.path"
              :src="
                getLogoPath(
                  generalAgentForm.generalAgentIcon.path,
                  generalAgentForm.generalAgentIcon.previewUrl,
                )
              "
              class="avatar"
            />
            <i
              v-if="!generalAgentForm.generalAgentIcon.path"
              class="el-icon-plus avatar-uploader-icon"
            ></i>
            <span style="margin-left: 12px; color: #aaa !important">
              {{ $t('infoSetting.hint.imgUpload') }}，{{
                $t('infoSetting.hint.logoSizeLimit')
              }}
            </span>
            <div
              style="
                text-align: left;
                font-size: 11px;
                color: #aaa;
                margin-top: -9px;
              "
            >
              {{ $t('infoSetting.hint.generalAgentIcon') }}
            </div>
          </el-upload>
        </el-form-item>
        <el-form-item
          :label="$t('infoSetting.form.generalAgentWelcome')"
          prop="generalAgentWelcome"
        >
          <el-input
            v-model="generalAgentForm.generalAgentWelcome"
            style="width: 300px"
          />
          <div style="font-size: 11px; color: #aaa; margin-top: -9px">
            {{ $t('infoSetting.hint.generalAgentWelcome') }}
          </div>
        </el-form-item>
      </el-form>
      <div class="card-footer">
        <el-button
          :loading="generalAgentLoading"
          type="primary"
          @click="handleSubmitGeneralAgent()"
        >
          {{ $t('infoSetting.form.save') }}
        </el-button>
      </div>
    </el-card>
  </div>
</template>
<script>
import { setPlatformInfo } from '@/api/setInfo';
import { uploadAvatar } from '@/api/user';
import { mapActions, mapGetters } from 'vuex';
import { avatarSrc } from '@/utils/util';

export default {
  data() {
    const checkImage = (val, rule, value, callback) => {
      // 判断编辑器值，编辑器默认没输入时是 <p><br></p>
      if (val) {
        return callback();
      } else {
        callback(new Error(this.$t('infoSetting.form.uploadHint')));
      }
    };
    return {
      tabLoading: false,
      homeLoading: false,
      loginLoading: false,
      generalAgentLoading: false,
      tabForm: {
        tabTitle: '',
        tabLogo: {},
      },
      form: {
        homeName: '',
        homeLogo: {},
        homeBgColor: '',
      },
      loginForm: {
        loginBg: {},
        loginLogo: {},
        loginButtonColor: '',
      },
      generalAgentForm: {
        generalAgentIcon: {},
        generalAgentMenuName: '',
        generalAgentWelcome: '',
      },
      radio: '0',
      tabRules: {
        tabTitle: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
        ],
        tabLogo: [
          {
            required: true,
            validator: (...params) =>
              checkImage(this.tabForm.tabLogo.path, ...params),
            trigger: 'change',
          },
        ],
      },
      rules: {
        homeName: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
        ],
        homeLogo: [
          {
            required: true,
            validator: (...params) =>
              checkImage(this.form.homeLogo.path, ...params),
            trigger: 'change',
          },
        ],
        homeBgColor: [
          {
            required: true,
            message: this.$t('common.select.placeholder'),
            trigger: 'change',
          },
        ],
      },
      loginRules: {
        loginBg: [
          {
            required: true,
            validator: (...params) =>
              checkImage(this.loginForm.loginBg.path, ...params),
            trigger: 'change',
          },
        ],
        loginLogo: [
          {
            required: true,
            validator: (...params) =>
              checkImage(this.loginForm.loginLogo.path, ...params),
            trigger: 'change',
          },
        ],
        loginButtonColor: [
          {
            required: true,
            message: this.$t('common.select.placeholder'),
            trigger: 'change',
          },
        ],
      },
      generalAgentRules: {
        generalAgentMenuName: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
        ],
        generalAgentIcon: [
          {
            required: true,
            validator: (...params) =>
              checkImage(
                this.generalAgentForm.generalAgentIcon.path,
                ...params,
              ),
            trigger: 'change',
          },
        ],
        generalAgentWelcome: [
          {
            required: true,
            message: this.$t('common.input.placeholder'),
            trigger: 'blur',
          },
        ],
      },
    };
  },
  created() {
    this.getCommonInfo();
  },
  beforeDestroy() {
    // 释放本地预览的Blob URL，避免内存泄漏
    const forms = [
      this.tabForm.tabLogo,
      this.loginForm.loginBg,
      this.loginForm.loginLogo,
      this.form.homeLogo,
      this.generalAgentForm.generalAgentIcon,
    ];
    forms.forEach(item => {
      if (item && item.previewUrl) URL.revokeObjectURL(item.previewUrl);
    });
  },
  watch: {
    commonInfo: {
      handler(val) {
        const {
          home = {},
          tab = {},
          login = {},
          generalAgent = {},
        } = val ? val.data || {} : {};

        // 服务端数据刷新时释放旧的本地预览URL
        [
          this.tabForm.tabLogo,
          this.form.homeLogo,
          this.loginForm.loginBg,
          this.loginForm.loginLogo,
          this.generalAgentForm.generalAgentIcon,
        ].forEach(item => {
          if (item && item.previewUrl) URL.revokeObjectURL(item.previewUrl);
        });

        this.tabForm.tabTitle = tab.title;
        this.tabForm.tabLogo = tab.logo || {};

        this.form.homeName = home.title;
        this.form.homeLogo = home.logo || {};
        const color = home.backgroundColor || this.$config.backgroundColor;
        this.radio = color.includes('linear-gradient') ? '1' : '0';
        this.form.homeBgColor = color;

        this.loginForm.loginBg = login.background || {};
        this.loginForm.loginLogo = login.logo || {};
        this.loginForm.loginButtonColor = login.loginButtonColor || '';

        this.generalAgentForm.generalAgentIcon = generalAgent.logo || {};
        this.generalAgentForm.generalAgentMenuName =
          generalAgent.menuName || '';
        this.generalAgentForm.generalAgentWelcome =
          generalAgent.welcomeText || '';
      },
      deep: true,
    },
  },
  computed: {
    ...mapGetters('user', ['commonInfo']),
  },
  methods: {
    ...mapActions('user', ['getCommonInfo']),
    /** 获取图片显示地址，优先使用本地预览URL以保证清晰度 */
    getLogoPath(path, previewUrl) {
      return previewUrl || (path ? avatarSrc(path) : '');
    },
    uploadAvatar(file) {
      const formData = new FormData();
      const config = { headers: { 'Content-Type': 'multipart/form-data' } };
      formData.append('avatar', file);
      return uploadAvatar(formData, config);
    },
    /** 校验文件尺寸，bg类型限制1MB，logo类型限制100KB */
    checkFileSize(file, type) {
      const isBg = type === 'bg';
      const limit = isBg ? 1 * 1024 * 1024 : 100 * 1024;
      if (file.size > limit) {
        this.$message.warning(
          isBg
            ? this.$t('infoSetting.hint.bgSizeLimit')
            : this.$t('infoSetting.hint.logoSizeLimit'),
        );
        return false;
      }
      return true;
    },
    handleUploadLogo(data) {
      if (data.file) {
        if (!this.checkFileSize(data.file, 'logo')) return;
        const previewUrl = URL.createObjectURL(data.file);
        this.uploadAvatar(data.file).then(res => {
          this.form.homeLogo = { ...res.data, previewUrl };
        });
      }
    },
    handleUploadLoginBg(data) {
      if (data.file) {
        if (!this.checkFileSize(data.file, 'bg')) return;
        const previewUrl = URL.createObjectURL(data.file);
        this.uploadAvatar(data.file).then(res => {
          this.loginForm.loginBg = { ...res.data, previewUrl };
        });
      }
    },
    handleUploadLoginLogo(data) {
      if (data.file) {
        if (!this.checkFileSize(data.file, 'logo')) return;
        const previewUrl = URL.createObjectURL(data.file);
        this.uploadAvatar(data.file).then(res => {
          this.loginForm.loginLogo = { ...res.data, previewUrl };
        });
      }
    },
    handleUploadLabelIcon(data) {
      if (data.file) {
        if (!this.checkFileSize(data.file, 'logo')) return;
        const previewUrl = URL.createObjectURL(data.file);
        this.uploadAvatar(data.file).then(res => {
          this.tabForm.tabLogo = { ...res.data, previewUrl };
        });
      }
    },
    handleUploadGeneralAgentIcon(data) {
      if (data.file) {
        if (!this.checkFileSize(data.file, 'logo')) return;
        const previewUrl = URL.createObjectURL(data.file);
        this.uploadAvatar(data.file).then(res => {
          this.generalAgentForm.generalAgentIcon = {
            ...res.data,
            previewUrl,
          };
        });
      }
    },
    handleUploadError() {
      this.$message.error(this.$t('common.message.uploadError'));
    },
    async submitData(type, data) {
      try {
        this[`${type}Loading`] = true;
        const res = await setPlatformInfo(type, data);
        if (res.code === 0)
          this.$message.success(this.$t('common.message.success'));
        if (type !== 'login') this.getCommonInfo();
      } finally {
        this[`${type}Loading`] = false;
      }
    },
    handleSubmitTab() {
      this.$refs.tabForm.validate(valid => {
        if (valid) this.submitData('tab', this.tabForm);
      });
    },
    handleSubmit() {
      this.$refs.form.validate(valid => {
        if (valid) this.submitData('home', this.form);
      });
    },
    handelSubmitLogin() {
      this.$refs.loginForm.validate(valid => {
        if (valid) this.submitData('login', this.loginForm);
      });
    },
    handleSubmitGeneralAgent() {
      this.$refs.generalAgentForm.validate(valid => {
        if (valid) this.submitData('general-agent', this.generalAgentForm);
      });
    },
  },
};
</script>

<style lang="scss" scoped>
.docPage-card {
  margin-bottom: 20px;
  .card-title {
    font-size: 18px;
    font-weight: bold;
  }
  .card-footer {
    text-align: right;
  }
}
.avatar-uploader ::v-deep .el-upload {
  text-align: left !important;
}
.avatar-uploader-icon {
  font-size: 24px;
  color: #8c939d;
  width: 120px;
  height: 120px;
  line-height: 120px;
  text-align: center;
  border: 1px dashed rgb(217, 217, 217);
  border-radius: 5px;
}
.avatar {
  width: 120px;
  height: 120px;
  display: inline-block;
  vertical-align: middle;
  border: 1px dashed rgb(217, 217, 217);
  border-radius: 5px;
  object-fit: cover;
  background-color: #f8f8f8;
  will-change: transform;
}
.el-upload {
  width: 100%;
}
</style>
