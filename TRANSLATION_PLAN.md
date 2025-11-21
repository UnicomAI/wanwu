# Wanwu AI Agent Platform - English Translation Plan

## Executive Summary

The Wanwu AI Agent Platform currently has its UI entirely in Chinese. This document outlines a comprehensive, phased approach to translate the entire application to English. The project is massive, involving frontend, backend, documentation, and configuration files across multiple services.

## Current State Analysis

### ✅ Already in Place
- **Vue i18n Infrastructure**: The frontend already has internationalization setup
  - `/web/src/lang/index.js` - i18n configuration
  - `/web/src/lang/zh.js` - Chinese translations (1,212 lines)
  - `/web/src/lang/en.js` - Partial English translations (exists but incomplete)
  - Currently defaults to Chinese (`localStorage.getItem('locale') || ZH`)

### 🔴 Needs Translation

#### 1. **Frontend (Vue.js) - HIGH PRIORITY**
- **Location**: `/web/src/lang/`
- **Files**: 
  - `zh.js` - Complete Chinese translations (1,212 lines)
  - `en.js` - Needs completion to match `zh.js`
- **Impact**: Direct user-facing UI text
- **Estimated Lines**: ~1,212 translation keys

#### 2. **Backend Go Services - MEDIUM PRIORITY**
- **Locations**:
  - `/internal/` - Service implementations
  - `/pkg/` - Shared packages
  - `/proto/` - gRPC proto definitions
- **Content**:
  - Error messages with Chinese text
  - Chinese comments (marked with `[EN]` for some)
  - i18n error code mappings
- **Estimated Files**: ~50+ Go files with Chinese content

#### 3. **Python Services (Agent & RAG) - MEDIUM PRIORITY**
- **Locations**:
  - `/agent/agent_open_source/` - Agent service
  - `/rag/rag_open_source/` - RAG service
- **Content**:
  - Chinese comments
  - Log messages in Chinese
  - Error messages
- **Estimated Files**: ~30+ Python files

#### 4. **Documentation - MEDIUM PRIORITY**
- **Location**: `/configs/microservice/bff-service/static/manual/`
- **Files**: 35+ markdown documentation files
  - Model management
  - Knowledge base
  - Workflows
  - Agent configuration
  - Tool library
  - Safety guardrails
- **Impact**: User onboarding and help
- **Estimated Pages**: 35+ documentation pages

#### 5. **Configuration & Scripts - LOW PRIORITY**
- Error code proto files with Chinese descriptions
- Build scripts with Chinese comments

---

## 📋 Translation Strategy - 8 Phase Approach

### **Phase 1: Frontend UI Translation (CRITICAL PATH)** ⏰ 2-3 days
**Goal**: Complete English translation of all user-facing UI text

**Tasks**:
1. Compare `zh.js` and `en.js` to identify missing translations
2. Translate all missing keys in `en.js` to match `zh.js` structure
3. Review and improve existing English translations
4. Test language switching functionality
5. Update default language setting to English

**Files to Modify**:
- `/web/src/lang/en.js` - Complete translation
- `/web/src/lang/index.js` - Change default from `ZH` to `'en'`
- `/web/src/lang/constants.js` - Verify language codes

**Validation Checklist**:
- [ ] All UI elements display in English
- [ ] No Chinese characters in UI (except user content)
- [ ] Language switcher works between EN/ZH
- [ ] Forms, buttons, tooltips all translated
- [ ] Error messages appear in English

---

### **Phase 2: Backend Error Messages** ⏰ 2-3 days
**Goal**: Translate all backend error messages and API responses

**Tasks**:
1. Review i18n error code structure in `/pkg/i18n/`
2. Identify all error message files
3. Create English translations for error codes
4. Update proto files with English descriptions
5. Test error message display

**Files to Modify**:
- `/pkg/i18n/*.go` - Error message translations
- `/proto/err-code/err-code.proto` - Error descriptions
- Service-specific error messages in `/internal/`

**Validation Checklist**:
- [ ] API errors return English messages
- [ ] gRPC status codes have English descriptions
- [ ] Log files show English error messages

---

### **Phase 3: Code Comments Translation** ⏰ 3-4 days
**Goal**: Translate all Chinese comments in Go and Python code

