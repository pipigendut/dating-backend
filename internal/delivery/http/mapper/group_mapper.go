package mapper

import (
	"github.com/pipigendut/dating-backend/internal/delivery/http/dto/v1"
	"github.com/pipigendut/dating-backend/internal/entities"
)

func ToGroupResponse(g *entities.Group, storage StorageURLProvider) dtov1.GroupResponse {
	resp := dtov1.GroupResponse{
		ID:         g.ID,
		EntityID:   g.EntityID,
		Name:       g.Name,
		CreatedBy:  g.CreatedBy,
		Members:    make([]dtov1.UserResponse, 0),
		MainPhotos: make([]string, 0),
	}

	for _, m := range g.Members {
		if m.User != nil {
			userResp := ToUserResponse(m.User, storage)
			resp.Members = append(resp.Members, userResp)
			if userResp.MainPhoto != "" {
				resp.MainPhotos = append(resp.MainPhotos, userResp.MainPhoto)
			}
		}
	}

	return resp
}
