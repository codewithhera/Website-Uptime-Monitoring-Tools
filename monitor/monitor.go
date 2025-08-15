package monitor

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// Website represents a website to monitor
type Website struct {
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
}

// CheckResult represents the result of a website check
type CheckResult struct {
	WebsiteID    string
	Status       string
	ResponseTime int
	Timestamp    time.Time
	Error        error
}

// MonitorEngine manages the monitoring of multiple websites
type MonitorEngine struct {
	websites     map[string]*Website
	mutex        sync.RWMutex
	resultChan   chan CheckResult
	stopChan     chan bool
	httpClient   *http.Client
	userAgents   []string
	running      bool
}

// NewMonitorEngine creates a new monitoring engine
func NewMonitorEngine() *MonitorEngine {
	// Create HTTP client with timeout and TLS config
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// Common user agents to rotate
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:89.0) Gecko/20100101 Firefox/89.0",
		"Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Edge/91.0.864.59",
	}

	return &MonitorEngine{
		websites:   make(map[string]*Website),
		resultChan: make(chan CheckResult, 1000),
		stopChan:   make(chan bool),
		httpClient: client,
		userAgents: userAgents,
		running:    false,
	}
}

// AddWebsite adds a website to monitor
func (me *MonitorEngine) AddWebsite(website *Website) {
	me.mutex.Lock()
	defer me.mutex.Unlock()
	me.websites[website.ID] = website
}

// RemoveWebsite removes a website from monitoring
func (me *MonitorEngine) RemoveWebsite(id string) {
	me.mutex.Lock()
	defer me.mutex.Unlock()
	delete(me.websites, id)
}

// GetWebsite gets a website by ID
func (me *MonitorEngine) GetWebsite(id string) (*Website, bool) {
	me.mutex.RLock()
	defer me.mutex.RUnlock()
	website, exists := me.websites[id]
	return website, exists
}

// GetAllWebsites returns all websites
func (me *MonitorEngine) GetAllWebsites() map[string]*Website {
	me.mutex.RLock()
	defer me.mutex.RUnlock()
	
	// Create a copy to avoid race conditions
	websites := make(map[string]*Website)
	for id, website := range me.websites {
		websites[id] = website
	}
	return websites
}

// UpdateWebsiteStatus updates the status of a website
func (me *MonitorEngine) UpdateWebsiteStatus(id, status string, responseTime int) {
	me.mutex.Lock()
	defer me.mutex.Unlock()
	
	if website, exists := me.websites[id]; exists {
		website.Status = status
		website.LastResponseTime = responseTime
		website.LastCheckTime = time.Now()
	}
}

// Start begins monitoring all websites
func (me *MonitorEngine) Start() {
	me.mutex.Lock()
	if me.running {
		me.mutex.Unlock()
		return
	}
	me.running = true
	me.mutex.Unlock()

	// Start result processor
	go me.processResults()

	// Start monitoring goroutines for each website
	for _, website := range me.GetAllWebsites() {
		if website.Enabled {
			go me.monitorWebsite(website)
		}
	}
}

// Stop stops monitoring all websites
func (me *MonitorEngine) Stop() {
	me.mutex.Lock()
	if !me.running {
		me.mutex.Unlock()
		return
	}
	me.running = false
	me.mutex.Unlock()

	close(me.stopChan)
}

// GetResultChannel returns the result channel for external processing
func (me *MonitorEngine) GetResultChannel() <-chan CheckResult {
	return me.resultChan
}

// monitorWebsite monitors a single website in a goroutine
func (me *MonitorEngine) monitorWebsite(website *Website) {
	ticker := time.NewTicker(time.Duration(website.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	// Perform initial check
	me.checkWebsite(website)

	for {
		select {
		case <-ticker.C:
			// Check if website still exists and is enabled
			if currentWebsite, exists := me.GetWebsite(website.ID); exists && currentWebsite.Enabled {
				me.checkWebsite(currentWebsite)
			} else {
				// Website was removed or disabled, stop monitoring
				return
			}
		case <-me.stopChan:
			return
		}
	}
}

// checkWebsite performs a single check on a website
func (me *MonitorEngine) checkWebsite(website *Website) {
	start := time.Now()
	
	// Create request with random user agent
	req, err := http.NewRequest("GET", website.URL, nil)
	if err != nil {
		me.resultChan <- CheckResult{
			WebsiteID:    website.ID,
			Status:       "down",
			ResponseTime: 0,
			Timestamp:    time.Now(),
			Error:        err,
		}
		return
	}

	// Set random user agent
	userAgent := me.userAgents[rand.Intn(len(me.userAgents))]
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	// Perform request
	resp, err := me.httpClient.Do(req)
	responseTime := int(time.Since(start).Milliseconds())

	var status string
	if err != nil {
		status = "down"
		responseTime = 0
	} else {
		defer resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			status = "up"
		} else {
			status = "down"
		}
	}

	// Send result
	me.resultChan <- CheckResult{
		WebsiteID:    website.ID,
		Status:       status,
		ResponseTime: responseTime,
		Timestamp:    time.Now(),
		Error:        err,
	}
}

// processResults processes check results
func (me *MonitorEngine) processResults() {
	for result := range me.resultChan {
		// Update website status
		me.UpdateWebsiteStatus(result.WebsiteID, result.Status, result.ResponseTime)
		
		// Log result (can be extended to save to JSON files)
		if result.Error != nil {
			fmt.Printf("[%s] Website %s is %s (Error: %v)\n", 
				result.Timestamp.Format("2006-01-02 15:04:05"), 
				result.WebsiteID, result.Status, result.Error)
		} else {
			fmt.Printf("[%s] Website %s is %s (Response time: %dms)\n", 
				result.Timestamp.Format("2006-01-02 15:04:05"), 
				result.WebsiteID, result.Status, result.ResponseTime)
		}
	}
}

