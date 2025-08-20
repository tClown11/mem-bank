# Scripts 目录

AI Memory Bank 系统的脚本工具集合，包含部署、安全检查等功能。

## 文件说明

- `deploy.sh` - 统一部署脚本
- `check-secrets.sh` - 安全检查脚本
- `README.md` - 本说明文档

## 部署指南

AI Memory Bank 系统提供简化的部署方案，支持本地开发和生产环境。

## 快速开始

### 本地开发 (推荐)
```bash
# 使用 Docker Compose 一键部署
./scripts/deploy.sh

# 或者指定开发环境
./scripts/deploy.sh -m docker -e dev
```

### 生产环境
```bash
# 使用 Kubernetes 部署到生产环境
./scripts/deploy.sh -m k8s -e prod
```

## 命令参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-m, --mode` | 部署模式: `docker` 或 `k8s` | `docker` |
| `-e, --env` | 环境: `dev`, `staging`, `prod` | `dev` |
| `-c, --clean` | 清理现有部署 | - |
| `-h, --help` | 显示帮助信息 | - |

## 使用示例

```bash
# 本地开发环境 (Docker Compose)
./scripts/deploy.sh

# Kubernetes 生产环境  
./scripts/deploy.sh -m k8s -e prod

# 清理所有部署
./scripts/deploy.sh -c

# 查看帮助
./scripts/deploy.sh -h
```

## 连接信息

### Docker Compose 模式
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379  
- **API**: http://localhost:8080

### Kubernetes 模式
- **PostgreSQL**: localhost:30432
- **Redis**: localhost:30379

### 数据库凭据
- **用户名**: mem_bank_user
- **密码**: mem_bank_password
- **数据库**: mem_bank

## 数据持久化

数据文件存储在 `./data/` 目录：
- PostgreSQL: `./data/postgres/`
- Redis: `./data/redis/`

## 依赖要求

### Docker Compose 模式
- Docker
- Docker Compose

### Kubernetes 模式  
- kubectl
- 可访问的 Kubernetes 集群

## 故障排除

### 检查服务状态
```bash
# Docker Compose
docker-compose ps
docker-compose logs

# Kubernetes
kubectl get pods -n mem-bank
kubectl logs -n mem-bank <pod-name>
```

### 重置部署
```bash
# 清理并重新部署
./scripts/deploy.sh -c
./scripts/deploy.sh
```

## 安全检查脚本

### check-secrets.sh

用于检查代码中是否包含潜在的敏感信息，防止意外提交密码、密钥等。

```bash
# 检查当前暂存的文件
./scripts/check-secrets.sh

# 设置为 Git pre-commit hook（推荐）
ln -sf ../../scripts/check-secrets.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

**检查项目：**
- 密码、密钥、令牌等敏感字符串
- 硬编码的 IP 地址
- Base64 编码的可疑数据
- 私钥文件内容

**使用建议：**
- 设置为 Git pre-commit hook 自动检查
- 在 CI/CD 流程中集成此检查
- 发现问题时，将敏感信息移动到环境变量或 .env 文件中