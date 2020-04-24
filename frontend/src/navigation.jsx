import { h, Fragment } from 'preact';
import { useState, useEffect } from 'preact/hooks';

export function Navigation({children}) {
	return <Fragment>
		<div class="search-bar">
			search
		</div>
		<nav>
			nav
		</nav>
		{children}
	</Fragment>
}