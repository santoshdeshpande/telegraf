//go:build !custom
// +build !custom

//go:generate ../../../tools/readme_config_includer/generator
package fanuc

import "C"

import (
	_ "embed"
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Fanuc struct {
	Machines []string `toml:"machines"`
	Timeout  int      `toml:"timeout"`
}

func (*Fanuc) SampleConfig() string {
	fmt.Println("Calling SampleConfig")
	return sampleConfig
}

func (m *Fanuc) Description() string {
	return "Read metrics from Fanuc CNC machines"
}

func (m *Fanuc) Gather(acc telegraf.Accumulator) error {
	for _, machine := range m.Machines {
		info, err := ExtractMachineData(machine)
		if err != nil {
			return err
		}
		tags := map[string]string{
			"machineIp": info.MachineIP,
			"machineId": info.Id,
		}
		fields := map[string]interface{}{
			"powerOnTime": info.PowerOnTime,
			"cuttingTime": info.CuttingTime,
			"operatingTime": info.OperatingTime,
			"mode": info.StatusInfo.Mode,
			"execution": info.StatusInfo.Execution,
		}
		acc.AddFields("fanuc", fields, tags)
	}
	return nil
}

func init() {
	inputs.Add("fanuc", func() telegraf.Input {
		return &Fanuc{}
	})
}

func init() {
	inputs.Add("fanuc", func() telegraf.Input {
		return &Fanuc{}
	})
}
