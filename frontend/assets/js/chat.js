document.addEventListener("DOMContentLoaded", function () {
  // Check authentication
  const token = localStorage.getItem("token");
  const currentUserId = localStorage.getItem("userId");
  window.currentUserId = parseInt(currentUserId);
  console.log("window.currentUserId", window.currentUserId);
  const currentUsername = localStorage.getItem("username");

  if (!token || !currentUserId) {
    window.location.href = "index.html";
    return;
  }

  // Set current user display
  document.getElementById("current-user").textContent = currentUsername;

  // Global state
  let selectedUserId = null;
  let allUsers = [];
  let uploadedMediaFile = null;
  let broadcastMediaFile = null;

  // Make these globally accessible
  window.selectedUserId = selectedUserId;
  window.allUsers = allUsers;
  window.currentUserId = currentUserId;

  // DOM elements
  const userList = document.getElementById("user-list");
  const onlineUsersCount = document.getElementById("online-count");
  const messagesContainer = document.getElementById("messages-container");
  const messageForm = document.getElementById("message-form");
  const messageInput = document.getElementById("message-text");
  const sendButton = document.getElementById("send-button");
  const chatTitle = document.getElementById("chat-title");
  const userStatus = document.getElementById("selected-user-status");
  const mediaUpload = document.getElementById("media-upload");
  const mediaPreview = document.getElementById("media-preview");
  const logoutBtn = document.getElementById("logout-btn");
  const broadcastMessage = document.getElementById("broadcast-message");
  const broadcastFile = document.getElementById("broadcast-file");
  const sendBroadcastBtn = document.getElementById("send-broadcast");

  let socket = null;
  let reconnectTimeout = null;
  let isConnected = false;

  function connectWebSocket() {
    // Close existing socket if any
    if (socket && socket.readyState !== WebSocket.CLOSED) {
      socket.close();
    }

    console.log("Connecting to WebSocket...");
    socket = new WebSocket(`ws://${window.location.host}/ws?token=${token}`);
    window.socket = socket;

    socket.onopen = function () {
      console.log("WebSocket connection opened");
      isConnected = true;

      // Clear any reconnect timeout
      if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
        reconnectTimeout = null;
      }

      // Fetch initial data after connection
      fetchUsers();
      // Request current online users after a short delay
      setTimeout(() => {
        requestOnlineUsers();
      }, 500);
    };

    socket.onclose = function (event) {
      console.log("WebSocket connection closed:", event.code, event.reason);
      isConnected = false;

      // Only reconnect for unexpected closures
      if (event.code !== 1000 && event.code !== 1001) {
        // Attempt to reconnect after 3 seconds for unexpected closures
        console.log("Attempting to reconnect in 3 seconds...");
        reconnectTimeout = setTimeout(connectWebSocket, 3000);
      }
    };

    socket.onerror = function (error) {
      console.error("WebSocket error:", error);
      isConnected = false;
    };

    socket.onmessage = function (event) {
      try {
        console.log("Raw WebSocket message:", event.data);

        // Check if the message starts with <!DOCTYPE - indicates HTML error page
        if (
          typeof event.data === "string" &&
          event.data.trim().startsWith("<!DOCTYPE")
        ) {
          console.error("Received HTML instead of JSON from WebSocket");
          return;
        }

        const data = JSON.parse(event.data);
        console.log("Parsed WebSocket message:", data);

        // Route the event
        routeEvent(data);
      } catch (e) {
        console.error("Error processing WebSocket message:", e);
        console.log("Raw message data:", event.data);
      }
    };
  }
  // Request online users
  function requestOnlineUsers() {
    if (isConnected && socket && socket.readyState === WebSocket.OPEN) {
      console.log("Requesting online users...");
      sendEvent("get_online_users", {});
    }
  }
  // Fetch users
  async function fetchUsers() {
    try {
      const response = await fetch("/api/users", {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();

        if (data && Array.isArray(data)) {
          allUsers = data.filter((user) => user.id !== parseInt(currentUserId));
          allUsers.forEach((user) => {
            user.status = "offline"; // Default to offline until WebSocket updates
          });
          window.allUsers = allUsers;

          // Clear existing user list
          userList.innerHTML = "";

          // Populate user list
          allUsers.forEach((user) => {
            const userElement = createUserListItem(user);
            userList.appendChild(userElement);
          });

          // Populate broadcast users after fetching user list
          populateBroadcastUsers();

          // Update online count
          updateOnlineCount();
        } else {
          console.error("Invalid user data format:", data);
          allUsers = [];
        }
      } else {
        console.error("Failed to fetch users:", response.status);
      }
    } catch (error) {
      console.error("Error fetching users:", error);
      allUsers = [];
    }
  }

  // Helper function to create a user list item
  function createUserListItem(user) {
    const li = document.createElement("li");
    li.className = "user-item";
    li.dataset.userId = user.id;
    li.id = `user-${user.id}`;

    const statusClass = user.status === "online" ? "online" : "offline";

    li.innerHTML = `
      <span class="status-dot ${statusClass}"></span>
      <span class="user-name">${user.username}</span>
    `;

    li.addEventListener("click", () => selectUser(user.id));

    return li;
  }

  // Update online count
  function updateOnlineCount() {
    if (!onlineUsersCount) return;
    const onlineUsers = allUsers.filter((user) => user.status === "online");
    onlineUsersCount.textContent = `${onlineUsers.length} online`;
  }

  // Populate broadcast users with improved UI
  function populateBroadcastUsers() {
    const container = document.getElementById("broadcast-users-container");
    if (!container) return;

    container.innerHTML = "";

    allUsers.forEach((user) => {
      const userItem = document.createElement("div");
      userItem.className = "broadcast-user-item";

      const statusClass = user.status === "online" ? "online" : "offline";

      userItem.innerHTML = `
        <input type="checkbox" id="broadcast-user-${user.id}" value="${user.id}">
        <label for="broadcast-user-${user.id}">
          ${user.username}
        </label>
        <span class="status-dot ${statusClass}"></span>
      `;

      userItem.addEventListener("click", (e) => {
        if (e.target.type !== "checkbox") {
          const checkbox = userItem.querySelector("input[type='checkbox']");
          checkbox.checked = !checkbox.checked;
        }
        updateSelectedCount();
      });

      container.appendChild(userItem);
    });

    updateSelectedCount();
  }

  // Update selected count display
  function updateSelectedCount() {
    const countElement = document.getElementById("selected-count");
    if (!countElement) return;

    const checkboxes = document.querySelectorAll(
      "#broadcast-users-container input[type='checkbox']:checked"
    );
    const count = checkboxes.length;

    countElement.textContent = `${count} user${
      count !== 1 ? "s" : ""
    } selected`;
    countElement.className =
      count > 0 ? "selected-count has-selection" : "selected-count";
  }

  // Select a user and load conversation
  function selectUser(userId) {
    console.log(`Selecting user with ID: ${userId}`);
    const user = allUsers.find((u) => u.id === parseInt(userId));

    if (!user) {
      console.error(`User with ID ${userId} not found in allUsers:`, allUsers);
      return;
    }

    // Notify about conversation change
    // if (
    //   window.currentConversationWith &&
    //   window.currentConversationWith !== user.id
    // ) {
    //   notifyConversationClosed();
    // }

    selectedUserId = user.id;
    window.selectedUserId = selectedUserId;

    // Notify backend about conversation opened
    //notifyConversationOpened(user.id);

    // Update UI
    document.querySelectorAll(".user-item").forEach((item) => {
      item.classList.remove("active", "new-message");
      if (parseInt(item.dataset.userId) === user.id) {
        item.classList.add("active");
      }
    });

    chatTitle.textContent = user.username;

    // Update status indicator
    const statusDot = userStatus.querySelector(".status-dot");
    const statusText = userStatus.querySelector(".status-text");

    if (statusDot && statusText) {
      statusDot.className = `status-dot ${user.status || "offline"}`;
      statusText.textContent = user.status === "online" ? "Online" : "Offline";
    }

    // Enable message input
    messageInput.disabled = false;
    sendButton.disabled = false;

    // Load conversation
    loadConversation(user.id);
  }

  // Load conversation
  async function loadConversation(userId) {
    try {
      messagesContainer.innerHTML =
        '<div class="loading">Loading messages...</div>';

      const url = `/api/messages/conversation?user_id=${userId}`;
      console.log(`Fetching conversation from: ${url}`);

      const response = await fetch(url, {
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
      });

      if (!response.ok) {
        if (response.status === 401) {
          console.log("Unauthorized - redirecting to login");
          localStorage.removeItem("token");
          localStorage.removeItem("userId");
          localStorage.removeItem("username");
          window.location.href = "index.html";
          return;
        }

        try {
          const errorData = await response.json();
          throw new Error(
            errorData.error || `Server error: ${response.status}`
          );
        } catch (parseError) {
          throw new Error(`Server error: ${response.status}`);
        }
      }

      const data = await response.json();
      console.log(`Received conversation data:`, data);

      // Ensure all messages have a status field before rendering
      const messagesWithStatus = (data.messages || []).map((message) => ({
        ...message,
        status: message.status || "sent", // Default to "sent" if no status
      }));

      renderMessages(messagesWithStatus);
    } catch (error) {
      console.error("Error loading conversation:", error);
      messagesContainer.innerHTML = `<div class="error">Failed to load messages: ${error.message}</div>`;
    }
  }

  // Add message to conversation without full reload
  function addMessageToConversation(message) {
    const messageEl = createMessageElement(message);
    messagesContainer.appendChild(messageEl);

    // Scroll to bottom
    messagesContainer.scrollTo(0, messagesContainer.scrollHeight);

    // Notify that new messages were rendered
    setTimeout(() => {
      document.dispatchEvent(new CustomEvent("messagesRendered"));
    }, 100);
  }

  // Create message element
  function createMessageElement(message) {
    const messageEl = document.createElement("div");
    const isSentByMe = message.sender_id == window.currentUserId;
    console.log("Message is sent by me:", isSentByMe);
    console.log("sender_id:", message.sender_id);
    console.log("currentUserId:", window.currentUserId);
    messageEl.className = `message ${isSentByMe ? "sent" : "received"}`;
    messageEl.setAttribute("data-message-id", message.id);

    // Message content
    const contentEl = document.createElement("div");
    contentEl.className = "message-content";
    contentEl.textContent = message.content || "";
    messageEl.appendChild(contentEl);

    // Media if present
    if (message.media_url) {
      const mediaEl = document.createElement("div");

      if (message.media_url.match(/\.(jpeg|jpg|gif|png)$/i)) {
        const img = document.createElement("img");
        img.src = message.media_url;
        img.className = "message-media";
        mediaEl.appendChild(img);
      } else if (message.media_url.match(/\.(mp4|webm|ogg)$/i)) {
        const video = document.createElement("video");
        video.src = message.media_url;
        video.className = "message-media";
        video.controls = true;
        mediaEl.appendChild(video);
      } else {
        const link = document.createElement("a");
        link.href = message.media_url;
        link.textContent = "Download attachment";
        link.target = "_blank";
        mediaEl.appendChild(link);
      }
      messageEl.appendChild(mediaEl);
    }

    // Message meta section (timestamp + status)
    const metaEl = document.createElement("div");
    metaEl.className = "message-meta";

    const timeEl = document.createElement("div");
    timeEl.className = "message-time";
    const date = new Date(message.created_at);
    timeEl.textContent = date.toLocaleTimeString([], {
      hour: "2-digit",
      minute: "2-digit",
    });
    metaEl.appendChild(timeEl);

    // ALWAYS add status indicator for sent messages
    if (isSentByMe) {
      const statusEl = document.createElement("div");
      // Use the status from the message, or default to "sent"
      const messageStatus = message.status || "sent";
      statusEl.className = `message-status ${messageStatus}`;
      statusEl.textContent = getStatusText(messageStatus);
      statusEl.title = getStatusTitle(messageStatus);
      metaEl.appendChild(statusEl);

      console.log(
        `Created message ${message.id} with status: ${messageStatus}`
      );
    }

    messageEl.appendChild(metaEl);
    return messageEl;
  }

  // Make addMessageToConversation globally accessible
  window.addMessageToConversation = addMessageToConversation;

  // Render messages
  function renderMessages(messages) {
    messagesContainer.innerHTML = "";

    if (!messages || messages.length === 0) {
      messagesContainer.innerHTML =
        '<div class="no-messages">No messages yet</div>';
      return;
    }

    messages.forEach((message) => {
      const messageEl = createMessageElement(message);
      messagesContainer.appendChild(messageEl);
    });

    // Scroll to bottom
    messagesContainer.scrollTop = messagesContainer.scrollHeight;

    // Notify that messages were rendered
    setTimeout(() => {
      document.dispatchEvent(new CustomEvent("messagesRendered"));
    }, 100);
  }

  // Handle media selection
  async function handleMediaUpload(file, isForBroadcast = false) {
    if (!file) return;

    // Validate file type and size
    const validTypes = ["image/jpeg", "image/png", "image/gif", "video/mp4"];
    if (!validTypes.includes(file.type)) {
      alert("Unsupported file type. Please upload an image or MP4 video.");
      return;
    }

    if (file.size > 10 * 1024 * 1024) {
      alert("File too large. Maximum size is 10MB.");
      return;
    }

    // Store the file and show indicator
    if (isForBroadcast) {
      broadcastMediaFile = file;
      showMediaIndicator(file, true);
    } else {
      uploadedMediaFile = file;
      showMediaIndicator(file, false);
    }
  }

  // Show media indicator
  function showMediaIndicator(file, isForBroadcast) {
    const fileName = file.name;
    const fileSize = (file.size / 1024 / 1024).toFixed(2);
    const isImage = file.type.startsWith("image/");
    const isVideo = file.type.startsWith("video/");

    const icon = isImage ? "🖼️" : isVideo ? "🎥" : "📎";

    const indicatorHTML = `
      <div class="media-indicator">
        <span class="media-icon">${icon}</span>
        <span class="media-info">${fileName} (${fileSize} MB)</span>
        <button class="remove-media-btn" onclick="removeMediaFile(${isForBroadcast})">✕</button>
      </div>
    `;

    if (isForBroadcast) {
      const broadcastContainer = document.querySelector(".broadcast-actions");
      if (!broadcastContainer) return;

      let indicatorContainer = broadcastContainer.querySelector(
        ".broadcast-media-indicator"
      );

      if (!indicatorContainer) {
        indicatorContainer = document.createElement("div");
        indicatorContainer.className = "broadcast-media-indicator";
        broadcastContainer.appendChild(indicatorContainer);
      }

      indicatorContainer.innerHTML = indicatorHTML;
    } else {
      let indicatorContainer = document.querySelector(
        ".message-media-indicator"
      );

      if (!indicatorContainer) {
        indicatorContainer = document.createElement("div");
        indicatorContainer.className = "message-media-indicator";
        messageForm.parentNode.insertBefore(
          indicatorContainer,
          messageForm.nextSibling
        );
      }

      indicatorContainer.innerHTML = indicatorHTML;
    }
  }

  // Remove media file and indicator
  function removeMediaFile(isForBroadcast) {
    if (isForBroadcast) {
      broadcastMediaFile = null;
      const indicator = document.querySelector(".broadcast-media-indicator");
      if (indicator) {
        indicator.innerHTML = "";
      }
      broadcastFile.value = "";
    } else {
      uploadedMediaFile = null;
      const indicator = document.querySelector(".message-media-indicator");
      if (indicator) {
        indicator.innerHTML = "";
      }
      mediaUpload.value = "";
    }
  }

  // Make removeMediaFile globally accessible
  window.removeMediaFile = removeMediaFile;

  // Upload media file to server
  async function uploadMediaFile(file) {
    const formData = new FormData();
    formData.append("file", file);

    const response = await fetch("/api/messages/upload", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
      },
      body: formData,
    });

    if (!response.ok) {
      if (response.status === 401) {
        localStorage.removeItem("token");
        localStorage.removeItem("userId");
        localStorage.removeItem("username");
        window.location.href = "index.html";
        return null;
      }
      throw new Error("Failed to upload file");
    }

    const data = await response.json();
    return data.url;
  }

  // Send message via WebSocket - FIXED VERSION
  async function sendMessage(e) {
    e.preventDefault();

    const content = messageInput.value.trim();

    if (!content && !uploadedMediaFile) {
      return;
    }

    if (!selectedUserId) {
      alert("Please select a user to send a message");
      return;
    }

    if (!isConnected) {
      alert("WebSocket is not connected. Please wait and try again.");
      return;
    }

    try {
      let mediaUrl = null;

      // Upload media file if present
      if (uploadedMediaFile) {
        try {
          mediaUrl = await uploadMediaFile(uploadedMediaFile);
        } catch (error) {
          alert("Error uploading media: " + error.message);
          return;
        }
      }

      console.log("Sending message via WebSocket:", {
        type: "send_message",
        payload: {
          message: content,
          to: selectedUserId,
          media_url: mediaUrl,
        },
      });

      // Send via WebSocket with correct format
      const event = {
        type: "send_message",
        payload: {
          message: content,
          to: selectedUserId,
          media_url: mediaUrl || "",
        },
      };

      if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify(event));

        // Clear input and media file
        messageInput.value = "";
        uploadedMediaFile = null;

        // Clear media indicator and file input
        const indicator = document.querySelector(".message-media-indicator");
        if (indicator) {
          indicator.innerHTML = "";
        }
        mediaUpload.value = "";
      } else {
        alert("WebSocket is not connected");
      }
    } catch (error) {
      console.error("Error sending message:", error);
      alert("Error sending message: " + error.message);
    }
  }

  // Send broadcast via WebSocket
  async function sendBroadcast() {
    const selectedCheckboxes = document.querySelectorAll(
      "#broadcast-users-container input[type='checkbox']:checked"
    );
    const receiverIds = Array.from(selectedCheckboxes).map((cb) =>
      parseInt(cb.value)
    );

    const content = broadcastMessage.value.trim();

    if (receiverIds.length === 0) {
      alert("Please select at least one user");
      return;
    }

    if (!content && !broadcastMediaFile) {
      alert("Please enter a message or attach media");
      return;
    }

    if (!isConnected) {
      alert("WebSocket is not connected. Please wait and try again.");
      return;
    }

    try {
      let mediaUrl = null;

      if (broadcastMediaFile) {
        try {
          mediaUrl = await uploadMediaFile(broadcastMediaFile);
        } catch (error) {
          alert("Error uploading media: " + error.message);
          return;
        }
      }

      console.log("Sending broadcast via WebSocket:", {
        type: "broadcast_message",
        payload: {
          message: content,
          receiver_ids: receiverIds,
          media_url: mediaUrl,
        },
      });

      // Send via WebSocket with correct format
      const event = {
        type: "broadcast_message",
        payload: {
          message: content,
          receiver_ids: receiverIds,
          media_url: mediaUrl || "",
        },
      };

      if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify(event));

        // Clear form
        broadcastMessage.value = "";
        broadcastMediaFile = null;

        // Clear selections
        const checkboxes = document.querySelectorAll(
          "#broadcast-users-container input[type='checkbox']"
        );
        checkboxes.forEach((cb) => (cb.checked = false));
        updateSelectedCount();

        // Clear media indicator and file input
        const indicator = document.querySelector(".broadcast-media-indicator");
        if (indicator) {
          indicator.innerHTML = "";
        }
        broadcastFile.value = "";

        alert("Broadcast message sent successfully!");
      } else {
        alert("WebSocket is not connected");
      }
    } catch (error) {
      console.error("Error sending broadcast:", error);
      alert("Error sending broadcast: " + error.message);
    }
  }

  // Logout function
  function logout() {
    if (window.currentConversationWith) {
      notifyConversationClosed();
    }
    if (socket) {
      isConnected = false; // Prevent reconnection
      socket.close(1000, "User logout"); // Normal closure
    }
    if (reconnectTimeout) {
      clearTimeout(reconnectTimeout);
    }
    localStorage.removeItem("token");
    localStorage.removeItem("userId");
    localStorage.removeItem("username");
    window.location.href = "index.html";
  }

  // Event listeners
  messageForm.addEventListener("submit", sendMessage);
  mediaUpload.addEventListener("change", (e) =>
    handleMediaUpload(e.target.files[0])
  );
  broadcastFile.addEventListener("change", (e) =>
    handleMediaUpload(e.target.files[0], true)
  );
  sendBroadcastBtn.addEventListener("click", sendBroadcast);
  logoutBtn.addEventListener("click", logout);

  // Event listeners for broadcast UI
  const selectAllBtn = document.getElementById("select-all-btn");
  const selectNoneBtn = document.getElementById("select-none-btn");

  if (selectAllBtn) {
    selectAllBtn.addEventListener("click", () => {
      const checkboxes = document.querySelectorAll(
        "#broadcast-users-container input[type='checkbox']"
      );
      checkboxes.forEach((cb) => (cb.checked = true));
      updateSelectedCount();
    });
  }

  if (selectNoneBtn) {
    selectNoneBtn.addEventListener("click", () => {
      const checkboxes = document.querySelectorAll(
        "#broadcast-users-container input[type='checkbox']"
      );
      checkboxes.forEach((cb) => (cb.checked = false));
      updateSelectedCount();
    });
  }

  // Add cleanup on page unload to prevent reconnection attempts
  window.addEventListener("beforeunload", () => {
    if (socket) {
      isConnected = false; // Prevent reconnection
      socket.close(1000, "Page unload"); // Normal closure
    }
    if (reconnectTimeout) {
      clearTimeout(reconnectTimeout);
    }
  });

  // Make functions globally accessible
  window.addMessageToConversation = addMessageToConversation;
  window.loadConversation = loadConversation;
  window.selectUser = selectUser;
  window.updateOnlineCount = updateOnlineCount;

  // Initialize WebSocket connection
  connectWebSocket();
});

// Helper functions for message status
function getStatusText(status) {
  switch (status) {
    case "sent":
      return "✓";
    case "delivered":
      return "✓✓";
    case "read":
      return "✓✓";
    default:
      return "";
  }
}

function getStatusTitle(status) {
  switch (status) {
    case "sent":
      return "Message sent";
    case "delivered":
      return "Message delivered";
    case "read":
      return "Message read";
    default:
      return "";
  }
}
