<template>
  <div class="organization-tree-select">
    <el-input
      v-model="keyword"
      class="organization-search"
      clearable
      size="small"
      prefix-icon="el-icon-search"
      :placeholder="$t('adminCenter.common.searchOrg')"
    />
    <div v-loading="loading" class="organization-tree-wrap">
      <div
        v-if="organizationTreeData.length === 0 && !loading"
        class="organization-tree-empty"
      >
        {{ $t('common.noData') }}
      </div>
      <el-tree
        v-else
        ref="organizationTree"
        class="organization-tree"
        :data="treeData"
        :props="treeProps"
        :node-key="nodeKey"
        :current-node-key="currentKey"
        :default-checked-keys="value"
        highlight-current
        show-checkbox
        check-strictly
        :default-expand-all="defaultExpandAll"
        :default-expanded-keys="defaultExpandedKeys"
        :expand-on-click-node="false"
        @check="handleCheck"
        @node-click="handleNodeClick"
      >
        <span slot-scope="{ data }" class="organization-tree-node">
          <el-tooltip
            :content="getNodeLabel(data)"
            placement="top"
            :open-delay="tooltipOpenDelay"
          >
            <span class="organization-tree-node__name">
              {{ getNodeLabel(data) }}
            </span>
          </el-tooltip>
          <span
            v-if="hasChildren(data) && !isAllOrgNode(data)"
            class="organization-tree-node__actions"
          >
            <el-tooltip
              :content="$t('adminCenter.actions.selectAll')"
              placement="top"
              :open-delay="tooltipOpenDelay"
            >
              <el-button
                type="text"
                size="mini"
                :aria-label="$t('adminCenter.actions.selectAll')"
                @click.stop="handleSubtreeCheck(data, true)"
              >
                <svg-icon icon-class="checks" />
              </el-button>
            </el-tooltip>
            <el-tooltip
              :content="$t('adminCenter.actions.deselectAll')"
              placement="top"
              :open-delay="tooltipOpenDelay"
            >
              <el-button
                type="text"
                size="mini"
                :aria-label="$t('adminCenter.actions.deselectAll')"
                @click.stop="handleSubtreeCheck(data, false)"
              >
                <svg-icon icon-class="square-off" />
              </el-button>
            </el-tooltip>
          </span>
        </span>
      </el-tree>
    </div>
    <div class="selected-organization-summary">
      {{
        $t('adminCenter.common.selectedOrgs', {
          count: selectedCount,
        })
      }}
    </div>
  </div>
</template>

<script>
import { fetchOrgTree } from '@/api/permission/org';

const ALL_ORG_ID = -1;

