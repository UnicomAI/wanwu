<template>
  <div>
    <el-dialog
      :title="$t('app.modelSet')"
      :visible.sync="dialogVisible"
      width="50%"
      v-bind="$attrs"
      :before-close="handleClose"
    >
      <span>
        <el-form
          :model="ruleForm"
          ref="ruleForm"
          label-width="100px"
          class="demo-ruleForm"
        >
          <el-form-item
            :label="item.label"
            :prop="item.props"
            v-for="(item, index) in modelSet"
            :key="index"
          >
            <el-row>
              <el-col :span="1">
                <el-tooltip
                  class="item"
                  effect="light"
                  :content="item.desc"
                  placement="bottom"
                >
                  <span class="el-icon-question question"></span>
                </el-tooltip>
              </el-col>
              <el-col :span="2">
                <el-switch v-model="ruleForm[item.btnProps]"></el-switch>
              </el-col>
              <el-col :span="20">
                <el-slider
                  v-if="!item.hideSlider"
                  v-model="ruleForm[item.props]"
                  show-input
                  :min="item.min"
                  :max="item.max"
                  :step="item.step"
                  :disabled="!ruleForm[item.btnProps]"
                ></el-slider>
              </el-col>
            </el-row>
          </el-form-item>
        </el-form>
      </span>
      <span slot="footer" class="dialog-footer">
        <el-button
          :type="isRuleFormChanged ? 'primary' : ''"
          :disabled="!isRuleFormChanged"
          @click="resetForm"
        >
          {{ $t('common.button.reset') }}
        </el-button>
        <el-button @click="dialogVisible = false">
          {{ $t('common.button.cancel') }}
        </el-button>
        <el-button type="primary" @click="submit">
          {{ $t('common.button.confirm') }}
        </el-button>
      </span>
    </el-dialog>
  </div>
</template>
<script>
const getDefaultRuleForm = () => ({
  temperature: 0.7,
  topP: 1,
  frequencyPenalty: 0,
  presencePenalty: 0,
  maxTokens: 512,
  temperatureEnable: false,
  topPEnable: false,
  presencePenaltyEnable: false,
  maxTokensEnable: false,
  frequencyPenaltyEnable: false,
});

const cloneForm = form => JSON.parse(JSON.stringify(form));

export default {
  props: {
    modelform: {
      type: Object,
      default: null,
    },
    limitMaxTokens: {
      type: Number,
      default: 4096,
    },
  },
  data() {
    return {
      dialogVisible: false,
      ruleForm: getDefaultRuleForm(),
      initialRuleForm: getDefaultRuleForm(),
    };
  },
  computed: {
    isRuleFormChanged() {
      return (
        JSON.stringify(this.ruleForm) !== JSON.stringify(this.initialRuleForm)
      );
    },
    modelSet() {
      const baseModelSet = [
        {
          label: this.$t('app.temperature'),
          desc: this.$t('app.temperatureDesc'),
          props: 'temperature',
          btnProps: 'temperatureEnable',
          min: 0,
          max: 2,
          step: 0.1,
        },
        {
          label: 'TopP',
          desc: this.$t('app.topDesc'),
          props: 'topP',
          btnProps: 'topPEnable',
          min: 0,
          max: 1,
          step: 0.01,
        },
        {
          label: this.$t('app.frequencyPenalty'),
          desc: this.$t('app.frequencyPenaltyDesc'),
          props: 'frequencyPenalty',
          btnProps: 'frequencyPenaltyEnable',
          min: -2,
          max: 2,
          step: 1,
        },
        {
          label: this.$t('app.punishmentExists'),
          desc: this.$t('app.punishmentExistsDesc'),
          props: 'presencePenalty',
          btnProps: 'presencePenaltyEnable',
          min: -2,
          max: 2,
          step: 1,
        },
        {
          label: this.$t('app.maxTokens'),
          desc: this.$t('app.maxTokensDesc'),
          props: 'maxTokens',
          btnProps: 'maxTokensEnable',
          min: 1,
          max: this.limitMaxTokens,
          step: 1,
        },
      ];
      if ('thinkingEnable' in this.ruleForm) {
        baseModelSet.push({
          label: this.$t('app.thinking'),
          desc: this.$t('app.thinkingDesc'),
          props: 'thinking',
          btnProps: 'thinkingEnable',
          hideSlider: true,
        });
      }

      return baseModelSet;
    },
  },
  methods: {
    showDialog() {
      this.dialogVisible = true;
      const data =
        this.modelform !== null
          ? cloneForm(this.modelform)
          : getDefaultRuleForm();
      this.ruleForm = data;
      this.initialRuleForm = cloneForm(data);
    },
    resetForm() {
      this.ruleForm = cloneForm(this.initialRuleForm);
    },
    handleClose() {
      this.dialogVisible = false;
    },
    submit() {
      this.dialogVisible = false;
      this.$emit('setModelSet', this.ruleForm);
    },
  },
};
</script>
<style lang="scss" scoped>
::v-deep {
  .el-input-number--small {
    line-height: 28px !important;
  }
}
.question {
  cursor: pointer;
  color: #ccc;
}
</style>
