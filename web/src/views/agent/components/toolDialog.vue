<template>
  <div>
    <el-dialog
      :title="$t('agent.toolDialog.addTool')"
      :visible.sync="dialogVisible"
      width="50%"
      :before-close="handleClose"
      custom-class="tool-dialog"
    >
      <div class="tool-typ">
        <div class="toolbtn">
          <div
            v-for="(item, index) in toolList"
            :key="index"
            @click="clickTool(item, index)"
            :class="[{ active: activeValue === item.value }]"
          >
            {{ item.name }}
            <span>/</span>
            {{ showToolNum(item.value) }}
          </div>
        </div>
        <el-input
          v-model="toolName"
          :placeholder="$t('agent.toolDialog.searchTool')"
          class="tool-input"
          suffix-icon="el-icon-search"
          @keyup.enter.native="searchTool"
          clearable
        ></el-input>
      </div>
      <div class="toolContent">
        <div @click="goCreate" class="createTool">
          <span class="el-icon-plus add"></span>
          {{ createText() }}
        </div>
        <template v-for="(items, type) in contentMap">
          <template v-if="activeValue === type">
            <div
              v-for="item in items"
              :key="item[type + 'Id'] || item.id"
              class="toolContent_item"
            >
              <template
                v-if="
                  type === AGENT_TOOL_TYPE.WORKFLOW ||
                  type === AGENT_TOOL_TYPE.SKILL
                "
              >
                <div class="tool_box">
                  <div class="tool_img">
                    <img
                      :src="avatarSrc(item.avatar.path)"
                      v-if="item.avatar && item.avatar.path"
                    />
                  </div>
                  <div class="tool-info">
                    <div class="tool-title-row">
                      <div class="tool-title">
                        {{
                          type === AGENT_TOOL_TYPE.SKILL
                            ? item.skillName
                            : item.name
                        }}
                      </div>
                      <template
                        v-if="type === AGENT_TOOL_TYPE.SKILL && item.author"
                      >
                        <el-divider direction="vertical"></el-divider>
                        <span class="tool-author">{{ item.author }}</span>
                      </template>
                    </div>
                    <span
                      class="tag"
                      v-if="
                        type === AGENT_TOOL_TYPE.WORKFLOW &&
                        tagMap[item.appType]
                      "
                    >
                      {{ tagMap[item.appType] }}
                    </span>
                    <div
                      class="tag skill-type-tag"
                      v-if="
                        type === AGENT_TOOL_TYPE.SKILL &&
                        skillTypeTagMap[item.skillType]
                      "
                    >
                      {{ skillTypeTagMap[item.skillType] }}
                    </div>
                    <el-tooltip
                      v-if="toolDescription(item)"
                      effect="dark"
                      :open-delay="300"
                      :content="toolDescription(item)"
                      placement="top-start"
                    >
                      <span class="tool-description">
                        {{ toolDescription(item) }}
                      </span>
                    </el-tooltip>
                  </div>
                </div>
                <div>
                  <el-button
                    type="text"
                    @click="openTool($event, item, type)"
                    v-if="!item.checked"
                  >
                    {{ $t('agent.toolDialog.add') }}
                  </el-button>
                  <el-button type="text" v-else style="color: #ccc">
                    {{ $t('agent.toolDialog.added') }}
                  </el-button>
                </div>
              </template>
              <el-collapse
                v-model="item.activeNames"
                @change="handleToolChange"
                v-else
                class="tool_collapse"
              >
                <el-collapse-item :name="item.toolId">
                  <template slot="title">
                    <div class="tool_img">
                      <img
                        :src="avatarSrc(item.avatar.path)"
                        v-if="item.avatar && item.avatar.path"
                      />
                    </div>
                    <div class="tool-name-container">
                      <div class="tool-name-main">
                        <div class="tool-title-line">
                          <h3 class="tool-name">{{ item.toolName }}</h3>
                          <svg-icon
                            v-if="isParentSelected(item, type)"
                            class-name="selected-parent-icon"
                            icon-class="circle-check"
                          />
                          <span
                            v-if="item.loading"
                            class="el-icon-loading loading-text"
                          ></span>
                        </div>
                        <span v-if="type === AGENT_TOOL_TYPE.TOOL" class="tag">
                          {{
                            item.toolType === 'builtin'
                              ? $t('agent.toolDialog.builtinTools')
                              : $t('agent.toolDialog.customTools')
                          }}
                        </span>
                        <el-tooltip
                          v-if="toolDescription(item)"
                          effect="dark"
                          :open-delay="300"
                          :content="toolDescription(item)"
                          placement="top-start"
                        >
                          <span class="tool-description">
                            {{ toolDescription(item) }}
                          </span>
                        </el-tooltip>
                      </div>
                    </div>
                  </template>
                  <template v-if="item.children && item.children.length">
                    <div
                      v-for="(tool, index) in item.children"
                      class="tool-action-item"
                      :key="'tool' + index"
                    >
                      <div style="padding-right: 5px">
                        <p>
                          <span>{{ tool.name }}</span>
                          <el-tooltip
                            class="item"
                            effect="dark"
                            :content="tool.description"
                            placement="top-start"
                            v-if="
                              tool.description && tool.description.length > 0
                            "
                          >
                            <span class="el-icon-info desc-info"></span>
                          </el-tooltip>
                        </p>
                      </div>
                      <div>
                        <el-button
                          type="text"
                          @click="openTool($event, item, type, tool)"
                          v-if="!tool.checked"
                        >
                          {{ $t('agent.toolDialog.add') }}
                        </el-button>
                        <el-button type="text" v-else style="color: #ccc">
                          {{ $t('agent.toolDialog.added') }}
                        </el-button>
                      </div>
                    </div>
                  </template>
                </el-collapse-item>
              </el-collapse>
            </div>
          </template>
        </template>
      </div>
    </el-dialog>
  </div>
