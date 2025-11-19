<template>
  <overview @getCommonInfo="handleCommonInfo">
    <template #default="{ commonInfo }">
      <div class="auth-box">
        <p class="auth-header">
          <span style="font-weight: bold">{{ $t('register.title') }}</span>
        </p>
        <div class="auth-form">
          <el-form ref="form" :model="form" :rules="rules" label-position="top">
            <el-form-item class="auth-form-item" prop="username">
              <img class="auth-icon" src="@/assets/imgs/user.png" alt=""/>
              <el-input
                v-model.trim="form.username"
                :placeholder="$t('common.input.placeholder') + $t('register.form.username')" clearable
              />
            </el-form-item>
            <el-form-item class="auth-form-item" prop="email">
              <img class="auth-icon" src="@/assets/imgs/user.png" alt=""/>
              <el-input
                v-model.trim="form.email"
                :placeholder="$t('common.input.placeholder') + $t('register.form.email')" clearable/>
            </el-form-item>
            <el-form-item class="auth-form-item" prop="code">
              <img class="auth-icon" src="@/assets/imgs/code.png" alt=""/>
              <el-input
                style="width: calc(100% - 90px)"
                v-model.trim="form.code"
                @keyup.enter.native="addByEnterKey"
                :placeholder="$t('common.input.placeholder') + $t('register.form.code')" clearable/>
              <el-button
                style="height: 32px; width: 80px; margin-left: 10px; vertical-align: middle; padding-left: 8px; padding-top: 8px"
                @click="requestEmailCode"
                :disabled="isCooldown"
              >
                {{ isCooldown ? `${cooldownTime}s` : $t('register.action') + $t('register.form.code') }}
              </el-button>
            </el-form-item>
          </el-form>
          <p class="message" v-if="codeSentMessage">{{ codeSentMessage }}</p>
          <div class="nav-bt">
            {{ $t('register.askAccount') }}
            <span :style="{ color: 'var(--color)', cursor: 'pointer' }" @click="$router.push({path: `/login`})">
              {{ $t('register.login') }}
            </span>
          </div>
          <div class="auth-bt">
            <p
              class="primary-bt" :style="`background: ${commonInfo.login.loginButtonColor} !important`"
              @click="doRegister">
              {{ $t('register.button') }}
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
import {registerCode, register} from "@/api/user"

export default {
  components: {overview},
  data() {
    return {
      form: {
        username: '',
        email: '',
        code: '',
      },
      rules: {
        username: [
          {required: true, message: this.$t('common.input.placeholder'), trigger: 'blur'},
          {min: 2, max: 20, message: this.$t('common.hint.userNameLimit'), trigger: 'blur'},
          {pattern: /^(?!_)[a-zA-Z0-9_.\u4e00-\u9fa5]+$/, message: this.$t('common.hint.userName'), trigger: "blur"} // 结尾：(?!.*?_$)
        ],
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
        ]
      },
      isCooldown: false,
      cooldownTime: 60,
      cooldownTimer: '',
      codeSentMessage: '',
      basePath: this.$basePath,
    }
  },
  methods: {
    handleCommonInfo(commonInfo) {
      // If功能未开启，重定向到Login页
      if (!commonInfo.register.email.status) {
        this.$router.push({path: `/login`})
      }
    },
    addByEnterKey(e) {
      if (e.keyCode === 13) {
        this.doRegister()
      }
    },
    doRegister() {
      this.$refs.form.validate((valid) => {
        if (!valid) return
        register(this.form).then(res => {
          if (res.code === 0) {
            this.$router.push({path: `/login`})
          }
        })
      })
    },
    requestEmailCode() {
      let count = 0
      this.$refs.form.validateField(['email', 'username'], (err) => {
        if (!err) count++
        if (count === 2) {
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
          const data = {
            email: this.form.email,
            username: this.form.username
          }
          registerCode(data).then(res => {
            if (res.code === 0) {
              this.codeSentMessage = this.$t('common.hint.codeSent')
            }
          })
        }
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
.message {
  color: red;
  width: 100%;
  text-align: left;
  margin-bottom: 10px;
}
</style>
