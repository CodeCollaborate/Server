package datahandling

import (
	"time"

	"github.com/CodeCollaborate/Server/modules/dbfs"
	"github.com/CodeCollaborate/Server/modules/rabbitmq"
	"github.com/CodeCollaborate/Server/utils"
)

var projectRequestsSetup = false

// TODO(wongb): Create & Use a Project struct

// initProjectRequests populates the requestMap from requestmap.go with the appropriate constructors for the project methods
func initProjectRequests() {
	if projectRequestsSetup {
		return
	}

	authenticatedRequestMap["Project.Create"] = func(req *abstractRequest) (request, error) {
		return commonJSON(new(projectCreateRequest), req)
	}

	authenticatedRequestMap["Project.Rename"] = func(req *abstractRequest) (request, error) {
		return commonJSON(new(projectRenameRequest), req)
	}

	authenticatedRequestMap["Project.GetPermissionsConstants"] = func(req *abstractRequest) (request, error) {
		return commonJSON(new(projectGetPermissionConstantsRequest), req)
	}

	authenticatedRequestMap["Project.GrantPermissions"] = func(req *abstractRequest) (request, error) {
		return commonJSON(new(projectGrantPermissionsRequest), req)
	}

	authenticatedRequestMap["Project.RevokePermissions"] = func(req *abstractRequest) (request, error) {
		return commonJSON(new(projectRevokePermissionsRequest), req)
	}

	authenticatedRequestMap["Project.GetOnlineClients"] = func(req *abstractRequest) (request, error) {
		return commonJSON(new(projectGetOnlineClientsRequest), req)
	}

	authenticatedRequestMap["Project.Lookup"] = func(req *abstractRequest) (request, error) {
		return commonJSON(new(projectLookupRequest), req)
	}

	authenticatedRequestMap["Project.GetFiles"] = func(req *abstractRequest) (request, error) {
		return commonJSON(new(projectGetFilesRequest), req)
	}

	authenticatedRequestMap["Project.Subscribe"] = func(req *abstractRequest) (request, error) {
		return commonJSON(new(projectSubscribeRequest), req)
	}

	authenticatedRequestMap["Project.Unsubscribe"] = func(req *abstractRequest) (request, error) {
		return commonJSON(new(projectUnsubscribeRequest), req)
	}

	authenticatedRequestMap["Project.Delete"] = func(req *abstractRequest) (request, error) {
		return commonJSON(new(projectDeleteRequest), req)
	}

	projectRequestsSetup = true
}

// Project.Create
type projectCreateRequest struct {
	Name string
	abstractRequest
}

func (p *projectCreateRequest) setAbstractRequest(req *abstractRequest) {
	p.abstractRequest = *req
}

func (p projectCreateRequest) process(db dbfs.DBFS) ([]dhClosure, error) {
	projectID, err := db.MySQLProjectCreate(p.SenderID, p.Name)
	if err != nil {
		//if err == project already exists {
		// TODO(shapiro): implement a specific error for this on the mysql.go side
		//}
		return []dhClosure{toSenderClosure{msg: newEmptyResponse(servfail, p.Tag)}}, nil
	}

	res := response{
		Status: success,
		Tag:    p.Tag,
		Data: struct {
			ProjectID int64
		}{
			ProjectID: projectID,
		},
	}.wrap()

	return []dhClosure{toSenderClosure{msg: res}}, nil
}

// Project.Rename
type projectRenameRequest struct {
	ProjectID int64
	NewName   string
	abstractRequest
}

func (p *projectRenameRequest) setAbstractRequest(req *abstractRequest) {
	p.abstractRequest = *req
}

func (p projectRenameRequest) process(db dbfs.DBFS) ([]dhClosure, error) {
	// TODO: check if permission high enough on project

	err := db.MySQLProjectRename(p.ProjectID, p.NewName)
	if err != nil {
		return []dhClosure{toSenderClosure{msg: newEmptyResponse(servfail, p.Tag)}}, err
	}

	res := newEmptyResponse(success, p.Tag)
	not := notification{
		Resource:   p.Resource,
		Method:     p.Method,
		ResourceID: p.ProjectID,
		Data: struct {
			NewName string
		}{
			NewName: p.NewName,
		},
	}.wrap()

	return []dhClosure{toSenderClosure{msg: res}, toRabbitChannelClosure{msg: not, projectID: p.ProjectID}}, nil
}

// Project.GetPermissionConstants
type projectGetPermissionConstantsRequest struct {
	abstractRequest
}

func (p *projectGetPermissionConstantsRequest) setAbstractRequest(req *abstractRequest) {
	p.abstractRequest = *req
}

func (p projectGetPermissionConstantsRequest) process(db dbfs.DBFS) ([]dhClosure, error) {
	// TODO (non-immediate/required): figure out how we want to do projectGetPermissionConstantsRequest
	utils.LogWarn("ProjectGetPermissionConstants not implemented", nil)

	return []dhClosure{toSenderClosure{msg: newEmptyResponse(unimplemented, p.Tag)}}, nil
}

