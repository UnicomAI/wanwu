import { uploadChunks,mergeChunks,clearChunks } from '@/api/chunkFile'
import axios from "axios";
import {i18n} from "@/lang"
export default {
    data() {
        return {
            isStop: false, // CheckYesNoCancelRequest
            fileList:[],//FileList
            fileIndex:0,//FileIndex
            isChunk:true,//CheckYesNoYes切片Upload
            isExpire:false,//合并InterfaceYesNoAddisExpiredParameter，用来Checkminio存储FileYesNo 期
            // maxSizeBytes: 20 * 1024 * 1024,//可切片size
            maxSizeBytes:0,//可切片size
            chunkSize: 4 * 1024 * 1024,//切片size1MB
            file: null,//currentFile
            totalChunks: 0,//所Has切片数
            uploadedChunks: 0,
            MAX_CONCURRENT:5,//MaxConcurrency数
            chunks:[],//所Has切片
            nextChunkIndex: 0, // 下一toProcess of 块Index
            uploadQueue: [], // currentin progressProcess of RequestQueue
            failChunk:[],//UploadFailed切片
            cancelSources: [], // 存储eachRequest of Cancel令牌源
            resList:[],//RecordBackSuccess of Filename
            uuid:'',//GeneratecurrentFile of uuid
        }
    },
    created() {
      // ListenPageRefreshORCloseEvent
      window.addEventListener('beforeunload', this.cancelAllRequests);
    },
    beforeDestroy() {
      // RemoveEventListen器
      window.removeEventListener('beforeunload', this.cancelAllRequests);
      // ensure在ComponentDestroy when Cancel all requests
      this.cancelAllRequests();
    },
    methods: {
        async startUpload(fileIndex=0){//startUpload切片
          this.isStop = false;
          this.fileIndex = fileIndex;
          this.file = this.fileList[this.fileIndex];
          this.uploadedChunks = 0;
          this.nextChunkIndex = 0;
          this.uploadQueue = []; // InitQueue
          this.failChunk = [];
          this.isChunk = true;
          this.uuid = this.$guid();
          //CheckYesNoNeed切片
          if(this.file.size < this.maxSizeBytes){
            this.isChunk = false;
            this.uploadFile()
            return
          }
          //Get切片
          this.chunks = this.createFileChunks(this.file);
          // Start初始 of MAX_CONCURRENTRequest
          for (let i = 0; i < Math.min(this.MAX_CONCURRENT, this.chunks.length); i++) {
            this.processNextChunk();
          }
        },
        createFileChunks(file) {//Create切片
            this.totalChunks  = Math.ceil(file.size / this.chunkSize);
            const chunks = [];
            let start = 0;
            while (start < file.size) {
              const chunkIndex = chunks.length;
              const groupNumber = Math.floor(chunkIndex / this.MAX_CONCURRENT) + 1; // 计算批次号
              chunks.push({
                index: chunks.length,
                group:groupNumber,
                chunk: file.raw.slice(start, start + this.chunkSize)
              });
              start += this.chunkSize;
            }
            return chunks;
        },
        async processNextChunk() {//进row下一切片Upload
          //Ifcurrent切片Filealready经Upload完StopUpload
          if (this.nextChunkIndex >= this.chunks.length){
            //所HasExecute完之after，Failed切片进rowRetry
            if(this.failChunk.length !== 0){
                this.resetUpload()
            }
            return
          } 

          const chunk = this.chunks[this.nextChunkIndex++];
          const uploadPromise = this.uploadChunk(chunk).then(() => {
              this.processNextChunk(); // RecurseCall以Process下一块
          }).catch(error =>{
            //Network问题，Parameter问题导致 of Failed
            if(this.isStop) return;
            this.failChunk.push(chunk);
            this.processNextChunk();
          });
          
          this.uploadQueue.push(uploadPromise);
          // AwaitQueuein of any一RequestComplete,ignorealreadyCompleteORFailed of RequestError
          await Promise.race(this.uploadQueue.map(promise => promise.catch(() => {}))); 
          // RemovealreadyComplete of Request
          this.uploadQueue = this.uploadQueue.filter(promise => !promise.isFulfilled);
        },
        clearFile(index){//ClearFile
          // let formData = new FormData();
          const file = this.fileList[index]
          const hash = `${this.uuid}.${file.name.split(".").pop()}`
          // formData.append('chunkName', hash);//前端uuid+Fileafter缀,标识一次Upload批次
          // formData.append('version', 0);//前端uuid+Fileafter缀,标识一次Upload批次
          const formData = {
            chunkName:hash,
            version:0
          }
          clearChunks(formData).then(res =>{
            if(res.code === 0 && res.data.status === 1){
                this.$message.success(i18n.t('fileChunk.fileClear'))
                this.fileList.splice(index, 1);
                this.$refs["upload"].updateFile(index);
                if(this.fileList.length > 0){
                  this.startUpload(index);
                }
            }
          })
        },
        async uploadChunk(chunkData) {//Upload切片
              const source = axios.CancelToken.source();//Create一Cancel令牌
              this.cancelSources.push(source);
              const config =  source.token

              let formData = new FormData();
              const hash = `${this.uuid}.${this.file.name.split(".").pop()}`
              formData.append('chunkName', hash);//前端uuid+Fileafter缀,标识一次Upload批次
              formData.append('fileName', this.file.name);//原始FileName
              formData.append('files', chunkData.chunk);//File
              formData.append('concurrentTotal',this.MAX_CONCURRENT);
              formData.append('chunkSize',chunkData.chunk.size);
              formData.append('concurrentNo', chunkData.group);//ConcurrencyUploadThread of No.
              formData.append('sequence', chunkData.index + 1);//拆分小File of No.
              formData.append('version',0);
              try{
                const res = await uploadChunks(formData,config);// 传递 AbortSigna
                if(res.code === 0 && res.data.status === 1){
                  this.uploadedChunks++;//用来CheckExecuteSuccess of 切片 of Count
                  if(Math.floor((this.uploadedChunks*100) / this.totalChunks) >= 100){
                    this.fileList[this.fileIndex].percentage = 99
                  }else{
                    this.fileList[this.fileIndex].percentage = Math.floor((this.uploadedChunks*100) / this.totalChunks);
                  }
                  if(this.uploadedChunks === this.totalChunks){//If都alreadyUploadComplete，合并File
                    await this.mergeChunks()
                  }

                  //CompleteRequest，cancelSourcesDelete一token
                  const index = this.cancelSources.indexOf(source);
                  if(index !== -1){
                    this.cancelSources.splice(index, 1);
                  }
                  source.cancel()

                }else{
                  throw new Error(`Upload failed with status ${res.data.status}`);
                }
              }catch(error){
                throw error;
              }
          },
        async resetUpload(){//Failed切片Retry
          const failedChunksCopy = [...this.failChunk];
          this.failChunk = [];
          for(const chunk of failedChunksCopy){
            try{
              await this.uploadChunk(chunk);
            }catch(error){
              //clickResumeButtonResumeFailedList里面 of 切片
              this.failChunk.push(chunk);
              //RetryFailedShowRetry、ResumeButton
              this.fileList[this.fileIndex]['showRetry'] = 'true';
              this.fileList[this.fileIndex]['showResume'] = 'true';
            }
          }
        },
        async mergeChunks(){//合并切片
          try{
            let file_size =  this.fileList[this.fileIndex]['size'];
            const formData = {
              chunkName:`${this.uuid}.${this.file.name.split(".").pop()}`,
              chunkTotal:this.totalChunks,
              fileName:this.file.name,
              fileSize:this.file.size,
              isExpired:false
            }

            await mergeChunks(formData).then(res =>{
              if(res.code === 0){
                this.$message.success(`${this.file.name}`+i18n.t('fileChunk.uploadFinish'));
                this.fileList[this.fileIndex].percentage = 100;
                this.fileList[this.fileIndex]['progressStatus'] = 'success';
                this.fileList[this.fileIndex]['showRetry'] = 'false';
                this.fileList[this.fileIndex]['showResume'] = 'false';
                this.fileListSize += (file_size/1024/1024).toFixed(5);
                this.resList.push({name:res.data.fileName});
                //接片合并完之after走UploadInterface
                this.uploadFile(res.data.fileName,this.file.name,res.data.filePath)
              }else{
                this.$message.error(`${this.file.name}`+ i18n.t('fileChunk.uploadFail'))
                this.fileList[this.fileIndex]['showRemerge'] = 'true';
              }
            })
          }catch(error){
            this.$message.error(`${this.file.name}`+ i18n.t('fileChunk.uploadFail'))
            this.fileList[this.fileIndex]['showRemerge'] = 'true';
          }
        },
        cancelAllRequests() {//Cancel all requests
          this.isStop = true;
          if (this.cancelSources.length > 0) {
            for (let i = 0; i < this.cancelSources.length; i++) {
              this.cancelSources[i].cancel();
            }
          }
          this.cancelSources = [];
        },
    }
}
