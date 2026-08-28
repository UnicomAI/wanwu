## Description: <br>
Search WeChat Official Account articles using OpenClaw's web search, Tavily API, and web fetch capabilities with compliance-focused design. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[jixsonwang](https://clawhub.ai/user/jixsonwang) <br>

### License/Terms of Use: <br>


## Use Case: <br>
Developers and agents use this skill to search public WeChat Official Account articles, return recent matching article metadata, and format results for downstream review or use. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: Search terms and result fetching may be sent to external search or fetch providers. <br>
Mitigation: Install only when external provider use is acceptable for the intended queries and data handling policy. <br>
Risk: The skill can use a configured Tavily API key and delegate work to local helper tooling. <br>
Mitigation: Verify the local Tavily helper skill and OpenClaw CLI are trusted before deployment, and scope API keys appropriately. <br>


## Reference(s): <br>
- [Wechat Search ClawHub Page](https://clawhub.ai/jixsonwang/wechat-search) <br>
- [README](artifact/README.md) <br>


## Skill Output: <br>
**Output Type(s):** [text, markdown, JSON, shell commands, configuration, guidance] <br>
**Output Format:** [Text, JSON, or Markdown search results with article titles, snippets, source names, and URLs.] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [Defaults to five results, supports configurable result limits, date filters, output format selection, and search strategy selection.] <br>

## Skill Version(s): <br>
1.0.3 (source: server release metadata and _meta.json) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
