package programs

import (
	"github.com/pufferpanel/pufferd/environments"
)

type ProgramData struct {
	Data            map[string]DataObject  `json:"data"`
	Display         string                 `json:"display"`
	EnvironmentData map[string]interface{} `json:"environment"`
	InstallData     InstallSection         `json:"install"`
	Type            string                 `json:"type"`
	Identifier      string                 `json:"id"`
	RunData         RunObject              `json:"run"`

	Environment environments.Environment `json:"-"`
}

type DataObject struct {
	Description string      `json:"desc"`
	Display     string      `json:"display"`
	Internal    bool        `json:"internal"`
	Required    bool        `json:"required"`
	Value       interface{} `json:"value"`
}

type RunObject struct {
	Arguments               []string                 `json:"arguments"`
	Program                 string                   `json:"program"`
	Stop                    string                   `json:"stop"`
	Enabled                 bool                     `json:"enabled"`
	AutoStart               bool                     `json:"autostart"`
	AutoRestartFromCrash    bool                     `json:"autorecover"`
	AutoRestartFromGraceful bool                     `json:"autorestart"`
	Pre                     []map[string]interface{} `json:"pre"`
}

type InstallSection struct {
	Operations []map[string]interface{} `json:"commands"`
}

func (p ProgramData) DataToMap() map[string]interface{} {
	var result = make(map[string]interface{}, len(p.Data))

	for k, v := range p.Data {
		result[k] = v.Value
	}

	return result
}

func CreateProgram() ProgramData{
	return ProgramData{
		RunData: RunObject{
			Enabled: true,
			AutoStart: true,
			Pre: make([]map[string]interface{}, 0),
		},
		Type: "standard",
		Data: make(map[string]DataObject, 0),
		Display: "Unknown server",
		InstallData: InstallSection{
			Operations: make([]map[string]interface{}, 0),
		},
	}
}
