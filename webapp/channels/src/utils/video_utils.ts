// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

export type VideoProvider = 'youtube' | 'vimeo' | 'dailymotion' | 'twitch' | 'bilibili' | 'tiktok' | 'instagram' | 'facebook' | 'twitter' | 'streamable' | 'youku' | 'iqiyi' | 'tencent' | 'generic';

export interface VideoInfo {
    provider: VideoProvider;
    videoId?: string;
    embedUrl?: string;
    originalUrl: string;
}

/**
 * Detects if a URL is a video link and returns information about it
 */
export function getVideoInfo(url: string): VideoInfo | null {
    if (!url) {
        return null;
    }

    // YouTube detection
    const youtubeMatch = url.match(/(?:youtube\.com\/(?:[^/]+\/.+\/|(?:v|e(?:mbed)?)\/|.*[?&]v=)|youtu\.be\/)([^"&?/\s]{11})/i);
    if (youtubeMatch) {
        return {
            provider: 'youtube',
            videoId: youtubeMatch[1],
            embedUrl: `https://www.youtube.com/embed/${youtubeMatch[1]}`,
            originalUrl: url,
        };
    }

    // Vimeo detection
    const vimeoMatch = url.match(/vimeo\.com\/(?:video\/)?(\d+)/i);
    if (vimeoMatch) {
        return {
            provider: 'vimeo',
            videoId: vimeoMatch[1],
            embedUrl: `https://player.vimeo.com/video/${vimeoMatch[1]}`,
            originalUrl: url,
        };
    }

    // Dailymotion detection
    const dailymotionMatch = url.match(/dailymotion\.com\/(?:video|hub)\/([^_\s]+)/i);
    if (dailymotionMatch) {
        return {
            provider: 'dailymotion',
            videoId: dailymotionMatch[1],
            embedUrl: `https://www.dailymotion.com/embed/video/${dailymotionMatch[1]}`,
            originalUrl: url,
        };
    }

    // Twitch detection (videos and clips)
    const twitchVideoMatch = url.match(/twitch\.tv\/videos\/(\d+)/i);
    if (twitchVideoMatch) {
        return {
            provider: 'twitch',
            videoId: twitchVideoMatch[1],
            embedUrl: `https://player.twitch.tv/?video=${twitchVideoMatch[1]}&parent=${window.location.hostname}`,
            originalUrl: url,
        };
    }

    const twitchClipMatch = url.match(/twitch\.tv\/\w+\/clip\/(\w+)/i);
    if (twitchClipMatch) {
        return {
            provider: 'twitch',
            videoId: twitchClipMatch[1],
            embedUrl: `https://clips.twitch.tv/embed?clip=${twitchClipMatch[1]}&parent=${window.location.hostname}`,
            originalUrl: url,
        };
    }

    // Bilibili detection (支持多种格式)
    // BV号格式: https://www.bilibili.com/video/BV1xx411c7XZ
    // av号格式: https://www.bilibili.com/video/av12345678
    const bilibiliMatch = url.match(/bilibili\.com\/video\/((?:BV|av)[\w]+)/i);
    if (bilibiliMatch) {
        const videoId = bilibiliMatch[1];
        const isAv = videoId.startsWith('av');
        const paramName = isAv ? 'aid' : 'bvid';
        const paramValue = isAv ? videoId.substring(2) : videoId;
        
        return {
            provider: 'bilibili',
            videoId,
            embedUrl: `https://player.bilibili.com/player.html?${paramName}=${paramValue}&high_quality=1&danmaku=0&autoplay=0`,
            originalUrl: url,
        };
    }

    // Bilibili 短链接: https://b23.tv/xxxxx
    const bilibiliShortMatch = url.match(/b23\.tv\/(\w+)/i);
    if (bilibiliShortMatch) {
        // Note: 短链接需要重定向，这里返回原链接，让后端处理
        return {
            provider: 'bilibili',
            videoId: bilibiliShortMatch[1],
            embedUrl: url, // 需要后端解析
            originalUrl: url,
        };
    }

    // TikTok detection
    const tiktokMatch = url.match(/tiktok\.com\/@[\w.-]+\/video\/(\d+)/i);
    if (tiktokMatch) {
        return {
            provider: 'tiktok',
            videoId: tiktokMatch[1],
            embedUrl: `https://www.tiktok.com/embed/v2/${tiktokMatch[1]}`,
            originalUrl: url,
        };
    }

    // Instagram detection (video posts and reels)
    const instagramMatch = url.match(/instagram\.com\/(?:p|reel)\/([\w-]+)/i);
    if (instagramMatch) {
        return {
            provider: 'instagram',
            videoId: instagramMatch[1],
            embedUrl: `https://www.instagram.com/p/${instagramMatch[1]}/embed`,
            originalUrl: url,
        };
    }

    // Facebook video detection
    const facebookMatch = url.match(/facebook\.com\/.*\/videos\/(\d+)/i);
    if (facebookMatch) {
        return {
            provider: 'facebook',
            videoId: facebookMatch[1],
            embedUrl: `https://www.facebook.com/plugins/video.php?href=${encodeURIComponent(url)}`,
            originalUrl: url,
        };
    }

    // Twitter/X video detection
    const twitterMatch = url.match(/(?:twitter\.com|x\.com)\/\w+\/status\/(\d+)/i);
    if (twitterMatch) {
        return {
            provider: 'twitter',
            videoId: twitterMatch[1],
            embedUrl: `https://platform.twitter.com/embed/Tweet.html?id=${twitterMatch[1]}`,
            originalUrl: url,
        };
    }

    // Streamable detection
    const streamableMatch = url.match(/streamable\.com\/(\w+)/i);
    if (streamableMatch) {
        return {
            provider: 'streamable',
            videoId: streamableMatch[1],
            embedUrl: `https://streamable.com/e/${streamableMatch[1]}`,
            originalUrl: url,
        };
    }

    // Youku detection (优酷)
    const youkuMatch = url.match(/youku\.com\/.*(?:id_|vid=)([\w=]+)/i);
    if (youkuMatch) {
        return {
            provider: 'youku',
            videoId: youkuMatch[1],
            embedUrl: `https://player.youku.com/embed/${youkuMatch[1]}`,
            originalUrl: url,
        };
    }

    // iQiyi detection (爱奇艺)
    const iqiyiMatch = url.match(/iqiyi\.com\/.*[\/_]([\w]+)\.html/i);
    if (iqiyiMatch) {
        return {
            provider: 'iqiyi',
            videoId: iqiyiMatch[1],
            embedUrl: `https://www.iqiyi.com/common/flashplayer/20150916/share_player.html?vid=${iqiyiMatch[1]}`,
            originalUrl: url,
        };
    }

    // Tencent Video detection (腾讯视频)
    const tencentMatch = url.match(/v\.qq\.com\/.*\/([a-z0-9]+)\.html/i);
    if (tencentMatch) {
        return {
            provider: 'tencent',
            videoId: tencentMatch[1],
            embedUrl: `https://v.qq.com/txp/iframe/player.html?vid=${tencentMatch[1]}`,
            originalUrl: url,
        };
    }

    return null;
}

/**
 * Check if a URL is a video link
 */
export function isVideoLink(url: string): boolean {
    return getVideoInfo(url) !== null;
}

/**
 * Get embed dimensions for a video
 */
export function getVideoEmbedDimensions(aspectRatio: number = 16 / 9): {width: number; height: number} {
    const maxWidth = 560;
    const width = Math.min(maxWidth, window.innerWidth - 40);
    const height = Math.round(width / aspectRatio);

    return {width, height};
}

/**
 * Get friendly display name for video provider
 */
export function getProviderDisplayName(provider: VideoProvider): string {
    const displayNames: Record<VideoProvider, string> = {
        youtube: 'YouTube',
        vimeo: 'Vimeo',
        dailymotion: 'Dailymotion',
        twitch: 'Twitch',
        bilibili: '哔哩哔哩 (Bilibili)',
        tiktok: 'TikTok',
        instagram: 'Instagram',
        facebook: 'Facebook',
        twitter: 'Twitter/X',
        streamable: 'Streamable',
        youku: '优酷 (Youku)',
        iqiyi: '爱奇艺 (iQiyi)',
        tencent: '腾讯视频 (Tencent Video)',
        generic: 'Video',
    };

    return displayNames[provider] || 'Video';
}

/**
 * Check if provider requires special handling
 */
export function providerRequiresAuth(provider: VideoProvider): boolean {
    // Some platforms may require authentication to view embedded content
    return ['instagram', 'facebook'].includes(provider);
}
