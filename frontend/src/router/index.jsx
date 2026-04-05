import { Routes, Route, Navigate } from 'react-router-dom'
import KitchenSinkPage from '../pages/KitchenSink/KitchenSinkPage.jsx'
import CatalogPage from '../pages/Catalog/CatalogPage.jsx'
import GamesPage from '../pages/Games/GamesPage.jsx'
import ShmonPage from '../pages/Shmon/ShmonPage.jsx'
import EduPage from '../pages/Edu/EduPage.jsx'

export default function AppRouter() {
  return (
    <Routes>
      <Route path="/"              element={<KitchenSinkPage />} />
      <Route path="/catalog"       element={<CatalogPage />} />
      <Route path="/games"         element={<GamesPage />} />
      <Route path="/games/shmon"   element={<ShmonPage />} />
      {/* Остальные игры — заглушка, будут реализованы позже */}
      <Route path="/games/:id"     element={<GamesPage />} />
      <Route path="/edu"           element={<EduPage />} />
      <Route path="*"              element={<Navigate to="/" replace />} />
    </Routes>
  )
}