**Tasks**:
1. Scan all Go files for Chinese comments
2. Scan all Python files for Chinese comments
3. Replace with English equivalents
4. Maintain `[EN]` markers for clarity
5. Update inline documentation

**Files to Modify**:
- `/internal/**/*.go` - ~50 files
- `/pkg/**/*.go` - ~20 files
- `/agent/**/*.py` - ~30 files
- `/rag/**/*.py` - ~20 files

**Approach**:
- Use AI translation tool (like lingo.dev) for bulk translation
- Manual review for technical accuracy
- Keep original Chinese in comments as `// 原文：中文` if context needed

---

### **Phase 4: User Documentation Translation** ⏰ 5-7 days
**Goal**: Translate all user-facing documentation

**Tasks**:
1. Create `/configs/microservice/bff-service/static/manual_en/` directory
2. Translate all 35+ markdown files
3. Update screenshot paths if needed
4. Verify links and references
5. Create English version of help documentation

**Files to Create/Modify**:
- Create English versions of all `.md` files in `manual/`
- Update documentation index to support EN/ZH switching

**Priority Order**:
1. Platform introduction (`0.平台介绍.md`)
2. Model management (`1.模型管理.md`)
3. Knowledge base (`2.知识库/*.md`)
4. Agent creation (`7.智能体.md`)
5. Workflow (`6.工作流/*.md`)
6. Settings (`11.设置.md`)

---

### **Phase 5: Proto Definitions & API Docs** ⏰ 2 days
**Goal**: Translate gRPC proto file comments and API documentation

**Tasks**:
1. Translate comments in all `.proto` files
2. Update field descriptions
3. Generate updated API documentation
4. Update Swagger/OpenAPI specs if present

**Files to Modify**:
- `/proto/**/*.proto` - 17 proto files
- `/docs/**/*.yaml` - API documentation

---

### **Phase 6: Configuration Files** ⏰ 1 day
**Goal**: Translate configuration file comments and descriptions

**Tasks**:
1. Review YAML/JSON config files
2. Translate comment strings
3. Update default values if language-specific

**Files to Modify**:
- `/configs/**/*.yaml`
- Environment variable descriptions in `.env.bak`

---

### **Phase 7: Python Service Logs & Messages** ⏰ 2 days
**Goal**: Translate Python logging and user-facing messages

**Tasks**:
1. Identify all logger messages
2. Replace with English equivalents
3. Update exception messages
4. Test log output

**Files to Modify**:
- All Python files with `logger.info()`, `logger.error()`, etc.
- Exception messages in try/except blocks

---

### **Phase 8: Testing & Quality Assurance** ⏰ 3-4 days
**Goal**: Comprehensive testing of English version

**Tasks**:
1. Full application walkthrough in English
2. Test all user flows:
   - User registration/login
   - Model management
   - Knowledge base creation
   - Agent creation
   - Workflow design
   - Tool configuration
3. Verify error handling displays English
4. Test documentation accessibility
5. Browser testing (Chrome, Firefox, Safari)
6. Mobile responsiveness check

**Test Cases**:
- [ ] New user onboarding (English only)
- [ ] Create agent end-to-end
- [ ] Upload document to knowledge base
- [ ] Design and publish workflow
- [ ] Error scenarios (invalid input, network errors)
- [ ] Help documentation navigation

---

## 🛠️ Tools & Resources

### Recommended Translation Tools

1. **lingo.dev** (as suggested)
   - Good for bulk translation of UI strings
   - Maintains context for technical terms
   - Can preserve formatting

2. **DeepL API**
   - High-quality technical translation
   - Better for documentation
   - Paid but worth it for quality

3. **Google Cloud Translation API**
   - Good for code comments
   - Batch processing support

4. **Manual Review Tools**
   - VS Code with Chinese-English dictionary extension
   - Glossary of technical terms (create custom)

### Translation Glossary (Critical Terms)

Create a glossary for consistency:

