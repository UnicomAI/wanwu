<template>
  <CommonLayout
    :showAside="false"
  >
    <template #main-content>
      <div class="app-content">
        <div class="app-header">
           <div class="header-top">
              <div class="taglist_warp">
                <div v-for="item in tagList" class="tagList" @click="handleTagClick(item)" :class="{'white':item.value === active}">
                  <img :src="item.value === active ? item.activeImg : item.unactiveImg" class="h-icon" />
                  <span>{{item.name}}</span>
                </div>
              </div>
                <!-- <SearchInput :placeholder="placeholder" style="width:200px;" @handleSearch="handleSearch"/> -->
           </div>
           <div>
            <el-tabs v-model="activeName" @tab-click="handleClick">
              <el-tab-pane label="智能体" name="agent">
                <AppList :appData="listData" :isShowTool="false" :appFrom="'explore'" />
              </el-tab-pane>
              <el-tab-pane label="文本问答" name="rag">
                <AppList :appData="listData" :isShowTool="false" :appFrom="'explore'" />
              </el-tab-pane>
              <!-- <el-tab-pane label="工作流" name="workflow">
                 <AppList :appData="listData" :isShowTool="false" :appFrom="'explore'" />
              </el-tab-pane> -->
            </el-tabs>
           </div>
        </div>
      </div>
    </template>
  </CommonLayout>
</template>

<script>
import {mapActions, mapGetters} from 'vuex'
import CommonLayout from '@/components/exploreContainer.vue'
import SearchInput from "@/components/searchInput.vue"
import AppList from "@/components/appList.vue"
import CreateTotalDialog from "@/components/createTotalDialog.vue"
import { getExplorList } from "@/api/explore"

export default {
  components: { SearchInput, CreateTotalDialog, AppList,CommonLayout },
  data() {
    return {
      placeholder:'输入应用名称搜索',
      asideTitle:'历史应用',
      activeName:'agent',
      searchValue:'',
      active:'all',
      tagList:[
        {
          name:'全部',
          value:'all',
          activeImg:require('@/assets/imgs/all_active.png'),
          unactiveImg:require('@/assets/imgs/all_unactive.png')
        },
        {
          name:'我收藏的',
          value:'favorite',
          activeImg:require('@/assets/imgs/mine_active.png'),
          unactiveImg:require('@/assets/imgs/mine_unactive.png')
        },
        {
          name:'私密发布的',
          value:'private',
          activeImg:require('@/assets/imgs/start_active.png'),
          unactiveImg:require('@/assets/imgs/start_unactive.png')
        },
        {
          name:'历史应用',
          value:'history',
          activeImg:require('@/assets/imgs/history_active.png'),
          unactiveImg:require('@/assets/imgs/history_unactive.png')
        }
      ],
      historyList:[],
      listData:[]
    }
  },
  watch: {
    historyAppList:{
      handler(val){
        if(val){
          this.historyList = val;
        }
      }
    }
  },
  computed: {
    // ...mapGetters('app', ['historyAppList'])
  },
  created(){
    // this.getHistoryList();
    this.getExplorData(this.activeName,this.active)
  },
  mounted() {
  },
  methods: {
    // ...mapActions('app', ['getHistoryList']),
    handleSearch(value){
      this.searchValue = value;
      this.getExplorData(this.activeName,this.active)
    },
    historyClick(n){
      if(!n.path)return;
      this.$router.push({path: n.path});
    },
    handleClick(){
      this.getExplorData(this.activeName,this.active)
    },
    handleTagClick(item){
      this.active = item.value
      this.getExplorData(this.activeName,this.active)
    },
    getExplorData(appType,searchType){
      const data = {name:this.searchValue,appType,searchType}
      getExplorList(data).then(res =>{
          if(res.code === 0){
            this.listData = res.data.list || []
          }
      }).catch((err) =>{
          this.$message.error(err)
      })
    }
  }
}
</script>
<style lang="scss" scoped>
/deep/{
  .el-tabs__content{
    overflow: unset;
  }
  .table-search-input{
    height:30px;
  }
}
.white{
  background: #fff;
  color: #384BF7;
}
.explore-aside-app{
  .appList:hover{
    background-color: $color_opacity !important;
  }
  .appList{
    margin:10px 20px;
    padding:10px;
    border-radius: 6px;
    margin-bottom: 6px;
    display: flex;
    gap: 8px;
    align-items: center;
    cursor: pointer;
    .appImg{
      width:30px;
      height:30px;
      border-radius: 50%;
      object-fit: cover;
    }
    .appName{
      display: block;
      max-width: 130px;
      overflow: hidden;
      white-space: nowrap;
      pointer-events: none;
      text-overflow: ellipsis;
    }
  }
  
}
.app-content{
  padding: 20px;
  margin:20px;
  .header-top{
    display: flex;
    justify-content: space-between;
    padding:15px 0;
    border-bottom: 1px solid #eaeaea;
    .tagList:nth-child(1){
      margin-left: 0 !important;
    }
    .taglist_warp{
      display:flex;
      .tagList{
        margin:10px;
        padding:10px;
        border-radius:6px;
        cursor: pointer;
        display:flex;
        align-items:center;
        .h-icon{
          margin-right: 5px;
          width:14px;
        }
      }
    }
  }
}
</style>