package main

import (
	"net"
	"fmt"
	"bufio"
	"errors"
	"strings"
	"encoding/json"
)


func HaproxyCmd(cmd string) (string, error){

	// connect to haproxy
	conn, err_conn := net.Dial("unix", "/tmp/haproxy.stats.sock")
	defer conn.Close()

	if err_conn != nil {
		return "", errors.New("Unable to connect to Haproxy socket")
	} else {

		fmt.Fprint(conn, cmd)

		response := ""

		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			response += (scanner.Text() + "\n")
		}
		if err := scanner.Err(); err != nil {
			return "", err
		} else {
			return response, nil
		}

	}
}

/*

	Backends

 */


func SetWeight(backend string, server string, weight string) (string, error){
	result, err := HaproxyCmd("set weight " + backend + "/" + server + " " + weight +"\n")

	if err != nil {
		return "", err
	} else {
		fmt.Println(result)
		return result, nil
	}

}



/*

	Stats

 */

// get the basic stats in CSV format
func GetStats() ([]StatsGroup, error) {

	var empty []StatsGroup
	result, err := HaproxyCmd("show stat -1\n")

	if err != nil {
		return empty, err
	} else {

		json_result, err := parse_csv(strings.Trim(result,"# "))
		if err != nil {
			return empty, err
		} else {
			return json_result, nil
		}

	}

}

func GetStatsBackend() ([]StatsGroup, error) {

	var empty []StatsGroup
	result, err := HaproxyCmd("show stat -1 2 -1\n")
	if err != nil {
		return empty, err
	} else {

		json_result, err := parse_csv(strings.Trim(result,"# "))

		if err != nil {
			return empty, err
		} else {
			return json_result, nil
		}

	}
}

func GetStatsFrontend() ([]StatsGroup, error) {

	var empty []StatsGroup
	result, err := HaproxyCmd("show stat -1 1 -1\n")
	if err != nil {
		return empty, err
	} else {

		json_result, err := parse_csv(strings.Trim(result,"# "))
		if err != nil {
			return empty, err
		} else {
			return json_result, nil
		}

	}
}


func GetInfo() (Info, error) {
	var Info Info
	result, err := HaproxyCmd("show info \n")
	if err != nil {
		return Info, err
	} else {
		result, err := parse_multi_line(result)
		if err != nil {
			return Info, err
		} else {
			err := json.Unmarshal([]byte(result), &Info)
			if err != nil {
				return Info, err
			} else {
				return Info, nil
			}
		}
	}

}


