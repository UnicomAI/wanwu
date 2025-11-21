<template>
  <overview>
    <template #default="{ commonInfo }">
      <div class="auth-box">
        <p class="auth-header">
          <span style="font-weight: bold">{{ $t('login.title') }}</span>
        </p>
        <div class="auth-form">
          <el-form ref="form" :model="form" label-position="top">
            <el-form-item class="auth-form-item">
              <img class="auth-icon" src="@/assets/imgs/user.png" alt=""/>
              <el-input
                v-model.trim="form.username"
                :placeholder="$t('common.input.placeholder') + $t('login.form.username')"/>
            </el-form-item>
            <el-form-item class="auth-form-item">
              <img class="auth-icon" src="@/assets/imgs/pwd.png" alt=""/>
              <el-input
                :type="isShowPwd ? '' : 'password'"
                class="auth-pwd-input"
                v-model.trim="form.password"
                :placeholder="$t('common.input.placeholder') + $t('login.form.password')"/>
              <img
                v-if="!isShowPwd" class="pwd-icon" src="@/assets/imgs/showPwd.png" alt=""
                @click="isShowPwd = true"/>
              <img
                v-else class="pwd-icon" src="@/assets/imgs/hidePwd.png" alt=""
                @click="isShowPwd = false"/>
            </el-form-item>
            <el-form-item class="auth-form-item">
              <img class="auth-icon" src="@/assets/imgs/code.png" alt=""/>
              <el-input
                style="width: calc(100% - 90px)"
                v-model.trim="form.code"
                @keyup.enter.native="addByEnterKey"
                :placeholder="$t('common.input.placeholder') + $t('login.form.code')"/>
              <span style="display: inline-block; height: 32px; width: 80px; margin-left: 10px; vertical-align: middle">
                <img style="width: 100%; height: 100%" v-if="codeData.b64" :src="codeData.b64" @click="getImgCode"/>
              </span>
            </el-form-item>
          </el-form>
          <div class="nav-bt">
            <span v-if="commonInfo.register.email.status">
                {{ $t('login.askAccount') }}
              <span :style="{ color: 'var(--color)', cursor: 'pointer' }" @click="$router.push({path: `/register`})">
                {{ $t('login.register') }}
              </span>
            </span>
            <span
              v-if="commonInfo.resetPassword.email.status"
              :style="{ color: 'var(--color)', cursor: 'pointer', float: 'right' }"
              @click="$router.push({path: `/reset`})">
              {{ $t('login.forgetPassword') }}
            </span>
          </div>
          <div class="auth-bt">
            <p
              :class="['primary-bt', {'disabled': isDisabled()}]"
              :style="`background: ${commonInfo.login.loginButtonColor} !important`"
              @click="doLogin">
              {{ $t('login.button') }}
            </p>
          </div>
          <div class="bottom-text">{{ commonInfo.login.platformDesc }}</div>
        </div>
        <dialog2FA ref="dialog2FA"></dialog2FA>
      </div>
    </template>
  </overview>
</template>

<script>
import dialog2FA from './2FADialog'
import overview from '@/views/auth/layout'
import {mapActions, mapState} from 'vuex'
import {getImgVerCode} from "@/api/user"
import {urlEncrypt} from "@/utils/crypto";
import {redirectUrl} from "@/utils/util";

export default {
  components: {overview, dialog2FA},
  data() {
    return {
      form: {
        username: '',
        password: '',
        code: '',
      },
      isShowPwd: false,
      codeData: {
        key: '',
        b64: ''
      },
      params: {
        client_id: '',
        redirect_uri: '',
        scope: '',
        response_type: '',
        state: '',
        client_name: ''
      }
    }
  },
  created() {
    // IfalreadyLogin，重定向toHasPermission of Page
    // if (this.$store.state.user.token && localStorage.getItem("access_cert") && !this.$store.state.user.is2FA) redirectUrl()

    this.getImgCode()
  },
  watch: {
    $route: {
      handler() {
        this.params = this.$route.query
        if (
          this.$store.state.user.token &&
          localStorage.getItem("access_cert") &&
          !this.$store.state.user.is2FA &&
          this.params.client_id
        ) this.$router.push({
          path: "/oauth",
          query: this.params
        });

      },
      // 深度ObserveListen
      deep: true
    }
  },
  mounted() {
    this.params = this.$route.query
    if (
      this.$store.state.user.token &&
      localStorage.getItem("access_cert") &&
      !this.$store.state.user.is2FA &&
      this.params.client_id
    ) this.$router.push({
      path: "/oauth",
      query: this.params
    });
  },
  computed: {
    ...mapState('login', ['commonInfo'])
  },
  methods: {
    ...mapActions('user', ['LoginIn','LoginIn2FA1']),
    isDisabled() {
      const {username, password, code} = this.form
      return !(username && password && code)
    },
    addByEnterKey(e) {
      if (e.keyCode === 13) {
        this.doLogin()
      }
    },
    // GetImageVerification Code
    async getImgCode() {
      const res = await getImgVerCode()
      this.codeData = res.data || {}
    },
    async doLogin() {
      if (this.isDisabled()) return

      const data = {
        username: this.form.username,
        password: urlEncrypt(this.form.password),
        key: this.codeData.key,
        code: this.form.code
      }

      try {
        if (this.commonInfo.loginEmail.email.status) {
          const {isEmailCheck, isUpdatePassword} = await this.LoginIn2FA1(data)
          this.$refs.dialog2FA.showDialog(isEmailCheck, isUpdatePassword, this.params)
        } else await this.LoginIn({ loginInfo: data, params: this.params })
      } catch (e) {
        await this.getImgCode()
      }
    },
  }
}
</script>

<style lang="scss" scoped>

</style>
