package delegate

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

func validHash(s string) bool { b, err := hex.DecodeString(s); return err == nil && len(b) == 32 }
func LoadState(path string, chainID uint64) (*State, error) {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return NewState(chainID), nil
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	dec := json.NewDecoder(io.LimitReader(f, 16<<20))
	dec.DisallowUnknownFields()
	var s State
	if err = dec.Decode(&s); err != nil {
		return nil, fmt.Errorf("state is unreadable (preserved): %w", err)
	}
	var extra any
	if dec.Decode(&extra) != io.EOF {
		return nil, errors.New("invalid trailing state data")
	}
	if s.Version != 1 || s.ChainID != chainID || s.Pillars == nil {
		return nil, errors.New("state schema or chainId mismatch")
	}
	for _, entries := range s.Pillars {
		for id, e := range entries {
			if e != nil && e.CopyObservation != nil {
				o := e.CopyObservation
				if o.Follow == "" || !o.Vote.Valid() || o.ObservedAt.IsZero() {
					return nil, errors.New("invalid copy observation")
				}
			}
			if !validHash(id) || e == nil || (e.Cast != nil && !e.Cast.Valid()) || (e.ManualVote != nil && !e.ManualVote.Valid()) || (e.ManualConfirmed && e.ManualVote == nil) || e.Failures < 0 || e.Failures > 32 || (e.BlockHash != "" && !validHash(e.BlockHash)) || (e.Pending != nil && (!e.Pending.Vote.Valid() || !validHash(e.Pending.BlockHash) || e.Pending.At.IsZero())) {
				return nil, errors.New("invalid state entry")
			}
		}
	}
	return &s, nil
}

func SaveState(path string, s *State) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".az-state-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(f.Name()) }()
	if _, err = f.Write(append(b, '\n')); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(f.Name(), path); err != nil {
		return err
	}
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer func() { _ = d.Close() }()
	return d.Sync()
}

// The stable lock file is deliberately never unlinked: unlinking a held lock
// would allow another process to lock a different inode at the same path.
func LockState(path string) (func(), error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	fd, err := unix.Open(path+".lock", unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0600)
	if err != nil {
		return nil, err
	}
	if err = unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = unix.Close(fd)
		return nil, errors.New("state is locked by another process")
	}
	return func() { _ = unix.Flock(fd, unix.LOCK_UN); _ = unix.Close(fd) }, nil
}
