// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {Client4} from 'mattermost-redux/client';

import type {ActionFuncAsync} from 'types/store';

import ActionTypes from 'action_types/read_receipts';
import type {ReadCursor} from 'types/read_receipts';

import type {DispatchFunc, GetStateFunc} from 'mattermost-redux/types/actions';

export function advanceReadCursor(channelId: string, lastPostSeq?: number, postId?: string): ActionFuncAsync<ReadCursor> {
    return async (dispatch) => {
        try {
            const cursor = await Client4.advanceReadCursor(channelId, lastPostSeq, postId);

            dispatch({
                type: ActionTypes.READ_CURSOR_ADVANCED,
                data: cursor,
            });

            return {data: cursor};
        } catch (error) {
            return {error};
        }
    };
}

export function getReadCursor(channelId: string): ActionFuncAsync<ReadCursor> {
    return async (dispatch) => {
        try {
            const cursor = await Client4.getReadCursor(channelId);

            dispatch({
                type: ActionTypes.RECEIVED_READ_CURSOR,
                data: cursor,
            });

            return {data: cursor};
        } catch (error) {
            return {error};
        }
    };
}

export function receivedReadCursorFromWebSocket(cursor: ReadCursor) {
    return {
        type: ActionTypes.RECEIVED_READ_CURSOR,
        data: cursor,
    };
}

// Track in-flight requests to prevent duplicate API calls
const pendingRequests = new Set<string>();

// Batch queue for efficient bulk fetching
let batchQueue: string[] = [];
let batchTimeout: NodeJS.Timeout | null = null;
let batchDispatch: DispatchFunc | null = null;

export function fetchReadReceiptsCount(postId: string): ActionFuncAsync<{count: number}> {
    return async (dispatch, getState) => {
        // Check if already fetched
        const state = getState();
        const existingCount = state.views.readReceipts?.postReadCounts?.[postId];
        if (existingCount !== undefined) {
            return {data: {count: existingCount}};
        }

        // Prevent duplicate requests
        if (pendingRequests.has(postId)) {
            return {data: {count: 0}};
        }

        pendingRequests.add(postId);
        batchDispatch = dispatch;

        // Add to batch queue
        if (!batchQueue.includes(postId)) {
            batchQueue.push(postId);
        }

        // Schedule batch processing
        if (!batchTimeout) {
            batchTimeout = setTimeout(() => {
                processBatchQueue();
            }, 100); // Wait 100ms to collect more requests
        }

        // Return immediately - data will be dispatched when batch completes
        return {data: {count: 0}};
    };
}

async function processBatchQueue() {
    if (batchQueue.length === 0 || !batchDispatch) {
        batchTimeout = null;
        return;
    }

    const postIds = [...batchQueue];
    const dispatch = batchDispatch;
    batchQueue = [];
    batchTimeout = null;

    try {
        const results = await Client4.getBatchPostReadReceiptsCounts(postIds);
        
        // Dispatch all results
        Object.entries(results).forEach(([postId, count]) => {
            dispatch({
                type: ActionTypes.RECEIVED_READ_RECEIPTS_COUNT,
                data: {
                    postId,
                    count,
                },
            });
            pendingRequests.delete(postId);
        });
    } catch (error) {
        // On error, clear pending and try individual requests as fallback
        for (const postId of postIds) {
            pendingRequests.delete(postId);
            try {
                const result = await Client4.getPostReadReceiptsCount(postId);
                dispatch({
                    type: ActionTypes.RECEIVED_READ_RECEIPTS_COUNT,
                    data: {
                        postId,
                        count: result?.count || 0,
                    },
                });
            } catch (individualError) {
                // Silently fail individual requests
            }
        }
    }
}
