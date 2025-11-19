<!--BatchPlugin-->
<template>
    <div class="rl">
        <div class="editable-box">
            <div  v-if="fileType === 'image/*'" class="echo-img-box">
                <div v-for="(file,i) in fileList" class="echo-img-item" :key="'file'+i">
                    <el-image class="echo-img" :src="file.fileUrl" @click="showBigImg(file.fileUrl)"  :preview-src-list="[file.imgUrl]"></el-image>
                    <i class="el-icon-close echo-close" @click="clearFile"></i>
                </div>
            </div>
            <div v-if="fileType === 'audio/*'" class="echo-audio-box">
                <audio  id="audio" controls>
                    <source :src="fileUrl" type="video/mp3">
                    <source :src="fileUrl" type="audio/ogg">
                    <source :src="fileUrl" type="audio/mpeg">
                    {{$t('agent.autioTips')}}
                </audio>
                <i class="el-icon-close echo-close" @click="clearFile"></i>
            </div>
            <div v-if="fileType === 'doc/*'" class="echo-img-box echo-doc-box">
                <img :src="require('@/assets/imgs/fileicon.png')" class="docIcon">
                <div class="docInfo">
                    <p class="docInfo_name">FileName：{{fileList[0]['name']}}</p>
                    <p class="docInfo_size">File大小：{{fileList[0]['size'] > 1024 ?(fileList[0]['size'] / (1024 * 1024 )).toFixed(5) + ' MB' : fileList[0]['size'] + ' bytes'}}</p>
                </div>
                <span class="el-icon-loading loading-icon" v-if="fileLoading"></span>
                <i class="el-icon-close echo-close" @click="clearFile"></i>
            </div>
            <!-- 问答Input Box -->
            <div class="editable-wp flex">
                <div class="editable-wp-left rl">
                     <!-- <i  class="el-icon-upload2 upload-icon" @click="preUpload"></i> -->
                     <img class="upload-icon" :src="require('@/assets/imgs/uploadIcon.png')" @click="preUpload" v-if="type !== 'webChat'"/>
                </div>
                <div class="editable-wp-right rl">
                    <div class="aibase-textarea editable--input" ref="editor"  @blur="onBlur" v-on:input="getPrompt"  @keydown="textareaKeydown($event)" @dragenter.prevent @dragover.prevent @drop.prevent.stop="handleDrop" contenteditable="true"></div>
                    <span class="editable--placeholder" v-if="!promptValue">{{placeholder}}</span>
                    <i v-if="promptValue" class="el-icon-close editable--close" @click.stop="clearInput"></i>
                    <div class="edtable--wrap">
                        <el-button size="mini" :class="{'btnActive':isActive,'btnAnactive':!isActive}" @click="linkSearch" v-if="showModelSelect && !isPower && isLink">{{$t('agent.connectInternect')}}</el-button>
                        <!-- <img class="editable--send" :src="require('@/assets/imgs/send.png')" @click="preSend" /> -->
                        <el-button type="primary" class="editable--send" @click="preSend"><span>发送</span> <img :src="require('@/assets/imgs/sendIcon.png')" /></el-button>
                    </div>
                </div>
                <!-- 覆盖层，Model下线禁止点击 -->
                <!-- <div class="overlay" v-if="modelParams === null"></div> -->
            </div>
        </div>
        <transition name="el-zoom-in-bottom">
            <div class="perfectReminder-item-box" v-show="randomReminderShow">
                <div class="perfectReminder-item" v-for="(n) in randomReminderList" :key='n.id'  :style="`background-color:${colorArr[n.random]}`">
                    <el-popover
                            placement="top-start"
                            width="300"
                            :visible-arrow="false"
                            trigger="hover"
                            :open-delay="500"
                            :content="n.prompt && n.prompt.replaceAll('{','').replaceAll('}','')">
                        <span style="font-size: 15px"  slot="reference" @click="setRandomReminder(n)">{{n.title}}</span>
                    </el-popover>
                </div>
                <span class="refresh" @click="getReminderList"><i class="el-icon-loading" v-show="refreshLoading"></i>&nbsp;{{$t('agent.next')}}</span>
            </div>
        </transition>

        <upload-dialog ref="upload" :fileTypeArr="fileTypeArr" @setFileId="setFileId" @setFile="setFile"></upload-dialog>

    </div>
