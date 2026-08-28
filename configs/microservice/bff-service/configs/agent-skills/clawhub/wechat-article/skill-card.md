## Description: <br>
Fetches WeChat public account articles, extracts the title and body, and outputs the result as Markdown. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[haweiYu](https://clawhub.ai/user/haweiYu) <br>

### License/Terms of Use: <br>


## Use Case: <br>
Developers and content operators use this skill to turn public WeChat article links into clean Markdown for analysis, archiving, or reuse in agent workflows. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: The default extraction method sends the supplied WeChat article URL to r.jina.ai. <br>
Mitigation: Use only public or non-sensitive article URLs, and avoid confidential, access-controlled, tokenized, or internal links. <br>
Risk: Some WeChat articles may be blocked by anti-crawling controls or captcha challenges. <br>
Mitigation: Expect extraction failures for protected articles and review errors before relying on the Markdown output. <br>


## Reference(s): <br>
- [ClawHub skill page](https://clawhub.ai/haweiYu/wechat-article) <br>
- [Publisher profile](https://clawhub.ai/user/haweiYu) <br>
- [r.jina.ai extraction service](https://r.jina.ai/) <br>


## Skill Output: <br>
**Output Type(s):** [text, markdown, shell commands, guidance] <br>
**Output Format:** [Markdown document written to standard output, with status and error messages on standard error.] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [Includes article title, original link, optional publication metadata when available, and body content.] <br>

## Skill Version(s): <br>
1.0.0 (source: server release metadata and artifact _meta.json) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
