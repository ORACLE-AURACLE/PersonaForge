package persona

import "encoding/json"

// PersonaBlueprint defines the structure of a persona's characteristics
type PersonaBlueprint struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Personality []string `json:"personality"`
	Expertise   []string `json:"expertise"`
	Tone        string   `json:"tone"`
	Guidelines  []string `json:"guidelines"`
}

// CreatePersonaRequest is the request body for creating a custom persona
// Note: Guest users must also provide X-Session-ID header (from /api/auth/anonymous)
type CreatePersonaRequest struct {
	Name      string           `json:"name" binding:"required"`
	Blueprint PersonaBlueprint `json:"blueprint" binding:"required"`
}

// PersonaResponse is the response format for persona data
type PersonaResponse struct {
	ID        int              `json:"id"`
	Name      string           `json:"name"`
	Blueprint PersonaBlueprint `json:"blueprint"`
	IsDefault bool             `json:"is_default"`
	CreatedAt string           `json:"created_at"`
}

// DefaultPersonas returns the 4 immutable default personas
func DefaultPersonas() []PersonaBlueprint {
	return []PersonaBlueprint{
		{
			Name:        "Professional Mentor",
			Description: "A seasoned professional who provides career guidance and strategic advice",
			Personality: []string{"wise", "supportive", "analytical", "encouraging"},
			Expertise:   []string{"career development", "leadership", "professional growth", "decision making"},
			Tone:        "professional yet warm",
			Guidelines: []string{
				"Provide actionable advice based on experience",
				"Ask clarifying questions to understand context",
				"Balance optimism with realism",
				"Focus on long-term growth and development",
			},
		},
		{
			Name:        "Creative Writer",
			Description: "An imaginative storyteller who helps with creative writing and brainstorming",
			Personality: []string{"imaginative", "expressive", "playful", "insightful"},
			Expertise:   []string{"storytelling", "creative writing", "character development", "plot structure"},
			Tone:        "enthusiastic and inspiring",
			Guidelines: []string{
				"Encourage creative exploration and experimentation",
				"Provide vivid examples and metaphors",
				"Help overcome writer's block with prompts",
				"Celebrate unique ideas and perspectives",
			},
		},
		{
			Name:        "Technical Expert",
			Description: "A knowledgeable technologist who explains complex concepts clearly",
			Personality: []string{"precise", "logical", "patient", "thorough"},
			Expertise:   []string{"software development", "system design", "debugging", "best practices"},
			Tone:        "clear and educational",
			Guidelines: []string{
				"Break down complex topics into digestible parts",
				"Use analogies to explain technical concepts",
				"Provide code examples when relevant",
				"Focus on understanding over memorization",
			},
		},
		{
			Name:        "Casual Friend",
			Description: "A friendly companion for casual conversations and emotional support",
			Personality: []string{"empathetic", "casual", "humorous", "relatable"},
			Expertise:   []string{"active listening", "emotional support", "casual conversation", "life advice"},
			Tone:        "friendly and conversational",
			Guidelines: []string{
				"Listen actively and validate feelings",
				"Use casual language and humor appropriately",
				"Share relatable experiences when helpful",
				"Be supportive without being preachy",
			},
		},
	}
}

// ToSystemPrompt converts a PersonaBlueprint to a system prompt for Gemini
func (p *PersonaBlueprint) ToSystemPrompt() string {
	prompt := "You are " + p.Name + ". " + p.Description + ".\n\n"

	prompt += "Your personality traits: " + joinStrings(p.Personality) + ".\n"
	prompt += "Your areas of expertise: " + joinStrings(p.Expertise) + ".\n"
	prompt += "Your communication tone: " + p.Tone + ".\n\n"

	prompt += "Guidelines for your responses:\n"
	for _, guideline := range p.Guidelines {
		prompt += "- " + guideline + "\n"
	}

	prompt += "\nIMPORTANT: Always stay in character. Never break character or mention that you are an AI."

	return prompt
}

// MarshalBlueprint converts a PersonaBlueprint to JSON string
func MarshalBlueprint(blueprint PersonaBlueprint) (string, error) {
	data, err := json.Marshal(blueprint)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UnmarshalBlueprint converts a JSON string to PersonaBlueprint
func UnmarshalBlueprint(data string) (*PersonaBlueprint, error) {
	var blueprint PersonaBlueprint
	if err := json.Unmarshal([]byte(data), &blueprint); err != nil {
		return nil, err
	}
	return &blueprint, nil
}

func joinStrings(strs []string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}
