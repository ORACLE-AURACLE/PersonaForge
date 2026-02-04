import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { authenticateWithGoogle } from "../apis/api";
import logo from "../assets/images/Main-Logo.svg";

export default function GoogleAuth() {
  const [idToken, setIdToken] = useState("");
  const [sessionId, setSessionId] = useState(
    localStorage.getItem("session_id") || "",
  );
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const response = await authenticateWithGoogle(idToken, sessionId || null);
      if (response.success === false) {
        setError(response.message);
      } else {
        // Assuming response includes token or session_id
        if (response.token) {
          localStorage.setItem("token", response.token);
        }
        if (response.session_id) {
          localStorage.setItem("session_id", response.session_id);
        }
        navigate("/personas"); // Redirect to personas after successful auth
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <section className="auth-page">
      <div className="auth-container">
        <img src={logo} alt="PersonaForge" className="logo" />
        <h1>Google Authentication</h1>
        <p>Enter your Google ID token to authenticate.</p>

        <form onSubmit={handleSubmit} className="auth-form">
          <div className="form-group">
            <label htmlFor="idToken">ID Token:</label>
            <input
              type="text"
              id="idToken"
              value={idToken}
              onChange={(e) => setIdToken(e.target.value)}
              required
              placeholder="Enter your Google ID token"
            />
          </div>

          <div className="form-group">
            <label htmlFor="sessionId">Session ID (optional):</label>
            <input
              type="text"
              id="sessionId"
              value={sessionId}
              onChange={(e) => setSessionId(e.target.value)}
              placeholder="Leave blank to use current session"
            />
          </div>

          <button type="submit" disabled={loading}>
            {loading ? "Authenticating..." : "Authenticate"}
          </button>
        </form>

        {error && (
          <div className="error-message">
            <p>Error: {error}</p>
          </div>
        )}

        <button className="back-btn" onClick={() => navigate(-1)}>
          ← Back
        </button>
      </div>
    </section>
  );
}