// Project.GrantPermissions
type projectGrantPermissionsRequest struct {
	ProjectID       int64
	GrantUsername   string
	PermissionLevel int
	abstractRequest
}

func (p projectGrantPermissionsRequest) process(db dbfs.DBFS) ([]dhClosure, error) {
	// TODO: check if permission high enough on project

	err := db.MySQLProjectGrantPermission(p.ProjectID, p.GrantUsername, p.PermissionLevel, p.SenderID)
	if err != nil {
		return []dhClosure{toSenderClosure{msg: newEmptyResponse(servfail, p.Tag)}}, err
	}

	res := newEmptyResponse(success, p.Tag)
	not := notification{
		Resource:   p.Resource,
		Method:     p.Method,
		ResourceID: p.ProjectID,
		Data: struct {
			GrantUsername   string
			PermissionLevel int
		}{
			GrantUsername:   p.GrantUsername,
			PermissionLevel: p.PermissionLevel,
		},
	}.wrap()

	return []dhClosure{toSenderClosure{msg: res}, toRabbitChannelClosure{msg: not, projectID: p.ProjectID}}, nil
}

func (p *projectGrantPermissionsRequest) setAbstractRequest(req *abstractRequest) {
	p.abstractRequest = *req
}

// Project.RevokePermissions
type projectRevokePermissionsRequest struct {
	ProjectID      int64
	RevokeUsername string
	abstractRequest
}

func (p projectRevokePermissionsRequest) process(db dbfs.DBFS) ([]dhClosure, error) {
	// TODO: check if permission high enough on project
	err := db.MySQLProjectRevokePermission(p.ProjectID, p.RevokeUsername, p.SenderID)

	if err != nil {
		return []dhClosure{toSenderClosure{msg: newEmptyResponse(servfail, p.Tag)}}, err
	}

	res := newEmptyResponse(success, p.Tag)
	not := notification{
		Resource:   p.Resource,
		Method:     p.Method,
		ResourceID: p.ProjectID,
		Data: struct {
			RevokeUsername string
		}{
			RevokeUsername: p.RevokeUsername,
		},
	}.wrap()

	return []dhClosure{toSenderClosure{msg: res}, toRabbitChannelClosure{msg: not, projectID: p.ProjectID}}, nil
}

func (p *projectRevokePermissionsRequest) setAbstractRequest(req *abstractRequest) {
	p.abstractRequest = *req
}

// Project.GetOnlineClients
type projectGetOnlineClientsRequest struct {
	ProjectID int64
	abstractRequest
}

func (p projectGetOnlineClientsRequest) process(db dbfs.DBFS) ([]dhClosure, error) {
	// TODO: implement on redis (and actually implement redis)
	utils.LogWarn("ProjectGetOnlineClients not implemented", nil)

	return []dhClosure{toSenderClosure{msg: newEmptyResponse(unimplemented, p.Tag)}}, nil
}

func (p *projectGetOnlineClientsRequest) setAbstractRequest(req *abstractRequest) {
	p.abstractRequest = *req
}

// Project.Lookup
type projectLookupRequest struct {
	ProjectIDs []int64
	abstractRequest
}

// this request returns a slice of results for the projects we found, so we need the object that goes in that slice
type projectLookupResult struct {
	ProjectID   int64
	Name        string
	Permissions map[string](dbfs.ProjectPermission)
}

func (p projectLookupRequest) process(db dbfs.DBFS) ([]dhClosure, error) {
	/*
		We could do
			data := make([]interface{}, len(p.ProjectIDs))
		but it seems like poor practice and makes the object oriented side of my brain cry
	*/
	resultData := make([]projectLookupResult, len(p.ProjectIDs))

	var errOut error
	i := 0
	for _, id := range p.ProjectIDs {
		// TODO: see note at modules/dbfs/mysql.go:307
		name, permissions, err := db.MySQLProjectLookup(id, p.SenderID)
		if err != nil {
			errOut = err
		} else {
			resultData[i] = projectLookupResult{
				ProjectID:   id,
				Name:        name,
				Permissions: permissions}
			i++
		}
	}
	// shrink to cut off remainder left by errors
	resultData = resultData[:i]

	if errOut != nil {
		if len(resultData) == 0 {
			res := response{
				Status: fail,
				Tag:    p.Tag,
				Data: struct {
					Projects []projectLookupResult
				}{
					Projects: resultData,
				},
			}.wrap()
			return []dhClosure{toSenderClosure{msg: res}}, nil
		}
		res := response{
			Status: partialfail,
			Tag:    p.Tag,
			Data: struct {
				Projects []projectLookupResult
			}{
				Projects: resultData,
			},
		}.wrap()
		return []dhClosure{toSenderClosure{msg: res}}, nil
	}

	res := response{
		Status: success,
		Tag:    p.Tag,
		Data: struct {
			Projects []projectLookupResult
		}{
			Projects: resultData,
		},
	}.wrap()

	return []dhClosure{toSenderClosure{msg: res}}, nil
}

