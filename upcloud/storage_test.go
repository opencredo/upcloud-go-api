package upcloud

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUnmarshalStorage tests that Storages and Storage struct are unmarshaled correctly
func TestUnmarshalStorage(t *testing.T) {
	originalJSON := `
{
  "storages": {
    "storage": [
      {
        "access": "private",
        "license": 0,
        "size": 10,
        "state": "online",
        "tier": "hdd",
        "title": "Operating system disk",
        "type": "normal",
        "uuid": "01eff7ad-168e-413e-83b0-054f6a28fa23",
        "zone": "uk-lon1"
      },
      {
        "access" : "private",
        "created" : "2019-09-17T14:34:43Z",
        "license" : 0,
        "origin" : "01eff7ad-168e-413e-83b0-054f6a28fa23",
        "size" : 10,
        "state" : "online",
        "title" : "On demand backup",
        "type" : "backup",
        "uuid" : "01287ad1-496c-4b5f-bb67-0fc2e3494740",
        "zone" : "uk-lon1"
      },
      {
        "access": "private",
        "license": 0,
        "part_of_plan": "yes",
        "size": 50,
        "state": "online",
        "tier": "maxiops",
        "title": "Databases",
        "type": "normal",
        "uuid": "01f3286c-a5ea-4670-8121-d0b9767d625b",
        "zone": "fi-hel1"
      }
    ]
  }
}`

	storages := Storages{}
	err := json.Unmarshal([]byte(originalJSON), &storages)
	assert.NoError(t, err)
	assert.Len(t, storages.Storages, 3)

	testData := []Storage{
		{
			Access:  StorageAccessPrivate,
			License: 0.0,
			Size:    10,
			State:   StorageStateOnline,
			Tier:    StorageTierHDD,
			Title:   "Operating system disk",
			Type:    StorageTypeNormal,
			UUID:    "01eff7ad-168e-413e-83b0-054f6a28fa23",
			Zone:    "uk-lon1",
		},
		{
			Access:  StorageAccessPrivate,
			License: 0.0,
			Origin:  "01eff7ad-168e-413e-83b0-054f6a28fa23",
			Size:    10,
			State:   StorageStateOnline,
			Title:   "On demand backup",
			Type:    StorageTypeBackup,
			UUID:    "01287ad1-496c-4b5f-bb67-0fc2e3494740",
			Zone:    "uk-lon1",
		},
		{
			Access:     StorageAccessPrivate,
			License:    0.0,
			PartOfPlan: "yes",
			Size:       50,
			State:      StorageStateOnline,
			Tier:       StorageTierMaxIOPS,
			Title:      "Databases",
			Type:       StorageTypeNormal,
			UUID:       "01f3286c-a5ea-4670-8121-d0b9767d625b",
			Zone:       "fi-hel1",
		},
	}

	for i, data := range testData {
		storage := storages.Storages[i]
		assert.Equal(t, data.Access, storage.Access)
		assert.Equal(t, data.License, storage.License)
		assert.Equal(t, data.Size, storage.Size)
		assert.Equal(t, data.Title, storage.Title)
		assert.Equal(t, data.Type, storage.Type)
		assert.Equal(t, data.UUID, storage.UUID)
		assert.Equal(t, data.PartOfPlan, storage.PartOfPlan)
		assert.Equal(t, data.State, storage.State)
		assert.Equal(t, data.Tier, storage.Tier)
		assert.Equal(t, data.Zone, storage.Zone)
	}
}

