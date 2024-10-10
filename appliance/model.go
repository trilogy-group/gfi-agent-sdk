/**
 * This software and associated documentation files (the “Software”),
 * including GFI AppManager, is the property of GFI USA, LLC and its affiliates.
 * No part of the Software may be copied, modified, distributed, sold, or otherwise
 * used except as expressly permitted by the terms of the software license agreement.
 */

package appliance

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/trilogy-group/gfi-agent-sdk/utils"
)

type Status byte

type JSONTime time.Time

const (
	NotRegistered Status = iota
	Registering
	Registered
)

type Appliance interface {
	// return id of the appliance
	Id() string

	// return type of the appliance like Kerio-Connect, Languard, etc
	Type() string

	// returns path to the dir
	Dir() string

	// Create a special user acccount with the appliance
	SignUp() error

	// Logout
	RemoveAccount() error

	// returns true if already signed up with appliance else false
	SignUpStatus() bool

	// return api endpoint of the appliance like http://localhost:4040
	ConnectInfo() (string, map[string]string, error)

	// true if appliance is up and can be connected else false
	ConnectionStatus() bool

	// true if appliance is registered
	RegistrationStatus() Status

	// return metrics
	Insights() []*Insight

	HasLicense() bool

	// return notifications
	Notifications() []*Notification

	// return appliance info
	Info() (*ApplianceInfo, error)

	// this should remove config
	Remove() error

	// return private key
	PrivateKey() string

	// return public key
	PublicKey() string

	// return account password
	Password() string

	// check for registration status at equal interval
	// this gets invoked only after register endpoint is invoked
	UpdateRegistrationStatus(status Status) error

	UpdateSignUpStatus(status bool) error

	UpdateSerialNumber(serialNumber string) error

	GetHardwareInfo() error

	// return hardware serial number
	SerialNumber() string

	SaveConfigs() error

	ReloadConfigs() error

	// this function is called after agent successfully publishes insights to kinesis
	InsightsPublished(insights []*Insight)

	// dir where the appliance is installed currently
	ServerDir() string

	Clone() Appliance

	IsSelfManagedInstallation() bool

	GetAppManagerUIBaseUrl() (string, error)

	GetApiServerBaseUrl() (string, error)

	ModifyApplianceResponse(r *http.Response) error

	HandleByLocalApi(r *http.Request) (*int, interface{})

	CheckUpdate() error
}

type ApplianceInfo struct {
	User              int64    `json:"User"`
	LicensedUsers     int64    `json:"LicensedUsers"`
	Type              string   `json:"type"`
	ApplianceId       string   `json:"applianceId"`
	Uptime            float64  `json:"Uptime"`
	Version           string   `json:"Version"`
	Expiry            string   `json:"Expiry"`
	Timestamp         JSONTime `json:"timestamp"`
	AdminUiUrl        string   `json:"AdminUiUrl"`
	ProductLicenseKey string   `json:"ProductLicenseKey"`
	AgentVersion      string   `json:"AgentVersion"`
}

func (I *ApplianceInfo) String() string {
	return fmt.Sprintf("{version=%s, expiry=%s, uptime=%.2f, user=%d, licensedUsers=%d, appliance=%s, adminUiUrl=%s, productLicenseKey=%s agentVersion=%s}", I.Version, I.Expiry, I.Uptime, I.User, I.LicensedUsers, I.ApplianceId, I.AdminUiUrl, I.ProductLicenseKey, I.AgentVersion)
}

type Dimension struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Metric struct {
	Unit  string  `json:"unit"`
	Value float64 `json:"value"`
}

type Insight struct {
	Name        string       `json:"name"`
	Type        string       `json:"type"`
	ApplianceId string       `json:"applianceId"`
	Metric      *Metric      `json:"metric"`
	Dimensions  []*Dimension `json:"dimensions"`
	Timestamp   JSONTime     `json:"timestamp"`
	RepeatTime  time.Time    `json:"repeatTime"`
}

