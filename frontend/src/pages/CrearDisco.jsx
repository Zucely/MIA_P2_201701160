import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

function CrearDisco() {
  const navigate = useNavigate()
  const [form, setForm] = useState({ size: '', unit: 'M', fit: 'FF', path: '' })
  const [mensaje, setMensaje] = useState('')

  const handleChange = (e) => {
    setForm({ ...form, [e.target.name]: e.target.value })
  }

  const handleSubmit = async () => {
    try {
      const res = await fetch('http://localhost:8080/api/mkdisk', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ...form, size: parseInt(form.size) })
      })
      const data = await res.json()
      setMensaje(data.message)
    } catch (err) {
      setMensaje('Error al conectar con el servidor')
    }
  }

  return (
    <div style={{ padding: '2rem' }}>
      <button onClick={() => navigate('/discos')}>← Volver</button>
      <h1>Crear Disco</h1>
      <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', maxWidth: '400px' }}>
        <label>
          Size:
          <input name="size" type="number" value={form.size} onChange={handleChange} />
        </label>
        <label>
          Unit:
          <select name="unit" value={form.unit} onChange={handleChange}>
            <option value="M">MB</option>
            <option value="K">KB</option>
          </select>
        </label>
        <label>
          Fit:
          <select name="fit" value={form.fit} onChange={handleChange}>
            <option value="FF">First Fit</option>
            <option value="BF">Best Fit</option>
            <option value="WF">Worst Fit</option>
          </select>
        </label>
        <label>
          Path:
          <input name="path" type="text" value={form.path} onChange={handleChange} placeholder="/home/discos/disco1.mia" />
        </label>
        <button onClick={handleSubmit}>Crear Disco</button>
        {mensaje && <p>{mensaje}</p>}
      </div>
    </div>
  )
}

export default CrearDisco