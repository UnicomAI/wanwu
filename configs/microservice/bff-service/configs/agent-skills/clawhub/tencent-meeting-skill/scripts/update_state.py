"""
用户本地版本更新偏好状态维护
状态文件路径：~/.tencent-meeting-mcp/update-state.json
"""

import json
import os
import sys
import time
from pathlib import Path
from typing import Any, Dict, Optional

# 状态文件根目录
STATE_DIR = Path.home() / ".tencent-meeting-mcp"
STATE_FILE = STATE_DIR / "update-state.json"

# snooze 各 level 对应的静默时间（秒）
SNOOZE_LEVEL_DURATIONS = {
    1: 24 * 60 * 60,         # 24 小时
    2: 48 * 60 * 60,         # 48 小时
    3: 7 * 24 * 60 * 60,     # 7 天
}


def _default_state() -> Dict[str, Any]:
    """返回默认的状态结构"""
    return {
        "optionalUpdateCheck": True,
        "autoUpgrade": False,
        "snooze": None,
        "lastCheck": None,
        "justUpgradedFrom": None,
    }


def _ensure_dir() -> None:
    """确保状态目录存在"""
    try:
        STATE_DIR.mkdir(parents=True, exist_ok=True)
    except OSError as e:
        # 目录创建失败不阻塞主流程，仅在 stderr 打印警告，避免污染 stdout
        print(f"[warn] update-state mkdir failed: {e}", file=sys.stderr)


def load_update_state() -> Dict[str, Any]:
    """
    加载用户更新偏好状态。

    - 文件不存在时返回默认状态
    - JSON 损坏时返回默认状态，不抛异常
    """
    if not STATE_FILE.exists():
        return _default_state()

    try:
        with open(STATE_FILE, "r", encoding="utf-8") as f:
            data = json.load(f)
        if not isinstance(data, dict):
            return _default_state()
        # 合并默认值，保证字段完整
        merged = _default_state()
        merged.update(data)
        return merged
    except (json.JSONDecodeError, OSError):
        # JSON 损坏或读取失败时返回默认值，避免影响主流程
        return _default_state()


def save_update_state(state: Dict[str, Any]) -> None:
    """
    持久化用户更新偏好状态到磁盘。

    写文件失败时仅在 stderr 打印警告，不报错也不污染 stdout。
    """
    _ensure_dir()
    try:
        with open(STATE_FILE, "w", encoding="utf-8") as f:
            json.dump(state, f, ensure_ascii=False, indent=2)
    except OSError as e:
        print(f"[warn] update-state write failed: {e}", file=sys.stderr)


def is_optional_update_enabled(state: Dict[str, Any]) -> bool:
    """是否启用可选更新提醒"""
    return bool(state.get("optionalUpdateCheck", True))


def is_auto_upgrade_enabled(state: Dict[str, Any]) -> bool:
    """是否开启默认自动更新"""
    return bool(state.get("autoUpgrade", False))


def is_snoozed(
    state: Dict[str, Any],
    latest_version: str,
    now: Optional[int] = None,
) -> bool:
    """
    判断指定 latest_version 当前是否处于 snooze 静默期内。

    规则：
    - state.snooze 为空：未 snooze
    - snooze.version != latest_version：远端已有新版本，旧 snooze 失效
    - snooze.version == latest_version：按 timestamp + level 对应静默时长判断
    """
    snooze = state.get("snooze")
    if not snooze or not isinstance(snooze, dict):
        return False

    if snooze.get("version") != latest_version:
        return False

    timestamp = snooze.get("timestamp")
    level = snooze.get("level", 1)
    if not isinstance(timestamp, (int, float)):
        return False

    duration = SNOOZE_LEVEL_DURATIONS.get(int(level), SNOOZE_LEVEL_DURATIONS[3])
    now_ts = now if now is not None else int(time.time())
    return now_ts < int(timestamp) + duration

