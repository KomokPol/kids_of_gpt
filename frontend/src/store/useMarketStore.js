import { create } from 'zustand'
import { getMarketItems, MARKET_CATEGORIES } from '../api/market.js'

/**
 * useMarketStore — стор барахолки
 */
const useMarketStore = create((set, get) => ({
  // ── State ──────────────────────────────────────────────────────────────────
  items:          [],
  activeCategory: 'Всё',
  loading:        false,
  error:          null,
  // id товара → количество купленных (для отслеживания stock)
  purchased:      {},

  // ── Actions ────────────────────────────────────────────────────────────────
  fetchItems: async () => {
    set({ loading: true, error: null })
    try {
      const items = await getMarketItems()
      set({ items, loading: false })
    } catch (err) {
      set({ loading: false, error: err.message })
    }
  },

  setCategory: (category) => set({ activeCategory: category }),

  /**
   * Купить товар
   * @param {string} itemId
   * @param {function} onBalanceChange — колбэк(delta) для списания с баланса
   * @returns {'ok'|'no_balance'|'no_stock'|'no_rank'}
   */
  buyItem: (itemId, onBalanceChange) => {
    const { items, purchased } = get()
    const item = items.find(i => i.id === itemId)
    if (!item) return 'error'

    // Проверка стока
    const boughtCount = purchased[itemId] ?? 0
    if (item.stock !== -1 && boughtCount >= item.stock) return 'no_stock'

    // Списываем лаве
    onBalanceChange(-item.price)

    set(s => ({
      purchased: {
        ...s.purchased,
        [itemId]: (s.purchased[itemId] ?? 0) + 1,
      },
    }))

    return 'ok'
  },

  // Получить отфильтрованные товары
  getFiltered: (category) => {
    const { items } = get()
    if (category === 'Всё') return items
    return items.filter(i => i.category === category)
  },
}))

export default useMarketStore
