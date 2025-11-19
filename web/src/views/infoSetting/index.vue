<template>
  <div>
    <!--tabConfiguration-->
    <el-card shadow="never" class="docPage-card">
      <div slot="header">
        <span class="card-title">{{$t('infoSetting.tabSet')}}</span>
      </div>
      <el-form label-width="120px" :model="tabForm" :rules="tabRules" ref="tabForm">
        <el-form-item :label="$t('infoSetting.form.labelTitle')" prop="tabTitle">
          <el-input v-model="tabForm.tabTitle" style="width: 300px" />
          <div style="font-size: 11px; color: #aaa; margin-top: -9px">
            {{$t('infoSetting.hint.labelTitle')}}
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
            <img v-if="tabForm.tabLogo.path" :src="getLogoPath(tabForm.tabLogo.path)" class="avatar">
            <i v-if="!tabForm.tabLogo.path" class="el-icon-plus avatar-uploader-icon"></i>
            <span style="margin-left: 12px; color: #aaa !important;">
              {{$t('infoSetting.hint.imgUpload')}}
            </span>
            <div style="text-align: left; font-size: 11px; color: #aaa; margin-top: -9px">
              {{$t('infoSetting.hint.labelIcon')}}
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
          {{$t('infoSetting.form.save')}}
        </el-button>
      </div>
    </el-card>
    <!--登录页Configuration-->
    <el-card shadow="never" class="docPage-card">
      <div slot="header">
        <span class="card-title">{{$t('infoSetting.loginBgSet')}}</span>
      </div>
      <el-form label-width="120px" :model="loginForm" :rules="loginRules" ref="loginForm">
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
            <img v-if="loginForm.loginBg.path" :src="getLogoPath(loginForm.loginBg.path)" class="avatar">
            <i v-if="!loginForm.loginBg.path" class="el-icon-plus avatar-uploader-icon"></i>
            <span style="margin-left: 12px; color: #aaa !important;">
              {{$t('infoSetting.hint.imgUpload')}}
            </span>
            <div style="text-align: left; font-size: 11px; color: #aaa; margin-top: -9px">
              {{$t('infoSetting.hint.loginBg')}}
            </div>
          </el-upload>
        </el-form-item>
        <el-form-item :label="$t('infoSetting.form.loginLogo')" prop="loginLogo">
          <el-upload
            class="avatar-uploader"
            action=""
            name="files"
            :show-file-list="false"
            :http-request="handleUploadLoginLogo"
            :on-error="handleUploadError"
            accept=".png,.jpg,.jpeg"
          >
            <img v-if="loginForm.loginLogo.path" :src="getLogoPath(loginForm.loginLogo.path)" class="avatar">
            <i v-if="!loginForm.loginLogo.path" class="el-icon-plus avatar-uploader-icon"></i>
            <span style="margin-left: 12px; color: #aaa !important;">
              {{$t('infoSetting.hint.imgUpload')}}
            </span>
            <div style="text-align: left; font-size: 11px; color: #aaa; margin-top: -9px">
              {{$t('infoSetting.hint.loginLogo')}}
            </div>
          </el-upload>
        </el-form-item>
        <el-form-item :label="$t('infoSetting.form.logoWelcome')" prop="loginWelcomeText">
          <el-input v-model="loginForm.loginWelcomeText" style="width: 300px" />
          <div style="font-size: 11px; color: #aaa; margin-top: -9px">
            {{$t('infoSetting.hint.logoWelcome')}}
          </div>
        </el-form-item>
        <el-form-item :label="$t('infoSetting.form.loginButtonColor')" prop="loginButtonColor">
          <el-color-picker v-model="loginForm.loginButtonColor" show-alpha></el-color-picker>
          <div style="text-align: left; font-size: 11px; color: #aaa; margin-top: -9px">
            {{$t('infoSetting.hint.loginButtonColor')}}
          </div>
        </el-form-item>
      </el-form>
      <div class="card-footer">
        <el-button
          :loading="loginLoading"
          type="primary"
          @click="handelSubmitLogin()"
        >
          {{$t('infoSetting.form.save')}}
        </el-button>
      </div>
    </el-card>
    <!--平台Configuration-->
    <el-card shadow="never" class="docPage-card">
      <div slot="header">
        <span class="card-title">{{$t('infoSetting.platformSet')}}</span>
      </div>
      <el-form label-width="120px" :model="form" :rules="rules" ref="form">
        <el-form-item :label="$t('infoSetting.form.platformTitle')" prop="homeName">
          <el-input v-model="form.homeName" style="width: 300px" />
          <div style="font-size: 11px; color: #aaa; margin-top: -9px">
            {{$t('infoSetting.hint.platformTitle')}}
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
            <img v-if="form.homeLogo.path" :src="getLogoPath(form.homeLogo.path)" class="avatar">
            <i v-if="!form.homeLogo.path" class="el-icon-plus avatar-uploader-icon"></i>
            <span style="margin-left: 12px; color: #aaa !important;">
              {{$t('infoSetting.hint.imgUpload')}}
            </span>
            <div style="text-align: left; font-size: 11px; color: #aaa; margin-top: -9px">
              {{$t('infoSetting.hint.logo')}}
            </div>
          </el-upload>
        </el-form-item>
        <el-form-item :label="$t('infoSetting.form.bgColor')" prop="homeBgColor">
          <el-radio-group v-model="radio">
            <el-radio label="0">{{$t('infoSetting.hint.oneColor')}}</el-radio>
            <el-radio label="1">{{$t('infoSetting.hint.linearColor')}}</el-radio>
          </el-radio-group>
          <div v-if="radio === '0'">
            <el-color-picker v-model="form.homeBgColor" show-alpha></el-color-picker>
          </div>
          <div v-if="radio === '1'">
            <el-input v-model="form.homeBgColor" style="width: 300px" />
            <span :style="`display: inline-block; vertical-align: middle; width: 32px; height: 32px; border-radius: 2px; margin-left: 14px; background: ${form.homeBgColor}`"></span>
          </div>
          <div style="text-align: left; font-size: 11px; color: #aaa; margin-top: -9px">
            {{$t('infoSetting.hint.bgColor')}}
          </div>
        </el-form-item>
      </el-form>
      <div class="card-footer">
        <el-button
          :loading="homeLoading"
          type="primary"
          @click="handleSubmit()"
        >
          {{$t('infoSetting.form.save')}}
        </el-button>
      </div>
    </el-card>
  </div>
