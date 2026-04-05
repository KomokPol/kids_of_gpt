/**
 * Механика званий Zondex
 * Звание определяется по количеству респекта пользователя
 */

export const RANKS = [
  {
    id: 'новичок',
    title: 'Новичок',
    emoji: '🔰',
    minRespect: 0,
    color: '#888888',
  },
  {
    id: 'мужик',
    title: 'Мужик',
    emoji: '👤',
    minRespect: 100,
    color: '#aaaaaa',
  },
  {
    id: 'бродяга',
    title: 'Бродяга',
    emoji: '🎒',
    minRespect: 300,
    color: '#c0a060',
  },
  {
    id: 'авторитет',
    title: 'Авторитет',
    emoji: '⭐',
    minRespect: 700,
    color: '#f5c518',
  },
  {
    id: 'положенец',
    title: 'Положенец',
    emoji: '🏅',
    minRespect: 1500,
    color: '#e0b015',
  },
  {
    id: 'смотрящий',
    title: 'Смотрящий',
    emoji: '👁️',
    minRespect: 3000,
    color: '#9b7fe8',
  },
  {
    id: 'вор',
    title: 'Вор в законе',
    emoji: '♠️',
    minRespect: 6000,
    color: '#e74c3c',
  },
]

/**
 * Получить звание по количеству респекта
 * @param {number} respect
 * @returns {typeof RANKS[0]}
 */
export function getRankByRespect(respect) {
  let rank = RANKS[0]
  for (const r of RANKS) {
    if (respect >= r.minRespect) rank = r
    else break
  }
  return rank
}

/**
 * Получить следующее звание
 * @param {number} respect
 * @returns {typeof RANKS[0] | null}
 */
export function getNextRank(respect) {
  const current = getRankByRespect(respect)
  const idx = RANKS.findIndex(r => r.id === current.id)
  return RANKS[idx + 1] ?? null
}

/**
 * Прогресс до следующего звания (0–100)
 * @param {number} respect
 * @returns {number}
 */
export function getRankProgress(respect) {
  const current = getRankByRespect(respect)
  const next = getNextRank(respect)
  if (!next) return 100
  const range = next.minRespect - current.minRespect
  const progress = respect - current.minRespect
  return Math.min(100, Math.round((progress / range) * 100))
}

/**
 * Проверить, достаточно ли звания для доступа к товару
 * @param {number} respect
 * @param {string} requiredRankId
 * @returns {boolean}
 */
export function hasRankAccess(respect, requiredRankId) {
  if (!requiredRankId) return true
  const userRank = getRankByRespect(respect)
  const requiredRank = RANKS.find(r => r.id === requiredRankId)
  if (!requiredRank) return true
  return respect >= requiredRank.minRespect
}
