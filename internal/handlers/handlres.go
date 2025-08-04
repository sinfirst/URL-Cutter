package handlers

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/netip"

	"github.com/sinfirst/URL-Cutter/internal/config"
	"github.com/sinfirst/URL-Cutter/internal/middleware/jwtgen"
	"github.com/sinfirst/URL-Cutter/internal/models"
)

// Storage интерфейс для взаимодействия с хранилищем данных
type Storage interface {
	SetURL(ctx context.Context, key, value string, userID int) error
	GetURL(ctx context.Context, key string) (string, error)
	GetByUserID(ctx context.Context, userID int) ([]models.ShortenOrigURLs, error)
	GetCountURLs(ctx context.Context) (int, error)
}

// Cтруктура для бизнес логики
type Handler struct {
	storage  Storage
	config   config.Config
	deleteCh chan string
}

func NewHandler(storage Storage, config config.Config, deleteCh chan string) Handler {
	handler := Handler{storage: storage, config: config, deleteCh: deleteCh}
	return handler
}
func (h *Handler) BatchShortenURL(ctx context.Context, requests []models.ShortenRequestForBatch) ([]models.ShortenResponceForBatch, error) {
	var responces []models.ShortenResponceForBatch
	for _, req := range requests {
		shortURL := fmt.Sprintf("%x", md5.Sum([]byte(req.OriginalURL)))[:8]
		responces = append(responces, models.ShortenResponceForBatch{
			CorrelationID: req.CorrelationID,
			ShortURL:      h.config.Host + "/" + shortURL,
		})
		err := h.storage.SetURL(ctx, shortURL, req.OriginalURL, 0)
		if err != nil {
			return nil, err
		}
	}
	return responces, nil
}

func (h *Handler) GetHandler(ctx context.Context, idGet string) (string, error) {
	origURL, err := h.storage.GetURL(ctx, idGet)
	return origURL, err
}

func (h *Handler) PostHandler(ctx context.Context, body []byte, UserID int) (string, int, error) {
	shortURL := fmt.Sprintf("%x", md5.Sum(body))[:8]
	if _, err := h.storage.GetURL(ctx, shortURL); err == nil {
		return fmt.Sprintf("%s/%s", h.config.Host, shortURL), 1, err
	}
	err := h.storage.SetURL(ctx, shortURL, string(body), UserID)
	if err != nil {
		return "", 2, err
	}
	return fmt.Sprintf("%s/%s", h.config.Host, shortURL), 0, nil
}

func (h *Handler) JSONPostHandler(ctx context.Context, input models.OriginalURL) ([]byte, int, error) {
	var output models.ResultURL

	shortURL := fmt.Sprintf("%x", md5.Sum([]byte(input.URL)))[:8]
	output = models.ResultURL{Result: h.config.Host + "/" + shortURL}
	JSONResponse, err := json.Marshal(output)
	if err != nil {
		return nil, 2, err
	}

	if _, err := h.storage.GetURL(ctx, shortURL); err == nil {
		return nil, 1, nil
	}

	err = h.storage.SetURL(ctx, shortURL, string(input.URL), 0)
	if err != nil {
		return nil, 3, err
	}

	return JSONResponse, 0, nil
}

func (h *Handler) DBPing() error {
	db, err := sql.Open("pgx", h.config.DatabaseDsn)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Ping()

	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) GetUserUrls(ctx context.Context, userID int) ([]models.ShortenOrigURLs, error) {

	urlsFromDB, err := h.storage.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	url := models.ShortenOrigURLs{OriginalURL: urlsFromDB[len(urlsFromDB)-1].OriginalURL, ShortURL: h.config.Host + "/" + urlsFromDB[len(urlsFromDB)-1].ShortURL}
	return []models.ShortenOrigURLs{url}, nil

}

// DeleteUrls удаляет запрошенные урлы
func (h *Handler) DeleteUrls(body []byte) error {
	var urlIDs []string

	err := json.Unmarshal(body, &urlIDs)
	if err != nil {
		return err
	}
	for _, id := range urlIDs {
		h.deleteCh <- id
	}
	return nil
}

func (h *Handler) GetStats(ctx context.Context, ipFromRequest string) (models.ServerStats, int, error) {
	if h.config.TrustedSubnet == "" {
		return models.ServerStats{}, 1, nil
	}
	ipSubnet := h.config.TrustedSubnet

	ip, err := netip.ParseAddr(ipFromRequest)
	if err != nil {
		return models.ServerStats{}, 2, err
	}
	network, err := netip.ParsePrefix(ipSubnet)
	if err != nil {
		return models.ServerStats{}, 2, err
	}
	if ok := network.Contains(ip); !ok {
		return models.ServerStats{}, 1, nil
	}

	countURLs, err := h.storage.GetCountURLs(ctx)
	if err != nil {
		return models.ServerStats{}, 2, err
	}
	users := jwtgen.UsersID
	resp := models.ServerStats{URLs: countURLs, Users: users}
	return resp, 0, nil
}
