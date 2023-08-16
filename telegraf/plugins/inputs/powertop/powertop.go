package powertop

import (
	"encoding/csv"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type EnergyTop struct {
	Scheduler        float64
	DisplaySoftPower bool   `toml:"display_soft_power"`
	PathToCsv        string `toml:"path_to_csv"`
}

var EnergyTopConfig = `
  ## Set the scheduler
  #  scheduler = 20.0

  ## Display more details
  # displaySoftPower = false

  ## csv file path (include the file name)
  # path_to_csv = "/usr/src/powertop.csv"
`

func (e *EnergyTop) ConfigSample() string {
	return EnergyTopConfig
}

func (e *EnergyTop) Explain() string {
	return "fetch powertop report"
}

func (e *EnergyTop) Collect(acc telegraf.Accumulator) error {
	// generate csv file
	err := GeneratePowertopFile(e.PathToCsv, e.Scheduler)
	if err != nil {
		return err
	}
	now := time.Now()
	if e.DisplaySoftPower {
		err = AddSoftPowerConsumerFields(acc, e.PathToCsv, now)
		if err != nil {
			return err
		}
	}
	err = AddDevicePowerReportFields(acc, e.PathToCsv, now)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	inputs.Add("topower", func() telegraf.Input {
		return &EnergyTop{
			Scheduler:        20.0,
			DisplaySoftPower: false,
			PathToCsv:        "/usr/src/powertop.csv",
		}
	})
}

func AddSoftPowerConsumerFields(acc telegraf.Accumulator, csvPath string, now time.Time) error {
	out, err := FilterPowertopOutput("sed", "-n", "/Usage;Wakeups\\/s;GPU ops\\/s;/,/^$/p", csvPath)
	if err != nil {
		return err
	}
	// split without header
	var total float64
	for _, line := range out[1:] {
		// if my value is empty
		if line[7] == "" || line[7] == "0" {
			continue
		}
		tags := map[string]string{
			"category":     line[5],
			"description:": line[6],
		}
		fields := map[string]interface{}{
			"usage/s":                    ExtractAndConvertPrefix(line[0]),
			"wakeups/s":                  ExtractAndConvertPrefix(line[1]),
			"(software) PW estimate [W]": ExtractAndConvertPrefix(line[7]),
		}
		total += ExtractAndConvertPrefix(line[7])
		acc.AddFields("software power consumer", fields, tags, now)
	}
	acc.AddFields(
		"software power consumer",
		map[string]interface{}{
			"total power [W]": total,
		},
		map[string]string{},
		now)

	return nil
}

func AddDevicePowerReportFields(acc telegraf.Accumulator, csvPath string, now time.Time) error {
	out, err := FilterPowertopOutput("sed", "-n", "/Usage;Device Name;PW Estimate/,/^_/p", csvPath)
	if err != nil {
		return err
	}
	// split without header
	for _, line := range out[1 : len(out)-1] {
		// if value is empty, stop
		if strings.Trim(line[2], " ") == "" {
			break
		}
		tags := map[string]string{
			"device name": line[1],
		}
		fields := map[string]interface{}{
			"usage [%]":                ExtractAndConvertPrefix(line[0]),
			"(device) PW estimate [W]": ExtractAndConvertPrefix(line[2]),
		}
		acc.AddFields("device power report", fields, tags, now)
	}
	return nil
}

func GeneratePowertopFile(name string, scheduler float64) error {
	_, err := exec.Command("sudo", "powertop", fmt.Sprintf("--csv=%v", name), fmt.Sprintf("--time=%v", scheduler)).Output()
	if err != nil {
		return err
	}
	// check if file "name".csv exist
	if _, err = os.Stat(fmt.Sprintf("%v", name)); err != nil {
		return err
	}
	return nil
}

func FilterPowertopOutput(name string, arg ...string) ([][]string, error) {
	outputFilter, err := exec.Command(name, arg...).Output()
	if err != nil {
		return nil, err
	}
	if string(outputFilter) == "" {
		return nil, fmt.Errorf("Data not found")
	}
	r := csv.NewReader(strings.NewReader(string(outputFilter)))
	r.Comma = ';'
	r.FieldsPerRecord = -1
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	return records, nil
}

func ExtractAndConvertPrefix(data string) float64 {
	reg_result := regexp.MustCompile(`([0-9]+\.?[0-9]*)[\s]*([a-zA-Z \/]+)?`).FindSubmatch([]byte(data))
	if len(reg_result) == 0 {
		return 0
	}
	val, err := strconv.ParseFloat(string(reg_result[1]), 64)
	if err != nil {
		fmt.Println(err)
	}
	if len(reg_result) > 2 {
		format := string(reg_result[2])
		if strings.ContainsAny(format, "n") {
			val *= 1e-9
			return val
		}
		if strings.ContainsAny(format, "u") {
			val *= 1e-6
			return val
		}
		if strings.ContainsAny(format, "m") {
			val *= 1e-3
			return val
		}
		if strings.ContainsAny(format, "k") {
			val *= 1e3
			return val
		}
	}
	return val
}

