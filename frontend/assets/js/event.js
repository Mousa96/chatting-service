// event.js - Enhanced WebSocket Event Handling with Message Status
class Event {
  constructor(type, payload) {
    this.type = type;
    this.payload = payload;
  }
}

// Global state for tracking current conversation
window.currentConversationWith = null;
window.isUserActive = true;

// Remove focus/blur handlers - user is considered online when connected
function routeEvent(event) {
  console.log("Routing event:", event);

  if (!event.type) {
    console.error("Event type is undefined");
    return;
  }
  console.log("Event type:", event.type);
  switch (event.type) {
    case "receive_message": // Individual messages (including from broadcasts)
      handleNewMessage(event.payload);
      break;
    case "status_change": // Message status updates
      handleStatusChange(event.payload);
      break;
    case "user_status":
      handleUserStatusChange(event.payload);
      break;
    case "typing":
      handleTypingIndicator(event.payload);
      break;
    case "error":
      handleError(event.payload);
      break;
    default:
      console.log("Unsupported event type:", event.type);
      break;
  }
}

function handleNewMessage(messageData) {
  console.log("Received new message:", messageData);

  // Parse the message payload if it's a string
  let message;
  if (typeof messageData === "string") {
    try {
      message = JSON.parse(messageData);
    } catch (e) {
      console.error("Failed to parse message data:", e);
      return;
    }
  } else {
    message = messageData;
  }

  // Convert backend field names to frontend format and ensure status is set
  const normalizedMessage = {
    id: message.id,
    sender_id: message.sender_id,
    receiver_id: message.receiver_id,
    content: message.content,
    media_url: message.media_url,
    status: message.status || "sent", // Always ensure we have a status
    created_at: message.created_at,
    updated_at: message.updated_at,
  };

  console.log("Normalized message with status:", normalizedMessage);

  // Add message to conversation if it's the active one or if it's from/to the current user
  const currentUserId = parseInt(window.currentUserId);
  const isInCurrentConversation =
    window.selectedUserId &&
    (normalizedMessage.sender_id === window.selectedUserId ||
      normalizedMessage.receiver_id === window.selectedUserId);

  if (isInCurrentConversation) {
    if (window.addMessageToConversation) {
      window.addMessageToConversation(normalizedMessage);
    } else {
      console.warn(
        "addMessageToConversation function not available, reloading conversation"
      );
      if (window.loadConversation) {
        window.loadConversation(window.selectedUserId);
      }
    }
  }

  // Show notification for new messages when not viewing that conversation
  if (
    normalizedMessage.sender_id !== currentUserId &&
    window.selectedUserId !== normalizedMessage.sender_id
  ) {
    showMessageNotification(normalizedMessage);
  }

  // FIXED: Auto-mark messages as delivered/read only if we're the recipient
  if (
    normalizedMessage.receiver_id === currentUserId &&
    window.selectedUserId === normalizedMessage.sender_id &&
    window.isUserActive
  ) {
    // Mark as delivered after a short delay
    setTimeout(() => {
      sendEvent("update_status", {
        message_id: normalizedMessage.id,
        status: "delivered",
      });
    }, 500);

    // Mark as read after user has "seen" it
    setTimeout(() => {
      if (
        window.isUserActive &&
        window.selectedUserId === normalizedMessage.sender_id
      ) {
        sendEvent("message_read", {
          message_id: normalizedMessage.id,
        });
      }
    }, 2000);
  }
}

function handleStatusChange(statusData) {
  console.log("Message status changed:", statusData);

  let statusChange;
  if (typeof statusData === "string") {
    try {
      statusChange = JSON.parse(statusData);
    } catch (e) {
      console.error("Failed to parse status data:", e);
      return;
    }
  } else {
    statusChange = statusData;
  }

  updateMessageStatusInUI(statusChange.message_id, statusChange.status);
}

function updateMessageStatusInUI(messageId, status) {
  console.log(`Updating message ${messageId} status to ${status}`);

  // Find the message element and update its status
  const messageElement = document.querySelector(
    `[data-message-id="${messageId}"]`
  );

  if (messageElement) {
    let statusElement = messageElement.querySelector(".message-status");

    // Create status element if it doesn't exist (for sent messages that were missing it)
    if (!statusElement) {
      statusElement = document.createElement("div");
      statusElement.className = "message-status";

      // Add to message meta section
      const messageMeta = messageElement.querySelector(".message-meta");
      if (messageMeta) {
        messageMeta.appendChild(statusElement);
      } else {
        // Create meta section if it doesn't exist
        const metaEl = document.createElement("div");
        metaEl.className = "message-meta";
        metaEl.appendChild(statusElement);
        messageElement.appendChild(metaEl);
      }

      console.log(`Created missing status element for message ${messageId}`);
    }

    // Update status text and class
    statusElement.textContent = getStatusText(status);
    statusElement.className = `message-status ${status}`;
    statusElement.title = getStatusTitle(status);

    // Add visual feedback animation
    statusElement.classList.add("updated");
    setTimeout(() => {
      statusElement.classList.remove("updated");
    }, 500);

    console.log(`Updated UI for message ${messageId} to status ${status}`);
  } else {
    console.warn(`Message element with ID ${messageId} not found in DOM`);
  }
}

