package service

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain is the main test method
func TestMain(m *testing.M) {
	retCode := m.Run()

	// Optionally perform teardown
	deleteResources := os.Getenv("UPCLOUD_GO_SDK_TEST_DELETE_RESOURCES")
	if deleteResources == "yes" {
		log.Print("UPCLOUD_GO_SDK_TEST_DELETE_RESOURCES defined, deleting all resources ...")
		teardown()
	}

	os.Exit(retCode)
}

// Configures the test environment
func getService() *Service {
	user, password := getCredentials()

	c := client.New(user, password)
	c.SetTimeout(time.Second * 300)

	return New(c)
}

// records the API interactions of the test
func record(t *testing.T, fixture string, f func(*testing.T, *Service)) {
	r, err := recorder.New("fixtures/" + fixture)
	require.NoError(t, err)

	r.AddFilter(func(i *cassette.Interaction) error {
		delete(i.Request.Headers, "Authorization")
		return nil
	})

	defer func() {
		err := r.Stop()
		require.NoError(t, err)
	}()

	user, password := getCredentials()

	httpClient := cleanhttp.DefaultClient()
	httpClient.Transport = r

	c := client.NewWithHTTPClient(user, password, httpClient)
	c.SetTimeout(time.Second * 300)

	f(t, New(c))
}

// Tears down the test environment by removing all resources
func teardown() {
	svc := getService()

	log.Print("Deleting all servers ...")
	servers, err := svc.GetServers()
	handleError(err)

	for _, server := range servers.Servers {
		// Try to ensure the server is not in maintenance state
		log.Printf("Waiting for server with UUID %s to leave maintenance state ...", server.UUID)
		serverDetails, err := svc.WaitForServerState(&request.WaitForServerStateRequest{
			UUID:           server.UUID,
			UndesiredState: upcloud.ServerStateMaintenance,
			Timeout:        time.Minute * 15,
		})
		handleError(err)

		// Stop the server if it's still running
		if serverDetails.State != upcloud.ServerStateStopped {
			log.Printf("Stopping server with UUID %s ...", server.UUID)
			stopServer(svc, server.UUID)
		}

		// Delete the server
		log.Printf("Deleting the server with UUID %s ...", server.UUID)
		deleteServer(svc, server.UUID)
	}

	// Delete all private storage devices
	log.Print("Deleting all storage devices ...")
	storages, err := svc.GetStorages(&request.GetStoragesRequest{
		Access: upcloud.StorageAccessPrivate,
	})
	handleError(err)

	for _, storage := range storages.Storages {
		// Wait for the storage to come online so we can delete it
		if storage.State != upcloud.StorageStateOnline {
			log.Printf("Waiting for storage %s to come online ...", storage.UUID)
			_, err = svc.WaitForStorageState(&request.WaitForStorageStateRequest{
				UUID:         storage.UUID,
				DesiredState: upcloud.StorageStateOnline,
				Timeout:      time.Minute * 15,
			})
			handleError(err)
		}

		log.Printf("Deleting the storage with UUID %s ...", storage.UUID)
		deleteStorage(svc, storage.UUID)
	}

	// Delete all tags
	log.Print("Deleting all tags ...")
	deleteAllTags(svc)
}

// TestGetAccount tests that the GetAccount() method returns proper data
func TestGetAccount(t *testing.T) {
	record(t, "getaccount", func(t *testing.T, svc *Service) {

		account, err := svc.GetAccount()
		require.NoError(t, err)

		username, _ := getCredentials()

		if account.UserName != username {
			t.Errorf("TestGetAccount expected %s, got %s", username, account.UserName)
		}

		assert.NotZero(t, account.ResourceLimits.Cores)
		assert.NotZero(t, account.ResourceLimits.Memory)
		assert.NotZero(t, account.ResourceLimits.Networks)
		assert.NotZero(t, account.ResourceLimits.PublicIPv6)
		assert.NotZero(t, account.ResourceLimits.StorageHDD)
		assert.NotZero(t, account.ResourceLimits.StorageSSD)
	})
}

// TestGetZones tests that the GetZones() function returns proper data
func TestGetZones(t *testing.T) {
	record(t, "getzones", func(t *testing.T, svc *Service) {
		zones, err := svc.GetZones()
		require.NoError(t, err)
		assert.NotEmpty(t, zones.Zones)

		var found bool
		for _, z := range zones.Zones {
			if z.Description == "Helsinki #1" && z.ID == "fi-hel1" {
				found = true
				assert.True(t, bool(z.Public))
				break
			}
		}
		assert.True(t, found)
	})
}

// TestGetPriceZones tests that GetPriceZones() function returns proper data
func TestGetPriceZones(t *testing.T) {
	record(t, "getpricezones", func(t *testing.T, svc *Service) {
		zones, err := svc.GetPriceZones()
		require.NoError(t, err)
		assert.NotEmpty(t, zones.PriceZones)

		var found bool
		var zone upcloud.PriceZone
		for _, z := range zones.PriceZones {
			if z.Name == "fi-hel1" {
				found = true
				zone = z
				break
			}
		}
		assert.True(t, found)
		assert.NotZero(t, zone.Firewall.Amount)
		assert.NotZero(t, zone.Firewall.Price)
		assert.NotZero(t, zone.IPv4Address.Amount)
		assert.NotZero(t, zone.IPv4Address.Price)
	})
}

// TestGetTimeZones ensures that the GetTimeZones() function returns proper data
func TestGetTimeZones(t *testing.T) {
	record(t, "gettimezones", func(t *testing.T, svc *Service) {
		zones, err := svc.GetTimeZones()
		require.NoError(t, err)
		assert.NotEmpty(t, zones.TimeZones)

		var found bool
		for _, z := range zones.TimeZones {
			if z == "Pacific/Wallis" {
				found = true
				break
			}
		}
		assert.True(t, found)
	})
}

