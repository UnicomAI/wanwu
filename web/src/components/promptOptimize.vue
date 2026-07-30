<template>
  <div>
    <el-dialog
      title=""
      :visible.sync="dialogVisible"
      width="700"
      append-to-body
      :close-on-click-modal="false"
      :before-close="handleClose"
    >
      <div class="prompt-optimize-dialog-content">
        <div>
          <modelSelect
            v-model="modelId"
            :options="modelOptions"
            :placeholder="$t('tempSquare.selectModel')"
            @visible-change="visibleChange"
            :loading-text="$t('tempSquare.loadingText')"
            :loading="modelLoading"
            :filterable="true"
            style="width: 400px; margin-right: 10px"
          />
          <el-button
            :disabled="!(prompt && modelId)"
            size="mini"
            type="primary"
            @click="promptOptimize"
            :loading="loading"
          >
            {{ $t('tempSquare.optimize') }}
          </el-button>
        </div>
        <div v-if="optimizedPrompt" class="answer-content">
          <el-input
            type="textarea"
            :rows="16"
            v-model="optimizedPrompt"
          ></el-input>
        </div>
      </div>
      <span slot="footer" class="dialog-footer">
        <el-button
          :disabled="!optimizedPrompt"
          type="primary"
          @click="doSubmit"
        >
          {{ $t('tempSquare.replace') }}
        </el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { selectModelList } from '@/api/modelAccess';
import Print from '@/utils/printPlus2.js';
import { fetchEventSource } from '@/sse/index.js';
import { USER_API } from '@/utils/requestConstants';
import ModelSelect from '@/components/modelSelect.vue';
import { formatReqUrl } from '@/utils/util';

export default {
  components: { ModelSelect },
  data() {
    return {
      dialogVisible: false,
      modelId: '',
      prompt: '',
      optimizedPrompt: '',
      modelOptions: [],
      modelLoading: false,
      loading: false,
      eventSource: null,
      ctrlAbort: null,
    };
  },
  created() {
    this.getModelData();
  },
  beforeDestroy() {
    this.stopEventSource();
    this._print && this._print.stop();
  },
  methods: {
    getModelData() {
      this.modelLoading = true;
      selectModelList()
        .then(res => {
          this.modelOptions = (res.data.list || []).filter(
            item => item.config && item.config.visionSupport !== 'support',
          );
          this.modelLoading = false;
        })
        .catch(() => (this.modelLoading = false));
    },
    visibleChange(val) {
      if (val) this.getModelData();
    },
    openDialog(item) {
      const { prompt } = item || {};
      this.dialogVisible = true;
      this.prompt = prompt || '';
      this.optimizedPrompt = '';
      this.modelId = '';
    },
    handleClose() {
      this.stopEventSource();
      this.prompt = '';
      this.dialogVisible = false;
      this.loading = false;
    },
    stopEventSource() {
      this.ctrlAbort && this.ctrlAbort.abort();
      this.eventSource = null;
    },
    promptOptimize() {
      const params = {
        modelId: this.modelId,
        prompt: this.prompt,
      };
      this.loading = true;

      const origin = window.location.origin;
      const user = this.$store.state.user || {};
      const token = user.token;
      const userInfo = user.userInfo || {};
      this._print = new Print({});
      let endStr = '';

      this.ctrlAbort = new AbortController();
      this.eventSource = new fetchEventSource(
        origin + formatReqUrl(`${USER_API}/prompt/optimize`),
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: 'Bearer ' + token,
            'x-user-id': userInfo.uid,
            'x-org-id': userInfo.orgId,
          },
          openWhenHidden: true,
          signal: this.ctrlAbort.signal,
          body: JSON.stringify({ ...params }),
          onopen: async e => {
            if (e.status !== 200) {
              this.$message.error(e.statusText);
              this.stopEventSource();
              this.loading = false;
            }
          },
          onmessage: e => {
            if (e && e.data) {
              const data = JSON.parse(e.data) || {};
              let _sentence = data.response;
              this._print.print(
                {
                  response: _sentence,
                  finish: data.finish,
                },
                {},
                worldObj => {
                  endStr = endStr + worldObj.world;
                  this.optimizedPrompt = endStr;
                  if (Boolean(worldObj.finish)) this.loading = false;
                },
              );
            }
          },
          onerror: () => {
            this.loading = false;
            this.stopEventSource();
          },
        },
      );
    },
    doSubmit() {
      this.$emit('promptSubmit', this.optimizedPrompt);
      this.handleClose();
    },
  },
};
</script>

<style lang="scss" scoped>
.prompt-optimize-dialog-content {
  .model-select {
    width: calc(100% - 100px);
    margin-right: 15px;
  }
  .answer-content {
    padding: 15px 0;
  }
}
.model-option-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;

  .model-name {
    flex-shrink: 0;
    font-weight: 500;
  }

  .model-select-tags {
    display: flex;
    flex-wrap: nowrap;
    gap: 4px;
    flex-shrink: 0;
    margin-top: 4px;

    .model-select-tag {
      background-color: #f0f2ff;
      color: $color;
      border-radius: 4px;
      padding: 2px 8px;
      font-size: 10px;
      line-height: 1.2;
    }
  }
}
</style>
