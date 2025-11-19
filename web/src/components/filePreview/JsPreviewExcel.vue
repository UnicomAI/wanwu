<template>
  <div style="height: 100vh; overflow: auto;" ref="container">
    <VueOfficeExcel
      :src="page.url"
      :options="options"
      style="height: 100vh;"
      @rendered="renderedHandler"
      @error="errorHandler"
    />
  </div>
</template>

<script>
import  VueOfficeExcel  from '@vue-office/excel';
import '@vue-office/excel/lib/index.css';

export default {
  name: 'ExcelPreview',
  components: {
    VueOfficeExcel
  },

  data() {
    return {
      page: {},
      options: {}
    };
  },

  mounted() {
    // GetRouteParameter
    this.page = this.$route.query || {};
    this.initPreviewer();
  },

  methods: {
    initPreviewer() {
      const highlightRowNum = this.page.rownum ? parseInt(this.page.rownum) - 1 : null; // 高亮行Index（0开始）
      let dataIndex = -1;

      this.options = {
        xls: false,
        minColLength: 0,
        minRowLength: 0,
        widthOffset: 20,
        heightOffset: 10,
        transformData: (workbookData) => {
          // Find指定 sheet
          dataIndex = workbookData.findIndex(item => item.name === this.page.sheetName);

          if (dataIndex === -1) {
            console.warn('未找到指定 of  sheet:', this.page.sheetName);
            return workbookData;
          }

          // 高亮指定行
          if (highlightRowNum !== null && workbookData[dataIndex].rows[highlightRowNum]) {
            const row = workbookData[dataIndex].rows[highlightRowNum];
            if (row.cells) {
              for (let key in row.cells) {
                const cell = row.cells[key];
                const styleKey = cell.style;

                // 确保 style 存在
                if (styleKey && workbookData[dataIndex].styles[styleKey]) {
                  workbookData[dataIndex].styles[styleKey].bgcolor = '#ff0000';
                  workbookData[dataIndex].styles[styleKey].color = '#fff';
                }
              }
            }
          }

          return workbookData;
        }
      };
    },

    renderedHandler() {
      console.log('Excel 渲染完成');
    },

    errorHandler(err) {
      console.error('Excel 加载Failed:', err);
    }
  },

  watch: {
    // If你想ListenRoute变化（For example url Parameter变了），重新Init
    '$route.query': {
      handler(newQuery) {
        this.page = newQuery || {};
        this.initPreviewer();
      },
      immediate: false
    }
  }
};
</script>

<style lang="scss">
canvas {
  width: 100% !important;
}

.x-spreadsheet-sheet {
  width: 100% !important;
}

.highlighted-row {
  background-color: #ff0000 !important;
  color: #ffffff !important;
  animation: pulse 1s infinite alternate;
}
</style>