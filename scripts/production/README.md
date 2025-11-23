# Production Hardening Scripts

Thư mục này chứa các scripts và cấu hình cho **Production Hardening** của IBN Network.

## 📁 Files

### Scripts

- **`generate-certs.sh`** - Tạo self-signed SSL certificates
- **`setup-letsencrypt.sh`** - Setup Let's Encrypt certificates (production)
- **`backup-production.sh`** - Automated backup script
- **`restore-production.sh`** - Restore từ backup
- **`production-checklist.sh`** - Kiểm tra production readiness

## 🚀 Quick Start

### 1. Generate SSL Certificates

```bash
./generate-certs.sh
```

### 2. Setup Automated Backup

```bash
# Test backup
./backup-production.sh

# Add to crontab
crontab -e
# Add: 0 2 * * * /home/exp2/ibn/scripts/production/backup-production.sh
```

### 3. Run Production Checklist

```bash
./production-checklist.sh
```

## 📚 Documentation

Xem chi tiết tại: `/home/exp2/ibn/docs/v1.0.1/production-hardening.md`

## ⚙️ Configuration

### Backup Directory

Mặc định: `/backup/ibn-network`

Có thể thay đổi bằng biến môi trường:
```bash
export BACKUP_DIR=/custom/backup/path
./backup-production.sh
```

### Retention Policy

Mặc định: 30 ngày

Có thể thay đổi:
```bash
export RETENTION_DAYS=60
./backup-production.sh
```

## 🔧 Troubleshooting

### Backup Issues

```bash
# Check logs
tail -f /var/log/ibn-backup.log

# Manual backup test
BACKUP_DIR=/tmp/test-backup ./backup-production.sh
```

### Certificate Issues

```bash
# Verify certificate
openssl x509 -in ../../certs/keycloak.crt -text -noout

# Regenerate
rm -rf ../../certs/*
./generate-certs.sh
```

## 📝 Notes

- Tất cả scripts cần quyền thực thi: `chmod +x *.sh`
- Backup script cần Docker đang chạy
- Let's Encrypt script cần quyền root (sudo)
- Checklist script không cần quyền đặc biệt

