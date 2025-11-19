<template>
  <overview @getCommonInfo="handleCommonInfo">
    <template #default="{ commonInfo }">
      <div class="auth-box">
        <p class="auth-header">
          <span style="font-weight: bold">{{ $t('reset.title') }}</span>
        </p>
        <div class="auth-form">
          <el-form ref="form" :model="form" :rules="rules" label-position="top">
            <el-form-item class="auth-form-item" prop="email">
              <img class="auth-icon" src="@/assets/imgs/user.png" alt=""/>
              <el-input
                v-model.trim="form.email"
                :placeholder="$t('common.input.placeholder') + $t('reset.form.email')" clearable
              />
            </el-form-item>
            <el-form-item class="auth-form-item" prop="code">
              <img class="auth-icon" src="@/assets/imgs/code.png" alt=""/>
              <el-input
                style="width: calc(100% - 90px)"
                v-model.trim="form.code"
                @keyup.enter.native="addByEnterKey"
                :placeholder="$t('common.input.placeholder') + $t('reset.form.code')" clearable
              />
              <el-button
                style="height: 32px; width: 80px; margin-left: 10px; vertical-align: middle; padding-left: 8px; padding-top: 8px"
                @click="requestEmailCode({email: form.email})"
                :disabled="isCooldown"
              >
                {{ isCooldown ? `${cooldownTime}s` : $t('reset.action') + $t('reset.form.code') }}
              </el-button>
            </el-form-item>
            <el-form-item
              class="auth-form-item"
              prop="password1">
              <img class="auth-icon" src="@/assets/imgs/pwd.png" alt=""/>
              <el-input
                :type="isShowPwd1 ? '' : 'password'"
                class="auth-pwd-input"
                v-model.trim="form.password1"
                :placeholder="$t('reset.pwd1Placeholder')"
              />
              <img
                v-if="!isShowPwd1" class="pwd-icon" src="@/assets/imgs/showPwd.png" alt=""
                @click="isShowPwd1 = true"/>
              <img
                v-else class="pwd-icon" src="@/assets/imgs/hidePwd.png" alt=""
                @click="isShowPwd1 = false"/>
            </el-form-item>
            <el-form-item
              class="auth-form-item"
              prop="password2">
              <img class="auth-icon" src="@/assets/imgs/pwd.png" alt=""/>
              <el-input
                :type="isShowPwd2 ? '' : 'password'"
                class="auth-pwd-input"
                v-model.trim="form.password2"
                :placeholder="$t('reset.action2') + $t('reset.form.password')"/>
              <img
                v-if="!isShowPwd2" class="pwd-icon" src="@/assets/imgs/showPwd.png" alt=""
                @click="isShowPwd2 = true"/>
              <img
                v-else class="pwd-icon" src="@/assets/imgs/hidePwd.png" alt=""
                @click="isShowPwd2 = false"/>
            </el-form-item>
          </el-form>
          <div class="nav-bt">
            {{ $t('reset.askAccount') }}
            <span :style="{ color: 'var(--color)', cursor: 'pointer' }" @click="$router.push({path: `/login`})">
              {{ $t('reset.login') }}
            </span>
          </div>
          <div class="auth-bt">
            <p class="primary-bt" :style="`background: ${commonInfo.login.loginButtonColor} !important`"
               @click="doReset">
              {{ $t('reset.button') }}
            </p>
          </div>
          <div class="bottom-text">{{ commonInfo.login.platformDesc }}</div>
        </div>
      </div>
    </template>
  </overview>
</template>

<script>
import overview from '@/views/auth/layout'
import {resetCode, reset} from "@/api/user"
import {urlEncrypt} from "@/utils/crypto";

export default {
  components: {overview},
  data() {
    let checkPassword2 = (rule, value, callback) => {
      if (this.form.password1 !== this.form.password2) callback(new Error(this.$t('resetPwd.differError')))
      callback()
    }
    let checkPassword1 = (rule, value, callback) => {
      let reg = /^(?=.*[a-zA-Z])(?=.*\d)(?=.*[~!@#$%^&*()_+`\-={}:";'<>?,./]).{8,20}$/
      if (!reg.test(value)) {
        callback(new Error(this.$t('resetPwd.pwdError')))
      } else {
        return callback()
      }
    }
    return {
      form: {
        email: '',
        code: '',
        password1: '',
        password2: ''
      },
      rules: {
        email: [
          {required: true, message: this.$t('common.input.placeholder'), trigger: 'blur'},
          {
            pattern: /^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(.[a-zA-Z0-9_-]+)+$/,
            message: this.$t('common.hint.emailError'),
            trigger: "blur"
          }
        ],
        code: [
          {required: true, message: this.$t('common.input.placeholder'), trigger: 'blur'}
        ],
        password1: [
          {required: true, message: this.$t('common.input.placeholder'), trigger: 'blur'},
          {validator: checkPassword1, trigger: "blur"}
        ],
        password2: [
          {required: true, message: this.$t('common.input.placeholder'), trigger: 'blur'},
          {validator: checkPassword1, trigger: "blur"},
          {validator: checkPassword2, trigger: "blur"},
        ],
      },
      isCooldown: false,
      cooldownTime: 60,
      cooldownTimer: '',
      codeSentMessage: '',
      isShowPwd1: false,
      isShowPwd2: false,
      codeData: {
        key: '',
        b64: ''
      },
      basePath: this.$basePath
    }
  },
  methods: {
    handleCommonInfo(commonInfo) {
      // If功能未开启，重定向到Login页
      if (!commonInfo.resetPassword.email.status) {
        this.$router.push({path: `/login`})
      }
    },
    addByEnterKey(e) {
      if (e.keyCode === 13) {
        this.doReset()
      }
    },
    doReset() {
      this.$refs.form.validate((valid) => {
        if (!valid) return
        const data = {
          email: this.form.email,
          password: urlEncrypt(this.form.password1),
          code: this.form.code
        }
        reset(data).then(res => {
          if (res.code === 0) {
            this.$router.push({path: `/login`})
          }
        })
      })
    },
    requestEmailCode(data) {
      this.$refs.form.validateField(['email'], err => {
        if (err) return
        this.codeSentMessage = this.$t('common.hint.codeSent')
        this.isCooldown = true
        this.cooldownTimer = setInterval(() => {
          if (this.cooldownTime > 1) {
            this.cooldownTime--
          } else {
            this.isCooldown = false
            this.cooldownTime = 60
            clearInterval(this.cooldownTimer)
          }
        }, 1000)
        resetCode(data)
      })
    }
  },
  beforeDestroy() {
    clearInterval(this.cooldownTimer)
    this.codeSentMessage = ''
  }
}
</script>

<style lang="scss" scoped>
.auth-box {
  height: 550px !important;
}
</style>
