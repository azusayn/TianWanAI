package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gopkg.in/yaml.v3"
)

// Config represents the structure of config.yaml
type Config struct {
	Tianwan1    []string `yaml:"tianwan1"`
	Tianwan2    []string `yaml:"tianwan2"`
	AlertServer string   `yaml:"alert_server"`
	ExcelPath   string   `yaml:"excel_path"`
	FilterMap   []string `yaml:"filter_map"`
}

// LoadConfig loads configuration from YAML file
func LoadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// read from excel file.
type ModelType string

const (
	ModelHelmet      ModelType = "安全帽"
	ModelMouse       ModelType = "老鼠"
	ModelShortSleeve ModelType = "短袖"
	ModelPonding     ModelType = "积水"
	ModelFall        ModelType = "倒地"
	ModelSafetyBelt  ModelType = "安全带"
	ModelCigar       ModelType = "吸烟"
	ModelGesture     ModelType = "手势"
	ModelSmoke       ModelType = "烟雾"
	ModelFire        ModelType = "火焰"
)

func convertNamesToUrl(m ModelType) string {
	switch m {
	case ModelHelmet:
		return "helmet"
	case ModelMouse:
		return "mouse"
	case ModelShortSleeve:
		return "tshirt"
	case ModelPonding:
		return "ponding"
	case ModelFall:
		return "fall"
	case ModelSafetyBelt:
		return "safetybelt"
	case ModelCigar:
		return "cigar"
	case ModelGesture:
		return "gesture"
	case ModelSmoke:
		return "smoke"
	case ModelFire:
		return "fire"
	}
	return ""
}

type CameraInfo struct {
	DeviceName string
	RtspURL    string
	Models     []string
}

func ReadCameraInfoFromExcel(filePath string) ([]CameraInfo, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	var cameras []CameraInfo
	defaultModels := []string{"cigar", "gesture", "smoke", "fire"}

	for i, row := range rows {
		if i == 0 {
			continue
		}

		if len(row) < 12 {
			continue
		}

		deviceName := row[7]
		rtspURL := row[10]
		modelsStr := row[11]

		if deviceName == "" || rtspURL == "" {
			continue
		}

		var models []string
		modelMap := make(map[string]bool)

		if modelsStr != "" {
			modelList := strings.Split(modelsStr, "、")
			for _, m := range modelList {
				m = strings.TrimSpace(m)
				if m != "" {
					u := convertNamesToUrl(ModelType(m))
					models = append(models, u)
					modelMap[u] = true
				}
			}
		}

		for _, dm := range defaultModels {
			if !modelMap[dm] {
				models = append(models, dm)
			}
		}

		camera := CameraInfo{
			DeviceName: deviceName,
			RtspURL:    rtspURL,
			Models:     models,
		}

		cameras = append(cameras, camera)
	}

	return cameras, nil
}

func readCamerasFromFile(filePath string, filterList []string) []CameraInfo {
	cameras, err := ReadCameraInfoFromExcel(filePath)
	if err != nil {
		slog.Error("failed to read cameras' info from excel", "error", err)
		return nil
	}

	// 将 filterList 转换为 map 以便快速查找
	filterMap := make(map[string]bool)
	for _, device := range filterList {
		filterMap[device] = true
	}

	var filteredCameras []CameraInfo
	for _, c := range cameras {
		if filterMap[c.DeviceName] {
			continue
		}
		filteredCameras = append(filteredCameras, c)
	}
	return filteredCameras
}

type InferenceServer struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	ModelType   string    `json:"model_type"`
	Description string    `json:"description,omitempty"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// InferenceServerBinding represents a binding between camera and inference server with threshold
type InferenceServerBinding struct {
	ServerID     string  `json:"server_id"`
	Threshold    float64 `json:"threshold"`
	MaxThreshold float64 `json:"max_threshold"`
}
type CameraConfig struct {
	ID                      string                   `json:"id"`
	Name                    string                   `json:"name"`
	RTSPUrl                 string                   `json:"rtsp_url"`
	InferenceServerBindings []InferenceServerBinding `json:"inference_server_bindings,omitempty"`
	Enabled                 bool                     `json:"enabled"`
	Running                 bool                     `json:"running"`
	CreatedAt               time.Time                `json:"created_at"`
	UpdatedAt               time.Time                `json:"updated_at"`
}

// AlertServerConfig represents the global alert server configuration
type AlertServerConfig struct {
	URL       string    `json:"url"`
	Enabled   bool      `json:"enabled"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DataStore struct {
	Cameras          map[string]*CameraConfig    `json:"cameras"`
	InferenceServers map[string]*InferenceServer `json:"inference_servers"`
	AlertServer      *AlertServerConfig          `json:"alert_server,omitempty"`
}

