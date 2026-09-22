# Validación técnica de conductores para cwind (IEEE 738)

**Fecha de consulta de todas las fuentes:** 21 de septiembre de 2026.

**Marco físico del proyecto:** cwind utiliza IEEE Std 738-2023 como referencia
central para el balance térmico del conductor: calentamiento Joule, convección,
radiación y calentamiento solar. La ecuación térmica no se sustituye por la
norma de fabricación del conductor. IRAM, ASTM, IEC y las fichas de fabricante
se utilizan para identificar la construcción, el diámetro, las áreas, la
resistencia y los ratings publicados del conductor seleccionado.

**Alcance de esta investigación:** se revisaron fichas técnicas y catálogos de
fabricante disponibles públicamente (Prysmian Argentina "Prysalac", Nexans,
Southwire y APAR-India). La falta de un dato en una ficha no significa que
IEEE 738 no lo defina: significa que ese parámetro no quedó identificado para
ese producto en la documentación técnica consultada. Cuando un valor no pudo
confirmarse directamente en la ficha correspondiente, se lo marcó como
**NO ENCONTRADO** o con confianza **MEDIA**, sin completar el dato por
inferencia.

**Criterio de implementación:** la implementación debe auditarse ecuación por
ecuación contra IEEE 738-2023. Por eso este documento distingue entre la
referencia física IEEE 738-2023 y la reproducibilidad del rating comercial. No
se debe afirmar conformidad certificada de un rating de fabricante si la ficha
no publica todas las condiciones necesarias para reproducirlo.

**Nota crítica de reproducibilidad:** en NINGÚN caso se encontró la condición
térmica completa que exige el balance de IEEE 738 (Tc, Tamb, viento, ángulo de
viento, radiación solar, elevación solar y presión/altitud) publicada en conjunto
por un mismo fabricante. Esto limita la reproducción del rating comercial, no el
uso de IEEE 738 como núcleo físico del cálculo de cwind.

**Separación de ratings en cwind:** `StaticRating` no debe reemplazarse
automáticamente por un valor encontrado en otra ficha. Para validar un caso
concreto, cwind debe cargar ese valor como `ReferenceRatingA`, junto con
`ReferenceRatingSource`, `ReferenceRatingEdition` y las condiciones exactas de
la fuente. Así se evita mezclar rating comercial, referencia de validación y
límite operativo.

### Casos incorporados en cwind

La base de datos incorpora ahora tres registros trazables:

| ID | Caso | Rating de referencia | Fuente | Alcance |
|---|---|---:|---|---|
| `prysal-150` | PRYSAL AAAC, 37×2,25 mm, IRAM 2212 | 386 A | Prysmian Argentina, PRYSAL, edición 2021 | Caso publicado por TR IEC 61597:95; faltan ángulo de viento, elevación solar y presión numérica |
| `prysal-240` | PRYSAL AAAC, 37×2,85 mm, IRAM 2212 | 520 A | Prysmian Argentina, PRYSAL, edición 2021 | Caso publicado por TR IEC 61597:95; faltan ángulo de viento, elevación solar y presión numérica |
| `acsr-795-drake` | ACSR Drake 26/7, caso Nexans | 907 A | Nexans/Centelsa | Caso parcial; se conserva aparte del rating Prysmian de 907 A bajo condiciones incompatibles |

Los registros PRYSAL no reemplazan los registros `cadla-*`. El nombre
nominal `AAAC 150` o `AAAC 240` no es suficiente para identificar una
geometría. La comparación de Drake tampoco demuestra que 907 A sea un límite
operativo: es únicamente el rating de referencia del caso Nexans cargado.

---

## 1. Tabla comparativa general

