import { createContext, useContext, useReducer, useEffect } from 'react'

// ─── Shape ───────────────────────────────────────────────────────────────────
/**
 * @typedef {Object} UserState
 * @property {string|null}  login    — логин пользователя
 * @property {string|null}  iconUrl  — URL аватарки
 * @property {number}       xp       — опыт
 * @property {number}       balance  — баланс в лаве
 * @property {boolean}      loading  — идёт загрузка
 * @property {string|null}  error    — ошибка загрузки
 */

/** @type {UserState} */
const INITIAL_STATE = {
  login:   null,
  iconUrl: null,
  xp:      0,
  balance: 0,
  loading: false,
  error:   null,
}

// ─── Reducer ─────────────────────────────────────────────────────────────────
function userReducer(state, action) {
  switch (action.type) {
    case 'FETCH_START':
      return { ...state, loading: true, error: null }

    case 'FETCH_SUCCESS':
      return {
        ...state,
        loading: false,
        login:   action.payload.login,
        iconUrl: action.payload.iconUrl ?? null,
        xp:      action.payload.xp ?? 0,
        balance: action.payload.balance ?? 0,
      }

    case 'FETCH_ERROR':
      return { ...state, loading: false, error: action.payload }

    case 'SET_BALANCE':
      return { ...state, balance: action.payload }

    case 'SET_XP':
      return { ...state, xp: action.payload }

    default:
      return state
  }
}

// ─── Context ─────────────────────────────────────────────────────────────────
const UserContext = createContext(null)

// ─── Mock fetch (заглушка) ───────────────────────────────────────────────────
/**
 * Заглушка — имитирует GET /api/user
 * Когда будет готов бэкенд, заменить на:
 *   const res = await fetch('/api/user')
 *   return res.json()
 */
async function fetchUserMock() {
  await new Promise(r => setTimeout(r, 400)) // имитация сетевой задержки
  return {
    login:   'Балабанов',
    iconUrl: null,           // null → покажет инициалы
    xp:      575,
    balance: 1240,
  }
}

// ─── Provider ────────────────────────────────────────────────────────────────
export function UserProvider({ children }) {
  const [state, dispatch] = useReducer(userReducer, INITIAL_STATE)

  /**
   * Загрузить данные пользователя.
   * Вызывается при монтировании и может быть вызвана повторно (refresh).
   */
  async function loadUser() {
    dispatch({ type: 'FETCH_START' })
    try {
      const data = await fetchUserMock()
      dispatch({ type: 'FETCH_SUCCESS', payload: data })
    } catch (err) {
      dispatch({ type: 'FETCH_ERROR', payload: err.message })
    }
  }

  // Загружаем при старте приложения
  useEffect(() => {
    loadUser()
  }, [])

  const value = {
    ...state,
    loadUser,
    dispatch,
  }

  return (
    <UserContext.Provider value={value}>
      {children}
    </UserContext.Provider>
  )
}

// ─── Hook ─────────────────────────────────────────────────────────────────────
export function useUser() {
  const ctx = useContext(UserContext)
  if (!ctx) throw new Error('useUser must be used within <UserProvider>')
  return ctx
}
