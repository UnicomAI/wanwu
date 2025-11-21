# Quick Start - EHS Workflows

## ✅ What's Ready

Your platform now shows **ONLY 7 EHS workflows** - all in English!

## 🚀 Quick Restart

```bash
# Navigate to project directory
cd /Users/mohankumarv/Desktop/SAFVR/wanwu

# Restart the application
docker-compose restart

# OR restart specific service
docker restart bff-service
```

## 👀 What You'll See in UI

### Categories (EHS-focused):
1. **Incident Investigation** - Incident reporting from images
2. **Compliance** - OSHA search & policy generation  
3. **Corrective Actions** - Action plan generation
4. **Training** - Training record formatting
5. **Audit** - Evidence compilation
6. **Reporting** - Leadership reports

### 7 Workflows:
1. ✅ Incident Report from Image
2. ✅ OSHA Regulation Search
3. ✅ Safety Policy Generator
4. ✅ Corrective Action Plan Generator
5. ✅ Training Record Formatter
6. ✅ Audit Evidence Compiler
7. ✅ Leadership Safety Report

## 🔧 Configuration Needed

After restart, configure each workflow:

### All Workflows Need:
- **LLM Model** - Select your AI model (GPT-4, Claude, etc.)

### Web Search Workflows Need:
- **OSHA Regulation Search** - Configure Perplexity MCP or web search service
- **Safety Policy Generator** - Configure Perplexity MCP or web search service

### Document Parsing Workflows Need:
- **Audit Evidence Compiler** - Configure document parsing service

## 📝 Test Each Workflow

1. Go to Workflows page
2. Select "Incident Report from Image"
3. Upload a test image
4. Verify it generates a report
5. Repeat for other workflows

## 🗂️ File Locations

```
Active Workflows: /configs/.../workflow-template/
Archived (backup): /configs/.../workflow-template-archived/
Config: /configs/.../workflow_template_config.yaml
```

## 📚 Documentation

- `UI_CLEANUP_SUMMARY.md` - What was changed
- `EHS_WORKFLOW_TRANSFORMATION.md` - Detailed workflow info
- `WORKFLOW_PLAN.md` - Original planning doc

## ⚠️ Troubleshooting

**If old workflows still appear:**
1. Hard refresh browser: Ctrl+F5 (Windows) or Cmd+Shift+R (Mac)
2. Clear browser cache completely
3. Restart the application
4. Check `workflow_template_config.yaml` contains only EHS workflows

**If workflows don't work:**
1. Configure LLM model in each workflow
2. Configure MCP services for web search workflows
3. Check Docker logs: `docker logs bff-service`

## 🎯 Success Criteria

✅ Only 7 workflows visible in UI  
✅ All categories are EHS-focused  
✅ All text is in English  
✅ No Chinese characters appear  
✅ Old workflows (tourism, education, etc.) are hidden  

---

**Ready to use!** 🎉
