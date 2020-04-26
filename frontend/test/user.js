import assert from 'assert';
import { test } from './micro-test.js';

import * as ticker from "../src/ticker.js";

export default function testUsers() {
	test('Login and a basic function should work', async ctx => {
		const session = await ticker.createSessionFromCredentials("admin@ticker.io", "password");
		const response = await session.createTicket({
			QueueID: 0,
			Subject: "Test ticket",
			Requestor: "test@example.com",
			Reply: [{
				Body: "",
				RenderType: ""
			}],
			Comment: [{
				Body: ""
			}]
		});

		console.log(response);
		session.close();
	});
}