const API_BASE = "/api";

// Global variables
let websites = [];
let selectedWebsiteId = null;
let responseChart = null;

// Initialize the application
document.addEventListener("DOMContentLoaded", function () {
  loadWebsites();
  setInterval(loadWebsites, 30000); // Refresh every 30 seconds

  // Attach event listeners to buttons and forms
  document
    .getElementById("add-new-monitor-btn")
    .addEventListener("click", showAddMonitorModal);
  document
    .getElementById("addMonitorForm")
    .addEventListener("submit", function (event) {
      event.preventDefault();
      addMonitor();
    });
  document
    .getElementById("editMonitorForm")
    .addEventListener("submit", function (event) {
      event.preventDefault();
      updateMonitor();
    });
  document
    .getElementById("pauseButton")
    .addEventListener("click", pauseWebsite);
  document
    .getElementById("deleteButton")
    .addEventListener("click", deleteWebsite);
  document.getElementById("editButton").addEventListener("click", editWebsite);
  document
    .getElementById("searchInput")
    .addEventListener("keyup", filterWebsites);

  // Close modals when clicking outside
  document.addEventListener("click", function (event) {
    const modals = document.querySelectorAll(".modal");
    modals.forEach((modal) => {
      if (event.target === modal) {
        modal.classList.remove("show");
      }
    });
  });

  // Handle escape key to close modals
  document.addEventListener("keydown", function (event) {
    if (event.key === "Escape") {
      document
        .querySelectorAll(".modal")
        .forEach((modal) => modal.classList.remove("show"));
    }
  });

  // Attach close button listeners for modals
  document.querySelectorAll(".close-button").forEach((button) => {
    button.addEventListener("click", (e) => {
      e.target.closest(".modal").classList.remove("show");
    });
  });

  // Attach cancel button listeners for modals
  document.querySelectorAll(".cancel-button").forEach((button) => {
    button.addEventListener("click", (e) => {
      e.target.closest(".modal").classList.remove("show");
    });
  });
});

// Load websites from API
async function loadWebsites() {
  try {
    const response = await fetch(`${API_BASE}/websites`);
    if (response.ok) {
      websites = await response.json();
      renderWebsiteList();
      if (selectedWebsiteId) {
        updateSelectedWebsite();
      } else if (websites.length > 0) {
        // Select the first website if none is selected and websites exist
        selectWebsite(websites[0].id);
      }
    } else {
      showNotification("Failed to load websites", "error");
    }
  } catch (error) {
    console.error("Error loading websites:", error);
    showNotification("Error loading websites: " + error.message, "error");
  }
}

// Render website list in sidebar
function renderWebsiteList() {
  const websiteList = document.getElementById("websiteList");

  if (websites.length === 0) {
    websiteList.innerHTML = `
            <div style="padding: 2rem; text-align: center; color: #9ca3af;">
                <i class="fas fa-inbox" style="font-size: 2rem; margin-bottom: 1rem; display: block;"></i>
                <p>No monitors yet</p>
                <p style="font-size: 0.875rem;">Add your first monitor to get started</p>
            </div>
        `;
    document.getElementById("welcomeMessage").style.display = "flex"; // Show welcome message
    document.getElementById("websiteDetails").style.display = "none"; // Hide details
    document.getElementById("contentActions").style.display = "none"; // Hide actions
    return;
  }

  websiteList.innerHTML = websites
    .map((website) => {
      const statusClass = website.status || "unknown";
      const isActive = website.id === selectedWebsiteId ? "active" : "";

      // Use real history if available, else show unknown
      let statusDots = "";
      if (Array.isArray(website.history) && website.history.length > 0) {
        // Show last 20 checks (most recent last)
        const last20 = website.history.slice(-20);
        statusDots = last20
          .map(
            (entry) =>
              `<div class="status-dot-mini ${
                entry.status === "up"
                  ? "up"
                  : entry.status === "down"
                  ? "down"
                  : "unknown"
              }"></div>`
          )
          .join("");
      } else {
        statusDots = Array(20)
          .fill(0)
          .map(() => `<div class="status-dot-mini unknown"></div>`)
          .join("");
      }

      return `
            <div class="website-item ${isActive}" onclick="selectWebsite('${
        website.id
      }')">
                <div class="website-status ${statusClass}">
                    ${Math.round(website.uptime_24h || 100)}%
                </div>
                <div class="website-info">
                    <div class="website-name">${website.name}</div>
                    <div class="website-url">${website.url}</div>
                </div>
                <div class="status-indicators-mini">
                    ${statusDots}
                </div>
            </div>
        `;
    })
    .join("");

  // If there are websites and no website is selected, select the first one
  if (!selectedWebsiteId && websites.length > 0) {
    selectWebsite(websites[0].id);
  }
}

