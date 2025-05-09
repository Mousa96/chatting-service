import React, { useState, useEffect } from "react";

function Chat() {
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState("");
  const [receiverId, setReceiverId] = useState("");
  const [file, setFile] = useState(null);
  const [broadcastMode, setBroadcastMode] = useState(false);
  const [receiverIds, setReceiverIds] = useState("");
  const [currentUserId, setCurrentUserId] = useState(null);
  const [selectedUser, setSelectedUser] = useState(null);
  const [allUsers, setAllUsers] = useState(new Set());

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (token) {
      const payload = JSON.parse(atob(token.split(".")[1]));
      setCurrentUserId(payload.user_id);
    }
    loadMessageHistory();
  }, []);

  const loadMessageHistory = async () => {
    try {
      const response = await fetch(
        "http://localhost:8080/api/messages/history",
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("token")}`,
          },
        }
      );
      if (response.ok) {
        const data = await response.json();
        setMessages(data.messages);

        // Update the list of all users
        const users = new Set();
        data.messages.forEach((msg) => {
          if (msg.sender_id !== currentUserId) users.add(msg.sender_id);
          if (msg.receiver_id !== currentUserId) users.add(msg.receiver_id);
        });
        setAllUsers(users);
      }
    } catch (error) {
      console.error("Failed to load messages:", error);
    }
  };

  const loadConversation = async (otherUserId) => {
    try {
      const response = await fetch(
        `http://localhost:8080/api/messages/conversation?user_id=${otherUserId}`,
        {
          headers: {
            Authorization: `Bearer ${localStorage.getItem("token")}`,
          },
        }
      );
      if (response.ok) {
        const data = await response.json();
        setMessages(data.messages || []);
        setSelectedUser(otherUserId);
        setReceiverId(otherUserId);
      }
    } catch (error) {
      console.error("Failed to load conversation:", error);
    }
  };

  const handleSend = async (e) => {
    e.preventDefault();
    if ((!newMessage.trim() && !file) || (!broadcastMode && !receiverId))
      return;

    const token = localStorage.getItem("token");
    try {
      let mediaUrl = "";
      if (file) {
        const formData = new FormData();
        formData.append("media", file);
        const uploadResp = await fetch(
          "http://localhost:8080/api/messages/upload",
          {
            method: "POST",
            headers: { Authorization: `Bearer ${token}` },
            body: formData,
          }
        );
        if (!uploadResp.ok) {
          const errorData = await uploadResp.text();
          console.error("Upload error:", errorData);
          throw new Error(`Failed to upload file: ${uploadResp.statusText}`);
        }
        const data = await uploadResp.json();
        mediaUrl = data.url;
      }

      const messageData = broadcastMode
        ? {
            receiver_ids: receiverIds
              .split(",")
              .map((id) => parseInt(id.trim())),
            content: newMessage.trim(),
            media_url: mediaUrl,
          }
        : {
            receiver_id: parseInt(receiverId),
            content: newMessage.trim(),
            media_url: mediaUrl,
          };

      console.log("Sending message data:", messageData);

      const endpoint = broadcastMode
        ? "http://localhost:8080/api/messages/broadcast"
        : "http://localhost:8080/api/messages/";

      const response = await fetch(endpoint, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(messageData),
      });

      if (!response.ok) {
        const errorData = await response.text();
        console.error("Server response:", errorData);
        throw new Error(`Failed to send message: ${response.statusText}`);
      }

      setNewMessage("");
      setFile(null);

      if (selectedUser) {
        await loadConversation(selectedUser);
      } else {
        await loadMessageHistory();
      }
    } catch (error) {
      console.error("Failed to send message:", error);
      alert(error.message);
    }
  };

  const handleUserSelect = (userId) => {
    setSelectedUser(userId);
    loadConversation(userId);
  };

  return (
    <div style={{ display: "flex", gap: "20px" }}>
      <div
        style={{
          width: "200px",
          borderRight: "1px solid #ccc",
          padding: "10px",
        }}
      >
        <h3>Conversations</h3>
        <button
          onClick={() => {
            setSelectedUser(null);
            loadMessageHistory();
          }}
          style={{
            width: "100%",
            padding: "8px",
            marginBottom: "8px",
            backgroundColor: !selectedUser ? "#e3f2fd" : "#fff",
          }}
        >
          All Messages
        </button>
        {Array.from(allUsers).map((userId) => (
          <button
            key={userId}
            onClick={() => handleUserSelect(userId)}
            style={{
              width: "100%",
              padding: "8px",
              marginBottom: "8px",
              backgroundColor: selectedUser === userId ? "#e3f2fd" : "#fff",
            }}
          >
            User {userId}
          </button>
        ))}
      </div>

      <div style={{ flex: 1 }}>
        <div>Your User ID: {currentUserId}</div>
        <div
          style={{
            height: "400px",
            overflowY: "scroll",
            border: "1px solid #ccc",
            padding: "10px",
            marginBottom: "10px",
          }}
        >
          {messages.map((msg) => (
            <div
              key={msg.id}
              style={{
                marginBottom: "10px",
                textAlign: msg.sender_id === currentUserId ? "right" : "left",
                backgroundColor:
                  msg.sender_id === currentUserId ? "#e3f2fd" : "#f5f5f5",
                padding: "8px",
                borderRadius: "4px",
              }}
            >
              <strong>
                {msg.sender_id === currentUserId
                  ? "You"
                  : `User ${msg.sender_id}`}
                {" → "}
                {msg.receiver_id === currentUserId
                  ? "You"
                  : `User ${msg.receiver_id}`}
              </strong>
              {msg.content && <p>{msg.content}</p>}
              {msg.media_url && (
                <div>
                  {msg.media_url.match(/\.(jpg|jpeg|png|gif)$/i) ? (
                    <img
                      src={`http://localhost:8080${msg.media_url}`}
                      alt="attachment"
                      style={{
                        maxWidth: "200px",
                        maxHeight: "200px",
                        objectFit: "contain",
                      }}
                    />
                  ) : (
                    <a
                      href={`http://localhost:8080${msg.media_url}`}
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      Download Attachment
                    </a>
                  )}
                </div>
              )}
              <small style={{ color: "#666" }}>
                {new Date(msg.created_at).toLocaleString()}
              </small>
            </div>
          ))}
        </div>

        <form onSubmit={handleSend}>
          <div>
            <label>
              <input
                type="checkbox"
                checked={broadcastMode}
                onChange={(e) => setBroadcastMode(e.target.checked)}
              />
              Broadcast Mode
            </label>
          </div>

          {broadcastMode ? (
            <input
              type="text"
              value={receiverIds}
              onChange={(e) => setReceiverIds(e.target.value)}
              placeholder="Receiver IDs (comma-separated)"
            />
          ) : (
            <input
              type="number"
              value={receiverId}
              onChange={(e) => setReceiverId(e.target.value)}
              placeholder="Receiver ID"
            />
          )}

          <input
            type="text"
            value={newMessage}
            onChange={(e) => setNewMessage(e.target.value)}
            placeholder="Message"
          />

          <input type="file" onChange={(e) => setFile(e.target.files[0])} />

          <button type="submit">Send</button>
        </form>
      </div>
    </div>
  );
}

export default Chat;
