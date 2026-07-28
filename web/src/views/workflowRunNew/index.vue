<template>
  <div class="wp right-content-full">
    <iframe id="inlineFrameExample" :src="workflowUrl"></iframe>
  </div>
</template>

<script>
export default {
  name: 'ExploreWorkflow',
  data() {
    const query = `?workflow_id=${this.$route.query.id}&space_id=${this.$store.state.user.userInfo.orgId}`;
    return {
      workflowUrl:
        process.env.NODE_ENV === 'development'
          ? `http://localhost:8282/workflow/run${query}`
          : window.location.origin + `/workflow/run${query}`,
    };
  },
};
</script>

<style lang="scss" scoped>
.wp {
  height: calc(100vh - 3px);
  background: #f4f5f6;
  iframe {
    width: 100%;
    height: 100%;
    border: none;
  }
}
</style>
