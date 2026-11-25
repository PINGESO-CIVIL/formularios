#!/bin/sh
# Revisa el formato (gofmt) y ejecuta go vet sobre los módulos Go del repo:
# la API y el spa-server de la web. Lo usan lefthook (pre-commit) y pnpm verify.
estado=0
for modulo in apps/api apps/web/servidor; do
  sin_formato=$(cd "$modulo" && gofmt -l .)
  if [ -n "$sin_formato" ]; then
    echo "Archivos Go sin formato en $modulo. Ejecuta: cd $modulo; gofmt -w ."
    echo "$sin_formato"
    estado=1
  fi
  (cd "$modulo" && go vet ./...) || estado=1
done
exit $estado
