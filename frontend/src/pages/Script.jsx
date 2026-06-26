import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

function Script() {
  const navigate = useNavigate()
  const [contenido, setContenido] = useState('')
  const [mensaje, setMensaje] = useState('')

  return (
    <div style={{ padding: '2rem' }}>
      <button onClick={() => navigate('/')}>← Volver</button>
      <h1>Cargar Script</h1>
      <textarea
        rows={15}
        cols={60}
        value={contenido}
        onChange={(e) => setContenido(e.target.value)}
        placeholder="Pega aquí el contenido de tu script .smia"
      />
      <br />
      <button onClick={() => setMensaje('Endpoint de script pendiente')}>
        Ejecutar Script
      </button>
      {mensaje && <p>{mensaje}</p>}
    </div>
  )
}

export default Script