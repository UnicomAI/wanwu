<template>
    <!-- 远景大Model -->
    <div class="full-content flex">
        <el-main class="scroll">
            <div class="smart-center" style="padding:0;">
                <!--基础Configuration回显-->
                <div v-show="echo" class="session rl echo">
                    <Prologue  :editForm="editForm" @setProloguePrompt="setProloguePrompt" :isBigModel="true" />
                </div>
                <!--conversation-->
                <div v-show="!echo" class="center-session">
                    <SessionComponentSe
                            ref="session-com"
                            class="component"
                            :sessionStatus="sessionStatus"
                            @clearHistory="clearHistory"
                            @refresh="refresh"
                            :type="type"
                            @queryCopy="queryCopy"
                            :defaultUrl="editForm.avatar.path"
                    />
                </div>
                <!--Input Box-->
                <div class="center-editable">
                    <div v-show="stopBtShow" class="stop-box">
                        <span v-show="sessionStatus === 0" class="stop" @click="preStop">
                            <img class="stop-icon mdl" :src="require('@/assets/imgs/stop.png')"/>
                            <span class="mdl">{{$t('agent.stop')}}</span>
                        </span>
                        <span v-show="sessionStatus !== 0" class="stop" @click="refresh">
                            <img class="stop-icon mdl" :src="require('@/assets/imgs/refresh.png')"/>
                            <span class="mdl">{{$t('agent.refresh')}}</span>
                        </span>
                    </div>
                    <EditableDivV3
                            ref="editable"
                            source="perfectReminder"
                            :fileTypeArr="fileTypeArr"
                            :currentModel="currentModel"
                            :isModelDisable="isModelDisable"
                            :showModelSelect="false"
                            :type="type"
                            @preSend="preSend"
                            @modelChange="modelChange"
                            @setSessionStatus="setSessionStatus"
                    />
                    <div v-if="appUrlInfo" class="appUrlInfo">
                        <span v-if='appUrlInfo.copyrightEnable'>Copyright所have: {{appUrlInfo.copyright}}</span>
                        <span v-if='appUrlInfo.privacyPolicyEnable'>Privacy Policy: <a :href="appUrlInfo.privacyPolicy" target="_blank" style="color:var(--color);">{{appUrlInfo.privacyPolicy}}</a></span>
                        <span v-if="appUrlInfo.disclaimerEnable">Disclaimer: {{appUrlInfo.disclaimer}}</span>
                    </div>
                </div>
            </div>
        </el-main>
    </div>
</template>

