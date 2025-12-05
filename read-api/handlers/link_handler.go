package handlers

import (
	"errors"
	"linkfast/read-api/dtos"
	"linkfast/read-api/services"
	"linkfast/read-api/utils/consts"
	"linkfast/read-api/utils/res"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
)

type LinkHandler interface {
	GetByID(c *fiber.Ctx) error
	GetByShotCode(c *fiber.Ctx) error
}

type linkHandler struct {
	service services.LinkService
}

func NewLinkHandler(service services.LinkService) LinkHandler {
	return &linkHandler{service: service}
}

func (h *linkHandler) GetByID(c *fiber.Ctx) error {
	traceID, ok := c.Locals("trace_id").(string)
	if !ok {
		traceID = "unknown_trace"
	}

	idParam := c.Params("id")
	if idParam == "" {
		res := res.ResponseHttp[string]{
			Timestamp: time.Now(),
			Payload:   "",
			Code:      fiber.StatusBadRequest,
			Status:    false,
			Message:   "Id is required",
			Version:   1,
			TraceID:   traceID,
			Path:      "",
		}

		return c.Status(fiber.StatusBadRequest).JSON(res)
	}

	id, err := c.ParamsInt("id", 0)
	if err != nil || id <= 0 {
		res := res.ResponseHttp[string]{
			Timestamp: time.Now(),
			Payload:   "",
			Code:      fiber.StatusBadRequest,
			Status:    false,
			Message:   "Id is required",
			Version:   1,
			TraceID:   traceID,
			Path:      "",
		}

		return c.Status(fiber.StatusBadRequest).JSON(res)
	}

	link, err := h.service.GetById(c.Context(), int64(id))

	if err != nil {
		if errors.Is(err, consts.ErrRecordNotFound) {
			res := res.ResponseHttp[string]{
				Timestamp: time.Now(),
				Payload:   "",
				Code:      fiber.StatusNotFound,
				Status:    false,
				Message:   "Link not found",
				Version:   1,
				TraceID:   traceID,
				Path:      "",
			}

			return c.Status(fiber.StatusNotFound).JSON(res)
		}

		res := res.ResponseHttp[string]{
			Timestamp: time.Now(),
			Payload:   err.Error(),
			Code:      fiber.StatusInternalServerError,
			Status:    false,
			Message:   "Error internal in server! Try again later",
			Version:   1,
			TraceID:   traceID,
			Path:      "",
		}

		return c.Status(fiber.StatusInternalServerError).JSON(res)
	}

	var linkDTO dtos.LinkDto
	if err := copier.Copy(&linkDTO, &link); err != nil {
		res := res.ResponseHttp[string]{
			Timestamp: time.Now(),
			Payload:   err.Error(),
			Code:      fiber.StatusInternalServerError,
			Status:    false,
			Message:   "Error internal in server! Try again later",
			Version:   1,
			TraceID:   traceID,
			Path:      "",
		}

		return c.Status(fiber.StatusInternalServerError).JSON(res)
	}

	res := res.ResponseHttp[dtos.LinkDto]{
		Timestamp: time.Now(),
		Payload:   linkDTO,
		Code:      fiber.StatusOK,
		Status:    false,
		Message:   "Link found",
		Version:   1,
		TraceID:   traceID,
		Path:      "",
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *linkHandler) GetByShotCode(c *fiber.Ctx) error {
	traceID, ok := c.Locals("trace_id").(string)
	if !ok {
		traceID = "unknown_trace"
	}

	shortCode := c.Params("code")
	if shortCode == "" {
		res := res.ResponseHttp[string]{
			Timestamp: time.Now(),
			Payload:   "",
			Code:      fiber.StatusBadRequest,
			Status:    false,
			Message:   "Code is required",
			Version:   1,
			TraceID:   traceID,
			Path:      "",
		}

		return c.Status(fiber.StatusBadRequest).JSON(res)
	}

	link, err := h.service.GetByCode(c.Context(), shortCode)

	if err != nil {
		if errors.Is(err, consts.ErrRecordNotFound) {
			res := res.ResponseHttp[string]{
				Timestamp: time.Now(),
				Payload:   "",
				Code:      fiber.StatusNotFound,
				Status:    false,
				Message:   "Link not found",
				Version:   1,
				TraceID:   traceID,
				Path:      "",
			}

			return c.Status(fiber.StatusNotFound).JSON(res)
		}

		res := res.ResponseHttp[string]{
			Timestamp: time.Now(),
			Payload:   err.Error(),
			Code:      fiber.StatusInternalServerError,
			Status:    false,
			Message:   "Error internal in server! Try again later",
			Version:   1,
			TraceID:   traceID,
			Path:      "",
		}

		return c.Status(fiber.StatusInternalServerError).JSON(res)
	}

	c.Set(fiber.HeaderCacheControl, "no-store, no-cache, must-revalidate, max-age=0")
	return c.Redirect(link.LONG_URL, fiber.StatusTemporaryRedirect)
}
