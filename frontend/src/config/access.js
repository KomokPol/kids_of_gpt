/**
 * Матрица доступа к разделам по рангу
 *
 * Каждый раздел требует минимальный ранг.
 * Новичок (0 респекта) — только Шарага и Звания.
 */

import { RANKS, getRankByRespect } from './ranks.js'

/**
 * Конфигурация доступа к разделам
 * requiredRankId: null = доступно всем
 */
export const SECTION_ACCESS = {
    edu: { requiredRankId: null, label: 'Шарага' },
    ranks: { requiredRankId: null, label: 'Звания' },
    market: { requiredRankId: 'мужик', label: 'Барахолка' },
    games: { requiredRankId: 'бродяга', label: 'Игровой зал' },
    'games/shmon': { requiredRankId: 'бродяга', label: 'Шмон' },
    catalog: { requiredRankId: 'авторитет', label: 'ЗОНАФИЛЬМ' },
    'kitchen-sink': { requiredRankId: 'смотрящий', label: 'Kitchen Sink' },
}

/**
 * Проверить доступ к разделу по количеству респекта
 * @param {string} sectionId — ключ из SECTION_ACCESS
 * @param {number} respect
 * @returns {boolean}
 */
export function hasAccess(sectionId, respect) {
    const section = SECTION_ACCESS[sectionId]
    if (!section) return true
    if (!section.requiredRankId) return true
    const required = RANKS.find(r => r.id === section.requiredRankId)
    if (!required) return true
    return respect >= required.minRespect
}

/**
 * Получить требуемый ранг для раздела
 * @param {string} sectionId
 * @returns {typeof RANKS[0] | null}
 */
export function getRequiredRank(sectionId) {
    const section = SECTION_ACCESS[sectionId]
    if (!section?.requiredRankId) return null
    return RANKS.find(r => r.id === section.requiredRankId) ?? null
}
