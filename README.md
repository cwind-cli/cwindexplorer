# cwind

`cwind` es una herramienta CLI para evaluar ampacidad térmica dinámica (DLR) a
partir de una traza georreferenciada, un conductor y un rango de fechas. El
objetivo es apoyar análisis históricos y comparación de escenarios meteorológicos,
no reemplazar mediciones de campo, estudios normativos ni decisiones operativas
sin criterio profesional.

## Alcance y rigor científico

El cálculo utiliza el balance térmico de IEEE Std 738-2023 como núcleo físico:
calentamiento Joule, convección, radiación y calentamiento solar, con datos
meteorológicos históricos de Open-Meteo.
Esto significa que:

- la salida depende de la geometría de la traza,
- del conductor seleccionado,
- de la zona horaria y de los datos meteorológicos del período consultado,
- y de los supuestos del modelo interno.

No debe interpretarse como una medición en tiempo real, validación de
conformidad ni orden operativa. La herramienta es útil para análisis,
comparación de escenarios y revisión técnica preliminar.

### Cómo interpretar la comparación de referencia

La comparación de referencia contrasta el `StaticRating` del conductor con el resultado del
balance térmico en las condiciones declaradas en `RatingConditions`. No es una
prueba de que el conductor "cumple 1%" con IEEE: el rating del fabricante puede
haber sido calculado con otra edición de la norma, otra geometría eléctrica,
otra resistencia AC, otra emisividad/absortividad o condiciones redondeadas.

En particular:

- `Tc`, `Tamb`, `V` y la condición solar deben representar el mismo caso de
  referencia. Si falta uno de ellos, el motor usa un valor predeterminado y el
  resultado no es comparable de forma trazable.
- La comparación usa las propiedades ópticas del conductor y el `Tc`
  extraído de la referencia, en lugar de sustituirlos silenciosamente por los
  valores por defecto del estándar.
- El cálculo operativo no aplica `CalibrationFactor`. El rating declarado queda
  como referencia independiente y la diferencia se reporta como desvío de
  validación. El campo histórico puede conservarse en YAML por compatibilidad,
  pero no modifica la ampacidad física.
- `StaticRating` representa el rating comercial/límite de catálogo usado por
  el flujo operativo existente. `ReferenceRatingA` representa, cuando está
  disponible, el valor exacto que debe usarse para validar una ficha técnica
  concreta. Si `ReferenceRatingA` no está definido, la validación usa
  `StaticRating` como compatibilidad hacia atrás.
- La fuente y la edición de la referencia se conservan en
  `ReferenceRatingSource` y `ReferenceRatingEdition`; no se deben reemplazar
  valores de catálogo sin registrar el documento y sus condiciones.
- En los reportes este valor se presenta como **desvío de validación** y no como
  “margen de error”. Un resultado de `+27%` significa que el cálculo físico
  produjo 27% más ampacidad que el rating declarado en las condiciones de
  referencia; `-3%` significa que produjo 3% menos. No es un margen de seguridad
  ni una incertidumbre estadística.
- La comparación nunca debe leerse como una elección entre dos límites
  operativos. Por ejemplo, `modelo de referencia 433 A` frente a `rating
  declarado 470 A` no autoriza a operar a 433 A ni a 470 A: solo muestra que
  ambos supuestos no fueron reproducidos de forma equivalente. La comparación
  se muestra únicamente en el detalle técnico y se etiqueta
  explícitamente como **no límite operativo**.
- Cuando la validación no reproduce el rating, el límite operativo debe provenir
  de un estudio independiente y documentado, de los límites administrativos,
  térmicos, mecánicos y del equipamiento instalado. Si no existe ese estudio,
  el reporte no debe interpretarse como autorización de corriente.
- Un desvío grande no implica automáticamente que el modelo esté "mal": indica
  que el caso de referencia no está completamente especificado o que el rating
  publicado pertenece a otro conjunto de supuestos. Debe conservarse el
  resultado calculado, el rating, las condiciones, la fuente y la fecha para
  poder auditar la comparación.

Para una validación defendible, primero hay que reproducir una tabla de casos
del fabricante (no solo un punto), verificar unidades y resistencia AC, y
reportar error máximo, sesgo y dispersión. La referencia física del proyecto es
IEEE 738-2023; lo que permanece bajo auditoría es la correspondencia exacta
entre cada ficha comercial y las condiciones completas de ese estándar. cwind
no debe presentarse como una certificación de un rating de fabricante ni como
una autorización operativa.

La traza se procesa localmente. El CLI solo envía coordenadas representativas y
fechas al proveedor meteorológico; no sube el archivo CSV/KML completo.

## Estado de la versión

