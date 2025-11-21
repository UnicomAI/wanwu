<template>
  <div class="text-to-image-example">
    <div class="page-header">
      <h1>Text-to-Image Configuration Example</h1>
      <el-button 
        type="primary" 
        @click="showDialog"
        icon="el-icon-plus"
      >
        Open Text-to-Image Configuration
      </el-button>
    </div>

    <!-- ConfigurationInformationShow -->
    <div v-if="configData" class="config-display">
      <h3>Current Configuration:</h3>
      <div class="config-item">
        <strong>API Key:</strong> {{ configData.apiKey }}
      </div>
      <div class="config-item">
        <strong>ParameterCount:</strong> {{ configData.parameters.length }}
      </div>
    </div>

    <!-- Text-to-Image Configuration Dialog -->
    <TextToImageDialog
      :visible.sync="dialogVisible"
      :initial-api-key="initialApiKey"
      @confirm="handleConfirm"
      @api-key-confirm="handleApiKeyConfirm"
      @api-key-update="handleApiKeyUpdate"
      @close="handleClose"
    />
  </div>
</template>

<script>
import TextToImageDialog from '@/components/TextToImageDialog.vue'

export default {
  name: 'TextToImageExample',
  components: {
    TextToImageDialog
  },
  data() {
    return {
      dialogVisible: false,
      initialApiKey: 'sk-1234567890abcdef', // Mocked existing API key
      configData: null
    }
  },
  methods: {
    showDialog() {
      this.dialogVisible = true;
    },
    handleConfirm(data) {
      console.log('ConfigurationConfirm:', data);
      this.configData = data;
      this.$message.success('Text-to-Image configuration saved');
    },
    handleApiKeyConfirm(apiKey) {
      console.log('API Key Confirm:', apiKey);
    },
    handleApiKeyUpdate(apiKey) {
      console.log('API Key Update:', apiKey);
    },
    handleClose() {
      console.log('Dialog closed');
    }
  }
}
</script>

<style lang="scss" scoped>
.text-to-image-example {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 30px;
    padding-bottom: 20px;
    border-bottom: 1px solid #e4e7ed;

    h1 {
      margin: 0;
      color: #333;
    }
  }

  .config-display {
    background: #f8f9fa;
    padding: 20px;
    border-radius: 8px;
    margin-top: 20px;

    h3 {
      margin: 0 0 15px 0;
      color: #333;
    }

    .config-item {
      margin-bottom: 10px;
      padding: 8px 0;
      
      strong {
        color: #666;
        margin-right: 10px;
      }
    }
  }
}
</style>

