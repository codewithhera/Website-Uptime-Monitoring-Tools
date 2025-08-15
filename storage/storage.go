package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
	"uptime-monitor/monitor"
)

// HistoryEntry represents a single monitoring history entry
type HistoryEntry struct {
	Timestamp    time.Time `json:"timestamp"`
	Status       string    `json:"status"`
	ResponseTime int       `json:"response_time_ms"`
}

// Storage manages JSON file storage for websites and history
type Storage struct {
	dataDir     string
	websitesFile string
	mutex       sync.RWMutex
}

// NewStorage creates a new storage instance
func NewStorage(dataDir string) *Storage {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create data directory: %v", err))
	}

	return &Storage{
		dataDir:     dataDir,
		websitesFile: filepath.Join(dataDir, "websites.json"),
	}
}

// SaveWebsites saves all websites to JSON file
func (s *Storage) SaveWebsites(websites map[string]*monitor.Website) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Convert map to slice for JSON serialization
	websiteList := make([]*monitor.Website, 0, len(websites))
	for _, website := range websites {
		websiteList = append(websiteList, website)
	}

	data, err := json.MarshalIndent(websiteList, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal websites: %v", err)
	}

	// Write to temporary file first, then rename for atomic operation
	tempFile := s.websitesFile + ".tmp"
	if err := ioutil.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write websites file: %v", err)
	}

	if err := os.Rename(tempFile, s.websitesFile); err != nil {
		return fmt.Errorf("failed to rename websites file: %v", err)
	}

	return nil
}

// LoadWebsites loads all websites from JSON file
func (s *Storage) LoadWebsites() (map[string]*monitor.Website, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Check if file exists
	if _, err := os.Stat(s.websitesFile); os.IsNotExist(err) {
		// Return empty map if file doesn't exist
		return make(map[string]*monitor.Website), nil
	}

	data, err := ioutil.ReadFile(s.websitesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read websites file: %v", err)
	}

	var websiteList []*monitor.Website
	if err := json.Unmarshal(data, &websiteList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal websites: %v", err)
	}

	// Convert slice to map
	websites := make(map[string]*monitor.Website)
	for _, website := range websiteList {
		websites[website.ID] = website
	}

	return websites, nil
}

// SaveHistory saves a history entry for a website
func (s *Storage) SaveHistory(websiteID string, entry HistoryEntry) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	historyFile := filepath.Join(s.dataDir, fmt.Sprintf("history_%s.json", websiteID))

	// Load existing history
	var history []HistoryEntry
	if data, err := ioutil.ReadFile(historyFile); err == nil {
		json.Unmarshal(data, &history)
	}

	// Add new entry
	history = append(history, entry)

	// Keep only last 1000 entries to prevent unlimited growth
	if len(history) > 1000 {
		history = history[len(history)-1000:]
	}

	// Save updated history
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %v", err)
	}

	// Write to temporary file first, then rename for atomic operation
	tempFile := historyFile + ".tmp"
	if err := ioutil.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write history file: %v", err)
	}

	if err := os.Rename(tempFile, historyFile); err != nil {
		return fmt.Errorf("failed to rename history file: %v", err)
	}

	return nil
}

// LoadHistory loads history for a website
func (s *Storage) LoadHistory(websiteID string) ([]HistoryEntry, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	historyFile := filepath.Join(s.dataDir, fmt.Sprintf("history_%s.json", websiteID))

	// Check if file exists
	if _, err := os.Stat(historyFile); os.IsNotExist(err) {
		// Return empty slice if file doesn't exist
		return []HistoryEntry{}, nil
	}

	data, err := ioutil.ReadFile(historyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read history file: %v", err)
	}

	var history []HistoryEntry
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("failed to unmarshal history: %v", err)
	}

	return history, nil
}

// GetRecentHistory gets recent history entries for a website
func (s *Storage) GetRecentHistory(websiteID string, hours int) ([]HistoryEntry, error) {
	history, err := s.LoadHistory(websiteID)
	if err != nil {
		return nil, err
	}

	// Filter entries from the last N hours
	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)
	var recentHistory []HistoryEntry

	for _, entry := range history {
		if entry.Timestamp.After(cutoff) {
			recentHistory = append(recentHistory, entry)
		}
	}

	return recentHistory, nil
}

// DeleteWebsiteHistory deletes all history for a website
func (s *Storage) DeleteWebsiteHistory(websiteID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	historyFile := filepath.Join(s.dataDir, fmt.Sprintf("history_%s.json", websiteID))
	
	if err := os.Remove(historyFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete history file: %v", err)
	}

	return nil
}

// CalculateUptime calculates uptime percentage for a website over a given period
func (s *Storage) CalculateUptime(websiteID string, hours int) (float64, error) {
	history, err := s.GetRecentHistory(websiteID, hours)
	if err != nil {
		return 0, err
	}

	if len(history) == 0 {
		return 100.0, nil // Assume 100% if no data
	}

	upCount := 0
	for _, entry := range history {
		if entry.Status == "up" {
			upCount++
		}
	}

	return float64(upCount) / float64(len(history)) * 100.0, nil
}

// GetAverageResponseTime calculates average response time for a website over a given period
func (s *Storage) GetAverageResponseTime(websiteID string, hours int) (float64, error) {
	history, err := s.GetRecentHistory(websiteID, hours)
	if err != nil {
		return 0, err
	}

	if len(history) == 0 {
		return 0, nil
	}

	totalTime := 0
	validEntries := 0

	for _, entry := range history {
		if entry.Status == "up" && entry.ResponseTime > 0 {
			totalTime += entry.ResponseTime
			validEntries++
		}
	}

	if validEntries == 0 {
		return 0, nil
	}

	return float64(totalTime) / float64(validEntries), nil
}

// CleanupOldHistory removes history files for websites that no longer exist
func (s *Storage) CleanupOldHistory(existingWebsiteIDs map[string]bool) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	files, err := ioutil.ReadDir(s.dataDir)
	if err != nil {
		return fmt.Errorf("failed to read data directory: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			// Check if it's a history file
			if len(file.Name()) > 8 && file.Name()[:8] == "history_" {
				// Extract website ID from filename
				websiteID := file.Name()[8 : len(file.Name())-5] // Remove "history_" prefix and ".json" suffix
				
				// If website doesn't exist anymore, delete the history file
				if !existingWebsiteIDs[websiteID] {
					historyFile := filepath.Join(s.dataDir, file.Name())
					if err := os.Remove(historyFile); err != nil {
						fmt.Printf("Warning: failed to delete old history file %s: %v\n", historyFile, err)
					}
				}
			}
		}
	}

	return nil
}

