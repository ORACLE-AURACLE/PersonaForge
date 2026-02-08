import { useParams, useNavigate } from "react-router-dom";
import { useState, useEffect, useRef } from "react";
import {
  getPersonaById,
  sendMessage,
  getChatHistory,
  getInsights,
  createAnonymousSession,
} from "../apis/api";
import logo from "../assets/images/Main-Logo.svg";
import "../App.css";
import { parseMarkdown } from "../assets/markdown/markdown";

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
  const [insightsLoading, setInsightsLoading] = useState(false);
  const [sending, setSending] = useState(false);

  const [sessionId, setLocalSessionId] = useState(null);
  const [isMobile, setIsMobile] = useState(window.innerWidth <= 768);
  const [headerVisible, setHeaderVisible] = useState(true);

  const chatEndRef = useRef(null);
  const chatContentRef = useRef(null);

  const personaId = parseInt(id, 10);

  // ----------------------------
  // Scroll to bottom on new messages
  // ----------------------------
  useEffect(() => {
    if (chatEndRef.current) {
      chatEndRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [messages]);

  // ----------------------------
  // Handle window resize
  // ----------------------------
  useEffect(() => {
    const onResize = () => setIsMobile(window.innerWidth <= 768);
    window.addEventListener("resize", onResize);
    return () => window.removeEventListener("resize", onResize);
  }, []);

  // ----------------------------
  // Initialize session + fetch persona/chat
  // ----------------------------
  useEffect(() => {
    const init = async () => {
      let sid = contextSessionId || localStorage.getItem("session_id");
      if (!sid) {
        const response = await createAnonymousSession();
        if (response?.session_id) {
          sid = response.session_id;
          localStorage.setItem("session_id", sid);
          if (setSessionId) setSessionId(sid);
        } else {
          setError("Failed to create session");
          setLoading(false);
          return;
        }
      }
      setLocalSessionId(sid);
      await fetchData(sid);
    };

    init();
  }, [contextSessionId, id, setSessionId]);

  // ----------------------------
  // Fetch persona, chat history, and insights
  // ----------------------------
  const fetchData = async (sid) => {
    try {
      const personaResponse = await getPersonaById(id);
      if (personaResponse?.status !== "success") {
        setError(personaResponse?.message || "Failed to fetch persona");
        return;
      }
      setPersona(personaResponse.data);

      // Load insights from localStorage
      const storedInsights = localStorage.getItem(`insights_${sid}_${id}`);
      if (storedInsights) setInsights(JSON.parse(storedInsights));

      // Fetch insights from backend
      setInsightsLoading(true);
      const insightsResponse = await getInsights(sid, id);
      if (insightsResponse?.status === "success" && insightsResponse.data) {
        let parsed = [];
        if (Array.isArray(insightsResponse.data)) parsed = insightsResponse.data;
        else if (typeof insightsResponse.data === "string")
          parsed = parseAnalysis(insightsResponse.data);
        else if (insightsResponse.data.analysis)
          parsed = parseAnalysis(insightsResponse.data.analysis);

        setInsights(parsed);
        localStorage.setItem(`insights_${sid}_${id}`, JSON.stringify(parsed));
      }
      setInsightsLoading(false);

      // Fetch chat history
      const historyResponse = await getChatHistory(sid, id);
      if (historyResponse?.success === false) {
        setError(historyResponse.message);
      } else {
        const transformedMessages = Array.isArray(historyResponse.data?.messages)
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
  // Parse insights into structured format
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
  // Send message & handle loading persona state
  // ----------------------------
  const handleSend = async () => {
    if (!message.trim() || !sessionId) return;

    setSending(true);
    setError(null);

    const userMessage = message;
    setMessage("");

    // Step 1: Add user message + persona loading message
    setMessages((prev) => [
      ...prev,
      { text: userMessage, sender: "user" },
      { text: "Thinking...", sender: "persona", loading: true },
    ]);

    try {
      const response = await sendMessage({
        message: userMessage,
        persona_id: personaId,
        session_id: sessionId,
      });

      if (response?.status !== "success") {
        setError(response?.message || "Failed to send message");
        // Remove loading persona message
        setMessages((prev) => prev.filter((msg) => !msg.loading));
        return;
      }

      const assistantMessage = response.data.message.content;

      // Step 2: Replace loading persona message with actual response
      setMessages((prev) =>
        prev.map((msg) =>
          msg.loading ? { text: assistantMessage, sender: "persona" } : msg
        )
      );

      // Step 3: Fetch and update insights
      setInsightsLoading(true);
      const insightsResponse = await getInsights(sessionId, personaId);

      let parsed = [];
      if (insightsResponse?.status === "success" && insightsResponse.data) {
        if (Array.isArray(insightsResponse.data)) parsed = insightsResponse.data;
        else if (typeof insightsResponse.data === "string")
          parsed = parseAnalysis(insightsResponse.data);
        else if (insightsResponse.data.analysis)
          parsed = parseAnalysis(insightsResponse.data.analysis);
      }

      setInsights((prev) => {
        const existingTitles = new Set(prev.map((i) => i.title));
        const newInsights = parsed.filter((i) => !existingTitles.has(i.title));
        const updated = [...prev, ...newInsights];
        localStorage.setItem(
          `insights_${sessionId}_${personaId}`,
          JSON.stringify(updated)
        );
        return updated;
      });

      setInsightsLoading(false);
    } catch (err) {
      console.error("Error in handleSend:", err);
      setError(err.message);
      setMessages((prev) => prev.filter((msg) => !msg.loading));
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

      <div className={`headCenterHero ${!headerVisible ? "hidden" : ""}`} style={{ color: "black" }}>
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
                    className={`message ${msg.sender === "user" ? "user-msg" : "persona-msg"}`}
                  >
                    {msg.loading ? (
                      <div className="persona-loading">
                        <span className="spinner-small" /> Thinking...
                      </div>
                    ) : (
                      parseMarkdown(msg.text)
                    )}
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
            ) : insightsLoading ? (
              <div className="insights-loading">
                <p>Generating insights...</p>
                <div className="insight-skeleton">
                  <div className="skeleton-title"></div>
                  <div className="skeleton-text"></div>
                  <div className="skeleton-text short"></div>
                </div>
              </div>
            ) : insights.length === 0 ? (
              <p className="insights-placeholder">
                Insights will appear here as you converse with {firstName}.
              </p>
            ) : (
              insights.map((insight, idx) => (
                <div key={idx} className="insight-item fade-in">
                  <h5 style={{ marginBottom: "0.25rem" }}>{insight.title}</h5>
                  {parseMarkdown(insight.text)}
                </div>
              ))
            )}
          </aside>
        )}
      </div>
    </section>
  );
}
