import copy
import os
import random
import time
import configparser
from abc import ABC, abstractmethod
from pprint import pformat
from typing import Any, Dict, Iterator, List, Literal, Optional, Tuple, Union

from utils.actions.llm.schema import DEFAULT_SYSTEM_MESSAGE, SYSTEM, USER, Message
from utils.actions.log import logger
from utils.actions.settings import DEFAULT_MAX_INPUT_TOKENS
from utils.actions.utils.tokenization_qwen import tokenizer
from utils.actions.utils.utils import (extract_text_from_message, format_as_multimodal_message, has_chinese_messages,
                                    merge_generate_cfgs, print_traceback)

LLM_REGISTRY = {}

config = configparser.ConfigParser()
config.read('config.ini')

MODEL_NAME_CONFIG = config["MODELS"]["default_llm"]
MODEL_NAME = os.getenv('CUAI_DEFAULT_LLM_MODEL_ID', MODEL_NAME_CONFIG)

MODEL_URL_CONFIG = config["MODELS"]["model_url"]
MODEL_URL = os.getenv('CUAI_DEFAULT_LLM_MODEL_URL', MODEL_URL_CONFIG)

def register_llm(model_type):

    def decorator(cls):
        LLM_REGISTRY[model_type] = cls
        return cls

    return decorator


class ModelServiceError(Exception):

    def __init__(self,
                 exception: Optional[Exception] = None,
                 code: Optional[str] = None,
                 message: Optional[str] = None):
        if exception is not None:
            super().__init__(exception)
        else:
            super().__init__(f'\nError code: {code}. Error message: {message}')
        self.exception = exception
        self.code = code
        self.message = message


