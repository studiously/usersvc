package classsvc

import (
	"context"
	"net/url"
	"strings"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/google/uuid"
	"github.com/studiously/classsvc/models"
	"github.com/studiously/introspector"
)

// Endpoints collects all of the endpoints that compose a class service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
//
// In a server, it's useful for functions that need to operate on a per-endpoint
// basis. For example, you might pass an Endpoints to a function that produces
// an http.Handler, with each method (endpoint) wired up to a specific path. (It
// is probably a mistake in design to invoke the Service methods on the
// Endpoints struct in a server.)
//
// In a client, it's useful to collect individually constructed endpoints into a
// single type that implements the Service interface. For example, you might
// construct individual endpoints using transport/http.NewClient, combine them
// into an Endpoints, and return it to the caller as a Service.
type Endpoints struct {
	ListClassesEndpoint endpoint.Endpoint
	GetClassEndpoint    endpoint.Endpoint
	CreateClassEndpoint endpoint.Endpoint
	UpdateClassEndpoint endpoint.Endpoint
	DeleteClassEndpoint endpoint.Endpoint
	JoinClassEndpoint   endpoint.Endpoint
	SetRoleEndpoint     endpoint.Endpoint
	LeaveClassEndpoint  endpoint.Endpoint
	ListMembersEndpoint endpoint.Endpoint
	GetMemberEndpoint   endpoint.Endpoint
}

func MakeServerEndpoints(s Service) Endpoints {
	return Endpoints{
		ListClassesEndpoint: MakeListClassesEndpoint(s),
		GetClassEndpoint:    MakeGetClassEndpoint(s),
		CreateClassEndpoint: MakeCreateClassEndpoint(s),
		UpdateClassEndpoint: MakeUpdateClassEndpoint(s),
		DeleteClassEndpoint: MakeDeleteClassEndpoint(s),
		JoinClassEndpoint:   MakeJoinClassEndpoint(s),
		SetRoleEndpoint:     MakeSetRoleEndpoint(s),
		LeaveClassEndpoint:  MakeLeaveClassEndpoint(s),
		ListMembersEndpoint: MakeListMembersEndpoint(s),
		GetMemberEndpoint:   MakeGetMemberEndpoint(s),
	}
}

// MakeClientEndpoints returns an Endpoints struct where each endpoint invokes
// the corresponding method on the remote instance, via a transport/http.Client.
// Useful in a classsvc client.
func MakeClientEndpoints(instance string) (Endpoints, error) {
	if !strings.HasPrefix(instance, "http") {
		instance = "http://" + instance
	}
	tgt, err := url.Parse(instance)
	if err != nil {
		return Endpoints{}, err
	}
	tgt.Path = ""

	options := []httptransport.ClientOption{
		httptransport.ClientBefore(introspector.FromHTTPContext()),
	}

	return Endpoints{
		ListClassesEndpoint: httptransport.NewClient("GET", tgt, EncodeListClassesRequest, DecodeListClassesResponse, options...).Endpoint(),
		GetClassEndpoint:    httptransport.NewClient("GET", tgt, EncodeGetClassRequest, DecodeGetClassResponse, options...).Endpoint(),
		CreateClassEndpoint: httptransport.NewClient("POST", tgt, EncodeCreateClassRequest, DecodeCreateClassResponse, options...).Endpoint(),
		UpdateClassEndpoint: httptransport.NewClient("PATCH", tgt, EncodeUpdateClassRequest, DecodeUpdateClassResponse, options...).Endpoint(),
		DeleteClassEndpoint: httptransport.NewClient("DELETE", tgt, EncodeDeleteClassRequest, DecodeDeleteClassResponse, options...).Endpoint(),
		JoinClassEndpoint:   httptransport.NewClient("POST", tgt, EncodeJoinClassRequest, DecodeJoinClassResponse, options...).Endpoint(),
		SetRoleEndpoint:     httptransport.NewClient("PATCH", tgt, EncodeSetRoleRequest, DecodeSetRoleResponse, options...).Endpoint(),
		LeaveClassEndpoint:  httptransport.NewClient("DELETE", tgt, EncodeLeaveClassRequest, DecodeLeaveClassResponse, options...).Endpoint(),
		ListMembersEndpoint: httptransport.NewClient("GET", tgt, EncodeListMembersRequest, DecodeListMembersResponse, options...).Endpoint(),
		GetMemberEndpoint:   httptransport.NewClient("GET", tgt, EncodeGetMemberRequest, DecodeGetMemberResponse, options...).Endpoint(),
	}, nil
}

func (e Endpoints) ListClasses(ctx context.Context) ([]uuid.UUID, error) {
	response, err := e.ListClassesEndpoint(ctx, nil)
	if err != nil {
		return nil, err
	}
	resp := response.(listClassesResponse)
	return resp.Classes, resp.Error
}

func (e Endpoints) GetClass(ctx context.Context, classID uuid.UUID) (*models.Class, error) {
	request := getClassRequest{ClassID: classID}
	response, err := e.GetClassEndpoint(ctx, request)
	if err != nil {
		return nil, err
	}
	resp := response.(getClassResponse)
	return resp.Class, resp.Error
}

