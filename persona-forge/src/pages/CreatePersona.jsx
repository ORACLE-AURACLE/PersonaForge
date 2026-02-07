import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { createPersona, createAnonymousSession } from "../apis/api";
import logo from "../assets/images/Main-Logo.svg"; // 👈 adjust path if needed

export default function CreatePersona() {
  const navigate = useNavigate();

  const [name, setName] = useState("");
  const [form, setForm] = useState({
    description: "",
    tone: "",
    personality: "",
    expertise: "",
    guidelines: "",
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const handleChange = (field, value) => {
    setForm((prev) => ({ ...prev, [field]: value }));
  };

  const toArray = (text) =>
    text
      .split("\n")
      .map((v) => v.trim())
      .filter(Boolean);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      // Ensure anonymous session exists for guest users
      if (!localStorage.getItem("session_id")) {
        await createAnonymousSession();
      }

      const payload = {
        name,
        blueprint: {
          description: form.description,
          tone: form.tone,
          personality: toArray(form.personality),
          expertise: toArray(form.expertise),
          guidelines: toArray(form.guidelines),
        },
      };

      const response = await createPersona(name, payload.blueprint);
      const newPersonaId = response?.data?.id;
      if (!newPersonaId) {
        throw new Error("Failed to retrieve new persona ID");
      }
      navigate(`/personas/${newPersonaId}`);
    } catch (err) {
      setError(err.message || "Failed to create persona");
    } finally {
      setLoading(false);
    }
  };

  return (
    <section className="">
      <div className="persona-brand">
        <img src={logo} alt="PersonaForge logo" className="persona-logo" />
      </div>

      <div className="create-persona-page">
        <header className="persona-hero">
          <h1>Create a Custom Persona</h1>
          <p>
            Design a persona that reflects the audience you want to understand.
            Be specific about their context, constraints, and motivations.
          </p>
        </header>

        <div className="create-persona-container">
          {/* 🔒 FORM — UNCHANGED */}
          <form onSubmit={handleSubmit} className="create-persona-form">
            <input
              className="input"
              placeholder="Persona name (e.g. Lena)"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />

            <textarea
              className="textarea"
              placeholder="Description"
              value={form.description}
              onChange={(e) => handleChange("description", e.target.value)}
              required
            />

            <input
              className="input"
              placeholder="Tone"
              value={form.tone}
              onChange={(e) => handleChange("tone", e.target.value)}
              required
            />

            <textarea
              className="textarea"
              placeholder="Personality (one per line)"
              value={form.personality}
              onChange={(e) => handleChange("personality", e.target.value)}
              required
            />

            <textarea
              className="textarea"
              placeholder="Expertise (one per line)"
              value={form.expertise}
              onChange={(e) => handleChange("expertise", e.target.value)}
              required
            />

            <textarea
              className="textarea"
              placeholder="Guidelines (one per line)"
              value={form.guidelines}
              onChange={(e) => handleChange("guidelines", e.target.value)}
              required
            />

            <button className="primary-btn" disabled={loading}>
              {loading ? "Creating..." : "Create Persona"}
            </button>
          </form>

          {error && <p className="error-message">{error}</p>}
        </div>
      </div>
      {/* 🔶 HEADER / HERO (NEW) */}
    </section>
  );
}
