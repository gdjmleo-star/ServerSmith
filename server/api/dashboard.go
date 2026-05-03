package api

import (
	"net/http"
	"serversmith/db"
)

type CostPerGBItem struct {
	Carrier      string  `json:"carrier"`
	TotalQuotaGB float64 `json:"total_quota_gb"`
	MonthlyRent  float64 `json:"monthly_rent"`
	CostPerGB    float64 `json:"cost_per_gb"`
}

type DashboardSummary struct {
	TotalServers  int            `json:"total_servers"`
	OnlineCount   int            `json:"online_count"`
	OfflineCount  int            `json:"offline_count"`
	TotalRent     float64        `json:"total_monthly_rent"`
	TotalRentNode float64        `json:"node_rent"`
	TotalRentProj float64        `json:"project_rent"`
	TotalQuotaGB  float64        `json:"total_quota_gb"`
	TotalUsedGB   float64        `json:"total_used_gb"`
	AlertsUnread  int            `json:"alerts_unread"`
	CostPerGB     []CostPerGBItem `json:"cost_per_gb"`
}

type CarrierSummary struct {
	CarrierType  string  `json:"carrier_type"`
	ServerCount  int     `json:"server_count"`
	TotalQuotaGB float64 `json:"total_quota_gb"`
	UsedGB       float64 `json:"used_gb"`
	RemainingGB  float64 `json:"remaining_gb"`
	UsagePercent float64 `json:"usage_percent"`
	TotalRent    float64 `json:"total_rent"`
	CostPerGB    float64 `json:"cost_per_gb"`
	Abnormal     int     `json:"abnormal_count"`
}

type AlertResponse struct {
	ID              int64   `json:"id"`
	ServerID        *int64  `json:"server_id"`
	ServerName      string  `json:"server_name,omitempty"`
	AlertType       string  `json:"alert_type"`
	Message         string  `json:"message"`
	Level           string  `json:"level"`
	CreatedAt       string  `json:"created_at"`
	Acknowledged    bool    `json:"acknowledged"`
	AcknowledgedAt  *string `json:"acknowledged_at"`
}

func HandleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}

	var s DashboardSummary

	db.DB.QueryRow("SELECT COUNT(*) FROM servers").Scan(&s.TotalServers)
	db.DB.QueryRow("SELECT COUNT(*) FROM servers WHERE status='online'").Scan(&s.OnlineCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM servers WHERE status='offline'").Scan(&s.OfflineCount)
	db.DB.QueryRow("SELECT COALESCE(SUM(monthly_rent), 0) FROM servers").Scan(&s.TotalRent)
	db.DB.QueryRow("SELECT COALESCE(SUM(monthly_rent), 0) FROM servers WHERE server_type='node'").Scan(&s.TotalRentNode)
	db.DB.QueryRow("SELECT COALESCE(SUM(monthly_rent), 0) FROM servers WHERE server_type='project'").Scan(&s.TotalRentProj)
	db.DB.QueryRow("SELECT COALESCE(SUM(p.total_quota), 0) FROM server_plans p JOIN servers s ON s.id=p.server_id WHERE p.plan_type='traffic' AND s.server_type='node'").Scan(&s.TotalQuotaGB)
	db.DB.QueryRow("SELECT COALESCE(SUM(p.used_bytes), 0) FROM server_plans p JOIN servers s ON s.id=p.server_id WHERE p.plan_type='traffic' AND s.server_type='node'").Scan(&s.TotalUsedGB)
	db.DB.QueryRow("SELECT COUNT(*) FROM alerts WHERE acknowledged=0").Scan(&s.AlertsUnread)

	s.TotalUsedGB = s.TotalUsedGB / 1_000_000_000

	// Compute cost_per_gb by carrier
	costRows, err := db.DB.Query(`
		SELECT s.carrier_type, COALESCE(SUM(s.monthly_rent),0), COALESCE(SUM(p.total_quota),0)
		FROM servers s
		JOIN server_plans p ON p.server_id = s.id
		WHERE s.server_type='node' AND s.carrier_type IS NOT NULL AND p.plan_type='traffic'
		GROUP BY s.carrier_type ORDER BY s.carrier_type
	`)
	if err == nil {
		defer costRows.Close()
		for costRows.Next() {
			var item CostPerGBItem
			if scanErr := costRows.Scan(&item.Carrier, &item.MonthlyRent, &item.TotalQuotaGB); scanErr == nil {
				if item.TotalQuotaGB > 0 {
					item.CostPerGB = item.MonthlyRent / item.TotalQuotaGB
				}
				s.CostPerGB = append(s.CostPerGB, item)
			}
		}
	}
	if s.CostPerGB == nil {
		s.CostPerGB = []CostPerGBItem{}
	}

	writeOK(w, s)
}

func HandleDashboardByCarrier(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}

	rows, err := db.DB.Query(`
		SELECT
			s.carrier_type,
			COUNT(s.id),
			COALESCE(SUM(p.total_quota), 0),
			COALESCE(SUM(p.used_bytes), 0),
			COALESCE(SUM(s.monthly_rent), 0)
		FROM servers s
		JOIN server_plans p ON p.server_id = s.id
		WHERE s.server_type = 'node'
		  AND s.carrier_type IS NOT NULL
		  AND p.plan_type = 'traffic'
		GROUP BY s.carrier_type
		ORDER BY s.carrier_type
	`)
	if err != nil {
		writeError(w, 500, "query error: "+err.Error())
		return
	}
	defer rows.Close()

	var carriers []CarrierSummary
	for rows.Next() {
		var cs CarrierSummary
		var totalBytes int64
		if err := rows.Scan(&cs.CarrierType, &cs.ServerCount, &cs.TotalQuotaGB, &totalBytes, &cs.TotalRent); err != nil {
			continue
		}
		cs.UsedGB = float64(totalBytes) / 1_000_000_000
		cs.RemainingGB = cs.TotalQuotaGB - cs.UsedGB
		if cs.RemainingGB < 0 {
			cs.RemainingGB = 0
		}
		if cs.TotalQuotaGB > 0 {
			cs.UsagePercent = (cs.UsedGB / cs.TotalQuotaGB) * 100
		}
		if cs.TotalQuotaGB > 0 {
			cs.CostPerGB = cs.TotalRent / cs.TotalQuotaGB
		}

		// Count abnormal nodes (>80% usage)
		db.DB.QueryRow(`
			SELECT COUNT(*) FROM server_plans p
			JOIN servers s ON s.id=p.server_id
			WHERE s.carrier_type=? AND p.plan_type='traffic'
			  AND CAST(p.used_bytes AS REAL) / (p.total_quota * 1000000000) > 0.8
		`, cs.CarrierType).Scan(&cs.Abnormal)

		carriers = append(carriers, cs)
	}

	if carriers == nil {
		carriers = []CarrierSummary{}
	}
	writeOK(w, carriers)
}


