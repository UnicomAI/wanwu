
import os
import cv2
import math
import numpy as np
from scipy.signal import argrelextrema

import logging

from utils.constant import MAX_SENTENCE_SIZE
from utils.prompts import IMAGE_TEXT_EXTRACT_PROMPT
from langchain.text_splitter import CharacterTextSplitter
from pathlib import Path
#from minio_utils import upload_local_file
from utils import minio_utils
from utils import multimodal_utils
from settings import NO_CONTENT
logger = logging.getLogger(__name__)

def information_entropy(img):
    """
    计算图像的信息熵

    Parameters
    ----------
    img : ndarray
        输入灰度值图像的像素矩阵

    Returns
    -------
    res : double
        图像信息熵
    """

    prob = np.zeros(256, )

    # 计算各灰度值下出现的概率
    for i in range(img.shape[0]):
        for j in range(img.shape[1]):
            ind = img[i][j]
            prob[ind] += 1
    prob = prob / (img.shape[0] * img.shape[1])

    # 计算信息熵
    res = 0
    for i in range(prob.shape[0]):
        if prob[i] != 0:
            res -= prob[i] * math.log2(prob[i])
    return res


def rel_change(a, b):
    if max(a, b) == 0:  # 防止除以0
        return 0
    return (b - a) / max(a, b)


class Frame:
    def __init__(self, id, diff):
        self.id = id
        self.diff = diff

    def __lt__(self, other):
        return self.id < other.id

    def __gt__(self, other):
        return other.__lt__(self)

    def __eq__(self, other):
        return self.id == other.id

    def __ne__(self, other):
        return not self.__eq__(other)



def smooth(data, window_len=13, window='hanning'):
    s = np.r_[2 * data[0] - data[window_len:1:-1], data, 2 * data[-1] - data[-1:-window_len:-1]]

    if window == 'flat':
        win = np.ones(window_len, 'd')
    elif window == 'hanning':
        win = getattr(np, window)(window_len)

    y = np.convolve(win / win.sum(), s, mode='same')
    return y[window_len - 1: -window_len + 1]


def get_histogram_similarity(img1, img2):
    """
    计算两幅图像的直方图相似度（越接近1越相似）
    """
    # 转为灰度图
    gray1 = cv2.cvtColor(img1, cv2.COLOR_BGR2GRAY)
    gray2 = cv2.cvtColor(img2, cv2.COLOR_BGR2GRAY)

    # 计算直方图
    hist1 = cv2.calcHist([gray1], [0], None, [256], [0, 256])
    hist2 = cv2.calcHist([gray2], [0], None, [256], [0, 256])

    # 归一化
    cv2.normalize(hist1, hist1, 0, 1, cv2.NORM_MINMAX)
    cv2.normalize(hist2, hist2, 0, 1, cv2.NORM_MINMAX)

    # 比较直方图（相关性）
    similarity = cv2.compareHist(hist1, hist2, cv2.HISTCMP_CORREL)
    return similarity

