package cmd

import (
	"log"
	"os"

	"github.com/budimanjojo/talhelper/pkg/config"
	"github.com/budimanjojo/talhelper/pkg/secret"
	"github.com/spf13/cobra"
	"github.com/talos-systems/talos/pkg/machinery/config/configloader"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/generate"
	"gopkg.in/yaml.v3"
)

var (
	gensecretPatchCfg bool
	gensecretFromCfg  string
	gensecretCfgFile  string
)

var gensecretCmd = &cobra.Command{
	Use:   "gensecret",
	Short: "Generate Talos cluster secrets",
	Run: func(cmd *cobra.Command, args []string) {
		var s *generate.SecretsBundle
		var err error
		switch gensecretFromCfg {
		case "":
			s, err = secret.NewSecretBundle(generate.NewClock())
			if err != nil {
				log.Fatalf("failed to generate secret bundle: %s", err)
			}
		default:
			cfg, err := configloader.NewFromFile(gensecretFromCfg)
			if err != nil {
				log.Fatalf("failed to load Talos cluster node config file: %s", err)
			}

			s = secret.NewSecretFromCfg(generate.NewClock(), cfg)
		}

		secret.PrintSortedSecrets(s)

		if gensecretPatchCfg {
			cf, err := os.ReadFile(gensecretCfgFile)
			if err != nil {
				log.Fatalf("failed to read file %s: %s", gensecretCfgFile, err)
			}

			var m config.TalhelperConfig

			err = yaml.Unmarshal(cf, &m)
			if err != nil {
				log.Fatalf("failed to unmarshal config file: %s", err)
			}

			cfg, err := m.ApplyInlinePatch([]byte(secret.SecretPatch))
			if err != nil {
				log.Fatalf("failed to patch config file %s: %s", gensecretCfgFile, err)
			}

			cfg, err = m.Encode(cfg)
			if err != nil {
				log.Fatalf("failed to encode config file %s: %s", gensecretCfgFile, err)
			}

			err = os.WriteFile(gensecretCfgFile, cfg, 0700)
			if err != nil {
				log.Fatalf("failed to write config file %s: %s", gensecretCfgFile, err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(gensecretCmd)

	gensecretCmd.Flags().StringVarP(&gensecretCfgFile, "config-file", "c", "talconfig.yaml", "File containing configurations for talhelper")
	gensecretCmd.Flags().StringVarP(&gensecretFromCfg, "from-configfile", "f", "", "Talos cluster node configuration file to generate secret from")
	gensecretCmd.Flags().BoolVarP(&gensecretPatchCfg, "patch-configfile", "p", false, "Whether to generate inline patches into config file")
}
