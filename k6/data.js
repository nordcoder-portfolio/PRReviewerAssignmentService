import http from 'k6/http';
import { check } from 'k6';

import { BASE_URL, NUM_TEAMS, USERS_PER_TEAM } from './constants.js';
import { isTeamExistsError } from './utils/errors.js';

export function setupTestData() {
    const teams = [];
    const users = [];

    for (let i = 1; i <= NUM_TEAMS; i++) {
        const teamName = `team-${i}`;
        const members = [];

        for (let j = 1; j <= USERS_PER_TEAM; j++) {
            const userId = `u-${i}-${j}`;
            const username = `user_${i}_${j}`;
            const isActive = true;

            members.push({ user_id: userId, username, is_active: isActive });

            users.push({
                user_id: userId,
                username,
                team_name: teamName,
                is_active: isActive,
            });
        }

        const res = http.post(
            `${BASE_URL}/team/add`,
            JSON.stringify({
                team_name: teamName,
                members,
            }),
            {
                headers: { 'Content-Type': 'application/json' },
                tags: { operation: 'team_add' },
            },
        );

        check(res, {
            'team/add 201 or TEAM_EXISTS': (r) =>
                r.status === 201 ||
                (r.status === 400 && isTeamExistsError(r)),
        });

        teams.push({ team_name: teamName, members });
    }

    return { teams, users };
}
