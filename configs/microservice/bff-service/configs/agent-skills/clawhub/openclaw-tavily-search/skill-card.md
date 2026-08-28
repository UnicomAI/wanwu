## Description: <br>
Web search via Tavily API for looking up sources, finding links, and returning a small set of relevant results with titles, URLs, snippets, and optional short answers. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[Jacky1n7](https://clawhub.ai/user/Jacky1n7) <br>

### License/Terms of Use: <br>


## Use Case: <br>
Developers and agent users use this skill when they need Tavily-backed web search from an OpenClaw workspace, especially when Brave web search is unavailable or undesired. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: Search queries and the Tavily API key are sent to Tavily. <br>
Mitigation: Use a dedicated Tavily API key and avoid including secrets, private customer data, unreleased project details, or other sensitive information in search queries. <br>


## Reference(s): <br>
- [ClawHub skill page](https://clawhub.ai/Jacky1n7/openclaw-tavily-search) <br>
- [Tavily Search API endpoint](https://api.tavily.com/search) <br>


## Skill Output: <br>
**Output Type(s):** [Shell commands, JSON, Markdown, Guidance] <br>
**Output Format:** [Command output as raw JSON, brave-like JSON, or compact Markdown] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [Results include query, titles, URLs, snippets or content, and optional short answer summaries; max results are capped at 10 by the script.] <br>

## Skill Version(s): <br>
0.1.0 (source: server release evidence) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
