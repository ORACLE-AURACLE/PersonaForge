import { useParams, useNavigate } from "react-router-dom";
import { useState, useEffect, useContext } from "react";
import {
  getPersonaById,
  sendMessage,
  getChatHistory,
  getInsights,
} from "../apis/api";
import { SessionContext } from "../App";
import { v4 as uuidv4 } from "uuid"; // to generate guest session IDs
import logo from "../assets/images/Main-Logo.svg";
import "../App.css";
import ReactMarkdown from "react-markdown";

export default function PersonaChat() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { sessionId: contextSessionId, setSessionId } =
    useContext(SessionContext);

  const [persona, setPersona] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const [activeTab, setActiveTab] = useState("chat");
  const [message, setMessage] = useState("");
  const [messages, setMessages] = useState([]);
  const [insights, setInsights] = useState([]);
  const [sending, setSending] = useState(false);

  const personaId = parseInt(id, 10);
  const [isMobile, setIsMobile] = useState(window.innerWidth <= 768);

  // ----------------------------
  // Helper: parse analysis string safely
  // ----------------------------
  const parseAnalysis = (analysis) => {
    if (!analysis || typeof analysis !== "string") return [];
    const sections = analysis.split("\n\n").filter((s) => s.trim());
    return sections.map((section) => {
      const lines = section.split("\n");
      const title = lines[0]
        .replace(/^\d+\.\s*\*\*/, "")
        .replace(/\*\*$/, "")
        .trim();
      const text = lines.slice(1).join("\n").trim();
      return { title, text };
    });
  };

  // ----------------------------------
  // Ensure we have a sessionId
  // ----------------------------------
  const [sessionId, setLocalSessionId] = useState(null);

  useEffect(() => {
    let id = contextSessionId || localStorage.getItem("session_id");
    if (!id) {
      id = uuidv4(); // generate a new guest session
      localStorage.setItem("session_id", id);
      if (setSessionId) setSessionId(id); // update context if available
      console.log("Generated new guest sessionId:", id);
    } else {
      console.log("Using existing sessionId:", id);
    }
    setLocalSessionId(id);
  }, [contextSessionId, setSessionId]);

  // ----------------------------
  // Handle window resize (mobile)
  // ----------------------------
  useEffect(() => {
    const onResize = () => setIsMobile(window.innerWidth <= 768);
    window.addEventListener("resize", onResize);
    return () => window.removeEventListener("resize", onResize);
  }, []);

  // ----------------------------
  // Fetch persona & chat history
  // ----------------------------
  useEffect(() => {
    const fetchData = async () => {
      try {
        const personaResponse = await getPersonaById(id);
        if (personaResponse?.status !== "success") {
          setError(personaResponse?.message || "Failed to fetch persona");
          return;
        }
        setPersona(personaResponse.data);

        if (sessionId) {
          const historyResponse = await getChatHistory(sessionId, id);
          if (historyResponse?.success === false) {
            setError(historyResponse.message);
          } else {
            setMessages(
              Array.isArray(historyResponse.data)
                ? historyResponse.data.map((msg) => ({
                    ...msg,
                    timestamp: msg.timestamp || new Date(),
                  }))
                : [],
            );
          }
        }
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };
    if (sessionId) {
      fetchData();
    }
  }, [id, sessionId]);

  // ----------------------------
  // Send message & fetch insights
  // ----------------------------
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

      console.log("Sending message:", requestBody);
      const response = await sendMessage(requestBody);

      if (response?.status !== "success") {
        setError(response?.message || "Failed to send message");
        return;
      }

      const assistantMessage = response.data.message.content;

      setMessages((prev) => [
        ...prev,
        { text: message, sender: "user", timestamp: new Date() },
        { text: assistantMessage, sender: "persona", timestamp: new Date() },
      ]);
      setMessage("");

      // Fetch insights safely
      if (sessionId) {
        console.log(
          "Fetching insights for sessionId:",
          sessionId,
          "personaId:",
          personaId,
        );
        const insightsResponse = await getInsights(sessionId, personaId);
        console.log("Insights API response:", insightsResponse);

        let parsed = [];
        if (insightsResponse?.status === "success" && insightsResponse.data) {
          if (Array.isArray(insightsResponse.data)) {
            parsed = insightsResponse.data; // already structured
          } else if (typeof insightsResponse.data === "string") {
            parsed = parseAnalysis(insightsResponse.data); // parse string safely
          } else if (
            typeof insightsResponse.data === "object" &&
            insightsResponse.data.analysis
          ) {
            parsed = parseAnalysis(insightsResponse.data.analysis); // parse the analysis string from object
          } else {
            console.warn(
              "Unexpected insights data type:",
              insightsResponse.data,
            );
          }
        } else {
          console.log("No insights returned or API failed");
        }

        setInsights(parsed);
      }
    } catch (err) {
      console.error("Error in handleSend:", err);
      setError(err.message);
    } finally {
      setSending(false);
    }
  };

  if (loading) return <p>Loading persona...</p>;
  if (error) return <p>Error: {error}</p>;
  if (!persona) return <p>Persona not found.</p>;

  const firstName = persona?.name?.split(" ")[0] || "this persona";

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

  const formatTimestamp = (timestamp) => {
    if (!timestamp) return "";
    const date = new Date(timestamp);
    return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  };

  return (
    <section className="chat-page">
      <div className="personaTop">
        <img src={logo} alt="PersonaForge" className="logo" />
        <button className="back-btn" onClick={() => navigate(-1)}>
          ← Change Persona
        </button>
      </div>

      <div className="headCenterHero" style={{ color: "black" }}>
        <div className="headerCenter">
          <h3>{persona.name}</h3>
        </div>
      </div>

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
            {insights.length > 0 && <span className="insight-dot" />}
          </button>
        </div>
      </div>

      <div className="chat-body">
        {(!isMobile || activeTab === "chat") && (
          <div className="chat-panel">
            {messages.length === 0 ? (
              <div className="chat-empty">
                <p>Start a conversation with {firstName}</p>
              </div>
            ) : (
              <div className="chat-content">
                {messages.map((msg, idx) => (
                  <div
                    key={idx}
                    className={`message ${msg.sender === "user" ? "user-msg" : "persona-msg"}`}
                  >
                    {msg.sender === "persona" ? (
                      <div className="persona-bubble">
                        <ReactMarkdown children={msg.text} />
                      </div>
                    ) : (
                      <div className="user-bubble">
                        <ReactMarkdown children={msg.text} />
                      </div>
                    )}
                    <div className="message-timestamp">
                      {formatTimestamp(msg.timestamp)}
                    </div>
                  </div>
                ))}
              </div>
            )}

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

        {(!isMobile || activeTab === "insights") && (
          <aside className={`insights-panel ${!sessionId ? "disabled" : ""}`}>
            <h4>Insights</h4>
            {!sessionId ? (
              <p className="insights-placeholder">
                No session available to fetch insights.
              </p>
            ) : insights.length === 0 ? (
              <p className="insights-placeholder">
                Insights will appear here as you converse with {firstName}.
              </p>
            ) : (
              insights.map((insight, idx) => (
                <div key={idx} className="insight-item">
                  <h5 style={{ marginBottom: "0.25rem" }}>{insight.title}</h5>
                  <ReactMarkdown children={insight.text} />
                </div>
              ))
            )}
          </aside>
        )}
      </div>
    </section>
  );
}
