import { create } from 'zustand'
import { takeLoan, repayLoan, LOAN_INTEREST_RATE, LOAN_INTEREST_INTERVAL_MS, LOAN_BAN_LIMIT } from '../api/loan.js'

/**
 * useLoanStore — управление долгом пользователя
 *
 * State:
 *   debt            {number}       — текущий долг (отрицательное число или 0)
 *   nextInterestAt  {number|null}  — timestamp следующего начисления процентов
 *   loading         {boolean}
 *   error           {string|null}
 *
 * Actions:
 *   borrow(amount)  — занять лаве (долг уходит в минус)
 *   repay(amount)   — погасить долг
 *   tickInterest()  — начислить проценты (+5% к долгу)
 *   startTimer()    — запустить интервал начисления процентов
 *   stopTimer()     — остановить интервал
 */

const LOAN_AMOUNT = 500 // фиксированная сумма займа

const useLoanStore = create((set, get) => ({
  // ── State ──────────────────────────────────────────────────────────────────
  debt:           0,
  nextInterestAt: null,
  loading:        false,
  error:          null,
  _timerId:       null,

  // ── Computed ───────────────────────────────────────────────────────────────
  get isBanned() {
    return get().debt <= LOAN_BAN_LIMIT
  },

  get debtPercent() {
    // Процент заполнения до лимита бана (0–100)
    const { debt } = get()
    if (debt >= 0) return 0
    return Math.min(100, Math.round((Math.abs(debt) / Math.abs(LOAN_BAN_LIMIT)) * 100))
  },

  // ── Actions ────────────────────────────────────────────────────────────────
  borrow: async () => {
    set({ loading: true, error: null })
    try {
      await takeLoan(LOAN_AMOUNT)
      const now = Date.now()
      set(s => ({
        loading: false,
        debt: s.debt - LOAN_AMOUNT,
        nextInterestAt: now + LOAN_INTEREST_INTERVAL_MS,
      }))
      get().startTimer()
    } catch (err) {
      set({ loading: false, error: err.message })
    }
  },

  repay: async (amount) => {
    set({ loading: true, error: null })
    try {
      await repayLoan(amount)
      set(s => {
        const newDebt = Math.min(0, s.debt + amount)
        return {
          loading: false,
          debt: newDebt,
          nextInterestAt: newDebt < 0 ? s.nextInterestAt : null,
        }
      })
      if (get().debt >= 0) get().stopTimer()
    } catch (err) {
      set({ loading: false, error: err.message })
    }
  },

  tickInterest: () => {
    set(s => {
      if (s.debt >= 0) return s
      const interest = Math.floor(s.debt * LOAN_INTEREST_RATE) // отрицательное * 0.05 = отрицательное
      const newDebt = s.debt + interest // долг растёт
      return {
        debt: newDebt,
        nextInterestAt: Date.now() + LOAN_INTEREST_INTERVAL_MS,
      }
    })
  },

  startTimer: () => {
    const { _timerId, tickInterest } = get()
    if (_timerId) return // уже запущен
    const id = setInterval(() => {
      tickInterest()
    }, LOAN_INTEREST_INTERVAL_MS)
    set({ _timerId: id })
  },

  stopTimer: () => {
    const { _timerId } = get()
    if (_timerId) {
      clearInterval(_timerId)
      set({ _timerId: null })
    }
  },
}))

export { LOAN_AMOUNT, LOAN_BAN_LIMIT }
export default useLoanStore
