package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"uptime-monitor/controllers"
	"uptime-monitor/monitor"
	"uptime-monitor/notification"
	_ "uptime-monitor/routers" // This import ensures the init() function in routers/router.go runs
	"uptime-monitor/storage"

	"github.com/astaxie/beego"
)

func main() {
	// Initialize storage
	dataDir := "./data"
	stor := storage.NewStorage(dataDir)

	// Initialize notification manager
	notificationConfig := notification.NotificationConfig{
		SMTPHost:     beego.AppConfig.String("smtp_host"),
		SMTPPort:     beego.AppConfig.String("smtp_port"),
		SMTPUsername: beego.AppConfig.String("smtp_username"),
		SMTPPassword: beego.AppConfig.String("smtp_password"),
		FromEmail:    beego.AppConfig.String("from_email"),
	}
	notificationManager := notification.NewNotificationManager(notificationConfig)

	// Initialize monitor engine
	monitorEngine := monitor.NewMonitorEngine()

	// Load existing websites from storage
	websites, err := stor.LoadWebsites()
	if err != nil {
		log.Printf("Warning: Failed to load websites from storage: %v", err)
	} else {
		for _, website := range websites {
			monitorEngine.AddWebsite(website)
		}
		log.Printf("Loaded %d websites from storage", len(websites))
	}

	// --- START: Add this section to add sample websites programmatically ---
	// Only add samples if no websites are loaded yet (e.g., on first run)
	if len(websites) == 0 {
		log.Println("Adding sample websites...")
		sampleURLs := []struct { Name string; URL string } {
			{Name: "Google", URL: "https://www.google.com"},
			{Name: "YouTube", URL: "https://www.youtube.com"},
			{Name: "Facebook", URL: "https://www.facebook.com"},
			{Name: "Baidu", URL: "https://www.baidu.com"},
			{Name: "Wikipedia", URL: "https://www.wikipedia.org"},
			{Name: "Reddit", URL: "https://www.reddit.com"},
			{Name: "Yahoo", URL: "https://www.yahoo.com"},
			{Name: "Amazon", URL: "https://www.amazon.com"},
			{Name: "Twitter", URL: "https://www.twitter.com"},
			{Name: "Instagram", URL: "https://www.instagram.com"},
			{Name: "LinkedIn", URL: "https://www.linkedin.com"},
			{Name: "Netflix", URL: "https://www.netflix.com"},
			{Name: "Microsoft", URL: "https://www.microsoft.com"},
			{Name: "Apple", URL: "https://www.apple.com"},
			{Name: "Twitch", URL: "https://www.twitch.tv"},
			{Name: "eBay", URL: "https://www.ebay.com"},
			{Name: "Pinterest", URL: "https://www.pinterest.com"},
			{Name: "Bing", URL: "https://www.bing.com"},
			{Name: "Stack Overflow", URL: "https://www.stackoverflow.com"},
			{Name: "GitHub", URL: "https://www.github.com"},
			{Name: "Medium", URL: "https://www.medium.com"},
			{Name: "WordPress", URL: "https://www.wordpress.com"},
			{Name: "Blogger", URL: "https://www.blogger.com"},
			{Name: "Tumblr", URL: "https://www.tumblr.com"},
			{Name: "Quora", URL: "https://www.quora.com"},
			{Name: "GitLab", URL: "https://www.gitlab.com"},
			{Name: "Bitbucket", URL: "https://www.bitbucket.org"},
			{Name: "Atlassian", URL: "https://www.atlassian.com"},
			{Name: "Docker", URL: "https://www.docker.com"},
			{Name: "Kubernetes", URL: "https://www.kubernetes.io"},
			{Name: "AWS", URL: "https://www.aws.amazon.com"},
			{Name: "Azure", URL: "https://azure.microsoft.com"},
			{Name: "Google Cloud", URL: "https://cloud.google.com"},
			{Name: "DigitalOcean", URL: "https://www.digitalocean.com"},
			{Name: "Heroku", URL: "https://www.heroku.com"},
			{Name: "Netlify", URL: "https://www.netlify.com"},
			{Name: "Vercel", URL: "https://www.vercel.com"},
			{Name: "Cloudflare", URL: "https://www.cloudflare.com"},
			{Name: "Stripe", URL: "https://www.stripe.com"},
			{Name: "PayPal", URL: "https://www.paypal.com"},
			{Name: "Shopify", URL: "https://www.shopify.com"},
			{Name: "Etsy", URL: "https://www.etsy.com"},
			{Name: "Walmart", URL: "https://www.walmart.com"},
			{Name: "Target", URL: "https://www.target.com"},
			{Name: "Best Buy", URL: "https://www.bestbuy.com"},
			{Name: "CNN", URL: "https://www.cnn.com"},
			{Name: "BBC", URL: "https://www.bbc.com"},
			{Name: "NY Times", URL: "https://www.nytimes.com"},
			{Name: "The Guardian", URL: "https://www.theguardian.com"},
			{Name: "Reuters", URL: "https://www.reuters.com"},
			{Name: "Bloomberg", URL: "https://www.bloomberg.com"},
			{Name: "Forbes", URL: "https://www.forbes.com"},
			{Name: "TechCrunch", URL: "https://www.techcrunch.com"},
			{Name: "Wired", URL: "https://www.wired.com"},
			{Name: "The Verge", URL: "https://www.theverge.com"},
			{Name: "Engadget", URL: "https://www.engadget.com"},
			{Name: "Ars Technica", URL: "https://www.arstechnica.com"},
			{Name: "Mozilla", URL: "https://www.mozilla.org"},
			{Name: "Opera", URL: "https://www.opera.com"},
			{Name: "Brave", URL: "https://www.brave.com"},
			{Name: "DuckDuckGo", URL: "https://www.duckduckgo.com"},
			{Name: "ProtonMail", URL: "https://www.protonmail.com"},
			{Name: "Signal", URL: "https://www.signal.org"},
			{Name: "Telegram", URL: "https://www.telegram.org"},
			{Name: "WhatsApp", URL: "https://www.whatsapp.com"},
			{Name: "Zoom", URL: "https://www.zoom.us"},
			{Name: "Slack", URL: "https://www.slack.com"},
			{Name: "Trello", URL: "https://www.trello.com"},
			{Name: "Asana", URL: "https://www.asana.com"},
			{Name: "Notion", URL: "https://www.notion.so"},
			{Name: "Figma", URL: "https://www.figma.com"},
			{Name: "Adobe", URL: "https://www.adobe.com"},
			{Name: "Autodesk", URL: "https://www.autodesk.com"},
			{Name: "Blender", URL: "https://www.blender.org"},
			{Name: "GIMP", URL: "https://www.gimp.org"},
			{Name: "Inkscape", URL: "https://www.inkscape.org"},
			{Name: "LibreOffice", URL: "https://www.libreoffice.org"},
			{Name: "OpenOffice", URL: "https://www.openoffice.org"},
			{Name: "Ubuntu", URL: "https://www.ubuntu.com"},
			{Name: "Debian", URL: "https://www.debian.org"},
			{Name: "Fedora", URL: "https://www.fedora.org"},
			{Name: "CentOS", URL: "https://www.centos.org"},
			{Name: "Red Hat", URL: "https://www.redhat.com"},
			{Name: "SUSE", URL: "https://www.suse.com"},
			{Name: "Kernel.org", URL: "https://www.kernel.org"},
			{Name: "GNU", URL: "https://www.gnu.org"},
			{Name: "FSF", URL: "https://www.fsf.org"},
			{Name: "Apache", URL: "https://www.apache.org"},
			{Name: "Nginx", URL: "https://www.nginx.com"},
			{Name: "MySQL", URL: "https://www.mysql.com"},
			{Name: "PostgreSQL", URL: "https://www.postgresql.org"},
			{Name: "MongoDB", URL: "https://www.mongodb.com"},
			{Name: "Redis", URL: "https://www.redis.io"},
			{Name: "Elastic", URL: "https://www.elastic.co"},
			{Name: "Grafana", URL: "https://www.grafana.com"},
			{Name: "Prometheus", URL: "https://www.prometheus.io"},
			{Name: "Jenkins", URL: "https://www.jenkins.io"},
			{Name: "Travis CI", URL: "https://www.travis-ci.com"},
			{Name: "CircleCI", URL: "https://www.circleci.com"},
			{Name: "GitHub Pages", URL: "https://www.github.io"},
			{Name: "Google.org", URL: "https://www.google.org"},
			{Name: "UN", URL: "https://www.un.org"},
			{Name: "WHO", URL: "https://www.who.int"},
			{Name: "NASA", URL: "https://www.nasa.gov"},
			{Name: "SpaceX", URL: "https://www.spacex.com"},
			{Name: "Tesla", URL: "https://www.tesla.com"},
			{Name: "OpenAI", URL: "https://www.openai.com"},
			{Name: "DeepMind", URL: "https://www.deepmind.com"},
			{Name: "Hugging Face", URL: "https://www.huggingface.co"},
			{Name: "Kaggle", URL: "https://www.kaggle.com"},
			{Name: "Coursera", URL: "https://www.coursera.org"},
			{Name: "edX", URL: "https://www.edx.org"},
			{Name: "Udemy", URL: "https://www.udemy.com"},
			{Name: "Khan Academy", URL: "https://www.khanacademy.org"},
			{Name: "W3Schools", URL: "https://www.w3schools.com"},
			{Name: "MDN Web Docs", URL: "https://www.developer.mozilla.org"},
			{Name: "PHP", URL: "https://www.php.net"},
			{Name: "Python", URL: "https://www.python.org"},
			{Name: "Ruby", URL: "https://www.ruby-lang.org"},
			{Name: "Java", URL: "https://www.java.com"},
			{Name: "Go", URL: "https://www.golang.org"},
			{Name: "Rust", URL: "https://www.rust-lang.org"},
			{Name: "TypeScript", URL: "https://www.typescriptlang.org"},
			{Name: "Node.js", URL: "https://www.nodejs.org"},
			{Name: "React", URL: "https://www.react.dev"},
			{Name: "Angular", URL: "https://www.angular.io"},
			{Name: "Vue.js", URL: "https://www.vuejs.org"},
			{Name: "jQuery", URL: "https://www.jquery.com"},
			{Name: "Bootstrap", URL: "https://www.bootstrapcdn.com"},
			{Name: "Tailwind CSS", URL: "https://www.tailwindcss.com"},
			{Name: "Material UI", URL: "https://www.material-ui.com"},
			{Name: "Ant Design", URL: "https://www.ant.design"},
			{Name: "D3.js", URL: "https://www.d3js.org"},
			{Name: "Chart.js", URL: "https://www.chartjs.org"},
			{Name: "Highcharts", URL: "https://www.highcharts.com"},
			{Name: "Mapbox", URL: "https://www.mapbox.com"},
			{Name: "OpenStreetMap", URL: "https://www.openstreetmap.org"},
			{Name: "Google Maps", URL: "https://www.google.maps"},
			{Name: "Weather.com", URL: "https://www.weather.com"},
			{Name: "AccuWeather", URL: "https://www.accuweather.com"},
			{Name: "Time and Date", URL: "https://www.timeanddate.com"},
			{Name: "World Bank", URL: "https://www.worldbank.org"},
			{Name: "IMF", URL: "https://www.imf.org"},
			{Name: "WTO", URL: "https://www.wto.org"},
			{Name: "Wikipedia Main", URL: "https://www.wikipedia.org/wiki/Main_Page"},
			{Name: "Example.com", URL: "https://www.example.com"},
			{Name: "Example.org", URL: "https://www.example.org"},
			{Name: "Example.net", URL: "https://www.example.net"},
		}

		for i, site := range sampleURLs {
			id := fmt.Sprintf("website_%d", time.Now( ).UnixNano() + int64(i)) // Unique ID
			website := &monitor.Website{
				ID:                id,
				Name:              site.Name,
				URL:               site.URL,
				IntervalSeconds:   60, // Default check interval
				Status:            "unknown",
				LastCheckTime:     time.Time{},
				LastResponseTime:  0,
				NotificationEmails: []string{}, // No email notifications by default
				SlackWebhook:      "",         // No Slack notifications by default
				Enabled:           true,
			}
			monitorEngine.AddWebsite(website)
		}
		// Save the newly added sample websites to storage
		if err := stor.SaveWebsites(monitorEngine.GetAllWebsites()); err != nil {
			log.Printf("Error saving sample websites: %v", err)
		}
		log.Printf("Added %d sample websites.", len(sampleURLs))
	}
	// --- END: Add this section ---

	// Set up controllers with dependencies
	// IMPORTANT: Create the controller instance *after* monitorEngine and stor are initialized
	websiteController := &controllers.WebsiteController{
		MonitorEngine: monitorEngine,
		Storage:       stor,
	}

	// Register controller instance with Beego after initialization
	// These routes MUST be registered here in main.go, not in init() of router.go
	beego.Router("/api/websites", websiteController, "get:GetAll;post:Post;options:Options")
	beego.Router("/api/websites/:id", websiteController, "get:Get;put:Put;delete:Delete;options:Options")
	beego.Router("/api/websites/:id/history", websiteController, "get:GetHistory;options:Options")

	// Start notification manager
	notificationManager.Start()

	// Start monitor engine
	monitorEngine.Start()

	// Handle monitoring results and notifications
	go func() {
		previousStatus := make(map[string]string)
		
		for result := range monitorEngine.GetResultChannel() {
			// Save history
			historyEntry := storage.HistoryEntry{
				Timestamp:    result.Timestamp,
				Status:       result.Status,
				ResponseTime: result.ResponseTime,
			}
			
			if err := stor.SaveHistory(result.WebsiteID, historyEntry); err != nil {
				log.Printf("Error saving history for %s: %v", result.WebsiteID, err)
			}

			// Check for status changes and send notifications
			if prevStatus, exists := previousStatus[result.WebsiteID]; exists && prevStatus != result.Status {
				website, websiteExists := monitorEngine.GetWebsite(result.WebsiteID)
				if websiteExists {
					event := notification.StatusChangeEvent{
						WebsiteID:    result.WebsiteID,
						WebsiteName:  website.Name,
						WebsiteURL:   website.URL,
						OldStatus:    prevStatus,
						NewStatus:    result.Status,
						ResponseTime: result.ResponseTime,
						Timestamp:    result.Timestamp,
						Emails:       website.NotificationEmails,
						SlackWebhook: website.SlackWebhook,
					}
					notificationManager.SendStatusChange(event)
				}
			}
			
			previousStatus[result.WebsiteID] = result.Status
		}
	}()

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nShutting down gracefully...")
		
		// Stop monitor engine
		monitorEngine.Stop()
		
		// Stop notification manager
		notificationManager.Stop()
		
		// Save current state
		websites := monitorEngine.GetAllWebsites()
		if err := stor.SaveWebsites(websites); err != nil {
			log.Printf("Error saving websites during shutdown: %v", err)
		}
		
		// Give some time for cleanup
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}()

	// Configure Beego
	beego.BConfig.WebConfig.DirectoryIndex = true
	// The static path is set in routers/router.go init() function
	beego.BConfig.Listen.HTTPAddr = "0.0.0.0"
	beego.BConfig.Listen.HTTPPort = 8081 // Ensure this matches your desired port

	fmt.Println("Starting Uptime Monitor on http://0.0.0.0:8081" )
	fmt.Printf("Monitoring %d websites\n", len(websites))
	
	// Start Beego
	beego.Run()
}
