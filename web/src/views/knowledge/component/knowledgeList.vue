<template>
  <div class="app-card-container">
    <div class="app-card">
      <div class="smart rl smart-create">
        <div class="app-card-create" @click="showCreate">
          <div class="create-img-wrap">
            <img class="create-type" src="@/assets/imgs/create_knowledge.svg" alt="" />
            <img class="create-img" src="@/assets/imgs/create_icon.png" alt="" />
            <div class="create-filter"></div>
          </div>
          <span>CreateKnowledge Base</span>
        </div>
      </div>
      <template v-if="listData && listData.length">
        <div class="smart rl" 
        v-for="(n,i) in listData" 
        :key="`${i}sm`" 
        @click.stop="toDocList(n)">
          <div>
              <img  class="logo" :src="require('@/assets/imgs/knowledgeIcon.png')" />
              <p :class="['smartDate']">{{n.docCount || 0}}Document</p>
          </div>
          <div class="info rl">
            <p class="name" :title="n.name">
              {{n.name}}
            </p>
            <el-tooltip
              v-if="n.description"
              popper-class="instr-tooltip tooltip-cover-arrow"
              effect="dark"
              :content="n.description"
              placement="bottom-start"
            >
              <p class="desc">{{n.description}}</p>
            </el-tooltip>
          </div>
          <div class="tags">
            <span :class="['smartDate','tagList']" v-if="formattedTagNames(n.knowledgeTagList).length === 0" @click.stop="addTag(n.knowledgeId,n)">
              <span class="el-icon-price-tag icon-tag"></span>
              AddTag
            </span>
            <span v-else @click.stop="addTag(n.knowledgeId,n)">{{formattedTagNames(n.knowledgeTagList) }}</span>
          </div>
          <div class="editor">
            <el-tooltip class="item" effect="dark" :content="n.orgName" placement="right-start">
              <span style="margin-right:52px; color:#999;font-size:12px;">{{n.orgName.length > 10 ? n.orgName.substring(0, 10) + '...' : n.orgName}}</span>
            </el-tooltip>
            <div v-if="n.share" class="publishType" style="right:22px;">
                <span v-if="n.share" class="publishType-tag"><span class="el-icon-unlock"></span> Public</span>
                <span v-else class="publishType-tag"><span class="el-icon-lock"></span> Private</span>
            </div>
            <el-dropdown @command="handleClick($event, n)" placement="top">
              <span class="el-dropdown-link">
                <i class="el-icon-more icon edit-icon" @click.stop></i>
              </span>
              <el-dropdown-menu slot="dropdown">
                <el-dropdown-item command="edit" v-if="[30].includes(n.permissionType)">{{$t('common.button.edit')}}</el-dropdown-item>
                <el-dropdown-item command="delete" v-if="[30].includes(n.permissionType)">{{$t('common.button.delete')}}</el-dropdown-item>
                <el-dropdown-item command="power" >Permission</el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </div>
        </div>
      </template>
    </div>
    <el-empty class="noData" v-if="!(listData && listData.length)" :description="$t('common.noData')"></el-empty>
    <tagDialog ref="tagDialog" @relodaData="relodaData" type="knowledge" :title="title"/>
    <PowerManagement ref="powerManagement"/>
  </div>
</template>

<script>
import { delKnowledgeItem } from "@/api/knowledge";
import { AppType } from "@/utils/commonSet"
import tagDialog from './tagDialog.vue';
import PowerManagement from './power/index.vue';
import {mapActions} from 'vuex';
export default {
  components:{tagDialog, PowerManagement},
  props:{
    appData:{
      type:Array,
      required:true,
      default:[]
    }
  },
  watch:{
    appData:{
      handler:function(val){
        this.listData = val
      },
      immediate:true,
      deep:true
    }
  },
  data(){
    return{
      apptype:AppType,
      basePath: this.$basePath,
      listData:[],
      title:'CreateTag'
    }
  },
  
  methods:{
  ...mapActions("app", ["setPermissionType","clearPermissionType"]),
  formattedTagNames(data){
    if(data.length === 0){
      return [];
    }
    const tags = data.filter(item => item.selected).map(item =>  item.tagName ).join(', ');
    if (tags.length > 30) {
        return tags.slice(0, 30) + '...';
    }
    return tags;
  },
  addTag(id,n){
    if([0].includes(n.permissionType)){
      this.$message.warning('无OperationPermission')
      return;
    }
    this.$nextTick(() =>{
      this.$refs.tagDialog.showDiaglog(id);
    })
  },
  showCreate(){
    this.$parent.showCreate();
  },
    handleClick(command,n){
      switch (command){
        case 'edit':
          this.editItem(n);
          break;
        case 'delete':
          this.deleteItem(n.knowledgeId)
          break;
        case 'power':
          this.showPowerManagement(n);
          break;
      }
    },
    editItem(row) {
      this.$emit('editItem', row)
    },
    relodaData(){
      this.$emit('reloadData');
    },
    deleteItem(knowledgeId){
      this.$confirm(this.$t('knowledgeManage.delKnowledgeTips'), this.$t('knowledgeManage.tip'), {
        confirmButtonText: this.$t('common.confirm.confirm'),
        cancelButtonText: this.$t('common.confirm.cancel'),
        type: "warning",
        beforeClose:(action, instance, done) =>{
          if(action === 'confirm'){
            instance.confirmButtonLoading = true;
            delKnowledgeItem({knowledgeId})
              .then(res =>{
                if(res.code === 0){
                  this.$message.success(this.$t('knowledgeManage.operateSuccess'));
                  this.$emit('reloadData')
                }
              })
              .catch(() => {})
              .finally(() =>{
                done();
                setTimeout(() => {
                  instance.confirmButtonLoading = false;
                }, 300);
              })
          }else{
            done()
          }
        }
      }).then(() => {})
    },
    toDocList(n){
      this.$router.push({path:`/knowledge/doclist/${n.knowledgeId}`});
      this.setPermissionType(n.permissionType)
    },
    showPowerManagement(knowledgeItem) {
      this.$refs.powerManagement.knowledgeId = knowledgeItem.knowledgeId;
      this.$refs.powerManagement.knowledgeName = knowledgeItem.knowledgeName;
      this.$refs.powerManagement.permissionType = knowledgeItem.permissionType;
      this.$refs.powerManagement.showDialog();
    },
  }
}
</script>

<style lang="scss" scoped>
@import "@/style/appCard.scss";
.app-card {
  .smart {
    height: 152px;
    .smartDate{
      // text-align:center;
      padding-top:3px;
      color:#888888;
    }
    .info {
      padding-right: 0;
    }
    .desc{
      padding-top: 5px;
    }
    .logo{
      border-radius:50%;
      background: #F1F4FF;
      box-shadow: none;
      padding: 5px !important;
      width: 65px !important;
      height: 65px !important;
    }
    .tagList{
      cursor: pointer;
      .icon-tag{
        transform: rotate(-40deg);
        margin-right:3px;
      }
    }
    .tagList:hover{
        color:$color;
    }
    .tag-knowledge{
      background:#826fff!important;
    }
  }
}
</style>
