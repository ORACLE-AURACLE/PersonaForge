import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { getPersonas, deletePersona, getPersonasBySession } from "../apis/api";
import logo from "../assets/images/Main-Logo.svg";
import "../App.css";

export default function Personas() {
  const navigate = useNavigate();
  const [personas, setPersonas] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchPersonas = async () => {
      try {
        // Fetch all default personas
        const allResponse = await getPersonas();
        let allPersonas = [];
        if (allResponse.success === false) {
          setError(allResponse.message);
          return;
        } else if (Array.isArray(allResponse)) {
          allPersonas = allResponse;
        } else if (allResponse.data && Array.isArray(allResponse.data)) {
          allPersonas = allResponse.data;
        } else {
          setError("Unexpected response format for all personas");
          return;
        }

        // Ensure personality is always an array for all
        allPersonas = allPersonas.map((p) => {
          if (
            p.blueprint?.personality &&
            typeof p.blueprint.personality === "string"
          ) {
            p.blueprint.personality = p.blueprint.personality
              .split(",")
              .map((tag) => tag.trim());
          }
          return p;
        });

        // Fetch personas by session
        const sessionId = localStorage.getItem("session_id");
        let sessionPersonas = [];
        if (sessionId) {
          const sessionResponse = await getPersonasBySession(sessionId);
          if (sessionResponse.success === false) {
            // Don't set error for session, just skip
          } else if (Array.isArray(sessionResponse)) {
            sessionPersonas = sessionResponse;
          } else if (sessionResponse.data && Array.isArray(sessionResponse.data)) {
            sessionPersonas = sessionResponse.data;
          }
        }

        // Merge results, assuming no duplicates
        const mergedPersonas = [...allPersonas, ...sessionPersonas];

        setPersonas(mergedPersonas);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchPersonas();
  }, []);

  const handleDelete = async (id) => {
    if (!window.confirm("Are you sure you want to delete this persona?"))
      return;
    try {
      const response = await deletePersona(id);
      if (response.success === false) {
        setError(response.message);
      } else {
        setPersonas(personas.filter((p) => p.id !== id));
      }
    } catch (err) {
      setError(err.message);
    }
  };

  return (
    <section className="persona-page">
      <div className="persona-top">
        <img src={logo} alt="PersonaForge" className="logo" />
      </div>

      <button className="back-btn" onClick={() => navigate(-1)}>
        ← Back
      </button>

      <div className="personaContent">
        <div className="persona-header">
          <h1>Choose a Default Persona</h1>
          <p>
            Select someone to talk to. Each persona has distinct motivations,
            constraints, and decision-making patterns.
          </p>
          <h2>OR</h2>
          {/* NEW: Create Custom Persona Button */}
          <button
            className="create-persona-btn"
            onClick={() => navigate("/personas/create")}
          >
            + Create Custom Persona
          </button>
        </div>

        {loading && <p style={{ color: "black" }}>Loading personas...</p>}
        {error && (
          <div className="error-message">
            <p>Error: {error}</p>
          </div>
        )}

        <div className="persona-grid">
          {personas.map((persona) => (
            <div
              key={persona.id}
              className="personaCard"
              onClick={() => navigate(`/personas/${persona.id}`)}
            >
              {/* DELETE (custom personas only) */}
              {persona.custom && (
                <button
                  className="persona-delete"
                  onClick={(e) => {
                    e.stopPropagation(); // ⛔ prevent card navigation
                    handleDelete(persona.id);
                  }}
                  aria-label="Delete persona"
                >
                  ✕
                </button>
              )}

              <h3>{persona.name}</h3>

             

              <p>
                {persona.description ||
                  persona.blueprint?.description ||
                  "No description available."}
              </p>

              {/* Personality Tags */}
              <div className="persona-tags">
                {Array.isArray(persona.blueprint?.personality) &&
                persona.blueprint.personality.length > 0 ? (
                  persona.blueprint.personality.map((tag) => (
                    <span key={tag} className="persona-tag">
                      {tag}
                    </span>
                  ))
                ) : (
                  <span className="no-tags">No personality data</span>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