`v0.1.0-beta.1` es una beta funcional. Está preparada para pruebas técnicas
reproducibles, revisión de código y evaluación de la interfaz, no para
autorizar la operación de una línea.

## Capacidades principales

### Límites administrativos mín/máx separados
- `--admin-min`: derateo manual por calor extremo (ej. 580 A)
- `--admin-max`: límite por TC/equipamiento instalado (ej. 600 A)
- Ambos opcionales, se usan en cascada: DLR < mín → **grave**; mín ≤ DLR < máx → **normal**; máx ≤ DLR < conductor → **moderado**; DLR ≥ conductor → **oportunidad**

### Corrección por pendiente del terreno
- Obtiene elevaciones por vértice usando Open-Elevation API (SRTM)
- Calcula pendientes entre tramos, usa **máxima absoluta**
- Corrige convección natural: `Gr_eff = Gr * cos(θ)` donde θ = atan(pendiente%)

### Comparación con el rating del conductor
- Compara `StaticRating` contra `RatingConditions` como control de trazabilidad
- No modifica la ampacidad física ni fuerza una coincidencia con el fabricante
- Campos YAML: `calibration_error_pct`, `calibration_conditions`, `calibrated_at`

### Elevación y presión
- Promedio de elevaciones de TODOS los vértices ponderado por longitud de tramo
- Se muestra en encabezado: `altura: promedio: X m`

### Reportes
- Formato compacto sin marcos, tabla horizontal con deltas vs mín/máx/conductor
- Estados descriptivos sin depender del color: **grave**, **normal**, **moderado**,
  **oportunidad**
- CSV/JSON incluyen `state`, `delta_min_pct`, `delta_max_pct`, `delta_cond_pct`

