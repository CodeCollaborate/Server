package datahandling

import (
	"time"

	"github.com/CodeCollaborate/Server/modules/dbfs"
)

var fileRequestsSetup = false
var newFileVersion int64 = 1

// File aggregates information relating to an individual file
// TODO(wongb): Change all responses and notifications to use this struct; add creator and creation date
type File struct {
	FileID       int64
	Filename     string
	RelativePath string
	Version      int64
}

// initProjectRequests populates the requestMap from requestmap.go with the appropriate constructors for the project methods
func initFileRequests() {
	if fileRequestsSetup {
		return
	}

	authenticatedRequestMap["File.Create"] = func(req *abstractRequest) (request, error) {
		return commonJSON(new(fileCreateRequest), req)
	}

	authenticatedRequestMap["File.Rename"] = func(req *abstractRequest) (request, error) {
		return commonJSON(new(fileRenameRequest), req)
	}

	authenticatedRequestMap["File.Move"] = func(req *abstractRequest) (request, error) {
		return commonJSON(new(fileMoveRequest), req)
	}

	authenticatedRequestMap["File.Delete"] = func(req *abstractRequest) (request, error) {
		return commonJSON(new(fileDeleteRequest), req)
	}

	authenticatedRequestMap["File.Change"] = func(req *abstractRequest) (request, error) {
		return commonJSON(new(fileChangeRequest), req)
	}

	authenticatedRequestMap["File.Pull"] = func(req *abstractRequest) (request, error) {
		return commonJSON(new(filePullRequest), req)
	}

	fileRequestsSetup = true
}

// File.Create
type fileCreateRequest struct {
	Name         string
	RelativePath string
	ProjectID    int64
	FileBytes    []byte
	abstractRequest
}

func (f *fileCreateRequest) setAbstractRequest(req *abstractRequest) {
	f.abstractRequest = *req
}

func (f fileCreateRequest) process(db dbfs.DBFS) ([]dhClosure, error) {
	// TODO (normal/required): check if permission high enough on project
	res := new(serverMessageWrapper)
	res.Timestamp = time.Now().Unix()
	res.Type = "Response"

	not := new(serverMessageWrapper)
	not.Timestamp = res.Timestamp
	not.Type = "Notification"

	_, err := db.FileWrite(f.RelativePath, f.Name, f.ProjectID, f.FileBytes)
	if err != nil {
		res.ServerMessage = response{
			Status: fail,
			Tag:    f.Tag,
			Data:   struct{}{}}
		return []dhClosure{toSenderClosure{msg: res}}, nil
	}

	fileID, err := db.MySQLFileCreate(f.SenderID, f.Name, f.RelativePath, f.ProjectID)
	if err != nil {
		res.ServerMessage = response{
			Status: fail,
			Tag:    f.Tag,
			Data:   struct{}{}}
		return []dhClosure{toSenderClosure{msg: res}}, nil
	}

	err = db.CBInsertNewFile(fileID, newFileVersion, make([]string, 0))

	if err != nil {
		res.ServerMessage = response{
			Status: servfail,
			Tag:    f.Tag,
			Data:   struct{}{}}
		return []dhClosure{toSenderClosure{msg: res}}, nil
	}
	res.ServerMessage = response{
		Status: success,
		Tag:    f.Tag,
		Data: struct {
			FileID int64
		}{
			FileID: fileID,
		}}
	not.ServerMessage = notification{
		Resource:   f.Resource,
		Method:     f.Method,
		ResourceID: f.ProjectID,
		Data: struct {
			File File
		}{
			File: File{
				FileID:       fileID,
				Filename:     f.Name,
				RelativePath: f.RelativePath,
				Version:      newFileVersion,
			},
		}}
	return []dhClosure{toSenderClosure{msg: res}, toRabbitChannelClosure{msg: not, routingKey: f.ProjectID}}, nil
}

// File.Rename
type fileRenameRequest struct {
	FileID  int64
	NewName string
	abstractRequest
}

func (f *fileRenameRequest) setAbstractRequest(req *abstractRequest) {
	f.abstractRequest = *req
}

