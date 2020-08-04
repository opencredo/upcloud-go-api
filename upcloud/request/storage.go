package request

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
)

// GetStoragesRequest represents a request for retrieving all or some storages
type GetStoragesRequest struct {
	// If specified, only storages with this access type will be retrieved
	Access string
	// If specified, only storages with this type will be retrieved
	Type string
	// If specified, only storages marked as favorite will be retrieved
	Favorite bool
}

// RequestURL implements the Request interface
func (r *GetStoragesRequest) RequestURL() string {
	if r.Access != "" {
		return fmt.Sprintf("/storage/%s", r.Access)
	}

	if r.Type != "" {
		return fmt.Sprintf("/storage/%s", r.Type)
	}

	if r.Favorite {
		return "/storage/favorite"
	}

	return "/storage"
}

// GetStorageDetailsRequest represents a request for retrieving details about a piece of storage
type GetStorageDetailsRequest struct {
	UUID string
}

// RequestURL implements the Request interface
func (r *GetStorageDetailsRequest) RequestURL() string {
	return fmt.Sprintf("/storage/%s", r.UUID)
}

// CreateStorageRequest represents a request to create a storage device
type CreateStorageRequest struct {
	XMLName xml.Name `xml:"storage" json:"-"`

	Size       int                 `xml:"size" json:"size,string"`
	Tier       string              `xml:"tier,omitempty" json:"tier,omitempty"`
	Title      string              `xml:"title" json:"title,omitempty"`
	Zone       string              `xml:"zone" json:"zone"`
	BackupRule *upcloud.BackupRule `xml:"backup_rule,omitempty" json:"backup_rule,omitempty"`
}

// RequestURL implements the Request interface
func (r *CreateStorageRequest) RequestURL() string {
	return "/storage"
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r CreateStorageRequest) MarshalJSON() ([]byte, error) {
	type localCreateStorageRequest CreateStorageRequest
	v := struct {
		CreateStorageRequest localCreateStorageRequest `json:"storage"`
	}{}
	v.CreateStorageRequest = localCreateStorageRequest(r)

	return json.Marshal(&v)
}

// ModifyStorageRequest represents a request to modify a storage device
type ModifyStorageRequest struct {
	XMLName xml.Name `xml:"storage" json:"-"`
	UUID    string   `xml:"-" json:"-"`

	Title      string              `xml:"title,omitempty" json:"title,omitempty"`
	Size       int                 `xml:"size,omitempty" json:"size,omitempty,string"`
	BackupRule *upcloud.BackupRule `xml:"backup_rule,omitempty" json:"backup_rule,omitempty"`
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r ModifyStorageRequest) MarshalJSON() ([]byte, error) {
	type localModifyStorageRequest ModifyStorageRequest
	v := struct {
		ModifyStorageRequest localModifyStorageRequest `json:"storage"`
	}{}
	v.ModifyStorageRequest = localModifyStorageRequest(r)

	return json.Marshal(&v)
}

// RequestURL implements the Request interface
func (r *ModifyStorageRequest) RequestURL() string {
	return fmt.Sprintf("/storage/%s", r.UUID)
}

// AttachStorageRequest represents a request to attach a storage device to a server
type AttachStorageRequest struct {
	XMLName    xml.Name `xml:"storage_device" json:"-"`
	ServerUUID string   `xml:"-" json:"-"`

	Type        string `xml:"type,omitempty" json:"type,omitempty"`
	Address     string `xml:"address,omitempty" json:"address,omitempty"`
	StorageUUID string `xml:"storage,omitempty" json:"storage,omitempty"`
	BootDisk    int    `xml:"-" json:"boot_disk,omitempty,string"`
}

// RequestURL implements the Request interface
func (r *AttachStorageRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/storage/attach", r.ServerUUID)
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r AttachStorageRequest) MarshalJSON() ([]byte, error) {
	type localAttachStorageRequest AttachStorageRequest
	v := struct {
		AttachStorageRequest localAttachStorageRequest `json:"storage_device"`
	}{}
	v.AttachStorageRequest = localAttachStorageRequest(r)

	return json.Marshal(&v)
}

// DetachStorageRequest represents a request to detach a storage device from a server
type DetachStorageRequest struct {
	XMLName    xml.Name `xml:"storage_device" json:"-"`
	ServerUUID string   `xml:"-" json:"-"`

	Address string `xml:"address" json:"address"`
}

