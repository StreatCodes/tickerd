import assert from 'assert';
import { test, runTests } from './micro-test.js';

import testUsers from './user.js';

testUsers();

runTests();