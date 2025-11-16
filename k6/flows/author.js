import http from 'k6/http';
import { check, sleep } from 'k6';

import { BASE_URL } from '../constants.js';
import {
    randomChoice,
    randomBetween,
    generatePrId,
    safeJson,
} from '../utils/common.js';
import { isPrExistsError } from '../utils/errors.js';

export function authorFlow(data) {
    const { users } = data;

    const author = randomChoice(users);
    const prId = generatePrId();

    const createRes = http.post(
        `${BASE_URL}/pullRequest/create`,
        JSON.stringify({
            pull_request_id: prId,
            pull_request_name: `Load test PR by ${author.user_id}`,
            author_id: author.user_id,
        }),
        {
            headers: { 'Content-Type': 'application/json' },
            tags: { operation: 'create_pr' },
        },
    );

    const createBody = safeJson(createRes);
    const createdPr = createBody && createBody.pr;
    const reviewers = createdPr && createdPr.assigned_reviewers
        ? createdPr.assigned_reviewers
        : [];

    check(createRes, {
        'create PR: 201 or PR_EXISTS (409)': (r) =>
            r.status === 201 ||
            (r.status === 409 && isPrExistsError(r)),
    });

    if (createRes.status === 201) {
        check(createRes, {
            'create PR domain: pr object present': () => !!createdPr,
            'create PR domain: reviewers <= 2': () =>
                !!createdPr && reviewers.length <= 2,
            'create PR domain: no author among reviewers': () =>
                !!createdPr && !reviewers.includes(author.user_id),
        });

        const teamRes = http.get(
            `${BASE_URL}/team/get?team_name=${encodeURIComponent(author.team_name)}`,
            { tags: { operation: 'get_team_for_create' } },
        );

        const teamBody = safeJson(teamRes);
        const members = teamBody && teamBody.members ? teamBody.members : [];
        const membersById = {};
        for (const m of members) {
            membersById[m.user_id] = m;
        }

        check(teamRes, {
            'create PR domain: team/get 200': (r) => r.status === 200,
        });

        if (Math.random() < 0.5) {
            const mergePayload = JSON.stringify({ pull_request_id: prId });

            const mergeRes1 = http.post(
                `${BASE_URL}/pullRequest/merge`,
                mergePayload,
                {
                    headers: { 'Content-Type': 'application/json' },
                    tags: { operation: 'merge_pr' },
                },
            );
            const mergeBody1 = safeJson(mergeRes1);
            const mergedPr1 = mergeBody1 && mergeBody1.pr;

            check(mergeRes1, {
                'merge PR #1: 200': (r) => r.status === 200,
                'merge PR #1: status MERGED': () =>
                    !!mergedPr1 && mergedPr1.status === 'MERGED',
            });

            const mergeRes2 = http.post(
                `${BASE_URL}/pullRequest/merge`,
                mergePayload,
                {
                    headers: { 'Content-Type': 'application/json' },
                    tags: { operation: 'merge_pr' },
                },
            );
            const mergeBody2 = safeJson(mergeRes2);
            const mergedPr2 = mergeBody2 && mergeBody2.pr;

            check(mergeRes2, {
                'merge PR #2: idempotent 200': (r) => r.status === 200,
                'merge PR #2: still MERGED': () =>
                    !!mergedPr2 && mergedPr2.status === 'MERGED',
            });
        }
    }

    sleep(randomBetween(0.2, 1.5));
}
