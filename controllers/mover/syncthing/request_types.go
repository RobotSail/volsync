/*
Copyright 2021 The VolSync authors.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package syncthing

// syncthing config type

type SyncthingDevice struct {
	DeviceID                 string   `json:"deviceID"`
	Addresses                []string `json:"addresses"`
	Compression              string   `json:"compression"`
	CertName                 string   `json:"certName"`
	Introducer               bool     `json:"introducer"`
	SkipIntroductionRemovals bool     `json:"skipIntroductionRemovals"`
	IntroducedBy             string   `json:"introducedBy"`
	Paused                   bool     `json:"paused"`
	AllowedNetworks          []string `json:"allowedNetworks"`
	AutoAcceptFolders        bool     `json:"autoAcceptFolders"`
	MaxSendKbps              int      `json:"maxSendKbps"`
	MaxRecvKbps              int      `json:"maxRecvKbps"`
	IgnoredFolders           []string `json:"ignoredFolders"`
	MaxRequestKiB            int      `json:"maxRequestKiB"`
	Untrusted                bool     `json:"untrusted"`
	RemoteGUIPort            int      `json:"remoteGUIPort"`
}

type SyncthingSize struct {
	Value int    `json:"value"`
	Unit  string `json:"unit"`
}

type SyncthingVersioning struct {
	Type             string            `json:"type"`
	Params           map[string]string `json:"params"`
	CleanupIntervalS int               `json:"cleanupIntervalS"`
	FsPath           string            `json:"fsPath"`
	FsType           string            `json:"fsType"`
}

type SyncthingFolder struct {
	ID                    string              `json:"id"`
	Label                 string              `json:"label"`
	FilesystemType        string              `json:"filesystemType"`
	Path                  string              `json:"path"`
	Type                  string              `json:"type"`
	Devices               []SyncthingDevice   `json:"devices"`
	RescanIntervalS       int                 `json:"rescanIntervalS"`
	FsWatcherEnabled      bool                `json:"fsWatcherEnabled"`
	FsWatcherDelayS       int                 `json:"fsWatcherDelayS"`
	IgnorePerms           bool                `json:"ignorePerms"`
	AutoNormalize         bool                `json:"autoNormalize"`
	MinDiskFree           SyncthingSize       `json:"minDiskFree"`
	Versioning            SyncthingVersioning `json:"versioning"`
	Copiers               int                 `json:"copiers"`
	PullerMaxPendingKiB   int                 `json:"pullerMaxPendingKiB"`
	Hashers               int                 `json:"hashers"`
	Order                 string              `json:"order"`
	IgnoreDelete          bool                `json:"ignoreDelete"`
	ScanProgressIntervalS int                 `json:"scanProgressIntervalS"`
}

type SyncthingOptions struct {
	ListenAddresses         []string      `json:"listenAddresses"`
	GlobalAnnServers        []string      `json:"globalAnnServers"`
	GlobalAnnEnabled        bool          `json:"globalAnnEnabled"`
	NATEnabled              bool          `json:"natEnabled"`
	NATLeaseMinutes         int           `json:"natLeaseMinutes"`
	NATRenewalMinutes       int           `json:"natRenewalMinutes"`
	NATTimeout              int           `json:"natTimeout"`
	URAccepted              int           `json:"urAccepted"`
	URUniqueID              string        `json:"urUniqueID"`
	URDeclined              int           `json:"urDeclined"`
	URPaused                int           `json:"urPaused"`
	RestartOnWakeup         bool          `json:"restartOnWakeup"`
	AutoUpgradeIntervalH    int           `json:"autoUpgradeIntervalH"`
	KeepTemporariesH        int           `json:"keepTemporariesH"`
	CacheIgnoredFiles       bool          `json:"cacheIgnoredFiles"`
	ProgressUpdateIntervalS int           `json:"progressUpdateIntervalS"`
	SymlinksEnabled         bool          `json:"symlinksEnabled"`
	LimitBandwidthInLan     bool          `json:"limitBandwidthInLan"`
	MinHomeDiskFree         SyncthingSize `json:"minHomeDiskFree"`
	URURL                   string        `json:"urURL"`
	URInitialDelayS         int           `json:"urInitialDelayS"`
	URPostInsecurely        bool          `json:"urPostInsecurely"`
	URInitialTimeoutS       int           `json:"urInitialTimeoutS"`
	URReconnectIntervalS    int           `json:"urReconnectIntervalS"`
	OverwriteRemoteDevice   bool          `json:"overwriteRemoteDevice"`
	TempIndexMinBlocks      int           `json:"tempIndexMinBlocks"`
	UnackedNotificationIDs  []string      `json:"unackedNotificationIDs"`
	DefaultFolderIgnores    []string      `json:"defaultFolderIgnores"`
	DefaultFolderExcludes   []string      `json:"defaultFolderExcludes"`
	DefaultFolderIncludes   []string      `json:"defaultFolderIncludes"`
	OverwriteRemoteTimeoutS int           `json:"overwriteRemoteTimeoutS"`
	OverwriteRemoteIgnores  bool          `json:"overwriteRemoteIgnores"`
	OverwriteRemoteExcludes bool          `json:"overwriteRemoteExcludes"`
	OverwriteRemoteIncludes bool          `json:"overwriteRemoteIncludes"`
	DefaultReadOnly         bool          `json:"defaultReadOnly"`
	IgnoredFiles            []string      `json:"ignoredFiles"`
	MaxConflicts            int           `json:"maxConflicts"`
	MaxKnownDeleted         int           `json:"maxKnownDeleted"`
	MaxChangeKiB            int           `json:"maxChangeKiB"`
	MaxSendKiB              int           `json:"maxSendKiB"`
	MaxRecvKiB              int           `json:"maxRecvKiB"`
	MaxRequestKiB           int           `json:"maxRequestKiB"`
	ReconnectIntervalS      int           `json:"reconnectIntervalS"`
	ReconnectFailureCount   int           `json:"reconnectFailureCount"`
	ReconnectBackoffMinS    int           `json:"reconnectBackoffMinS"`
	ReconnectBackoffMaxS    int           `json:"reconnectBackoffMaxS"`
	ReconnectBackoffMaxExp  int           `json:"reconnectBackoffMaxExp"`
	ReconnectBackoffJitter  bool          `json:"reconnectBackoffJitter"`
	ReconnectBackoffMult    int           `json:"reconnectBackoffMult"`
	ReconnectBackoffExp     int           `json:"reconnectBackoffExp"`
	ReconnectBackoffFloor   int           `json:"reconnectBackoffFloor"`
	ReconnectBackoffCeiling int           `json:"reconnectBackoffCeiling"`
	ReconnectBackoffMax     int           `json:"reconnectBackoffMax"`
	ReconnectBackoffMin     int           `json:"reconnectBackoffMin"`
}

type SyncthingConfig struct {
	Version              string            `json:"version"`
	Folders              []SyncthingFolder `json:"folders"`
	Devices              []SyncthingDevice `json:"devices"`
	Options              SyncthingOptions  `json:"options"`
	Device               SyncthingDevice   `json:"device"`
	RemoteIgnoredDevices []string          `json:"remoteIgnoredDevices"`
	Defaults             SyncthingFolder   `json:"defaults"`
}
