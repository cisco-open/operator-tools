// Copyright © 2020 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package reconciler_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/banzaicloud/operator-tools/pkg/utils"
	"github.com/go-logr/logr"
	"github.com/pborman/uuid"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var testNamespace = "test-" + uuid.New()[:8]
var controlNamespace = "control"
var log logr.Logger

func TestMain(m *testing.M) {
	log = utils.Log
	utils.GlobalLogLevel = 2
	err := beforeSuite()
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
	code := m.Run()
	err = afterSuite()
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
	os.Exit(code)
}

func beforeSuite() error {
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
	}

	var err error

	cfg, err = testEnv.Start()
	if err != nil {
		return err
	}
	if cfg == nil {
		return fmt.Errorf("failed to start testenv, config is nil")
	}

	k8sClient, err = client.New(cfg, client.Options{Scheme: clientgoscheme.Scheme})
	if err != nil {
		return err
	}
	if k8sClient == nil {
		return fmt.Errorf("failed to create k8s config")
	}

	for _, ns := range []string{controlNamespace, testNamespace} {
		err := k8sClient.Create(context.TODO(), &corev1.Namespace{
			ObjectMeta: v1.ObjectMeta{
				Name: ns,
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func afterSuite() error {
	return testEnv.Stop()
}

func assertConfigMapList(t *testing.T, a func(l *corev1.ConfigMapList)) {
	l := &corev1.ConfigMapList{}

	if err := k8sClient.List(context.TODO(), l, client.InNamespace(controlNamespace)); err != nil {
		t.Fatalf("+%v", err)
	}

	a(l)
}

func assertSecretList(t *testing.T, a func(l *corev1.SecretList)) {
	l := &corev1.SecretList{}

	if err := k8sClient.List(context.TODO(), l, client.InNamespace(controlNamespace)); err != nil {
		t.Fatalf("+%v", err)
	}

	a(l)
}