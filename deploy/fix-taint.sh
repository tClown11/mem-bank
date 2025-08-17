#!/bin/bash

# 修复 control-plane 节点污点问题

echo "正在移除 control-plane 节点的污点..."

# 移除 control-plane 污点，允许在控制节点上调度 Pod
kubectl taint nodes --all node-role.kubernetes.io/control-plane-

echo "污点已移除，control-plane 节点现在可以运行工作负载了"

# 检查节点状态
echo ""
echo "=== 节点状态 ==="
kubectl get nodes -o wide

echo ""
echo "=== 节点污点信息 ==="
kubectl describe nodes | grep -i taint