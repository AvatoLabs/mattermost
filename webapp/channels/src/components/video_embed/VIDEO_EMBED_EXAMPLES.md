# 视频嵌入示例

## 测试链接

### 🇨🇳 中国平台

#### Bilibili (哔哩哔哩)
```
https://www.bilibili.com/video/BV1xx411c7XZ
https://www.bilibili.com/video/av170001
https://b23.tv/abc123
```

#### Youku (优酷)
```
https://v.youku.com/v_show/id_XMzI1NjY4NzQ4MA==.html
```

#### iQiyi (爱奇艺)
```
https://www.iqiyi.com/v_19rr8kn0eo.html
```

#### Tencent Video (腾讯视频)
```
https://v.qq.com/x/page/a0123456789.html
```

### 🌍 国际平台

#### YouTube
```
https://www.youtube.com/watch?v=dQw4w9WgXcQ
https://youtu.be/dQw4w9WgXcQ
```

#### Vimeo
```
https://vimeo.com/123456789
```

#### TikTok
```
https://www.tiktok.com/@username/video/1234567890123456789
```

#### Instagram
```
https://www.instagram.com/p/ABC123def456/
https://www.instagram.com/reel/ABC123def456/
```

#### Twitter/X
```
https://twitter.com/user/status/1234567890123456789
https://x.com/user/status/1234567890123456789
```

#### Twitch
```
https://www.twitch.tv/videos/1234567890
https://www.twitch.tv/channel/clip/ClipName
```

#### Facebook
```
https://www.facebook.com/username/videos/1234567890
```

#### Streamable
```
https://streamable.com/abc123
```

## 功能特性

### ✅ 自动识别
- 粘贴任何支持的视频链接
- 自动检测平台类型
- 生成嵌入式播放器

### ✅ 响应式设计
- 自动适应屏幕宽度
- 保持 16:9 比例
- 移动端友好

### ✅ 平台特定优化
- **Bilibili**: 自动隐藏弹幕、高清画质、禁用自动播放
- **YouTube**: 标准嵌入播放器
- **Twitch**: 包含父域名参数以支持嵌入
- **TikTok**: 使用 v2 嵌入 API

### ✅ 用户体验
- 懒加载 iframe（性能优化）
- 平台品牌色彩
- 悬停效果
- 需要认证的平台会显示提示

### ✅ 安全性
- 所有外部链接使用 `rel="noopener noreferrer"`
- iframe 沙箱保护
- HTTPS 强制

## 使用方式

在 Mattermost 消息中直接粘贴视频链接：

```
用户: 看看这个视频 https://www.bilibili.com/video/BV1xx411c7XZ
```

系统会自动渲染为：

```
┌─────────────────────────────────────┐
│  [Bilibili 视频播放器]              │
│                                     │
│  ▶️ 视频内容                        │
│                                     │
└─────────────────────────────────────┘
```

## 故障排除

### Bilibili 短链接 (b23.tv)
- 短链接需要重定向解析
- 如果无法嵌入，会显示"在新标签页中打开"链接

### Instagram/Facebook
- 可能需要用户登录
- 会显示认证提示

### 企业防火墙
- 某些平台可能被企业防火墙拦截
- 建议在内网测试时配置白名单

## 开发者信息

### 添加新平台

1. 在 `video_utils.ts` 中添加平台类型
2. 在 `getVideoInfo()` 中添加 URL 匹配逻辑
3. 在 `getProviderDisplayName()` 中添加显示名称
4. （可选）在 `video_embed.scss` 中添加平台特定样式

### 测试

```typescript
import {getVideoInfo} from 'utils/video_utils';

const info = getVideoInfo('https://www.bilibili.com/video/BV1xx411c7XZ');
console.log(info);
// {
//   provider: 'bilibili',
//   videoId: 'BV1xx411c7XZ',
//   embedUrl: 'https://player.bilibili.com/player.html?bvid=BV1xx411c7XZ&high_quality=1&danmaku=0&autoplay=0',
//   originalUrl: 'https://www.bilibili.com/video/BV1xx411c7XZ'
// }
```
