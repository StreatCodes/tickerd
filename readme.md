## Ticker

A lightweight ticketing system which gets updates in real time, good search and supports creating tickets from an email. The frontend has few dependencies and has been built with small libraries to keep load times minimal. Once authenticated all frontend communication happens over a websocket to allow for live updates.

### Build
 - `go build` from the root directory will build the server
 - `npm i` from the `frontend/` directory to install frontend dependencies
 - `npm run prod` will build the frontend