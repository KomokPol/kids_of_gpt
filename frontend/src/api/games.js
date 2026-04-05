/**
 * API-слой для игр
 * TODO: заменить на fetch('/api/games') когда будет готов бэкенд
 */

export const GAME_CATEGORIES = ['Все игры', 'Слоты', 'Карты', 'Рулетка']

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
      requiredRank: 'бродяга',   // базовый доступ к играм
    },
    {
      id: 'smotrящий',
      title: 'Смотрящий',
      description: 'Блэкджек на арго',
      emoji: '🃏',
      category: 'Карты',
      badge: 'НОВОЕ',
      free: false,
      requiredRank: 'авторитет',
    },
    {
      id: 'prigovor',
      title: 'Приговор',
      description: 'Рулетка по статьям УК',
      emoji: '⚖️',
      category: 'Рулетка',
      badge: 'БЕСПЛАТНО',
      free: true,
      requiredRank: 'бродяга',
    },
    {
      id: 'etap',
      title: 'Этап',
      description: 'Угадай срок по делу',
      emoji: '🚌',
      category: 'Карты',
      badge: null,
      free: false,
      requiredRank: 'авторитет',
    },
    {
      id: 'obshak',
      title: 'Общак',
      description: 'Покер по понятиям',
      emoji: '💰',
      category: 'Карты',
      badge: null,
      free: false,
      requiredRank: 'положенец',
    },
    {
      id: 'zona',
      title: 'Зона',
      description: 'Слоты с зонными символами',
      emoji: '🔒',
      category: 'Слоты',
      badge: null,
      free: false,
      requiredRank: 'положенец',
    },
  ]
}
