package service

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"hidevops.io/hiboot/pkg/app"
	"hidevops.io/hiboot/pkg/log"
	"hidevops.io/hioak/starter/docker"
	scmgit "hidevops.io/hioak/starter/scm/git"
	"hidevops.io/mio/node/protobuf"
	"hidevops.io/mio/node/types"
	miov1alpha1 "hidevops.io/mioclient/pkg/apis/mio/v1alpha1"
	//docker_types "github.com/docker/docker/api/types"

	pkg_utils "hidevops.io/mio/node/utils"
	"io"
	"os"
	"strings"
)

type BuildConfigService interface {
	Clone(sourceCodePullRequest *protobuf.SourceCodePullRequest, cloneFunc scmgit.CloneFunc) (string, error)
	Compile(compileRequest *protobuf.CompileRequest) error
	ImageBuild(imageBuildRequest *protobuf.ImageBuildRequest) error
	ImagePush(imagePushRequest *protobuf.ImagePushRequest) error
}

type buildConfigServiceImpl struct {
	BuildConfigService
	imageClient *docker.ImageClient
}

func init() {
	log.SetLevel(log.DebugLevel)
	app.Component(newBuildService)
}

func newBuildService(imageClient *docker.ImageClient) BuildConfigService {
	return &buildConfigServiceImpl{
		imageClient: imageClient,
	}
}

func (b *buildConfigServiceImpl) Clone(sourceCodePullRequest *protobuf.SourceCodePullRequest, cloneFunc scmgit.CloneFunc) (string, error) {
	fmt.Printf("\n[INFO] clone %s start:\n", sourceCodePullRequest.Url)

	if sourceCodePullRequest.Token != "" {
		//CMD
		Path, err := pkg_utils.CloneBYCMD(sourceCodePullRequest)
		if err != nil {
			fmt.Printf("Error clone %s filed:\n", sourceCodePullRequest.Url)
			return "", err
		} else {
			return Path, nil
		}
	}

	//go-git
	passwordAuth := transport.AuthMethod(&http.BasicAuth{
		Username: sourceCodePullRequest.Username,
		Password: sourceCodePullRequest.Password},
	)
	//	tokenAuth := transport.AuthMethod(&http.TokenAuth{Token: sourceCodePullRequest.Password})

	referenceName := fmt.Sprintf("refs/heads/%s", sourceCodePullRequest.Branch)
	codePath, err := scmgit.NewRepository(cloneFunc).Clone(&git.CloneOptions{URL: sourceCodePullRequest.Url,
		ReferenceName:     plumbing.ReferenceName(referenceName),
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Depth:             int(sourceCodePullRequest.Depth),
		Auth:              passwordAuth,
	},
		sourceCodePullRequest.DstDir)

	if err != nil {
		fmt.Printf("\nError clone %s filed:\n", sourceCodePullRequest.Url)
		os.RemoveAll(codePath)
		return "", err
	}
	fmt.Printf("\n[INFO] clone %s succeed\n", sourceCodePullRequest.Url)
	return codePath, nil
}

func (b *buildConfigServiceImpl) Compile(compileRequest *protobuf.CompileRequest) error {
	fmt.Println("\n[INFO] compile start:")

	execCommand := func(CommandName string, Params []string) error {
		cmd, bufioReader, err := pkg_utils.ExecCommand(CommandName, Params)
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
	codeType := os.Getenv("CODE_TYPE")
	if codeType == "" {
		return fmt.Errorf("env CODE_TYPE get filed")
	}

	if types.JAVA == codeType {

		pomXmlInfo, err := pkg_utils.GetPomXmlInfo("pom.xml")
		if err != nil {
			return err
		}

		projectName := fmt.Sprintf("%s-%s.%s", pomXmlInfo.ArtifactId, pomXmlInfo.Version, pomXmlInfo.Packaging)

		fmt.Println("[INFO] project name ", projectName)
		compileRequest.CompileCmd = append(compileRequest.CompileCmd, &protobuf.BuildCommand{ExecType: string(string(miov1alpha1.Script)),
			Script: fmt.Sprintf("cp target/%s app.%s", projectName, pomXmlInfo.Packaging),
		})

	}

	for _, cmd := range compileRequest.CompileCmd {
		if cmd.ExecType == string(miov1alpha1.Script) {
			fmt.Println("$ compile script:\n", cmd.Script)
			scriptPath, err := pkg_utils.GenScript(cmd.Script)
			if err != nil {
				return err
			}

			if err := execCommand("chmod", []string{"+x", scriptPath}); err != nil {
				return err
			}

			if err := execCommand("sh", []string{"-c", scriptPath}); err != nil {
				fmt.Println("Error compile filed")
				return err
			}
			os.RemoveAll(scriptPath)
			continue
		}

		if err := execCommand(cmd.CommandName, cmd.Params); err != nil {
			fmt.Println("\nError compile filed")
			return err
		}
	}

	fmt.Print("\n[INFO] compile succeed\n")
	return nil
}

func (b *buildConfigServiceImpl) ImageBuild(imageBuildRequest *protobuf.ImageBuildRequest) error {
	fmt.Printf("\n[INFO] image %v start build:\n", imageBuildRequest.Tags)

	buildImage := &docker.Image{
		Tags:       imageBuildRequest.Tags,
		BuildFiles: pkg_utils.GetBuildFileBYDockerfile(imageBuildRequest.DockerFile),
		Username:   imageBuildRequest.Username,
		Password:   imageBuildRequest.Password,
	}

	file, err := os.Create("Dockerfile")
	if err != nil {
		return err
	}
	if _, err := file.Write([]byte(strings.Join(imageBuildRequest.DockerFile, "\n"))); err != nil {
		return err
	}
	file.Close()
	//defer os.RemoveAll("Dockerfile")

	imageBuildResponse, err := b.imageClient.BuildImage(buildImage)
	if err != nil {
		fmt.Printf("\nError image %v build filed\n", imageBuildRequest.Tags)
		return err
	}
	defer imageBuildResponse.Body.Close()

	if _, err := io.Copy(os.Stdout, imageBuildResponse.Body); err != nil {
		return err
	}

	return nil
}

func (b *buildConfigServiceImpl) ImagePush(imagePushRequest *protobuf.ImagePushRequest) error {

	fmt.Printf("\n[INFO] image %v start push:\n", imagePushRequest.Tags)

	for _, imageName := range imagePushRequest.Tags {
		imageInfo := strings.Split(imageName, ":")

		pushImage := &docker.Image{
			Username:  imagePushRequest.Username,
			Password:  imagePushRequest.Password,
			FromImage: imageInfo[0],
			Tag:       imageInfo[1],
		}

		if err := b.imageClient.PushImage(pushImage); err != nil {
			fmt.Printf("\nError image %s push filed\n", imageName)
			return err
		}
	}
	//fmt.Printf("\n[INFO] image %v push succeed\n", imagePushRequest.Tags)
	return nil
}
