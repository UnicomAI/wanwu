<template>
  <div class="doc-page-container">
    <div class="doc-header">
      <div class="doc-header__left">
        <img v-if="homeLogoPath" style="height: 50px;" :src="basePath + '/user/api' + homeLogoPath"/>
        <span v-if="homeTitle">
          {{homeTitle}}
        </span>
      </div>
      <div class="doc-header__right">
        <el-select
          filterable
          remote
          clearable
          placeholder=""
          :remote-method="visibleChange"
          style="width: 500px;"
          class="top-search-input"
          popper-class="top-search-popover"
          :popper-append-to-body="false"
          v-model="searchText"
        >
          <i
            slot="prefix"
            class="el-input__icon el-icon-search"
            style="font-weight: bolder; font-size: 16px;"
          />
          <div class="header_search-option" v-if="searchText">
            <el-option v-if="searchList" :value="searchText" style="height: auto; background: #fff">
              <div v-if="docLoading" style="text-align: center; padding: 50px 0">
                <i class="el-icon-loading" style="font-size: 28px; "></i>
              </div>
              <div v-if="!docLoading && !(searchList && searchList.length)" style="text-align: center; padding: 50px 0">
                <span style="font-size: 14px; font-weight: normal; color: #999">{{$t('header.noData')}}</span>
              </div>
              <div
                v-if="!docLoading && searchList && searchList.length"
                v-for="(item, index) in searchList"
                :key="`search${item.title + index}`"
              >
                <div class="header_search-title">{{item.title}}</div>
                <div class="header_search-item" v-for="(it, i) in item.list" :key="`it${it.title + i}`">
                  <div class="header_search-item-left">
                    {{it.title}}
                  </div>
                  <div class="header_search-item-right" @click="jumpMenu(it.url)" v-html="it.content" />
                </div>
              </div>
            </el-option>
          </div>
        </el-select>
      </div>
    </div>
    <div class="doc-outer-container">
      <div :class="['doc-inner-container']">
        <el-aside style="min-width: 200px; width: auto; max-width: 300px" class="full-menu-aside">
          <el-menu
            :default-openeds="defaultOpeneds"
            :default-active="activeIndex"
            class="el-menu-vertical-demo"
          >
            <div v-for="(n,i) in menuList" :key="`${i}ml`">
              <!--有下一级-->
              <el-submenu v-if="n.children && checkPerm(n.perm)" :index="n.index">
                <template slot="title">
                  <i :class="n.icon || 'el-icon-menu'"></i>
                  <span>{{n.name}}</span>
                </template>
                <div v-for="(m,j) in n.children" v-if="checkPerm(m.perm)" :key="`${j}cl`">
                  <el-submenu v-if="m.children" :index="m.index" class="menu-indent">
                    <template slot="title">{{m.name}}</template>
                    <div v-for="(p,k) in m.children" :key="`${k}pl`" v-if="checkPerm(p.perm)">
                      <el-submenu v-if="p.children" :index="p.index" class="menu-indent-sub">
                        <template slot="title">{{p.name}}</template>
                        <el-menu-item
                          v-for="(item, index) in p.children"
                          :key="`${index}itemEl`"
                          :index="item.index"
                          v-if="checkPerm(item.perm)"
                          @click="menuClick(item)"
                        >
                          {{item.name}}
                        </el-menu-item>
                      </el-submenu>
                      <el-menu-item v-else :index="p.index" @click="menuClick(p)">{{p.name}}</el-menu-item>
                    </div >
                  </el-submenu>
                  <el-menu-item v-else :index="m.index" @click="menuClick(m)" class="menu-indent-item">{{m.name}}</el-menu-item>
                </div >
              </el-submenu>
              <!--No下一级-->
              <el-menu-item :index="n.index" v-if="!n.children && checkPerm(n.perm)" @click="menuClick(n)">
                <i :class="n.icon || 'el-icon-menu'"></i>
                <span slot="title">{{n.name}}</span>
              </el-menu-item>
            </div>
          </el-menu>
        </el-aside>
        <!-- 右侧Content -->
        <div class="doc-page-main">
          <DocPage />
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import DocPage from "./components/docPage.vue"
import { checkPerm } from "@/router/permission"
import { DOC_FIRST_KEY } from "./constants"
import { getDocMenu, getDocSearchContent } from "@/api/docs"
import { fetchPermFirPath, fetchCurrentPathIndex } from "@/utils/util"
import { mapGetters } from "vuex"

