# Kubernetes 部署文件

## 文件说明

### 部署配置文件
- `postgres.yaml` - PostgreSQL 数据库部署文件（使用 mem_bank namespace）
- `redis.yaml` - Redis 数据库部署文件（使用 mem_bank namespace）

### 部署脚本
- `deploy.sh` - 宿主机执行的一键部署脚本
- `deploy-inside.sh` - 虚拟机内部执行的部署脚本
- `cleanup.sh` - 宿主机执行的一键清理脚本
- `cleanup-inside.sh` - 虚拟机内部执行的清理脚本

## 使用方法

### 🚀 快速部署（推荐）
```bash
# 一键部署数据库到 mem_bank namespace
./deploy/deploy.sh
```

### 🗑️ 清理部署
```bash
# 一键清理数据库部署
./deploy/cleanup.sh
```

### 📋 手动部署
```bash
# 1. 挂载项目目录
multipass mount . kube-master:/mnt/mem_bank

# 2. 创建数据目录
mkdir -p ./data/postgres ./data/redis
chmod -R 777 ./data/

# 3. 进入虚拟机手动部署
multipass shell kube-master
cd /mnt/mem_bank
./deploy/deploy-inside.sh
```

### 🔧 直接在虚拟机中操作
```bash
# 进入虚拟机
multipass shell kube-master

# 部署
/mnt/mem_bank/deploy/deploy-inside.sh

# 清理
/mnt/mem_bank/deploy/cleanup-inside.sh
```

## 连接信息

### PostgreSQL
- **地址**: localhost:30432
- **用户名**: mem_bank_user
- **密码**: mem_bank_password
- **数据库**: mem_bank
- **Namespace**: mem_bank
- **连接命令**: `psql -h localhost -p 30432 -U mem_bank_user -d mem_bank`

### Redis
- **地址**: localhost:30379
- **Namespace**: mem_bank
- **连接命令**: `redis-cli -h localhost -p 30379`

## 数据存储

数据库文件存储在：
- PostgreSQL: `./data/postgres/` ↔ `/mnt/mem_bank/data/postgres/`
- Redis: `./data/redis/` ↔ `/mnt/mem_bank/data/redis/`

这些目录通过 multipass mount 实时同步到虚拟机中。

## 架构优势

### 🔄 分层执行
- **宿主机脚本**: 处理挂载、权限、用户交互
- **虚拟机内脚本**: 直接执行 kubectl 命令，避免网络延迟

### 🎯 命名空间隔离
- 所有资源部署在 `mem_bank` namespace 中
- 便于管理和清理，避免污染默认命名空间

### ⚡ 性能优化
- 虚拟机内部执行避免了网络往返延迟
- 使用超时机制防止命令卡顿
- 智能检查现有部署状态