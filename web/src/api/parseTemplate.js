import service from '@/utils/request';
import { USER_API } from '@/utils/requestConstants';

export const getParseTemplateList = params => {
  return service({
    url: `${USER_API}/knowledge/parseTemplate`,
    method: 'get',
    params,
  });
};

export const createParseTemplate = data => {
  return service({
    url: `${USER_API}/knowledge/parseTemplate`,
    method: 'post',
    data,
  });
};

export const updateParseTemplate = data => {
  return service({
    url: `${USER_API}/knowledge/parseTemplate`,
    method: 'put',
    data,
  });
};

export const deleteParseTemplate = data => {
  return service({
    url: `${USER_API}/knowledge/parseTemplate`,
    method: 'delete',
    data,
  });
};

export const updateKnowledgeParseTemplate = data => {
  return service({
    url: `${USER_API}/knowledge/parseTemplate/bind`,
    method: 'put',
    data,
  });
};
