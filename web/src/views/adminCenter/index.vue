<template>
  <div class="page-wrapper">
    <div class="page-title">
      <i class="el-icon-arrow-left" @click="$router.go(-1)" />
      <img class="page-title-img" src="@/assets/imgs/org.png" alt="" />
      <span class="page-title-name">{{ $t('adminCenter.title') }}</span>
    </div>
    <div class="page-wrapper-content">
      <div class="page-container">
        <section
          v-for="(groupName, groupKey) in visibleGroups"
          :key="groupKey"
          class="group-section"
        >
          <!-- 分组标题（带蓝色侧边条） -->
          <div class="group-header">
            <div class="blue-bar"></div>
            <h2>{{ groupName }}</h2>
          </div>

          <!-- 卡片网格布局 -->
          <div class="card-grid">
            <!-- 过滤并渲染属于当前大组的卡片 -->
            <div
              v-for="(module, moduleKey) in getModulesByGroup(groupKey)"
              :key="moduleKey"
              class="module-card"
            >
              <!-- 卡片头部（图标 + 标题） -->
              <div class="card-header">
                <div class="icon-wrapper">
                  <svg-icon
                    class="module-icon"
                    :icon-class="moduleIcons[moduleKey]"
                  />
                </div>
                <h3>{{ module.title }}</h3>
              </div>

              <!-- 卡片子菜单链接 -->
              <div
                :class="[
                  'card-links',
                  moduleKey === 'resourcePool' ? 'grid-layout' : 'flex-layout',
                ]"
              >
                <!-- 仅遍历叶子节点链接 -->
                <a
                  v-for="(linkObj, linkKey) in getSubLinks(module)"
                  :key="linkKey"
                  href="#"
                  class="menu-link"
                  @click.prevent="handleNavigate(linkObj)"
                >
                  {{ linkObj.title }}
                </a>
              </div>
            </div>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<script>
import { checkPerm, PERMS } from '@/router/permission';

