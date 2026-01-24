import { useNavigate } from "react-router-dom";
import logo from "../assets/images/PersonaForge.svg";
import "../App.css";

export default function Personas() {
  const navigate = useNavigate();

  return (
    <section className="persona-page">
      {/* Top left stack */}
      <div className="persona-top">
        <img src={logo} alt="PersonaForge" className="logo-img" />
        <button className="back-btn" onClick={() => navigate(-1)}>
          ← Back
        </button>
      </div>

      {/* Header */}
      <div className="persona-content">
        <div className="persona-header">
          <h1>Choose a Persona</h1>
          <p>
            Select someone to talk to. Each persona has distinct motivations,
            constraints, and decision-making patterns.
          </p>
        </div>

        {/* Cards */}
        <div className="persona-grid">
          <div
            className="persona-card"
            onClick={() => navigate("/personas/amaka")}
          >
            <h3>Amaka Okonkwo</h3>
            <span className="persona-meta">
              34 · Product Manager at B2B SaaS
            </span>
            <p>
              Makes data-driven decisions centered on user needs. Values clarity
              over complexity.
            </p>
            <div className="persona-tags">
              <span>Pragmatic</span>
              <span>ROI-driven</span>
              <span>Team-oriented</span>
            </div>
          </div>

          <div
            className="persona-card"
            onClick={() => navigate("/personas/daniel")}
          >
            <h3>Daniel Chen</h3>
            <span className="persona-meta">28 · Indie Creator & Coder</span>
            <p>
              Runs a solo studio and ships fast. Focuses on speed, quality, and
              scalability.
            </p>
            <div className="persona-tags">
              <span>Creative</span>
              <span>Cost-conscious</span>
              <span>Hands-on</span>
            </div>
          </div>

          <div
            className="persona-card"
            onClick={() => navigate("/personas/priya")}
          >
            <h3>Priya Sharma</h3>
            <span className="persona-meta">41 · UX Research Lead</span>
            <p>
              Leads research teams and translates insights into product
              strategy.
            </p>
            <div className="persona-tags">
              <span>Analytical</span>
              <span>User-first</span>
              <span>Detail-oriented</span>
            </div>
          </div>

          <div
            className="persona-card"
            onClick={() => navigate("/personas/marcus")}
          >
            <h3>Marcus Thompson</h3>
            <span className="persona-meta">38 · Founder, Wellness App</span>
            <p>
              Builds purpose-driven products and values authenticity and trust.
            </p>
            <div className="persona-tags">
              <span>Visionary</span>
              <span>Growth-focused</span>
              <span>Community-led</span>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