type Notification struct {
	Type        string       `json:"type"`
	Name        string       `json:"name"`
	Severity    string       `json:"severity"`
	Message     string       `json:"message"`
	Dimensions  []*Dimension `json:"dimensions"`
	Timestamp   JSONTime     `json:"timestamp"`
	ApplianceId string       `json:"applianceId"`
}

type Config struct {
	// unique identifier for the appliance
	Id string `toml:"id"`

	// type of the appliance
	Type string `toml:"type"`

	// path to the installation dir for the appliance
	ServerDir string `toml:"serverDir"`

	// private and public key for making calls to the app manager backend
	PrivateKey string `toml:"privateKey"`
	PublicKey  string `toml:"publicKey"`

	// username and password to authenticate with the appliance
	Username          string `toml:"username"`
	Password          string `toml:"password"`
	PasswordEncrypted string `toml:"passwordEncrypted"`

	// true if agent has signed up with the appliance
	SignUpStatus bool `toml:"signupStatus"`

	// current status of registration [REGISTERED, NOT_REGISTERED, REGISTERING]
	RegistrationStatus Status `toml:"registrationStatus"`

	// Hardware box serial number
	SerialNumber string `toml:"serialNumber"`

	AgentSupportedVersion string `toml:"agentSupportedVersion"`
}

type ConfigManager struct {
	Name string
	Type string
	Path string
}

func NewConfigManager(name string, configType string, path string) *ConfigManager {
	return &ConfigManager{
		Name: name,
		Type: configType,
		Path: path,
	}
}

func (C *ConfigManager) FullPath() string {
	return filepath.Join(C.Path, C.Name+"."+C.Type)
}

func (C *ConfigManager) Unmarshal(config interface{}) error {
	_, err := toml.DecodeFile(C.FullPath(), config)
	return err
}

type ConfigWithPassword interface {
	GetPassword() string
	SetPassword(value string)
	GetPasswordEncrypted() string
	SetPasswordEncrypted(value string)
}

var mutex sync.Mutex
var encryptionKey = [32]byte{
	188, 74, 186, 252, 160, 2, 151, 205, 140, 118, 207, 210, 51, 108, 180, 149,
	213, 181, 104, 174, 206, 5, 190, 24, 153, 29, 195, 153, 235, 96, 250, 163,
}

func Encrypt(plaintext []byte, key *[32]byte) (ciphertext []byte, err error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func Decrypt(ciphertext []byte, key *[32]byte) (plaintext []byte, err error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	return gcm.Open(nil,
		ciphertext[:gcm.NonceSize()],
		ciphertext[gcm.NonceSize():],
		nil,
	)
}

func Base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func Base64Decode(str string) (string, bool) {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", true
	}
	return string(data), false
}

func EncryptPassword(password string) (string, error) {
	encrypted, error := Encrypt([]byte(password), &encryptionKey)
	if error != nil {
		return "", errors.New("failed to encrypt config data: " + error.Error())
	}

	return Base64Encode(string(encrypted)), nil
}

func DecryptPassword(password string) (string, error) {
	decoded, hasError := Base64Decode(password)
	if hasError {
		return "", errors.New("failed to decode config data")
	}

	decrypted, error := Decrypt([]byte(decoded), &encryptionKey)
	if error != nil {
		return "", errors.New("failed to further decode config data")
	}

	result := string(decrypted)
	return result, nil
}

func (C *ConfigManager) LoadApplianceConfig(config ConfigWithPassword) error {
	err := C.Unmarshal(config)
	if err != nil {
		return err
	}

	if config != nil {
		if len(config.GetPassword()) > 0 {
			err = C.SaveApplianceConfig(config)
			if err != nil {
				return err
			}
		}

		encrypted := config.GetPasswordEncrypted()
		if len(encrypted) > 0 {
			decrypted, err := DecryptPassword(encrypted)
			if err != nil {
				return err
			}
			config.SetPassword(decrypted)
		}
	}

	return nil
}

