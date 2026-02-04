import { useParams, useNavigate } from "react-router-dom";
import { useState, useEffect, useContext } from "react";
import {
  getPersonaById,
  sendMessage,
  getChatHistory,
  getInsights,
} from "../apis/api";
import { SessionContext } from "../App";
import logo from "../assets/images/Main-Logo.svg";
import identifier from "../assets/images/Background.svg";
import "../App.css";

export default function PersonaChat() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { sessionId } = useContext(SessionContext);

  const [persona, setPersona] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const [activeTab, setActiveTab] = useState("chat");
  const [message, setMessage] = useState("");
  const [messages, setMessages] = useState([]);
  const [insights, setInsights] = useState([]);
  const [sending, setSending] = useState(false);

  const isAnonymous = !localStorage.getItem("token");

  const personaId = parseInt(id, 10);
  const [isMobile, setIsMobile] = useState(window.innerWidth <= 768);

  useEffect(() => {
    const onResize = () => setIsMobile(window.innerWidth <= 768);
    window.addEventListener("resize", onResize);
    return () => window.removeEventListener("resize", onResize);
  }, []);

  // Fetch persona and chat history
  useEffect(() => {
    const fetchData = async () => {
      try {
        const personaResponse = await getPersonaById(id);
        if (personaResponse?.status !== "success") {
          setError(personaResponse?.message || "Failed to fetch persona");
          return;
        }

        setPersona(personaResponse.data);

        if (!isAnonymous) {
          const historyResponse = await getChatHistory(id);
          if (historyResponse?.success === false) {
            setError(historyResponse.message);
          } else {
            setMessages(
              Array.isArray(historyResponse.data) ? historyResponse.data : [],
            );
          }
        }
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, [id, isAnonymous]);

  const handleSend = async () => {
    if (!message.trim() || !sessionId) return;

    setSending(true);
    setError(null);

    try {
      const requestBody = {
        message,
        persona_id: personaId,
        session_id: sessionId,
      };
      const response = await sendMessage(requestBody);

      if (response?.status !== "success") {
        setError(response?.message || "Failed to send message");
        return;
      }

      const assistantMessage = response.data.message.content;

      // Add user message immediately for chat-like feel
      setMessages((prev) => [
        ...prev,
        { text: message, sender: "user" },
        { text: assistantMessage, sender: "persona" },
      ]);

      setMessage("");

      if (!isAnonymous) {
        const insightsResponse = await getInsights(sessionId);
        if (insightsResponse?.success !== false) {
          setInsights(
            Array.isArray(insightsResponse.data) ? insightsResponse.data : [],
          );
        }
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setSending(false);
    }
  };

  if (loading) return <p>Loading persona...</p>;
  if (error) return <p>Error: {error}</p>;
  if (!persona) return <p>Persona not found.</p>;

  const firstName = persona?.name?.split(" ")[0] || "this persona";

  // Helper to render structured messages with headings, subtitles, bullets
  const formatMessage = (text) =>
    text.split("\n").map((line, idx) => {
      if (line.startsWith("## "))
        return <h4 key={idx}>{line.replace("## ", "")}</h4>;
      if (line.startsWith("# "))
        return <h3 key={idx}>{line.replace("# ", "")}</h3>;
      if (line.startsWith("- "))
        return <li key={idx}>{line.replace("- ", "")}</li>;
      return <p key={idx}>{line}</p>;
    });

  return (
    <section className="chat-page">
      {/* HEADER */}
      <div className="personaTop">
        <img src={logo} alt="PersonaForge" className="logo" />
        <button className="back-btn" onClick={() => navigate(-1)}>
          ← Change Persona
        </button>
      </div>

      <div className="headCenterHero" style={{ color: "black" }}>
        <div className="headerCenter">
          <h3>{persona.name}</h3>
          {/* <p>{persona.blueprint?.description || "No description available."}</p> */}
        </div>
      </div>

      {/* MOBILE TOGGLE */}
      <div className="mobile-toggle">
        <div className="toggle-pill">
          <button
            className={`toggle-btn ${activeTab === "chat" ? "active" : ""}`}
            onClick={() => setActiveTab("chat")}
          >
            Chat
          </button>

          <button
            className={`toggle-btn ${activeTab === "insights" ? "active" : ""}`}
            onClick={() => setActiveTab("insights")}
          >
            Insights
            {/* optional notification dot */}
            {insights.length > 0 && <span className="insight-dot" />}
          </button>
        </div>
      </div>

      {/* BODY */}
      <div className="chat-body">
        {/* CHAT PANEL */}
        {(!isMobile || activeTab === "chat") && (
          <div className="chat-panel">
            {messages.length === 0 ? (
              <div className="chat-empty">
                <p>Start a conversation with {firstName}</p>
                <span>
                  Ask about their needs, pitch your idea, or explore how they
                  make decisions.
                </span>
              </div>
            ) : (
              <div className="chat-content">
                {messages.map((msg, index) => (
                  <div
                    key={index}
                    style={{ color: "black" }}
                    className={`message ${
                      msg.sender === "user" ? "user-msg" : "persona-msg"
                    }`}
                  >
                    {msg.sender === "persona" ? (
                      <div className="persona-bubble">
                        {formatMessage(msg.text)}
                      </div>
                    ) : (
                      <div className="user-bubble">{msg.text}</div>
                    )}
                  </div>
                ))}
              </div>
            )}

            {/* INPUT */}
            <div className="chat-input">
              <input
                placeholder="Type your message…"
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && !sending && handleSend()}
                disabled={sending}
              />
              <button onClick={handleSend} disabled={sending}>
                {sending ? <span className="spinner" /> : "➤"}
              </button>
            </div>
          </div>
        )}



        {/* INSIGHTS PANEL */}
        {(!isMobile || activeTab === "insights") &&
          (!isAnonymous ? (
            <aside className="insights-panel">
              <h4>Insights</h4>
              {isAnonymous ? (
                <>
                  <p className="insights-placeholder">
                    Sign in to unlock deep insights about motivations,
                    objections, and decision drivers.
                  </p>

                  <button
                    className="insights-auth-btn"
                    onClick={() => navigate("/auth/google-login")}
                  >
                    Sign in with Google
                  </button>
                </>
              ) : insights.length === 0 ? (
                <p className="insights-placeholder">
                  Insights will appear here as you converse with {firstName}.
                </p>
              ) : (
                insights.map((insight, index) => (
                  <div key={index} className="insight-item">
                    <strong>{insight.title}</strong>
                    <p>{insight.text}</p>
                  </div>
                ))
              )}
            </aside>
          ) : (
            <aside className="insights-panel disabled">
              <h4>Insights</h4>
              <p className="insights-placeholder">
                Sign in to unlock conversation insights and history.
              </p>
            </aside>
          ))}
      </div>
    </section>
  );
}
