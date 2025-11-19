<template>
  <div class="page-wrapper">
    <div class="page-title">
      <img class="page-title-img" :src="typeObj[type] ? typeObj[type].img : require('@/assets/imgs/task.png')" alt="" />
      <span class="page-title-name">{{typeObj[type] ? typeObj[type].title : $t('appSpace.title')}}</span>
    </div>
    <div class="hide-loading-bg" style="padding: 20px" v-loading="loading">
      <search-input :placeholder="$t('appSpace.search')" ref="searchInput" @handleSearch="handleSearch" />
      <div class="header-right">
        <el-button size="mini" type="primary" @click="showImport" v-if="type === 'workflow'">
          {{$t('common.button.import')}}
        </el-button>
        <el-button size="mini" type="primary" @click="showCreate" icon="el-icon-plus" v-if="validateAgent()">
          {{$t('common.button.create')}}
        </el-button>
      </div>
      <!-- <div v-if="type === 'agent'" class="agent_type_switch">
          <div v-for="item in agentSwitch" :class="{'agentActive':item.type === agnet_type }" class="agent_type_item" @click="agentType_change(item)" :key="item.type">{{item.name}}</div>
      </div> -->
      <AppList
        :agnetType="agnet_type"
        :type="type"
        :showCreate="showCreate"
        :appData="listData"
        :isShowPublished="true"
        :isShowTool="true"
        @reloadData="getTableData"
      />
      <CreateTotalDialog ref="createTotalDialog" />
      <UploadFileDialog
        @reloadData="getTableData"
        :title="$t('appSpace.workflowExport')"
        ref="uploadFileDialog"
      />
    </div>
  </div>
</template>

<script>
import SearchInput from "@/components/searchInput.vue"
import AppList from "@/components/appList.vue"
import CreateTotalDialog from "@/components/createTotalDialog.vue"
import UploadFileDialog from "@/components/uploadFileDialog.vue"
import { getAppSpaceList,agnetTemplateList } from "@/api/appspace"
import { mapGetters} from 'vuex'
import {fetchPermFirPath} from "@/utils/util";

export default {
  components: { SearchInput, CreateTotalDialog, UploadFileDialog, AppList },
  data() {
    return {
      type: '',
      loading: false,
      listData:[],
      typeObj: {
        workflow: {title: 'Workflow', img: require('@/assets/imgs/workflow_icon.svg')},
        rag: {title: 'Text Q&A', img: require('@/assets/imgs/rag.svg')},
        agent: {title: 'Agent', img: require('@/assets/imgs/agent.svg')}
      },
      currentTypeObj: {},
      agnet_type:'auto',
      agentSwitch:[
        {
          type:'template',
          name:'Agent Template'
        },
        {
          type:'auto',
          name:'CustomAgent'
        }
      ]
    }
  },
  watch: {
    $route: {
      handler(val) {
        const {type} = val ? val.params || {} : {}
        this.listData = []
        this.type = type
        this.$refs.searchInput.value = ''
        this.justifyRenderPage(type)
        // this.getTypeData();
        this.getTableData()
      },
      // 深度ObserveListen
      deep: true
    },
    fromList:{
      handler(val){
        if(val !== ''){
          this.type = val;
          // this.getTypeData();
          this.getTableData()
        }
      }
    }
  },
  computed: {
    ...mapGetters('app', ['fromList']),
  },
  mounted() {
    const {type} = this.$route.params || {}
    this.type = type
    this.justifyRenderPage(type)
    // this.getTypeData();
    this.getTableData();
  },
  methods: {
    justifyRenderPage(type) {
      if (!['workflow', 'agent', 'rag'].includes(type)) {
        const {path} = fetchPermFirPath()
        this.$router.push({path})
      }
    },
    getTypeData(){
      if(this.type === 'agent'){
      this.agnet_type = 'template'
      this.getAgentTemplate();
      }else{
        this.getTableData();
      }
    },
    handleSearch(){
      if(this.type === 'agent' && this.agnet_type === 'template'){
        this.getAgentTemplate();
      }else{
        this.getTableData();
      }
    },
    validateAgent(){
      if(this.type === 'agent' && this.agnet_type === 'template'){
        return false
      }
      return true
    },
    getAgentTemplate(){
      this.loading = true
      const searchInput = this.$refs.searchInput.value
      agnetTemplateList({category:'',name:searchInput}).then(res =>{
        if(res.code === 0){
          this.loading = false
          this.listData = res.data 
            ? res.data.list.map(item => ({
                ...item,
                isShowCopy: false
              })) 
            : [];
          }
      }).catch(() => {
        this.loading = false
        this.listData = []
      })
    },
    agentType_change(item){
      this.agnet_type = item.type;
      this.$refs.searchInput.value = '';
      if(this.agnet_type === 'auto'){
        this.getTableData();
      }else{
        this.getAgentTemplate();
      }
    },
    getTableData() {
      this.loading = true
      const searchInput = this.$refs.searchInput
      const searchInfo = {
        appType: this.type === 'all' ? '' : this.type,
        ...searchInput.value && {name: searchInput.value}
      }
      getAppSpaceList(searchInfo).then(res => {
        this.loading = false
        this.listData = res.data ? res.data.list || [] : []
      }).catch(() => {
        this.loading = false
        this.listData = []
      })
    },
    showImport() {
      this.$refs.uploadFileDialog.openDialog()
    },
    showCreate() {
      switch (this.type) {
        case 'agent':
          this.$refs.createTotalDialog.showCreateIntelligent()
          break
        case 'rag':
          this.$refs.createTotalDialog.showCreateTxtQues()
          break
        case 'workflow':
          this.$refs.createTotalDialog.showCreateWorkflow()
          break
        default:
          this.$refs.createTotalDialog.openDialog()
          break
      }
    }
  }
}
</script>
<style lang="scss" scoped>
.header-right {
  display: inline-block;
  float: right;
}
.agent_type_switch{
  margin-top:20px;
  width:300px;
  height:40px;
  border-bottom: 1px solid #333;
  display:flex;
  justify-content:space-between;
  .agent_type_item{
    cursor: pointer;
    height:100%;
    width:50%;
    color:#333;
    border-radius:4px;
    text-align:center;
    line-height:40px;
    font-size: 14px;
  }
  .agentActive{
    color: #fff;
    background: #333;
    border-radius: 0;
    font-weight: bold;
  }
}
</style>