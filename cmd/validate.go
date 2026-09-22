package cmd

import (
	"fmt"
	"github.com/cwind-cli/cwind/internal/conductors"
	"github.com/cwind-cli/cwind/internal/geo"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use: "validate", Short: "Valida una traza y un conductor sin consultar weather",
	RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("trace")
		id, _ := cmd.Flags().GetString("conductor")
		if p == "" {
			return fmt.Errorf("falta --trace")
		}
		t, err := geo.ParseTraceFileSimple(p)
		if err != nil {
			return err
		}
		if id != "" {
			conductors.EnsureCalibration()
			cond, ok := conductors.GetMerged(id)
			if !ok {
				return fmt.Errorf("conductor %q no encontrado", id)
			}
			if err := cond.Validate(); err != nil {
				return err
			}
		}
		fmt.Fprintf(cmd.OutOrStdout(), "válido: %d puntos, azimut %.1f°\n", len(t), t.Azimuth())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().String("trace", "", "Archivo CSV o KML")
	validateCmd.Flags().String("conductor", "", "ID del conductor")
}
