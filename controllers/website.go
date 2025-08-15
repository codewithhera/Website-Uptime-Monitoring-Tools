package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"uptime-monitor/monitor"
	"uptime-monitor/storage"

	"github.com/astaxie/beego"
)

// WebsiteController handles website-related API endpoints
type WebsiteController struct {
	beego.Controller
	MonitorEngine *monitor.MonitorEngine
	Storage       *storage.Storage
}

// WebsiteResponse represents the API response for a website
type WebsiteResponse struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	URL               string    `json:"url"`
	IntervalSeconds   int       `json:"interval_seconds"`
	Status            string    `json:"status"`
	LastCheckTime     time.Time `json:"last_check_time"`
	LastResponseTime  int       `json:"last_response_time_ms"`
	NotificationEmails []string `json:"notification_emails"`
	SlackWebhook      string    `json:"slack_webhook"`
	Enabled           bool      `json:"enabled"`
	Uptime24h         float64   `json:"uptime_24h"`
	Uptime30d         float64   `json:"uptime_30d"`
	AvgResponseTime24h float64  `json:"avg_response_time_24h"`
	History           []storage.HistoryEntry `json:"history,omitempty"`
}

// CreateWebsiteRequest represents the request to create a website
type CreateWebsiteRequest struct {
	Name              string   `json:"name"`
	URL               string   `json:"url"`
	IntervalSeconds   int      `json:"interval_seconds"`
	NotificationEmails []string `json:"notification_emails"`
	SlackWebhook      string   `json:"slack_webhook"`
}

// UpdateWebsiteRequest represents the request to update a website
type UpdateWebsiteRequest struct {
	Name              string   `json:"name"`
	URL               string   `json:"url"`
	IntervalSeconds   int      `json:"interval_seconds"`
	NotificationEmails []string `json:"notification_emails"`
	SlackWebhook      string   `json:"slack_webhook"`
	Enabled           bool     `json:"enabled"`
}

// GetAll returns all websites
func (c *WebsiteController) GetAll() {
	// Enable CORS
	c.Ctx.Output.Header("Access-Control-Allow-Origin", "*")
	c.Ctx.Output.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Ctx.Output.Header("Access-Control-Allow-Headers", "Content-Type")

	websites := c.MonitorEngine.GetAllWebsites()
	var response []WebsiteResponse

	for _, website := range websites {
		// Calculate uptime and average response time
		uptime24h, _ := c.Storage.CalculateUptime(website.ID, 24)
		uptime30d, _ := c.Storage.CalculateUptime(website.ID, 24*30)
		avgResponseTime24h, _ := c.Storage.GetAverageResponseTime(website.ID, 24)

		// Load recent history (last 50 checks)
		history, _ := c.Storage.GetRecentHistory(website.ID, 24) // last 24h

		response = append(response, WebsiteResponse{
			ID:                website.ID,
			Name:              website.Name,
			URL:               website.URL,
			IntervalSeconds:   website.IntervalSeconds,
			Status:            website.Status,
			LastCheckTime:     website.LastCheckTime,
			LastResponseTime:  website.LastResponseTime,
			NotificationEmails: website.NotificationEmails,
			SlackWebhook:      website.SlackWebhook,
			Enabled:           website.Enabled,
			Uptime24h:         uptime24h,
			Uptime30d:         uptime30d,
			AvgResponseTime24h: avgResponseTime24h,
			// Add history to response
			History:           history,
		})
	}

	c.Data["json"] = response
	c.ServeJSON()
}

// Get returns a specific website by ID
func (c *WebsiteController) Get() {
	// Enable CORS
	c.Ctx.Output.Header("Access-Control-Allow-Origin", "*")
	c.Ctx.Output.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Ctx.Output.Header("Access-Control-Allow-Headers", "Content-Type")

	id := c.Ctx.Input.Param(":id")
	website, exists := c.MonitorEngine.GetWebsite(id)
	
	if !exists {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "Website not found"}
		c.ServeJSON()
		return
	}

	// Calculate uptime and average response time
	uptime24h, _ := c.Storage.CalculateUptime(website.ID, 24)
	uptime30d, _ := c.Storage.CalculateUptime(website.ID, 24*30)
	avgResponseTime24h, _ := c.Storage.GetAverageResponseTime(website.ID, 24)

	// Load recent history (last 50 checks)
	history, _ := c.Storage.GetRecentHistory(website.ID, 24)

	response := WebsiteResponse{
		ID:                website.ID,
		Name:              website.Name,
		URL:               website.URL,
		IntervalSeconds:   website.IntervalSeconds,
		Status:            website.Status,
		LastCheckTime:     website.LastCheckTime,
		LastResponseTime:  website.LastResponseTime,
		NotificationEmails: website.NotificationEmails,
		SlackWebhook:      website.SlackWebhook,
		Enabled:           website.Enabled,
		Uptime24h:         uptime24h,
		Uptime30d:         uptime30d,
		AvgResponseTime24h: avgResponseTime24h,
		History:           history,
	}

	c.Data["json"] = response
	c.ServeJSON()
}

