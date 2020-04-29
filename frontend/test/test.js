import assert from 'assert';
import { test, runTests } from './micro-test.js';

import testUsers from './user.js';
import testQueues from './queue.js';

testUsers();
testQueues();

test('Connection should close', async ctx => {
	ctx.session.close();
});

runTests();