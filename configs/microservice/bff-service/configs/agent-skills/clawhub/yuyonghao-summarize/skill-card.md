## Description: <br>
Summarize URLs or files with the summarize CLI (web, PDFs, images, audio, YouTube). <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[yuyonghao-123](https://clawhub.ai/user/yuyonghao-123) <br>

### License/Terms of Use: <br>
MIT-0 <br>


## Use Case: <br>
Developers and external users use this skill to ask an agent to run the summarize CLI against URLs, local files, PDFs, images, audio, or YouTube links and return concise summaries. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: Summarizing local files, private URLs, regulated data, or sensitive media may send content to the selected model provider or optional extraction services. <br>
Mitigation: Use only approved providers for sensitive content, avoid confidential inputs unless permitted, and disable Firecrawl or Apify fallbacks when those services should not receive content. <br>


## Reference(s): <br>
- [Summarize homepage](https://summarize.sh) <br>
- [ClawHub skill listing](https://clawhub.ai/yuyonghao-123/yuyonghao-summarize) <br>


## Skill Output: <br>
**Output Type(s):** [Text, Markdown, JSON, Shell commands, Configuration guidance] <br>
**Output Format:** [Markdown guidance with shell commands; CLI output can be plain text or JSON when --json is used.] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [Supports length controls, output token limits, model selection, API-key configuration, and optional Firecrawl or Apify extraction fallbacks.] <br>

## Skill Version(s): <br>
0.1.0 (source: server release metadata and package.json) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
