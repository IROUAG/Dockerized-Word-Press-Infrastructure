package powertop

import (
	"encoding/csv"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	_ "math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Powertop struct {
	Timer             float64
	ShowSoftwarePower bool   `toml:"show_software_power"`
	CsvPath           string `toml:"csv_path"`
}

var PowertopConfig = `
  ## Set the timer
  #  timer = 20.0

  ## Add more details
  # showSoftwarePower = false

  ## file csv path (include the name of the file)
  # csv_path = "/usr/src/powertop.csv"
`

func (p *Powertop) SampleConfig() string {
	return PowertopConfig
}

func (p *Powertop) Description() string {
	return "get powertop report"
}

func (p *Powertop) Gather(acc telegraf.Accumulator) error {
	// generate csv file
	err := CreatePowertopFile(p.CsvPath, p.Timer)
	if err != nil {
		return err
	}
	now := time.Now()
	if p.ShowSoftwarePower {
		err = AddFieldsSoftwarePowerConsumer(acc, p.CsvPath, now)
		if err != nil {
			return err
		}
	}
	err = AddFieldsDevicePowerReport(acc, p.CsvPath, now)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	inputs.Add("powertop", func() telegraf.Input {
		return &Powertop{
			Timer:             20.0,
			ShowSoftwarePower: false,
			CsvPath:           "/usr/src/powertop.csv",
		}
	})
}

func AddFieldsSoftwarePowerConsumer(acc telegraf.Accumulator, csvPath string, now time.Time) error {
	out, err := PowertopOutputFilter("sed", "-n", "/Usage;Wakeups\\/s;GPU ops\\/s;/,/^$/p", csvPath)
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
		//fmt.Println(line)
		tags := map[string]string{
			"category":     line[5],
			"description:": line[6],
		}
		fields := map[string]interface{}{
			"usage/s":                    FindAndConvertPrefix(line[0]),
			"wakeups/s":                  FindAndConvertPrefix(line[1]),
			"(software) PW estimate [W]": FindAndConvertPrefix(line[7]),
		}
		total += FindAndConvertPrefix(line[7])
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

func AddFieldsDevicePowerReport(acc telegraf.Accumulator, csvPath string, now time.Time) error {
	out, err := PowertopOutputFilter("sed", "-n", "/Usage;Device Name;PW Estimate/,/^_/p", csvPath)
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
			"usage [%]":                FindAndConvertPrefix(line[0]),
			"(device) PW estimate [W]": FindAndConvertPrefix(line[2]),
		}
		acc.AddFields("device power report", fields, tags, now)
	}
	return nil
}

func CreatePowertopFile(name string, timer float64) error {
	_, err := exec.Command("sudo", "powertop", fmt.Sprintf("--csv=%v", name), fmt.Sprintf("--time=%v", timer)).Output()
	if err != nil {
		return err
	}
	// check if file "name".csv exist
	if _, err = os.Stat(fmt.Sprintf("%v", name)); err != nil {
		return err
	}
	return nil
}

func PowertopOutputFilter(name string, arg ...string) ([][]string, error) {
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

func FindAndConvertPrefix(data string) float64 {
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

