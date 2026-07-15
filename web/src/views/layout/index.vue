<template>
  <div class="layout full-menu" :style="`background: ${bgColor}`">
    <el-container class="outer-container">
      <!-- 左侧内容 -->
      <div
        v-if="isShowMenu"
        :class="[
          'left-aside-container',
          { 'left-aside-container-isCollapse': isCollapse },
        ]"
      >
        <div class="left-header-container">
          <div style="padding-top: 12px; text-align: center; height: 46px">
            <div v-if="homeLogoPath">
              <img
                v-if="!isCollapse"
                style="max-height: 46px; max-width: 80%"
                :src="avatarSrc(homeLogoPath)"
              />
              <img
                v-else
                :src="avatarSrc(tabLogoPath)"
                style="height: 36px; width: 36px; object-fit: contain"
              />
            </div>
          </div>
          <!-- 组织切换 -->
          <div class="menu-org-select-wrapper" v-if="!isCollapse">
            <img style="width: 16px" src="@/assets/imgs/org_user.svg" alt="" />
            <ChangeOrg
              :org="org"
              :orgList="orgList"
              :getCurrentOrgName="getCurrentOrgName"
              :changeOrg="changeOrg"
            />
          </div>
          <el-popover
            v-else
            popper-class="menu-org-popover"
            placement="right"
            width="220"
            trigger="click"
          >
            <ChangeOrg
              :org="org"
              :orgList="orgList"
              :getCurrentOrgName="getCurrentOrgName"
              :changeOrg="changeOrg"
            />
            <div slot="reference">
              <div style="text-align: center; cursor: pointer">
                <img
                  class="menu-org-popover-icon"
                  src="@/assets/imgs/org_user.svg"
                  alt=""
                />
              </div>
            </div>
          </el-popover>
          <div v-if="isCollapse" class="collapse-menu-divider"></div>
        </div>
        <div @mouseenter="menuHover = true" @mouseleave="menuHover = false">
          <!-- 菜单 -->
          <el-aside
            v-if="menuList && menuList.length"
            :class="[
              'full-menu-aside',
              { 'full-menu-isCollapse': isCollapse },
              { 'menu-hover': menuHover },
            ]"
          >
            <!-- 展开状态：扁平化分组列表 -->
            <div v-if="!isCollapse" class="flat-menu-list">
              <template v-for="(n, i) in menuList" v-if="checkPerm(n.perm)">
                <!-- 有子级：渲染每个子项 -->
                <div v-if="n.children" :key="`${i}ml`" class="menu-group">
                  <div class="menu-group-title">{{ n.name }}</div>
                  <div
                    v-for="(m, j) in n.children"
                    :key="`${j}cl`"
                    v-if="checkPerm(m.perm)"
                    :class="[
                      'flat-menu-item',
                      { 'flat-menu-item-active': activeIndex === m.index },
                    ]"
                    @click="menuClick(m)"
                  >
                    <div class="menu-svg">
                      <svg-icon
                        class="menu-icon"
                        :icon-class="m.icon || 'menu_default'"
                      />
                    </div>
                    <span class="menu-withIcon-title">{{ m.name }}</span>
                  </div>
                </div>
                <!-- 无子级的顶级菜单项 -->
                <div
                  v-else
                  :key="`${i}ml-top`"
                  :class="[
                    'flat-menu-item',
                    { 'flat-menu-item-active': activeIndex === n.index },
                  ]"
                  @click="menuClick(n)"
                >
                  <div class="menu-svg">
                    <svg-icon class="menu-icon" :icon-class="n.icon" />
                  </div>
                  <span class="menu-withIcon-title">{{ n.name }}</span>
                </div>
              </template>
            </div>

            <!-- 收起状态：仅图标 + 分组分隔 -->
            <div v-else class="collapse-menu-list">
              <template v-for="(n, i) in menuList" v-if="checkPerm(n.perm)">
                <!-- 分组分隔线（非第一组） -->
                <div
                  v-if="i > 0"
                  :key="`${i}divider`"
                  class="collapse-menu-divider"
                ></div>
                <!-- 有子级：渲染每个子项图标 -->
                <template v-if="n.children">
                  <div
                    v-for="(m, j) in n.children"
                    :key="`${i}-${j}cl`"
                    v-if="checkPerm(m.perm)"
                  >
                    <el-tooltip
                      placement="right"
                      trigger="click"
                      width="auto"
                      :content="m.name"
                    >
                      <div
                        :class="[
                          'collapse-menu-item',
                          {
                            'collapse-menu-item-active':
                              activeIndex === m.index,
                          },
                        ]"
                        @click="menuClick(m)"
                      >
                        <svg-icon
                          class="menu-icon"
                          :icon-class="m.icon || 'menu_default'"
                        />
                      </div>
                    </el-tooltip>
                  </div>
                </template>
                <!-- 无子级 -->
                <div v-else :key="`${i}top`">
                  <el-tooltip
                    placement="right"
                    trigger="click"
                    width="auto"
                    :content="n.name"
                  >
                    <div
                      :class="[
                        'collapse-menu-item',
                        {
                          'collapse-menu-item-active': activeIndex === n.index,
                        },
                      ]"
                      @click="menuClick(n)"
                    >
                      <svg-icon class="menu-icon" :icon-class="n.icon" />
                    </div>
                  </el-tooltip>
                </div>
              </template>
            </div>
          </el-aside>
        </div>
        <div
          :class="['left-bottom-container', { 'menu-isCollapse': isCollapse }]"
        >
          <el-popover placement="top" width="220" trigger="click">
            <div
              :class="[
                'menu--popover-wrap',
                { 'wrap-last': popoverList.length === index + 1 },
              ]"
              v-for="(it, index) in popoverList"
              :key="'popoverList' + index"
            >
              <div
                v-if="checkPerm(item.perm)"
                v-for="item in it"
                :key="item.name"
              >
                <!-- 有子菜单的项 -->
                <el-popover
                  v-if="item.children && item.children.length"
                  placement="right"
                  width="140"
                  trigger="click"
                  :popper-class="'menu-submenu-popover'"
                >
                  <div
                    v-for="child in item.children"
                    :key="child.name"
                    class="menu--popover-item"
                    style="padding-left: 4px"
                    @click="menuClick(child)"
                  >
                    <img
                      class="menu--popover-item-img"
                      :src="child.img"
                      alt=""
                    />
                    <span class="menu--popover-item-name">
                      {{ child.name }}
                    </span>
                  </div>
                  <div slot="reference" class="menu--popover-item">
                    <img
                      class="menu--popover-item-img"
                      :src="item.img"
                      alt=""
                    />
                    <span class="menu--popover-item-name">{{ item.name }}</span>
                    <img
                      v-if="item.icon"
                      class="menu--popover-item-icon"
                      :src="item.icon"
                      alt=""
                    />
                  </div>
                </el-popover>
                <!-- 无子菜单的普通项 -->
                <div v-else class="menu--popover-item" @click="menuClick(item)">
                  <img class="menu--popover-item-img" :src="item.img" alt="" />
                  <el-tooltip
                    v-if="item.isTip"
                    effect="dark"
                    :content="item.tipContent"
                    placement="top-start"
                  >
                    <span
                      style="display: inline-block; width: 150px"
                      class="menu--popover-item-name"
                    >
                      {{ item.name }}
                    </span>
                  </el-tooltip>
                  <span v-if="!item.isTip" class="menu--popover-item-name">
                    {{ item.name }}
                  </span>
                  <img
                    v-if="item.icon"
                    class="menu--popover-item-icon"
                    :src="item.icon"
                    alt=""
                  />
                  <span v-if="item.version" class="menu--popover-item-version">
                    {{ version || '' }}
                  </span>
                </div>
              </div>
            </div>
            <div
              v-if="!isCollapse"
              class="header-user-content"
              slot="reference"
            >
              <img
                class="header-icon"
                v-if="userAvatar"
                :src="avatarSrc(userAvatar)"
              />
              <div class="header-user-name">{{ userInfo.userName }}</div>
              <i class="el-icon-more header-more"></i>
            </div>
            <div class="user-content-isCollapse" v-else slot="reference">
              <img
                class="header-icon"
                v-if="userAvatar"
                :src="avatarSrc(userAvatar)"
              />
            </div>
          </el-popover>
          <div class="left-bottom-menu-icon" @click="changeMenuCollapse">
            <img
              src="@/assets/imgs/left_menu_icon.svg"
              alt=""
              v-if="!isCollapse"
            />
            <img src="@/assets/imgs/right_menu_icon.svg" alt="" v-else />
          </div>
        </div>
      </div>
      <!-- 容器 -->
      <el-container :class="['inner-container']">
        <!-- 右侧内容 -->
        <el-main :class="[{ 'no-header-main': !isShowMenu }]">
          <div class="page-container">
            <div class="right-page-content">
              <keep-alive :include="cachedViews">
                <router-view :key="routerViewKey"></router-view>
              </keep-alive>
              <div id="container" class="qk-container"></div>
            </div>
          </div>
        </el-main>
      </el-container>
    </el-container>
    <AboutDialog ref="aboutDialog" />
  </div>