// RequestURL implements the Request interface
func (r *DetachStorageRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/storage/detach", r.ServerUUID)
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r DetachStorageRequest) MarshalJSON() ([]byte, error) {
	type localDetachStorageRequest DetachStorageRequest
	v := struct {
		DetachStorageRequest localDetachStorageRequest `json:"storage_device"`
	}{}
	v.DetachStorageRequest = localDetachStorageRequest(r)

	return json.Marshal(&v)
}

//DeleteStorageRequest represents a request to delete a storage device
type DeleteStorageRequest struct {
	UUID string
}

// RequestURL implements the Request interface
func (r *DeleteStorageRequest) RequestURL() string {
	return fmt.Sprintf("/storage/%s", r.UUID)
}

// CloneStorageRequest represents a requests to clone a storage device
type CloneStorageRequest struct {
	XMLName xml.Name `xml:"storage" json:"-"`
	UUID    string   `xml:"-" json:"-"`

	Zone  string `xml:"zone" json:"zone"`
	Tier  string `xml:"tier,omitempty" json:"tier,omitempty"`
	Title string `xml:"title" json:"title"`
}

// RequestURL implements the Request interface
func (r *CloneStorageRequest) RequestURL() string {
	return fmt.Sprintf("/storage/%s/clone", r.UUID)
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r CloneStorageRequest) MarshalJSON() ([]byte, error) {
	type localCloneStorageRequest CloneStorageRequest
	v := struct {
		CloneStorageRequest localCloneStorageRequest `json:"storage"`
	}{}
	v.CloneStorageRequest = localCloneStorageRequest(r)

	return json.Marshal(&v)
}

// TemplatizeStorageRequest represents a request to templatize a storage device
type TemplatizeStorageRequest struct {
	XMLName xml.Name `xml:"storage" json:"-"`
	UUID    string   `xml:"-" json:"-"`

	Title string `xml:"title" json:"title"`
}

// RequestURL implements the Request interface
func (r *TemplatizeStorageRequest) RequestURL() string {
	return fmt.Sprintf("/storage/%s/templatize", r.UUID)
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r TemplatizeStorageRequest) MarshalJSON() ([]byte, error) {
	type localTemplatizeStorageRequest TemplatizeStorageRequest
	v := struct {
		TemplatizeStorageRequest localTemplatizeStorageRequest `json:"storage"`
	}{}
	v.TemplatizeStorageRequest = localTemplatizeStorageRequest(r)

	return json.Marshal(&v)
}

// WaitForStorageStateRequest represents a request to wait for a storage to enter a specific state
type WaitForStorageStateRequest struct {
	UUID         string
	DesiredState string
	Timeout      time.Duration
}

// LoadCDROMRequest represents a request to load a storage as a CD-ROM in the CD-ROM device of a server
type LoadCDROMRequest struct {
	XMLName    xml.Name `xml:"storage_device" json:"-"`
	ServerUUID string   `xml:"-" json:"-"`

	StorageUUID string `xml:"storage" json:"storage"`
}

// RequestURL implements the Request interface
func (r *LoadCDROMRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/cdrom/load", r.ServerUUID)
}

// MarshalJSON is a custom marshaller that deals with
// deeply embedded values.
func (r LoadCDROMRequest) MarshalJSON() ([]byte, error) {
	type localLoadCDROMRequest LoadCDROMRequest
	v := struct {
		LoadCDROMRequest localLoadCDROMRequest `json:"storage_device"`
	}{}
	v.LoadCDROMRequest = localLoadCDROMRequest(r)

	return json.Marshal(&v)
}

// EjectCDROMRequest represents a request to load a storage as a CD-ROM in the CD-ROM device of a server
type EjectCDROMRequest struct {
	ServerUUID string
}

// RequestURL implements the Request interface
func (r *EjectCDROMRequest) RequestURL() string {
	return fmt.Sprintf("/server/%s/cdrom/eject", r.ServerUUID)
}

// CreateBackupRequest represents a request to create a backup of a storage device
type CreateBackupRequest struct {
	XMLName xml.Name `xml:"storage"`
	UUID    string   `xml:"-"`

	Title string `xml:"title"`
}

// RequestURL implements the Request interface
func (r *CreateBackupRequest) RequestURL() string {
	return fmt.Sprintf("/storage/%s/backup", r.UUID)
}

// RestoreBackupRequest represents a request to restore a storage from the specified backup
type RestoreBackupRequest struct {
	UUID string
}

// RequestURL implements the Request interface
func (r *RestoreBackupRequest) RequestURL() string {
	return fmt.Sprintf("/storage/%s/restore", r.UUID)
}
