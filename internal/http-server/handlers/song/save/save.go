package savehandler

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"effective_mobile/internal/lib/api/response"
	"effective_mobile/internal/lib/logger/sl"
	"effective_mobile/internal/service"
	"effective_mobile/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Group string `json:"group" validate:"required"`
	Song  string `json:"song" validate:"required"`
}

type Response struct {
	response.Response
	ID int `json:"id,omitempty"`
}

type SongDetail struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

type SongSaver interface {
	SaveSong(group, song string) (int, error)
}

func New(log *slog.Logger, songSaver SongSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.song.save.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			response.Error(w, r, http.StatusBadRequest, "empty request")

			return
		}

		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			response.Error(w, r, http.StatusInternalServerError, "failed to decode request")

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			response.Error(w, r, http.StatusBadRequest, response.ValidationErrors(validateErr))

			return
		}

		id, err := songSaver.SaveSong(req.Group, req.Song)
		if errors.Is(err, storage.ErrSongExists) {
			log.Info("song already exists", slog.String("song", fmt.Sprintf("%s - %s", req.Group, req.Song)))

			response.Error(w, r, http.StatusBadRequest, "song already exists")

			return
		}
		if errors.Is(err, service.ErrInvalidDateFormat) {
			log.Info("externalAPI response invalid date format", slog.String("song", fmt.Sprintf("%s - %s", req.Group, req.Song)))

			response.Error(w, r, http.StatusInternalServerError, "externalAPI server error")

			return
		}
		if err != nil {
			log.Error("failed to add song", sl.Err(err))

			response.Error(w, r, http.StatusInternalServerError, "failed to add song")

			return
		}

		log.Info("song added", slog.Int("id", id))

		responseOK(w, r, id)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, id int) {
	render.JSON(w, r, Response{
		Response: response.OK(),
		ID:       id,
	})
}
