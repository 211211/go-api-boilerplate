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

type google struct {
	client user_proto.UserClient
	jwt    jwt.Jwt
}

func (g *google) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	accessToken := r.FormValue("accessToken")
	data, e := getProfile(accessToken, "https://www.googleapis.com/oauth2/v2/userinfo")
	if e != nil {
		response.WithError(r.Context(), errors.Wrap(e, "Invalid access token", errors.INVALID))
		return
	}

	identity := &identity.Identity{}
	identity.FromGoogleData(data)

	token, e := g.jwt.Encode(identity)
	if e != nil {
		response.WithError(r.Context(), errors.Wrap(e, "Generate token failure", errors.INTERNAL))
		return
	}

	payload := &commandPayload{token, data}
	_, e = g.client.DispatchCommand(r.Context(), &user_proto.DispatchCommandRequest{
		Name:    user_grpc.RegisterUserWithGoogle,
		Payload: payload.toJSON(),
	})

	if e != nil {
		response.WithError(r.Context(), errors.Wrap(e, "Invalid request", errors.INVALID))
		return
	}

	response.WithPayload(r.Context(), &responsePayload{token, identity})
	return
}

// NewGoogle creates google auth handler
func NewGoogle(c user_proto.UserClient, j jwt.Jwt) http.Handler {
	return &google{c, j}
}
