#!/bin/bash

# 修复 namespace 名称问题的脚本

echo "正在清理错误的资源..."

# 删除错误创建的 PV
kubectl delete pv postgres-pv --ignore-not-found=true
kubectl delete pv redis-pv --ignore-not-found=true

echo "错误资源已清理，现在可以重新部署了"
echo "请运行: ./deploy/deploy-inside.sh"