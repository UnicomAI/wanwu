import router from './index';
import { store } from '@/store/index';
import { fetchPermFirPath } from '@/utils/util';
import { PERMS as menuPerms } from './constants';
import { basePath } from '@/utils/config';

const white_list = [
  basePath + '/aibase',
  '/oauth',
  '/login',
  '/webChat',
  '/register',
  '/reset',
  '/templateSquare',
];
export const PERMS = menuPerms;

export const checkPerm = perm => {
  // 不传权限点，表示不需要权限控制，返回 true
  if (!perm) return true;
  // 传权限点，判断是否在权限列表，在返回 true，否则 false
  const permission = store.getters['user/permission'];
  const orgPermission = permission.orgPermission;
  if (orgPermission && orgPermission.length) {
    return Array.isArray(perm)
      ? perm.some(item => orgPermission.includes(item))
      : orgPermission.includes(perm);
  }
  return false;
};

export const formatPerms = perms => {
  return perms && perms.length ? perms.map(item => item.perm) : [];
};

router.beforeEach(async (to, from, next) => {
  let token = '';
  let access_cert =
    localStorage.getItem('access_cert') &&
    JSON.parse(localStorage.getItem('access_cert'));
  if (access_cert) {
    token = access_cert.user.token;
  }
  if (token && !access_cert.user.is2FA) {
    if (to.path === '/') {
      const { path } = fetchPermFirPath();
      next({ path });
    } else {
      next();
    }
  } else {
    if (
      white_list.some(item => {
        return to.path.indexOf(item) > -1;
      })
    ) {
      next();
    } else {
      window.location.href =
        window.location.origin + basePath + '/aibase/login';
    }
  }
});

// 管理员中心列表页查询参数保持清空
router.afterEach(to => {
  const shouldKeepAdminListState =
    to.path.startsWith('/adminCenter/') || to.path === '/workflow';
  if (!shouldKeepAdminListState) {
    store.commit('adminCenter/CLEAR_ACTIVE_LIST');
  }
});