</template>

<script>
// import { start } from 'qiankun'
import { mapActions, mapGetters } from 'vuex';
import { checkPerm, PERMS } from '@/router/permission';
import { menuList as rawMenuList } from './menu';
import { changeLang } from '@/api/user';
import {
  fetchPermFirPath,
  fetchCurrentPathIndex,
  replaceIcon,
  replaceTitle,
  redirectUserInfoPage,
  avatarSrc,
} from '@/utils/util';
import ChangeLang from '@/components/changeLang.vue';
import ChangeOrg from '@/components/changeOrg.vue';
import AboutDialog from '@/components/aboutDialog.vue';
import { DOC_FIRST_KEY } from '@/views/docCenter/constants';

export default {
  name: 'Layout',
  components: { ChangeLang, ChangeOrg, AboutDialog },
  data() {
    return {
      menuHover: false,
      isCollapse: false,
      homeLogoPath: '',
      tabLogoPath: '',
      bgColor: '',
      version: '',
      orgList: [],
      org: { orgId: '' },
      defaultOpeneds: [],
      menuKey: 'menu_key',
      activeIndex: '',
      isShowMenu: true,
      cachedViews: [],
      popoverList: [
        [
          {
            name: this.$t('menu.account'),
            path: '/userInfo',
            img: require('@/assets/imgs/user_icon.svg'),
          },
          {
            name: this.$t('menu.setting'),
            path: '/permission',
            img: require('@/assets/imgs/setting_icon.svg'),
            isTip: true,
            tipContent: this.$t('menu.settingTip'),
            perm: PERMS.ADMIN_CENTER,
          },
          {
            name: this.$t('menu.operationManage'),
            path: '/operation',
            img: require('@/assets/imgs/operationManage.svg'),
            perm: PERMS.OPERATION,
          },
        ],
        [
          {
            name: this.$t('menu.helpDoc'),
            img: require('@/assets/imgs/helpDoc_icon.svg'),
            icon: require('@/assets/imgs/link_icon.png'),
            redirect: () => {
              // open('https://github.com/UnicomAI/wanwu/tree/main/docs/manual')
              open(
                location.origin +
                  `${this.$basePath}/aibase/docCenter/pages/${DOC_FIRST_KEY}`,
              );
            },
          },
          {
            name: this.$t('menu.openSource'),
            img: require('@/assets/imgs/openSource.svg'),
            icon: require('@/assets/imgs/link_icon.png'),
            perm: PERMS.OPEN_SOURCE,
            children: [
              {
                name: 'Github',
                img: require('@/assets/imgs/github_icon.svg'),
                redirect: () => {
                  open('https://github.com/UnicomAI/wanwu');
                },
              },
              {
                name: 'Gitee',
                img: require('@/assets/imgs/gitee_icon.svg'),
                redirect: () => {
                  open('https://gitee.com/UnicomAI/wanwu');
                },
              },
            ],
          },
          {
            name: this.$t('menu.about'),
            img: require('@/assets/imgs/about_icon.svg'),
            version: 'version',
            redirect: () => {
              this.$refs.aboutDialog && this.$refs.aboutDialog.openDialog();
            },
          },
        ],
        [
          {
            name: this.$t('header.logout'),
            img: require('@/assets/imgs/logout_icon.svg'),
            redirect: () => {
              this.logout();
            },
          },
        ],
      ],
    };
  },
  watch: {
    $route: {
      handler(val) {
        this.justifyIsShowMenu(val.path);
        this.getMenuList(val.path);
        this.redirectUserInfo();
        this.initScroll();
        // 缓存当前路由组件
        this.addCachedView(val);
      },
      deep: true,
      immediate: true,
    },
    orgInfo: {
      handler(val) {
        this.orgList = val.orgs || [];
      },
      deep: true,
    },
    commonInfo: {
      handler(val) {
        const { home = {}, tab = {}, about = {} } = val.data || {};
        this.homeLogoPath = home.logo ? home.logo.path : '';
        this.tabLogoPath = tab.logo ? tab.logo.path : '';
        this.bgColor = home.backgroundColor || this.$config.backgroundColor;
        this.version = about.version || '1.0';
        replaceIcon(tab.logo ? tab.logo.path : '');
        replaceTitle(tab.title);
      },
      deep: true,
    },
    permission: {
      handler() {
        // 如果没修改过密码，重新向到修改密码
        this.redirectUserInfo();
      },
      deep: true,
    },
  },
  computed: {
    // 共用组件的路由（如 /tool、/prompt、/mcpService）通过 routeType 区分实例，
    // 避免 Vue 复用同一组件导致状态残留；其他路由返回 undefined 保持原行为
    routerViewKey() {
      const { routeType } = this.$route.meta || {};
      return routeType || undefined;
    },
    ...mapGetters('user', [
      'orgInfo',
      'userInfo',
      'commonInfo',
      'permission',
      'userAvatar',
    ]),
    menuList() {
      const MENU_NAME_MAP = {
        'generalAgent-wanwuAgent':
          this.commonInfo?.data?.generalAgent?.menuName,
      };

      if (!Object.values(MENU_NAME_MAP).some(Boolean)) return rawMenuList;

      return rawMenuList.map(group => ({
        ...group,
        children: (group.children ?? []).map(c =>
          MENU_NAME_MAP[c.index] ? { ...c, name: MENU_NAME_MAP[c.index] } : c,
        ),
      }));
    },
  },
  async created() {
    // 设置语言
    // await this.setLanguage()

    // 获取菜单
    this.getCurrentMenu();
    this.setLocalMenuCollapse();

    // 只有登陆状态下才查询接口，否则会一直刷新
    if (localStorage.getItem('access_cert')) this.getPermissionInfo();

    // 设置组织列表以及当前的组织
    this.orgList = this.orgInfo.orgs || [];
    this.org.orgId = this.userInfo.orgId;

    // 获取平台名称以及 logo 等信息
    this.getCommonInfo();
  },
  /* 保证容器 DIV 在 qiankun start 时一定存在 */
  mounted() {
    /* start() */
  },
  methods: {
    avatarSrc,
    ...mapActions('user', ['LoginOut', 'getPermissionInfo', 'getCommonInfo']),
    checkPerm,
    logout() {
      localStorage.removeItem('access_cert');
      location.href = location.origin + this.$basePath + '/aibase/login';
    },
    setLocalMenuCollapse() {
      this.isCollapse = localStorage.getItem('menu_collapse') === 'true';
    },
    initScroll() {
      const pageContainer = document.querySelector('.el-main .page-container');
      if (pageContainer) {
        pageContainer.scrollTop = 0;
        pageContainer.scrollLeft = 0;
      }
    },
    getCurrentOrgName() {
      return this.orgList.find(item => item.id === this.org.orgId)?.name;
    },
    redirectUserInfo() {
      redirectUserInfoPage(this.permission.isUpdatePassword, () => {
        return null;
      });
    },
    justifyDocPages(val) {
      const path = `${this.$basePath}/aibase` + val;
      return val && path.includes(`${this.$basePath}/aibase/docCenter/pages`);
    },
    justifyIsShowMenu(path) {
      const notShowArr = ['/workflow'];
      let isShowMenu = true;
      if (this.justifyDocPages(path)) {
        isShowMenu = false;
      } else {
        for (let item of notShowArr) {
          if (item === path) {
            isShowMenu = false;
            break;
          }
        }
      }
      this.isShowMenu = isShowMenu;
    },
    async setLanguage() {
      const langCode = localStorage.getItem('locale');
      // 主要解决本地和线上两个 localStorage 语言不同问题，使用用户本地缓存的语言
      if (langCode) await changeLang({ language: langCode });
    },
    setMenuOpeneds(menus) {
      this.defaultOpeneds = menus.map(item => item.index);
    },
    menuClick(item) {
      // 菜单切换时销毁所有缓存组件
      this.cachedViews = [];
      this.$nextTick(() => {
        if (item.redirect) {
          item.redirect();
          this.changeMenuIndex(item.index);
        } else if (item.path) this.$router.push({ path: item.path });
      });
    },
    getCurrentMenu() {
      const { path } = this.$route || {};
      // 获取当前菜单
      this.getMenuList(path);
    },
    getMenuList(path) {
      const menus = this.menuList;

      this.setMenuOpeneds(menus);

      // 给当前 activeIndex 赋值
      this.changeMenuIndex(fetchCurrentPathIndex(path, menus));
    },
    changeMenuIndex(index) {
      this.activeIndex = index;
    },
    addCachedView(route) {
      const cacheName = route.meta?.cacheName;
      if (cacheName && !this.cachedViews.includes(cacheName)) {
        this.cachedViews.push(cacheName);
      }
    },
    async changeOrg(orgId) {
      this.$store.state.user.userInfo.orgId = orgId;
      // 切换组织更新权限，跳转有权限的页面；如果是用模型跳转用模型，其他跳转模型开发平台
      await this.getPermissionInfo();

      // 更新 storage 用户信息中组织 id
      const info = JSON.parse(localStorage.getItem('access_cert'));
      info.user.orgId = orgId;
      localStorage.setItem('access_cert', JSON.stringify(info));

      const { path } = fetchPermFirPath();
      // 如果当前页面 path 与第一个有权限的 path 相同，需要刷新页面以确保数据为新切换组织的
      if (path === this.$route.path) {
        // 切换组织后需清除 query，避免旧会话参数残留
        if (Object.keys(this.$route.query).length) {
          await this.$router.replace({ path });
        }
        location.reload();
        return;
      }
      // 切换组织, 根据当前路径有权限的第一个路径找到对应的 menu
      this.getMenuList(path);
      this.menuClick({ path });
    },
    changeMenuCollapse() {
      this.isCollapse = !this.isCollapse;
      if (!this.isCollapse) {
        this.setMenuOpeneds(this.menuList);
      }
      localStorage.setItem('menu_collapse', this.isCollapse);
    },
  },
};
</script>

