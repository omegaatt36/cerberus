package gemini

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Service is a wrapper around the Gemini client
type Service struct {
	model  string
	client *genai.Client
}

// NewService creates a new Gemini service
func NewService(ctx context.Context, apiKey, model string) (*Service, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %v", err)
	}

	fullModelName := func(name string) string {
		if strings.ContainsRune(name, '/') {
			return name
		}
		return "models/" + name
	}(model)

	var exist bool
	modelIterator := client.ListModels(ctx)
	for {
		m, err := modelIterator.Next()
		if err != nil {
			break
		}

		if m.Name == fullModelName {
			exist = true
			fmt.Println("found")
			break
		}
	}

	if !exist {
		return nil, fmt.Errorf("model %s not found", model)
	}

	return &Service{
		client: client,
		model:  model,
	}, nil
}

// Close closes the Gemini client
func (g *Service) Close() error {
	return g.client.Close()
}

// GetEmotionScore analyzes the emotion of a given input string or emoji
func (g *Service) GetEmotionScore(ctx context.Context, input string) (int, error) {
	const formatGetEmotionScorePrompt = `Analyze the emotion in the following text or emoji and provide a score from 0 to 100, where 0 is very negative and 100 is very positive. Only respond with the number, no other text. Text to analyze: %s`

	resp, err := g.client.GenerativeModel(g.model).GenerateContent(ctx, genai.Text(fmt.Sprintf(formatGetEmotionScorePrompt, input)))
	if err != nil {
		return 0, fmt.Errorf("error generating content: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return 0, fmt.Errorf("no response received from Gemini")
	}

	scoreStr := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			scoreStr += string(textPart)
		}
	}

	score, err := strconv.Atoi(strings.TrimSpace(scoreStr))
	if err != nil {
		return 0, fmt.Errorf("failed to parse score: %v", err)
	}

	if score < 0 || score > 100 {
		return 0, fmt.Errorf("invalid score received: %d", score)
	}

	return score, nil
}

// GenerateTaskSuggestion generates a task suggestion based on the emotion score and description
func (g *Service) GenerateTaskSuggestion(ctx context.Context, emoji string, description string, score int) (string, error) {
	const formatGenerateTaskSuggestionPrompt = `你是一個超級厲害的情緒分析大師，同時也是一個網路迷因和梗圖專家。你的任務是解讀用戶的 emoji '%s' 與他可能的的心情描述 '%s'，以及 emotion score '%d' (0-100, where 0 is very negative and 100 is very positive)。

你的回應應該既搞笑又有用，讓用戶忍俊不禁的同時也能獲得實際的幫助。

請根據用戶選擇的 emoji，生成一個「使用正體中文」、按照以下的四個階段回應，：

用一句帶有流行梗的話來描述這個 emoji 可能代表的心情。
附上一個與當前情緒相關，表達對用戶情緒的理解。
提供 1-2 個能夠改善或維持心情的建議，但要用誇張幽默的方式表達。
用一個流行的網路用語來鼓勵用戶，為回應畫上完美的句點。

範例輸入：
	emoji: ':sweet_smile:', description: 'Feeling embarrassed about the situation', emotion score: 50

好的範例輸出：
	看來你正在經歷一場尷尬力量大爆發啊，尷尬到連汗都變成了表情符號！
	就像那個黑人問號的迷因一樣，我現在腦子裡全是問號。究竟發生了什麼讓你如此尷尬呢？
	不如我們來玩個尷尬大逃亡如何？第一步，深呼吸。第二步，假裝你是在演一部超級英雄電影，而尷尬是你必須戰勝的終極大魔王！
	記住，尷尬讓你更強大！你現在就是尷尬界的一代宗師，指定是修煉滿一百年的那種。加油，尷尬大師！

不好的範例輸出：
	階段 1：流行梗描述
	看來你正處於佛系狀態，萬事看淡，無慾無求，天下任我行！

	階段 2：情緒理解
	就像那個無所謂臉的迷因，你現在就是超級佛系，對一切事情都佛系到不行。

	階段 3：誇張建議
	不如我們來展開一場佛系修行之旅吧！首先，我們要學會對一切事物都說「沒關係」。其次，我們要培養「佛擋殺佛」的氣勢，遇事鎮定自若，泰山崩於前而面不改色！

	階段 4：流行網路用語
	佛系少年，加油！承包你一年的好佛氣，佛力無邊！

在創作回應時，請注意以下幾點：
	- 必須符合範例輸出的格式，不包含 listed notation 或 step 等字眼。
	- 請不要將輸入的 emoji 和 description 直接生成在回應上。
	- 請不要出現「附圖」等等你無法顯示的內容。
	- 語氣要親切幽默，就像在跟瀏覽量最高的臉書梗圖粉專對話一樣。
	- 盡量使用當前流行的網路用語和迷因，但要確保它們是廣為人知的。
	- 建議雖然要幽默，但還是要有實際可行性，不能太離譜。
	- 對於負面情緒，用幽默來緩解，但不要嘲笑用戶的感受。
	- 對於正面情緒，用誇張的方式讚美，讓用戶笑得更開心。
	- 可以適當使用一些無厘頭的幽默，但要確保不會冒犯到用戶。`
	resp, err := g.client.GenerativeModel(g.model).GenerateContent(ctx,
		genai.Text(fmt.Sprintf(formatGenerateTaskSuggestionPrompt, emoji, description, score)))
	if err != nil {
		return "", fmt.Errorf("failed to generate task suggestion: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response received for task suggestion")
	}

	var suggestion string
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			suggestion += string(textPart)
		}
	}

	return strings.TrimSpace(suggestion), nil
}

// GenerateDailySummary generates a summary using Gemini based on the average score
func (g *Service) GenerateDailySummary(ctx context.Context, average float64) (string, error) {
	const formatGenerateDailySummaryPrompt = `Based on the average emotion score of %.2f (0-100, where 0 is very negative and 100 is very positive), provide a brief summary in Traditional Chinese about the overall mood and a general suggestion for improvement. Keep it concise and positive.`

	resp, err := g.client.GenerativeModel(g.model).GenerateContent(ctx, genai.Text(fmt.Sprintf(formatGenerateDailySummaryPrompt, average)))
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response received for summary")
	}

	var summary string
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			summary += string(textPart)
		}
	}

	return strings.TrimSpace(summary), nil
}
