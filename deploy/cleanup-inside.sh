#!/bin/bash

# 在 multipass 虚拟机内部执行的清理脚本

set -e

# 定义 namespace
NAMESPACE="mem-bank"

echo "在虚拟机内部开始清理数据库部署..."

# 1. 删除 Redis 资源
echo "1. 删除 Redis 资源..."
kubectl delete -f /mnt/mem_bank/deploy/redis.yaml --ignore-not-found=true

# 2. 删除 PostgreSQL 资源
echo "2. 删除 PostgreSQL 资源..."
kubectl delete -f /mnt/mem_bank/deploy/postgres.yaml --ignore-not-found=true

# 3. 可选：删除 namespace（会删除所有相关资源）
echo ""
read -p "是否删除整个 namespace '$NAMESPACE'？这会删除其中的所有资源 (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "删除 namespace $NAMESPACE..."
    kubectl delete namespace "$NAMESPACE" --ignore-not-found=true
fi

# 4. 等待资源删除完成
echo "4. 等待资源删除完成..."
sleep 10

# 5. 显示清理状态
echo "5. 清理完成！当前状态："
echo ""

# 检查 namespace 是否还存在
if kubectl get namespace "$NAMESPACE" >/dev/null 2>&1; then
    echo "=== Namespace $NAMESPACE 中的资源状态 ==="
    kubectl get pods,svc,pv,pvc -n "$NAMESPACE" 2>/dev/null || echo "Namespace $NAMESPACE 中暂无资源"
else
    echo "Namespace $NAMESPACE 已被删除"
fi

echo ""
echo "清理脚本执行完成！"