<style lang="scss" scoped>
.disabled {
  cursor: not-allowed !important;
}
.full-menu.layout {
  height: 100%;
  .outer-container {
    height: 100%;

    .collapse-menu-divider {
      width: 28px;
      height: 0.5px;
      background-color: #d8d8d8;
      margin: 8px auto;
    }

    .left-aside-container {
      position: relative;
      width: 208px;
      background: #fff;
      border-right: 1px solid #d8d8d8;
      transition: width 0.2s linear;
      .left-header-container {
        position: absolute;
        top: 0;
        width: 100%;
        z-index: 10;
      }
      .left-bottom-container {
        width: 100%;
        height: 68px;
        display: flex;
        align-items: center;
        justify-content: space-around;
        padding: 0 16px;
        .left-bottom-menu-icon {
          img {
            width: 18px;
            height: 18px;
            cursor: pointer;
            margin-top: 5px;
          }
        }
        .header-user-content {
          display: flex;
          align-items: center;
          padding: 0 16px;
          height: 40px;
          width: 142px;
          border-radius: 30px;
          background: #fff;
          color: #1f1f1f;
          box-shadow: 0 2px 8px 0 rgba(16, 18, 25, 0.102);
          font-size: 14px;
          cursor: pointer;
          margin-right: 16px;

          .header-user-name {
            width: 46px;
            overflow: hidden;
            white-space: nowrap;
            text-overflow: ellipsis;
          }

          .header-icon {
            width: 26px;
            height: 26px;
            border-radius: 50%;
            margin-right: 10px;
          }

          .header-more {
            margin-left: 8px;
            transform: rotate(90deg);
            font-size: 16px;
            color: #a2a7b4;
          }
        }
      }
      .left-bottom-container.menu-isCollapse {
        display: block;
        text-align: center;
        height: 90px;
        padding-top: 8px;
        .user-content-isCollapse {
          padding-bottom: 10px;
          border-bottom: 1px solid #d8d8d8;
          margin-bottom: 5px;
          cursor: pointer;
        }
        .header-icon {
          width: 28px;
          height: 28px;
          border-radius: 50%;
          margin: 0 auto;
        }
      }
      .menu-org-popover-icon {
        width: 36px;
        margin-top: 10px;
        padding: 8px;
        border-radius: 12px;
        border: 1px solid #e6e9ed;
      }
      .menu-org-select-wrapper ::v-deep {
        display: flex;
        align-items: center;
        margin: 15px 14px 0;
        border: 1px solid #e6e9ed;
        border-radius: 12px;
        padding-left: 12px;
        .el-select .el-input.is-focus .el-input__inner,
        .el-input__inner,
        .el-input__inner:focus {
          border: none !important;
          outline: none !important;
          padding-left: 12px;
          background-color: rgba(255, 255, 255, 0) !important;
        }
      }
    }

    .left-aside-container.left-aside-container-isCollapse {
      width: 65px;
      .el-aside.full-menu-aside {
        height: calc(100vh - 200px) !important;
      }
    }

    /*element ui 样式重写*/
    .el-aside.full-menu-aside {
      height: calc(100vh - 178px);
      width: 100% !important;
      border-radius: 10px 0 0 10px;
      margin-top: 110px;
      position: relative;
      overflow-y: auto;
      overflow-x: hidden;
      &::-webkit-scrollbar-thumb {
        background-color: transparent;
        transition: background-color 0.2s;
      }
    }

    .menu-hover.el-aside.full-menu-aside {
      &::-webkit-scrollbar-thumb {
        background-color: rgba(16, 18, 25, 0.12) !important;
      }
    }

    /* ===== 展开态：扁平化分组列表 ===== */
    .flat-menu-list {
      padding: 4px 0;
      width: 200px;
      margin-left: 3px;
      .menu-group {
        margin-bottom: 4px;
      }
      .menu-group-title {
        font-size: 12px;
        color: #686f82;
        padding: 16px 16px 6px 24px;
        line-height: 1;
        font-weight: 400;
        white-space: nowrap;
      }
      .menu-group:first-of-type {
        .menu-group-title {
          padding-top: 5px;
        }
      }
      .flat-menu-item {
        display: flex;
        align-items: center;
        height: 36px;
        padding: 0 8px;
        cursor: pointer;
        border-radius: 12px;
        margin: 3px 16px;
        font-size: 14px;
        color: $menu_text_color;
        &:hover {
          background-color: #f7f8fa;
        }
        .menu-svg {
          padding-top: 2px;
          flex-shrink: 0;
          .menu-icon {
            font-size: 16px;
            margin-right: 8px;
          }
          .svg-icon {
            color: #868d9c;
          }
        }
        .menu-withIcon-title {
          display: inline-block;
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
          font-size: 14px;
          font-weight: 400;
          color: $menu_text_color;
        }
      }
      .flat-menu-item-active {
        background-color: $menu_bg_color !important;
        .menu-withIcon-title {
          font-weight: 500 !important;
          color: $menu_text_color !important;
        }
        .menu-svg .svg-icon {
          color: $menu_text_color !important;
        }
      }
    }

    /* ===== 收起态：仅图标 + 分组分隔 ===== */
    .collapse-menu-list {
      padding: 6px 0 4px 0;
      display: flex;
      flex-direction: column;
      align-items: center;
      width: 40px;
      margin-left: 12px;
      .collapse-menu-item {
        width: 40px;
        height: 32px;
        border-radius: 12px;
        display: flex;
        align-items: center;
        justify-content: center;
        cursor: pointer;
        margin: 2px 0;
        .menu-icon {
          font-size: 18px;
        }
        .svg-icon {
          color: #868d9c;
        }
        &:hover {
          background-color: #f7f8fa;
        }
      }
      .collapse-menu-item-active {
        background-color: #f7f8fa;
        .svg-icon {
          color: $menu_text_color !important;
        }
      }
    }

    .inner-container {
      width: calc(100% - 80px);
      height: 100%;
      border-radius: 10px;
      .el-header {
        line-height: 60px;
        height: 60px;
        box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.05);
        z-index: 2;
        background: url('@/assets/imgs/nav_bg.png');
        background-size: 100% 100%;
      }
      .el-main {
        position: relative;
        margin: 0 !important;
        padding: 0 !important;
        width: 100%;
        height: 100%;
        overflow: auto;
        .page-container {
          height: 100%;
          overflow-x: auto;
          .right-page-content {
            min-width: 1250px;
            min-height: calc(100% - 32px);
            padding: 16px;
          }
        }
      }
      .no-header-main {
        height: 100%;
        .page-container {
          height: 100%;
          .right-page-content {
            height: 100%;
            padding: 0;
          }
        }
      }
    }
  }
}

