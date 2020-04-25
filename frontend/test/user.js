import assert from 'assert';
import { test } from './micro-test.js';

import * as ticker from "../src/ticker.js";

export default function testUsers() {
	test('Login and a basic function should work', async ctx => {
		const session = await ticker.createSessionFromCredentials("admin@ticker.io", "password")
		const response = await session.echo("hello server");
		console.log(response);
	});
}