<template>
    <div class="page-wrapper">
        <div class="page-title">
        {{$t('knowledgeManage.knowledge')}}
        </div>
        <el-container>
            <el-header class="classifyTitle">
                <div class="content_title">
                <h2 class="marginRight">{{$t('knowledgeManage.name')}}</h2>
                <el-link
                    type="danger"
                    :underline="false"
                    @click="handleUpload(true)"
                >{{$t('knowledgeManage.fileUpload')}}</el-link>
                <el-link
                    type="danger"
                    :underline="false"
                    @click="handleHit"
                    style="padding:0 15px;"
                >{{$t('knowledgeManage.hitTest')}}</el-link>
                <el-link
                    type="danger"
                    :underline="false"
                    @click="handleRefresh"
                    style="padding:0 15px;"
                ><span class="el-icon-refresh"></span></el-link>
                </div>
                <div class="searchInfo">
                <el-select
                    @change="changeOption($event)"
                    v-model="docQuery.status"
                    :placeholder="$t('knowledgeManage.please')"
                    style="width:150px;"
                    class="marginRight cover-input-icon"
                >
                    <el-option
                    v-for="item in knowLegOptions"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                    />
                </el-select>
                <el-input
                    style="width:150px;"
                    :placeholder="$t('yuanjing.inputTips')"
                    suffix-icon="el-icon-search"
                    class="cover-input-icon"
                    clearable
                    v-model="docQuery.keyword"
                    @clear="searchDoc"
                    @keyup.enter.native="searchDoc"
                >
                </el-input>
                </div>
            </el-header>
            <el-main
                class="noPadding"
                v-loading="tableLoading"
            >
                <el-alert
                :title="title_tips"
                type="warning"
                show-icon
                style="margin-bottom:10px;"
                v-if="showTips"
                ></el-alert>
                <el-table
                :data="tableData"
                style="width: 100%"
                :header-cell-style="{ background: '#F9F9F9', color: '#999999' }"
                >
                <el-table-column
                    prop="docName"
                    :label="$t('knowledgeManage.fileName')"
                    min-width="350"
                >
                    <template slot-scope="scope">
                    <el-popover
                        placement="bottom"
                        :content="scope.row.docName"
                        trigger="hover"
                        width="200"
                    >
                        <span slot="reference">{{scope.row.docName.length>20?scope.row.docName.slice(0,20)+'...':scope.row.docName}}</span>
                    </el-popover>
                    </template>
                </el-table-column>
                <el-table-column
                    prop="fileSize"
                    :label="$t('knowledgeManage.fileSize')"
                ></el-table-column>
                <!-- <el-table-column
                    prop="oriDocName"
                    label="Original File Name"
                    width="120"
                >
                    <template slot-scope="scope">
                    <el-popover
                        placement="bottom"
                        :content="scope.row.docName"
                        trigger="hover"
                        width="200"
                    >
                        <span slot="reference">{{scope.row.oriDocName.length>10?scope.row.oriDocName.slice(0,10)+'...':scope.row.oriDocName}}</span>
                    </el-popover>
                    </template>
                </el-table-column> -->
                <el-table-column
                    prop="type"
                    :label="$t('knowledgeManage.fileStyle')"
                    width="80"
                >
                </el-table-column>
                <el-table-column
                    prop="uploadTime"
                    :label="$t('knowledgeManage.importTime')"
                    width="180"
                >
                </el-table-column>
                <el-table-column
                    prop="status"
                    :label="$t('knowledgeManage.currentStatus')"
                >
                    <template slot-scope="scope">
                    <span :class="[[4,5].includes(scope.row.status)?'error':'']">{{filterStatus(scope.row.status)}}</span>
                    <el-tooltip
                        class="item"
                        effect="light"
                        :content="scope.row.errorMsg?scope.row.errorMsg:''"
                        placement="top"
                        v-if="scope.row.status === 5"
                        popper-class="custom-tooltip"
                    >
                        <span
                        class="el-icon-warning"
                        style="margin-left:5px;color:#E6A23C;"
                        ></span>
                    </el-tooltip>
                    </template>
                </el-table-column>
                <el-table-column
                    :label="$t('knowledgeManage.operate')"
                    width="260"
                >
                    <template slot-scope="scope">
                    <el-button
                        size="mini"
                        round
                        @click="handleDel(scope.row)"
                        :disabled="[2,3].includes(Number(scope.row.status))"
                        :type="[2,3].includes(Number(scope.row.status))?'info':''"
                    >{{$t('createApp.delete')}}</el-button>
                    <el-button
                        size="mini"
                        round
                        :type="[0,3,5].includes(Number(scope.row.status))?'info':''"
                        :disabled="[0,3,5,-2].includes(Number(scope.row.status))"
                        @click="handleView(scope.row)"
                    >{{$t('knowledgeManage.view')}}</el-button>
                    </template>
                </el-table-column>
                </el-table>
                <!-- 分页 -->
                <Pagination
                class="pagination table-pagination"
                ref="pagination"
                :listApi="listApi"
                :page_size="10"
                @refreshData="refreshData"
                />
            </el-main>
        </el-container>
    </div>
</template>
<script>
export default {
    data(){
        return{

        }
    }
}
</script>