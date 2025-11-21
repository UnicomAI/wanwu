<template>
  <div class="templateSquare" :style="`background: ${isPublic ? bgColor : 'none'}`">
    <div class="page-wrapper">
      <div class="page-title">
        <img class="page-title-img" :src="require('@/assets/imgs/template_square.svg')" alt="" />
        <span class="page-title-name">{{$t('menu.templateSquare')}}</span>
      </div>
      <!-- tabs -->
      <div class="templateSquare-tabs">
        <div :class="['templateSquare-tab',{ 'active': type === workflow }]" @click="tabClick(workflow)">
          {{$t('tempSquare.workflow')}}
        </div>
        <div :class="['templateSquare-tab',{ 'active': type === prompt }]" @click="tabClick(prompt)">
          {{$t('tempSquare.prompt')}}
        </div>
      </div>

      <TempSquare :isPublic="isPublic" :type="workflow" ref="tempSquare" v-if="type === workflow" />
      <PromptTempSquare :isPublic="isPublic" :type="prompt" ref="promptTempSquare" v-if="type === prompt" />
    </div>
  </div>
</template>
<script>
import TempSquare from './tempSquare.vue'
import PromptTempSquare from "./promptTempSquare.vue"
import { WORKFLOW, PROMPT } from "./constants"

export default {
  components: { TempSquare, PromptTempSquare },
  data() {
    return {
      isPublic: true,
      bgColor: 'linear-gradient(1deg, rgb(247, 252, 255) 50%, rgb(233, 246, 254) 98%)',
      workflow: WORKFLOW,
      prompt: PROMPT,
      type: ''
    };
  },
  created() {
    this.isPublic = this.$route.path.includes('/public/')
    this.type = this.$route.query.type || PROMPT //WORKFLOW
  },
  methods: {
    tabClick(type) {
      this.type = type
    },
  },
};
</script>
<style lang="scss">
.templateSquare {
  width: 100%;
  height: 100%;
  
  .templateSquare-tabs {
    margin: 20px;

    .templateSquare-tab {
      display: inline-block;
      vertical-align: middle;
      width: 160px;
      height: 40px;
      border-bottom: 1px solid #333;
      line-height: 40px;
      text-align: center;
      cursor: pointer;
    }

    .active {
      background: #333;
      color: #fff;
      font-weight: bold;
    }
  }
}
</style>