| conductor | fabricante (fuente) | tipo | diámetro_m | área_al_mm2 | área_acero_mm2 | diámetro_núcleo_m | R20_ohm_m | RAC_ohm_m | alpha_1_C | temperatura_max_C | rating_A | Tc_C | Tamb_C | viento_m_s | ángulo_viento_deg | radiación_W_m2 | elevación_solar_deg | presión_Pa | norma | edición | confianza | reproducibilidad |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| ACSR 150/25 | Prysmian Arg. (Prysalac) | ACSR | 0,0171 | 150 (nominal) | 25 (nominal) | NO ENCONTRADO | 0,000194 | NO ENCONTRADO | NO ENCONTRADO (normativo Al 1350-H19 ≈0,00403/°C) | NO ENCONTRADO | 415 | NO ENCONTRADO | 40 | 0,6 | NO ENCONTRADO | NO ENCONTRADO | NO ENCONTRADO | NO ENCONTRADO | IRAM 2187 | Dic. 2008 | MEDIO | NO reproducible |
| ACSR 240/40 | Prysmian Arg. (Prysalac) | ACSR | 0,0219 | 240 (nominal) | 40 (nominal) | NO ENCONTRADO | 0,000119 | NO ENCONTRADO | NO ENCONTRADO (ídem) | NO ENCONTRADO | 565 | NO ENCONTRADO | 40 | 0,6 | NO ENCONTRADO | NO ENCONTRADO | NO ENCONTRADO | NO ENCONTRADO | IRAM 2187 | Dic. 2008 | MEDIO | NO reproducible |
| ACSR 795 kcmil Drake | Nexans / Prysmian Arg. / Southwire | ACSR | 0,02813 | 402,92 | ~52,2 (calc., ver nota) | NO ENCONTRADO | 0,000071 (Nexans) / 0,0000702 (Prysmian) | 0,0000863 (distribuidor, MEDIO) | NO ENCONTRADO (normativo Al 1350-H19 ≈0,00403/°C) | NO ENCONTRADO | 907 (2 fuentes coinciden, condiciones distintas – ver contradicción) | 75 (solo Nexans) | 25 (Nexans) / 40 (Prysmian) | 0,61 (Nexans) / 0,6 (Prysmian) | NO ENCONTRADO | 1000 (Nexans, "1kW/m²") | NO ENCONTRADO | nivel del mar (ambas) | ASTM B232 | N/D en fuente | ALTO (geometría) / MEDIO (rating) | PARCIAL — ver sección individual |
| ACSR 300/50 | Prysmian Arg. (Prysalac) | ACSR | 0,0245 | 300 (nominal) | 50 (nominal) | NO ENCONTRADO | 0,0000949 | NO ENCONTRADO | NO ENCONTRADO | NO ENCONTRADO | 650 | NO ENCONTRADO | 40 | 0,6 | NO ENCONTRADO | NO ENCONTRADO | NO ENCONTRADO | NO ENCONTRADO | IRAM 2187 | Dic. 2008 | MEDIO | NO reproducible |
| ACSR 435/55 | Prysmian Arg. (Prysalac) | ACSR | 0,0288 | 435 (nominal) | 55 (nominal) | NO ENCONTRADO | 0,0000666 | NO ENCONTRADO | NO ENCONTRADO | NO ENCONTRADO | 765 | NO ENCONTRADO | 40 | 0,6 | NO ENCONTRADO | NO ENCONTRADO | NO ENCONTRADO | NO ENCONTRADO | IRAM 2187 | Dic. 2008 | MEDIO | NO reproducible |
| ACSR 477 kcmil Hawk | Prysmian Arg. (Prysalac) | ACSR | 0,02178 | 242 (nominal, kcmil) | NO ENCONTRADO (numérico) | NO ENCONTRADO | 0,000117 | NO ENCONTRADO | NO ENCONTRADO | NO ENCONTRADO | 659 | NO ENCONTRADO | 40 | 0,6 | NO ENCONTRADO | NO ENCONTRADO | NO ENCONTRADO | NO ENCONTRADO | ASTM B232 | N/D en fuente | MEDIO | NO reproducible |
| AAAC 150 mm² (EN 50182 AL3 "150/37"; CADLA IRAM 2212 — geometría cruzada) | APAR (India) — datasheet EN 50182 Tipo AL3; geometría también IRAM 2212 vía Industrias MH (Arg.) | AAAC (6201-T81) | 0,0158 | 147,1 (real, no 150 exacto) | — (no aplica, sin alma de acero) | — | 0,0002256 | NO ENCONTRADO (solo I admisible a Tc dado) | 0,0036 (normativo, fuente secundaria) | NO ENCONTRADO (fabricante da 130 °C cortocircuito y 90 °C IRAM 2177 en fuente distinta, no confirmado para esta ficha) | 284 (@Tc 75°C) / 348 (@Tc 85°C) | 75 / 85 (dos valores) | 45 | 0,56 | NO ENCONTRADO | 1045 | NO ENCONTRADO | NO ENCONTRADO | EN 50182 tipo AL3 | N/D en fuente | ALTO (geometría/rating APAR) / MEDIO (cruce con IRAM 2212) | REPRODUCIBLE (condiciones completas salvo ángulo viento, elevación solar y presión) |
| AAAC 240 mm² (EN 50182 AL3 — 2 variantes nacionales distintas) | APAR (India) — datasheet EN 50182 Tipo AL3; geometría también IRAM 2212 vía Industrias MH (Arg.) | AAAC (6201-T81) | 0,0203 (var. DE/AT, 61 hilos) / 0,0203 (var. IT, 37 hilos) | 242,5 (DE/AT) / 244,4 (IT) | — | — | 0,0001373 (DE/AT) / 0,0001358 (IT) | NO ENCONTRADO | 0,0036 (normativo, fuente secundaria) | NO ENCONTRADO | 378 (DE/AT @75°C) / 380 (IT @75°C) | 75 / 85 | 45 | 0,56 | NO ENCONTRADO | 1045 | NO ENCONTRADO | NO ENCONTRADO | EN 50182 tipo AL3 | N/D en fuente | ALTO (APAR) / MEDIO (cruce IRAM) | REPRODUCIBLE (condiciones completas salvo ángulo viento, elevación solar y presión) |

**Notas a la tabla:**
- Las secciones de los conductores IRAM (150/25, 240/40, 300/50, 435/55) son **designaciones nominales**, no áreas medidas independientemente en la fuente consultada; no se recalcularon a partir del diámetro de hilo porque el catálogo no publica el área efectiva por separado.
- El área de acero de Drake según Nexans es 455,11 mm² (total) − 402,92 mm² (aluminio) = 52,19 mm² (cálculo propio, mostrado explícitamente; no aparece como dato directo de fábrica).
- La fila de Hawk no trae área de acero individual en la fuente consultada (Prysmian solo da formación en N°×mm, no mm²), y no se calculó por no tener el diámetro de hilo de acero con certeza suficiente separado del de aluminio en el fragmento extraído.

---

## 2. Secciones individuales por conductor

### 2.1 ACSR 150/25

**Fuentes utilizadas:**
- Prysmian Argentina, catálogo "Prysalac", tabla "Cables según norma IRAM 2187", edición diciembre 2008. URL: https://ar.prysmian.com/sites/default/files/atoms/files/4LA_4_4_Prysalac.pdf (fuente primaria de fabricante, ALTO para geometría/masa/resistencia/rating publicado).

**Datos encontrados:** formación aluminio 26×2,7 mm; formación acero 7×2,1 mm; diámetro exterior 17,1 mm; masa aprox. 600 kg/km; carga de rotura calculada 5464 kg; resistencia eléctrica máxima a 20 °C en c.c. 0,194 Ω/km; intensidad de corriente admisible 415 A, nota (1) del catálogo: "para temperatura ambiente de 40 °C, cables expuestos al sol, al nivel del mar y viento de 0,6 m/seg".

**Datos faltantes:** área efectiva de aluminio y de acero en mm² (solo se da la designación nominal "150/25"), diámetro del núcleo de acero, resistencia AC, coeficiente de temperatura de la resistencia, emisividad, absortividad, temperatura máxima de diseño del conductor (Tc), ángulo de viento, radiación solar en W/m², elevación solar, altitud/presión, frecuencia.

**Contradicciones entre fuentes:** ninguna — solo se ubicó una fuente primaria con datos completos para esta designación exacta.

**Valores publicados vs. asumidos:** Tamb=40 °C y viento=0,6 m/s son valores **publicados** por el fabricante (no asumidos por mí). Tc, radiación, ángulo de viento y demás no se asumieron: se dejaron como NO ENCONTRADO.

**¿Es reproducible el rating?** No. Falta Tc, radiación solar, ángulo de viento, emisividad y absortividad — variables imprescindibles para reproducir el cálculo IEEE 738. Estado: **rating no reproducible de forma completa con la información disponible**.

**Advertencias de calidad:** el catálogo es de 2008; no hay evidencia de que siga vigente sin cambios. La sección "150/25" corresponde a designación nominal según IRAM 2187, no verificada por área real medida.

---

### 2.2 ACSR 240/40

**Fuentes utilizadas:** igual fuente que 150/25 (Prysmian Argentina, Prysalac, IRAM 2187, dic. 2008).