func (C *ConfigManager) Remove() error {
	return utils.FS.RemoveFile(C.FullPath())
}

func (C *ConfigManager) SaveApplianceConfig(config ConfigWithPassword) error {
	if config != nil {
		encrypted, err := EncryptPassword(config.GetPassword())
		if err != nil {
			return errors.New("failed to save config: " + err.Error())
		}

		config.SetPasswordEncrypted(encrypted)
		config.SetPassword("")
	}

	return C.Save(config)
}

func (C *ConfigManager) Save(config interface{}) error {
	mutex.Lock()
	dir := filepath.Dir(C.FullPath())
	if !utils.FS.CreateDir(dir) {
		return fmt.Errorf("could not create dir: %s", dir)
	}
	f, err := os.Create(C.FullPath())
	if err != nil {
		mutex.Unlock()
		return err
	}

	if err := toml.NewEncoder(f).Encode(config); err != nil {
		mutex.Unlock()
		return err
	}
	if err := f.Close(); err != nil {
		mutex.Unlock()
		return err
	}
	mutex.Unlock()
	return nil
}

const (
	ApplianceConfigName = "config"
	ApplianceConfigType = "toml"
	CommonConfigName    = "config"
	CommonConfigType    = "toml"
)

func (C *Config) Save(dir string) error {
	cfgMgr := NewConfigManager(ApplianceConfigName, ApplianceConfigType, dir)
	configToSave := *C
	return cfgMgr.SaveApplianceConfig(&configToSave)
}

func (C *Config) Reload(dir string) error {
	cfgMgr := NewConfigManager(ApplianceConfigName, ApplianceConfigType, dir)
	return cfgMgr.LoadApplianceConfig(C)
}

func (C *Config) GetPassword() string {
	return C.Password
}

func (C *Config) SetPassword(value string) {
	C.Password = value
}

func (C *Config) GetPasswordEncrypted() string {
	return C.PasswordEncrypted
}

func (C *Config) SetPasswordEncrypted(value string) {
	C.PasswordEncrypted = value
}

func (C *CommonConfig) Save(dir string) error {
	cfgMgr := NewConfigManager(CommonConfigName, CommonConfigType, dir)
	return cfgMgr.Save(C)
}

// GetSupportedVersion returns the supported version of the appliance
func (C *Config) GetAgentSupportedVersion() string {
	return C.AgentSupportedVersion
}

// SetSupportedVersion sets the supported version of the appliance
func (C *Config) SetAgentSupportedVersion(value string) {
	C.AgentSupportedVersion = value
}

type CommonConfig struct {
	// unique identifier for the machine
	MachineId string `toml:"machineId"`
	// enable or disable agent auto update
	EnableUpdate *bool `toml:"enableUpdate"`
	EnableSentry bool  `toml:"enableSentry"`
}

type MetricInsight struct {
	Name       string
	Metric     *Metric
	Dimensions []*Dimension
	Timestamp  time.Time
}

func (M *Metric) String() string {
	return fmt.Sprintf("{unit=%s, value=%.2f}", M.Unit, M.Value)
}

func (D *Dimension) String() string {
	return fmt.Sprintf("{name=%s, value=%s}", D.Name, D.Value)
}
func (I *MetricInsight) String() string {
	return fmt.Sprintf("{name=%s, metric=%s, dims=%s}", I.Name, I.Metric, I.Dimensions)
}

func (I *Insight) String() string {
	return fmt.Sprintf("{name=%s, metric=%s, dims=%s, appliance=%s}", I.Name, I.Metric, I.Dimensions, I.ApplianceId)
}

func (N *Notification) String() string {
	return fmt.Sprintf("{name=%s, severity=%s, msg='%s', dims=%s, appliance=%s}", N.Name, N.Severity, N.Message, N.Dimensions, N.ApplianceId)
}
