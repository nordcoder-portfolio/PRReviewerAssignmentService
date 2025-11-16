import { safeJson } from './common.js';

export function isTeamExistsError(res) {
    const body = safeJson(res);
    return (
        body &&
        body.error &&
        body.error.code === 'TEAM_EXISTS'
    );
}

export function isPrExistsError(res) {
    const body = safeJson(res);
    return (
        body &&
        body.error &&
        body.error.code === 'PR_EXISTS'
    );
}