func (e Endpoints) CreateClass(ctx context.Context, name string) (*uuid.UUID, error) {
	request := createClassRequest{Name: name}
	response, err := e.CreateClassEndpoint(ctx, request)
	if err != nil {
		return nil, err
	}
	resp := response.(createClassResponse)
	return resp.ClassID, resp.Error
}

func (e Endpoints) UpdateClass(ctx context.Context, classID uuid.UUID, name *string, currentUnit *uuid.UUID) error {
	request := updateClassRequest{ClassID: classID, Name: name, CurrentUnit: currentUnit}
	response, err := e.UpdateClassEndpoint(ctx, request)
	if err != nil {
		return err
	}
	resp := response.(updateClassResponse)
	return resp.Error
}

func (e Endpoints) DeleteClass(ctx context.Context, classID uuid.UUID) error {
	request := deleteClassRequest{ClassID: classID}
	response, err := e.DeleteClassEndpoint(ctx, request)
	if err != nil {
		return err
	}
	resp := response.(deleteClassResponse)
	return resp.Error
}

func (e Endpoints) JoinClass(ctx context.Context, classID uuid.UUID) error {
	request := joinClassRequest{ClassID: classID}
	response, err := e.JoinClassEndpoint(ctx, request)
	if err != nil {
		return err
	}
	resp := response.(joinClassResponse)
	return resp.Error
}

func (e Endpoints) SetRole(ctx context.Context, classID, userID uuid.UUID, role models.UserRole) error {
	request := setRoleRequest{UserID: userID, ClassID: classID, Role: role}
	response, err := e.SetRoleEndpoint(ctx, request)
	if err != nil {
		return err
	}
	resp := response.(setRoleResponse)
	return resp.Error
}

func (e Endpoints) LeaveClass(ctx context.Context, userID *uuid.UUID, classID uuid.UUID) error {
	request := leaveClassRequest{ClassID: classID, UserID: userID}
	response, err := e.LeaveClassEndpoint(ctx, request)
	if err != nil {
		return err
	}
	resp := response.(leaveClassResponse)
	return resp.Error
}

func (e Endpoints) ListMembers(ctx context.Context, classID uuid.UUID) ([]*models.Member, error) {
	request := listMembersRequest{ClassID: classID}
	response, err := e.ListMembersEndpoint(ctx, request)
	if err != nil {
		return nil, err
	}
	resp := response.(listMembersResponse)
	return resp.Members, resp.Error
}

//func (e Endpoints) GetRole(ctx context.Context, userID, classID uuid.UUID) (*models.UserRole, error) {
//	request := getRoleRequest{UserID: userID, ClassID: classID}
//	response, err := e.GetRoleEndpoint(ctx, request)
//	if err != nil {
//		return nil, err
//	}
//	resp := response.(getRoleResponse)
//	return resp.Role, resp.Error
//}
//
//func (e Endpoints) IsOwner(ctx context.Context, userID, classID uuid.UUID) (bool, error) {
//	request := isOwnerRequest{UserID: userID, ClassID: classID}
//	response, err := e.IsOwnerEndpoint(ctx, request)
//	if err != nil {
//		return nil, err
//	}
//	resp := response.(isOwnerResponse)
//	return resp.IsOwner, resp.Error
//}

func (e Endpoints) GetMember(ctx context.Context, classID, userID uuid.UUID) (*models.Member, error) {
	request := getMemberRequest{UserID: userID, ClassID: classID}
	response, err := e.GetMemberEndpoint(ctx, request)
	if err != nil {
		return nil, err
	}
	resp := response.(getMemberResponse)
	return resp.Member, resp.Error
}

func MakeListClassesEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		classes, e := s.ListClasses(ctx)
		return listClassesResponse{classes, e}, nil
	}
}

type listClassesResponse struct {
	Classes []uuid.UUID
	Error   error `json:"error,omitempty"`
}

func (r listClassesResponse) error() error {
	return r.Error
}

func MakeGetClassEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getClassRequest)
		class, e := s.GetClass(ctx, req.ClassID)
		return getClassResponse{class, e}, nil
	}
}

type getClassRequest struct {
	ClassID uuid.UUID `json:"class_id,omitempty"`
}

type getClassResponse struct {
	Class *models.Class `json:"class,omitempty"`
	Error error `json:"error,omitempty"`
}

func (r getClassResponse) error() error {
	return r.Error
}

func MakeCreateClassEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(createClassRequest)
		id, e := s.CreateClass(ctx, req.Name)
		return createClassResponse{id, e}, nil
	}
}

type createClassRequest struct {
	Name string `json:"name"`
}

type createClassResponse struct {
	ClassID *uuid.UUID `json:"class_id,omitempty"`
	Error   error `json:"error,omitempty"`
}

func (r createClassResponse) error() error {
	return r.Error
}

func MakeUpdateClassEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(updateClassRequest)
		e := s.UpdateClass(ctx, req.ClassID, req.Name, req.CurrentUnit)
		return updateClassResponse{e}, nil
	}
}