</template>

<script>
    import commonMixin from '@/mixins/common'
    import uploadDialog from './uploadBatchDialog'
    import { getModelList } from '@/api/cubm';
    import {mapGetters} from 'vuex'
    export default {
        props: {
            source: {type:String},
            fileTypeArr: {type: Array, required: false,default: () => { return []}},
            showModelSelect:{type:Boolean,default:true},
            currentModel:{type:Object,default: () => { return null}},
            isModelDisable:{type:Boolean,default:false},
            type:{type:String}
        },
        mixins: [commonMixin],
        components:{uploadDialog},
        data(){
            return{
                basePath: this.$basePath,
                isActive:false,
                isPower: this.$platform ==='YWD_RAG' || this.$platform ==='HW_RAG',
                isLink:false,
                modelParams:null,
                modleOptions:[],
                colorArr:['#dca3c2','#aaa9db','#d1a69b','#7894cf','#4fbed9',
                    '#ebb8bd','#9b9655','#3bb4b7','#61aac5','#d79ae5',
                    '#51a2da','#89b0f9','#738cbd'],
                placeholder:'Please enterContent,用Ctrl+Enter可换行',
                promptHtml:'',
                promptValue:'',
                randomReminderList:[],  //随机8个Tip词
                randomReminderShow:false,
                refreshLoading:false,
               //File
                hasFile:false,
                fileIdList:null,
                fileType:'',
                fileList:[],
                fileUrl:'',
                imgUrl:'',
                modelType:'',
                fileLoading:false,
                isDragging:false
            }
        },
        watch:{
            currentModel:{
                handler(val){
                    if(val === null){
                        this.modelParams = null;
                        return
                    }
                    this.modleOptions.map(modelItem =>{
                        if(modelItem.modelId == val.modelId && modelItem.modelVersion == val.modelVersion ){
                            this.modelParams = modelItem;
                        }
                    })
                    this.modelType = this.modelParams.modelSeries === 'deepseek' ? 'deepseek' : 'bigModel';
                    this.$emit('getModelType',this.modelType)
                }
            }
        },
        computed: {
            ...mapGetters("app", ["maxPicNum"]),
        },
        created(){
            // this.isLink = this.commonInfo.data.useModel.useInternet === 1 ? true : false;
            // this.getModelData()
        },
        mounted(){
            // this.$nextTick(() => {
            // this.$setupDragAndDrop({
            // containerSelector: '.editable--input',
            // maxImageFiles: this.maxPicNum,
            // onFiles: (files) => {
            //     this.isDragging = true
            //     this.processFiles(files)
            //     }
            // })
            //  }) 
        },
        methods:{
            // ProcessFile of Method，提取出来供 handleDrop And onFiles Use
            processFiles(files) {
                if (!files || files.length === 0) return
                
                const picked = files
                const fileObjs = picked.map(f => ({
                    fileName: f.name, name: f.name, size: f.size, type: f.type,
                    fileUrl: URL.createObjectURL(f), imgUrl: URL.createObjectURL(f)
                }))
                const ext = (picked[0].name.split('.').pop() || '').toLowerCase()
                const mime = picked[0].type
                let ftype = ''
                if ((mime && mime.indexOf('image/') === 0) || ['jpg','jpeg','png','gif','webp','bmp','svg'].indexOf(ext) > -1) ftype = 'image/*'
                else if ((mime && mime.indexOf('audio/') === 0) || ['mp3','wav','ogg'].indexOf(ext) > -1) ftype = 'audio/*'
                else ftype = 'doc/*'
                this.fileType = ftype
                this.fileList = fileObjs
                this.fileUrl = fileObjs[0].fileUrl
                this.hasFile = true
                this.fileLoading = true

            },
            // ProcessDrag到Input Box of File
            handleDrop(event) {
                /* const dt = event.dataTransfer
                if (!dt || !dt.files) return
                
                const fileList = dt.files
                const files = Array.prototype.slice.call(fileList)
                if (files.length === 0) return
                
                // CallFileProcessMethod
                this.processFiles(files)
                */
            },
            linkSearch(){
                this.isActive = !this.isActive;
            },
            getModelData(type=''){
                getModelList({modelType:''}).then(res =>{
                    if(res.code === 0){
                        this.modleOptions = res.data.list || [];
                        if(type === ''){
                            this.modelParams = this.modleOptions[0];
                        }
                        this.modelType = this.modelParams.modelSeries === 'deepseek' ? 'deepseek' : 'bigModel';
                        this.$emit('getModelType',this.modelType);
                    }
                }).catch(err =>{
                })
            },
            visibleChange(val){
                if(val){
                    this.getModelData('refresh');
                }
            },
            modelChange(){
                this.modelType = this.modelParams.modelSeries === 'deepseek' ? 'deepseek' : 'bigModel';
                this.$emit('getModelType',this.modelType)
                this.$emit('modelChange')
            },
            showBigImg(url){
                console.log(url)
            },
            clearFile(){
                this.fileIdList = null
                this.fileList = []
                this.fileType = ''
                this.fileUrl = ''
                this.imgUrl = ''
                this.hasFile = false
            },
            /*showFileUpload(status){
              this.hasFile = status
            },*/
            preUpload(){
              this.$refs['upload'].openDialog()
            },
            setFileId(fileIdList){
                this.fileIdList = fileIdList;
                this.fileUrl = this.fileIdList[this.fileIdList.length-1].fileUrl;
                //this.imgUrl = fileIdList.imgUrl;
                let fileType =  this.fileIdList[this.fileIdList.length-1]['fileName'].split('.').pop() || '';
                if(["jpeg", "PNG", "png", "JPG", "jpg"].includes(fileType)){//'bmp','webp'
                    this.fileType = 'image/*'
                }
                if(["mp3", "wav"].includes(fileType)){
                    this.fileType = 'audio/*'
                }
                if(['txt','csv','xlsx','doc','docx','html','pptx','pdf'].includes(fileType)){
                    this.fileType = 'doc/*'
                }
                //this.$emit('setSessionStatus',0)   //UploadImage后，算法自动BackResult，此 when 将statusSetting为0
            },
            setFile(fileList){
                this.fileList = fileList;
                if(this.fileList.length > 0){
                    this.hasFile = true;
                }
            },
            getFileList(){
                return this.fileList
            },
            getFileIdList(){
                return this.fileIdList
            },
            setRandomReminder(n){
              this.setPrompt(n.prompt)
              this.randomReminderShow = false
            },
            clearInput(){
                this.$refs.editor.innerHTML = ''
                this.promptHtml = ''
                this.promptValue = ''
                this.randomReminderShow && (this.randomReminderShow = false)
            },
            getContentInBraces(str) {
                const regex = /{([^}]+)}/g;
                let match;
                const matches = [];

                while ((match = regex.exec(str))) {
                    matches.push(match[1]); // 第一个括号中 of ContentYes我们想要 of Result
                }

                return matches;
            },
            setPrompt(data){
              this.clearInput()
              this.promptValue = data
              this.$refs.editor.innerHTML = data.replaceAll('{','<div class="light-input" contenteditable="true">').replaceAll('}','</div>')
              // let matchArr = this.getContentInBraces(data)
            },
            getPrompt(){
                if(this.source === 'perfectReminder'){
                    if(this.$refs.editor.innerHTML === '/'){
                        this.getReminderList()
                    }else {
                        this.randomReminderShow = false
                    }
                }
                let prompt = this.$refs.editor.innerText
                this.promptValue = prompt
                return prompt
            },
            getPromptBak(){
                let prompt = ''
                if(this.source === 'perfectReminder'){
                    if(this.$refs.editor.innerHTML === '/'){
                        this.getReminderList()
                    }else {
                        this.randomReminderShow = false
                    }
                    prompt = this.$refs.editor.innerHTML.replaceAll('<div class="light-input" contenteditable="true">','').replaceAll('</div>','').replaceAll(' ','')
                }else{
                    prompt = this.$refs.editor.innerHTML  //从DialogCopy 来 of 会换行，临 when Process，后期优化
                }
                let prompt2 = prompt.replace('<div><br></div>','')
                this.promptValue = this.$refs.editor.innerHTML
                return prompt2
            },
            onBlur(){
                //勿删，Define此MethodUsed forGet焦点
            },
            async getReminderList(){
                //Show8个Tip词
                this.refreshLoading = true
                let res = await this.$api.expand.getPerfectReminderV2({pageNo:1,pageSize:1000,randomNum:8})
                if(res.code === 0){
                    this.refreshLoading = false
                    this.randomReminderShow = true
                    this.randomReminderList = res.data.list.map(item=>{
                        return {
                            ...item,
                            random:parseInt(Math.random(13)*10),
                        }
                    })
                }
            },
            //换行并重新定位光标位置
            textareaRange() {
                var el = this.$refs.editor
                var range = document.createRange();
                //BackUser当前 of 选Area
                var sel =  document.getSelection();
                //Get当前光标位置
                var offset = sel.focusOffset
                //div当前Content
                var content = el.innerHTML　　　　　//Add换行符\n
                el.innerHTML = content.slice(0,offset)+'\n'+content.slice(offset)　　　　　//Setting光标为当前位置
                range.setStart(el.childNodes[0], offset+1);
                //使得选Area(光标)开始With结束位置重叠
                range.collapse(true);
                //Remove现HasOther of 选Area
                sel.removeAllRanges();
                //加入光标 of 选Area
                sel.addRange(range);
            },
            //Listen按KeyOperation
            textareaKeydown (event) {
                if(event.ctrlKey && event.keyCode === 13){
                    //ctrl+enter
                    this.textareaRange()
                }else if (event.keyCode === 13) {
                    //enter
                    this.preSend()
                    event.preventDefault() // 阻止浏览器Default换行Operation
                    return false
                }
            },
            getModelInfo(){
                return this.modelParams || null
            },
            sendUseSearch(){
                return this.isActive;
            },
            preSend(){
                this.hasFile = false;
                this.$emit('preSend');
            },
            goModelList(){//跳转到Service管理
                location.href = window.location.origin + `${this.$basePath}/aibase/portal/training/releaseTable`
            }
        }
    }
