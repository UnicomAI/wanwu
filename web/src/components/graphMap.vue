<template>
  <div class="graph-map-container">
    <div class="graph-map-content">
      <div class="graph-header" v-if="showHeader">
        <span class="el-icon-arrow-left back" @click="goBack"></span>
        <span class="header-text">
          {{headerName}}
        </span>
      </div>
      <div class="graph-toolbar" v-if="showToolBar">
        <span class="graphStatus">
          <span class="statusNum">{{data.processingCount || 0}}</span>
          <span class="statusText">{{$t('knowledgeManage.graph.processing')}}</span>
        </span>
        <span class="graphStatus">
          <span class="statusNum">{{data.successCount || 0}}</span>
          <span class="statusText">{{$t('knowledgeManage.graph.finished')}}</span>
        </span>
        <span class="graphStatus">
          <span class="statusNum">{{data.failCount || 0}}</span>
          <span class="statusText">{{$t('knowledgeManage.graph.failed')}}</span>
        </span>
        <el-divider direction="vertical"></el-divider>
        <el-tooltip :content="$t('knowledgeManage.graph.zoomIn')" placement="top">
          <el-button 
            icon="el-icon-zoom-in" 
            @click="zoomIn"
          ></el-button>
        </el-tooltip>
        <span class="zoom-display">{{ zoomPercentage }}%</span>
        <el-tooltip :content="$t('knowledgeManage.graph.zoomOut')" placement="top">
          <el-button 
            icon="el-icon-zoom-out" 
            @click="zoomOut"
          ></el-button>
        </el-tooltip>
        <el-divider direction="vertical"></el-divider>
        <el-tooltip :content="$t('knowledgeManage.graph.fitView')" placement="top">
          <el-button 
            icon="el-icon-full-screen" 
            @click="fitView"
          ></el-button>
        </el-tooltip>
        <el-divider direction="vertical"></el-divider>
        <el-tooltip :content="$t('knowledgeManage.graph.resetZoom')" placement="top">
          <el-button 
            icon="el-icon-refresh-left" 
            @click="resetZoom"
          ></el-button>
        </el-tooltip>
        <el-divider direction="vertical"></el-divider>
        <el-tooltip :content="$t('knowledgeManage.graph.refresh')" placement="top">
          <el-button 
            icon="el-icon-refresh" 
            :loading="loading"
            @click="refreshData"
          ></el-button>
        </el-tooltip>
      </div>
      <div ref="graphContainer" class="graph-container" v-loading="loading"></div>
    </div>
  </div>
</template>

<script>
import G6 from '@antv/g6'
import { transformGraphData } from '@/utils/graphDataTransform'

