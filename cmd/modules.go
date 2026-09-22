package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type Modulo interface {
	Nombre() string
	Disponible() bool
	Ejecutar(cmd *cobra.Command) error
}

type Sector struct {
	Modulo
	Submodulos []Modulo
}

var sectors []Sector

func RegisterSector(s Sector) {
	sectors = append(sectors, s)
}

func GetSectores() []Sector {
	return sectors
}

func FindSector(nombre string) *Sector {
	for i := range sectors {
		if sectors[i].Nombre() == nombre {
			return &sectors[i]
		}
	}
	return nil
}

func FindSubmodulo(sectorNombre, submoduloNombre string) Modulo {
	sector := FindSector(sectorNombre)
	if sector == nil {
		return nil
	}
	for _, m := range sector.Submodulos {
		if m.Nombre() == submoduloNombre {
			return m
		}
	}
	return nil
}

type moduloBase struct {
	nombre     string
	disponible bool
	ejecutarFn func(*cobra.Command) error
}

func (m *moduloBase) Nombre() string   { return m.nombre }
func (m *moduloBase) Disponible() bool { return m.disponible }
func (m *moduloBase) Ejecutar(c *cobra.Command) error {
	if !m.disponible {
		fmt.Fprintf(c.OutOrStdout(), "\n")
		fmt.Fprintf(c.OutOrStdout(), styleAddMuted.Render("═══════════════════════════════════════════════════════════════════\n"))
		fmt.Fprintf(c.OutOrStdout(), styleWarning.Render("  Este sector está en desarrollo.\n"))
		fmt.Fprintf(c.OutOrStdout(), styleAddMuted.Render("  Consultá las próximas versiones de cwind en cwind-ai.web.app/novedades\n"))
		fmt.Fprintf(c.OutOrStdout(), styleAddMuted.Render("═══════════════════════════════════════════════════════════════════\n"))
		fmt.Fprintf(c.OutOrStdout(), "\n")
		return nil
	}
	if m.ejecutarFn != nil {
		return m.ejecutarFn(c)
	}
	return nil
}

func NuevoModulo(nombre string, disponible bool, fn func(*cobra.Command) error) Modulo {
	return &moduloBase{
		nombre:     nombre,
		disponible: disponible,
		ejecutarFn: fn,
	}
}

type sectorConSubmodulos struct {
	moduloBase
	submodulos []Modulo
}

func (s *sectorConSubmodulos) SubmodulosList() []Modulo {
	return s.submodulos
}

func NuevoSectorConSubmodulos(nombre string, submodulos []Modulo) Sector {
	return Sector{
		Modulo:     &moduloBase{nombre: nombre, disponible: true},
		Submodulos: submodulos,
	}
}
