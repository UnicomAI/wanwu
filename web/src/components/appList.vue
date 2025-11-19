<template>
  <div class="app-card-container">
    <div class="app-card">
      <div class="smart rl smart-create" v-if="isShowTool && validateAgnet()">
        <div class="app-card-create" @click="showCreate">
          <div class="create-img-wrap">
            <img v-if="type" class="create-type" :src="require(`@/assets/imgs/create_${type}.png`)" alt="" />
            <img class="create-img" src="@/assets/imgs/create_icon.png" alt="" />
            <div class="create-filter"></div>
          </div>
          <span>{{`${$t('common.button.add')}${apptype[type] || ''}`}}</span>
        </div>
      </div>
      <div
        v-if="listData && listData.length"
        class="smart rl"
        v-for="(n,i) in listData"
        :key="`${i}sm`"
        :style="`cursor: ${isCanClick(n) ? 'pointer' : 'default'} !important;`"
        @click.stop="isCanClick(n) && toEdit(n)"
        @mouseenter="mouseEnter(n)"
        @mouseleave="mouseLeave(n)"
      >
        <el-image v-if="n.avatar && n.avatar.path" class="logo" lazy :src="basePath + '/user/api/' + n.avatar.path" :key="`${i}-${n.appId}-avatar`"></el-image>
        <span :class="['tag-app', `${n.appType}-tag`]">{{apptype[n.appType] || ''}}</span>
        <img
          v-if="apptype[n.appType]"
          class="tag-img"
          src="@/assets/imgs/rectangle.png"
          alt=""
        />
        <div class="info rl">
          <p
            class="name-wrap"
            :title="n.name"
          >
            <span class="name">{{n.name}}</span>
            <i
              v-if="isShowPublished && n.publishType"
              class="el-icon-success published-icon"
            />
          </p>
          <el-tooltip
            v-if="n.desc"
            popper-class="instr-tooltip tooltip-cover-arrow"
            effect="dark"
            :content="n.desc"
            placement="bottom-start"
          >
            <p class="desc">{{n.desc}}</p>
          </el-tooltip>
        </div>
        <div class="tags">
          <span :class="['smartDate']">{{n.createdAt}}</span>
          <div
            v-if="!isShowTool"
            class="favorite-wrap"
          >
          <el-tooltip class="item" effect="dark" :content="n.user.userName" placement="top-start">
            <span class="user-name">{{n.user ? n.user.userName.length>6 ? n.user.userName.substring(0,6)+'...' : n.user.userName  : ''}}</span>
          </el-tooltip>
            <img
              v-if="!n.isFavorite"
              class="favorite"
              src="@/assets/imgs/like.png"
              alt=""
              @click="handelMark($event, n, i)"
            />
            <img
              v-else
              class="favorite"
              src="@/assets/imgs/like_active.png"
              alt=""
              @click="handelMark($event, n, i)"
            />
          </div>
        </div>
        <div v-if="isShowPublished && n.publishType && type !== 'workflow'" class="publishType">
            <span v-if="n.publishType === 'private'" class="publishType-tag"><span class="el-icon-lock"></span> Private</span>
            <span v-else class="publishType-tag"><span class="el-icon-unlock"></span> Public</span>
        </div>
        <div
          class="editor"
          v-if="isShowTool"
        >
          <el-dropdown
            @command="handleClick($event, n)"
            placement="top"
          >
            <span class="el-dropdown-link">
              <i
                class="el-icon-more icon edit-icon"
                @click.stop
              />
            </span>
            <el-dropdown-menu slot="dropdown">
              <el-dropdown-item
                command="edit"
                v-if="isCanClick(n)"
              >
                {{$t('common.button.edit')}}
              </el-dropdown-item>
              <el-dropdown-item
                command="delete"
                v-if="n.appId !== 'example'"
              >
                {{$t('common.button.delete')}}
              </el-dropdown-item>
              <el-dropdown-item
                command="copy"
              >
                {{$t('common.button.copy')}}
              </el-dropdown-item>
              <el-dropdown-item
                command="publish"
                v-if="n.appType === 'workflow' && !n.publishType && n.appId !== 'example'"
              >
                {{$t('common.button.publish')}}
              </el-dropdown-item>
              <el-dropdown-item
                command="cancelPublish"
                v-if="n.publishType && n.appId !== 'example'"
              >
                {{$t('common.button.cancelPublish')}}
              </el-dropdown-item>
               <el-dropdown-item
                command="publishSet"
              >
                PublishConfiguration
              </el-dropdown-item>
              <el-dropdown-item
                command="export"
                v-if="n.appType === 'workflow'"
              >
                {{$t('common.button.export')}}
              </el-dropdown-item>
            </el-dropdown-menu>
          </el-dropdown>
        </div>
        <div class="copy-editor" v-if="n.appType === 'agentTemplate' && n.isShowCopy" @click.stop="copyTemplate(n)">
          <span class="el-icon-plus add"></span>
          <span>Copy</span>
        </div>
      </div>
    </div>
    <el-empty class="noData" v-if="!(listData && listData.length)" :description="$t('common.noData')"></el-empty>
    <el-dialog
      :title="$t('list.tips')"
      :visible.sync="dialogVisible"
      width="400px"
      append-to-body
      :close-on-click-modal="false"
      :before-close="handleClose"
      class="createTotalDialog"
    >
      <div style="margin-top: -20px">
        <div v-for="item in publishList" :key="item.key" style="margin-bottom: 5px">
          <el-radio :label="item.key" v-model="publishType">{{item.value}}</el-radio>
        </div>
        <div style="text-align: right; margin-top: 15px; margin-bottom: -10px">
          <el-button size="mini" type="primary" @click="doPublish">{{$t('common.button.confirm')}}</el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { AppType } from "@/utils/commonSet";
