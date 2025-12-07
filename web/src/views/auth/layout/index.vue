<template>
  <div class="auth">
    <div class="overview">
      <img :src="backgroundSrc" alt="">
    </div>
    <div class="blur-overlay"></div>
    <div class="auth-modal">
      <div class="header__left">
        <img
          v-if="commonInfo.login.logo && commonInfo.login.logo.path"
          style="height: 60px; margin: 0 15px 0 22px"
          :src="basePath + '/user/api' + commonInfo.login.logo.path"
          alt=""/>
      </div>
      <div class="auth-content">
        <slot :commonInfo="commonInfo"/>
      </div>
    </div>
  </div>
</template>

<script>
import {mapState, mapActions} from 'vuex'
import ChangeLang from "@/components/changeLang.vue"
import {replaceTitle, replaceIcon, avatarSrc} from "@/utils/util";
import { getCommonInfo } from '@/api/user'

export default {
  components: {ChangeLang},
  data() {
    return {
      backgroundSrc: require('@/assets/imgs/auth_bg.png'),
      basePath: this.$basePath
    }
  },
  computed: {
    ...mapState('login', ['commonInfo']),
    ...mapState('user', ['lang'])
  },
  watch: {
    'lang': {
      handler(val) {
        if (val) {
          /*this.getImgCode()
          this.getLogoInfo()*/
        }
      },
      immediate: true
    }
  },
  created() {
    this.getCommonInfo().then(() => {
      const { tab = {}, login = {} } = this.commonInfo || {}
      const { logo = {}, title = '' } = tab || {}
      const { background = {} } = login || {}

      background.path && this.setAuthBg(background.path)
      title && replaceTitle(title)
      logo.path && replaceIcon(logo.path)
      this.$emit('getCommonInfo', this.commonInfo)
    })
  },
  methods: {
    ...mapActions('login', ['getCommonInfo']),
    setDefaultImage() {
      this.backgroundSrc = require('@/assets/imgs/auth_bg.png')
    },
    setAuthBg(backgroundPath) {
      if (!backgroundPath) {
        this.setDefaultImage()
        return
      }
      this.backgroundSrc = avatarSrc(backgroundPath)
    },
  }
}
</script>

<style lang="scss" scoped>
@import "@/style/auth.scss";
.overview {
  position: relative;
  height: 100%;
  overflow: hidden;
  //background-color: #000;
  z-index: 10;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    background-size: 100% 100%;
  }

  .overview-desc {
    width: 800px;
    position: absolute;
    bottom: 56px;
    left: 56px;
    color: #fff;
    text-align: center;
    opacity: .8;
    letter-spacing: 1px;

    .desc {
      font-size: 30px;
      text-align: left;

      p:nth-child(1) {
        font-size: 22px;
      }

      p:nth-child(2) {
        font-size: 30px;
        margin: 10px 0;
      }

      p:nth-child(3) {
        font-size: 18px;
      }
    }

    .tabs {
      display: flex;
      font-size: 27px;
      margin-top: 30px;
      color: #fff;

      .tab {
        width: 1.63rem;
        min-width: 163px;
        margin-right: 20px;
        border: 1px solid #fff;
        cursor: pointer;

        p:nth-child(1) {
          font-size: 18px;
          padding: 4px 0 3px 0;
        }

        p:nth-child(2) {
          font-size: 12px;
          font-weight: 400;
          padding: 0 0 6px 0;
        }
      }
    }
  }
}

.auth {
  height: 100%;
}

.blur-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.3);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  z-index: 100;
}

.auth-modal {
  position: fixed;
  top: 0;
  bottom: 0;
  left: 0;
  right: 0;
  width: 100%;
  height: 100%;
  z-index: 1000;
  display: flex;
  flex-direction: column;

  .header__left {
    position: absolute;
    top: 16px;
    left: 10px;
    color: #fff;
    font-weight: bold;
    display: flex;
    align-items: center;
    height: 60px;
    z-index: 1001;
  }

  .auth-content {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
  }
}
</style>
