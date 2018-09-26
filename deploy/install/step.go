package install

import (
	"fmt"
	"strings"
	"time"

	pb "github.com/CS-SI/SafeScale/broker"
	brokerclient "github.com/CS-SI/SafeScale/broker/client"

	"github.com/CS-SI/SafeScale/deploy/install/enums/Action"
)

type stepTargets map[string]string

func (st stepTargets) parse() (string, string, string, error) {
	var masterT, privnodeT, pubnodeT string

	switch strings.ToLower(st[targetMasters]) {
	case "":
		fallthrough
	case "false":
		fallthrough
	case "no":
		fallthrough
	case "none":
		fallthrough
	case "0":
		masterT = "0"
	case "any":
		fallthrough
	case "one":
		fallthrough
	case "1":
		masterT = "1"
	case "all":
		fallthrough
	case "*":
		masterT = "*"
	default:
		return "", "", "", fmt.Errorf("invalid value '%s' for target '%s'", masterT, targetMasters)
	}

	switch strings.ToLower(st[targetPrivateNodes]) {
	case "":
		fallthrough
	case "false":
		fallthrough
	case "no":
		fallthrough
	case "none":
		privnodeT = "0"
	case "any":
		fallthrough
	case "one":
		fallthrough
	case "1":
		privnodeT = "1"
	case "all":
		fallthrough
	case "*":
		privnodeT = "*"
	default:
		return "", "", "", fmt.Errorf("invalid value '%s' for target '%s'", privnodeT, targetPrivateNodes)
	}

	switch strings.ToLower(st[targetPublicNodes]) {
	case "":
		fallthrough
	case "false":
		fallthrough
	case "no":
		fallthrough
	case "none":
		fallthrough
	case "0":
		pubnodeT = "0"
	case "any":
		fallthrough
	case "one":
		fallthrough
	case "1":
		pubnodeT = "1"
	case "all":
		fallthrough
	case "*":
		pubnodeT = "*"
	default:
		return "", "", "", fmt.Errorf("invalid value '%s' for target '%s'", pubnodeT, targetPublicNodes)
	}

	if masterT == "0" && privnodeT == "0" && pubnodeT == "0" {
		return "", "", "", fmt.Errorf("no targets identified")
	}
	return masterT, privnodeT, pubnodeT, nil
}

type step struct {
	Worker             *worker
	Name               string
	Action             Action.Enum
	Targets            stepTargets
	Script             string
	WallTime           time.Duration
	YamlKey            string
	OptionsFileContent string
}

// Run executes the step on all the concerned hosts
func (is *step) Run(v Variables) (stepErrors, error) {
	// Determine list of hosts concerned by the step
	hostsList, err := is.identifyHosts(is.Worker, is.Targets)
	if err != nil {
		return nil, err
	}

	// Empty results
	results := stepErrors{}

	broker := brokerclient.New()
	for _, host := range hostsList {
		// Updates variables in step script
		command, err := replaceVariablesInString(is.Script, v)
		if err != nil {
			return results, fmt.Errorf("failed to finalize installer script: %s", err.Error())
		}

		// If options file is defined, upload it to the remote host
		if is.OptionsFileContent != "" {
			err := UploadStringToRemoteFile(is.OptionsFileContent, host, "/var/tmp/options.json", "cladm", "gpac", "ug+rw-x,o-rwx")
			if err != nil {
				return results, err
			}
		}

		// Uploads then executes command
		filename := fmt.Sprintf("/var/tmp/%s_add.sh", is.Worker.component.BaseFilename())
		err = UploadStringToRemoteFile(command, host, filename, "", "", "")
		if err != nil {
			return results, err
		}
		//if debug {
		if true {
			command = fmt.Sprintf("sudo bash %s", filename)
		} else {
			command = fmt.Sprintf("sudo bash %s; rc=$?; sudo rm -f %s /var/tmp/options.json; exit $rc", filename, filename)
		}

		// Executes the script on the remote host
		retcode, _, _, err := broker.Ssh.Run(host.Name, command, brokerclient.DefaultConnectionTimeout, is.WallTime)
		if err != nil {
			return results, err
		}
		err = nil
		ok := retcode == 0
		if !ok {
			err = fmt.Errorf("installer step failed (retcode=%d)", retcode)
		}
		results[host.Name] = err
	}
	return results, nil
}

// IdentifyHosts identifies hosts concerned by the step
func (is *step) identifyHosts(w *worker, targets stepTargets) ([]*pb.Host, error) {
	//specs := is.Worker.component.Specs()

	masterT, privnodeT, pubnodeT, err := targets.parse()
	if err != nil {
		return nil, err
	}

	hostsList := []*pb.Host{}
	switch masterT {
	case "1":
		host, err := is.Worker.AvailableMaster()
		if err != nil {
			return nil, err
		}
		hostsList = append(hostsList, host)
	case "*":
		all, err := is.Worker.AllMasters()
		if err != nil {
			return nil, err
		}
		hostsList = append(hostsList, all...)
	}
	switch privnodeT {
	case "1":
		host, err := is.Worker.AvailableNode(false)
		if err != nil {
			return nil, err
		}
		hostsList = append(hostsList, host)
	case "*":
		hosts, err := is.Worker.AllNodes(false)
		if err != nil {
			return nil, err
		}
		hostsList = append(hostsList, hosts...)
	}
	switch pubnodeT {
	case "1":
		host, err := is.Worker.AvailableNode(true)
		if err != nil {
			return nil, err
		}
		hostsList = append(hostsList, host)
	case "*":
		hosts, err := is.Worker.AllNodes(true)
		if err != nil {
			return nil, err
		}
		hostsList = append(hostsList, hosts...)
	}

	return hostsList, nil
}