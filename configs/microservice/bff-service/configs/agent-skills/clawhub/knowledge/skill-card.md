## Description: <br>
Local knowledge-base integration for document retrieval, knowledge ingestion, and switching between local retrieval and AnythingLLM conversation modes. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[zhbgyj](https://clawhub.ai/user/zhbgyj) <br>

### License/Terms of Use: <br>


## Use Case: <br>
External users and developers use this skill to query a local knowledge base, list documents, view status and statistics, and switch between local retrieval and AnythingLLM conversation mode. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: Core runtime behavior depends on a missing local helper, so users cannot verify from this package alone how documents and queries are handled. <br>
Mitigation: Install only after inspecting and trusting the missing helper and the local service it calls. <br>
Risk: Queries or uploaded documents may involve sensitive knowledge-base content and may be routed through local retrieval or AnythingLLM mode. <br>
Mitigation: Avoid sensitive documents until storage, deletion, and provider routing behavior are understood and approved. <br>
Risk: The skill expects a local knowledge-base path and a service at 127.0.0.1:8001, which may not exist or may be configured differently in the user's environment. <br>
Mitigation: Verify the local service, knowledge-base location, and mode-switching behavior in a controlled environment before relying on the skill. <br>


## Reference(s): <br>
- [ClawHub Skill Page](https://clawhub.ai/zhbgyj/knowledge) <br>


## Skill Output: <br>
**Output Type(s):** [Text, Guidance] <br>
**Output Format:** [Plain text responses with Markdown-style lists] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [May include retrieval answers, source labels, mode/status messages, document lists, and knowledge-base statistics.] <br>

## Skill Version(s): <br>
1.0.0 (source: server release metadata) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
