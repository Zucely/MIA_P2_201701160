import { useNavigate } from 'react-router-dom'

function Discos() {
  const navigate = useNavigate()

  return (
    <div style={{ padding: '2rem' }}>
        <button onClick={() => navigate('/')}>← Volver</button>
      <h1>Discos</h1>
      <button onClick={() => navigate('/discos/crear')}>+ Crear nuevo disco</button>
      <div style={{ marginTop: '2rem' }}>
        <p>No hay discos aún.</p>
      </div>
    </div>
  )
}

export default Discos