package scheduler

import (
	"fmt"
	"github.com/prometheus/common/log"
	"gopkg.in/src-d/go-git.v4"
	"hidevops.io/hiboot/pkg/app"
	utilsio "hidevops.io/hiboot/pkg/utils/io"
	"hidevops.io/hioak/starter/kube"
	"hidevops.io/mio/node/protobuf"
	"hidevops.io/mio/node/service"
	pkg_utils "hidevops.io/mio/node/utils"
	corev1 "k8s.io/api/core/v1"
	"os"
	"time"
)

type buildSchedulerImpl struct {
	BuildService
	buildConfigService service.BuildConfigService
	buildConfigClient  service.BuildConfigClient
	secret             *kube.Secret
	token              kube.Token
}

func newBuildSchedulerImpl(buildConfigService service.BuildConfigService,
	buildConfigClient service.BuildConfigClient, secret *kube.Secret, token kube.Token) BuildService {
	return &buildSchedulerImpl{
		buildConfigClient:  buildConfigClient,
		buildConfigService: buildConfigService,
		secret:             secret,
		token:              token,
	}
}

func init() {
	app.Component(newBuildSchedulerImpl)
}

type BuildService interface {
	SourceCodePull(sourceCodePullRequest *protobuf.SourceCodePullRequest)
	SourceCodeCompiles(compileRequest *protobuf.CompileRequest)
	SourceCodeImageBuild(imageBuildRequest *protobuf.ImageBuildRequest)
	SourceCodeImagePush(imagePushRequest *protobuf.ImagePushRequest)
	SourceCodeTest(testRequest *protobuf.TestsRequest)
	EnvMakeUp(testRequest *protobuf.CommandRequest)
}

var dataCache = make(map[string]interface{})

const (
	CODEPATH string = "codepath"
)

func fmtName(str, namespace, name string) string {
	return fmt.Sprintf("%s-%s-%s", str, namespace, name)
}

func (b *buildSchedulerImpl) SourceCodePull(sourceCodePullRequest *protobuf.SourceCodePullRequest) {

	//////////////////////////
	scriptPath := os.Getenv("Script_Path")
	log.Debug("script path", scriptPath)
	if scriptPath != "" {
		sourceCodeTestRequest := protobuf.TestsRequest{
			TestCmd: []*protobuf.TestCommand{
				&protobuf.TestCommand{CommandName: "chmod", Params: []string{"+x", scriptPath}},
				&protobuf.TestCommand{CommandName: "sh", Params: []string{"-c", scriptPath}},
			},
		}
		if err := pkg_utils.TestStart(&sourceCodeTestRequest); err != nil {
			log.Debug("Error", fmt.Errorf("script %s start failed", scriptPath))
			log.Debug("Error", err)
			//return
		}
		time.Sleep(time.Second * 3)
	}
	//////////////////////

	//TODO update status
	bc, err := b.buildConfigClient.UpdateBuildStatus(sourceCodePullRequest.Namespace, sourceCodePullRequest.Name, service.SourceCodePull, service.Created)
	if err != nil {
		fmt.Println("Error ", err)
		////TODO update status
		b.buildConfigClient.UpdateBuildStatus(sourceCodePullRequest.Namespace, sourceCodePullRequest.Name, service.SourceCodePull, service.Fail)
		return
	}

	secret, err := b.secret.Get(bc.Labels["mio.io/buildConfig.name"], sourceCodePullRequest.Namespace)
	if err != nil {
		fmt.Println("Error ", err)
		////TODO update status
		b.buildConfigClient.UpdateBuildStatus(sourceCodePullRequest.Namespace, sourceCodePullRequest.Name, service.SourceCodePull, service.Fail)
		return
	}
	sourceCodePullRequest.Username = string(secret.Data[corev1.BasicAuthUsernameKey])
	sourceCodePullRequest.Password = string(secret.Data[corev1.BasicAuthPasswordKey])
	sourceCodePullRequest.Token = string(secret.Data[corev1.ServiceAccountTokenKey])

	codePath, err := b.buildConfigService.Clone(sourceCodePullRequest, git.PlainClone)
	if err != nil {
		fmt.Println("Error ", err)
		//TODO update status
		b.buildConfigClient.UpdateBuildStatus(sourceCodePullRequest.Namespace, sourceCodePullRequest.Name, service.SourceCodePull, service.Fail)
	}
	//TODO update status
	b.buildConfigClient.UpdateBuildStatus(sourceCodePullRequest.Namespace, sourceCodePullRequest.Name, service.SourceCodePull, service.Success)

	fmt.Println("code path", codePath)
	dataCache[fmtName(CODEPATH, sourceCodePullRequest.Namespace, sourceCodePullRequest.Name)] = codePath
	fmt.Println("data cache: ", dataCache)
}

