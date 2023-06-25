package server

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/fake"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"text/template"
)

func TestCreateOrDetectNamespace(t *testing.T) {
	s := NewServer()
	s.Kube.Client = fake.NewSimpleClientset()
	s.createOrDetectJ8aNamespace()
}

func TestCreateOrDetectServiceTypeLoadBalancer(t *testing.T) {
	s := NewServer()
	s.Kube.Client = fake.NewSimpleClientset()
	s.createOrDetectJ8aServiceTypeLoadBalancer()
}

func TestCreateOrDetectJ8aDeployment(t *testing.T) {
	s := NewServer()
	s.Kube.Client = fake.NewSimpleClientset()
	s.createOrDetectJ8aDeployment()
}

func TestCreateOrDetectJ8aIngressClass(t *testing.T) {
	s := NewServer()
	s.Kube.Client = fake.NewSimpleClientset()
	s.createOrDetectJ8aIngressClass()
}

func TestJ8aReadInitialConfig(t *testing.T) {
	s := getInitialJ8aConfig()
	m := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(s), &m)

	if err != nil {
		t.Errorf("should have decoded initial config without error, got: %v", err)
	}

	if _, ok := m["connection"]; !ok {
		t.Error("config should contain key connection")
	}

	if _, ok := m["resources"]; !ok {
		t.Error("config should contain key resources")
	}

	if _, ok := m["routes"]; !ok {
		t.Error("config should contain key routes")
	}
}

func TestJ8aReadConfigFile(t *testing.T) {
	s := getTemplateJ8aConfig()

	//render a template.
	tmpl, err := template.New("j8aConfigTemplate").Parse(s)
	if err != nil {
		t.Errorf("should have parsed config without error, got: %v", err)
	}

	data := map[string]interface{}{
		"TLS":       "",
		"ROUTES":    "- path: /\n    host: www.emoji😊😊😊.org\n    resource: upstreamresource",
		"RESOURCES": "upstreamresource:\n    - url:\n        scheme: http://\n        host: localhost\n        port: 60083",
	}

	var output strings.Builder
	err = tmpl.Execute(&output, data)
	if err != nil {
		t.Errorf("Failed to execute template: %v", err)
	}
	finalConfig := output.String()
	t.Logf("normal. final config from template parsing\n: %v", finalConfig)

	//find the right j8a to run, arm or amd
	j8a_path, err := locateJ8a()
	if err != nil {
		t.Errorf("unable to build path to executable j8a for testing, cause: %v", err)
	} else {
		t.Logf("located exec: %v", j8a_path)
	}

	//check the config on the cli
	cmd := exec.Command(j8a_path, "-o")
	env := os.Environ()
	env = append(env, fmt.Sprintf("J8ACFG_YML=%v", finalConfig))
	cmd.Env = env

	o, oe := cmd.CombinedOutput()
	if oe != nil {
		t.Errorf("error from j8a validation:\n%v", string(o))
		t.Errorf("unable to validate j8a config using external binary, cause: %v", oe)
	} else {
		t.Logf("success from j8a validation:\n%v", string(o))
	}

}

func locateJ8a() (string, error) {
	//determine arm or amd
	var j8a_path string
	switch runtime.GOOS {
	case "darwin":
		//universal binary for arm and intel Macs
		j8a_path = "resources/tools/j8a_1.1.0_darwin_all/j8a"
	case "arm":
		//linux arm
		j8a_path = "resources/tools/j8a_1.1.0_linux_arm64/j8a"
	default:
		j8a_path = "resources/tools/j8a_1.1.0_linux_amd64/j8a"
	}

	//find the executable in the tree
	cwd, _ := os.Getwd()
	if strings.HasSuffix(cwd, "/server") {
		cwd = cwd[:len(cwd)-7]
	}
	j8a_path = cwd + "/" + j8a_path

	//return fs check on absolute path
	_, err := os.Stat(j8a_path)
	if os.IsNotExist(err) {
		j8a_path = "../" + j8a_path
		_, err = os.Stat(j8a_path)
		if os.IsNotExist(err) {
			return "", err
		}
	}
	return j8a_path, err
}
