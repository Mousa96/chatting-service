import React, { useState } from "react";
import Login from "./components/Login";
import Chat from "./components/Chat";

function App() {
  const [isLoggedIn, setIsLoggedIn] = useState(!!localStorage.getItem("token"));

  const handleLogin = () => {
    setIsLoggedIn(true);
  };

  const handleLogout = () => {
    localStorage.removeItem("token");
    setIsLoggedIn(false);
  };

  return (
    <div style={{ maxWidth: "800px", margin: "0 auto", padding: "20px" }}>
      {!isLoggedIn ? (
        <Login onLogin={handleLogin} />
      ) : (
        <div>
          <button onClick={handleLogout} style={{ marginBottom: "20px" }}>
            Logout
          </button>
          <Chat />
        </div>
      )}
    </div>
  );
}

export default App;
