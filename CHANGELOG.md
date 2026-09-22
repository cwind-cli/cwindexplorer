# Changelog

## 0.1.0-beta.1

Primera beta funcional para análisis histórico y comparación de escenarios de
ampacidad térmica dinámica.

### Incluye

- CLI interactiva con wizard de selección.
- Dashboard TUI de dos paneles para el flujo DLR.
- Reporte con scroll independiente, teclado y rueda del mouse.
- Salida text, JSON y CSV.
- Validación de trazas y base de conductores.
- Consulta meteorológica histórica mediante Open-Meteo.
- Cálculo de azimut, pendiente y presión atmosférica derivada de la elevación
  para el balance térmico.
- Separación entre cálculo físico y factor histórico de calibración.
- Referencias documentadas para conductores PRYSAL y ACSR Drake sin reemplazar
  automáticamente los ratings comerciales.
- Tests unitarios, integración de CLI, `go vet` y workflow de CI.

### Alcance

Esta beta utiliza IEEE Std 738-2023 como núcleo del balance térmico. No se
presenta como certificación de ratings comerciales, medición de campo, estudio
normativo completo o autorización operativa.
