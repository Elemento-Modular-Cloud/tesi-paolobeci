package schema

// -------- STORAGE --------
// Elemento storage representation
type StorageVolume struct {
	VolumeID  string   `json:"volumeID"`
	CreatorID string   `json:"creator_id"`
	Size      int      `json:"size"` // Bytes
	Name      string   `json:"name"`
	Format	  string   `json:"format"`
	Private   bool     `json:"private"`
	Bootable  bool     `json:"bootable"`
	Readonly  bool     `json:"readonly"`
	Shareable bool     `json:"shareable"`
	Clonable  bool     `json:"clonable"`
	Alg       string   `json:"alg"`
	Bus       string   `json:"bus"`
	Cloudinit bool     `json:"cloudinit"`
	Ceph      bool     `json:"ceph"`
	Exported  bool     `json:"exported"`
	SizeOnDisk int     `json:"sizeOnDisk"` // Bytes
	LastUpdated string `json:"lastUpdated"`
	ServerUrl string   `json:"serverurl"`
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

type CreateStorageResponse struct{
	Name		string	`json:"name"`
	Private 	bool	`json:"private"`
	Clonable	bool 	`json:"clonable"`
	Alg			string	`json:"alg"`
	Format		string	`json:"img_format"`
	Bus			string	`json:"bus"`
	Size		int		`json:"size"`
	Shareable	bool	`json:"shareable"`
	Readonly	bool	`json:"readonly"`
	Bootable	bool	`json:"bootable"`
	VolumeID	string	`json:"vid"`
	CreatorID	string	`json:"creatorID"`
}

// -------- CRETATE STORAGE WITH IMAGE --------
type CreateStorageImageRequest struct {
	Name		string	`json:"name"`
	Private 	bool	`json:"private"`
	Clonable	bool 	`json:"clonable"`
	Alg			string	`json:"alg"`
	Format		string	`json:"format"`
	Bus			string	`json:"bus"`
	Size		int		`json:"size"`
	Url			string	`json:"url"`
}

type CreateStorageImageResponse struct{
	Name		string	`json:"name"`
	Private 	bool	`json:"private"`
	Clonable	bool 	`json:"clonable"`
	Alg			string	`json:"alg"`
	Format		string	`json:"img_format"`
	Bus			string	`json:"bus"`
	Size		int		`json:"size"`
	Url			string	`json:"url"`
	VolumeID	string	`json:"vid"`		
}

// -------- CREATE CLOUDINIT VOLUME -------
type CreateStorageCloudInitRequest struct {
	Name			string	`json:"name"`
	Private 		bool	`json:"private"`
	Bootable		bool	`json:"bootable"`
	Clonable		bool 	`json:"clonable"`
	Alg				string	`json:"alg"`
	ExpectedFiles	int		`json:"expectedFiles"`
}

type CreateStorageCloudInitResponse struct {
	VolumeID		string	`json:"vid"`
}

type FeedFileIntoCloudInitStorageRequest struct {
	VolumeID		string	`json:"vid"`
}

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
