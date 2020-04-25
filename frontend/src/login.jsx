import { h } from 'preact';
import { useState, useEffect } from 'preact/hooks';

import {createSessionFromCredentials} from './ticker';
import { showNotification } from "./notification";

async function login(e) {
	e.preventDefault();

	try {
		const session = await createSessionFromCredentials(
			e.target['email'].value,
			e.target['password'].value
		);
		
		return session;
	} catch(e) {
		showNotification('error', e.message);
	}
}

export function Login({setSession}) {
	const handleSubmission = async e => {
		const session = await login(e);

		if(session !== undefined) {
			setSession(session);
		}
	}

	return <div class="login-box">
		<p>Ticker</p>
		<form onSubmit={handleSubmission}>
			<input type="email" name="email" />
			<input type="password" name="password" />
			<button type="submid">Login</button>
		</form>
	</div>
}