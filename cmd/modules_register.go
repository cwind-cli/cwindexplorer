package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	RegisterSector(NuevoSectorConSubmodulos("Transmisión", []Modulo{
		NuevoModulo("Conductores (DLR)", true, ejecutarConductoresDLR),
		NuevoModulo("Transformadores (DTR)", false, nil),
	}))

	RegisterSector(NuevoSectorConSubmodulos("Generación", []Modulo{
		NuevoModulo("Eólica", false, nil),
		NuevoModulo("Solar", false, nil),
	}))
}

func ejecutarConductoresDLR(cmd *cobra.Command) error {
	tracePath, err := askTraceFile()
	if err != nil {
		return err
	}

	cond, condID, err := askConductor()
	if err == errBack {
		return fmt.Errorf("volver al menú anterior")
	}
	if err != nil {
		return err
	}

	adminMin, adminMax, err := askAdminLimit()
	if err != nil {
		return err
	}

	from, to, err := askDateRange()
	if err != nil {
		return err
	}

	return runReportDirect(cmd, tracePath, cond, condID, adminMin, adminMax, from, to)
}
