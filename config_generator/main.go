package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

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

// Usage:
func readCamerasFromFile(filePath string) []CameraInfo {
	cameras, err := ReadCameraInfoFromExcel(filePath)
	if err != nil {
		slog.Error("failed to read cameras' info from excel")
	}

	// filter map
	filterMap := map[string]bool{
		"5M1DTW102TV": true,
		"6M2DTW101TV": true,
		"6M2DTW104TV": true,
		"5M0DTW126TV": true,
		"5M0DTW134TV": true,
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

// generate config file.
type AlertServer struct {
	Url       string `json:"url"`
	Enabled   bool   `json:"enabled"`
	UpdatedAt string `json:"updated_at"`
}

type InferenceServer struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Url       string `json:"url"`
	ModelType string `json:"model_type"`
	Enabled   bool   `json:"enabled"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type InferenceServers map[string]InferenceServer

type InferenceServerBinding struct {
	ServerId     string  `json:"server_id"`
	Threshold    float32 `json:"threshold"`
	MaxThreshold float32 `json:"max_threshold"`
}

type Camera struct {
	Id                      string                   `json:"id"`
	Name                    string                   `json:"name"`
	RtspUrl                 string                   `json:"rtsp_url"`
	InferenceServerBindings []InferenceServerBinding `json:"inference_server_bindings"`
	Enabled                 bool                     `json:"enabled"`
	Running                 bool                     `json:"running"`
	CreatedAt               string                   `json:"created_at"`
	UpdatedAt               string                   `json:"updated_at"`
}

type Cameras map[string]Camera

type ServerConfig struct {
	Cameras          Cameras          `json:"cameras"`
	InferenceServers InferenceServers `json:"inference_servers"`
	AlertServer      AlertServer      `json:"alert_server"`
}

// TODO: move these functions to 'common' package
func GetCurrentTime() string {
	return time.Now().Format(time.RFC3339Nano)
}
func GenerateUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

func main() {
	serverConfig := ServerConfig{
		Cameras:          make(Cameras),
		InferenceServers: make(InferenceServers),
		AlertServer: AlertServer{
			Url:       "http://192.168.1.82:80/api/account/ai/alarm",
			Enabled:   false,
			UpdatedAt: GetCurrentTime(),
		},
	}

	// generate 'inference_servers' section
	availableAServerAddrs := []string{
		"192.168.1.86:8901",
		"192.168.1.86:8902",
		"192.168.1.86:8903",
		"192.168.1.86:8904",
		"192.168.1.86:8905",
		"192.168.1.86:8906",
		"192.168.1.86:8907",
		"192.168.1.86:8908",
		"192.168.1.86:8909",
		"192.168.1.86:8910",
		"192.168.1.86:8911",
		"192.168.1.86:8912",
		"192.168.1.86:8913",
		"192.168.1.86:8914",
	}
	availableBServerAddrs := []string{
		"192.168.1.86:8915",
		"192.168.1.86:8916",
		"192.168.1.86:8917",
		"192.168.1.86:8918",
	}
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

	var availableAServerIds []string
	var availableBServerIds []string
	for i, addr := range availableAServerAddrs {
		for _, t := range aServerModelTypes {
			now := GetCurrentTime()
			id := fmt.Sprintf("inf_%s_%s", t, GenerateUUID())
			// TODO: it should be noted that both the fire and the smoke
			// models are using the same url: /smoke
			modelUrl := t
			if t == "fire" {
				modelUrl = "smoke"
			}
			serverConfig.InferenceServers[id] = InferenceServer{
				Id:        id,
				Name:      fmt.Sprintf("%s%d", t, i+1),
				Url:       fmt.Sprintf("http://%s/%s", addr, modelUrl),
				ModelType: t,
				Enabled:   true,
				CreatedAt: now,
				UpdatedAt: now,
			}
			availableAServerIds = append(availableAServerIds, id)
		}
	}

	for i, addr := range availableBServerAddrs {
		t := "safetybelt"
		now := GetCurrentTime()
		id := fmt.Sprintf("inf_%s_%s", t, GenerateUUID())
		serverConfig.InferenceServers[id] = InferenceServer{
			Id:        id,
			Name:      fmt.Sprintf("%s%d", t, i+1),
			Url:       fmt.Sprintf("http://%s/%s", addr, t),
			ModelType: t,
			Enabled:   true,
			CreatedAt: now,
			UpdatedAt: now,
		}
		availableBServerIds = append(availableBServerIds, id)
	}

	// generate 'cameras' section
	ia, ib := 0, 0
	for _, c := range readCamerasFromFile("docs/info.xlsx") {
		now := GetCurrentTime()
		// basic info
		cid := fmt.Sprintf("cam_%s", GenerateUUID())
		camera := Camera{
			Id:        cid,
			Name:      c.DeviceName,
			RtspUrl:   c.RtspURL,
			Enabled:   true,
			Running:   true,
			CreatedAt: now,
			UpdatedAt: now,
		}
		// bindings
		for _, m := range c.Models {
			binding := InferenceServerBinding{
				Threshold: 0.5,
				// TODO: 0 means no max limit
				MaxThreshold: 0,
			}
			// arrange to type B server
			if m == "safetybelt" {
				binding.ServerId = availableBServerIds[ib]
				ib = (ib + 1) % len(availableBServerIds)
			} else {
				// arrange to type A server
				binding.ServerId = availableAServerIds[ia]
				ia = (ia + 1) % len(availableAServerIds)
			}
			camera.InferenceServerBindings = append(camera.InferenceServerBindings, binding)
		}
		serverConfig.Cameras[cid] = camera
	}

	// write server config to local file
	data, err := json.MarshalIndent(serverConfig, "", "  ")
	if err != nil {
		slog.Error("failed to marshal server config to .json file")
		return
	}

	if err := os.WriteFile("config.json", data, 0644); err != nil {
		slog.Error("failed to write json data to local file")
		return
	}
}
