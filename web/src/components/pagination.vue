<template>
  <div class="container my-pagination">
    <el-pagination
      :current-page="pageNo"
      :page-size="pageSize"
      :page-sizes="[5, 10, 20, 50]"
      :style="{ float: 'right', padding: '20px 5px' }"
      :total="total"
      @current-change="handleCurrentChange"
      @size-change="handleSizeChange"
      layout="total, sizes, prev, pager, next, jumper"
    ></el-pagination>
  </div>
</template>

<script>
import infoList from '@/mixins/infoList';
export default {
  mixins: [infoList],
  components: {},
  name: '',
  props: {
    listApi: '',
  },
  data() {
    return {};
  },
  mounted() {},
  methods: {},
};
</script>

<style lang="scss" scoped>
.container {
  .con_table_pagination {
    margin-top: 5px;
  }
}
.my-pagination ::v-deep {
  .el-pagination.is-background .el-pager li:not(.disabled).active {
    background-color: $color;
    color: #fff;
  }

  .el-pagination .el-select .el-input .el-input__inner {
    padding-right: 25px;
    width: 109px;
    border-color: #cccccc;
  }

  .el-pager li:hover {
    color: $color;
  }

  .el-pager li.active {
    color: $color;
    cursor: default;
    border: 1px solid $color;
  }

  .el-pager .number,
  button {
    background: #ffffff !important;
  }

  .el-pagination__editor.el-input .el-input__inner {
    height: 28px;
    background: #ffffff;
    border: 1px solid #cccccc;
  }
  .el-pagination.is-background .el-pager li:not(.disabled):hover {
    color: $color;
  }
}
</style>
