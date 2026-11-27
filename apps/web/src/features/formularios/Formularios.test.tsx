import { expect, test } from "vitest"
import { renderToStaticMarkup } from "react-dom/server"
import { Formularios } from "./Formularios"
test("muestra el título de la aplicación", () => {
  expect(renderToStaticMarkup(<Formularios />)).toContain("<h1>Formularios</h1>")
})
