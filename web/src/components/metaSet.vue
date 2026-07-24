<template>
  <div class="metaSet">
    <div v-if="!readonly" class="tool-typ">
      <el-button
        icon="el-icon-plus"
        type="primary"
        @click="addMetaItem"
        size="small"
      >
        {{ $t('metaSet.add') }}
      </el-button>
      <el-switch
        v-model="metaDataFilterParams.filterEnable"
        active-color="var(--color)"
        @change="switchChange"
      ></el-switch>
    </div>
    <div class="docMetaData">
      <div
        v-if="readonly && !metaDataFilterParams.metaFilterParams.length"
        class="emptyMetaTip"
      >
        {{ $t('metaSet.notConfigured') }}
      </div>
      <div
        v-else
        :class="[
          'docMetaBox',
          metaDataFilterParams.metaFilterParams.length > 1
            ? 'docMetaContainer'
            : '',
        ]"
      >
        <div
          v-for="(item, index) in metaDataFilterParams.metaFilterParams"
          class="docItem"
          :key="`metaItem-${index}`"
        >
          <div class="docItem_data">
            <el-select
              v-model="item.key"
              :disabled="readonly"
              :placeholder="$t('common.select.placeholder') + 'key'"
              @change="keyChange($event, item, index)"
            >
              <template #prefix>
                <img class="prefix" src="@/assets/imgs/key.png" />
              </template>
              <el-option
                key="createNew"
                :label="$t('metaData.createMetaData')"
                value="createNew"
              ></el-option>
              <el-option
                v-for="meta in keyOptions"
                :key="meta.metaKey"
                :label="meta.metaKey + ' | ' + '[ ' + meta.metaValueType + ' ]'"
                :value="meta.metaKey"
              ></el-option>
            </el-select>
          </div>
          <el-divider direction="vertical"></el-divider>
          <div class="docItem_data">
            <el-select
              v-model="item.condition"
              :disabled="readonly"
              :placeholder="$t('metaSet.conditionPlaceholder')"
              @change="conditionChange($event, item)"
            >
              <template #prefix>
                <img
                  class="prefix"
                  src="@/assets/imgs/condition.png"
                  style="width: 18px"
                />
              </template>
              <el-option
                v-for="item in conditionOptions[item.type]"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              ></el-option>
            </el-select>
          </div>
          <el-divider direction="vertical"></el-divider>
          <div class="docItem_data">
            <el-input
              v-model="item.value"
              v-if="item.type === 'string' || item.type === ''"
              @blur="metaValueBlur(item)"
              :placeholder="$t('common.input.placeholder') + 'value'"
              :disabled="
                readonly ||
                item.condition === 'empty' ||
                item.condition === 'not empty'
              "
            >
              <template #prefix>
                <img
                  class="prefix"
                  src="@/assets/imgs/value.png"
                  style="width: 16px"
                />
              </template>
            </el-input>
            <el-input
              v-model="item.value"
              v-if="item.type === 'number'"
              @blur="metaValueBlur(item)"
              type="number"
              placeholder="number"
              :disabled="
                readonly ||
                item.condition === 'empty' ||
                item.condition === 'not empty'
              "
            >
              <template #prefix>
                <img
                  class="prefix"
                  src="@/assets/imgs/number.png"
                  style="width: 16px"
                />
              </template>
            </el-input>
            <el-date-picker
              v-if="item.type === 'time'"
              v-model="item.value"
              align="right"
              format="yyyy-MM-dd HH:mm:ss"
              value-format="timestamp"
              type="datetime"
              :placeholder="$t('common.datePicker.placeholder')"
              :disabled="
                readonly ||
                item.condition === 'empty' ||
                item.condition === 'not empty'
              "
            ></el-date-picker>
          </div>
          <el-divider direction="vertical"></el-divider>
          <div class="docItem_data docItem_data_btn">
            <span
              v-if="!readonly"
              class="el-icon-delete setBtn"
              @click="delMetaItem(index)"
            ></span>
          </div>
        </div>
        <el-select
          v-if="metaDataFilterParams.metaFilterParams.length > 1"
          v-model="metaDataFilterParams.filterLogicType"
          :disabled="readonly"
          class="orAnd"
          :placeholder="$t('metaSet.conditionPlaceholder')"
        >
          <el-option
            v-for="item in conditions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          ></el-option>
        </el-select>
      </div>
    </div>
    <createMetadata ref="createMetadata" @updateMeta="getList"></createMetadata>
  </div>
