# 摇摆熊 · Coco Music

轻量、快速的 Apple Music 风格 Web 音乐播放器（PWA）。

**Author / Brand：摇摆熊**

- 前端：Vue 3 + Vite + TypeScript + Tailwind + PWA
- 后端：Go BFF（仅作站点接口聚合，具体媒体源由部署环境配置）
- 收藏 / 下载：浏览器本地（localStorage + IndexedDB），不经服务端存歌

演示：<https://music.52131415.xyz>

---

## 公开说明

本仓库是 **摇摆熊** 的 Coco Music 前端与站点服务骨架。

- 仓库内 **不包含** 私有媒体源、采集接口地址、密钥或内网 upstream 配置
- 部署时请在服务器本地通过环境变量 / `EnvironmentFile` 注入 `COCO_PLAY_BASE` 等配置
- 客户端下载与播放优先走第三方直链，避免把整轨音频从本站出站中转

## 功能

- 站友搜索榜（按本站搜索热度）
- 搜索 / 播放 / 滚动歌词
- 本地收藏与客户端下载（可离线回放已缓存曲目）

## 本地开发

```bash
# 1) 准备本地 env（不要提交）
cp .env.example .env
# 编辑 .env，填入你自己的上游基址

# 2) 后端
cd backend && set -a && source ../.env && set +a && go run ./cmd/server

# 3) 前端
cd frontend && npm install && npm run dev
```

## 环境变量（本地 / 服务器）

| 变量 | 说明 |
|------|------|
| `ADDR` | 监听地址，默认 `:18280` |
| `COCO_PLAY_BASE` | **必填**。私有上游 API 基址（仅部署机配置，勿写入公开仓库） |
| `UPSTREAM_PUBLIC` | 可选。上游的公网同源前缀（若需要） |
| `STATIC_DIR` | 前端 `dist` 目录 |
| `PUBLIC_ORIGIN` | 站点公网 origin |
| `DATA_DIR` | 运行时数据目录 |
| `CORS_ORIGINS` | 逗号分隔 |

示例文件：`.env.example`、`deploy/coco-music.service.example`。

## 许可与署名

© 摇摆熊. 个人项目，未经允许请勿将私有接口配置二次公开。