function getStatusText(status) {
  switch (status) {
    case "sending":
      return "⏳";
    case "sent":
      return "✓";
    case "delivered":
      return "✓✓";
    case "read":
      return "✓✓";
    default:
      return "✓"; // Default to sent
  }
}

function getStatusTitle(status) {
  switch (status) {
    case "sending":
      return "Sending message...";
    case "sent":
      return "Message sent";
    case "delivered":
      return "Message delivered";
    case "read":
      return "Message read";
    default:
      return "Message sent";
  }
}

function handleUserStatusChange(payload) {
  console.log("User status changed:", payload);

  let statusChange;
  if (typeof payload === "string") {
    try {
      statusChange = JSON.parse(payload);
    } catch (e) {
      console.error("Failed to parse user status data:", e);
      return;
    }
  } else {
    statusChange = payload;
  }

  if (statusChange && statusChange.user_id && statusChange.status) {
    updateUserStatusInUI(statusChange.user_id, statusChange.status);

    // Update online count after status change
    if (window.updateOnlineCount) {
      window.updateOnlineCount();
    }
  }
}

function handleTypingIndicator(payload) {
  console.log("Typing indicator:", payload);

  let typingData;
  if (typeof payload === "string") {
    try {
      typingData = JSON.parse(payload);
    } catch (e) {
      console.error("Failed to parse typing data:", e);
      return;
    }
  } else {
    typingData = payload;
  }

  if (typingData && typingData.user_id && typingData.is_typing !== undefined) {
    showTypingIndicator(typingData.user_id, typingData.is_typing);
  }
}

function handleError(payload) {
  console.error("WebSocket error received:", payload);

  let errorMessage;
  if (typeof payload === "string") {
    errorMessage = payload;
  } else if (payload && payload.message) {
    errorMessage = payload.message;
  } else {
    errorMessage = "Unknown error occurred";
  }

  // Show error message to user
  alert("Error: " + errorMessage);
}

function showMessageNotification(message) {
  // Visual notification for new messages
  const user = window.allUsers?.find((u) => u.id === message.sender_id);
  if (user) {
    // Add visual indicator to user list
    const userElement = document.getElementById(`user-${user.id}`);
    if (userElement && !userElement.classList.contains("active")) {
      userElement.classList.add("new-message");
    }
  }

  // Optional: Show browser notification if supported
  if (Notification.permission === "granted") {
    const userName = user ? user.username : `User ${message.sender_id}`;
    new Notification(`New message from ${userName}`, {
      body: message.content || "Sent a media file",
      icon: "/favicon.ico",
    });
  }
}

function updateUserStatusInUI(userId, status) {
  console.log(`Updating user ${userId} status to ${status}`);

  const userElement = document.getElementById(`user-${userId}`);
  if (userElement) {
    const statusDot = userElement.querySelector(".status-dot");
    if (statusDot) {
      statusDot.className = `status-dot ${status}`;
    }

    // Update user object in allUsers array
    if (window.allUsers) {
      const user = window.allUsers.find((u) => u.id === userId);
      if (user) {
        user.status = status;
      }
    }

    // Update broadcast user status too
    const broadcastUserElement = document.querySelector(
      `#broadcast-user-${userId}`
    );
    if (broadcastUserElement) {
      const broadcastStatusDot =
        broadcastUserElement.parentElement.querySelector(".status-dot");
      if (broadcastStatusDot) {
        broadcastStatusDot.className = `status-dot ${status}`;
      }
    }

    // If this is the currently selected user, update header status
    if (window.selectedUserId === userId) {
      const headerStatusDot = document.querySelector(
        "#selected-user-status .status-dot"
      );
      const headerStatusText = document.querySelector(
        "#selected-user-status .status-text"
      );

      if (headerStatusDot) {
        headerStatusDot.className = `status-dot ${status}`;
      }
      if (headerStatusText) {
        headerStatusText.textContent =
          status === "online" ? "Online" : "Offline";
      }
    }
  }

  // Update online count
  if (window.updateOnlineCount) {
    window.updateOnlineCount();
  }
}

function showTypingIndicator(userId, isTyping) {
  // Only show if we're currently chatting with this user
  if (window.selectedUserId === userId) {
    let indicator = document.getElementById("typing-indicator");

    if (isTyping) {
      if (!indicator) {
        indicator = document.createElement("div");
        indicator.id = "typing-indicator";
        indicator.className = "typing-indicator";

        const user = window.allUsers?.find((u) => u.id === userId);
        const userName = user ? user.username : `User ${userId}`;
        indicator.textContent = `${userName} is typing...`;

        const chatHeader = document.querySelector(".chat-header");
        if (chatHeader) {
          chatHeader.appendChild(indicator);
        }
      }
      indicator.style.display = "block";
    } else {
      if (indicator) {
        indicator.style.display = "none";
      }
    }
  }
}

