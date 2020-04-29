import assert from 'assert';
import { test } from './micro-test.js';

import * as ticker from "../src/ticker.js";

export default function testQueues() {
	test('Creating a queue with an empty name should fail', async ctx => {
		await assert.rejects(
			ctx.session.createQueue({
				Name: "",
				Email: "john@example.com"
			}),
			{message: "Error validating new queue: Name can't be blank"}
		);
	});

	test('Creating a queue with invalid emails should fail', async ctx => {
		const invalidEmails = [
			"plainaddress",
			"#@%^%#$@#$@#.com",
			"@example.com",
			"email.example.com",
		]
		
		for(const email of invalidEmails) {
			await assert.rejects(
				ctx.session.createQueue({
					Name: "Test",
					Email: email
				}),
				{message: /^Error validating new queue: Invalid Email address:/}
			);
		}
	});

	test('Creating a valid queue should succeed', async ctx => {
		const queue = await ctx.session.createQueue({
			Name: "Test",
			Email: "email@example.com"
		});

		assert.equal(typeof queue.ID, 'number');
		assert.equal(queue.Name, "Test");
		assert.equal(queue.Email, "email@example.com");

		ctx.testQueueID = queue.ID;
	});

	test('Creating a second queue with the same name or email should fail', async ctx => {
		await assert.rejects(
			ctx.session.createQueue({
				Name: "unqiue",
				Email: "email@example.com"
			}),
			{message: "Error saving queue: already exists"}
		);

		await assert.rejects(
			ctx.session.createQueue({
				Name: "Test",
				Email: "unique@example.com"
			}),
			{message: "Error saving queue: already exists"}
		);
	});

	test('Deleting a non-existant queue should fail', async ctx => {
		await assert.rejects(
			ctx.session.deleteQueue(123111),
			{message: "Queue does not exist"}
		);
	});

	test('Deleting a valid queue should succeed', async ctx => {
		await ctx.session.deleteQueue(ctx.testQueueID);
	});

	test('Updating a non-existant queue should fail', async ctx => {
		await assert.rejects(
			ctx.session.updateQueue({
				ID: 123111,
				Name: "Test",
				Email: "email@example.com"
			}),
			{message: "Queue does not exist"}
		);
	});

	test('Updating a queue should return the updated queue', async ctx => {
		const queue = await ctx.session.createQueue({
			Name: "Test",
			Email: "email@example.com"
		});

		const updatedQueue = await ctx.session.updateQueue({
			ID: queue.ID,
			Name: "Update name",
			Email: "updated@example.com"
		});

		assert.equal(updatedQueue.ID, queue.ID);
		assert.equal(updatedQueue.Name, "Update name");
		assert.equal(updatedQueue.Email, "updated@example.com");

		const finalQueues = await ctx.session.listQueues()
		const finalQueue = finalQueues.find(q => q.ID === queue.ID);

		assert.equal(finalQueue.ID, queue.ID);
		assert.equal(finalQueue.Name, "Update name");
		assert.equal(finalQueue.Email, "updated@example.com");
	});
}