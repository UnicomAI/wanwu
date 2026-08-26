import { md } from '@/mixins/markdown-it';
import { i18n } from '@/lang';
import {
  convertLatexSyntax,
  parseSub,
  parseSubConversation,
} from '@/utils/util.js';
import { processToolResultBlocks } from '@/utils/toolResultProcessor.js';
import { AGENT_MESSAGE_CONFIG } from '@/components/stream/constants';

const normalizeSearchList = searchList => {
  if (typeof searchList !== 'string') return searchList || [];

  try {
    return JSON.parse(searchList || '[]');
  } catch (error) {
    return [];
  }
};

const renderMainResponse = (response, index) => {
  return md.render(
    parseSub(
      processToolResultBlocks(convertLatexSyntax(response || '')),
      index,
    ),
  );
};

const renderSubResponse = (response, index, searchList, subId) => {
  return md.render(
    parseSubConversation(
      processToolResultBlocks(convertLatexSyntax(response || '')),
      index,
      normalizeSearchList(searchList),
      subId,
    ),
  );
};

const renderNestedSubTextResponse = (response, index, searchList, subId) => {
  return md.render(
    parseSubConversation(
      convertLatexSyntax(response || ''),
      index,
      searchList,
      subId,
    ),
  );
};

const buildMainMessageSequence = (record, index) => {
  const sequence = [];
  let fullResponse = '';
  let hasHistoryError = false;

  if (record.responseList && record.responseList.length) {
    record.responseList.forEach(item => {
      fullResponse += item.response || '';
      if (item.errMessage || item.errResponse) hasHistoryError = true;

      sequence.push({
        type: 'main',
        order: item.order,
        renderedContent: renderMainResponse(item.response, index),
        errMsg: item.errMessage,
        errResponse: item.errResponse,
      });
    });
  } else if (record.response) {
    fullResponse = record.response;
  }

  return { sequence, fullResponse, hasHistoryError };
};

const transformSubConversation = (conversation, index) => {
  const searchList = normalizeSearchList(conversation.searchList);
  const response = conversation.response || '';
  const citationsTagList = (response.match(/\【([0-9]{0,2})\^\】/g) || []).map(
    item => Number(item.match(/\【([0-9]{0,2})\^\】/)[1]),
  );
  const messageSequence = [];

  if (response) {
    messageSequence.push({
      type: 'main',
      order: conversation.order,
      response,
      renderedContent: renderSubResponse(
        response,
        index,
        searchList,
        conversation.id,
      ),
    });
  }

  return {
    ...conversation,
    citationsTagList,
    messageSequence,
    searchList,
    response: renderSubResponse(response, index, searchList, conversation.id),
  };
};

const assembleSubConversationTree = (subConversions, index) => {
  const sequence = [];
  const textConversationType = AGENT_MESSAGE_CONFIG.SUB_TEXT.CONVERSATION_TYPE;

  subConversions.forEach(conversation => {
    if (conversation.parentId) {
      const parent = subConversions.find(
        item => item.id === conversation.parentId,
      );
      if (!parent) return;

      if (conversation.conversationType === textConversationType) {
        parent.messageSequence.push({
          type: 'main',
          order: conversation.order,
          response: conversation.response,
          renderedContent: conversation.response
            ? renderNestedSubTextResponse(
                conversation.response,
                index,
                conversation.searchList,
                conversation.id,
              )
            : '',
        });
      } else {
        parent.messageSequence.push({
          type: 'sub',
          id: conversation.id,
          order: conversation.order,
        });
      }
      parent.messageSequence.sort((a, b) => (a.order || 0) - (b.order || 0));
    } else if (conversation.conversationType !== textConversationType) {
      sequence.push({
        type: 'sub',
        id: conversation.id,
        order: conversation.order,
      });
    }
  });

  return sequence.sort((a, b) => (a.order || 0) - (b.order || 0));
};

const transformAgentHistoryItem = (record, index) => {
  const { sequence, fullResponse, hasHistoryError } = buildMainMessageSequence(
    record,
    index,
  );
  const subConversions = (record.subConversationList || []).map(item =>
    transformSubConversation(item, index),
  );
  const subSequence = assembleSubConversationTree(subConversions, index);

  return {
    ...record,
    error: record.error || hasHistoryError,
    query: record.prompt,
    finish: 1,
    response: renderMainResponse(fullResponse, index),
    oriResponse: fullResponse,
    searchList: normalizeSearchList(record.searchList),
    fileList: record.requestFiles,
    gen_file_url_list: record.responseFileUrls || [],
    subConversions,
    messageSequence: [...sequence, ...subSequence].sort(
      (a, b) => (a.order || 0) - (b.order || 0),
    ),
    isOpen: true,
    toolText: i18n.t('agent.tooled'),
    thinkText: i18n.t('agent.thinked'),
    showScrollBtn: null,
  };
};

/**
 * 将智能体历史接口记录转换为 streamMessageField 可渲染的数据结构。
 * @param {Object[]} records 智能体历史接口的 list 数据。
 * @returns {Object[]} streamMessageField 历史数据。
 */
export const transformAgentHistory = records => {
  if (!Array.isArray(records)) return [];
  return records.map((record, index) =>
    transformAgentHistoryItem(record, index),
  );
};
