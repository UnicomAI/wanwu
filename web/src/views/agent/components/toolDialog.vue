<template>
    <div>
        <el-dialog
            title="AddTool"
            :visible.sync="dialogVisible"
            width="50%"
            :before-close="handleClose">
            <div class="tool-typ">
                <div class="toolbtn">
                    <div v-for="(item,index) in toolList" :key="index" @click="clickTool(item,index)" :class="[{'active':activeValue === item.value}]">
                        {{item.name}}
                    </div>
                </div>
                <el-input v-model="toolName" placeholder="SearchTool" class="tool-input" suffix-icon="el-icon-search" @keyup.enter.native="searchTool" clearable></el-input>
            </div>
            <div class="toolContent">
                <div @click="goCreate" class="createTool"><span class="el-icon-plus add"></span>{{createText()}}</div>
                <template v-for="(items, type) in contentMap">
                    <div 
                    v-if="activeValue === type"
                    v-for="(item,i) in items"
                    :key="item[type + 'Id'] || item.id"
                    class="toolContent_item"
                    >
                        <template v-if="type === 'workflow'">
                                <div class="tool_box">
                                   <div class="tool_img">
                                        <img :src="'/user/api/' + item.avatar.path" v-show="item.avatar && item.avatar.path" />
                                    </div>
                                    <span>{{item.name}}</span>
                                </div>
                                <div>
                                    <el-button type="text" @click="openTool($event,item,type)" v-if="!item.checked">Add</el-button>
                                    <el-button type="text" v-else style="color:#ccc;">alreadyAdd</el-button>
                                </div>
                        </template>
                        <el-collapse  @change="handleToolChange" v-else class="tool_collapse">
                            <el-collapse-item  :name="item.toolId">
                                <template slot="title">
                                   <div class="tool_img">
                                        <img :src="'/user/api/' + item.avatar.path" v-show="item.avatar && item.avatar.path" />
                                    </div>
                                   <h3>{{item.toolName}}</h3>
                                   <span v-if="item.loading" class="el-icon-loading loading-text"></span>
                                </template>
                                <template v-if="item.children && item.children.length">
                                    <div v-for="(tool,index) in item.children" class="tool-action-item" :key="'tool'+index">
                                        <div style="padding-right:5px;">
                                            <p>
                                             <span>{{tool.name}}</span>
                                             <el-tooltip class="item" effect="dark" :content="tool.description" placement="top-start" v-if="tool.description && tool.description.length > 0">
                                                <span class="el-icon-info desc-info"></span>
                                             </el-tooltip>
                                            </p>
                                        </div>
                                        <div>
                                        <el-button type="text" @click="openTool($event,item,type,tool)" v-if="!tool.checked">Add</el-button>
                                        <el-button type="text" v-else style="color:#ccc;">alreadyAdd</el-button>
                                        </div>
                                    </div>
                                </template>
                            </el-collapse-item>
                        </el-collapse>
                    </div>
                </template>
            </div>
        </el-dialog>
    </div>
