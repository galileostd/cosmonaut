package db

import "time"

type Job struct {
    ID          string     `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
    ComponentID string     `gorm:"type:uuid;index" json:"component_id"`
    Name        string     `json:"name"`
    Status      string     `gorm:"default:'queued'" json:"status"`
    SubmitTime  time.Time  `json:"submit_time"`
    StartTime   *time.Time `json:"start_time"`
    EndTime     *time.Time `json:"end_time"`
    Duration    string     `json:"duration"`
    ErrorMsg    string     `json:"error_msg"`
    Logs        string     `gorm:"type:text" json:"logs"`
    CreatedAt   time.Time  `json:"created_at"`
}

type User struct {
    ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
    Username  string    `gorm:"unique;not null" json:"username"`
    Email     string    `gorm:"unique;not null" json:"email"`
    Password  string    `gorm:"not null" json:"-"`
    FullName  string    `json:"full_name"`
    Role      string    `gorm:"default:'user'" json:"role"`
    CreatedAt time.Time `json:"created_at"`
}