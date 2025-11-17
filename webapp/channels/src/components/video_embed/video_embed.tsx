// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {memo} from 'react';

import {getVideoInfo, getVideoEmbedDimensions, getProviderDisplayName, providerRequiresAuth} from 'utils/video_utils';

import './video_embed.scss';

export type Props = {
    url: string;
    show?: boolean;
    className?: string;
};

const VideoEmbed = ({url, show = true, className = ''}: Props) => {
    if (!show) {
        return null;
    }

    const videoInfo = getVideoInfo(url);
    if (!videoInfo || !videoInfo.embedUrl) {
        return null;
    }

    const {width, height} = getVideoEmbedDimensions();
    const providerName = getProviderDisplayName(videoInfo.provider);
    const requiresAuth = providerRequiresAuth(videoInfo.provider);

    // For Bilibili short links, show a message to open in new tab
    if (videoInfo.provider === 'bilibili' && videoInfo.embedUrl === url) {
        return (
            <div className={`VideoEmbed VideoEmbed--fallback ${className}`}>
                <a
                    href={url}
                    target='_blank'
                    rel='noopener noreferrer'
                    className='VideoEmbed__fallback-link'
                >
                    在新标签页中打开 Bilibili 视频
                </a>
            </div>
        );
    }

    return (
        <div className={`VideoEmbed VideoEmbed--${videoInfo.provider} ${className}`}>
            {requiresAuth && (
                <div className='VideoEmbed__auth-notice'>
                    此平台可能需要登录才能查看内容
                </div>
            )}
            <div
                className='VideoEmbed__wrapper'
                style={{
                    paddingBottom: `${(height / width) * 100}%`,
                }}
            >
                <iframe
                    className='VideoEmbed__iframe'
                    src={videoInfo.embedUrl}
                    width={width}
                    height={height}
                    frameBorder='0'
                    allow='accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture'
                    allowFullScreen
                    title={`${providerName} 视频`}
                    loading='lazy'
                />
            </div>
        </div>
    );
};

export default memo(VideoEmbed);
