## Description: <br>
Wechat Publisher helps an agent publish Markdown articles to WeChat Official Account drafts through wenyan-cli, with theme, code highlighting, and image upload support. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[0731coderlee-sudo](https://clawhub.ai/user/0731coderlee-sudo) <br>

### License/Terms of Use: <br>


## Use Case: <br>
Developers and content operators use this skill to prepare Markdown with required WeChat frontmatter, choose presentation themes, load WeChat credentials, and publish the content as a draft for review in the WeChat Official Account backend. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: WeChat credentials may be exposed if WECHAT_APP_SECRET is stored in git, chat logs, or shared plaintext files. <br>
Mitigation: Use session environment variables or a secret store where possible, and avoid committing or sharing credential files. <br>
Risk: The selected Markdown file, frontmatter, cover image, and embedded local or remote images are uploaded to WeChat when publishing. <br>
Mitigation: Review the article, cover path, embedded image paths, and image URLs before running the publish workflow. <br>
Risk: Publishing depends on the third-party wenyan-cli package. <br>
Mitigation: Install only when the dependency is trusted for the target environment. <br>


## Reference(s): <br>
- [ClawHub skill page](https://clawhub.ai/0731coderlee-sudo/wechat-publisher) <br>
- [wenyan-cli](https://github.com/caol64/wenyan-cli) <br>
- [wenyan documentation](https://wenyan.yuzhi.tech) <br>
- [wenyan upload configuration](https://yuzhi.tech/docs/wenyan/upload) <br>
- [WeChat Official Account API documentation](https://developers.weixin.qq.com/doc/offiaccount/) <br>
- [Themes reference](references/themes.md) <br>
- [Troubleshooting guide](references/troubleshooting.md) <br>


## Skill Output: <br>
**Output Type(s):** [Text, Markdown, Shell commands, Configuration, Guidance] <br>
**Output Format:** [Markdown guidance with inline bash commands and configuration examples] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [May invoke wenyan-cli to upload the selected Markdown file, frontmatter cover image, and embedded local or remote images to a WeChat Official Account draft.] <br>

## Skill Version(s): <br>
0.1.0 (source: server-resolved release metadata) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