type updateClassRequest struct {
	ClassID     uuid.UUID
	Name        *string `json:"class,omitempty"`
	CurrentUnit *uuid.UUID `json:"current_unit,omitempty"`
}

type updateClassResponse struct {
	Error error `json:"error,omitempty"`
}

func (r updateClassResponse) error() error {
	return r.Error
}

func MakeDeleteClassEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(deleteClassRequest)
		e := s.DeleteClass(ctx, req.ClassID)
		return deleteClassResponse{e}, nil
	}
}

type deleteClassRequest struct {
	ClassID uuid.UUID `json:"class_id"`
}

type deleteClassResponse struct {
	Error error `json:"error,omitempty"`
}

func (r deleteClassResponse) error() error {
	return r.Error
}

func MakeJoinClassEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(joinClassRequest)
		e := s.JoinClass(ctx, req.ClassID)
		return joinClassResponse{e}, nil
	}
}

type joinClassRequest struct {
	ClassID uuid.UUID `json:"class"`
}

type joinClassResponse struct {
	Error error `json:"error,omitempty"`
}

func (r joinClassResponse) error() error {
	return r.Error
}

func MakeLeaveClassEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(leaveClassRequest)
		e := s.LeaveClass(ctx, req.UserID, req.ClassID)
		return leaveClassResponse{e}, nil
	}
}

type leaveClassRequest struct {
	UserID  *uuid.UUID `json:"user,omitempty"`
	ClassID uuid.UUID `json:"class"`
}

type leaveClassResponse struct {
	Error error `json:"error,omitempty"`
}

func (r leaveClassResponse) error() error {
	return r.Error
}

func MakeSetRoleEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(setRoleRequest)
		e := s.SetRole(ctx, req.UserID, req.ClassID, req.Role)
		return setRoleResponse{e}, nil
	}
}

type setRoleRequest struct {
	UserID  uuid.UUID `json:"user"`
	ClassID uuid.UUID `json:"class"`
	Role    models.UserRole `json:"role"`
}

type setRoleResponse struct {
	Error error `json:"error,omitempty"`
}

func (r setRoleResponse) error() error {
	return r.Error
}

func MakeListMembersEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(listMembersRequest)
		members, e := s.ListMembers(ctx, req.ClassID)
		return listMembersResponse{members, e}, nil
	}
}

type listMembersRequest struct {
	ClassID uuid.UUID `json:"id"`
}

type listMembersResponse struct {
	Members []*models.Member `json:"members"`
	Error   error `json:"error,omitempty"`
}

func (r listMembersResponse) error() error {
	return r.Error
}

func MakeGetMemberEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getMemberRequest)
		member, e := s.GetMember(ctx, req.ClassID, req.UserID)
		return getMemberResponse{member, e}, nil
	}
}

type getMemberRequest struct {
	UserID  uuid.UUID
	ClassID uuid.UUID
}

type getMemberResponse struct {
	Member *models.Member `json:"member,omitempty"`
	Error  error `json:"error,omitempty"`
}

//func MakeGetRoleEndpoint(s Service) endpoint.Endpoint {
//	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
//		req := request.(getRoleRequest)
//		role, e := s.GetRole(ctx, req.UserID, req.ClassID)
//		return getRoleResponse{role, e}, nil
//	}
//}
//
//type getRoleRequest struct {
//	UserID  uuid.UUID
//	ClassID uuid.UUID
//}
//
//type getRoleResponse struct {
//	Role  *models.UserRole `json:"role,omitempty"`
//	Error error `json:"error,omitempty"`
//}
//
//func (r getRoleResponse) error() error {
//	return r.Error
//}
//
//func MakeIsOwnerEndpoint(s Service) endpoint.Endpoint {
//	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
//		req := request.(isOwnerRequest)
//		isOwner, e := s.IsOwner(ctx, req.UserID, req.ClassID)
//		return isOwnerResponse{isOwner, e}, nil
//	}
//}
//
//type isOwnerRequest struct {
//	UserID  uuid.UUID
//	ClassID uuid.UUID
//}
//
//type isOwnerResponse struct {
//	IsOwner bool `json:"role,omitempty"`
//	Error   error `json:"error,omitempty"`
//}
//
//func (r isOwnerResponse) error() error {
//	return r.Error
//}

//func MakeGetMemberEndpoint(s Service) endpoint.Endpoint {
//	return func(ctx context.Context, request interface{}) (interface{}, error) {
//		req := request.(getMemberRequest)
//		member, e := s.GetMember(ctx, req.ClassID, req.UserID)
//		return getMemberResponse{member, e}, nil
//	}
//}
//
//type getMemberRequest struct {
//	ClassID uuid.UUID `json:"class_id"`
//	UserID  uuid.UUID `json:"user_id"`
//}
//
//type getMemberResponse struct {
//	Member *models.Member `json:"member,omitempty"`
//	Error  error `json:"error,omitempty"`
//}
//
//func (r getMemberResponse) error() error {
//	return r.Error
//}
