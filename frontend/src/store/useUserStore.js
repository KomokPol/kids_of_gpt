import { create } from 'zustand'
import { getUser } from '../api/user.js'

/**
 * useUserStore — Zustand стор пользователя
 *
 * State:
 *   login   {string|null}  — логин
 *   iconUrl {string|null}  — URL аватарки (null → инициалы в Avatar)
 *   respect {number}       — респект (очки опыта)
 *   balance {number}       — баланс в лаве
 *   loading {boolean}      — идёт загрузка
 *   error   {string|null}  — ошибка
 *
 * Actions:
 *   fetchUser()        — загрузить данные пользователя через api/user.js
 *   setBalance(n)      — обновить баланс локально
 *   setRespect(n)      — обновить респект локально
 */
const useUserStore = create((set) => ({
  // ── State ──────────────────────────────────────────────────────────────────
  login:   null,
  iconUrl: null,
  respect: 0,
  balance: 0,
  loading: false,
  error:   null,

  // ── Actions ────────────────────────────────────────────────────────────────
  fetchUser: async () => {
    set({ loading: true, error: null })
    try {
      const data = await getUser()
      set({
        loading:  false,
        login:    data.login   ?? null,
        iconUrl:  data.iconUrl ?? null,
        respect:  data.respect ?? 0,
        balance:  data.balance ?? 0,
      })
    } catch (err) {
      set({ loading: false, error: err.message })
    }
  },

  setBalance: (balance) => set({ balance }),
  setRespect: (respect) => set({ respect }),
}))

export default useUserStore
