
//TODO
export function reducer(state, message) {
	return state;
}

//Fetches the initial websocket state and establishes a websocket connection for live updates
export function connect() {
	const ws = new WebSocket(`ws://${window.location.host}/ws`);
	window.onbeforeunload = () => {console.log('Closing websocket connection'); ws.close();}

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
		window.setTimeout(() => {
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
	
	const res = await fetch('/login', {
		method: 'POST',
		body: JSON.stringify(body)
	});

	if(res.status !== 200) {
		throw new Error(await res.text());
	} else {
		const token = await res.json()
		window.localStorage.setItem('ticker-token', token);

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
		this.ws = new WebSocket(`ws://${window.location.host}/ws?token=${token}`)
		
		this.ws.onopen = e => {
			this.setConnected();
			console.log(`Websocket connection established`);
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
				handler.reject(new Error(message.Error));
			} else {
				handler.resolve(message.Result);
			}
		}
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
}