# 配置文件说明

## 文件结构

- `config.go` - Go 配置结构定义和加载逻辑
- `config.yaml` - 应用配置文件（支持环境变量替换）

## 配置方式

### 1. 环境变量 (推荐)

复制项目根目录的 `.env.example` 到 `.env`：

```bash
cp .env.example .env
```

编辑 `.env` 文件设置你的配置值。

### 2. 直接修改配置文件

你也可以直接编辑 `config.yaml` 文件中的默认值。

## 环境变量优先级

环境变量 > config.yaml 默认值

## 主要配置项

### 服务器配置
- `SERVER_HOST` - 服务器监听地址 
- `SERVER_PORT` - 服务器端口
- `SERVER_MODE` - 运行模式 (debug/release)

### 数据库配置  
- `DB_HOST`, `DB_PORT` - 数据库连接信息
- `DB_USER`, `DB_PASSWORD` - 数据库认证信息
- `DB_NAME` - 数据库名称

### 安全配置
- `JWT_SECRET` - JWT 签名密钥（生产环境必须修改）
- `BCRYPT_COST` - 密码哈希强度

详细说明请参考 `.env.example` 文件。