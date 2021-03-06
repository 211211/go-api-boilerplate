package socialmedia

import (
	"net/http"

	"github.com/vardius/go-api-boilerplate/pkg/common/application/errors"
	"github.com/vardius/go-api-boilerplate/pkg/common/application/http/response"
	"github.com/vardius/go-api-boilerplate/pkg/common/application/jwt"
	"github.com/vardius/go-api-boilerplate/pkg/common/application/security/identity"
	user_proto "github.com/vardius/go-api-boilerplate/pkg/user/infrastructure/proto"
	user_grpc "github.com/vardius/go-api-boilerplate/pkg/user/interfaces/grpc"
)

type facebook struct {
	client user_proto.UserClient
	jwt    jwt.Jwt
}

func (f *facebook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	accessToken := r.FormValue("accessToken")
	data, e := getProfile(accessToken, "https://graph.facebook.com/me")
	if e != nil {
		response.WithError(r.Context(), errors.Wrap(e, "Invalid access token", errors.INVALID))
		return
	}

	identity := &identity.Identity{}
	identity.FromFacebookData(data)

	token, e := f.jwt.Encode(identity)
	if e != nil {
		response.WithError(r.Context(), errors.Wrap(e, "Generate token failure", errors.INTERNAL))
		return
	}

	payload := &commandPayload{token, data}
	_, e = f.client.DispatchCommand(r.Context(), &user_proto.DispatchCommandRequest{
		Name:    user_grpc.RegisterUserWithFacebook,
		Payload: payload.toJSON(),
	})

	if e != nil {
		response.WithError(r.Context(), errors.Wrap(e, "Invalid request", errors.INVALID))
		return
	}

	response.WithPayload(r.Context(), &responsePayload{token, identity})
	return
}

// NewFacebook creates facebook auth handler
func NewFacebook(c user_proto.UserClient, j jwt.Jwt) http.Handler {
	return &facebook{c, j}
}
