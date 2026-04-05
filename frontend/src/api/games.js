/**
 * API-слой для игр
 * TODO: заменить на fetch('/api/games') когда будет готов бэкенд
 */

export const GAME_CATEGORIES = ['Все игры', 'Слоты', 'Карты', 'Рулетка']

/** @typedef {'slots'|'cards'|'roulette'} GameCategory */
/** @typedef {{id:string, title:string, description:string, emoji:string, category:GameCategory, badge:'ХИТ'|'НОВОЕ'|'БЕСПЛАТНО'|null, free:boolean}} Game */

export async function getGames() {
  await new Promise(r => setTimeout(r, 200))
  return [
    {
      id: 'shmon',
      title: 'Шмон',
      description: 'Слоты с тюремными символами',
      emoji: '🎰',
      category: 'Слоты',
      badge: 'ХИТ',
      free: false,
    },
    {
      id: 'smotrящий',
      title: 'Смотрящий',
      description: 'Блэкджек на арго',
      emoji: '🃏',
      category: 'Карты',
      badge: 'НОВОЕ',
      free: false,
    },
    {
      id: 'prigovor',
      title: 'Приговор',
      description: 'Рулетка по статьям УК',
      emoji: '⚖️',
      category: 'Рулетка',
      badge: 'БЕСПЛАТНО',
      free: true,
    },
    {
      id: 'etap',
      title: 'Этап',
      description: 'Угадай срок по делу',
      emoji: '🚌',
      category: 'Карты',
      badge: null,
      free: false,
    },
    {
      id: 'obshak',
      title: 'Общак',
      description: 'Покер по понятиям',
      emoji: '💰',
      category: 'Карты',
      badge: null,
      free: false,
    },
    {
      id: 'zona',
      title: 'Зона',
      description: 'Слоты с зонными символами',
      emoji: '🔒',
      category: 'Слоты',
      badge: null,
      free: false,
    },
  ]
}
