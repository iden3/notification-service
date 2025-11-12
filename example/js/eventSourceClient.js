/**
 * EventSource client with automatic reconnection and ping/pong handling
 */
class EventSourceClient {
  constructor(url, options = {}) {
    this.url = url;
    this.options = {
      jwzToken: options.jwzToken || null,
      autoReconnect: options.autoReconnect !== false,
      reconnectInterval: options.reconnectInterval || 3000,
      maxReconnectAttempts: options.maxReconnectAttempts || Infinity,
      pingTimeout: options.pingTimeout || 30000, // 30 seconds
      ...options,
    };

    this.eventSource = null;
    this.reconnectAttempts = 0;
    this.reconnectTimer = null;
    this.pingTimeoutTimer = null;
    this.connectionStartTime = null;
    this.isManualDisconnect = false;

    // Event handlers
    this.onOpen = null;
    this.onMessage = null;
    this.onPing = null;
    this.onError = null;
    this.onReconnecting = null;
    this.onReconnected = null;
    this.onClose = null;
  }

  /**
   * Build URL with query parameters if needed
   */
  buildUrl() {
    const url = new URL(this.url);
    // Add any query parameters here if needed
    return url.toString();
  }

  /**
   * Connect to the EventSource
   */
  connect() {
    this.isManualDisconnect = false;
    this.connectionStartTime = Date.now();

    try {
      const url = this.buildUrl();

      // If jwzToken is provided, use fetch-based EventSource with Authorization header
      if (this.options.jwzToken) {
        this.connectWithAuth(url);
      } else {
        this.connectNative(url);
      }
    } catch (error) {
      console.error("[EventSource] Connection error", error);
      if (this.onError) {
        this.onError(error);
      }
      this.handleDisconnect();
    }
  }

  /**
   * Connect using native EventSource (no custom headers)
   */
  connectNative(url) {
    this.eventSource = new EventSource(url);
    this.setupEventListeners();
  }

