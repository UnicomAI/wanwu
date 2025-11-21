<!--BatchPlugin-->
<template>
    <div class="rl">
        <div class="editable-box">
            <div  v-if="fileType === 'image/*'" class="echo-img-box">
                <el-image class="echo-img" :src="fileUrl" @click="showBigImg(fileUrl)"  :preview-src-list="[fileUrl]"></el-image>
                <i class="el-icon-close echo-close" @click="clearFile"></i>
            </div>
            <div v-if="fileType === 'audio/*'" class="echo-audio-box">
                <audio  id="audio" controls>
                    <source :src="fileUrl" type="video/mp3">
                    <source :src="fileUrl" type="audio/ogg">
                    <source :src="fileUrl" type="audio/mpeg">
                    {{$t('editavleDiv.autioTips')}}
                </audio>
                <i class="el-icon-close echo-close" @click="clearFile"></i>
            </div>
            <div v-if="fileType === 'doc/*'" class="echo-img-box echo-doc-box">
                <img :src='require('@/assets/imgs/docFile.png')' class="docIcon">
                <div class="docInfo">
                    <p class="docInfo_name">FileName：{{fileList[0]['name']}}</p>
                    <p class="docInfo_size">Filesize：{{fileList[0]['size'] > 1024 ?(fileList[0]['size'] / (1024 * 1024 )).toFixed(2) + ' MB' : fileList[0]['size'] + ' bytes'}}</p>
                </div>
                <i class="el-icon-close echo-close" @click="clearFile"></i>
            </div>
            <!-- 问答Input Box -->
            <div class="editable-wp flex">
                <div class="editable-wp-left rl" >
                     <!-- <i  class="el-icon-upload2 upload-icon" @click="preUpload"></i> -->
                </div>
                <div class="editable-wp-right rl">
                    <div class="aibase-textarea editable--input" ref="editor"  @blur="onBlur" v-on:input="getPrompt"  @keydown="textareaKeydown($event)" contenteditable="true"></div>
                    <span class="editable--placeholder" v-if="!promptValue">{{placeholder}}</span>
                    <i v-if="promptValue" class="el-icon-close editable--close" @click.stop="clearInput"></i>
                    <div class="edtable--wrap">
                        <el-button size="mini" :class="{'btnActive':isActive,'btnAnactive':!isActive}" @click="linkSearch" v-if="showModelSelect && !isPower && isLink">{{$t('editavleDiv.connectInternect')}}</el-button>
                        <!-- <img class="editable--send" :src="require('@/assets/imgs/send.png')" @click="preSend" /> -->
                        <el-button type="primary" class="editable--send" @click="preSend"><span>send</span> <img :src="require('@/assets/imgs/sendIcon.png')" /></el-button>
                    </div>
                </div>
                <!-- override层，Model下线禁止click -->
                <!-- <div class="overlay" v-if="modelParams === null"></div> -->
            </div>
        </div>
        <transition name="el-zoom-in-bottom">
            <div class="perfectReminder-item-box" v-show="randomReminderShow">
                <div class="perfectReminder-item" v-for="(n,i) in randomReminderList" :key='n.id'  :style="`background-color:${colorArr[n.random]}`">
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
    import uploadDialog from './uploadBatchDialog'
    export default {
        props: {
            source: {type:String},
            fileTypeArr: {type: Array, required: false,default: () => { return []}},
            showModelSelect:{type:Boolean,default:true},
            currentModel:{type:Object,default: () => { return null}},
            isModelDisable:{type:Boolean,default:false}
        },
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
                placeholder:'Please enterContent,用Ctrl+Enter可换row',
                promptHtml:'',
                promptValue:'',
                randomReminderList:[],  //随机8Tip词
                randomReminderShow:false,
                refreshLoading:false,
               //File
                hasFile:false,
                fileIdList:[],
                fileType:'',
                fileList:[],
                fileUrl:'',
                modelType:''
            }
        },
        methods:{
            linkSearch(){
                this.isActive = !this.isActive;
            },
            showBigImg(url){
                console.log(url)
            },
            clearFile(){
                this.fileIdList = []
                this.fileList = []
                this.fileType = ''
                this.fileUrl = ''
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
                this.fileUrl = fileIdList[0].downloadUrl;

                let fileType = this.fileUrl.split('.')[this.fileUrl.split('.').length-1]
                if(["jpeg", "PNG", "png", "JPG", "jpg",'bmp','webp'].includes(fileType)){
                    this.fileType = 'image/*'
                }
                if(["mp3", "wav"].includes(fileType)){
                    this.fileType = 'audio/*'
                }
                if(['txt','csv','xlsx','doc','docx','html','pptx','pdf'].includes(fileType)){
                    this.fileType = 'doc/*'
                }
                //this.$emit('setSessionStatus',0)   //UploadImageafter，算法自动BackResult，此 when 将statusSettingis0
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
                    matches.push(match[1]); // first括号in of ContentYes我们想to of Result
                }

                return matches;
            },
            setPrompt(data){
              this.clearInput()
              this.promptValue = data
              this.$refs.editor.innerHTML = data.replaceAll('{','<div class="light-input" contenteditable="true">').replaceAll('}','</div>')
              //  console.log(data,this.getContentInBraces(data))
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
                    prompt = this.$refs.editor.innerHTML  //从DialogCopy 来 of 会换row，临 when Process，after期优化
                }
                let prompt2 = prompt.replace('<div><br></div>','')
                this.promptValue = this.$refs.editor.innerHTML
                return prompt2
            },
            onBlur(){
                //勿删，Define此MethodUsed forGet焦点
            },
            async getReminderList(){
                //Show8Tip词
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
            //换row并restart定位光标位置
            textareaRange() {
                var el = this.$refs.editor
                var range = document.createRange();
                //BackUsercurrent of 选Area
                var sel =  document.getSelection();
                //Getcurrent光标位置
                var offset = sel.focusOffset
                //divcurrentContent
                var content = el.innerHTML　　　　　//Add换row符\n
                el.innerHTML = content.slice(0,offset)+'\n'+content.slice(offset)　　　　　//Setting光标iscurrent位置
                range.setStart(el.childNodes[0], offset+1);
                //使得选Area(光标)startWithend位置重叠
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
                    event.preventDefault() // preventbrowserDefault换rowOperation
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
            goModelList(){//jumptoService管理
                location.href = window.location.origin + `${this.$basePath}/aibase/portal/training/releaseTable`
            }
        }
    }
</script>
<style lang="scss" scoped>
    .tips{color:#ccc;}
    //Model Selection框自适应
    .auto-width-select{
        min-width:250px;
        max-width:450px;
    }
    .editable-box{
        border:1px solid #d3d7dd ;
        .echo-img-box{
            position: absolute;
            width: 90px;
            height: 90px;
            top:-95px;
            .echo-img{
                width: 100%;
                height: 100%;
                object-fit: contain;
                background: #ffff;
                box-shadow: 1px 1px 10px #9b9a9a;
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
            padding:2px 50px 5px 5px;
            .docIcon{
                width:30px;
                height:30px;
                margin-right:10px;
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
                margin: 11px;
                padding: 3px;
                border-radius: 4px;
                border: 1px solid $color;
                color: #fff;
                background: $color;
                font-size: 13px;
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
