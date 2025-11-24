# Fix: Database Operation Failed (720700801) & Knowledge Base 400 Error

## Root Causes Identified

### 1. **MySQL Connection Issues**
- MySQL container restarted recently (only 2 minutes uptime)
- Services lost database connections and got stale connection pools
- Error: `dial tcp 172.19.0.7:3306: connect: connection refused`
- Error: `invalid connection`, `broken pipe`

### 2. **RAG Service Memory Issues**
- RAG workers being killed: `Worker was sent SIGKILL! Perhaps out of memory?`
- This causes "向量库校验失败" (Vector database validation failed) when creating knowledge bases
- Embedding model validation fails because RAG service is unstable

### 3. **Kafka Also Restarted**
- Kafka only has 10 seconds uptime
- May cause message queue issues

## Immediate Fixes

### Fix 1: Restart Backend Services (Already Done)
```bash
docker-compose restart workflow knowledge-service assistant-service
```
✅ **Status**: Completed - Services restarted successfully

### Fix 2: Wait for MySQL to Stabilize
MySQL needs a few minutes to fully initialize and accept connections. Check status:
```bash
docker logs mysql-wanwu --tail 20
docker exec mysql-wanwu mysqladmin ping -h localhost -uroot -pWanwu123456
```

### Fix 3: Increase RAG Service Memory
The RAG service is running out of memory. Edit `docker-compose.yaml`:

```yaml
rag:
  # ... existing config ...
  deploy:
    resources:
      limits:
        memory: 4G  # Increase from current limit
      reservations:
        memory: 2G  # Increase reservation
```

Then restart:
```bash
docker-compose up -d rag
```

### Fix 4: Verify Embedding Model Configuration
After services stabilize, check if embedding model is properly configured:

1. Go to Model Management in UI
2. Verify the embedding model (ID: 2) is:
   - **Status**: Enabled
   - **Type**: Embedding (NOT LLM)
   - **API endpoint**: Responding correctly

If you're using OpenRouter for embedding:
- Model name: `openai/text-embedding-ada-002`
- Model Type: **Embedding** (not LLM)
- Inference URL: `https://openrouter.ai/api/v1`

## Verification Steps

### 1. Check All Services Are Healthy
```bash
docker-compose ps
```
All services should show `(healthy)` status.

### 2. Check MySQL Connections
```bash
docker exec mysql-wanwu mysql -uroot -pWanwu123456 -e "SHOW PROCESSLIST;"
```

### 3. Check RAG Service Logs
```bash
docker logs rag-wanwu --tail 50
```
Should NOT see SIGKILL errors.

### 4. Test Knowledge Base Creation
Try creating a knowledge base again through the UI.

### 5. Test Workflow List
Navigate to workflow section - should load without 720700801 error.

## Long-term Solutions

### 1. Add Health Check Delays
Services should wait for MySQL to be fully ready before starting. Add to `docker-compose.yaml`:

```yaml
knowledge-service:
  depends_on:
    mysql:
      condition: service_healthy
```

### 2. Configure Connection Pool Retry
Backend services should have database connection retry logic with exponential backoff.

### 3. Monitor Memory Usage
```bash
docker stats --no-stream
```

If RAG consistently uses >3GB, consider:
- Reducing concurrent workers
- Using smaller embedding models
- Adding swap space

### 4. Add Persistent Volume for MySQL
Ensure MySQL data persists across restarts:
```yaml
mysql:
  volumes:
    - mysql-data:/var/lib/mysql
```

## Current Status Summary

| Issue | Status | Action Needed |
|-------|--------|---------------|
| MySQL connection lost | 🟡 Recovering | Wait 2-3 minutes |
| Services restarted | ✅ Done | Monitor logs |
| RAG memory issues | 🔴 Critical | Increase memory limit |
| Embedding model config | ⚠️ Check | Verify model type is "Embedding" |

## Quick Recovery Command

If issues persist after 5 minutes:
```bash
# Full restart of affected services
docker-compose restart mysql redis kafka es
sleep 30
docker-compose restart workflow knowledge-service assistant-service rag model-service bff-service
```

## Error Code Reference

- **720700801**: Database operation failed (workflow service)
- **141001**: Knowledge base creation failed
- **400**: Bad request (usually validation failure)

The "向量库校验失败" error means the RAG service couldn't validate the embedding model, likely because:
1. RAG service is out of memory
2. Embedding model is misconfigured (wrong type)
3. Embedding model API is not responding