def diff_exaction(video_path, use_thresh=True, thresh=0.6, use_local_maximal=True, len_window=50, frame_interval=5,
                  similarity_threshold=0.95):
    """
    提取关键帧，并去重相似帧。

    Parameters
    ----------
    video_path : str
        视频路径
    similarity_threshold : float
        图像相似度阈值，超过该值视为重复帧（默认0.95）
    """

    cap = cv2.VideoCapture(video_path)
    ind = 0
    curr_frame, prev_frame = None, None
    frame_diffs = []
    frames = []
    success, frame = cap.read()

    last_keyframe = None  # 存储上一张关键帧图像用于比较

    while success:
        luv = cv2.cvtColor(frame, cv2.COLOR_BGR2LUV)
        curr_frame = luv

        if curr_frame is not None and prev_frame is not None:
            diff = cv2.absdiff(curr_frame, prev_frame)
            diff_sum = np.sum(diff)
            diff_sum_mean = diff_sum / (diff.shape[0] * diff.shape[1])
            frame_diffs.append(diff_sum_mean)
            frames.append(Frame(ind, diff_sum_mean))

        prev_frame = curr_frame
        ind += 1
        success, frame = cap.read()

    cap.release()

    keyframe_id_set = set()

    # 根据阈值筛选
    if use_thresh:
        for i in range(1, len(frames)):
            if rel_change(float(frames[i - 1].diff), float(frames[i].diff)) >= thresh:
                keyframe_id_set.add(frames[i].id)

    # 局部极值筛选
    if use_local_maximal:
        diff_array = np.array(frame_diffs)
        # 自适应窗口：短视频帧数 < len_window 时 smooth 会返回空数组（切片两端越过中点），
        # 导致局部极值筛选整路失效、抽不到任何关键帧。将窗口限制为不超过数据长度，
        # 并保证 hanning 窗至少为 3。长视频（帧数 >= len_window）行为完全不变。
        actual_window = max(3, min(len_window, len(diff_array)))
        if len(diff_array) >= actual_window:
            sm_diff_array = smooth(diff_array, actual_window)
            frame_indexes = np.asarray(argrelextrema(sm_diff_array, np.greater))[0]

            for i in frame_indexes:
                if i < len(frames):
                    keyframe_id_set.add(frames[i].id)

    # 第一次筛选：按 frame_interval 控制最小间隔
    reduced_keyframes = set()
    prev_keyframe_id = -frame_interval
    for keyframe_id in sorted(keyframe_id_set):
        if keyframe_id - prev_keyframe_id >= frame_interval:
            reduced_keyframes.add(keyframe_id)
            prev_keyframe_id = keyframe_id

    # 第二次筛选：去除视觉上相似的帧
    final_keyframes = set()
    cap = cv2.VideoCapture(video_path)
    success, frame = cap.read()
    idx = 0
    saved_frames = []

    for keyframe_id in sorted(reduced_keyframes):
        while success and idx <= keyframe_id:
            if idx == keyframe_id:
                if last_keyframe is None:
                    # 第一个关键帧直接保留
                    final_keyframes.add(keyframe_id)
                    last_keyframe = frame.copy()
                    saved_frames.append(frame)
                else:
                    # 比较相似度
                    similarity = get_histogram_similarity(last_keyframe, frame)
                    if similarity < similarity_threshold:
                        final_keyframes.add(keyframe_id)
                        last_keyframe = frame.copy()
                        saved_frames.append(frame)
                    # else: 相似度过高，跳过
                break
            idx += 1
            success, frame = cap.read()

    cap.release()
    return final_keyframes


