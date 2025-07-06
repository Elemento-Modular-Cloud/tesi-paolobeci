package ecloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Elemento-Modular-Cloud/tesi-paolobeci/ecloud/schema"
)

// VolumeStatus specifies a volume's status.
type VolumeStatus string

const (
	// VolumeStatusCreating is the status when a volume is being created.
	VolumeStatusCreating VolumeStatus = "creating"

	// VolumeStatusAvailable is the status when a volume is available.
	VolumeStatusAvailable VolumeStatus = "available"
)

// VolumeClient is a client for the volumes API.
type VolumeClient struct {
	client *Client
}

// GetByID retrieves a volume by its ID. If the volume does not exist, nil is returned.
func (c *VolumeClient) GetByID(ctx context.Context, id string) (*schema.StorageVolume, error) {
	reqBody := schema.GetStorageByIDRequest{
		VolumeID: id,
	}
	resp, err := c.client.GetStorageByID(reqBody)
	if err != nil {
		return nil, err
	}

	// Return the volume from the response
	return &resp.Volume, nil
}

// VolumeCreateOpts specifies options for creating a new volume.
type VolumeCreateOpts struct {
	Name      string
	Size      int // Size in GB
	Bootable  bool
	Readonly  bool
	Shareable bool
	Private   bool
	Labels    map[string]string
}

// Create creates a new volume.
func (c *VolumeClient) Create(ctx context.Context, opts VolumeCreateOpts) (*schema.StorageVolume, *Response, error) {
	if err := opts.Validate(); err != nil {
		return nil, nil, err
	}

	// Prepare the can create request
	reqBodyCanCreate := schema.CanCreateStorageRequest{
		Size: opts.Size,
	}

	// First check if we can create the storage volume
	_, err := c.client.CanCreateStorage(reqBodyCanCreate)
	if err != nil {
		return nil, nil, fmt.Errorf("the config provided cannot be created: %w", err)
	}

	// Prepare the create request
	reqBody := schema.CreateStorageRequest{
		Name:      opts.Name,
		Size:      opts.Size,
		Bootable:  opts.Bootable,
		Readonly:  opts.Readonly,
		Shareable: opts.Shareable,
		Private:   opts.Private,
	}

	// Create the storage volume
	_, err = c.client.CreateStorage(reqBody)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create storage volume: %w", err)
	}

	// Get the created volume with retry logic
	var createdVolume *schema.StorageVolume
	maxRetries := 10
	retryDelay := time.Millisecond * 500

	for i := 0; i < maxRetries; i++ {
		volumes, err := c.client.GetStorage()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to retrieve created volume: %w", err)
		}

		// Search for the volume with the matching unique name
		for _, vol := range *volumes {
			if vol.Name == reqBody.Name {
				createdVolume = &vol
				break
			}
		}

		if createdVolume != nil {
			break
		}

		// Wait before retrying
		time.Sleep(retryDelay)
	}

	if createdVolume == nil {
		return nil, nil, fmt.Errorf("created volume with name '%s' not found in storage list after %d retries", reqBody.Name, maxRetries)
	}

	return createdVolume, &Response{}, nil
}

// Validate checks if options are valid.
func (o VolumeCreateOpts) Validate() error {
	if o.Name == "" {
		return errors.New("missing name")
	}
	if o.Size <= 0 {
		return errors.New("size must be greater than 0")
	}
	return nil
}

// List returns a list of volumes.
func (c *VolumeClient) List(ctx context.Context) ([]*schema.StorageVolume, *Response, error) {
	body, err := c.client.GetStorage()
	if err != nil {
		return nil, nil, err
	}

	volumes := make([]*schema.StorageVolume, 0, len(*body))
	for _, vol := range *body {
		volumes = append(volumes, &vol)
	}
	return volumes, &Response{}, nil
}