// TestGetPlans ensures that the GetPlans() functions returns proper data
func TestGetPlans(t *testing.T) {
	record(t, "getplans", func(t *testing.T, svc *Service) {
		plans, err := svc.GetPlans()
		require.NoError(t, err)
		assert.NotEmpty(t, plans.Plans)

		var found bool
		var plan upcloud.Plan
		for _, p := range plans.Plans {
			if p.Name == "1xCPU-1GB" {
				found = true
				plan = p
				break
			}
		}
		assert.True(t, found)

		assert.Equal(t, 1, plan.CoreNumber)
		assert.Equal(t, 1024, plan.MemoryAmount)
		assert.Equal(t, 1024, plan.PublicTrafficOut)
		assert.Equal(t, 25, plan.StorageSize)
		assert.Equal(t, upcloud.StorageTierMaxIOPS, plan.StorageTier)
	})
}

// TestGetServerConfigurations ensures that the GetServerConfigurations() function returns proper data
func TestGetServerConfigurations(t *testing.T) {
	record(t, "getserverconfigurations", func(t *testing.T, svc *Service) {
		configurations, err := svc.GetServerConfigurations()
		require.NoError(t, err)
		assert.NotEmpty(t, configurations.ServerConfigurations)

		var found bool
		for _, sc := range configurations.ServerConfigurations {
			if sc.CoreNumber == 1 && sc.MemoryAmount == 1024 {
				found = true
				break
			}
		}
		assert.True(t, found)
	})
}

// TestGetServerDetails ensures that the GetServerDetails() function returns proper data
func TestGetServerDetails(t *testing.T) {
	record(t, "getserverdetails", func(t *testing.T, svc *Service) {
		d, err := createServer(svc, "getserverdetails")
		require.NoError(t, err)

		serverDetails, err := svc.GetServerDetails(&request.GetServerDetailsRequest{
			UUID: d.UUID,
		})
		require.NoError(t, err)

		assert.Contains(t, serverDetails.Title, "getserverdetails")
		assert.Equal(t, "fi-hel2", serverDetails.Zone)
	})
}

// TestCreateStopStartServer ensures that StartServer() and StopServer() behave
// as expect and return proper data
// The test:
//   - Creates a server
//   - Stops the server
//   - Starts the server
//   - Checks the details of the started server and that it is in the
//     correct state.
func TestCreateStopStartServer(t *testing.T) {
	record(t, "createstartstopserver", func(t *testing.T, svc *Service) {
		d, err := createServer(svc, "createstartstopserver")
		require.NoError(t, err)

		stopServerDetails, err := svc.StopServer(&request.StopServerRequest{
			UUID:     d.UUID,
			Timeout:  15 * time.Minute,
			StopType: upcloud.StopTypeHard,
		})
		require.NoError(t, err)
		assert.Contains(t, stopServerDetails.Title, "createstartstopserver")
		assert.Equal(t, "fi-hel2", stopServerDetails.Zone)
		// We shouldn't have transitioned state yet.
		assert.Equal(t, upcloud.ServerStateStarted, stopServerDetails.State)

		waitServerDetails, err := svc.WaitForServerState(&request.WaitForServerStateRequest{
			UUID:         d.UUID,
			DesiredState: upcloud.ServerStateStopped,
			Timeout:      15 * time.Minute,
		})
		require.NoError(t, err)
		assert.Contains(t, waitServerDetails.Title, "createstartstopserver")
		assert.Equal(t, "fi-hel2", waitServerDetails.Zone)
		assert.Equal(t, upcloud.ServerStateStopped, waitServerDetails.State)

		startServerDetails, err := svc.StartServer(&request.StartServerRequest{
			UUID: d.UUID,
		})
		require.NoError(t, err)

		assert.Contains(t, startServerDetails.Title, "createstartstopserver")
		assert.Equal(t, "fi-hel2", startServerDetails.Zone)
		assert.Equal(t, upcloud.ServerStateStarted, startServerDetails.State)
	})
}

func TestStartAvoidHost(t *testing.T) {
	record(t, "startavoidhost", func(t *testing.T, svc *Service) {
		serverDetails, err := createServer(svc, "TestStartAvoidHost")
		require.NoError(t, err)
		assert.NotZero(t, serverDetails.Host)

		_, err = svc.StopServer(&request.StopServerRequest{
			UUID:     serverDetails.UUID,
			StopType: upcloud.StopTypeHard,
		})
		require.NoError(t, err)

		_, err = svc.WaitForServerState(&request.WaitForServerStateRequest{
			UUID:         serverDetails.UUID,
			DesiredState: upcloud.ServerStateStopped,
			Timeout:      15 * time.Minute,
		})
		require.NoError(t, err)

		postServerDetails, err := svc.StartServer(&request.StartServerRequest{
			UUID:      serverDetails.UUID,
			AvoidHost: serverDetails.Host,
		})
		require.NoError(t, err)
		assert.NotZero(t, postServerDetails.Host)
		assert.NotEqual(t, serverDetails.Host, postServerDetails.Host)
	})
}

// TestCreateRestartServer ensures that RestartServer() behaves as expect and returns
// proper data
// The test:
//   - Creates a server
//   - Restarts the server
//   - Checks the details of the restarted server and that it is in the
//     correct state.
func TestCreateRestartServer(t *testing.T) {
	record(t, "createrestartserver", func(t *testing.T, svc *Service) {
		d, err := createServer(svc, "createrestartserver")
		require.NoError(t, err)

		restartServerDetails, err := svc.RestartServer(&request.RestartServerRequest{
			UUID:          d.UUID,
			Timeout:       15 * time.Minute,
			StopType:      upcloud.StopTypeHard,
			TimeoutAction: request.RestartTimeoutActionIgnore,
		})
		require.NoError(t, err)
		assert.Contains(t, restartServerDetails.Title, "createrestartserver")
		assert.Equal(t, "fi-hel2", restartServerDetails.Zone)
		// We shouldn't have transitioned state yet.
		assert.Equal(t, upcloud.ServerStateStarted, restartServerDetails.State)

		waitServerDetails, err := svc.WaitForServerState(&request.WaitForServerStateRequest{
			UUID:           d.UUID,
			UndesiredState: upcloud.ServerStateStarted,
			Timeout:        15 * time.Minute,
		})
		require.NoError(t, err)
		assert.Contains(t, waitServerDetails.Title, "createrestartserver")
		assert.Equal(t, "fi-hel2", waitServerDetails.Zone)

		waitServerDetails2, err := svc.WaitForServerState(&request.WaitForServerStateRequest{
			UUID:         waitServerDetails.UUID,
			DesiredState: upcloud.ServerStateStarted,
			Timeout:      15 * time.Minute,
		})
		require.NoError(t, err)
		assert.Contains(t, waitServerDetails2.Title, "createrestartserver")
		assert.Equal(t, "fi-hel2", waitServerDetails2.Zone)
		assert.Equal(t, upcloud.ServerStateStarted, waitServerDetails2.State)
	})
}

