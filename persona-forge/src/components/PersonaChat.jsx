import { useParams } from 'react-router-dom'
import { useNavigate } from 'react-router-dom';
import { useState } from 'react';
import logo from '../assets/images/PersonaForge.svg';

const PERSONAS = {
  amaka: {
    name: 'Amaka Okonkwo',
    role: 'Product Manager at Mid-Size SaaS',
  },
  daniel: {
    name: 'Daniel Chen',
    role: 'Founder & Creator',
  },
  priya: {
    name: 'Priya Sharma',
    role: 'UX Researcher',
  },
    marcus: {
    name: 'Marcus Thompson',
    role: 'Founder seeking structured insights and authentic connection',
  },
};

export default function PersonaChat() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState('chat'); // mobile toggle
  const [message, setMessage] = useState('');
  const [messages, setMessages] = useState([]); // persona responses
  const [insights, setInsights] = useState([]); // initial empty

  const persona = PERSONAS[id] || PERSONAS['amaka']; // fallback to amaka
  const firstName = persona.name.split(' ')[0];

  const handleSend = () => {
    if (!message.trim()) return;

  
    setMessages([...messages, mockResponse]);

    if (insights.length === 0) {
      setInsights([
        {
          title: 'Core Motivation',
          text: `${firstName} values efficiency and team impact over personal gains.`,
        },
        {
          title: 'Key Concern',
          text: 'Integration friction is a deal-breaker. She needs seamless adoption.',
        },
        {
          title: 'Decision Factor',
          text: 'Will consult her team before committing.',
        },
      ]);
    }

    setMessage('');
  };

  return (
    <section className="chat-page">
      {/* HEADER */}
      <header className="chat-header">
        <div className="header-left">
          <img src={logo} alt="PersonaForge" className="logo-img" />
          <button className="back-btn" onClick={() => navigate(-1)}>
            ← Change Persona
          </button>
        </div>

        <div className="header-center">
          <h3>{persona.name}</h3>
          <span>{persona.role}</span>
        </div>
      </header>

      {/* MOBILE TOGGLE */}
      <div className="mobile-toggle">
        <button
          className={activeTab === 'chat' ? 'active' : ''}
          onClick={() => setActiveTab('chat')}
        >
          Chat
        </button>
        <button
          className={activeTab === 'insights' ? 'active' : ''}
          onClick={() => setActiveTab('insights')}
        >
          Insights
        </button>
      </div>

      {/* BODY */}
      <div className="chat-body">
        {/* CHAT */}
        <div className={`chat-panel ${activeTab !== 'chat' ? 'hide-mobile' : ''}`}>
          {messages.length === 0 ? (
            <div className="chat-empty">
              <p>Start a conversation with {firstName}</p>
              <span>
                Ask about their needs, pitch your idea, or explore how they make decisions.
              </span>
            </div>
          ) : (
            <div className="chat-content">
              {messages.map((msg, index) => (
                <div key={index} className="persona-msg">
                  {msg}
                </div>
              ))}
            </div>
          )}

          <div className="chat-input">
            <input 
              placeholder="Type your message…" 
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSend()}
            />
            <button onClick={handleSend}>➤</button>
          </div>
        </div>

        {/* INSIGHTS */}
        <aside className={`insights-panel ${activeTab !== 'insights' ? 'hide-mobile' : ''}`}>
          <h4>Insights</h4>
          {insights.length === 0 ? (
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
      </div>
    </section>
  );
}