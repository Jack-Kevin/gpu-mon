/*
 * Copyright 2018 Xiaomi, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package common

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/micro/go-config"
	"github.com/micro/go-config/source/file"
)

type falconConf struct {
	Agent string `json:"agent"`
}

type metricConf struct {
	IgnoreMetrics []string `json:"ignoreMetrics"`
	EndPoint      string   `json:"endpoint"`
}

type logConf struct {
	Level string `json:"level"`
	Dir   string `json:"dir"`
}

//Conf 用户定义的配置
type Conf struct {
	Falcon       falconConf
	Metric       metricConf
	Log          logConf
	MetricFilter map[string]struct{}
	IsCrontab    bool
}

var (
	globalConf Conf
	configLock = new(sync.RWMutex)
)

// Config 返回全局配置
func Config() *Conf {
	configLock.Lock()
	defer configLock.Unlock()
	return &globalConf
}

func getCurrpath() string {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return ""
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return ""
	}
	index := strings.LastIndex(path, string(os.PathSeparator))
	currPath := path[:index]
	return currPath
}

func readConfigFile(configPath string) (err error) {
	if !fileExist(configPath) {
		return fmt.Errorf("config file %s is not exicted", configPath)
	}
	err = config.Load(file.NewSource(file.WithPath(configPath)))
	if err != nil {
		return err
	}
	return initGlobalConfig()
}

func initGlobalConfig() (err error) {
	err = config.Get("falcon").Scan(&globalConf.Falcon)
	if err != nil {
		return err
	}
	err = config.Get("metric").Scan(&globalConf.Metric)
	if err != nil {
		return err
	}
	err = config.Get("log").Scan(&globalConf.Log)
	if err != nil {
		return err
	}

	logdir := globalConf.Log.Dir
	if !filepath.IsAbs(logdir) {
		currentPath := getCurrpath()
		globalConf.Log.Dir = filepath.Join(currentPath, logdir)
	}

	return err
}

func initIgnoreMetrics() {
	metricFilter := make(map[string]struct{})
	for _, metric := range globalConf.Metric.IgnoreMetrics {
		metricFilter[metric] = struct{}{}
	}
	globalConf.MetricFilter = metricFilter

}

func initConfig(configPath string, isCrontab bool) error {
	err := readConfigFile(configPath)
	if err != nil {
		return fmt.Errorf("unable to read Config file: %v", err)
	}
	globalConf.IsCrontab = isCrontab
	initIgnoreMetrics()
	return nil
}
