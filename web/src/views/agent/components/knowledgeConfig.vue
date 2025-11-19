<template>
    <div class="knowledgeEnhance ">
        <div class="knowledge-upload" v-loading="fileLoading" element-loading-background="rgba(255, 255, 255, 0.8)">
            <div class="el-upload-box">
                <!--算法ParseFile能力不够，改iseach次只Upload一File-->
                <el-upload
                        :disabled="!assistantId"
                        :class="!assistantId?'not-allow':'allow'"
                        action="#"
                        drag
                        :on-change="fileChange"
                        accept=".pdf, .docx, .txt"
                        :auto-upload="false"
                        :limit="1"
                        :show-file-list="false"
                        :file-list="fileList"
                        ref="uploadFile">
                    <i class="el-icon-upload" :title="assistantId?$t('createApp.knowledgeConfigTips'):''"></i><span style="margin-left: 4px;">{{$t('createApp.knowledgeConfigUploud')}}</span>
                </el-upload>
                <p v-if="!assistantId" class="disabled-tip">{{$t('createApp.knowledgeConfigTips1')}}</p>

            </div>
            <div v-if="fileResList.length" class="knowledge-file">
                <p class="file__item" v-for="(n,i) in fileResList" :key="`${i}fl`" @mouseenter="fileMouseEnter(n)" @mouseleave="fileMouseLeave(n)">
                    <span>{{i+1}}&nbsp;&nbsp;</span><span>{{n.name}}</span>
                    <i  v-show="!n.hover || n.status==='loading'" :class="['file-icon',fileStatus[n.status]]"></i>
                    <i v-show="n.hover && n.status!=='loading'" class="el-icon-error file-icon del" @click="preDelete(n,i)"></i>
                </p>
            </div>

        </div>
    </div>
</template>
<script>
    import { knowledgeFileUpload, deleteKnowledgeFile,getKnowledgeFileList } from '@/api/chat'

    export default {
        props:{
            assistantId:{
                type:String,
                required:false
            }
        },
        data(){
            return{
                fileStatus:{
                    success:'el-icon-circle-check success',
                    fail:'el-icon-circle-close fail',
                    loading:'el-icon-loading'
                },
                configForm:{
                    threshold:0.3, //ws传
                    topK:3, //ws传
                    sentenceSize:100  //UploadFile传
                },
                headers:{
                    'Authorization': "Bearer " + JSON.parse(localStorage.getItem('access_cert')).user.token,
                    'X-User-Id': JSON.parse(localStorage.getItem('access_cert')).user.userInfo.uid,
                    "x-org-id": this.$store.state.user.userInfo.orgId
                },
                //UploadFile
                limitFileType:['.txt','.docx','.pdf'],
                fileLoading:false,
                fileList:[],
                fileResList:[]
            }
        },
        computed:{
            fileData(){
                return {
                    'assistantId': this.assistantId
                }
            },
        },
        methods:{
            setAssistantId(id){
                this.assistantId = id
            },
            setFileList(list){
                this.fileResList = list ? list.map(n=>{
                    return {
                        ...n,
                        name:n.fileName,
                        hover:false,
                        status:'success'
                    }
                }) : []
            },
            beforeAvatarUpload(file, limit) {
                const isLegal = limit.some(item=>{
                    return file.name.endsWith(item)
                })
                !isLegal && this.$message.warning(this.$t('createApp.fileError'));
                return isLegal;
            },
            fileChange(file, fileList) {
                if (!this.beforeAvatarUpload(file, this.limitFileType)) {
                    return;
                }
                let index = this.fileResList.length
                this.$set(this.fileResList,index,{name:file.name,fileId:'',status:'loading'})
                this.doUploadFile(file,index)
                this.fileList = []
            },
            async doUploadFile(file,index){
                var formData = new FormData()
                var config = {
                    headers: { 'Content-Type': 'multipart/form-data' }
                }
                formData.append('assistantId', this.assistantId)
                formData.append('files', file.raw)
                let res = await knowledgeFileUpload(formData, config)
                if(res.code === 0){
                    let row = {name:file.name,fileId:res.data.list[0],hover:false,status:'success'}
                    this.$set(this.fileResList,index,row)
                }else{
                    let row = {name:file.name,fileId:'',hover:false,status:'fail'}
                    this.$set(this.fileResList,index,row)
                }
            },
            async preDelete(n,i){
                switch (n.status){
                    case 'success':
                        this.$confirm(this.$t('createApp.deleteFileTips'),this.$t('knowledgeManage.tip'), {
                            confirmButtonText: this.$t('createApp.confirm'),
                            cancelButtonText: this.$t('createApp.cancel'),
                            type: 'warning'
                        }).then(async() => {
                            let fileId = n.fileId
                            this.fileLoading = true
                            let res = await deleteKnowledgeFile({fileId, assistantId:this.assistantId})
                            if(res.code === 0){
                                this.fileLoading = false
                                this.spliceFile(i)
                            }
                        }).catch(() => {

                        });

                        break;
                    case 'fail':
                        this.spliceFile(i)
                        break;
                }
            },
            //Delete指定下标 of File
            spliceFile(index){
                this.fileResList.splice(index,1)
            },
            fileMouseEnter(n){
                this.fileResList.forEach(m=>{ m.hover = false})
                n.hover = true
            },
            fileMouseLeave(n){
                n.hover = false
            },
            //---自动Upload自带Callback，FileUploadFailed也会ShowSuccessicon，弃用---
            async beforeRemove(file,fileList){
                let fileId = file.fileId || file.response.data.list[0]
                let res = await deleteKnowledgeFile({fileId, assistantId:this.assistantId})
                return res.code === 0
            },
            async onSuccess(response,file,fileList){
                console.log('===>',response,file,fileList)
                //await getKnowledgeFileList({assistantId:this.assistantId,pageNo:1,pageSize:1000})
                if(response.code === 0){
                    this.fileList = fileList
                }else{
                    this.$message.error(response.msg)
                }
            },

        }
    }
