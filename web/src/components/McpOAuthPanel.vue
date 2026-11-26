<template>
  <section v-if="providers.length" aria-label="MCP OAuth connections" class="mcp-oauth-panel">
    <h3>OAuth 授权连接 / OAuth connections</h3>
    <p>请在系统浏览器登录万悟并授权。无需复制 API Key；授权后自动保存连接。</p>
    <article v-for="p in providers" :key="p.id" class="oauth-card">
      <strong>{{ p.name }}</strong> <el-tag size="mini">OAuth · Streamable HTTP</el-tag>
      <p>{{ p.resource }}</p>
      <el-button type="primary" :loading="busy === p.id" @click="prepare(p)">一键授权 / Authorize</el-button>
      <el-button v-if="authorized[p.id]" @click="connect(p)">保存连接 / Save connection</el-button>
      <el-button v-if="authorized[p.id]" @click="disconnect(p)">解绑 / Disconnect</el-button>
      <div v-if="active === p.id && authorizationUrl" role="status">
        <p>授权已准备好。点击下方链接打开登录页；如在桌面内嵌页，请复制链接到系统浏览器。</p>
        <a :href="authorizationUrl" target="_blank" rel="noopener noreferrer" class="oauth-open">打开登录授权页 / Open authorization</a>
        <el-button size="small" @click="copyLink">复制授权链接</el-button>
        <p>正在等待授权回调（最长 10 分钟）…</p>
      </div>
      <p v-if="authorized[p.id]" role="status">✓ 已授权 / Authorized</p>
      <p v-if="messages[p.id]" role="alert">{{ messages[p.id] }}</p>
    </article>
  </section>
</template>
<script>
import { getOAuthProviders, mcpOAuthAction } from '@/api/mcp';
export default {
  name: 'McpOAuthPanel',
  props: { providerIds: { type: Array, default: () => [] } },
  data: () => ({ providers: [], authorized: {}, messages: {}, busy: '', active: '', authorizationUrl: '', timer: null, deadline: 0, disposed: false }),
  async mounted() {
    try {
      const res = await getOAuthProviders();
      this.providers = (res.data || []).filter(p => !this.providerIds.length || this.providerIds.includes(p.id));
      for (const p of this.providers) {
        const status = await mcpOAuthAction('status', { provider: p.id });
        this.$set(this.authorized, p.id, Boolean(status.data.authorized));
      }
    } catch (_) { /* OAuth is optional; existing connectors remain available. */ }
  },
  beforeDestroy() { this.disposed = true; clearTimeout(this.timer); },
  methods: {
    async prepare(p) {
      clearTimeout(this.timer); this.active = ''; this.authorizationUrl = ''; this.busy = p.id;
      this.$set(this.messages, p.id, '');
      try {
        const res = await mcpOAuthAction('prepare', { provider: p.id });
        const u = new URL(res.data.authorizationUrl);
        if (u.protocol !== 'https:') throw new Error('Invalid authorization URL');
        this.authorizationUrl = u.href; this.active = p.id; this.deadline = Date.now() + 600000;
        this.$set(this.authorized, p.id, false);
        this.poll(p);
      } catch (_) { this.$set(this.messages, p.id, '准备授权失败，请检查服务配置后重试 / Unable to prepare authorization.'); }
      finally { this.busy = ''; }
    },
    async poll(p) {
      if (this.disposed || this.active !== p.id) return;
      if (Date.now() >= this.deadline) { this.authorizationUrl = ''; this.$set(this.messages, p.id, '授权等待已超时，请重新授权。'); return; }
      try {
        const res = await mcpOAuthAction('status', { provider: p.id });
        if (res.data.authorized) {
          await mcpOAuthAction('connect', { provider: p.id });
          this.$set(this.authorized, p.id, true); this.authorizationUrl = ''; this.active = '';
          this.$set(this.messages, p.id, '连接已保存，可在智能体中选择使用 / Connection saved.');
          this.$emit('connected'); return;
        }
      } catch (_) { this.$set(this.messages, p.id, '暂未完成保存，正在重试；请保持登录。'); }
      if (!this.disposed && this.active === p.id) this.timer = setTimeout(() => this.poll(p), 2000);
    },
    async connect(p) {
      try { await mcpOAuthAction('connect', { provider: p.id }); this.$set(this.messages, p.id, '连接已保存'); this.$emit('connected'); }
      catch (_) { this.$set(this.messages, p.id, '保存失败，请检查是否有同名连接并重试。'); }
    },
    async disconnect(p) {
      try { await this.$confirm('解绑将停用此授权下的全部 MCP 连接；进行中的调用可能完成。是否继续？', 'Disconnect OAuth', { type: 'warning' }); }
      catch (_) { return; }
      clearTimeout(this.timer); this.active = ''; this.authorizationUrl = '';
      try { await mcpOAuthAction('disconnect', { provider: p.id }); this.$set(this.messages, p.id, '已解绑，后续调用需要重新授权。'); }
      catch (_) { this.$set(this.messages, p.id, '解绑结果未确认，请重试；如提示远端撤销失败，请到服务商账户检查授权。'); }
      finally { const res = await mcpOAuthAction('status', { provider: p.id }).catch(() => null); if (res) this.$set(this.authorized, p.id, Boolean(res.data.authorized)); }
    },
    async copyLink() {
      try { await navigator.clipboard.writeText(this.authorizationUrl); this.$message.success('授权链接已复制'); }
      catch (_) { this.$message.info('请右键授权链接，选择复制链接地址'); }
    },
  },
};
</script>
<style scoped>
.mcp-oauth-panel { padding: 20px; }
.oauth-card { padding: 16px; margin: 12px 0; border: 1px solid #ddd; border-radius: 8px; }
.oauth-card p { margin: 12px 0; overflow-wrap: anywhere; }
.oauth-open { display: inline-block; padding: 12px 18px; margin: 8px 12px 8px 0; background: #4252c6; color: white; border-radius: 6px; }
</style>
