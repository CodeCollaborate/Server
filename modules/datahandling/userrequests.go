package datahandling

import (
	"fmt"

	"time"

	"github.com/CodeCollaborate/Server/modules/dbfs"
)

var userRequestsSetup = false

// initProjectRequests populates the requestMap from requestmap.go with the appropriate constructors for the project methods
func initUserRequests() {
	if userRequestsSetup {
		return
	}

	unauthenticatedRequestMap["User.Register"] = func(req *abstractRequest) (request, error) {
		return commonJSON(new(userRegisterRequest), req)
	}

	unauthenticatedRequestMap["User.Login"] = func(req *abstractRequest) (request, error) {
		return commonJSON(new(userLoginRequest), req)
	}

	authenticatedRequestMap["User.Lookup"] = func(req *abstractRequest) (request, error) {
		return commonJSON(new(userLookupRequest), req)
	}

	authenticatedRequestMap["User.Projects"] = func(req *abstractRequest) (request, error) {
		return commonJSON(new(userProjectsRequest), req)
	}

	userRequestsSetup = true
}

// User.Register
type userRegisterRequest struct {
	Username  string
	FirstName string
	LastName  string
	Email     string
	Password  string
	abstractRequest
}

func (f *userRegisterRequest) setAbstractRequest(req *abstractRequest) {
	f.abstractRequest = *req
}

func (f userRegisterRequest) process() (*serverMessageWrapper, *serverMessageWrapper, error) {

	newUser := dbfs.UserMeta{
		Username:  f.Username,
		FirstName: f.FirstName,
		LastName:  f.LastName,
		Email:     f.Email,
		Password:  f.Password}

	err := dbfs.MySQLUserRegister(newUser)

	res := new(serverMessageWrapper)
	res.Timestamp = time.Now()
	res.Type = "Responce"
	if err != nil {
		if err == dbfs.ErrNoDbChange {
			res.ServerMessage = response{Status: conflict, Tag: f.Tag}
		} else {
			res.ServerMessage = response{Status: fail, Tag: f.Tag}
		}
	} else {
		res.ServerMessage = response{Status: success, Tag: f.Tag}
	}
	return res, nil, err
}

// User.Login
type userLoginRequest struct {
	Username string
	Password string
	abstractRequest
}

func (f *userLoginRequest) setAbstractRequest(req *abstractRequest) {
	f.abstractRequest = *req
}

func (f userLoginRequest) process() (*serverMessageWrapper, *serverMessageWrapper, error) {
	// TODO implement login logic
	fmt.Printf("Recieved login request from %s\n", f.Username)
	return nil, nil, nil
}

// User.Lookup
type userLookupRequest struct {
	Usernames []string
	abstractRequest
}

func (f *userLookupRequest) setAbstractRequest(req *abstractRequest) {
	f.abstractRequest = *req
}

func (f userLookupRequest) process() (*serverMessageWrapper, *serverMessageWrapper, error) {
	users := make([]dbfs.UserMeta, len(f.Usernames))
	index := 0
	var erro error
	for _, username := range f.Usernames {
		usr, err := dbfs.MySQLUserLookup(username)
		if err != nil {
			erro = err
		} else {
			users[index] = usr
			index++
		}
	}
	// shrink as needed
	users = users[:index+1]

	res := new(serverMessageWrapper)
	res.Timestamp = time.Now()
	res.Type = "Responce"

	if len(users) < 0 {
		res.ServerMessage = response{Status: fail, Tag: f.Tag}
	} else {
		if erro != nil {
			// at least 1 value failed
			// return what we can but
			// tell the client whatever they don't get back failed
			res.ServerMessage = response{
				Status: servfail,
				Tag:    f.Tag,
				Data:   {"Users": users}}
		} else {
			res.ServerMessage = response{
				Status: success,
				Tag:    f.Tag,
				Data:   {"Users": users}}
		}
	}
	return res, nil, erro
}

// User.Projects
type userProjectsRequest struct {
	abstractRequest
}

func (f *userProjectsRequest) setAbstractRequest(req *abstractRequest) {
	f.abstractRequest = *req
}

func (f userProjectsRequest) process() (*serverMessageWrapper, *serverMessageWrapper, error) {
	projects, err := dbfs.MySQLUserProjects(f.SenderID)

	res := new(serverMessageWrapper)
	res.Timestamp = time.Now()
	res.Type = "Responce"

	if err != nil {
		res.ServerMessage = response{
			Status: servfail,
			Tag:    f.Tag,
			Data:   {"Projects": projects}}
	} else {
		res.ServerMessage = response{
			Status: success,
			Tag:    f.Tag,
			Data:   {"Projects": projects}}
	}

	return res, nil, err
}
