package mcpoauth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// FileStore is a durable store for a single-host deployment. All BFF processes
// must use the same local POSIX directory and encryption key. Do not use NFS or
// independent replica volumes: flock is deliberately not a distributed lease.
// Lock files are permanent; unlinking them could create two locks for one grant.
type FileStore struct {
	dir   string
	codec *Codec
}

func NewFileStore(dir string, codec *Codec) (*FileStore, error) {
	if !filepath.IsAbs(dir) || codec == nil {
		return nil, ErrInvalidConfig
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode().Perm()&0077 != 0 {
		return nil, errors.New("OAuth store directory must be private (0700)")
	}
	return &FileStore{dir, codec}, nil
}
func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func (s *FileStore) locked(ctx context.Context, key string, fn func() error) error {
	f, err := os.OpenFile(filepath.Join(s.dir, key+".lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	for {
		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			break
		}
		if err != syscall.EWOULDBLOCK && err != syscall.EAGAIN {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	if err = ctx.Err(); err != nil {
		return err
	}
	return fn()
}
func (s *FileStore) write(key string, data []byte) error {
	f, err := os.CreateTemp(s.dir, ".oauth-")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(name, filepath.Join(s.dir, key)); err != nil {
		return err
	}
	d, err := os.Open(s.dir)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}
func (s *FileStore) PutPending(ctx context.Context, state string, p Pending) error {
	key := "pending-" + digest([]byte(state))
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	sealed, err := s.codec.sealBytes(p.Owner, state, b)
	if err != nil {
		return err
	}
	return s.locked(ctx, "pending", func() error {
		entries, e := os.ReadDir(s.dir)
		if e != nil {
			return e
		}
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), "pending-") && !strings.HasSuffix(entry.Name(), ".lock") {
				info, e := entry.Info()
				if e != nil {
					return e
				}
				if time.Since(info.ModTime()) > 10*time.Minute {
					if e = os.Remove(filepath.Join(s.dir, entry.Name())); e != nil {
						return e
					}
				}
			}
		}
		return s.write(key, sealed)
	})
}
func (s *FileStore) Take(ctx context.Context, state string, owner Owner) (p Pending, err error) {
	key := "pending-" + digest([]byte(state))
	err = s.locked(ctx, "pending", func() error {
		b, e := os.ReadFile(filepath.Join(s.dir, key))
		if e != nil {
			return ErrInvalidState
		}
		plain, e := s.codec.openBytes(owner, state, b)
		if e != nil {
			return ErrInvalidState
		}
		if json.Unmarshal(plain, &p) != nil || p.Owner != owner {
			return ErrInvalidState
		}
		if e = os.Remove(filepath.Join(s.dir, key)); e != nil {
			return e
		}
		return nil
	})
	return
}
func (s *FileStore) Update(ctx context.Context, owner Owner, resource string, fn func(*Grant) (*Grant, error)) error {
	key := "grant-" + digest(binding(owner, resource))
	return s.locked(ctx, key, func() error {
		var old *Grant
		b, err := os.ReadFile(filepath.Join(s.dir, key))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if err == nil {
			old, err = s.codec.Open(owner, resource, b)
			if err != nil {
				return err
			}
		}
		next, err := fn(old)
		if err != nil {
			return err
		}
		if next == nil {
			return errors.New("OAuth store requires a tombstone instead of deleting a grant")
		}
		sealed, err := s.codec.Seal(owner, resource, next)
		if err != nil {
			return err
		}
		return s.write(key, sealed)
	})
}
func (s *FileStore) Delete(ctx context.Context, o Owner, r string) error {
	return s.Update(ctx, o, r, func(*Grant) (*Grant, error) {
		generation, e := randomString()
		return &Grant{Generation: generation, Epoch: generation}, e
	})
}
