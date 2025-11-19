import {fetchEventSource} from "../sse/index.js";
import { store } from '@/store/index'
import Print from '../utils/printPlus2.js'
import {parseSub, convertLatexSyntax} from "@/utils/util.js"
import {mapActions, mapGetters} from 'vuex'
import {i18n} from "@/lang"

var originalFetch = window.fetch;

import {md} from './marksown-it'
import $ from './jquery.min.js'
import { file } from "jszip";
import {OPENURL_API, USER_API} from "@/utils/requestConstants"

export default {
    data() {
        return {
            isTestChat:false,
            defaultUrl: '/img/smart/logo.png',
            inputVal: '',
            eventSource: null,
            ctrlAbort: null,
            sseParams: {},
            sseResponse:{},
            echo: true,
            conversationId: '', //会话id
            chatList: [],
            reminderList: [],
            queryFilePath: '',
            stopBtShow: false,
            origin:window.location.origin,
            reconnectCount:0,
            isEnd: true,
            sseApi: `${USER_API}/assistant/stream`,
            rag_sseApi:`${USER_API}/rag/chat`,
            token:store.getters['user/token'],
            lastIndex:0,
            query:'',
            isStoped : false,
            access_token:'',
            runResponse: "",
            fileList: [],  // FileList
        };
    },
    created() {
        const vuex = JSON.parse(localStorage.getItem("access_cert"));
        if(vuex){
            this.access_token = vuex.user.token;
        }

    },
    mounted(){
        //this.addVisibilitychangeEvent()
    },
    beforeDestroy() {
        this.setStoreSessionStatus(-1)
        this.stopEventSource()
        this._print && this._print.stop()
    },
    computed: {
        ...mapGetters('app', ['sessionStatus']),
    },
    methods: {
        ...mapActions("app", ["setStoreSessionStatus"]),
        newFetch(url, options){
                // 可以Call原始 of  fetch Function
                if(this.isStoped){
                    return 
                }
                return originalFetch(url, options).then(response => {
                    // 可以在这里ModifyResponseOR者Add额外 of Process
                    let query = this.query

                    if(response.status != 200){
                        let me = this
                        try{
                            const stream = response.body
                           
                            const reader = stream.getReader();
                            const decoder = new TextDecoder('utf-8');

                            function readStream() {
                                reader.read().then(({ done, value }) => {
                                if (done) {
                                    console.log('Stream complete');
                                    reader.releaseLock()
                                    return;
                                }
                        
                                // Decode and process each chunk of data.
                                const decodedValue = decoder.decode(value, { stream: true });
                                                               
                                if(decodedValue){
                                    let msg = JSON.parse(decodedValue).msg
                                    me.setStoreSessionStatus(-1)
                                    var fillData = {
                                        "query": query,
                                        qa_type:0,
                                        finish:1,
                                        response:msg,  //Non代码TextUseCustomConvert规Then，不Usemarkdown,(markdownRender会导致卡顿ORStyle丢失)
                                        oriResponse:"",
                                        searchList:[]  // 滤ContainsyunyingshangFile of 出处
                                    }
                                    this.runResponse = msg
                                }                                
                                readStream();
                        
                                // Continue reading the stream.
                                }).catch(err => {
                                    console.error('Reading stream failed1:', err);
                                });
                            }
                        
                        readStream();
                        me.isStoped = true
                        }catch(e){
                            console.error('Reading stream failed:', e);
                        }

                    }

                    return response;
                }).catch((err)=>{
                    this.$message.warning("Connection failed, please try again later")
                    this.isEnd = true
                    this.setStoreSessionStatus(-1)
                    this.runDisabled = false
                });
        },
        ...mapActions("app", ["setStoreSessionStatus"]),
        queryCopy(text){
            this.setPrompt(text)
        },
        /* 滤掉markdown中Custom of 行号*/
        getContentInBraces(shtml) {
            let temp = document.createElement('div')
            temp.setAttribute('id','temp')
            temp.innerHTML = shtml
            document.body.appendChild(temp)
            $(temp).find('.line-num').remove()
            return temp.innerText
        },
        addCopyClick(){
            let copys =document.getElementsByClassName('copy-btn')
            for(var i=0;i<copys.length;i++){
                copys[i].addEventListener('click', (e)=> {
                    let innerText = e.target.parentNode.nextElementSibling.innerText
                    this.$copy(innerText)
                    e.target.innerText ='Copy successful'
                    //document.getElementById('temp') && document.getElementById('temp').remove() //Copy完成把临 when Generate of 元素Delete
                    setTimeout(()=>{
                        e.target.innerText ='Copy'
                    },1500)
                })
            }
        },
        // 填充开场白
        setProloguePrompt(val) {
            this.$refs['editable'].setPrompt(val)
            this.preSend()
        },
        //GetUpload of File
        getFileIdList() {
            let list = this.$refs['editable'].getFileIdList()  //[{fileId:"b67e79615"url:"b4d0a8.jpg"}]
            let fileIds = []
            this.queryFilePath = ''
            if (list.length) {
                fileIds = list.map(n => {
                    return n.fileId
                })
                this.queryFilePath = list[0].url
            }
            return fileIds.join(',')
        },
        mouseEnter(n) {
            n.hover = true
        },
        mouseLeave(n) {
            n.hover = false
        },
        setSessionStatus(status) {
            this.setStoreSessionStatus(status)
        },
        setSseParams(data) {
            this.sseParams = data
        },
        doragSend(){
            this.stopBtShow = true
            this.isStoped= false
            let _history = this.$refs['session-com'].getList()
            this.sendEventStream(this.inputVal,'', _history.length)
        },
        sendEventStream(prompt, msgStr, lastIndex){
            if (this.sessionStatus === 0) {
                this.$message.warning('The previous question has not been answered yet!')
                return
            }

            this.sseResponse = {}
            this.setStoreSessionStatus(0)
            this.clearInput()
            let params = {
                query: prompt, 
                pending: true, 
                responseLoading: true, 
                requestFileUrls:[],
                fileList:this.fileList,
                pendingResponse:''
            }
            this.$refs['session-com'].pushHistory(params)
            let endStr = ''
            this._print = new Print({
                onPrintEnd: () => {
                    // this.setStoreSessionStatus(-1)
                }
            })

            let history_list = this.$refs['session-com'].getSessionData();
            const history = history_list['history'].length > 1 ? history_list['history'][history_list['history'].length - 2]['history'] : [];
            this.ctrlAbort = new AbortController();
            const userInfo = this.$store.state.user.userInfo || {};
            this.eventSource = new fetchEventSource(this.origin + this.rag_sseApi, {
                method: 'POST',
                headers: {
                    "Content-Type": 'application/json',
                    'Authorization': 'Bearer ' + this.token,
                    "x-user-id": userInfo.uid,
                    "x-org-id": userInfo.orgId
                },
                signal: this.ctrlAbort.signal,
                body: JSON.stringify({...this.sseParams,'history':history}),
                openWhenHidden: true, //Page退to后台保持Join
                onopen: async(e) => {
                    //console.log("SSE connection established~",new Date().getTime());
                    if (e.status !== 200) {
                        try {
                            const errorData = await e.json();
                            let commonData = {
                                ...this.sseParams,
                                "query": prompt,
                            }
                            let fillData = {
                                ...commonData,
                                "response": errorData.msg                                
                            }
                            this.$refs['session-com'].replaceLastData(lastIndex, fillData)
                        } catch (e) {
                            const text = await e.text();
                            this.$message.error(text || 'Unknown error');
                        }

                        this.stopEventSource();
                        this.setStoreSessionStatus(-1);
                        return;
                    }
                },
                onmessage: (e) => {
                    if (e && e.data) {
                        let data;
                        try {
                            data = JSON.parse(e.data);
                            console.log('===>',new Date().getTime(), data);
                        } catch (error) {
                            return; // IfParseFailed，DirectlyBack，不Process这条Message
                        }
                        
                        this.sseResponse = data;
                        //待Replace of Data，Need前端组装
                        let commonData = {
                            ...this.sseResponse,
                            ...this.sseParams,
                            "query": prompt,
                            "fileName":'',
                            "fileSize":'',
                            "response": '',
                            "filepath": '',
                            "requestFileUrls":'',
                            "searchList": this.sseResponse.data && this.sseResponse.data.searchList ? this.sseResponse.data.searchList: [],
                            "gen_file_url_list": [],
                            "thinkText":'Thinking',
                            "isOpen":true,
                            "citations":[]
                        }
                        if(data.code === 0 || data.code === 1){
                            //finish 0：进行中  1：Close   2:Sensitive WordClose
                            let _sentence = data.data.output;
                                this._print.print(
                                    {
                                        response:_sentence,
                                        finish:data.finish
                                    },
                                    commonData,
                                    (worldObj,search_list) => {
                                        this.setStoreSessionStatus(0)
                                        endStr += worldObj.world  
                                        endStr = convertLatexSyntax(endStr)
                                        endStr = parseSub(endStr, lastIndex)     
                                        let fillData = {
                                            ...commonData,
                                            "response":md.render(endStr),
                                            oriResponse:endStr,
                                            finish:worldObj.finish,
                                            searchList:(search_list && search_list.length) ? search_list.map(n => ({
                                                  ...n,
                                                  snippet: md.render(n.snippet)
                                                }))
                                            : []
                                        }
                                        this.$refs['session-com'].replaceLastData(lastIndex, fillData)
                                        if(worldObj.isEnd && worldObj.finish === 1){
                                            this.setStoreSessionStatus(-1)
                                        }
                                    })
                            // this.$nextTick(()=>{
                            //     this.$refs['session-com'].scrollBottom()
                            // })
                        }else if(data.code === 7 || data.code === -1){
                            this.setStoreSessionStatus(-1)
                            let fillData = {
                                ...commonData,
                                "response": data.message                            
                            }
                            this.$refs['session-com'].replaceLastData(lastIndex, fillData)
                        }
                    }
                },
                onclose: () => {
                    console.log('===> eventSource onClose')
                    this.setStoreSessionStatus(-1)//Close后改变Status
                    this.sseOnCloseCallBack()
                },
                onerror: (e) => {
                    console.log("Service connection exception, please try again!");
                    if (e.readyState === EventSource.CLOSED) {
                        console.log("connection is closed");
                    } else {
                        console.log("Error occured", e);
                    }
                    this.stopEventSource()//前端主动CloseJoin
                    this.setStoreSessionStatus(-1)//Close后改变Status
                }
            });
                      
        },
        doSend(params) {
            this.stopBtShow = true
            this.isStoped= false
            let _history = this.$refs['session-com'].getList()
            this.sendEventSource(this.inputVal, '', _history.length)
        },
        sendEventSource(prompt, msgStr, lastIndex) {
            console.log('####  sendEventSource',new Date().getTime())
            const userInfo = this.$store.state.user.userInfo || {}
            if (this.sessionStatus === 0) {
                this.$message.warning('The previous question has not been answered yet!')
                return
            }

            this.sseResponse = {}
            //Send问题后不允许继续提问
            this.setStoreSessionStatus(0)

            this.clearInput()

            let params = {
                query: prompt, 
                pending: true, 
                responseLoading: true, 
                requestFileUrls: this.queryFilePath?[this.queryFilePath]:[],
                fileList:this.fileList,
                pendingResponse:''
            }
            //正式Environment传ModelParameter
            this.$refs['session-com'].pushHistory(params)
            
            let endStr = ''
            this._print = new Print({
                onPrintEnd: () => {
                }
            })

            let data = null;
            let headers = null;
            //判断YesYesIs Notopenurl对话
            if(this.type === 'agentChat'){
                this.sseApi = `${USER_API}/assistant/stream`;
                const trial = this.isTestChat ? true : false
                data = {
                    ...this.sseParams,
                    prompt,
                    trial
                };
                headers = {
                    "Content-Type": 'application/json',
                    'Authorization': 'Bearer ' + this.token,
                    "x-user-id": userInfo.uid,
                    "x-org-id": userInfo.orgId
                }
            }else{
                this.sseApi = `${OPENURL_API}/agent/${this.sseParams.assistantId}/stream`;
                data = {
                   conversationId:this.sseParams.conversationId, 
                   prompt
                }
                headers = {
                    'X-Client-ID':this.getHeaderConfig().headers['X-Client-ID']
                };
            }

            this.ctrlAbort = new AbortController();
            this.eventSource = new fetchEventSource(this.origin + this.sseApi, {
                method: 'POST',
                headers,
                signal: this.ctrlAbort.signal,
                body: JSON.stringify(data),
                openWhenHidden: true, //Page退to后台保持Join
                ...(this.type === 'webChat' && { isOpenUrl: true }),
                onopen: async(e) => {
                    console.log("SSE connection established~",new Date().getTime());
                    if (e.status !== 200) {
                        try {
                            const errorData = await e.json();
                            let commonData = {
                                ...this.sseParams,
                                "query": prompt,
                            }
                            let fillData = {
                                ...commonData,
                                "response": errorData.msg                                
                            }
                            this.$refs['session-com'].replaceLastData(lastIndex, fillData)
                        } catch (e) {
                            const text = await e.text();
                            this.$message.error(text || 'Unknown error');
                        }

                        this.stopEventSource();
                        this.setStoreSessionStatus(-1);
                        return;
                    }
                },
                onmessage: (e) => {
                    if (e && e.data) {
                        let data = JSON.parse(e.data)
                        console.log('===>',new Date().getTime(),data)
                        this.sseResponse = data
                        //待Replace of Data，Need前端组装
                        let commonData = {
                            ...data,
                            ...this.sseParams,
                            "query": prompt,
                            "fileList":this.fileList,
                            "response": '',
                            "filepath": data.file_url || '',
                            "requestFileUrls": this.queryFilePath?[this.queryFilePath] : data.requestFileUrls,
                            "searchList": data.search_list || [],
                            "gen_file_url_list":data.gen_file_url_list || [],
                            "thinkText":i18n.t('agent.thinking'),
                            'toolText':'Using tool...',
                            "isOpen":true,
                            "showScrollBtn":null,
                            "citations":[]
                        }

                        if(data.code === 0){
                            //finish 0：进行中  1：Close   2:Sensitive WordClose
                            let _sentence = data.response
                                this._print.print(
                                    {
                                        response:_sentence,
                                        finish:data.finish
                                    },
                                    commonData,
                                    (worldObj,search_list) => {
                                        this.setStoreSessionStatus(0)
                                        endStr += worldObj.world
                                        endStr = convertLatexSyntax(endStr)
                                        endStr = parseSub(endStr, lastIndex)
                                        const finalResponse = String(endStr)
                                        let fillData = {
                                            ...commonData,
                                            response: [0,1,2,3,4,6,20,21,10].includes(commonData.qa_type)?md.render(finalResponse):finalResponse.replaceAll('\n-','<br/>•').replaceAll('\n','<br/>'),
                                            finish:worldObj.finish,
                                            oriResponse:endStr,
                                            searchList:(search_list && search_list.length) ? search_list.map(n => ({
                                                  ...n,
                                                  snippet: md.render(n.snippet)
                                                }))
                                            : []
                                        }
                                        this.$refs['session-com'].replaceLastData(lastIndex, fillData)
                                        if(worldObj.finish !== 0){
                                            if(worldObj.finish === 4){
                                                let fillData = {
                                                    ...commonData,
                                                    "response":i18n.t('yuanjing.sensitiveTips')
                                                }
                                                this.$refs['session-com'].replaceLastData(lastIndex, fillData)
                                            }
                                            this.setStoreSessionStatus(-1)
                                        }
                                        if(worldObj.isEnd && worldObj.finish === 1){
                                          this.setStoreSessionStatus(-1)
                                       }
                                    })

                            this.$nextTick(()=>{
                                this.$refs['session-com'].scrollBottom()
                            })

                        }else if(data.code === 7 || data.code === -1 || data.code === 1){
                            this.setStoreSessionStatus(-1)
                            let fillData = {
                                ...commonData,
                                "response": data.message                               
                            }
                            this.$refs['session-com'].replaceLastData(lastIndex, fillData)
                        }
                    }
                },
                onclose: () => {
                    console.log('===> eventSource onClose')
                    this.setStoreSessionStatus(-1)//Close后改变Status
                    this.sseOnCloseCallBack()
                },
                onerror: (e) => {
                    console.log("Service connection exception, please try again!");
                    if (e.readyState === EventSource.CLOSED) {
                        console.log("connection is closed");
                    } else {
                        console.log("Error occured", e);
                    }
                    this.stopEventSource()//前端主动CloseJoin
                    this.setStoreSessionStatus(-1)//Close后改变Status
                }
            });
        },
        preStop() {
            //Get已经拿到 of 全部回答,一次性回显出来
            this.sseOnCloseCallBack(true)
        },
        sseOnCloseCallBack(isStoped){
            this.stopEventSource()
            //图文问答不Use打字机
           /* if(this.sseResponse.qa_type === 6){
                return
            }*/
            //主动Stop
            if(isStoped) {
                this.stopAndEcho()
            }else{
                //收到onclose,ANDUse of Yes文生代码
                if(this.sseResponse.qa_type === 4){
                    this.stopAndEcho()
                }else{
                    //Interface405Etc
                    let history_list = []
                    let lastIndex = history_list.length - 1
                    let lastRQ = history_list[lastIndex]
                    let endStr =this._print.getAllworld()
                    endStr = convertLatexSyntax(endStr)
                    // ReplaceTag
                    endStr = parseSub(endStr)
                    // IfBackHasResult，Then在结束 when 不Show“本次回答已终止”
                    this.runResponse = md.render(endStr)
                    this.runDisabled = false
                    this.setStoreSessionStatus(-1)
                }
            }
        },
        stopAndEcho(){
            //暂存已经收到 of 所Hasresponse
            let endResponse = this._print.getAllworld()

            this._print && this._print.stop()

            setTimeout(()=>{
                this.setStoreSessionStatus(-1)

                let history_list = []
                let lastIndex = history_list.length - 1
                let lastRQ = history_list[lastIndex]
                if(endResponse){
                    endResponse = convertLatexSyntax(endResponse)
                    // ReplaceTag
                    endResponse = parseSub(endResponse)
                    this.runResponse = md.render(endResponse)
                    this.runDisabled = false
                    this.$nextTick(()=>{
                        this.addCopyClick()
                    })

                }else{
                    if(Object.keys(this.sseResponse).length !== 0 && this.sseResponse.code !== 7){
                        this.runResponse ="This response has been terminated"
                        this.setStoreSessionStatus(-1)
                    }else{
                        this.stopEventSource();
                        this.setStoreSessionStatus(-1)
                        this.$refs['session-com'].stopPending();
                    }
                }
            },15)
        },
        stopEventSource() {
            this.ctrlAbort && this.ctrlAbort.abort()
            this.eventSource = null
        },
        refreshLastSession(){
            let endResponse = this._print.getAllworld()
            let history_list =[]
            let lastIndex = history_list.length - 1
            let lastRQ = history_list[lastIndex]
            // this.$refs['session-com'].replaceLastData(lastIndex, {
            //     ...lastRQ,
            //     response: endResponse
            // })
        },
        setPrompt(data) {
            this.$refs['editable'].setPrompt(data)
        },
        clearInput() {
            this.$refs.editable.clearInput()
            this.$refs.editable.clearFile()
            this.inputVal = ''
            this.fileId = ''
        },
        clearPageHistory(){
            this.$refs['session-com'] && this.$refs['session-com'].clearData()
            this.$refs.editable && this.clearInput()
        },
        clearHistory() {
            this.stopBtShow = false
            this.clearPageHistory()
            // this.doDeleteHistory()
        },
        refresh() {
            let history_list = this.$refs['session-com'].getList();
            let _history = history_list[history_list.length - 1];
            let inputVal = _history.query;
            let fileInfo = _history.fileInfo ? _history.fileInfo : [];
            let fileList = _history.fileList ? _history.fileList : [];
            this.preSend(inputVal,fileList,fileInfo);
        }
    }
};