/**
 * Main application logic
 */
class NotificationApp {
  constructor() {
    this.client = null;
    this.messageCount = 0;
    this.pingCount = 0;
    this.reconnectCount = 0;
    this.connectionTimeInterval = null;

    this.initializeElements();
    this.attachEventListeners();
  }

  /**
   * Initialize DOM element references
   */
  initializeElements() {
    this.elements = {
      statusDot: document.getElementById("statusDot"),
      statusText: document.getElementById("statusText"),
      connectBtn: document.getElementById("connectBtn"),
      disconnectBtn: document.getElementById("disconnectBtn"),
      clearBtn: document.getElementById("clearBtn"),
      serverUrl: document.getElementById("serverUrl"),
      jwzToken: document.getElementById("jwzToken"),
      autoReconnect: document.getElementById("autoReconnect"),
      messageCount: document.getElementById("messageCount"),
      pingCount: document.getElementById("pingCount"),
      reconnectCount: document.getElementById("reconnectCount"),
      connectionTime: document.getElementById("connectionTime"),
      notifications: document.getElementById("notifications"),
      logs: document.getElementById("logs"),
    };
  }

  /**
   * Attach event listeners to UI elements
   */
  attachEventListeners() {
    this.elements.connectBtn.addEventListener("click", () => this.connect());
    this.elements.disconnectBtn.addEventListener("click", () =>
      this.disconnect()
    );
    this.elements.clearBtn.addEventListener("click", () => this.clearLogs());
  }

  /**
   * Connect to the notification service
   */
  connect() {
    const url = this.elements.serverUrl.value.trim();
    if (!url) {
      this.addLog("error", "Please enter a valid server URL");
      return;
    }

    const jwzToken = this.elements.jwzToken.value.trim();
    const autoReconnect = this.elements.autoReconnect.checked;

    this.client = new EventSourceClient(url, {
      jwzToken: jwzToken,
      autoReconnect: autoReconnect,
      reconnectInterval: 3000,
      pingTimeout: 30000,
    });

    // Set up event handlers
    this.client.onOpen = () => {
      this.updateStatus("connected", "Connected");
      this.addLog("success", "Connected to notification service");
      this.elements.connectBtn.disabled = true;
      this.elements.disconnectBtn.disabled = false;
      this.startConnectionTimer();
    };

    this.client.onMessage = (data, event) => {
      this.messageCount++;
      this.updateMessageCount();
      this.addNotification(data, event);
      this.addLog("info", `Notification received: ${this.formatData(data)}`);
    };

    this.client.onPing = (data, event) => {
      this.pingCount++;
      this.updatePingCount();
      this.addLog("ping", `Ping received${data ? ": " + data : ""}`);
    };

    this.client.onError = (error) => {
      this.updateStatus("error", "Error");
      this.addLog(
        "error",
        `Connection error: ${error.message || "Unknown error"}`
      );
    };

    this.client.onReconnecting = (attempt) => {
      this.updateStatus("connecting", `Reconnecting (${attempt})...`);
      this.addLog("warning", `Reconnecting... Attempt ${attempt}`);
    };

    this.client.onReconnected = (attempt) => {
      this.reconnectCount++;
      this.updateReconnectCount();
      this.addLog("success", `Reconnected after ${attempt} attempts`);
    };

    this.client.onClose = () => {
      this.updateStatus("disconnected", "Disconnected");
      this.addLog("info", "Disconnected from notification service");
      this.elements.connectBtn.disabled = false;
      this.elements.disconnectBtn.disabled = true;
      this.stopConnectionTimer();
    };

    // Start connection
    this.updateStatus("connecting", "Connecting...");
    this.addLog("info", `Connecting to ${url}...`);
    this.client.connect();
  }

  /**
   * Disconnect from the notification service
   */
  disconnect() {
    if (this.client) {
      this.client.disconnect();
      this.client = null;
    }
    this.stopConnectionTimer();
  }

