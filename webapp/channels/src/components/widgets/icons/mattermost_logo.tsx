// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';
import {useIntl} from 'react-intl';

import unionLogo from 'images/union_logo.svg';

export default function MattermostLogo(props: React.HTMLAttributes<HTMLSpanElement>) {
    const {formatMessage} = useIntl();
    return (
        <span {...props}>
            <img
                src={unionLogo}
                alt={formatMessage({id: 'generic_icons.mattermost', defaultMessage: 'Mattermost Logo'})}
                style={{width: '32px', height: '32px'}}
            />
        </span>
    );
}