export default {
  data() {
    return {
      menuData: {
        group: {
          globalManage: this.$t('adminCenter.group.globalManage'),
          productBackend: this.$t('adminCenter.group.productBackend'),
        },
        pageModules: {
          // --- 全局管理 ---
          personnelManage: {
            title: this.$t('adminCenter.pageModules.personnelManage.title'),
            group: 'globalManage',
            perm: PERMS.ADMIN_CENTER,
            organization: {
              title: this.$t(
                'adminCenter.pageModules.personnelManage.organization',
              ),
              path: '/permission?key=org',
            },
            user: {
              title: this.$t('adminCenter.pageModules.personnelManage.user'),
              path: '/permission?key=user',
            },
            role: {
              title: this.$t('adminCenter.pageModules.personnelManage.role'),
              path: '/permission?key=role',
            },
          },
          // operationManage: {
          //   title: this.$t('adminCenter.pageModules.operationManage.title'),
          //   group: 'globalManage',
          //   approvalProcess: {
          //     title: this.$t(
          //       'adminCenter.pageModules.operationManage.approvalProcess',
          //     ),
          //     path: '/admin/operation/approval',
          //   },
          // },
          platformConfig: {
            title: this.$t('adminCenter.pageModules.platformConfig.title'),
            group: 'globalManage',
            appearance: {
              title: this.$t(
                'adminCenter.pageModules.platformConfig.appearance',
              ),
              path: '/permission?key=platformConfig',
              perm: PERMS.SETTING,
            },
            oAuth: {
              title: this.$t('oauth.title'),
              path: '/permission?key=oAuth',
              perm: PERMS.OAUTH,
            },
          },
          // --- 产品后台 ---
          modelService: {
            title: this.$t('adminCenter.pageModules.modelService.title'),
            group: 'productBackend',
            modelManage: {
              title: this.$t(
                'adminCenter.pageModules.modelService.modelManage',
              ),
              path: '/adminCenter/modelConfig',
            },
          },
          resourcePool: {
            title: this.$t('adminCenter.pageModules.resourcePool.title'),
            group: 'productBackend',
            knowledge: {
              title: this.$t(
                'adminCenter.pageModules.resourcePool.knowledge.title',
              ),
              path: '/adminCenter/knowledge',
            },
            mcp: {
              title: this.$t('adminCenter.pageModules.resourcePool.mcp.title'),
              path: '/adminCenter/mcp',
            },
            tool: {
              title: this.$t('adminCenter.pageModules.resourcePool.tool.title'),
              path: '/adminCenter/tool',
            },
            prompt: {
              title: this.$t(
                'adminCenter.pageModules.resourcePool.prompt.title',
              ),
              path: '/adminCenter/prompt',
            },
            skill: {
              title: this.$t(
                'adminCenter.pageModules.resourcePool.skill.title',
              ),
              path: '/adminCenter/skill',
            },
            safety: {
              title: this.$t(
                'adminCenter.pageModules.resourcePool.safety.title',
              ),
              path: '/adminCenter/safety',
            },
          },
          appDevelopment: {
            title: this.$t('adminCenter.pageModules.appDevelopment.title'),
            group: 'productBackend',
            agent: {
              title: this.$t(
                'adminCenter.pageModules.appDevelopment.agent.title',
              ),
              path: '/adminCenter/agent',
            },
            workflow: {
              title: this.$t(
                'adminCenter.pageModules.appDevelopment.workflow.title',
              ),
              path: '/adminCenter/workflow',
            },
            rag: {
              title: this.$t(
                'adminCenter.pageModules.appDevelopment.rag.title',
              ),
              path: '/adminCenter/rag',
            },
          },
        },
      },
      moduleIcons: {
        personnelManage: 'users',
        operationManage: 'chart-bar-popular',
        platformConfig: 'adjustments',
        modelService: 'cpu',
        resourcePool: 'box',
        appDevelopment: 'brand-appstore',
      },
    };
  },
  computed: {
    visibleGroups() {
      const groups = {};
      Object.keys(this.menuData.group).forEach(groupKey => {
        if (Object.keys(this.getModulesByGroup(groupKey)).length) {
          groups[groupKey] = this.menuData.group[groupKey];
        }
      });
      return groups;
    },
  },
  methods: {
    checkPerm,
    // 过滤出属于特定分组的模块
    getModulesByGroup(groupKey) {
      const filtered = {};
      Object.keys(this.menuData.pageModules).forEach(key => {
        const module = this.menuData.pageModules[key];
        if (
          module.group === groupKey &&
          this.checkPerm(module.perm) &&
          Object.keys(this.getSubLinks(module)).length
        ) {
          filtered[key] = module;
        }
      });
      return filtered;
    },
    // 仅仅筛选出属于路由配置对象的真正叶子节点
    getSubLinks(module) {
      const links = {};
      Object.keys(module).forEach(key => {
        // 排除掉卡片的 title 属性和用来分组的 group 属性，剩下的就是叶子节点对象
        if (
          key !== 'title' &&
          key !== 'group' &&
          key !== 'perm' &&
          this.checkPerm(module[key].perm)
        ) {
          links[key] = module[key];
        }
      });
      return links;
    },
    // 核心路由/跳转方法
    handleNavigate(linkItem) {
      this.$router.push(linkItem.path);
    },
  },
  mounted() {},
};
</script>

<style lang="scss" scoped>
@import './styles/common.scss';

.page-title {
  .el-icon-arrow-left {
    margin-right: 10px;
    font-size: 15px;
    cursor: pointer;
    color: $color_title;
  }
}
.group-section {
  margin-bottom: 48px;

  .card-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 24px;

    .module-card {
      background-color: #ffffff;
      border: 1px solid #f3f4f6;
      border-radius: 12px;
      padding: 24px;
      box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
      transition: box-shadow 0.2s ease;
      max-width: 330px;
      width: 100%;

      &:hover {
        box-shadow:
          0 4px 6px -1px rgba(0, 0, 0, 0.1),
          0 2px 4px -1px rgba(0, 0, 0, 0.06);
      }

      .card-header {
        display: flex;
        align-items: center;
        gap: 12px;
        margin-bottom: 16px;

        .icon-wrapper {
          color: #9ca3af;
          display: flex;
          align-items: center;

          .module-icon {
            font-size: 24px;
          }
        }

        h3 {
          font-size: 16px;
          font-weight: 700;
          color: #1f2937;
          margin: 0;
        }
      }

      .card-links {
        .menu-link {
          font-size: 14px;
          color: $color;
          text-decoration: none;
          font-weight: 500;
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
          cursor: pointer;

          &:hover {
            text-decoration: none;
            color: #1d4ed8;
          }
        }

        &.flex-layout {
          display: flex;
          flex-wrap: wrap;
          column-gap: 16px;
          row-gap: 8px;
        }

        &.grid-layout {
          display: grid;
          grid-template-columns: repeat(3, minmax(0, 1fr));
          row-gap: 12px;
          column-gap: 8px;
        }
      }
    }
  }
}
</style>
