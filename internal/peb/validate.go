package peb

import (
	"errors"
	"fmt"
	"strings"
)

var ErrCycle = errors.New("cycle detected in blocked-by relationships")
var ErrInvalidReference = errors.New("referenced peb(s) not found")

func IsInvalidReference(err error) bool {
	return err != nil && err.Error() == ErrInvalidReference.Error()
}

type Store interface {
	Get(id string) (*Peb, bool)
}

func ValidateBlockedBy(store Store, peb *Peb, blockedBy []string) error {
	for _, id := range blockedBy {
		if _, ok := store.Get(id); !ok {
			return fmt.Errorf("%w: %s", ErrInvalidReference, id)
		}
	}
	return nil
}

func HasInvalidReference(err error) bool {
	return err != nil && strings.Contains(err.Error(), ErrInvalidReference.Error())
}

func CheckCycle(store Store, pebID string, blockedBy []string) error {
	visited := make(map[string]bool)
	return checkCycleDFS(store, pebID, blockedBy, visited)
}

func checkCycleDFS(store Store, currentID string, blockedBy []string, visited map[string]bool) error {
	visited[currentID] = true

	for _, blockingID := range blockedBy {
		if blockingID == currentID {
			return ErrCycle
		}

		if visited[blockingID] {
			continue
		}

		blockingPeb, ok := store.Get(blockingID)
		if !ok {
			continue
		}

		if err := checkCycleDFS(store, currentID, blockingPeb.BlockedBy, visited); err != nil {
			return err
		}
	}

	visited[currentID] = false
	return nil
}