**Datos encontrados:** formación Al 26×3,45 mm; formación acero 7×2,68 mm; diámetro exterior 21,9 mm; masa aprox. 980 kg/km; carga de rotura calculada 8675 kg; R20 c.c. 0,119 Ω/km; corriente admisible 565 A, mismas condiciones de nota (1) (Tamb 40 °C, sol, nivel del mar, viento 0,6 m/s).

**Datos faltantes:** idénticos a 150/25 (ver arriba).

**Contradicciones:** ninguna.

**Publicado vs. asumido:** igual criterio que 150/25.

**Reproducibilidad:** NO reproducible — mismas carencias que 150/25.

**Advertencias:** mismas que 150/25.

---

### 2.3 ACSR 795 kcmil Drake

**Fuentes utilizadas:**
1. Nexans Colombia (Centelsa by Nexans), ficha de producto "795 KCMIL (Drake)". URL: https://www.nexans.co/en/products/.../product~200021~.html (fabricante, ALTO).
2. Prysmian Argentina, "Prysalac", tabla "Cables según norma ASTM B232 cincado A". URL: https://ar.prysmian.com/sites/default/files/atoms/files/4LA_4_4_Prysalac.pdf (fabricante, ALTO).
3. Southwire, ficha de producto "795-26/7 ACSR/GA2 DRAKE". URL: https://www.southwire.com/wire-cable/bare-aluminum-overhead-transmission-distribution/acsr/p/10154314 (fabricante, ALTO para geometría/normas, sin resistencia/rating visibles en la página HTML consultada).
4. Distribuidor "Wire & Cable Your Way" (no fabricante, MEDIO), usado solo para resistencia AC a 75 °C.

**Datos encontrados:**
- Construcción: 26 hilos de aluminio 1350-H19 (diámetro 4,442 mm / 0,1749 in) + 7 hilos de acero galvanizado (diámetro 3,454 mm / 0,136 in según Nexans y datos ASTM estándar coincidentes; el documento de Prysmian muestra "7 x 4,54" que es inconsistente con el estándar Drake y probablemente un error de OCR/transcripción del PDF original — se señala como **contradicción/anomalía de fuente**, no se usa ese valor).
- Diámetro exterior: 28,13 mm (Nexans) = 1,108 in (Southwire, ficha "Nom. OD 1107 mils / 28,13 mm") — coincidentes.
- Sección de aluminio: 402,92 mm² (Nexans). Sección total del conductor: 455,11 mm² (Nexans). Sección de acero: 455,11 − 402,92 = 52,19 mm² (cálculo propio mostrado, no dato directo).
- Masa: 1629 kg/km (Nexans) vs. 1627 kg/km (Southwire, "Nominal Total Weight 1093 lbs/1000ft = 1627 kg/km") vs. 1625,6 kg/km (distribuidor Yifang, MEDIO). Diferencias menores (<0,3 %), atribuibles a redondeo.
- Resistencia DC a 20 °C: 0,071 Ω/km (Nexans, "Max. DC resistance of the conductor at 20°C") vs. 0,0702 Ω/km (Prysmian). Diferencia de ~1,1 %, ambas de fabricante — se conservan ambas.
- Resistencia AC: 0,0263 Ω/1000 ft a 75 °C = 0,0863 Ω/km (fuente de distribuidor, no de fabricante — confianza MEDIA).
- Carga de rotura: 31.500 lb (distribuidor) = 14.288,9 kg, que coincide con el valor "14288" listado en la tabla de Prysmian (aunque la alineación de columnas de ese PDF es ambigua en la extracción de texto) — confianza MEDIA-ALTA por concordancia cruzada.
- Ampacidad publicada: **907 A** en dos fuentes de fabricante distintas — pero con condiciones de referencia declaradas diferentes:
  - Nexans: Tamb=25 °C, Tc=75 °C, radiación solar=1 kW/m², coeficiente de absorción/emisividad=0,5, viento=0,61 m/s (610 mm/s), 60 Hz, nivel del mar.
  - Prysmian: Tamb=40 °C, "cables expuestos al sol", nivel del mar, viento=0,6 m/s (sin Tc, sin radiación numérica, sin emisividad/absortividad, sin frecuencia).
- Normas: ASTM B230, B232, B498, B500 (Southwire y Nexans coinciden).

**Contradicción explícita a destacar:** el mismo valor de 907 A aparece bajo dos conjuntos de condiciones ambientales *distintos* (Tamb 25 °C vs. 40 °C) en dos fabricantes distintos. Esto es matemáticamente sospechoso para un cálculo IEEE 738 real (a mayor Tamb, la ampacidad admisible para el mismo Tc debería ser menor), lo que sugiere que al menos una de las dos notas de condiciones es una leyenda genérica de catálogo no recalculada específicamente para este ítem. **No se puede usar 907 A como validado bajo ninguna de las dos condiciones sin recalcular.**

**Datos faltantes:** ángulo de viento, elevación solar, altitud/presión, coeficiente de temperatura de la resistencia (alpha), emisividad y absortividad de Prysmian (Nexans sí da 0,5/0,5), frecuencia en Prysmian.

**¿Reproducible?** Parcialmente: la ficha de Nexans es la más completa (Tc, Tamb, radiación, viento, coeficientes ópticos, frecuencia), pero le falta el ángulo de viento, la elevación solar y la altitud/presión, que IEEE 738 requiere para el término de convección y de radiación solar completos. Estado: **rating no reproducible de forma completa con la información disponible**, aunque Nexans es la fuente más cercana a permitir una reproducción aproximada.

**Advertencias de calidad:** el valor de resistencia AC (distribuidor) no debe tratarse como dato de fabricante. La discrepancia de diámetro de hilo de acero en Prysmian (7×4,54 mm) es un error de fuente, no un valor alternativo válido — descartado.

---

### 2.4 ACSR 300/50

**Fuente:** Prysmian Argentina, Prysalac, IRAM 2187, dic. 2008 (misma fuente que 150/25 y 240/40).

**Datos encontrados:** Al 26×3,86 mm; acero 7×3,0 mm; diámetro exterior 24,5 mm; masa 1230 kg/km; carga de rotura 10.700 kg; R20 c.c. 0,0949 Ω/km; corriente admisible 650 A bajo la misma nota (1) (Tamb 40 °C, sol, nivel del mar, viento 0,6 m/s).

**Datos faltantes / contradicciones / reproducibilidad:** idénticos en naturaleza a 150/25 y 240/40 — ver esas secciones. Estado: **rating no reproducible de forma completa con la información disponible**.

---

