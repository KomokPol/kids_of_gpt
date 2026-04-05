/**
 * API-слой для Барахолки (маркетплейс)
 * TODO: заменить на fetch('/api/market') когда будет готов бэкенд
 */

export const MARKET_CATEGORIES = ['Всё', 'Еда', 'Шмотки', 'Инструменты', 'Статус']

/**
 * @typedef {{
 *   id: string,
 *   title: string,
 *   description: string,
 *   emoji: string,
 *   category: string,
 *   price: number,
 *   requiredRank: string|null,  // null = доступно всем
 *   stock: number,              // -1 = безлимит
 * }} MarketItem
 */

export async function getMarketItems() {
  await new Promise(r => setTimeout(r, 200))
  return [
    // ── Еда (доступна всем) ──────────────────────────────────────────────────
    {
      id: 'balanda',
      title: 'Баланда',
      description: 'Похлёбка неустановленного состава. Сытно.',
      emoji: '🍲',
      category: 'Еда',
      price: 10,
      requiredRank: null,
      stock: -1,
    },
    {
      id: 'bread',
      title: 'Пайка хлеба',
      description: '200г чёрного. Свежесть под вопросом.',
      emoji: '🍞',
      category: 'Еда',
      price: 15,
      requiredRank: null,
      stock: -1,
    },
    {
      id: 'tea',
      title: 'Чай зонный',
      description: 'Крепкий. Валюта и напиток одновременно.',
      emoji: '🍵',
      category: 'Еда',
      price: 30,
      requiredRank: null,
      stock: 50,
    },
    {
      id: 'potato',
      title: 'Варёная картошка',
      description: 'Без соли. Но горячая.',
      emoji: '🥔',
      category: 'Еда',
      price: 20,
      requiredRank: null,
      stock: -1,
    },
    {
      id: 'buckwheat',
      title: 'Гречка с чем-то',
      description: 'Что-то серое сверху.',
      emoji: '🥣',
      category: 'Еда',
      price: 25,
      requiredRank: null,
      stock: -1,
    },
    // ── Шмотки ──────────────────────────────────────────────────────────────
    {
      id: 'tshirt',
      title: 'Роба',
      description: 'Стандартная. Серая. Твоя.',
      emoji: '👕',
      category: 'Шмотки',
      price: 80,
      requiredRank: null,
      stock: 20,
    },
    {
      id: 'boots',
      title: 'Кирзачи',
      description: 'Тяжёлые. Надёжные. Вечные.',
      emoji: '👢',
      category: 'Шмотки',
      price: 150,
      requiredRank: 'мужик',
      stock: 10,
    },
    {
      id: 'jacket',
      title: 'Телогрейка',
      description: 'Тепло. Статусно. Уважение.',
      emoji: '🧥',
      category: 'Шмотки',
      price: 300,
      requiredRank: 'бродяга',
      stock: 5,
    },
    // ── Инструменты ─────────────────────────────────────────────────────────
    {
      id: 'spoon',
      title: 'Ложка',
      description: 'Алюминиевая. Заточенная по краям.',
      emoji: '🥄',
      category: 'Инструменты',
      price: 50,
      requiredRank: null,
      stock: -1,
    },
    {
      id: 'needle',
      title: 'Игла с нитью',
      description: 'Зашить робу. Или что-то ещё.',
      emoji: '🪡',
      category: 'Инструменты',
      price: 40,
      requiredRank: 'мужик',
      stock: 30,
    },
    {
      id: 'radio',
      title: 'Радиоприёмник',
      description: 'Ловит три станции. Одна нормальная.',
      emoji: '📻',
      category: 'Инструменты',
      price: 500,
      requiredRank: 'авторитет',
      stock: 3,
    },
    // ── Статус ───────────────────────────────────────────────────────────────
    {
      id: 'tattoo',
      title: 'Наколка',
      description: 'Самодельная. Значение знают не все.',
      emoji: '💉',
      category: 'Статус',
      price: 200,
      requiredRank: 'бродяга',
      stock: -1,
    },
    {
      id: 'ring',
      title: 'Перстень',
      description: 'Алюминиевый. Но смотрящий оценит.',
      emoji: '💍',
      category: 'Статус',
      price: 800,
      requiredRank: 'авторитет',
      stock: 5,
    },
    {
      id: 'crown',
      title: 'Корона',
      description: 'Символ власти на зоне. Только для своих.',
      emoji: '👑',
      category: 'Статус',
      price: 3000,
      requiredRank: 'смотрящий',
      stock: 1,
    },
  ]
}