// TestErrorHandling checks that the correct error type is returned from service methods
func TestErrorHandling(t *testing.T) {
	record(t, "errorhandling", func(t *testing.T, svc *Service) {
		// Perform a bogus request that will certainly fail
		_, err := svc.StartServer(&request.StartServerRequest{
			UUID: "invalid",
		})

		// Check that the correct error type is returned
		expectedErrorType := "*upcloud.Error"
		actualErrorType := reflect.TypeOf(err).String()

		if actualErrorType != expectedErrorType {
			t.Errorf("TestErrorHandling expected %s, got %s", expectedErrorType, actualErrorType)
		}
	})
}

// TestCreateModifyDeleteServer performs the following actions:
//
// - creates a server
// - modifies the server
// - stops the server
// - deletes the server
func TestCreateModifyDeleteServer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	record(t, "createmodifydeleteserver", func(t *testing.T, svc *Service) {
		// Create a server
		serverDetails, err := createServer(svc, "TestCreateModifyDeleteServer")
		require.NoError(t, err)
		t.Logf("Server %s with UUID %s created", serverDetails.Title, serverDetails.UUID)

		// Get details about the storage (UUID is required for testing)
		if len(serverDetails.StorageDevices) == 0 {
			t.Errorf("Server %s with UUID %s has no storages attached", serverDetails.Title, serverDetails.UUID)
		}

		firstStorage := serverDetails.StorageDevices[0]
		storageUUID := firstStorage.UUID

		t.Logf("First storage of server with UUID %s has UUID %s", serverDetails.UUID, storageUUID)

		// Modify the server
		t.Log("Modifying the server ...")

		newTitle := "Modified server"
		_, err = svc.ModifyServer(&request.ModifyServerRequest{
			UUID:  serverDetails.UUID,
			Title: newTitle,
		})

		require.NoError(t, err)
		t.Log("Waiting for the server to exit maintenance state ...")

		serverDetails, err = svc.WaitForServerState(&request.WaitForServerStateRequest{
			UUID:         serverDetails.UUID,
			DesiredState: upcloud.ServerStateStarted,
			Timeout:      time.Minute * 15,
		})

		require.NoError(t, err)
		assert.Equal(t, newTitle, serverDetails.Title)
		t.Logf("Server is now modified, new title is %s", serverDetails.Title)

		// Stop the server
		t.Logf("Stopping server with UUID %s ...", serverDetails.UUID)
		err = stopServer(svc, serverDetails.UUID)
		require.NoError(t, err)
		t.Log("Server is now stopped")

		// Delete the server
		t.Logf("Deleting the server with UUID %s...", serverDetails.UUID)
		err = deleteServer(svc, serverDetails.UUID)
		require.NoError(t, err)
		t.Log("Server is now deleted")

		// Check if the storage still exists
		storages, err := svc.GetStorages(&request.GetStoragesRequest{
			Access: upcloud.StorageAccessPrivate,
		})
		require.NoError(t, err)

		found := false
		for _, storage := range storages.Storages {
			if storage.UUID == storageUUID {
				found = true
				break
			}
		}
		assert.Truef(t, found, "Storage with UUID %s not found. It should still exist after deleting server with UUID %s", storageUUID, serverDetails.UUID)

		t.Log("Storage still exists")
	})
}

// TestCreateDeleteServerAndStorage performs the following actions:
//
// - creates a server
// - deletes the server including storage
func TestCreateDeleteServerAndStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	record(t, "createdeleteserverandstorage", func(t *testing.T, svc *Service) {
		// Create a server
		serverDetails, err := createServer(svc, "TestCreateDeleteServerAndStorage")
		require.NoError(t, err)
		t.Logf("Server %s with UUID %s created", serverDetails.Title, serverDetails.UUID)

		// Get details about the storage (UUID is required for testing)
		assert.NotEmptyf(t, serverDetails.StorageDevices, "Server %s with UUID %s has no storages attached", serverDetails.Title, serverDetails.UUID)

		firstStorage := serverDetails.StorageDevices[0]
		storageUUID := firstStorage.UUID
		t.Logf("First storage of server with UUID %s has UUID %s", serverDetails.UUID, storageUUID)

		// Stop the server
		t.Logf("Stopping server with UUID %s ...", serverDetails.UUID)
		err = stopServer(svc, serverDetails.UUID)
		require.NoError(t, err)
		t.Log("Server is now stopped")

		// Delete the server and storage
		t.Logf("Deleting the server with UUID %s, including storages...", serverDetails.UUID)
		err = deleteServerAndStorages(svc, serverDetails.UUID)
		require.NoError(t, err)
		t.Log("Server is now deleted")

		// Check if the storage was deleted
		storages, err := svc.GetStorages(&request.GetStoragesRequest{
			Access: upcloud.StorageAccessPrivate,
		})
		require.NoError(t, err)

		found := false
		for _, storage := range storages.Storages {
			if storage.UUID == storageUUID {
				found = true
				break
			}
		}
		assert.Falsef(t, found, "Storage with UUID %s still exists. It should have been deleted with server with UUID %s", storageUUID, serverDetails.UUID)

		t.Log("Storage was deleted, too")
	})
}

