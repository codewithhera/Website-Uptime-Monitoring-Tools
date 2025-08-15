package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"sync"
	"time"
)

// NotificationConfig holds configuration for notifications
type NotificationConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
}

// SlackMessage represents a Slack webhook message
type SlackMessage struct {
	Text        string       `json:"text"`
	Username    string       `json:"username,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Attachment represents a Slack message attachment
type Attachment struct {
	Color     string  `json:"color"`
	Title     string  `json:"title"`
	Text      string  `json:"text"`
	Timestamp int64   `json:"ts"`
	Fields    []Field `json:"fields,omitempty"`
}

// Field represents a field in a Slack attachment
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// StatusChangeEvent represents a website status change
type StatusChangeEvent struct {
	WebsiteID    string
	WebsiteName  string
	WebsiteURL   string
	OldStatus    string
	NewStatus    string
	ResponseTime int
	Timestamp    time.Time
	Emails       []string
	SlackWebhook string
}

// NotificationManager manages sending notifications
type NotificationManager struct {
	config       NotificationConfig
	eventQueue   chan StatusChangeEvent
	stopChan     chan bool
	running      bool
	mutex        sync.RWMutex
	httpClient   *http.Client
	lastNotified map[string]time.Time // Track last notification time per website to prevent spam
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(config NotificationConfig) *NotificationManager {
	return &NotificationManager{
		config:       config,
		eventQueue:   make(chan StatusChangeEvent, 100),
		stopChan:     make(chan bool),
		running:      false,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		lastNotified: make(map[string]time.Time),
	}
}

// Start begins processing notification events
func (nm *NotificationManager) Start() {
	nm.mutex.Lock()
	if nm.running {
		nm.mutex.Unlock()
		return
	}
	nm.running = true
	nm.mutex.Unlock()

	go nm.processEvents()
}

// Stop stops processing notification events
func (nm *NotificationManager) Stop() {
	nm.mutex.Lock()
	if !nm.running {
		nm.mutex.Unlock()
		return
	}
	nm.running = false
	nm.mutex.Unlock()

	close(nm.stopChan)
}

// SendStatusChange queues a status change notification
func (nm *NotificationManager) SendStatusChange(event StatusChangeEvent) {
	// Check if we should throttle notifications for this website
	nm.mutex.RLock()
	lastNotified, exists := nm.lastNotified[event.WebsiteID]
	nm.mutex.RUnlock()

	// Don't send notifications more than once every 5 minutes for the same website
	if exists && time.Since(lastNotified) < 5*time.Minute {
		return
	}

	select {
	case nm.eventQueue <- event:
		nm.mutex.Lock()
		nm.lastNotified[event.WebsiteID] = time.Now()
		nm.mutex.Unlock()
	default:
		fmt.Printf("Warning: notification queue is full, dropping event for %s\n", event.WebsiteID)
	}
}

// processEvents processes notification events from the queue
func (nm *NotificationManager) processEvents() {
	for {
		select {
		case event := <-nm.eventQueue:
			nm.handleStatusChange(event)
		case <-nm.stopChan:
			return
		}
	}
}

// handleStatusChange handles a single status change event
func (nm *NotificationManager) handleStatusChange(event StatusChangeEvent) {
	// Send email notifications
	if len(event.Emails) > 0 && nm.config.SMTPHost != "" {
		go nm.sendEmailNotification(event)
	}

	// Send Slack notification
	if event.SlackWebhook != "" {
		go nm.sendSlackNotification(event)
	}
}

// sendEmailNotification sends an email notification
func (nm *NotificationManager) sendEmailNotification(event StatusChangeEvent) {
	if nm.config.SMTPHost == "" || nm.config.SMTPUsername == "" {
		fmt.Printf("Warning: SMTP not configured, skipping email notification for %s\n", event.WebsiteID)
		return
	}

	subject := fmt.Sprintf("Website %s is %s", event.WebsiteName, strings.ToUpper(event.NewStatus))
	
	var body string
	if event.NewStatus == "up" {
		body = fmt.Sprintf(`Website %s (%s) is now UP!

Status changed from %s to %s at %s
Response time: %dms

This is an automated notification from your uptime monitoring system.`,
			event.WebsiteName,
			event.WebsiteURL,
			strings.ToUpper(event.OldStatus),
			strings.ToUpper(event.NewStatus),
			event.Timestamp.Format("2006-01-02 15:04:05"),
			event.ResponseTime)
	} else {
		body = fmt.Sprintf(`Website %s (%s) is DOWN!

Status changed from %s to %s at %s

Please check your website immediately.

This is an automated notification from your uptime monitoring system.`,
			event.WebsiteName,
			event.WebsiteURL,
			strings.ToUpper(event.OldStatus),
			strings.ToUpper(event.NewStatus),
			event.Timestamp.Format("2006-01-02 15:04:05"))
	}

	// Create email message
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		nm.config.FromEmail,
		strings.Join(event.Emails, ","),
		subject,
		body)

	// Send email
	auth := smtp.PlainAuth("", nm.config.SMTPUsername, nm.config.SMTPPassword, nm.config.SMTPHost)
	addr := fmt.Sprintf("%s:%s", nm.config.SMTPHost, nm.config.SMTPPort)
	
	err := smtp.SendMail(addr, auth, nm.config.FromEmail, event.Emails, []byte(message))
	if err != nil {
		fmt.Printf("Error sending email notification for %s: %v\n", event.WebsiteID, err)
	} else {
		fmt.Printf("Email notification sent for %s to %v\n", event.WebsiteID, event.Emails)
	}
}

// sendSlackNotification sends a Slack webhook notification
func (nm *NotificationManager) sendSlackNotification(event StatusChangeEvent) {
	var color string
	var emoji string
	var title string

	if event.NewStatus == "up" {
		color = "good"
		emoji = ":white_check_mark:"
		title = fmt.Sprintf("%s Website %s is UP", emoji, event.WebsiteName)
	} else {
		color = "danger"
		emoji = ":x:"
		title = fmt.Sprintf("%s Website %s is DOWN", emoji, event.WebsiteName)
	}

	fields := []Field{
		{Title: "Website", Value: event.WebsiteName, Short: true},
		{Title: "URL", Value: event.WebsiteURL, Short: true},
		{Title: "Status Change", Value: fmt.Sprintf("%s → %s", strings.ToUpper(event.OldStatus), strings.ToUpper(event.NewStatus)), Short: true},
		{Title: "Time", Value: event.Timestamp.Format("2006-01-02 15:04:05"), Short: true},
	}

	if event.NewStatus == "up" && event.ResponseTime > 0 {
		fields = append(fields, Field{Title: "Response Time", Value: fmt.Sprintf("%dms", event.ResponseTime), Short: true})
	}

	attachment := Attachment{
		Color:     color,
		Title:     title,
		Timestamp: event.Timestamp.Unix(),
		Fields:    fields,
	}

	message := SlackMessage{
		Username:    "Uptime Monitor",
		IconEmoji:   ":computer:",
		Attachments: []Attachment{attachment},
	}

	// Send to Slack
	jsonData, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("Error marshaling Slack message for %s: %v\n", event.WebsiteID, err)
		return
	}

	resp, err := nm.httpClient.Post(event.SlackWebhook, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending Slack notification for %s: %v\n", event.WebsiteID, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Slack webhook returned status %d for %s\n", resp.StatusCode, event.WebsiteID)
	} else {
		fmt.Printf("Slack notification sent for %s\n", event.WebsiteID)
	}
}

// UpdateConfig updates the notification configuration
func (nm *NotificationManager) UpdateConfig(config NotificationConfig) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	nm.config = config
}