export default {
  name: 'OrganizationTreeSelect',
  props: {
    value: {
      type: Array,
      default: () => [],
    },
    currentKey: {
      type: [String, Number],
      default: '',
    },
    nodeKey: {
      type: String,
      default: 'orgId',
    },
    treeProps: {
      type: Object,
      default: () => ({
        children: 'children',
        label: 'name',
      }),
    },
    defaultExpandAll: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      keyword: '',
      loading: false,
      organizationTreeData: [],
      tooltipOpenDelay: 400,
    };
  },
  computed: {
    selectedCount() {
      return this.value.filter(key => !this.isAllOrgKey(key)).length;
    },
    defaultExpandedKeys() {
      return [
        ALL_ORG_ID,
        ...this.organizationTreeData.map(node => node[this.nodeKey]),
      ];
    },
    filteredOrganizationTreeData() {
      return this.filterTree(this.organizationTreeData);
    },
    treeData() {
      const childrenKey = this.getChildrenKey();
      const labelKey = this.treeProps.label || 'name';
      return [
        {
          [this.nodeKey]: ALL_ORG_ID,
          [labelKey]: this.$t('adminCenter.options.common.all'),
          [childrenKey]: this.filteredOrganizationTreeData,
        },
      ];
    },
  },
  watch: {
    keyword() {
      this.emitCheckedKeys([]);
      this.$nextTick(() => {
        if (this.$refs.organizationTree) {
          this.$refs.organizationTree.setCheckedKeys([]);
          this.syncAllOrgNodeCheckState([]);
        }
        this.syncSearchExpansion();
        this.syncCurrentKey();
      });
    },
    value() {
      this.syncValueSelection();
    },
    currentKey() {
      this.syncCurrentKey();
    },
  },
  mounted() {
    this.loadOrganizationTree();
  },
  methods: {
    /** 加载完整组织树，并在加载完成后同步外部选中状态。 */
    async loadOrganizationTree() {
      this.loading = true;
      try {
        const res = await fetchOrgTree();
        this.organizationTreeData = res.data || [];
        this.syncValueSelection();
        this.syncCurrentKey();
      } catch (error) {
        this.organizationTreeData = [];
        this.$emit('load-error', error);
      } finally {
        this.loading = false;
      }
    },
    /** 获取当前树数据使用的子节点字段名。 */
    getChildrenKey() {
      return this.treeProps.children || 'children';
    },
    /** 获取节点的子节点数组，非数组值统一视为空数组。 */
    getChildren(data) {
      const children = data[this.getChildrenKey()];
      return Array.isArray(children) ? children : [];
    },
    /** 获取节点的展示名称。 */
    getNodeLabel(data) {
      const labelKey = this.treeProps.label || 'name';
      return data[labelKey] || '';
    },
    /** 判断指定节点键是否为虚拟“全部”组织键。 */
    isAllOrgKey(key) {
      return String(key) === String(ALL_ORG_ID);
    },
    /** 判断指定节点是否为虚拟“全部”节点。 */
    isAllOrgNode(data) {
      return this.isAllOrgKey(data[this.nodeKey]);
    },
    /** 判断节点在当前展示树中是否包含子节点。 */
    hasChildren(data) {
      return this.getChildren(data).length > 0;
    },
    /** 按搜索关键词递归筛选树，并保留匹配节点的祖先路径。 */
    filterTree(nodes) {
      const keyword = this.keyword.trim().toLowerCase();
      if (!keyword) return nodes;

      const childrenKey = this.getChildrenKey();
      return nodes.reduce((filteredNodes, node) => {
        const children = this.filterTree(this.getChildren(node));
        const isMatched = this.getNodeLabel(node)
          .toLowerCase()
          .includes(keyword);

        if (isMatched || children.length) {
          filteredNodes.push({
            ...node,
            [childrenKey]: children,
          });
        }
        return filteredNodes;
      }, []);
    },
    /** 递归收集节点及其后代的真实组织 ID，排除虚拟“全部”节点。 */
    getNodeKeys(nodes, keys = []) {
      nodes.forEach(node => {
        const key = node[this.nodeKey];
        if (!this.isAllOrgKey(key)) {
          keys.push(key);
        }
        this.getNodeKeys(this.getChildren(node), keys);
      });
      return keys;
    },
    /** 获取完整组织树中的全部真实组织 ID。 */
    getAllOrganizationKeys() {
      return this.getNodeKeys(this.organizationTreeData);
    },
    /** 获取当前搜索结果中的全部真实组织 ID。 */
    getCurrentOrganizationKeys() {
      return this.getNodeKeys(this.filteredOrganizationTreeData);
    },
    /** 根据选中键列表返回当前展示树中对应的节点数据。 */
    getCheckedNodes(checkedKeys) {
      const checkedKeySet = new Set(checkedKeys.map(String));
      const checkedNodes = [];
      const collect = nodes => {
        nodes.forEach(node => {
          if (checkedKeySet.has(String(node[this.nodeKey]))) {
            checkedNodes.push(node);
          }
          collect(this.getChildren(node));
        });
      };
      collect(this.treeData);
      return checkedNodes;
    },
    /** 根据完整树是否全选，自动添加或移除虚拟“全部”组织键。 */
    normalizeAllOrgSelection(keys) {
      const checkedKeys = keys.filter(key => !this.isAllOrgKey(key));
      const allOrganizationKeys = this.getAllOrganizationKeys();
      const selectedKeySet = new Set(checkedKeys.map(String));
      const isAllOrganizationsSelected =
        allOrganizationKeys.length > 0 &&
        allOrganizationKeys.every(key => selectedKeySet.has(String(key)));

      return isAllOrganizationsSelected
        ? [...checkedKeys, ALL_ORG_ID]
        : checkedKeys;
    },
    /** 按字符串形式的键值去重，并保留原始键类型与顺序。 */
    uniqueCheckedKeys(keys) {
      const keySet = new Set();
      return keys.filter(key => {
        const normalizedKey = String(key);
        if (keySet.has(normalizedKey)) return false;
        keySet.add(normalizedKey);
        return true;
      });
    },
    /** 忽略键类型与顺序，比较两组已选组织键是否等价。 */
    areCheckedKeysEqual(firstKeys, secondKeys) {
      const firstKeySet = new Set(firstKeys.map(String));
      const secondKeySet = new Set(secondKeys.map(String));
      return (
        firstKeySet.size === secondKeySet.size &&
        [...firstKeySet].every(key => secondKeySet.has(key))
      );
    },
    /** 归一化选中结果，同步树控件并向外发出选中变更事件。 */
    emitCheckedKeys(keys) {
      const checkedKeys = this.uniqueCheckedKeys(
        this.normalizeAllOrgSelection(keys),
      );
      if (this.$refs.organizationTree) {
        this.$refs.organizationTree.setCheckedKeys(checkedKeys);
        this.syncAllOrgNodeCheckState(checkedKeys);
      }
      this.$emit('input', checkedKeys);
      this.$emit('change', {
        checkedKeys,
        checkedNodes: this.getCheckedNodes(checkedKeys),
        halfCheckedKeys: [],
        halfCheckedNodes: [],
      });
    },
    /** 将外部 v-model 选中值归一化后同步回树控件。 */
    syncValueSelection() {
      const checkedKeys = this.uniqueCheckedKeys(
        this.normalizeAllOrgSelection(this.value),
      );
      if (this.areCheckedKeysEqual(checkedKeys, this.value)) {
        this.syncCheckedKeys();
        return;
      }
      this.emitCheckedKeys(checkedKeys);
    },
    /** 处理节点勾选；“全部”节点按当前搜索结果执行全选或全不选。 */
    handleCheck(node, checkedInfo) {
      if (this.isAllOrgNode(node)) {
        const currentResultKeys = this.getCurrentOrganizationKeys();
        const selectedKeySet = new Set(this.value.map(String));
        const isCurrentResultFullySelected =
          currentResultKeys.length > 0 &&
          currentResultKeys.every(key => selectedKeySet.has(String(key)));
        const currentResultKeySet = new Set(currentResultKeys.map(String));
        const nextKeys = isCurrentResultFullySelected
          ? this.value.filter(key => !currentResultKeySet.has(String(key)))
          : this.value.concat(
              currentResultKeys.filter(key => !selectedKeySet.has(String(key))),
            );
        this.emitCheckedKeys(nextKeys);
        return;
      }

      this.emitCheckedKeys(checkedInfo.checkedKeys || []);
    },
    /** 批量勾选或取消勾选当前展示范围内的节点及其后代。 */
    handleSubtreeCheck(node, shouldCheck) {
      const subtreeKeys = this.getNodeKeys([node]);
      const selectedKeySet = new Set(this.value.map(String));
      let nextKeys;

      if (shouldCheck) {
        nextKeys = this.value.concat(
          subtreeKeys.filter(key => !selectedKeySet.has(String(key))),
        );
      } else {
        const subtreeKeySet = new Set(subtreeKeys.map(String));
        nextKeys = this.value.filter(key => !subtreeKeySet.has(String(key)));
      }

      this.emitCheckedKeys(nextKeys);
    },
    /** 递归设置一组树节点及其后代的展开状态。 */
    setNodesExpanded(nodes, expanded) {
      nodes.forEach(node => {
        node.expanded = expanded;
        this.setNodesExpanded(node.childNodes || [], expanded);
      });
    },
    /** 搜索时展开所有筛选结果；清空搜索时恢复默认展开层级。 */
    syncSearchExpansion() {
      const tree = this.$refs.organizationTree;
      if (!tree) return;

      const rootNode = tree.getNode(ALL_ORG_ID);
      if (!rootNode) return;

      if (this.keyword.trim()) {
        this.setNodesExpanded(tree.store.root.childNodes, true);
        return;
      }

      rootNode.expanded = true;
      rootNode.childNodes.forEach(node => {
        node.expanded = true;
        this.setNodesExpanded(node.childNodes || [], false);
      });
    },
    /** 同步当前点击节点，并向外透传节点点击事件。 */
    handleNodeClick(data) {
      this.$emit('update:currentKey', data[this.nodeKey]);
      this.$emit('node-click', data);
    },
    /** 根据完整树和当前搜索结果的选中状态设置“全部”节点的半选状态。 */
    syncAllOrgNodeCheckState(checkedKeys) {
      const allOrgNode = this.$refs.organizationTree.getNode(ALL_ORG_ID);
      if (!allOrgNode) return;

      const selectedKeySet = new Set(checkedKeys.map(String));
      const allOrganizationKeys = this.getAllOrganizationKeys();
      const isAllOrganizationsSelected =
        allOrganizationKeys.length > 0 &&
        allOrganizationKeys.every(key => selectedKeySet.has(String(key)));
      const hasSelectedCurrentResult = this.getCurrentOrganizationKeys().some(
        key => selectedKeySet.has(String(key)),
      );

      allOrgNode.indeterminate =
        !isAllOrganizationsSelected && hasSelectedCurrentResult;
    },
    /** 在树控件渲染完成后写入外部传入的选中键。 */
    syncCheckedKeys() {
      this.$nextTick(() => {
        if (this.$refs.organizationTree) {
          this.$refs.organizationTree.setCheckedKeys(this.value);
          this.syncAllOrgNodeCheckState(this.value);
        }
      });
    },
    /** 在树控件渲染完成后同步当前高亮节点。 */
    syncCurrentKey() {
      this.$nextTick(() => {
        if (!this.$refs.organizationTree) return;
        const hasCurrentKey =
          this.currentKey !== '' &&
          this.currentKey !== null &&
          this.currentKey !== undefined;
        this.$refs.organizationTree.setCurrentKey(
          hasCurrentKey ? this.currentKey : null,
        );
      });
    },
  },
};
</script>