// TestCreateModifyDeleteStorage performs the following actions:
//
// - creates a new storage disk
// - modifies the storage
// - deletes the storage
func TestCreateModifyDeleteStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	record(t, "createmodifydeletestorage", func(t *testing.T, svc *Service) {
		// Create some storage
		storageDetails, err := createStorage(svc)
		require.NoError(t, err)
		t.Logf("Storage %s with UUID %s created", storageDetails.Title, storageDetails.UUID)

		// Modify the storage
		t.Log("Modifying the storage ...")

		newTitle := "New fancy title"
		storageDetails, err = svc.ModifyStorage(&request.ModifyStorageRequest{
			UUID:  storageDetails.UUID,
			Title: newTitle,
		})
		require.NoError(t, err)
		assert.Equal(t, newTitle, storageDetails.Title)
		t.Logf("Storage with UUID %s modified successfully, new title is %s", storageDetails.UUID, storageDetails.Title)

		// Delete the storage
		t.Log("Deleting the storage ...")
		err = deleteStorage(svc, storageDetails.UUID)
		require.NoError(t, err)
		t.Log("Storage is now deleted")
	})
}

// TestAttachDetachStorage performs the following actions:
//
// - creates a server
// - stops the server
// - creates a new storage disk
// - attaches the storage
// - detaches the storage
// - deletes the storage
// - deletes the server
func TestAttachDetachStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	record(t, "attachdetachstorage", func(t *testing.T, svc *Service) {
		// Create a server
		serverDetails, err := createServer(svc, "TestAttachDetachStorage")
		require.NoError(t, err)
		t.Logf("Server %s with UUID %s created", serverDetails.Title, serverDetails.UUID)

		// Stop the server
		t.Logf("Stopping server with UUID %s ...", serverDetails.UUID)
		err = stopServer(svc, serverDetails.UUID)
		require.NoError(t, err)
		t.Log("Server is now stopped")

		// Create some storage
		storageDetails, err := createStorage(svc)
		require.NoError(t, err)
		t.Logf("Storage %s with UUID %s created", storageDetails.Title, storageDetails.UUID)

		// Attach the storage
		t.Logf("Attaching storage %s", storageDetails.UUID)

		serverDetails, err = svc.AttachStorage(&request.AttachStorageRequest{
			StorageUUID: storageDetails.UUID,
			ServerUUID:  serverDetails.UUID,
			Type:        upcloud.StorageTypeDisk,
			Address:     "scsi:0:0",
		})
		require.NoError(t, err)
		t.Logf("Storage attached to server with UUID %s", serverDetails.UUID)

		// Detach the storage
		t.Logf("Detaching storage %s", storageDetails.UUID)

		_, err = svc.DetachStorage(&request.DetachStorageRequest{
			ServerUUID: serverDetails.UUID,
			Address:    "scsi:0:0",
		})
		require.NoError(t, err)
		t.Logf("Storage %s detached", storageDetails.UUID)
	})
}

// TestCloneStorage performs the following actions:
//
// - creates a storage device
// - clones the storage device
// - deletes the clone and the storage device
func TestCloneStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	record(t, "clonestorage", func(t *testing.T, svc *Service) {
		// Create storage
		storageDetails, err := createStorage(svc)
		require.NoError(t, err)
		t.Logf("Storage %s with UUID %s created", storageDetails.Title, storageDetails.UUID)

		// Clone the storage
		t.Log("Cloning storage ...")

		clonedStorageDetails, err := svc.CloneStorage(&request.CloneStorageRequest{
			UUID:  storageDetails.UUID,
			Title: "Cloned storage",
			Zone:  "fi-hel2",
			Tier:  upcloud.StorageTierMaxIOPS,
		})
		require.NoError(t, err)
		err = waitForStorageOnline(svc, clonedStorageDetails.UUID)
		require.NoError(t, err)
		t.Logf("Storage cloned as %s", clonedStorageDetails.UUID)
	})
}

// TestTemplatizeServerStorage performs the following actions:
//
// - creates a server
// - templatizes the server's storage
// - deletes the new storage
// - stops and deletes the server
func TestTemplatizeServerStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	record(t, "templatizeserverstorage", func(t *testing.T, svc *Service) {
		// Create server
		serverDetails, err := createServer(svc, "TestTemplatizeServerStorage")
		require.NoError(t, err)
		t.Logf("Server %s with UUID %s created", serverDetails.Title, serverDetails.UUID)

		// Stop the server
		t.Logf("Stopping server with UUID %s ...", serverDetails.UUID)
		err = stopServer(svc, serverDetails.UUID)
		require.NoError(t, err)
		t.Log("Server is now stopped")

		// Get extended service details
		serverDetails, err = svc.GetServerDetails(&request.GetServerDetailsRequest{
			UUID: serverDetails.UUID,
		})
		require.NoError(t, err)

		// Templatize the server's first storage device
		require.NotEmpty(t, serverDetails.StorageDevices)
		t.Log("Templatizing storage ...")

		storageDetails, err := svc.TemplatizeStorage(&request.TemplatizeStorageRequest{
			UUID:  serverDetails.StorageDevices[0].UUID,
			Title: "Templatized storage",
		})
		require.NoErrorf(t, err, "Error: %#v", err)

		err = waitForStorageOnline(svc, storageDetails.UUID)
		require.NoError(t, err)
		t.Logf("Storage templatized as %s", storageDetails.UUID)
	})
}

