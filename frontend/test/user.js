import assert from 'assert';
import { test } from './micro-test.js';

import * as ticker from "../src/ticker.js";

export default function testUsers() {
	test('Login and a basic function should work', async ctx => {
		ctx.session = await ticker.createSessionFromCredentials("admin@ticker.io", "password");
		//TODO move below to ticket.js
		//TODO add a status function or something here
		// const response = await ctx.session.createTicket({
		// 	QueueID: 0,
		// 	Subject: "Test ticket",
		// 	Requestor: "test@example.com",
		// 	Reply: [{
		// 		Body: "",
		// 		RenderType: ""
		// 	}],
		// 	Comment: [{
		// 		Body: ""
		// 	}]
		// });
	});
}