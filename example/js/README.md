# EventSource (SSE) Demo Application

A vanilla JavaScript application demonstrating how to use Server-Sent Events (SSE) / EventSource API to receive real-time notifications from the notification service.

## Features

- ✅ **EventSource Connection**: Connect to SSE endpoint with automatic setup
- ✅ **Ping/Pong Handling**: Detect and handle keepalive ping messages
- ✅ **Auto-Reconnect**: Automatic reconnection with configurable retry logic
- ✅ **Connection Timeout**: Detect stale connections when pings stop arriving
- ✅ **Visual Feedback**: Real-time connection status and statistics
- ✅ **Message Display**: View all received notifications with timestamps
- ✅ **Event Logging**: Detailed log of all connection events
- ✅ **JWZ Authentication**: Bearer token support via Authorization header

## How to Run

### Option 1: Simple HTTP Server (Python)

```bash
cd example/js
python3 -m http.server 8000
```

Then open http://localhost:8000 in your browser.

### Option 2: Simple HTTP Server (Node.js)

```bash
cd example/js
npx http-server -p 8000
```

Then open http://localhost:8000 in your browser.

### Option 3: Open Directly

Simply open `index.html` in your browser (may have CORS restrictions).

## Usage

1. **Configure Connection**:
   - Enter your notification service URL (default: `http://localhost:8080/api/v1/subscribe`)
   - Optionally add a JWZ authentication token (sent as `Authorization: Bearer <token>`)
   - Toggle auto-reconnect on/off

2. **Connect**:
   - Click the "Connect" button
   - Watch the status indicator turn green
   - Connection timer starts

3. **Receive Notifications**:
   - Notifications appear in the "Notifications" panel
   - Each message shows timestamp and content
   - Stats update in real-time

4. **Monitor Connection**:
   - View ping messages in the event log
   - Track reconnection attempts
   - Monitor connection duration

## How It Works

### EventSource API

The application uses either native browser `EventSource` API or a fetch-based polyfill when authentication is required:

**Without Authentication:**
```javascript
const eventSource = new EventSource('http://localhost:8080/api/v1/subscribe');

eventSource.onmessage = (event) => {
    console.log('Message:', event.data);
};
```

**With JWZ Token (uses fetch with Authorization header):**
```javascript
const client = new EventSourceClient(url, {
    jwzToken: 'your-jwz-token-here'
});
// Sends: Authorization: Bearer your-jwz-token-here
```

### Ping/Pong Handling

The client expects periodic ping messages from the server. If no message is received within the `pingTimeout` period (default: 30 seconds), the connection is considered stale and reconnection is triggered.

### Reconnection Logic

When a connection is lost:
1. The client detects the disconnection
2. If auto-reconnect is enabled, it waits `reconnectInterval` milliseconds
3. Attempts to reconnect with exponential backoff
4. Continues until `maxReconnectAttempts` is reached or connection succeeds

### Custom Event Types

The application listens for multiple event types:
- **`message`**: Default event type for general messages
- **`ping`**: Keepalive/heartbeat messages
- **`notification`**: Custom notification events

## Configuration Options

```javascript
{
    jwzToken: null,               // JWZ token sent as Authorization header
    autoReconnect: true,          // Enable auto-reconnection
    reconnectInterval: 3000,      // Wait 3s before reconnecting
    maxReconnectAttempts: Infinity, // Unlimited retry attempts
    pingTimeout: 30000            // 30s timeout without messages
}
```

## Backend Requirements

Your notification service should:

1. **Send periodic pings** (recommended every 10-15 seconds):
   ```
   event: ping
   data: keep-alive
   
   ```

2. **Send notifications** with proper SSE format:
   ```
   event: notification
   data: {"id": "123", "message": "Hello"}
   
   ```

3. **Handle CORS and Authentication** if serving from different origin:
   ```go
   w.Header().Set("Access-Control-Allow-Origin", "*")
   w.Header().Set("Access-Control-Allow-Headers", "Authorization")
   w.Header().Set("Content-Type", "text/event-stream")
   w.Header().Set("Cache-Control", "no-cache")
   w.Header().Set("Connection", "keep-alive")
   ```

## Browser Compatibility

EventSource is supported in all modern browsers:
- Chrome/Edge: ✅
- Firefox: ✅
- Safari: ✅
- Opera: ✅

**Note on Authentication:**
- Native EventSource does **not** support custom headers
- When a JWZ token is provided, the client automatically switches to a fetch-based implementation that supports the `Authorization` header
- This provides seamless authentication while maintaining the same API

## File Structure

```
example/js/
├── index.html              # Main HTML page
├── styles.css              # Styling
├── eventSourceClient.js    # EventSource wrapper with reconnection
├── app.js                  # Application logic and UI handling
└── README.md              # This file
```

## Troubleshooting

### Connection fails immediately
- Check if the backend URL is correct
- Verify the backend is running and accessible
- Check browser console for CORS errors

### No messages received
- Ensure backend is sending SSE-formatted messages
- Check if backend sends `Content-Type: text/event-stream`
- Verify backend doesn't close connection immediately

### Reconnection not working
- Check if auto-reconnect is enabled
- Look for errors in the event log
- Verify network connectivity

## License

This example is part of the iden3/notification-service project.
