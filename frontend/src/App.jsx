import { BrowserRouter, Routes, Route } from 'react-router-dom'
import Home from './pages/Home'
import Discos from './pages/Discos'
import CrearDisco from './pages/CrearDisco'
import Script from './pages/Script'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/discos" element={<Discos />} />
        <Route path="/discos/crear" element={<CrearDisco />} />
        <Route path="/script" element={<Script />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App