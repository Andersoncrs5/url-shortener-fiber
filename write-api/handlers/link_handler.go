package handlers

import (
	"errors"
	"linkfast/write-api/dtos"
	"linkfast/write-api/services"
	"linkfast/write-api/utils/consts"
	"linkfast/write-api/utils/res"
	"log"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
)

var validater = validator.New()

type LinkHandler interface {
	Create(c *fiber.Ctx) error
	GetByID(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
	GetByShotCode(c *fiber.Ctx) error
}

type linkHandler struct {
	service services.LinkService
}

func NewTaskHandler(service services.LinkService) LinkHandler {
	return &linkHandler{service: service}
}

func (h *linkHandler) GetByID(c *fiber.Ctx) error {
	var dto dtos.LinkDto

	traceID, ok := c.Locals("trace_id").(string)
	if !ok {
		traceID = "unknown_trace"
	}

	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response := res.ResponseHttp[string]{
			Timestamp: time.Now(),
			Payload:   err.Error(),
			Code:      fiber.StatusBadRequest,
			Status:    false,
			Message:   "Id is required",
			Version:   1,
			TraceID:   traceID,
			Path:      "",
		}

		return c.Status(fiber.StatusBadRequest).JSON(response)
	}

	link, err_get := h.service.GetByID(id)
	if err_get != nil {
		response := res.ResponseHttp[string]{
			Timestamp: time.Now(),
			Payload:   err_get.Error(),
			Code:      0,
			Status:    false,
			Message:   err_get.Error(),
			Version:   1,
			TraceID:   traceID,
			Path:      "",
		}

		if errors.Is(err_get, consts.ErrRecordNotFound) {
			response.Code = fiber.StatusNotFound
		}

		if errors.Is(err_get, consts.ErrInternal) {
			response.Code = fiber.StatusInternalServerError
		}

		return c.Status(response.Code).JSON(response)
	}

	err_copy := copier.Copy(&dto, link)
	if err_copy != nil {
		log.Printf("Error the copy of Links to LinkDto: %v", err_copy)
		response := res.ResponseHttp[string]{
			Timestamp: time.Now(),
			Payload:   "",
			Code:      fiber.StatusInternalServerError,
			Status:    false,
			Message:   "Error internal in server! Try again later",
			Version:   1,
			TraceID:   traceID,
			Path:      "",
		}

		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}

	response := res.ResponseHttp[dtos.LinkDto]{
		Timestamp: time.Now(),
		Payload:   dto,
		Code:      fiber.StatusOK,
		Status:    true,
		Message:   "Link found",
		Version:   1,
		TraceID:   traceID,
		Path:      "",
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func (h *linkHandler) GetByShotCode(c *fiber.Ctx) error {
	var dto dtos.LinkDto
	traceID, ok := c.Locals("trace_id").(string)
	if !ok {
		traceID = "unknown_trace"
	}

	code := c.Params("code")

	link, err := h.service.GetByShotCode(code)

	if err != nil {
		response := res.ResponseHttp[string]{
			Timestamp: time.Now(),
			Payload:   err.Error(),
			Code:      0,
			Status:    false,
			Message:   err.Error(),
			Version:   1,
			TraceID:   traceID,
			Path:      "",
		}

		if errors.Is(err, consts.ErrInternal) {
			response.Code = fiber.StatusInternalServerError
		}

		if errors.Is(err, consts.ErrRecordNotFound) {
			response.Code = fiber.StatusNotFound
		}

		return c.Status(response.Code).JSON(response)
	}

	err_parse := copier.Copy(&dto, link)
	if err_parse != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			res.ResponseHttp[string]{
				Timestamp: time.Now(),
				Payload:   err_parse.Error(),
				Code:      fiber.StatusInternalServerError,
				Status:    false,
				Message:   "Error internal in server",
				Version:   1,
				TraceID:   traceID,
				Path:      "",
			},
		)
	}

	response := res.ResponseHttp[dtos.LinkDto]{
		Timestamp: time.Now(),
		Payload:   dto,
		Code:      200,
		Status:    true,
		Message:   "Link found",
		Version:   1,
		TraceID:   traceID,
		Path:      "",
	}

	return c.Status(200).JSON(response)
}

