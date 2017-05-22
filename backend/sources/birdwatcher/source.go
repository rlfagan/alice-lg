package birdwatcher

import (
	"github.com/ecix/alice-lg/backend/api"
)

type Birdwatcher struct {
	config Config
	client *Client
}

func NewBirdwatcher(config Config) *Birdwatcher {
	client := NewClient(config.Api)

	birdwatcher := &Birdwatcher{
		config: config,
		client: client,
	}
	return birdwatcher
}

func (self *Birdwatcher) Status() (api.StatusResponse, error) {
	bird, err := self.client.GetJson("/status")
	if err != nil {
		return api.StatusResponse{}, err
	}

	apiStatus, err := parseApiStatus(bird, self.config)
	if err != nil {
		return api.StatusResponse{}, err
	}

	birdStatus := bird["status"].(map[string]interface{})

	// Get special fields
	serverTime, _ := parseServerTime(
		birdStatus["current_server"],
		SERVER_TIME_SHORT,
		self.config.Timezone,
	)

	lastReboot, _ := parseServerTime(
		birdStatus["last_reboot"],
		SERVER_TIME_SHORT,
		self.config.Timezone,
	)

	lastReconfig, _ := parseServerTime(
		birdStatus["last_reconfig"],
		SERVER_TIME_EXT,
		self.config.Timezone,
	)

	// Make status response
	status := api.Status{
		ServerTime:   serverTime,
		LastReboot:   lastReboot,
		LastReconfig: lastReconfig,
		Backend:      "bird",
		Version:      birdStatus["version"].(string),
		Message:      birdStatus["message"].(string),
		RouterId:     birdStatus["router_id"].(string),
	}

	response := api.StatusResponse{
		Api:    apiStatus,
		Status: status,
	}

	return response, nil
}

// Get bird BGP protocols
func (self *Birdwatcher) Neighbours() (api.NeighboursResponse, error) {
	bird, err := self.client.GetJson("/protocols/bgp")
	if err != nil {
		return api.NeighboursResponse{}, err
	}

	apiStatus, err := parseApiStatus(bird, self.config)
	if err != nil {
		return api.NeighboursResponse{}, err
	}

	neighbours, err := parseNeighbours(bird, self.config)
	if err != nil {
		return api.NeighboursResponse{}, err
	}

	return api.NeighboursResponse{
		Api:        apiStatus,
		Neighbours: neighbours,
	}, nil
}

// Get filtered and exported routes
func (self *Birdwatcher) Routes(neighbourId string) (api.RoutesResponse, error) {
	// Exported
	bird, err := self.client.GetJson("/routes/protocol/" + neighbourId)
	if err != nil {
		return api.RoutesResponse{}, err
	}

	// Use api status from first request
	apiStatus, err := parseApiStatus(bird, self.config)
	if err != nil {
		return api.RoutesResponse{}, err
	}

	exported, err := parseRoutes(bird, self.config)
	if err != nil {
		return api.RoutesResponse{}, err
	}

	// Filtered
	bird, err = self.client.GetJson("/routes/filtered/" + neighbourId)
	if err != nil {
		return api.RoutesResponse{}, err
	}

	filtered, err := parseRoutes(bird, self.config)
	if err != nil {
		return api.RoutesResponse{}, err
	}

	return api.RoutesResponse{
		Api:      apiStatus,
		Exported: exported,
		Filtered: filtered,
	}, nil
}
