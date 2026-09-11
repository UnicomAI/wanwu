<template>
  <div class="page-wrapper page-wrapper-pr-none">
    <!--<div class="page-title">
      <img class="page-title-img" src="@/assets/imgs/knowledge.svg" alt="" />
      <span class="page-title-name">{{ $t('knowledgeManage.knowledge') }}</span>
    </div>-->
    <div style="padding: 20px">
      <div class="tabs" style="padding-bottom: 20px">
        <div
          :class="['tab', { active: category === KNOWLEDGE }]"
          @click="tabClick(KNOWLEDGE)"
        >
          {{ $t('menu.knowledge') }}
        </div>
        <div
          :class="['tab', { active: category === QA }]"
          @click="tabClick(QA)"
        >
          {{ $t('knowledgeManage.qaDatabase.title') }}
        </div>
        <!--        <div-->
        <!--          :class="['tab', { active: category === DB }]"-->
        <!--          @click="tabClick(DB)"-->
        <!--        >-->
        <!--          {{ $t('knowledgeManage.dbDatabase.title') }}-->
        <!--        </div>-->
      </div>
      <div class="search-box header-form-pr">
        <div class="no-border-input">
          <search-input
            class="cover-input-icon"
            :placeholder="
              $t('knowledgeManage.searchPlaceholder', {
                category: CATEGORY_TYPE_OBJ[category],
              })
            "
            ref="searchInput"
            @handleSearch="getTableData"
          />
          <el-select
            style="margin-right: 15px"
            v-model="tagIds"
            :placeholder="$t('knowledgeManage.selectTag')"
            multiple
            @visible-change="tagChange"
            @remove-tag="removeTag"
            v-if="category === KNOWLEDGE"
          >
            <el-option
              v-for="item in tagOptions"
              :key="item.tagId"
              :label="item.tagName"
              :value="item.tagId"
            ></el-option>
          </el-select>
          <el-select
            style="margin-right: 15px"
            v-model="external"
            :placeholder="$t('knowledgeManage.selectExternal')"
            @visible-change="getTableData"
            v-if="category === KNOWLEDGE"
          >
            <el-option
              v-for="item in externalOptions"
              :key="item.value"
              :label="item.name"
              :value="item.value"
            ></el-option>
          </el-select>
        </div>
        <div>
          <el-button
            size="mini"
            type="primary"
            @click="$router.push('/knowledge/parseTemplate')"
            v-if="category === KNOWLEDGE"
          >
            {{ $t('knowledgeManage.parseTemplate.title') }}
          </el-button>
          <el-button
            size="mini"
            type="primary"
            @click="$router.push('/knowledge/keyword')"
            v-if="category === KNOWLEDGE"
          >
            {{ $t('knowledgeManage.keyWordManage') }}
          </el-button>
          <el-button
            size="mini"
            type="primary"
            @click="showDrawer()"
            icon="el-icon-plus"
            v-if="category === KNOWLEDGE"
          >
            {{ $t('knowledgeManage.externalAPI.title') }}
          </el-button>
        </div>
      </div>
      <knowledgeList
        :appData="knowledgeData"
        @editItem="editItem"
        @exportItem="exportItem"
        @reloadData="getTableData"
        ref="knowledgeList"
        v-loading="tableLoading"
        :category="category"
      />
      <createKnowledge
        ref="createKnowledge"
        @reloadData="getTableData"
        @createExternalApi="showDrawer"
        :category="category"
      />
      <externalAPIDrawer ref="externalAPIDrawer" @update="updateExternalAPI" />
    </div>
  </div>
</template>
<script>
import { getKnowledgeList, tagList, exportDoc } from '@/api/knowledge';
import SearchInput from '@/components/searchInput.vue';
import knowledgeList from './component/knowledgeList.vue';
import createKnowledge from './component/create.vue';
import { qaDocExport } from '@/api/qaDatabase';
import ExternalAPIDrawer from '@/components/externalAPIDrawer.vue';
import {
  ALL,
  INTERNAL,
  EXTERNAL,
  KNOWLEDGE,
  QA,
  DB,
  CATEGORY_TYPE_OBJ,
} from '@/views/knowledge/constants';
export default {
  name: 'Knowledge',
  components: {
    ExternalAPIDrawer,
    SearchInput,
    knowledgeList,
    createKnowledge,
  },
  provide() {
    return {
      reloadKnowledgeData: this.getTableData,
    };
  },
  data() {
    return {
      KNOWLEDGE,
      QA,
      DB,
      CATEGORY_TYPE_OBJ,
      knowledgeData: [],
      tableLoading: false,
      tagOptions: [],
      tagIds: [],
      category: KNOWLEDGE,
      external: ALL,
      externalOptions: [
        {
          name: this.$t('knowledgeManage.all'),
          value: ALL,
        },
        {
          name: this.$t('knowledgeManage.internal'),
          value: INTERNAL,
        },
        {
          name: this.$t('knowledgeManage.external'),
          value: EXTERNAL,
        },
      ],
    };
  },
  activated() {
    this.getTableData();
    this.getList();
  },
  methods: {
    showDrawer() {
      this.$refs.externalAPIDrawer.showDrawer();
    },
    updateExternalAPI() {
      this.$refs.createKnowledge.getExternalAPIList();
    },
    tabClick(status) {
      this.category = status;
      this.$refs.searchInput.clearValue();
      this.tagIds = [];
      this.external = ALL;
      this.getTableData();
    },
    getList() {
      tagList({ knowledgeId: '', tagName: '' }).then(res => {
        if (res.code === 0) {
          this.tagOptions = res.data.knowledgeTagList || [];
        }
      });
    },
    tagChange(val) {
      if (!val && this.tagIds.length > 0) {
        this.getTableData();
      } else {
        this.getList();
      }
    },
    removeTag() {
      this.getTableData();
    },
    getTableData() {
      const searchInput = this.$refs.searchInput.value;
      this.tableLoading = true;
      getKnowledgeList({
        name: searchInput,
        tagId: this.tagIds,
        category: this.category,
        external: this.external,
      })
        .then(res => {
          this.knowledgeData = res.data.knowledgeList || [];
          this.tableLoading = false;
        })
        .catch(error => {
          this.tableLoading = false;
        });
    },
    clearIptValue() {
      this.$refs.searchInput.clearValue();
    },
    editItem(row) {
      this.$refs.createKnowledge.showDialog(row);
    },
    exportItem(row) {
      const params = {
        knowledgeId: row.knowledgeId,
      };
      const exportApi = this.category === KNOWLEDGE ? exportDoc : qaDocExport;
      exportApi(params).then(res => {
        if (res.code === 0) {
          this.$message.success(this.$t('common.message.success'));
        }
      });
    },
    showCreate() {
      this.$refs.createKnowledge.showDialog();
    },
  },
};
</script>
<style lang="scss" scoped>
@import '@/style/tabs.scss';
.search-box {
  display: flex;
  justify-content: space-between;
  padding-bottom: 20px;
}

::v-deep {
  .el-loading-mask {
    background: none !important;
  }
}
</style>
