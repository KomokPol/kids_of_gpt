import { create } from 'zustand'

/**
 * Символы барабана Шмона
 * emoji, id, вес (чем больше — тем чаще выпадает)
 */
export const SYMBOLS = [
  { id: 'diamond', emoji: '💎', label: 'Бриллиант', weight: 1 },
  { id: 'money',   emoji: '💰', label: 'Общак',     weight: 3 },
  { id: 'key',     emoji: '🗝️', label: 'Ключ',      weight: 5 },
  { id: 'lock',    emoji: '🔒', label: 'Замок',     weight: 8 },
]

/**
 * Таблица выплат:
 * три одинаковых символа → множитель ставки
 * два бриллианта (любые 2 💎) → 1.5x
 */
export const PAYOUTS = [
  { match: ['diamond', 'diamond', 'diamond'], multiplier: 10, label: '💎 💎 💎', desc: '× 10 ставки' },
  { match: ['money',   'money',   'money'],   multiplier: 5,  label: '💰 💰 💰', desc: '× 5 ставки' },
  { match: ['key',     'key',     'key'],     multiplier: 3,  label: '🗝️ 🗝️ 🗝️', desc: '× 3 ставки' },
  { match: ['lock',    'lock',    'lock'],    multiplier: 2,  label: '🔒 🔒 🔒', desc: '× 2 ставки' },
  { match: 'two_diamonds',                   multiplier: 1.5, label: 'любые 2 💎', desc: '× 1.5 ставки' },
]

const MIN_BET = 10
const MAX_BET = 500
const BET_STEP = 10
const SPIN_DURATION_MS = 800

// Взвешенный случайный выбор символа
function randomSymbol() {
  const totalWeight = SYMBOLS.reduce((s, sym) => s + sym.weight, 0)
  let r = Math.random() * totalWeight
  for (const sym of SYMBOLS) {
    r -= sym.weight
    if (r <= 0) return sym.id
  }
  return SYMBOLS[SYMBOLS.length - 1].id
}

// Рассчитать выигрыш
function calcWin(reels, bet) {
  const [a, b, c] = reels
  // Три одинаковых
  if (a === b && b === c) {
    const payout = PAYOUTS.find(p => Array.isArray(p.match) && p.match[0] === a)
    if (payout) return Math.floor(bet * payout.multiplier)
  }
  // Два бриллианта
  const diamonds = reels.filter(r => r === 'diamond').length
  if (diamonds >= 2) return Math.floor(bet * 1.5)
  return 0
}

const useShmonStore = create((set, get) => ({
  // ── State ──────────────────────────────────────────────────────────────────
  reels:       ['lock', 'lock', 'lock'],   // текущие символы
  spinning:    false,
  bet:         50,
  lastWin:     null,   // сумма последнего выигрыша (null = ещё не крутили)
  lastResult:  null,   // 'win' | 'lose'

  // Статистика сессии
  wins:        0,
  losses:      0,
  record:      0,      // максимальный выигрыш за сессию
  respectGained: 0,    // респект за сессию

  // ── Actions ────────────────────────────────────────────────────────────────
  setBet: (bet) => {
    const clamped = Math.max(MIN_BET, Math.min(MAX_BET, bet))
    set({ bet: clamped })
  },

  increaseBet: () => {
    const { bet } = get()
    set({ bet: Math.min(MAX_BET, bet + BET_STEP) })
  },

  decreaseBet: () => {
    const { bet } = get()
    set({ bet: Math.max(MIN_BET, bet - BET_STEP) })
  },

  /**
   * spin — крутить барабаны
   * @param {function} onBalanceChange — колбэк(delta) для обновления баланса в useUserStore
   */
  spin: async (onBalanceChange) => {
    const { bet, spinning, record } = get()
    if (spinning) return

    // Списываем ставку
    onBalanceChange(-bet)
    set({ spinning: true, lastWin: null, lastResult: null })

    // Имитация вращения — промежуточные случайные символы
    await new Promise(r => setTimeout(r, SPIN_DURATION_MS))

    // Финальный результат
    const newReels = [randomSymbol(), randomSymbol(), randomSymbol()]
    const win = calcWin(newReels, bet)

    if (win > 0) {
      onBalanceChange(win)
    }

    const respectDelta = win > 0 ? Math.floor(win / 10) : 0

    set(s => ({
      reels:        newReels,
      spinning:     false,
      lastWin:      win,
      lastResult:   win > 0 ? 'win' : 'lose',
      wins:         win > 0 ? s.wins + 1 : s.wins,
      losses:       win === 0 ? s.losses + 1 : s.losses,
      record:       Math.max(record, win),
      respectGained: s.respectGained + respectDelta,
    }))

    return { win, respectDelta }
  },
}))

export { MIN_BET, MAX_BET, BET_STEP }
export default useShmonStore