func (h *linkHandler) Create(c *fiber.Ctx) error {
	traceID, ok := c.Locals("trace_id").(string)
	if !ok {
		traceID = "unknown_trace"
	}

	var req dtos.CreateLinkDto
	dto := new(dtos.LinkDto)

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			res.ResponseHttp[string]{
				Timestamp: time.Now(),
				Payload:   err.Error(),
				Code:      fiber.StatusBadRequest,
				Status:    false,
				Message:   "Inputs invalids",
				Version:   1,
				TraceID:   traceID,
				Path:      "",
			},
		)
	}

	if err := validater.Struct(req); err != nil {
		errors := []string{}
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, err.Field()+" failed on "+err.Tag())
		}

		return c.Status(fiber.StatusBadRequest).JSON(
			res.ResponseHttp[[]string]{
				Timestamp: time.Now(),
				Payload:   errors,
				Code:      fiber.StatusBadRequest,
				Status:    false,
				Message:   "Inputs invalids",
				Version:   1,
				TraceID:   traceID,
				Path:      "",
			},
		)
	}

	body, err_create := h.service.Create(req)
	if err_create != nil {
		res := res.ResponseHttp[string]{
			Timestamp: time.Now(),
			Payload:   err_create.Error(),
			Code:      1,
			Status:    false,
			Message:   err_create.Error(),
			Version:   1,
			TraceID:   traceID,
			Path:      "",
		}

		if errors.Is(err_create, consts.ErrInternal) {
			res.Code = fiber.StatusInternalServerError
		}

		return c.Status(1).JSON(res)
	}

	err_parse := copier.Copy(dto, body)
	if err_parse != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			res.ResponseHttp[string]{
				Timestamp: time.Now(),
				Payload:   err_parse.Error(),
				Code:      fiber.StatusInternalServerError,
				Status:    false,
				Message:   "Error internal in server",
				Version:   1,
				TraceID:   traceID,
				Path:      "",
			},
		)
	}

	response := res.ResponseHttp[dtos.LinkDto]{
		Timestamp: time.Now(),
		Payload:   *dto,
		Code:      fiber.StatusCreated,
		Status:    true,
		Message:   "Link created",
		Version:   1,
		TraceID:   traceID,
		Path:      "",
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

func (h *linkHandler) Delete(c *fiber.Ctx) error {
	traceID, ok := c.Locals("trace_id").(string)
	if !ok {
		traceID = "unknown_trace"
	}

	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response := res.ResponseHttp[string]{
			Timestamp: time.Now(),
			Payload:   err.Error(),
			Code:      fiber.StatusBadRequest,
			Status:    false,
			Message:   "Id is required",
			Version:   1,
			TraceID:   traceID,
			Path:      "",
		}

		return c.Status(fiber.StatusBadRequest).JSON(response)
	}

	link, err_get := h.service.GetByID(id)
	if err_get != nil {
		response := res.ResponseHttp[string]{
			Timestamp: time.Now(),
			Payload:   err_get.Error(),
			Code:      0,
			Status:    false,
			Message:   err_get.Error(),
			Version:   1,
			TraceID:   traceID,
			Path:      "",
		}

		if errors.Is(err_get, consts.ErrRecordNotFound) {
			response.Code = fiber.StatusNotFound
		}

		return c.Status(response.Code).JSON(response)
	}

	err_delete := h.service.Delete(&link)
	if err_delete != nil {
		response := res.ResponseHttp[string]{
			Timestamp: time.Now(),
			Payload:   err_get.Error(),
			Code:      fiber.StatusInternalServerError,
			Status:    false,
			Message:   err_get.Error(),
			Version:   1,
			TraceID:   traceID,
			Path:      "",
		}

		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}

	response := res.ResponseHttp[string]{
		Timestamp: time.Now(),
		Payload:   "",
		Code:      fiber.StatusOK,
		Status:    false,
		Message:   "Link deleted",
		Version:   1,
		TraceID:   traceID,
		Path:      "",
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