</template>
<script>
import { metaSelect } from '@/api/knowledge';
import createMetadata from '@/components/createMetadata.vue';

export default {
  components: {
    createMetadata,
  },
  props: {
    knowledgeId: {
      type: String,
      required: true,
      default: '',
    },
    currentMetaData: {
      type: Object,
      default: () => ({}),
    },
    readonly: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      metaDataFilterParams: {
        filterEnable: false,
        filterLogicType: 'and',
        metaFilterParams: [],
      },
      keyOptions: [],
      conditions: [
        {
          value: 'and',
          label: this.$t('metaSet.and'),
        },
        {
          value: 'or',
          label: this.$t('metaSet.or'),
        },
      ],
      conditionOptions: {
        time: [
          {
            value: 'is',
            label: this.$t('metaSet.is'),
          },
          {
            value: 'before',
            label: this.$t('metaSet.before'),
          },
          {
            value: 'after',
            label: this.$t('metaSet.after'),
          },
          {
            value: 'empty',
            label: this.$t('metaSet.empty'),
          },
          {
            value: 'not empty',
            label: this.$t('metaSet.notEmpty'),
          },
        ],
        string: [
          {
            value: 'is',
            label: this.$t('metaSet.is'),
          },
          {
            value: 'is not',
            label: this.$t('metaSet.not'),
          },
          {
            value: 'contains',
            label: this.$t('metaSet.contains'),
          },
          {
            value: 'not contains',
            label: this.$t('metaSet.notContains'),
          },
          {
            value: 'start with',
            label: this.$t('metaSet.startsWith'),
          },
          {
            value: 'end with',
            label: this.$t('metaSet.endsWith'),
          },
          {
            value: 'empty',
            label: this.$t('metaSet.empty'),
          },
          {
            value: 'not empty',
            label: this.$t('metaSet.notEmpty'),
          },
        ],
        number: [
          {
            value: '=',
            label: this.$t('metaSet.equal'),
          },
          {
            value: '≠',
            label: this.$t('metaSet.notEqual'),
          },
          {
            value: '>',
            label: this.$t('metaSet.moreThan'),
          },
          {
            value: '≥',
            label: this.$t('metaSet.moreThanOrEqual'),
          },
          {
            value: '<',
            label: this.$t('metaSet.lessThan'),
          },
          {
            value: '≤',
            label: this.$t('metaSet.lessThanOrEqual'),
          },
          {
            value: 'empty',
            label: this.$t('metaSet.empty'),
          },
          {
            value: 'not empty',
            label: this.$t('metaSet.notEmpty'),
          },
        ],
      },
    };
  },
  watch: {
    currentMetaData: {
      handler: function (val, old) {
        if (val === null) {
          this.metaDataFilterParams = {
            filterEnable: false,
            filterLogicType: 'and',
            metaFilterParams: [],
          };
          return;
        }

        if (Object.keys(val).length > 0) {
          const copy = JSON.parse(JSON.stringify(val));
          copy.metaFilterParams = copy.metaFilterParams || [];
          this.metaDataFilterParams = copy;
        }
      },
      immediate: true,
      deep: true,
    },
    knowledgeId: {
      handler(val, old) {
        if (val && !this.readonly) {
          if (val !== old) {
            this.getList();
          }
        }
      },
      immediate: true,
    },
  },
  created() {
    // this.getList()
  },
  methods: {
    switchChange(val) {
      if (!val) {
        this.metaDataFilterParams.metaFilterParams = [];
      }
    },
    conditionChange(e, item) {
      item.value = '';
    },
    getMetaData() {
      this.metaDataFilterParams.metaFilterParams =
        this.metaDataFilterParams.metaFilterParams.map(item => {
          if (item.type === 'time') {
            return {
              ...item,
              value: String(item.value),
            };
          }
          return item;
        });
      return { metaDataFilterParams: this.metaDataFilterParams };
    },
    getList() {
      metaSelect({ knowledgeId: this.knowledgeId }).then(res => {
        if (res.code === 0) {
          this.keyOptions = res.data.knowledgeMetaList || [];
        }
      });
    },
    metaValueBlur(item) {
      if (item.condition === 'empty' || item.condition === 'not empty') {
        return true;
      } else {
        if (!item.value) {
          this.$message.warning(this.$t('metaSet.valuePlaceholder'));
          return;
        }
      }
    },
    isEmpty(value) {
      if (
        value === null ||
        value === 'null' ||
        value === undefined ||
        value === ''
      )
        return true;
      return false;
    },
    validateRequiredFields(data) {
      return data.some(field => {
        if (
          field &&
          typeof field === 'object' &&
          (field.condition === 'empty' || field.condition === 'not empty')
        ) {
          return false;
        }
        if (field && typeof field === 'object' && 'value' in field) {
          return this.isEmpty(field.value);
        }
      });
    },
    addMetaItem() {
      if (this.metaDataFilterParams.filterEnable === false) {
        this.$message.warning(this.$t('metaSet.filterEnable'));
        return;
      }
      if (this.metaDataFilterParams.metaFilterParams.length > 0) {
        if (
          this.validateRequiredFields(
            this.metaDataFilterParams.metaFilterParams,
          )
        ) {
          this.$message.warning(this.$t('metaSet.filterValidate'));
          return;
        }
      }
      this.metaDataFilterParams.metaFilterParams.push({
        key: '',
        type: '',
        condition: '',
        value: '',
      });
    },
    clearData() {
      this.metaDataFilterParams.metaFilterParams = [];
      this.metaDataFilterParams.filterLogicType = 'and';
      this.metaDataFilterParams.filterEnable = false;
    },
    keyChange(val, item) {
      if (val === 'createNew') {
        this.$refs.createMetadata.showDialog(this.knowledgeId);
        item.key = '';
        return;
      }
      item.value = '';
      item.condition = '';
      item.type = this.keyOptions
        .filter(i => i.metaKey === val)
        .map(e => e.metaValueType)[0];
    },
    delMetaItem(index) {
      this.metaDataFilterParams.metaFilterParams.splice(index, 1);
      if (this.metaDataFilterParams.metaFilterParams.length === 0) {
        this.metaDataFilterParams.filterLogicType = 'and';
      }
    },
  },
};
</script>
<style lang="scss" scoped>
::v-deep {
  .el-dialog__body {
    padding: 10px 20px;
  }

  .el-divider--vertical {
    margin: 0;
  }

  .el-input__prefix {
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .el-input__inner {
    padding: 15px 20px 15px 30px !important;
  }

  .el-icon-time {
    color: #727ff9;
  }
}

.metaSet {
  width: 100%;

  .tool-typ {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .docMetaData {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .docMetaBox {
      width: 100%;
    }

    .emptyMetaTip {
      width: 100%;
      padding: 16px 0;
      text-align: center;
      color: #909399;
      font-size: 14px;
    }

    .docMetaContainer {
      position: relative;
      margin-left: 65px;
      margin-top: 15px;
    }

    .docMetaContainer::after {
      content: '';
      display: block;
      position: absolute;
      left: -20px;
      top: 50%;
      bottom: 0;
      width: 15px;
      height: 90%;
      transform: translateY(-50%);
      border-top-left-radius: 8px;
      border-bottom-left-radius: 8px;
      border: 1px solid rgba(16, 24, 40, 0.1411764706);
      border-right-width: 0;
    }

    .docItem {
      display: flex;
      flex: 1;
      align-items: center;
      border-radius: 8px;
      background: #f7f8fa;
      margin-top: 10px;
      width: 100%;

      .docItem_data {
        width: 32%;
        display: flex;
        align-items: center;
        flex-wrap: wrap;
        padding: 5px 10px;

        .el-input,
        .el-select,
        .el-date-picker {
          width: 100%;
        }

        .prefix {
          width: 14px;
          margin-left: 5px;
        }

        .docItem_data_label {
          margin-right: 8px;
          display: flex;
          align-items: center;

          .question {
            color: #aaadcc;
            margin-left: 2px;
            cursor: pointer;
          }
        }

        .setBtn {
          font-size: 16px;
          cursor: pointer;
          color: $btn_bg;
        }
      }

      .docItem_data_type {
        width: 80px;
      }

      .docItem_data_btn {
        display: flex;
        justify-content: center;
        width: 30px;

        .el-icon-delete {
          margin-left: 5px;
        }
      }
    }

    .orAnd {
      width: 65px;
      position: absolute;
      left: 0;
      top: 50%;
      transform: translate(-110%, -50%);
      padding: 6px 2px;
      z-index: 1;
    }
  }
}
</style>
