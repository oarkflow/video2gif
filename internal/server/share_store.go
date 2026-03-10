package server

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/oarkflow/video2gif/internal/config"
)

type ShareStore struct {
	dir      string
	mu       sync.RWMutex
	sessions map[string]*ShareSession
}

type shareSessionDisk struct {
	ID         string               `json:"id"`
	FileName   string               `json:"file_name"`
	VideoFile  string               `json:"video_file"`
	CutRanges  []config.ClipSegment `json:"cut_ranges"`
	Comments   []ShareComment       `json:"comments"`
	CreatedAt  time.Time            `json:"created_at"`
	ExpiresAt  time.Time            `json:"expires_at"`
	CreatedBy  string               `json:"created_by,omitempty"`
	PublicView bool                 `json:"public_view"`
}

func NewShareStore(dir string) (*ShareStore, error) {
	if dir == "" {
		return nil, errors.New("share dir is required")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	store := &ShareStore{
		dir:      dir,
		sessions: make(map[string]*ShareSession),
	}
	if err := store.Load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *ShareStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions = make(map[string]*ShareSession)
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.dir, entry.Name()))
		if err != nil {
			continue
		}
		var disk shareSessionDisk
		if err := json.Unmarshal(data, &disk); err != nil {
			continue
		}
		videoPath := filepath.Join(s.dir, filepath.Base(disk.VideoFile))
		if disk.ID == "" || disk.VideoFile == "" {
			continue
		}
		if _, err := os.Stat(videoPath); err != nil {
			continue
		}
		s.sessions[disk.ID] = &ShareSession{
			ID:         disk.ID,
			FileName:   disk.FileName,
			VideoPath:  videoPath,
			CutRanges:  append([]config.ClipSegment(nil), disk.CutRanges...),
			Comments:   append([]ShareComment(nil), disk.Comments...),
			CreatedAt:  disk.CreatedAt,
			ExpiresAt:  disk.ExpiresAt,
			CreatedBy:  disk.CreatedBy,
			PublicView: disk.PublicView,
		}
	}
	return nil
}

func (s *ShareStore) Save(session *ShareSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked(session)
}

func (s *ShareStore) Get(id string, now time.Time) (*ShareSession, bool, error) {
	s.mu.RLock()
	session, ok := s.sessions[id]
	if !ok {
		s.mu.RUnlock()
		return nil, false, nil
	}
	clone := cloneShareSession(session)
	s.mu.RUnlock()
	if clone.ExpiresAt.IsZero() || now.Before(clone.ExpiresAt) {
		return clone, true, nil
	}
	if err := s.Delete(id); err != nil {
		return nil, false, err
	}
	return nil, false, nil
}

func (s *ShareStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.deleteLocked(id)
}

func (s *ShareStore) CleanupExpired(now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, session := range s.sessions {
		if !session.ExpiresAt.IsZero() && !now.Before(session.ExpiresAt) {
			if err := s.deleteLocked(id); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *ShareStore) saveLocked(session *ShareSession) error {
	if session == nil {
		return errors.New("session is required")
	}
	videoFile := filepath.Base(session.VideoPath)
	disk := shareSessionDisk{
		ID:         session.ID,
		FileName:   session.FileName,
		VideoFile:  videoFile,
		CutRanges:  append([]config.ClipSegment(nil), session.CutRanges...),
		Comments:   append([]ShareComment(nil), session.Comments...),
		CreatedAt:  session.CreatedAt,
		ExpiresAt:  session.ExpiresAt,
		CreatedBy:  session.CreatedBy,
		PublicView: session.PublicView,
	}
	data, err := json.MarshalIndent(disk, "", "  ")
	if err != nil {
		return err
	}
	metaPath := filepath.Join(s.dir, session.ID+".json")
	tmpPath := metaPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, metaPath); err != nil {
		return err
	}
	s.sessions[session.ID] = cloneShareSession(session)
	return nil
}

func (s *ShareStore) deleteLocked(id string) error {
	session, ok := s.sessions[id]
	if !ok {
		return nil
	}
	delete(s.sessions, id)
	if strings.TrimSpace(session.VideoPath) != "" {
		if err := os.Remove(session.VideoPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	metaPath := filepath.Join(s.dir, id+".json")
	if err := os.Remove(metaPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func cloneShareSession(session *ShareSession) *ShareSession {
	if session == nil {
		return nil
	}
	clone := *session
	clone.CutRanges = append([]config.ClipSegment(nil), session.CutRanges...)
	clone.Comments = append([]ShareComment(nil), session.Comments...)
	return &clone
}
