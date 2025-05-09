import React, { useState } from "react";
import "../styles/Message.css";

function Message({ message, currentUser }) {
  // Check if the current user is the sender
  const isCurrentUser = message.sender_id === currentUser;

  // State to track if the image failed to load
  const [imageError, setImageError] = useState(false);

  // Create the correct URL to the backend server
  const getImageUrl = () => {
    if (!message.media_url) return "";

    // Extract just the filename (last part after any slash)
    const filename = message.media_url.split("/").pop();

    // Use the correct API endpoint
    return `http://localhost:8080/api/uploads/${filename}`;
  };

  // Handle image loading errors
  const handleImageError = () => {
    console.error(`Error loading image: ${getImageUrl()}`);
    console.log(`Media URL from DB: ${message.media_url}`);
    console.log(`Full message object: `, message);
    setImageError(true);
  };

  // Format the status text explicitly
  const formatStatus = (status) => {
    switch (status) {
      case "sent":
        return "Sent ✓";
      case "delivered":
        return "Delivered ✓✓";
      case "read":
        return "Read ✓✓";
      default:
        return status;
    }
  };

  return (
    <div className={`message ${isCurrentUser ? "own-message" : ""}`}>
      <div className="message-content">
        {message.content && <p className="message-text">{message.content}</p>}

        {message.media_url && (
          <div className="message-media">
            {imageError ? (
              <div className="image-placeholder">
                <p>Image unavailable</p>
                <small className="debug-info">
                  Media URL: {message.media_url}
                </small>
              </div>
            ) : (
              <img
                src={getImageUrl()}
                alt="Message media"
                onError={handleImageError}
              />
            )}
          </div>
        )}

        <div className="message-meta">
          <span className="message-time">
            {new Date(message.created_at).toLocaleTimeString([], {
              hour: "2-digit",
              minute: "2-digit",
            })}
          </span>
          {message.status && (
            <span className={`status-text ${message.status}`}>
              • {formatStatus(message.status)}
            </span>
          )}
        </div>
      </div>
    </div>
  );
}

export default Message;
