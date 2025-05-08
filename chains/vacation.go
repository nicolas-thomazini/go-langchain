package chains

import (
	"context"
	"errors"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

var Vacations []*Vacation

func GetVacationFromDb(id uuid.UUID) (Vacation, error) {
	idx := slices.IndexFunc(Vacations, func(v *Vacation) bool { return v.Id == id })
	if idx < 0 {
		return Vacation{}, errors.New("ID not found")
	}
	return *Vacations[idx], nil
}

func GenerateVacationIdeaChange(id uuid.UUID, budget int, season string, hobbies []string) {
	log.Printf("Generating new vacation with ID: %s", id)

	v := &Vacation{Id: id, Completed: false, Idea: ""}
	Vacations = append(Vacations, v)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithBaseURL("https://api.together.ai/v1"),
		openai.WithModel("meta-llama/Llama-3.3-70B-Instruct-Turbo-Free"),
		openai.WithToken(os.Getenv("TOGETHER_API_KEY")),
	)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	system_message_prompt_string := "You are an AI travel agent that will help me create a vacation idea.\n" +
		"My favorite season is {{.season}}.\n" +
		"My hobbies include {{.hobbies}}.\n" +
		"My budget is {{.budget}} dollars.\n"
	system_message_prompt := prompts.NewSystemMessagePromptTemplate(system_message_prompt_string, []string{"season", "hobbies", "budget"})

	human_message_prompt_string := "Write a travel itinerary for me"
	human_message_prompt := prompts.NewHumanMessagePromptTemplate(human_message_prompt_string, []string{})

	chat_prompt := prompts.NewChatPromptTemplate([]prompts.MessageFormatter{system_message_prompt, human_message_prompt})

	vals := map[string]any{
		"season":  season,
		"budget":  budget,
		"hobbies": strings.Join(hobbies, ", "),
	}

	msgs, err := chat_prompt.FormatMessages(vals)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	content := []llms.MessageContent{
		llms.TextParts(msgs[0].GetType(), msgs[0].GetContent()),
		llms.TextParts(msgs[1].GetType(), msgs[1].GetContent()),
	}

	completion, err := llm.GenerateContent(ctx, content)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	v.Idea = completion.Choices[0].Content
	v.Completed = true
	log.Printf("Generation for %s is done! Vacation idea: %s", v.Id, v.Idea)
}
