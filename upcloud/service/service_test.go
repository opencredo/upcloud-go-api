package service

import (
	"fmt"
	"github.com/jalle19/upcloud-go-sdk/upcloud"
	"github.com/jalle19/upcloud-go-sdk/upcloud/client"
	"github.com/jalle19/upcloud-go-sdk/upcloud/request"
	"os"
	"testing"
	"time"
)

// The service object used by the tests
var svc *Service

/**
Configures the test environment
*/
func init() {
	user, password := getCredentials()

	c := client.New(user, password)
	c.SetTimeout(time.Second * 30)
	svc = New(c)
}

/**
TestCreateServer performs the following actions:

- creates a server
- waits until the server has started
- modifies the server
- stops the server
- deletes the server

*/
func TestCreateServer(t *testing.T) {
	// Create a server
	createServerRequest := request.CreateServerRequest{
		Zone:             "fi-hel1",
		Title:            "Integration test server #1",
		Hostname:         "debian.example.com",
		PasswordDelivery: request.PasswordDeliveryNone,
		StorageDevices: []request.CreateServerStorageDevice{
			{
				Action:  request.CreateStorageDeviceActionClone,
				Storage: "01000000-0000-4000-8000-000030060200",
				Title:   "disk1",
				Size:    30,
				Tier:    request.CreateStorageDeviceTierMaxIOPS,
			},
		},
		IPAddresses: []request.CreateServerIPAddress{
			{
				Access: upcloud.IPAddressAccessPrivate,
				Family: upcloud.IPAddressFamilyIPv4,
			},
			{
				Access: upcloud.IPAddressAccessPublic,
				Family: upcloud.IPAddressFamilyIPv4,
			},
			{
				Access: upcloud.IPAddressAccessPublic,
				Family: upcloud.IPAddressFamilyIPv6,
			},
		},
	}

	// Create the server and block until it has started
	serverDetails, err := svc.CreateServer(&createServerRequest)

	if err != nil {
		panic(err)
	}

	t.Log(fmt.Sprintf("Server %s with UUID %s created", serverDetails.Title, serverDetails.UUID))
	t.Log("Waiting for server to start ...")

	err = svc.WaitForServerState(&request.WaitForServerStateRequest{
		UUID:         serverDetails.UUID,
		DesiredState: upcloud.ServerStateStarted,
		Timeout:      time.Minute * 5,
	})

	if err != nil {
		panic(err)
	}

	t.Log("Server is now active")

	// Modify the server
	t.Log("Modifying the server ...")

	serverDetails, err = svc.ModifyServer(&request.ModifyServerRequest{
		UUID:  serverDetails.UUID,
		Title: "Modified server",
	})

	if err != nil {
		panic(err)
	}

	t.Log("Waiting for the server to exit maintenance state ...")

	err = svc.WaitForServerState(&request.WaitForServerStateRequest{
		UUID:         serverDetails.UUID,
		DesiredState: upcloud.ServerStateStarted,
		Timeout:      time.Minute * 5,
	})

	if err != nil {
		panic(err)
	}

	t.Log(fmt.Sprintf("Server is now modified, new title is %s", serverDetails.Title))

	// Stop the server
	t.Log("Force stopping the server ...")

	serverDetails, err = svc.StopServer(&request.StopServerRequest{
		UUID:     serverDetails.UUID,
		StopType: request.ServerStopTypeHard,
		Timeout:  time.Minute * 5,
	})

	if err != nil {
		panic(err)
	}

	t.Log("Waiting for the server to stop ...")

	err = svc.WaitForServerState(&request.WaitForServerStateRequest{
		UUID:         serverDetails.UUID,
		DesiredState: upcloud.ServerStateStopped,
		Timeout:      time.Minute * 5,
	})

	if err != nil {
		panic(err)
	}

	t.Log("Server is now stopped")

	// Delete the server
	t.Log("Deleting the server ...")

	err = svc.DeleteServer(&request.DeleteServerRequest{
		UUID: serverDetails.UUID,
	})

	if err != nil {
		panic(err)
	}

	t.Log("Server is now deleted")
}

/**
TestCreateStorage performs the following actions:

- creates a piece of storage
- modifies the storage
- deletes the storage

*/
func TestCreateStorage(t *testing.T) {
	// Create some storage
	createStorageRequest := request.CreateStorageRequest{
		Tier:  upcloud.StorageTierMaxIOPS,
		Title: "Test storage",
		Size:  50,
		Zone:  "fi-hel1",
	}

	storageDetails, err := svc.CreateStorage(&createStorageRequest)

	if err != nil {
		panic(err)
	}

	t.Log(fmt.Sprintf("Storage %s with UUID %s created", storageDetails.Title, storageDetails.UUID))

	// Modify the storage
	t.Log("Modifying the storage ...")

	storageDetails, err = svc.ModifyStorage(&request.ModifyStorageRequest{
		UUID:  storageDetails.UUID,
		Title: "New fancy title",
	})

	if err != nil {
		panic(err)
	}

	t.Log(fmt.Sprintf("Storage with UUID %s modified successfully, new title is %s", storageDetails.UUID, storageDetails.Title))

	// Delete the storage
	t.Log("Deleting the storage ...")

	err = svc.DeleteStorage(&request.DeleteStorageRequest{
		UUID: storageDetails.UUID,
	})

	if err != nil {
		panic(err)
	}

	t.Log("Storage is now deleted")
}

/**
Reads the API username and password from the environment, panics if they are not available
*/
func getCredentials() (string, string) {
	user := os.Getenv("UPCLOUD_GO_SDK_TEST_USER")
	password := os.Getenv("UPCLOUD_GO_SDK_TEST_PASSWORD")

	if user == "" || password == "" {
		panic("Unable to retrieve credentials from the environment, ensure UPCLOUD_GO_SDK_TEST_USER and UPCLOUD_GO_SDK_TEST_PASSWORD are exported")
	}

	return user, password
}
