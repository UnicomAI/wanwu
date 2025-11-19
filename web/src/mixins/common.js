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
      
      // CheckFileYesNoisImageType
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

          // securitylimit：Count/size/Type白Name单
          const maxImageFiles = Number(opt.maxImageFiles || opt.maxFiles || 3) // ImageTypeFile of MaxCount
          const maxSizeMB = Number(opt.maxSizeMB || 50) // singleFileDefault 50MB
          const maxSize = maxSizeMB * 1024 * 1024
          const allowExt = (opt.acceptExt || ['jpg','jpeg','png','gif','webp','bmp','svg','mp3','wav','ogg','txt','pdf','doc','docx','xlsx','xls','pptx','csv','html']).map(function(s){return String(s).toLowerCase()})

          // First找出firstHas效File，CheckFileType
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
          
          // CheckfirstHas效FileYesNoisImageType
          var isImageType = isImageFile(firstValidFile)
          // ImageType：Use maxImageFiles limit；NonImageType：limitis1File（overrideprevious）
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
            
            // CheckFilesizeYesNoexistANDHas效（MustYesNumberAND >= 0）
            if (typeof f.size !== 'number' || f.size < 0 || isNaN(f.size)) {
              rejected.push(f)
              continue
            }
            
            var ext = (f.name && f.name.split('.').pop() || '').toLowerCase()
            var okType = allowExt.indexOf(ext) > -1 || (f.type && (f.type.indexOf('image/') === 0 || f.type.indexOf('audio/') === 0))
            
            // CheckFileTypeAndsize（此 when  f.size alreadyensureYesHas效Number）
            if (!okType || f.size > maxSize) {
              rejected.push(f)
              continue
            }
            
            // CheckcurrentFileYesNoisImageType，ensureFileTypeconsistent
            var currentFileIsImage = isImageFile(f)
            if (!isImageType && currentFileIsImage) {
              // IffirstFileIs NotImage，butcurrentFileYesImage，reject（keepTypeconsistent）
              rejected.push(f)
              continue
            }
            if (isImageType && !currentFileIsImage) {
              // IffirstFileYesImage，butcurrentFileIs NotImage，reject（keepTypeconsistent）
              rejected.push(f)
              continue
            }
            
            // CheckCountlimit
            // NonImageType：can onlyUpload1File（overrideprevious）；ImageType：Use maxImageFiles limit
            if (safeFiles.length >= maxFiles) {
              rejected.push(f)
              continue
            }
            
            safeFiles.push(f)
          }

          // Tiprejected拒File
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

          // override前releaseold of  ObjectURL，avoidMemoryleak
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
     * @param {Date|string|number} date - date
     * @param {string} format - formatString
     * @returns {string} - formatafter of dateString
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
     * deep copyobject
     * @param {any} obj - tocopy of object
     * @returns {any} - copyafter of object
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
     * debouncefunction
     * @param {Function} func - todebounce of function
     * @param {number} delay - delay when 间（milliseconds）
     * @returns {Function} - debounceafter of function
     */
    $debounce(func, delay = 300) {
      let timeoutId
      return function (...args) {
        clearTimeout(timeoutId)
        timeoutId = setTimeout(() => func.apply(this, args), delay)
      }
    },

    /**
     * throttlefunction
     * @param {Function} func - tothrottle of function
     * @param {number} delay - delay when 间（milliseconds）
     * @returns {Function} - throttleafter of function
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
     * GetFilesizeformatString
     * @param {number} bytes - bytes
     * @returns {string} - formatafter of Filesize
     */
    $formatFileSize(bytes) {
      if (bytes === 0) return '0 B'
      const k = 1024
      const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
    },

    /**
     * validateemailformat
     * @param {string} email - emailaddress
     * @returns {boolean} - YesNovalid
     */
    $isValidEmail(email) {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
      return emailRegex.test(email)
    },

    /**
     * validatephoneformat
     * @param {string} phone - phone
     * @returns {boolean} - YesNovalid
     */
    $isValidPhone(phone) {
      const phoneRegex = /^1[3-9]\d{9}$/
      return phoneRegex.test(phone)
    },

    /**
     * scrolltopagetop
     */
    $scrollToTop() {
      window.scrollTo({
        top: 0,
        behavior: 'smooth'
      })
    },

    /**
     * Copytexttoclipboard
     * @param {string} text - toCopy of text
     * @returns {Promise} - Copyresult
     */
    async $copyToClipboard(text) {
      try {
        if (navigator.clipboard) {
          await navigator.clipboard.writeText(text)
          this.$success('Copy successful')
        } else {
          // compatibleoldbrowser
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
     * Processcitationclickevent
     * @param {Event} e - clickevent
     * @param {Object} options - Configurationoptions
     * @param {number} options.sessionStatus - sessionStatus
     * @param {Object} options.sessionData - sessionData
     * @param {string} options.citationSelector - citationelementselector，defaultis '.citation'
     * @param {string} options.subTagSelector - 子Tagselector，defaultis '.subTag'
     * @param {string} options.scrollElementId - scrollcontainerID，defaultis 'timeScroll'
     * @param {Function} options.onToggleCollapse - togglecollapseStatus of callbackfunction
     */
    $handleCitationClick(e, options = {}) {
      const {
        sessionStatus = 0,
        sessionData = null,
        citationSelector = '.citation',
        scrollElementId = 'timeScroll',
        onToggleCollapse = null
      } = options
      // ChecksessionStatus
      if (sessionStatus === 0) return

      // Findnearest of Referenceelement
      const citationElement = e.target.closest(citationSelector)
      if (!citationElement) return

      // GetTagIndex
      const tagIndex = parseInt(citationElement.textContent, 10)
      if (isNaN(tagIndex) || tagIndex <= 0) return

      // GetparentIndexAndcollapseStatus
      const parentsIndexAttr = citationElement.getAttribute('data-parents-index')
      const parentsIndex = parentsIndexAttr ? parseInt(parentsIndexAttr, 10) : null
      // Check parentsIndex YesNoHas效
      if (isNaN(parentsIndex)) return
      
      // ChecksessionDatastructure
      if (!sessionData || 
          !sessionData.history || 
          !sessionData.history[parentsIndex] || 
          !sessionData.history[parentsIndex].searchList || 
          !sessionData.history[parentsIndex].searchList[tagIndex - 1]
        ) {
        return
      }
      // togglecollapseStatus - strictlyaccording toComponentin of collapseClickMethodlogic
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

      // scrolltoBottom
      const timeScrollElement = document.getElementById(scrollElementId)
      if (timeScrollElement) {
        timeScrollElement.scrollTop = timeScrollElement.scrollHeight
      }

      // preventEventbubbling
      e.stopPropagation()
    }
  },

  computed: {
    /**
     * YesNoIs Emptyobject
     * @returns {Function} - Checkfunction
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
