<!--问答输入框-->
<template>
  <div class="rl chat-input-wrapper">
    <div v-if="visibleClearHistory" class="chat-input-wrapper-left">
      <el-tooltip
        class="item"
        effect="dark"
        :content="$t('app.clearChat')"
        placement="top"
      >
        <el-button
          circle
          class="chat-clear-btn"
          :disabled="!hasHistory || sessionStatus === 0"
          @click="handleClearHistory"
        >
          <svg class="chat-clear-icon">
            <use xlink:href="#icon-chatClear" />
          </svg>
        </el-button>
      </el-tooltip>
    </div>
    <div class="editable-box">
      <!-- file preview -->
      <div v-if="fileList.length" class="echo-img-box mixed-file-box">
        <div
          v-for="group in visibleMixedFileGroups"
          :key="group.type"
          :class="['mixed-file-group', 'mixed-file-' + group.type + '-group']"
        >
          <div
            v-for="item in group.files"
            :class="[
              'echo-img-item',
              'mixed-file-item',
              {
                'doc-file-item':
                  !isImageFile(item.file) && !isAudioFile(item.file),
              },
            ]"
            :key="item.file.uid || 'file' + item.index"
          >
            <el-image
              v-if="isImageFile(item.file)"
              class="echo-img"
              :src="item.file.fileUrl"
              :preview-src-list="[item.file.imgUrl || item.file.fileUrl]"
            ></el-image>
            <audio
              v-else-if="isAudioFile(item.file)"
              controls
              class="echo-audio"
            >
              <source :src="item.file.fileUrl" type="video/mp3" />
              <source :src="item.file.fileUrl" type="audio/ogg" />
              <source :src="item.file.fileUrl" type="audio/mpeg" />
              {{ $t('agent.autioTips') }}
            </audio>
            <div v-else class="docInfo-container">
              <FileIcon
                :type="getFileIconType(item.file)"
                size="32px"
                class="docIcon"
              />
              <div class="docInfo">
                <p class="docInfo_name">
                  {{ item.file.name }}
                </p>
                <p class="docInfo_size">
                  {{ formatFileSize(item.file.size) }}
                </p>
              </div>
            </div>
            <i
              class="el-icon-close echo-close"
              @click.stop="removeFile(item.index)"
            ></i>
            <span
              class="el-icon-loading loading-icon-img"
              v-if="item.file.uploadStatus === 'uploading'"
            ></span>
          </div>
        </div>
      </div>
      <el-divider class="editable-box-divider" v-if="fileList.length" />
      <!-- 问答输入框 -->
      <div
        class="editable-wp flex"
        :style="{
          'pointer-events': fileLoading || disableClick ? 'none' : 'auto',
        }"
      >
        <div
          class="editable-wp-right rl"
          :class="{ 'multi-line-layout': isMultiLine }"
          draggable="true"
        >
          <div class="input-and-clear-box">
            <div
              class="aibase-textarea editable--input"
              ref="editor"
              @input="getPrompt"
              @blur="onBlur"
              @keydown="textareaKeydown($event)"
              @dragenter.prevent
              @dragover.prevent
              @drop.prevent.stop="handleDrop"
              contenteditable="true"
            ></div>
            <span
              class="editable--placeholder"
              v-if="!promptValue || !promptValue.trim()"
            >
              {{ placeholder }}
            </span>
            <i
              class="el-icon-close editable--close"
              @click.stop="clearInput"
            ></i>
          </div>
          <div class="edtable--wrap">
            <streamUploadField
              v-if="visibleUpload"
              :fileTypeArr="fileTypeArr"
              :type="type"
              :maxImageSize="maxImageSize"
              :maxFileSize="maxFileSize"
              :maxPicNum="maxPicNum"
              :maxFileNum="maxFileNum"
              :confirmedFileList="fileList"
              :confirmedFileIdList="fileIdList"
              @setFileId="setFileId"
              @setFile="setFile"
            ></streamUploadField>
            <el-divider v-if="visibleUpload" direction="vertical"></el-divider>
            <el-button class="editable-send-btn" circle plain @click="preSend">
              <svg class="editable-send-icon">
                <use xlink:href="#icon-chatSend" />
              </svg>
            </el-button>
          </div>
        </div>
      </div>
    </div>
    <transition name="el-zoom-in-bottom">
      <div class="perfectReminder-item-box" v-show="randomReminderShow">
        <div
          class="perfectReminder-item"
          v-for="n in randomReminderList"
          :key="n.id"
          :style="`background-color:${colorArr[n.random]}`"
          @click="setRandomReminder(n)"
        >
          <el-popover
            placement="top-start"
            width="300"
            :visible-arrow="false"
            trigger="hover"
            :open-delay="500"
            :content="
              n.prompt && n.prompt.replaceAll('{', '').replaceAll('}', '')
            "
          >
            <span style="font-size: 15px" slot="reference">
              {{ n.title || n.name }}
            </span>
          </el-popover>
        </div>
        <span class="refresh" @click="getReminderList">
          <i class="el-icon-loading" v-show="refreshLoading"></i>
          &nbsp;{{ $t('agent.next') }}
        </span>
      </div>
    </transition>
  </div>
