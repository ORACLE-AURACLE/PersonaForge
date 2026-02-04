const BASE_URL = "https://persona-forge-ffce.onrender.com";

// Helper function to handle API responses and errors
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

// Helper function to make API calls
const apiCall = async (endpoint, options = {}) => {
  const sessionId = localStorage.getItem("session_id");
  const headers = {
    "Content-Type": "application/json",
    ...options.headers,
  };
  if (sessionId) {
    headers["Authorization"] = "Bearer " + sessionId; // Assuming sessionId is the auth token
  }
  try {
    const response = await fetch(`${BASE_URL}${endpoint}`, {
      headers,
      ...options,
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

// Authentication APIs
export const createAnonymousSession = async () => {
  try {
    const response = await apiCall("/api/auth/anonymous", { method: "POST" });
    
    // Assuming the response contains a field called 'sessionId'
    if (response?.sessionId) {
      localStorage.setItem("sessionId", response.sessionId);
    } else {
      console.warn("No session ID returned from API");
    }

    return response;
  } catch (error) {
    console.error("Error creating anonymous session:", error);
    throw error;
  }
};


export const authenticateWithGoogle = async (idToken, sessionId = null) => {
  const body = { id_token: idToken };
  if (sessionId) body.session_id = sessionId;
  return apiCall("/api/auth/google", {
    method: "POST",
    body: JSON.stringify(body),
  });
};

// Personas APIs
export const getPersonas = async () => {
  return apiCall("/api/personas");
};

export const createPersona = async (name, blueprint) => {
  const sessionId = localStorage.getItem("session_id");

  return apiCall("/api/personas", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...(sessionId && { "X-Session-ID": sessionId })
    },
    body: JSON.stringify({
      name,
      blueprint
    })
  });
};



export const getPersonaById = async (id) => {
  return apiCall(`/api/personas/${id}`);
};

export const deletePersona = async (id) => {
  return apiCall(`/api/personas/${id}`, { method: "DELETE" });
};

// Chat APIs
export const sendMessage = async ({ message, persona_id, session_id }) => {
  return apiCall("/api/chat", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
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

// Insights API
export const getInsights = async (sessionId) => {
  return apiCall("/api/insight", {
    method: "POST",
    body: JSON.stringify({ session_id: sessionId }),
  });
};
