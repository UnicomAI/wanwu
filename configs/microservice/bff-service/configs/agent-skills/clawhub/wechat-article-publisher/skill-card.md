## Description: <br>
Extracts articles from Markdown files or webpages, formats them with WeChat-friendly templates, and creates or publishes WeChat Official Account drafts. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[zhenglong2015](https://clawhub.ai/user/zhenglong2015) <br>

### License/Terms of Use: <br>
MIT-0 <br>


## Use Case: <br>
Developers, content operators, and publishing teams use this skill to turn local Markdown or webpage content into WeChat Official Account drafts, preview rendered HTML, and optionally submit articles for publication. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: The skill can create drafts and optionally publish to a live WeChat Official Account. <br>
Mitigation: Use --dry-run first, manually review the rendered preview and target account, and use --publish only after approval. <br>
Risk: WeChat app credentials are required in config.json and could be exposed if handled carelessly. <br>
Mitigation: Keep config.json private and out of version control, and restrict access to the account credentials. <br>
Risk: Remote webpages or Markdown with remote images may introduce untrusted content into the publishing workflow. <br>
Mitigation: Prefer reviewed local Markdown and a reviewed local --cover-image; avoid untrusted URLs and remote images. <br>


## Reference(s): <br>
- [ClawHub skill page](https://clawhub.ai/zhenglong2015/wechat-article-publisher) <br>


## Skill Output: <br>
**Output Type(s):** [JSON, Markdown, Shell commands, Configuration] <br>
**Output Format:** [JSON status output and preview HTML generated through command-line workflows] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [May create WeChat drafts, submit publication requests, return draft_media_id, publish_id, status, and dry-run preview_html.] <br>

## Skill Version(s): <br>
1.0.0 (source: ClawHub release evidence) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
