import {
  getDraftConversationLogDetail,
  getDraftConversationLogList,
  getDraftConversationLogUserSelect,
} from '@/api/agent';
import {
  getAdminAssistantConversationLogDetail,
  getAdminAssistantConversationLogList,
  getAdminAssistantConversationLogUserSelect,
} from '@/api/adminCenter';

export const draftConversationLogService = {
  getDetail: getDraftConversationLogDetail,
  getList: getDraftConversationLogList,
  getUsers: getDraftConversationLogUserSelect,
};

export const adminConversationLogService = {
  getDetail: getAdminAssistantConversationLogDetail,
  getList: getAdminAssistantConversationLogList,
  getUsers: getAdminAssistantConversationLogUserSelect,
};
