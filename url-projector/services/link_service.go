package services

import (
	"context"
	"errors"
	"linkfast/url-projector/cdc"
	"linkfast/url-projector/repositories"
	"linkfast/url-projector/utils/consts"
	"log"
)

type LinkService interface {
	ApplyLogic(ctx context.Context, envelope cdc.Envelope)
}

type linkService struct {
	repo repositories.LinkRepository
}

func NewLinkService(repo repositories.LinkRepository) LinkService {
	return &linkService{
		repo: repo,
	}
}

// ApplyLogic: Recebe o contexto.
func (l *linkService) ApplyLogic(ctx context.Context, envelope cdc.Envelope) {
	op := envelope.Payload.Op

	switch op {
	case "c", "u", "r":
		// Chamada atualizada para passar o contexto
		l.Upsert(ctx, envelope)

	case "d":
		if idFloat, ok := envelope.Payload.Before["id"].(float64); ok {
			// Chamada atualizada para passar o contexto
			l.Delete(ctx, int64(idFloat))
		} else {
			log.Printf("ERROR: Delete operation received, but 'id' not found in 'Before' payload or is not numeric. Op: %s", op)
		}

	default:
		log.Printf("INFO: No action taken! Op received: %s", op)
	}
}

// Upsert: Recebe o contexto.
func (l *linkService) Upsert(ctx context.Context, envelope cdc.Envelope) {
	link := cdc.GetLinkFromAfter(envelope)

	// Chamada de repositório atualizada para passar o contexto
	if _, err := l.repo.Upsert(ctx, &link); err != nil {
		log.Printf("ERROR: Failed to upsert document [ID: %d, Short Code: %s]: %v", link.ID, link.SHORT_CODE, err)
		return
	}

	log.Printf("SUCCESS: Document Upserted (Op: %s) [ID: %d, Short Code: %s]", envelope.Payload.Op, link.ID, link.SHORT_CODE)
}

// Delete: Recebe o contexto.
func (l *linkService) Delete(ctx context.Context, id int64) {
	// Chamada de repositório atualizada para passar o contexto
	err := l.repo.Delete(ctx, id)

	if err != nil {
		if errors.Is(err, consts.ErrRecordNotFound) {
			log.Printf("INFO: Document not found [ID: %d]. Already deleted.", id)
			return
		}

		log.Printf("ERROR: Failed to delete document [ID: %d]: %v", id, err)
		return
	}

	log.Printf("SUCCESS: Document deleted [ID: %d]", id)
}