// TestLoadEjectCDROM performs the following actions:
//
// - creates a server
// - stops the server
// - attaches a CD-ROM device
// - loads a CD-ROM
// - ejects the CD-ROM
func TestLoadEjectCDROM(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	record(t, "loadejectcdrom", func(t *testing.T, svc *Service) {
		// Create the server
		serverDetails, err := createServer(svc, "TestLoadEjectCDROM")
		require.NoError(t, err)
		t.Logf("Server %s with UUID %s created", serverDetails.Title, serverDetails.UUID)

		// Stop the server
		t.Logf("Stopping server with UUID %s ...", serverDetails.UUID)
		err = stopServer(svc, serverDetails.UUID)
		require.NoError(t, err)
		t.Log("Server is now stopped")

		// Attach CD-ROM device
		t.Logf("Attaching CD-ROM device to server with UUID %s", serverDetails.UUID)
		_, err = svc.AttachStorage(&request.AttachStorageRequest{
			ServerUUID: serverDetails.UUID,
			Type:       upcloud.StorageTypeCDROM,
		})
		require.NoError(t, err)
		t.Log("CD-ROM is now attached")

		// Load the CD-ROM
		t.Log("Loading CD-ROM into CD-ROM device")
		_, err = svc.LoadCDROM(&request.LoadCDROMRequest{
			ServerUUID:  serverDetails.UUID,
			StorageUUID: "01000000-0000-4000-8000-000030060101",
		})
		require.NoError(t, err)
		t.Log("CD-ROM is now loaded")

		// Eject the CD-ROM
		t.Log("Ejecting CD-ROM from CD-ROM device")
		_, err = svc.EjectCDROM(&request.EjectCDROMRequest{
			ServerUUID: serverDetails.UUID,
		})
		require.NoError(t, err)
		t.Log("CD-ROM is now ejected")
	})
}

// TestCreateRestoreBackup performs the following actions:
//
// - creates a storage device
// - creates a backup of the storage device
// - gets backup storage details
// - restores the backup
//
func TestCreateRestoreBackup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	record(t, "createrestorebackup", func(t *testing.T, svc *Service) {
		// Create the storage
		storageDetails, err := createStorage(svc)
		require.NoError(t, err)
		t.Logf("Storage %s with UUID %s created", storageDetails.Title, storageDetails.UUID)

		// Create a backup
		t.Logf("Creating backup of storage with UUID %s ...", storageDetails.UUID)

		timeBeforeBackup, err := utcTimeWithSecondPrecision()
		require.NoError(t, err)

		// Because we are recording the API tests we need to store the 'before'
		// time for the later check. We're storing it in the Title field.
		backupDetails, err := svc.CreateBackup(&request.CreateBackupRequest{
			UUID:  storageDetails.UUID,
			Title: fmt.Sprintf("backup-%d", timeBeforeBackup.UnixNano()),
		})
		require.NoError(t, err)

		err = waitForStorageOnline(svc, storageDetails.UUID)
		require.NoError(t, err)

		t.Logf("Created backup with UUID %s", backupDetails.UUID)

		// Get backup storage details
		t.Logf("Getting details of backup storage with UUID %s ...", backupDetails.UUID)

		backupStorageDetails, err := svc.GetStorageDetails(&request.GetStorageDetailsRequest{
			UUID: backupDetails.UUID,
		})
		require.NoError(t, err)

		assert.Equalf(
			t,
			backupStorageDetails.Origin,
			storageDetails.UUID,
			"The origin UUID %s of backup storage UUID %s does not match the actual origin UUID %s",
			backupStorageDetails.Origin,
			backupDetails.UUID,
			storageDetails.UUID,
		)
		t.Logf("Backup storage origin UUID OK")

		err = svc.RestoreBackup(&request.RestoreBackupRequest{
			UUID: backupDetails.UUID,
		})
		assert.NoError(t, err)

		err = waitForStorageOnline(svc, backupDetails.Origin)
		require.NoError(t, err)
	})
}

// TestGetIPAddresses performs the following actions:
// - creates a server
// - retrieves all IP addresses
// - compares the retrieved IP addresses with the created server's
//   ip addresses
func TestGetIPAddresses(t *testing.T) {
	record(t, "getipaddresses", func(t *testing.T, svc *Service) {
		serverDetails, err := createServer(svc, "TestGetIPAddresses")
		require.NoError(t, err)
		assert.Greater(t, len(serverDetails.IPAddresses), 0)

		ipAddresses, err := svc.GetIPAddresses()
		require.NoError(t, err)
		var foundCount int
		for _, sip := range serverDetails.IPAddresses {
			for _, gip := range ipAddresses.IPAddresses {
				if sip.Address == gip.Address {
					foundCount++
					if sip.Access == upcloud.IPAddressAccessPrivate {
						// Workaround during transition
						sip.Access = upcloud.IPAddressAccessUtility
					}
					assert.Equal(t, sip.Access, gip.Access)
					assert.Equal(t, sip.Family, gip.Family)
					break
				}
			}
		}
		assert.Equal(t, len(serverDetails.IPAddresses), foundCount)

		for _, ip := range serverDetails.IPAddresses {
			require.NotEmpty(t, ip.Address)
			ipAddress, err := svc.GetIPAddressDetails(&request.GetIPAddressDetailsRequest{
				Address: ip.Address,
			})
			require.NoError(t, err)

			assert.Equal(t, ip.Address, ipAddress.Address)
			if ip.Access == upcloud.IPAddressAccessPrivate {
				// Workaround during transition
				ip.Access = upcloud.IPAddressAccessUtility
			}
			assert.Equal(t, ip.Access, ipAddress.Access)
			assert.Equal(t, ip.Family, ipAddress.Family)
		}
	})
}

