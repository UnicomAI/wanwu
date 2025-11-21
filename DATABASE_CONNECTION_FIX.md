# Database Connection and i18n Nil Pointer Fixes

## Issues Fixed

### 1. MySQL Connection Refused for operate-service
**Problem:** `operate-service` was starting before MySQL was ready, causing "dial tcp 172.19.0.7:3306: connect: connection refused" errors.

**Root Cause:** Missing `depends_on` configuration in docker-compose.yaml

**Fix:** Added MySQL dependency with health check condition
```yaml
operate-service:
  depends_on:
    mysql:
      condition: service_healthy
```

**File Modified:** `docker-compose.yaml` (lines 506-509)

---

### 2. i18n Nil Pointer Dereference Panic
**Problem:** `[i18n] panic: runtime error: invalid memory address or nil pointer dereference`

**Root Cause:** The i18n `ByCodeOrKey` function was accessing `_i18n` without checking if it was `nil` first. While there was a check inside, it needed to be at the very start of the function.

**Fix:** Added defensive nil check at the beginning of `ByCodeOrKey()`
```go
func ByCodeOrKey(lang Lang, code err_code.Code, key string, args []string) string {
    // Defensive nil check to prevent panic
    if _i18n == nil {
        return fmt.Sprintf("[i18n not initialized] lang(%v) err_code(%v) text_key(%v) args: %v", lang, code, key, args)
    }
    // ... rest of function
}
```

**File Modified:** `pkg/i18n/api.go` (lines 69-95)

---

## Deployment Steps

### 1. Rebuild Backend Image (Currently Running)
```bash
make docker-image-backend
```

### 2. Restart Affected Services
After the build completes, restart the services:

```bash
# Stop and remove affected containers
docker-compose down app-service operate-service bff-service

# Start services with updated configuration
docker-compose up -d app-service operate-service bff-service
```

### 3. Verify Fixes
Check logs to confirm no more errors:

```bash
# Check operate-service starts successfully
docker logs operate-service --tail 50

# Check app-service
docker logs app-service --tail 50

# Check bff-service for i18n errors
docker logs bff-service --tail 50
```

### 4. Test Application Features
- Test "Get App Favorite List" endpoint
- Test "Get Custom Configuration" endpoint
- Verify no more connection refused errors
- Verify no more i18n panic errors

---

## Technical Details

### Why operate-service needed MySQL dependency
The service connects to MySQL in `main.go:56-59`:
```go
db, err := db.New(config.Cfg().DB)
if err != nil {
    log.Fatalf("init db err: %v", err)
}
```

Without the `depends_on` health check, Docker starts `operate-service` immediately, but MySQL might not be ready to accept connections yet.

### Why i18n was panicking
When errors occurred before services were fully initialized, or during early startup phases, error formatting code would call:
```go
i18n.ByCodeOrKey(lang, code, key, args)
```

If `_i18n` was `nil` (not yet initialized), accessing `_i18n.Keys` or `_i18n.CodeKeys` would cause a nil pointer dereference panic.

The fix ensures a graceful error message instead of a panic, making debugging easier.

---

## Status
- ✅ docker-compose.yaml updated
- ✅ pkg/i18n/api.go updated with nil check
- 🔄 Backend image rebuilding (in progress)
- ⏳ Service restart pending (after build completes)
