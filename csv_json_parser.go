package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"bytes"
	"encoding/json"
)

type StatsGroup struct {
 Pxname string `json:"pxname"`
 Svname string `json:"svname"`
 Qcur string `json:"qcur"`
 Qmax string `json:"qmax"`
 Scur string `json:"scur"`
 Smax string `json:"smax"`
 Slim string `json:"slim"`
 Stot string `json:"stot"`
 Bin string `json:"bin"`
 Bout string `json:"bout"`
 Dreq string `json:"dreq"`
 Dresp string `json:"dresp"`
 Ereq string `json:"ereq"`
 Econ string `json:"econ"`
 Eresp string `json:"eresp"`
 Wretr string `json:"wretr"`
 Wredis string `json:"wredis"`
 Status string `json:"status"`
 Weight string `json:"weight"`
 Act string `json:"act"`
 Bck string `json:"bck"`
 Chkfail string `json:"chkfail"`
 Chkdown string `json:"chkdown"`
 Lastchg string `json:"lastchg"`
 Downtime string `json:"downtime"`
 Qlimit string `json:"qlimit"`
 Pid string `json:"pid"`
 Iid string `json:"iid"`
 Sid string `json:"sid"`
 Throttle string `json:"throttle"`
 Lbtot string `json:"lbtot"`
 Tracked string `json:"tracked"`
 _Type string `json:"type"`
 Rate string `json:"rate"`
 Rate_lim string `json:"rate_lim"`
 Rate_max string `json:"rate_max"`
 Check_status string `json:"check_status"`
 Check_code string `json:"check_code"`
 Check_duration string `json:"check_duration"`
 Hrsp_1xx string `json:"hrsp_1xx"`
 Hrsp_2xx string `json:"hrsp_2xx"`
 Hrsp_3xx string `json:"hrsp_3xx"`
 Hrsp_4xx string `json:"hrsp_4xx"`
 Hrsp_5xx string `json:"hrsp_5xx"`
 Hrsp_other string `json:"hrsp_other"`
 Hanafail string `json:"hanafail"`
 Req_rate string `json:"req_rate"`
 Req_rate_max string `json:"req_rate_max"`
 Req_tot string `json:"req_tot"`
 Cli_abrt string `json:"cli_abrt"`
 Srv_abrt string `json:"srv_abrt"`
 Comp_in string `json:"comp_in"`
 Comp_out string `json:"comp_out"`
 Comp_byp string `json:"comp_byp"`
 Comp_rsp string `json:"comp_rsp"`
 Lastsess string `json:"lastsess"`
 Last_chk string `json:"last_chk"`
 Last_agt string `json:"last_agt"`
 Qtime string `json:"qtime"`
 Ctime string `json:"ctime"`
 Rtime string `json:"rtime"`
 Ttime string `json:"ttime"`
}



// parses the raw stats CSV output to a Json Object
func parse_csv(csvInput string) ([]StatsGroup, error){

	csvReader := csv.NewReader(strings.NewReader(csvInput))
	lineCount := 0
	var headers []string
	var result bytes.Buffer
	var item bytes.Buffer
	var empty []StatsGroup
	var statsAll []StatsGroup
	result.WriteString("[")


	for {
		// read just one record, but we could ReadAll() as well
		record, err := csvReader.Read()

		if err == io.EOF {
			result.Truncate(int(len(result.String())-1))
			result.WriteString("]")
			break
		} else if err != nil {
			fmt.Println("Error:", err)
			return empty, err
		}

		if lineCount == 0 {
			headers = record[:]
			lineCount += 1
		} else
		{
			item.WriteString("{")
			for i := 0; i < len(headers); i++ {
				item.WriteString("\"" + headers[i] + "\": \"" + record[i] + "\"")
				if i == (len(headers)-1) {
					item.WriteString("}")
				} else {
					item.WriteString(",")
				}
			}
			result.WriteString(item.String() + ",")
			item.Reset()
			lineCount += 1
		}
	}
	var jsonBlob = []byte(result.String())
	err := json.Unmarshal(jsonBlob, &statsAll)
	if err != nil {
		fmt.Println("error:", err)
	}
	return statsAll, nil
}