// Post creates a new website
func (c *WebsiteController) Post() {
	// Enable CORS
	c.Ctx.Output.Header("Access-Control-Allow-Origin", "*")
	c.Ctx.Output.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Ctx.Output.Header("Access-Control-Allow-Headers", "Content-Type")

	var request CreateWebsiteRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &request); err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "Invalid JSON"}
		c.ServeJSON()
		return
	}

	// Validate request
	if request.Name == "" || request.URL == "" {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "Name and URL are required"}
		c.ServeJSON()
		return
	}

	if request.IntervalSeconds < 30 {
		request.IntervalSeconds = 60 // Default to 60 seconds
	}

	// Generate unique ID
	id := fmt.Sprintf("website_%d", time.Now().UnixNano())

	website := &monitor.Website{
		ID:                id,
		Name:              request.Name,
		URL:               request.URL,
		IntervalSeconds:   request.IntervalSeconds,
		Status:            "unknown",
		LastCheckTime:     time.Time{},
		LastResponseTime:  0,
		NotificationEmails: request.NotificationEmails,
		SlackWebhook:      request.SlackWebhook,
		Enabled:           true,
	}

	// Add to monitor engine
	c.MonitorEngine.AddWebsite(website)

	// Save to storage
	websites := c.MonitorEngine.GetAllWebsites()
	if err := c.Storage.SaveWebsites(websites); err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "Failed to save website"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]string{"id": id, "message": "Website created successfully"}
	c.ServeJSON()
}

// Put updates an existing website
func (c *WebsiteController) Put() {
	// Enable CORS
	c.Ctx.Output.Header("Access-Control-Allow-Origin", "*")
	c.Ctx.Output.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Ctx.Output.Header("Access-Control-Allow-Headers", "Content-Type")

	id := c.Ctx.Input.Param(":id")
	website, exists := c.MonitorEngine.GetWebsite(id)
	
	if !exists {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "Website not found"}
		c.ServeJSON()
		return
	}

	var request UpdateWebsiteRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &request); err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "Invalid JSON"}
		c.ServeJSON()
		return
	}

	// Update website
	if request.Name != "" {
		website.Name = request.Name
	}
	if request.URL != "" {
		website.URL = request.URL
	}
	if request.IntervalSeconds >= 30 {
		website.IntervalSeconds = request.IntervalSeconds
	}
	website.NotificationEmails = request.NotificationEmails
	website.SlackWebhook = request.SlackWebhook
	website.Enabled = request.Enabled

	// Save to storage
	websites := c.MonitorEngine.GetAllWebsites()
	if err := c.Storage.SaveWebsites(websites); err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "Failed to update website"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]string{"message": "Website updated successfully"}
	c.ServeJSON()
}

// Delete removes a website
func (c *WebsiteController) Delete() {
	// Enable CORS
	c.Ctx.Output.Header("Access-Control-Allow-Origin", "*")
	c.Ctx.Output.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Ctx.Output.Header("Access-Control-Allow-Headers", "Content-Type")

	id := c.Ctx.Input.Param(":id")
	_, exists := c.MonitorEngine.GetWebsite(id)
	
	if !exists {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "Website not found"}
		c.ServeJSON()
		return
	}

	// Remove from monitor engine
	c.MonitorEngine.RemoveWebsite(id)

	// Delete history
	c.Storage.DeleteWebsiteHistory(id)

	// Save to storage
	websites := c.MonitorEngine.GetAllWebsites()
	if err := c.Storage.SaveWebsites(websites); err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "Failed to delete website"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]string{"message": "Website deleted successfully"}
	c.ServeJSON()
}

// GetHistory returns history for a website
func (c *WebsiteController) GetHistory() {
	// Enable CORS
	c.Ctx.Output.Header("Access-Control-Allow-Origin", "*")
	c.Ctx.Output.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Ctx.Output.Header("Access-Control-Allow-Headers", "Content-Type")

	id := c.Ctx.Input.Param(":id")
	hoursStr := c.GetString("hours", "24")
	
	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours < 1 {
		hours = 24
	}

	history, err := c.Storage.GetRecentHistory(id, hours)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "Failed to get history"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = history
	c.ServeJSON()
}

// Options handles CORS preflight requests
func (c *WebsiteController) Options() {
	c.Ctx.Output.Header("Access-Control-Allow-Origin", "*")
	c.Ctx.Output.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Ctx.Output.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Ctx.Output.SetStatus(200)
}

