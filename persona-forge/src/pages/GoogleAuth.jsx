import { useNavigate } from "react-router-dom";
import { GoogleLogin } from "@react-oauth/google";
import { authenticateWithGoogle } from "../apis/api";
import logo from "../assets/images/Main-Logo.svg";

export default function GoogleAuth() {
  const handleAnonymousSession = () => {
  let sessionId = localStorage.getItem("session_id");

  if (!sessionId) {
    sessionId = crypto.randomUUID();
    localStorage.setItem("session_id", sessionId);
  }

  navigate("/personas");
};

  const navigate = useNavigate();

  const handleGoogleSuccess = async (credentialResponse) => {
    try {
      const idToken = credentialResponse.credential;
      const sessionId = localStorage.getItem("session_id");

      const response = await authenticateWithGoogle(idToken, sessionId);

      if (response.success === false) {
        throw new Error(response.message);
      }

      if (response.token) {
        localStorage.setItem("token", response.token);
      }

      if (response.session_id) {
        localStorage.setItem("session_id", response.session_id);
      }

      navigate("/personas");
    } catch (err) {
      console.error("Google auth failed:", err);
    }
  };

  return (
    <section className="auth-page">
      <div className="auth-container">
        <img src={logo} alt="PersonaForge" className="logo" />
        <h1>Continue with Google</h1>

        <GoogleLogin
          onSuccess={handleGoogleSuccess}
          onError={() => console.log("Google Login Failed")}
        />
        <div className="auth-divider">
  <span>or</span>
</div>

<button
  className="anonymous-btn"
  onClick={handleAnonymousSession}
>
  Create Anonymous Session
</button>

      </div>
    </section>
  );
}
