<template>
  <main style="max-width:640px;margin:60px auto;padding:24px" aria-live="polite">
    <h1>MCP OAuth</h1>
    <p>{{ message }}</p>
    <template v-if="!done">
      <p>请使用发起授权时的同一个万悟账号和组织。</p>
      <a :href="loginUrl" target="_blank" rel="noopener noreferrer">在此浏览器登录万悟 / Sign in to Wanwu</a>
      <p><el-button type="primary" :disabled="!code || !state" :loading="busy" @click="finish">完成授权 / Complete authorization</el-button></p>
    </template>
  </main>
</template>
<script>
import { mcpOAuthAction } from '@/api/mcp';
export default {
  data: () => ({ code: '', state: '', busy: false, done: false, message: '正在完成授权，请勿关闭此页。', loginUrl: '' }),
  mounted() {
    const q = new URLSearchParams(window.location.search);
    if (q.getAll('code').length === 1 && q.getAll('state').length === 1 && !q.has('error')) { this.code = q.get('code'); this.state = q.get('state'); }
    // Strip credentials from browser history before any subsequent navigation.
    window.history.replaceState(null, '', window.location.pathname);
    this.loginUrl = this.$router.resolve('/login').href;
    if (!this.code || !this.state) { this.message = '授权已取消或回调无效，请返回连接器重新授权。'; return; }
    if (localStorage.getItem('access_cert')) this.finish();
    else this.message = '请先在此浏览器登录万悟，再点击「完成授权」。';
  },
  methods: {
    async finish() {
      this.busy = true;
      try {
        await mcpOAuthAction('complete', { state: this.state, code: this.code });
        this.code = ''; this.state = ''; this.done = true;
        this.message = '授权完成，请返回万悟连接器页面。Authorization completed. Return to Wanwu.';
      } catch (_) { this.message = '未能完成授权。请确认账号和组织与发起时一致；如果授权已过期，请返回连接器重新授权。'; }
      finally { this.busy = false; }
    },
  },
};
</script>
