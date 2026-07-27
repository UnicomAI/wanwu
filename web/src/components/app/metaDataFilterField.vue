<template>
  <el-dialog
    :visible.sync="metaSetVisible"
    width="80%"
    class="metaSetVisible"
    :before-close="handleMetaClose"
  >
    <template #title>
      <div class="metaHeader">
        <h3>{{ $t('agent.form.configMetaDataFilter') }}</h3>
        <span>
          {{
            category === 0
              ? $t('agent.form.metaDataFilterDesc')
              : $t('agent.form.metaDataQaFilterDesc')
          }}
        </span>
      </div>
    </template>
    <metaSet
      ref="metaSet"
      :knowledgeId="knowledgeId"
      :currentMetaData="metaData"
      :readonly="readonly"
    />
    <span slot="footer" class="dialog-footer">
      <el-button @click="handleMetaClose">
        {{ $t('common.button.cancel') }}
      </el-button>
      <el-button v-if="!readonly" type="primary" @click="submitMeta">
        {{ $t('common.button.confirm') }}
      </el-button>
    </span>
  </el-dialog>
</template>

<script>
import metaSet from '@/components/metaSet';
export default {
  name: 'MetaDataFilterField',
  components: {
    metaSet,
  },
  props: {
    knowledgeId: {
      type: String,
      default: '',
    },
    metaData: {
      type: Object,
      default: () => {},
    },
    category: {
      type: Number,
      default: 0,
    },
    readonly: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      metaSetVisible: false,
      currentKnowledgeId: '',
      currentMetaData: {},
    };
  },
  methods: {
    handleMetaClose() {
      this.metaSetVisible = false;
    },
    submitMeta() {
      const metaData = this.$refs.metaSet.getMetaData();
      if (
        this.$refs.metaSet.validateRequiredFields(
          metaData['metaDataFilterParams']['metaFilterParams'],
        )
      ) {
        this.$message.warning(this.$t('agent.form.incompleteInfo'));
        return;
      }
      this.$emit('submitMetaData', metaData);
      this.metaSetVisible = false;
    },
    showDialog() {
      this.metaSetVisible = true;
    },
  },
};
</script>
