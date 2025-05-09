export const updateMessageStatus = async (messageId, status) => {
  try {
    console.log("Updating message status:", { messageId, status });
    const response = await fetch("/api/messages/status", {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${localStorage.getItem("token")}`,
      },
      body: JSON.stringify({
        message_id: messageId,
        status: status,
      }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      console.error("Server response:", response.status, errorText);
      throw new Error("Failed to update message status");
    }

    // Check if there's a response body before parsing
    const contentType = response.headers.get("content-type");
    if (contentType && contentType.includes("application/json")) {
      return await response.json();
    }
    return { status: "success" };
  } catch (error) {
    console.error("Error updating message status:", error);
    throw error;
  }
};
