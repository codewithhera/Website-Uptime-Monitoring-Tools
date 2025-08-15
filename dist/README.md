# Uptime Monitor

A high-performance website uptime monitoring tool built with Go and Beego framework. Designed to monitor 4000-5000 websites concurrently with thread safety, JSON file storage, email/Slack notifications, and a modern dashboard UI.

## Features

### Core Monitoring
- **High Concurrency**: Monitor 4000-5000 websites simultaneously using Go goroutines
- **Thread Safety**: Robust concurrent access protection using mutexes and channels
- **Smart Request Handling**: User agent rotation and rate limiting to avoid being blocked
- **Real-time Status Detection**: HTTP status code monitoring with response time tracking
- **Configurable Intervals**: Customizable check intervals (minimum 30 seconds)

### Data Storage
- **JSON File Storage**: Simple, reliable local storage without complex database setup
- **Historical Data**: Automatic history tracking with configurable retention
- **Data Integrity**: Atomic file operations and concurrent access protection
- **Automatic Cleanup**: Old history cleanup for removed websites

### Notifications
- **Email Alerts**: SMTP-based email notifications for status changes
- **Slack Integration**: Webhook-based Slack notifications with rich formatting
- **Smart Throttling**: Prevents notification spam with configurable delays
- **Status Change Detection**: Only notifies on actual up/down transitions

### Dashboard UI
- **Modern Interface**: Dark theme UI matching professional monitoring tools
- **Real-time Updates**: Live status updates every 30 seconds
- **Interactive Charts**: Response time visualization with Chart.js
- **Search & Filter**: Quick website search and filtering capabilities
- **Mobile Responsive**: Works on desktop and mobile devices

### API
- **RESTful API**: Complete CRUD operations for website management
- **CORS Support**: Cross-origin requests enabled for frontend integration
- **JSON Responses**: Structured API responses with error handling
- **History Endpoints**: Access to historical monitoring data

## Installation

### Prerequisites
- Go 1.18 or later
- Linux/macOS/Windows

### Quick Start

1. **Download and extract** the application files
2. **Navigate** to the application directory
3. **Run** the build script:
   ```bash
   chmod +x build.sh
   ./build.sh
   ```
4. **Start** the application:
   ```bash
   cd dist
   ./start.sh
   ```
5. **Open** your browser to `http://localhost:8080`

### Manual Build

```bash
# Clone or download the source code
cd uptime-monitor

# Install dependencies
go mod tidy

# Build the application
go build -o uptime-monitor

# Run the application
./uptime-monitor
```

## Configuration

### Application Settings

Edit `conf/app.conf` to configure the application:

```ini
appname = uptime-monitor
httpport = 8080
runmode = dev

# SMTP Configuration for email notifications
smtp_host = smtp.gmail.com
smtp_port = 587
smtp_username = your-email@gmail.com
smtp_password = your-app-password
from_email = your-email@gmail.com
```

### Environment Variables

You can also use environment variables:
- `SMTP_HOST`: SMTP server hostname
- `SMTP_PORT`: SMTP server port
- `SMTP_USERNAME`: SMTP username
- `SMTP_PASSWORD`: SMTP password
- `FROM_EMAIL`: From email address

## Usage

### Adding Monitors

