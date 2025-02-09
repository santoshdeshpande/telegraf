package fanuc

/*
#cgo LDFLAGS: -L/usr/lib -lfwlib32 -Wl,-rpath=./../../../
#include <stdlib.h>
#include "./fwlib32.h"
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"
	"unsafe"
)

type FocasInfo struct {
	TimeStamp     time.Time
	MachineIP     string
	Id            string
	Path          Path
	SpindleData   []Spindle
	Axes          []Axes
	PowerOnTime   uint32
	CuttingTime   uint32
	OperatingTime uint32
	CycleTime     uint32
	StatusInfo    Status
}

type Path struct {
	PathNo    int
	MaxPathNo int
}

type Spindle struct {
	Name      string
	SpindleNo int
}

type Axes struct {
	Name   string
	AxesNo int
}

type Status struct {
	Mode      string
	Execution string
}

func ExtractMachineData(ip string) (*FocasInfo, error) {
	//get the current time in UTC format
	t := time.Now().UTC()
	focasInfo := FocasInfo{MachineIP: ip, TimeStamp: t}
	handle, err := connect(ip, 8193)
	defer C.cnc_freelibhndl(handle)
	if err != nil {
		return nil, fmt.Errorf("error connecting to machine %s: %v", ip, err)
	}

	id, err := getCncId(handle)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	focasInfo.Id = id

	// _, err := getPath(handle)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return nil, err
	// }
	// // focasInfo.Path = path

	// spindles, err := getSpindleNames(handle)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return nil, err
	// }
	// focasInfo.SpindleData = spindles

	// axes, err := getAxisNames(handle)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return nil, err
	// }
	// focasInfo.Axes = axes

	ctime, err := getRdTimer(handle, 0)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	focasInfo.PowerOnTime = ctime

	ctime, err = getRdTimer(handle, 1)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	focasInfo.OperatingTime = ctime

	ctime, err = getRdTimer(handle, 2)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	focasInfo.CuttingTime = ctime
	ctime, err = getRdTimer(handle, 3)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	focasInfo.CycleTime = ctime

	statInfo, err := getStatInfo(handle)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	focasInfo.StatusInfo = statInfo
	return &focasInfo, nil
}

func setupLib() error {
	log_level := 0
	log_fname := C.CString("focas.log")
	defer C.free(unsafe.Pointer(log_fname))
	if ret := C.cnc_startupprocess(C.long(log_level), log_fname); ret != C.EW_OK {
		return fmt.Errorf("cnc_startupprocess failed (%d)", ret)
	}
	return nil
}

func connect(ipAddress string, port C.ushort) (C.ushort, error) {
	err := setupLib()
	if err != nil {
		return 0, err
	}
	ip := C.CString(ipAddress)
	defer C.free(unsafe.Pointer(ip))
	var handle C.ushort
	ret := C.cnc_allclibhndl3(ip, port, 10, &handle)
	if ret != C.EW_OK {
		return 0, fmt.Errorf("cnc_allclibhndl3 failed (%d)", ret)
	}
	return handle, nil
}

func getCncId(handle C.ushort) (string, error) {
	cncId := make([]C.ulong, 4)
	ret := C.cnc_rdcncid(handle, &cncId[0])
	if ret != 0 {
		return "", fmt.Errorf("cnc_rdcncid failed with error code %d", ret)
	}
	hexParts := make([]string, len(cncId))
	for i, id := range cncId {
		hexParts[i] = fmt.Sprintf("%X", id) // Convert to uppercase hex
	}
	id := strings.Join(hexParts, "-")

	return id, nil
}

func getPath(handle C.ushort) (Path, error) {
	var pathNo C.short
	var maxPathNo C.short
	ret := C.cnc_getpath(handle, &pathNo, &maxPathNo)
	if ret != C.EW_OK {
		return Path{}, fmt.Errorf("cnc_getpath failed (%d)", ret)
	}
	return Path{int(pathNo), int(maxPathNo)}, nil
}

func getSpindleNames(handle C.ushort) ([]Spindle, error) {
	var dataNum C.short = 4
	names := make([]C.ODBSPDLNAME, dataNum)
	ret := C.cnc_rdspdlname(handle, &dataNum, &names[0])
	if ret != 0 {
		return nil, fmt.Errorf("cnc_rdspdlname failed with error code: %d", ret)
	}
	spindles := make([]Spindle, dataNum)
	for i := 0; i < int(dataNum); i++ {
		data := names[i]
		spindles[i] = Spindle{C.GoString(&data.name), int(i + 1)}
	}
	return spindles, nil
}

func getAxisNames(handle C.ushort) ([]Axes, error) {
	var dataNum C.short = 8 // get the data for atmost 8 axes
	names := make([]C.ODBAXISNAME, dataNum)
	ret := C.cnc_rdaxisname(handle, &dataNum, &names[0])
	if ret != 0 {
		return nil, fmt.Errorf("cnc_rdaxisname failed with error code: %d", ret)
	}
	axes := make([]Axes, dataNum)
	for i := 0; i < int(dataNum); i++ {
		data := names[i]
		axes[i] = Axes{C.GoString(&data.name), int(i + 1)}
	}
	return axes, nil
}

type GIODBPSD struct {
	Datano C.short
	Type   C.short
	Raw    [4]uint8
	U      [512]uint8
}

func getRdParamDoubleWord(handle C.ushort, paramNo C.short) (uint32, error) {
	var param C.IODBPSD
	ret := C.cnc_rdparam(handle, paramNo, 0, 8, &param)
	if ret != C.EW_OK {
		return 0, fmt.Errorf("cnc_rdparam failed (%d)", ret)
	}
	p := (*GIODBPSD)(unsafe.Pointer(&param))
	ldata := binary.LittleEndian.Uint32(p.Raw[:4])
	return ldata, nil
}

func getRdTimer(handle C.ushort, typ C.short) (uint32, error) {
	var param C.IODBTIME
	ret := C.cnc_rdtimer(handle, typ, &param)
	if ret != C.EW_OK {
		return 0, fmt.Errorf("cnc_rdtimer failed (%d)", ret)
	}
	return uint32(param.minute + param.msec*60), nil
}

func getStatInfo(handle C.ushort) (Status, error) {
	var stat C.ODBST
	ret := C.cnc_statinfo(handle, &stat)
	if ret != C.EW_OK {
		return Status{}, fmt.Errorf("cnc_statinfo failed (%d)", ret)
	}
	return Status{convertMode(int(stat.aut)), convertExecution(stat)}, nil
}

func convertMode(aut int) string {
	var mode string
	switch aut {
	case 0:
		mode = "MANUAL_DATA_INPUT"
	case 1, 10:
		mode = "AUTOMATIC"
	case 3:
		mode = "EDIT"
	default:
		mode = "MANUAL"
	}
	return mode
}

func convertExecution(data C.ODBST) string {
	execution := "UNAVAILABLE"
	switch int(data.emergency) {
	case 0:
		switch int(data.run) {
		case 0:
			execution = "READY"
		case 1:
			execution = "STOPPED"
		case 2:
			execution = "FEED_HOLD"
		case 3:
			execution = "ACTIVE"

		}
	case 1:
		execution = "STOPPED"
	}
	return execution
}
