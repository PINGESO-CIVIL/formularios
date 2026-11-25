import { StrictMode } from "react"
import { createRoot } from "react-dom/client"

import { AppDesarrollo } from "./app/app-desarrollo"
import "./estilos/tema.css"

const raiz = document.getElementById("raiz")
if (!raiz) throw new Error("No existe el elemento #raiz")

createRoot(raiz).render(
  <StrictMode>
    <AppDesarrollo />
  </StrictMode>,
)
