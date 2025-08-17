# Multipass + Kubernetes 数据库部署指南

## 背景说明

在 multipass 虚拟机环境中运行 Kubernetes 集群，需要部署 PostgreSQL 和 Redis 数据库服务，要求满足：

1. **数据持久化**：数据库数据不能因容器重启而丢失
2. **宿主机访问**：从宿主机能直接连接数据库服务
3. **数据同步**：数据库文件需要同步到宿主机的 ./data 目录便于管理

## 架构设计

```
宿主机 (./data) 
    ↓ multipass mount 挂载
multipass 虚拟机 (/mnt/mem_bank/data)
    ↓ hostPath 映射
Kubernetes 持久卷
    ↓ 
数据库 Pod (PostgreSQL/Redis)
    ↓ NodePort 服务暴露
宿主机访问 (localhost:30432, localhost:30379)
```

## 为什么这样设计

### 1. 使用 multipass mount 挂载
- **实时同步**：虚拟机和宿主机文件实时同步
- **备份方便**：数据直接存在宿主机，便于备份和版本管理
- **开发效率**：可以直接在宿主机查看和管理数据库文件

### 2. 使用 hostPath 持久卷
- **数据持久化**：Pod 重启后数据不丢失
- **性能优势**：本地存储比网络存储性能更好
- **配置简单**：不需要复杂的存储集群

### 3. 使用 NodePort 服务
- **外部访问**：允许从虚拟机外部访问数据库
- **端口固定**：提供固定端口便于应用连接

## 操作步骤

### 第一步：挂载项目目录到虚拟机

```bash
# 在宿主机项目根目录执行
multipass mount . kube-master:/mnt/mem_bank
```

**作用**：将当前项目目录挂载到虚拟机的 /mnt/mem_bank 路径

### 第二步：在虚拟机中准备数据目录

```bash
# 进入虚拟机
multipass shell kube-master

# 创建数据目录
sudo mkdir -p /mnt/mem_bank/data/postgres /mnt/mem_bank/data/redis

# 设置权限（数据库容器用户ID通常是999）
sudo chown -R 999:999 /mnt/mem_bank/data/postgres
sudo chown -R 999:999 /mnt/mem_bank/data/redis

# 验证挂载是否成功
ls -la /mnt/mem_bank/
```

### 第三步：创建 Kubernetes 部署文件

#### PostgreSQL 部署 (k8s/postgres.yaml)

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: postgres-pv
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: local-storage
  hostPath:
    path: /mnt/mem_bank/data/postgres
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: local-storage
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
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
          value: "mem_bank"
        - name: POSTGRES_USER
          value: "mem_bank_user"
        - name: POSTGRES_PASSWORD
          value: "mem_bank_password"
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
      volumes:
      - name: postgres-storage
        persistentVolumeClaim:
          claimName: postgres-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
spec:
  type: NodePort
  ports:
  - port: 5432
    targetPort: 5432
    nodePort: 30432
  selector:
    app: postgres
```

#### Redis 部署 (k8s/redis.yaml)

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: redis-pv
spec:
  capacity:
    storage: 5Gi
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: local-storage
  hostPath:
    path: /mnt/mem_bank/data/redis
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: redis-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
  storageClassName: local-storage
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
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
        volumeMounts:
        - name: redis-storage
          mountPath: /data
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "200m"
      volumes:
      - name: redis-storage
        persistentVolumeClaim:
          claimName: redis-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: redis-service
spec:
  type: NodePort
  ports:
  - port: 6379
    targetPort: 6379
    nodePort: 30379
  selector:
    app: redis
```

### 第四步：部署到 Kubernetes

```bash
# 在虚拟机中执行
kubectl apply -f /mnt/mem_bank/k8s/postgres.yaml
kubectl apply -f /mnt/mem_bank/k8s/redis.yaml

# 检查部署状态
kubectl get pods
kubectl get services
kubectl get pv,pvc
```

## 最终效果

### 1. 数据持久化
- PostgreSQL 数据：宿主机 ./data/postgres/ ↔ 虚拟机 /mnt/mem_bank/data/postgres/
- Redis 数据：宿主机 ./data/redis/ ↔ 虚拟机 /mnt/mem_bank/data/redis/

### 2. 宿主机访问
- **PostgreSQL**：localhost:30432
  ```bash
  psql -h localhost -p 30432 -U mem_bank_user -d mem_bank
  ```
- **Redis**：localhost:30379
  ```bash
  redis-cli -h localhost -p 30379
  ```

### 3. 数据同步
- 数据库文件实时同步到宿主机 ./data 目录
- 可在宿主机直接备份、查看、管理数据
- 支持版本控制

## 验证方法

```bash
# 1. 检查挂载状态
multipass info kube-master

# 2. 检查 Pod 状态
multipass exec kube-master -- kubectl get pods

# 3. 测试数据库连接
psql -h localhost -p 30432 -U mem_bank_user -d mem_bank
redis-cli -h localhost -p 30379

# 4. 验证数据同步
echo "test" > ./data/test.txt
multipass exec kube-master -- cat /mnt/mem_bank/data/test.txt
```

## 故障排除

### 权限问题
```bash
multipass exec kube-master -- sudo chown -R 999:999 /mnt/mem_bank/data/postgres
multipass exec kube-master -- sudo chown -R 999:999 /mnt/mem_bank/data/redis
```

### 挂载问题
```bash
multipass umount kube-master:/mnt/mem_bank
multipass mount . kube-master:/mnt/mem_bank
```

### 网络访问问题
```bash
multipass list
multipass exec kube-master -- kubectl get svc
```

## 方案优势

1. **开发友好**：数据文件在宿主机可见，便于调试
2. **生产就绪**：使用标准 K8s 资源，易于迁移
3. **数据安全**：数据持久化在宿主机，虚拟机重建不丢失
4. **网络隔离**：数据库在虚拟机中运行，通过固定端口访问
5. **资源控制**：通过 K8s 限制控制数据库资源使用