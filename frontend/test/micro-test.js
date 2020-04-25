const colorReset = "\x1b[0m";
const colorFgRed = "\x1b[31m";
const colorFgGreen = "\x1b[32m";

const tests = [];

const toms = (started, ended) => `${ended - started}ms`;

export function test(name, func) {
	tests.push({
		name: name,
		func: func
	});
}

export async function runTests() {
	const ctx = {};

	const started = Date.now();
	let passed = 0;

	for(const test of tests) {
		const testStarted = Date.now();

		try {
			await test.func(ctx);
			console.log(`${colorFgGreen}OK${colorReset} (${toms(testStarted, Date.now())}) - ${test.name}`);
			passed += 1;
		} catch(e) {
			console.log(`${colorFgRed}ERR (${toms(testStarted, Date.now())}) - ${test.name}\n${e.message}\n${e.stack}${colorReset}`);
		}
	}

	console.log(`\n${passed}/${tests.length} tests passed in ${toms(started, Date.now())}`);

	if(passed !== tests.length) {
		process.exit(1);
	}
}