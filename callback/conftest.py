"""pytest 配置：确保工程根目录在 sys.path 上，使 ``utils``/``configs``/``callback``
等顶层导入在测试中可用（与 gunicorn 以工程根为 cwd 运行的行为一致）。
"""
import os
import sys

_ROOT = os.path.dirname(os.path.abspath(__file__))
if _ROOT not in sys.path:
    sys.path.insert(0, _ROOT)

# 加载配置，使 config.callback_cfg 可用（build_prompt 等模块在 import 时即读取配置）
from configs.config import load_config  # noqa: E402

load_config()