  /**
   * Update connection status
   */
  updateStatus(status, text) {
    this.elements.statusDot.className = `status-dot ${status}`;
    this.elements.statusText.textContent = text;
  }

  /**
   * Update message count
   */
  updateMessageCount() {
    this.elements.messageCount.textContent = this.messageCount;
  }

  /**
   * Update ping count
   */
  updatePingCount() {
    this.elements.pingCount.textContent = this.pingCount;
  }

  /**
   * Update reconnect count
   */
  updateReconnectCount() {
    this.elements.reconnectCount.textContent = this.reconnectCount;
  }

  /**
   * Start connection time counter
   */
  startConnectionTimer() {
    this.stopConnectionTimer();
    this.connectionTimeInterval = setInterval(() => {
      if (this.client) {
        const duration = this.client.getConnectionDuration();
        this.elements.connectionTime.textContent =
          this.formatDuration(duration);
      }
    }, 1000);
  }

  /**
   * Stop connection time counter
   */
  stopConnectionTimer() {
    if (this.connectionTimeInterval) {
      clearInterval(this.connectionTimeInterval);
      this.connectionTimeInterval = null;
    }
    this.elements.connectionTime.textContent = "00:00";
  }

  /**
   * Format duration in seconds to MM:SS
   */
  formatDuration(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${String(mins).padStart(2, "0")}:${String(secs).padStart(2, "0")}`;
  }

  /**
   * Add notification to the list
   */
  addNotification(data, event) {
    const notificationEl = document.createElement("div");
    notificationEl.className = "notification-item";

    const headerEl = document.createElement("div");
    headerEl.className = "notification-header";
    headerEl.innerHTML = `
            <span>Event Type: ${event.type || "message"}</span>
            <span>${this.getCurrentTime()}</span>
        `;

    const bodyEl = document.createElement("div");
    bodyEl.className = "notification-body";
    bodyEl.textContent = this.formatData(data);

    notificationEl.appendChild(headerEl);
    notificationEl.appendChild(bodyEl);

    this.elements.notifications.insertBefore(
      notificationEl,
      this.elements.notifications.firstChild
    );

    // Keep only last 50 notifications
    while (this.elements.notifications.children.length > 50) {
      this.elements.notifications.removeChild(
        this.elements.notifications.lastChild
      );
    }
  }

  /**
   * Add log entry
   */
  addLog(type, message) {
    const logEl = document.createElement("div");
    logEl.className = `log-entry ${type}`;

    const timeEl = document.createElement("span");
    timeEl.className = "log-time";
    timeEl.textContent = this.getCurrentTime();

    const messageEl = document.createElement("span");
    messageEl.className = "log-message";
    messageEl.textContent = message;

    logEl.appendChild(timeEl);
    logEl.appendChild(messageEl);

    this.elements.logs.insertBefore(logEl, this.elements.logs.firstChild);

    // Keep only last 100 log entries
    while (this.elements.logs.children.length > 100) {
      this.elements.logs.removeChild(this.elements.logs.lastChild);
    }
  }

  /**
   * Clear all logs and notifications
   */
  clearLogs() {
    this.elements.notifications.innerHTML = "";
    this.elements.logs.innerHTML = "";
    this.messageCount = 0;
    this.pingCount = 0;
    this.reconnectCount = 0;
    this.updateMessageCount();
    this.updatePingCount();
    this.updateReconnectCount();
    this.addLog("info", "Logs cleared");
  }

  /**
   * Get current time string
   */
  getCurrentTime() {
    const now = new Date();
    return now.toLocaleTimeString("en-US", { hour12: false });
  }

  /**
   * Format data for display
   */
  formatData(data) {
    if (typeof data === "object") {
      return JSON.stringify(data, null, 2);
    }
    return String(data);
  }
}

// Initialize the app when DOM is ready
document.addEventListener("DOMContentLoaded", () => {
  window.app = new NotificationApp();
});