class BaseChatModel(ABC):
    """The base class of LLM"""

    def __init__(self,model_name = MODEL_NAME, model_url = MODEL_URL,cfg: Optional[Dict] = None):
        cfg = cfg or {}
        self.model = cfg.get('model', '').strip()
        generate_cfg = copy.deepcopy(cfg.get('generate_cfg', {}))
        self.max_retries = generate_cfg.pop('max_retries', 0)
        self.generate_cfg = generate_cfg
        self.model_name = model_name
        self.model_url = model_url

    def chat(
        self,
        messages: List[Union[Message, Dict]],
        functions: Optional[List[Dict]] = None,
        function_calls_list=None,
        model_name = None,
        model_url = None,
        stream: bool = True,
        delta_stream: bool = False,
        extra_generate_cfg: Optional[Dict] = None,
    ) -> Union[List[Message], List[Dict], Iterator[List[Message]], Iterator[List[Dict]]]:
        """LLM chat interface.

        Args:
            messages: Inputted messages.
            functions: Inputted functions for function calling. OpenAI format supported.
            stream: Whether to use streaming generation.
            delta_stream: Whether to stream the response incrementally.
              (1) When False (recommended): Stream the full response every iteration.
              (2) When True: Stream the chunked response, i.e, delta responses.
            extra_generate_cfg: Extra LLM generation hyper-paramters.

        Returns:
            the generated message list response by llm.
        """

        if stream and delta_stream:
            logger.warning(
                'Support for `delta_stream=True` is deprecated. '
                'Please use `stream=True and delta_stream=False` or `stream=False` instead. '
                'Using `delta_stream=True` makes it difficult to implement advanced postprocessing and retry mechanisms.'
            )

        generate_cfg = merge_generate_cfgs(base_generate_cfg=self.generate_cfg, new_generate_cfg=extra_generate_cfg)
        if 'lang' in generate_cfg:
            lang: Literal['en', 'zh'] = generate_cfg.pop('lang')
        else:
            lang: Literal['en', 'zh'] = 'zh' if has_chinese_messages(messages) else 'en'

        messages = copy.deepcopy(messages)

        _return_message_type = 'dict'
        new_messages = []
        for msg in messages:
            if isinstance(msg, dict):
                new_messages.append(Message(**msg))
            else:
                new_messages.append(msg)
                _return_message_type = 'message'
        messages = new_messages

        if messages[0].role != SYSTEM:
            messages = [Message(role=SYSTEM, content=DEFAULT_SYSTEM_MESSAGE)] + messages

        # Not precise. It's hard to estimate tokens related with function calling and multimodal items.
        max_input_tokens = generate_cfg.pop('max_input_tokens', DEFAULT_MAX_INPUT_TOKENS)
        if max_input_tokens > 0:
            messages = _truncate_input_messages_roughly(
                messages=messages,
                max_tokens=max_input_tokens,
            )

        if functions:
            fncall_mode = True
        else:
            fncall_mode = False
        if 'function_choice' in generate_cfg:
            fn_choice = generate_cfg['function_choice']
            valid_fn_choices = [f.get('name', f.get('name_for_model', None)) for f in (functions or [])]
            valid_fn_choices = ['auto', 'none'] + [f for f in valid_fn_choices if f]
            if fn_choice not in valid_fn_choices:
                raise ValueError(f'The value of function_choice must be one of the following: {valid_fn_choices}. '
                                 f'But function_choice="{fn_choice}" is received.')
            if fn_choice == 'none':
                fncall_mode = False

        # Note: the preprocessor's behavior could change if it receives function_choice="none"
        # messages = self._preprocess_messages(messages, lang=lang, generate_cfg=generate_cfg)
        # print(f"messages的类型为：{type(messages)}")


        if not fncall_mode:
            for k in ['parallel_function_calls', 'function_choice']:
                if k in generate_cfg:
                    del generate_cfg[k]

        def _call_model_service():
            if fncall_mode:
                return self._chat_with_functions(
                    messages=messages,
                    functions=functions,
                    function_calls_list=function_calls_list,
                    model_name = model_name,
                    model_url = model_url,
                    stream=stream,
                    delta_stream=delta_stream,
                    generate_cfg=generate_cfg,
                    lang=lang,
                )
            else:
                return self._chat(
                    messages,
                    model_name = model_name,
                    model_url = model_url,
                    stream=stream,
                    delta_stream=delta_stream,
                    generate_cfg=generate_cfg,
                )

        if stream and delta_stream:
            # No retry for delta streaming
            output = _call_model_service()
        elif stream and (not delta_stream):
            output = retry_model_service_iterator(_call_model_service, max_retries=self.max_retries)
        else:
            output = retry_model_service(_call_model_service, max_retries=self.max_retries)

        if isinstance(output, list):
            assert not stream
            logger.debug(f'LLM Output:\n{pformat([_.model_dump() for _ in output], indent=2)}')
            output = self._postprocess_messages(output, fncall_mode=fncall_mode, generate_cfg=generate_cfg)
            return self._convert_messages_to_target_type(output, _return_message_type)
        else:
            assert stream
            if delta_stream:
                # Hack: To avoid potential errors during the postprocessing of stop words when delta_stream=True.
                # Man, we should never have implemented the support for `delta_stream=True` in the first place!
                generate_cfg = copy.deepcopy(generate_cfg)  # copy to avoid conflicts with `_call_model_service`
                assert 'skip_stopword_postproc' not in generate_cfg
                generate_cfg['skip_stopword_postproc'] = True
            output = self._postprocess_messages_iterator(output, fncall_mode=fncall_mode, generate_cfg=generate_cfg)
            return self._convert_messages_iterator_to_target_type(output, _return_message_type)

    def _chat(
        self,
        messages: List[Union[Message, Dict]],
        model_name ,
        model_url ,
        stream: bool,
        delta_stream: bool,
        generate_cfg: dict
    ) -> Union[List[Message], Iterator[List[Message]]]:
        if stream:
            return self._chat_stream(messages, model_name = model_name,model_url = model_url,delta_stream=delta_stream, generate_cfg=generate_cfg)
        else:
            return self._chat_no_stream(messages,model_name = model_name,model_url = model_url, generate_cfg=generate_cfg)

    @abstractmethod
    def _chat_with_functions(
        self,
        messages: List[Union[Message, Dict]],
        functions: List[Dict],
        function_calls_list: List[Dict],
        model_name,
        model_url,
        stream: bool,
        delta_stream: bool,
        generate_cfg: dict,
        lang: Literal['en', 'zh'],
    ) -> Union[List[Message], Iterator[List[Message]]]:
        raise NotImplementedError

    @abstractmethod
    def _chat_stream(
        self,
        messages: List[Message],
        model_name,
        model_url,
        delta_stream: bool,
        generate_cfg: dict,
    ) -> Iterator[List[Message]]:
        raise NotImplementedError

    @abstractmethod
    def _chat_no_stream(
        self,
        messages: List[Message],
        model_name,
        model_url,
        generate_cfg: dict,
    ) -> List[Message]:
        raise NotImplementedError

    def _preprocess_messages(self, messages: List[Message], lang: Literal['en', 'zh'],
                             generate_cfg: dict) -> List[Message]:
        messages = [format_as_multimodal_message(msg, add_upload_info=True, lang=lang) for msg in messages]
        return messages

    def _postprocess_messages(
        self,
        messages: List[Message],
        fncall_mode: bool,
        generate_cfg: dict,
    ) -> List[Message]:
        messages = [format_as_multimodal_message(msg, add_upload_info=False) for msg in messages]
        if not generate_cfg.get('skip_stopword_postproc', False):
            stop = generate_cfg.get('stop', [])
            messages = _postprocess_stop_words(messages, stop=stop)
        return messages

    def _postprocess_messages_iterator(
        self,
        messages: Iterator[List[Message]],
        fncall_mode: bool,
        generate_cfg: dict,
    ) -> Iterator[List[Message]]:
        pre_msg = []
        for pre_msg in messages:
            post_msg = self._postprocess_messages(pre_msg, fncall_mode=fncall_mode, generate_cfg=generate_cfg)
            if post_msg:
                yield post_msg
        logger.debug(f'LLM Output:\n{pformat([_.model_dump() for _ in pre_msg], indent=2)}')

    def _convert_messages_to_target_type(self, messages: List[Message],
                                         target_type: str) -> Union[List[Message], List[Dict]]:
        if target_type == 'message':
            return [Message(**x) if isinstance(x, dict) else x for x in messages]
        elif target_type == 'dict':
            return [x.model_dump() if not isinstance(x, dict) else x for x in messages]
        else:
            raise NotImplementedError

    def _convert_messages_iterator_to_target_type(
            self, messages_iter: Iterator[List[Message]],
            target_type: str) -> Union[Iterator[List[Message]], Iterator[List[Dict]]]:
        for messages in messages_iter:
            yield self._convert_messages_to_target_type(messages, target_type)


