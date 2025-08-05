package handlers

import (
	"context"
	"crypto/md5"
	"database/sql"
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

func (h *Handler) Shortener(ctx context.Context, origURL string, userID int) (string, error) {
	shortURL := fmt.Sprintf("%x", md5.Sum([]byte(origURL)))[:8]
	if _, err := h.storage.GetURL(ctx, shortURL); err == nil {
		return "", fmt.Errorf("conflict")
	}
	err := h.storage.SetURL(ctx, shortURL, origURL, userID)
	return fmt.Sprintf("%s/%s", h.config.Host, shortURL), err
}

func (h *Handler) GetHandler(ctx context.Context, idGet string) (string, error) {
	origURL, err := h.storage.GetURL(ctx, idGet)
	return origURL, err
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
func (h *Handler) DeleteUrls(urlIDs []string) {
	for _, id := range urlIDs {
		h.deleteCh <- id
	}
}

func (h *Handler) GetStats(ctx context.Context, ipFromRequest string) (int, int, error) {
	if h.config.TrustedSubnet == "" {
		return 0, 0, fmt.Errorf("forbidden")
	}
	ipSubnet := h.config.TrustedSubnet

	ip, err := netip.ParseAddr(ipFromRequest)
	if err != nil {
		return 0, 0, err
	}
	network, err := netip.ParsePrefix(ipSubnet)
	if err != nil {
		return 0, 0, err
	}
	if ok := network.Contains(ip); !ok {
		return 0, 0, fmt.Errorf("forbidden")
	}

	countURLs, err := h.storage.GetCountURLs(ctx)
	if err != nil {
		return 0, 0, err
	}
	return countURLs, jwtgen.UsersID, nil
}
