# Quick Fix Reference - Categories Not Showing

## ⚡ Problem
Old categories (Government, Industry, Tourism, etc.) still showing in UI instead of new EHS categories.

## ✅ Solution (Completed)

### What Was Wrong
The frontend Vue files had hardcoded old category lists that weren't reading from the backend config.

### What Was Fixed
1. ✅ Updated `web/src/views/templateSquare/tempSquare.vue` 
2. ✅ Updated `web/src/views/templateSquare/promptTempSquare.vue`
3. ✅ Restarted `bff-service` (backend)
4. ✅ Restarted `nginx-wanwu` (frontend)

---

## 🚀 What To Do Now

### 1. Wait 30 Seconds
Let the services fully start up.

### 2. Clear Browser Cache (CRITICAL!)
```
Windows/Linux: Ctrl + F5
Mac: Cmd + Shift + R
```

### 3. Go To Template Gallery
```
http://localhost:8081/templateSquare
```

### 4. You Should See
```
Tabs: [All] [Incident Investigation] [Compliance] [Corrective Actions] 
      [Training] [Audit] [Reporting] [Safety Inspection]
```

---

## ❌ Still Not Working?

### Try This (in order):

**1. Hard Refresh (Again)**
```
Ctrl + F5 (Windows/Linux)
Cmd + Shift + R (Mac)
```

**2. Clear All Browser Cache**
- Open browser settings
- Privacy → Clear browsing data
- Select "Cached images and files"
- Clear

**3. Check Services**
```bash
docker ps | grep -E "bff-service|nginx-wanwu"
```
Both should show "Up" and "(healthy)"

**4. Restart Services Again**
```bash
docker restart bff-service nginx-wanwu
```
Wait 30 seconds, then hard refresh browser.

**5. Check Logs**
```bash
docker logs bff-service --tail 20
docker logs nginx-wanwu --tail 20
```

**6. Full System Restart**
```bash
cd /Users/mohankumarv/Desktop/SAFVR/wanwu
docker-compose restart
```
Wait 1 minute, then hard refresh browser.

---

## 🎯 Expected Categories

### Workflow Templates:
- All
- Incident Investigation
- Compliance
- Corrective Actions
- Training
- Audit
- Reporting
- Safety Inspection

### Each Category Shows:
- **All**: 7 workflows
- **Incident Investigation**: 1 workflow
- **Compliance**: 2 workflows
- **Corrective Actions**: 1 workflow
- **Training**: 1 workflow
- **Audit**: 1 workflow
- **Reporting**: 1 workflow
- **Safety Inspection**: 0 workflows (for future)

---

## 📞 Debug Info

If categories still wrong, check:

```bash
# 1. Check workflow config has correct categories
cat /Users/mohankumarv/Desktop/SAFVR/wanwu/configs/microservice/bff-service/configs/workflow_template_config.yaml | grep "category:"

# Should show: incident_investigation, compliance (x2), corrective_actions, training, audit, reporting

# 2. Check frontend Vue file was updated
grep "incident_investigation" /Users/mohankumarv/Desktop/SAFVR/wanwu/web/src/views/templateSquare/tempSquare.vue

# Should find matches

# 3. Check services are running
docker ps --filter "name=wanwu" --format "table {{.Names}}\t{{.Status}}"

# All should show "Up" and most should show "(healthy)"
```

---

## ✅ Success Checklist

After clearing cache, verify:

- [ ] Browser cache cleared (hard refresh)
- [ ] Services are running (docker ps)
- [ ] Navigate to http://localhost:8081/templateSquare
- [ ] See 8 category tabs
- [ ] Categories are in English
- [ ] Categories are EHS-focused
- [ ] No old categories visible
- [ ] Clicking categories filters workflows

---

## 🎉 You're Done When...

You see this at the top of Template Gallery:
```
[All] [Incident Investigation] [Compliance] [Corrective Actions]
[Training] [Audit] [Reporting] [Safety Inspection]
```

And NO old categories like:
- ❌ Government
- ❌ Industry  
- ❌ Education
- ❌ Tourism
- ❌ Data
- ❌ Creation
- ❌ Search

---

**Last Updated:** November 21, 2025  
**Status:** Fixed - Services Restarted  
**Action Required:** Clear browser cache!
