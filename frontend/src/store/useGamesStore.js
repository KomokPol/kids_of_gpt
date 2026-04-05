import { create } from 'zustand'
import { getGames, GAME_CATEGORIES } from '../api/games.js'

/**
 * useGamesStore — список игр и активный фильтр категории
 */
const useGamesStore = create((set, get) => ({
  // ── State ──────────────────────────────────────────────────────────────────
  games:          [],
  activeCategory: 'Все игры',
  loading:        false,
  error:          null,

  // ── Computed ───────────────────────────────────────────────────────────────
  get filteredGames() {
    const { games, activeCategory } = get()
    if (activeCategory === 'Все игры') return games
    return games.filter(g => g.category === activeCategory)
  },

  // ── Actions ────────────────────────────────────────────────────────────────
  fetchGames: async () => {
    set({ loading: true, error: null })
    try {
      const games = await getGames()
      set({ games, loading: false })
    } catch (err) {
      set({ loading: false, error: err.message })
    }
  },

  setCategory: (category) => set({ activeCategory: category }),
}))

export default useGamesStore
