#!/bin/bash

# 在 multipass 虚拟机内部执行的部署脚本

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

# 定义 namespace
NAMESPACE="mem-bank"

info "在虚拟机内部开始数据库部署..."

# 1. 检查并修复 control-plane 污点
info "1. 检查 control-plane 节点污点..."
if kubectl describe nodes | grep -q "node-role.kubernetes.io/control-plane:NoSchedule"; then
    warning "检测到 control-plane 污点，正在移除..."
    kubectl taint nodes --all node-role.kubernetes.io/control-plane- || true
    success "control-plane 污点已移除"
else
    success "control-plane 节点可正常调度"
fi

# 2. 检查 namespace 是否存在
info "2. 检查 namespace 状态..."
if kubectl get namespace "$NAMESPACE" >/dev/null 2>&1; then
    success "Namespace $NAMESPACE 已存在"
else
    info "Namespace $NAMESPACE 不存在，将自动创建"
fi

# 3. 检查现有部署
info "3. 检查现有部署..."
POSTGRES_EXISTS=$(kubectl get deployment postgres -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l || echo "0")
REDIS_EXISTS=$(kubectl get deployment redis -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l || echo "0")

if [ "$POSTGRES_EXISTS" -gt 0 ] || [ "$REDIS_EXISTS" -gt 0 ]; then
    warning "检测到现有数据库部署:"
    [ "$POSTGRES_EXISTS" -gt 0 ] && echo "  - PostgreSQL 部署已存在"
    [ "$REDIS_EXISTS" -gt 0 ] && echo "  - Redis 部署已存在"
    
    echo ""
    read -p "是否重新部署？这将重启数据库服务 (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        info "跳过部署，显示当前状态..."
        kubectl get pods,svc,pv,pvc -n "$NAMESPACE" 2>/dev/null || echo "Namespace $NAMESPACE 中暂无资源"
        exit 0
    fi
fi

# 4. 部署 PostgreSQL
info "4. 部署 PostgreSQL..."
kubectl apply -f /mnt/mem_bank/deploy/postgres.yaml
success "PostgreSQL 部署配置已应用"

# 5. 部署 Redis
info "5. 部署 Redis..."
kubectl apply -f /mnt/mem_bank/deploy/redis.yaml
success "Redis 部署配置已应用"

# 6. 等待 Pod 启动
info "6. 等待 Pod 启动..."
info "等待 PostgreSQL Pod 就绪..."
if kubectl wait --for=condition=ready pod -l app=postgres -n "$NAMESPACE" --timeout=300s; then
    success "PostgreSQL Pod 已就绪"
else
    error "PostgreSQL Pod 启动超时"
fi

info "等待 Redis Pod 就绪..."
if kubectl wait --for=condition=ready pod -l app=redis -n "$NAMESPACE" --timeout=300s; then
    success "Redis Pod 已就绪"
else
    error "Redis Pod 启动超时"
fi

# 7. 显示部署状态
info "7. 部署完成！查看状态："
echo ""
echo "=== Pods 状态 ==="
kubectl get pods -n "$NAMESPACE"

echo ""
echo "=== Services 状态 ==="
kubectl get services -n "$NAMESPACE"

echo ""
echo "=== PV/PVC 状态 ==="
kubectl get pv,pvc -n "$NAMESPACE"

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
echo -e "${GREEN}数据存储位置:${NC} /mnt/mem_bank/data/postgres/ 和 /mnt/mem_bank/data/redis/"

info "部署脚本执行完成！"