<!--问答文件上传-->
<template>
  <div class="fileUpload">
    <!-- 上传触发按钮 -->
    <slot :openDialog="openDialog">
      <el-button
        class="chat-upload-btn"
        icon="el-icon-circle-plus-outline"
        circle
        plain
        @click="openDialog"
      ></el-button>
    </slot>

    <el-dialog
      custom-class="upload-dialog"
      :visible.sync="dialogVisible"
      width="800px"
      append-to-body
      :before-close="handleClose"
    >
      <div
        v-loading="loading"
        element-loading-background="rgba(255, 255, 255, 0.5)"
      >
        <div class="dialog-body">
          <p class="upload-title">{{ $t('common.fileUpload.uploadFile') }}</p>
          <el-upload
            :class="['upload-box']"
            drag
            action=""
            :show-file-list="false"
            :auto-upload="false"
            :limit="uploadLimit"
            :accept="tipsArr"
            :file-list="fileList"
            :on-change="uploadOnChange"
            :on-exceed="uploadOnExceed"
          >
            <div v-if="fileUrl" class="echo-img-box">
              <div class="echo-img">
                <div class="type-img-container">
                  <el-button
                    v-show="canScroll"
                    icon="el-icon-arrow-left "
                    @click="prev($event)"
                    circle
                    class="scroll-btn left"
                    size="mini"
                    type="primary"
                  ></el-button>
                  <div
                    class="type-img"
                    ref="imgList"
                    :style="{ justifyContent: !canScroll ? 'center' : 'unset' }"
                  >
                    <div
                      v-for="(f, idx) in fileList"
                      :key="f.uid || idx"
                      class="type-img-item"
                    >
                      <img v-if="isImageUploadFile(f)" :src="f.fileUrl" />
                      <audio
                        v-else-if="isAudioUploadFile(f)"
                        controls
                        class="type-audio"
                      >
                        <source :src="f.fileUrl" type="video/mp3" />
                        <source :src="f.fileUrl" type="audio/ogg" />
                        <source :src="f.fileUrl" type="audio/mpeg" />
                        {{ $t('common.fileUpload.audioTips') }}
                      </audio>
                      <div v-else class="docFile">
                        <FileIcon
                          :type="getFileIconType(f)"
                          size="60px"
                          class="docIcon"
                        />
                      </div>
                      <p class="type-img-info">
                        <el-tooltip
                          class="item"
                          effect="dark"
                          :content="f.name"
                          placement="top-start"
                        >
                          <span>
                            {{
                              f.name.length > 6
                                ? f.name.slice(0, 6) + '...'
                                : f.name
                            }}
                          </span>
                        </el-tooltip>
                        <span>[{{ getFileSizeDisplay(f.size) }}]</span>
                      </p>
                    </div>
                  </div>
                  <el-button
                    v-show="canScroll"
                    icon="el-icon-arrow-right"
                    @click="next($event)"
                    circle
                    class="scroll-btn right"
                    size="mini"
                    type="primary"
                  ></el-button>
                </div>
              </div>
              <div class="tips">
                <el-progress
                  :percentage="file.percentage"
                  v-if="file.percentage !== 100"
                  :status="file.progressStatus"
                  max="100"
                  style="width: 360px; margin: 0 auto"
                ></el-progress>
                <p
                  v-if="visibleContinueUploadTips"
                  class="continue-upload-tips"
                >
                  {{ $t('common.fileUpload.clickBlankContinueUpload') }}
                </p>
                <template v-if="hasUploadLimitTips">
                  <p>{{ uploadLimitTips }}</p>
                </template>
              </div>
            </div>
            <div v-else>
              <i class="el-icon-upload"></i>
              <p>
                {{
                  $t('common.fileUpload.uploadText') +
                  $t('common.fileUpload.uploadClick')
                }}
              </p>
              <div class="tips">
                <p v-if="visibleSizeLimitTips">{{ sizeLimitTips }}</p>
                <p>
                  {{ $t('common.fileUpload.typeFileTip1') }}
                  <span>{{ tipsArr }}</span>
                  {{ $t('common.fileUpload.typeFileTip') }}
                </p>
              </div>
            </div>
          </el-upload>
        </div>
        <div class="dialog-footer">
          <el-button
            type="primary"
            :disabled="!fileUrl || !allFilesUploaded"
            @click="doBatchUpload"
          >
            {{ $t('common.fileUpload.submitBtn') }}
          </el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import uploadChunk from '@/mixins/uploadChunk';