</script>
<style lang="scss" scoped>
    .tips{color:#ccc;}
    //Model选择框自适应
    .auto-width-select{
        min-width:250px;
        max-width:450px;
    }
    .editable-box{
        border:1px solid #d3d7dd ;
        .loading-icon{
            font-size:18px;
            color:$color;
            margin-left:10px;
        }
        .echo-img-box{
            position: absolute;
            display:flex;
            top:-70px;
            justify-content: flex-start;
            align-items: center;
            gap:10px;
            .echo-img-item{
                height:60px;
                width:60px;
                display:flex;
            }
            .echo-img{
                width: 100%;
                height: 100%;
                object-fit: contain;
                background: #ffff;
                box-shadow: 1px 1px 10px #9b9a9a;
                border-radius: 4px;
            }
            .echo-close{
                position: absolute;
                right: 0;
                top:0;
                background-color: #333;
                color: #fff;
                cursor: pointer;
            }
            .fileid-icon{
                line-height: 20px;
                position: absolute;
                bottom: 0;
                text-align: center;
                background: #3333337a;
                width: 100%;
                color: #67C23A;
                i{
                    font-weight: bold;
                    font-size: 16px;
                }
            }
        }
        .echo-doc-box{
            background:#fff;
            width: auto;
            border:1px solid #DCDFE6;
            border-radius:5px;
            display:flex;
            justify-content: space-between;
            align-items: center;
            padding:10px 50px 10px 5px;
            .docIcon{
                width:30px;
                height:30px;
            }
            .docInfo{
               .docInfo_name{
                  color:#333;
                }
                .docInfo_size{
                  color:#bbbbbb;
                }
            }
        }
        .echo-audio-box{
            position: absolute;
            width: 300px;
            height: 40px;
            top: -60px;
            audio{
                width: 100%;
            }
            .echo-close{
                position: absolute;
                top:0;
                right:0;
                background-color: #333;
                color: #fff;
            }
        }
        .editable-wp{
            position:relative;
            .overlay{
                position:absolute;
                top: 0;
                left: 0;
                right: 0;
                bottom: 0;
                z-index: 9999;
                background-color:rgba(255,255,255,.4);
                border:1px solid #DCDFE6;
                border-radius:6px;
                pointer-events: auto;
            }
        }
        .editable-wp-left{
            min-width: 20px;
            .upload-icon{
                margin: 5px 5px 5px 11px;
                padding:3px;
                border-radius: 4px;
                cursor: pointer;
                // border: 1px solid $color;
                // color: #fff;
                // background: $color;
                // font-size: 13px;
            }
        }
        .editable-wp-right{
            flex:1;
        }
    }
    .aibase-textarea{
        padding: 10px 35px 35px 0;
    }
    .editable--placeholder{
        left: 0 !important;
    }
    .edtable--wrap{
        width: 100%;
        padding-bottom:10px;
        height:35px;
        position:absolute;
        bottom:0;
        left:0;
        display:flex;
        justify-content:space-between;
        align-items:center;
        z-index:999;
    }
    .model-box{
        padding:10px 0;
    }
    .btnActive{
        color: #E60001!important;
        border: 1px solid rgb(228, 165, 165)!important;
        background: linear-gradient(111deg, rgba(255, 58, 58, 0.2) 0%, #fff 25%, #fff 69%, rgba(255, 58, 58, 0.2) 100%)!important;
        // background: (111deg, rgba(255, 58, 58, 0.2) -12%, #fff 25%, #fff 69%, rgba(255, 58, 58, 0.2) 113%)
    }
    .btnAnactive{
        color: #606266!important;
        border: 1px solid #DCDFE6!important;
        background:#ffffff!important;
    }
    .editable-box /deep/.light-input{
        border: 1px solid deepskyblue;
        padding: 2px 14px  2px 10px;
        margin: 0 5px;
        border-radius: 4px;
        display: inline-block;
        box-shadow: 1px 1px 10px #d3ebf3;

    }
    .perfectReminder-item-box{
        position: absolute;
        width: 100%;
        height: 174px;
        top: -176px;
        left: 0;
        padding: 22px 20px 40px 20px;
        overflow: hidden;
        background: #fff;
        box-shadow: 1px 1px 10px #dce7f5;
        border-radius: 6px 6px 0 0;
        .perfectReminder-item{
            width: calc((100% - 80px)/4);
            height: 46px;
            line-height: 46px;
            text-align: center;
            position: relative;
            margin: 5px 10px;
            display: inline-block;
            background-color: #dfebfb;
            /*color: #3888f9;*/
            color: #fff;
            cursor: pointer;
            border-radius: 9px;
        }
        .perfectReminder-active{
            border: 1px solid #EC0B0C;
            overflow: hidden;
            i,span{
                color: #EC0B0C;
            }
        }
        .refresh{
            position: absolute;
            right: 30px;
            bottom: 10px;
            cursor: default;
            color: #62a1fb;
        }
    }



</style>
