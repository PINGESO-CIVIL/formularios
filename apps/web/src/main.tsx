import { createRoot } from "react-dom/client"
import { Formularios } from "./features/formularios/Formularios"
import "./styles.css"
async function iniciar() {
  if (import.meta.env.DEV && import.meta.env.VITE_MOCKS === "true") {
    const { worker } = await import("./mocks/browser")
    await worker.start({ onUnhandledRequest: "bypass" })
  }
  createRoot(document.getElementById("root")!).render(<Formularios />)
}
void iniciar()
