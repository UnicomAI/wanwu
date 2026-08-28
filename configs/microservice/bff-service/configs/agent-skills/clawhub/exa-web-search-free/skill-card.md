## Description: <br>
Provides no-key Exa MCP access for web search, code context lookup, and company research. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[Whiteknight07](https://clawhub.ai/user/Whiteknight07) <br>

### License/Terms of Use: <br>


## Use Case: <br>
Developers, researchers, and agents use this skill to configure and call Exa MCP for current web search, code examples, documentation lookup, and company research. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: Search terms and research inputs are sent to Exa's external service. <br>
Mitigation: Avoid secrets, private code, confidential business information, and sensitive personal data in queries. <br>
Risk: Optional people-search, crawling, and deep researcher tools can broaden collection or research scope. <br>
Mitigation: Use those tools only for legitimate, policy-compliant purposes and review outputs before relying on them. <br>


## Reference(s): <br>
- [ClawHub skill page](https://clawhub.ai/Whiteknight07/exa-web-search-free) <br>
- [Exa documentation](https://exa.ai/docs) <br>
- [Exa MCP server GitHub reference](https://github.com/exa-labs/exa-mcp-server) <br>
- [Exa MCP server npm package](https://www.npmjs.com/package/exa-mcp-server) <br>
- [Exa Search Examples](references/examples.md) <br>


## Skill Output: <br>
**Output Type(s):** [text, markdown, shell commands, configuration, guidance] <br>
**Output Format:** [Markdown guidance with inline mcporter shell commands and Exa MCP search-result text.] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [Requires the mcporter binary and sends search or research queries to Exa's external service.] <br>

## Skill Version(s): <br>
1.0.1 (source: server release evidence) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