export default {
  components: { DocPage },
  data() {
    return{
      basePath: this.$basePath,
      homeLogoPath: '',
      homeTitle: '',
      defaultOpeneds: [],
      menuList: [],
      docMenuList: [],
      docLoading: false,
      searchList: [],
      searchText: '',
      activeIndex: '0',
    }
  },
  watch: {
    $route: {
      handler (val) {
        this.changeMenuIndex(fetchCurrentPathIndex(val.path, this.menuList))
      },
      // 深度ObserveListen
      deep: true
    },
    commonInfo:{
      handler(val) {
        const { home = {} } = val.data || {}
        this.homeLogoPath = home.logo ? home.logo.path : ''
        this.homeTitle = home.title || ''
      },
      deep: true
    }
  },
  computed: {
    ...mapGetters('user', ['commonInfo']),
  },
  async created() {
    // GetMenu
    this.getCurrentMenu()
  },
  methods: {
    checkPerm,
    jumpMenu(url) {
      // location.href = url
      const [_, path] = url.split(`${this.$basePath}/aibase`)
      this.$router.push({path})
    },
    loadSearchMenu() {
      this.docLoading = true
      this.searchList = []
      getDocSearchContent({content: this.searchText}).then(res => {
        this.docLoading = false
        this.searchList = res.data || []
      }).catch(() => {
        this.docLoading = false
      })
    },
    visibleChange(val) {
      this.searchText = val
      if (val) {
        this.loadSearchMenu()
      }
    },
    menuClick(item) {
      if (item.redirect) {
        item.redirect()
      } else {
        // Document中心Back不带Page path 前缀，跳转加上 path 前缀，避免点击PathDirectly拼到当前Link后面Etc问题
        this.$router.push({path: `/docCenter/pages/${item.path}`})
      }
    },
    getDocMenu() {
      return getDocMenu().then(res => {
        this.docMenuList = res.data || []
      })
    },
    getCurrentMenu() {
      const route = this.$route
      this.getDocMenu().then(() => {
        this.getMenuList(route)
      })
    },
    getMenuList() {
      const { params, path } = this.$route || {}
      const { id } = params || {}
      let val = path
      // Get当前MenuList
      const menus = this.docMenuList
      this.menuList = menus
      this.defaultOpeneds = menus.map(item => item.index)

      // 跳转到Document中心第一个Menu栏
      if (id === DOC_FIRST_KEY) {
        const { path } = fetchPermFirPath(menus)
        val = path
        this.$router.push({path})
      }

      // 给当前 activeIndex 赋Value
      this.changeMenuIndex(fetchCurrentPathIndex(val, menus))
    },
    changeMenuIndex(index) {
      this.activeIndex = index
    },
  }
}
</script>

<style lang="scss" scoped>
.doc-page-container {
  height:100%;
  .doc-outer-container{
    height: calc(100% - 90px);
    background: #fff;
    width: calc(100% - 130px);
    margin: 0 auto;
    border-radius: 8px;
    /*element ui Style重写*/
    .doc-inner-container{
      height: 100%;
      display: flex;
      .el-aside.full-menu-aside {
        height: 100%;
        background-color: #fff;
        overflow-y: auto;
        overflow-x: auto;
        border-right: 1px solid #ededed;
        border-radius: 8px 0 0 0;
        .el-menu{
          min-height: 100%;
          width: fit-content;
          overflow-x: auto;
          overflow-y: hidden;
          .menu-indent /deep/ .el-submenu__title,
          .menu-indent-item {
            padding-left: 49px !important;
          }
          .menu-indent-sub /deep/ .el-submenu__title{
            padding-left: 60px !important;
          }
        }
      }
      .doc-page-main {
        width: 100%;
        height: 100%;
        min-height: 580px;
        overflow: auto;
        padding-top: 30px;
        background: rgba(255, 255, 255, 0);
        border-radius: 8px;
      }
      /deep/ .el-submenu__title,
      /deep/ .el-menu-item span,
      /deep/ .el-submenu__title span {
        font-size: 14px !important;
      }
      /deep/ .el-menu-item.is-active {
        background-color: rgba(230, 0, 1, 0.03) !important;
      }
      /deep/ .el-menu-item:focus {
        background-color: rgba(230, 0, 1, 0.03) !important;
      }
      /deep/ .el-menu-item:hover,
      /deep/ .el-submenu__title:hover {
        background-color: #f6f6f6 !important;
      }
      /deep/ .el-submenu__title {
        span {
          font-size: 14px !important;
        }
      }
      /deep/ .el-submenu.is-active .el-submenu__title {
        border-bottom-color: $color !important;
      }
    }
  }
  .doc-outer-container /deep/ {
    .el-submenu.is-active,
    .el-submenu.is-active > .el-submenu__title,
    .el-submenu.is-active > .el-submenu__title i:first-child,
    .el-submenu.is-active > .el-submenu__title .el-submenu__icon-arrow {
      color: $color !important;
    }
  }
}
.doc-header{
  padding: 20px 0;
  display: flex;
  justify-content: center;
  .doc-header__left {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 30%;
    span {
      font-size: 22px;
      margin-left: 18px;
      font-weight: bold;
      color: $color_title;
    }
  }
  .doc-header__right {
    width: 60%;
    display: flex;
    align-items: center;
  }
  .top-search-input /deep/ {
    .el-input__inner {
      border-radius: 20px;
    }
    .top-search-popover {
      left: auto !important;
      right: -150px !important;
    }

    .popper__arrow {
      display: none !important;
    }
    .el-select-dropdown__wrap {
      max-height: 550px !important;
    }
    .el-select-dropdown__item {
      background: rgba(255, 255, 255, 0) !important;
      padding: 4px 10px;
    }
    .el-input__suffix-inner {
      display: inline-block;
    }
    .header_search-option {
      width: 800px;
      .el-icon-loading {
        color: $color;
      }
      .header_search-title {
        color: #fff;
        background-color: #5b6bee;
        padding: 0 10px;
        font-weight: bold;
        font-size: 16px;
      }
      .header_search-item {
        display: flex;
        justify-content: space-between;
        color: #333;
        border-bottom: 1px solid #d8d6d6;
        .header_search-item-left {
          background-color: #f1f1f1;
          width: 180px;
          text-align: right;
          padding: 0 10px;
          font-weight: bold;
        }
        .header_search-item-right {
          width: calc(100% - 180px);
          text-align: left;
          padding: 3px 10px;
          color: #666;
          font-weight: bold;
          white-space: normal; /* 保留Null白符序列，但Yes正常换行 */
          word-break: break-all;
          span {
            color: $color;
            text-decoration: underline;
          }
          p {
            line-height: 20px;
          }
          a {
            color: $color;
          }
          img {
            width: 100%;
          }
        }
        .header_search-item-right:hover {
          background-color: #f1f1f1;
        }
      }
    }
  }
}
</style>
