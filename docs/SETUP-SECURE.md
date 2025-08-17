# Secure Setup Guide

This guide helps you set up the mem_bank project securely by avoiding hardcoded secrets and following security best practices.

## 🔐 Initial Security Setup

### 1. Environment Configuration

**For Local Development:**

```bash
# Copy the example environment file
cp .env.example .env

# Edit .env with your actual values
nano .env
```

**For Production:**

```bash
# Use the secure configuration file
cp configs/config.secure.yaml configs/config.yaml

# Set environment variables in your deployment system
export DB_PASSWORD="your_secure_password"
export JWT_SECRET="$(openssl rand -base64 32)"
```

### 2. Database Setup

**For Local Development:**

```bash
# Use the current postgres.yaml (already configured)
kubectl apply -f deploy/postgres.yaml
kubectl apply -f deploy/redis.yaml
```

**For Production:**

```bash
# Create secrets first
cp deploy/postgres-secret.example.yaml deploy/postgres-secret.yaml

# Edit secrets with base64 encoded values
# Generate secure passwords:
echo -n "your_secure_db_password" | base64

# Apply secrets and secure deployment
kubectl apply -f deploy/postgres-secret.yaml
kubectl apply -f deploy/postgres-secure.example.yaml
kubectl apply -f deploy/redis.yaml
```

### 3. Application Deployment

**For Production:**

```bash
# Create application secrets
cp deploy/app-secret.example.yaml deploy/app-secret.yaml

# Generate and encode JWT secret
JWT_SECRET=$(openssl rand -base64 32)
echo -n "$JWT_SECRET" | base64

# Edit app-secret.yaml with the encoded values
# Then deploy:
kubectl apply -f deploy/app-secret.yaml
kubectl apply -f deploy/app-deployment.example.yaml
```

## 🛡️ Security Checklist

### Before Deployment

- [ ] All secrets stored in environment variables or Kubernetes secrets
- [ ] No hardcoded passwords in configuration files
- [ ] SSL/TLS enabled for database connections (production)
- [ ] Strong, unique passwords generated
- [ ] JWT secret is cryptographically secure (min 32 chars)
- [ ] Resource limits configured for containers
- [ ] Health checks implemented
- [ ] Network policies applied (production)

### File Security

```bash
# Check for accidentally committed secrets
./scripts/check-secrets.sh

# Install pre-commit hooks (recommended)
pip install pre-commit
pre-commit install
```

### Access Security

- [ ] Database accessible only from application pods
- [ ] NodePort services disabled in production
- [ ] RBAC configured for service accounts
- [ ] Container runs as non-root user
- [ ] Read-only root filesystem where possible

## 🔧 Configuration Management

### Environment Variables Priority

1. **System environment variables** (highest priority)
2. **Kubernetes secrets/configmaps**
3. **Configuration files**
4. **Default values** (lowest priority)

### Secure Configuration Pattern

```yaml
# ✅ Good: Uses environment variable with fallback
password: ${DB_PASSWORD:}

# ❌ Bad: Hardcoded password
password: "hardcoded_password"

# ✅ Good: Environment variable with secure default
jwt_secret: ${JWT_SECRET:}

# ❌ Bad: Weak or exposed secret
jwt_secret: "weak_secret"
```

## 🚨 Security Monitoring

### Logging Security Events

The application logs:
- Authentication attempts
- Failed database connections
- Invalid JWT tokens
- Rate limit violations

### Regular Security Tasks

```bash
# Update dependencies
go mod tidy

# Check for vulnerabilities
go list -json -m all | nancy sleuth

# Rotate secrets (quarterly)
# 1. Generate new JWT secret
# 2. Update Kubernetes secret
# 3. Restart application pods
```

## 🆘 Emergency Procedures

### If Secrets Are Compromised

1. **Immediate Actions:**
   ```bash
   # Rotate JWT secret
   kubectl patch secret app-secret -p='{"data":{"JWT_SECRET":"'$(openssl rand -base64 32 | base64 -w 0)'"}}'
   
   # Restart pods to pick up new secret
   kubectl rollout restart deployment/mem-bank-api
   ```

2. **Change Database Passwords:**
   ```bash
   # Connect to database and change password
   # Update Kubernetes secret
   # Restart database and application
   ```

3. **Review Logs:**
   ```bash
   # Check for unauthorized access
   kubectl logs -l app=mem-bank-api --since=24h | grep -i "auth\|error\|fail"
   ```

## 📚 Additional Resources

- [Security Guidelines](./SECURITY.md)
- [Kubernetes Security Best Practices](https://kubernetes.io/docs/concepts/security/)
- [OWASP Application Security](https://owasp.org/www-project-top-ten/)
- [Go Security Checklist](https://github.com/securego/gosec)

## 🔍 Troubleshooting

### Common Issues

**Issue: Application can't connect to database**
```bash
# Check secret values
kubectl get secret postgres-secret -o yaml

# Verify environment variables in pod
kubectl exec -it deployment/mem-bank-api -- env | grep DB_
```

**Issue: JWT authentication failing**
```bash
# Verify JWT secret is set
kubectl exec -it deployment/mem-bank-api -- env | grep JWT_SECRET
```

**Issue: Pre-commit hooks failing**
```bash
# Update hooks
pre-commit autoupdate

# Run manually
pre-commit run --all-files
```