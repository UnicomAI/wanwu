import service from '@/utils/request';
import { USER_API } from '@/utils/requestConstants';

/**
 * 创建对话日志导出任务。
 * @param {Object} data
 * @param {string} data.appId - 应用 ID。
 * @param {string} data.appType - 应用类型。
 * @param {string[]} [data.logIds] - 待导出的对话日志 ID；不传时导出全部。
 */
export const exportConversationLogs = data => {
  return service({
    url: `${USER_API}/conversation/log/export`,
    method: 'post',
    data,
  });
};

/**
 * 删除对话日志导出记录。
 * @param {Object} data
 * @param {string[]} data.exportRecordIds - 导出记录 ID 列表。
 */
export const deleteConversationLogExportRecords = data => {
  return service({
    url: `${USER_API}/conversation/log/export/record`,
    method: 'delete',
    data,
  });
};

/**
 * 获取对话日志导出记录列表。
 * @param {Object} params - 查询参数。
 * @param {string} params.appId - 应用 ID。
 * @param {string} params.appType - 应用类型。
 * @param {number} params.pageNo - 页码。
 * @param {number} params.pageSize - 每页条数。
 */
export const getConversationLogExportRecordList = params => {
  return service({
    url: `${USER_API}/conversation/log/export/record/list`,
    method: 'get',
    params,
  });
};
