package generate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/budimanjojo/talhelper/pkg/config"
	"github.com/budimanjojo/talhelper/pkg/patcher"
	"github.com/budimanjojo/talhelper/pkg/talos"
	tconfig "github.com/talos-systems/talos/pkg/machinery/config"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1"
)

func GenerateConfig(c *config.TalhelperConfig, outDir, mode string) error {
	var cfg []byte
	var cfgDump tconfig.Provider
	input, err := talos.NewClusterInput(c)
	if err != nil {
		return err
	}

	for _, node := range c.Nodes {
		fileName := c.ClusterName + "-" + node.Hostname + ".yaml"
		cfgFile := outDir + "/" + fileName

		cfg, err = talos.GenerateNodeConfigBytes(&node, input)
		if err != nil {
			return err
		}

		if node.InlinePatch != nil {
			cfg, err = patcher.YAMLInlinePatcher(node.InlinePatch, cfg)
			if err != nil {
				return err
			}
		}

		if len(node.ConfigPatches) != 0 {
			cfg, err = patcher.YAMLPatcher(node.ConfigPatches, cfg)
			if err != nil {
				return err
			}
		}

		if node.ControlPlane {
			cfg, err = patcher.YAMLInlinePatcher(c.ControlPlane.InlinePatch, cfg)
			if err != nil {
				return err
			}
			cfg, err = patcher.YAMLPatcher(c.ControlPlane.ConfigPatches, cfg)
			if err != nil {
				return err
			}
		} else {
			cfg, err = patcher.YAMLInlinePatcher(c.Worker.InlinePatch, cfg)
			if err != nil {
				return err
			}
			cfg, err = patcher.YAMLPatcher(c.Worker.ConfigPatches, cfg)
			if err != nil {
				return err
			}
		}

		cfgDump, err = talos.LoadTalosConfig(cfg)
		if err != nil {
			return nil
		}

		var m v1alpha1.Config
		cfg, err = talos.ReEncodeTalosConfig(cfg, &m)
		if err != nil {
			return nil
		}

		err = dumpFile(cfgFile, cfg)
		if err != nil {
			return err
		}

		fmt.Printf("generated config for %s in %s\n", node.Hostname, cfgFile)
	}

	machineCert := cfgDump.Machine().Security().CA()

	clientCfg, err := talos.GenerateClientConfigBytes(c, input, machineCert)
	if err != nil {
		return nil
	}

	fileName := "talosconfig"
	err = dumpFile(outDir + "/" + fileName, clientCfg)
	if err != nil {
		return err
	}

	fmt.Printf("generated client config in %s\n", outDir + "/" + fileName)

	return nil
}

func dumpFile(path string, file []byte) error {
	dirName := filepath.Dir(path)

	_, err := os.Stat(dirName)
	if err != nil {
		err := os.MkdirAll(dirName, 0700)
		if err != nil {
			return err
		}
	}

	err = os.WriteFile(path, file, 0600)
	if err != nil {
		return err
	}

	return nil
}
