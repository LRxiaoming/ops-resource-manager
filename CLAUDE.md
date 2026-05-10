# AI Project

Enterprise projects including ops-resource-manager and frontend services.

## Docker Deployment (Windows/WSL)

所有容器化服务必须通过 WSL Ubuntu root 部署：

```bash
wsl -e bash -c "echo '<password>' | su -c 'cd /mnt/d/ai_project/<project> && docker compose up -d'" root
```

**注意:** WSL 上 docker.sock 需要 root 权限，非 root 用户会遇到权限问题。