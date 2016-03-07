package metadata

import "time"

type File struct {
	Id        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	FullPath  string    `json:"full_path"`
	DirectUrl string    `json:"direct_url"`
	Url       string    `json:"url"`
	Views     int       `json:"views"`
}