def compute_prompt_decision(
    state: Dict[str, Any],
    latest_version: str,
    now: Optional[int] = None,
) -> Dict[str, Any]:
    """
    根据当前本地偏好计算"是否需要向用户弹出版本更新提示"的决策。

    这是给本地伪工具 get_skill_update_preference 使用的唯一权威决策入口，
    模型不应自己解析 update-state.json 做时间戳运算，必须以本函数返回为准。

    返回结构：
        {
            "shouldPrompt": bool,
            "reason": str,                # disabled / auto_upgrade / snoozed
                                          # / version_changed / expired / no_snooze
                                          # / no_latest_version
            "snooze": {                   # 可选，仅在存在 snooze 时返回
                "version": str,
                "level": int,
                "timestamp": int,
                "deadline": int,          # timestamp + duration
                "remainingSeconds": int,  # deadline - now，负数代表已过期
                "active": bool,           # 即 is_snoozed 的结果（仅在版本一致时为 True）
            },
        }

    判定优先级与 references/version_management.md 保持一致：
      1) optionalUpdateCheck == False     -> shouldPrompt=False, reason=disabled
      2) autoUpgrade == True              -> shouldPrompt=False, reason=auto_upgrade
      3) snooze 存在且版本一致且未过期    -> shouldPrompt=False, reason=snoozed
      4) snooze 存在但版本不一致          -> shouldPrompt=True,  reason=version_changed
      5) snooze 存在但已过期              -> shouldPrompt=True,  reason=expired
      6) snooze 不存在                    -> shouldPrompt=True,  reason=no_snooze
      7) latest_version 为空              -> shouldPrompt=False, reason=no_latest_version
         （安全兜底：缺少版本号时不提示，避免误弹）
    """
    now_ts = now if now is not None else int(time.time())
    decision: Dict[str, Any] = {"shouldPrompt": False, "reason": ""}

    # 1) 用户已关闭可选更新提醒
    if not is_optional_update_enabled(state):
        decision["reason"] = "disabled"
        return decision

    # 2) 用户已开启自动更新（由调用方决定是否静默升级，本工具仅返回决策）
    if is_auto_upgrade_enabled(state):
        decision["reason"] = "auto_upgrade"
        return decision

    # 缺少 latest_version：安全兜底，不弹提示
    if not isinstance(latest_version, str) or not latest_version:
        decision["reason"] = "no_latest_version"
        return decision

    snooze = state.get("snooze")
    if not isinstance(snooze, dict) or not snooze:
        # 6) 无 snooze 记录
        decision["shouldPrompt"] = True
        decision["reason"] = "no_snooze"
        return decision

    snooze_version = snooze.get("version")
    timestamp = snooze.get("timestamp")
    level = snooze.get("level", 1)
    if not isinstance(timestamp, (int, float)):
        # snooze 数据异常，按无 snooze 处理
        decision["shouldPrompt"] = True
        decision["reason"] = "no_snooze"
        return decision

    duration = SNOOZE_LEVEL_DURATIONS.get(int(level), SNOOZE_LEVEL_DURATIONS[3])
    deadline = int(timestamp) + duration
    remaining = deadline - now_ts
    active = (snooze_version == latest_version) and (now_ts < deadline)

    decision["snooze"] = {
        "version": snooze_version,
        "level": int(level),
        "timestamp": int(timestamp),
        "deadline": deadline,
        "remainingSeconds": remaining,
        "active": active,
    }

    if snooze_version != latest_version:
        # 4) 服务端已发新版本，旧 snooze 失效
        decision["shouldPrompt"] = True
        decision["reason"] = "version_changed"
        return decision

    if active:
        # 3) 仍处于静默期内
        decision["shouldPrompt"] = False
        decision["reason"] = "snoozed"
        return decision

    # 5) 同版本但已过期
    decision["shouldPrompt"] = True
    decision["reason"] = "expired"
    return decision

