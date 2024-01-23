package types

type ApiReponses struct {
	Timestamp   int64    `json:"timestamp"`
	Success     bool     `json:"success"`
	Query       string   `json:"query"`
	Lang        string   `json:"lang"`
	JobId       string   `json:"job_id"`
	JobStatus   string   `json:"job_status"`
	JobProgress int      `json:"job_progress"`
	Places      []*Place `json:"places"`
}

type CreateJobRequest struct {
	Query       string `json:"query"`
	Lang        string `json:"lang"`
	Limit       int    `json:"limit"`
	Timeout     int    `json:"timeout"`
	Cooldown    int    `json:"cooldown"`
	Concurrency int    `json:"concurrency"`
	Emails      bool   `json:"emails"`
}
