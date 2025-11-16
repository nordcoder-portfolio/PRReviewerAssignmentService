import { options as baseOptions } from './config.js';
import { setupTestData } from './data.js';

import { authorFlow } from './flows/author.js';
import { reviewerFlow } from './flows/reviewer.js';
import { adminFlow } from './flows/admin.js';

export const options = baseOptions;

export function setup() {
    return setupTestData();
}

export { authorFlow, reviewerFlow, adminFlow };
