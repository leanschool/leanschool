package storage

import "errors"

// ErrOptimisticLock is returned when an UPDATE finds a version mismatch,
// indicating that another request modified the resource concurrently.
// Callers should translate this to HTTP 409 Conflict.
var ErrOptimisticLock = errors.New("optimistic lock conflict: resource was modified by another request")