// TestUnmarshalStorageDetails tests that StorageDetails struct is unmarshaled correctly
func TestUnmarshalStorageDetails(t *testing.T) {
	originalJSON := `
	{
		"storage": {
		  "access": "private",
		  "backup_rule": {
			  "interval": "daily",
			  "time": "0400",
			  "retention": "1"
		  },
		  "backups": {
			"backup": [
              "37c96670-9c02-4d5d-8f60-291d38f9a80c",
              "ecfda9f2-e071-4bbb-b38f-079ed26eb32a"
			]
		  },
		  "license": 0,
		  "servers": {
			"server": [
			  "00798b85-efdc-41ca-8021-f6ef457b8531"
			]
		  },
		  "size": 10,
		  "state": "online",
		  "tier": "maxiops",
		  "title": "Operating system disk",
		  "type": "normal",
		  "uuid": "01d4fcd4-e446-433b-8a9c-551a1284952e",
		  "zone": "fi-hel1"
		}
	  }
	`

	storageDeviceDetails := StorageDetails{}
	err := json.Unmarshal([]byte(originalJSON), &storageDeviceDetails)
	assert.NoError(t, err)

	assert.Equal(t, StorageAccessPrivate, storageDeviceDetails.Access)
	assert.Equal(t, 0.0, storageDeviceDetails.License)
	assert.Equal(t, 10, storageDeviceDetails.Size)
	assert.Equal(t, StorageStateOnline, storageDeviceDetails.State)
	assert.Equal(t, StorageTierMaxIOPS, storageDeviceDetails.Tier)
	assert.Equal(t, "Operating system disk", storageDeviceDetails.Title)
	assert.Equal(t, StorageTypeNormal, storageDeviceDetails.Type)
	assert.Equal(t, "01d4fcd4-e446-433b-8a9c-551a1284952e", storageDeviceDetails.UUID)
	assert.Equal(t, "fi-hel1", storageDeviceDetails.Zone)

	assert.Equal(t, BackupRuleIntervalDaily, storageDeviceDetails.BackupRule.Interval)
	assert.Equal(t, 1, storageDeviceDetails.BackupRule.Retention)
	assert.Equal(t, "0400", storageDeviceDetails.BackupRule.Time)

	assert.Equal(t, 2, len(storageDeviceDetails.BackupUUIDs))
	assert.Equal(t, "37c96670-9c02-4d5d-8f60-291d38f9a80c", storageDeviceDetails.BackupUUIDs[0])
	assert.Equal(t, "ecfda9f2-e071-4bbb-b38f-079ed26eb32a", storageDeviceDetails.BackupUUIDs[1])

	assert.Equal(t, 1, len(storageDeviceDetails.ServerUUIDs))
	assert.Equal(t, "00798b85-efdc-41ca-8021-f6ef457b8531", storageDeviceDetails.ServerUUIDs[0])
}

//
// Are the following tests needed?
//
/*
// TestUnmarshalServerStorageDevice tests that ServerStorageDevice objects are properly unmarshaled
func TestUnmarshalServerStorageDevice(t *testing.T) {
	originalJSON := `
		{
			"storage_device": {
				"address": "virtio:0",
				"part_of_plan": "yes",
				"storage": "01c8df16-d1c6-4223-9bfc-d3c06b208c88",
				"storage_size": 30,
				"storage_title": "test-disk0",
				"type": "disk"
			}
		}
	`

	storageDevice := ServerStorageDevice{}
	err := json.Unmarshal([]byte(originalJSON), &storageDevice)

	assert.Nil(t, err)
	assert.Equal(t, "virtio:0", storageDevice.Address)
	assert.Equal(t, "yes", storageDevice.PartOfPlan)
	assert.Equal(t, "01c8df16-d1c6-4223-9bfc-d3c06b208c88", storageDevice.UUID)
	assert.Equal(t, 30, storageDevice.Size)
	assert.Equal(t, "test-disk0", storageDevice.Title)
	assert.Equal(t, StorageTypeDisk, storageDevice.Type)
}

// TestMarshalCreateStorageDevice tests that CreateStorageDevice objects are correctly marshaled. We don't need to
// test unmarshaling because these data structures are never returned from the API.
func TestMarshalCreateStorageDevice(t *testing.T) {
	storage := CreateServerStorageDevice{
		Action:  CreateServerStorageDeviceActionClone,
		Storage: "01000000-0000-4000-8000-000030060200",
		Title:   "disk1",
		Size:    30,
		Tier:    StorageTierMaxIOPS,
	}

	expectedXML := "<storage_device><action>clone</action><storage>01000000-0000-4000-8000-000030060200</storage><title>disk1</title><size>30</size><tier>maxiops</tier></storage_device>"

	actualXML, err := xml.Marshal(storage)
	assert.Nil(t, err)
	assert.Equal(t, expectedXML, string(actualXML))
}

// TestMarshalBackupRule tests that BackupRule objects are properly marshaled
func TestMarshalBackupRule(t *testing.T) {
	backupRule := BackupRule{
		Interval:  BackupRuleIntervalDaily,
		Time:      "0430",
		Retention: 30,
	}

	ruleXML, err := xml.Marshal(backupRule)
	assert.Nil(t, err)
	assert.Equal(t, "<backup_rule><interval>daily</interval><time>0430</time><retention>30</retention></backup_rule>", string(ruleXML))
}

// TestUnmarshalBackupRule tests that BackupRule objects are properly unmarshaled
func TestUnmarshalBackupRule(t *testing.T) {
	originalXML := "<backup_rule><interval>daily</interval><time>0430</time><retention>30</retention></backup_rule>"

	backupRule := BackupRule{}
	err := xml.Unmarshal([]byte(originalXML), &backupRule)
	assert.Nil(t, err)
	assert.Equal(t, BackupRuleIntervalDaily, backupRule.Interval)
	assert.Equal(t, "0430", backupRule.Time)
	assert.Equal(t, 30, backupRule.Retention)
}
*/
