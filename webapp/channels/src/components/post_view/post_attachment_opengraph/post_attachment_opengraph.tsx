// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import classNames from 'classnames';
import React, {memo, useRef} from 'react';
import {useIntl} from 'react-intl';

import {CloseIcon, MenuDownIcon, MenuRightIcon, AccountOutlineIcon, CalendarOutlineIcon} from '@mattermost/compass-icons/components';
import type {
    OpenGraphMetadata,
    OpenGraphMetadataImage,
    OpenGraphMetadataVideo,
    OpenGraphMetadataAudio,
    Post,
    PostImage,
} from '@mattermost/types/posts';

import AutoHeightSwitcher from 'components/common/auto_height_switcher';
import ExternalImage from 'components/external_image';
import ExternalLink from 'components/external_link';
import VideoEmbed from 'components/video_embed';
import WithTooltip from 'components/with_tooltip';

import {PostTypes} from 'utils/constants';
import {isSystemMessage} from 'utils/post_utils';
import {makeUrlSafe} from 'utils/url';
import {isVideoLink} from 'utils/video_utils';

import {getNearestPoint} from './get_nearest_point';

import './post_attachment_opengraph.scss';

const DIMENSIONS_NEAREST_POINT_IMAGE = {
    height: 80,
    width: 80,
};

const LARGE_IMAGE_RATIO = 4 / 3;
const LARGE_IMAGE_WIDTH = 150;

export type Props = {
    postId: string;
    link: string;
    currentUserId?: string;
    post: Post;
    openGraphData?: OpenGraphMetadata;
    enableLinkPreviews?: boolean;
    previewEnabled?: boolean;
    isEmbedVisible?: boolean;
    toggleEmbedVisibility: () => void;
    actions: {
        editPost: (post: { id: string; props: Record<string, any> }) => void;
    };
    isInPermalink?: boolean;
    imageCollapsed?: boolean;
};

type ImageMetadata = Partial<OpenGraphMetadataImage> & PostImage;

export function getBestImage(openGraphData?: OpenGraphMetadata, imagesMetadata?: Record<string, PostImage>) {
    if (!openGraphData?.images?.length) {
        return null;
    }

    // Get the dimensions from the post metadata if they weren't provided by the website as part of the OpenGraph data
    const images = openGraphData.images.map((image: OpenGraphMetadataImage) => {
        const imageUrl = image.secure_url || image.url;

        return {
            ...image,
            height: image.height || imagesMetadata?.[imageUrl]?.height || -1,
            width: image.width || imagesMetadata?.[imageUrl]?.width || -1,
            format: image.type?.split('/')[1] || image.type || '',
            frameCount: 0,
        };
    });

    return getNearestPoint<ImageMetadata>(DIMENSIONS_NEAREST_POINT_IMAGE, images);
}

export const getIsLargeImage = (data: ImageMetadata|null) => {
    if (!data) {
        return false;
    }

    const {height, width} = data;

    return width >= LARGE_IMAGE_WIDTH && (width / height) >= LARGE_IMAGE_RATIO;
};

const PostAttachmentOpenGraph = ({openGraphData, post, actions, link, isInPermalink, previewEnabled, ...rest}: Props) => {
    const {formatMessage} = useIntl();
    const {current: bestImageData} = useRef<ImageMetadata>(getBestImage(openGraphData, post.metadata.images));
    const isPreviewRemoved = post?.props?.[PostTypes.REMOVE_LINK_PREVIEW] === 'true';

    // block of early return statements
    if (!rest.enableLinkPreviews || !previewEnabled || isPreviewRemoved) {
        return null;
    }

    if (!post || isSystemMessage(post)) {
        return null;
    }

    if (!openGraphData) {
        return null;
    }

    const handleRemovePreview = async (e: React.MouseEvent<HTMLButtonElement>) => {
        e.preventDefault();

        // prevent the button-click to trigger visiting the link
        e.stopPropagation();
        const props = Object.assign({}, post.props);
        props[PostTypes.REMOVE_LINK_PREVIEW] = 'true';

        const patchedPost = {
            id: post.id,
            props,
        };

        return actions.editPost(patchedPost);
    };

    const safeLink = makeUrlSafe(openGraphData?.url || link);

    return (
        <ExternalLink
            className='PostAttachmentOpenGraph'
            role='link'
            href={safeLink}
            title={openGraphData?.title || openGraphData?.url || link}
            location='post_attachment_opengraph'
        >
            {rest.currentUserId === post.user_id && !isInPermalink && (
                <WithTooltip
                    title={formatMessage({id: 'link_preview.remove_link_preview', defaultMessage: 'Remove link preview'})}
                >
                    <button
                        type='button'
                        className='remove-button style--none'
                        aria-label='Remove'
                        onClick={handleRemovePreview}
                        data-testid='removeLinkPreviewButton'
                    >
                        <CloseIcon
                            size={14}
                            color={'currentColor'}
                        />
                    </button>
                </WithTooltip>
            )}
            <PostAttachmentOpenGraphBody
                isInPermalink={isInPermalink}
                sitename={openGraphData?.site_name}
                title={openGraphData?.title || openGraphData?.url || link}
                description={openGraphData?.description}
                author={openGraphData?.article?.author}
                publishedTime={openGraphData?.article?.published_time}
            />
            <PostAttachmentOpenGraphImage
                imageMetadata={bestImageData}
                title={openGraphData?.title}
                isInPermalink={isInPermalink}
                isEmbedVisible={rest.isEmbedVisible}
                toggleEmbedVisibility={rest.toggleEmbedVisibility}
            />
            {/* Check URL directly for video links since backend strips video metadata */}
            <PostAttachmentOpenGraphVideo
                url={openGraphData?.url || link}
                isInPermalink={isInPermalink}
                isEmbedVisible={rest.isEmbedVisible}
            />
            <PostAttachmentOpenGraphAudio
                audioMetadata={openGraphData?.audios?.[0]}
                isInPermalink={isInPermalink}
            />
        </ExternalLink>
    );
};