import FileIcon from '@/components/FileIcon.vue';
import { getFileIconType } from '@/utils/util';

export default {
  components: { FileIcon },
  props: {
    fileTypeArr: {
      type: Array,
      required: false,
      default: () => [],
    },
    type: { type: String },
    maxImageSize: {
      type: [Number, String],
      required: false,
      default: null,
    },
    // 文件大小限制（null,undefined,-1，0表示不限制）
    maxFileSize: {
      type: [Number, String],
      required: false,
      default: null,
    },
    // 图片数量限制（-1表示不限制）
    maxPicNum: {
      type: Number,
      required: false,
      default: -1, // -1不限制
    },
    maxFileNum: {
      type: Number,
      required: false,
      default: -1, // -1不限制
    },
    confirmedFileList: {
      type: Array,
      required: false,
      default: () => [],
    },
    confirmedFileIdList: {
      type: Array,
      required: false,
      default: () => [],
    },
  },
  mixins: [uploadChunk],
  data() {
    return {
      canScroll: false,
      fileIdList: [],
      fileList: [],
      fileType: '',
      loading: false,
      isUploading: false,
      dialogVisible: false,
      fileUrl: '',
      tipsArr: '',
      tipsObj: {
        'image/*': ['jpg', 'jpeg', 'png'],
        'audio/*': ['wav', 'mp3'],
        'doc/*': ['txt', 'csv', 'xlsx', 'docx', 'html', 'pptx', 'pdf'],
      },
      fileInfo: [],
      lastFileType: '',
      imgUrl: '',
    };
  },
  watch: {
    fileTypeArr: {
      handler(val, oldVal) {
        this.setFileType(val);
      },
      immediate: true,
    },
  },
  computed: {
    normalizedMaxPicNum() {
      const maxPicNum = Number(this.maxPicNum);
      return Number.isFinite(maxPicNum) ? maxPicNum : 3;
    },
    hasImageLimit() {
      return this.normalizedMaxPicNum >= 0;
    },
    normalizedMaxFileNum() {
      const maxFileNum = Number(this.maxFileNum);
      return Number.isFinite(maxFileNum) ? maxFileNum : -1;
    },
    hasFileLimit() {
      return this.normalizedMaxFileNum >= 0;
    },
    displayMaxPicNum() {
      return this.effectiveImageLimit;
    },
    displayMaxFileNum() {
      return this.normalizedMaxFileNum;
    },
    supportsImageUpload() {
      return this.fileTypeArr.includes('image/*');
    },
    isImageOnlyUpload() {
      return this.supportsImageUpload && this.fileTypeArr.length === 1;
    },
    hasUploadLimitTips() {
      return (
        this.hasFileLimit ||
        (this.supportsImageUpload && this.hasEffectiveImageLimit)
      );
    },
    uploadLimitTips() {
      if (
        this.hasFileLimit &&
        this.supportsImageUpload &&
        this.hasEffectiveImageLimit &&
        !this.isImageOnlyUpload
      ) {
        return this.$t('app.uploadFileAndImgLimitTips', {
          fileNum: this.displayMaxFileNum,
          imageNum: this.displayMaxPicNum,
        });
      }
      if (this.supportsImageUpload && this.hasEffectiveImageLimit) {
        return this.$t('app.imgLimitOnly', { num: this.displayMaxPicNum });
      }
      if (this.hasFileLimit) {
        return this.$t('app.uploadFileLimitTips', {
          num: this.displayMaxFileNum,
        });
      }
      return '';
    },
    effectiveImageLimit() {
      const limits = [];
      if (this.hasImageLimit) limits.push(this.normalizedMaxPicNum);
      if (this.hasFileLimit) limits.push(this.normalizedMaxFileNum);
      return limits.length ? Math.min(...limits) : -1;
    },
    hasEffectiveImageLimit() {
      return this.effectiveImageLimit >= 0;
    },
    visibleContinueUploadTips() {
      if (!this.fileList.length) return false;
      if (!this.hasFileLimit) return true;
      return this.fileList.length < this.normalizedMaxFileNum;
    },
    uploadLimit() {
      return undefined;
    },
    maxImageSizeMB() {
      const maxSize = Number(this.maxImageSize);
      return maxSize > 0 ? maxSize : 0;
    },
    maxImageSizeBytes() {
      return this.maxImageSizeMB ? this.maxImageSizeMB * 1024 * 1024 : 0;
    },
    maxFileSizeMB() {
      const maxSize = Number(this.maxFileSize);
      return maxSize > 0 ? maxSize : 0;
    },
    maxFileSizeBytes() {
      return this.maxFileSizeMB ? this.maxFileSizeMB * 1024 * 1024 : 0;
    },
    visibleImageSizeLimit() {
      return this.maxImageSizeMB > 0;
    },
    visibleFileSizeLimit() {
      return this.maxFileSizeMB > 0;
    },
    visibleSizeLimitTips() {
      return this.visibleImageSizeLimit || this.visibleFileSizeLimit;
    },
    sizeLimitTips() {
      const tips = [];
      if (this.visibleImageSizeLimit) {
        tips.push(
          this.$t('app.imageSizeModelLimit', { maxSize: this.maxImageSizeMB }),
        );
      }
      if (this.visibleFileSizeLimit) {
        tips.push(
          this.$t('app.uploadFileSizeLimitTips', {
            maxSize: this.maxFileSizeMB,
          }),
        );
      }
      return tips.join('，');
    },
    allFilesUploaded() {
      return (
        this.fileList.length > 0 &&
        this.fileList.every(file => file.percentage === 100)
      );
    },
  },
  methods: {
    getFileIconType,
    checkScrollable() {
      this.$nextTick(() => {
        const container = this.$refs.imgList;
        if (container) {
          this.canScroll = container.scrollWidth > container.clientWidth;
        }
      });
    },
    prev(e) {
      e.stopPropagation();
      this.$refs.imgList.scrollBy({
        left: -200,
        behavior: 'smooth',
      });
    },
    next(e) {
      e.stopPropagation();
      this.$refs.imgList.scrollBy({
        left: 200,
        behavior: 'smooth',
      });
    },
    setFileType(fileTypeArr) {
      if (fileTypeArr.length) {
        this.tipsArr = '';
        let tips_arr = [];
        fileTypeArr.forEach(item => {
          const extensions = (this.tipsObj[item] || [item]).map(
            ext => '.' + ext,
          );
          tips_arr = tips_arr.concat(extensions);
        });
        this.tipsArr = tips_arr.join(', ');
      }
    },
    openDialog() {
      this.syncConfirmedFiles();
      this.dialogVisible = true;
    },
    syncConfirmedFiles() {
      this.fileList = (this.confirmedFileList || []).map(file =>
        this.normalizeUploadFile({
          ...file,
          percentage: file.percentage !== undefined ? file.percentage : 100,
          progressStatus: file.progressStatus || 'success',
          uploadStatus: file.uploadStatus || 'success',
          uploaded: file.uploaded !== undefined ? file.uploaded : true,
        }),
      );
      this.fileInfo = (this.confirmedFileIdList || []).map(item => ({
        ...item,
      }));
      this.fileIdList = this.fileInfo.map(item => ({ ...item }));
      const lastFile = this.fileList[this.fileList.length - 1];
      this.fileType = (lastFile && lastFile.fileType) || '';
      this.fileUrl = (lastFile && lastFile.fileUrl) || '';
      this.imgUrl = (lastFile && lastFile.imgUrl) || '';
      this.isUploading = false;
      this.checkScrollable();
    },
    clearFile() {
      this.fileIdList = [];
      this.fileList = [];
      this.fileType = '';
      this.fileUrl = '';
      this.imgUrl = '';
      this.fileInfo = [];
      this.isUploading = false;
      this.canScroll = false;
    },
    handleClose() {
      this.clearFile();
      this.dialogVisible = false;
    },
    showFileLimitMessage() {
      this.$message.warning(
        this.$t('app.uploadFileLimitTips', { num: this.displayMaxFileNum }),
      );
    },
    showImageLimitMessage() {
      this.$message.warning(
        this.$t('app.uploadImgTips', { num: this.displayMaxPicNum }),
      );
    },
    showEffectiveImageLimitMessage() {
      if (
        this.hasFileLimit &&
        (!this.hasImageLimit ||
          this.normalizedMaxFileNum <= this.normalizedMaxPicNum)
      ) {
        this.showFileLimitMessage();
        return;
      }
      this.showImageLimitMessage();
    },
    uploadOnExceed() {
      this.showEffectiveImageLimitMessage();
    },
    uploadOnChange(file, fileList) {
      const filename = file.name;
      const nextFileType = this.getFileType(filename);

      const validateResult = this.validateUploadFile(file, nextFileType);
      if (!validateResult.valid) {
        this.showUploadValidateMessage(validateResult);
        this.removeUploadFile(file, fileList);
        return;
      }

      const nextFileList = fileList.map(item => {
        const existingFile = this.fileList.find(
          oldFile => oldFile.uid === item.uid,
        );
        if (existingFile) {
          ['uploaded', 'uploadStatus', 'percentage', 'progressStatus'].forEach(
            key => {
              if (existingFile[key] !== undefined)
                this.$set(item, key, existingFile[key]);
            },
          );
          ['fileUrl', 'imgUrl', 'fileType'].forEach(key => {
            if (existingFile[key]) this.$set(item, key, existingFile[key]);
          });
        }
        return this.normalizeUploadFile(item);
      });

      if (
        this.hasFileLimit &&
        nextFileList.length > this.normalizedMaxFileNum
      ) {
        this.showFileLimitMessage();
        this.removeUploadFile(file, fileList);
        return;
      }

      if (
        this.hasEffectiveImageLimit &&
        nextFileList.filter(item => this.isImageUploadFile(item)).length >
          this.effectiveImageLimit
      ) {
        this.showEffectiveImageLimitMessage();
        this.removeUploadFile(file, fileList);
        return;
      }

      this.fileList = nextFileList;
      const currentFile = this.normalizeUploadFile(file);
      this.fileType = currentFile.fileType;
      this.imgUrl = currentFile.imgUrl || '';
      this.fileUrl = currentFile.fileUrl;
      this.checkScrollable();

      this.triggerNextUpload();
    },
    removeUploadFile(file, fileList) {
      const index = fileList.indexOf(file);
      if (index > -1) {
        fileList.splice(index, 1);
      }
    },
    createUploadFile(rawFile) {
      const uid = rawFile.uid || this.$guid();
      rawFile.uid = uid;
      return this.normalizeUploadFile({
        name: rawFile.name,
        size: rawFile.size,
        uid,
        raw: rawFile,
        percentage: 0,
        progressStatus: 'active',
        uploadStatus: 'pending',
        uploaded: false,
      });
    },
    normalizeUploadFile(file) {
      const fileType = file.fileType || this.getFileType(file.name);
      this.$set(file, 'fileType', fileType);
      if (!file.fileUrl && (file.raw || file.url)) {
        this.$set(
          file,
          'fileUrl',
          file.raw ? URL.createObjectURL(file.raw) : file.url,
        );
      }
      if (fileType === 'image/*' && !file.imgUrl) {
        this.$set(file, 'imgUrl', file.fileUrl);
      }
      return file;
    },
    isImageUploadFile(file) {
      return (
        (file?.fileType || this.getFileType(file?.name || '')) === 'image/*'
      );
    },
    isAudioUploadFile(file) {
      return (
        (file?.fileType || this.getFileType(file?.name || '')) === 'audio/*'
      );
    },
    getFileSizeDisplay(fileSize) {
      return fileSize > 1024
        ? `${(fileSize / (1024 * 1024)).toFixed(2)} MB`
        : `${fileSize} bytes`;
    },
    validateUploadFile(file, fileType) {
      const filename = (file && file.name) || '';
      const acceptedExtensions = this.tipsArr
        .split(',')
        .map(ext => ext.trim().toLowerCase())
        .filter(Boolean);
      const isAccepted = acceptedExtensions.some(ext =>
        filename.toLowerCase().endsWith(ext),
      );
      if (!isAccepted) {
        return { valid: false, type: 'fileType' };
      }

      const nextFileType = fileType || this.getFileType(filename);
      if (nextFileType === 'image/*' && this.isImageOverSize(file)) {
        return { valid: false, type: 'imageSize' };
      }
      if (nextFileType !== 'image/*' && this.isFileOverSize(file)) {
        return { valid: false, type: 'fileSize' };
      }
      return { valid: true, type: nextFileType };
    },
    showUploadValidateMessage(validateResult) {
      if (validateResult.type === 'imageSize') {
        this.$message.warning(
          this.$t('knowledgeManage.multiKnowledgeDatabase.imageSizeLimit', {
            maxSize: this.maxImageSizeMB,
          }),
        );
        return;
      }

      if (validateResult.type === 'fileSize') {
        this.$message.warning(
          this.$t('app.uploadFileSizeLimitTips', {
            maxSize: this.maxFileSizeMB,
          }),
        );
        return;
      }

      this.$message.warning(
        this.$t('common.fileUpload.typeFileTip1') +
          this.tipsArr +
          this.$t('common.fileUpload.typeFileTip'),
      );
    },
    getFileType(filename) {
      const fileTypeLower = (filename.split('.').pop() || '').toLowerCase();
      if (this.tipsObj['image/*'].includes(fileTypeLower)) return 'image/*';
      if (this.tipsObj['audio/*'].includes(fileTypeLower)) return 'audio/*';
      if ([...this.tipsObj['doc/*'], 'md'].includes(fileTypeLower)) {
        return 'doc/*';
      }
      return '';
    },
    isImageOverSize(file) {
      return (
        this.maxImageSizeBytes && file && file.size > this.maxImageSizeBytes
      );
    },
    isFileOverSize(file) {
      return this.maxFileSizeBytes && file && file.size > this.maxFileSizeBytes;
    },
    uploadFile(fileName, oldFileName, fiePath) {
      // 文件上传完成后，释放队列锁并继续调度下一个 pending 文件
      const currentFile = this.fileList[this.fileIndex] || {};
      if (!currentFile.uid) {
        this.isUploading = false;
        this.triggerNextUpload();
        return;
      }
      this.lastFileType = currentFile.fileType || this.fileType;
      this.$set(currentFile, 'uploadStatus', 'success');
      this.$set(currentFile, 'uploaded', true);
      this.$set(currentFile, 'percentage', 100);
      const fileInfoItem = {
        uid: currentFile.uid,
        fileType: currentFile.fileType,
        fileName,
        oldFileName,
        fileSize: currentFile.size,
        fileUrl: fiePath,
      };
      if (currentFile.fileType === 'image/*') {
        fileInfoItem.imgUrl = currentFile.imgUrl || currentFile.fileUrl;
      }
      const index = this.fileInfo.findIndex(
        item => item.uid === currentFile.uid,
      );
      if (index > -1) {
        this.$set(this.fileInfo, index, fileInfoItem);
      } else {
        this.fileInfo.push(fileInfoItem);
      }
      this.isUploading = false;
      this.triggerNextUpload();
    },
    isPendingUpload(file) {
      if (!file) return false;
      if (file.uploadStatus) return file.uploadStatus === 'pending';
      return file.percentage !== 100;
    },
    triggerNextUpload() {
      if (this.isUploading || !this.fileList || this.fileList.length === 0) {
        return;
      }
      const nextPendingIndex = this.fileList.findIndex(file =>
        this.isPendingUpload(file),
      );
      if (nextPendingIndex === -1) return;

      this.maxSizeBytes = 0;
      this.isExpire = true;
      this.isUploading = true;
      this.$set(this.fileList[nextPendingIndex], 'uploadStatus', 'uploading');
      this.$set(this.fileList[nextPendingIndex], 'uploaded', true);
      this.startUpload(nextPendingIndex, this.type === 'webChat');
    },
    doBatchUpload() {
      const sortedFileInfo = this.fileList
        .map(file => this.fileInfo.find(item => item.uid === file.uid))
        .filter(Boolean);
      this.$emit('setFileId', sortedFileInfo);
      this.$emit('setFile', this.fileList);
      this.handleClose();
    },
    getFileIdList() {
      return this.fileIdList;
    },
  },
};
</script>

