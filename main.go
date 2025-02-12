package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/sethvargo/go-envconfig"
	flag "github.com/spf13/pflag"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

type Config struct {
	Log struct {
		Level    string `env:"LOG_LEVEL, default=info"`
		Encoding string `env:"LOG_ENCODING, default=json"`
	}
	File             string            `env:"FILE, default=/dev/stdin"`
	AllowFailure     bool              `env:"ALLOW_FAILURE"`
	LabelSelector    map[string]string `env:"LABEL_SELECTOR"`
	FolderAnnotation string            `env:"FOLDER_ANNOTATION"`
}

var (
	config = &Config{}
)

func init() {
	flag.StringVarP(&config.Log.Level, "log-level", "", "", "Define the log level (default is warning) [debug,info,warn,error]")
	flag.StringVarP(&config.Log.Encoding, "log-encoding", "", "", "Define the log format (default is json) [json,console]")
	flag.StringVarP(&config.File, "file", "f", "", "Path to input")
	flag.BoolVar(&config.AllowFailure, "allow-failure", false, "Do not exit > 0 if an error occurred")
	flag.StringToStringVarP(&config.LabelSelector, "label-selector", "l", nil, "Filter resources by labels")
	flag.StringVarP(&config.FolderAnnotation, "folder-annotation", "a", "", "Name of the folder annotation key")
}

func main() {
	ctx := context.TODO()
	if err := envconfig.Process(ctx, config); err != nil {
		log.Fatal(err)
	}

	flag.Parse()

	logger, err := buildLogger()
	must(err)

	f, err := os.Open(config.File)
	must(err)

	scheme := kruntime.NewScheme()
	factory := serializer.NewCodecFactory(scheme)
	decoder := factory.UniversalDeserializer()

	multidocReader := utilyaml.NewYAMLReader(bufio.NewReader(f))

	selector := &metav1.LabelSelector{
		MatchLabels: config.LabelSelector,
	}

	for {
		resourceYAML, err := multidocReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			must(err)
		}

		if len(resourceYAML) == 0 {
			continue
		}

		obj := corev1.ConfigMap{}
		_, gvk, err := decoder.Decode(
			[]byte(resourceYAML),
			nil,
			&obj)

		must(err)
		logger.V(1).Info("check resource", "gvk", gvk.String())

		if gvk.Kind == "ConfigMap" && gvk.Group == "" && gvk.Version == "v1" {
			logger.V(1).Info("validate configmap", "name", obj.Name, "namespace", obj.Namespace)

			if len(config.LabelSelector) > 0 && !matches(obj.Labels, selector) {
				logger.V(1).Info("skip resource, not matching label selector", "name", obj.Name, "namespace", obj.Namespace)
				continue
			}

			for _, v := range obj.Data {
				d := &dashboard{}
				if err := json.Unmarshal([]byte(v), d); err != nil {
					must(fmt.Errorf("failed unmarshal dashboard %s.%s: %w", obj.Name, obj.Namespace, err))
				}

				if name, ok := obj.Annotations[config.FolderAnnotation]; ok {
					d.Folder = name
					logger.V(1).Info("found folder annotation", "folder", name, "name", obj.Name, "namespace", obj.Namespace)
				}

				if hasUid(d.Uid) {
					must(fmt.Errorf("duplicate uid `%s` found in %s.%s", d.Uid, obj.Name, obj.Namespace))
				}

				if hasTitle(d.Title, d.Folder) {
					must(fmt.Errorf("duplicate name/folder `%s (%s)` found in %s.%s", d.Title, d.Folder, obj.Name, obj.Namespace))
				}

				dashboards = append(dashboards, d)
			}
		}
	}
}

var dashboards []*dashboard

func hasTitle(title, folder string) bool {
	for _, v := range dashboards {
		if v.Title == title && v.Folder == folder {
			return true
		}
	}

	return false
}

func hasUid(uid string) bool {
	if uid == "" {
		return false
	}

	for _, v := range dashboards {
		if v.Uid == uid {
			return true
		}
	}

	return false
}

type dashboard struct {
	Folder string
	Title  string `json:"title"`
	Uid    string `json:"uid"`
}

func matches(labels map[string]string, selector *metav1.LabelSelector) bool {
	if selector == nil {
		return true
	}

	for kS, vS := range selector.MatchLabels {
		var match bool
		for kL, vL := range labels {
			if kS == kL && vS == vL {
				match = true
			}
		}

		if !match {
			return false
		}
	}

	return true
}

func buildLogger() (logr.Logger, error) {
	logOpts := zap.NewDevelopmentConfig()
	logOpts.Encoding = config.Log.Encoding

	err := logOpts.Level.UnmarshalText([]byte(config.Log.Level))
	if err != nil {
		return logr.Discard(), err
	}

	zapLog, err := logOpts.Build()
	if err != nil {
		return logr.Discard(), err
	}

	return zapr.NewLogger(zapLog), nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
