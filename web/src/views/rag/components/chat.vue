<template>
    <!-- 远景大Model -->
    <div class="full-content flex">
        <el-main class="scroll">
            <div class="smart-center">
                <!--基础Configuration回显-->
                <div v-show="echo" class="session rl echo">
                    <Prologue  :editForm="editForm" @setProloguePrompt="setProloguePrompt" :isBigModel="true" />
                </div>
                <!--conversation-->
                <div  v-show="!echo" class="center-session">
                    <SessionComponentSe
                            ref="session-com"
                            class="component"
                            :sessionStatus="sessionStatus"
                            @clearHistory="clearHistory"
                            @refresh="refresh"
                            @queryCopy="queryCopy"
                            :defaultUrl="editForm.avatar.path"
                    />
                </div>
                <!--Input Box-->
                <div class="center-editable">
                    <div v-show="stopBtShow" class="stop-box">
                        <span v-show="sessionStatus === 0" class="stop" @click="preStop">
                            <img class="stop-icon mdl" :src="require('@/assets/imgs/stop.png')"/>
                            <span class="mdl">StopGenerate</span>
                        </span>
                        <!-- <span v-show="sessionStatus !== 0" class="stop" @click="refresh">
                            <img class="stop-icon mdl" :src="require('@/assets/imgs/refresh.png')"/>
                            <span class="mdl">{{$t('yuanjing.refresh')}}</span>
                        </span> -->
                    </div>
                    <EditableDivV3
                            ref="editable"
                            source="perfectReminder"
                            :fileTypeArr="fileTypeArr"
                            :currentModel="currentModel"
                            :isModelDisable="isModelDisable"
                            :showModelSelect="false"
                            @preSend="preSend"
                            @modelChange="modelChange"
                            @getModelType="getModelType"
                            @setSessionStatus="setSessionStatus"
                    />
                </div>
            </div>
        </el-main>
    </div>
</template>

