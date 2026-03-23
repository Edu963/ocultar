package gateway

import "errors"

type Refiner struct { Endpoint string; APIKey string }

func (r *Refiner) Refine(payload string) (string, error) {
    if r.Endpoint == "" { return "", errors.New("refinery not configured") }
    return payload, nil // Placeholder for testing
}