// TODO: move these functions to 'common' package
func GetCurrentTime() string {
	return time.Now().Format(time.RFC3339Nano)
}
func GenerateUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

type AvailableServer struct {
	ID        string
	modelType string
}

func findAvailableServerId(servers []AvailableServer, modelType string) string {
	for _, s := range servers {
		if s.modelType == modelType {
			return s.ID
		}
	}
	return ""
}

func main() {
	// cli args.
	configPath := flag.String("c", "config.yaml", "配置文件路径 (默认: config.yaml)")
	outputPath := flag.String("o", "tianwan_config.json", "输出文件路径 (默认: tianwan_config.json)")
	flag.Parse()

	// load configuration from YAML file
	config, err := LoadConfig(*configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err, "path", *configPath)
		return
	}
	slog.Info("loading config from: " + *configPath)
	serverConfig := DataStore{
		Cameras:          make(map[string]*CameraConfig),
		InferenceServers: make(map[string]*InferenceServer),
		AlertServer: &AlertServerConfig{
			URL:       config.AlertServer,
			Enabled:   false,
			UpdatedAt: time.Now(),
		},
	}

	// generate 'inference_servers' section
	aServerModelTypes := []string{
		"mouse",
		"ponding",
		"cigar",
		"gesture",
		"fall",
		"tshirt",
		"helmet",
		"smoke",
		"fire",
	}

	allAvailableServers := make(map[string][]AvailableServer)
	for i, addr := range config.Tianwan1 {
		var availableServerByIp []AvailableServer
		for _, t := range aServerModelTypes {
			id := fmt.Sprintf("inf_%s_%s", t, GenerateUUID())
			// TODO: it should be noted that both the fire and the smoke
			// models are using the same url: /smoke
			modelUrl := t
			if t == "fire" {
				modelUrl = "smoke"
			}
			serverConfig.InferenceServers[id] = &InferenceServer{
				ID:        id,
				Name:      fmt.Sprintf("%s%d", t, i+1),
				URL:       fmt.Sprintf("http://%s/%s", addr, modelUrl),
				ModelType: t,
				Enabled:   true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			availableServerByIp = append(availableServerByIp, AvailableServer{
				ID:        id,
				modelType: t,
			})
		}
		allAvailableServers[addr] = availableServerByIp
	}

	for i, addr := range config.Tianwan2 {
		var availableServerByIp []AvailableServer

		t := "safetybelt"
		id := fmt.Sprintf("inf_%s_%s", t, GenerateUUID())
		serverConfig.InferenceServers[id] = &InferenceServer{
			ID:        id,
			Name:      fmt.Sprintf("%s%d", t, i+1),
			URL:       fmt.Sprintf("http://%s/%s", addr, t),
			ModelType: t,
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		availableServerByIp = append(availableServerByIp, AvailableServer{
			ID:        id,
			modelType: t,
		})
		allAvailableServers[addr] = availableServerByIp
	}

	// generate 'cameras' section
	ia, ib := 0, 0
	for _, c := range readCamerasFromFile(config.ExcelPath, config.FilterMap) {
		// basic info
		cid := fmt.Sprintf("cam_%s", GenerateUUID())
		camera := CameraConfig{
			ID:        cid,
			Name:      c.DeviceName,
			RTSPUrl:   c.RtspURL,
			Enabled:   true,
			Running:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		// bindings
		for _, m := range c.Models {
			binding := InferenceServerBinding{
				Threshold:    0.5,
				MaxThreshold: 0,
			}
			// arrange to type B server
			if m == "safetybelt" {
				ip := config.Tianwan2[ib]
				binding.ServerID = findAvailableServerId(allAvailableServers[ip], m)
				ib = (ib + 1) % len(config.Tianwan2)
			} else {
				// arrange to type A server
				ip := config.Tianwan1[ia]
				binding.ServerID = findAvailableServerId(allAvailableServers[ip], m)
				ia = (ia + 1) % len(config.Tianwan1)
			}
			camera.InferenceServerBindings = append(camera.InferenceServerBindings, binding)
		}
		serverConfig.Cameras[cid] = &camera
	}

	// write server config to local file
	data, err := json.MarshalIndent(serverConfig, "", "  ")
	if err != nil {
		slog.Error("failed to marshal server config to .json file")
		return
	}

	if err := os.WriteFile(*outputPath, data, 0644); err != nil {
		slog.Error("failed to write json data to local file")
		return
	}

	slog.Info("config.json generated successfully")
}
