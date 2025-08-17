#!/bin/bash

# Kubernetes 数据库清理脚本

set -e

echo "开始清理 Kubernetes 数据库部署..."

# 1. 检查 multipass 挂载
echo "1. 检查 multipass 挂载状态..."
if ! multipass exec kube-master -- test -d /mnt/mem_bank; then
    echo "项目目录未挂载，正在挂载..."
    multipass mount . kube-master:/mnt/mem_bank
fi

# 2. 设置内部脚本可执行权限
multipass exec kube-master -- chmod +x /mnt/mem_bank/deploy/cleanup-inside.sh

# 3. 在虚拟机内部执行清理
echo "3. 在虚拟机内部执行清理脚本..."
multipass exec kube-master -- /mnt/mem_bank/deploy/cleanup-inside.sh

echo ""
echo "注意: 数据文件仍保留在 ./data/ 目录中"
echo "如需完全清理数据，请手动删除 ./data/postgres/ 和 ./data/redis/ 目录"