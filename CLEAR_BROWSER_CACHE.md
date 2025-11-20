# Complete Browser Cache Clear Instructions

## ✅ Server-Side Fixed
- Database: All 4 workflows have English text ✅
- Redis: Cache cleared ✅  
- Workflow Service: Restarted ✅

## Browser-Side Issue
The workflow editor stores data in **browser local storage** which doesn't clear with Cmd+Shift+R.

## Complete Browser Clear (Choose One Method)

### Method 1: Clear Site Data (Recommended)
1. Open the workflow page: `http://localhost:8081`
2. Press `F12` to open Developer Tools
3. Go to **Application** tab (Chrome) or **Storage** tab (Firefox)
4. In the left sidebar, find **Local Storage** → `http://localhost:8081`
5. Right-click on it → **Clear**
6. Also clear **Session Storage** the same way
7. Close Developer Tools
8. Close the browser tab completely
9. Open a new tab and go to `http://localhost:8081`

### Method 2: Clear All Browser Data
1. **Chrome/Edge**:
   - Press `Cmd + Shift + Delete` (Mac) or `Ctrl + Shift + Delete` (Windows)
   - Time range: **All time**
   - Check: ✅ Cookies and other site data
   - Check: ✅ Cached images and files
   - Click **Clear data**

2. **Firefox**:
   - Press `Cmd + Shift + Delete` (Mac) or `Ctrl + Shift + Delete` (Windows)
   - Time range: **Everything**
   - Check: ✅ Cookies
   - Check: ✅ Cache
   - Click **Clear Now**

### Method 3: Incognito/Private Window (Quick Test)
1. Open a new Incognito/Private window
2. Go to `http://localhost:8081`
3. Log in and check the workflow
4. This bypasses all caching

### Method 4: Different Browser (Quick Test)
If you're using Chrome, try Firefox or Safari to verify the fix worked.

## After Clearing Cache

1. Go to: `http://localhost:8081`
2. Log in
3. Open your workflow
4. You should now see:
   - **Start** (not 开始)
   - **End** (not 结束)

## Why This Is Needed

The workflow editor is a Single Page Application (SPA) that caches data in:
1. Browser cache (cleared by Cmd+Shift+R)
2. **Local Storage** (NOT cleared by Cmd+Shift+R) ← THIS IS THE ISSUE
3. Session Storage (NOT cleared by Cmd+Shift+R)
4. IndexedDB (NOT cleared by Cmd+Shift+R)

## Verification

After clearing, if you still see Chinese text, run this in browser console:
```javascript
// Open console (F12 → Console tab)
localStorage.clear();
sessionStorage.clear();
location.reload(true);
```

---

**Issue**: Workflow nodes showing Chinese text  
**Root Cause**: Browser local storage caching old workflow data  
**Fix**: Clear browser local storage completely
