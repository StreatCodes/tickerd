import fetch from 'node-fetch';
import WebSocket from 'ws';

let TESTPREFIX = '';
let WSPREFIX = 'ws://localhost:8080';
if(typeof window === 'undefined') {
	TESTPREFIX = 'http://localhost:8080';
} else {
	WSPREFIX = `ws://${window.location.host}`
}

/**
 * @typedef Ticket
 * @type {object}
 * @property {number} ID - Ticket ID, null to auto increment
 * @property {number} QueueID - The ID of the Queue to assign the ticket to
 * @property {string} Subject - Ticket subject
 * @property {string} Requestor - Email of the requestor
 * @property {string} Status - new|open|closed omit for new
 * @property {number} Priority - 0-255 priority
 * @property {string} CreatedAt - A datetime, null for Now()
 * @property {Reply[]} Replies - Array of Replies
 * @property {Comment[]} Comments - Array of comments
 * 
 * @typedef Reply
 * @type {object}
 * @property {string} Body - The body of the reply
 * @property {string} RenderType - The type of the body contents html|plaintext
 * @property {string[]} AttachmentHashes - List of attachment hashes
 * @property {string} CreatedAt - A datetime, null for Now()
 * 
 * @typedef Comment
 * @type {object}
 * @property {string} Body - The body of the reply
 * @property {string[]} AttachmentHashes - List of attachment hashes
 * @property {string} CreatedAt - A datetime, null for Now()
 * @property {string} EditedAt - Date the comment was last edited (can be null)
 * 
 * @typedef Queue
 * @type {object}
 * @property {number} ID - ID of the queue
 * @property {string} Name - Queue name (must be unique)
 * @property {string} Email - Email associated with the queue (emails with this To Addr will be routed to this queue)
 */

//TODO
export function reducer(state, message) {
	return state;
}

//Fetches the initial websocket state and establishes a websocket connection for live updates
export function connect() {
	const ws = new WebSocket(`${WSPREFIX}/ws`);
	if(typeof window !== 'undefined') {
		window.onbeforeunload = () => {console.log('Closing websocket connection'); ws.close();}
	}

	//Get full state and mark them as inserts
	ws.onopen = async function() {
		showNotification('success', 'Websocket connection established, fetching full state');
	}

	ws.onerror = function(e) {
		console.log(`Websocket error:`)
		console.log(e)
	}

	ws.onclose = function(e) {
		showNotification('error', 'Websocket closed, scheduling reconnect');
		setTimeout(() => {
			connect();
		}, 5000);
	}

	ws.onmessage = function(e) {
		const message = JSON.parse(e.data);
		console.log(message);
	}
}

export async function createSessionFromToken(token) {
	return new Session(token);
}

export async function createSessionFromCredentials(email, password) {
	const body = {
		Email: email,
		Password: password
	};
	
	const res = await fetch(`${TESTPREFIX}/login`, {
		method: 'POST',
		body: JSON.stringify(body)
	});

	if(res.status !== 200) {
		throw new Error(await res.text());
	} else {
		const token = await res.json()
		// window.localStorage.setItem('ticker-token', token);

		return new Session(token);
	}
}

class Session {
	constructor(token) {
		this.messageCount = 0;
		this.token = token;
		this.setConnected = null;
		this.connected = new Promise(resolve => this.setConnected = resolve);
		this.pendingMessage = new Map();
		this.ws = new WebSocket(`${WSPREFIX}/ws?token=${token}`)

		this.ws.onopen = e => {
			this.setConnected();
		}

		this.ws.onclose = e => {
			this.setConnected = null;
			this.connected = new Promise(resolve => this.setConnected = resolve);
			console.log(`Websocket connection closed`);
		}

		//reject/resolve responses with matching IDs from the server
		this.ws.onmessage = e => {
			const message = JSON.parse(e.data);
			if(typeof message.ID === 'undefined' || message.ID === null) {
				console.error('Received WS message with no ID');
				console.log(message);
			}
			const handler = this.pendingMessage.get(message.ID);

			if(typeof handler === 'undefined') {
				console.error('Received WS message with no matching handler');
				console.log(message);
			}

			if(message.Error !== null) {
				handler.reject(new Error(message.Error.trim()));
			} else {
				handler.resolve(message.Result);
			}
		}
	}

	close() {
		this.ws.close()
	}

	//Create new promise and reject/resolve responses with matching IDs from the server
	async _sendMessage(method, params) {
		//Wait to be connected
		await this.connected;
		
		const count = this.messageCount;
		this.messageCount += 1;

		return new Promise((resolve, reject) => {
			this.pendingMessage.set(count, {resolve, reject});
			const request = {ID: count, Method: method, Params: params};
			this.ws.send(JSON.stringify(request));
		});
	}

	async me() {
		return this._sendMessage('me', null);
	}
	async echo(message) {
		return this._sendMessage('echo', message);
	}


	/**
	 * Create a ticket including any replies and comments,
	 * which can be useful for importers
	 *
	 * @param {Ticket} ticket
	 * @memberof Session
	 */
	async createTicket(ticket) {
		return this._sendMessage('createTicket', ticket);
	}

	/**
	 * Create a queue
	 *
	 * @param {Queue} queue
	 * @memberof Session
	 */
	async createQueue(queue) {
		return this._sendMessage('createQueue', queue);
	}

	/**
	 * Update a queue
	 *
	 * @param {Queue} queue
	 * @memberof Session
	 */
	async updateQueue(queue) {
		return this._sendMessage('updateQueue', queue);
	}

	/**
	 * List all queues
	 *
	 * @memberof Session
	 */
	async listQueues() {
		return this._sendMessage('listQueues', null);
	}
	

	/**
	 * Delete a queue
	 *
	 * @param {number} queueID
	 * @memberof Session
	 */
	async deleteQueue(queueID) {
		return this._sendMessage('deleteQueue', queueID);
	}
}