type BodyProps = {
    title: string;
    isInPermalink?: boolean;
    sitename?: string;
    description?: string;
    author?: string;
    publishedTime?: string;
}

export const PostAttachmentOpenGraphBody = memo(({title, isInPermalink, sitename = '', description = '', author, publishedTime}: BodyProps) => {
    const formatPublishedTime = (time?: string) => {
        if (!time) return null;
        try {
            const date = new Date(time);
            return date.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
        } catch {
            return null;
        }
    };

    return title ? (
        <div className={classNames('PostAttachmentOpenGraph__body', {isInPermalink})}>
            {(!isInPermalink && sitename) && <span className='sitename'>{sitename}</span>}
            <span className='title'>{title}</span>
            {description && <span className='description'>{description}</span>}
            {(author || publishedTime) && (
                <div className='metadata'>
                    {author && (
                        <span className='author'>
                            <AccountOutlineIcon size={14}/>
                            {author}
                        </span>
                    )}
                    {publishedTime && (
                        <span className='published-time'>
                            <CalendarOutlineIcon size={14}/>
                            {formatPublishedTime(publishedTime)}
                        </span>
                    )}
                </div>
            )}
        </div>
    ) : null;
});

type ImageProps = {
    title?: string;
    imageMetadata?: ImageMetadata|null;
    isInPermalink: Props['isInPermalink'];
    isEmbedVisible: Props['isEmbedVisible'];
    toggleEmbedVisibility: Props['toggleEmbedVisibility'];
}

export const PostAttachmentOpenGraphImage = memo(({imageMetadata, isInPermalink, toggleEmbedVisibility, isEmbedVisible = true, title = ''}: ImageProps) => {
    const {formatMessage} = useIntl();

    if (!imageMetadata || isInPermalink) {
        return null;
    }

    const large = getIsLargeImage(imageMetadata);
    const src = imageMetadata.secure_url || imageMetadata.url || '';

    const toggleImagePreview = (e: React.MouseEvent<HTMLButtonElement>) => {
        e.preventDefault();

        // prevent the button-click to trigger visiting the link
        e.stopPropagation();
        toggleEmbedVisibility();
    };

    const collapsedLabel = formatMessage({id: 'link_preview.image_preview', defaultMessage: 'Show image preview'});

    const imageCollapseButton = (
        <button
            className='preview-toggle style--none'
            onClick={toggleImagePreview}
        >
            {isEmbedVisible ? (
                <MenuDownIcon
                    size={18}
                    color='currentColor'
                />
            ) : (
                <>
                    <MenuRightIcon
                        size={18}
                        color='currentColor'
                    />
                    {collapsedLabel}
                </>
            )}
        </button>
    );

    const image = (
        <ExternalImage
            src={src}
            imageMetadata={imageMetadata}
        >
            {(source) => (
                <>
                    {large && imageCollapseButton}
                    <figure>
                        <img
                            src={source}
                            alt={title}
                        />
                    </figure>
                </>
            )}
        </ExternalImage>
    );

    return (
        <div className={classNames('PostAttachmentOpenGraph__image', {large, collapsed: !isEmbedVisible})}>
            {large ? (
                <AutoHeightSwitcher
                    showSlot={isEmbedVisible ? 1 : 2}
                    slot1={image}
                    slot2={imageCollapseButton}
                />
            ) : image}
        </div>
    );
});

type VideoProps = {
    url: string;
    isInPermalink?: boolean;
    isEmbedVisible?: boolean;
}

export const PostAttachmentOpenGraphVideo = memo(({url, isInPermalink, isEmbedVisible = true}: VideoProps) => {
    if (!url || isInPermalink || !isEmbedVisible) {
        return null;
    }

    // Check if it's a supported video platform (YouTube, Vimeo, Bilibili, etc.)
    if (isVideoLink(url)) {
        return (
            <VideoEmbed
                url={url}
                show={isEmbedVisible}
            />
        );
    }

    // No video detected
    return null;
});

type AudioProps = {
    audioMetadata?: OpenGraphMetadataAudio;
    isInPermalink?: boolean;
}

export const PostAttachmentOpenGraphAudio = memo(({audioMetadata, isInPermalink}: AudioProps) => {
    if (!audioMetadata || isInPermalink) {
        return null;
    }

    const src = audioMetadata.secure_url || audioMetadata.url;
    const type = audioMetadata.type || 'audio/mpeg';

    return (
        <div className='PostAttachmentOpenGraph__audio'>
            <audio
                controls
                preload='metadata'
            >
                <source src={src} type={type}/>
                Your browser does not support the audio tag.
            </audio>
        </div>
    );
});

export default PostAttachmentOpenGraph;
