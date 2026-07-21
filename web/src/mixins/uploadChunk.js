import {
  uploadChunks,
  mergeChunks,
  clearChunks,
  uploadChunksOpen,
  mergeChunksOpen,
  clearChunksOpen,
} from '@/api/chunkFile';
import axios from 'axios';
import { i18n } from '@/lang';
export default {
  data() {
    return {
      isStop: false, // 判断是否取消请求
      fileList: [], //文件列表
      fileIndex: 0, //文件索引
      isChunk: true, //判断是否是切片上传
      isExpire: false, //合并接口是否添加isExpired参数，用来判断minio存储文件是否过期
      // maxSizeBytes: 20 * 1024 * 1024,//可切片大小
      maxSizeBytes: 0, //可切片大小
      chunkSize: 4 * 1024 * 1024, //切片大小1MB
      file: null, //当前文件
      totalChunks: 0, //所有切片数
      uploadedChunks: 0,
      MAX_CONCURRENT: 5, //最大并发数
      chunks: [], //所有切片
      nextChunkIndex: 0, // 下一个要处理的块索引
      uploadQueue: [], // 当前正在处理的请求队列
      failChunk: [], //上传失败切片
      cancelSources: [], // 存储每个请求的取消令牌源
      resList: [], //记录返回成功的文件name
      uuid: '', //生成当前文件的uuid
      isOpenUrl: false, // 是否使用 Open 系列接口
      isRetryingFailedChunks: false,
      activeChunkUploads: 0,
    };
  },
  created() {
    // 监听页面刷新或关闭事件
    window.addEventListener('beforeunload', this.cancelAllRequests);
  },
  beforeDestroy() {
    // 移除事件监听器
    window.removeEventListener('beforeunload', this.cancelAllRequests);
    // 确保在组件销毁时取消所有请求
    this.cancelAllRequests();
  },
  methods: {
    async startUpload(fileIndex = 0, isOpenUrl = false) {
      //开始上传切片
      this.isOpenUrl = isOpenUrl;
      this.isStop = false;
      this.fileIndex = fileIndex;
      this.file = this.fileList[this.fileIndex];
      this.uploadedChunks = 0;
      this.nextChunkIndex = 0;
      this.uploadQueue = []; // 初始化队列
      this.failChunk = [];
      this.isChunk = true;
      this.uuid = this.$guid();
      //判断是否需要切片
      if (this.file.size < this.maxSizeBytes) {
        this.isChunk = false;
        this.uploadFile();
        this.fileList[fileIndex].percentage = 100;
        return;
      }
      //获取切片
      this.chunks = this.createFileChunks(this.file);
      // 启动初始的MAX_CONCURRENT个请求
      for (
        let i = 0;
        i < Math.min(this.MAX_CONCURRENT, this.chunks.length);
        i++
      ) {
        this.processNextChunk();
      }
    },
    createFileChunks(file) {
      //创建切片
      this.totalChunks = Math.ceil(file.size / this.chunkSize);
      const chunks = [];
      let start = 0;
      while (start < file.size) {
        const chunkIndex = chunks.length;
        const groupNumber = Math.floor(chunkIndex / this.MAX_CONCURRENT) + 1; // 计算批次号
        chunks.push({
          index: chunks.length,
          group: groupNumber,
          chunk: file.raw.slice(start, start + this.chunkSize),
        });
        start += this.chunkSize;
      }
      return chunks;
    },
    async processNextChunk() {
      if (this.isStop) return;
      //如果当前切片文件已经上传完停止上传
      if (this.nextChunkIndex >= this.chunks.length) {
        if (this.activeChunkUploads > 0) return;

        if (this.failChunk.length !== 0 && !this.isRetryingFailedChunks) {
          this.isRetryingFailedChunks = true;
          await this.resetUpload();
          this.isRetryingFailedChunks = false;

          if (this.failChunk.length !== 0 && this.handleUploadFailure) {
            this.handleUploadFailure();
          }
        }
        return;
      }

      const chunk = this.chunks[this.nextChunkIndex++];
      this.activeChunkUploads++;
      const uploadPromise = this.uploadChunk(chunk)
        .then(() => {
          if (this.isStop) return;
          this.processNextChunk(); // 递归调用以处理下一个块
        })
        .catch(error => {
          //网络问题，参数问题导致的失败
          if (this.isStop) return;
          this.failChunk.push(chunk);
          this.processNextChunk();
        })
        .finally(() => {
          this.activeChunkUploads--;
          if (!this.isStop) this.processNextChunk();
        });

      this.uploadQueue.push(uploadPromise);
      // 等待队列中的任意一个请求完成,忽略已完成或失败的请求错误
      await Promise.race(
        this.uploadQueue.map(promise => promise.catch(() => {})),
      );
      // 移除已完成的请求
      this.uploadQueue = this.uploadQueue.filter(
        promise => !promise.isFulfilled,
      );
    },
    clearFile(index) {
      //清除文件
      // let formData = new FormData();
      const file = this.fileList[index];
      const hash = `${this.uuid}.${file.name.split('.').pop()}`;
      // formData.append('chunkName', hash);//前端uuid+文件后缀,标识一次上传批次
      // formData.append('version', 0);//前端uuid+文件后缀,标识一次上传批次
      const formData = {
        chunkName: hash,
        version: 0,
      };
      const clearRequest = this.isOpenUrl
        ? clearChunksOpen(formData)
        : clearChunks(formData);
      clearRequest.then(res => {
        if (res.code === 0 && res.data.status === 1) {
          this.$message.success(i18n.t('fileChunk.fileClear'));
          this.fileList.splice(index, 1);
          this.$refs['upload'].updateFile(index);
          if (this.fileList.length > 0) {
            this.startUpload(index);
          }
        }
      });
    },
    async uploadChunk(chunkData) {
      //上传切片
      const source = axios.CancelToken.source(); //创建一个取消令牌
      this.cancelSources.push(source);
      const config = source.token;

      let formData = new FormData();
      const hash = `${this.uuid}.${this.file.name.split('.').pop()}`;
      formData.append('chunkName', hash); //前端uuid+文件后缀,标识一次上传批次
      formData.append('fileName', this.file.name); //原始文件名称
      formData.append('files', chunkData.chunk); //文件
      formData.append('concurrentTotal', this.MAX_CONCURRENT);
      formData.append('chunkSize', chunkData.chunk.size);
      formData.append('concurrentNo', chunkData.group); //并发上传线程的序号
      formData.append('sequence', chunkData.index + 1); //拆分小文件的序号
      formData.append('version', 0);
      try {
        const res = this.isOpenUrl
          ? await uploadChunksOpen(formData, config)
          : await uploadChunks(formData, config); // 传递 AbortSigna
        if (res.code === 0 && res.data.status === 1) {
          this.uploadedChunks++; //用来判断执行成功的切片的数量
          if (
            Math.floor((this.uploadedChunks * 100) / this.totalChunks) >= 100
          ) {
            this.fileList[this.fileIndex].percentage = 99;
          } else {
            this.fileList[this.fileIndex].percentage = Math.floor(
              (this.uploadedChunks * 100) / this.totalChunks,
            );
          }
          if (this.uploadedChunks === this.totalChunks) {
            //如果都已上传完成，合并文件
            await this.mergeChunks();
          }

          //完成请求，cancelSources删除一个token
          const index = this.cancelSources.indexOf(source);
          if (index !== -1) {
            this.cancelSources.splice(index, 1);
          }
          source.cancel();
        } else {
          throw new Error(`Upload failed with status ${res.data.status}`);
        }
      } catch (error) {
        if (this.isStop || (error && error.name === 'CanceledError')) {
          return;
        }
        throw error;
      }
    },
    async resetUpload() {
      //失败切片重试
      const failedChunksCopy = [...this.failChunk];
      this.failChunk = [];
      for (const chunk of failedChunksCopy) {
        if (this.isStop) break;
        try {
          await this.uploadChunk(chunk);
        } catch (error) {
          //点击续传按钮续传失败列表里面的切片
          this.failChunk.push(chunk);
          //重试失败显示重试、续传按钮
          this.fileList[this.fileIndex]['showRetry'] = 'true';
          this.fileList[this.fileIndex]['showResume'] = 'true';
        }
      }
    },
    async mergeChunks() {
      if (this.isStop) return;
      //合并切片
      try {
        let file_size = this.fileList[this.fileIndex]['size'];
        const formData = {
          chunkName: `${this.uuid}.${this.file.name.split('.').pop()}`,
          chunkTotal: this.totalChunks,
          fileName: this.file.name,
          fileSize: this.file.size,
          isExpired: false,
        };

        const mergeRequest = this.isOpenUrl
          ? mergeChunksOpen(formData)
          : mergeChunks(formData);
        await mergeRequest.then(res => {
          if (res.code === 0) {
            this.$message.success(
              `${this.file.name}` + i18n.t('fileChunk.uploadFinish'),
            );
            this.fileList[this.fileIndex].percentage = 100;
            this.fileList[this.fileIndex]['progressStatus'] = 'success';
            this.fileList[this.fileIndex]['showRetry'] = 'false';
            this.fileList[this.fileIndex]['showResume'] = 'false';
            this.fileListSize += (file_size / 1024 / 1024).toFixed(5);
            this.resList.push({ name: res.data.fileName });
            //接片合并完之后走上传接口
            this.uploadFile(
              res.data.fileName,
              this.file.name,
              res.data.filePath,
            );
          } else {
            this.$message.error(
              `${this.file.name}` + i18n.t('fileChunk.uploadFail'),
            );
            if (this.handleUploadFailure) {
              this.handleUploadFailure({ showMessage: false });
            } else {
              this.fileList[this.fileIndex]['showRemerge'] = 'true';
              this.uploadNextTask && this.uploadNextTask();
            }
          }
        });
      } catch (error) {
        this.$message.error(
          `${this.file.name}` + i18n.t('fileChunk.uploadFail'),
        );
        if (this.handleUploadFailure) {
          this.handleUploadFailure({ showMessage: false });
        } else {
          this.fileList[this.fileIndex]['showRemerge'] = 'true';
        }
      }
    },
    cancelAllRequests() {
      //取消所有请求
      this.isStop = true;
      if (this.cancelSources.length > 0) {
        for (let i = 0; i < this.cancelSources.length; i++) {
          try {
            this.cancelSources[i].cancel();
          } catch (error) {}
        }
      }
      this.cancelSources = [];
      this.failChunk = [];
      this.uploadQueue = [];
    },
    cancelAndRestartNextRequests() {
      this.cancelAllRequests();
      if (this.fileList.length < 1) {
        this.fileIndex = 0;
        this.file = null;
      } else {
        this.fileIndex = Math.max(0, this.fileIndex - 1);
        this.file = this.fileList[this.fileIndex] || null;
        this.startUpload(this.fileIndex);
      }
    },
  },
};