| Chinese | English | Notes |
|---------|---------|-------|
| 智能体 | Agent | Core concept |
| 知识库 | Knowledge Base | |
| 工作流 | Workflow | |
| 模型管理 | Model Management | |
| 应用广场 | App Marketplace | |
| 模板广场 | Template Gallery | |
| 资源库 | Resource Library | |
| 工作室 | Studio | |
| 安全护栏 | Safety Guardrails | |
| 文本问答 | Text Q&A | Or RAG |
| 组织管理 | Organization Management | |
| 角色管理 | Role Management | |
| 用户管理 | User Management | |
| 提示词 | Prompt | |
| 分段 | Chunk / Segment | Use "Chunk" for technical, "Segment" for UI |
| 命中测试 | Hit Test | |
| 元数据 | Metadata | |
| 向量检索 | Vector Search | |
| 全文检索 | Full-text Search | |
| 混合检索 | Hybrid Search | |
| 创建 | Create | |
| 编辑 | Edit | |
| 删除 | Delete | |
| 发布 | Publish | |
| 保存 | Save | |
| 确定 | Confirm | |
| 取消 | Cancel | |

---

## 📊 Resource Estimation

### Time Investment (Single Person)
- **Phase 1**: 2-3 days (16-24 hours)
- **Phase 2**: 2-3 days (16-24 hours)
- **Phase 3**: 3-4 days (24-32 hours)
- **Phase 4**: 5-7 days (40-56 hours)
- **Phase 5**: 2 days (16 hours)
- **Phase 6**: 1 day (8 hours)
- **Phase 7**: 2 days (16 hours)
- **Phase 8**: 3-4 days (24-32 hours)

**Total**: 20-26 days (160-208 hours) for a single developer

### Team Approach (Recommended)
With 3-4 people working in parallel:
- **Frontend Developer**: Phase 1 (2-3 days)
- **Backend Developer**: Phases 2, 5, 6 (5 days)
- **Technical Writer**: Phase 4 (5-7 days)
- **QA Engineer**: Phase 8 (3-4 days)
- **Everyone**: Phase 3, 7 (split work)

**Total Team Time**: ~2 weeks with parallel work

---

## 🚀 Quick Start Implementation

### Immediate Actions (Day 1)

1. **Set up translation infrastructure**:
   ```bash
   # Create backup of current Chinese version
   git checkout -b feature/english-translation
   
   # Create translation tracking file
   touch TRANSLATION_PROGRESS.md
   ```

2. **Start with Phase 1** (Frontend):
   ```bash
   cd web/src/lang
   # Back up original
   cp en.js en.js.backup
   cp zh.js zh.js.backup
   ```

3. **Use AI Assistant for bulk translation**:
   - Feed `zh.js` content to Claude/ChatGPT or lingo.dev
   - Request JSON-preserving translation
   - Review output manually
   - Update `en.js`

4. **Test immediately**:
   ```bash
   cd web
   npm run dev
   # Test language switcher
   # Verify UI elements
   ```

---

## ⚠️ Critical Considerations

### 1. **Maintain Backwards Compatibility**
- Keep Chinese (`zh.js`) intact
- Ensure language switcher works
- Default can be English, but support both

### 2. **Context-Aware Translation**
- Don't blindly translate technical terms
- Keep consistency across the application
- Some terms better left in English (API, URL, JSON, etc.)

### 3. **Cultural Adaptations**
- Date formats (MM/DD/YYYY vs DD/MM/YYYY)
- Number formatting (commas vs periods)
- Example data should be English-appropriate

### 4. **Testing Strategy**
- Test with fresh user account
- Have native English speaker review
- Check for Chinglish (Chinese-English hybrid mistakes)

### 5. **Documentation Quality**
- Screenshots may need retaking in English
- Examples should use English sample data
- Links should point to English resources when available

---

## 📈 Success Metrics

### Phase 1 Complete When:
- [ ] All UI text appears in English
- [ ] No untranslated Chinese in interface
- [ ] Language switch works seamlessly
- [ ] Forms validate in English

### Project Complete When:
- [ ] New English user can use entire platform without seeing Chinese
- [ ] All documentation available in English
- [ ] Error messages appear in English
- [ ] Code comments readable by English-speaking developers
- [ ] API documentation fully English
- [ ] No Chinese text in logs (except user content)

---

## 🔄 Maintenance Plan

### Ongoing Translation Requirements

1. **New Features**: 
   - Add translations to both `en.js` and `zh.js` simultaneously
   - Update documentation in both languages

2. **Error Messages**:
   - Add both languages to i18n error codes
   - Test in both languages before deployment

3. **Documentation**:
   - Maintain parallel English/Chinese versions
   - Update both with each release

4. **Code Comments**:
   - Write new comments in English
   - Consider removing `[EN]` markers once fully English

---

## 💡 Automation Opportunities