### 2.5 ACSR 435/55

**Fuente:** Prysmian Argentina, Prysalac, IRAM 2187, dic. 2008.

**Datos encontrados:** Al 54×3,2 mm; acero 7×3,2 mm; diámetro exterior 28,8 mm; masa 1640 kg/km; carga de rotura 13.688 kg; R20 c.c. 0,0666 Ω/km; corriente admisible 765 A, nota (1) idéntica a las anteriores.

**Observación:** en este conductor el diámetro de hilo de aluminio y de acero coinciden (3,2 mm ambos), lo cual es geométricamente coherente con una construcción 54 Al + 7 acero de dos capas de aluminio sobre un núcleo de 7 hilos de igual diámetro — consistente con el patrón de otros ACSR de alta relación de aluminio/acero.

**Datos faltantes / reproducibilidad:** iguales al resto de la serie IRAM. Estado: **rating no reproducible de forma completa con la información disponible**.

---

### 2.6 ACSR 477 kcmil Hawk

**Fuente:** Prysmian Argentina, Prysalac, tabla "Cables según norma ASTM B232 cincado A" (misma publicación que Drake, sección ASTM en vez de IRAM).

**Datos encontrados:** sección nominal Al 477,0 kcmil = 242 mm² (columna "mm²" del catálogo); formación Al 26×3,44 mm, acero 7×2,68 mm; diámetro exterior 21,78 mm; masa 976,1 kg/km; carga de rotura calculada 8845 kg; R20 c.c. 0,117 Ω/km; corriente admisible 659 A, bajo la misma nota (1) del catálogo (Tamb 40 °C, sol, nivel del mar, viento 0,6 m/s).

**Contraste con otras fuentes:** no se ubicó una segunda ficha de fabricante (tipo Southwire/Nexans) específica para Hawk 477 kcmil en esta sesión de investigación; solo se dispone de la fuente Prysmian. Esto **reduce la confianza relativa frente a Drake**, que sí tiene triple verificación cruzada.

**Datos faltantes:** área de acero en mm² (no publicada por separado), diámetro de núcleo, resistencia AC, alpha, emisividad, absortividad, Tc, ángulo de viento, radiación, elevación solar, presión/altitud, frecuencia.

**Reproducibilidad:** **rating no reproducible de forma completa con la información disponible**.

**Advertencia:** dado que solo hay una fuente, cualquier error tipográfico del PDF original (como el detectado en la fila de Drake del mismo catálogo) no puede contrastarse aquí. Se recomienda verificar contra una segunda ficha de fabricante (p. ej. Southwire o General Cable) antes de usar este dato para validación crítica.

---

### 2.7 AAAC/CADLA 150 mm²

**Fuentes utilizadas:**
1. APAR Industries (India), ficha técnica "All Aluminum Alloy Conductor (AAAC)", tabla "EN 50182-Type-AL3 (Used in Germany)" y "(Used in Italy)". URL: https://apar.com/wp-content/uploads/2021/02/5.-AAAC.pdf (fabricante, ALTO).
2. Industrias MH (Argentina), ficha "Conductor desnudo aleación de aluminio", con tabla de secciones IRAM incluyendo 150 mm². URL: https://www.industriasmh.com.ar/pdf/conductor-desnudo-aleacion-aluminio.pdf — **la información identificada para este caso solo aporta geometría**, por lo que la fuente se trata como confianza MEDIA y no se utiliza para asignar resistencia ni ampacidad.

**Aclaración de trazabilidad importante:** el requerimiento original pide "AAAC/CADLA 150 mm²". CADLA es la denominación comercial usual en Argentina para conductor de aleación de aluminio bajo IRAM 2212 (material) e IRAM 2177 (construcción), habitualmente aleación 6201 o equivalente. **La documentación técnica identificada para CADLA no contiene el conjunto completo de datos eléctricos y térmicos requerido para reproducir el caso.** Por eso, los valores de resistencia y ampacidad que se listan a continuación **provienen de la ficha de APAR (India) bajo la norma EN 50182 tipo AL3, "150/37"**, que **no es la misma norma que IRAM 2212/2177**, aunque la geometría (37 hilos × 2,25 mm, diámetro exterior 15,8 mm) coincide de forma notable con los datos geométricos identificados para la sección nominal argentina de 150 mm². Esta coincidencia no demuestra identidad de producto: **no se debe asumir que la resistencia AC, la ampacidad o las condiciones de ensayo de APAR sean válidas bajo el nombre comercial "CADLA" sin confirmación documental específica.**

**Datos encontrados (APAR, EN 50182 tipo AL3):**
- Sección real: 147,1 mm² (nominal "150"); 37 hilos de 2,25 mm; diámetro exterior 15,8 mm; masa 405,3 kg/km; carga de rotura 43,40 kN (≈4425 kg); resistencia DC a 20 °C: 0,2256 Ω/km.
- Corriente admisible: **284 A a Tc=75 °C** y **348 A a Tc=85 °C**.
- Condiciones declaradas explícitamente por el fabricante para AMBOS valores de corriente: viento 0,56 m/s, elevación 0 m, emisividad 0,45, absortividad 0,80, Tamb 45 °C, radiación solar 1045 W/m².
- Material: aleación de aluminio 6201-T81 (confirmado también por Chalco Aluminum y por el propio código de conductor).
- Coeficiente de temperatura de la resistencia: no publicado en esta ficha; una fuente secundaria (Panghao Conductor, fabricante chino, confianza MEDIA) da 0,0036 /°C para AAAC en general — se marca como **valor normativo/genérico**, no medido para este conductor específico.

**Datos faltantes:** ángulo de viento, elevación solar, altitud/presión, frecuencia, resistencia AC explícita (solo se da la ampacidad ya calculada, no la Ω/km a la temperatura de operación), temperatura máxima nominal de operación continua (una fuente distinta, industriasmh, menciona 80 °C de operación nominal y 130 °C de cortocircuito para IRAM 2212/2177, pero **no se confirmó que corresponda a esta ficha de APAR** — se deja como referencia cruzada, no como dato propio de la ficha usada para el rating).

**Contradicciones:** ninguna dentro de la misma fuente; la única "contradicción" es de alcance normativo (EN 50182 vs. IRAM 2212), señalada arriba.

**¿Reproducible el rating?** Parcialmente — esta es la ficha más completa de todo el conjunto de 8 conductores (Tc, Tamb, viento, radiación, emisividad y absortividad, todos publicados), pero **le siguen faltando el ángulo de viento, la elevación solar y la altitud/presión**, que IEEE 738 exige para el balance completo. Por eso continúa clasificado como **no reproducible al 100 %**, aunque con confianza ALTA en la mayoría de las variables.

