import { useParams, useNavigate } from "react-router-dom";
import { useState, useEffect, useRef } from "react";
import {
  getPersonaById,
  sendMessage,
  getChatHistory,
  getInsights,
  createAnonymousSession,
} from "../apis/api";
import { v4 as uuidv4 } from "uuid"; // to generate guest session IDs
import logo from "../assets/images/Main-Logo.svg";
import "../App.css";
import ReactMarkdown from "react-markdown";

export default function PersonaChat({ contextSessionId, setSessionId }) {
  const { id } = useParams();
  const navigate = useNavigate();

  const [persona, setPersona] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const [activeTab, setActiveTab] = useState("chat");
  const [message, setMessage] = useState("");
  const [messages, setMessages] = useState([]);
  const [insights, setInsights] = useState([]);
  const [sending, setSending] = useState(false);

  const [sessionId, setLocalSessionId] = useState(null);
  const [isMobile, setIsMobile] = useState(window.innerWidth <= 768);
  const [headerVisible, setHeaderVisible] = useState(true);

  const chatEndRef = useRef(null); // for auto-scroll
  const chatContentRef = useRef(null); // for scroll detection

  const personaId = parseInt(id, 10);

  // ----------------------------
  // Scroll to bottom whenever messages update
  // ----------------------------
  useEffect(() => {
    if (chatEndRef.current) {
      chatEndRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [messages]);

  const fetchData = async (sid) => {
    try {
      console.log("Session ID:", sid, "Persona ID:", id);

      const personaResponse = await getPersonaById(id);
      console.log("Persona API response:", personaResponse);
      if (personaResponse?.status !== "success") {
        setError(personaResponse?.message || "Failed to fetch persona");
        return;
      }
      setPersona(personaResponse.data);

      // Use getChatHistory instead of direct fetch for consistency
      const historyResponse = await getChatHistory(sid, id);
      console.log("Chat history API response:", historyResponse);
      if (historyResponse?.success === false) {
        setError(historyResponse.message);
      } else {
        // Transform API data to component format
        const transformedMessages = Array.isArray(
          historyResponse.data?.messages,
        )
          ? historyResponse.data.messages.map((msg) => ({
              text: msg.content,
              sender: msg.role === "user" ? "user" : "persona",
            }))
          : [];
        setMessages(transformedMessages);
      }
    } catch (err) {
      console.error("Error fetching data:", err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  // ----------------------------
  // Handle window resize (mobile)
  // ----------------------------
  useEffect(() => {
    const onResize = () => setIsMobile(window.innerWidth <= 768);
    window.addEventListener("resize", onResize);
    return () => window.removeEventListener("resize", onResize);
  }, []);

  // ----------------------------
  // Initialize sessionId + fetch persona & chat history
  // ----------------------------
  useEffect(() => {
    const init = async () => {
      let sid = contextSessionId || localStorage.getItem("session_id");
      if (!sid) {
        console.log("No sessionId found, creating anonymous session...");
        const response = await createAnonymousSession();
        console.log("createAnonymousSession response:", response);
        if (response?.session_id) {
          sid = response.session_id;
          localStorage.setItem("session_id", sid);
          if (setSessionId) setSessionId(sid);
          console.log("Created anonymous sessionId:", sid);
        } else {
          console.error(
            "Failed to create anonymous session, response:",
            response,
          );
          setError("Failed to create session");
          setLoading(false);
          return;
        }
      } else {
        console.log("Using existing sessionId:", sid);
      }
      setLocalSessionId(sid);

      await fetchData(sid);
    };

    init();
  }, [contextSessionId, id, setSessionId]);

  // ----------------------------
  // Parse insights safely
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
        { text: message, sender: "user" },
        { text: assistantMessage, sender: "persona" },
      ]);
      setMessage("");

      // Fetch insights safely
      const insightsResponse = await getInsights(sessionId, personaId);
      console.log("Insights API response:", insightsResponse);

      let parsed = [];
      if (insightsResponse?.status === "success" && insightsResponse.data) {
        if (Array.isArray(insightsResponse.data))
          parsed = insightsResponse.data;
        else if (typeof insightsResponse.data === "string")
          parsed = parseAnalysis(insightsResponse.data);
        else if (insightsResponse.data.analysis)
          parsed = parseAnalysis(insightsResponse.data.analysis);
      }
      setInsights(parsed);
    } catch (err) {
      console.error("Error in handleSend:", err);
      setError(err.message);
    } finally {
      setSending(false);
    }
  };

  if (loading)
    return (
      <div className="loading-spinner">
        <div className="spinner"></div>
      </div>
    );
  if (error) return <p>Error: {error}</p>;
  if (!persona) return <p>Persona not found.</p>;

  const firstName = persona?.name?.split(" ")[0] || "this persona";

  return (
    <section className="chat-page">
      <div className="personTopDiv">
        <div className={`personaTop ${!headerVisible ? "hidden" : ""}`}>
          <img src={logo} alt="PersonaForge" className="logo" />
          <button className="back-btn" onClick={() => navigate(-1)}>
            ← Change Persona
          </button>
        </div>
      </div>

      <div
        className={`headCenterHero ${!headerVisible ? "hidden" : ""}`}
        style={{ color: "black" }}
      >
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
              <div className="chat-content" ref={chatContentRef}>
                {messages.map((msg, idx) => (
                  <div
                    key={idx}
                    className={`message ${
                      msg.sender === "user" ? "user-msg" : "persona-msg"
                    }`}
                  >
                    <ReactMarkdown children={msg.text} />
                  </div>
                ))}
                <div ref={chatEndRef} />
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
