// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';
import {FormattedMessage} from 'react-intl';

import './post_read_indicator.scss';

type Props = {
    postId: string;
    readCount?: number;
    onClick?: () => void;
};

export default class PostReadIndicator extends React.PureComponent<Props> {
    render() {
        const {readCount, onClick} = this.props;

        if (!readCount || readCount === 0) {
            return null;
        }

        return (
            <button
                className='post-read-indicator'
                onClick={onClick}
                aria-label='View read receipts'
            >
                <i className='icon icon-check-all'/>
                <span className='read-count'>
                    {readCount === 1 ? (
                        <FormattedMessage
                            id='post.read_indicator.one'
                            defaultMessage='1 read'
                        />
                    ) : (
                        <FormattedMessage
                            id='post.read_indicator.many'
                            defaultMessage='{count} read'
                            values={{count: readCount}}
                        />
                    )}
                </span>
            </button>
        );
    }
}
