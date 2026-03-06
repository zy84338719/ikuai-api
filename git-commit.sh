#!/bin/bash
set -e

cd /opt/project/ikuai-aio/sdk

echo "=== Git Status ==="
git status

echo ""
echo "=== Git Diff (stat) ==="
git diff --stat

echo ""
echo "=== Adding all changes ==="
git add -A

echo ""
echo "=== Git Status after add ==="
git status

echo ""
echo "=== Creating commit ==="
git commit -m "feat: 添加多个新服务实现和测试

- 添加 VPN 服务实现 (PPTP/L2TP 客户端管理)
- 添加日志服务实现 (多种日志查询功能)
- 添加 Docker 服务实现 (容器/镜像/网络管理)
- 添加虚拟机服务实现 (完整的 VM 生命周期管理)
- 添加 UPnP 服务实现 (UPnP 映射管理)
- 添加各服务的单元测试
- 改进错误处理和类型定义
- 完善文档和示例代码"

echo ""
echo "=== Pushing to remote ==="
git push origin main

echo ""
echo "=== Done! ==="