// TestAttachModifyReleaseIPAddress performs the following actions
//
// - creates a server
// - assigns an additional IP address to it
// - modifies the PTR record of the IP address
// - deletes the IP address
func TestAttachModifyReleaseIPAddress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	record(t, "attachmodifyreleaseipaddress", func(t *testing.T, svc *Service) {
		// Create the server
		serverDetails, err := createServer(svc, "TestAttachModifyReleaseIPAddress")
		require.NoError(t, err)
		t.Logf("Server %s with UUID %s created", serverDetails.Title, serverDetails.UUID)

		// Stop the server
		t.Logf("Stopping server with UUID %s ...", serverDetails.UUID)
		err = stopServer(svc, serverDetails.UUID)
		require.NoError(t, err)
		t.Log("Server is now stopped")

		// Assign an IP address
		t.Log("Assigning IP address to server ...")
		ipAddress, err := svc.AssignIPAddress(&request.AssignIPAddressRequest{
			Access:     upcloud.IPAddressAccessPublic,
			Family:     upcloud.IPAddressFamilyIPv6,
			ServerUUID: serverDetails.UUID,
		})
		require.NoError(t, err)
		t.Logf("Assigned IP address %s to server with UUID %s", ipAddress.Address, serverDetails.UUID)

		// Modify the PTR record
		t.Logf("Modifying PTR record for address %s ...", ipAddress.Address)
		ipAddress, err = svc.ModifyIPAddress(&request.ModifyIPAddressRequest{
			IPAddress: ipAddress.Address,
			PTRRecord: "such.pointer.example.com",
		})
		require.NoError(t, err)
		t.Logf("PTR record modified, new record is %s", ipAddress.PTRRecord)

		// Release the IP address
		t.Log("Releasing the IP address ...")
		err = svc.ReleaseIPAddress(&request.ReleaseIPAddressRequest{
			IPAddress: ipAddress.Address,
		})
		require.NoError(t, err)
		t.Log("The IP address is now released")
	})
}