func (b *buildSchedulerImpl) SourceCodeCompiles(compileRequest *protobuf.CompileRequest) {

	codePathKey := fmtName(CODEPATH, compileRequest.Namespace, compileRequest.Name)

	if _, ok := dataCache[codePathKey]; !ok {
		fmt.Println(fmt.Errorf("code path: %s not found", codePathKey))
		//TODO update status
		b.buildConfigClient.UpdateBuildStatus(compileRequest.Namespace, compileRequest.Name, service.Compile, service.Fail)
		return
	} else {
		if err := utilsio.ChangeWorkDir(dataCache[codePathKey].(string)); err != nil {
			fmt.Println("Error ", err)
			//TODO update status
			b.buildConfigClient.UpdateBuildStatus(compileRequest.Namespace, compileRequest.Name, service.Compile, service.Fail)
		}
	}

	if err := b.buildConfigService.Compile(compileRequest); err != nil {
		fmt.Println("Error ", err)
		//TODO update status
		b.buildConfigClient.UpdateBuildStatus(compileRequest.Namespace, compileRequest.Name, service.Compile, service.Fail)
		return
	} else {
		//TODO update status
		b.buildConfigClient.UpdateBuildStatus(compileRequest.Namespace, compileRequest.Name, service.Compile, service.Success)
	}
}

func (b *buildSchedulerImpl) SourceCodeImageBuild(imageBuildRequest *protobuf.ImageBuildRequest) {
	codePathKey := fmtName(CODEPATH, imageBuildRequest.Namespace, imageBuildRequest.Name)

	if _, ok := dataCache[codePathKey]; !ok {
		fmt.Println(fmt.Errorf("code path not found"))
		//TODO update status
		b.buildConfigClient.UpdateBuildStatus(imageBuildRequest.Namespace, imageBuildRequest.Name, service.ImageBuild, service.Fail)
		return
	}

	if err := b.buildConfigService.ImageBuild(imageBuildRequest); err != nil {
		fmt.Println("Error ", err)
		//TODO update status
		b.buildConfigClient.UpdateBuildStatus(imageBuildRequest.Namespace, imageBuildRequest.Name, service.ImageBuild, service.Fail)
		return
	} else {
		//TODO update status
		b.buildConfigClient.UpdateBuildStatus(imageBuildRequest.Namespace, imageBuildRequest.Name, service.ImageBuild, service.Success)
	}

	//delete(dataCache, codePathKey)
}

func (b *buildSchedulerImpl) SourceCodeImagePush(imagePushRequest *protobuf.ImagePushRequest) {

	if "unused" == imagePushRequest.Username {
		imagePushRequest.Password = fmt.Sprintf("%s", b.token)
	}

	if err := b.buildConfigService.ImagePush(imagePushRequest); err != nil {
		fmt.Println("Error ", err)
		//TODO update status
		b.buildConfigClient.UpdateBuildStatus(imagePushRequest.Namespace, imagePushRequest.Name, service.ImagePush, service.Fail)
		return
	} else {
		//TODO update status
		b.buildConfigClient.UpdateBuildStatus(imagePushRequest.Namespace, imagePushRequest.Name, service.ImagePush, service.Success)
	}
}

func (b *buildSchedulerImpl) SourceCodeTest(testRequest *protobuf.TestsRequest) {
	codePathKey := fmtName(CODEPATH, testRequest.Namespace, testRequest.Name)

	if _, ok := dataCache[codePathKey]; !ok {
		fmt.Println(fmt.Errorf("code path not found"))
		//TODO update status
		b.buildConfigClient.UpdateBuildStatus(testRequest.Namespace, testRequest.Name, service.SourceCodeTest, service.Fail)
		return
	}

	if err := pkg_utils.TestStart(testRequest); err != nil {
		fmt.Println("Error", err)
		//TODO update status
		b.buildConfigClient.UpdateBuildStatus(testRequest.Namespace, testRequest.Name, service.SourceCodeTest, service.Fail)
	}

	b.buildConfigClient.UpdateBuildStatus(testRequest.Namespace, testRequest.Name, service.SourceCodeTest, service.Success)
}

func (b *buildSchedulerImpl) EnvMakeUp(commandRequest *protobuf.CommandRequest) {

	codePathKey := fmtName(CODEPATH, commandRequest.Namespace, commandRequest.Name)

	if _, ok := dataCache[codePathKey]; !ok {
		fmt.Println(fmt.Errorf("code path not found"))
		//TODO update status
		b.buildConfigClient.UpdateBuildStatus(commandRequest.Namespace, commandRequest.Name, service.Command, service.Fail)
		return
	}

	if err := pkg_utils.StartCmd(commandRequest); err != nil {
		fmt.Println("Error", err)
		//TODO update status
		b.buildConfigClient.UpdateBuildStatus(commandRequest.Namespace, commandRequest.Name, service.Command, service.Fail)
	}

	b.buildConfigClient.UpdateBuildStatus(commandRequest.Namespace, commandRequest.Name, service.Command, service.Success)
}