</template>
<script>
import { getList } from '@/api/workflow.js';
import { addWorkFlowInfo, addMcp,addCustomBuiltIn,toolList,toolActionList,mcptoolList,mcpActionList } from "@/api/agent";
import { getExplorationFlowList} from "@/api/workflow";
export default {
    props:['assistantId'],
    data(){
        return {
            toolName:'',
            dialogVisible:false,
            toolIndex:0,
            activeValue:'tool',
            workFlowInfos:[],
            mcpInfos:[],
            customInfos:[],
            mcpList:[],
            workFlowList:[],
            customList:[],
            builtInInfos:[],
            toolList:[
                {
                    value:'tool',
                    name:'Tool'
                },
                {
                    value:'mcp',
                    name:'MCP'
                },
                {
                    value:'workflow',
                    name:'Workflow'
                }
            ]
        }
    },
    computed:{
         contentMap() {
            return {
            tool: this.customInfos,
            builtIn: this.builtInInfos,
            mcp: this.mcpInfos,
            workflow: this.workFlowInfos
            }
        }
    },
    created(){
        this.getMcpSelect('');
        this.getWorkflowList('');
        this.getCustomList('')
    },
    methods:{
        handleToolChange(id){
            let toolId = id[0];
            if(this.activeValue === 'tool'){
                const targetItem = this.customInfos.find(item => item.toolId === toolId)
                if(targetItem) {
                    const { toolId, toolType } = targetItem
                    const index = this.customInfos.findIndex(item => item.toolId === toolId)
                    this.getToolAction(toolId, toolType, index)
                }
            }else if(this.activeValue === 'mcp'){
                const targetItem = this.mcpInfos.find(item => item.toolId === toolId)
                if(targetItem) {
                    const { toolId, toolType } = targetItem
                    const index = this.mcpInfos.findIndex(item => item.toolId === toolId)
                    this.getMcpAction(toolId, toolType, index)
                }
            }
           
        },
        getCustomList(name){//GetCustomAnd内置Tool
            toolList({name}).then(res =>{
                if(res.code === 0){
                    this.customInfos  = (res.data.list || []).map(m => ({
                        ...m,
                        loading:false,
                        children:[]
                    }))
                }
            }).catch(() =>{

            })
        },
        getToolAction(toolId,toolType,index){
            this.$set(this.customInfos[index], 'loading',true)
            toolActionList({toolId,toolType}).then(res =>{
                if(res.code === 0){
                    this.$set(this.customInfos[index], 'children', res.data.actions)
                    this.$set(this.customInfos[index], 'loading',false)
                    this.customInfos[index]['children'].forEach(m => {
                        m.checked = this.customList.some(item => item.actionName === m.name && item.toolId === toolId)
                    })
                    
                }
            }).catch(() =>{
                this.$set(this.customInfos[index], 'loading',false)
            })
        },
        goCreate(){
            if(this.activeValue === 'tool'){
                this.$router.push({path:'/tool?type=tool&mcp=custom'})
            }else if(this.activeValue === 'mcp'){
                this.$router.push({path:'/tool?type=mcp&mcp=integrate'})
            }else{
                this.$router.push({path:'/appSpace/workflow'})
            }
        },
        createText(){
            if(this.activeValue === 'tool'){
                return 'CreateCustomTool'
            }else if(this.activeValue === 'mcp'){
                return 'ImportMCP'
            }else{
                return 'CreateWorkflow'
            }
        },
        openTool(e,item,type,action){
            if(!e) return;
            if(type === 'workflow'){
                this.addWorkFlow(item)
            }else if(type === 'mcp'){
                this.addMcpItem(item,action)
            }else{
                if(item.needApiKeyInput && !item.apiKey.length){
                    this.$message.warning('该内置Tool暂notbindAPI Key，会导致CallFailed!')
                }
                this.addCustomBuiltIn(item,action)
            }
        },
        addCustomBuiltIn(n,action){
            //AddCustomToolAnd内置Tool
            addCustomBuiltIn({assistantId:this.assistantId,actionName:action.name,toolId:n.toolId,toolType:n.toolType}).then(res =>{
                if(res.code === 0){
                    this.$set(action, 'checked', true);
                    this.$forceUpdate();
                    this.$message.success('ToolAddSuccess');
                    this.$emit('updateDetail');
                }
            }).catch(() =>{

            })
        },
        addMcpItem(n,action){
            addMcp({assistantId:this.assistantId,actionName:action.name,mcpId:n.toolId,mcpType:n.toolType}).then(res =>{
                if(res.code === 0){
                    this.$set(action, 'checked', true);
                    this.$forceUpdate();
                    this.$message.success('ToolAddSuccess');
                    this.$emit('updateDetail');
                }
            }).catch(() =>{

            })
        },
        addWorkFlow(n){
            this.doCreateWorkFlow(n,n.appId)
        },
        async doCreateWorkFlow(n,workFlowId, schema){
            let params = {
                assistantId: this.assistantId,
                workFlowId
            };
            let res = await addWorkFlowInfo(params);
            if (res.code === 0) {
                n.checked = true;
                this.$message.success(this.$t('agent.addPluginTips'));
                this.$emit('updateDetail');
            }
        },
        searchTool(){
            if(this.activeValue === 'tool'){
                this.getCustomList(this.toolName)
            }else if(this.activeValue === 'mcp'){
                this.getMcpSelect(this.toolName)
            }else{
                this.getWorkflowList(this.toolName)
            }
        },
        getMcpSelect(name){//GetmcpTool
            mcptoolList({name}).then(res => {
                if(res.code === 0){
                    this.mcpInfos = (res.data.list || []).map(m => ({
                        ...m,
                        children: [],
                        loading:false
                    }));
                }
               
            }).catch(err => {
            })
            },
        getMcpAction(toolId,toolType,index){
            this.$set(this.mcpInfos[index], 'loading', true)
            mcpActionList({toolId,toolType}).then(res => {
                if(res.code === 0){
                    this.$set(this.mcpInfos[index], 'children', res.data.actions)
                    this.$set(this.mcpInfos[index], 'loading', false)
                    this.mcpInfos[index]['children'].forEach(m => {
                        m.checked = this.mcpList.some(item => item.actionName === m.name && item.mcpId === toolId)
                    })
                    
                }
            }).catch(() =>{
                this.$set(this.mcpInfos[index], 'loading', false)
            })
        },
        getWorkflowList(name) {
                getExplorationFlowList({name,appType:'workflow',searchType:'all'}).then(res =>{
                    if (res.code === 0) {
                        this.workFlowInfos = (res.data.list || []).map(m => ({
                            ...m,
                            checked: this.workFlowList.some(item => item.workFlowId === m.appId)
                        }));
                    }
                })
        },
        showDialog(row){
            this.dialogVisible = true;
            //this.setMcp(row.mcpInfos);
            this.setWorkflow(row.workFlowInfos);
            //this.setCustom(row.customInfos)
            this.mcpList = row.mcpInfos || [];
            this.workFlowList = row.workFlowInfos || [];
            this.customList  = row.customInfos || [];
        },

        setWorkflow(data){
            this.workFlowInfos = this.workFlowInfos.map(m => ({
            ...m,
            checked: data.some(item => item.workFlowId === m.appId)
            }));
        },

        handleClose(){
            this.toolIndex = -1;
            this.activeValue = 'tool';
            this.dialogVisible = false;
        },
        clickTool(item,i){
            this.toolIndex = i;
            this.activeValue = item.value;
            if(this.activeValue === 'tool'){
                this.getCustomList('')
            }else if(this.activeValue === 'mcp'){
                this.getMcpSelect('')
            }else{
                this.getWorkflowList('')
            }
        }
    }
}
</script>
<style lang="scss" scoped>
/deep/{
    .el-dialog__body{
        padding:10px 20px;
    }
    .tool_collapse{
        width:100% !important;
        border:none !important;
    }
    .el-collapse-item__header{
        background:none!important;
        border-bottom:none!important;
    }
    .el-collapse-item__wrap{
         border-bottom:none!important;
         background:none!important;
    }
    .el-collapse-item__content{
        padding-bottom:0!important;
    }
   
}
.createTool{
    padding:10px;
    cursor: pointer;
    .add{
        padding-right:5px;
    }
}
.createTool:hover{
    color: $color;
}
.tool-typ{
    display:flex;
    justify-content:space-between;
    padding:10px 0;
    border-bottom: 1px solid #dbdbdb;
    .toolbtn{
        display:flex;
        justify-content:flex-start;
        gap:20px;
        div{
            text-align: center;
            padding:5px 20px;
            border-radius:6px;
            border:1px solid #ddd;
            cursor: pointer;
        }
    }
    .tool-input{
        width:200px;
    }
}
.toolContent{
    padding:10px 0;
    max-height:400px;
    overflow-y:auto;
    .toolContent_item{
        padding:5px 20px;
        border:1px solid #dbdbdb;
        border-radius:6px;
        margin-bottom:10px;
        cursor: pointer;
        display: flex;
        align-items:center;
        justify-content:space-between;
        .tool_box{
            display:flex;
            align-items:center;
        }
        .tool_img{
            width:35px;
            height:35px;
            background:#eee;
            border-radius:50%;
            display:inline-block;
            margin-right:5px;
            img{
                width:100%;
                height:100%;
                border-radius:50%;
                object-fit: cover;
            }
        }
        .loading-text{
            margin-left:4px;
            color:var($color)
        }
        .tool-action-item{
            display: flex;
            align-items:center;
            justify-content:space-between;
            border-top:1px solid #eee;
            padding:5px 0;
            .desc-info{
                color:#ccc;
                margin-left:4px;
                cursor:pointer;
            }
        }
    }
    .toolContent_item:hover{
        background:$color_opacity;
    }
}
.active{
    border:1px solid $color !important;
    color: #fff;
    background:$color;
}
</style>