<template>
  <section class="qcc-oauth-card" aria-label="企查查 MCP OAuth">
    <h3>企查查 MCP <el-tag size="small">OAuth</el-tag></h3>
    <p>登录企查查并授权，一次连接企业、风险、知识产权、经营及董监高 5 类数据服务。查询范围以账号权限为准。</p>
    <p>无需查找或粘贴 API Key。请使用系统浏览器完成授权。</p>
    <McpOAuthPanel v-if="configured" :provider-ids="['qcc']" @connected="$emit('connected')" />
    <el-alert v-else :closable="false" type="info" title="请管理员启用企查查 OAuth 配置，设置本部署的 HTTPS 回调与加密凭据存储。启用后此卡片显示授权入口。" />
  </section>
</template>
<script>
import McpOAuthPanel from '@/components/McpOAuthPanel.vue';
import { getOAuthProviders } from '@/api/mcp';
export default {
  name: 'QccMcpCard',
  components: { McpOAuthPanel },
  data: () => ({ configured: false }),
  async mounted() {
    try { const res = await getOAuthProviders(); this.configured = (res.data || []).some(p => p.id === 'qcc'); }
    catch (_) { this.configured = false; }
  },
};
</script>
<style scoped>
.qcc-oauth-card { margin: 16px 24px; padding: 20px; border: 1px solid #d9e2f2; border-radius: 10px; }
.qcc-oauth-card p { margin: 12px 0; }
</style>
