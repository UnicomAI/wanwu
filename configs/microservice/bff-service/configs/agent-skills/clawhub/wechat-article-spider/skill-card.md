## Description: <br>
Converts a user-provided WeChat public-account article URL into a local Markdown file with downloaded images. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[chenchaoqun](https://clawhub.ai/user/chenchaoqun) <br>

### License/Terms of Use: <br>
MIT-0 <br>


## Use Case: <br>
Developers and content maintainers use this skill to archive WeChat public-account articles as Markdown with local image assets for documentation or offline review. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: The skill fetches user-provided URLs and downloads referenced images. <br>
Mitigation: Use only trusted article URLs and run dependencies in a virtual environment. <br>
Risk: The default or supplied output path controls where Markdown and image files are written. <br>
Mitigation: Pass a dedicated output directory to avoid saving or overwriting files somewhere unexpected. <br>
Risk: WeChat anti-scraping behavior or dynamic image loading can cause incomplete exports. <br>
Mitigation: Retry later when fetching fails and review the generated Markdown and image folder for missing content. <br>


## Reference(s): <br>
- [ClawHub release page](https://clawhub.ai/chenchaoqun/wechat-article-spider) <br>
- [Publisher profile](https://clawhub.ai/user/chenchaoqun) <br>


## Skill Output: <br>
**Output Type(s):** [markdown, files, shell commands] <br>
**Output Format:** [Markdown file plus local image files and command-line status text] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [Writes the article Markdown file to the selected output directory and stores downloaded images under an images subdirectory.] <br>

## Skill Version(s): <br>
1.0.0 (source: frontmatter and server release metadata) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