**Advertencias de calidad:** no usar este dato bajo la etiqueta comercial "CADLA" en documentos regulatorios argentinos sin verificación directa con el fabricante local, dado que la ficha de origen es de un fabricante indio bajo norma europea.

---

### 2.8 AAAC/CADLA 240 mm²

**Fuentes:** igual que 150 mm² (APAR, EN 50182 tipo AL3; Industrias MH para geometría cruzada de la sección nominal argentina de 240 mm²).

**Datos encontrados — dos variantes nacionales dentro de la misma ficha APAR, EN 50182 tipo AL3, ambas etiquetadas como sección nominal "240":**

| Variante | Sección real (mm²) | Hilos | Ø hilo (mm) | Ø exterior (mm) | Masa (kg/km) | Carga rotura (kN) | R20 (Ω/km) | I @75°C (A) | I @85°C (A) |
|---|---|---|---|---|---|---|---|---|---|
| "Usada en Alemania/Austria" (243-AL3) | 242,5 | 61 | 2,25 | 20,3 | 670,3 | 71,55 | 0,1373 | 378 | 470 |
| "Usada en Italia" (244-AL3, "240/37") | 244,4 | 37 | 2,90 | 20,30 | 673,30 | 72,10 | 0,1358 | 380 | 473 |

**Contradicción explícita:** para la **misma sección nominal (240 mm²) y la misma norma EN 50182 tipo AL3**, el propio fabricante APAR publica **dos construcciones geométricas distintas** (61 hilos de 2,25 mm vs. 37 hilos de 2,90 mm) con resultados eléctricos ligeramente distintos (R20: 0,1373 vs. 0,1358 Ω/km; ampacidad: 378 vs. 380 A a 75 °C). **No deben promediarse ni combinarse: son productos distintos que comparten sección nominal y norma pero difieren en cableado.** Se conservan ambas explícitamente, tal como exige la consigna.

**Cruce con Industrias MH (Argentina, IRAM 2212, confianza MEDIA, solo geometría):** sección nominal 240 mm², 37 hilos, Ø 2,85 mm, diámetro exterior 20,0 mm, masa 650,8 kg/km. Esta geometría (37 hilos) es más cercana a la variante "Italia" de APAR (37 hilos) que a la variante "Alemania/Austria" (61 hilos), aunque el diámetro de hilo difiere (2,85 mm IRAM vs. 2,90 mm EN 50182-Italia) y la masa difiere (650,8 vs. 673,3 kg/km) — **no son el mismo producto**, solo son geométricamente parecidos. No se debe usar la resistencia ni la ampacidad de APAR como si fueran valores certificados de CADLA IRAM 2212.

**Condiciones de ampacidad (ambas variantes APAR):** viento 0,56 m/s, elevación 0 m, emisividad 0,45, absortividad 0,80, Tamb 45 °C, radiación solar 1045 W/m², Tc = 75 °C y 85 °C (dos columnas).

**Datos faltantes:** los mismos que en 150 mm² (ángulo de viento, elevación solar, presión/altitud, frecuencia, alpha específico del producto).

**Reproducibilidad:** igual que 150 mm² — parcialmente reproducible, con las mismas tres variables faltantes.

**Advertencias de calidad:** la existencia de dos construcciones distintas bajo el mismo nombre nominal "240" en la misma ficha de fabricante es un recordatorio de que la designación nominal en mm² **no identifica unívocamente la geometría** — cwind debe validarse contra la construcción específica (número de hilos y diámetro), no solo contra la sección nominal.

---

## 3. Bloques YAML

```yaml
id: acsr_150_25
name: "ACSR 150/25"
manufacturer: "Prysmian Argentina (Prysalac)"
type: ACSR
standard: "IRAM 2187"
standard_edition: "Diciembre 2008"
source: "Catálogo técnico de fabricante (PDF)"
source_url: "https://ar.prysmian.com/sites/default/files/atoms/files/4LA_4_4_Prysalac.pdf"
source_date: "2026-09-21"
diameter_m: 0.0171
core_diameter_m: null
al_area_mm2: null
steel_area_mm2: null
total_area_mm2: null
al_layers: null
al_wires: 26
steel_wires: 7
mass_kg_per_km: 600
r20_ohm_per_m: 0.000194
rac_ohm_per_m: null
rac_temperature_c: null
frequency_hz: null
alpha_per_c: null
max_temperature_c: null
emissivity: null
absorptivity: null
declared_rating_a: 415
rating_conditions:
  tc_c: null
  tamb_c: 40
  wind_speed_mps: 0.6
  wind_angle_deg: null
  solar_radiation_wm2: null
  solar_elevation_deg: null
  pressure_pa: null
  altitude_m: null
  natural_convection: null
  source: "Nota (1) del catálogo Prysalac: 'Para temperatura ambiente de 40º C, cables expuestos al sol, al nivel del mar y viento de 0,6 m/seg'"
validation_status: "rating no reproducible de forma completa con la información disponible"
confidence: MEDIO
missing_data:
  - "área efectiva de aluminio y acero en mm2"
  - "resistencia AC"
  - "coeficiente de temperatura de la resistencia"
  - "Tc"
  - "ángulo de viento"
  - "radiación solar"
  - "elevación solar"
  - "altitud/presión"
  - "emisividad"
  - "absortividad"
  - "frecuencia"
notes: "Diámetro y formación coinciden con la designación nominal IRAM 2187; sección nominal, no medida."
```

```yaml
id: acsr_240_40
name: "ACSR 240/40"
manufacturer: "Prysmian Argentina (Prysalac)"
type: ACSR
standard: "IRAM 2187"
standard_edition: "Diciembre 2008"
source: "Catálogo técnico de fabricante (PDF)"
source_url: "https://ar.prysmian.com/sites/default/files/atoms/files/4LA_4_4_Prysalac.pdf"
source_date: "2026-09-21"
diameter_m: 0.0219
core_diameter_m: null
al_area_mm2: null
steel_area_mm2: null
total_area_mm2: null
al_layers: null
al_wires: 26
steel_wires: 7
mass_kg_per_km: 980
r20_ohm_per_m: 0.000119
rac_ohm_per_m: null
rac_temperature_c: null
frequency_hz: null
alpha_per_c: null
max_temperature_c: null
emissivity: null
absorptivity: null
declared_rating_a: 565
rating_conditions:
  tc_c: null
  tamb_c: 40
  wind_speed_mps: 0.6
  wind_angle_deg: null
  solar_radiation_wm2: null
  solar_elevation_deg: null
  pressure_pa: null
  altitude_m: null
  natural_convection: null
  source: "Nota (1) del catálogo Prysalac (idéntica a 150/25)"
validation_status: "rating no reproducible de forma completa con la información disponible"
confidence: MEDIO
missing_data:
  - "área efectiva de aluminio y acero en mm2"
  - "resistencia AC"
  - "coeficiente de temperatura de la resistencia"
  - "Tc"
  - "ángulo de viento"
  - "radiación solar"
  - "elevación solar"
  - "altitud/presión"
  - "emisividad"
  - "absortividad"
  - "frecuencia"
notes: "Igual observación que 150/25."
```

