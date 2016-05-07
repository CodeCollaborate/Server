package dbfs

import (
	"errors"
	"time"
)

// NoDbChange : No rows or values in the DB were changed, which was an unexpected result
var ErrNoDbChange = errors.New("No entries were correctly altered")

// DbNotInitialized : active db connection does not exist
var ErrDbNotInitialized = errors.New("The database was not propperly initialized before execution")

// Project is the type which represents a row in the MySQL `Project` table
type Project struct {
	ProjectID       int64
	ProjectName     string
	PermissionLevel int
}

// ProjectPermission is the type which represents the permission relationship on projects
type ProjectPermission struct {
	Username        string
	PermissionLevel int
	GrantedBy       string
	GrantedDate     time.Time
}

// File is the type which represents a row in the MySQL `File` table
type File struct {
	FileID       int64
	Creator      string
	CreationDate time.Time
	RelativePath string
	ProjectID    int64
	Filename     string
}