func TestAttachModifyReleaseFloatingIPAddress(t *testing.T) {
	record(t, "attachmodifyreleasefloatingipaddress", func(t *testing.T, svc *Service) {
		// Create the first server
		serverDetails1, err := createServer(svc, "TestAttachModifyReleaseIPAddress1")
		require.NoError(t, err)
		t.Logf("Server 1 %s with UUID %s created", serverDetails1.Title, serverDetails1.UUID)

		// Create the second server
		serverDetails2, err := createServer(svc, "TestAttachModifyReleaseIPAddress2")
		require.NoError(t, err)
		t.Logf("Server 2 %s with UUID %s created", serverDetails2.Title, serverDetails2.UUID)

		var mac string
		for _, ip := range serverDetails1.IPAddresses {
			if ip.Access == upcloud.IPAddressAccessPublic && ip.Family == upcloud.IPAddressFamilyIPv4 {
				ipDetails, err := svc.GetIPAddressDetails(&request.GetIPAddressDetailsRequest{
					Address: ip.Address,
				})
				require.NoError(t, err)
				mac = ipDetails.MAC
				break
			}
		}
		require.NotEmpty(t, mac)

		assignedIP, err := svc.AssignIPAddress(&request.AssignIPAddressRequest{
			Family:   upcloud.IPAddressFamilyIPv4,
			Floating: true,
			MAC:      mac,
		})
		require.NoError(t, err)

		postAssignServerDetails1, err := svc.GetServerDetails(&request.GetServerDetailsRequest{
			UUID: serverDetails1.UUID,
		})
		require.NoError(t, err)

		var found bool
		for _, inf := range postAssignServerDetails1.Networking.Interfaces {
			for _, ip := range inf.IPAddresses {
				if ip.Address == assignedIP.Address {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		assert.True(t, found)

		var mac2 string
		for _, ip := range serverDetails2.IPAddresses {
			if ip.Access == upcloud.IPAddressAccessPublic && ip.Family == upcloud.IPAddressFamilyIPv4 {
				ipDetails, err := svc.GetIPAddressDetails(&request.GetIPAddressDetailsRequest{
					Address: ip.Address,
				})
				require.NoError(t, err)
				mac2 = ipDetails.MAC
				break
			}
		}
		require.NotEmpty(t, mac2)

		_, err = svc.ModifyIPAddress(&request.ModifyIPAddressRequest{
			IPAddress: assignedIP.Address,
			MAC:       mac2,
		})
		require.NoError(t, err)

		postModifyServerDetails1, err := svc.GetServerDetails(&request.GetServerDetailsRequest{
			UUID: serverDetails1.UUID,
		})
		require.NoError(t, err)

		found = false
		for _, inf := range postModifyServerDetails1.Networking.Interfaces {
			for _, ip := range inf.IPAddresses {
				if ip.Address == assignedIP.Address {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		assert.False(t, found)

		postModifyServerDetails2, err := svc.GetServerDetails(&request.GetServerDetailsRequest{
			UUID: serverDetails2.UUID,
		})
		require.NoError(t, err)

		found = false
		for _, inf := range postModifyServerDetails2.Networking.Interfaces {
			for _, ip := range inf.IPAddresses {
				if ip.Address == assignedIP.Address {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		assert.True(t, found)

		// Unassign IP
		unassignIP, err := svc.ModifyIPAddress(&request.ModifyIPAddressRequest{
			IPAddress: assignedIP.Address,
		})
		require.NoError(t, err)
		assert.Empty(t, unassignIP.ServerUUID)
		assert.Empty(t, unassignIP.MAC)

		err = svc.ReleaseIPAddress(&request.ReleaseIPAddressRequest{
			IPAddress: assignedIP.Address,
		})
		require.NoError(t, err)
	})
}

// TestFirewallRules performs the following actions:
//
// - creates a server
// - adds a firewall rule to the server
// - gets details about the firewall rule
// - deletes the firewall rule
//
func TestFirewallRules(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	record(t, "firewallrules", func(t *testing.T, svc *Service) {
		// Create the server
		serverDetails, err := createServer(svc, "TestFirewallRules")
		require.NoError(t, err)
		t.Logf("Server %s with UUID %s created", serverDetails.Title, serverDetails.UUID)

		// Create firewall rule
		t.Logf("Creating firewall rule #1 for server with UUID %s ...", serverDetails.UUID)
		_, err = svc.CreateFirewallRule(&request.CreateFirewallRuleRequest{
			ServerUUID: serverDetails.UUID,
			FirewallRule: upcloud.FirewallRule{
				Direction: upcloud.FirewallRuleDirectionIn,
				Action:    upcloud.FirewallRuleActionAccept,
				Family:    upcloud.IPAddressFamilyIPv4,
				Protocol:  upcloud.FirewallRuleProtocolTCP,
				Position:  1,
				Comment:   "This is the comment",
			},
		})
		require.NoError(t, err)
		t.Log("Firewall rule created")

		// Get list of firewall rules for this server
		firewallRules, err := svc.GetFirewallRules(&request.GetFirewallRulesRequest{
			ServerUUID: serverDetails.UUID,
		})
		require.NoError(t, err)
		assert.Len(t, firewallRules.FirewallRules, 1)
		assert.Equal(t, "This is the comment", firewallRules.FirewallRules[0].Comment)

		// Get details about the rule
		t.Log("Getting details about firewall rule #1 ...")
		firewallRule, err := svc.GetFirewallRuleDetails(&request.GetFirewallRuleDetailsRequest{
			ServerUUID: serverDetails.UUID,
			Position:   1,
		})
		require.NoError(t, err)
		assert.Equal(t, "This is the comment", firewallRule.Comment)
		t.Logf("Got firewall rule details, comment is %s", firewallRule.Comment)

		// Delete the firewall rule
		t.Log("Deleting firewall rule #1 ...")
		err = svc.DeleteFirewallRule(&request.DeleteFirewallRuleRequest{
			ServerUUID: serverDetails.UUID,
			Position:   1,
		})
		require.NoError(t, err)
		t.Log("Firewall rule #1 deleted")
	})
}

// TestCreateTag tests the creation of a single tag
func TestCreateTag(t *testing.T) {
	record(t, "createtag", func(t *testing.T, svc *Service) {
		svc.DeleteTag(&request.DeleteTagRequest{
			Name: "testTag",
		})

		tag, err := svc.CreateTag(&request.CreateTagRequest{
			Tag: upcloud.Tag{
				Name: "testTag",
			},
		})
		require.NoError(t, err)
		assert.Equal(t, "testTag", tag.Name)
	})
}

// TestGetTags tests that GetTags returns multiple tags and it, at least, contains the 3
// we create.
func TestGetTags(t *testing.T) {
	record(t, "gettags", func(t *testing.T, svc *Service) {
		testData := []string{
			"testgettags_tag1",
			"testgettags_tag2",
			"testgettags_tag3",
		}

		for _, tag := range testData {
			// Delete all the tags we're about to create.
			// We don't care about errors.
			svc.DeleteTag(&request.DeleteTagRequest{
				Name: tag,
			})
		}

		for _, tag := range testData {
			_, err := svc.CreateTag(&request.CreateTagRequest{
				Tag: upcloud.Tag{
					Name:        tag,
					Description: tag + " description",
				},
			})

			require.NoError(t, err)
		}

		tags, err := svc.GetTags()
		require.NoError(t, err)
		// There may be other tags so the length must be
		// greater than or equal to.
		assert.GreaterOrEqual(t, len(tags.Tags), len(testData))
		for _, expectedTag := range testData {
			var found bool
			for _, tag := range tags.Tags {
				if tag.Name == expectedTag {
					found = true
					assert.Equal(t, expectedTag+" description", tag.Description)
					break
				}
			}
			assert.True(t, found)
		}

		for _, tag := range tags.Tags {
			err := svc.DeleteTag(&request.DeleteTagRequest{
				Name: tag.Name,
			})
			require.NoError(t, err)
		}
	})
}

// TestTagging tests that all tagging-related functionality works correctly. It performs the following actions:
//   - creates a server
//   - creates three tags
//   - assigns the first tag to the server
//   - renames the second tag
//   - deletes the third tag
//   - untags the first tag from the server
func TestTagging(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	record(t, "tagging", func(t *testing.T, svc *Service) {
		// Create the server
		serverDetails, err := createServer(svc, "TestTagging")
		require.NoError(t, err)
		t.Logf("Server %s with UUID %s created", serverDetails.Title, serverDetails.UUID)

		// Remove all existing tags
		t.Log("Deleting any existing tags ...")
		err = deleteAllTags(svc)
		require.NoError(t, err)

		// Create three tags
		tags := []string{
			"tag1",
			"tag2",
			"tag3",
		}

		for _, tag := range tags {
			t.Logf("Creating tag %s", tag)
			tagDetails, err := svc.CreateTag(&request.CreateTagRequest{
				Tag: upcloud.Tag{
					Name: tag,
				},
			})

			require.NoError(t, err)
			assert.Equal(t, tag, tagDetails.Name)
			t.Logf("Tag %s created", tagDetails.Name)
		}

		// Assign the first tag to the server
		serverDetails, err = svc.TagServer(&request.TagServerRequest{
			UUID: serverDetails.UUID,
			Tags: []string{
				"tag1",
			},
		})
		require.NoError(t, err)
		assert.Contains(t, serverDetails.Tags, "tag1")
		var utilityCount int
		for _, ip := range serverDetails.IPAddresses {
			assert.NotEqual(t, upcloud.IPAddressAccessPrivate, ip.Access)
			if ip.Access == upcloud.IPAddressAccessUtility {
				utilityCount++
			}
		}
		assert.NotZero(t, utilityCount)
		t.Logf("Server %s is now tagged with tag %s", serverDetails.Title, "tag1")

		// Rename the second tag
		tagDetails, err := svc.ModifyTag(&request.ModifyTagRequest{
			Name: "tag2",
			Tag: upcloud.Tag{
				Name: "tag2_renamed",
			},
		})

		require.NoError(t, err)
		assert.Equal(t, "tag2_renamed", tagDetails.Name)
		t.Logf("Tag tag2 renamed to %s", tagDetails.Name)

		// Delete the third tag
		err = svc.DeleteTag(&request.DeleteTagRequest{
			Name: "tag3",
		})

		require.NoError(t, err)
		t.Log("Tag tag3 deleted")

		// Untag the server
		t.Logf("Removing tag %s from server %s", "tag1", serverDetails.UUID)
		serverDetails, err = svc.UntagServer(&request.UntagServerRequest{
			UUID: serverDetails.UUID,
			Tags: []string{
				"tag1",
			},
		})
		require.NoError(t, err)
		assert.NotContains(t, serverDetails.Tags, "tag1")
		utilityCount = 0
		for _, ip := range serverDetails.IPAddresses {
			assert.NotEqual(t, upcloud.IPAddressAccessPrivate, ip.Access)
			if ip.Access == upcloud.IPAddressAccessUtility {
				utilityCount++
			}
		}
		assert.NotZero(t, utilityCount)
		t.Logf("Server %s is now untagged", serverDetails.Title)
	})
}

// Creates a server and returns the details about it, panic if creation fails
func createServer(svc *Service, name string) (*upcloud.ServerDetails, error) {
	title := "uploud-go-sdk-integration-test-" + name
	hostname := strings.ToLower(title + ".example.com")

	createServerRequest := request.CreateServerRequest{
		Zone:             "fi-hel2",
		Title:            title,
		Hostname:         hostname,
		PasswordDelivery: request.PasswordDeliveryNone,
		StorageDevices: []upcloud.CreateServerStorageDevice{
			{
				Action:  upcloud.CreateServerStorageDeviceActionClone,
				Storage: "01000000-0000-4000-8000-000030060200",
				Title:   "disk1",
				Size:    30,
				Tier:    upcloud.StorageTierMaxIOPS,
			},
		},
		Networking: &request.CreateServerNetworking{
			Interfaces: []request.CreateServerInterface{
				{
					IPAddresses: []request.CreateServerIPAddress{
						{
							Family: upcloud.IPAddressFamilyIPv4,
						},
					},
					Type: upcloud.IPAddressAccessUtility,
				},
				{
					IPAddresses: []request.CreateServerIPAddress{
						{
							Family: upcloud.IPAddressFamilyIPv4,
						},
					},
					Type: upcloud.IPAddressAccessPublic,
				},
				{
					IPAddresses: []request.CreateServerIPAddress{
						{
							Family: upcloud.IPAddressFamilyIPv6,
						},
					},
					Type: upcloud.IPAddressAccessPublic,
				},
			},
		},
	}

	// Create the server and block until it has started
	serverDetails, err := svc.CreateServer(&createServerRequest)
	if err != nil {
		return nil, err
	}

	// Wait for the server to start
	serverDetails, err = svc.WaitForServerState(&request.WaitForServerStateRequest{
		UUID:         serverDetails.UUID,
		DesiredState: upcloud.ServerStateStarted,
		Timeout:      time.Minute * 15,
	})
	if err != nil {
		return nil, err
	}

	return serverDetails, nil
}

// Stops the specified server (forcibly)
func stopServer(svc *Service, uuid string) error {
	serverDetails, err := svc.StopServer(&request.StopServerRequest{
		UUID:     uuid,
		Timeout:  time.Minute * 15,
		StopType: request.ServerStopTypeHard,
	})
	if err != nil {
		return err
	}

	_, err = svc.WaitForServerState(&request.WaitForServerStateRequest{
		UUID:         serverDetails.UUID,
		DesiredState: upcloud.ServerStateStopped,
		Timeout:      time.Minute * 15,
	})
	if err != nil {
		return err
	}

	return nil
}

// Deletes the specified server
func deleteServer(svc *Service, uuid string) error {
	err := svc.DeleteServer(&request.DeleteServerRequest{
		UUID: uuid,
	})

	return err
}

// Deletes the specified server and storages
func deleteServerAndStorages(svc *Service, uuid string) error {
	err := svc.DeleteServerAndStorages(&request.DeleteServerAndStoragesRequest{
		UUID: uuid,
	})

	return err
}

// Creates a piece of storage and returns the details about it, panic if creation fails
func createStorage(svc *Service) (*upcloud.StorageDetails, error) {
	createStorageRequest := request.CreateStorageRequest{
		Tier:  upcloud.StorageTierMaxIOPS,
		Title: "Test storage",
		Size:  10,
		Zone:  "fi-hel2",
		BackupRule: &upcloud.BackupRule{
			Interval:  upcloud.BackupRuleIntervalDaily,
			Time:      "0430",
			Retention: 30,
		},
	}

	storageDetails, err := svc.CreateStorage(&createStorageRequest)

	if err != nil {
		return nil, err
	}

	return storageDetails, nil
}

// Deletes the specified storage
func deleteStorage(svc *Service, uuid string) error {
	err := svc.DeleteStorage(&request.DeleteStorageRequest{
		UUID: uuid,
	})

	return err
}

// deleteAllTags deletes all existing tags
func deleteAllTags(svc *Service) error {
	tags, err := svc.GetTags()
	if err != nil {
		return err
	}

	for _, tagDetails := range tags.Tags {
		err = svc.DeleteTag(&request.DeleteTagRequest{
			Name: tagDetails.Name,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// Waits for the specified storage to come online
func waitForStorageOnline(svc *Service, uuid string) error {
	_, err := svc.WaitForStorageState(&request.WaitForStorageStateRequest{
		UUID:         uuid,
		DesiredState: upcloud.StorageStateOnline,
		Timeout:      time.Minute * 15,
	})

	return err
}

// Returns the current UTC time with second precision (milliseconds truncated).
// This is the format we usually get from the UpCloud API.
func utcTimeWithSecondPrecision() (time.Time, error) {
	utc, err := time.LoadLocation("UTC")
	if err != nil {
		return time.Time{}, err
	}

	t := time.Now().In(utc).Truncate(time.Second)

	return t, err
}

// Handles the error by panicing, thus stopping the test execution
func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

// Reads the API username and password from the environment, panics if they are not available
func getCredentials() (string, string) {
	user := os.Getenv("UPCLOUD_GO_SDK_TEST_USER")
	password := os.Getenv("UPCLOUD_GO_SDK_TEST_PASSWORD")

	if user == "" || password == "" {
		panic("Unable to retrieve credentials from the environment, ensure UPCLOUD_GO_SDK_TEST_USER and UPCLOUD_GO_SDK_TEST_PASSWORD are exported")
	}

	return user, password
}