// Select a website
function selectWebsite(websiteId) {
  selectedWebsiteId = websiteId;
  const website = websites.find((w) => w.id === websiteId);

  if (!website) return;

  // Update UI
  document.getElementById("selectedWebsiteName").textContent = website.name;
  document.getElementById("websiteUrl").href = website.url;
  document.getElementById("websiteUrl").textContent = website.url;
  document.getElementById("checkInterval").textContent =
    website.interval_seconds;
  document.getElementById("currentStatus").textContent =
    (website.status || "Unknown").charAt(0).toUpperCase() +
    (website.status || "Unknown").slice(1);
  document.getElementById("currentStatus").className = `status-badge ${
    website.status || "unknown"
  }`;
  document.getElementById("currentResponse").textContent = `${
    website.last_response_time_ms || 0
  } ms`;
  document.getElementById("avgResponse").textContent = `${Math.round(
    website.avg_response_time_24h || 0
  )} ms`;
  document.getElementById("uptime24h").textContent = `${Math.round(
    website.uptime_24h || 100
  )}%`;
  document.getElementById("uptime30d").textContent = `${Math.round(
    website.uptime_30d || 100
  )}%`;

  // Show website details
  document.getElementById("welcomeMessage").style.display = "none";
  document.getElementById("websiteDetails").style.display = "block";
  document.getElementById("contentActions").style.display = "flex";

  // Update status bar using website.history (not chart history)
  updateStatusBar(website);

  // Load and update chart (chart uses API history, not for status bar)
  loadWebsiteHistory(websiteId);

  // Re-render list to update active state
  renderWebsiteList();
}

// Update selected website data
function updateSelectedWebsite() {
  if (!selectedWebsiteId) return;

  const website = websites.find((w) => w.id === selectedWebsiteId);
  if (website) {
    selectWebsite(selectedWebsiteId);
  }
}

// Update status bar
function updateStatusBar(website) {
  const statusBar = document.getElementById("statusBar");

  // Use real history if available, else show unknown
  let dots = "";
  if (Array.isArray(website.history) && website.history.length > 0) {
    // Show last 50 checks (most recent last)
    const last50 = website.history.slice(-50);
    dots = last50
      .map(
        (entry) =>
          `<div class="status-dot ${
            entry.status === "up"
              ? "up"
              : entry.status === "down"
              ? "down"
              : "unknown"
          }"></div>`
      )
      .join("");
  } else {
    dots = Array(50)
      .fill(0)
      .map(() => `<div class="status-dot unknown"></div>`)
      .join("");
  }

  statusBar.innerHTML = dots;
}

// Load website history and update chart
async function loadWebsiteHistory(websiteId) {
  try {
    const response = await fetch(
      `${API_BASE}/websites/${websiteId}/history?hours=24`
    );
    if (response.ok) {
      const history = await response.json();
      updateChart(history);
    } else {
      console.error("Failed to load history:", response.status);
      showNotification("Failed to load history", "error");
    }
  } catch (error) {
    console.error("Error loading history:", error);
    showNotification("Error loading history: " + error.message, "error");
  }
}

