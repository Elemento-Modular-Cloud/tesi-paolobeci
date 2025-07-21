package ecloud

import (
	"context"
	"errors"
	"fmt"

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
	Url       string
}

// Create creates a new volume.
func (c *VolumeClient) Create(ctx context.Context, opts VolumeCreateOpts) (string, *Response, error) {
	if err := opts.Validate(); err != nil {
		return "", nil, err
	}

	// Prepare the can create request
	reqBodyCanCreate := schema.CanCreateStorageRequest{
		Size: opts.Size,
	}

	// First check if we can create the storage volume
	_, err := c.client.CanCreateStorage(reqBodyCanCreate)
	if err != nil {
		return "", nil, fmt.Errorf("the config provided cannot be created: %w", err)
	}

	if opts.Url == "" {
		reqBody := schema.CreateStorageRequest{
			Name:      opts.Name,
			Size:      opts.Size,
			Bootable:  opts.Bootable,
			Readonly:  opts.Readonly,
			Shareable: opts.Shareable,
			Private:   opts.Private,
		}

		// Create the storage volume
		createdVolume, err := c.client.CreateStorage(reqBody)
		if err != nil {
			return "", nil, fmt.Errorf("failed to create storage volume: %w", err)
		}
		return createdVolume.VolumeID, &Response{}, nil

	} else {
		reqBody := schema.CreateStorageImageRequest{
			Name:     opts.Name,
			Size:     opts.Size,
			Alg:      "cp",
			Format:   "qcow2",
			Bus:      "virtio",
			Clonable: true,
			Private:  false,
			Url:      opts.Url,
		}

		// Create the boot volume
		createdVolume, err := c.client.CreateStorageImage(reqBody)
		if err != nil {
			return "", nil, fmt.Errorf("failed to create storage volume: %w", err)
		}

		fmt.Printf("CreateStorageImage response: %+v\n", createdVolume)
		return createdVolume.VolumeID, &Response{}, nil
	}
}

// CloudInitCreateOpts specifies options for creating a new cloud-init.
type CloudInitCreateOpts struct {
	Name      string
}

func (c *VolumeClient) CreateCloudInit(ctx context.Context, opts CloudInitCreateOpts) (string, *Response, error) {
	reqBody := schema.CreateStorageCloudInitRequest{
		Name: opts.Name,
		Private: false,
		Bootable: true,
		Clonable: false,
		Alg: "no",
		ExpectedFiles: 2,  // Minimum number of files accepted are 2
	}
	createdVolume, err := c.client.CreateStorageCloudInit(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create cloud-init volume: %w", err)
	}

	return createdVolume.VolumeID, &Response{}, nil
}

func (c *VolumeClient) FeedFileIntoCloudInitStorage(ctx context.Context, volumeID string) (string, *Response, error) {
	reqBody := schema.FeedFileIntoCloudInitStorageRequest{
		VolumeID: volumeID,
	}
	response, err := c.client.FeedFileIntoCloudInitStorage(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create cloud-init volume: %w", err)
	}

	return response, &Response{}, nil
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
