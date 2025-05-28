package schema

// -------- STORAGE --------
type StorageVolume struct {
	VolumeID  string   `json:"volumeID"`
	Name      string   `json:"name"`
	CreatorID string   `json:"creator_id"`
	Private   bool     `json:"private"`
	Bootable  bool     `json:"bootable"`
	Readonly  bool     `json:"readonly"`
	Shareable bool     `json:"shareable"`
	Size      int      `json:"size"` // Bytes
	Serverurl string   `json:"serverurl"`
	Server    string   `json:"server"`
	Own       bool     `json:"own"`
	Nservers  int      `json:"nservers"`
	Servers   []string `json:"servers"`
}

// -------- HEALTH CHECK --------
type HealthCheckStorageResponse struct {
	Status string `json:"status"`
}

// -------- CAN CREATE STORAGE --------
type CanCreateStorageRequest struct {
	Size int `json:"size"` // GB
}

type CanCreateStorageResponse int

// -------- CREATE STORAGE --------
type CreateStorageRequest struct {
	Name      string `json:"name"`
	Size      int    `json:"size"` // GB
	Bootable  bool   `json:"bootable"`
	Readonly  bool   `json:"readonly"`
	Shareable bool   `json:"shareable"`
	Private   bool   `json:"private"`
}

type CreateStorageResponse struct{}

// -------- GET STORAGE --------
type GetStorageResponse []StorageVolume

// -------- GET STORAGE BY ID --------
type GetStorageByIDRequest struct {
	VolumeID string `json:"volume_id"`
}

type GetStorageByIDResponse struct {
	Volume StorageVolume `json:"volume"`
}

// -------- DELETE STORAGE --------
type DeleteStorageRequest struct {
	VolumeID string `json:"volume_id"`
}

type DeleteStorageResponse struct{}
