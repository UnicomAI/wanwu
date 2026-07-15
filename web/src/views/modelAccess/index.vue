<template>
  <div class="page-wrapper modelAccess">
    <div class="table-wrap list-common wrap-fullheight">
      <!--<div class="page-title">
        <img class="page-title-img" src="@/assets/imgs/model.svg" alt="" />
        <span class="page-title-name">{{ $t('modelAccess.title') }}</span>
      </div>-->
      <div class="tabs tabs-x-top">
        <div
          v-for="item in isSystem
            ? tabList.filter(({ type }) => !type)
            : tabList"
          :key="item.type"
          :class="['tab', { active: type === item.type }]"
          @click="tabClick(item.type)"
        >
          {{ item.name }}
        </div>
      </div>
      <div class="table-box">
        <div class="table-form">
          <el-select
            v-model="params.provider"
            :placeholder="$t('modelAccess.table.publisher')"
            class="modelAccess-select no-border-select"
            clearable
            @change="searchData()"
          >
            <el-option
              v-for="item in providerList"
              :key="item.key"
              :label="item.name"
              :value="item.key"
            />
          </el-select>

          <el-select
            v-model="params.modelType"
            :placeholder="$t('modelAccess.table.modelType')"
            class="modelAccess-select no-border-select"
            style="margin-left: 15px"
            clearable
            @change="searchData()"
          >
            <el-option
              v-for="item in modelTypeList"
              :key="item.key"
              :label="item.name"
              :value="item.key"
            />
          </el-select>

          <el-select
            v-model="params.scopeType"
            :placeholder="$t('modelAccess.table.scopeType')"
            class="modelAccess-select no-border-select"
            style="margin-left: 15px"
            clearable
            @change="searchData()"
          >
            <el-option
              v-for="item in getScopeTypeList(isSystem)"
              :key="item.key"
              :label="item.name"
              :value="item.key"
            ></el-option>
          </el-select>
          <div
            style="
              width: 100%;
              float: right;
              text-align: right;
              margin-top: -30px;
            "
          >
            <el-input
              v-model="params.displayName"
              prefix-icon="el-icon-search"
              class="no-border-input"
              style="width: 240px; margin-right: 10px"
              :placeholder="$t('modelAccess.table.modelName')"
              @keyup.enter.native="searchName"
              @clear="searchData()"
              clearable
            />
            <el-button
              class="add-bt"
              size="mini"
              type="primary"
              @click="goModelComparison"
            >
              <img
                style="
                  width: 14px;
                  margin-right: 5px;
                  display: inline-block;
                  vertical-align: middle;
                "
                src="@/assets/imgs/modelComparison.png"
                alt=""
              />
              <span style="display: inline-block; vertical-align: middle">
                {{ $t('modelExprience.modelComparison') }}
              </span>
            </el-button>
            <el-button
              class="add-bt"
              size="mini"
              type="primary"
              @click="preInsert"
            >
              <img
                style="
                  width: 14px;
                  margin-right: 5px;
                  display: inline-block;
                  vertical-align: middle;
                "
                src="@/assets/imgs/modelImport.png"
                alt=""
              />
              <span style="display: inline-block; vertical-align: middle">
                {{ $t('modelAccess.import') }}
              </span>
            </el-button>
          </div>
        </div>
        <div class="card-wrapper scroll-card-container">
          <div class="card-item card-item-create">
            <div class="app-card-create" @click="preInsert">
              <div class="create-img-wrap">
                <img
                  class="create-type"
                  src="@/assets/imgs/card_add_icon.svg"
                  alt=""
                />
                <img
                  class="create-img"
                  src="@/assets/imgs/create_model.svg"
                  alt=""
                />
              </div>
              <span>{{ $t('modelAccess.import') }}</span>
            </div>
          </div>
          <div
            v-if="tableData && tableData.length"
            class="card-item"
            v-for="(item, index) in tableData"
            :key="item.model + index"
            @click="preUpdate(item)"
          >
            <div class="card-top">
              <img
                class="common-card-logo ml-10"
                :src="
                  item.avatar && item.avatar.path
                    ? avatarSrc(item.avatar.path)
                    : defaultLogo
                "
              />
              <div
                :class="
                  item.modelType === LLM && item.allowEdit
                    ? 'card-title-with-select'
                    : 'card-title'
                "
              >
                <el-tooltip
                  placement="top"
                  :content="item.displayName || item.model"
                >
                  <div class="card-name">
                    {{ item.displayName || item.model }}
                  </div>
                </el-tooltip>
              </div>
              <div class="card-top-right" @click.stop="">
                <el-switch
                  v-if="item.allowEdit"
                  @change="
                    val => {
                      changeStatus(item, val);
                    }
                  "
                  style="width: 32px"
                  v-model="item.isActive"
                  active-text=""
                  inactive-text=""
                />
                <el-checkbox
                  style="margin-left: 10px; margin-top: -2px"
                  v-if="item.modelType === LLM"
                  :disabled="!item.isActive"
                  :model-value="checkModelSelection(item.modelId)"
                  @change="setModelCheck(item.modelId)"
                ></el-checkbox>
                <el-dropdown
                  v-if="item.allowEdit"
                  @command="handleCommand"
                  placement="top"
                >
                  <span class="el-dropdown-link">
                    <i class="el-icon-more more"></i>
                  </span>
                  <el-dropdown-menu slot="dropdown">
                    <el-dropdown-item :command="{ type: 'edit', item }">
                      <i class="el-icon-edit-outline card-opera-icon"></i>
                      {{ $t('common.button.edit') }}
                    </el-dropdown-item>
                    <el-dropdown-item
                      class="card-delete"
                      :command="{ type: 'delete', item }"
                    >
                      <i class="el-icon-delete card-opera-icon" />
                      {{ $t('common.button.delete') }}
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </el-dropdown>
              </div>
            </div>
            <div class="card-middle">
              <div
                v-if="item.tags"
                class="card-type"
                v-for="(it, itIndex) in item.tags"
                :style="{
                  color: getTagColor(itIndex).color,
                  background: getTagColor(itIndex).backgroundColor,
                }"
              >
                {{ it.text }}
              </div>
            </div>
            <div class="card-bottom">
              <el-tooltip
                placement="top"
                :content="providerObj[item.provider] || '--'"
              >
                <div
                  :class="[
                    'card-bottom-provider',
                    { 'no-publishData': !item.updatedAt },
                  ]"
                >
                  {{ $t('modelAccess.table.publisher') }}:
                  {{ providerObj[item.provider] || '--' }}
                </div>
              </el-tooltip>
              <div>
                {{ item.updatedAt ? item.updatedAt.split(' ')[0] : '--' }}
                {{ $t('modelAccess.table.update') }}
              </div>
            </div>
            <div class="card-btn" v-if="item.modelType === LLM">
              <el-button
                size="mini"
                type="primary"
                round
                :disabled="!item.isActive"
                @click.stop="goModelExprience(item.modelId)"
              >
                {{ $t('modelExprience.createConversation') }}
              </el-button>
            </div>
          </div>
        </div>
        <el-empty
          class="noData"
          v-if="!(tableData && tableData.length)"
          :description="$t('common.noData')"
        ></el-empty>
      </div>
      <CreateSelectDialog ref="createSelectDialog" @showCreate="showCreate" />
      <CreateDialog ref="createDialog" @reloadData="searchData" />
    </div>
  </div>
