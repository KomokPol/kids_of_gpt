import { Routes, Route, Navigate } from 'react-router-dom'
import HomePage from '../pages/Home/HomePage.jsx'
import CatalogPage from '../pages/Catalog/CatalogPage.jsx'
import GamesPage from '../pages/Games/GamesPage.jsx'
import ShmonPage from '../pages/Shmon/ShmonPage.jsx'
import EduPage from '../pages/Edu/EduPage.jsx'
import MarketPage from '../pages/Market/MarketPage.jsx'
import RanksPage from '../pages/Ranks/RanksPage.jsx'
import AccessGuard from '../ui/AccessGuard/AccessGuard.jsx'

export default function AppRouter() {
  return (
    <Routes>
      {/* Главная — доступна всем */}
      <Route path="/"              element={<HomePage />} />

      {/* Шарага — доступна всем */}
      <Route path="/edu"           element={<EduPage />} />

      {/* Звания — доступны всем */}
      <Route path="/ranks"         element={<RanksPage />} />

      {/* Барахолка — с ранга Мужик */}
      <Route path="/market"        element={
        <AccessGuard sectionId="market"><MarketPage /></AccessGuard>
      } />

      {/* Игры — с ранга Бродяга */}
      <Route path="/games"         element={
        <AccessGuard sectionId="games"><GamesPage /></AccessGuard>
      } />
      <Route path="/games/shmon"   element={
        <AccessGuard sectionId="games/shmon"><ShmonPage /></AccessGuard>
      } />
      <Route path="/games/:id"     element={
        <AccessGuard sectionId="games"><GamesPage /></AccessGuard>
      } />

      {/* Видеомагнитофон — с ранга Авторитет */}
      <Route path="/catalog"       element={
        <AccessGuard sectionId="catalog"><CatalogPage /></AccessGuard>
      } />

      <Route path="*"              element={<Navigate to="/" replace />} />
    </Routes>
  )
}
