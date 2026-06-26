import { useNavigate } from 'react-router-dom'

function Home() {
  const navigate = useNavigate()

  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100vh' }}>
      <h1>MIA - Sistema de Archivos</h1>
      <p>¿Qué deseas hacer?</p>
      <div style={{ display: 'flex', gap: '2rem', marginTop: '2rem' }}>
        <button onClick={() => navigate('/script')}>
          📄 Cargar Script
        </button>
        <button onClick={() => navigate('/discos')}>
          🖥️ Ver Sistema de Archivos
        </button>
      </div>
    </div>
  )
}

export default Home