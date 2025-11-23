# Docker Compose Files - Backup/Legacy

Thư mục này chứa các docker-compose files cũ đã được thay thế.

## Files trong thư mục này:

- `docker-compose.yml` - Main orchestrator (legacy)
- `docker-compose.network.yml` - Fabric network (đã gộp vào dev/prod)
- `docker-compose.services.yml` - Backend services (đã gộp vào dev/prod)
- `docker-compose.gateway.yml` - API Gateway (đã gộp vào dev/prod)
- `docker-compose.monitoring.yml` - Monitoring (đã gộp vào dev/prod)

## Sử dụng file mới:

**Development:**
```bash
docker-compose -f docker-compose.dev.yml up -d
```

**Production:**
```bash
docker-compose -f docker-compose.prod.yml up -d
```

## Khi nào dùng file cũ?

Chỉ dùng các file trong thư mục này khi:
- Cần debug riêng từng module
- CI/CD cần test từng phần
- Reference cho cấu trúc

## Khôi phục (nếu cần):

```bash
cp docker-compose-backup/*.yml .
```
