#!/usr/bin/env node
/**
 * 腾讯文档 MCP 幻灯片创建辅助脚本（跨平台）
 *
 * 功能：
 *   完成幻灯片创建的完整流程：
 *   1. 调用 create_slide 提交生成/编辑任务
 *   2. 自动轮询 slide_progress 查询进度
 *   3. 输出最终结果（file_url 或错误信息）
 *
 * 用法：
 *   node generate_slide.js --description "用户描述" [--reference_context "参考材料"] [--session_id "已有会话ID"]
 *
 * 参数说明：
 *   --description       (必填) 用户对PPT的主题和要求描述
 *   --reference_context  (可选) 生成PPT的参考资料
 *   --session_id         (可选) 多轮编辑时传入之前返回的session_id
 *
 * 依赖：
 *   - Node.js (>= 14)
 *   - mcporter（已配置 tencent-docs 服务）
 *
 * 输出（成功时）：
 *   SLIDE_COMPLETED
 *   SESSION_ID:<session_id>
 *   FILE_URL:<file_url>
 *
 * 输出（失败时）：
 *   SLIDE_FAILED
 *   ERROR:<error_message>
 */

"use strict";

const { execSync } = require("child_process");

// ── 常量 ──────────────────────────────────────────────────────────────────
const POLL_INTERVAL_MS = 20 * 1000; // 轮询间隔 20 秒
const MAX_POLL_DURATION_MS = 20 * 60 * 1000; // 单次调用最长轮询等待 20 分钟（仅限本次脚本执行，不影响 session_id 生命周期）
const MCP_SERVICE = "tencent-docs";

// ── 参数解析 ──────────────────────────────────────────────────────────────
function parseArgs() {
  const args = process.argv.slice(2);
  const params = {};
  for (let i = 0; i < args.length; i++) {
    if (args[i] === "--description" && i + 1 < args.length) {
      params.description = args[++i];
    } else if (args[i] === "--reference_context" && i + 1 < args.length) {
      params.reference_context = args[++i];
    } else if (args[i] === "--session_id" && i + 1 < args.length) {
      params.session_id = args[++i];
    }
  }
  return params;
}

// ── mcporter 调用封装 ────────────────────────────────────────────────────
function mcpCall(tool, argsObj) {
  const argsJson = JSON.stringify(argsObj);
  const cmd = `mcporter call "${MCP_SERVICE}" "${tool}" --args '${argsJson.replace(/'/g, "'\\''")}'`;
  try {
    const stdout = execSync(cmd, { encoding: "utf-8", timeout: 60000 });
    return JSON.parse(stdout.trim());
  } catch (err) {
    const msg = err.stderr || err.stdout || err.message || "unknown error";
    throw new Error(`mcporter call ${tool} failed: ${msg}`);
  }
}

// ── 等待指定毫秒 ────────────────────────────────────────────────────────
function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// ── 主流程 ────────────────────────────────────────────────────────────────
async function main() {
  const params = parseArgs();

  // 参数校验
  if (!params.description) {
    console.log("SLIDE_FAILED");
    console.log("ERROR:missing_argument - 必须提供 --description 参数");
    process.exit(1);
  }

  // ── Step 1: 调用 create_slide ──────────────────────────────────────────
  const createArgs = { description: params.description };
  if (params.reference_context) {
    createArgs.reference_context = params.reference_context;
  }
  if (params.session_id) {
    createArgs.session_id = params.session_id;
  }

  const mode = params.session_id ? "编辑" : "创建";
  console.log(`⏳ 正在${mode}幻灯片...`);

  let createResult;
  try {
    createResult = mcpCall("create_slide", createArgs);
  } catch (err) {
    console.log("SLIDE_FAILED");
    console.log(`ERROR:create_slide_failed - ${err.message}`);
    process.exit(1);
  }

  const sessionId = createResult.session_id;
  if (!sessionId) {
    console.log("SLIDE_FAILED");
    console.log(
      `ERROR:no_session_id - 未获取到 session_id，create_slide 返回: ${JSON.stringify(createResult)}`
    );
    process.exit(1);
  }

  const traceId = createResult.trace_id || "";
  console.log(`✅ 任务已提交，session_id: ${sessionId}`);
  if (traceId) {
    console.log(`🔗 trace_id: ${traceId}`);
  }
  console.log("");

  // ── Step 2: 轮询 slide_progress ───────────────────────────────────────
  const startTime = Date.now();
  let pollCount = 0;

  while (Date.now() - startTime < MAX_POLL_DURATION_MS) {
    await sleep(POLL_INTERVAL_MS);
    pollCount++;

    console.log(
      `⏳ 正在生成中，第 ${pollCount} 次轮询，已等待 ${Math.round((Date.now() - startTime) / 1000)}s ...`
    );

    let progressResult;
    try {
      progressResult = mcpCall("slide_progress", { session_id: sessionId });
    } catch (err) {
      console.log(
        `⚠️ 第 ${pollCount} 次轮询异常: ${err.message}，将继续重试...`
      );
      continue;
    }

    const status = progressResult.status;

    switch (status) {
      case "completed": {
        const fileUrl = progressResult.file_url || "";
        console.log("");
        console.log("SLIDE_COMPLETED");
        console.log(`SESSION_ID:${sessionId}`);
        console.log(`FILE_URL:${fileUrl}`);
        console.log("");
        console.log(`✅ 幻灯片生成完成！`);
        console.log(`📎 链接: ${fileUrl}`);
        if (params.session_id) {
          console.log(`💡 此为多轮编辑结果，session_id 保持不变: ${sessionId}`);
        } else {
          console.log(
            `💡 如需后续编辑此PPT，请在下次调用时传入 --session_id ${sessionId}`
          );
        }
        process.exit(0);
        break;
      }
      case "failed":
        console.log("");
        console.log("SLIDE_FAILED");
        console.log(
          `ERROR:generation_failed - 幻灯片生成失败: ${progressResult.error || "未知错误"}`
        );
        process.exit(1);
        break;
      case "not_found":
        console.log("");
        console.log("SLIDE_FAILED");
        console.log(
            "ERROR:session_not_found - session_id 不正确"
        );
        process.exit(1);
        break;
      case "400008":
        console.log("");
        console.log("DO_NOT_RETRY");
        console.log(
            "ERROR:400008 - 积分已消耗完毕，超级会员专享2000积分/月，立即购买：https://docs.qq.com/vip/asset-center?tab=ai&fromPage=offsite&part_aid=offsite_claw"
        );
        process.exit(1);
        break;
      case "in_progress":
        // 继续轮询
        break;
      default:
        console.log(`⚠️ 未知状态: ${status}，继续轮询...`);
        break;
    }
  }

  // 超时
  console.log("");
  console.log("SLIDE_FAILED");
  console.log(
    `ERROR:timeout - 本次轮询超时（已等待 ${MAX_POLL_DURATION_MS / 60000} 分钟），session_id 仍然有效，可重新执行脚本继续轮询`
  );
  console.log(`SESSION_ID:${sessionId}`);
  process.exit(1);
}

main().catch((err) => {
  console.log("SLIDE_FAILED");
  console.log(`ERROR:unexpected - ${err.message}`);
  process.exit(1);
});