<script>
    import SessionComponentSe from './SessionComponentSe'
    import EditableDivV3 from './EditableDivV3'
    import {delConversation,createConversation,getConversationHistory,delOpenurlConversation,openurlConversation,OpenurlConverHistory} from "@/api/agent";
    import Prologue from './Prologue'
    import sseMethod from '@/mixins/sseMethod'
    import {md} from '@/mixins/marksown-it'
    import {mapGetters} from 'vuex'

    export default {
        inject:{
            getHeaderConfig:{
                default:() => null
            }
        },
        props:{
            editForm:{
                type:Object,
                default:null
            },
            chatType:{
                type:String,
                default:''
            },
            type:{
                type:String,
                default:'agentChat'
            },
            appUrlInfo:{
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
        watch: {
            // 'editForm.visionsupport': {
            //     handler(newVal) {
            //         if ( newVal === 'support') {
            //             this.fileTypeArr = ['doc/*','image/*'];
            //         } else {
            //             this.fileTypeArr = ['doc/*'];
            //         }
            //     },
            //     immediate: true
            // }
        },
        data() {
            return {
                amswerNum:0,
                isModelDisable:false,
                currentModel:null,
                echo: true,
                fileTypeArr: ['doc/*','image/*'],
                hasDrawer: false,
                drawer: true,
            }
        },

        methods: {
            createConversion(){
                if (this.echo) {
                    this.$message({
                        type: 'info',
                        message: 'alreadytoggle最新session',
                        customClass: 'dark-message',
                        iconClass: 'none',
                        duration: 1500
                    })
                    return
                }
                this.conversationId = ''
                this.echo = true
                this.clearPageHistory()
                this.$emit('setHistoryStatus')
            },
            //toggleconversation
            conversionClick(n) {
                // this.isModelDisable = true;
                if (this.sessionStatus === 0) {
                    //this.$message.warning('上问题not答完')
                    return
                }else{
                    this.stopBtShow = false
                }

                this.$emit('setHistoryStatus')
                this.amswerNum = 0
                n.active = true
                this.clearPageHistory()
                this.echo = false
                this.conversationId = n.conversationId
                this.getConversationDetail(this.conversationId,true)
            },
            async getConversationDetail(id,loading){
                loading && this.$refs['session-com'].doLoading()
                let res = null;
                if(this.type === "agentChat"){
                    res = await getConversationHistory({conversationId: id, pageSize: 1000, pageNo: 1});
                }else{
                    const config = this.getHeaderConfig();
                    res = await OpenurlConverHistory({conversationId: id},this.editForm.assistantId,config);
                }
                
                if (res.code === 0) {
                    let history = res.data.list ? res.data.list.map(n => {
                        return {
                            ...n,
                            query: n.prompt,
                            response:[0,1,2,3,4,5,6,20,21,10].includes(n.qa_type)?md.render(n.response):n.response.replaceAll('\n-','\n•'),
                            oriResponse:n.response,
                            searchList: n.searchList ? n.searchList : [],
                            fileList: n.requestFiles,
                            "gen_file_url_list":n.responseFileUrls || [],
                            "isOpen":true,
                            toolText:'alreadyuseTool',
                            thinkText:'already深度思考',
                            showScrollBtn:null
                        }
                    }) : []
                    this.$refs['session-com'].replaceHistory(history)
                    this.$nextTick(()=>{
                        this.addCopyClick()
                    })
                }
            },
            //Delete Conversation
            async preDelConversation(n) {
                if (this.sessionStatus === 0) {
                    return
                }
                let res = null;
                if (this.type === "agentChat") {
                    res = await delConversation({conversationId: n.conversationId})
                }else{
                    const config = this.getHeaderConfig();
                    res = await delOpenurlConversation({conversationId:n.conversationId},this.editForm.assistantId,config)
                }
                
                if (res.code === 0) {
                    this.$emit('reloadList')
                    if(this.conversationId === n.conversationId){
                        this.conversationId = ''
                        this.$refs['session-com'].clearData()
                    }
                    this.echo = true
                }
            },
            /*------session------*/
            async preSend(val,fileList,fileInfo) {
                this.inputVal = val || this.$refs['editable'].getPrompt()
                this.fileId = fileInfo || [];
                this.isTestChat = this.chatType === 'test' ? true :false;
                this.fileList = fileList || this.$refs['editable'].getFileList();
                if (!this.inputVal) {
                    this.$message.warning('Please enterContent');
                    return
                }
                if (!this.verifiyFormParams()) {
                    return;
                }
                //IfYes新session，FirstCreate
                if (!this.conversationId && this.chatType === 'chat') {
                    let res = null;
                    if (this.type === "agentChat") {
                        res = await createConversation({prompt: this.inputVal,assistantId:this.editForm.assistantId})
                    }else{
                        const config = this.getHeaderConfig();
                        res = await openurlConversation({prompt: this.inputVal},this.editForm.assistantId,config)
                    }
                    
                    if (res.code === 0) {
                        this.conversationId = res.data.conversationId
                        this.$emit('reloadList',true)
                        this.setParams()
                    }
                } else {
                    this.setParams()
                }
            },
            verifiyFormParams(){
                if (this.chatType === 'chat') return true;
                const conditions = [
                    { check: !this.editForm.modelParams, message: 'Please selectModel' },
                    { check: !this.editForm.prologue, message: 'Please enter开场白' }
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
                let fileId = this.$refs['editable'].getFileIdList() || this.fileId;
                this.useSearch = this.$refs['editable'].sendUseSearch();
                this.setSseParams({conversationId: this.conversationId, fileInfo:fileId,assistantId:this.editForm.assistantId})
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
            }
        }
    }
</script>

<style lang="scss" scoped>
@import '@/style/chat.scss';
.appUrlInfo{
    margin-top:10px;
    display:flex;
    justify-content:center;
    span{
        cursor: pointer;
        color:#bbb;
        margin-right:15px;
    }
}
</style>
