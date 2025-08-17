#!/bin/bash

# Multipass + Kubernetes 数据库部署脚本

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 输出函数
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

info "开始部署数据库到 Kubernetes..."

# 1. 检查 multipass 挂载
info "1. 检查 multipass 挂载状态..."
if ! multipass exec kube-master -- test -d /mnt/mem_bank; then
    info "挂载项目目录到 kube-master..."
    multipass mount . kube-master:/mnt/mem_bank
    success "项目目录挂载完成"
else
    success "项目目录已挂载"
fi

# 2. 检查和创建数据目录
info "2. 检查数据目录状态..."
DATA_EXISTS=false
if [ -d "./data/postgres" ] && [ -d "./data/redis" ]; then
    DATA_EXISTS=true
    warning "数据目录已存在"
    
    # 检查数据目录是否有数据
    if [ "$(ls -A ./data/postgres 2>/dev/null)" ] || [ "$(ls -A ./data/redis 2>/dev/null)" ]; then
        warning "检测到现有数据库数据"
        echo -e "${YELLOW}现有数据目录:${NC}"
        [ "$(ls -A ./data/postgres 2>/dev/null)" ] && echo "  - PostgreSQL: ./data/postgres/ (有数据)"
        [ "$(ls -A ./data/redis 2>/dev/null)" ] && echo "  - Redis: ./data/redis/ (有数据)"
        
        echo ""
        read -p "是否继续部署？现有数据将被保留 (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            info "部署已取消"
            exit 0
        fi
    fi
else
    info "创建数据目录..."
    mkdir -p ./data/postgres ./data/redis
    success "数据目录创建完成"
fi

# 设置目录权限
chmod -R 777 ./data/

# 3. 进入虚拟机执行部署
info "3. 进入虚拟机执行数据库部署..."

# 设置内部脚本可执行权限
multipass exec kube-master -- chmod +x /mnt/mem_bank/deploy/deploy-inside.sh

# 在虚拟机内部执行部署
info "正在虚拟机内部执行部署脚本..."
multipass exec kube-master -- /mnt/mem_bank/deploy/deploy-inside.sh

echo ""
success "部署完成！"
echo ""
echo -e "${GREEN}=== 连接信息 ===${NC}"
echo -e "${BLUE}PostgreSQL:${NC} localhost:30432"
echo "  用户名: mem_bank_user"
echo "  密码: mem_bank_password"
echo "  数据库: mem_bank"
echo -e "  连接命令: ${YELLOW}psql -h localhost -p 30432 -U mem_bank_user -d mem_bank${NC}"
echo ""
echo -e "${BLUE}Redis:${NC} localhost:30379"
echo -e "  连接命令: ${YELLOW}redis-cli -h localhost -p 30379${NC}"
echo ""
echo -e "${GREEN}数据存储位置:${NC} ./data/postgres/ 和 ./data/redis/"

# 9. 验证连接
info "9. 验证数据库连接..."
echo "正在测试数据库连接..."

# 测试 Redis 连接
if command -v redis-cli >/dev/null 2>&1; then
    if timeout 5 redis-cli -h localhost -p 30379 ping >/dev/null 2>&1; then
        success "Redis 连接测试成功"
    else
        warning "Redis 连接测试失败，请检查服务状态"
    fi
else
    info "未安装 redis-cli，跳过 Redis 连接测试"
fi

# 测试 PostgreSQL 连接
if command -v psql >/dev/null 2>&1; then
    if timeout 5 psql -h localhost -p 30432 -U mem_bank_user -d mem_bank -c "SELECT 1;" >/dev/null 2>&1; then
        success "PostgreSQL 连接测试成功"
    else
        warning "PostgreSQL 连接测试失败，请检查服务状态"
    fi
else
    info "未安装 psql，跳过 PostgreSQL 连接测试"
fi