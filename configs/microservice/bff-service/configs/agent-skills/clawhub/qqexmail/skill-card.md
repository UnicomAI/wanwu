## Description: <br>
通过 IMAP/SMTP 收发腾讯企业邮箱（exmail.qq.com）邮件，支持发送邮件、收取邮件列表和按 UID 获取邮件正文，凭证从环境变量读取。 <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[hunger09](https://clawhub.ai/user/hunger09) <br>

### License/Terms of Use: <br>
MIT-0 <br>


## Use Case: <br>
Developers and agent users use this skill to send Tencent Enterprise Mail messages, list recent inbox messages, and retrieve full message bodies when working with exmail.qq.com accounts. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: The SMTP sender disables TLS certificate verification while using mailbox credentials. <br>
Mitigation: Review and fix TLS verification before using the send script; avoid sending mail until certificate validation is enforced. <br>
Risk: EXMAIL_AUTH_CODE functions like an app password for the mailbox. <br>
Mitigation: Provide it only through environment variables or a local secret store, never commit it, and rotate it if exposed. <br>
Risk: Mailbox subjects, senders, summaries, and full message bodies can be printed to stdout. <br>
Mitigation: Avoid terminal logs, transcripts, redirected files, or pipes that may persist sensitive email content. <br>
Risk: The send script can transmit content to unintended recipients if arguments are wrong. <br>
Mitigation: Confirm every recipient, subject, and message body before execution. <br>


## Reference(s): <br>
- [ClawHub skill page](https://clawhub.ai/hunger09/qqexmail) <br>
- [ClawHub publisher profile](https://clawhub.ai/user/hunger09) <br>


## Skill Output: <br>
**Output Type(s):** [Shell commands, Configuration instructions, Text, Markdown, Guidance] <br>
**Output Format:** [Markdown guidance with inline shell commands and plain-text command output] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [May print email metadata, message summaries, or full message bodies to stdout depending on the script used.] <br>

## Skill Version(s): <br>
1.0.0 (source: server release evidence and package.json) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