def _postprocess_stop_words(messages: List[Message], stop: List[str]) -> List[Message]:
    messages = copy.deepcopy(messages)

    # Make sure it stops before stop words.
    trunc_messages = []
    for msg in messages:
        truncated = False
        trunc_content = []
        for i, item in enumerate(msg.content):
            item_type, item_text = item.get_type_and_value()
            if item_type == 'text':
                truncated, item.text = _truncate_at_stop_word(text=item_text, stop=stop)
            trunc_content.append(item)
            if truncated:
                break
        msg.content = trunc_content
        trunc_messages.append(msg)
        if truncated:
            break
    messages = trunc_messages

    # It may ends with partial stopword 'Observation' when the full stopword is 'Observation:'.
    # The following post-processing step removes partial stop words.
    partial_stop = []
    for s in stop:
        s = tokenizer.tokenize(s)[:-1]
        if s:
            s = tokenizer.convert_tokens_to_string(s)
            partial_stop.append(s)
    partial_stop = sorted(set(partial_stop))
    last_msg = messages[-1].content
    for i in range(len(last_msg) - 1, -1, -1):
        item_type, item_text = last_msg[i].get_type_and_value()
        if item_type == 'text':
            for s in partial_stop:
                if item_text.endswith(s):
                    last_msg[i].text = item_text[:-len(s)]
            break

    return messages