<script>
    import SessionComponentSe from './SessionComponentSe'
    import EditableDivV3 from './EditableDivV3'
    import {getConversationList, createConversation, getConversationDetail, deleteConversation,deleteConversationHistory} from '@/api/cubm'
    import Prologue from './Prologue'
    import sseMethod from '@/mixins/sseMethod'
    import {md} from '@/mixins/marksown-it'
    import {mapActions, mapGetters} from 'vuex'
    // import { getTemplateList } from '@/api/prompt'

    export default {
        props:{
            chatType:{
                type: String,
                default:''
            },
            editForm:{
                type:Object,
                default:null
            }
        },
        components: {
            SessionComponentSe,
            EditableDivV3,
            Prologue
        },
        mixins: [sseMethod],
        computed: {
            ...mapGetters('app', ['sessionStatus']),
            ...mapGetters('menu', ['basicInfo']),
            ...mapGetters('user', ['commonInfo']),
        },
        data() {
            return {
                amswerNum:0,
                isModelDisable:false,
                currentModel:null,
                echo: true,
                basicForm: {
                    avatar: '123',//img/tr/deepseek-icon.png-UseInterfaceGet of Image
                    instructions: '456',//我Yes您 of 智能小助手，可以帮您思考文案，With您聊天，还可以答疑Result。For example您可以问我：
                    name: 'ffff',//你好，我YesDeepSeek-UseInterfaceGet of 文案
                    description: 'fsdfdggfh'
                },
                expandForm: {
                    starterPrompts: [
                        // {value: 'Such as何打粉底不卡粉?'},
                        // {value: '我想问怎么化妆皮肤不干?'},
                        // {value: '写一故宫一日游攻略'}
                    ]
                },
                // fileTypeArr: ['image/*','doc/*'],
                fileTypeArr: ['doc/*'],
                hasDrawer: false,
                drawer: true,
                sseApi: this.$basePath + '/use/model/api/v1/chatllm/stream',
            }
        },
        watch: {
            $route: {
                handler(val, oldval) {

                },
                deep: true
            }
        },
        created() {
            // this.getReminderList(() => {
            //     this.hasDrawer = true
            // })
            // this.getConversationList()
        },
        methods: {
            getModelType(type){
                const dataInfo = this.commonInfo.data.useModel;
                if(type ==='deepseek'){
                   this.expandForm.starterPrompts = dataInfo.useModels[1]['welcomeQuestions']
                   this.basicForm.instructions = dataInfo.useModels[1]['welcomeDesc'];
                   this.basicForm.name = dataInfo.useModels[1]['welcomeText'];
                   this.basicForm.avatar = this.$basePath + '/use/model/api' + dataInfo.useModels[1]['welcomeLogoPath'];
                }else{
                    this.expandForm.starterPrompts = dataInfo.useModels[0]['welcomeQuestions']
                   this.basicForm.instructions = dataInfo.useModels[0]['welcomeDesc'];
                   this.basicForm.name = dataInfo.useModels[0]['welcomeText'];
                   this.basicForm.avatar = this.$basePath + '/use/model/api' + dataInfo.useModels[0]['welcomeLogoPath'];
                }
            },
            //conversationList
            async getConversationList(noInit) {
                let res = await getConversationList({assistantId: this.assistantId, pageSize: 1000, pageNo: 1})
                if (res.code === 0) {
                    if(res.data.list && res.data.list.length > 0){
                        this.chatList = res.data.list.map(n => {
                            return {...n, hover: false, active: false}
                         })
                        if (noInit) {
                            this.chatList[0].active = true  //noInit Yestrue when ，左侧Default选infirst,butYes不to调InterfaceRefreshDetails
                        } else {
                            this.conversionClick[this.chatList[0]]
                        }
                    }else{
                        this.chatList = []
                    }
                }
            },
            //New Conversation
            preCreateConversation() {
                if (this.echo) {
                    this.$message({
                        type: 'info',
                        message: this.$t('yuanjing.changeDialogMsg'),
                        customClass: 'dark-message',
                        iconClass: 'none',
                        duration: 1500
                    })
                    return
                }
                this.isModelDisable = false
                this.conversationId = ''
                this.currentModel = null
                this.echo = true
                this.clearPageHistory()
                this.chatList.forEach(m => {
                    m.active = false
                })
            },
            //toggleconversation
            async conversionClick(n) {
                this.isModelDisable = true;
                if (this.sessionStatus === 0) {
                    //this.$message.warning('上问题not答完')
                    return
                }else{
                    this.stopBtShow = false
                }

                this.chatList.forEach(m => {
                    m.active = false
                })
                this.amswerNum = 0
                n.active = true
                this.clearPageHistory()
                this.echo = false

                this.conversationId = n.conversationId
                this.getConversationDetail(this.conversationId,true)
            },
            async getConversationDetail(id,loading){
                loading && this.$refs['session-com'].doLoading()
                let res = await getConversationDetail({conversationId: id, pageSize: 1000, pageNo: 1})
                if (res.code === 0) {
                    let history = res.data.list ? res.data.list.map(n => {
                        return {
                            ...n,
                            query: n.prompt,
                            //response:n.qa_type===4?(marked(n.response)).replaceAll('\\n','<br/>'):n.response.replaceAll('\n-','<br/>•').replaceAll('\n','<br/>'),
                            response:[0,1,2,3,4,5,6,20,21,10].includes(n.qa_type)?md.render(n.response):n.response.replaceAll('\n-','\n•'),
                            oriResponse:n.response,
                            searchList: n.searchList ? JSON.parse(n.searchList) : [],
                            filepath: n.responseFileUrls,
                            "gen_file_url_list":n.responseFileUrls || [],
                            "isOpen":true
                        }
                    }) : []

                    //toggleHistory Record，select对应Model
                    if(res.data.list && res.data.list !== null){
                        this.currentModel = {
                        modelId:res.data.list[0]['modelId'],
                        modelVersion:res.data.list[0]['modelVersion']
                        }
                    }else{
                        this.currentModel = null
                    }
                    this.$refs['session-com'].replaceHistory(history)
                    this.$nextTick(()=>{
                        this.addCopyClick()
                    })
                }
            },
            //Delete Conversation
            async preDelConversation(n) {
                //todo 给所Has of clickEvent统一Add拦截
                if (this.sessionStatus === 0) {
                    //this.$message.warning('上问题not答完')
                    return
                }
                let res = await deleteConversation({conversationId: n.conversationId})
                if (res.code === 0) {
                    this.getConversationList()
                    if(this.conversationId === n.conversationId){
                        this.conversationId = ''
                        this.$refs['session-com'].clearData()
                    }
                    this.echo = true
                }
            },
            /*------session------*/
            async preSend(val,fileId,fileInfo) {
                this.inputVal = val || this.$refs['editable'].getPrompt()
                if (!this.inputVal) {
                    this.$message.warning('Please enterContent');
                    return
                }
                if (!this.verifiyFormParams()) {
                    return;
                }
                // this.setParams()
                this.setSseParams({ragId: this.editForm.appId, question: this.inputVal})
                this.doragSend()
                this.echo = false
            },
            verifiyFormParams(){
                if (this.chatType === 'chat') return true;
                const { matchType, priorityMatch, rerankModelId } = this.editForm.knowledgeConfig;
                const isMixPriorityMatch = matchType === 'mix' && priorityMatch;
                const conditions = [
                    { check: !this.editForm.modelParams, message: 'Please selectModel' },
                    { check: !isMixPriorityMatch && !rerankModelId, message: 'Please selectrerankModel' },
                    { check: this.editForm.knowledgebases.length === 0, message: 'Please selectKnowledge Base' }
                ];
                for (const condition of conditions) {
                    if (condition.check) {
                    this.$message.warning(condition.message);
                    return false;
                    }
                }
                return true;
            },
            modelChange(){//toggleModelNew Conversation
                this.preCreateConversation()
            },
            setParams() {
                ++this.amswerNum;
                if(this.amswerNum > 0){
                    this.isModelDisable = true
                }
                let fileId = this.getFileIdList() || this.fileId;
                this.useSearch = this.$refs['editable'].sendUseSearch()
                this.modelParams = this.$refs['editable'].getModelInfo()
                this.isBigModel = true;
                this.setSseParams({conversationId: this.conversationId, fileId})
                this.doSend()
                this.echo = false
            },
            /*--右侧Tip词--*/
            showDrawer() {
                this.drawer = true
            },
            hideDrawer() {
                this.drawer = false
            },
            async getReminderList(cb) {
                let res = await getTemplateList({pageNo:0,pageSize:0,title:''})
                if (res.code === 0) {
                    this.reminderList = res.data.list||[]
                    cb && cb()
                    console.log(new Date().getTime())
                }
            },
            reminderClick(n) {
                this.$refs['editable'].setPrompt(n.prompt)
            },
            async doDeleteHistory(){
                let res = await deleteConversationHistory({conversationId:this.conversationId})
                if(res.code === 0){
                    this.$message.success(this.$t('yuanjing.deleteTips'))
                }
            },
        }
    }
</script>

<style lang="scss" scoped>
@import '@/style/chat.scss';
</style>
