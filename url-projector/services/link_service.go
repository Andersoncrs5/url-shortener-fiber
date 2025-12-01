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

func (l *linkService) ApplyLogic(ctx context.Context, envelope cdc.Envelope) {
	op := envelope.Payload.Op

	switch op {
	case "c", "u", "r":
		l.Upsert(ctx, envelope)

	case "d":
		if idFloat, ok := envelope.Payload.Before["id"].(int64); ok {
			l.Delete(ctx, idFloat)
		} else {
			log.Printf("ERROR: Delete operation received, but 'id' not found in 'Before' payload or is not numeric. Op: %s", op)
		}

	default:
		log.Printf("INFO: No action taken! Op received: %s", op)
	}
}

func (l *linkService) Upsert(ctx context.Context, envelope cdc.Envelope) {
	link, err_after := cdc.GetLinkFromAfter(envelope)

	if err_after != nil {
		log.Fatalf("Error the get after of envelope! Details error: %v", err_after)
	}

	if _, err := l.repo.Upsert(ctx, &link); err != nil {
		log.Printf("ERROR: Failed to upsert document [ID: %d, Short Code: %s]: %v", link.ID, link.SHORT_CODE, err)
		return
	}

	log.Printf("SUCCESS: Document Upserted (Op: %s) [ID: %d, Short Code: %s]", envelope.Payload.Op, link.ID, link.SHORT_CODE)
}

func (l *linkService) Delete(ctx context.Context, id int64) {
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
