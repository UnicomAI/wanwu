/**
 * Common mixins methods
 * Provide common utility methods and lifecycle hooks used in the project
 */
import { i18n } from '@/lang'

export default {
  data() {
    return {
    }
  },

  methods: {
    /**
     * Common drag and drop encapsulation
     * @param {Object} opt
     * @param {string} [opt.containerSelector='.editable-wp'] - Container selector for listening to drag and drop
     * @param {(files:Array,ctx:Object)=>void} [opt.onFiles] - Callback for handling files after they are dropped
     * @param {number} [opt.maxImageFiles=3] - Maximum number of image files to upload
     * @param {number} [opt.maxFiles=3] - Default maximum upload quantity (deprecated, use maxImageFiles first)
     */
    $setupDragAndDrop(opt = {}) {
      const { containerSelector = '.editable-wp', onFiles } = opt
      const wrap = this.$el && this.$el.querySelector ? this.$el.querySelector(containerSelector) : null
      if (!wrap) return () => {}

      const prevent = (e) => { e.preventDefault(); e.stopPropagation(); wrap.classList.add('is-dropping'); }
      const leave = () => { wrap.classList.remove('is-dropping'); }
      
      // 判断FileYesNo为ImageType
      const isImageFile = (f) => {
        if (!f || !f.name) return false
        var ext = (f.name.split('.').pop() || '').toLowerCase()
        var imageExts = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp', 'svg']
        return imageExts.indexOf(ext) > -1 || (f.type && f.type.indexOf('image/') === 0)
      }
      
      const onDrop = async (e) => {
        prevent(e)
        try {
          const dt = e && e.dataTransfer
          const fileList = (dt && dt.files) ? dt.files : []
          const rawFiles = Array.prototype.slice.call(fileList)
          if (!rawFiles.length) return

          // 安全限制：Count/大小/Type白Name单
          const maxImageFiles = Number(opt.maxImageFiles || opt.maxFiles || 3) // ImageTypeFile of MaxCount
          const maxSizeMB = Number(opt.maxSizeMB || 50) // 单个FileDefault 50MB
          const maxSize = maxSizeMB * 1024 * 1024
          const allowExt = (opt.acceptExt || ['jpg','jpeg','png','gif','webp','bmp','svg','mp3','wav','ogg','txt','pdf','doc','docx','xlsx','xls','pptx','csv','html']).map(function(s){return String(s).toLowerCase()})

          // First找出第一个Has效File，判断FileType
          var firstValidFile = null
          for (var j = 0; j < rawFiles.length; j++) {
            var tempFile = rawFiles[j]
            if (!tempFile || !tempFile.name) continue
            if (typeof tempFile.size !== 'number' || tempFile.size < 0 || isNaN(tempFile.size)) continue
            var tempExt = (tempFile.name.split('.').pop() || '').toLowerCase()
            var tempOkType = allowExt.indexOf(tempExt) > -1 || (tempFile.type && (tempFile.type.indexOf('image/') === 0 || tempFile.type.indexOf('audio/') === 0))
            if (tempOkType && tempFile.size <= maxSize) {
              firstValidFile = tempFile
              break
            }
          }
          
          // IfNoHas效File，DirectlyBack
          if (!firstValidFile) {
            if (this && this.$message && this.$message.warning) {
              this.$message.warning(i18n.t('agent.fileTypeNotSupported'))
            }
            return
          }
          
          // 判断第一个Has效FileYesNo为ImageType
          var isImageType = isImageFile(firstValidFile)
          // ImageType：Use maxImageFiles 限制；NonImageType：限制为1个File（覆盖之前）
          var maxFiles = isImageType ? maxImageFiles : 1
          
          //  滤Illegal/ 大File
          const safeFiles = []
          const rejected = []
          for (var i = 0; i < rawFiles.length; i++) {
            var f = rawFiles[i]
            
            // CheckFileObjectYesNo完整（MustHas name Property）
            if (!f || !f.name) {
              rejected.push(f)
              continue
            }
            
            // CheckFile大小YesNo存在ANDHas效（MustYesNumberAND >= 0）
            if (typeof f.size !== 'number' || f.size < 0 || isNaN(f.size)) {
              rejected.push(f)
              continue
            }
            
            var ext = (f.name && f.name.split('.').pop() || '').toLowerCase()
            var okType = allowExt.indexOf(ext) > -1 || (f.type && (f.type.indexOf('image/') === 0 || f.type.indexOf('audio/') === 0))
            
            // CheckFileTypeAnd大小（此 when  f.size 已确保YesHas效Number）
            if (!okType || f.size > maxSize) {
              rejected.push(f)
              continue
            }
            
            // Check当前FileYesNo为ImageType，确保FileType一致
            var currentFileIsImage = isImageFile(f)
            if (!isImageType && currentFileIsImage) {
              // If第一个FileIs NotImage，但当前FileYesImage，拒绝（保持Type一致）
              rejected.push(f)
              continue
            }
            if (isImageType && !currentFileIsImage) {
              // If第一个FileYesImage，但当前FileIs NotImage，拒绝（保持Type一致）
              rejected.push(f)
              continue
            }
            
            // CheckCount限制
            // NonImageType：只能Upload1个File（覆盖之前）；ImageType：Use maxImageFiles 限制
            if (safeFiles.length >= maxFiles) {
              rejected.push(f)
              continue
            }
            
            safeFiles.push(f)
          }

          // Tip被拒File
          if (rejected.length && this && this.$message && this.$message.warning) {
            if (!isImageType && rawFiles.length > 1) {
              this.$message.warning(i18n.t('agent.fileTypeNotSupportedTips'))
            } else if (isImageType && safeFiles.length < rawFiles.length) {
              this.$message.warning(i18n.t('agent.fileTypeNotSupportedTips1'))
            } else {
              this.$message.warning(i18n.t('agent.fileTypeNotSupportedTips2'))
            }
          }

          if (!safeFiles.length) return

          // 覆盖前释放旧 of  ObjectURL，避免Memory泄漏
          try {
            var currentList = this && this.fileList
            if (currentList && currentList.forEach) {
              currentList.forEach(function(f){
                try { if (f && f.fileUrl) URL.revokeObjectURL(f.fileUrl) } catch(e) {}
                try { if (f && f.imgUrl) URL.revokeObjectURL(f.imgUrl) } catch(e) {}
              })
            }
          } catch(err) {}

          if (typeof onFiles === 'function') {
            onFiles(safeFiles, { event: e, wrap })
          }
        } finally {
          leave()
        }
      }

      wrap.addEventListener('dragenter', prevent)
      wrap.addEventListener('dragover', prevent)
      wrap.addEventListener('dragleave', leave)
      wrap.addEventListener('drop', onDrop)

      const cleanup = () => {
        wrap.removeEventListener('dragenter', prevent)
        wrap.removeEventListener('dragover', prevent)
        wrap.removeEventListener('dragleave', leave)
        wrap.removeEventListener('drop', onDrop)
      }

      this.$once('hook:beforeDestroy', () => {
        try { cleanup() } catch (e) {}
      })

      return cleanup
    },
    /**
     * Format date
     * @param {Date|string|number} date - 日期
     * @param {string} format - 格式String
     * @returns {string} - 格式化后 of 日期String
     */
    $formatDate(date, format = 'YYYY-MM-DD HH:mm:ss') {
      if (!date) return ''
      const d = new Date(date)
      if (isNaN(d.getTime())) return ''
      
      const year = d.getFullYear()
      const month = String(d.getMonth() + 1).padStart(2, '0')
      const day = String(d.getDate()).padStart(2, '0')
      const hours = String(d.getHours()).padStart(2, '0')
      const minutes = String(d.getMinutes()).padStart(2, '0')
      const seconds = String(d.getSeconds()).padStart(2, '0')
      
      return format
        .replace('YYYY', year)
        .replace('MM', month)
        .replace('DD', day)
        .replace('HH', hours)
        .replace('mm', minutes)
        .replace('ss', seconds)
    },

    /**
     * 深拷贝对象
     * @param {any} obj - 要拷贝 of 对象
     * @returns {any} - 拷贝后 of 对象
     */
    $deepClone(obj) {
      if (obj === null || typeof obj !== 'object') return obj
      if (obj instanceof Date) return new Date(obj.getTime())
      if (obj instanceof Array) return obj.map(item => this.$deepClone(item))
      if (typeof obj === 'object') {
        const clonedObj = {}
        for (const key in obj) {
          if (obj.hasOwnProperty(key)) {
            clonedObj[key] = this.$deepClone(obj[key])
          }
        }
        return clonedObj
      }
    },

    /**
     * 防抖函数
     * @param {Function} func - 要防抖 of 函数
     * @param {number} delay - 延迟 when 间（毫秒）
     * @returns {Function} - 防抖后 of 函数
     */
    $debounce(func, delay = 300) {
      let timeoutId
      return function (...args) {
        clearTimeout(timeoutId)
        timeoutId = setTimeout(() => func.apply(this, args), delay)
      }
    },

    /**
     * 节流函数
     * @param {Function} func - 要节流 of 函数
     * @param {number} delay - 延迟 when 间（毫秒）
     * @returns {Function} - 节流后 of 函数
     */
    $throttle(func, delay = 300) {
      let lastCall = 0
      return function (...args) {
        const now = Date.now()
        if (now - lastCall >= delay) {
          lastCall = now
          func.apply(this, args)
        }
      }
    },
    /**
     * GetFile大小格式化String
     * @param {number} bytes - 字节数
     * @returns {string} - 格式化后 of File大小
     */
    $formatFileSize(bytes) {
      if (bytes === 0) return '0 B'
      const k = 1024
      const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
    },

    /**
     * 验证邮箱格式
     * @param {string} email - 邮箱地址
     * @returns {boolean} - YesNo有效
     */
    $isValidEmail(email) {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
      return emailRegex.test(email)
    },

    /**
     * 验证手机号格式
     * @param {string} phone - 手机号
     * @returns {boolean} - YesNo有效
     */
    $isValidPhone(phone) {
      const phoneRegex = /^1[3-9]\d{9}$/
      return phoneRegex.test(phone)
    },

    /**
     * 滚动到页面顶部
     */
    $scrollToTop() {
      window.scrollTo({
        top: 0,
        behavior: 'smooth'
      })
    },

    /**
     * Copy文本到剪贴板
     * @param {string} text - 要Copy of 文本
     * @returns {Promise} - Copy结果
     */
    async $copyToClipboard(text) {
      try {
        if (navigator.clipboard) {
          await navigator.clipboard.writeText(text)
          this.$success('Copy successful')
        } else {
          // 兼容旧浏览器
          const textArea = document.createElement('textarea')
          textArea.value = text
          document.body.appendChild(textArea)
          textArea.select()
          document.execCommand('copy')
          document.body.removeChild(textArea)
          this.$success('Copy successful')
        }
      } catch (error) {
        this.$error('Copy failed')
        console.error('Copy failed:', error)
      }
    },

    /**
     * Process引用点击事件
     * @param {Event} e - 点击事件
     * @param {Object} options - Configuration选项
     * @param {number} options.sessionStatus - 会话Status
     * @param {Object} options.sessionData - 会话Data
     * @param {string} options.citationSelector - 引用元素选择器，默认为 '.citation'
     * @param {string} options.subTagSelector - 子标签选择器，默认为 '.subTag'
     * @param {string} options.scrollElementId - 滚动容器ID，默认为 'timeScroll'
     * @param {Function} options.onToggleCollapse - 切换折叠Status of 回调函数
     */
    $handleCitationClick(e, options = {}) {
      const {
        sessionStatus = 0,
        sessionData = null,
        citationSelector = '.citation',
        scrollElementId = 'timeScroll',
        onToggleCollapse = null
      } = options
      // Check会话Status
      if (sessionStatus === 0) return

      // Find最近 of Reference元素
      const citationElement = e.target.closest(citationSelector)
      if (!citationElement) return

      // GetTagIndex
      const tagIndex = parseInt(citationElement.textContent, 10)
      if (isNaN(tagIndex) || tagIndex <= 0) return

      // Get父级IndexAnd折叠Status
      const parentsIndexAttr = citationElement.getAttribute('data-parents-index')
      const parentsIndex = parentsIndexAttr ? parseInt(parentsIndexAttr, 10) : null
      // Check parentsIndex YesNoHas效
      if (isNaN(parentsIndex)) return
      
      // Check会话Data结构
      if (!sessionData || 
          !sessionData.history || 
          !sessionData.history[parentsIndex] || 
          !sessionData.history[parentsIndex].searchList || 
          !sessionData.history[parentsIndex].searchList[tagIndex - 1]
        ) {
        return
      }
      // 切换折叠Status - 严格按照Component中 of collapseClickMethod逻辑
      const searchItem = sessionData.history[parentsIndex].searchList[tagIndex - 1]
      const currentCollapse = searchItem.collapse
      const newCollapse = !currentCollapse
      if (onToggleCollapse && typeof onToggleCollapse === 'function') {
        onToggleCollapse(searchItem, newCollapse)
      } else {
        const updatedItem = { ...searchItem, collapse: newCollapse }
        if (this.$set) {
          this.$set(sessionData.history[parentsIndex].searchList, tagIndex - 1, updatedItem)
        } else {
          sessionData.history[parentsIndex].searchList[tagIndex - 1] = updatedItem
        }
      }

      // 滚动到Bottom
      const timeScrollElement = document.getElementById(scrollElementId)
      if (timeScrollElement) {
        timeScrollElement.scrollTop = timeScrollElement.scrollHeight
      }

      // 阻止Event冒泡
      e.stopPropagation()
    }
  },

  computed: {
    /**
     * YesNoIs Empty对象
     * @returns {Function} - 判断函数
     */
    $isEmpty() {
      return (obj) => {
        if (obj === null || obj === undefined) return true
        if (typeof obj === 'string') return obj.trim() === ''
        if (Array.isArray(obj)) return obj.length === 0
        if (typeof obj === 'object') return Object.keys(obj).length === 0
        return false
      }
    }
  },

  mounted() {

  },

  beforeDestroy() {

  }
}