export default {
  name: 'GraphMap',
  props: {
    data: {
      type: Object,
      default: function() {
        return {
        }
      }
    },
    showHeader:{
      type:Boolean,
      default:true
    },
    showToolBar:{
      type:Boolean,
      default:true
    },
    knowledgeId: {
      type: [String, Number],
      default: null
    },
    autoFit: {
      type: Boolean,
      default: true
    },
    layout: {
      type: String,
      default: 'force'
    },
    nodeStyle: {
      type: Object,
      default: function() {
        return {}
      }
    },
    edgeStyle: {
      type: Object,
      default: function() {
        return {}
      }
    },
    performance: {
      type: Object,
      default: function() {
        return {
          largeThreshold: 500,
          hideLabelZoom: 0.5,
          enableWorkerLayout: true,
          simplifyEdgeOnLarge: true,
          disableAnimateOnLarge: true
        }
      }
    }
  },
  data() {
    return {
      graph: null,
      loading: false,
      zoom: 1,
      lastLabelSwitchRAF: 0,
      graphData:{
        nodes:[],
        edges:[]
      }
    }
  },
  computed: {
    headerName() {
      if (this.$route && this.$route.query && this.$route.query.name) {
        return this.$route.query.name
      }
    },
    zoomPercentage() {
      return Math.round(this.zoom * 100)
    }
  },
  watch: {
    data: {
      handler(newData) {
        // If graph 还没Init，Await mounted 后 of  initGraph Process
        if (!this.graph) {
          return
        }
        
        if (newData && newData.graph) {
          const transformedData = transformGraphData(newData.graph)
          this.graphData = transformedData
          this.updateGraphData(transformedData)
        }
      },
      deep: true,
      immediate: false
    }
  },
  mounted() {
    this.initGraph()
  },
  beforeDestroy() {
    if (this.graph) {
      this.graph.destroy()
      this.graph = null
    }
    window.removeEventListener('resize', this.handleResize)
  },
  methods: {
    goBack(){
      this.$emit('goBack')
    },
    initGraph() {
      if (!this.$refs.graphContainer) {
        return
      }

      const container = this.$refs.graphContainer
      const width = container.clientWidth || 800
      const height = container.clientHeight || 600

      this.registerCustomNodes()
      const Graph = G6.Graph || G6

      const graphConfig = {
        container,
        width,
        height,
        animate: false,
        modes: {
          default: [
            'drag-canvas',
            'zoom-canvas',
            'drag-node',
            'click-select',
            'brush-select'
          ]
        },
        layout: {
          type: this.layout,
          preventOverlap: true,
          nodeSize: 50,
          nodeSpacing: 60,
          autoFit: 'view',
          ...this.getLayoutConfig()
        },
        defaultNode: {
          type: 'circle',
          size: 50,
          labelCfg: {
            position: 'center',
            offset: 0,
            style: {
              fill: '#333',
              fontSize: 12,
              fontWeight: 'normal'
            }
          },
          style: {
            fill: '#e4e4fe',
            stroke: '#d4d0e9',
            lineWidth: 2
          },
          ...this.nodeStyle
        },
        defaultEdge: {
          type: 'line',
          style: {
            stroke: '#c7cad9',
            lineWidth: 2,
            endArrow: false
          },
          labelCfg: {
            autoRotate: true,
            refY: -10,
            style: {
              fill: '#666',
              fontSize: 14,
              fontWeight: 'normal',
              background: {
                fill: '#fff',
                padding: [2, 4, 2, 4],
                radius: 2,
                stroke: '#e4e7ed',
                lineWidth: 1
              }
            }
          },
          ...this.edgeStyle
        },
        nodeStateStyles: {
          hover: {
            fill: '#bebefe',
            stroke: '#d4d0e9',
            lineWidth: 2
          },
          selected: {
            fill: '#bebefe',
            stroke: '#c7cad9',
            lineWidth: 2
          }
        },
        edgeStateStyles: {
          hover: {
            stroke: '#c7cad9',
            lineWidth: 3
          },
          selected: {
            stroke: '#c7cad9',
            lineWidth: 3
          }
        }
      }

      this.graph = new Graph(graphConfig)
 
      this.bindEvents()

      if (
        this.data &&
        this.data.graph &&
        (Array.isArray(this.data.graph.nodes) ||
          Array.isArray(this.data.graph.edges))
      ) {
        const transformedData = transformGraphData(this.data.graph)
        this.graph.data(transformedData)
        this.graph.render()
        
        this.$nextTick(() => {
          // 适应View，让图谱占满整个屏幕
          this.graph.fitView(20, true)
          // Setting缩放比例为100%
          this.zoom = 1
        })
      }

      window.addEventListener('resize', this.handleResize)
    },

    registerCustomNodes() {
    },
 
    getLayoutConfig() {
      const configs = {
        force: {
          preventOverlap: true,
          nodeSize: 50,
          nodeSpacing: 60,
          linkDistance: 30,
          nodeStrength: -50,
          edgeStrength: 0.8,
          collideStrength: 0.8,
          alpha: 0.3,
          alphaDecay: 0.02,
          alphaMin: 0.001
        },
        dagre: {
          rankdir: 'TB',
          nodesep: 60,
          ranksep: 60
        },
        circular: {
          radius: 150,
          startRadius: 10,
          endRadius: 200
        },
        radial: {
          unitRadius: 60,
          nodeSize: 50
        }
      }
      return configs[this.layout] || configs.force
    },
    bindEvents() {
       if (!this.graph) return
 
       this.graph.on('node:click', (e) => {
         const node = e.item
         this.$emit('node-click', node.getModel())
       })
 
       this.graph.on('node:mouseenter', (e) => {
         const node = e.item
         this.graph.setItemState(node, 'hover', true)
       })
 
       this.graph.on('node:mouseleave', (e) => {
         const node = e.item
         this.graph.setItemState(node, 'hover', false)
       })
 
       this.graph.on('edge:click', (e) => {
         const edge = e.item
         this.$emit('edge-click', edge.getModel())
       })
 
       this.graph.on('canvas:click', () => {
         this.graph.getNodes().forEach(node => {
           this.graph.setItemState(node, 'selected', false)
         })
         this.graph.getEdges().forEach(edge => {
           this.graph.setItemState(edge, 'selected', false)
         })
       })
 
       this.graph.on('viewportchange', () => {
         if (this.graph) {
           const hideLabelZoom = (this.performance && this.performance.hideLabelZoom) || 0.5
           const minZoom = hideLabelZoom > 0 ? hideLabelZoom : 0.5
           let zoom = this.graph.getZoom()
           const maxZoom = 2
           // If缩放Less ThanMin限制，自动调整到Min缩放，以画布中心为缩放中心
           if (zoom < minZoom) {
             zoom = minZoom
             const width = this.graph.getWidth()
             const height = this.graph.getHeight()
             const centerPoint = {
               x: width / 2,
               y: height / 2
             }
             this.graph.zoomTo(zoom, centerPoint)
           } else if (zoom > maxZoom) {
             zoom = maxZoom
             const width = this.graph.getWidth()
             const height = this.graph.getHeight()
             const centerPoint = {
               x: width / 2,
               y: height / 2
             }
             this.graph.zoomTo(zoom, centerPoint)
           }
          this.zoom = zoom
          this.$emit('zoom-change', zoom)
          this.toggleLabelsByZoom()
        }
      })
    },

    updateGraphData(data) {
      if (!this.graph) return

      const safeData = {
        nodes: Array.isArray(data && data.nodes)
          ? data.nodes.map(n => ({ ...n }))
          : [],
        edges: Array.isArray(data && data.edges)
          ? data.edges.map(e => ({ ...(e || {}) }))
          : []
      }
      this.graph.data(safeData)
      this.graph.render()
      
      this.$nextTick(() => {
        if (this.autoFit) {
          // Setting缩放比例为100%
          this.zoom = 1
        } else {
          // If不自动适应，只Setting缩放为100%
          const graphWidth = this.graph.getWidth()
          const graphHeight = this.graph.getHeight()
          const centerPoint = {
            x: graphWidth / 2,
            y: graphHeight / 2
          }
          this.graph.zoomTo(1, centerPoint)
          this.zoom = 1
        }
      })
    },
 
    toggleLabelsByZoom() {
      if (!this.graph) return
      const threshold = (this.performance && this.performance.hideLabelZoom) || 0
      if (threshold <= 0) return
      const now = (typeof performance !== 'undefined' && performance.now) ? performance.now() : Date.now()
      if (now - this.lastLabelSwitchRAF < 16) return
      this.lastLabelSwitchRAF = now
      const shouldHide = this.zoom < threshold
      this.graph.setAutoPaint(false)
      this.graph.getNodes().forEach((n) => {
        const model = n.getModel()
        const hasLabel = !!model.originalLabel
        const currentLabel = model.label
        if (shouldHide && currentLabel) {
          this.graph.updateItem(n, { label: '' })
        } else if (!shouldHide && !currentLabel && hasLabel) {
          this.graph.updateItem(n, { label: model.originalLabel })
        }
      })
      this.graph.setAutoPaint(true)
      this.graph.paint()
    },

    isLargeData(data) {
      const nodesLen = (data && data.nodes && data.nodes.length) || 0
      const threshold = (this.performance && this.performance.largeThreshold) || 500
      return nodesLen >= threshold
    },

    zoomIn() {
      if (!this.graph) return
      const currentZoom = this.graph.getZoom()
      const newZoom = Math.min(currentZoom * 1.2, 2)
      this.graph.zoomTo(newZoom)
    },
 
    zoomOut() {
      if (!this.graph) return
      const currentZoom = this.graph.getZoom()
      const hideLabelZoom = (this.performance && this.performance.hideLabelZoom) || 0.5
      // Min缩放限制为 hideLabelZoom，确保label能够Show
      const minZoom = hideLabelZoom > 0 ? hideLabelZoom : 0.5
      const calculatedZoom = currentZoom * 0.8
      const newZoom = Math.max(calculatedZoom, minZoom)
      
      // If达到Min缩放限制（That is计算出 of 缩放Less ThanMinValue），以画布中心为缩放中心
      const isAtMinLimit = calculatedZoom < minZoom
      if (isAtMinLimit) {
        const width = this.graph.getWidth()
        const height = this.graph.getHeight()
        const centerPoint = {
          x: width / 2,
          y: height / 2
        }
        this.graph.zoomTo(newZoom, centerPoint)
      } else {
        this.graph.zoomTo(newZoom)
      }
    },
 
    fitView() {
      if (!this.graph) return
      this.graph.fitView(20)
    },
 
    resetZoom() {
      if (!this.graph) return
      this.graph.zoomTo(1)
      this.fitView()
    },
 
    refreshData() {
      this.loading = true
      this.$emit('refresh', () => {
        this.loading = false
        if (this.autoFit) {
          this.$nextTick(() => {
            this.fitView()
          })
        }
      })
    },
     
    finishRefresh() {
      this.loading = false
      if (this.autoFit && this.graph) {
        this.$nextTick(() => {
          this.fitView()
        })
      }
    },
 
    handleResize() {
      if (!this.graph || !this.$refs.graphContainer) return
      const container = this.$refs.graphContainer
      const width = container.clientWidth
      const height = container.clientHeight
      
      this.graph.changeSize(width, height)
    },
    getGraph() {
      return this.graph
    },
    downloadImage(fileName = 'graph') {
      if (!this.graph) return
      
      if (typeof this.graph.downloadFullImage === 'function') {
        this.graph.downloadFullImage(fileName, 'image/png', {
          backgroundColor: '#fff',
          padding: [10, 10, 10, 10]
        })
      } else if (typeof this.graph.downloadImage === 'function') {
        this.graph.downloadImage(fileName)
      }
    }
  }
}
</script>

