<template>
  <svg class="icon-file" aria-hidden="true" :style="{ fontSize: size }">
    <use :xlink:href="iconHref"></use>
  </svg>
</template>

<script>
// 支持的文件类型列表（与 iconfont 图标对应）
export const SUPPORTED_TYPES = [
  // 基础与文件夹
  'md', // Markdown 文件
  'markdown', // Markdown 文件（自绘）
  'unknown', // 未知文件类型
  'dir', // 文件夹 (关闭)
  'diropen', // 文件夹 (打开)
  // 办公文档
  'doc', // Word 文档 (.doc)
  'docx', // Word 文档 (.docx)
  'xls', // Excel 表格 (.xls)
  'xlsx', // Excel 表格 (.xlsx)
  'csv', // CSV 纯文本表格
  'pdf', // PDF 文档
  'ppt', // PPT 演示文稿 (.ppt)
  'pptx', // PPT 演示文稿 (.pptx)
  'ofd', // OFD 版式文档
  'txt', // 纯文本文件
  // 多媒体 - 音频
  'aac', // AAC 音频
  'flac', // FLAC 无损音频
  'm4a', // M4A 音频
  'midi', // MIDI 音乐序列
  'mp3', // MP3 音频
  'ogg', // OGG 格式音频
  'wav', // WAV 音频
  'mp2', // MP2 音频
  // 多媒体 - 视频
  'mp4', // MP4 视频
  'mkv', // MKV 视频
  'ogv', // OGV 视频
  'webm', // WEBM 视频
  // 多媒体 - 图片
  'gif', // GIF 动图
  'jpg', // JPG 图片
  'png', // PNG 图片
  'webp', // WEBP 图片
  'svg', // SVG 矢量图
  // 开发文件
  'js', // JavaScript 脚本
  'html', // HTML 页面
  'php', // PHP 脚本
  'asp', // ASP 脚本
  // 压缩包
  'zip', // ZIP 压缩包
  'rar', // RAR 压缩包
];

export default {
  name: 'FileIcon',
  props: {
    type: {
      type: String,
      default: 'unknown',
    },
    size: {
      type: String,
      default: '1em',
    },
  },
  computed: {
    // 验证并规范化文件类型
    validatedType() {
      return SUPPORTED_TYPES.includes(this.type) ? this.type : 'unknown';
    },
    iconHref() {
      return `#icon-${this.validatedType}`;
    },
  },
};
</script>

<style lang="scss" scoped>
.icon-file {
  width: 1em;
  height: 1em;
  vertical-align: -0.15em;
  fill: currentColor;
  overflow: hidden;
}
</style>