### Experiencia de usuario
- Colores terminal: verde cwind (#00d47e), violaceos, naranjas, cian
- Dashboard interactivo de dos paneles: el reporte queda en un viewport
  independiente a la derecha y el wizard/acciones a la izquierda
- Wizard `conductors add` con navegación "← Volver" (Esc)
- 4 opciones de rango: 1h, 12h, 24h, custom
- Banner simplificado (solo logo + versión)

## Instalación

Se distribuye como binario compilado y se publica en GitHub Releases junto con
`checksums.txt`.

```sh
# ver la versión instalada
cwind --version

# o compilar localmente
# go build -o cwind .
```

## Comandos disponibles

```sh
cwind --help
cwind validate --help
cwind report --help
cwind conductors list
cwind conductors add
cwind doctor
```

### Interfaz interactiva dividida

Cuando cwind se ejecuta en una terminal interactiva con al menos 80 columnas,
el flujo de DLR usa un dashboard dividido:

- el panel izquierdo mantiene el contexto de la sesión: conductor, traza,
  límites administrativos y acciones;
- el panel derecho muestra el reporte y sus resultados;
- el reporte tiene viewport independiente: `↑`/`↓`, `j`/`k`, `PgUp`/`PgDn`,
  `Home`/`End` y rueda del mouse desplazan sus líneas;
- las acciones del panel izquierdo se seleccionan con `↑`/`↓` y `Enter`;
- `Esc`/`q` finaliza la sesión sin duplicar el reporte;
- en terminales de menos de 80 columnas, redirecciones, pipes y CI se usa automáticamente la
  salida plana anterior, sin códigos de pantalla ni cambios en CSV/JSON.

La división es una mejora de presentación, no cambia las ecuaciones, los
límites ni los datos exportados.

### 1) `doctor`

Comprueba que la base de conductores y la instalación local estén disponibles.

```sh
cwind doctor
```

### 2) `conductors list`

Lista los conductores registrados en la base local.

```sh
cwind conductors list
```

### 3) `conductors add`

Agrega un conductor personalizado con wizard interactivo. Incluye:
- Datos básicos (nombre, tipo, diámetro, R20, T máx, rating)
- Superficie (emisividad, absortividad)
- ACSR opcional (núcleo acero, áreas Al/Acero)
- Condiciones de rating y fuente
- **Calibración automática** contra rating declarado
- Navegación con "← Volver" (Esc)

```sh
cwind conductors add
```

### 4) `validate`

Valida que la traza sea legible y que el conductor exista y cumpla con los
parámetros mínimos esperados, sin consultar datos meteorológicos.

```sh
cwind validate --trace linea.csv --conductor acsr-795-drake
```

### 5) `report`

Genera un reporte horario de ampacidad dinámica para un rango de fechas.

```sh
cwind report \
  --trace linea.csv \
  --conductor acsr-795-drake \
  --from 2024-01-01 \
  --to 2024-01-02 \
  --format json
```

Se puede especificar `text`, `json` o `csv`.

## Flags importantes del `report`

### `--admin-min` y `--admin-max`

Límites administrativos separados en Amperes (opcionales):

```sh
cwind report --trace linea.csv --conductor alac-240-40 \
  --from 2024-01-01 --to 2024-01-02 \
  --admin-min 580 --admin-max 600
```

El header mostrará: `límite: mín: 580 A, máx: 600 A, conductor: 645 A`

### `--line-azimuth`

Permite fijar manualmente el azimut de la línea en grados. Si no se indica,
`cwind` calcula el azimut a partir de la traza.

```sh
cwind report --trace linea.csv --conductor acsr-795-drake \
  --from 2024-01-01 --to 2024-01-02 \
  --line-azimuth 60
```

### `--timezone`

Define la zona horaria usada para interpretar las fechas y para la consulta de
weather. Debe usar un identificador IANA válido, por ejemplo:

```sh
--timezone America/Argentina/Buenos_Aires
--timezone UTC
```

### `--cache-dir` y `--cache-ttl`

El cache meteorológico es local y opcional. Si se activa, el CLI guarda las
respuestas del proveedor para reutilizarlas en ejecuciones posteriores.

```sh
cwind report \
  --trace linea.csv \
  --conductor acsr-795-drake \
  --from 2024-01-01 \
  --to 2024-01-02 \
  --cache-dir .cwind-cache \
  --cache-ttl 24h
```

- `--cache-dir ""` desactiva el cache.
- `--cache-ttl 0` lo desactiva en forma explícita.
- El cache no contiene la traza ni un archivo de trabajo completo; solo guarda
  respuestas meteorológicas recuperadas del proveedor.

### `--stats`

Muestra percentiles P05/P10/P50/P90 en el resumen.

### `--p05`

Muestra percentil P05 (variante ultra-conservadora).

### `--detalle-tecnico`

Muestra parámetros técnicos del conductor (R20, α, diámetro) y la comparación
con el rating de referencia cuando existe:

```
Comparación de referencia: 644.8 A vs 645 A IMSA  │  desvío: -0.03%
```

## Salida y estructura del reporte

### Formato Text (nuevo formato compacto)

```
conductor:     ALAC 240/40 (IMSA) — familia ACSR/IRAM 2187  (desvío 3.28%)
límite: mín: 580 A, máx: 600 A, conductor: 645 A
período:    2024-01-01 00:00 a 2024-01-01 23:00  (24 h)
traza:    -34.605 -58.395  (altura: promedio: 36 m)  azimut: 141°

hora     sol                     viento      t°amb    t°_cndut (~)    dlr (a)    mín(580a)    máx(600a)    cndut(645a)    estado
────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
15:00     800 w/m²   0.6 m/s  35°c        68°c                   400 a        -31.03%        -33.33%          -37.98%            grave
...
```

**Estados descriptivos:**
- **grave**: DLR < admin_min
- **normal**: admin_min ≤ DLR < admin_max
- **moderado**: admin_max ≤ DLR < conductor
- **oportunidad**: DLR ≥ conductor

Los colores de la interfaz son decorativos y no son necesarios para interpretar
el estado ni sus límites.

**Columnas de deltas:** % vs admin_mín, % vs admin_máx, % vs conductor

### JSON

Incluye nuevos metadatos y por hora `state` + deltas:

```json
{
  "limite_admin_min_a": 580,
  "limite_admin_max_a": 600,
  "limite_conductor_a": 645,
  "elevation_avg_m": 36.5,
  "results": [
    {
      "time": "2024-01-01T15:00:00-03:00",
      "ampacity_a": 400.0,
      "tc_est_at_limit_c": 68.0,
      "state": "grave",
      "delta_min_pct": -31.0,
      "delta_max_pct": -33.3,
      "delta_cond_pct": -38.0
    }
  ]
}
```

### CSV

Nuevas columnas: `state`, `delta_min_pct`, `delta_max_pct`, `delta_cond_pct`
Metadatos nuevos: `# limite_admin_min_a`, `# limite_admin_max_a`, `# limite_conductor_a`, `# elevation_avg_m`

## Flujo recomendado

1. Ejecutar `cwind validate` para verificar la traza y el conductor.
2. Ejecutar `cwind report` con el período de interés.
3. Revisar el resumen y, si corresponde, exportar JSON/CSV.
4. Usar la salida solo como soporte técnico y comparar con estudios locales,
   mediciones de campo y criterios operativos.

## Desarrollo

```sh
go test ./...
go vet ./...
```

Para facilitar la validación local:

```sh
gofmt -w .
go test ./...
go test -race ./...
go vet ./...
```

Los builds de CI cubren Windows, Linux y macOS en amd64/arm64. Los binarios
oficiales se publican en GitHub Releases con un archivo `checksums.txt`, que los
instaladores deben verificar antes de instalar el binario final.