// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {getVideoInfo, isVideoLink, getProviderDisplayName} from './video_utils';

describe('video_utils', () => {
    describe('getVideoInfo', () => {
        test('should detect Bilibili BV video', () => {
            const url = 'https://www.bilibili.com/video/BV1xx411c7XZ';
            const info = getVideoInfo(url);

            expect(info).not.toBeNull();
            expect(info?.provider).toBe('bilibili');
            expect(info?.videoId).toBe('BV1xx411c7XZ');
            expect(info?.embedUrl).toContain('player.bilibili.com');
            expect(info?.embedUrl).toContain('bvid=BV1xx411c7XZ');
        });

        test('should detect Bilibili av video', () => {
            const url = 'https://www.bilibili.com/video/av170001';
            const info = getVideoInfo(url);

            expect(info).not.toBeNull();
            expect(info?.provider).toBe('bilibili');
            expect(info?.videoId).toBe('av170001');
            expect(info?.embedUrl).toContain('aid=170001');
        });

        test('should detect YouTube video', () => {
            const url = 'https://www.youtube.com/watch?v=dQw4w9WgXcQ';
            const info = getVideoInfo(url);

            expect(info).not.toBeNull();
            expect(info?.provider).toBe('youtube');
            expect(info?.videoId).toBe('dQw4w9WgXcQ');
        });

        test('should detect Vimeo video', () => {
            const url = 'https://vimeo.com/123456789';
            const info = getVideoInfo(url);

            expect(info).not.toBeNull();
            expect(info?.provider).toBe('vimeo');
            expect(info?.videoId).toBe('123456789');
        });

        test('should detect TikTok video', () => {
            const url = 'https://www.tiktok.com/@user/video/1234567890';
            const info = getVideoInfo(url);

            expect(info).not.toBeNull();
            expect(info?.provider).toBe('tiktok');
        });

        test('should return null for non-video URL', () => {
            const url = 'https://example.com/page';
            const info = getVideoInfo(url);

            expect(info).toBeNull();
        });
    });

    describe('isVideoLink', () => {
        test('should return true for Bilibili link', () => {
            expect(isVideoLink('https://www.bilibili.com/video/BV1xx411c7XZ')).toBe(true);
        });

        test('should return false for regular link', () => {
            expect(isVideoLink('https://example.com')).toBe(false);
        });
    });

    describe('getProviderDisplayName', () => {
        test('should return correct display name for Bilibili', () => {
            expect(getProviderDisplayName('bilibili')).toBe('哔哩哔哩 (Bilibili)');
        });

        test('should return correct display name for YouTube', () => {
            expect(getProviderDisplayName('youtube')).toBe('YouTube');
        });
    });
});
