import { renderToStaticMarkup } from "react-dom/server"
import { describe, expect, it } from "vitest"

import { AppDesarrollo } from "./app-desarrollo"

describe("AppDesarrollo", () => {
  it("muestra el título y usa solo clases del tema", () => {
    const html = renderToStaticMarkup(<AppDesarrollo />)
    expect(html).toContain("Formularios")
    expect(html).toContain("bg-fondo")
    expect(html).toContain("text-texto")
  })
})
