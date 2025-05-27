document.addEventListener("DOMContentLoaded", function () {
  // Check if user is already logged in
  const token = localStorage.getItem("token");
  if (token) {
    window.location.href = "chat.html";
    return;
  } else {
    console.log("token not found");
  }

  // Tab switching
  const tabs = document.querySelectorAll(".tab");
  const tabContents = document.querySelectorAll(".tab-content");

  tabs.forEach((tab) => {
    tab.addEventListener("click", () => {
      const tabId = tab.getAttribute("data-tab");

      // Update active tab
      tabs.forEach((t) => t.classList.remove("active"));
      tab.classList.add("active");

      // Show corresponding form
      tabContents.forEach((content) => {
        content.classList.add("hidden");
        if (content.id === `${tabId}-form`) {
          content.classList.remove("hidden");
        }
      });
    });
  });

  // Login form submission
  document
    .getElementById("login-form")
    .addEventListener("submit", async (e) => {
      e.preventDefault();

      const username = document.getElementById("login-username").value;
      const password = document.getElementById("login-password").value;

      try {
        const response = await fetch("/api/auth/login", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ username, password }),
        });

        if (!response.ok) {
          throw new Error("Login failed");
        }

        const data = await response.json();
        console.log("data", data);
        localStorage.setItem("token", data.token);
        localStorage.setItem("userId", data.user.id);
        localStorage.setItem("username", username);

        window.location.href = "chat.html";
      } catch (error) {
        alert("Login failed: " + error.message);
      }
    });

  // Register form submission
  document
    .getElementById("register-form")
    .addEventListener("submit", async (e) => {
      e.preventDefault();
      const username = document.getElementById("register-username").value;
      const password = document.getElementById("register-password").value;

      try {
        const response = await fetch("/api/auth/register", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ username, password }),
        });

        if (!response.ok) {
          throw new Error("Registration failed");
        }

        // Show login tab after successful registration
        document.querySelector('[data-tab="login"]').click();
        alert("Registration successful! Please login.");
      } catch (error) {
        alert("Registration failed: " + error.message);
      }
    });
});