<style lang="scss" scoped>
.upload-dialog {
  .dialog-body {
    padding: 0 20px;
    .upload-title {
      text-align: center;
      font-size: 18px;
      margin-bottom: 20px;
    }
    .upload-box {
      height: 230px;
      width: 100% !important;
      background-color: #fff;
      ::v-deep .el-upload-dragger {
        height: 100%;
        width: 100%;
      }
      .el-upload-dragger {
        .el-icon-upload {
          margin: 46px 0 10px 0 !important;
          font-size: 32px !important;
          line-height: 36px !important;
          color: $color;
        }
        .el-upload__text {
          margin-top: -10px;
        }
      }
    }

    .echo-img-box {
      background-color: transparent !important;
      .echo-img {
        .type-img-container {
          width: 100%;
          position: relative;
          .scroll-btn {
            position: absolute;
            top: 50%;
            transform: translateY(-32px);
            &.left {
              left: 5px;
            }
            &.right {
              right: 5px;
            }
          }
          .type-img {
            display: flex;
            gap: 10px;
            width: 100%;
            overflow-x: hidden;
            scroll-behavior: smooth;
            .type-img-item {
              width: auto !important;
              flex-shrink: 0;
              margin-bottom: 10px;
            }
            .type-img-info {
              display: flex;
              gap: 5px;
              justify-content: center;
              span {
                color: $color;
              }
            }
          }
        }
        img,
        video {
          width: auto;
          height: 80px;
          margin: 10px auto;
          border-radius: 4px;
          background-color: transparent;
        }
        audio {
          width: 300px;
          height: 54px;
          margin: 50px auto;
        }
      }
      .docFile {
        height: 100px;
        display: flex;
        align-items: center;
        justify-content: center;
        .docIcon {
          flex: 0 0 auto;
        }
      }
    }
    .tips {
      position: absolute;
      bottom: 16px;
      left: 0;
      right: 0;
      padding: 0 24px;
      line-height: 20px;
      p {
        margin: 4px 0 0;
        color: #9d8d8d !important;
      }
    }
  }
  .dialog-footer {
    text-align: center;
    margin: 30px 0 20px 0;
  }
}

.chat-upload-btn {
  padding: 8px;
  color: rgba(15, 21, 40, 0.82);
  border: none;
  &:hover {
    background-color: rgba(87, 104, 161, 0.08) !important;
    color: rgba(15, 21, 40, 0.82);
  }
  ::v-deep i {
    font-size: 16px;
  }
}
</style>
