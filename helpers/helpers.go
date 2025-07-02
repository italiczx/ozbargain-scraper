package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"log"

	"github.com/PuerkitoBio/goquery"
)

type Deal struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Timestamp string `json:"timestamp"`
	Votes     int    `json:"votes"`
	Category  string `json:"category"`
}

// Discord embed structures
type DiscordEmbed struct {
	Title       string              `json:"title"`
	URL         string              `json:"url"`
	Description string              `json:"description,omitempty"`
	Color       int                 `json:"color"`
	Fields      []DiscordEmbedField `json:"fields"`
	Footer      *DiscordEmbedFooter `json:"footer,omitempty"`
}

type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type DiscordEmbedFooter struct {
	Text string `json:"text"`
}

type DiscordWebhookPayload struct {
	Content string `json:"content"`
	Embeds []DiscordEmbed `json:"embeds"`
}

const (
	UrlPrefix = "https://www.ozbargain.com.au"
)

func ParseDeals(resp *http.Response, category string) ([]Deal, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	deals := []Deal{}

	doc.Find("ul.ozblist li").Each(func(i int, s *goquery.Selection) {
		deal := Deal{}

		deal.Title = strings.TrimSpace(s.Find("div.title a").Text())
		deal.URL = UrlPrefix + s.Find("div.title a").AttrOr("href", "")
		deal.Timestamp = strings.TrimSpace(s.Find("ul.meta li.timestamp").Text())
		deal.Category = category
		// Parse votes - extract number from text like "+341"
		votesText := strings.TrimSpace(s.Find("ul.meta li.votes").Text())
		if votesText != "" {
			// Remove the "+" sign and convert to int
			votesText = strings.TrimPrefix(votesText, "+")
			if votes, err := strconv.Atoi(votesText); err == nil {
				deal.Votes = votes
			}
		}

		// Only add deals that have a title and URL
		if deal.Title != "" && deal.URL != "" {
			deals = append(deals, deal)
		}
	})

	return deals, nil
}

func FormatAndPostToDiscord(deals []Deal, asTable bool) {
	discordWebhookURL := os.Getenv("DISCORD_WEBHOOK_URL")

	if len(deals) == 0 {
		log.Printf("No deals to post")
		return
	}

	var payload DiscordWebhookPayload

	if asTable {
		// Add clickable links after the table
		var linksBuilder strings.Builder
		linksBuilder.WriteString("\n**🔗 Links:**\n")
		for i, deal := range deals {
			linksBuilder.WriteString(fmt.Sprintf("%d. [%s](%s) *(%d votes)*\n", i+1, deal.Title, deal.URL, deal.Votes))
		}

		// Combine table and links into description (which has higher character limit)
		var description strings.Builder
		description.WriteString(linksBuilder.String())

		embed := DiscordEmbed{
			Title:       fmt.Sprintf("🔥 %s Deals (%d found)", deals[0].Category, len(deals)),
			Description: description.String(),
			Color:       0x00A86B, // OzBargain green color
			Footer: &DiscordEmbedFooter{
				Text: "OzBargain Deal Alert",
			},
		}

		payload = DiscordWebhookPayload{
			Embeds:  []DiscordEmbed{embed},
		}
	} else {
		// Original format - separate embeds for each deal
		embeds := []DiscordEmbed{}

		for _, deal := range deals {
			// Create color based on vote count (green for high votes, orange for medium, red for low)
			color := 0xFF0000 // Red for low votes
			if deal.Votes >= 200 {
				color = 0x00FF00 // Green for high votes
			} else if deal.Votes >= 100 {
				color = 0xFFA500 // Orange for medium votes
			}

			embed := DiscordEmbed{
				Title: deal.Title,
				URL:   deal.URL,
				Color: color,
				Fields: []DiscordEmbedField{
					{
						Name:   "👍 Votes",
						Value:  fmt.Sprintf("%d", deal.Votes),
						Inline: true,
					},
					{
						Name:   "🕒 Posted",
						Value:  deal.Timestamp,
						Inline: true,
					},
				},
				Footer: &DiscordEmbedFooter{
					Text: "OzBargain Deal Alert - " + deal.Category,
				},
			}

			embeds = append(embeds, embed)
		}

		payload = DiscordWebhookPayload{
			Embeds:  embeds,
		}
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return
	}

	response, err := http.Post(discordWebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error posting to Discord: %v", err)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return
	}

	if response.StatusCode != http.StatusNoContent {
		log.Printf("Discord webhook returned status %d: %s", response.StatusCode, string(body))
	} else {
		log.Printf("Successfully posted %d deals to Discord", len(deals))
	}
}