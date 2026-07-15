<template>
  <div class="createDialog">
    <el-dialog
      class="reset-dialog-margin"
      :visible.sync="dialogVisible"
      width="800px"
      append-to-body
      :close-on-click-modal="false"
      :before-close="handleClose"
    >
      <template slot="title">
        <div class="dialog-title-wrapper">
          <span class="dialog-title">{{ $t('modelAccess.dialog.title') }}</span>
          <LinkText :text="$t('modelAccess.dialog.helpDoc')" type="model" />
        </div>
      </template>
      <div class="provider-search-wrapper">
        <el-input
          v-model="params.provider"
          prefix-icon="el-icon-search"
          class="no-border-input"
          style="width: calc(50% - 12px); margin-right: 18px"
          :placeholder="$t('modelAccess.dialog.providerName')"
          @keyup.enter.native="searchProviderList"
          @clear="searchProviderList()"
          clearable
        />
        <el-select
          v-model="params.modelType"
          :placeholder="$t('modelAccess.table.modelType')"
          class="no-border-select"
          style="width: calc(50% - 10px)"
          clearable
          @change="searchProviderList()"
        >
          <el-option
            v-for="item in modelTypeList"
            :key="item.key"
            :label="item.name"
            :value="item.key"
          />
        </el-select>
      </div>
      <div class="provider-card-wrapper">
        <div class="provider-card-content">
          <div
            :class="[
              'provider-card-item',
              { 'is-active': item.key === currentObj.key },
            ]"
            v-if="providerType && providerType.length"
            v-for="item in providerType"
            :key="item.key"
            @click="setCurrentSelected(item)"
          >
            <div class="provider-card-top">
              <div class="provider-card-top-left">
                <img
                  class="provider-card-img"
                  :src="providerImgObj[item.key]"
                  alt=""
                />
                <div class="provider-card-name">{{ item.name }}</div>
              </div>
              <div
                class="provider-check-icon"
                v-if="item.key === currentObj.key"
              >
                <i class="el-icon-check"></i>
              </div>
            </div>
            <div style="margin-top: 10px">
              <span
                class="provider-card-tag"
                v-for="it in item.children"
                :key="it.key"
              >
                {{ it.name }}
              </span>
            </div>
          </div>
          <el-empty
            class="noData"
            v-if="!(providerType && providerType.length)"
            :description="$t('common.noData')"
          ></el-empty>
        </div>
      </div>
      <span
        slot="footer"
        class="dialog-footer"
        style="padding-top: 0; margin-top: -10px"
      >
        <el-button @click="handleClose">
          {{ $t('common.button.cancel') }}
        </el-button>
        <el-button
          :disabled="!currentObj.key"
          type="primary"
          @click="handleConfirm"
        >
          {{ $t('common.button.confirm') }}
        </el-button>
      </span>
    </el-dialog>
  </div>
</template>
<script>
import {
  PROVIDER_TYPE,
  YUAN_JING,
  PROVIDER_IMG_OBJ,
  MODEL_TYPE,
} from '../constants';
import LinkText from '@/components/linkText.vue';
import { fetchProviderList } from '@/api/modelAccess';

export default {
  components: { LinkText },
  data() {
    return {
      dialogVisible: false,
      providerImgObj: PROVIDER_IMG_OBJ,
      providerType: [], // PROVIDER_TYPE,
      currentObj: {}, // PROVIDER_TYPE[0],
      yuanjing: YUAN_JING,
      modelTypeList: MODEL_TYPE,
      params: {
        provider: '',
        modelType: '',
      },
    };
  },
  methods: {
    openDialog() {
      this.dialogVisible = true;
      this.searchProviderList();
    },
    handleClose() {
      this.dialogVisible = false;
    },
    setCurrentSelected(item) {
      this.currentObj = item;
    },
    searchProviderList() {
      fetchProviderList(this.params).then(res => {
        this.providerType = res.data.list || [];
        this.currentObj = this.providerType[0] || {};
      });
    },
    handleConfirm() {
      this.handleClose();
      this.$emit('showCreate', this.currentObj, this.providerType);
    },
  },
};
</script>
<style lang="scss" scoped>
.provider-search-wrapper {
  padding: 0 22px 10px 24px;
}
.provider-card-wrapper {
  height: calc(100vh - 248px);
  overflow-y: auto;
  .provider-card-content {
    padding-left: 24px;
    padding-top: 5px;
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-start;
  }
  .noData {
    width: 100%;
    height: calc(100vh - 258px);
  }
}
.provider-card-item {
  width: calc(50% - 20px);
  margin-bottom: 20px;
  margin-right: 20px;
  border-radius: 8px;
  padding: 15px 10px 15px 20px;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  box-shadow: 0px 8px 10px 4px rgba(80, 98, 161, 0.07);
  height: 168px;
  .provider-card-top {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .provider-card-top-left {
    display: flex;
    align-items: center;
    justify-content: flex-start;
  }
  .provider-card-img {
    width: 50px;
    height: 50px;
    object-fit: contain;
    padding: 10px 6px;
    background: #ffffff;
    box-shadow: 0 1px 4px 0 rgba(0, 0, 0, 0.15);
    border-radius: 8px;
    border: 0 solid #d9d9d9;
    margin-right: 16px;
  }
  .provider-check-icon {
    width: 20px;
    height: 20px;
    line-height: 20px;
    border-radius: 50%;
    background-color: #fff;
    text-align: center;
    box-shadow: 0px 2px 8px 0px rgba(15, 17, 20, 0.1);
    margin-top: -5px;
    margin-right: 14px;
    i {
      color: $color;
      font-size: 13px;
    }
  }
  .provider-card-name {
    font-size: 16px;
    font-weight: bold;
    color: $color_title;
    margin-bottom: 5px;
  }
  .provider-card-tag {
    display: inline-block;
    margin-right: 8px;
    margin-top: 5px;
    font-size: 12px;
    border-radius: 4px;
    padding: 2px 8px;
    background: #f7f8fa;
    color: #55575f;
  }
}
.provider-card-item:hover,
.provider-card-item.is-active {
  box-shadow: 0 1px 4px 0 rgba(0, 0, 0, 0.15);
  border: 1px solid #d8d8d8;
}
.dialog-title-wrapper {
  display: flex;
  align-items: flex-end;
  .dialog-title {
    color: $color_title;
    font-size: 18px;
    font-weight: bold;
    margin-bottom: -2px;
  }
}
</style>
