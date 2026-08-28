## Description: <br>
Fetches webpage content as text or Markdown through named third-party relay services when normal fetching is blocked or filtered. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[MRTommyWU](https://clawhub.ai/user/MRTommyWU) <br>

### License/Terms of Use: <br>


## Use Case: <br>
Developers and agents use this skill to retrieve webpage content in Markdown-like text when ordinary web fetching is blocked, filtered, or affected by Cloudflare-style protections. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: Requested URLs and returned page content are processed by third-party relay services. <br>
Mitigation: Use this skill only for public or non-sensitive webpages where sharing the URL and retrieved content with the named relay service is acceptable. <br>
Risk: Private dashboards, authenticated pages, signed links, or token-bearing URLs could expose secrets to external services. <br>
Mitigation: Do not use this skill with intranet resources, authenticated pages, password-reset links, signed URLs, or any URL containing credentials, tokens, or other secrets. <br>


## Reference(s): <br>
- [ClawHub skill page](https://clawhub.ai/MRTommyWU/web-content-fetcher) <br>
- [Jina AI Reader endpoint](https://r.jina.ai/) <br>
- [markdown.new endpoint](https://markdown.new/) <br>
- [defuddle.md endpoint](https://defuddle.md/) <br>


## Skill Output: <br>
**Output Type(s):** [text, markdown, shell commands, code, guidance] <br>
**Output Format:** [Markdown or plain text returned from webpage relay services, with optional shell or Python command usage guidance.] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [Returned content depends on the selected relay service and the accessibility of the target URL.] <br>

## Skill Version(s): <br>
1.0.1 (source: server release evidence) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
