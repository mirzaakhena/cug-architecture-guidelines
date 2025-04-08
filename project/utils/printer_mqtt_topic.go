package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// MQTTTopicData represents metadata about an MQTT topic
type MQTTTopicData struct {
	Topic        string
	Description  string
	Tag          string
	HandlerType  string // Type of handler (e.g., "heartbeat", "pong", "discovery")
	Subscription bool   // Whether this is a subscription topic
}

// MQTTTopicPrinter collects and displays MQTT topic information
type MQTTTopicPrinter struct {
	topics []MQTTTopicData
}

// NewMQTTTopicPrinter creates a new MQTT topic printer
func NewMQTTTopicPrinter() *MQTTTopicPrinter {
	return &MQTTTopicPrinter{
		topics: []MQTTTopicData{},
	}
}

// Add adds a new MQTT topic to the printer
func (p *MQTTTopicPrinter) Add(topicData MQTTTopicData) *MQTTTopicPrinter {
	p.topics = append(p.topics, topicData)
	return p
}

// PrintMQTTTopicTable prints a formatted table of MQTT topics
func (p MQTTTopicPrinter) PrintMQTTTopicTable() MQTTTopicPrinter {
	// Define column widths
	tagWidth := 20
	topicWidth := 40
	descWidth := 40

	// Print table header
	headerFormat := fmt.Sprintf("%%-%ds %%-%ds %%s\n", tagWidth, topicWidth)
	fmt.Printf(headerFormat, "Tag", "Topic", "Description")
	fmt.Println(strings.Repeat("-", tagWidth+topicWidth+descWidth+4))

	// Print each row
	rowFormat := fmt.Sprintf("%%-%ds %%-%ds %%s\n", tagWidth, topicWidth)
	for _, topic := range p.topics {
		tag := TruncateOrPad(topic.Tag, tagWidth)
		topicPath := TruncateOrPad(topic.Topic, topicWidth)
		desc := TruncateOrPad(topic.Description, descWidth)

		fmt.Printf(rowFormat, tag, topicPath, desc)
	}
	fmt.Println()

	return p
}

// GenerateMQTTTopicSchema generates a JSON schema for MQTT topics
func (p MQTTTopicPrinter) GenerateMQTTTopicSchema() string {
	schema := map[string]any{
		"mqtt_topics": []map[string]any{},
	}

	for _, topic := range p.topics {
		topicSchema := map[string]any{
			"topic":        topic.Topic,
			"description":  topic.Description,
			"tag":          topic.Tag,
			"subscription": topic.Subscription,
		}
		schema["mqtt_topics"] = append(schema["mqtt_topics"].([]map[string]any), topicSchema)
	}

	jsonSchema, _ := json.MarshalIndent(schema, "", "  ")
	return string(jsonSchema)
}

// PublishMQTTTopicSchema publishes the MQTT topic schema to a specified URL
func (p MQTTTopicPrinter) PublishMQTTTopicSchema(mux *http.ServeMux, baseURL, apiURL string) MQTTTopicPrinter {
	handler := func(w http.ResponseWriter, req *http.Request) {

		w.Header().Set("Content-Type", "text/json")
		w.Write([]byte(p.GenerateMQTTTopicSchema()))
	}

	mux.HandleFunc("GET "+apiURL, handler)

	fmt.Printf("\nMQTT TOPIC SCHEMA available at: %s%s\n", baseURL, apiURL)
	return p
}