</template>
<script>
import commonMixin from '@/mixins/common';
import uploadChunk from '@/mixins/uploadChunk';
import streamUploadField from './streamUploadField';
import { mapGetters } from 'vuex';
import {
  getPromptTemplateList,
  getPromptBuiltInList,
} from '@/api/promptTemplate';
import FileIcon from '@/components/FileIcon.vue';
import { formatFileSize, getFileIconType } from '@/utils/util';

export default {
  props: {
    source: { type: String },
    fileTypeArr: {
      type: Array,
      required: false,
      default: () => {
        return [];
      },
    },
    type: { type: String },
    disableClick: { type: Boolean, default: false },
    supportReminder: { type: Boolean, default: false },
    hasHistory: { type: Boolean, default: false },
    visibleClearHistory: { type: Boolean, default: true },
    visibleUpload: { type: Boolean, default: true },
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
    // 图片数量限制（-1 表示不限制）
    maxPicNum: {
      type: Number,
      required: false,
      default: -1,
    },
    maxFileNum: {
      type: Number,
      required: false,
      default: -1,
    },
  },
  mixins: [commonMixin, uploadChunk],
  components: { streamUploadField, FileIcon },
  data() {
    return {
      // placeholder: '请输入内容,用Ctrl+Enter可换行',
      promptValue: '',
      randomReminderShow: false,
      refreshLoading: false,
      hasFile: false,
      fileIdList: [],
      fileType: '',
      fileList: [],
      fileUrl: '',
      fileLoading: false,
      isUploading: false,
      isDragging: false,
      lastFileType: '',
      dragConfigured: false,
      dragCleanup: null,
      colorArr: [
        '#dca3c2',
        '#aaa9db',
        '#d1a69b',
        '#7894cf',
        '#4fbed9',
        '#ebb8bd',
        '#9b9655',
        '#3bb4b7',
        '#61aac5',
        '#d79ae5',
        '#51a2da',
        '#89b0f9',
        '#738cbd',
      ],
      randomReminderList: [], //随机8个提示词
      _resizeObserver: null, // 输入框尺寸变化监听器
      isMultiLine: false,
      breakLength: 0, //记录触发换行时的字符长度
    };
  },
  watch: {
    maxPicNum: {
      handler(val) {
        if (val === undefined || val === null || val === '') return;
        this.initDrag(val);
      },
      immediate: true,
    },
  },
  computed: {
    ...mapGetters('app', ['sessionStatus']),
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
    effectiveImageLimit() {
      const limits = [];
      if (this.hasImageLimit) limits.push(this.normalizedMaxPicNum);
      if (this.hasFileLimit) limits.push(this.normalizedMaxFileNum);
      return limits.length ? Math.min(...limits) : -1;
    },
    hasEffectiveImageLimit() {
      return this.effectiveImageLimit >= 0;
    },
    // 展示层按文件类型分组，保留原始 index，避免影响上传队列顺序。
    mixedFileGroups() {
      return this.fileList.reduce(
        (groups, file, index) => {
          const item = { file, index };
          if (this.isImageFile(file)) {
            groups.images.push(item);
          } else {
            groups.nonImages.push(item);
          }
          return groups;
        },
        { nonImages: [], images: [] },
      );
    },
    // 控制 mixed-file-box 的展示顺序：非图片在上，图片在下。
    visibleMixedFileGroups() {
      return [
        { type: 'non-image', files: this.mixedFileGroups.nonImages },
        { type: 'image', files: this.mixedFileGroups.images },
      ].filter(group => group.files.length);
    },
    placeholder() {
      return this.supportReminder
        ? this.$t('common.input.modelChatPlaceholder2')
        : this.$t('common.input.modelChatPlaceholder1');
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
  },
  mounted() {
    if (this.supportReminder) {
      this.originPromptList = [];
      this.getReminderList();
    }
    // 监听输入框尺寸变化，用以改变输入区单行/多行布局切换
    this.$nextTick(() => {
      if (this.$refs.editor) {
        // 初始单行基准高度
        const rect = this.$refs.editor.getBoundingClientRect();
        this._singleLineHeight = rect.height;

        this._resizeObserver = new ResizeObserver(([entry]) => {
          const height = entry.target.getBoundingClientRect().height;
          const currentLength = (this.promptValue || '').trim().length;
          // 通知父组件输入框高度变化
          this.$emit('inputHeightChange', height);

          if (
            !this.isMultiLine &&
            height > this._singleLineHeight &&
            currentLength !== 0
          ) {
            this.isMultiLine = true;
            this.breakLength = currentLength;
          } else if (this.isMultiLine && currentLength < this.breakLength) {
            // 当内容长度回退到触发点以下时，尝试恢复单行布局
            this.isMultiLine = false;
          }
        });
        this._resizeObserver.observe(this.$refs.editor);
      }
    });
  },
  beforeDestroy() {
    // 清理观察器，防止内存泄漏
    if (this._resizeObserver) {
      this._resizeObserver.disconnect();
    }
    if (typeof this.dragCleanup === 'function') {
      this.dragCleanup();
    }
  },
  methods: {
    getFileIconType,
    formatFileSize,
    initDrag(maxFiles) {
      this.$nextTick(() => {
        if (typeof this.dragCleanup === 'function') {
          this.dragCleanup();
        }
        this.dragCleanup = this.$setupDragAndDrop({
          containerSelector: '.editable-wp',
          maxImageFiles: maxFiles,
          maxSizeMB: Number.MAX_SAFE_INTEGER,
          onFiles: files => {
            this.isDragging = true;
            this.processFiles(files);
          },
        });
        this.dragConfigured = true;
      });
    },
    processFiles(files) {
      if (!files || files.length === 0) return;
      const picked = this.filterFileCount(
        this.filterImageCount(this.filterFileSize(this.filterImageSize(files))),
      );
      if (!picked.length) return;
      const fileObjs = picked.map(f => {
        const fileUrl = URL.createObjectURL(f);
        const fileObj = {
          raw: f,
          uid: f.uid || this.$guid(),
          percentage: 0,
          progressStatus: 'active',
          uploadStatus: 'pending',
          uploaded: false,
          fileName: f.name,
          name: f.name,
          size: f.size,
          type: f.type,
          fileUrl,
        };
        fileObj.fileType = this.getFileType(fileObj);
        if (fileObj.fileType === 'image/*') {
          fileObj.imgUrl = fileUrl;
        }
        return fileObj;
      });
      this.fileList = [...this.fileList, ...fileObjs];
      this.fileType = this.fileList[this.fileList.length - 1]?.fileType || '';
      this.fileUrl = this.fileList[this.fileList.length - 1]?.fileUrl || '';
      this.hasFile = true;
      this.fileLoading = true;
      this.triggerNextUpload();
    },
    isImageFile(file) {
      if (!file || !file.name) return false;
      const ext = (file.name.split('.').pop() || '').toLowerCase();
      return (
        file.fileType === 'image/*' ||
        (file.type && file.type.indexOf('image/') === 0) ||
        ['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp', 'svg'].indexOf(ext) > -1
      );
    },
    isAudioFile(file) {
      if (!file || !file.name) return false;
      const ext = (file.name.split('.').pop() || '').toLowerCase();
      return (
        file.fileType === 'audio/*' ||
        (file.type && file.type.indexOf('audio/') === 0) ||
        ['mp3', 'wav', 'ogg'].indexOf(ext) > -1
      );
    },
    getFileType(file) {
      if (!file || !file.name) return '';
      if (this.isImageFile(file)) return 'image/*';
      if (this.isAudioFile(file)) return 'audio/*';
      return 'doc/*';
    },
    showFileLimitMessage() {
      this.$message.warning(
        this.$t('app.uploadFileLimitTips', {
          num: this.normalizedMaxFileNum,
        }),
      );
    },
    showImageLimitMessage() {
      this.$message.warning(
        this.$t('app.uploadImgTips', { num: this.normalizedMaxPicNum }),
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
    filterImageCount(files) {
      const picked = Array.prototype.slice.call(files || []);
      if (!this.hasEffectiveImageLimit) return picked;

      let imageCount = this.fileList.filter(file =>
        this.isImageFile(file),
      ).length;
      const validFiles = picked.filter(file => {
        if (!this.isImageFile(file)) return true;
        imageCount += 1;
        return imageCount <= this.effectiveImageLimit;
      });

      if (validFiles.length !== picked.length) {
        this.showEffectiveImageLimitMessage();
      }

      return validFiles;
    },
    filterImageSize(files) {
      const picked = Array.prototype.slice.call(files || []);
      if (!this.maxImageSizeBytes) return picked;

      const validFiles = picked.filter(file => {
        return !this.isImageFile(file) || file.size <= this.maxImageSizeBytes;
      });

      if (validFiles.length !== picked.length) {
        this.$message.warning(
          this.$t('knowledgeManage.multiKnowledgeDatabase.imageSizeLimit', {
            maxSize: this.maxImageSizeMB,
          }),
        );
      }

      return validFiles;
    },
    filterFileSize(files) {
      const picked = Array.prototype.slice.call(files || []);
      if (!this.maxFileSizeBytes) return picked;

      const validFiles = picked.filter(file => {
        return this.isImageFile(file) || file.size <= this.maxFileSizeBytes;
      });

      if (validFiles.length !== picked.length) {
        this.$message.warning(
          this.$t('app.uploadFileSizeLimitTips', {
            maxSize: this.maxFileSizeMB,
          }),
        );
      }

      return validFiles;
    },
    filterFileCount(files) {
      const picked = Array.prototype.slice.call(files || []);
      if (!this.hasFileLimit) return picked;

      const remainCount = this.normalizedMaxFileNum - this.fileList.length;
      if (remainCount <= 0) {
        this.showFileLimitMessage();
        return [];
      }
      if (picked.length <= remainCount) {
        return picked;
      }

      this.showFileLimitMessage();
      return picked.slice(0, remainCount);
    },
    uploadFile(fileName, oldFileName, fiePath) {
      // 文件上传完成后，释放队列锁并继续调度下一个 pending 文件
      const currentFile = this.fileList[this.fileIndex] || {};
      if (!currentFile.uid) {
        this.isUploading = false;
        this.fileLoading = false;
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
      const index = this.fileIdList.findIndex(
        item => item.uid === currentFile.uid,
      );
      if (index > -1) {
        this.$set(this.fileIdList, index, fileInfoItem);
      } else {
        this.fileIdList.push(fileInfoItem);
      }
      this.isUploading = false;
      this.fileLoading = this.fileList.some(file => this.isPendingUpload(file));
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
      if (nextPendingIndex === -1) {
        this.fileLoading = false;
        return;
      }

      this.maxSizeBytes = 0;
      this.isExpire = true;
      this.isUploading = true;
      this.fileLoading = true;
      this.$set(this.fileList[nextPendingIndex], 'uploadStatus', 'uploading');
      this.$set(this.fileList[nextPendingIndex], 'uploaded', true);
      this.startUpload(nextPendingIndex);
    },
    // 处理拖拽到输入框的文件
    handleDrop(event) {
      const dt = event.dataTransfer;
      if (!dt || !dt.files) return;

      const fileList = dt.files;
      const files = Array.prototype.slice.call(fileList);
      if (files.length === 0) return;

      // 调用文件处理方法
      this.processFiles(files);
    },
    setPrompt(data) {
      this.clearInput();
      this.promptValue = data;
      this.$refs.editor.innerHTML = data
        .replaceAll('{', '<div class="light-input" contenteditable="true">')
        .replaceAll('}', '</div>');
    },
    getPrompt() {
      if (this.supportReminder) {
        if (this.$refs.editor.innerHTML === '/') {
          this.openReminderDialog();
        } else {
          this.randomReminderShow = false;
        }
      }
      let prompt = this.$refs.editor.innerText;
      this.promptValue = prompt;
      return prompt;
    },
    clearFile() {
      this.fileIdList = [];
      this.fileList = [];
      this.fileType = '';
      this.fileUrl = '';
      this.hasFile = false;
      this.fileLoading = false;
      this.isUploading = false;
    },
    removeFile(index) {
      const removedFile = this.fileList[index];
      this.fileList.splice(index, 1);
      const fileInfoIndex = this.fileIdList.findIndex(
        item => item.uid === removedFile?.uid,
      );
      if (fileInfoIndex > -1) {
        this.fileIdList.splice(fileInfoIndex, 1);
      } else {
        this.fileIdList.splice(index, 1);
      }

      if (this.fileList.length === 0) {
        this.clearFile();
        return;
      }

      const lastFile = this.fileList[this.fileList.length - 1];
      this.fileType = lastFile.fileType || this.getFileType(lastFile);
      this.fileUrl = lastFile.fileUrl;
      this.hasFile = true;
    },
    setFileId(fileIdList) {
      this.fileIdList = (fileIdList || []).map(item => ({
        ...item,
        fileType: item.fileType || this.getFileType({ name: item.fileName }),
      }));
      const lastFileId = this.fileIdList[this.fileIdList.length - 1];
      this.fileUrl = lastFileId?.fileUrl || '';
      this.fileType = lastFileId?.fileType || '';
    },
    setFile(fileList) {
      this.fileList = (fileList || []).map(file => ({
        ...file,
        fileType: file.fileType || this.getFileType(file),
        uploadStatus: file.uploadStatus || 'success',
        uploaded: file.uploaded !== undefined ? file.uploaded : true,
      }));
      const lastFile = this.fileList[this.fileList.length - 1];
      if (lastFile) {
        this.fileType = lastFile.fileType || this.getFileType(lastFile);
        this.fileUrl = lastFile.fileUrl;
        this.hasFile = true;
      }
    },
    getFileList() {
      return this.fileList;
    },
    getFileIdList() {
      return this.fileList
        .map(file => this.fileIdList.find(item => item.uid === file.uid))
        .filter(Boolean);
    },
    clearInput() {
      this.$refs.editor.innerHTML = '';
      this.promptValue = '';
      this.isMultiLine = false;
    },
    onBlur() {
      //勿删，定义此方法用于获取焦点
    },
    //换行并重新定位光标位置
    textareaRange() {
      let el = this.$refs.editor;
      let range = document.createRange();
      let sel = document.getSelection();
      let offset = sel.focusOffset;
      let content = el.innerHTML;
      el.innerHTML = content.slice(0, offset) + '\n' + content.slice(offset);
      range.setStart(el.childNodes[0], offset + 1);
      range.collapse(true);
      sel.removeAllRanges();
      sel.addRange(range);
    },
    textareaKeydown(event) {
      if (event.ctrlKey && event.keyCode === 13) {
        this.textareaRange();
      } else if (event.keyCode === 13) {
        this.preSend();
        event.preventDefault();
        return false;
      }
    },
    preSend() {
      this.hasFile = false;
      this.$emit('preSend');
    },
    setRandomReminder(n) {
      this.setPrompt(n.prompt);
      this.randomReminderShow = false;
    },
    openReminderDialog() {
      this.randomReminderShow = true;
      !this.refreshLoading && this.getRandomReminderList();
    },
    getReminderList() {
      this.refreshLoading = true;
      const p1 = new Promise(resolve => {
        getPromptTemplateList({
          name: '',
        })
          .then(res => {
            if (res.code === 0) {
              resolve(res.data.list || []);
              return;
            }
            resolve([]);
          })
          .catch(() => {
            resolve([]);
          });
      });
      const p2 = new Promise(resolve => {
        getPromptBuiltInList({
          name: '',
          category: 'all',
        })
          .then(res => {
            if (res.code === 0) {
              resolve(res.data.list || []);
              return;
            }
            resolve([]);
          })
          .catch(() => {
            resolve([]);
          });
      });
      Promise.all([p1, p2])
        .then(([list1, list2]) => {
          this.originPromptList = [...list1, ...list2];
        })
        .finally(() => {
          this.refreshLoading = false;
          this.randomReminderShow && this.getRandomReminderList();
        });
    },
    // 从全量提示词中随机获取8个
    getRandomReminderList() {
      this.randomReminderShow = true;
      if (this.refreshLoading) {
        return;
      }
      const recommendCount = 8;
      if (this.originPromptList.length <= recommendCount)
        return this.originPromptList.slice();
      const shuffled = this.originPromptList.slice();
      for (let i = shuffled.length - 1; i > 0; i--) {
        const index = Math.floor(Math.random() * (i + 1));
        [shuffled[i], shuffled[index]] = [shuffled[index], shuffled[i]];
      }
      this.randomReminderList = shuffled.slice(0, recommendCount).map(item => ({
        ...item,
        random: parseInt(Math.random(13) * 10),
      }));
    },
    handleClearHistory() {
      this.$emit('clearHistory');
    },
  },
};
</script>
<style lang="scss" scoped>
.tips {
  color: #ccc;
}
.auto-width-select {
  min-width: 250px;
  max-width: 450px;
}
.editable-box {
  border: 1px solid #d3d7dd;
  .loading-icon {
    font-size: 18px;
    color: $color;
    margin-left: 10px;
  }
  .editable-box-divider {
    margin: 6px 0;
  }
  .echo-img-box {
    position: absolute;
    display: flex;
    top: -65px;
    justify-content: flex-start;
    align-items: center;
    gap: 10px;
    .echo-img-item {
      height: 60px;
      width: 60px;
      display: flex;
      position: relative;
      flex-shrink: 0;
      .loading-icon-img {
        position: absolute;
        right: 50%;
        top: 50%;
        transform: translate(50%, -50%);
        color: $color;
        font-size: 18px;
        animation: loading 1s linear infinite;
      }
    }
    .echo-img {
      width: 100%;
      height: 100%;
      object-fit: contain;
      background: #ffff;
      border-radius: 8px;
    }
    .mixed-file-item.doc-file-item {
      width: 220px;
    }
    .docInfo-container {
      width: 100%;
      height: 100%;
      box-sizing: border-box;
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 8px 24px 8px 8px;
      background: #fff;
      overflow: hidden;
      background-color: #f7f7fa;
      border-radius: 8px;
    }
    .docIcon {
      flex: 0 0 32px;
      width: 32px;
      height: 32px;
      object-fit: contain;
    }
    .docInfo {
      min-width: 0;
      flex: 1;
      line-height: 1.4;
      p {
        margin: 0;
      }
      .docInfo_name {
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        color: #333;
      }
      .docInfo_size {
        margin-top: 2px;
        color: #bbbbbb;
      }
    }
    .echo-close {
      position: absolute;
      right: -4px;
      top: -4px;
      background-color: #333;
      color: #fff;
      cursor: pointer;
      border-radius: 15px;
      font-size: 12px;
      padding: 2px;
    }
    .fileid-icon {
      line-height: 20px;
      position: absolute;
      bottom: 0;
      text-align: center;
      background: #3333337a;
      width: 100%;
      color: #67c23a;
      i {
        font-weight: bold;
        font-size: 16px;
      }
    }
  }
  .echo-img-box.mixed-file-box {
    position: static;
    top: auto;
    width: 100%;
    box-sizing: border-box;
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
    padding: 8px 10px 0;
    overflow: visible;
  }
  .mixed-file-group {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-start;
    align-items: center;
    gap: 10px;
    &.mixed-file-non-image-group {
      max-height: 360px;
      overflow-y: auto;
      padding-top: 4px;
    }
  }
  .echo-doc-box {
    background: #fff;
    width: auto;
    border: 1px solid #dcdfe6;
    border-radius: 5px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 10px 50px 10px 5px;
    .docIcon {
      width: 30px;
      height: 30px;
    }
    .docInfo {
      .docInfo_name {
        color: #333;
      }
      .docInfo_size {
        color: #bbbbbb;
      }
    }
  }
  .echo-audio-box {
    position: absolute;
    width: 300px;
    height: 40px;
    top: -60px;
    audio {
      width: 100%;
    }
    .echo-close {
      position: absolute;
      top: 0;
      right: 0;
      background-color: #333;
      color: #fff;
    }
  }
  .editable-wp {
    position: relative;
    .overlay {
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      z-index: 9999;
      background-color: rgba(255, 255, 255, 0.4);
      border: 1px solid #dcdfe6;
      border-radius: 6px;
      pointer-events: auto;
    }
  }
  .editable-wp-left {
    min-width: 20px;
    .upload-icon {
      margin: 5px 5px 5px 11px;
      padding: 3px;
      border-radius: 4px;
      cursor: pointer;
    }
  }
  .editable-wp-right {
    display: flex;
    flex-direction: row;
    flex: 1;
    padding: 4px 10px !important;
    align-items: center;
    gap: 4px;

    &.multi-line-layout {
      flex-direction: column;
      align-items: flex-start;
    }
  }

  .input-and-clear-box {
    display: flex;
    flex: 1;
    align-items: flex-end;
    position: relative;
  }
  ::v-deep .light-input {
    border: 1px solid deepskyblue;
    padding: 2px 14px 2px 10px;
    margin: 0 5px;
    border-radius: 4px;
    display: inline-block;
    box-shadow: 1px 1px 10px #d3ebf3;
  }
}
.aibase-textarea {
  min-height: 22px !important;
  height: auto !important;
  padding: 0px 10px 0 0;
  flex: 1;
  word-break: break-all;
}

.editable--placeholder {
  position: absolute;
  left: 0px !important;
  top: 0px !important;
  line-height: 22px;
  max-width: calc(100% - 25px);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  pointer-events: none;
}

.editable--close {
  position: static !important;
  z-index: 20;
  margin: 0 0 2px 5px;
  cursor: pointer;
  font-size: 18px;
  color: #dbdada;
  padding: 2px;
  flex-shrink: 0;
}

.edtable--wrap {
  height: 35px;
  display: flex;
  align-items: center;
  flex-shrink: 0;

  .multi-line-layout & {
    width: 100%;
    justify-content: flex-end;
  }
}
.model-box {
  padding: 10px 0;
}
.btnActive {
  color: #e60001 !important;
  border: 1px solid rgb(228, 165, 165) !important;
  background: linear-gradient(
    111deg,
    rgba(255, 58, 58, 0.2) 0%,
    #fff 25%,
    #fff 69%,
    rgba(255, 58, 58, 0.2) 100%
  ) !important;
}
.btnAnactive {
  color: #606266 !important;
  border: 1px solid #dcdfe6 !important;
  background: #ffffff !important;
}
.perfectReminder-item-box {
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
  .perfectReminder-item {
    width: calc((100% - 80px) / 4);
    height: 46px;
    line-height: 46px;
    text-align: center;
    position: relative;
    margin: 5px 10px;
    display: inline-block;
    background-color: #dfebfb;
    color: #fff;
    cursor: pointer;
    border-radius: 9px;
  }
  .perfectReminder-active {
    border: 1px solid #ec0b0c;
    overflow: hidden;
    i,
    span {
      color: #ec0b0c;
    }
  }
  .refresh {
    position: absolute;
    right: 30px;
    bottom: 10px;
    cursor: pointer;
    color: #62a1fb;
  }
}

.chat-input-wrapper {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chat-input-wrapper-left {
  margin-right: 20px;
}

.chat-clear-btn {
  padding: 8px;
  font-size: 14px;
  border-radius: 99px;
  line-height: 0;
  border-color: rgba(68, 83, 130, 0.25);
  color: rgba(15, 21, 40, 0.82);
  &:hover {
    background-color: rgba(68, 83, 130, 0.05);
    border-color: rgba(68, 83, 130, 0.5);
    color: rgba(15, 21, 40, 0.82);
  }
  .chat-clear-icon {
    width: 14px;
    height: 14px;
    fill: currentColor;
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

.editable-send-btn {
  padding: 8px;
  border: none;
  color: rgb(81, 71, 255);
  line-height: 0;
  display: inline-flex;
  justify-content: center;
  align-items: center;
  &:hover {
    background-color: rgba(87, 104, 161, 0.08) !important;
  }
  .editable-send-icon {
    width: 18px;
    height: 18px;
    fill: currentColor;
  }
}
</style>