### Scripts to Create

1. **Translation Validator**:
   ```bash
   # Script to check for Chinese characters in JS/Vue files
   ./scripts/check-chinese-in-code.sh
   ```

2. **i18n Key Sync Checker**:
   ```bash
   # Ensure en.js and zh.js have same keys
   ./scripts/validate-i18n-keys.js
   ```

3. **Documentation Sync Tool**:
   ```bash
   # Track which docs are updated in one language but not other
   ./scripts/sync-docs.py
   ```

---

## 📞 Next Steps

### To Begin Implementation:

1. **Review this plan** with your team
2. **Assign roles** to team members
3. **Set up project tracking** (Jira, GitHub Projects, etc.)
4. **Create translation glossary** (spend 2-3 hours on this)
5. **Start Phase 1** immediately
6. **Daily standups** to track progress
7. **Weekly demos** of translated sections

### First Week Goals:
- ✅ Complete Phase 1 (Frontend UI)
- ✅ Start Phase 2 (Backend errors)
- ✅ Create translation glossary
- ✅ Set up quality review process

---

## 📋 Progress Tracking

Use this checklist to track overall progress:

### Phase 1: Frontend UI
- [ ] Audit zh.js vs en.js differences
- [ ] Translate all missing keys
- [ ] Review existing translations
- [ ] Test language switcher
- [ ] Update default language
- [ ] QA verification

### Phase 2: Backend Errors
- [ ] Identify all error message sources
- [ ] Create English error translations
- [ ] Update proto files
- [ ] Test API error responses
- [ ] Verify log messages

### Phase 3: Code Comments
- [ ] Scan Go files
- [ ] Scan Python files
- [ ] Batch translate comments
- [ ] Manual review
- [ ] Test builds

### Phase 4: Documentation
- [ ] Create manual_en directory
- [ ] Translate platform intro
- [ ] Translate model management
- [ ] Translate knowledge base docs
- [ ] Translate workflow docs
- [ ] Translate agent docs
- [ ] Translate settings docs
- [ ] Update screenshots
- [ ] Verify all links

### Phase 5: Proto & API Docs
- [ ] Translate proto comments
- [ ] Update API docs
- [ ] Regenerate docs
- [ ] Verify accuracy

### Phase 6: Configuration
- [ ] Review config files
- [ ] Translate comments
- [ ] Update examples

### Phase 7: Python Logs
- [ ] Identify log messages
- [ ] Translate messages
- [ ] Update exceptions
- [ ] Test output

### Phase 8: QA
- [ ] Full app walkthrough
- [ ] User flow testing
- [ ] Error scenario testing
- [ ] Browser compatibility
- [ ] Mobile testing
- [ ] Documentation testing
- [ ] Final review

---

## 🎯 AI Translation Prompt Template

Use this template with lingo.dev or Claude/ChatGPT:

```
TRANSLATION TASK: Chinese to English for Wanwu AI Agent Platform

CONTEXT: Enterprise-grade AI agent development platform with microservice architecture

REQUIREMENTS:
1. Translate ONLY the VALUES, preserve all KEYS
2. Maintain JSON structure exactly
3. Keep technical terms in English: API, JSON, URL, Token, Workflow, Agent, RAG, MCP
4. Use formal but friendly tone (B2B SaaS product)
5. Preserve placeholders like {count}, {{variable}}
6. Keep HTML tags unchanged
7. Use American English spelling

GLOSSARY (USE EXACTLY):
- 智能体 → Agent
- 知识库 → Knowledge Base
- 工作流 → Workflow
- 模型 → Model
- 应用 → Application
- 组织 → Organization
- 角色 → Role
- 用户 → User
- 提示词 → Prompt
- 分段 → Chunk
- 命中测试 → Hit Test
- 元数据 → Metadata

SOURCE TEXT:
[Paste zh.js content here]

Please provide the translated JSON with all keys unchanged and values in English.
```

---

## Contact & Support

For questions about this translation plan:
- Create GitHub Issue in the wanwu repository
- Tag with `translation` and `documentation`
- Reference this plan document

**Document Version**: 1.0  
**Created**: 2025-01-19  
**Last Updated**: 2025-01-19  
**Status**: Ready for Implementation

---

**Good luck with the translation project! Break it into phases, celebrate small wins, and maintain quality throughout.** 🚀
