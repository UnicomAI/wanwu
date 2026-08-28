import Imap from 'imap';
import { simpleParser } from 'mailparser';

const account = process.env.EXMAIL_ACCOUNT;
const authCode = process.env.EXMAIL_AUTH_CODE;

if (!account || !authCode) {
  console.error('请设置环境变量 EXMAIL_ACCOUNT 和 EXMAIL_AUTH_CODE');
  process.exit(1);
}

function parseArgs() {
  const args = process.argv.slice(2);
  let uid = null;
  for (let i = 0; i < args.length; i++) {
    if (args[i] === '--uid' && args[i + 1]) {
      uid = parseInt(args[i + 1], 10);
      i++;
    }
  }
  return { uid };
}

function stripHtml(html) {
  if (!html) return '';
  return html
    .replace(/<style[^>]*>[\s\S]*?<\/style>/gi, '')
    .replace(/<script[^>]*>[\s\S]*?<\/script>/gi, '')
    .replace(/<[^>]+>/g, ' ')
    .replace(/\s+/g, ' ')
    .trim();
}

async function fetchBody(uid) {
  const imap = new Imap({
    user: account,
    password: authCode,
    host: 'imap.exmail.qq.com',
    port: 993,
    tls: true,
    connTimeout: 30000,
    authTimeout: 30000,
  });

  return new Promise((resolve, reject) => {
    let bodyText = '';
    let messageReceived = false;

    imap.once('ready', () => {
      imap.openBox('INBOX', false, (err) => {
        if (err) {
          imap.end();
          return reject(err);
        }

        const fetch = imap.fetch([uid], { bodies: '' });

        fetch.on('message', (msg) => {
          messageReceived = true;
          msg.on('body', (stream) => {
            simpleParser(stream, (err, parsed) => {
              if (err) return reject(err);
              bodyText = parsed.text || stripHtml(parsed.html) || '(无正文)';
            });
          });
        });

        fetch.once('error', (e) => {
          imap.end();
          reject(e);
        });

        fetch.once('end', () => {
          // 等待解析完成
          setTimeout(() => {
            if (!messageReceived) {
              bodyText = '(未找到该 UID 的邮件)';
            }
            imap.end();
          }, 500);
        });
      });
    });

    imap.once('error', reject);
    imap.once('end', () => resolve(bodyText));
    imap.connect();
  });
}

async function main() {
  const { uid } = parseArgs();

  if (!uid || isNaN(uid)) {
    console.error('用法: node scripts/get-body.js --uid <邮件UID>');
    console.error('提示: UID 可从 receive.js 的输出中获取');
    process.exit(1);
  }

  try {
    const body = await fetchBody(uid);
    console.log(body);
  } catch (err) {
    console.error('获取邮件正文失败:', err.message);
    process.exit(1);
  }
}

main();