func (p *projectLookupRequest) setAbstractRequest(req *abstractRequest) {
	p.abstractRequest = *req
}

// Project.GetFiles
type projectGetFilesRequest struct {
	ProjectID int64
	abstractRequest
}

type fileLookupResult struct {
	FileID       int64
	Filename     string
	Creator      string
	CreationDate time.Time
	RelativePath string
	Version      int64
}

func (p projectGetFilesRequest) process(db dbfs.DBFS) ([]dhClosure, error) {
	files, err := db.MySQLProjectGetFiles(p.ProjectID)
	if err != nil {
		res := response{
			Status: fail,
			Tag:    p.Tag,
			Data: struct {
				Files []fileLookupResult
			}{
				Files: make([]fileLookupResult, 0),
			},
		}.wrap()

		return []dhClosure{toSenderClosure{msg: res}}, nil
	}

	resultData := make([]fileLookupResult, len(files))

	i := 0
	var errOut error
	for _, file := range files {
		version, err := db.CBGetFileVersion(file.FileID)
		if err != nil {
			errOut = err
		} else {
			resultData[i] = fileLookupResult{
				FileID:       file.FileID,
				Filename:     file.Filename,
				Creator:      file.Creator,
				CreationDate: file.CreationDate,
				RelativePath: file.RelativePath,
				Version:      version}
			i++
		}
	}
	// shrink to cut off remainder left by errors
	resultData = resultData[:i]

	if errOut != nil {
		if len(resultData) == 0 {
			res := response{
				Status: fail,
				Tag:    p.Tag,
				Data: struct {
					Files []fileLookupResult
				}{
					Files: resultData,
				},
			}.wrap()
			return []dhClosure{toSenderClosure{msg: res}}, nil
		}
		res := response{
			Status: partialfail,
			Tag:    p.Tag,
			Data: struct {
				Files []fileLookupResult
			}{
				Files: resultData,
			},
		}.wrap()
		return []dhClosure{toSenderClosure{msg: res}}, nil
	}
	res := response{
		Status: success,
		Tag:    p.Tag,
		Data: struct {
			Files []fileLookupResult
		}{
			Files: resultData,
		},
	}.wrap()

	return []dhClosure{toSenderClosure{msg: res}}, nil
}

func (p *projectGetFilesRequest) setAbstractRequest(req *abstractRequest) {
	p.abstractRequest = *req
}

// Project.Subscribe
type projectSubscribeRequest struct {
	ProjectID int64
	abstractRequest
}

func (p projectSubscribeRequest) process(db dbfs.DBFS) ([]dhClosure, error) {
	subscribeClosure := rabbitChannelSubscribeClosure{
		key: rabbitmq.RabbitProjectQueueName(p.ProjectID),
		tag: p.Tag,
	}
	return []dhClosure{subscribeClosure}, nil
}

func (p *projectSubscribeRequest) setAbstractRequest(req *abstractRequest) {
	p.abstractRequest = *req
}

// Project.Unsubscribe
type projectUnsubscribeRequest struct {
	ProjectID int64
	abstractRequest
}

func (p projectUnsubscribeRequest) process(db dbfs.DBFS) ([]dhClosure, error) {
	unsubscribeClosure := rabbitChannelUnsubscribeClosure{
		key: rabbitmq.RabbitProjectQueueName(p.ProjectID),
		tag: p.Tag,
	}
	return []dhClosure{unsubscribeClosure}, nil
}

func (p *projectUnsubscribeRequest) setAbstractRequest(req *abstractRequest) {
	p.abstractRequest = *req
}

// Project.Delete
type projectDeleteRequest struct {
	ProjectID int64
	abstractRequest
}

func (p projectDeleteRequest) process(db dbfs.DBFS) ([]dhClosure, error) {
	err := db.MySQLProjectDelete(p.ProjectID, p.SenderID)
	if err != nil {
		if err == dbfs.ErrNoDbChange {
			return []dhClosure{toSenderClosure{msg: newEmptyResponse(fail, p.Tag)}}, err
		}
		return []dhClosure{toSenderClosure{msg: newEmptyResponse(servfail, p.Tag)}}, err

	}

	res := newEmptyResponse(success, p.Tag)
	not := notification{
		Resource:   p.Resource,
		Method:     p.Method,
		ResourceID: p.ProjectID,
		Data:       struct{}{},
	}.wrap()

	return []dhClosure{toSenderClosure{msg: res}, toRabbitChannelClosure{msg: not, projectID: p.ProjectID}}, nil
}

func (p *projectDeleteRequest) setAbstractRequest(req *abstractRequest) {
	p.abstractRequest = *req
}