func (f fileRenameRequest) process(db dbfs.DBFS) ([]dhClosure, error) {
	res := new(serverMessageWrapper)
	res.Timestamp = time.Now().Unix()
	res.Type = "Response"

	not := new(serverMessageWrapper)
	not.Timestamp = res.Timestamp
	not.Type = "Notification"

	res.ServerMessage = response{
		Status: fail,
		Tag:    f.Tag,
		Data:   struct{}{}}

	fileMeta, err := db.MySQLFileGetInfo(f.FileID)
	if err != nil {
		return []dhClosure{toSenderClosure{msg: res}}, err
	}

	// TODO (normal/required): check if permission high enough on project (fileMeta.ProjectID)

	err = db.MySQLFileRename(f.FileID, f.NewName)
	if err != nil {
		return []dhClosure{toSenderClosure{msg: res}}, err
	}

	err = db.FileMove(fileMeta.RelativePath, fileMeta.Filename, fileMeta.RelativePath, f.NewName, fileMeta.ProjectID)
	if err != nil {
		return []dhClosure{toSenderClosure{msg: res}}, err
	}

	res.ServerMessage = response{
		Status: success,
		Tag:    f.Tag,
		Data:   struct{}{}}
	not.ServerMessage = notification{
		Resource:   f.Resource,
		Method:     f.Method,
		ResourceID: f.FileID,
		Data: struct {
			NewName string
		}{
			NewName: f.NewName,
		}}
	return []dhClosure{toSenderClosure{msg: res}, toRabbitChannelClosure{msg: not, routingKey: fileMeta.ProjectID}}, nil
}

// File.Move
type fileMoveRequest struct {
	FileID  int64
	NewPath string
	abstractRequest
}

func (f *fileMoveRequest) setAbstractRequest(req *abstractRequest) {
	f.abstractRequest = *req
}

func (f fileMoveRequest) process(db dbfs.DBFS) ([]dhClosure, error) {
	res := new(serverMessageWrapper)
	res.Timestamp = time.Now().Unix()
	res.Type = "Response"

	not := new(serverMessageWrapper)
	not.Timestamp = res.Timestamp
	not.Type = "Notification"

	res.ServerMessage = response{
		Status: fail,
		Tag:    f.Tag,
		Data:   struct{}{}}

	fileMeta, err := db.MySQLFileGetInfo(f.FileID)
	if err != nil {
		return []dhClosure{toSenderClosure{msg: res}}, err
	}

	// TODO (normal/required): check if permission high enough on project (fileMeta.ProjectID)

	err = db.MySQLFileMove(f.FileID, f.NewPath)
	if err != nil {
		return []dhClosure{toSenderClosure{msg: res}}, err
	}

	err = db.FileMove(fileMeta.RelativePath, fileMeta.Filename, f.NewPath, fileMeta.Filename, fileMeta.ProjectID)
	if err != nil {
		return []dhClosure{toSenderClosure{msg: res}}, err
	}

	res.ServerMessage = response{
		Status: success,
		Tag:    f.Tag,
		Data:   struct{}{}}
	not.ServerMessage = notification{
		Resource:   f.Resource,
		Method:     f.Method,
		ResourceID: f.FileID,
		Data: struct {
			NewPath string
		}{
			NewPath: f.NewPath,
		}}
	return []dhClosure{toSenderClosure{msg: res}, toRabbitChannelClosure{msg: not, routingKey: fileMeta.ProjectID}}, nil
}

// File.Delete
type fileDeleteRequest struct {
	FileID int64
	abstractRequest
}

func (f *fileDeleteRequest) setAbstractRequest(req *abstractRequest) {
	f.abstractRequest = *req
}

func (f fileDeleteRequest) process(db dbfs.DBFS) ([]dhClosure, error) {
	res := new(serverMessageWrapper)
	res.Timestamp = time.Now().Unix()
	res.Type = "Response"

	not := new(serverMessageWrapper)
	not.Timestamp = res.Timestamp
	not.Type = "Notification"

	res.ServerMessage = response{
		Status: fail,
		Tag:    f.Tag,
		Data:   struct{}{}}

	fileMeta, err := db.MySQLFileGetInfo(f.FileID)
	if err != nil {
		return []dhClosure{toSenderClosure{msg: res}}, err
	}

	// TODO (normal/required): check if permission high enough on project (fileMeta.ProjectID)

	err = db.FileDelete(fileMeta.RelativePath, fileMeta.Filename, fileMeta.ProjectID)
	if err != nil {
		return []dhClosure{toSenderClosure{msg: res}}, err
	}

	err = db.MySQLFileDelete(f.FileID)
	if err != nil {
		return []dhClosure{toSenderClosure{msg: res}}, err
	}

	err = db.CBDeleteFile(f.FileID)
	if err != nil {
		return []dhClosure{toSenderClosure{msg: res}}, err
	}

	res.ServerMessage = response{
		Status: success,
		Tag:    f.Tag,
		Data:   struct{}{}}
	not.ServerMessage = notification{
		Resource:   f.Resource,
		Method:     f.Method,
		ResourceID: f.FileID,
		Data:       struct{}{}}
	return []dhClosure{toSenderClosure{msg: res}, toRabbitChannelClosure{msg: not, routingKey: fileMeta.ProjectID}}, nil
}

