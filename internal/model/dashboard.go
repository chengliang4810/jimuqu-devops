package model

type HomeStatsTotal struct {
	DeployCount          int64   `json:"deploy_count"`
	SuccessCount         int64   `json:"success_count"`
	FailedCount          int64   `json:"failed_count"`
	RunningCount         int64   `json:"running_count"`
	QueuedCount          int64   `json:"queued_count"`
	ProjectCount         int64   `json:"project_count"`
	SuccessRate          float64 `json:"success_rate"`
	AverageDeployPerProj float64 `json:"average_deploy_per_project"`
	AverageDurationSec   int64   `json:"average_deploy_duration_seconds"`
	LastDeployAt         string  `json:"last_deploy_at"`
}

type HomeStatsDaily struct {
	Date         string  `json:"date"`
	DeployCount  int64   `json:"deploy_count"`
	SuccessCount int64   `json:"success_count"`
	FailedCount  int64   `json:"failed_count"`
	RunningCount int64   `json:"running_count"`
	QueuedCount  int64   `json:"queued_count"`
	SuccessRate  float64 `json:"success_rate"`
}

type HomeStatsHourly struct {
	Hour         string  `json:"hour"`
	DeployCount  int64   `json:"deploy_count"`
	SuccessCount int64   `json:"success_count"`
	FailedCount  int64   `json:"failed_count"`
	RunningCount int64   `json:"running_count"`
	QueuedCount  int64   `json:"queued_count"`
	SuccessRate  float64 `json:"success_rate"`
}

type HomeProjectRank struct {
	ProjectID    int64   `json:"project_id"`
	ProjectName  string  `json:"project_name"`
	Branch       string  `json:"branch"`
	DeployCount  int64   `json:"deploy_count"`
	SuccessCount int64   `json:"success_count"`
	FailedCount  int64   `json:"failed_count"`
	RunningCount int64   `json:"running_count"`
	QueuedCount  int64   `json:"queued_count"`
	SuccessRate  float64 `json:"success_rate"`
	LastDeployAt string  `json:"last_deploy_at"`
}

type HomeDashboard struct {
	Total    HomeStatsTotal    `json:"total"`
	Daily    []HomeStatsDaily  `json:"daily"`
	Hourly   []HomeStatsHourly `json:"hourly"`
	Projects []HomeProjectRank `json:"projects"`
}
