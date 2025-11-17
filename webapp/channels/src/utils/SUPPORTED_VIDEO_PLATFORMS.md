# 支持的视频平台列表

## 🌍 国际平台

### YouTube
- **标准链接**: `https://www.youtube.com/watch?v=VIDEO_ID`
- **短链接**: `https://youtu.be/VIDEO_ID`
- **嵌入链接**: `https://www.youtube.com/embed/VIDEO_ID`
- **示例**: `https://www.youtube.com/watch?v=dQw4w9WgXcQ`

### Vimeo
- **标准链接**: `https://vimeo.com/VIDEO_ID`
- **视频链接**: `https://vimeo.com/video/VIDEO_ID`
- **示例**: `https://vimeo.com/123456789`

### Dailymotion
- **视频链接**: `https://www.dailymotion.com/video/VIDEO_ID`
- **Hub链接**: `https://www.dailymotion.com/hub/VIDEO_ID`
- **示例**: `https://www.dailymotion.com/video/x8abcde`

### Twitch
- **视频回放**: `https://www.twitch.tv/videos/VIDEO_ID`
- **精彩片段**: `https://www.twitch.tv/CHANNEL/clip/CLIP_ID`
- **示例**: `https://www.twitch.tv/videos/1234567890`

### TikTok
- **视频链接**: `https://www.tiktok.com/@USERNAME/video/VIDEO_ID`
- **示例**: `https://www.tiktok.com/@user/video/1234567890123456789`

### Instagram
- **帖子**: `https://www.instagram.com/p/POST_ID`
- **Reels**: `https://www.instagram.com/reel/REEL_ID`
- **示例**: `https://www.instagram.com/p/ABC123def456/`

### Facebook
- **视频链接**: `https://www.facebook.com/.../videos/VIDEO_ID`
- **示例**: `https://www.facebook.com/username/videos/1234567890`

### Twitter/X
- **推文链接**: `https://twitter.com/USERNAME/status/TWEET_ID`
- **X链接**: `https://x.com/USERNAME/status/TWEET_ID`
- **示例**: `https://twitter.com/user/status/1234567890123456789`

### Streamable
- **视频链接**: `https://streamable.com/VIDEO_ID`
- **示例**: `https://streamable.com/abc123`

---

## 🇨🇳 中国平台

### Bilibili (哔哩哔哩)
- **BV号格式**: `https://www.bilibili.com/video/BV1xx411c7XZ`
- **av号格式**: `https://www.bilibili.com/video/av12345678`
- **短链接**: `https://b23.tv/xxxxx`
- **示例**: `https://www.bilibili.com/video/BV1xx411c7XZ`
- **特性**: 
  - 自动隐藏弹幕
  - 高清画质
  - 响应式播放器

### Youku (优酷)
- **视频链接**: `https://v.youku.com/v_show/id_VIDEO_ID.html`
- **示例**: `https://v.youku.com/v_show/id_XMzI1NjY4NzQ4MA==.html`

### iQiyi (爱奇艺)
- **视频链接**: `https://www.iqiyi.com/v_VIDEO_ID.html`
- **示例**: `https://www.iqiyi.com/v_19rr8kn0eo.html`

### Tencent Video (腾讯视频)
- **视频链接**: `https://v.qq.com/x/page/VIDEO_ID.html`
- **示例**: `https://v.qq.com/x/page/a0123456789.html`

---

## 🎯 使用方式

在 Mattermost 消息中直接粘贴以上任何平台的视频链接，系统会自动：

1. ✅ 识别视频平台
2. ✅ 生成嵌入式播放器
3. ✅ 显示视频预览
4. ✅ 支持直接播放（无需跳转）

## 🔧 技术实现

所有平台的检测和嵌入逻辑都在 `video_utils.ts` 中实现，使用正则表达式匹配 URL 模式，并生成对应的嵌入链接。

## 📝 添加新平台

如果需要添加新的视频平台支持，请在 `video_utils.ts` 的 `getVideoInfo()` 函数中添加相应的检测逻辑。

## ⚠️ 注意事项

- 某些平台（如 Instagram、Facebook）可能需要用户登录才能查看内容
- Bilibili 短链接（b23.tv）需要后端支持重定向解析
- 部分平台的嵌入功能可能受到地区限制
- 建议在企业内网环境中测试各平台的可访问性
