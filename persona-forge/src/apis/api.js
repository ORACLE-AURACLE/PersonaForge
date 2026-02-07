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
      `${response.status}: ${errorData.message || "Backend error"}`,
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

  // Normalize the URL to avoid double slashes
  const url = `${BASE_URL.replace(/\/$/, "")}${endpoint}`;

  try {
    const response = await fetch(url, {
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
  const url = `${BASE_URL.replace(/\/$/, "")}/api/auth/google`;
  const res = await fetch(url, {
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
  const sessionId = localStorage.getItem("session_id");
  const url = `${BASE_URL.replace(/\/$/, "")}/api/personas`;
  const response = await fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...(sessionId && { "X-Session-ID": sessionId }),
    },
    body: JSON.stringify({ name, blueprint }),
  });
  return await handleResponse(response);
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

export const getChatHistory = async (sessionId, personaId) => {
  console.log("Calling getChatHistory with sessionId:", sessionId, "personaId:", personaId);
  const result = await apiCall(`/api/chat/${sessionId}/history?persona_id=${personaId}`);
  console.log("getChatHistory result:", result);
  return result;
};


// ----------------------------
// getInsights helper
// ----------------------------
export const getInsights = async (sessionId, personaId) => {
  const params = new URLSearchParams();
  if (sessionId) params.append("session_id", sessionId);
  if (personaId) params.append("persona_id", personaId);

  const url = `/api/insight?${params.toString()}`;
  console.log("GET insights URL:", url);
  return apiCall(url);
};
