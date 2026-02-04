import { BrowserRouter, Routes, Route } from "react-router-dom";
import { useState, useEffect, createContext } from "react";
import Home from "./pages/Home";
import Personas from "./pages/Personas";
import PersonaChat from "./pages/PersonaChat";
import GoogleAuth from "./pages/GoogleAuth";
import CreatePersona from "./pages/CreatePersona";
import { createAnonymousSession } from "./apis/api";
// index.js or App.js
import 'bootstrap/dist/css/bootstrap.min.css';
import "./App.css";

export const SessionContext = createContext();

function App() {
  const [sessionId, setSessionId] = useState(() => {
    const stored = localStorage.getItem("session_id");
    return stored && stored !== "undefined" ? stored : null;
  });
  const [error, setError] = useState(null);

  useEffect(() => {
  const initializeSession = async () => {
    if (!sessionId) {
      try {
        const response = await createAnonymousSession();

        if (response.status === "success") {
          const newSessionId = response.data.session_id; // <-- fix here
          setSessionId(newSessionId);
          localStorage.setItem("session_id", newSessionId);
        } else {
          setError(response.message || "Failed to create session");
        }
      } catch (err) {
        setError(err.message);
      }
    }
  };
  initializeSession();
}, [sessionId]);

  return (
    <SessionContext.Provider value={{ sessionId, setSessionId }}>
      <BrowserRouter>
        {error && (
          <div className="error-banner">
            <p>Error: {error}</p>
            <button onClick={() => setError(null)}>Dismiss</button>
          </div>
        )}
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/personas" element={<Personas />} />
          <Route path="/personas/:id" element={<PersonaChat />} />
          <Route path="/auth/google" element={<GoogleAuth />} />
          <Route path="/personas/create" element={<CreatePersona />} />
        </Routes>
      </BrowserRouter>
    </SessionContext.Provider>
  );
}

export default App;
