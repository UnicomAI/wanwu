import request from '@/utils/request';
import { USER_API } from '@/utils/requestConstants';

const NOTICE_API = `${USER_API}/notice`;

/* 获取当前组织下的未读总数及分类未读数。
 */
export const getUnreadMessageCount = () => {
  return request({
    url: `${NOTICE_API}/unread/count`,
    method: 'get',
  });
};

/* 获取消息中心分页列表，也可用于气泡中的未读消息预览。
 */
export const getMessageList = params => {
  return request({
    url: `${NOTICE_API}/list`,
    method: 'get',
    params,
  });
};

/* 将当前组织下全部可见未读消息标记为已读。
 */
export const markAllMessagesAsRead = () => {
  return request({
    url: `${NOTICE_API}/read-all`,
    method: 'put',
  });
};

/* 将单条消息标记为已读。data 格式：{ messageId: string }。
 */
export const markMessageAsRead = data => {
  return request({
    url: `${NOTICE_API}/read`,
    method: 'put',
    data,
  });
};

/* 批量删除消息。data 格式：{ ids: string[] }。
 */
export const deleteMessages = data => {
  return request({
    url: NOTICE_API,
    method: 'delete',
    data,
  });
};