<style lang="scss" scoped>
.graph-map-container {
  position: relative;
  width: 100%;
  height:100%;
  box-sizing: border-box;
  overflow: hidden;
  padding:10px;
  .graph-map-content{
     width: 100%;
     height:100%;
     background:#fff;
     border-radius:6px;
      .graph-header {
    width:calc(100% - 20px);
    padding:24px 0 14px 30px;
    border-bottom:1px solid #eeeeee;
    .back{
      font-size:20px;
      cursor:pointer;
    }
    .header-text {
      color: #434C6C;
      font-size: 18px;
      font-weight: bold;
      user-select: none;
    }
  }
  .graph-toolbar {
    position: absolute;
    bottom: 20px;
    left: 50%;
    transform: translateX(-50%);
    z-index: 10;
    display: flex;
    align-items: center;
    gap: 0;
    padding: 5px 24px;
    min-width: 400px;
    background: #ffffff;
    border: 1px solid $color;
    border-radius: 8px;
    box-shadow: none;
    .graphStatus{
      color:$color;
      display: flex;
      align-items: center;
      margin-right: 20px;
      font-size: 14px;
      .statusNum{
        display: flex;
        justify-content: center;
        align-items: center;
        border-radius: 4px;
        width: 20px;
        height: 20px;
        background: $color;
        color: #fff;
        margin-right: 6px;
        font-size: 12px;
        font-weight: bold;
      }
      .statusText{
        white-space: nowrap;
        font-size:14px;
      }
    }
    .el-button {
      margin: 0;
      padding: 8px 16px;
      color: $color;
      border-color: transparent;
      background-color: transparent;
      border-radius: 4px;
      font-size: 20px !important;
      height: auto;
      line-height: 1;
    }
    
    ::v-deep .el-button {
      i {
        font-size: 20px !important;
        color: $color !important;
        width: 20px !important;
        height: 20px !important;
      }
      
      [class*="el-icon-"] {
        font-size: 20px !important;
        color: $color !important;
      }
    }
    
    /deep/ .el-button {
      i {
        font-size: 22px !important;
        color: $color !important;
        width: 22px !important;
        height: 22px !important;
      }
      
      [class*="el-icon-"] {
        font-size: 22px !important;
        color: $color !important;
      }
    }
    
    .el-button:hover {
      color: #ffffff;
      background-color: $color;
    }
    
    ::v-deep .el-button:hover {
      i {
        color: #ffffff !important;
        font-size: 20px !important;
      }
      
      [class*="el-icon-"] {
        color: #ffffff !important;
        font-size: 20px !important;
      }
    }
    
    /deep/ .el-button:hover {
      i {
        color: #ffffff !important;
        font-size: 20px !important;
      }
      
      [class*="el-icon-"] {
        color: #ffffff !important;
        font-size: 20px !important;
      }
    }
    
    .el-button:focus {
      color: $color;
      background-color: transparent;
    }

    .el-divider {
      margin: 0 15px;
      height: 15px;
      background-color: $color;
    }
    .zoom-display {
      margin: 0 15px;
      font-size: 14px;
      color: $color;
      font-weight: 500;
      min-width: 45px;
      text-align: center;
    }
  }

  .graph-container {
    width: 100%;
    height: 100%;
  }
  }
}
</style>

<style lang="scss">
.graph-toolbar {
  .el-button {
    font-size: 20px !important;
  }
  
  .el-button i,
  .el-button [class*="el-icon-"] {
    font-size: 24px !important;
    width: 24px !important;
    height: 24px !important;
    line-height: 24px !important;
    display: inline-block !important;
  }
  
  .el-button:hover {
    background-color: $color !important;
    color: #ffffff !important;
    
    i,
    [class*="el-icon-"] {
      color: #ffffff !important;
      font-size: 24px !important;
    }
  }
  
  .el-button.is-loading i,
  .el-button.is-loading [class*="el-icon-"] {
    font-size: 24px !important;
  }
}
</style>