<style lang="scss" scoped>
.organization-tree-select {
  display: flex;
  flex: 1;
  min-height: 0;
  flex-direction: column;
}

.organization-search {
  flex: 0 0 auto;
  margin-bottom: 12px;

  ::v-deep .el-input__inner {
    height: 32px;
    border-color: #e5e7eb;
    border-radius: 12px !important;
    color: #333333;

    &:hover,
    &:focus {
      border-color: $color;
    }
  }

  ::v-deep .el-input__prefix {
    color: #9ca3af;
  }
}

.organization-tree-wrap {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 4px 8px 4px 0;
}

.organization-tree-empty {
  padding: 40px 16px;
  color: #999999;
  font-size: 13px;
  line-height: 20px;
  text-align: center;
}

.organization-tree {
  flex: 1;
  width: max-content;
  min-width: 100%;
  background: transparent;

  ::v-deep .el-tree-node__content {
    width: max-content;
    min-width: 100%;
    height: 36px;
    margin: 2px 0;
    border-radius: 12px;
    background: rgba(255, 255, 255, 0) !important;
    transition: background 0.15s linear;

    &:hover {
      background: #f5f7fa !important;
    }
  }

  ::v-deep .el-tree-node.is-current > .el-tree-node__content {
    background: $color_opacity !important;

    .organization-tree-node__name {
      color: $color;
      font-weight: 500;
    }
  }

  ::v-deep .el-tree-node__label {
    display: flex;
    flex: 0 0 auto;
    min-width: 0;
  }

  ::v-deep .el-tree-node__children {
    overflow: visible;
  }
}

.organization-tree-node {
  display: flex;
  width: max-content;
  min-width: 100%;
  flex: 0 0 auto;
  align-items: center;
  padding-right: 8px;

  &__name {
    min-width: 30px;
    max-width: 120px;
    flex: 0 0 auto;
    overflow: hidden;
    color: #333333;
    font-size: 13px;
    line-height: 20px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  &__actions {
    display: flex;
    flex: 0 0 auto;
    margin-left: 8px;
    gap: 4px;

    .el-button {
      display: inline-flex;
      width: 24px;
      height: 24px;
      margin: 0;
      padding: 4px;
      align-items: center;
      justify-content: center;
      border-radius: 4px;
      font-size: 12px;

      &:hover,
      &:focus {
        background: #eee !important;
      }
    }
  }
}

.selected-organization-summary {
  flex: 0 0 auto;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #eef0f5;
  color: #6b7280;
  font-size: 13px;
  line-height: 20px;
}
</style>