# ---------------------------------------------------------------------------
# 用户偏好「三态互斥」约束
# ---------------------------------------------------------------------------
# 用户的版本更新偏好任意时刻仅能处于以下其中一个状态：
#   A. 默认询问态：optionalUpdateCheck=True, autoUpgrade=False, snooze=None
#   B. 暂不更新  ：optionalUpdateCheck=True, autoUpgrade=False, snooze={...}
#   C. 自动更新  ：optionalUpdateCheck=True, autoUpgrade=True,  snooze=None
#   D. 永不更新  ：optionalUpdateCheck=False, autoUpgrade=False, snooze=None
#
# 所有写入函数必须保证：把当前状态切换到目标状态时，同步清空其它互斥字段，
# 避免出现「snooze + autoUpgrade」「snooze + 永不更新」等不可达组合。
# ---------------------------------------------------------------------------


def save_snooze(latest_version: str) -> Dict[str, Any]:
    """
    切换到 B 态：针对指定 latest_version 写入暂不更新状态。

    - 同一版本再次 snooze 时 level 递增（最高 3，对应最长静默时长）
    - 切换到新版本时 level 从 1 重新开始
    - 同时清空 autoUpgrade 并恢复 optionalUpdateCheck=True，
      保证用户最近一次「暂不更新」的意图压过此前的「自动更新」/「永不更新」
    """
    state = load_update_state()
    snooze = state.get("snooze") or {}

    if snooze.get("version") == latest_version:
        level = int(snooze.get("level", 1)) + 1
        if level > 3:
            level = 3
    else:
        level = 1

    # 三态互斥：进入 B 态前清掉 C/D 态字段
    state["optionalUpdateCheck"] = True
    state["autoUpgrade"] = False
    state["snooze"] = {
        "version": latest_version,
        "level": level,
        "timestamp": int(time.time()),
    }
    save_update_state(state)
    return state


def enable_auto_upgrade() -> Dict[str, Any]:
    """
    切换到 C 态：开启默认自动更新。

    三态互斥：清空 snooze 并恢复 optionalUpdateCheck=True，
    保证用户最近一次「自动更新」的意图压过此前的「暂不更新」/「永不更新」。
    """
    state = load_update_state()
    state["optionalUpdateCheck"] = True
    state["autoUpgrade"] = True
    state["snooze"] = None
    save_update_state(state)
    return state


def disable_optional_update_check() -> Dict[str, Any]:
    """
    切换到 D 态：关闭可选更新提醒（永不更新）。

    三态互斥：同时清空 autoUpgrade 与 snooze，
    保证用户最近一次「永不更新」的意图压过此前的「自动更新」/「暂不更新」。
    """
    state = load_update_state()
    state["optionalUpdateCheck"] = False
    state["autoUpgrade"] = False
    state["snooze"] = None
    save_update_state(state)
    return state


def enable_optional_update_check() -> Dict[str, Any]:
    """
    切换到 A 态：恢复默认询问态。

    用户语义为「回到默认行为」，因此除了打开 optionalUpdateCheck 外，
    一并清空 autoUpgrade 与 snooze，避免残留旧偏好导致下次提示行为不可预期。
    """
    state = load_update_state()
    state["optionalUpdateCheck"] = True
    state["autoUpgrade"] = False
    state["snooze"] = None
    save_update_state(state)
    return state


def save_last_check(check_result: Dict[str, Any]) -> None:
    """记录最近一次版本检查结果（用于排查/节流）"""
    state = load_update_state()
    state["lastCheck"] = {
        "timestamp": int(time.time()),
        "result": check_result,
    }
    save_update_state(state)


def mark_just_upgraded(old_version: str) -> None:
    """标记升级前版本，便于升级后展示提示"""
    state = load_update_state()
    state["justUpgradedFrom"] = old_version
    save_update_state(state)


def clear_after_upgrade() -> None:
    """升级完成后清理 snooze 与 justUpgradedFrom"""
    state = load_update_state()
    state["snooze"] = None
    state["justUpgradedFrom"] = None
    save_update_state(state)