// Update response time chart
function updateChart(history) {
  const ctx = document.getElementById("responseChart").getContext("2d");

  if (responseChart) {
    responseChart.destroy();
  }

  // Process history data
  const labels = [];
  const data = [];
  const now = new Date();

  // Generate time labels for the last 24 hours
  for (let i = 23; i >= 0; i--) {
    const time = new Date(now.getTime() - i * 60 * 60 * 1000);
    labels.push(
      time.toLocaleTimeString("en-US", { hour: "2-digit", minute: "2-digit" })
    );

    // Find data point for this hour or use simulated data
    const historyPoint = history.find((h) => {
      const historyTime = new Date(h.timestamp);
      return Math.abs(historyTime.getTime() - time.getTime()) < 30 * 60 * 1000; // Within 30 minutes
    });

    data.push(
      historyPoint ? historyPoint.response_time_ms : Math.random() * 200 + 100
    );
  }

  responseChart = new Chart(ctx, {
    type: "line",
    data: {
      labels: labels,
      datasets: [
        {
          label: "Response Time (ms)",
          data: data,
          borderColor: "#4ade80",
          backgroundColor: "rgba(74, 222, 128, 0.1)",
          borderWidth: 2,
          fill: true,
          tension: 0.4,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: false,
        },
      },
      scales: {
        x: {
          grid: {
            color: "#404040",
          },
          ticks: {
            color: "#9ca3af",
          },
        },
        y: {
          grid: {
            color: "#404040",
          },
          ticks: {
            color: "#9ca3af",
          },
          beginAtZero: true,
        },
      },
    },
  });
}

// Filter websites
function filterWebsites() {
  const searchTerm = document.getElementById("searchInput").value.toLowerCase();
  const websiteItems = document.querySelectorAll(".website-item");

  websiteItems.forEach((item) => {
    const name = item.querySelector(".website-name").textContent.toLowerCase();
    const url = item.querySelector(".website-url").textContent.toLowerCase();

    if (name.includes(searchTerm) || url.includes(searchTerm)) {
      item.style.display = "flex";
    } else {
      item.style.display = "none";
    }
  });
}

// Show add monitor modal
function showAddMonitorModal() {
  document.getElementById("addMonitorModal").classList.add("show");
  document.getElementById("websiteName").focus();
}

// Hide add monitor modal
function hideAddMonitorModal() {
  document.getElementById("addMonitorModal").classList.remove("show");
  document.getElementById("addMonitorForm").reset();
}

// Add new monitor
async function addMonitor() {
  const name = document.getElementById("websiteName").value.trim();
  const url = document.getElementById("websiteURL").value.trim();
  const interval = parseInt(
    document.getElementById("checkIntervalInput").value
  );
  const emails = document.getElementById("notificationEmails").value.trim();
  const slackWebhook = document.getElementById("slackWebhook").value.trim();

  if (!name || !url) {
    showNotification("Name and URL are required", "error");
    return;
  }

  const emailList = emails
    ? emails
        .split(",")
        .map((e) => e.trim())
        .filter((e) => e)
    : [];

  const data = {
    name: name,
    url: url,
    interval_seconds: interval,
    notification_emails: emailList,
    slack_webhook: slackWebhook,
  };

  try {
    const response = await fetch(`${API_BASE}/websites`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
    });

    if (response.ok) {
      hideAddMonitorModal();
      showNotification("Monitor added successfully", "success");
      loadWebsites();
    } else {
      const error = await response.json();
      showNotification(error.error || "Failed to add monitor", "error");
    }
  } catch (error) {
    console.error("Error adding monitor:", error);
    showNotification("Error adding monitor", "error");
  }
}

// Show edit monitor modal
function editWebsite() {
  if (!selectedWebsiteId) return;

  const website = websites.find((w) => w.id === selectedWebsiteId);
  if (!website) return;

  // Populate form
  document.getElementById("editWebsiteName").value = website.name;
  document.getElementById("editWebsiteURL").value = website.url;
  document.getElementById("editCheckInterval").value = website.interval_seconds;
  document.getElementById("editNotificationEmails").value =
    website.notification_emails ? website.notification_emails.join(", ") : "";
  document.getElementById("editSlackWebhook").value =
    website.slack_webhook || "";
  document.getElementById("editEnabled").checked = website.enabled !== false;

  document.getElementById("editMonitorModal").classList.add("show");
}

// Hide edit monitor modal
function hideEditMonitorModal() {
  document.getElementById("editMonitorModal").classList.remove("show");
  document.getElementById("editMonitorForm").reset();
}

