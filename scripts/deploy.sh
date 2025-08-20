#!/bin/bash

# AI Memory Bank 统一部署脚本
# 支持 Docker Compose (本地开发) 和 Kubernetes (生产环境)

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

# 显示使用说明
show_usage() {
    echo "AI Memory Bank 部署脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -m, --mode MODE        部署模式: docker 或 k8s (默认: docker)"
    echo "  -e, --env ENV          环境: dev, staging, prod (默认: dev)" 
    echo "  -c, --clean            清理现有部署"
    echo "  -h, --help             显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0                     # 使用 Docker Compose 部署开发环境"
    echo "  $0 -m k8s -e prod     # 使用 Kubernetes 部署生产环境"
    echo "  $0 -c                  # 清理现有部署"
}

# 检查依赖
check_dependencies() {
    local mode=$1
    
    if [ "$mode" = "docker" ]; then
        if ! command -v docker >/dev/null 2>&1; then
            error "Docker 未安装或未在 PATH 中"
            exit 1
        fi
        if ! command -v docker-compose >/dev/null 2>&1 && ! docker compose version >/dev/null 2>&1; then
            error "Docker Compose 未安装或未在 PATH 中"
            exit 1
        fi
    elif [ "$mode" = "k8s" ]; then
        if ! command -v kubectl >/dev/null 2>&1; then
            error "kubectl 未安装或未在 PATH 中"
            exit 1
        fi
        # 检查 Kubernetes 集群连接
        if ! kubectl cluster-info >/dev/null 2>&1; then
            error "无法连接到 Kubernetes 集群"
            exit 1
        fi
    fi
}

# 准备数据目录
prepare_directories() {
    info "准备数据目录..."
    mkdir -p ./data/postgres ./data/redis
    chmod -R 755 ./data/
    success "数据目录准备完成"
}

# Docker Compose 部署
deploy_docker() {
    local env=$1
    
    info "使用 Docker Compose 部署 ($env 环境)..."
    
    # 准备数据目录
    prepare_directories
    
    # 检查现有容器
    if docker-compose ps -q >/dev/null 2>&1; then
        warning "检测到现有 Docker Compose 服务"
        echo -n "是否先停止现有服务？(y/N): "
        read -r response
        if [[ "$response" =~ ^[Yy]$ ]]; then
            info "停止现有服务..."
            docker-compose down
        fi
    fi
    
    # 构建并启动服务
    info "启动服务..."
    docker-compose up -d --build
    
    # 等待服务启动
    info "等待服务启动..."
    sleep 10
    
    # 检查服务状态
    if docker-compose ps | grep -q "Up"; then
        success "服务启动成功！"
        echo ""
        echo -e "${GREEN}=== 连接信息 ===${NC}"
        echo -e "${BLUE}PostgreSQL:${NC} localhost:5432"
        echo "  用户名: mem_bank_user"
        echo "  密码: mem_bank_password"
        echo "  数据库: mem_bank"
        echo ""
        echo -e "${BLUE}Redis:${NC} localhost:6379"
        echo ""
        echo -e "${BLUE}API:${NC} http://localhost:8080"
        echo "  健康检查: http://localhost:8080/api/v1/health"
    else
        error "服务启动失败，请检查日志: docker-compose logs"
        exit 1
    fi
}

