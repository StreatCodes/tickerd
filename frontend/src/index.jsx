import { h, render, Fragment } from 'preact';
import { useState, useEffect, useReducer } from 'preact/hooks';
import { Router } from 'preact-router';

import {showNotification} from './notification.jsx'
import {Login} from './login.jsx';
import {Navigation} from './navigation.jsx';
import {Home} from './home.jsx';

import * as ws from './socket';

function Main() {
	const [tickets, dispatchTickets] = useReducer(ws.reducer, []);
	const [authenticating, setAuthenticating] = useState(true);
	const [session, setSession] = useState(null);
	const [userInfo, setUserInfo] = useState(null);

	useEffect(async () => {
		const token = window.localStorage.getItem('ticker-token');
		if(token !== null) {
			try {
				const session = await ws.createSessionFromToken(token);
				const userInfo = await session.me()
				setSession(session);
				setUserInfo(userInfo);
			} catch(e) {
				//TODO improve this
				showNotification('error', `Invalid session: ${e.message}`);
				window.localStorage.removeItem('ticker-token');
			}
		}
		setAuthenticating(false);
	}, []);

	if(authenticating) {
		return <div class="loading">Authenticating...</div>
	}

	if(session === null) {
		return <Login setSession={setSession} />
	}

	return <Navigation>
		<Router>
			<Home path="/" />
			<NotFound default />
		</Router>
	</Navigation>
}

function NotFound() {
	return <div>404</div>
}

render(<Main />, document.getElementById('app'));