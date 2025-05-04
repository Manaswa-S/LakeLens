package public

import (
	"lakelens/internal/services/public"
)


type PublicHandler struct {
	Public *public.PublicService
}

func NewPublicHandler(public *public.PublicService) *PublicHandler {
	return &PublicHandler{
		Public: public,
	}
}


