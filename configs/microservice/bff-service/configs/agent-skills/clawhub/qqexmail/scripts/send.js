import nodemailer from 'nodemailer';

const account = process.env.EXMAIL_ACCOUNT;
const authCode = process.env.EXMAIL_AUTH_CODE;

if (!account || !authCode) {
  console.error('请设置环境变量 EXMAIL_ACCOUNT 和 EXMAIL_AUTH_CODE');
  process.exit(1);
}

const transporter = nodemailer.createTransport({
  host: 'smtp.exmail.qq.com',
  port: 465,
  secure: true,
  auth: {
    user: account,
    pass: authCode,
  },
  tls: {
    rejectUnauthorized: false,
    servername: 'smtp.exmail.qq.com',
  },
});

async function readStdin() {
  const chunks = [];
  for await (const chunk of process.stdin) chunks.push(chunk);
  return Buffer.concat(chunks).toString('utf8').trim();
}

async function main() {
  const args = process.argv.slice(2);
  const useStdin = args.includes('--stdin');
  const filtered = args.filter((a) => a !== '--stdin');

  let to, subject, body;
  if (useStdin) {
    if (filtered.length < 2) {
      console.error('用法: node scripts/send.js <收件人> <主题> --stdin');
      process.exit(1);
    }
    [to, subject] = filtered;
    body = await readStdin();
  } else {
    if (filtered.length < 3) {
      console.error('用法: node scripts/send.js <收件人> <主题> <正文>');
      process.exit(1);
    }
    [to, subject, body] = filtered;
  }

  try {
    const info = await transporter.sendMail({
      from: account,
      to,
      subject,
      text: body,
    });
    console.log('已发送:', info.messageId);
  } catch (err) {
    console.error('发信失败:', err.message);
    process.exit(1);
  }
}

main();
