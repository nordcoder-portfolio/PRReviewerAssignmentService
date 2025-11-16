import http from 'k6/http';
import { check, sleep } from 'k6';

import { BASE_URL } from '../constants.js';
import {
    randomChoice,
    randomBetween,
    safeJson,
} from '../utils/common.js';

export function reviewerFlow(data) {
    const { users } = data;

    const reviewer = randomChoice(users);

    const listRes = http.get(
        `${BASE_URL}/users/getReview?user_id=${encodeURIComponent(
            reviewer.user_id,
        )}`,
        { tags: { operation: 'get_review' } },
    );

    check(listRes, {
        'getReview: 200': (r) => r.status === 200,
    });

    const listBody = safeJson(listRes);
    const prs = listBody && listBody.pull_requests ? listBody.pull_requests : [];

    if (!prs.length) {
        sleep(randomBetween(0.2, 1.0));
        return;
    }

    const prShort = randomChoice(prs);

    const reassignRes = http.post(
        `${BASE_URL}/pullRequest/reassign`,
        JSON.stringify({
            pull_request_id: prShort.pull_request_id,
            old_user_id: reviewer.user_id,
        }),
        {
            headers: { 'Content-Type': 'application/json' },
            tags: { operation: 'reassign' },
        },
    );

    check(reassignRes, {
        'reassign: 200/404/409 (domain ok)': (r) =>
            r.status === 200 || r.status === 404 || r.status === 409,
    });

    if (reassignRes.status === 200) {
        const body = safeJson(reassignRes);
        const updatedPr = body && body.pr;
        const replacedBy = body && body.replaced_by;
        const newReviewers =
            updatedPr && updatedPr.assigned_reviewers
                ? updatedPr.assigned_reviewers
                : [];

        check(reassignRes, {
            'reassign domain: replaced_by present': () => !!replacedBy,
            'reassign domain: replaced_by in reviewers': () =>
                !!replacedBy && newReviewers.includes(replacedBy),
            'reassign domain: replaced_by != old_user': () =>
                !!replacedBy && replacedBy !== reviewer.user_id,
        });

        const teamRes = http.get(
            `${BASE_URL}/team/get?team_name=${encodeURIComponent(
                reviewer.team_name,
            )}`,
            { tags: { operation: 'get_team_for_reassign' } },
        );

        const teamBody = safeJson(teamRes);
        const members = teamBody && teamBody.members ? teamBody.members : [];
        const membersById = {};
        for (const m of members) {
            membersById[m.user_id] = m;
        }

        check(teamRes, {
            'reassign domain: team/get 200': (r) => r.status === 200,
        });

        check(teamRes, {
            'reassign domain: replaced_by in same team & active': () =>
                !!replacedBy &&
                membersById[replacedBy] &&
                membersById[replacedBy].is_active === true,
        });
    }

    sleep(randomBetween(0.2, 1.5));
}
