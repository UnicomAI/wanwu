<template>
  <div>
    <!-- Search Area -->
    <slot name="search"></slot>

    <!-- Table Area -->
    <!--<el-scrollbar>
    <div style="max-height: 5rem;">-->
    <div>
      <el-table
        :data="table.data"
        :row-class-name="tableRowClassName"
        @select="handleSelect"
        @select-all="handleSelectAll"
        @cell-click="handCellClick"
      >
        <slot name="checkbox"></slot>

        <!-- Multi-select Area -->
        <slot name="selection" ></slot>

        <!-- Index Area -->
        <el-table-column v-if="index" type="index" :index="indexMethod" width="62" label="No."></el-table-column>

        <!-- Header -->
        <slot name="haed"></slot>

        <!-- Table Header Area -->
        <el-table-column
          v-for="(column, index) in table.column"
          :key="index"
          :label="column.label"
          :prop="column.prop"
          :width="column.width"
          :formatter="column.formatter"
          :fixed="column.fixed"
        >
          <span v-if="column.html" >{{column.formatter()}}</span>
        </el-table-column>

        <!-- SettingArea -->
        <slot name="other1"></slot>
        <slot name="other2"></slot>
        <slot name="other3"></slot>

        <!-- ButtonArea -->
        <slot name="button"></slot>
      </el-table>
    </div>
    <!--</el-scrollbar>-->
    <!-- Pagination Area -->
    <el-pagination
      @size-change="handleSizeChange"
      @current-change="handleCurrentChange"
      :current-page="page.pageNo"
      :page-sizes="[5, 10, 15, 20]"
      :page-size="page.pageSize"
      layout="total, sizes, prev, pager, next, jumper"
      :total="table.total"
      class="pagination"
    ></el-pagination>
  </div>
</template>

<script>
export default {
  name: "tableWithPagination",
  props: ["noCreate","table","index","select",'pageSize','noborder'],
  data() {
    return {
      // Table pagination parameters
      page: {
        pageNo: 1,
        pageSize: 5
      },
      // Table multi-select collection
      selection: []
    };
  },
  created: function() {
    // Initialize table load
    this.page.pageSize=this.pageSize||5
    if(!this.noCreate){
      this.handlePagination();
    }
  },
  methods: {
    // Index calculation for serial numbers
    indexMethod(index) {
      return (this.page.pageNo - 1) * this.page.pageSize + index + 1;
    },
    // Assign the calculated index back to each row
    tableRowClassName({ row, rowIndex }) {
      row.rowIndex = rowIndex;
    },
    // Execute pagination logic
    handlePagination() {
      this.selection = [];
      this.$emit("handlePagination", this.page);
    },
    handleSizeChange(val) {
      this.page.pageSize = val;
      this.handlePagination();
    },
    handleCurrentChange(val) {
      this.page.pageNo = val;
      this.handlePagination();
    },
    handleSelect(selection, row) {
      this.selection = selection;
    },
    handleSelectAll(selection) {
      this.selection = selection;
    },
    handCellClick(row, column, cell, event) {
      var params = {
        row: row,
        column: column,
        cell: cell,
        event: event
      };
      this.$emit("handCellClick", params);
    },
    handelInitPage(){
      this.page.pageNo= 1
      this.page.pageSize = this.pageSize||5
    }
  }
};
</script>

<style>
.pagination {
  margin-top: 20px;
  text-align: right;
}
.el-pager li.active{
  color: #D33A3A;
}
</style>