def _truncate_at_stop_word(text: str, stop: List[str]):
    truncated = False
    for s in stop:
        k = text.find(s)
        if k >= 0:
            truncated = True
            text = text[:k]
    return truncated, text


def _truncate_input_messages_roughly(messages: List[Message], max_tokens: int) -> List[Message]:
    sys_msg = messages[0]
    assert sys_msg.role == SYSTEM  # The default system is prepended if none exists
    if len([m for m in messages if m.role == SYSTEM]) >= 2:
        raise ModelServiceError(
            code='400',
            message='The input messages must contain no more than one system message. '
            ' And the system message, if exists, must be the first message.',
        )

    turns = []
    for m in messages[1:]:
        if m.role == USER:
            turns.append([m])
        else:
            if turns:
                turns[-1].append(m)
            else:
                raise ModelServiceError(
                    code='400',
                    message='The input messages (excluding the system message) must start with a user message.',
                )

    def _count_tokens(msg: Message) -> int:
        return tokenizer.count_tokens(extract_text_from_message(msg, add_upload_info=True))

    token_cnt = _count_tokens(sys_msg)
    truncated = []
    for i, turn in enumerate(reversed(turns)):
        cur_turn_msgs = []
        cur_token_cnt = 0
        for m in reversed(turn):
            cur_turn_msgs.append(m)
            cur_token_cnt += _count_tokens(m)
        # Check "i == 0" so that at least one user message is included
        if (i == 0) or (token_cnt + cur_token_cnt <= max_tokens):
            truncated.extend(cur_turn_msgs)
            token_cnt += cur_token_cnt
        else:
            break
    # Always include the system message
    truncated.append(sys_msg)
    truncated.reverse()

    if len(truncated) < 2:  # one system message + one or more user messages
        raise ModelServiceError(
            code='400',
            message='At least one user message should be provided.',
        )
    if token_cnt > max_tokens:
        raise ModelServiceError(
            code='400',
            message=f'The input messages exceed the maximum context length ({max_tokens} tokens) after '
            f'keeping only the system message and the latest one user message (around {token_cnt} tokens). '
            'To configure the context limit, please specifiy "max_input_tokens" in the model generate_cfg. '
            f'Example: generate_cfg = {{..., "max_input_tokens": {(token_cnt // 100 + 1) * 100}}}',
        )
    return truncated


def retry_model_service(
    fn,
    max_retries: int = 10,
) -> Any:
    """Retry a function"""

    num_retries, delay = 0, 1.0
    while True:
        try:
            return fn()

        except ModelServiceError as e:
            num_retries, delay = _raise_or_delay(e, num_retries, delay, max_retries)


def retry_model_service_iterator(
    it_fn,
    max_retries: int = 10,
) -> Iterator:
    """Retry an iterator"""

    num_retries, delay = 0, 1.0
    while True:
        try:
            for rsp in it_fn():
                yield rsp
            break

        except ModelServiceError as e:
            num_retries, delay = _raise_or_delay(e, num_retries, delay, max_retries)


def _raise_or_delay(
    e: ModelServiceError,
    num_retries: int,
    delay: float,
    max_retries: int = 10,
    max_delay: float = 300.0,
    exponential_base: float = 2.0,
) -> Tuple[int, float]:
    """Retry with exponential backoff"""

    if max_retries <= 0:  # no retry
        raise e

    # Bad request, e.g., incorrect config or input
    if e.code == '400':
        raise e

    # If harmful input or output detected, let it fail
    if e.code == 'DataInspectionFailed':
        raise e
    if 'inappropriate content' in str(e):
        raise e

    # Retry is meaningless if the input is too long
    if 'maximum context length' in str(e):
        raise e

    print_traceback(is_error=False)

    if num_retries >= max_retries:
        raise ModelServiceError(exception=Exception(f'Maximum number of retries ({max_retries}) exceeded.'))

    num_retries += 1
    jitter = 1.0 + random.random()
    delay = min(delay * exponential_base, max_delay) * jitter
    time.sleep(delay)
    return num_retries, delay
