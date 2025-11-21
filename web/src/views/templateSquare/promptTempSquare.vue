<template>
  <div class="tempSquare-management">
    <div class="tempSquare-content-box tempSquare-third">
      <div class="tempSquare-main">
        <div class="tempSquare-content">
          <div class="tempSquare-card-box">
            <div class="card-search card-search-cust">
              <div>
                <span
                  v-for="item in typeList"
                  :key="item.key"
                  :class="['tab-span', {'is-active': typeRadio === item.key}]"
                  @click="changeTab(item.key)"
                >
                  {{ item.name }}
                </span>
              </div>
              <search-input
                style="margin-right: 2px"
                :placeholder="$t('tempSquare.searchText')"
                ref="searchInput"
                @handleSearch="doGetPromptTempList"
              />
            </div>

            <div class="card-loading-box" v-if="list.length">
              <div class="card-box" v-loading="loading">
                <div
                  class="card"
                  v-for="(item, index) in list"
                  :key="index"
                  @click.stop="handleClick(item)"
                >
                  <div class="card-title">
                    <img
                      class="card-logo"
                      v-if="item.avatar && item.avatar.path"
                      :src="basePath + '/user/api/' + item.avatar.path"
                    />
                    <div class="mcp_detailBox">
                      <span class="mcp_name">{{ item.name }}</span>
                      <span class="mcp_from">
                        <label>
                          {{$t('tempSquare.author')}}：{{ item.author }}
                        </label>
                      </span>
                    </div>
                  </div>
                  <div class="card-des">{{ item.desc }}</div>
                  <div class="card-bottom" style="display: block; width: 100%; text-align: right">
                    <i
                      v-if="!isPublic"
                      class="el-icon-copy-document"
                      :title="$t('tempSquare.copy')"
                      @click.stop="copyPromptTemplate(item)"
                    ></i>
                  </div>
                </div>
              </div>
            </div>
            <div v-else class="empty">
              <el-empty :description="$t('common.noData')"></el-empty>
            </div>
          </div>
        </div>
      </div>
    </div>
    <CreatePrompt :type="promptType" ref="clonePromptDialog" />
  </div>
</template>
<script>
import { getPromptTempList } from "@/api/templateSquare"
import SearchInput from "@/components/searchInput.vue"
import CreatePrompt from "@/components/createApp/createPrompt.vue"
export default {
  components: { SearchInput, CreatePrompt },
  props: {
    isPublic: true,
    type: ''
  },
  data() {
    return {
      basePath: this.$basePath,
      category: this.$t('tempSquare.all'),
      list: [],
      loading: false,
      promptType: 'copy',
      typeRadio: 'all',
      typeList: [
        {name: this.$t('tempSquare.all'), key: 'all'},
        {name: this.$t('tempSquare.safety_inspection'), key: 'safety_inspection'},
        {name: this.$t('tempSquare.incident_investigation'), key: 'incident_investigation'},
        {name: this.$t('tempSquare.compliance'), key: 'compliance'},
        {name: this.$t('tempSquare.corrective_actions'), key: 'corrective_actions'},
        {name: this.$t('tempSquare.training'), key: 'training'},
        {name: this.$t('tempSquare.audit'), key: 'audit'},
        {name: this.$t('tempSquare.reporting'), key: 'reporting'},
      ]
    };
  },
  mounted() {
    this.doGetPromptTempList()
  },
  methods: {
    changeTab(key) {
      this.typeRadio = key
      this.$refs.searchInput.value = ''
      this.doGetPromptTempList()
    },
    doGetPromptTempList() {
      const searchInput = this.$refs.searchInput
      let params = {
        name: searchInput.value,
        category: this.typeRadio,
      }

      getPromptTempList(params)
        .then((res) => {
          const {list} = res.data || {}
          this.list = list || []
          this.loading = false
        })
        .catch(() => this.loading = false)
    },
    showPromptDialog(item) {
      this.$refs.clonePromptDialog.openDialog(item)
    },
    copyPromptTemplate(item) {
      this.promptType = 'copy'
      this.showPromptDialog(item)
    },
    handleClick(item) {
      this.promptType = 'detail'
      this.showPromptDialog(item)
    },
  },
};
</script>

<style lang="scss" scoped>
@import "@/style/tempSquare.scss";
</style>