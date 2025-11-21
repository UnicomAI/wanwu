<template>
  <div>
    <el-dialog
      :title="$t('tool.server.bind.title')"
      :visible.sync="dialogVisible"
      width="50%"
      :before-close="handleClose">
      <div class="tool-typ">
        <div class="toolbtn">
          <div
            v-for="(item,index) in toolList" :key="index" @click="clickTool(item)"
            :class="[{'active':activeValue === item.value}]">
            {{ item.name }}
          </div>
        </div>
        <el-input
          v-model="toolName" :placeholder="$t('tool.server.bind.search')" class="tool-input"
          suffix-icon="el-icon-search"
          @keyup.enter.native="searchTool" clearable/>
      </div>
      <div class="toolContent">
        <div @click="goCreate" class="createTool"><span class="el-icon-plus add"></span>{{ createText() }}</div>
        <template v-for="(items, type) in contentMap">
          <div
            v-if="activeValue === type"
            v-for="(item,i) in items"
            :key="item[type + 'Id'] || item.id"
            class="toolContent_item"
          >
            <el-collapse @change="handleToolChange" class="tool_collapse">
              <el-collapse-item :name="item.toolId">
                <template slot="title">
                  <div class="tool_img">
                    <img :src="'/user/api/' + item.avatar.path" v-show="item.avatar && item.avatar.path"/>
                  </div>
                  <h3>{{ item.toolName }}</h3>
                  <span v-if="item.loading" class="el-icon-loading loading-text"></span>
                </template>
                <template v-if="item.children && item.children.length">
                  <div v-for="(tool,index) in item.children" class="tool-action-item" :key="TOOL+index">
                    <div style="padding-right:5px;">
                      <p>
                        <span>{{ tool.name }}</span>
                        <el-tooltip class="item" effect="dark" :content="tool.description" placement="top-start"
                                    v-if="tool.description && tool.description.length > 0">
                          <span class="el-icon-info desc-info"></span>
                        </el-tooltip>
                      </p>
                    </div>
                    <div>
                      <el-button type="text" @click="openTool($event,item,type,tool)" v-if="!tool.checked">
                        {{ $t('tool.server.bind.add') }}
                      </el-button>
                      <el-button type="text" v-else style="color:#ccc;">
                        {{ $t('tool.server.bind.added') }}
                      </el-button>
                    </div>
                  </div>
                </template>
              </el-collapse-item>
            </el-collapse>
          </div>
        </template>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import {addServerTool} from "@/api/mcp";
import {toolActionList, toolList} from "@/api/agent";
import {MCP, PROMPT, TOOL} from "@/views/tool/constants";

