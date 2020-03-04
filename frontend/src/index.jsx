import { h, render, Fragment } from 'preact';
import { useState, useEffect, useReducer } from 'preact/hooks';
import { Router } from 'preact-router';

import {showNotification} from './notification.jsx'
import {Login} from './login.jsx';
import {Home} from './home.jsx';

import * as ws from './socket';

function Main() {
	const [tickets, dispatchTickets] = useReducer(ws.reducer, []);
	const [token, setToken] = useState(null);
	const [session, setSession] = useState(null);

	useEffect(() => {
		const token = window.localStorage.getItem('ticker-token');
		if(token !== null) {
			setToken(token);

			ws.createSessionFromToken(token).then(session => {
				setSession(session);
				session.me().then(res => console.log(res));
			}).catch(err => {
				showNotification('error', `Invalid session: ${err.message}`);
				// window.localStorage.removeItem('ticker-token');
				setToken(null);
			});
		}
	}, []);

	if(session === null) {
		return <Login setSession={setSession} />
	}

	// if(session === null) {
	// 	return <div class="loading">Logging you in</div>
	// }

	return <Router>
		<Home path="/" />
		<NotFound default />
	</Router>
}

function NotFound() {
	return <div>404</div>
}

render(<Main />, document.getElementById('app'));