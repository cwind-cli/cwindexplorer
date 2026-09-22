package cmd

import (
	"fmt"
	"github.com/cwind-cli/cwind/internal/conductors"
	"github.com/spf13/cobra"
	"os"
)

var doctorCmd = &cobra.Command{Use: "doctor", Short: "Comprueba la instalación local", RunE: func(cmd *cobra.Command, args []string) error {
	if len(conductors.ListMerged()) == 0 {
		return fmt.Errorf("base de conductores vacía")
	}
	if _, err := os.Stat("."); err != nil {
		return fmt.Errorf("directorio actual: %w", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "ok: configuración y base de conductores disponibles")
	return nil
}}

func init() { rootCmd.AddCommand(doctorCmd) }
