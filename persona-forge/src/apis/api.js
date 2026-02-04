// =======================
// ENV
// =======================
export const GOOGLE_CLIENT_ID = import.meta.env.VITE_GOOGLE_CLIENT_ID;
const BASE_URL = import.meta.env.VITE_API_BASE_URL;

// =======================
// HELPERS
// =======================
const handleResponse = async (response) => {
  if (!response.ok) {
    const errorData = await response
      .json()
      .catch(() => ({ message: "Unknown error" }));

    throw new Error(
      `${response.status}: ${errorData.message || "Backend error"}`
    );
  }
  return response.json();
};

const apiCall = async (endpoint, options = {}) => {
  const sessionId = localStorage.getItem("session_id");

  const headers = {
    "Content-Type": "application/json",
    ...options.headers,
    ...(sessionId && { Authorization: `Bearer ${sessionId}` }),
  };

  try {
    const response = await fetch(`${BASE_URL}${endpoint}`, {
      ...options,
      headers,
    });

    return await handleResponse(response);
  } catch (error) {
    console.error("API Error:", error);
    return {
      success: false,
      message: error.message,
      status: error.message.split(":")[0],
    };
  }
};

// =======================
// AUTH APIs
// =======================
export const createAnonymousSession = async () => {
  const response = await apiCall("/api/auth/anonymous", {
    method: "POST",
  });

  if (response?.session_id) {
    localStorage.setItem("session_id", response.session_id);
  }

  return response;
};

export const authenticateWithGoogle = async (idToken, sessionId) => {
  const res = await fetch(`${BASE_URL}/api/auth/google`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      id_token: idToken,
      session_id: sessionId || null,
    }),
  });

  return res.json();
};

// =======================
// PERSONAS APIs
// =======================
export const getPersonas = async () => {
  return apiCall("/api/personas");
};

export const createPersona = async (name, blueprint) => {
  return apiCall("/api/personas", {
    method: "POST",
    body: JSON.stringify({ name, blueprint }),
  });
};

export const getPersonaById = async (id) => {
  return apiCall(`/api/personas/${id}`);
};

export const deletePersona = async (id) => {
  return apiCall(`/api/personas/${id}`, { method: "DELETE" });
};

// =======================
// CHAT APIs
// =======================
export const sendMessage = async ({ message, persona_id, session_id }) => {
  return apiCall("/api/chat", {
    method: "POST",
    body: JSON.stringify({
      message,
      persona_id,
      session_id,
    }),
  });
};

export const getChatHistory = async (personaId) => {
  return apiCall(`/api/chat/history?persona_id=${personaId}`);
};

// =======================
// INSIGHTS API
// =======================
export const getInsights = async (sessionId) => {
  return apiCall("/api/insight", {
    method: "POST",
    body: JSON.stringify({ session_id: sessionId }),
  });
};