// Enhanced sendEvent function with conversation tracking
function sendEvent(eventName, payload) {
  console.log("Sending event:", { eventName, payload });

  if (!window.socket || window.socket.readyState !== WebSocket.OPEN) {
    console.error("WebSocket is not connected");
    return false;
  }

  const event = {
    type: eventName,
    payload: payload,
  };

  try {
    window.socket.send(JSON.stringify(event));
    return true;
  } catch (error) {
    console.error("Failed to send WebSocket event:", error);
    return false;
  }
}

// Function to notify backend when user opens a conversation
function notifyConversationOpened(userId) {
  window.currentConversationWith = userId;
  sendEvent("conversation_opened", { user_id: userId });

  // Mark any unread messages in this conversation as read after a delay
  setTimeout(() => {
    markConversationMessagesAsRead(userId);
    se;
  }, 1500);
}

// Function to notify backend when user closes a conversation
function notifyConversationClosed() {
  if (window.currentConversationWith) {
    sendEvent("conversation_closed", {});
    window.currentConversationWith = null;
  }
}

// FIXED: Only mark RECEIVED messages as read
function markConversationMessagesAsRead(userId) {
  if (!window.isUserActive || window.currentConversationWith !== userId) {
    return;
  }

  // Only mark received messages in the current conversation as read
  const messageElements = document.querySelectorAll(".message.received");

  messageElements.forEach((messageEl) => {
    const messageId = messageEl.getAttribute("data-message-id");

    // Only mark messages we received (not sent) and that aren't already read
    if (messageId && parseInt(messageId) > 0) {
      const statusEl = messageEl.querySelector(".message-status");

      // Only send read receipt if message isn't already marked as read
      if (!statusEl || !statusEl.classList.contains("read")) {
        sendEvent("message_read", {
          message_id: parseInt(messageId),
        });
      }
    }
  });
}

// FIXED: Enhanced ReadReceiptManager
class ReadReceiptManager {
  constructor() {
    this.readMessages = new Set();
    this.visibilityObserver = null;
    this.setupVisibilityTracking();
    this.setupEventListeners();
  }

  setupVisibilityTracking() {
    if ("IntersectionObserver" in window) {
      this.visibilityObserver = new IntersectionObserver(
        (entries) => {
          entries.forEach((entry) => {
            if (entry.isIntersecting && entry.intersectionRatio > 0.5) {
              const messageEl = entry.target;
              const messageId = parseInt(
                messageEl.getAttribute("data-message-id")
              );

              // CRITICAL FIX: Only mark RECEIVED messages as read
              const isFromOtherUser = messageEl.classList.contains("received");

              if (
                isFromOtherUser &&
                !this.readMessages.has(messageId) &&
                messageId > 0
              ) {
                // Mark as read after user has seen it for 1 second
                setTimeout(() => {
                  if (entry.isIntersecting && window.isUserActive) {
                    this.markMessageAsRead(messageId);
                  }
                }, 1000);
              }
            }
          });
        },
        {
          threshold: 0.5,
          rootMargin: "0px",
        }
      );
    }
  }

  setupEventListeners() {
    document.addEventListener("messagesRendered", () => {
      this.observeNewMessages();
    });
  }

  observeNewMessages() {
    if (!this.visibilityObserver) return;

    // Only observe RECEIVED messages (not sent messages)
    const messages = document.querySelectorAll(
      ".message.received:not([data-observed])"
    );

    messages.forEach((messageEl) => {
      this.visibilityObserver.observe(messageEl);
      messageEl.setAttribute("data-observed", "true");
    });
  }

  markMessageAsRead(messageId) {
    // Additional validation
    if (this.readMessages.has(messageId) || !messageId || messageId <= 0) {
      return;
    }

    this.readMessages.add(messageId);

    sendEvent("message_read", {
      message_id: messageId,
    });

    console.log(`Marked received message ${messageId} as read`);
  }

  cleanup() {
    if (this.visibilityObserver) {
      this.visibilityObserver.disconnect();
    }
    this.readMessages.clear();
  }
}

// Initialize the read receipt manager
let readReceiptManager;
document.addEventListener("DOMContentLoaded", () => {
  readReceiptManager = new ReadReceiptManager();

  // Request notification permission
  if ("Notification" in window && Notification.permission === "default") {
    Notification.requestPermission();
  }
});

// Global cleanup function
window.addEventListener("beforeunload", () => {
  if (window.currentConversationWith) {
    notifyConversationClosed();
  }
  if (readReceiptManager) {
    readReceiptManager.cleanup();
  }
});
