/**
 * API-слой для кредита (долга)
 * TODO: заменить на реальные fetch-запросы когда будет готов бэкенд
 */

/** Процентная ставка — +5% каждые 10 минут */
export const LOAN_INTEREST_RATE = 0.05
export const LOAN_INTEREST_INTERVAL_MS = 10 * 60 * 1000 // 10 минут
export const LOAN_BAN_LIMIT = -5000  // лимит бана

/** Взять кредит */
export async function takeLoan(amount) {
  await new Promise(r => setTimeout(r, 300))
  // TODO: POST /api/loan/take { amount }
  return { success: true, amount }
}

/** Погасить долг */
export async function repayLoan(amount) {
  await new Promise(r => setTimeout(r, 300))
  // TODO: POST /api/loan/repay { amount }
  return { success: true, amount }
}

/** Получить текущий долг */
export async function getLoanStatus() {
  await new Promise(r => setTimeout(r, 200))
  // TODO: GET /api/loan/status
  return {
    debt: 0,
    nextInterestAt: null,
  }
}