```yaml
id: acsr_795_drake
name: "ACSR 795 kcmil (Drake)"
manufacturer: "Nexans (Centelsa) / Prysmian Argentina / Southwire"
type: ACSR
standard: "ASTM B232 / B230 / B498 / B500"
standard_edition: "no especificada en las fichas consultadas"
source: "Fichas de producto de tres fabricantes"
source_url: "https://www.nexans.co/en/products/.../product~200021~.html ; https://ar.prysmian.com/sites/default/files/atoms/files/4LA_4_4_Prysalac.pdf ; https://www.southwire.com/wire-cable/bare-aluminum-overhead-transmission-distribution/acsr/p/10154314"
source_date: "2026-09-21"
diameter_m: 0.02813
core_diameter_m: null
al_area_mm2: 402.92
steel_area_mm2: 52.19
total_area_mm2: 455.11
al_layers: null
al_wires: 26
steel_wires: 7
mass_kg_per_km: 1629
r20_ohm_per_m: 0.000071
rac_ohm_per_m: 0.0000863
rac_temperature_c: 75
frequency_hz: 60
alpha_per_c: null
max_temperature_c: null
emissivity: 0.5
absorptivity: 0.5
declared_rating_a: 907
rating_conditions:
  tc_c: 75
  tamb_c: 25
  wind_speed_mps: 0.61
  wind_angle_deg: null
  solar_radiation_wm2: 1000
  solar_elevation_deg: null
  pressure_pa: null
  altitude_m: 0
  natural_convection: null
  source: "Nexans: 'Current capacity at ambient temperature 25°C, conductor temperature 75°C, solar emission 1kW/m², absorption and emissivity coefficients 0.5, wind speed 610 mm/sec, at sea level and at 60 Hz'"
validation_status: "rating no reproducible de forma completa con la información disponible (falta ángulo de viento, elevación solar y presión/altitud)"
confidence: ALTO (geometría) / MEDIO (rating, por contradicción de condiciones entre fuentes)
missing_data:
  - "ángulo de viento"
  - "elevación solar"
  - "altitud/presión"
  - "alpha (coeficiente de temperatura de la resistencia)"
  - "área de acero medida directamente (se calculó por diferencia)"
notes: "Prysmian publica el mismo 907 A bajo Tamb=40°C, sin Tc ni radiación explícitos — condiciones inconsistentes con las de Nexans para el mismo valor numérico. Se preservan ambas fuentes sin fusionarlas. El diámetro de hilo de acero de 4,54 mm en el PDF de Prysmian es inconsistente con el estándar Drake (3,454 mm según Nexans y ASTM) y se descarta como probable error de OCR."
```

```yaml
id: acsr_300_50
name: "ACSR 300/50"
manufacturer: "Prysmian Argentina (Prysalac)"
type: ACSR
standard: "IRAM 2187"
standard_edition: "Diciembre 2008"
source: "Catálogo técnico de fabricante (PDF)"
source_url: "https://ar.prysmian.com/sites/default/files/atoms/files/4LA_4_4_Prysalac.pdf"
source_date: "2026-09-21"
diameter_m: 0.0245
core_diameter_m: null
al_area_mm2: null
steel_area_mm2: null
total_area_mm2: null
al_layers: null
al_wires: 26
steel_wires: 7
mass_kg_per_km: 1230
r20_ohm_per_m: 0.0000949
rac_ohm_per_m: null
rac_temperature_c: null
frequency_hz: null
alpha_per_c: null
max_temperature_c: null
emissivity: null
absorptivity: null
declared_rating_a: 650
rating_conditions:
  tc_c: null
  tamb_c: 40
  wind_speed_mps: 0.6
  wind_angle_deg: null
  solar_radiation_wm2: null
  solar_elevation_deg: null
  pressure_pa: null
  altitude_m: null
  natural_convection: null
  source: "Nota (1) del catálogo Prysalac"
validation_status: "rating no reproducible de forma completa con la información disponible"
confidence: MEDIO
missing_data:
  - "área efectiva de aluminio y acero en mm2"
  - "resistencia AC"
  - "alpha"
  - "Tc"
  - "ángulo de viento"
  - "radiación solar"
  - "elevación solar"
  - "altitud/presión"
  - "emisividad"
  - "absortividad"
  - "frecuencia"
notes: "Misma familia de catálogo que 150/25, 240/40 y 435/55."
```

```yaml
id: acsr_435_55
name: "ACSR 435/55"
manufacturer: "Prysmian Argentina (Prysalac)"
type: ACSR
standard: "IRAM 2187"
standard_edition: "Diciembre 2008"
source: "Catálogo técnico de fabricante (PDF)"
source_url: "https://ar.prysmian.com/sites/default/files/atoms/files/4LA_4_4_Prysalac.pdf"
source_date: "2026-09-21"
diameter_m: 0.0288
core_diameter_m: null
al_area_mm2: null
steel_area_mm2: null
total_area_mm2: null
al_layers: null
al_wires: 54
steel_wires: 7
mass_kg_per_km: 1640
r20_ohm_per_m: 0.0000666
rac_ohm_per_m: null
rac_temperature_c: null
frequency_hz: null
alpha_per_c: null
max_temperature_c: null
emissivity: null
absorptivity: null
declared_rating_a: 765
rating_conditions:
  tc_c: null
  tamb_c: 40
  wind_speed_mps: 0.6
  wind_angle_deg: null
  solar_radiation_wm2: null
  solar_elevation_deg: null
  pressure_pa: null
  altitude_m: null
  natural_convection: null
  source: "Nota (1) del catálogo Prysalac"
validation_status: "rating no reproducible de forma completa con la información disponible"
confidence: MEDIO
missing_data:
  - "área efectiva de aluminio y acero en mm2"
  - "resistencia AC"
  - "alpha"
  - "Tc"
  - "ángulo de viento"
  - "radiación solar"
  - "elevación solar"
  - "altitud/presión"
  - "emisividad"
  - "absortividad"
  - "frecuencia"
notes: "Único conductor de la serie IRAM consultada con 54 hilos de aluminio (dos capas) en vez de 26."
```

