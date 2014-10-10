package main

import (
	"net"
	"fmt"
	"bufio"
	"errors"
	"strings"
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"bytes"
	"strconv"
)


var Binary string
var PidFile string


func SetPidFileName(c string) error {
	PidFile = c
	return nil
}

func SetBinaryFileName(c string) error {
	Binary = c
	return nil
}


// Executes a arbitrariry HAproxy command on the unix socket
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

// Sets the weight of a backend
func SetWeight(backend string, server string, weight int) (string, error){

	result, err := HaproxyCmd("set weight " + backend + "/" + server + " " + strconv.Itoa(weight) +"\n")


	if err != nil {
		return "", err
	} else {
		return result, nil
	}

}

/*

	Stats

 */

/* get the basic stats in CSV format

	@parameter statsType takes the form of:
	-	all
	-	frontend
	-	backend
*/
func GetStats(statsType string) ([]StatsGroup, error) {
	var Stats []StatsGroup
	var cmdString string

	switch statsType {
	case "all":
		cmdString = "show stat -1\n"
	case "backend":
		cmdString = "show stat -1 2 -1\n"
	case "frontend":
		cmdString = "show stat -1 1 -1\n"
	case "server":
		cmdString = "show stat -1 4 -1\n"
	}

	result, err := HaproxyCmd(cmdString)
	if err != nil {
		return Stats, err
	} else {
		result, err := parse_csv(strings.Trim(result,"# "))
		if err != nil {
			return Stats, err
		} else {
			err := json.Unmarshal([]byte(result), &Stats)
			if err != nil {
				return Stats, err
			} else {
				return Stats, nil
			}
		}

	}
}

/*

	Reload

 */

// Configuration reload
func Reload() error {

	pid, err := ioutil.ReadFile(PidFile)
	if err !=nil {
		return err
	}

	/* 	Setup all the command line parameters so we get an executable similar to
		/usr/local/bin/haproxy -f resources/haproxy_new.cfg -p resources/haproxy-private.pid -st 1234

	*/
	arg0 := "-f"
	arg1 := ConfigFile
	arg2 := "-p"
	arg3 := PidFile
	arg4 := "-D"
	arg5 := "-sf"
	arg6 := strings.Trim(string(pid),"\n")
	var cmd *exec.Cmd

	// If this is the first run, the PID value will be empty, otherwise it will be > 0
	if len(arg6) > 0 {
		cmd = exec.Command(Binary, arg0, arg1 ,arg2, arg3, arg4, arg5)
	} else {
		cmd = exec.Command(Binary, arg0, arg1 ,arg2, arg3 )
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmdErr := cmd.Run()
	if cmdErr != nil {
		fmt.Println(cmdErr.Error())
		return cmdErr
	}
	log.Info("HaproxyReload: " + out.String() + string(pid))
	return nil
}


/*

	Info

 */

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


