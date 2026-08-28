#!/usr/bin/env node
/*
 * 跨平台脚本：将 public_key.pem 转换为 public-key.js
 * 输出格式: window.__VEGA_PUBLIC_KEY__ = "<PEM 内容>";
 *
 * 公钥来源优先级：先读权威源 data-connection/public_key.pem，
 * 为空再回退 vega-server 根目录的 public_key.pem。
 * public-key.js 是派生产物，每次都覆盖写出（内容由公钥确定，可重建）。
 *
 * 用法:
 *   node generate-public-key-js.js
 */
const fs = require('fs');
const path = require('path');

const baseDir = __dirname;
const canonicalPemPath = path.join(baseDir, 'data-connection', 'public_key.pem');
const pemPath = path.join(baseDir, 'public_key.pem');
const outPath = path.join(baseDir, 'public-key.js');

function isNonEmpty(p) {
  try { return fs.statSync(p).size > 0; } catch { return false; }
}

// Resolve pem source: canonical store first, then the vega-server copy.
let sourcePem = null;
if (isNonEmpty(canonicalPemPath)) {
  sourcePem = canonicalPemPath;
} else if (isNonEmpty(pemPath)) {
  sourcePem = pemPath;
} else {
  console.error(`Public key not found or empty: ${canonicalPemPath}`);
  console.error('Run ./generate-keys.sh first.');
  process.exit(1);
}

const pem = fs.readFileSync(sourcePem, 'utf8').replace(/\r\n/g, '\n').trim();
const content = `window.__VEGA_PUBLIC_KEY__ = ${JSON.stringify(pem)};\n`;

fs.writeFileSync(outPath, content, { encoding: 'utf8' });

console.log(`Wrote ${outPath} (from ${sourcePem})`);
