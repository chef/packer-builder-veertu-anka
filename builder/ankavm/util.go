package ankavm

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

func stepError(ui packer.Ui, state multistep.StateBag, err error) multistep.StepAction {
	state.Put("error", err)
	ui.Error(err.Error())
	return multistep.ActionHalt
}

func convertDiskSizeToBytes(diskSize string) (error, uint64) {
	match, err := regexp.MatchString("^[0-9]+[g|G|m|M]$", diskSize)
	if err != nil {
		return err, uint64(0)
	}
	if !match {
		return fmt.Errorf("Input %s is not a valid disk size input", diskSize), 0
	}

	numericValue, err := strconv.Atoi(diskSize[:len(diskSize)-1])
	if err != nil {
		return err, uint64(0)
	}
	suffix := diskSize[len(diskSize)-1:]

	switch strings.ToUpper(suffix) {
	case "G":
		return nil, uint64(numericValue * 1024 * 1024 * 1024)
	case "M":
		return nil, uint64(numericValue * 1024 * 1024)
	default:
		return fmt.Errorf("Invalid disk size suffix: %s", suffix), uint64(0)
	}
}

func convertDiskSizeFromBytes(diskSize uint64) string {
	var suffixes [5]string
	suffixes[0] = "B"
	suffixes[1] = "K"
	suffixes[2] = "M"
	suffixes[3] = "G"
	suffixes[4] = "T"

	base := math.Log(float64(diskSize)) / math.Log(1024)
	getSize := round(math.Pow(1024, base-math.Floor(base)), .5, 2)
	getSuffix := suffixes[int(math.Floor(base))]

	return strconv.FormatFloat(getSize, 'f', -1, 64) + string(getSuffix)
}

func round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}
