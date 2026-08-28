## Description: <br>
本地知识库 supports local storage, search, deletion, statistics, and context recall for user-provided preferences, facts, tasks, and other memory. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[toughspider](https://clawhub.ai/user/toughspider) <br>

### License/Terms of Use: <br>
MIT-0 <br>


## Use Case: <br>
Agents use this skill to keep user-provided memory in a local SQLite knowledge base, then retrieve, summarize, or delete entries when the user asks. It is intended for local personal knowledge management and should not be used to store passwords, tokens, recovery codes, or highly sensitive personal data. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: Review before execution as proposals could introduce incorrect or misleading guidance into skills. <br>
Mitigation: Review and scan skill before deployment. <br>

## Reference(s): <br>
- [ClawHub skill page](https://clawhub.ai/toughspider/knowledge-spider) <br>


## Skill Output: <br>
**Output Type(s):** [text, markdown, json, guidance] <br>
**Output Format:** [JSON responses with natural-language messages and optional Markdown context snippets] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [Stores and searches local SQLite entries; returned memories may influence later agent responses, so users should make save, search, and delete requests precise.] <br>

## Skill Version(s): <br>
2.0.0 (source: server release metadata and skill.json) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