</script>

<style lang="scss" scoped>
    .knowledgeEnhance{
        height: 100%;
        .disabled-tip{
            text-align: left;
            margin:5px 0;
            color: #f0900c;
        }
        .knowledge-upload{
            position: relative;
            height: calc(100% - 64px);
            display: flex;
            flex-direction: column;
            .file-total{
                position: absolute;
                bottom: 7px;
                font-size: 12px;
                text-align: right;
                right: 0;
            }
            .el-upload-box{
                width: 360px;
                position: relative;
                text-align: center;
                color: #EC0B0C;
                cursor: pointer;
                i{
                    font-size: 16px;
                }
                .not-allow, .not-allow/deep/.el-upload-dragger{
                    cursor: not-allowed;
                }
                .allow, .allow/deep/.el-upload-dragger{
                    cursor: pointer;
                }
                /deep/{
                    .el-upload{
                        width: 100%;
                        height:100%;
                        color: #EC0B0C;
                    }
                    .el-upload-list {
                        background: #f1f1f1;
                    }
                    .el-upload-list__item{
                        text-align: left!important;
                    }
                    .el-upload-dragger {
                        border: 1px dashed #eb0a0b;
                        background-color: #FEF1F1;
                        width: 100%;
                        height: 100% !important;
                    }
                    .el-upload-dragger .el-icon-upload {
                        font-size: 20px!important;
                        color: #eb0a0b;
                        margin: 0!important;
                    }
                }
            }
            .knowledge-file{
                margin-top: 10px;
                background-color: #f1f1f1;
                border-radius: 6px;
                padding: 10px 15px;
                .file__item{
                    position: relative;
                    color: #666;
                    line-height: 24px;
                    &:hover>i{
                        color:#E60001
                    }
                    i{
                        position: absolute;
                        right: 0;
                        cursor: pointer;
                        padding: 3px 6px;
                    }
                    .del{
                        &:hover{
                            color:#E60001;
                        }
                    }
                    .success{
                        color: #67C23A;
                    }
                    .fail{
                        color: #E60001;
                    }
                }
            }
            .upload--bt-box{
                height: 70px;
                .upload--bt{
                    border: 1px solid #515a71;
                    border-radius: 21px;
                    padding: 4px 14px;
                    color: #ccc;
                    text-align: center;
                    cursor: pointer;
                    float: right;
                    margin: 10px 0;
                }
            }
        }
    }
    .el-form-item{
        display: flex;
        margin-bottom: 0;
        margin-top: -1px;
        border: 1px solid #ddd;
        /deep/{
            .el-form-item__label{
                display: inline-block;
                flex:1;
                border-right: 1px solid #ddd;
                text-align: center;
            }
            .el-form-item__content{
                flex:1;
            }
            .el-input__inner {
                border: 1px solid transparent!important;
            }
        }

    }
    .api-bt{
        cursor: pointer;
        color: #33a4df;
        font-size: 16px;
        padding: 5px;
    }
    .m-collapse {
        margin-top: 20px;
        border-top: 1px solid transparent!important;
        border-bottom: 1px solid transparent!important;
        /deep/ {
            .el-collapse-item__header {
                background-color: transparent!important;
                color: #fafafa!important;
                border-bottom: 1px solid transparent!important;
            }
            .el-collapse-item__wrap {
                background-color: transparent!important;
                border-bottom: 1px solid transparent!important;
            }
            .el-collapse-item__arrow {
                margin: 0 8px!important;
            }
        }
    }

    .knowledge-upload{
        /* border: 1px solid #3d4455;
         padding: 20px;
         border-radius: 4px;*/
    }
    .knowledgeEnhance /deep/ .knowledge-upload-box .knowledge-upload .knowledge-file{
        background-color: transparent!important;
    }
    .knowledgeEnhance .knowledge-upload-box .knowledge-upload .knowledge-file .file__item{
        color: #ccc!important;
    }
</style>