</template>
<script>
import { setPlatformInfo } from "@/api/setInfo"
import { uploadAvatar } from "@/api/user"
import { mapActions, mapGetters } from "vuex"
import {avatarSrc} from "@/utils/util";

export default {
  data() {
    const checkImage = (val, rule, value, callback) => {
      // 判断Edit器Value，Edit器Default没Enter when Yes <p><br></p>
      if (!val) {
        callback(new Error(this.$t('infoSetting.form.uploadHint')))
      } else {
        return callback()
      }
    }
    return {
      tabLoading: false,
      homeLoading: false,
      loginLoading: false,
      tabForm: {
        tabTitle: '',
        tabLogo: {},
      },
      form: {
        homeName: '',
        homeLogo: {},
        homeBgColor: ''
      },
      loginForm: {
        loginBg: {},
        loginLogo: {},
        loginButtonColor: '',
        loginWelcomeText: ''
      },
      radio: '0',
      tabRules: {
        tabTitle: [{required: true, message: this.$t('common.input.placeholder'), trigger: 'blur'}],
        tabLogo: [{required: true, validator: (...params) => checkImage(this.tabForm.tabLogo.path, ...params), trigger: 'change'}],
      },
      rules: {
        homeName: [{required: true, message: this.$t('common.input.placeholder'), trigger: 'blur'}],
        homeLogo: [{required: true, validator: (...params) => checkImage(this.form.homeLogo.path, ...params), trigger: 'change'}],
        homeBgColor: [{required: true, message: this.$t('common.select.placeholder'), trigger: 'change'}],
      },
      loginRules: {
        loginBg: [{required: true, validator: (...params) => checkImage(this.loginForm.loginBg.path, ...params), trigger: 'change'}],
        loginLogo: [{required: true, validator: (...params) => checkImage(this.loginForm.loginLogo.path, ...params), trigger: 'change'}],
        loginButtonColor: [{required: true, message: this.$t('common.select.placeholder'), trigger: 'change'}],
        loginWelcomeText: [{required: true, message: this.$t('common.input.placeholder'), trigger: 'blur'}],
      }
    }
  },
  created() {
    this.getCommonInfo()
  },
  watch: {
    commonInfo:{
      handler(val) {
        const { home = {}, tab = {}, login = {} } = val ? val.data || {} : {}

        this.tabForm.tabTitle = tab.title
        this.tabForm.tabLogo = tab.logo || {}

        this.form.homeName = home.title
        this.form.homeLogo = home.logo || {}
        const color = home.backgroundColor || 'linear-gradient(1deg, #FFFFFF 42%, #FFFFFF 42%, #EBEDFE 98%, #EEF0FF 98%)'
        this.radio = color.includes('linear-gradient') ? '1' : '0'
        this.form.homeBgColor = color

        this.loginForm.loginBg = login.background || {}
        this.loginForm.loginLogo = login.logo || {}
        this.loginForm.loginWelcomeText = login.welcomeText || ''
        this.loginForm.loginButtonColor = login.loginButtonColor || ''
      },
      deep: true
    }
  },
  computed: {
    ...mapGetters('user', ['commonInfo']),
  },
  methods: {
    ...mapActions('user', ['getCommonInfo']),
    getLogoPath(path) {
      return path ? avatarSrc(path) : ''
    },
    uploadAvatar(file) {
      const formData = new FormData()
      const config = {headers: { "Content-Type": "multipart/form-data" }}
      formData.append('avatar', file)
      return uploadAvatar(formData, config)
    },
    handleUploadLogo(data) {
      if (data.file) {
        this.uploadAvatar(data.file).then(res => {
          this.form.homeLogo = res.data || {}
        })
      }
    },
    handleUploadLoginBg(data) {
      if (data.file) {
        this.uploadAvatar(data.file).then(res => {
          this.loginForm.loginBg = res.data || {}
        })
      }
    },
    handleUploadLoginLogo(data) {
      if (data.file) {
        this.uploadAvatar(data.file).then(res => {
          this.loginForm.loginLogo = res.data || {}
        })
      }
    },
    handleUploadLabelIcon(data) {
      if (data.file) {
        this.uploadAvatar(data.file).then(res => {
          this.tabForm.tabLogo = res.data || {}
        })
      }
    },
    handleUploadError() {
      this.$message.error(this.$t('common.message.uploadError'))
    },
    async submitData(type, data) {
      try {
        this[`${type}Loading`] = true
        const res = await setPlatformInfo(type, data)
        if (res.code === 0) this.$message.success(this.$t('common.message.success'))
        if (type !== 'login') this.getCommonInfo()
      } finally {
        this[`${type}Loading`] = false
      }
    },
    handleSubmitTab() {
      this.$refs.tabForm.validate((valid) => {
        if (valid) this.submitData('tab', this.tabForm)
      })
    },
    handleSubmit() {
      this.$refs.form.validate((valid) => {
        if (valid) this.submitData('home', this.form)
      })
    },
    handelSubmitLogin() {
      this.$refs.loginForm.validate((valid) => {
        if (valid) this.submitData('login', this.loginForm)
      })
    }
  }
}
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
  .avatar-uploader /deep/ .el-upload {
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
  }
  .el-upload {
    width: 100%;
  }
</style>