# Kubernetes 部署
deploy_k8s() {
    local env=$1
    
    info "使用 Kubernetes 部署 ($env 环境)..."
    
    # 创建命名空间
    info "创建命名空间..."
    kubectl create namespace mem-bank --dry-run=client -o yaml | kubectl apply -f -
    
    # 应用配置
    info "部署数据库服务..."
    kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: mem-bank-config
  namespace: mem-bank
data:
  POSTGRES_DB: "mem_bank"
  POSTGRES_USER: "mem_bank_user"
  REDIS_HOST: "redis-service"
  REDIS_PORT: "6379"
---
apiVersion: v1
kind: Secret
metadata:
  name: mem-bank-secret
  namespace: mem-bank
type: Opaque
stringData:
  POSTGRES_PASSWORD: "mem_bank_password"
  JWT_SECRET: "your-jwt-secret-key-change-in-production"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
  namespace: mem-bank
spec:
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: pgvector/pgvector:pg15
        ports:
        - containerPort: 5432
        env:
        - name: POSTGRES_DB
          valueFrom:
            configMapKeyRef:
              name: mem-bank-config
              key: POSTGRES_DB
        - name: POSTGRES_USER
          valueFrom:
            configMapKeyRef:
              name: mem-bank-config
              key: POSTGRES_USER
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: mem-bank-secret
              key: POSTGRES_PASSWORD
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
  namespace: mem-bank
spec:
  type: NodePort
  ports:
  - port: 5432
    targetPort: 5432
    nodePort: 30432
  selector:
    app: postgres
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: mem-bank
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        ports:
        - containerPort: 6379
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
---
apiVersion: v1
kind: Service
metadata:
  name: redis-service
  namespace: mem-bank
spec:
  type: NodePort
  ports:
  - port: 6379
    targetPort: 6379
    nodePort: 30379
  selector:
    app: redis
EOF
    
    # 等待部署完成
    info "等待服务启动..."
    kubectl wait --for=condition=ready pod -l app=postgres -n mem-bank --timeout=300s
    kubectl wait --for=condition=ready pod -l app=redis -n mem-bank --timeout=300s
    
    success "Kubernetes 部署完成！"
    echo ""
    echo -e "${GREEN}=== 连接信息 ===${NC}"
    echo -e "${BLUE}PostgreSQL:${NC} localhost:30432"
    echo "  用户名: mem_bank_user"
    echo "  密码: mem_bank_password"
    echo "  数据库: mem_bank"
    echo ""
    echo -e "${BLUE}Redis:${NC} localhost:30379"
    echo ""
    echo "检查状态: kubectl get pods -n mem-bank"
}

# 清理部署
cleanup() {
    local mode=$1
    
    info "清理现有部署..."
    
    if [ "$mode" = "docker" ] || [ "$mode" = "all" ]; then
        if command -v docker-compose >/dev/null 2>&1; then
            info "停止 Docker Compose 服务..."
            docker-compose down -v 2>/dev/null || true
            success "Docker Compose 服务已停止"
        fi
    fi
    
    if [ "$mode" = "k8s" ] || [ "$mode" = "all" ]; then
        if command -v kubectl >/dev/null 2>&1; then
            info "清理 Kubernetes 资源..."
            kubectl delete namespace mem-bank --ignore-not-found=true
            success "Kubernetes 资源已清理"
        fi
    fi
    
    warning "数据文件仍保留在 ./data/ 目录中"
    echo -n "是否删除数据文件？(y/N): "
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        rm -rf ./data/
        success "数据文件已删除"
    fi
}

# 主函数
main() {
    local mode="docker"
    local env="dev"
    local clean_flag=false
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -m|--mode)
                mode="$2"
                shift 2
                ;;
            -e|--env)
                env="$2"
                shift 2
                ;;
            -c|--clean)
                clean_flag=true
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                error "未知参数: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    # 验证参数
    if [ "$mode" != "docker" ] && [ "$mode" != "k8s" ]; then
        error "无效的部署模式: $mode (支持: docker, k8s)"
        exit 1
    fi
    
    if [ "$env" != "dev" ] && [ "$env" != "staging" ] && [ "$env" != "prod" ]; then
        error "无效的环境: $env (支持: dev, staging, prod)"
        exit 1
    fi
    
    # 执行清理
    if [ "$clean_flag" = true ]; then
        cleanup "all"
        exit 0
    fi
    
    # 检查依赖
    check_dependencies "$mode"
    
    # 执行部署
    case $mode in
        docker)
            deploy_docker "$env"
            ;;
        k8s)
            deploy_k8s "$env"
            ;;
    esac
}

# 检查是否为 root 用户
if [ "$EUID" -eq 0 ]; then
    warning "不建议以 root 用户运行此脚本"
fi

# 检查是否在项目根目录
if [ ! -f "docker-compose.yml" ]; then
    error "请在项目根目录运行此脚本"
    exit 1
fi

# 运行主函数
main "$@"