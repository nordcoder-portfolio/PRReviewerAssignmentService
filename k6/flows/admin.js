import http from 'k6/http';
import { check, sleep } from 'k6';

import { BASE_URL } from '../constants.js';
import { randomChoice, randomBetween } from '../utils/common.js';

export function adminFlow(data) {
    const { teams } = data;

    const team = randomChoice(teams);

    const getRes = http.get(
        `${BASE_URL}/team/get?team_name=${encodeURIComponent(
            team.team_name,
        )}`,
        { tags: { operation: 'get_team' } },
    );

    check(getRes, {
        'get team: 200 or 404': (r) => r.status === 200 || r.status === 404,
    });

    if (team.members.length && Math.random() < 0.5) {
        const member = randomChoice(team.members);
        const newActive = Math.random() < 0.5;

        const setRes = http.post(
            `${BASE_URL}/users/setIsActive`,
            JSON.stringify({
                user_id: member.user_id,
                is_active: newActive,
            }),
            {
                headers: { 'Content-Type': 'application/json' },
                tags: { operation: 'set_active' },
            },
        );

        check(setRes, {
            'setIsActive: 200 or 404': (r) =>
                r.status === 200 || r.status === 404,
        });
    }

    sleep(randomBetween(1.0, 2.0));
}
