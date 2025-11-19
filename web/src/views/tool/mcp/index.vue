<template>
  <div class="page-wrapper mcp-management">
    <div class="common_bg">
      <!-- tabs -->
      <div class="mcp-tabs">
        <div :class="['mcp-tab',{ 'active': tabActive === 0 }]" @click="tabClick(0)">{{ $t('common.button.import') }}MCP</div>
        <div :class="['mcp-tab',{ 'active': tabActive === 1 }]" @click="tabClick(1)">{{ $t('common.button.add') }}MCP</div>
      </div>

      <customize ref="customize" v-if="tabActive === 0"/>
      <server ref="server" v-if="tabActive === 1"/>
    </div>
  </div>
</template>
<script>
import customize from './integrate'
import server from './server'
export default {
  data() {
    return {
      tabActive:0
    };
  },
  watch: {
    $route: {
      handler() {
        if (this.$route.query.mcp === "integrate") this.tabActive = 0
        else if (this.$route.query.mcp === "server") this.tabActive = 1
        else this.tabActive = 0
      },
      // 深度ObserveListen
      deep: true
    }
  },
  mounted() {
    if (this.$route.query.mcp === "integrate") this.tabActive = 0
    else if (this.$route.query.mcp === "server") this.tabActive = 1
    else this.tabActive = 0
  },
  methods: {
    tabClick(status){
      this.tabActive = status
    },
  },
  components: {
      customize,
      server
  },
};
</script>
<style lang="scss" scoped>
.mcp-tabs{
  .mcp-tab{
    width: 100px !important;
    height: 30px !important;
    line-height: 30px !important;
    font-size: 10px !important;
    color: #666666 !important;
    border-bottom: 1px solid #CCCCCC !important;
  }
  .active{
    background: #CCCCCC !important;
    color: #333 !important;
  }
}
</style>