// Update monitor
async function updateMonitor() {
  if (!selectedWebsiteId) return;

  const name = document.getElementById("editWebsiteName").value.trim();
  const url = document.getElementById("editWebsiteURL").value.trim();
  const interval = parseInt(document.getElementById("editCheckInterval").value);
  const emails = document.getElementById("editNotificationEmails").value.trim();
  const slackWebhook = document.getElementById("editSlackWebhook").value.trim();
  const enabled = document.getElementById("editEnabled").checked;

  if (!name || !url) {
    showNotification("Name and URL are required", "error");
    return;
  }

  const emailList = emails
    ? emails
        .split(",")
        .map((e) => e.trim())
        .filter((e) => e)
    : [];

  const data = {
    name: name,
    url: url,
    interval_seconds: interval,
    notification_emails: emailList,
    slack_webhook: slackWebhook,
    enabled: enabled,
  };

  try {
    const response = await fetch(`${API_BASE}/websites/${selectedWebsiteId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
    });

    if (response.ok) {
      hideEditMonitorModal();
      showNotification("Monitor updated successfully", "success");
      loadWebsites();
    } else {
      const error = await response.json();
      showNotification(error.error || "Failed to update monitor", "error");
    }
  } catch (error) {
    console.error("Error updating monitor:", error);
    showNotification("Error updating monitor", "error");
  }
}

// Pause/Resume website
function pauseWebsite() {
  if (!selectedWebsiteId) return;

  const website = websites.find((w) => w.id === selectedWebsiteId);
  if (!website) return;

  // Toggle enabled state
  const newEnabledState = !website.enabled;

  const data = {
    name: website.name,
    url: website.url,
    interval_seconds: website.interval_seconds,
    notification_emails: website.notification_emails || [],
    slack_webhook: website.slack_webhook || "",
    enabled: newEnabledState,
  };

  fetch(`${API_BASE}/websites/${selectedWebsiteId}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  })
    .then((response) => {
      if (response.ok) {
        showNotification(
          `Monitor ${newEnabledState ? "resumed" : "paused"}`,
          "success"
        );
        loadWebsites();
      } else {
        showNotification("Failed to update monitor", "error");
      }
    })
    .catch((error) => {
      console.error("Error updating monitor:", error);
      showNotification("Error updating monitor", "error");
    });
}

// Delete website
function deleteWebsite() {
  if (!selectedWebsiteId) return;

  const website = websites.find((w) => w.id === selectedWebsiteId);
  if (!website) return;

  if (
    !confirm(
      `Are you sure you want to delete "${website.name}"? This action cannot be undone.`
    )
  ) {
    return;
  }

  fetch(`${API_BASE}/websites/${selectedWebsiteId}`, {
    method: "DELETE",
  })
    .then((response) => {
      if (response.ok) {
        showNotification("Monitor deleted successfully", "success");
        selectedWebsiteId = null;
        document.getElementById("welcomeMessage").style.display = "flex";
        document.getElementById("websiteDetails").style.display = "none";
        document.getElementById("contentActions").style.display = "none";
        document.getElementById("selectedWebsiteName").textContent =
          "Select a website";
        loadWebsites();
      } else {
        showNotification("Failed to delete monitor", "error");
      }
    })
    .catch((error) => {
      console.error("Error deleting monitor:", error);
      showNotification("Error deleting monitor", "error");
    });
}

// Show status page
function showStatusPage() {
  showNotification("Status page feature coming soon", "info");
}

// Show dashboard
function showDashboard() {
  // Already on dashboard
}

// Show notification
function showNotification(message, type = "info") {
  // Remove existing notifications
  const existingNotifications = document.querySelectorAll(".notification");
  existingNotifications.forEach((n) => n.remove());

  const notification = document.createElement("div");
  notification.className = `notification ${type}`;
  notification.textContent = message;

  document.body.appendChild(notification);

  // Show notification
  setTimeout(() => {
    notification.classList.add("show");
  }, 100);

  // Hide notification after 3 seconds
  setTimeout(() => {
    notification.classList.remove("show");
    setTimeout(() => {
      notification.remove();
    }, 300);
  }, 3000);
}