// File.Change
type fileChangeRequest struct {
	FileID          int64
	Changes         []string
	BaseFileVersion int64
	abstractRequest
}

func (f *fileChangeRequest) setAbstractRequest(req *abstractRequest) {
	f.abstractRequest = *req
}

func (f fileChangeRequest) process(db dbfs.DBFS) ([]dhClosure, error) {
	res := new(serverMessageWrapper)
	res.Timestamp = time.Now().Unix()
	res.Type = "Response"

	not := new(serverMessageWrapper)
	not.Timestamp = res.Timestamp
	not.Type = "Notification"

	res.ServerMessage = response{
		Status: fail,
		Tag:    f.Tag,
		Data:   struct{}{}}

	fileMeta, err := db.MySQLFileGetInfo(f.FileID)
	if err != nil {
		return []dhClosure{toSenderClosure{msg: res}}, err
	}

	// TODO (normal/required): check if permission high enough on project (fileMeta.ProjectID)

	// TODO (normal/required): verify changes are valid changes
	version, err := db.CBAppendFileChange(f.FileID, f.BaseFileVersion, f.Changes)
	if err != nil {
		if err == dbfs.ErrVersionOutOfDate {
			res.ServerMessage = response{
				Status: versionOutOfDate,
				Tag:    f.Tag,
				Data:   struct{}{}}
		}
		return []dhClosure{toSenderClosure{msg: res}}, err
	}

	res.ServerMessage = response{
		Status: success,
		Tag:    f.Tag,
		Data: struct {
			FileVersion int64
		}{
			FileVersion: version,
		}}

	not.ServerMessage = notification{
		Resource:   f.Resource,
		Method:     f.Method,
		ResourceID: f.FileID,
		Data: struct {
			BaseFileVersion int64
			FileVersion     int64
			Changes         []string
		}{
			BaseFileVersion: f.BaseFileVersion,
			FileVersion:     version,
			Changes:         f.Changes,
		}}

	return []dhClosure{toSenderClosure{msg: res}, toRabbitChannelClosure{msg: not, routingKey: fileMeta.ProjectID}}, nil
}

// File.Pull
type filePullRequest struct {
	FileID int64
	abstractRequest
}

func (f *filePullRequest) setAbstractRequest(req *abstractRequest) {
	f.abstractRequest = *req
}

func (f filePullRequest) process(db dbfs.DBFS) ([]dhClosure, error) {
	// TODO (normal/required): check if permission high enough on project

	res := new(serverMessageWrapper)
	res.Timestamp = time.Now().Unix()
	res.Type = "Response"

	res.ServerMessage = response{
		Status: fail,
		Tag:    f.Tag,
		Data:   struct{}{}}

	fileMeta, err := db.MySQLFileGetInfo(f.FileID)
	if err != nil {
		return []dhClosure{toSenderClosure{msg: res}}, err
	}

	rawFile, err := db.FileRead(fileMeta.RelativePath, fileMeta.Filename, fileMeta.ProjectID)
	if err != nil {
		return []dhClosure{toSenderClosure{msg: res}}, err
	}

	changes, err := db.CBGetFileChanges(f.FileID)
	if err != nil {
		return []dhClosure{toSenderClosure{msg: res}}, err
	}

	res.ServerMessage = response{
		Status: success,
		Tag:    f.Tag,
		Data: struct {
			FileBytes []byte
			Changes   []string
		}{
			FileBytes: *rawFile,
			Changes:   changes,
		}}

	return []dhClosure{toSenderClosure{msg: res}}, nil
}
