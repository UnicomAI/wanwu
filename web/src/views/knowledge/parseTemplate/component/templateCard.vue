<template>
  <div class="template-card">
    <div class="card-row">
      <span class="card-label">
        {{ $t('knowledgeManage.parseTemplate.templateName') }}:
      </span>
      <el-input
        v-if="expanded && !form.builtIn"
        v-model="form.name"
        class="name-input"
        :placeholder="
          $t('knowledgeManage.parseTemplate.templateNamePlaceholder')
        "
        maxlength="50"
      />
      <span v-else class="name-text">{{ displayName }}</span>
      <div class="card-actions">
        <el-button type="text" @click="expanded = !expanded">
          {{ expanded ? $t('common.button.fold') : $t('common.button.expand') }}
        </el-button>
        <el-button type="text" v-if="!form.builtIn" @click="$emit('delete')">
          {{ $t('common.button.delete') }}
        </el-button>
      </div>
    </div>
    <div class="card-row card-body" v-if="expanded">
      <span class="card-label">
        {{ $t('knowledgeManage.parseTemplate.paramSetting') }}:
      </span>
      <div class="card-content">
        <el-form
          :model="form"
          ref="ruleForm"
          label-width="140px"
          label-position="left"
          @submit.native.prevent
        >
          <parseConfigForm
            ref="configForm"
            :value="form"
            :multiModal="multiModal"
            :mediaTypes="mediaTypes"
            :readonly="form.builtIn"
          />
        </el-form>
        <div class="card-footer" v-if="!form.builtIn">
          <el-button size="mini" type="primary" @click="handleSave">
            {{ $t('knowledgeManage.parseTemplate.save') }}
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
import parseConfigForm from '../../component/parseConfigForm.vue';

export default {
  name: 'TemplateCard',
  components: { parseConfigForm },
  props: {
    value: { type: Object, required: true },
    mediaType: { type: String, default: 'doc' },
    defaultExpanded: { type: Boolean, default: false },
  },
  data() {
    return {
      expanded: this.defaultExpanded,
      form: this.value,
    };
  },
  computed: {
    multiModal() {
      return this.mediaType !== 'doc';
    },
    mediaTypes() {
      return this.multiModal ? [this.mediaType] : [];
    },
    // 内置（默认）模板名字固定，不随库里存的值走
    displayName() {
      return this.form.builtIn
        ? this.$t('knowledgeManage.parseTemplate.builtIn')
        : this.form.name;
    },
  },
  watch: {
    value(val) {
      this.form = val;
    },
  },
  methods: {
    async handleSave() {
      if (!this.form.name) {
        this.$message.warning(
          this.$t('knowledgeManage.parseTemplate.nameRequired'),
        );
        return;
      }
      try {
        await this.$refs.ruleForm.validate();
      } catch {
        return;
      }
      this.$emit('save', this.form);
    },
  },
};
</script>
<style lang="scss" scoped>
.template-card {
  border: 1px solid #e6e8f0;
  border-radius: 8px;
  padding: 20px 24px;
  margin-bottom: 16px;
  background: #fff;
}
.card-row {
  display: flex;
  align-items: center;
}
.card-body {
  align-items: flex-start;
  margin-top: 20px;
}
.card-label {
  width: 90px;
  flex-shrink: 0;
  color: #333;
  line-height: 32px;
}
.name-input {
  width: 420px;
}
.name-text {
  color: #333;
}
.card-actions {
  margin-left: auto;
}
.card-content {
  flex: 1;
  min-width: 0;
}
.card-footer {
  display: flex;
  justify-content: flex-end;
}
</style>