export default {
  data() {
    return {
      mcpServerId: '',
      toolName: '',
      dialogVisible: false,
      activeValue: TOOL,
      customInfos: [],
      customList: [],
      toolList: [
        {
          value: TOOL,
          name: this.$t('menu.app.tool'),
        },
        // {
        //   value: 'agent',
        //   name: 'Agent'
        // },
        // {
        //   value: 'workflow',
        //   name: 'Workflow'
        // },
        // {
        //   value: 'rag',
        //   name: 'Text Q&A'
        // },
        // {
        //   value: 'builtin',
        //   name: '内置Tool'
        // },
      ]
    }
  },
  computed: {
    contentMap() {
      return {
        tool: this.customInfos,
      }
    },
    TOOL() {
      return TOOL
    },
  },
  created() {
    this.getCustomList('')
  },
  methods: {
    handleToolChange(id) {
      let toolId = id[0];
      if (this.activeValue === TOOL) {
        const targetItem = this.customInfos.find(item => item.toolId === toolId)
        if (targetItem) {
          const {toolId, toolType} = targetItem
          const index = this.customInfos.findIndex(item => item.toolId === toolId)
          this.getToolAction(toolId, toolType, index)
        }
      }
    },
    getToolAction(toolId, toolType, index) {
      this.$set(this.customInfos[index], 'loading', true)
      toolActionList({toolId, toolType}).then(res => {
        if (res.code === 0) {
          this.$set(this.customInfos[index], 'children', res.data.actions)
          this.$set(this.customInfos[index], 'loading', false)
          this.customInfos[index]['children'].forEach(m => {
            m.checked = this.customList.some(item => item.methodName === m.name && item.id === toolId)
          })

        }
      }).catch(() => {
        this.$set(this.customInfos[index], 'loading', false)
      })
    },
    openTool(e, item, type, action) {
      if (!e) return;
      if (type === TOOL) {
        if (item.needApiKeyInput && !item.apiKey.length) {
          this.$message.warning(this.$t('tool.server.bind.apiWarning'))
        }
        this.addCustomTool(item, action)
      }
    },
    showDialog(detail) {
      this.mcpServerId = detail.mcpServerId
      this.customList = detail.tools.filter(tool => tool.type === 'custom' || tool.type === 'builtin');

      this.customInfos.forEach((item, index) => {
        if (item.children && item.children.length > 0) {
          const toolId = item.toolId;
          this.customInfos[index]['children'].forEach(m => {
            m.checked = this.customList.some(item => item.methodName === m.name && item.id === toolId)
          });
        }
      })
      this.dialogVisible = true
    },
    getCustomList(name) {
      toolList({name}).then(res => {
        if (res.code === 0) {
          this.customInfos = (res.data.list || []).map(item => ({
            ...item,
            loading: false,
            children: []
          }))
        }
      })
    },
    goCreate() {
      if (this.activeValue === TOOL) {
        this.$router.push({path: '/tool?type=tool&mcp=custom'})
      }
    },
    createText() {
      if (this.activeValue === TOOL) {
        return this.$t('common.button.add') + this.$t('menu.app.tool')
      }
    },
    addCustomTool(item, method) {
      const params = {
        id: item.toolId,
        type: item.toolType,
        methodName: method.name,
        mcpServerId: this.mcpServerId
      }
      addServerTool(params).then(res => {
        if (res.code === 0) {
          this.$set(method, 'checked', true);
          this.$message.success(this.$t('tool.server.bind.addMsg'));
          this.$emit('handleFetch');
          this.$nextTick(() => {
            this.$forceUpdate();
          });
        }
      })
    },
    searchTool() {
      if (this.activeValue === TOOL) {
        this.getCustomList(this.toolName)
      }
    },
    handleClose() {
      this.activeValue = TOOL
      this.dialogVisible = false;
    },
    clickTool(item) {
      this.activeValue = item.value;
      if (this.activeValue === TOOL) {
        this.getCustomList('');
      }
    },
  }
}
</script>

<style lang="scss" scoped>
/deep/ {
  .el-dialog__body {
    padding: 10px 20px;
  }

  .tool_collapse {
    width: 100% !important;
    border: none !important;
  }

  .el-collapse-item__header {
    background: none !important;
    border-bottom: none !important;
  }

  .el-collapse-item__wrap {
    border-bottom: none !important;
    background: none !important;
  }

  .el-collapse-item__content {
    padding-bottom: 0 !important;
  }

}

.createTool {
  padding: 10px;
  cursor: pointer;

  .add {
    padding-right: 5px;
  }
}

.createTool:hover {
  color: $color;
}

.tool-typ {
  display: flex;
  justify-content: space-between;
  padding: 10px 0;
  border-bottom: 1px solid #dbdbdb;

  .toolbtn {
    display: flex;
    justify-content: flex-start;
    gap: 20px;

    div {
      text-align: center;
      padding: 5px 20px;
      border-radius: 6px;
      border: 1px solid #ddd;
      cursor: pointer;
    }
  }

  .tool-input {
    width: 200px;
  }
}

.toolContent {
  padding: 10px 0;
  max-height: 400px;
  overflow-y: auto;

  .toolContent_item {
    padding: 5px 20px;
    border: 1px solid #dbdbdb;
    border-radius: 6px;
    margin-bottom: 10px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: space-between;

    .tool_box {
      display: flex;
      align-items: center;
    }

    .tool_img {
      width: 35px;
      height: 35px;
      background: #eee;
      border-radius: 50%;
      display: inline-block;
      margin-right: 5px;

      img {
        width: 100%;
        height: 100%;
        border-radius: 50%;
        object-fit: cover;
      }
    }

    .loading-text {
      margin-left: 4px;
      color: var($color)
    }

    .tool-action-item {
      display: flex;
      align-items: center;
      justify-content: space-between;
      border-top: 1px solid #eee;
      padding: 5px 0;

      .desc-info {
        color: #ccc;
        margin-left: 4px;
        cursor: pointer;
      }
    }
  }

  .toolContent_item:hover {
    background: $color_opacity;
  }
}

.active {
  border: 1px solid $color !important;
  color: #fff;
  background: $color;
}
</style>