</template>
<script>
import {
  addWorkFlowInfo,
  addMcp,
  addCustomBuiltIn,
  addSkill,
  toolList,
  toolActionList,
  mcptoolList,
  mcpActionList,
  getWorkflowList,
} from '@/api/agent';
import { avatarSrc } from '@/utils/util';
import { getSkillSelectList } from '@/api/templateSquare';
import { AGENT_TOOL_TYPE } from '@/views/agent/constants';
export default {
  props: ['assistantId'],
  data() {
    return {
      AGENT_TOOL_TYPE,
      toolName: '',
      dialogVisible: false,
      toolIndex: 0,
      activeValue: AGENT_TOOL_TYPE.TOOL,
      workFlowInfos: [],
      mcpInfos: [],
      skillInfos: [],
      customInfos: [],
      mcpList: [],
      workFlowList: [],
      customList: [],
      skillList: [],
      builtInInfos: [],
      customCount: 0,
      mcpCount: 0,
      workflowCount: 0,
      skillCount: 0,
      toolList: [
        {
          value: AGENT_TOOL_TYPE.TOOL,
          name: this.$t('agent.toolDialog.tool'),
        },
        {
          value: AGENT_TOOL_TYPE.MCP,
          name: 'MCP',
        },
        {
          value: AGENT_TOOL_TYPE.WORKFLOW,
          name: this.$t('appSpace.workflow'),
        },
        {
          value: AGENT_TOOL_TYPE.SKILL,
          name: 'Skills',
        },
      ],
    };
  },
  computed: {
    contentMap() {
      return {
        [AGENT_TOOL_TYPE.TOOL]: this.customInfos,
        builtIn: this.builtInInfos,
        [AGENT_TOOL_TYPE.MCP]: this.mcpInfos,
        [AGENT_TOOL_TYPE.WORKFLOW]: this.workFlowInfos,
        [AGENT_TOOL_TYPE.SKILL]: this.skillInfos,
      };
    },
    tagMap() {
      return {
        [AGENT_TOOL_TYPE.WORKFLOW]: this.$t('appSpace.workflow'),
        chatflow: this.$t('appSpace.chat'),
      };
    },
    skillTypeTagMap() {
      return {
        acquired: this.$t('agent.toolDialog.skillAcquired'),
        custom: this.$t('agent.toolDialog.skillCustom'),
        builtin: this.$t('agent.toolDialog.skillBuiltin'),
      };
    },
  },
  created() {
    this.getMcpSelect('');
    this.getWorkflowList('');
    this.getCustomList('');
    this.getSkillList('');
  },
  methods: {
    avatarSrc,
    toolDescription(item) {
      return item.description || item.desc || '';
    },
    showToolNum(type) {
      if (type === AGENT_TOOL_TYPE.TOOL) {
        return this.customCount;
      } else if (type === AGENT_TOOL_TYPE.MCP) {
        return this.mcpCount;
      } else if (type === AGENT_TOOL_TYPE.WORKFLOW) {
        return this.workflowCount;
      } else {
        return this.skillCount;
      }
    },
    handleToolChange(id) {
      let toolId = id[0];
      if (this.activeValue === AGENT_TOOL_TYPE.TOOL) {
        const targetItem = this.customInfos.find(
          item => item.toolId === toolId,
        );
        if (targetItem) {
          const { toolId, toolType } = targetItem;
          const index = this.customInfos.findIndex(
            item => item.toolId === toolId,
          );
          this.getToolAction(toolId, toolType, index);
        }
      } else if (this.activeValue === AGENT_TOOL_TYPE.MCP) {
        const targetItem = this.mcpInfos.find(item => item.toolId === toolId);
        if (targetItem) {
          const { toolId, toolType } = targetItem;
          const index = this.mcpInfos.findIndex(item => item.toolId === toolId);
          this.getMcpAction(toolId, toolType, index);
        }
      }
    },
    sortBySelected(list, selectedIndexGetter) {
      return list
        .map((item, index) => ({
          item,
          index,
          selectedIndex: selectedIndexGetter(item),
        }))
        .sort((a, b) => {
          const aSelected = a.selectedIndex > -1;
          const bSelected = b.selectedIndex > -1;
          if (aSelected && bSelected) {
            return a.selectedIndex - b.selectedIndex;
          }
          if (aSelected !== bSelected) {
            return aSelected ? -1 : 1;
          }
          return a.index - b.index;
        })
        .map(({ item }) => item);
    },
    getToolSelectedIndex(toolId) {
      return this.customList.findIndex(item => item.toolId === toolId);
    },
    getMcpSelectedIndex(toolId) {
      return this.mcpList.findIndex(item => item.mcpId === toolId);
    },
    isParentSelected(item, type) {
      if (type === AGENT_TOOL_TYPE.TOOL) {
        return this.getToolSelectedIndex(item.toolId) > -1;
      }
      if (type === AGENT_TOOL_TYPE.MCP) {
        return this.getMcpSelectedIndex(item.toolId) > -1;
      }
      return false;
    },
    getWorkflowSelectedIndex(appId) {
      return this.workFlowList.findIndex(item => item.workFlowId === appId);
    },
    getSkillSelectedIndex(skillId) {
      return this.skillList.findIndex(item => item.skillId === skillId);
    },
    getToolActionSelectedIndex(toolId, actionName) {
      return this.customList.findIndex(
        item => item.toolId === toolId && item.actionName === actionName,
      );
    },
    getMcpActionSelectedIndex(toolId, actionName) {
      return this.mcpList.findIndex(
        item => item.mcpId === toolId && item.actionName === actionName,
      );
    },
    getCustomList(name) {
      //获取自定义和内置工具
      toolList({ name }).then(res => {
        if (res.code === 0) {
          const list = (res.data.list || []).map(m => ({
            ...m,
            loading: false,
            activeNames: [],
            children: [],
          }));
          this.customInfos = this.sortBySelected(list, item =>
            this.getToolSelectedIndex(item.toolId),
          );
        }
      });
    },
    getToolAction(toolId, toolType, index) {
      this.$set(this.customInfos[index], 'loading', true);
      toolActionList({ toolId, toolType })
        .then(res => {
          if (res.code === 0) {
            this.$set(this.customInfos[index], 'children', res.data.actions);
            this.$set(this.customInfos[index], 'loading', false);
            this.customInfos[index]['children'].forEach(m => {
              m.checked = this.customList.some(
                item => item.actionName === m.name && item.toolId === toolId,
              );
            });
            this.$set(
              this.customInfos[index],
              'children',
              this.sortBySelected(this.customInfos[index]['children'], item =>
                this.getToolActionSelectedIndex(toolId, item.name),
              ),
            );
          }
        })
        .catch(() => {
          this.$set(this.customInfos[index], 'loading', false);
        });
    },
    goCreate() {
      if (this.activeValue === AGENT_TOOL_TYPE.TOOL) {
        this.$router.push({ path: '/tool?tool=custom' });
      } else if (this.activeValue === AGENT_TOOL_TYPE.MCP) {
        this.$router.push({ path: '/mcpService?mcp=integrate' });
      } else if (this.activeValue === AGENT_TOOL_TYPE.WORKFLOW) {
        this.$router.push({ path: '/appSpace/workflow' });
      } else {
        this.$router.push({ path: '/skill?type=skill_custom' });
      }
    },
    createText() {
      if (this.activeValue === AGENT_TOOL_TYPE.TOOL) {
        return this.$t('agent.toolDialog.createAutoTool');
      } else if (this.activeValue === AGENT_TOOL_TYPE.MCP) {
        return this.$t('agent.toolDialog.importMcp');
      } else if (this.activeValue === AGENT_TOOL_TYPE.WORKFLOW) {
        return this.$t('agent.toolDialog.createWorkflow');
      } else {
        return this.$t('agent.toolDialog.addSkill');
      }
    },
    openTool(e, item, type, action) {
      if (!e) return;
      if (type === AGENT_TOOL_TYPE.WORKFLOW) {
        this.addWorkFlow(item);
      } else if (type === AGENT_TOOL_TYPE.SKILL) {
        this.addSkillItem(item);
      } else if (type === AGENT_TOOL_TYPE.MCP) {
        this.addMcpItem(item, action);
      } else {
        if (item.needApiKeyInput && !item.apiKey.length) {
          this.$message.warning(this.$t('agent.toolDialog.errorApiKey'));
        }
        this.addCustomBuiltIn(item, action);
      }
    },
    addCustomBuiltIn(n, action) {
      //添加自定义工具和内置工具
      addCustomBuiltIn({
        assistantId: this.assistantId,
        actionName: action.name,
        toolId: n.toolId,
        toolType: n.toolType,
      }).then(res => {
        if (res.code === 0) {
          this.$set(action, 'checked', true);
          this.customCount++;
          this.$forceUpdate();
          this.$message.success(this.$t('agent.toolDialog.addSuccess'));
          this.$emit('updateDetail');
        }
      });
    },
    addMcpItem(n, action) {
      addMcp({
        assistantId: this.assistantId,
        actionName: action.name,
        mcpId: n.toolId,
        mcpType: n.toolType,
      }).then(res => {
        if (res.code === 0) {
          this.$set(action, 'checked', true);
          this.mcpCount++;
          this.$forceUpdate();
          this.$message.success(this.$t('agent.toolDialog.addSuccess'));
          this.$emit('updateDetail');
        }
      });
    },
    addWorkFlow(n) {
      this.doCreateWorkFlow(n, n.appId);
    },
    async doCreateWorkFlow(n, workFlowId, schema) {
      let params = {
        assistantId: this.assistantId,
        workFlowId,
      };
      let res = await addWorkFlowInfo(params);
      if (res.code === 0) {
        n.checked = true;
        this.workflowCount++;
        this.$message.success(this.$t('agent.addWorkFlowTips'));
        this.$emit('updateDetail');
      }
    },
    // 添加skill
    addSkillItem(n) {
      addSkill({
        assistantId: this.assistantId,
        skillId: n.skillId,
        skillType: n.skillType,
      }).then(res => {
        if (res.code === 0) {
          const isFirst = this.skillCount === 0;
          this.$set(n, 'checked', true);
          this.skillCount++;
          this.$forceUpdate();
          if (isFirst) {
            this.$message.success(
              this.$t('agent.toolDialog.firstAddSkillTips'),
            );
          } else {
            this.$message.success(this.$t('agent.toolDialog.addSkillSuccess'));
          }
          this.$emit('updateDetail');
        }
      });
    },
    searchTool() {
      if (this.activeValue === AGENT_TOOL_TYPE.TOOL) {
        this.getCustomList(this.toolName);
      } else if (this.activeValue === AGENT_TOOL_TYPE.MCP) {
        this.getMcpSelect(this.toolName);
      } else if (this.activeValue === AGENT_TOOL_TYPE.WORKFLOW) {
        this.getWorkflowList(this.toolName);
      } else {
        this.getSkillList(this.toolName);
      }
    },
    getMcpSelect(name) {
      //获取mcp工具
      mcptoolList({ name }).then(res => {
        if (res.code === 0) {
          const list = (res.data.list || []).map(m => ({
            ...m,
            children: [],
            activeNames: [],
            loading: false,
          }));
          this.mcpInfos = this.sortBySelected(list, item =>
            this.getMcpSelectedIndex(item.toolId),
          );
        }
      });
    },
    getMcpAction(toolId, toolType, index) {
      this.$set(this.mcpInfos[index], 'loading', true);
      mcpActionList({ toolId, toolType })
        .then(res => {
          if (res.code === 0) {
            this.$set(this.mcpInfos[index], 'children', res.data.actions);
            this.$set(this.mcpInfos[index], 'loading', false);
            this.mcpInfos[index]['children'].forEach(m => {
              m.checked = this.mcpList.some(
                item => item.actionName === m.name && item.mcpId === toolId,
              );
            });
            this.$set(
              this.mcpInfos[index],
              'children',
              this.sortBySelected(this.mcpInfos[index]['children'], item =>
                this.getMcpActionSelectedIndex(toolId, item.name),
              ),
            );
          }
        })
        .catch(() => {
          this.$set(this.mcpInfos[index], 'loading', false);
        });
    },
    getWorkflowList(name) {
      getWorkflowList({ name }).then(res => {
        if (res.code === 0) {
          const list = (res.data.list || []).map(m => ({
            ...m,
            checked: this.workFlowList.some(
              item => item.workFlowId === m.appId,
            ),
          }));
          this.workFlowInfos = this.sortBySelected(list, item =>
            this.getWorkflowSelectedIndex(item.appId),
          );
        }
      });
    },
    getSkillList(name) {
      getSkillSelectList({ name }).then(res => {
        if (res.code === 0) {
          const list = (res.data.list || []).map(m => ({
            ...m,
            checked: this.skillList.some(item => item.skillId === m.skillId),
          }));
          this.skillInfos = this.sortBySelected(list, item =>
            this.getSkillSelectedIndex(item.skillId),
          );
        }
      });
    },
    syncToolsSelected(row = {}) {
      this.mcpList = row.mcpInfos || [];
      this.workFlowList = row.workFlowInfos || [];
      this.customList = row.customInfos || [];
      this.skillList = row.skillInfos || [];
      this.customCount = this.customList.length;
      this.mcpCount = this.mcpList.length;
      this.workflowCount = this.workFlowList.length;
      this.skillCount = this.skillList.length;
      this.refreshCheckedStatus();
    },
    refreshCheckedStatus() {
      this.workFlowInfos.forEach(item => {
        this.$set(
          item,
          'checked',
          this.workFlowList.some(
            selected => selected.workFlowId === item.appId,
          ),
        );
      });
      this.skillInfos.forEach(item => {
        this.$set(
          item,
          'checked',
          this.skillList.some(selected => selected.skillId === item.skillId),
        );
      });
      this.customInfos.forEach(item => {
        (item.children || []).forEach(action => {
          this.$set(
            action,
            'checked',
            this.customList.some(
              selected =>
                selected.toolId === item.toolId &&
                selected.actionName === action.name,
            ),
          );
        });
      });
      this.mcpInfos.forEach(item => {
        (item.children || []).forEach(action => {
          this.$set(
            action,
            'checked',
            this.mcpList.some(
              selected =>
                selected.mcpId === item.toolId &&
                selected.actionName === action.name,
            ),
          );
        });
      });
    },
    showDialog(row) {
      this.dialogVisible = true;
      this.syncToolsSelected(row);
    },
    handleClose() {
      this.toolIndex = -1;
      this.activeValue = AGENT_TOOL_TYPE.TOOL;
      this.dialogVisible = false;
    },
    clickTool(item, i) {
      this.toolIndex = i;
      this.activeValue = item.value;
      if (this.activeValue === AGENT_TOOL_TYPE.TOOL) {
        this.getCustomList('');
      } else if (this.activeValue === AGENT_TOOL_TYPE.MCP) {
        this.getMcpSelect('');
      } else if (this.activeValue === AGENT_TOOL_TYPE.WORKFLOW) {
        this.getWorkflowList('');
      } else {
        this.getSkillList('');
      }
    },
  },
};
</script>
<style lang="scss">
.tool-dialog {
  min-width: 700px;
}
</style>
<style lang="scss" scoped>
::v-deep {
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
    height: auto;
    line-height: normal;
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

    .tag {
      padding: 0 8px;
      background: rgba(139, 139, 149, 0.15);
      color: #4b4a58;
      font-size: 12px;
      border-radius: 6px;
      width: fit-content;
    }
    .tool_box {
      display: flex;
      align-items: center;
      min-width: 0;
      flex: 1;
      margin-right: 20px;
    }
    .tool-info {
      display: flex;
      flex-direction: column;
      gap: 4px;
      min-width: 0;
      flex: 1;
    }
    .tool-title-row {
      display: flex;
      align-items: center;
      line-height: 1.5;
      min-width: 0;
    }
    .tool-title {
      color: #333;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    .tool-author {
      color: #666;
      font-size: 13px;
    }
    .skill-type-tag {
      width: fit-content;
      line-height: 1.5;
    }
    .tool-description {
      display: block;
      min-width: 0;
      overflow: hidden;
      color: #8b8b95;
      font-size: 12px;
      line-height: 1.5;
      text-align: left;
      text-overflow: ellipsis;
      white-space: nowrap;
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
      color: var($color);
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

.tool-name-container {
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-width: 0;
  width: calc(100% - 48px);
  margin-right: 20px;
  .tool-name-main {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 0;
  }
  .tool-title-line {
    display: flex;
    align-items: center;
    min-width: 0;
  }
  .tool-name {
    height: auto;
    line-height: 1.5;
    margin: 0;
  }
  .selected-parent-icon {
    flex-shrink: 0;
    margin-left: 6px;
    color: #18a058;
    font-size: 15px;
  }
  .tag {
    height: auto;
    line-height: 1.5;
    width: fit-content;
  }
  .tool-description {
    display: block;
    min-width: 0;
    overflow: hidden;
    color: #8b8b95;
    font-size: 12px;
    line-height: 1.5;
    text-align: left;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}
</style>