```yaml
id: acsr_477_hawk
name: "ACSR 477 kcmil (Hawk)"
manufacturer: "Prysmian Argentina (Prysalac)"
type: ACSR
standard: "ASTM B232"
standard_edition: "no especificada en la ficha consultada"
source: "Catálogo técnico de fabricante (PDF), tabla 'Cables según norma ASTM B232 cincado A'"
source_url: "https://ar.prysmian.com/sites/default/files/atoms/files/4LA_4_4_Prysalac.pdf"
source_date: "2026-09-21"
diameter_m: 0.02178
core_diameter_m: null
al_area_mm2: 242
steel_area_mm2: null
total_area_mm2: null
al_layers: null
al_wires: 26
steel_wires: 7
mass_kg_per_km: 976.1
r20_ohm_per_m: 0.000117
rac_ohm_per_m: null
rac_temperature_c: null
frequency_hz: null
alpha_per_c: null
max_temperature_c: null
emissivity: null
absorptivity: null
declared_rating_a: 659
rating_conditions:
  tc_c: null
  tamb_c: 40
  wind_speed_mps: 0.6
  wind_angle_deg: null
  solar_radiation_wm2: null
  solar_elevation_deg: null
  pressure_pa: null
  altitude_m: null
  natural_convection: null
  source: "Nota (1) del catálogo Prysalac"
validation_status: "rating no reproducible de forma completa con la información disponible"
confidence: MEDIO (fuente única, sin verificación cruzada con otro fabricante)
missing_data:
  - "área de acero en mm2"
  - "diámetro de núcleo"
  - "resistencia AC"
  - "alpha"
  - "Tc"
  - "ángulo de viento"
  - "radiación solar"
  - "elevación solar"
  - "altitud/presión"
  - "emisividad"
  - "absortividad"
  - "frecuencia"
notes: "No se encontró una segunda ficha de fabricante para Hawk 477 kcmil en esta sesión; se recomienda verificación cruzada adicional (p. ej. Southwire, General Cable) antes de uso crítico."
```

```yaml
id: aaac_cadla_150
name: "AAAC 150 mm2 (EN 50182 tipo AL3, código '150/37'); posible equivalencia con CADLA IRAM 2212/2177 (geometría) — no confirmada para datos eléctricos"
manufacturer: "APAR Industries (India) — datos eléctricos/térmicos; Industrias MH (Argentina) — geometría cruzada IRAM, sin datos eléctricos"
type: AAAC
standard: "EN 50182 tipo AL3 (datos eléctricos); IRAM 2212 / IRAM 2177 (geometría cruzada, sin datos eléctricos confirmados)"
standard_edition: "no especificada en las fichas consultadas"
source: "Ficha técnica de fabricante (PDF) + fragmento indexado de ficha de fabricante argentino"
source_url: "https://apar.com/wp-content/uploads/2021/02/5.-AAAC.pdf ; https://www.industriasmh.com.ar/pdf/conductor-desnudo-aleacion-aluminio.pdf (solo fragmento, PDF bloqueado para acceso directo)"
source_date: "2026-09-21"
diameter_m: 0.0158
core_diameter_m: null
al_area_mm2: 147.1
steel_area_mm2: null
total_area_mm2: 147.1
al_layers: null
al_wires: 37
steel_wires: 0
mass_kg_per_km: 405.3
r20_ohm_per_m: 0.0002256
rac_ohm_per_m: null
rac_temperature_c: null
frequency_hz: null
alpha_per_c: 0.0036
max_temperature_c: null
emissivity: 0.45
absorptivity: 0.80
declared_rating_a: 284
rating_conditions:
  tc_c: 75
  tamb_c: 45
  wind_speed_mps: 0.56
  wind_angle_deg: null
  solar_radiation_wm2: 1045
  solar_elevation_deg: null
  pressure_pa: null
  altitude_m: 0
  natural_convection: null
  source: "APAR datasheet nota: 'Current capacity based on referenced conductor temperature, 0.56 m/s wind, 0 m Elevation, 0.45 Emissivity, 0.80 absorptivity, 45°C Ambient temperature, 1045 W/m² Solar radiation'"
validation_status: "rating no reproducible de forma completa con la información disponible (falta ángulo de viento, elevación solar y presión/altitud); es la ficha más completa del conjunto de 8 conductores"
confidence: ALTO (geometría y rating de APAR) / MEDIO (equivalencia con CADLA IRAM, solo por semejanza geométrica)
missing_data:
  - "ángulo de viento"
  - "elevación solar"
  - "altitud/presión (solo se da 'elevación 0 m', que se interpreta como altitud, no como presión atmosférica)"
  - "alpha específico del producto (el 0,0036 es un valor genérico de fuente secundaria, no de esta ficha)"
  - "confirmación de que este dato aplica bajo la marca comercial 'CADLA' en Argentina"
notes: "Existe también un segundo valor de ampacidad a Tc=85°C: 348 A, bajo las mismas demás condiciones. No usar esta ficha como equivalente certificado a CADLA IRAM sin validación del fabricante argentino."
```

