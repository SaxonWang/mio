package utils

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"hidevops.io/hiboot/pkg/log"
	utilsio "hidevops.io/hiboot/pkg/utils/io"
	"hidevops.io/mio/node/protobuf"
	miov1alpha1 "hidevops.io/mioclient/pkg/apis/mio/v1alpha1"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	//docker_types "github.com/docker/docker/api/types"

	"os/exec"
	"time"
)

type ResourceString struct {
	XMLName      xml.Name `xml:"project"`
	ModelVersion string   `xml:"modelVersion"`
	GroupId      string   `xml:"groupId"`
	ArtifactId   string   `xml:"artifactId"`
	Version      string   `xml:"version"`
	Packaging    string   `xml:"packaging"`
}

func GetPomXmlInfo(pomName string) (*ResourceString, error) {
	file, err := os.Open(pomName)
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil, err
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil, err
	}
	resource := &ResourceString{}
	err = xml.Unmarshal([]byte(data), resource)
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil, err
	}
	if resource.Packaging == "" {
		resource.Packaging = "jar"
	}

	//projectName := fmt.Sprintf("%s-%s.%s",resource.ArtifactId,resource.Version,resource.Packaging)
	return resource, nil
}

func ExecCommand(commandName string, params []string) (*exec.Cmd, *bufio.Reader, error) {
	cmd := exec.Command(commandName, params...)

	fmt.Println("$ ", cmd.Args)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	return cmd, bufio.NewReader(stdout), nil
}

func GenScript(scriptContent string) (string, error) {
	temporaryFile := fmt.Sprintf("./script-%d.sh", time.Now().Unix())
	fileObj, err := os.OpenFile(temporaryFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Debug("Failed to open the file", err.Error())
		os.Exit(2)
	}
	defer fileObj.Close()
	if _, err := fileObj.WriteString(scriptContent); err == nil {
		return temporaryFile, err
	}
	fileObj.Sync()
	return temporaryFile, nil
}

func GetBuildFileBYDockerfile(dockerfile []string) []string {
	var buildFiles []string
	buildFiles = append(buildFiles, "Dockerfile")
	for _, dockerfile := range dockerfile {
		if strings.Contains(dockerfile, "ADD") || strings.Contains(dockerfile, "COPY") {
			files := strings.Split(dockerfile, " ")

			for i, f := range files {
				if i == 0 {
					continue
				}

				if f != "" {
					buildFiles = append(buildFiles, f)
					break
				}
			}
		}
	}
	return buildFiles
}

func CloneBYCMD(sourceCodePullRequest *protobuf.SourceCodePullRequest) (string, error) {

	//git clone -b "分支" --depth=1 xxx.git "指定目录"

	//git clone http://oauth2:cf5c6dd52a297270fb5f94e1f763bb6acac3169561f98161258c4103b2b7b17f@gitlab.vpclub:8022/hidevopsio/hiadmin.git

	urls := strings.Split(sourceCodePullRequest.Url, "//")

	projectName := utilsio.Filename(sourceCodePullRequest.Url)
	projectName = utilsio.Basename(projectName)
	projectPath := filepath.Join(sourceCodePullRequest.DstDir, projectName)

	if _, err := os.Stat(projectPath); err == nil {
		return "", fmt.Errorf("file %s already exists", projectPath)
	}

	url := sourceCodePullRequest.Url
	if sourceCodePullRequest.Token != "" {
		url = fmt.Sprintf("%s//oauth2:%s@%s", urls[0], sourceCodePullRequest.Token, urls[1])
	}
	Params := []string{"clone", "-b", sourceCodePullRequest.Branch, "--depth=1", url, projectPath}

	cmd, bufioReader, err := ExecCommand("git", Params)
	if err != nil {
		return "", err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err = cmd.Start(); err != nil {
		return "", err
	}

	for {
		line, err2 := bufioReader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		fmt.Println(line)
	}
	if err = cmd.Wait(); err != nil {
		return "", err
	}

	if _, err := os.Stat(projectPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("clone %s fail", sourceCodePullRequest.Url)
		} else {
			return "", err
		}
	}

	return projectPath, nil
}

func TestStart(testRequest *protobuf.TestsRequest) error {
	execCommand := func(CommandName string, Params []string) error {
		cmd, bufioReader, err := ExecCommand(CommandName, Params)
		if err != nil {

			return err
		}

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err = cmd.Start(); err != nil {
			return err
		}
		for {
			line, err2 := bufioReader.ReadString('\n')
			if err2 != nil || io.EOF == err2 {
				break
			}

			fmt.Println(line)
		}
		if err = cmd.Wait(); err != nil {
			return err
		}
		return nil
	}

	for _, cmd := range testRequest.TestCmd {

		if cmd.ExecType == string(miov1alpha1.Script) {
			fmt.Println("$script:\n", cmd.Script)
			scriptPath, err := GenScript(cmd.Script)
			if err != nil {
				return err
			}
			//defer os.RemoveAll(scriptPath)

			if err := execCommand("chmod", []string{"+x", scriptPath}); err != nil {
				return err
			}

			if err := execCommand("sh", []string{"-c", scriptPath}); err != nil {
				fmt.Println("Error CMD filed")
				return err
			}
			os.RemoveAll(scriptPath)
			continue
		}

		if err := execCommand(cmd.CommandName, cmd.Params); err != nil {
			fmt.Println("Error compile filed")
			return err
		}
	}
	return nil
}

func StartCmd(commandRequest *protobuf.CommandRequest) error {
	execCommand := func(CommandName string, Params []string) error {
		cmd, bufioReader, err := ExecCommand(CommandName, Params)
		if err != nil {
			return err
		}

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err = cmd.Start(); err != nil {
			return err
		}
		for {
			line, err2 := bufioReader.ReadString('\n')
			if err2 != nil || io.EOF == err2 {
				break
			}

			fmt.Println(line)
		}
		if err = cmd.Wait(); err != nil {
			return err
		}
		return nil
	}

	for _, cmd := range commandRequest.CommandList {

		if cmd.ExecType == string(miov1alpha1.Script) {
			fmt.Println("$script:\n", cmd.Script)
			scriptPath, err := GenScript(cmd.Script)
			if err != nil {
				return err
			}
			defer os.RemoveAll(scriptPath)

			if err := execCommand("chmod", []string{"+x", scriptPath}); err != nil {
				return err
			}

			if err := execCommand("sh", []string{"-c", scriptPath}); err != nil {
				fmt.Println("Error CMD filed")
				return err
			}
			continue
		}

		if err := execCommand(cmd.CommandName, cmd.Params); err != nil {
			fmt.Println("Error compile filed")
			return err
		}
	}
	return nil
}
