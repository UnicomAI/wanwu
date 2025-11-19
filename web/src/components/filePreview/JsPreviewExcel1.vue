<template>
  <div style="width: 100%;height: 100%;overflow-x: auto;">
    <div id="comac-preview-excel" style="height: 100%;"></div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, onBeforeUnmount } from 'vue';
import jsPreviewExcel from "@js-preview/excel";
import '@js-preview/excel/lib/index.css';
import {useRoute} from "vue-router";
const page = ref({})
const myExcelPreviewer = ref(null);
const testSheets = [
  { name: 'Sheet1', rows: [{ cells: [{ value: 'Test1' }] }] },
  { name: 'Sheet2', rows: [{ cells: [{ value: 'Test2' }] }] }
];
const initPreviewer = () => {
  
  myExcelPreviewer.value = jsPreviewExcel.init(document.getElementById('comac-preview-excel'));
  const highlightRowNum = page.value.rownum ? page.value.rownum - 1 : null;// Highlighted row index
  let dataIndex = -1;
  myExcelPreviewer.value.setOptions({
    transformData: (workbookData) => {
      if(!workbookData || !Array.isArray(workbookData)){
        return [];
      }

      dataIndex = workbookData.findIndex(item => item.name === page.value.sheetName)
      if (highlightRowNum !== null && workbookData[dataIndex].rows[highlightRowNum]) {
          const row = workbookData[dataIndex].rows[highlightRowNum];
          const cells = row.cells
          if(row.cells){
            for(let key in cells){// Highlight the specified row
              workbookData[dataIndex].styles[cells[key].style].bgcolor = '#ff0000';
              workbookData[dataIndex].styles[cells[key].style].color = '#fff';
            }
          }
      }

      return workbookData
    }
  })

  myExcelPreviewer.value.preview(page.value.url)
};

watch(() => page.value.url, (newUrl) => {
  myExcelPreviewer.value.preview(newUrl)
});

onMounted(() => {
   const route = useRoute()
   page.value = route.query || {}
   initPreviewer();
    console.log(document.getElementById('comac-preview-excel'))
  })

onBeforeUnmount(() => {
  if (myExcelPreviewer.value) {
    myExcelPreviewer.value.destroy();
  }
});
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
