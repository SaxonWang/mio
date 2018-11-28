package service

import (
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-git.v4"
	"hidevops.io/hioak/starter/docker"
	"hidevops.io/hioak/starter/docker/fake"
	"hidevops.io/mio/node/protobuf"
	"hidevops.io/mio/node/service/mock"
	"io"
	"os"
	"testing"
)

//go:generate mockgen -destination mock/mock_build.go -package mock hidevops.io/mio/node/pkg/service BuildConfigService

func TestBuild(t *testing.T) {

	defer os.RemoveAll("Dockerfile")

	projectName := "demo"

	var cmdList []*protobuf.BuildCommand
	cmd1 := &protobuf.BuildCommand{
		CodeType:    "",
		CommandName: "pwd",
		Params:      []string{},
	}

	cmd2 := &protobuf.BuildCommand{
		ExecType: "script",
		Script: `if [[ $? == 0 ]]; then
          echo "Build Successful."
        else
          echo "Build Failed!"
          exit 1
        fi`,
	}

	cmdList = append(cmdList, cmd1)
	cmdList = append(cmdList, cmd2)

	compileRequest := protobuf.CompileRequest{
		CompileCmd: cmdList,
	}

	//
	imageBuildRequest := protobuf.ImageBuildRequest{
		App:        projectName,
		S2IImage:   "FROM ubuntu:16.04",
		Tags:       []string{"test:0.1"},
		DockerFile: []string{},
	}

	//
	imagePushRequest := protobuf.ImagePushRequest{
		Tags: []string{"test:0.1"},
	}

	cli, err := fake.NewClient()
	assert.Equal(t, nil, err)
	buildConfigimpl := buildConfigServiceImpl{
		imageClient: &docker.ImageClient{Client: cli},
	}

	mockCtl := gomock.NewController(t)
	m := mock.NewMockBuildConfigService(mockCtl)

	t.Run("should err is nil in the clone code", func(t *testing.T) {
		m.EXPECT().Clone(nil).Return("", nil)
		dstPath, err := m.Clone(nil)
		assert.Equal(t, "", dstPath)
		assert.Equal(t, nil, err)
	})

	t.Run("should err in the compile", func(t *testing.T) {
		err := buildConfigimpl.Compile(&compileRequest)
		assert.Equal(t, nil, err)

	})

	t.Run("should err not nil in the image build", func(t *testing.T) {

		cli.On("ImageBuild", nil, nil,
			nil).Return(types.ImageBuildResponse{}, errors.New("1"))
		err = buildConfigimpl.ImageBuild(&imageBuildRequest)
		assert.NotEqual(t, nil, err)
	})

	t.Run("should err not nil in the image push", func(t *testing.T) {

		var i io.ReadCloser
		cli.On("ImagePush", nil, nil, nil).Return(i, errors.New("1"))
		err = buildConfigimpl.ImagePush(&imagePushRequest)
		assert.NotEqual(t, nil, err)
	})
}

func TestBuildConfigServiceImpl_Compile(t *testing.T) {

	os.Setenv("CODE_TYPE", ".")

	b := new(buildConfigServiceImpl)
	cr := &protobuf.CompileRequest{

		CompileCmd: []*protobuf.BuildCommand{{CodeType: Command, CommandName: "pwd"}, {ExecType: "script", Script: "pwd"}},
	}

	err := b.Compile(cr)
	assert.Equal(t, nil, err)
}

func TestBuildConfigServiceImpl_Clone(t *testing.T) {
	b := new(buildConfigServiceImpl)
	sp := &protobuf.SourceCodePullRequest{}

	b.Clone(sp, func(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {

		return nil, nil
	})
}

func TestBuildConfigServiceImpl_Clone2(t *testing.T) {

	b := new(buildConfigServiceImpl)
	sourceCodePullRequest := &protobuf.SourceCodePullRequest{
		Url:      "http://gitlab.vpclub:8022/wanglulu/clone-test01.git",
		Branch:   "master",
		DstDir:   "/Users/mac/.gvm/pkgsets/go1.10/vpcloud/src/hidevops.io/",
		Username: "",
		Password: "",
		Token:    "test",
	}
	_, err := b.Clone(sourceCodePullRequest, func(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {
		return nil, nil
	})
	assert.NotEqual(t, nil, err)
}
