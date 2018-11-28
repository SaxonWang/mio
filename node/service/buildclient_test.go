package service

//go:generate mockgen -destination mock/mock_buildclient.go -package mock hidevops.io/mio/node/pkg/service BuildConfigClient

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"hidevops.io/hiboot/pkg/log"
	"hidevops.io/mio/node/service/mock"
	"hidevops.io/mioclient/pkg/apis/mio/v1alpha1"
	"hidevops.io/mioclient/pkg/client/clientset/versioned/fake"
	"hidevops.io/mioclient/starter/mio"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func TestBuildClient(t *testing.T) {
	projectName := "example-buildconfig17"

	build := &v1alpha1.Build{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BuildConfig",
			APIVersion: "mio.k8s.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      projectName,
			Namespace: "default",
			Labels:    map[string]string{"app": projectName},
		},
		Spec: v1alpha1.BuildSpec{
			App:       projectName,
			CodeType:  "golang",
			CloneType: "http",
			BaseImage: "FROM ubuntu:16.04",
			Tags:      []string{"test:2.1"},

			CompileCmd: []v1alpha1.CompileCmd{},

			DockerFile: []string{
				"FROM ubuntu:16.04",
				"WORKDIR /vpcloue",
				fmt.Sprintf("ADD %s /vpcloue/", projectName),
				fmt.Sprintf("BuildCommand /vpcloue/%s", projectName)},
			CloneConfig: v1alpha1.BuildCloneConfig{
				Url:      "http://gitlab.vpclub:8022/wanglulu/clone-test.git",
				Branch:   "master",
				DstDir:   "/Users/mac/.gvm/pkgsets/go1.10/vpcloud/src/hidevops.io/",
				Username: "",
				Password: "",
			},
		},
	}

	clientSet := fake.NewSimpleClientset().MioV1alpha1()
	buildClientSet := mio.NewBuild(clientSet)

	b := newBuildConfigClient(buildClientSet)

	t.Run("should create build success", func(t *testing.T) {
		_, err := b.Create(build)
		assert.Equal(t, nil, err)
	})

	t.Run("should get build success", func(t *testing.T) {
		_, err := b.Get(build.Name, build.Namespace)
		assert.Equal(t, nil, err)
	})

	t.Run("should watch build success", func(t *testing.T) {
		_, err := b.Watch(metav1.ListOptions{}, build.Namespace, build.Name)
		assert.Equal(t, nil, err)
	})

	t.Run("should update build success", func(t *testing.T) {
		_, err := b.Update(build.Name, build.Namespace, build)
		assert.Equal(t, nil, err)
	})

	t.Run("should UpdateBuildStatus build success", func(t *testing.T) {
		_, err := b.UpdateBuildStatus(build.Namespace, build.Name, SourceCodePull, Fail)
		assert.Equal(t, nil, err)
	})

	t.Run("should delete build success", func(t *testing.T) {
		err := b.Delete(build.Name, build.Namespace)
		assert.Equal(t, nil, err)
	})

}

func TestClient(t *testing.T) {
	projectName := "example-buildconfig17"

	build := &v1alpha1.Build{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BuildConfig",
			APIVersion: "mio.k8s.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      projectName,
			Namespace: "default",
			Labels:    map[string]string{"app": projectName},
		},
		Spec: v1alpha1.BuildSpec{
			App:       projectName,
			CodeType:  "golang",
			CloneType: "http",
			BaseImage: "FROM ubuntu:16.04",
			Tags:      []string{"test:2.1"},

			CompileCmd: []v1alpha1.CompileCmd{},

			DockerFile: []string{
				"FROM ubuntu:16.04",
				"WORKDIR /vpcloue",
				fmt.Sprintf("ADD %s /vpcloue/", projectName),
				fmt.Sprintf("BuildCommand /vpcloue/%s", projectName)},
			CloneConfig: v1alpha1.BuildCloneConfig{
				Url:      "http://gitlab.vpclub:8022/wanglulu/clone-test.git",
				Branch:   "master",
				DstDir:   "/Users/mac/.gvm/pkgsets/go1.10/vpcloud/src/hidevops.io/",
				Username: "",
				Password: "",
			},
		},
	}
	fmt.Println(build)

	mockCtl := gomock.NewController(t)
	m := mock.NewMockBuildConfigClient(mockCtl)

	t.Run("should create build success", func(t *testing.T) {
		m.EXPECT().Create(build).Return(build, nil)
		_, err := m.Create(build)
		assert.Equal(t, nil, err)
	})

	t.Run("should get build success", func(t *testing.T) {
		m.EXPECT().Get("", "").Return(build, nil)
		_, err := m.Get("", "")
		assert.Equal(t, nil, err)
	})

	t.Run("should watch build success", func(t *testing.T) {
		m.EXPECT().Watch(metav1.ListOptions{}, "", "").Return(nil, nil)
		_, err := m.Watch(metav1.ListOptions{}, "", "")
		assert.Equal(t, nil, err)
	})

	t.Run("should update build success", func(t *testing.T) {
		m.EXPECT().Update("", "", nil).Return(nil, nil)
		_, err := m.Update("", "", nil)
		assert.Equal(t, nil, err)
	})

	t.Run("should UpdateBuildStatus build success", func(t *testing.T) {

		m.EXPECT().UpdateBuildStatus("", "", "", "").Return(nil, nil)
		_, err := m.UpdateBuildStatus("", "", "", "")

		assert.Equal(t, nil, err)
	})

	t.Run("should delete build success", func(t *testing.T) {
		m.EXPECT().Delete("", "").Return(nil)
		err := m.Delete("", "")
		assert.Equal(t, nil, err)
	})

}