```yaml
id: aaac_cadla_240
name: "AAAC 240 mm2 (EN 50182 tipo AL3) — DOS variantes de fabricante bajo la misma sección nominal; posible equivalencia geométrica con CADLA IRAM 2212/2177 no confirmada para datos eléctricos"
manufacturer: "APAR Industries (India) — datos eléctricos/térmicos; Industrias MH (Argentina) — geometría cruzada IRAM, sin datos eléctricos"
type: AAAC
standard: "EN 50182 tipo AL3 (datos eléctricos); IRAM 2212 / IRAM 2177 (geometría cruzada, sin datos eléctricos confirmados)"
standard_edition: "no especificada en las fichas consultadas"
source: "Ficha técnica de fabricante (PDF) + fragmento indexado de ficha de fabricante argentino"
source_url: "https://apar.com/wp-content/uploads/2021/02/5.-AAAC.pdf ; https://www.industriasmh.com.ar/pdf/conductor-desnudo-aleacion-aluminio.pdf (solo fragmento, PDF bloqueado para acceso directo)"
source_date: "2026-09-21"
variant_de_at:
  diameter_m: 0.0203
  al_area_mm2: 242.5
  al_wires: 61
  wire_diameter_mm: 2.25
  mass_kg_per_km: 670.3
  r20_ohm_per_m: 0.0001373
  declared_rating_a_at_75c: 378
  declared_rating_a_at_85c: 470
variant_italia:
  diameter_m: 0.0203
  al_area_mm2: 244.4
  al_wires: 37
  wire_diameter_mm: 2.90
  mass_kg_per_km: 673.3
  r20_ohm_per_m: 0.0001358
  declared_rating_a_at_75c: 380
  declared_rating_a_at_85c: 473
steel_area_mm2: null
total_area_mm2: null
alpha_per_c: 0.0036
emissivity: 0.45
absorptivity: 0.80
rating_conditions:
  tc_c: [75, 85]
  tamb_c: 45
  wind_speed_mps: 0.56
  wind_angle_deg: null
  solar_radiation_wm2: 1045
  solar_elevation_deg: null
  pressure_pa: null
  altitude_m: 0
  natural_convection: null
  source: "APAR datasheet, nota idéntica a la de AAAC 150 mm²"
validation_status: "rating no reproducible de forma completa con la información disponible; existen DOS construcciones geométricas distintas bajo el mismo nombre nominal '240' en la misma ficha de fabricante — no combinar"
confidence: ALTO (ambas variantes, de fabricante) / MEDIO (equivalencia con CADLA IRAM, solo por semejanza geométrica parcial — la variante argentina de Industrias MH usa 37 hilos de 2,85 mm, más cercana a la variante 'Italia' que a la 'Alemania/Austria', pero con masa distinta: 650,8 kg/km IRAM vs. 673,3 kg/km EN 50182-Italia)
missing_data:
  - "ángulo de viento"
  - "elevación solar"
  - "altitud/presión"
  - "alpha específico del producto"
  - "confirmación de equivalencia con la marca comercial 'CADLA' argentina"
notes: "Contradicción documentada entre las dos variantes nacionales de la misma ficha APAR: NO se promedian ni combinan, se listan ambas por separado según exige la consigna."
```

---

## 4. Matriz final de prioridad para cwind

| conductor | disponibilidad de datos | calidad de fuentes | reproducibilidad del rating | parámetros críticos faltantes | prioridad de validación |
|---|---|---|---|---|---|
| ACSR 795 kcmil Drake | Alta (geometría y resistencia con 3 fuentes de fabricante cruzadas) | Alta, pero con una contradicción de condiciones de rating entre fabricantes | No reproducible (falta ángulo de viento, elevación solar, presión) | Ángulo de viento, elevación solar, presión/altitud, alpha, resolución de la contradicción Tamb 25 vs. 40 °C | **ALTA** — es el conductor con más datos disponibles y con una contradicción concreta que vale la pena resolver primero |
| AAAC 150 mm² (EN 50182 AL3) | Alta (ficha de fabricante más completa del conjunto: Tc, Tamb, viento, radiación, emisividad, absortividad) | Alta para APAR; media para la equivalencia con CADLA IRAM | No reproducible al 100 % (falta ángulo de viento, elevación solar, presión), pero es la más cercana | Ángulo de viento, elevación solar, presión/altitud, confirmación de equivalencia con CADLA argentina | **ALTA** — mejor candidato para una primera validación numérica de cwind por ser la ficha más completa |
| AAAC 240 mm² (EN 50182 AL3) | Alta, pero con ambigüedad geométrica (dos construcciones bajo el mismo nombre nominal) | Alta para cada variante individual; la ambigüedad reduce la utilidad si no se fija la construcción exacta | No reproducible al 100 %; además requiere fijar cuál de las dos variantes se usa | Ángulo de viento, elevación solar, presión/altitud, decisión explícita de qué variante (61×2,25 mm vs. 37×2,90 mm) representa el conductor real de interés | **MEDIA-ALTA** — buena base de datos, pero exige primero resolver cuál variante corresponde al conductor real antes de validar |
| ACSR 240/40 | Media (una sola fuente de fabricante, pero completa en geometría/resistencia/rating parcial) | Media — fuente única, catálogo de 2008 | No reproducible (faltan Tc, radiación, ángulo de viento, emisividad, absortividad, presión) | Tc, radiación solar, ángulo de viento, emisividad, absortividad, altitud/presión, alpha | **MEDIA** |
| ACSR 300/50 | Media (misma fuente única que 240/40) | Media | No reproducible | Idénticos a 240/40 | **MEDIA** |
| ACSR 150/25 | Media (misma fuente única) | Media | No reproducible | Idénticos a 240/40 | **MEDIA** |
| ACSR 435/55 | Media (misma fuente única) | Media | No reproducible | Idénticos a 240/40 | **MEDIA** |
| ACSR 477 kcmil Hawk | Media-Baja (única fuente, sin verificación cruzada, a diferencia de Drake) | Media — no se pudo contrastar con un segundo fabricante | No reproducible | Todos los de la serie IRAM/ASTM de Prysmian, más la verificación cruzada faltante | **BAJA** — se recomienda conseguir una segunda ficha de fabricante antes de usarlo para validar cwind |

---

## 5. Resumen de limitaciones de esta entrega

1. **Ningún conductor de los 8 tiene el conjunto completo de variables que exige IEEE 738** (Tc, Tamb, viento, ángulo de viento, radiación solar, elevación solar, altitud/presión) publicado por un único fabricante. El más cercano es **AAAC 150 mm² / 240 mm² (APAR, EN 50182 AL3)**, al que solo le faltan ángulo de viento, elevación solar y presión/altitud.
2. **La identificación exacta del producto comercial "CADLA" bajo IRAM 2212/2177 con datos eléctricos completos queda pendiente de documentación técnica específica**. Hasta contar con esa información, las geometrías y los datos de APAR deben mantenerse como referencias separadas y no intercambiables.
3. Se documentaron **dos contradicciones concretas** dignas de atención antes de calibrar cwind:
   - Drake 795 kcmil: mismo rating (907 A) bajo dos conjuntos de condiciones de referencia distintos (Nexans vs. Prysmian).
   - AAAC 240 mm²: dos construcciones geométricas distintas (61×2,25 mm vs. 37×2,90 mm) bajo la misma sección nominal y norma, en la misma ficha de fabricante.
4. No se inventó, infirió ni completó ningún valor NO ENCONTRADO; todos los campos vacíos se dejaron explícitamente así en las tablas y en los YAML.