1. Click **"Add New Monitor"** in the dashboard
2. Fill in the website details:
   - **Name**: Display name for the website
   - **URL**: Full URL to monitor (must include http:// or https://)
   - **Check Interval**: How often to check (minimum 30 seconds)
   - **Notification Emails**: Comma-separated email addresses
   - **Slack Webhook**: Slack webhook URL for notifications

### Managing Monitors

- **View Details**: Click on any website in the sidebar to view detailed metrics
- **Edit**: Use the "Edit" button to modify website settings
- **Pause/Resume**: Temporarily disable monitoring with the "Pause" button
- **Delete**: Remove a monitor permanently with the "Delete" button

### Dashboard Features

- **Status Overview**: Green/red indicators show current status
- **Response Time Chart**: Visual representation of response times over 24 hours
- **Uptime Metrics**: 24-hour and 30-day uptime percentages
- **Search**: Filter websites by name or URL
- **Real-time Updates**: Automatic refresh every 30 seconds

## API Reference

### Websites

#### Get All Websites
```
GET /api/websites
```

#### Get Website by ID
```
GET /api/websites/{id}
```

#### Create Website
```
POST /api/websites
Content-Type: application/json

{
  "name": "Example Site",
  "url": "https://example.com",
  "interval_seconds": 60,
  "notification_emails": ["admin@example.com"],
  "slack_webhook": "https://hooks.slack.com/services/..."
}
```

#### Update Website
```
PUT /api/websites/{id}
Content-Type: application/json

{
  "name": "Updated Name",
  "url": "https://example.com",
  "interval_seconds": 120,
  "notification_emails": ["admin@example.com"],
  "slack_webhook": "",
  "enabled": true
}
```

#### Delete Website
```
DELETE /api/websites/{id}
```

#### Get Website History
```
GET /api/websites/{id}/history?hours=24
```

## Architecture

### Components

1. **Monitor Engine** (`monitor/`): Core monitoring logic with goroutines
2. **Storage System** (`storage/`): JSON file-based data persistence
3. **Notification Manager** (`notification/`): Email and Slack notifications
4. **Web Server** (`controllers/`, `routers/`): Beego-based API and UI serving
5. **Frontend** (`static/`): HTML/CSS/JavaScript dashboard

### Concurrency Model

- **Worker Goroutines**: Each website runs in its own goroutine
- **Result Channel**: Centralized result processing
- **Mutex Protection**: Thread-safe access to shared data
- **Graceful Shutdown**: Clean shutdown with data persistence

### Performance Optimizations

- **Connection Pooling**: HTTP client reuses connections
- **Memory Management**: Efficient data structures and cleanup
- **Rate Limiting**: Prevents overwhelming target websites
- **Batch Operations**: Efficient JSON file updates

## Troubleshooting

### Common Issues

**Application won't start**
- Check if port 8080 is available
- Verify Go installation and version
- Check file permissions

**Websites not appearing**
- Check browser console for JavaScript errors
- Verify API endpoints are responding
- Check application logs

**Notifications not working**
- Verify SMTP settings in `conf/app.conf`
- Test Slack webhook URLs
- Check firewall settings for outbound connections

**High memory usage**
- Reduce number of monitored websites
- Increase check intervals
- Monitor for memory leaks in logs

### Logs

Application logs are printed to stdout. To save logs to a file:

```bash
./uptime-monitor > uptime-monitor.log 2>&1
```

### Performance Tuning

For monitoring large numbers of websites:

1. **Increase check intervals** to reduce load
2. **Monitor system resources** (CPU, memory, network)
3. **Use SSD storage** for better JSON file performance
4. **Adjust Go runtime settings** if needed:
   ```bash
   export GOMAXPROCS=4
   export GOGC=100
   ```

## Development

### Project Structure

```
uptime-monitor/
├── main.go              # Application entry point
├── conf/                # Configuration files
├── controllers/         # API controllers
├── monitor/            # Core monitoring engine
├── notification/       # Notification system
├── routers/           # URL routing
├── static/            # Frontend assets
├── storage/           # Data storage layer
├── build.sh           # Build script
└── README.md          # This file
```

### Building from Source

```bash
# Install dependencies
go mod tidy

# Run tests (if available)
go test ./...

# Build
go build -o uptime-monitor

# Run
./uptime-monitor
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

This project is open source. See LICENSE file for details.

## Support

For issues and questions:
1. Check the troubleshooting section
2. Review application logs
3. Create an issue with detailed information

---

**Built with Go and Beego Framework**

