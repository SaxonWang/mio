package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"hidevops.io/hiboot/pkg/log"
	"hidevops.io/mio/node/protobuf"
	"io/ioutil"
	"os"
	"testing"
)

func TestGetProjectNameByPom(t *testing.T) {
	pom := `
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
	xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
	<modelVersion>4.0.0</modelVersion>

	<groupId>com.example</groupId>
	<artifactId>hello-world</artifactId>
	<version>0.0.7-SNAPSHOT</version>
	<packaging>jar</packaging>

	<name>hello-world</name>
	<description>Demo project for Spring Boot</description>
</project>

`

	file, err := ioutil.TempFile("", "tmpfile")
	if err != nil {
		panic(err)
	}
	defer os.Remove(file.Name())

	if _, err := file.Write([]byte(pom)); err != nil {
		panic(err)
	}

	xmlInfo, err := GetPomXmlInfo(file.Name())
	assert.Equal(t, nil, err)
	fmt.Println(xmlInfo)
}

func TestBuildFileInit(t *testing.T) {
	file1 := "demo"
	file2 := "src/demo.jar"

	dockerFiles := []string{"FROM ubuntu:16.04",
		"WORKDIR /vpcloue",
		fmt.Sprintf("COPY  %s  /dest/%s", file1, file1),
		fmt.Sprintf("ADD      %s   /dest/%s", file2, file2),
	}

	log.Debugf("buildFile %v", GetBuildFileBYDockerfile(dockerFiles))
	number := 0
	for _, f := range GetBuildFileBYDockerfile(dockerFiles) {

		if f == file1 {
			number++
		}

		if f == file2 {
			number++
		}
	}
	assert.Equal(t, 2, number)
}

func TestTestStart(t *testing.T) {
	sourceCodeTestRequest := &protobuf.CommandRequest{
		CommandList: []*protobuf.Command{{
			CommandName: "pwd",
			Params:      []string{},
		}, {ExecType: "script", Script: "pwd"}},
	}
	err := StartCmd(sourceCodeTestRequest)

	assert.Equal(t, nil, err)
}

func TestStartCmd(t *testing.T) {
	sourceCodeTestRequest := &protobuf.TestsRequest{
		TestCmd: []*protobuf.TestCommand{{
			CommandName: "pwd",
			Params:      []string{},
		}, {ExecType: "script", Script: "pwd"}},
	}
	err := TestStart(sourceCodeTestRequest)

	assert.Equal(t, nil, err)
}