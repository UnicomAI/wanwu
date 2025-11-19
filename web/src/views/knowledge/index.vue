<template>
  <div class="page-wrapper">
    <div class="page-title">
      <img class="page-title-img" src="@/assets/imgs/knowledge.svg" alt="" />
      <span class="page-title-name">{{$t('knowledgeManage.knowledge')}}</span>
    </div>
    <div style="padding: 20px">
      <div class="search-box">
        <div class="no-border-input">
          <search-input class="cover-input-icon" :placeholder="$t('knowledgeManage.searchPlaceholder')" ref="searchInput" @handleSearch="getTableData" />
          <el-select v-model="tagIds" placeholder="Please select标签" multiple @visible-change="tagChange" @remove-tag="removeTag">
            <el-option
              v-for="item in tagOptions"
              :key="item.tagId"
              :label="item.tagName"
              :value="item.tagId">
            </el-option>
          </el-select>
        </div>
        <div>
          <el-button size="mini" type="primary" @click="$router.push('/knowledge/keyword')">{{$t('knowledgeManage.keyWordManage')}}</el-button>
          <el-button size="mini" type="primary" @click="showCreate()" icon="el-icon-plus">
            {{$t('common.button.create')}}
          </el-button>
        </div>
      </div>
      <knowledgeList :appData="knowledgeData" @editItem="showCreate" @reloadData="getTableData" ref="knowledgeList" v-loading="tableLoading" />
      <createKnowledge ref="createKnowledge" @reloadData="getTableData" />
    </div>
  </div>
</template>
<script>
import { getKnowledgeList,tagList } from "@/api/knowledge";
import SearchInput from "@/components/searchInput.vue"
import knowledgeList from './component/knowledgeList.vue';
import createKnowledge from './component/create.vue';
export default {
    components: { SearchInput,knowledgeList,createKnowledge },
    provide(){
      return {
        reloadKnowledgeData: this.getTableData
      }
    },
    data(){
       return{
        knowledgeData:[],
        tableLoading:false,
        tagOptions:[],
        tagIds:[]
       } 
    },
    mounted(){
      this.getTableData();
      this.getList();
    },
    methods:{
      getList(){
        tagList({knowledgeId:'',tagName:''}).then(res => {
            if(res.code === 0){
                this.tagOptions = res.data.knowledgeTagList || []
            }
          })
        },
        tagChange(val){
          if(!val && this.tagIds.length > 0){
              this.getTableData()
          }else{
            this.getList()
          }
        },
        removeTag(){
          this.getTableData()
        },
        getTableData(){
            const searchInput = this.$refs.searchInput.value
            this.tableLoading = true
            getKnowledgeList({name:searchInput,tagId:this.tagIds}).then(res => {
                this.knowledgeData = res.data.knowledgeList || [];
                this.tableLoading = false
            }).catch((error) =>{
                this.tableLoading = false
                this.$message.error(error)
            })
        },
        clearIptValue(){
          this.$refs.searchInput.clearValue()
        },
        showCreate(row){
            this.$refs.createKnowledge.showDialog(row)
        }
    }
}
</script>
<style lang="scss" scoped>
.search-box {
 display:flex;
 justify-content:space-between;
}

/deep/{
  .el-loading-mask{
    background: none !important;
  }
}

</style>
