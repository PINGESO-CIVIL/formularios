// App de desarrollo local (DAS §3): monta los microcomponentes con datos simulados.
// No se publica; los productos componen los microcomponentes en su propia app.
export function AppDesarrollo() {
  return (
    <main className="bg-fondo p-formularios font-formularios text-texto min-h-screen">
      <h1 className="text-2xl font-semibold">Formularios</h1>
      <p className="text-texto-suave mt-2">
        App de desarrollo de los microcomponentes. Los componentes se agregan en el Hito
        1.
      </p>
    </main>
  )
}