</template>

<script>
import Pagination from '@/components/pagination.vue';
import {
  fetchModelList,
  deleteModel,
  changeModelStatus,
  getModelDetail,
} from '@/api/modelAccess';
import CreateDialog from './components/createDialog.vue';
import CreateSelectDialog from './components/createSelectDialog.vue';
import {
  LLM,
  MODEL_TYPE_OBJ,
  PROVIDER_OBJ,
  PROVIDER_TYPE,
  MODEL_TYPE,
  SCOPE_TYPE_LIST,
  ORG,
  TAB_LIST,
  SCOPE_TYPE_OBJ,
} from './constants';
import { avatarSrc, getModelDefaultIcon } from '@/utils/util';
import { TagColorList } from '@/utils/commonSet';

export default {
  name: 'ModelAccess',
  components: { Pagination, CreateSelectDialog, CreateDialog },
  data() {
    return {
      LLM,
      listApi: fetchModelList,
      isSystem: this.$store.state.user.permission.isSystem || false,
      providerList: PROVIDER_TYPE,
      modelTypeList: MODEL_TYPE,
      basePath: this.$basePath,
      modelTypeObj: MODEL_TYPE_OBJ,
      providerObj: PROVIDER_OBJ,
      defaultLogo: getModelDefaultIcon(),
      tableData: [],
      params: {
        provider: '',
        modelType: '',
        displayName: '',
        scopeType: '',
      },
      loading: false,
      modelSelection: [],
      type: '',
      tabList: TAB_LIST,
    };
  },
  computed: {
    checkModelSelection() {
      return model => {
        return this.modelSelection.includes(model);
      };
    },
  },
  created() {
    this.type = this.$route.query.type || '';
  },
  mounted() {
    this.getTableData();
  },
  methods: {
    avatarSrc,
    getScopeTypeList() {
      // 系统管理员非组织筛选，普通用户全部-都可以筛选，公有模型非个人筛选，我的模型非全局筛选
      return this.isSystem
        ? SCOPE_TYPE_LIST.filter(item => item.key !== ORG)
        : SCOPE_TYPE_OBJ[this.type] || SCOPE_TYPE_LIST;
    },
    getTagColor(index) {
      return TagColorList[index % TagColorList.length];
    },
    tabClick(type) {
      this.type = type;
      this.clearParams();
      this.getTableData();
    },
    async getTableData(params) {
      this.loading = true;
      try {
        const res = await fetchModelList({ filterScope: this.type, ...params });
        const tableData = res.data ? res.data.list || [] : [];
        this.tableData = [...tableData];
      } finally {
        this.loading = false;
      }
    },
    clearParams() {
      for (let key in this.params) {
        this.params[key] = '';
      }
    },
    searchData(isCreate) {
      if (isCreate) {
        this.clearParams();
        this.type = '';
      }
      this.getTableData({ ...this.params });
    },
    searchName(e) {
      if (e.keyCode === 13) {
        this.searchData();
      }
    },
    handleCommand(value) {
      const { type, item } = value || {};
      switch (type) {
        case 'edit':
          this.preUpdate(item);
          break;
        case 'delete':
          this.preDel(item);
          break;
      }
    },
    preInsert() {
      this.$refs.createSelectDialog.openDialog();
    },
    showCreate(item, providerType) {
      this.$refs.createDialog &&
        this.$refs.createDialog.openDialog(item.key, null, providerType);
    },
    preUpdate(row) {
      const { modelId, provider } = row || {};

      getModelDetail({ modelId }).then(res => {
        const rowObj = res.data || {};
        const newRow = { ...rowObj, ...rowObj.config };
        this.$refs.createDialog &&
          this.$refs.createDialog.openDialog(provider, newRow);
      });
    },
    preDel(row) {
      this.$confirm(
        this.$t('modelAccess.confirm.delete'),
        this.$t('common.confirm.title'),
        {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          type: 'warning',
        },
      ).then(async () => {
        const { modelId } = row || {};
        let res = await deleteModel({ modelId });
        if (res.code === 0) {
          this.$message.success(this.$t('common.message.success'));
          await this.getTableData();
        }
      });
    },
    changeStatus(row, val) {
      this.$confirm(
        val
          ? this.$t('modelAccess.confirm.start')
          : this.$t('modelAccess.confirm.stop'),
        this.$t('common.confirm.title'),
        {
          confirmButtonText: this.$t('common.confirm.confirm'),
          cancelButtonText: this.$t('common.confirm.cancel'),
          type: 'warning',
        },
      )
        .then(async () => {
          const { modelId } = row || {};
          let res = await changeModelStatus({ modelId, isActive: val });
          if (res.code === 0) {
            this.$message.success(this.$t('common.message.success'));
            await this.searchData();
          }
        })
        .catch(() => {
          this.searchData();
        });
    },
    goModelExprience(modelId) {
      this.$router.push({
        path: 'modelAccess/modelExprience',
        query: { modelId },
      });
    },
    goModelComparison() {
      const length = this.modelSelection.length;
      if (!length) {
        this.$message.warning(this.$t('modelExprience.warning.selectModel'));
        return;
      }
      if (length > 4) {
        this.$message.warning(
          this.$t('modelExprience.tip.maxSelectModel').replace('@', 4),
        );
        return;
      }
      this.$router.push({
        path: 'modelAccess/modelExprience',
        query: { comparisonIds: this.modelSelection.join(',') },
      });
    },
    setModelCheck(modelId) {
      if (this.modelSelection.includes(modelId)) {
        this.modelSelection = this.modelSelection.filter(
          item => item !== modelId,
        );
      } else {
        this.modelSelection.push(modelId);
      }
    },
  },
};
</script>