import { deleteApp, appCancelPublish, copyAgnetTemplate, appPublish,copyTextQues,copyAgentApp } from "@/api/appspace";
import { copyWorkFlow, publishWorkFlow, copyExample, exportWorkflow } from "@/api/workflow";
import { setFavorite } from "@/api/explore";
export default {
  props:{
    type: String,
    showCreate: Function,
    agnetType:String,
    appData:{
      type:Array,
      required:true,
      default:[]
    },
    agent_type:'agent_template',
    isShowTool: false,
    isShowPublished: false,
    appFrom:{
      type:String,
      default:''
    }
  },
  watch: {
    appData: {
      handler: function (val) {
        this.listData = val;
      },
      immediate: true,
      deep: true,
    },
  },
  data() {
    return {
      apptype: AppType,
      basePath: this.$basePath,
      listData: [],
      row: {},
      publishType: 'private',
      dialogVisible: false,
      publishList: [
        {key: 'private', value: this.$t('workflow.publishText')},
        {key: 'organization', value: this.$t('workflow.publicOrgText')},
        {key: 'public', value: this.$t('workflow.publicTotalText')}
      ]
    };
  },
  methods: {
    copyTemplate(n){
      copyAgnetTemplate({assistantTemplateId:n.assistantTemplateId}).then(res =>{
        if(res.code === 0){
          this.$message.success('Copy successful')
          const id = res.data.assistantId
          this.$router.push({path:`/agent/test?id=${id}`})
        }
      })
    },
    mouseEnter(n){
      if(n.appType === 'agentTemplate'){
        n.isShowCopy = true;
      }
    },
    mouseLeave(n){
      if(n.appType === 'agentTemplate'){
        n.isShowCopy = false;
      }
    },
    validateAgnet(){
      if(this.type === 'agent' && this.agnetType === 'template'){
        return false
      }else{
        return true
      }
    },
    handleClose() {
      this.dialogVisible = false
    },
    isCanClick(n) {
      return this.isShowTool ? ((n.appType === 'workflow' && !n.publishType && n.appId !== 'example') || n.appType !== 'workflow') : true
    },
    // Shared delete method
    async handleDelete() {
      const params = {
        appId: this.row.appId,
        appType:this.row.appType
      };
      const res = await deleteApp(params);
      if (res.code === 0) {
        this.$message.success(this.$t("list.delSuccess"));
        this.$emit("reloadData");
      }
    },
    workflowEdit(row) {
      const querys = {
        id: row.appId,
      };
      this.$router.push({ path: "/workflow", query: querys });
    },
    workflowDelete(row) {
      this.row = row;
      this.$alert(this.$t("list.deleteTips"), this.$t("list.tips"), {
        confirmButtonText: this.$t("list.confirm"),
        callback: (action) => {
          if (action === "confirm") {
            this.handleDelete();
          }
        },
      });
    },
    async workflowCopy(row) {
      const params = {
        workflow_id: row.appId,
      }

      const isExample = false // For now always call the workflow copy interface even though templates are not supported yet
      const exampleParams = {
        configName: row.name + '_' + this.$t('common.copy.copyText'),
        configENName: "",
        configDesc: row.desc,
        isStream: false
      }

      const res = isExample
        ? await copyExample({...params, ...exampleParams})
        : await copyWorkFlow(params);

      if (res.code === 0) {
        this.$router.push({
          path: "/workflow",
          query: { id: isExample ? res.data.workflowID : res.data.workflow_id },
        });
      }
    },

    workflowPublish(row) {
      this.row = row
      this.dialogVisible = true
      this.publishType = 'private'
    },
    async doPublish() {
      const params = {
        appId: this.row.appId,
        appType: this.row.appType,
        publishType: this.publishType
      }
      const res = await appPublish(params)
      if (res.code === 0) {
        this.$message.success(this.$t("list.publicSuccess"))
        this.handleClose()
        this.$emit('reloadData')
      }
    },
    async cancelPublish(row) {
      let confirmed = true;
      const params = {
        appId: row.appId,
        appType: row.appType
      };
      
      // Workflow unpublish requires a confirmation dialog
      if(row.appType === 'workflow'){
        confirmed = await this.showDeleteConfirm('After unpublishing, agents that historically referenced this workflow will automatically cancel the reference, and this operation cannot be undone');
      }
      
      if(confirmed){
        const res = await appCancelPublish(params)
        if (res.code === 0) {
          this.$message.success(this.$t("common.message.success"))
          this.$emit("reloadData")
        }
      }
    },
    workflowExport(row) {
      exportWorkflow({workflow_id: row.appId}).then(response => {
        const blob = new Blob([response], { type: response.type })
        const url = URL.createObjectURL(blob);
        const link = document.createElement("a")
        link.href = url
        link.download = row.name + '.json'
        link.click()
        window.URL.revokeObjectURL(link.href)
      })
    },
    workflowOperation(method, row) {
      switch (method) {
        case "edit":
          this.workflowEdit(row);
          break;
        case "delete":
          this.workflowDelete(row);
          break;
        case "copy":
          this.workflowCopy(row);
          break;
        case "publish":
          this.workflowPublish(row);
          break;
        case "cancelPublish":
          this.cancelPublish(row);
          break;
        case "publishSet":
          this.$router.push({path:`/workflow/publishSet`, query: {appId: row.appId, appType: row.appType, name: row.name}})
          break;
        case 'export':
          this.workflowExport(row)
      }
    },
    async showDeleteConfirm(tips){
      try{
        await this.$alert(tips, this.$t("list.tips"), {
          confirmButtonText: this.$t("list.confirm"),
        });
        return true;
      }catch(err){
        return false;
      }
    },
    intelligentEdit(row) {
      this.$router.push({
          path: "/agent/test",
          query: { 
            id: row.appId,
            ...(row.publishType !== '' && {publish:true})
          }
      });
    },
    intelligentDelete(row) {
      this.row = row;
      this.handleDelete();
    },
    intelligentCopy(row){
      copyAgentApp({assistantId:row.appId}).then(res=>{
        if(res.code === 0){
          const id = res.data.assistantId;
          this.$message.success('Copy successful');
          this.$router.push({path:`/agent/test?id=${id}`})
        }
      }).catch(()=>{})
    },
    intelligentOperation(method, row) {
      switch (method) {
        case "edit":
          // AgentEdit
          this.intelligentEdit(row);
          break;
        case "delete":
          // AgentDelete
          this.intelligentDelete(row);
          break;
        case "copy":
          // AgentCopy
          this.intelligentCopy(row);
          break;
        case "cancelPublish":
          this.cancelPublish(row);
          break;
        case "publishSet":
          //PublishSetting
          this.$router.push({path:`/agent/publishSet`, query: {appId: row.appId, appType: row.appType, name: row.name}})
          break;
      }
    },
    txtQuesEdit(row) {
      this.$router.push({
          path: "/rag/test",
          query: { 
            id: row.appId,
            ...(row.publishType !== '' && {publish:true})
          }
      });
    },
    txtQuesDelete(row) {
      this.row = row;
      this.handleDelete();
    },
    txtQuesCopy(row){
      copyTextQues({ragId:row.appId}).then(res=>{
        if(res.code === 0){
          const id = res.data.ragId;
          this.$message.success('Copy successful');
          this.$router.push({path:`/rag/test?id=${id}`})
        }
      }).catch(()=>{})
    },
    txtQuesOperation(method, row) {
      switch (method) {
        case "edit":
          // Text Q&AEdit
          this.txtQuesEdit(row);
          break;
        case "delete":
          // Text Q&ADelete
          this.txtQuesDelete(row);
          break;
        case "copy":
          // Text Q&ACopy
          this.txtQuesCopy(row);
          break;
        case "cancelPublish":
          this.cancelPublish(row);
          break;
        case "publishSet":
          this.$router.push({path:`/rag/publishSet`, query: {appId: row.appId, appType: row.appType, name: row.name}})
          break;
      }
    },
    commonToChat(row){
      const type = row.appType;
      switch (type) {
        case "agent":
          this.$router.push({path:'/explore/agent', query:{id:row.appId}});
          break;
        case "rag":
          this.$router.push({path:'/explore/rag', query:{id:row.appId}});
          break;
        case "workflow":
          this.$router.push({path:'/explore/workflow', query:{id:row.appId}});
          break;
      }
    },
    commonMethods(method, row) {
      const type = row.appType;
      switch (type) {
        case "agent":
          this.intelligentOperation(method, row);
          break;
        case "rag":
          this.txtQuesOperation(method, row);
          break;
        case "workflow":
          this.workflowOperation(method, row);
          break;
      }
    },
    handleClick(command, row) {
      this.commonMethods(command, row);
    },
    toEdit(row) {
      if(this.appFrom === 'explore'){
        this.commonToChat(row)
      }else if(row.appType === 'agentTemplate'){
        this.$router.push({path:`/agent/templateDetail?id=${row.assistantTemplateId}`})
      }
      else{
        this.commonMethods("edit", row);
      }
    },
    handelMark(e,n, i) {
      e.stopPropagation();
      this.$confirm(
        n.isFavorite
          ? this.$t("explore.unFavorite")
          : this.$t("explore.favorite"),
        this.$t("common.confirm.title"),
        {
          confirmButtonText: this.$t("common.confirm.confirm"),
          cancelButtonText: this.$t("common.confirm.cancel"),
          type: "warning",
        }
      )
        .then(() => {
          setFavorite({
            appId: n.appId,
            appType:n.appType,
            isFavorite: !n.isFavorite,
          }).then((res) => {
            if (res.code === 0) {
              this.$message.success(
                n.isFavorite
                  ? this.$t("explore.delSuccess")
                  : this.$t("explore.setSuccess")
              );
              const list = [...this.listData];
              list[i].isFavorite = !n.isFavorite;
              this.listData = [...list];
              // this.getHistoryList();
            }
          });
        })
        .catch(() => {});
    },
  },
};
</script>

<style lang="scss" scoped>
@import "@/style/appCard.scss";
.noData {
  padding: 30px 0;
}
</style>