.menu--popover-wrap {
  border-bottom: 1px solid #ebebeb;
  padding: 4px 0 6px 0;
}
.menu--popover-wrap:first-of-type {
  padding-top: 0;
}
.menu--popover-wrap.wrap-last {
  padding-bottom: 0;
  border-bottom: none;
}
.menu--popover-item {
  font-size: 13px;
  color: $menu_text_color;
  height: 34px;
  line-height: 34px;
  cursor: pointer;
  border-radius: 4px;
  padding: 0 8px;
  .menu--popover-item-img {
    height: 16px;
    display: inline-block;
    vertical-align: middle;
    margin-right: 5px;
  }
  .menu--popover-item-name {
    font-size: 13px;
    color: $menu_text_color;
    display: inline-block;
    vertical-align: middle;
  }
  .menu--popover-item-icon {
    width: 16px;
    float: right;
    margin-top: 13px;
  }
  .menu--popover-item-version {
    font-size: 13px;
    float: right;
  }
  .menu--popover-item-version:after {
    display: inline-block;
    content: '';
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: #f59a23;
    margin-bottom: 2px;
  }
}
.menu--popover-item:hover ::v-deep {
  background: #f5f7fa !important;
  .el-input .el-input__inner {
    border: none !important;
  }
}
</style>