<style lang="scss" scoped>
@import '@/style/tabs.scss';
.routerview-container {
  top: 0;
}
.table-box {
  padding: 20px 20px 0;
  .table-form {
    width: 100%;
    padding-bottom: 20px;
    clear: both;
  }
  .table-header {
    font-size: 16px;
    font-weight: bold;
    color: #555;
  }
  .add-bt {
    margin: 0 2px 20px 5px;
  }
}
.modelAccess-select {
  width: 200px;
}
.mark-textArea ::v-deep {
  .el-textarea__inner {
    font-family: inherit;
    font-size: inherit;
  }
}
.card-wrapper {
  margin: 0 -10px;
}
.card-item {
  display: inline-block;
  width: calc((100% / 4) - 20px);
  height: 165px;
  vertical-align: middle;
  margin: 0 10px 20px;
  background: url('@/assets/imgs/card_bg.png');
  background-size: 100% 100%;
  box-shadow: 0 8px 10px 0 rgba(22, 52, 156, 0.07);
  border-radius: 8px;
  padding: 18px 10px 16px 14px;
  position: relative;
  cursor: pointer;
  overflow: hidden;
  .card-top {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .card-title {
    width: calc(100% - 90px);
  }
  .card-title-with-select {
    width: calc(100% - 140px);
  }
  .card-name {
    font-size: 16px;
    color: $create_card_text_color;
    font-weight: 600;
    width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    display: -webkit-box;
    -webkit-box-orient: vertical;
    line-clamp: 2;
    -webkit-line-clamp: 2;
    word-break: break-word;
  }
  .card-middle {
    padding-top: 10px;
  }
  .card-type {
    display: inline-block;
    padding: 0 3px;
    border-radius: 3px;
    color: $color;
    background: $color_opacity;
    margin-top: 5px;
    margin-right: 8px;
    font-size: 12px;
  }
  .card-top-right {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    margin-left: 5px;
  }
  .more {
    margin-left: 5px;
    cursor: pointer;
    transform: rotate(90deg);
    font-size: 16px;
    color: #8c8c8f;
    padding: 5px 5px 4px 5px;
    border-radius: 4px;
  }
  .more:hover {
    background: #fff;
    box-shadow: 0 4px 8px 0 rgba(0, 0, 0, 0.06);
  }
  .card-bottom {
    position: absolute;
    color: #686f82;
    bottom: 14px;
    left: 15px;
    right: 12px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    .card-bottom-provider {
      width: calc(100% - 105px);
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    .no-publishData.no-publishData {
      width: calc(100% - 45px);
    }
  }
  &:hover {
    .card-btn {
      transform: translateY(0);
      opacity: 1;
    }
  }
  .card-btn {
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    padding: 12px;
    transition: all 0.4s;
    transform: translateY(100%);
    opacity: 0;
    .el-button {
      width: 100%;
    }
  }
}
.card-item:hover {
  border: 1px solid $color;
}
.card-item-create {
  background: #fff;
  box-shadow: 0 8px 10px 0 rgba(80, 98, 161, 0.07);
  border: 1px solid $create_card_border_color;
  .app-card-create {
    width: 100%;
    height: 100%;
    text-align: center;
    display: flex;
    align-items: center;
    justify-content: center;
    .create-img-wrap {
      display: inline-block;
      vertical-align: middle;
      margin-right: 30px;
      position: relative;
      .create-img {
        width: 44px;
        height: 46px;
        border-radius: 6px;
        background: linear-gradient(44deg, #edc1ff 0%, #1486ff 100%);
        padding: 5px;
        box-shadow: 0 10px 16px 0 rgba(236, 190, 255, 0.5);
      }
      .create-type {
        width: 32px;
        height: 32px;
        position: absolute;
        background: linear-gradient(
          180deg,
          rgba(197, 222, 255, 0.3) 0%,
          rgba(255, 255, 255, 0.3) 100%
        );
        backdrop-filter: blur(8px);
        border: 1px solid #d3c2ff;
        border-radius: 5px;
        bottom: -8px;
        right: -16px;
      }
    }
    span {
      display: inline-block;
      vertical-align: middle;
      font-size: 16px;
      color: $create_card_text_color;
      font-weight: bold;
    }
  }
}
::v-deep .el-dropdown-menu__item.card-delete:hover {
  color: #ff4d4f !important;
  background: #fbeae8 !important;
}
.card-opera-icon {
  font-size: 15px;
}
.modelAccess .noData {
  width: 100%;
  text-align: center;
  margin-top: -60px;
  ::v-deep .el-empty__description p {
    color: #b3b1bc;
  }
}
</style>
