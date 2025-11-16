import { RESPONSE_SLI_MS, SUCCESS_SLI } from './constants.js';

export const options = {
    scenarios: {
        authors: {
            executor: 'constant-arrival-rate',
            rate: 6,
            timeUnit: '1s',
            duration: '10m',
            preAllocatedVUs: 5,
            maxVUs: 20,
            exec: 'authorFlow',
        },
        reviewers: {
            executor: 'constant-arrival-rate',
            rate: 2,
            timeUnit: '1s',
            duration: '10m',
            preAllocatedVUs: 3,
            maxVUs: 10,
            exec: 'reviewerFlow',
            startTime: '10s',
        },
        admins: {
            executor: 'constant-arrival-rate',
            rate: 2,
            timeUnit: '1s',
            duration: '10m',
            preAllocatedVUs: 1,
            maxVUs: 5,
            exec: 'adminFlow',
            startTime: '20s',
        },
    },
    thresholds: {
        [`http_req_duration{operation:create_pr}`]: [`p(99)<${RESPONSE_SLI_MS}`],
        [`http_req_duration{operation:get_review}`]: [`p(99)<${RESPONSE_SLI_MS}`],
        [`http_req_duration{operation:reassign}`]: [`p(99)<${RESPONSE_SLI_MS}`],
        [`http_req_duration{operation:merge_pr}`]: [`p(99)<${RESPONSE_SLI_MS}`],
        [`http_req_duration{operation:get_team}`]: [`p(99)<${RESPONSE_SLI_MS}`],
        [`http_req_duration{operation:set_active}`]: [`p(99)<${RESPONSE_SLI_MS}`],

        checks: [`rate>${SUCCESS_SLI}`],
    },
};
