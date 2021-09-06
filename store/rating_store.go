package store

import (
	"sync"
)

//RatingStore rates the laptop
type RatingStore interface {
	Add(laptopID string, score float64) (*Rating, error)
}

//Rating contains the rating information
type Rating struct {
	Count uint32
	Sum   float64
}

//InMemoryRatingScore stores the laptop ratingsin memory
type InMemoryRatingScore struct {
	mutex  sync.RWMutex
	rating map[string]*Rating
}

//NewInMemoryRatingStore returns a new InMemoryRatingStore
func NewInMemoryRatingStore() *InMemoryRatingScore {
	return &InMemoryRatingScore{
		rating: make(map[string]*Rating),
	}
}

//Add adds a new laptop score to the store and returns the rating
func (store *InMemoryRatingScore) Add(laptopID string, score float64) (*Rating, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	rating := store.rating[laptopID]
	if rating == nil {
		rating = &Rating{
			Count: 1,
			Sum:   score,
		}
	} else {
		rating.Count++
		rating.Sum += score
	}

	store.rating[laptopID] = rating
	return rating, nil
}