  /**
   * Connect using fetch with Authorization header
   */
  async connectWithAuth(url) {
    try {
      const response = await fetch(url, {
        headers: {
          Authorization: `Bearer ${this.options.jwzToken}`,
          Accept: "text/event-stream",
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      // Create a ReadableStream reader
      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = "";

      // Simulate EventSource ready state
      this.eventSource = {
        readyState: 1, // OPEN
        close: () => {
          this.eventSource.readyState = 2; // CLOSED
          reader.cancel();
        },
      };

      // Trigger onopen
      if (this.onOpen) {
        this.onOpen({ type: "open" });
      }
      this.reconnectAttempts = 0;
      this.resetPingTimeout();

      // Read stream
      const processStream = async () => {
        try {
          while (true) {
            const { done, value } = await reader.read();

            if (done) {
              console.log("[EventSource] Stream closed");
              this.handleDisconnect();
              break;
            }

            buffer += decoder.decode(value, { stream: true });
            const lines = buffer.split("\n");
            buffer = lines.pop() || ""; // Keep incomplete line in buffer

            let eventType = "message";
            let eventData = "";
            let eventId = null;

            for (const line of lines) {
              if (line.startsWith("event:")) {
                eventType = line.substring(6).trim();
              } else if (line.startsWith("data:")) {
                eventData += line.substring(5).trim() + "\n";
              } else if (line.startsWith("id:")) {
                eventId = line.substring(3).trim();
              } else if (line === "") {
                // Empty line means end of event
                if (eventData) {
                  eventData = eventData.trim();
                  this.resetPingTimeout();

                  const event = {
                    type: eventType,
                    data: eventData,
                    lastEventId: eventId,
                  };

                  if (eventType === "ping") {
                    this.handlePing(event);
                  } else {
                    this.handleMessage(event);
                  }

                  // Reset for next event
                  eventType = "message";
                  eventData = "";
                  eventId = null;
                }
              }
            }
          }
        } catch (error) {
          if (!this.isManualDisconnect) {
            console.error("[EventSource] Stream error", error);
            if (this.onError) {
              this.onError(error);
            }
            this.handleDisconnect();
          }
        }
      };

      processStream();
    } catch (error) {
      console.error("[EventSource] Fetch error", error);
      if (this.onError) {
        this.onError(error);
      }
      this.handleDisconnect();
    }
  }

  /**
   * Setup event listeners for native EventSource
   */
  setupEventListeners() {
    if (!this.eventSource) return;

    this.eventSource.onopen = (event) => {
      console.log("[EventSource] Connection opened", event);
      this.reconnectAttempts = 0;
      this.resetPingTimeout();

      if (this.onOpen) {
        this.onOpen(event);
      }
    };

    this.eventSource.onerror = (event) => {
      console.error("[EventSource] Error occurred", event);

      if (this.eventSource.readyState === EventSource.CLOSED) {
        console.log("[EventSource] Connection closed");
        this.handleDisconnect();
      }

      if (this.onError) {
        this.onError(event);
      }
    };

    this.eventSource.onmessage = (event) => {
      console.log("[EventSource] Message received", event.data);
      this.resetPingTimeout();
      this.handleMessage(event);
    };

    // Listen for custom event types
    this.eventSource.addEventListener("ping", (event) => {
      console.log("[EventSource] Ping received", event.data);
      this.resetPingTimeout();
      this.handlePing(event);
    });

    this.eventSource.addEventListener("notification", (event) => {
      console.log("[EventSource] Notification received", event.data);
      this.resetPingTimeout();
      this.handleMessage(event);
    });
  }

  /**
   * Handle incoming messages
   */
  handleMessage(event) {
    try {
      // Try to parse as JSON
      let data;
      try {
        data = JSON.parse(event.data);
      } catch (e) {
        // If not JSON, use raw data
        data = event.data;
      }

      if (this.onMessage) {
        this.onMessage(data, event);
      }
    } catch (error) {
      console.error("[EventSource] Error handling message", error);
    }
  }

  /**
   * Handle ping messages
   */
  handlePing(event) {
    if (this.onPing) {
      this.onPing(event.data, event);
    }
  }

  /**
   * Reset ping timeout timer
   */
  resetPingTimeout() {
    if (this.pingTimeoutTimer) {
      clearTimeout(this.pingTimeoutTimer);
    }

    this.pingTimeoutTimer = setTimeout(() => {
      console.warn("[EventSource] Ping timeout - no messages received");
      if (
        this.eventSource &&
        this.eventSource.readyState === EventSource.OPEN
      ) {
        // Connection seems stale, force reconnect
        this.disconnect();
        if (this.options.autoReconnect) {
          this.reconnect();
        }
      }
    }, this.options.pingTimeout);
  }

  /**
   * Handle disconnection and attempt reconnect
   */
  handleDisconnect() {
    this.clearPingTimeout();

    if (this.isManualDisconnect) {
      console.log("[EventSource] Manual disconnect - not reconnecting");
      return;
    }

    if (!this.options.autoReconnect) {
      console.log("[EventSource] Auto-reconnect disabled");
      return;
    }

    if (this.reconnectAttempts >= this.options.maxReconnectAttempts) {
      console.log("[EventSource] Max reconnect attempts reached");
      if (this.onClose) {
        this.onClose();
      }
      return;
    }

    this.reconnect();
  }

  /**
   * Attempt to reconnect
   */
  reconnect() {
    if (this.reconnectTimer) {
      return; // Already attempting to reconnect
    }

    this.reconnectAttempts++;
    console.log(
      `[EventSource] Attempting to reconnect (${this.reconnectAttempts})...`
    );

    if (this.onReconnecting) {
      this.onReconnecting(this.reconnectAttempts);
    }

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();

      if (this.onReconnected) {
        this.onReconnected(this.reconnectAttempts);
      }
    }, this.options.reconnectInterval);
  }

  /**
   * Clear ping timeout
   */
  clearPingTimeout() {
    if (this.pingTimeoutTimer) {
      clearTimeout(this.pingTimeoutTimer);
      this.pingTimeoutTimer = null;
    }
  }

  /**
   * Disconnect from EventSource
   */
  disconnect() {
    this.isManualDisconnect = true;

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    this.clearPingTimeout();

    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }

    this.connectionStartTime = null;

    if (this.onClose) {
      this.onClose();
    }
  }

  /**
   * Get current connection state
   */
  getState() {
    if (!this.eventSource) {
      return EventSource.CLOSED;
    }
    return this.eventSource.readyState;
  }

  /**
   * Check if connected
   */
  isConnected() {
    return this.eventSource && this.eventSource.readyState === EventSource.OPEN;
  }

  /**
   * Get connection duration in seconds
   */
  getConnectionDuration() {
    if (!this.connectionStartTime) {
      return 0;
    }
    return Math.floor((Date.now() - this.connectionStartTime) / 1000);
  }
}

// Make available globally
window.EventSourceClient = EventSourceClient;
