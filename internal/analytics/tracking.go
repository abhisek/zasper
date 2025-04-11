package analytics

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/posthog/posthog-go"
	"github.com/rs/zerolog/log"
	"github.com/zasper-io/zasper/internal/core"
)

const phEndPoint = "https://us.i.posthog.com"
const phAPIKey = "phc_ptZIEQ1RgjxThkHTSeuxSgx7lxHSvnQx8anw9bD7A6R"

var client posthog.Client
var trackingID string

// UsageStats struct
type UsageStats struct {
	NotebooksOpened   int
	TerminalsOpened   int
	CodeCellsExecuted int
	FilesOpened       int
}

var stats UsageStats

func SetUpUsage() {
	stats = UsageStats{}
}

type EventType string

const (
	EventNotebookOpened   EventType = "notebook_opened"
	EventTerminalOpened   EventType = "terminal_opened"
	EventCodeCellExecuted EventType = "code_cell_executed"
	EventFileOpened       EventType = "file_opened"
)

func IncrementUsageStat(eventType EventType) {
	switch eventType {
	case EventNotebookOpened:
		stats.NotebooksOpened++
	case EventTerminalOpened:
		stats.TerminalsOpened++
	case EventCodeCellExecuted:
		stats.CodeCellsExecuted++
	case EventFileOpened:
		stats.FilesOpened++
	default:
		log.Printf("Unknown event type: %s", eventType)
	}
}

func SendStatsToPostHog() {
	if client == nil {
		log.Info().Msg("PostHog client not initialized.")
		return
	}

	props := posthog.NewProperties().
		Set("notebooks_opened", stats.NotebooksOpened).
		Set("terminals_opened", stats.TerminalsOpened).
		Set("code_cells_executed", stats.CodeCellsExecuted).
		Set("files_opened", stats.FilesOpened).
		Set("timestamp", time.Now().Format(time.RFC3339)).
		Set("OS", core.Zasper.OSName)

	config, _ := core.ReadConfig()

	err := client.Enqueue(posthog.Capture{
		DistinctId: config.TrackingID,
		Event:      "session usage summary",
		Properties: props,
	})

	if err != nil {
		log.Debug().Msgf("PostHog enqueue failed: %v. Retrying once...", err)

		// Retry once after a short delay
		time.Sleep(2 * time.Second)
		err = client.Enqueue(posthog.Capture{
			DistinctId: config.TrackingID,
			Event:      "session usage summary",
			Properties: props,
		})

		if err != nil {
			log.Printf("PostHog enqueue retry also failed: %v", err)
		} else {
			log.Debug().Msg("PostHog retry succeeded.")
			stats = UsageStats{}
		}
	} else {
		log.Debug().Msg("PostHog session usage summary sent successfully.")
		stats = UsageStats{}
	}
}

// Function to generate or retrieve the tracking ID from the config
func GetAnonymousTrackingId() (string, error) {
	config, err := core.ReadConfig()
	if err != nil {
		log.Debug().Msgf("Error reading config file: %v", err)
		return "", err
	}

	if len(config.TrackingID) != 21 {
		// Generate a new UUID and trim it to 21 characters
		config.TrackingID = uuid.New().String()
		config.TrackingID = strings.ReplaceAll(config.TrackingID, "-", "")[:21]

		err = core.WriteConfig(config)
		if err != nil {
			return "", err
		}
	}

	return config.TrackingID, nil
}

func SetUpPostHogClient() error {
	if phAPIKey == "" {
		log.Error().Msg("POSTHOG_API_KEY is not set")
		return nil
	}

	var err error
	client, err = posthog.NewWithConfig(
		phAPIKey,
		posthog.Config{Endpoint: phEndPoint})
	if err != nil {
		log.Error().Msgf("Error setting up PostHog client: %v", err)
		return err
	}
	trackingID, err = GetAnonymousTrackingId()
	if err != nil {
		log.Error().Msgf("Error getting tracking ID: %v", err)
	}
	SetUpUsage()
	return nil
}

func TrackEvent(eventName string, properties map[string]interface{}) {
	if phAPIKey == "" {
		log.Warn().Msg("API key is missing. Event tracking skipped.")
		return
	}

	properties["source"] = "web"
	properties["OS"] = core.Zasper.OSName

	err := client.Enqueue(posthog.Capture{
		DistinctId: trackingID,
		Event:      eventName,
		Properties: properties,
	})

	if err != nil {
		log.Error().Msgf("Failed to send tracking event: %v", err)
		return
	}

	log.Debug().Msg("Sent tracking event successfully.")
}