def exact(video_path, parser_choices, multimodal_model_id):
    """
    按序列号抽取视频帧
    """
    directory = os.path.dirname(video_path)
    # 构建 keyframe 子目录路径
    keyframe_dir = os.path.join(directory, "keyframe")
    # 自动创建目录（如果不存在）
    os.makedirs(keyframe_dir, exist_ok=True)
    path_obj = Path(video_path)
    file_name = path_obj.stem
    cap = cv2.VideoCapture(video_path)
    total_frames = int(cap.get(cv2.CAP_PROP_FRAME_COUNT))
    cap.release()
    logger.info("total_frames=%s" % total_frames)
    if total_frames >= 1000:
        frame_interval = max(1, total_frames // 200)
    else:
        frame_interval = 5

    logger.info("---->frame_interval=%s" % frame_interval)
    # file_extension = path_obj.suffix
    page_chunks = []
    keyframe_id_set = diff_exaction(video_path, frame_interval=frame_interval)  # 控制提取的关键帧间隔
    cap = cv2.VideoCapture(video_path)
    success, frame = cap.read()
    idx = 0

    # signature = hashlib.md5(name).hexdigest()
    while (success):
        if idx in keyframe_id_set:
            save_name = f"{file_name}_{idx}.jpg"
            image_desc_text = ""
            image_minio_url = ""
            image_text = ""
            # file_path = "./keyframe/" + save_name
            image_filepath = os.path.join(keyframe_dir, save_name)
            # file_path = os.path.join(directory, f"/keyframe/{save_name}")
            cv2.imwrite(image_filepath, frame)

            keyframe_id_set.remove(idx)  # 只保留一个关键帧，不重复写入
            logger.info("========>keyframe_extract:image_filepath:%s" % image_filepath)
            try:
                minio_result = minio_utils.upload_local_file(image_filepath)
                if minio_result['code'] == 0:
                    image_minio_url = minio_result['download_link']
            except Exception as err:  # 更新状态失败
                import traceback
                logger.error('vedio exact upload minio error：' + repr(err))
                logger.error(traceback.format_exc())

            if "multimodal" in parser_choices and multimodal_model_id:
                image_desc_text = multimodal_utils.req_unicom_VL_plus(image_filepath, multimodal_model_id, IMAGE_TEXT_EXTRACT_PROMPT)
            if image_minio_url:
                if image_desc_text and NO_CONTENT not in image_desc_text:
                    image_text = f"![image]({image_minio_url}) 此画面的描述：{image_desc_text}"
                else:
                    image_text = f"![image]({image_minio_url})"
            elif image_desc_text and NO_CONTENT not in image_desc_text:
                image_text = f"画面的描述：{image_desc_text}"

            if image_text:
                page_chunk = {}
                page_chunk["text"] = image_text
                page_chunk["file_path"] = video_path
                page_chunk["type"] = "text"
                page_chunk["length"] = len(image_text)
                page_chunks.append(page_chunk)

            logger.info("========>image2:%s, text2:%s" % (save_name, image_text))

        idx = idx + 1
        success, frame = cap.read()
    cap.release()

    # logger.info("------>len=%s,text=%s" % (len(text), text))

    return page_chunks


def exact_text(video_path, parser_choices, multimodal_model_id):
    """
    按序列号抽取视频帧
    """
    directory = os.path.dirname(video_path)
    # 构建 keyframe 子目录路径
    keyframe_dir = os.path.join(directory, "keyframe")
    # 自动创建目录（如果不存在）
    os.makedirs(keyframe_dir, exist_ok=True)
    path_obj = Path(video_path)
    file_name = path_obj.stem
    cap = cv2.VideoCapture(video_path)
    total_frames = int(cap.get(cv2.CAP_PROP_FRAME_COUNT))
    cap.release()
    logger.info("total_frames=%s" % total_frames)
    if total_frames >= 1000:
        frame_interval = max(1, total_frames // 200)
    else:
        frame_interval = 5

    logger.info("---->frame_interval=%s" % frame_interval)
    keyframe_id_set = diff_exaction(video_path, frame_interval=frame_interval)  # 控制提取的关键帧间隔
    cap = cv2.VideoCapture(video_path)
    success, frame = cap.read()
    idx = 0
    text = ""
    # signature = hashlib.md5(name).hexdigest()
    while (success):
        if idx in keyframe_id_set:
            save_name = f"{file_name}_{idx}.jpg"
            image_desc_text = ""
            image_minio_url = ""
            image_text = ""
            image_filepath = os.path.join(keyframe_dir, save_name)
            cv2.imwrite(image_filepath, frame)

            keyframe_id_set.remove(idx)  # 只保留一个关键帧，不重复写入
            # image_text = ocr_utils.ocr_parser_text(file_path)
            try:
                minio_result = minio_utils.upload_local_file(image_filepath)
                if minio_result['code'] == 0:
                    image_minio_url = minio_result['download_link']

            except Exception as err:  # 更新状态失败
                import traceback
                logger.error('vedio exact_text upload minio error：' + repr(err))
                logger.error(traceback.format_exc())

            if "multimodal" in parser_choices and multimodal_model_id:
                image_desc_text = multimodal_utils.req_unicom_VL_plus(image_filepath, multimodal_model_id, IMAGE_TEXT_EXTRACT_PROMPT)
            if image_minio_url:
                if image_desc_text and NO_CONTENT not in image_desc_text:
                    image_text = f"![image]({image_minio_url}) 此画面的描述：{image_desc_text}"
                else:
                    image_text = f"![image]({image_minio_url})"
            elif image_desc_text and NO_CONTENT not in image_desc_text:
                image_text = f"画面的描述：{image_desc_text}"

            if image_text:
                text += image_text + "\n"

            logger.info("========>image:%s, text:%s" % (save_name, image_text))
            #time.sleep(1)
            # 【新增逻辑】：处理完后立即删除该图片文件
            # try:
            #     os.remove(file_path)
            #     # 可选：打印日志
            #     # logger.debug(f"已删除文件: {file_path}")
            # except Exception as e:
            #     logger.warning(f"无法删除文件 {file_path}: {e}")
        idx = idx + 1
        success, frame = cap.read()
    cap.release()

    return text


if __name__ == '__main__':
     text = exact("test.mp4")
